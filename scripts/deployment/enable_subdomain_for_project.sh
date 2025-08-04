#!/bin/bash

# Enable subdomain for a project's service
PROJECT_ID=$1
SERVICE_NAME=${2:-api}

if [ -z "$PROJECT_ID" ]; then
    echo "Usage: $0 <project_id> [service_name]"
    exit 1
fi

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "Enabling subdomain for service '$SERVICE_NAME' in project $PROJECT_ID..."

# Initialize MCP server and get service list
RESPONSE=$(./mcp-server <<EOF 2>&1
{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"service_list","arguments":{"project_id":"$PROJECT_ID"}},"id":2}
EOF
)

# Extract service ID from response
SERVICE_ID=$(echo "$RESPONSE" | grep -B2 -A2 "\"$SERVICE_NAME\"" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$SERVICE_ID" ]; then
    echo "Service '$SERVICE_NAME' not found in project"
    exit 1
fi

echo "Found service: $SERVICE_NAME (ID: $SERVICE_ID)"

# Enable subdomain
echo "Enabling subdomain access..."
ENABLE_RESPONSE=$(./mcp-server <<EOF 2>&1
{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"subdomain_enable","arguments":{"service_id":"$SERVICE_ID"}},"id":2}
EOF
)

# Check if successful
if echo "$ENABLE_RESPONSE" | grep -q "Process started"; then
    echo "✅ Subdomain activation started"
    
    # Wait a bit
    sleep 5
    
    # Check status
    echo ""
    echo "Checking subdomain status..."
    ./mcp-server <<EOF 2>&1 | grep -A30 '"result"' | grep -E "(subdomain_url|enabled|status)"
{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"subdomain_status","arguments":{"service_id":"$SERVICE_ID"}},"id":2}
EOF
else
    echo "❌ Failed to enable subdomain"
    echo "$ENABLE_RESPONSE" | grep -A5 "error"
fi