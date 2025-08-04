#!/bin/bash

# Deploy a single recipe to Zerops
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

TIMESTAMP=$(date +%Y%m%d%H%M)
RUNTIME=${1:-python}

echo "Deploying $RUNTIME recipe to Zerops..."

# Set API key
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# Step 1: Create project
PROJECT_NAME="demo-$RUNTIME-$TIMESTAMP"
echo "Creating project: $PROJECT_NAME"

./scripts/test_mcp_tools.sh project_create <<EOF
{
  "name": "$PROJECT_NAME",
  "region": "prg1"
}
EOF

# Read the output and extract project ID
echo ""
echo "Enter the project ID from above: "
read PROJECT_ID

if [ -z "$PROJECT_ID" ]; then
    echo "No project ID provided"
    exit 1
fi

# Step 2: Import services
echo ""
echo "Importing services for $RUNTIME..."

case $RUNTIME in
    python)
        IMPORT_FILE="test/fixtures/recipe-python/zerops-project-import.yml"
        SERVICE_NAME="api"
        ;;
    php)
        IMPORT_FILE="test/fixtures/recipe-php-main/zerops-project-import.yml"
        SERVICE_NAME="apacheapi"
        ;;
    nodejs)
        IMPORT_FILE="test/fixtures/recipe-nodejs-main/zerops-project-import.yml"
        SERVICE_NAME="api"
        ;;
    go)
        IMPORT_FILE="test/fixtures/recipe-go-main/zerops-project-import.yml"
        SERVICE_NAME="api"
        ;;
    *)
        echo "Unknown runtime: $RUNTIME"
        exit 1
        ;;
esac

# Read YAML and update project name
YAML_CONTENT=$(cat "$IMPORT_FILE" | sed "s/name: recipe-[^ ]*/name: $PROJECT_NAME/")

echo "Import YAML:"
echo "$YAML_CONTENT"
echo ""

# Create temp file with YAML
TEMP_YAML="/tmp/import_$TIMESTAMP.yml"
echo "$YAML_CONTENT" > "$TEMP_YAML"

# Import using file
./scripts/test_mcp_tools.sh project_import <<EOF
{
  "project_id": "$PROJECT_ID",
  "yaml": "$(cat "$TEMP_YAML" | sed 's/"/\\"/g' | tr '\n' ' ')"
}
EOF

rm -f "$TEMP_YAML"

echo ""
echo "Waiting for services to start (30s)..."
sleep 30

# Step 3: Get service list
echo ""
echo "Getting service list..."
./scripts/test_mcp_tools.sh service_list <<EOF
{
  "project_id": "$PROJECT_ID"
}
EOF

echo ""
echo "Enter the service ID for $SERVICE_NAME from above: "
read SERVICE_ID

if [ -z "$SERVICE_ID" ]; then
    echo "No service ID provided"
    exit 1
fi

# Step 4: Enable subdomain
echo ""
echo "Enabling subdomain access..."
./scripts/test_mcp_tools.sh subdomain_enable <<EOF
{
  "service_id": "$SERVICE_ID"
}
EOF

echo ""
echo "Waiting for subdomain to be enabled (20s)..."
sleep 20

# Step 5: Check subdomain status
echo ""
echo "Checking subdomain status..."
./scripts/test_mcp_tools.sh subdomain_status <<EOF
{
  "service_id": "$SERVICE_ID"
}
EOF

echo ""
echo "========================================="
echo "Deployment Summary"
echo "========================================="
echo "Project: $PROJECT_NAME (ID: $PROJECT_ID)"
echo "Service: $SERVICE_NAME (ID: $SERVICE_ID)"
echo ""
echo "To view in Zerops dashboard:"
echo "https://app.zerops.io"
echo ""
echo "To deploy code (requires VPN):"
echo "sudo zcli vpn up --projectId $PROJECT_ID"
echo "cd test/fixtures/recipe-$RUNTIME*"
echo "zcli push"
echo ""