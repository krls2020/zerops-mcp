#!/bin/bash

# Quick deployment script for testing
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Check API key
if [ -z "$ZEROPS_API_KEY" ]; then
    echo "Setting API key..."
    export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"
fi

echo "Testing MCP server connection..."
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"auth_validate","arguments":{}}}' | ./mcp-server

echo ""
echo "Creating Python demo project..."
PROJECT_NAME="demo-python-$(date +%Y%m%d%H%M)"

# Create project
echo "Creating project: $PROJECT_NAME"
CREATE_RESULT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"project_create\",\"arguments\":{\"name\":\"$PROJECT_NAME\",\"region\":\"prg1\"}}}" | ./mcp-server)
echo "Result: $CREATE_RESULT"

# Extract project ID (simple grep approach)
PROJECT_ID=$(echo "$CREATE_RESULT" | grep -o '"project_id":"[^"]*' | cut -d'"' -f4)
echo "Project ID: $PROJECT_ID"

if [ -z "$PROJECT_ID" ]; then
    echo "Failed to create project"
    exit 1
fi

# Import services
echo ""
echo "Importing services..."
IMPORT_YAML='project:
  name: '"$PROJECT_NAME"'
services:
  - hostname: api
    type: python@3.12
    enableSubdomainAccess: true
    envVariables:
      PORT: 8000
  - hostname: db
    type: postgresql@16
    mode: NON_HA'

# Create import request
IMPORT_REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "project_import",
    "arguments": {
      "project_id": "$PROJECT_ID",
      "yaml": "$(echo "$IMPORT_YAML" | sed 's/"/\\"/g' | sed ':a;N;$!ba;s/\n/\\n/g')"
    }
  }
}
EOF
)

echo "$IMPORT_REQUEST" | ./mcp-server

echo ""
echo "Waiting for services to start (30s)..."
sleep 30

# Get service list
echo ""
echo "Getting service list..."
SERVICE_LIST=$(echo "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"service_list\",\"arguments\":{\"project_id\":\"$PROJECT_ID\"}}}" | ./mcp-server)
echo "$SERVICE_LIST"

# Extract API service ID
SERVICE_ID=$(echo "$SERVICE_LIST" | grep -B2 '"hostname":"api"' | grep '"id"' | head -1 | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo ""
echo "API Service ID: $SERVICE_ID"

if [ -n "$SERVICE_ID" ]; then
    # Enable subdomain
    echo ""
    echo "Enabling subdomain access..."
    SUBDOMAIN_RESULT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"subdomain_enable\",\"arguments\":{\"service_id\":\"$SERVICE_ID\"}}}" | ./mcp-server)
    echo "$SUBDOMAIN_RESULT"
    
    # Wait for process
    echo "Waiting for subdomain to be enabled (15s)..."
    sleep 15
    
    # Check subdomain status
    echo ""
    echo "Checking subdomain status..."
    STATUS_RESULT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":5,\"method\":\"tools/call\",\"params\":{\"name\":\"subdomain_status\",\"arguments\":{\"service_id\":\"$SERVICE_ID\"}}}" | ./mcp-server)
    echo "$STATUS_RESULT"
fi

echo ""
echo "========================================="
echo "Deployment Summary"
echo "========================================="
echo "Project Name: $PROJECT_NAME"
echo "Project ID: $PROJECT_ID"
echo "Service ID: $SERVICE_ID"
echo ""
echo "To view in Zerops:"
echo "1. Go to https://app.zerops.io"
echo "2. Look for project: $PROJECT_NAME"
echo ""
echo "To deploy code (requires VPN):"
echo "sudo zcli vpn up --projectId $PROJECT_ID"
echo "cd test/fixtures/recipe-python"
echo "zcli push"