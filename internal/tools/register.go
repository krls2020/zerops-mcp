package tools

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
	"github.com/zeropsio/zerops-mcp-v3/internal/config"
	"github.com/zeropsio/zerops-mcp-v3/internal/zcli"
)

// RegisterAll registers all tools with the MCP server
func RegisterAll(s *server.MCPServer, cfg *config.Config) {
	// Create API client
	apiClient := api.NewClient(api.ClientOptions{
		BaseURL: cfg.ZeropsAPIURL,
		APIKey:  cfg.ZeropsAPIKey,
		Timeout: cfg.APITimeout,
		Debug:   cfg.Debug,
	})

	// Create zcli wrapper
	zcliWrapper := zcli.NewWithConfig(cfg.Debug, cfg.VPNWaitTime)

	// Register all tool categories
	RegisterAuthTools(s, apiClient)
	RegisterProjectTools(s, apiClient)
	RegisterServiceTools(s, apiClient)
	RegisterDeployTools(s, apiClient, zcliWrapper)
	RegisterConfigTools(s, apiClient)
	RegisterWorkflowTools(s, apiClient, zcliWrapper)
	RegisterSubdomainTools(s, apiClient)
	RegisterProcessTools(s, apiClient)
	
	// Register knowledge tools
	RegisterKnowledgeTools(s)
}

// GetServerInstructions returns instructions for LLMs
func GetServerInstructions() string {
	return `# Zerops MCP Server v3

The Zerops MCP server provides tools for managing projects, services, and deployments on the Zerops platform.

## 🎯 PRIORITY INSTRUCTION FOR ALL OPERATIONS 🎯

**When a user asks to deploy to Zerops:**

1. **ALWAYS START WITH KNOWLEDGE TOOLS** - Never create configurations manually
2. **USE PATTERNS FIRST** - Search for existing patterns before creating custom configs
3. **USE THE PATTERN'S SERVICES SECTION** - It contains the complete, tested configuration
4. **NEVER CREATE SERVICE YAML MANUALLY** - Always extract from the pattern
5. **VALIDATE EVERYTHING** - Use knowledge_validate_config before any import/deployment

**CRITICAL**: When you find a pattern with knowledge_search_patterns, it contains:
- A 'services' field with complete service configuration - USE THIS for project_import!
- A 'zeropsYml' field with deployment configuration - SAVE THIS as zerops.yml!

This prevents configuration errors and ensures compatibility with Zerops platform requirements.

## 🚨 CRITICAL: Configuration Creation Workflow 🚨

**NEVER attempt to create Zerops configurations manually!** The platform has specific requirements that must be followed exactly. Always use the knowledge tools to:

1. **ALWAYS check the actual project configuration first:**
   - Check .env, config files, or package.json to see what services are actually used
   - Look for database type (MySQL vs PostgreSQL), cache type, etc.
   - Don't blindly apply templates - verify they match the project needs

2. **First, identify what the user needs:**
   - Use 'knowledge_search_patterns' to find pre-built framework patterns
   - Use 'knowledge_get_runtime' for runtime-specific configurations
   - Use 'knowledge_get_service' for service details

3. **For service creation (import YAML):**
   - **Step 1**: Analyze project to identify actual services needed (check .env, config files)
   - **Step 2**: Use 'knowledge_search_patterns' to find the appropriate pattern
   - **Step 3**: **EXTRACT the 'services' field from the pattern** - it has the complete service configuration!
   - **Step 4**: If needed, adapt the services (e.g., replace postgresql with mariadb)
   - **Step 5**: Use 'project_import' with the pattern's services YAML
   
   **CRITICAL**: The pattern already contains a complete 'services' section with all the correct configurations. DO NOT create service configurations manually!

4. **For deployment (zerops.yml):**
   - **Step 1**: Use pattern's zerops.yml as base (from knowledge_search_patterns)
   - **Step 2**: Get runtime details with 'knowledge_get_runtime'
   - **Step 3**: Always include proper 'addToRunPrepare' for runtime files
   - **Step 4**: Validate with 'knowledge_validate_config'
   - **Step 5**: NEVER create zerops.yml from scratch - always base on knowledge

## Example Workflow for Creating a Laravel App

❌ **WRONG**: 
- Using PostgreSQL template when app uses MySQL
- Creating YAML configuration manually
- Writing zerops.yml from memory

✅ **CORRECT**:
1. Check .env file to see actual database (MySQL vs PostgreSQL)
2. 'knowledge_search_patterns tags=["laravel", "php"]'
3. If app uses MySQL: 'knowledge_get_service service_type="mariadb"'
4. Modify pattern to use MariaDB instead of PostgreSQL
5. 'project_import' with the modified YAML
6. Use pattern's zerops.yml for deployment

## Available Tools (41 total)
- **Authentication** (3): auth_validate, platform_info, region_list
- **Projects** (5): project_create, project_list, project_info, project_import, project_delete
- **Services** (6): service_list, service_info, service_logs, service_start, service_stop, service_delete
- **Deployment** (8): vpn_status, vpn_connect, vpn_disconnect, deploy_validate, deploy_push, deploy_status, deploy_logs, deploy_troubleshoot
- **Configuration** (5): config_templates, config_generate, config_validate, env_vars_show, config_nginx
- **Workflows** (3): workflow_create_app, workflow_clone, workflow_diagnose
- **Subdomain** (3): subdomain_enable, subdomain_disable, subdomain_status
- **Process** (1): process_status
- **Knowledge** (7): knowledge_get_runtime, knowledge_search_patterns, knowledge_validate_config, knowledge_resolve_dependencies, knowledge_get_service, knowledge_list_services, knowledge_get_docs

## Key Concepts
- **Projects**: Isolated environments containing multiple services
- **Services**: Individual applications, databases, or utilities
- **Deployment**: Code is deployed via zcli push (requires VPN)
- **Configuration**: Services configured via YAML import

## Important Rules
1. **Project Naming**: No restrictions - can use any characters
   - ✅ Good: my-project, MyApp-2024, Test_Project!
   - Projects can have spaces, special characters, etc.

2. **Service Naming**: Only lowercase letters and numbers (no hyphens, no uppercase)
   - ✅ Good: app, db01, cache1
   - ❌ Bad: my-app, web-server, MyApp

2. **Service Creation**: Services are created via YAML import, not direct API
   - Use project_import tool with proper YAML structure
   - Supports preprocessing functions for secrets: <@generateRandomString(<32>)>

3. **Environment Variables**: 
   - Cross-service references: ${servicename_variablename}
   - Database passwords are auto-generated
   - Can only be set during service creation
   - Use envSecrets for sensitive data with preprocessing: <@generateRandomString(<32>)>

4. **Database Services**:
   - **IMPORTANT**: Check what database the project actually uses!
     - Laravel: Check .env for DB_CONNECTION (mysql, pgsql, etc.)
     - Node.js: Check package.json dependencies or config files
     - Django: Check settings.py for DATABASE settings
   - PostgreSQL recommended for most apps (use if DB_CONNECTION=pgsql)
   - MariaDB only version 10.6 (NOT 11) - use if DB_CONNECTION=mysql
   - Use Valkey or KeyDB instead of Redis
   - Database name env var is 'dbName' (e.g., ${db_dbName})

5. **Deployment Requirements**:
   - VPN must be connected (use vpn_connect)
   - Code must be in git repository with at least one commit
   - Requires zerops.yml configuration file in project root
   - **SERVICE NAME IS REQUIRED**: Must specify service_name parameter matching hostname in zerops.yml
   - Files are deployed to /var/www in runtime containers
   - Use addToRunPrepare for files needed during prepareCommands
   - Git repository needs commits: git add . && git commit -m "Initial commit"
   - **CRITICAL**: Without commits, deploy will fail with "exit status 128"
   - **CRITICAL**: Without service_name, deploy will fail with "Please, select a service"

6. **Subdomain Access**:
   - Only for services with HTTP/HTTPS ports
   - Requires ports with httpSupport: true
   - Service must be in ACTIVE/RUNNING state
   - Use subdomain_status to check eligibility
   - Database services cannot have subdomain access
   - enableSubdomainAccess in YAML only works with buildFromGit
   - For zcli push deployments, use subdomain_enable after deployment
   - Subdomain URLs may take time to appear after enabling

7. **Special Notes**:
   - buildFromGit is only for Zerops recipes, not user code
   - PHP: Use php-apache@8.3 for Laravel
   - Always check error messages for resolution steps
   - **PHP Nginx Configuration**: For Laravel/Symfony/WordPress using php-nginx:
     - Use 'config_nginx' tool to get the proper nginx template
     - Save as site.conf.tmpl in project root
     - Add siteConfigPath: site.conf.tmpl to run section
     - Include site.conf.tmpl in deployFiles

## 🚨 CRITICAL: Pattern-First Deployment

**ALWAYS search for patterns FIRST!** Never create configurations manually.

### Why Patterns Are Essential:
Patterns include critical configurations that prevent deployment failures:
- ✅ prepareCommands for runtime dependency installation
- ✅ addToRunPrepare for files needed during preparation  
- ✅ Correct ports with httpSupport: true
- ✅ Environment variable mappings between services
- ✅ Framework-specific optimizations

Manual configurations often miss these, causing:
- ❌ ModuleNotFoundError (missing prepareCommands)
- ❌ Connection failures (wrong port config)
- ❌ Build failures (missing addToRunPrepare)

## 📋 Pre-Deployment Project Analysis

**BEFORE ANY DEPLOYMENT, ALWAYS ANALYZE THE PROJECT:**

1. **Check Database Configuration**:
   - Laravel: Read .env for DB_CONNECTION value
   - Django: Check settings.py for DATABASES config
   - Node.js: Look for database packages in package.json
   - Flask: Check config.py or .env for database URL

2. **Identify Required Services**:
   - Database type (PostgreSQL, MySQL/MariaDB, MongoDB)
   - Cache needs (Redis alternative: Valkey/KeyDB)
   - Message queue requirements
   - Search engine needs

3. **Framework Version**:
   - Check runtime version requirements
   - Verify framework-specific needs

## Standard Deployment Workflow

### 1. Project Analysis FIRST:
Analyze the project configuration files to understand actual requirements

### 2. Pattern Search Based on Analysis:
Use: knowledge_search_patterns tags=["framework-name"]
Examples: flask, django, express, laravel, nextjs, react, vue

### 3. Adapt Pattern to Project Needs:
When pattern doesn't match project:
1. Get service details: knowledge_get_service service_type="mariadb"
2. Replace pattern services with actual requirements
3. Keep pattern's deployment configuration (zeropsYml)

### 4. Complete Deployment Flow:
1. Analyze project: Check .env, config files, package.json
2. Search patterns: knowledge_search_patterns tags=["framework"]
3. **USE THE PATTERN'S SERVICES**: The pattern contains a complete 'services' field - use it!
4. Adapt if needed: If DB differs, modify the pattern's services (don't create from scratch)
5. Create project: project_create name="my-app" region="prg1"
6. Import services: project_import yaml="<pattern's services section>"
7. **WAIT for services**: Services take 10-30 seconds to initialize. Check with service_list
8. **CREATE zerops.yml**: Extract pattern's zeropsYml field and save as zerops.yml
9. Connect VPN: vpn_connect project_id="..."
10. Deploy: deploy_push project_id="..." service_name="app" working_dir="./"
    **CRITICAL**: service_name parameter is REQUIRED! Must match hostname from zerops.yml
11. Enable access: subdomain_enable (if needed)

### ⚠️ CRITICAL: Environment Variables in Zerops ⚠️

**NEVER create .env.production or similar files!** Zerops handles environment variables differently:

1. **Service Creation**: Environment variables are set during service import via YAML
2. **Runtime Injection**: Zerops automatically injects these into your application
3. **Cross-Service References**: Use ${servicename_variable} syntax in YAML
4. **No .env Files**: Do NOT create .env.production, .env.staging, etc.

**Example**: When importing services, environment variables are set like this:

    services:
      - hostname: app
        type: php-apache@8.3
        envVariables:
          APP_ENV: production
          DB_HOST: \${db_hostname}
          DB_PASSWORD: \${db_password}

### 📝 Required File: zerops.yml

**ALWAYS create zerops.yml from the pattern!** This file tells Zerops how to build and run your app:

1. Get pattern: knowledge_search_patterns
2. Look for the 'zeropsYml' field in the pattern response
3. Convert the zeropsYml content to YAML format
4. Save as 'zerops.yml' in project root
5. Modify only the 'setup' field to match your service hostname (e.g., 'app')

**Example extraction:**
Pattern returns: zeropsYml: { zerops: [{ setup: "api", build: {...}, run: {...} }] }
Save as zerops.yml with setup changed to match your service name

**What goes in zerops.yml:**
- Build configuration (base image, build commands)
- Deploy files specification
- Runtime configuration (start command, ports)
- **addToRunPrepare**: Files needed during prepareCommands (requirements.txt, package.json, etc.)
- NO environment variables (those are in service YAML)

**Common addToRunPrepare patterns:**
- Python: addToRunPrepare: ["requirements.txt"]
- Node.js: addToRunPrepare: ["package.json", "package-lock.json"]
- PHP: addToRunPrepare: ["composer.json", "composer.lock"]

### 5. When No Pattern Exists:
Only if no pattern is available:
1. Use knowledge_get_runtime for base configuration
2. Add framework-specific commands manually
3. Ensure you include ALL critical sections

## Quick Reference

- Find patterns: knowledge_search_patterns tags=["language", "framework"]
- Validate config: knowledge_validate_config
- List services: knowledge_list_services
- Get docs: knowledge_get_docs

## 📚 Example: Deploying a Laravel Application

Here's the CORRECT workflow for deploying a Laravel app:

1. Check .env file - sees DB_CONNECTION=mysql
2. knowledge_search_patterns tags=["laravel", "php"]
3. **EXTRACT services section from pattern** - it contains complete configuration!
4. Pattern has PostgreSQL but app needs MariaDB - replace in the services YAML
5. For Laravel: Copy APP_KEY from .env 
6. project_create name="myapp" region="prg1"
7. project_import with the PATTERN'S services YAML (modified for MariaDB)
8. Extract zerops.yml from pattern's zeropsYml field
9. Save as zerops.yml in project root
10. vpn_connect project_id="..."
11. deploy_push work_dir="./"

**NEVER manually create service configurations - always use the pattern's services section!**

**What the pattern contains:**
- services: Complete service configuration array
- zeropsYml: Complete deployment configuration
- The services section has ALL environment variables, ports, etc.
- Just extract and use it - don't recreate!

**What NOT to do:**
- ❌ Create .env.production
- ❌ Create service configurations manually
- ❌ Skip creating zerops.yml
- ❌ Put environment variables in zerops.yml
- ❌ Write your own service YAML - use the pattern!

## Common Issues
- "Service name invalid": Use only lowercase letters and numbers
- "VPN connection failed": Ensure zcli is installed and sudo access
- "Project not found": Verify project ID with project_list
- "Deploy failed": Check zerops.yml exists and VPN is connected
- "Service stack is not http": Add ports with httpSupport: true to enable subdomain
- "Could not open requirements file": Use addToRunPrepare in zerops.yml
- "Subdomain not working": Service must be deployed and in ACTIVE state
- "exit status 128": Git needs at least one commit (git add . && git commit -m "msg")
- "websocket: bad handshake": Log streaming error, deployment may still succeed
- "No action allowed, project will be deleted": Wait for services to initialize after import

Always provide actionable error messages with resolution steps and suggest the next tool to use.`
}