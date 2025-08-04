#!/bin/bash

# Test process_status tool

echo "Checking process status..."

# Set API key
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# Create a simple Python script to handle JSON-RPC communication
cat > /tmp/test_process_status.py << 'EOF'
import json
import subprocess
import sys
import time

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

# Send tool call to check process
tool_request = {
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
        "name": "process_status",
        "arguments": {"process_id": "qqZSwhU1RD2leBvVB4b9Hg"}
    }
}
proc.stdin.write(json.dumps(tool_request) + "\n")
proc.stdin.flush()

# Read response
response = proc.stdout.readline()

# Parse and display result
try:
    result = json.loads(response)
    if "result" in result:
        print("\nProcess Status:")
        for item in result["result"]:
            if item["type"] == "text":
                print(item["text"])
except:
    print("Raw response:", response)

# Wait and check subdomain status again
time.sleep(5)
print("\n\nChecking subdomain status again...")

# Send tool call
tool_request = {
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
        "name": "subdomain_status",
        "arguments": {"service_id": "ZSr67vv6RvSjnrgM0hAaUw"}
    }
}
proc.stdin.write(json.dumps(tool_request) + "\n")
proc.stdin.flush()

# Read response
response = proc.stdout.readline()

# Parse and display result
try:
    result = json.loads(response)
    if "result" in result:
        print("\nSubdomain Status:")
        for item in result["result"]:
            if item["type"] == "text":
                print(item["text"])
except:
    print("Raw response:", response)

# Cleanup
proc.terminate()
EOF

python3 /tmp/test_process_status.py