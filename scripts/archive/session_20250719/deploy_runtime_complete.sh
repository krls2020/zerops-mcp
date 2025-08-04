#!/bin/bash

# Complete deployment script with subdomain activation
set -e

RUNTIME=${1:-python}
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"
export DEPLOY_RUNTIME="$RUNTIME"

echo "================================================"
echo "DEPLOYING $RUNTIME WITH SUBDOMAIN ACCESS"
echo "================================================"
echo ""

# Build MCP server
echo "Building MCP server..."
/usr/local/go/bin/go build -o mcp-server cmd/mcp-server/main.go

# Run deployment
echo "Starting deployment..."
/usr/local/go/bin/go test -v ./test/integration -run TestDeployRuntime -timeout 5m 2>&1 | tee /tmp/deploy_${RUNTIME}_full.log

# Extract project info
PROJECT_ID=$(grep "ID:" /tmp/deploy_${RUNTIME}_full.log | grep -v "Service:" | head -1 | awk '{print $3}' | tr -d ')')
PROJECT_NAME=$(grep "Name: demo-" /tmp/deploy_${RUNTIME}_full.log | head -1 | awk '{print $3}')

echo ""
echo "================================================"
echo "CHECKING DEPLOYMENT STATUS"
echo "================================================"
echo ""

if [ -n "$PROJECT_ID" ]; then
    echo "Project created: $PROJECT_NAME (ID: $PROJECT_ID)"
    
    # Get service name from runtime
    case $RUNTIME in
        "php") SERVICE_NAME="apacheapi" ;;
        "ruby") SERVICE_NAME="app" ;;
        *) SERVICE_NAME="api" ;;
    esac
    
    # Check if subdomain was enabled
    if grep -q "Subdomain URL:" /tmp/deploy_${RUNTIME}_full.log; then
        SUBDOMAIN_URL=$(grep "Subdomain URL:" /tmp/deploy_${RUNTIME}_full.log | awk '{print $4}')
        echo "✅ Subdomain enabled: $SUBDOMAIN_URL"
    else
        echo ""
        echo "Checking service status and enabling subdomain if needed..."
        
        # Use MCP server to check and enable subdomain
        ./mcp-server <<EOF 2>/dev/null | grep -A20 "result" || true
{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"service_list","arguments":{"project_id":"$PROJECT_ID"}},"id":2}
EOF
        
        # Get service ID
        SERVICE_ID=$(./mcp-server <<EOF 2>/dev/null | grep -B5 -A5 "\"name\":\"$SERVICE_NAME\"" | grep '"id"' | head -1 | cut -d'"' -f4 || echo "")
{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"service_list","arguments":{"project_id":"$PROJECT_ID"}},"id":2}
EOF
        
        if [ -n "$SERVICE_ID" ]; then
            echo "Found service $SERVICE_NAME (ID: $SERVICE_ID)"
            
            # Enable subdomain
            echo "Enabling subdomain access..."
            ./mcp-server <<EOF 2>/dev/null | grep -A10 "result" || true
{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"subdomain_enable","arguments":{"service_id":"$SERVICE_ID"}},"id":2}
EOF
            
            sleep 5
            
            # Check subdomain status
            echo ""
            echo "Checking subdomain status..."
            ./mcp-server <<EOF 2>/dev/null | grep -A20 "result" || true
{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"subdomain_status","arguments":{"service_id":"$SERVICE_ID"}},"id":2}
EOF
        fi
    fi
    
    echo ""
    echo "================================================"
    echo "DEPLOYMENT COMPLETE"
    echo "================================================"
    echo ""
    echo "Project: $PROJECT_NAME"
    echo "Dashboard: https://app.zerops.io"
    echo ""
    echo "To verify deployment:"
    echo "1. Check if service is running in dashboard"
    echo "2. Look for subdomain URL in service details"
    echo "3. Test the application via subdomain"
else
    echo "❌ Deployment failed - check logs above"
fi

echo ""
echo "Full log: /tmp/deploy_${RUNTIME}_full.log"
echo ""