#!/bin/bash

# Deploy all runtime recipes - fixed version
set -e

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "================================================"
echo "DEPLOYING ALL 12 RUNTIME RECIPES TO ZEROPS"
echo "================================================"
echo ""

RUNTIMES=("python" "nodejs" "go" "php" "ruby" "java" "rust" "dotnet" "deno" "bun" "elixir" "gleam")
DEPLOYED_COUNT=0
FAILED_COUNT=0

for runtime in "${RUNTIMES[@]}"; do
    echo "================================================"
    echo "Deploying $runtime..."
    echo "================================================"
    
    export DEPLOY_RUNTIME="$runtime"
    
    if /usr/local/go/bin/go test -v ./test/integration -run TestDeployRuntime -timeout 2m 2>&1; then
        ((DEPLOYED_COUNT++))
        echo "✅ $runtime deployed successfully"
    else
        ((FAILED_COUNT++))
        echo "❌ $runtime deployment failed"
    fi
    
    echo ""
    sleep 2
done

echo ""
echo "================================================"
echo "DEPLOYMENT SUMMARY"
echo "================================================"
echo ""
echo "Successfully deployed: $DEPLOYED_COUNT of 12 runtimes"
echo "Failed: $FAILED_COUNT"
echo ""
echo "To view your projects:"
echo "1. Go to https://app.zerops.io"
echo "2. Look for projects starting with 'demo-'"
echo ""