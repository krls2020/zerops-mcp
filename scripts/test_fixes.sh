#!/bin/bash

# Test script for the fixes
echo "Testing Zerops MCP Server Fixes"
echo "================================"

# Set test API key
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# 1. Test pattern fixes
echo -e "\n1. Testing pattern fixes (should use 'hostname' not 'name')..."
/usr/local/go/bin/go test -v ./test/integration -run TestKnowledgeTools

# 2. Test service logs implementation
echo -e "\n2. Testing service logs (should handle errors gracefully)..."
/usr/local/go/bin/go test -v ./test/integration -run TestServiceTools/service_logs

# 3. Test error handling
echo -e "\n3. Testing error handling..."
/usr/local/go/bin/go test -v ./test/integration -run TestErrorHandling

echo -e "\nAll tests completed!"