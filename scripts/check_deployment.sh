#!/bin/bash

# Check deployment status for Python recipe

echo "=== Checking Zerops Deployment Status ==="
echo

# Get the latest test project
PROJECT_NAME=$(zcli project list 2>/dev/null | grep "test-deploy-" | head -1 | awk '{print $2}')

if [ -z "$PROJECT_NAME" ]; then
    echo "No test-deploy project found"
    exit 1
fi

echo "Found project: $PROJECT_NAME"
echo

# Get project details
echo "Project services:"
zcli service list --projectName "$PROJECT_NAME" 2>/dev/null || echo "Failed to list services"
echo

# Check VPN status
echo "VPN Status:"
zcli vpn status 2>/dev/null || echo "VPN not connected"
echo

echo "To connect to project VPN, run:"
echo "sudo zcli vpn up --projectName $PROJECT_NAME"
echo

echo "To view service logs:"
echo "zcli service log api --projectName $PROJECT_NAME"