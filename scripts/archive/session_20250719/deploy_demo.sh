#!/bin/bash

# Simple deployment demo script
set -e

cd "$(dirname "$0")"

# Set API key
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "Creating Python demo project in Zerops..."

# Create project
PROJECT_NAME="demo-python-$(date +%s)"
echo "Project name: $PROJECT_NAME"

# Call MCP server directly
PROJECT_RESULT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"project_create","arguments":{"name":"'"$PROJECT_NAME"'","region":"prg1"}}}' | ./mcp-server 2>/dev/null | grep -v "Starting")

echo "Create result:"
echo "$PROJECT_RESULT" | jq .

# Extract project ID
PROJECT_ID=$(echo "$PROJECT_RESULT" | jq -r '.result.project_id // empty')

if [ -z "$PROJECT_ID" ]; then
    echo "Failed to create project"
    exit 1
fi

echo ""
echo "Created project: $PROJECT_ID"

# Import services
echo ""
echo "Importing services..."

IMPORT_YAML='project:
  name: '"$PROJECT_NAME"'
services:
  - hostname: api
    type: python@3.12
    enableSubdomainAccess: true
    ports:
      - port: 8000
        httpSupport: true
  - hostname: db
    type: postgresql@16
    mode: NON_HA'

# Escape YAML for JSON
YAML_ESCAPED=$(echo "$IMPORT_YAML" | jq -Rs .)

IMPORT_RESULT=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"project_import","arguments":{"project_id":"'"$PROJECT_ID"'","yaml":'"$YAML_ESCAPED"'}}}' | ./mcp-server 2>/dev/null | grep -v "Starting")

echo "Import result:"
echo "$IMPORT_RESULT" | jq .

echo ""
echo "Waiting for services to start (30s)..."
sleep 30

# Get service list
SERVICE_LIST=$(echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"service_list","arguments":{"project_id":"'"$PROJECT_ID"'"}}}' | ./mcp-server 2>/dev/null | grep -v "Starting")

echo "Services:"
echo "$SERVICE_LIST" | jq '.result.services[] | {id: .id, hostname: .hostname, status: .status}'

# Get API service ID
SERVICE_ID=$(echo "$SERVICE_LIST" | jq -r '.result.services[] | select(.hostname == "api") | .id')

if [ -n "$SERVICE_ID" ]; then
    echo ""
    echo "Enabling subdomain for API service: $SERVICE_ID"
    
    SUBDOMAIN_RESULT=$(echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"subdomain_enable","arguments":{"service_id":"'"$SERVICE_ID"'"}}}' | ./mcp-server 2>/dev/null | grep -v "Starting")
    
    echo "Subdomain result:"
    echo "$SUBDOMAIN_RESULT" | jq .
    
    # Wait and check status
    echo "Waiting for subdomain (20s)..."
    sleep 20
    
    STATUS_RESULT=$(echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"subdomain_status","arguments":{"service_id":"'"$SERVICE_ID"'"}}}' | ./mcp-server 2>/dev/null | grep -v "Starting")
    
    echo ""
    echo "Subdomain status:"
    echo "$STATUS_RESULT" | jq .
fi

echo ""
echo "========================================="
echo "DEPLOYMENT COMPLETE!"
echo "========================================="
echo ""
echo "Project: $PROJECT_NAME"
echo "Project ID: $PROJECT_ID"
echo ""
echo "View in Zerops dashboard:"
echo "https://app.zerops.io"
echo ""
echo "Look for project: $PROJECT_NAME"
echo ""