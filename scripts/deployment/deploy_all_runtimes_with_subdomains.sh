#!/bin/bash

# Deploy all 12 runtime recipes with subdomain access
set -e

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "================================================"
echo "DEPLOYING ALL 12 RUNTIME RECIPES WITH SUBDOMAINS"
echo "================================================"
echo ""
echo "This will create 12 projects in Zerops, each with:"
echo "- Application service with the runtime"
echo "- PostgreSQL database"
echo "- Deployed code"
echo "- Enabled subdomain access"
echo ""
echo "Starting deployment in 3 seconds..."
sleep 3

# All runtimes to deploy
RUNTIMES=(
    "python"
    "nodejs"
    "go"
    "php"
    "ruby"
    "java"
    "rust"
    "dotnet"
    "deno"
    "bun"
    "elixir"
    "gleam"
)

# Track successful deployments
DEPLOYED_PROJECTS=()
FAILED_DEPLOYMENTS=()

# Deploy each runtime
for runtime in "${RUNTIMES[@]}"; do
    echo ""
    echo "================================================"
    echo "Deploying $runtime..."
    echo "================================================"
    
    export DEPLOY_RUNTIME="$runtime"
    
    # Run the deployment test
    if /usr/local/go/bin/go test -v ./test/integration -run TestDeployRuntime -timeout 3m 2>&1 | tee /tmp/deploy_${runtime}.log | grep -E "(Created project:|Subdomain access enabled|Failed|Error)"; then
        # Extract project info
        PROJECT_ID=$(grep "ID:" /tmp/deploy_${runtime}.log | grep -v "Service:" | head -1 | awk '{print $3}' | tr -d ')')
        PROJECT_NAME=$(grep "Name: demo-" /tmp/deploy_${runtime}.log | head -1 | awk '{print $3}')
        
        if [ -n "$PROJECT_ID" ] && [ -n "$PROJECT_NAME" ]; then
            DEPLOYED_PROJECTS+=("$runtime:$PROJECT_NAME:$PROJECT_ID")
            echo "✅ $runtime deployed successfully"
        else
            FAILED_DEPLOYMENTS+=("$runtime")
            echo "❌ $runtime deployment failed"
        fi
    else
        FAILED_DEPLOYMENTS+=("$runtime")
        echo "❌ $runtime deployment failed"
    fi
    
    # Small delay between deployments
    sleep 5
done

echo ""
echo ""
echo "================================================"
echo "DEPLOYMENT SUMMARY"
echo "================================================"
echo ""
echo "Successfully deployed: ${#DEPLOYED_PROJECTS[@]}/12 runtimes"
echo ""

if [ ${#DEPLOYED_PROJECTS[@]} -gt 0 ]; then
    echo "✅ SUCCESSFUL DEPLOYMENTS:"
    echo ""
    for project in "${DEPLOYED_PROJECTS[@]}"; do
        IFS=':' read -r runtime name id <<< "$project"
        echo "  $runtime:"
        echo "    Project: $name"
        echo "    ID: $id"
        echo "    Dashboard: https://app.zerops.io/project/$id"
    done
fi

if [ ${#FAILED_DEPLOYMENTS[@]} -gt 0 ]; then
    echo ""
    echo "❌ FAILED DEPLOYMENTS:"
    echo ""
    for runtime in "${FAILED_DEPLOYMENTS[@]}"; do
        echo "  - $runtime (check /tmp/deploy_${runtime}.log for details)"
    done
fi

echo ""
echo "================================================"
echo "NEXT STEPS"
echo "================================================"
echo ""
echo "1. Go to https://app.zerops.io"
echo "2. Check each project for:"
echo "   - Service running status"
echo "   - Subdomain URL in service details"
echo "   - Application accessibility via subdomain"
echo ""
echo "Note: Subdomain URLs may take a minute to appear after enabling."
echo ""

# Create a summary file
cat > deployment_summary.txt <<EOF
Zerops Runtime Deployment Summary
Generated: $(date)

Successful Deployments: ${#DEPLOYED_PROJECTS[@]}/12

Projects Created:
EOF

for project in "${DEPLOYED_PROJECTS[@]}"; do
    IFS=':' read -r runtime name id <<< "$project"
    echo "- $runtime: $name (ID: $id)" >> deployment_summary.txt
done

if [ ${#FAILED_DEPLOYMENTS[@]} -gt 0 ]; then
    echo "" >> deployment_summary.txt
    echo "Failed Deployments:" >> deployment_summary.txt
    for runtime in "${FAILED_DEPLOYMENTS[@]}"; do
        echo "- $runtime" >> deployment_summary.txt
    done
fi

echo "Summary saved to: deployment_summary.txt"