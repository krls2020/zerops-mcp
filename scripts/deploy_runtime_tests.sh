#!/bin/bash

# Script to deploy all runtime test recipes using Zerops MCP server
# This script simulates the MCP tool calls that would be made during deployment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FIXTURES_DIR="$PROJECT_ROOT/test/fixtures"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if MCP server is available
check_mcp_server() {
    echo -e "${YELLOW}Checking MCP server availability...${NC}"
    if ! pgrep -f "mcp-server" > /dev/null; then
        echo -e "${RED}MCP server is not running. Please start it first.${NC}"
        exit 1
    fi
    echo -e "${GREEN}MCP server is running.${NC}"
}

# Deploy a single runtime recipe
deploy_recipe() {
    local recipe_name=$1
    local recipe_dir="$FIXTURES_DIR/recipe-${recipe_name}"
    local project_name="test-${recipe_name}-$(date +%s)"
    
    echo -e "\n${YELLOW}Deploying ${recipe_name} recipe...${NC}"
    echo "Recipe directory: $recipe_dir"
    echo "Project name: $project_name"
    
    # Check if directory exists
    if [ ! -d "$recipe_dir" ]; then
        echo -e "${RED}Recipe directory not found: $recipe_dir${NC}"
        return 1
    fi
    
    # Check for import file
    local import_file=""
    if [ -f "$recipe_dir/zerops-project-import.yml" ]; then
        import_file="$recipe_dir/zerops-project-import.yml"
    elif [ -f "$recipe_dir/zerops-import.yml" ]; then
        import_file="$recipe_dir/zerops-import.yml"
    else
        echo -e "${RED}No import file found for $recipe_name${NC}"
        return 1
    fi
    
    echo "Using import file: $import_file"
    
    # Simulate MCP tool calls
    echo -e "\n${YELLOW}Simulating MCP tool calls:${NC}"
    echo "1. auth_validate"
    echo "2. project_create (name: $project_name)"
    echo "3. project_import (file: $import_file)"
    echo "4. vpn_connect (project: $project_name)"
    echo "5. deploy_push (dir: $recipe_dir)"
    echo "6. service_info (checking deployment status)"
    echo "7. subdomain_status (checking accessibility)"
    
    echo -e "${GREEN}✓ ${recipe_name} deployment simulation complete${NC}"
}

# Main execution
main() {
    echo -e "${YELLOW}=== Zerops Runtime Test Deployment Script ===${NC}"
    echo "This script simulates deploying all runtime test recipes"
    echo "using Zerops MCP server tools."
    
    # Check MCP server
    check_mcp_server
    
    # List of all runtime recipes to deploy
    # Note: Python uses different directory name pattern
    RUNTIMES=(
        "python"
        "php-main"
        "gleam-main"
        "dotnet-main"
        "deno-main"
        "elixir-main"
        "ruby-main"
        "rust-main"
        "java-main"
        "bun-main"
        "nodejs-main"
        "go-main"
    )
    
    echo -e "\n${YELLOW}Will deploy the following runtimes:${NC}"
    printf '%s\n' "${RUNTIMES[@]}"
    
    # Deploy each runtime
    for runtime in "${RUNTIMES[@]}"; do
        deploy_recipe "$runtime"
    done
    
    echo -e "\n${GREEN}=== All deployments simulated successfully ===${NC}"
    echo "To actually deploy, use the integration test or call MCP tools directly."
}

# Run main function
main "$@"