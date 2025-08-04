#!/bin/bash

# Deploy recipes using MCP tools directly
set -e

cd "$(dirname "$0")"

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"
TIMESTAMP=$(date +%s)

echo "Deploying runtime recipes to Zerops using MCP tools..."
echo ""

# Function to create and deploy a project
deploy_project() {
    local runtime=$1
    local recipe_dir=$2
    local import_file=$3
    local service_name=$4
    local project_name="demo-$runtime-$TIMESTAMP"
    
    echo "========================================="
    echo "Deploying $runtime"
    echo "========================================="
    
    # Create project
    echo "Creating project: $project_name"
    PROJECT_RESULT=$(cat <<EOF | ./mcp-server 2>/dev/null | grep -v "Starting" || true
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"project_create","arguments":{"name":"$project_name","region":"prg1"}}}
EOF
)
    
    PROJECT_ID=$(echo "$PROJECT_RESULT" | jq -r '.result.project_id // empty' 2>/dev/null || echo "")
    
    if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "null" ]; then
        echo "Failed to create project for $runtime"
        echo "Response: $PROJECT_RESULT"
        return 1
    fi
    
    echo "✓ Created project: $project_name (ID: $PROJECT_ID)"
    
    # Read and prepare import YAML
    IMPORT_YAML=$(cat "test/fixtures/$recipe_dir/$import_file")
    IMPORT_YAML=$(echo "$IMPORT_YAML" | sed "s/name: recipe-[^ ]*/name: $project_name/")
    
    # Import services
    echo "Importing services..."
    YAML_ESCAPED=$(echo "$IMPORT_YAML" | jq -Rs .)
    
    IMPORT_RESULT=$(cat <<EOF | ./mcp-server 2>/dev/null | grep -v "Starting" || true
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"project_import","arguments":{"project_id":"$PROJECT_ID","yaml":$YAML_ESCAPED}}}
EOF
)
    
    if echo "$IMPORT_RESULT" | grep -q '"error"'; then
        echo "Failed to import services"
        echo "Response: $IMPORT_RESULT"
        return 1
    fi
    
    echo "✓ Services imported"
    
    # Wait for services
    echo "Waiting for services to start..."
    sleep 20
    
    # Get service list
    SERVICE_LIST=$(cat <<EOF | ./mcp-server 2>/dev/null | grep -v "Starting" || true
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"service_list","arguments":{"project_id":"$PROJECT_ID"}}}
EOF
)
    
    # Extract service ID
    SERVICE_ID=$(echo "$SERVICE_LIST" | jq -r ".result.services[] | select(.name == \"$service_name\") | .id" 2>/dev/null | head -1)
    
    if [ -n "$SERVICE_ID" ] && [ "$SERVICE_ID" != "null" ]; then
        # Enable subdomain
        echo "Enabling subdomain access..."
        SUBDOMAIN_RESULT=$(cat <<EOF | ./mcp-server 2>/dev/null | grep -v "Starting" || true
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"subdomain_enable","arguments":{"service_id":"$SERVICE_ID"}}}
EOF
)
        
        if echo "$SUBDOMAIN_RESULT" | grep -q "process_id"; then
            echo "✓ Subdomain enable initiated"
        fi
    fi
    
    echo "✓ $runtime project ready in Zerops"
    echo "  Project: $project_name"
    echo "  ID: $PROJECT_ID"
    echo ""
    
    return 0
}

# Deploy all recipes
echo "Starting deployment of all runtime recipes..."
echo ""

# Deploy each runtime
deploy_project "python" "recipe-python" "zerops-project-import.yml" "api"
deploy_project "nodejs" "recipe-nodejs-main" "zerops-project-import.yml" "api"
deploy_project "go" "recipe-go-main" "zerops-project-import.yml" "api"
deploy_project "php" "recipe-php-main" "zerops-project-import.yml" "apacheapi"
deploy_project "ruby" "recipe-ruby-main" "zerops-import.yml" "app"
deploy_project "java" "recipe-java-main" "zerops-project-import.yml" "api"
deploy_project "rust" "recipe-rust-main" "zerops-project-import.yml" "api"
deploy_project "dotnet" "recipe-dotnet-main" "zerops-project-import.yml" "api"
deploy_project "deno" "recipe-deno-main" "zerops-project-import.yml" "api"
deploy_project "bun" "recipe-bun-main" "zerops-project-import.yml" "api"
deploy_project "elixir" "recipe-elixir-main" "zerops-project-import.yml" "api"
deploy_project "gleam" "recipe-gleam-main" "zerops-project-import.yml" "api"

echo ""
echo "========================================="
echo "ALL DEPLOYMENTS COMPLETE!"
echo "========================================="
echo ""
echo "View your projects at: https://app.zerops.io"
echo ""
echo "All projects are prefixed with 'demo-' followed by the runtime name."
echo "Each project includes:"
echo "- Application service running the specific runtime"
echo "- PostgreSQL database service"
echo "- Subdomain access (where enabled)"
echo ""
echo "To deploy code to any project:"
echo "1. Note the project ID from above"
echo "2. Connect VPN: sudo zcli vpn up --projectId PROJECT_ID"
echo "3. Go to recipe: cd test/fixtures/recipe-RUNTIME"
echo "4. Deploy: zcli push"
echo ""