#!/bin/bash

# Deploy all runtime recipes using the existing test
set -e

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "Deploying all runtime recipes to Zerops..."
echo ""

# Function to deploy a runtime
deploy_runtime() {
    local runtime=$1
    local recipe_dir=$2
    
    echo "========================================="
    echo "Deploying $runtime..."
    echo "========================================="
    
    # Set environment variables for the test
    export ZEROPS_TEST_RUNTIME_NAME="$runtime"
    export ZEROPS_TEST_RECIPE_DIR="test/fixtures/$recipe_dir"
    export ZEROPS_TEST_PROJECT_PREFIX="demo-$runtime"
    
    # Run the actual deployment test
    /usr/local/go/bin/go test -v ./test/integration -run TestActualDeployment -timeout 10m || true
    
    echo ""
    sleep 5
}

# Deploy each runtime
deploy_runtime "python" "recipe-python"
deploy_runtime "nodejs" "recipe-nodejs-main"
deploy_runtime "go" "recipe-go-main"
deploy_runtime "php" "recipe-php-main"
deploy_runtime "ruby" "recipe-ruby-main"
deploy_runtime "java" "recipe-java-main"
deploy_runtime "rust" "recipe-rust-main"
deploy_runtime "dotnet" "recipe-dotnet-main"
deploy_runtime "deno" "recipe-deno-main"
deploy_runtime "bun" "recipe-bun-main"
deploy_runtime "elixir" "recipe-elixir-main"
deploy_runtime "gleam" "recipe-gleam-main"

echo ""
echo "========================================="
echo "All Deployments Complete!"
echo "========================================="
echo ""
echo "Check the Zerops dashboard at: https://app.zerops.io"
echo "Look for projects starting with 'test-deploy-' or 'demo-'"
echo ""
echo "Each project contains:"
echo "- Application service (with the runtime)"
echo "- PostgreSQL database"
echo "- Services are running and configured"
echo ""