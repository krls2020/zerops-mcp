#!/bin/bash

# Test service_info tool for specific service ID

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# Service ID to check
SERVICE_ID="i9r270d4Rt2WNfbb1jpRGA"

echo "Testing service_info for service ID: $SERVICE_ID"
echo "================================="

# Create request
REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "service_info",
    "arguments": {
      "service_id": "$SERVICE_ID"
    }
  },
  "id": 1
}
EOF
)

# Send request to MCP server
echo "$REQUEST" | ./mcp-server 2>/dev/null | grep -v "Starting Zerops MCP server" | jq -r '.result.content[0].text' 2>/dev/null || echo "Failed to parse response"