#!/bin/bash

# Deploy a single runtime and verify it's working
set -e

RUNTIME=${1:-python}
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"
export DEPLOY_RUNTIME="$RUNTIME"

echo "================================================"
echo "DEPLOYING $RUNTIME TO ZEROPS WITH FULL CODE"
echo "================================================"
echo ""

# Run the deployment test
echo "Starting deployment..."
/usr/local/go/bin/go test -v ./test/integration -run TestDeployRuntime -timeout 5m 2>&1 | tee /tmp/deploy_${RUNTIME}.log

# Extract project info from log
PROJECT_ID=$(grep "ID:" /tmp/deploy_${RUNTIME}.log | grep -v "Service:" | head -1 | awk '{print $3}' | tr -d ')')
PROJECT_NAME=$(grep "Name: demo-" /tmp/deploy_${RUNTIME}.log | head -1 | awk '{print $3}')

echo ""
echo "================================================"
echo "DEPLOYMENT SUMMARY"
echo "================================================"
echo ""

if [ -n "$PROJECT_ID" ] && [ -n "$PROJECT_NAME" ]; then
    echo "✅ Project created successfully!"
    echo "   Name: $PROJECT_NAME"
    echo "   ID: $PROJECT_ID"
    echo ""
    echo "To check deployment status:"
    echo "1. Go to https://app.zerops.io"
    echo "2. Find project: $PROJECT_NAME"
    echo "3. Check if the $RUNTIME service is running"
    echo ""
    
    # Check if deployment was attempted
    if grep -q "Code deployed successfully" /tmp/deploy_${RUNTIME}.log; then
        echo "✅ Code deployment successful!"
    elif grep -q "websocket: bad handshake" /tmp/deploy_${RUNTIME}.log; then
        echo "⚠️  Deployment may have succeeded despite websocket error"
        echo "   Check the Zerops dashboard to verify"
    elif grep -q "Deployment failed" /tmp/deploy_${RUNTIME}.log; then
        echo "❌ Code deployment failed"
        echo "   Check the error in the log above"
    else
        echo "⚠️  Code deployment status unclear"
    fi
else
    echo "❌ Project creation failed"
    echo "   Check the log above for errors"
fi

echo ""
echo "Full log saved to: /tmp/deploy_${RUNTIME}.log"
echo ""