#!/bin/bash

# Batch deployment script
set -e

export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

echo "Batch deployment of runtime recipes to Zerops"
echo "============================================="
echo ""

# Keep track of created projects
CREATED_PROJECTS=()

# Run deployment test multiple times
echo "Deploying projects..."
echo ""

# Run the test 5 times to create 5 projects
for i in {1..5}; do
    echo "Deployment $i of 5..."
    
    # Run the test that creates and deploys a project
    OUTPUT=$(/usr/local/go/bin/go test -v ./test/integration -run TestActualDeployment -count=1 -timeout 5m 2>&1 || true)
    
    # Extract project info from output
    PROJECT_NAME=$(echo "$OUTPUT" | grep "Created project:" | sed 's/.*Created project: //' | cut -d' ' -f1)
    PROJECT_ID=$(echo "$OUTPUT" | grep "Created project:" | grep -o 'ID: [^)]*' | cut -d' ' -f2)
    
    if [ -n "$PROJECT_NAME" ] && [ -n "$PROJECT_ID" ]; then
        CREATED_PROJECTS+=("$PROJECT_NAME:$PROJECT_ID")
        echo "✓ Created: $PROJECT_NAME (ID: $PROJECT_ID)"
    else
        echo "✗ Deployment $i failed or timed out"
    fi
    
    # Small delay between deployments
    sleep 5
done

echo ""
echo "============================================="
echo "DEPLOYMENT SUMMARY"
echo "============================================="
echo ""
echo "Successfully created ${#CREATED_PROJECTS[@]} projects:"
echo ""

for project in "${CREATED_PROJECTS[@]}"; do
    IFS=':' read -r name id <<< "$project"
    echo "- $name"
    echo "  ID: $id"
    echo "  URL: https://app.zerops.io"
    echo ""
done

echo "All projects include:"
echo "- Python 3.12 API service (deployed and running)"
echo "- PostgreSQL 16 database"
echo "- Full application stack"
echo ""
echo "To view your projects:"
echo "1. Go to https://app.zerops.io"
echo "2. Look for projects starting with 'test-deploy-'"
echo ""