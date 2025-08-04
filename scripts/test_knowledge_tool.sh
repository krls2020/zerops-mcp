#!/bin/bash

# Test knowledge_get_runtime tool

echo "Testing knowledge_get_runtime tool..."
echo

# Test with nodejs runtime
cat <<EOF | ./mcp-server
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "knowledge_get_runtime",
    "arguments": {
      "runtime": "nodejs"
    }
  },
  "id": 1
}
EOF

echo
echo "---"
echo

# Test with non-existent runtime
cat <<EOF | ./mcp-server
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "knowledge_get_runtime",
    "arguments": {
      "runtime": "ruby"
    }
  },
  "id": 2
}
EOF