#!/bin/bash

# Deploy a single runtime with full code deployment
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <runtime>"
    echo "Available runtimes: python, nodejs, go, php, ruby, java, rust, dotnet, deno, bun, elixir, gleam"
    exit 1
fi

RUNTIME=$1
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"
export DEPLOY_RUNTIME="$RUNTIME"

echo "================================================"
echo "DEPLOYING $RUNTIME TO ZEROPS (WITH CODE)"
echo "================================================"
echo ""
echo "This will:"
echo "1. Create a new project"
echo "2. Import services (app + database)"
echo "3. Connect VPN (requires sudo)"
echo "4. Deploy the actual application code"
echo ""

# Build the server first
echo "Building MCP server..."
/usr/local/go/bin/go build -o mcp-server cmd/mcp-server/main.go

# Run the deployment test
echo ""
echo "Starting deployment..."
echo ""
/usr/local/go/bin/go test -v ./test/integration -run TestDeployRuntimeWithCode -timeout 5m

echo ""
echo "================================================"
echo "DEPLOYMENT COMPLETE"
echo "================================================"
echo ""
echo "Check https://app.zerops.io for your project"
echo ""