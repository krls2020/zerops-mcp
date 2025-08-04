#!/bin/bash

# Final deployment script for all 12 runtimes
set -e

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "================================================"
echo "DEPLOYING ALL 12 RUNTIME RECIPES TO ZEROPS"
echo "================================================"
echo ""

RUNTIMES=("python" "nodejs" "go" "php" "ruby" "java" "rust" "dotnet" "deno" "bun" "elixir" "gleam")
CREATED_PROJECTS=()

# Deploy each runtime
for runtime in "${RUNTIMES[@]}"; do
    echo ""
    echo "================================================"
    echo "Deploying $runtime..."
    echo "================================================"
    
    export DEPLOY_RUNTIME="$runtime"
    
    # Run the test and capture output
    OUTPUT=$(/usr/local/go/bin/go test -v ./test/integration -run TestDeployRuntime -timeout 2m 2>&1 || true)
    
    # Check if project was created successfully
    if echo "$OUTPUT" | grep -q "PROJECT CREATED SUCCESSFULLY"; then
        # Extract project info
        PROJECT_NAME=$(echo "$OUTPUT" | grep "Name: demo-" | awk '{print $2}')
        PROJECT_ID=$(echo "$OUTPUT" | grep "ID:" | tail -1 | awk '{print $2}')
        
        if [ -n "$PROJECT_NAME" ] && [ -n "$PROJECT_ID" ]; then
            CREATED_PROJECTS+=("$runtime:$PROJECT_NAME:$PROJECT_ID")
            echo "✅ $runtime deployed successfully"
            echo "   Project: $PROJECT_NAME"
            echo "   ID: $PROJECT_ID"
        else
            echo "⚠️  $runtime project created but info not captured"
        fi
    else
        echo "❌ $runtime deployment failed"
        echo "$OUTPUT" | grep -E "(Failed|Error|error)"
    fi
    
    # Small delay between deployments
    sleep 2
done

echo ""
echo ""
echo "================================================"
echo "DEPLOYMENT SUMMARY"
echo "================================================"
echo ""
echo "Successfully created ${#CREATED_PROJECTS[@]} projects:"
echo ""

for project in "${CREATED_PROJECTS[@]}"; do
    IFS=':' read -r runtime name id <<< "$project"
    echo "✅ $runtime"
    echo "   Project: $name"
    echo "   ID: $id"
    echo ""
done

echo "================================================"
echo "VIEW YOUR PROJECTS"
echo "================================================"
echo ""
echo "1. Go to: https://app.zerops.io"
echo "2. Log in with your Zerops account"
echo "3. You will see all the demo projects listed above"
echo ""
echo "Each project contains:"
echo "- Application service for the specific runtime"
echo "- PostgreSQL database service"
echo "- Core service (system)"
echo ""
echo "Projects are named: demo-{runtime}-{timestamp}"
echo ""