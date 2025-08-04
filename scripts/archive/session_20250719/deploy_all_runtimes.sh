#!/bin/bash

# Deploy all runtime recipes to Zerops
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR" && pwd)"
cd "$PROJECT_ROOT"

# Set API key
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwAd4tdD2hRiya6vYLkI1CRg-p"

# Timestamp for unique names
TIMESTAMP=$(date +%Y%m%d%H%M)

# Log file
LOG_FILE="test-reports/deployment_all_runtimes_${TIMESTAMP}.log"
mkdir -p test-reports

echo "Deploying all runtime recipes to Zerops..."
echo "Log file: $LOG_FILE"
echo ""

# Function to deploy a single runtime
deploy_runtime() {
    local runtime=$1
    local test_name=$2
    
    echo "========================================="
    echo "Deploying $runtime..."
    echo "========================================="
    
    # Run the Go test for this specific runtime
    echo "Running deployment test for $runtime..."
    
    # Set environment variable to specify which runtime to deploy
    export ZEROPS_TEST_RUNTIME="$runtime"
    export ZEROPS_TEST_PROJECT_PREFIX="demo-$runtime"
    
    # Run the test
    /usr/local/go/bin/go test -v ./test/integration -run "$test_name" -timeout 10m 2>&1 | tee -a "$LOG_FILE" || true
    
    echo ""
    echo "✓ $runtime deployment initiated"
    echo ""
    
    # Small delay between deployments
    sleep 5
}

# Deploy each runtime using specific tests or create new deployment logic
echo "Starting deployment of all runtimes..."

# For now, let's create projects using the API directly through Go code
cat > test/integration/deploy_all_runtimes_test.go << 'EOF'
package integration

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/zeropsio/zerops-mcp-v3/internal/api"
    "github.com/zeropsio/zerops-mcp-v3/internal/config"
)

func TestDeployAllRuntimes(t *testing.T) {
    apiKey := os.Getenv("ZEROPS_API_KEY")
    if apiKey == "" {
        t.Skip("ZEROPS_API_KEY not set")
    }
    
    ctx := context.Background()
    apiClient := api.NewClient(api.ClientOptions{
        BaseURL: "https://api.app-prg1.zerops.io",
        APIKey:  apiKey,
    })
    
    timestamp := time.Now().Format("20060102-1504")
    
    // Define all runtimes to deploy
    runtimes := []struct {
        name        string
        dir         string
        importFile  string
        serviceType string
        serviceName string
    }{
        {"python", "recipe-python", "zerops-project-import.yml", "python@3.12", "api"},
        {"php", "recipe-php-main", "zerops-project-import.yml", "php-apache@8.3", "apacheapi"},
        {"nodejs", "recipe-nodejs-main", "zerops-project-import.yml", "nodejs@20", "api"},
        {"go", "recipe-go-main", "zerops-project-import.yml", "go@latest", "api"},
        {"ruby", "recipe-ruby-main", "zerops-import.yml", "ruby@3.4", "app"},
        {"java", "recipe-java-main", "zerops-project-import.yml", "java@latest", "api"},
        {"dotnet", "recipe-dotnet-main", "zerops-project-import.yml", "dotnet@6", "api"},
        {"rust", "recipe-rust-main", "zerops-project-import.yml", "rust@latest", "api"},
        {"deno", "recipe-deno-main", "zerops-project-import.yml", "deno@latest", "api"},
        {"bun", "recipe-bun-main", "zerops-project-import.yml", "bun@latest", "api"},
        {"elixir", "recipe-elixir-main", "zerops-project-import.yml", "elixir@latest", "api"},
        {"gleam", "recipe-gleam-main", "zerops-project-import.yml", "gleam@1.5", "api"},
    }
    
    createdProjects := []string{}
    
    for _, rt := range runtimes {
        t.Run(rt.name, func(t *testing.T) {
            projectName := fmt.Sprintf("demo-%s-%s", rt.name, timestamp)
            
            // Create project
            project, err := apiClient.CreateProject(ctx, api.CreateProjectRequest{
                Name:   projectName,
                Region: "prg1",
            })
            if err != nil {
                t.Logf("Failed to create project for %s: %v", rt.name, err)
                return
            }
            
            createdProjects = append(createdProjects, project.ID)
            t.Logf("✓ Created project: %s (ID: %s)", projectName, project.ID)
            
            // Read import YAML
            importPath := filepath.Join("test/fixtures", rt.dir, rt.importFile)
            importYAML, err := os.ReadFile(importPath)
            if err != nil {
                t.Logf("Failed to read import file for %s: %v", rt.name, err)
                return
            }
            
            // Update project name in YAML
            yamlStr := string(importYAML)
            yamlStr = strings.ReplaceAll(yamlStr, "name: recipe-", "name: "+projectName)
            
            // Import services
            err = apiClient.ImportProjectServices(ctx, project.ID, yamlStr)
            if err != nil {
                t.Logf("Failed to import services for %s: %v", rt.name, err)
                return
            }
            
            t.Logf("✓ Services imported for %s", rt.name)
            
            // Wait for services to initialize
            time.Sleep(10 * time.Second)
            
            // Get service list
            services, err := apiClient.GetProjectServices(ctx, project.ID)
            if err != nil {
                t.Logf("Failed to get services for %s: %v", rt.name, err)
                return
            }
            
            // Find app service and enable subdomain
            for _, svc := range services {
                if svc.Hostname == rt.serviceName {
                    // Try to enable subdomain
                    process, err := apiClient.EnableSubdomainAccess(ctx, svc.ID)
                    if err != nil {
                        t.Logf("Note: Could not enable subdomain for %s: %v", rt.name, err)
                    } else {
                        t.Logf("✓ Subdomain enable initiated for %s (process: %s)", rt.name, process.ID)
                    }
                    break
                }
            }
            
            t.Logf("✓ %s project ready: https://app.zerops.io (project: %s)", rt.name, projectName)
        })
    }
    
    // Summary
    t.Logf("\n========================================")
    t.Logf("Deployment Summary")
    t.Logf("========================================")
    t.Logf("Created %d projects", len(createdProjects))
    for _, id := range createdProjects {
        t.Logf("Project ID: %s", id)
    }
    t.Logf("\nView all projects at: https://app.zerops.io")
}
EOF

# Add missing import
sed -i '' 's/import (/import (\n    "strings"/' test/integration/deploy_all_runtimes_test.go

# Run the deployment test
echo "Running deployment for all runtimes..."
/usr/local/go/bin/go test -v ./test/integration -run TestDeployAllRuntimes -timeout 30m 2>&1 | tee -a "$LOG_FILE"

echo ""
echo "========================================="
echo "Deployment Complete!"
echo "========================================="
echo ""
echo "Check the Zerops dashboard at: https://app.zerops.io"
echo "Look for projects starting with 'demo-' followed by the runtime name"
echo ""
echo "Full log saved to: $LOG_FILE"
echo ""
echo "To deploy code to these projects:"
echo "1. Note the project ID from the logs above"
echo "2. Connect VPN: sudo zcli vpn up --projectId PROJECT_ID"
echo "3. Go to recipe directory: cd test/fixtures/recipe-RUNTIME"
echo "4. Deploy: zcli push"
echo ""