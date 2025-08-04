#!/bin/bash

# Test service_info tool for specific service

echo "Getting service info for service ID: i9r270d4Rt2WNfbb1jpRGA..."

# Set API key
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# Create a simple Python script to handle JSON-RPC communication
cat > /tmp/test_specific_service.py << 'EOF'
import json
import subprocess
import sys

# Initialize the MCP server
proc = subprocess.Popen(
    ["./mcp-server"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True
)

# Send initialize request
init_request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
        "protocolVersion": "0.1.0",
        "capabilities": {},
        "clientInfo": {"name": "test", "version": "1.0"}
    }
}
proc.stdin.write(json.dumps(init_request) + "\n")
proc.stdin.flush()

# Read response
response = proc.stdout.readline()
print("Initialize response received")

# Send tool call
tool_request = {
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
        "name": "service_info",
        "arguments": {"service_id": "i9r270d4Rt2WNfbb1jpRGA"}
    }
}
proc.stdin.write(json.dumps(tool_request) + "\n")
proc.stdin.flush()

# Read response
response = proc.stdout.readline()

# Parse and display result
try:
    result = json.loads(response)
    if "result" in result and "content" in result["result"]:
        print("\nService Info:")
        for item in result["result"]["content"]:
            if item["type"] == "text":
                print(item["text"])
    else:
        print("Full response:", json.dumps(result, indent=2))
except Exception as e:
    print("Error parsing response:", e)
    print("Raw response:", response)

# Cleanup
proc.terminate()
EOF

python3 /tmp/test_specific_service.py