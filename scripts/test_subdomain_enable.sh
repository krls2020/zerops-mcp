#!/bin/bash

# Test subdomain_enable tool

echo "Testing subdomain_enable tool..."
echo

# Set API key
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# Test the subdomain_enable tool
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "subdomain_enable", "arguments": {"service_id": "ZSr67vv6RvSjnrgM0hAaUw"}}}' | ./mcp-server

echo
echo "Done"