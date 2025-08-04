#!/bin/bash

# Script to deploy all runtime recipes to Zerops for demonstration
# This will create actual projects and deploy real applications

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MCP_SERVER="$PROJECT_ROOT/mcp-server"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DEPLOYMENT_LOG="$PROJECT_ROOT/test-reports/deployment_log_$TIMESTAMP.txt"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Project tracking
CREATED_PROJECTS=()
DEPLOYMENT_RESULTS=()

# Ensure test-reports directory exists
mkdir -p "$PROJECT_ROOT/test-reports"

# Start logging
exec > >(tee -a "$DEPLOYMENT_LOG")
exec 2>&1

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to call MCP tool via test script
call_mcp_tool() {
    local tool=$1
    shift
    local args="$@"
    
    log_info "Calling MCP tool: $tool $args"
    
    # Using the test script to call tools
    cd "$PROJECT_ROOT"
    echo "$args" | ./scripts/test_mcp_tools.sh "$tool" 2>&1 || true
}

# Deploy a single recipe
deploy_recipe() {
    local runtime_name=$1
    local recipe_dir=$2
    local import_file=$3
    local service_name=$4
    local project_name="demo-${runtime_name,,}-${TIMESTAMP}"
    
    log_info "=== Deploying $runtime_name Recipe ==="
    log_info "Project name: $project_name"
    log_info "Recipe directory: $recipe_dir"
    
    # Step 1: Create project
    log_info "Creating project..."
    local create_result=$(cat <<EOF | "$MCP_SERVER" 2>&1 || true
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"project_create","arguments":{"name":"$project_name","region":"prg1"}}}
EOF
)
    
    # Extract project ID from result (this is simplified - in real usage you'd parse JSON)
    local project_id=$(echo "$create_result" | grep -o '"project_id":"[^"]*' | cut -d'"' -f4 || echo "")
    
    if [ -z "$project_id" ]; then
        log_error "Failed to create project for $runtime_name"
        DEPLOYMENT_RESULTS+=("$runtime_name: FAILED - Project creation failed")
        return 1
    fi
    
    CREATED_PROJECTS+=("$project_id:$project_name")
    log_success "Created project: $project_name (ID: $project_id)"
    
    # Step 2: Import services
    log_info "Importing services from $import_file..."
    local import_yaml=$(cat "$recipe_dir/$import_file")
    # Update project name in YAML
    import_yaml=$(echo "$import_yaml" | sed "s/name: recipe-[^ ]*/name: $project_name/")
    
    local import_result=$(cat <<EOF | "$MCP_SERVER" 2>&1 || true
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"project_import","arguments":{"project_id":"$project_id","yaml":"$(echo "$import_yaml" | jq -Rs .)"}}}
EOF
)
    
    if echo "$import_result" | grep -q "error"; then
        log_error "Failed to import services for $runtime_name"
        DEPLOYMENT_RESULTS+=("$runtime_name: FAILED - Service import failed")
        return 1
    fi
    
    log_success "Services imported successfully"
    
    # Step 3: Wait for services to initialize
    log_info "Waiting for services to initialize (30s)..."
    sleep 30
    
    # Step 4: Connect VPN (if not already connected)
    log_info "Checking VPN status..."
    local vpn_status=$(cat <<EOF | "$MCP_SERVER" 2>&1 || true
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vpn_status","arguments":{}}}
EOF
)
    
    if ! echo "$vpn_status" | grep -q "connected"; then
        log_warning "VPN not connected. Please connect manually with: sudo zcli vpn up --projectId $project_id"
        log_info "Skipping deployment for $runtime_name (VPN required)"
        DEPLOYMENT_RESULTS+=("$runtime_name: CREATED - Project ready, deployment pending (VPN required)")
        return 0
    fi
    
    # Step 5: Deploy application
    log_info "Deploying application from $recipe_dir..."
    local deploy_result=$(cat <<EOF | "$MCP_SERVER" 2>&1 || true
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"deploy_push","arguments":{"working_dir":"$recipe_dir","project_id":"$project_id"}}}
EOF
)
    
    if echo "$deploy_result" | grep -q "error"; then
        log_error "Failed to deploy application for $runtime_name"
        DEPLOYMENT_RESULTS+=("$runtime_name: PARTIAL - Project created, deployment failed")
        return 1
    fi
    
    log_success "Application deployed successfully"
    
    # Step 6: Enable subdomain
    log_info "Enabling subdomain access..."
    # Get service ID first
    local service_list=$(cat <<EOF | "$MCP_SERVER" 2>&1 || true
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"service_list","arguments":{"project_id":"$project_id"}}}
EOF
)
    
    # Find the app service (simplified parsing)
    local service_id=$(echo "$service_list" | grep -B2 "\"hostname\":\"$service_name\"" | grep '"id"' | head -1 | cut -d'"' -f4 || echo "")
    
    if [ -n "$service_id" ]; then
        local subdomain_result=$(cat <<EOF | "$MCP_SERVER" 2>&1 || true
{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"subdomain_enable","arguments":{"service_id":"$service_id"}}}
EOF
)
        
        local subdomain_url=$(echo "$subdomain_result" | grep -o '"url":"[^"]*' | cut -d'"' -f4 || echo "")
        if [ -n "$subdomain_url" ]; then
            log_success "Subdomain enabled: $subdomain_url"
            DEPLOYMENT_RESULTS+=("$runtime_name: SUCCESS - $subdomain_url")
        else
            DEPLOYMENT_RESULTS+=("$runtime_name: SUCCESS - Deployed (subdomain pending)")
        fi
    else
        DEPLOYMENT_RESULTS+=("$runtime_name: SUCCESS - Deployed")
    fi
    
    return 0
}

# Main execution
main() {
    echo "========================================"
    echo "Zerops Runtime Recipe Deployment"
    echo "========================================"
    echo "Timestamp: $(date)"
    echo "Log file: $DEPLOYMENT_LOG"
    echo ""
    
    # Check prerequisites
    if [ ! -f "$MCP_SERVER" ]; then
        log_error "MCP server not found. Building..."
        make -C "$PROJECT_ROOT" build
    fi
    
    if [ -z "$ZEROPS_API_KEY" ]; then
        log_error "ZEROPS_API_KEY not set. Please export your API key."
        exit 1
    fi
    
    # Check if zcli is available
    if ! command -v zcli &> /dev/null; then
        log_error "zcli not found. Please install Zerops CLI."
        exit 1
    fi
    
    log_info "Starting deployment of all runtime recipes..."
    
    # Define recipes to deploy
    declare -a RECIPES=(
        "Python|recipe-python|zerops-project-import.yml|api"
        "PHP|recipe-php-main|zerops-project-import.yml|apacheapi"
        "Gleam|recipe-gleam-main|zerops-project-import.yml|api"
        "DotNet|recipe-dotnet-main|zerops-project-import.yml|api"
        "Deno|recipe-deno-main|zerops-project-import.yml|api"
        "Elixir|recipe-elixir-main|zerops-project-import.yml|api"
        "Ruby|recipe-ruby-main|zerops-import.yml|app"
        "Rust|recipe-rust-main|zerops-project-import.yml|api"
        "Java|recipe-java-main|zerops-project-import.yml|api"
        "Bun|recipe-bun-main|zerops-project-import.yml|api"
        "NodeJS|recipe-nodejs-main|zerops-project-import.yml|api"
        "Go|recipe-go-main|zerops-project-import.yml|api"
    )
    
    # Deploy each recipe
    for recipe in "${RECIPES[@]}"; do
        IFS='|' read -r name dir import service <<< "$recipe"
        deploy_recipe "$name" "$PROJECT_ROOT/test/fixtures/$dir" "$import" "$service"
        echo ""
    done
    
    # Summary
    echo "========================================"
    echo "Deployment Summary"
    echo "========================================"
    echo ""
    
    for result in "${DEPLOYMENT_RESULTS[@]}"; do
        echo "$result"
    done
    
    echo ""
    echo "Created Projects:"
    for project in "${CREATED_PROJECTS[@]}"; do
        IFS=':' read -r id name <<< "$project"
        echo "  - $name (ID: $id)"
    done
    
    echo ""
    log_info "Deployment log saved to: $DEPLOYMENT_LOG"
    
    # Create summary report
    local summary_file="$PROJECT_ROOT/test-reports/deployment_summary_$TIMESTAMP.md"
    cat > "$summary_file" <<EOF
# Zerops Runtime Deployment Summary

**Date**: $(date)
**Total Recipes**: ${#RECIPES[@]}

## Deployment Results

| Runtime | Status | URL/Notes |
|---------|--------|-----------|
EOF
    
    for result in "${DEPLOYMENT_RESULTS[@]}"; do
        IFS=':' read -r runtime status <<< "$result"
        echo "| $runtime | $status |" >> "$summary_file"
    done
    
    cat >> "$summary_file" <<EOF

## Created Projects

EOF
    
    for project in "${CREATED_PROJECTS[@]}"; do
        IFS=':' read -r id name <<< "$project"
        echo "- **$name** (ID: \`$id\`)" >> "$summary_file"
    done
    
    cat >> "$summary_file" <<EOF

## Next Steps

1. Check each project in Zerops GUI: https://app.zerops.io
2. Test each application's health endpoint
3. Monitor logs and performance
4. Clean up test projects when done

## Cleanup Command

To delete all created projects:
\`\`\`bash
# Use project_delete tool for each project ID
\`\`\`
EOF
    
    log_success "Summary report saved to: $summary_file"
}

# Run main function
main "$@"