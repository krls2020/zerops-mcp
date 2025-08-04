#!/bin/bash

# Test script for Zerops MCP Knowledge Tools

echo "=== Testing Zerops MCP Knowledge Tools ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to test a tool
test_tool() {
    local tool_name=$1
    local params=$2
    local expected_content=$3
    
    echo -n "Testing $tool_name... "
    
    # Create request
    local request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "$tool_name",
    "arguments": $params
  },
  "id": 1
}
EOF
)
    
    # Run the tool
    result=$(echo "$request" | ./mcp-server 2>/dev/null | grep -A 100 '"result"' | head -20)
    
    if echo "$result" | grep -q "$expected_content"; then
        echo -e "${GREEN}✓ PASSED${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        echo "Expected to find: $expected_content"
        echo "Got: $result"
        return 1
    fi
}

# Build the server
echo "Building MCP server..."
/usr/local/go/bin/go build -o mcp-server cmd/mcp-server/main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful${NC}"
echo

# Test 1: Get Runtime
echo "1. Testing knowledge_get_runtime"
test_tool "knowledge_get_runtime" '{"runtime": "nodejs"}' "nodejs"

# Test 2: Search Patterns
echo "2. Testing knowledge_search_patterns"
test_tool "knowledge_search_patterns" '{"tags": ["php"]}' "laravel"

# Test 3: Validate Config
echo "3. Testing knowledge_validate_config"
test_tool "knowledge_validate_config" '{"config": {"services": [{"hostname": "app", "type": "nodejs@20"}]}}' "valid"

# Test 4: Get Docs
echo "4. Testing knowledge_get_docs"
test_tool "knowledge_get_docs" '{}' "Zerops Platform Knowledge Base"

# Test 5: Get Docs with Section
echo "5. Testing knowledge_get_docs with section"
test_tool "knowledge_get_docs" '{"section": "deployment"}' "Deployment"

echo
echo "=== Test Summary ==="
echo "All tests completed. Check results above."