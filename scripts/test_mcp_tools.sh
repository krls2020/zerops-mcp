#!/bin/bash

# Test MCP server tools manually

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# Build the server first
echo "Building MCP server..."
/usr/local/go/bin/go build -o mcp-server cmd/mcp-server/main.go

echo "Testing MCP server tools..."

# Test auth_validate
echo -e "\n=== Testing auth_validate tool ==="
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"auth_validate","arguments":{}}}' | ./mcp-server 2>&1 | grep -A5 -B5 "result"

# Test platform_info
echo -e "\n=== Testing platform_info tool ==="
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"platform_info","arguments":{}}}' | ./mcp-server 2>&1 | grep -A5 -B5 "result"

# Test region_list
echo -e "\n=== Testing region_list tool ==="
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"region_list","arguments":{}}}' | ./mcp-server 2>&1 | grep -A5 -B5 "result"

# Test knowledge_search_patterns
echo -e "\n=== Testing knowledge_search_patterns tool ==="
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"knowledge_search_patterns","arguments":{"tags":"django"}}}' | ./mcp-server 2>&1 | grep -A20 -B5 "result"

# List all tools
echo -e "\n=== Listing all tools ==="
echo '{"jsonrpc":"2.0","id":5,"method":"tools/list","params":{}}' | ./mcp-server 2>&1 | grep -A20 -B5 "result"