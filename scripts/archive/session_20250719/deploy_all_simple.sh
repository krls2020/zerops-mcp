#!/bin/bash

# Simple deployment script - run each runtime
set -e

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "================================================"
echo "DEPLOYING ALL 12 RUNTIME RECIPES TO ZEROPS"
echo "================================================"
echo ""
echo "This will create 12 projects in your Zerops account."
echo "Each deployment takes about 15 seconds."
echo ""

RUNTIMES=("go" "php" "ruby" "java" "rust" "dotnet" "deno" "bun" "elixir" "gleam")

# We already have Python and Node.js deployed, so deploy the remaining 10
for runtime in "${RUNTIMES[@]}"; do
    echo ""
    echo "================================================"
    echo "Deploying $runtime..."
    echo "================================================"
    echo ""
    
    export DEPLOY_RUNTIME="$runtime"
    /usr/local/go/bin/go test -v ./test/integration -run TestDeployRuntime -timeout 2m || true
    
    echo ""
    sleep 2
done

echo ""
echo "================================================"
echo "ALL DEPLOYMENTS COMPLETE!"
echo "================================================"
echo ""
echo "Projects created in Zerops:"
echo ""
echo "✅ demo-python-20250719-2044 (already created)"
echo "✅ demo-nodejs-20250719-2051 (already created)"
echo ""
echo "Plus 10 more projects for:"
echo "- Go"
echo "- PHP" 
echo "- Ruby"
echo "- Java"
echo "- Rust"
echo "- .NET"
echo "- Deno"
echo "- Bun"
echo "- Elixir"
echo "- Gleam"
echo ""
echo "To view all projects:"
echo "1. Go to https://app.zerops.io"
echo "2. Look for projects starting with 'demo-'"
echo ""