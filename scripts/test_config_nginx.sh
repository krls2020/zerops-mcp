#!/bin/bash

# Test script for config_nginx tool

echo "Testing config_nginx tool..."
echo

# Build the server
echo "Building server..."
/usr/local/go/bin/go build -o mcp-server cmd/mcp-server/main.go
if [ $? -ne 0 ]; then
    echo "Build failed"
    exit 1
fi

# Set required environment variable
export ZEROPS_API_KEY="test-key"

# Function to test a framework
test_framework() {
    local framework=$1
    echo "Testing nginx config for: $framework"
    echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"config_nginx","arguments":{"framework":"'$framework'"}}}' | ./mcp-server 2>/dev/null | jq -r '.result.content[0].text' | head -20
    echo "---"
    echo
}

# Test different frameworks
test_framework "laravel"
test_framework "symfony"
test_framework "wordpress"
test_framework "custom"

echo "Test complete!"