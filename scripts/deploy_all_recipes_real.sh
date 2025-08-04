#!/bin/bash

# Actual deployment script for all runtime recipes to Zerops
# This will create real projects and deploy applications

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MCP_SERVER="$PROJECT_ROOT/mcp-server"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DEPLOYMENT_LOG="$PROJECT_ROOT/test-reports/real_deployment_log_$TIMESTAMP.txt"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Tracking arrays
CREATED_PROJECTS=()
DEPLOYMENT_RESULTS=()
DEPLOYMENT_URLS=()

# Ensure directories exist
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

# Check prerequisites
check_prerequisites() {
    local errors=0
    
    if [ ! -f "$MCP_SERVER" ]; then
        log_error "MCP server not found. Building..."
        make -C "$PROJECT_ROOT" build || { log_error "Failed to build MCP server"; ((errors++)); }
    fi
    
    if [ -z "$ZEROPS_API_KEY" ]; then
        log_error "ZEROPS_API_KEY not set. Please export your API key."
        ((errors++))
    fi
    
    if ! command -v zcli &> /dev/null; then
        log_error "zcli not found. Please install Zerops CLI from https://app.zerops.io"
        ((errors++))
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq not found. Please install jq for JSON parsing."
        ((errors++))
    fi
    
    return $errors
}

# Parse JSON response to extract value
extract_json_value() {
    local json=$1
    local key=$2
    echo "$json" | jq -r ".$key // empty" 2>/dev/null || echo ""
}

# Call MCP tool and get response
call_mcp_tool() {
    local tool=$1
    local args=$2
    local request="{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool\",\"arguments\":$args}}"
    
    echo "$request" | "$MCP_SERVER" 2>/dev/null || echo "{\"error\":\"Tool call failed\"}"
}

# Deploy a single recipe
deploy_recipe() {
    local runtime_name=$1
    local recipe_dir=$2
    local import_file=$3
    local service_name=$4
    local project_name="demo-${runtime_name,,}-${TIMESTAMP:0:8}"
    
    log_info "=== Deploying $runtime_name Recipe ==="
    log_info "Project name: $project_name"
    log_info "Recipe directory: $recipe_dir"
    
    # Step 1: Validate authentication
    log_info "Validating authentication..."
    local auth_result=$(call_mcp_tool "auth_validate" "{}")
    if echo "$auth_result" | grep -q "error"; then
        log_error "Authentication failed for $runtime_name"
        DEPLOYMENT_RESULTS+=("$runtime_name: FAILED - Authentication error")
        return 1
    fi
    log_success "Authentication validated"
    
    # Step 2: Create project
    log_info "Creating project..."
    local create_args="{\"name\":\"$project_name\",\"region\":\"prg1\"}"
    local create_result=$(call_mcp_tool "project_create" "$create_args")
    
    local project_id=$(echo "$create_result" | grep -o '"project_id":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$project_id" ] || [ "$project_id" = "null" ]; then
        log_error "Failed to create project for $runtime_name"
        log_error "Response: $create_result"
        DEPLOYMENT_RESULTS+=("$runtime_name: FAILED - Project creation failed")
        return 1
    fi
    
    CREATED_PROJECTS+=("$project_id:$project_name")
    log_success "Created project: $project_name (ID: $project_id)"
    
    # Step 3: Import services
    log_info "Importing services from $import_file..."
    local import_yaml=$(cat "$recipe_dir/$import_file")
    # Update project name in YAML
    import_yaml=$(echo "$import_yaml" | sed "s/name: recipe-[^ ]*/name: $project_name/")
    
    # Escape YAML for JSON
    import_yaml_escaped=$(echo "$import_yaml" | jq -Rs .)
    local import_args="{\"project_id\":\"$project_id\",\"yaml\":$import_yaml_escaped}"
    local import_result=$(call_mcp_tool "project_import" "$import_args")
    
    if echo "$import_result" | grep -q "error"; then
        log_error "Failed to import services for $runtime_name"
        log_error "Response: $import_result"
        DEPLOYMENT_RESULTS+=("$runtime_name: FAILED - Service import failed")
        return 1
    fi
    
    log_success "Services imported successfully"
    
    # Step 4: Wait for services to initialize
    log_info "Waiting for services to initialize (45s)..."
    sleep 45
    
    # Step 5: Get service list to find service IDs
    log_info "Getting service list..."
    local list_args="{\"project_id\":\"$project_id\"}"
    local service_list=$(call_mcp_tool "service_list" "$list_args")
    
    # Extract service ID for the main app service
    local service_id=$(echo "$service_list" | jq -r ".services[] | select(.hostname == \"$service_name\") | .id" 2>/dev/null | head -1)
    
    if [ -z "$service_id" ] || [ "$service_id" = "null" ]; then
        log_warning "Could not find service ID for $service_name"
        DEPLOYMENT_RESULTS+=("$runtime_name: PARTIAL - Project created, services imported")
        return 0
    fi
    
    log_success "Found service ID: $service_id"
    
    # Step 6: Check VPN status
    log_info "Checking VPN status..."
    local vpn_status=$(call_mcp_tool "vpn_status" "{}")
    
    local is_connected=false
    if echo "$vpn_status" | grep -q "connected"; then
        # Check if connected to this project
        if echo "$vpn_status" | grep -q "$project_id"; then
            is_connected=true
            log_success "VPN already connected to this project"
        else
            log_warning "VPN connected to different project. Please disconnect and reconnect."
        fi
    fi
    
    if [ "$is_connected" = false ]; then
        log_warning "VPN not connected to project $project_id"
        log_info "Please connect manually with: sudo zcli vpn up --projectId $project_id"
        
        # Still enable subdomain access
        log_info "Enabling subdomain access..."
        local subdomain_args="{\"service_id\":\"$service_id\"}"
        local subdomain_result=$(call_mcp_tool "subdomain_enable" "$subdomain_args")
        
        if echo "$subdomain_result" | grep -q "process_id"; then
            log_info "Subdomain enable process started"
            
            # Wait for process to complete
            sleep 10
            
            # Get subdomain status
            local status_result=$(call_mcp_tool "subdomain_status" "$subdomain_args")
            local subdomain_url=$(echo "$status_result" | grep -o '"url":"[^"]*' | cut -d'"' -f4)
            
            if [ -n "$subdomain_url" ] && [ "$subdomain_url" != "null" ]; then
                log_success "Subdomain enabled: $subdomain_url"
                DEPLOYMENT_URLS+=("$runtime_name: $subdomain_url")
                DEPLOYMENT_RESULTS+=("$runtime_name: CREATED - Project ready at $subdomain_url (deployment pending - VPN required)")
            else
                DEPLOYMENT_RESULTS+=("$runtime_name: CREATED - Project ready (deployment pending - VPN required)")
            fi
        else
            DEPLOYMENT_RESULTS+=("$runtime_name: CREATED - Project ready (deployment pending - VPN required)")
        fi
        
        return 0
    fi
    
    # Step 7: Deploy application
    log_info "Deploying application from $recipe_dir..."
    local deploy_args="{\"working_dir\":\"$recipe_dir\",\"project_id\":\"$project_id\"}"
    local deploy_result=$(call_mcp_tool "deploy_push" "$deploy_args")
    
    if echo "$deploy_result" | grep -q "error"; then
        log_error "Failed to deploy application for $runtime_name"
        log_error "Response: $deploy_result"
        DEPLOYMENT_RESULTS+=("$runtime_name: PARTIAL - Project created, deployment failed")
        return 1
    fi
    
    log_success "Application deployed successfully"
    
    # Step 8: Enable subdomain
    log_info "Enabling subdomain access..."
    local subdomain_args="{\"service_id\":\"$service_id\"}"
    local subdomain_result=$(call_mcp_tool "subdomain_enable" "$subdomain_args")
    
    if echo "$subdomain_result" | grep -q "process_id"; then
        log_info "Subdomain enable process started"
        
        # Wait for process to complete
        sleep 10
        
        # Get subdomain status
        local status_result=$(call_mcp_tool "subdomain_status" "$subdomain_args")
        local subdomain_url=$(echo "$status_result" | grep -o '"url":"[^"]*' | cut -d'"' -f4)
        
        if [ -n "$subdomain_url" ] && [ "$subdomain_url" != "null" ]; then
            log_success "Subdomain enabled: $subdomain_url"
            DEPLOYMENT_URLS+=("$runtime_name: $subdomain_url")
            
            # Test the URL
            log_info "Testing application at $subdomain_url/status..."
            if curl -s -f "$subdomain_url/status" > /dev/null 2>&1; then
                log_success "Application is responding!"
                DEPLOYMENT_RESULTS+=("$runtime_name: SUCCESS - Deployed and running at $subdomain_url")
            else
                log_warning "Application not responding yet (may still be starting)"
                DEPLOYMENT_RESULTS+=("$runtime_name: SUCCESS - Deployed at $subdomain_url (starting up)")
            fi
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
    if ! check_prerequisites; then
        log_error "Prerequisites check failed. Please fix the issues above."
        exit 1
    fi
    
    log_info "Starting deployment of all runtime recipes..."
    echo ""
    
    # Define recipes to deploy
    declare -a RECIPES=(
        "Python|recipe-python|zerops-project-import.yml|api"
        "PHP|recipe-php-main|zerops-project-import.yml|apacheapi"
        "NodeJS|recipe-nodejs-main|zerops-project-import.yml|api"
        "Go|recipe-go-main|zerops-project-import.yml|api"
        "Ruby|recipe-ruby-main|zerops-import.yml|app"
        "Java|recipe-java-main|zerops-project-import.yml|api"
        "DotNet|recipe-dotnet-main|zerops-project-import.yml|api"
        "Rust|recipe-rust-main|zerops-project-import.yml|api"
        "Deno|recipe-deno-main|zerops-project-import.yml|api"
        "Bun|recipe-bun-main|zerops-project-import.yml|api"
        "Elixir|recipe-elixir-main|zerops-project-import.yml|api"
        "Gleam|recipe-gleam-main|zerops-project-import.yml|api"
    )
    
    # Check if we should deploy specific runtime
    if [ -n "$1" ]; then
        log_info "Deploying only $1 runtime"
        for recipe in "${RECIPES[@]}"; do
            IFS='|' read -r name dir import service <<< "$recipe"
            if [ "${name,,}" = "${1,,}" ]; then
                deploy_recipe "$name" "$PROJECT_ROOT/test/fixtures/$dir" "$import" "$service"
                break
            fi
        done
    else
        # Deploy all recipes
        for recipe in "${RECIPES[@]}"; do
            IFS='|' read -r name dir import service <<< "$recipe"
            deploy_recipe "$name" "$PROJECT_ROOT/test/fixtures/$dir" "$import" "$service"
            echo ""
        done
    fi
    
    # Summary
    echo ""
    echo "========================================"
    echo "Deployment Summary"
    echo "========================================"
    echo ""
    
    log_info "Deployment Results:"
    for result in "${DEPLOYMENT_RESULTS[@]}"; do
        echo "  $result"
    done
    
    echo ""
    if [ ${#CREATED_PROJECTS[@]} -gt 0 ]; then
        log_info "Created Projects:"
        for project in "${CREATED_PROJECTS[@]}"; do
            IFS=':' read -r id name <<< "$project"
            echo "  - $name (ID: $id)"
        done
    fi
    
    echo ""
    if [ ${#DEPLOYMENT_URLS[@]} -gt 0 ]; then
        log_info "Application URLs:"
        for url in "${DEPLOYMENT_URLS[@]}"; do
            echo "  $url"
        done
    fi
    
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
    
    if [ ${#DEPLOYMENT_URLS[@]} -gt 0 ]; then
        cat >> "$summary_file" <<EOF

## Application URLs

EOF
        for url in "${DEPLOYMENT_URLS[@]}"; do
            echo "- $url" >> "$summary_file"
        done
    fi
    
    cat >> "$summary_file" <<EOF

## Next Steps

1. Check each application's health endpoint: \`/status\`
2. Monitor logs in Zerops dashboard: https://app.zerops.io
3. Test database connectivity for each runtime
4. Clean up test projects when done

## Cleanup Commands

To delete all created projects:
\`\`\`bash
# For each project ID above:
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"project_delete","arguments":{"project_id":"PROJECT_ID","confirm":true}}}' | ./mcp-server
\`\`\`

## VPN Commands

If you need to deploy code:
\`\`\`bash
# Connect to a project
sudo zcli vpn up --projectId PROJECT_ID

# Disconnect when done
sudo zcli vpn down
\`\`\`
EOF
    
    log_success "Summary report saved to: $summary_file"
    
    # Show final instructions
    echo ""
    echo "========================================"
    echo "Deployment Complete!"
    echo "========================================"
    echo ""
    echo "To view your deployed applications:"
    echo "1. Go to https://app.zerops.io"
    echo "2. Check the created projects listed above"
    echo "3. Test each application using the URLs provided"
    echo ""
    
    if echo "${DEPLOYMENT_RESULTS[@]}" | grep -q "VPN required"; then
        echo "Note: Some deployments require VPN connection to complete."
        echo "Connect using: sudo zcli vpn up --projectId PROJECT_ID"
        echo ""
    fi
}

# Run main function
main "$@"