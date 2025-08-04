package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
	"github.com/zeropsio/zerops-mcp-v3/internal/zcli"
)

// RegisterWorkflowTools registers all workflow tools
func RegisterWorkflowTools(s *server.MCPServer, client *api.Client, zcliWrapper *zcli.ZCLIWrapper) {
	// Register workflow_create_app tool
	workflowCreateAppTool := mcp.NewTool(
		"workflow_create_app",
		mcp.WithDescription("Complete workflow to create a new application project with services"),
		mcp.WithString("project_name",
			mcp.Required(),
			mcp.Description("Name for the new project"),
		),
		mcp.WithString("app_type",
			mcp.Required(),
			mcp.Description("Application type (nodejs, php-laravel, python-django, static-react, go, dotnet)"),
		),
		mcp.WithString("app_hostname",
			mcp.Required(),
			mcp.Description("Hostname for the application service (lowercase letters and numbers only)"),
		),
		mcp.WithArray("additional_services",
			mcp.Description("Additional services to create (e.g., ['postgresql:db', 'valkey:cache'])"),
		),
		mcp.WithString("region",
			mcp.Description("Region ID (default: prg1, use region_list to see options)"),
		),
		mcp.WithObject("env_vars",
			mcp.Description("Environment variables for the application"),
		),
	)

	s.AddTool(workflowCreateAppTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectName, err := request.RequireString("project_name")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROJECT_NAME",
				"Project name is required",
				"Provide a valid project name",
			), nil
		}

		appType, err := request.RequireString("app_type")
		if err != nil {
			return ErrorResponse(
				"INVALID_APP_TYPE",
				"Application type is required",
				"Choose from: nodejs, php-laravel, python-django, static-react, go, dotnet",
			), nil
		}

		appHostname, err := request.RequireString("app_hostname")
		if err != nil {
			return ErrorResponse(
				"INVALID_APP_HOSTNAME",
				"Application hostname is required",
				"Provide a valid hostname (lowercase letters and numbers only)",
			), nil
		}

		// Validate hostname
		if !isValidServiceName(appHostname) {
			return ErrorResponse(
				"INVALID_HOSTNAME_FORMAT",
				fmt.Sprintf("Hostname '%s' contains invalid characters", appHostname),
				"Use only lowercase letters and numbers (e.g., 'app', 'api1')",
			), nil
		}

		region := request.GetString("region", "prg1")
		
		// Get additional services
		additionalServices := []string{}
		args := request.GetArguments()
		if servicesArray, ok := args["additional_services"].([]interface{}); ok {
			for _, s := range servicesArray {
				if str, ok := s.(string); ok {
					additionalServices = append(additionalServices, str)
				}
			}
		}

		// Get environment variables
		envVars := make(map[string]string)
		if envObj, ok := args["env_vars"]; ok {
			if envMap, ok := envObj.(map[string]interface{}); ok {
				for k, v := range envMap {
					envVars[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		var response strings.Builder
		response.WriteString(fmt.Sprintf("🚀 Creating %s application project '%s'\n\n", appType, projectName))

		// Step 1: Create project
		response.WriteString("Step 1/4: Creating project...\n")
		project, err := client.CreateProject(ctx, api.CreateProjectRequest{
			Name:     projectName,
			RegionID: region,
		})
		if err != nil {
			return ErrorResponseWithNext(
				"PROJECT_CREATE_FAILED",
				fmt.Sprintf("Failed to create project: %v", err),
				"Check project name and try again",
				"project_create",
			), nil
		}
		response.WriteString(fmt.Sprintf("✅ Project created with ID: %s\n\n", project.ID))

		// Step 2: Build import YAML
		response.WriteString("Step 2/4: Preparing services configuration...\n")
		
		// Start with the application service
		importYAML := fmt.Sprintf(`services:
  - hostname: %s
    type: %s
    mode: NON_HA`, appHostname, getServiceTypeForApp(appType))
		
		// Add environment variables if provided
		if len(envVars) > 0 {
			importYAML += "\n    envVariables:"
			for k, v := range envVars {
				importYAML += fmt.Sprintf("\n      %s: %s", k, v)
			}
		}

		// Add additional services
		for _, service := range additionalServices {
			parts := strings.Split(service, ":")
			if len(parts) != 2 {
				response.WriteString(fmt.Sprintf("⚠️  Invalid service format '%s', skipping...\n", service))
				continue
			}
			
			serviceType := parts[0]
			hostname := parts[1]
			
			if !isValidServiceName(hostname) {
				response.WriteString(fmt.Sprintf("⚠️  Invalid hostname '%s' for %s, skipping...\n", hostname, serviceType))
				continue
			}

			// Build basic service YAML structure based on type
			serviceConfig := buildServiceConfig(serviceType, hostname)
			if serviceConfig == "" {
				response.WriteString(fmt.Sprintf("⚠️  Unknown service type '%s', skipping...\n", serviceType))
				continue
			}
			
			// Add to import YAML
			importYAML += "\n" + serviceConfig
		}
		
		response.WriteString("✅ Services configuration prepared\n\n")

		// Step 3: Import services
		response.WriteString("Step 3/4: Creating services...\n")
		err = client.ImportProjectServices(ctx, project.ID, project.ClientID, importYAML)
		if err != nil {
			// Try to delete the project to clean up
			client.DeleteProject(ctx, project.ID) //nolint:errcheck
			return ErrorResponseWithNext(
				"IMPORT_FAILED",
				fmt.Sprintf("Failed to create services: %v", err),
				"The project was created but service import failed. Try manually with 'project_import'",
				"project_import",
			), nil
		}
		
		// Count created services
		serviceCount := 1 + len(additionalServices)
		response.WriteString(fmt.Sprintf("✅ Created %d services\n", serviceCount))
		response.WriteString(fmt.Sprintf("   - %s (%s)\n", appHostname, getServiceTypeForApp(appType)))
		for _, service := range additionalServices {
			parts := strings.Split(service, ":")
			if len(parts) == 2 {
				response.WriteString(fmt.Sprintf("   - %s (%s)\n", parts[1], parts[0]))
			}
		}
		response.WriteString("\n")

		// Step 4: Provide next steps
		response.WriteString("Step 4/4: Setup complete! Next steps:\n\n")
		response.WriteString("1. Find deployment pattern:\n")
		response.WriteString(fmt.Sprintf("   knowledge_search_patterns tags=[\"%s\"]\n\n", appType))
		response.WriteString("2. Extract zerops.yml from pattern and update setup name to match your service\n\n")
		response.WriteString("3. Connect to VPN:\n")
		response.WriteString(fmt.Sprintf("   vpn_connect project_id=\"%s\"\n\n", project.ID))
		response.WriteString("4. Deploy your application:\n")
		response.WriteString("   deploy_push work_dir=\"./\" config_path=\"./zerops.yml\"\n\n")

		// Add database connection info if PostgreSQL was added
		for _, service := range additionalServices {
			if strings.HasPrefix(service, "postgresql:") {
				hostname := strings.Split(service, ":")[1]
				response.WriteString("📝 Database connection info:\n")
				response.WriteString(fmt.Sprintf("   Host: ${%s_hostname}\n", hostname))
				response.WriteString(fmt.Sprintf("   User: ${%s_user}\n", hostname))
				response.WriteString(fmt.Sprintf("   Password: ${%s_password}\n", hostname))
				response.WriteString(fmt.Sprintf("   Database: ${%s_dbName}\n", hostname))
				response.WriteString("\n")
			}
		}

		return SuccessResponse(map[string]interface{}{
			"message":          fmt.Sprintf("Successfully created %s project '%s'", appType, projectName),
			"project_id":       project.ID,
			"project_name":     project.Name,
			"region":           region,
			"app_hostname":     appHostname,
			"services_created": serviceCount,
			"next_step":        fmt.Sprintf("Use 'knowledge_search_patterns tags=[\"%s\"]' to find deployment pattern", appType),
			"details":          response.String(),
		}), nil
	})

	// Register workflow_clone tool
	workflowCloneTool := mcp.NewTool(
		"workflow_clone",
		mcp.WithDescription("Clone an existing project with all its services"),
		mcp.WithString("source_project_id",
			mcp.Required(),
			mcp.Description("ID of the project to clone from"),
		),
		mcp.WithString("new_project_name",
			mcp.Required(),
			mcp.Description("Name for the new project"),
		),
		mcp.WithString("region",
			mcp.Description("Region for the new project (default: same as source)"),
		),
		mcp.WithBoolean("include_data",
			mcp.Description("Include data in cloned databases (default: false)"),
		),
	)

	s.AddTool(workflowCloneTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sourceProjectID, err := request.RequireString("source_project_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SOURCE_PROJECT",
				"Source project ID is required",
				"Provide a valid project ID to clone from",
			), nil
		}

		newProjectName, err := request.RequireString("new_project_name")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROJECT_NAME",
				"New project name is required",
				"Provide a name for the new project",
			), nil
		}

		includeData := request.GetBool("include_data", false)

		// Get source project details
		sourceProject, err := client.GetProject(ctx, sourceProjectID)
		if err != nil {
			return ErrorResponse(
				"SOURCE_PROJECT_NOT_FOUND",
				fmt.Sprintf("Failed to get source project: %v", err),
				"Check the project ID and try again",
			), nil
		}

		region := request.GetString("region", "prg1") // Default to prg1 if not specified

		// Get services from source project
		services, err := client.ListServices(ctx, sourceProjectID)
		if err != nil {
			return ErrorResponse(
				"SERVICES_FETCH_FAILED",
				fmt.Sprintf("Failed to get source project services: %v", err),
				"Check the project ID and permissions",
			), nil
		}

		steps := []string{}

		// Step 1: Create new project
		newProject, err := client.CreateProject(ctx, api.CreateProjectRequest{
			Name:     newProjectName,
			RegionID: region,
		})
		if err != nil {
			return ErrorResponse(
				"PROJECT_CREATION_FAILED",
				fmt.Sprintf("Failed to create new project: %v", err),
				"Check the project name and try again",
			), nil
		}
		steps = append(steps, fmt.Sprintf("✓ Created new project '%s' (ID: %s)", newProjectName, newProject.ID))

		// Step 2: Generate import YAML for all services
		var importYAML strings.Builder
		importYAML.WriteString("services:\n")
		
		for _, service := range services {
			// Skip system services
			if strings.HasPrefix(service.Name, "prg-") {
				continue
			}

			importYAML.WriteString(fmt.Sprintf("  - hostname: %s\n", service.Name))
			// Construct type from service stack type info
			serviceType := fmt.Sprintf("%s@%s", 
				service.ServiceStackTypeInfo.ServiceStackTypeName,
				service.ServiceStackTypeInfo.ServiceStackTypeVersionName)
			importYAML.WriteString(fmt.Sprintf("    type: %s\n", serviceType))
			
			// Add mode for services that support it
			if service.ServiceStackTypeInfo.ServiceStackTypeCategory == "database" || 
			   service.ServiceStackTypeInfo.ServiceStackTypeCategory == "cache" {
				importYAML.WriteString(fmt.Sprintf("    mode: %s\n", service.Mode))
			}
			
			// Add minContainers/maxContainers for services with scaling
			if service.MinContainers > 1 || service.MaxContainers > 1 {
				importYAML.WriteString(fmt.Sprintf("    minContainers: %d\n", service.MinContainers))
				if service.MaxContainers > service.MinContainers {
					importYAML.WriteString(fmt.Sprintf("    maxContainers: %d\n", service.MaxContainers))
				}
			}
			
			// Add subdomain access if enabled
			if service.SubdomainAccess {
				importYAML.WriteString("    enableSubdomainAccess: true\n")
			}
			
			// Handle object-storage specific fields
			if service.ServiceStackTypeInfo.ServiceStackTypeName == "object-storage" {
				// Note: We can't access the original objectStorageSize and policy from the service list
				// These will use defaults unless manually configured
				importYAML.WriteString("    objectStorageSize: 2\n")
				importYAML.WriteString("    objectStoragePolicy: private\n")
			}
			
			// Note: Environment variables will need to be set separately
			// as they can't be included in the import YAML
		}

		// Step 3: Import services
		err = client.ImportProjectServices(ctx, newProject.ID, newProject.ClientID, importYAML.String())
		if err != nil {
			// Cleanup: delete the new project
			client.DeleteProject(ctx, newProject.ID)
			return ErrorResponse(
				"SERVICE_IMPORT_FAILED",
				fmt.Sprintf("Failed to import services: %v", err),
				"Service configuration may be incompatible",
			), nil
		}
		// Count actual services (excluding system ones)
		serviceCount := 0
		for _, service := range services {
			if !strings.HasPrefix(service.Name, "prg-") {
				serviceCount++
			}
		}
		steps = append(steps, fmt.Sprintf("✓ Cloned %d services from source project", serviceCount))

		// Prepare notes about what wasn't cloned
		var notes strings.Builder
		notes.WriteString("\nImportant notes:\n")
		notes.WriteString("- Environment variables were NOT cloned (set them manually)\n")
		notes.WriteString("- Service data was NOT cloned")
		if includeData {
			notes.WriteString(" (data cloning not yet implemented)\n")
		} else {
			notes.WriteString("\n")
		}
		notes.WriteString("- Deployment configurations need to be set up separately\n")
		notes.WriteString("- SSL certificates need to be configured if used\n")

		return SuccessResponse(map[string]interface{}{
			"message":           fmt.Sprintf("Successfully cloned project '%s' to '%s'", sourceProject.Name, newProjectName),
			"source_project_id": sourceProjectID,
			"new_project_id":    newProject.ID,
			"new_project_name":  newProjectName,
			"region":            region,
			"services_cloned":   serviceCount,
			"steps":             strings.Join(steps, "\n"),
			"notes":             notes.String(),
			"next_step":         fmt.Sprintf("Configure environment variables and deploy your applications to project %s", newProject.ID),
		}), nil
	})

	// Register workflow_diagnose tool
	workflowDiagnoseTool := mcp.NewTool(
		"workflow_diagnose",
		mcp.WithDescription("Diagnose common issues and provide solutions"),
		mcp.WithString("issue_type",
			mcp.Required(),
			mcp.Description("Type of issue (deployment, service, vpn, connection, performance)"),
		),
		mcp.WithString("project_id",
			mcp.Description("Project ID (required for most diagnostics)"),
		),
		mcp.WithString("service_id",
			mcp.Description("Service ID (for service-specific issues)"),
		),
		mcp.WithObject("symptoms",
			mcp.Description("Additional symptoms or error messages"),
		),
	)

	s.AddTool(workflowDiagnoseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		issueType, err := request.RequireString("issue_type")
		if err != nil {
			return ErrorResponse(
				"INVALID_ISSUE_TYPE",
				"Issue type is required",
				"Choose from: deployment, service, vpn, connection, performance",
			), nil
		}

		projectID := request.GetString("project_id", "")
		serviceID := request.GetString("service_id", "")
		
		// Get symptoms
		symptoms := make(map[string]string)
		args := request.GetArguments()
		if sympObj, ok := args["symptoms"]; ok {
			if sympMap, ok := sympObj.(map[string]interface{}); ok {
				for k, v := range sympMap {
					symptoms[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		var response strings.Builder
		response.WriteString(fmt.Sprintf("🔍 Diagnosing %s issue...\n\n", issueType))

		switch strings.ToLower(issueType) {
		case "deployment":
			response.WriteString("📦 Deployment Diagnostics\n\n")
			
			// Check VPN status
			response.WriteString("1. Checking VPN status...\n")
			if zcliWrapper.IsVPNConnected(ctx) {
				response.WriteString("   ✅ VPN is connected\n")
			} else {
				response.WriteString("   ❌ VPN is NOT connected\n")
				response.WriteString("   → Solution: Use 'vpn_connect' tool with your project ID\n")
			}
			response.WriteString("\n")

			// Check for common deployment errors
			response.WriteString("2. Common deployment issues:\n")
			if errorMsg, ok := symptoms["error"]; ok {
				if strings.Contains(errorMsg, "zerops.yml not found") {
					response.WriteString("   ❌ Missing zerops.yml configuration\n")
					response.WriteString("   → Solution: Use 'knowledge_search_patterns' to find a deployment pattern\n")
				} else if strings.Contains(errorMsg, "authentication") {
					response.WriteString("   ❌ Authentication issue\n")
					response.WriteString("   → Solution: Ensure ZEROPS_API_KEY is set correctly\n")
				} else if strings.Contains(errorMsg, "build failed") {
					response.WriteString("   ❌ Build process failed\n")
					response.WriteString("   → Solution: Check build logs with 'deploy_logs'\n")
				}
			}
			response.WriteString("\n")

			// Provide deployment checklist
			response.WriteString("3. Deployment checklist:\n")
			response.WriteString("   ☐ VPN connected to correct project\n")
			response.WriteString("   ☐ zerops.yml exists and is valid\n")
			response.WriteString("   ☐ Service exists in project\n")
			response.WriteString("   ☐ Git repository initialized (if using git)\n")
			response.WriteString("   ☐ All dependencies are specified\n")
			response.WriteString("\n")

			response.WriteString("Recommended actions:\n")
			response.WriteString("1. Validate config: config_validate config_path=\"./zerops.yml\"\n")
			response.WriteString("2. Check VPN: vpn_status\n")
			response.WriteString("3. Try deployment: deploy_push work_dir=\"./\"\n")

		case "service":
			response.WriteString("🔧 Service Diagnostics\n\n")
			
			if serviceID != "" && projectID != "" {
				// Get service details
				service, err := client.GetService(ctx, serviceID)
				if err != nil {
					response.WriteString("❌ Could not fetch service details\n")
				} else {
					response.WriteString(fmt.Sprintf("Service: %s\n", service.Name))
					response.WriteString(fmt.Sprintf("Status: %s\n", service.Status))
					response.WriteString(fmt.Sprintf("Mode: %s\n", service.Mode))
					response.WriteString("\n")

					if service.Status != "running" {
						response.WriteString("⚠️  Service is not running!\n")
						response.WriteString("→ Solution: Use 'service_start' to start the service\n\n")
					}
				}
			}

			response.WriteString("Common service issues:\n")
			response.WriteString("1. Service won't start:\n")
			response.WriteString("   - Check logs: service_logs\n")
			response.WriteString("   - Verify environment variables\n")
			response.WriteString("   - Check resource limits\n\n")
			
			response.WriteString("2. Service crashes:\n")
			response.WriteString("   - Review application logs\n")
			response.WriteString("   - Check memory usage\n")
			response.WriteString("   - Verify dependencies\n\n")

			response.WriteString("3. Connection issues:\n")
			response.WriteString("   - Verify service naming (no hyphens!)\n")
			response.WriteString("   - Check environment variable references\n")
			response.WriteString("   - Ensure services are in same project\n")

		case "vpn":
			response.WriteString("🔒 VPN Diagnostics\n\n")
			
			// Check current VPN status
			response.WriteString("1. Current VPN status:\n")
			if zcliWrapper.IsVPNConnected(ctx) {
				response.WriteString("   ✅ VPN is connected\n")
				// Try to get connected project info
				if projectID != "" {
					response.WriteString(fmt.Sprintf("   Connected to project: %s\n", projectID))
				}
			} else {
				response.WriteString("   ❌ VPN is NOT connected\n")
			}
			response.WriteString("\n")

			response.WriteString("2. VPN troubleshooting steps:\n")
			response.WriteString("   a) Ensure zcli is installed:\n")
			response.WriteString("      - macOS: brew install zeropsio/tap/zcli\n")
			response.WriteString("      - Linux: See Zerops documentation\n\n")
			
			response.WriteString("   b) Check sudo permissions:\n")
			response.WriteString("      - VPN operations require sudo\n")
			response.WriteString("      - You may need to enter your password\n\n")
			
			response.WriteString("   c) Common VPN errors:\n")
			response.WriteString("      - 'Already connected': Disconnect first with vpn_disconnect\n")
			response.WriteString("      - 'Permission denied': Run with proper sudo access\n")
			response.WriteString("      - 'Invalid project': Check project ID is correct\n\n")

			response.WriteString("Recommended actions:\n")
			response.WriteString("1. Check status: vpn_status\n")
			response.WriteString("2. Disconnect if needed: vpn_disconnect\n")
			response.WriteString(fmt.Sprintf("3. Connect: vpn_connect project_id=\"%s\"\n", projectID))

		case "connection":
			response.WriteString("🔌 Connection Diagnostics\n\n")
			
			response.WriteString("Database connection checklist:\n")
			response.WriteString("1. Service naming:\n")
			response.WriteString("   ✓ Use only lowercase letters and numbers\n")
			response.WriteString("   ✗ No hyphens, underscores, or special characters\n\n")
			
			response.WriteString("2. Environment variables:\n")
			response.WriteString("   Use cross-service references:\n")
			response.WriteString("   - DB_HOST=${servicename_hostname}\n")
			response.WriteString("   - DB_USER=${servicename_user}\n")
			response.WriteString("   - DB_PASSWORD=${servicename_password}\n")
			response.WriteString("   - DB_DATABASE=${servicename_dbName}\n\n")
			
			response.WriteString("3. Common connection strings:\n")
			response.WriteString("   PostgreSQL:\n")
			response.WriteString("   postgresql://${db_user}:${db_password}@${db_hostname}:5432/${db_dbName}\n\n")
			response.WriteString("   MariaDB/MySQL:\n")
			response.WriteString("   mysql://${db_user}:${db_password}@${db_hostname}:3306/${db_dbName}\n\n")

			response.WriteString("4. Troubleshooting steps:\n")
			response.WriteString("   - Verify both services are running\n")
			response.WriteString("   - Check service logs for connection errors\n")
			response.WriteString("   - Ensure services are in the same project\n")
			response.WriteString("   - Verify environment variables are set correctly\n")

		case "performance":
			response.WriteString("⚡ Performance Diagnostics\n\n")
			
			response.WriteString("Performance optimization checklist:\n\n")
			
			response.WriteString("1. Application level:\n")
			response.WriteString("   - Enable production mode (NODE_ENV=production, etc.)\n")
			response.WriteString("   - Use build caching in zerops.yml\n")
			response.WriteString("   - Optimize build commands\n")
			response.WriteString("   - Minimize deployment package size\n\n")
			
			response.WriteString("2. Service configuration:\n")
			response.WriteString("   - Check service mode (HA vs NON_HA)\n")
			response.WriteString("   - Review resource allocation\n")
			response.WriteString("   - Monitor service logs for errors\n\n")
			
			response.WriteString("3. Database optimization:\n")
			response.WriteString("   - Use connection pooling\n")
			response.WriteString("   - Add appropriate indexes\n")
			response.WriteString("   - Monitor slow queries\n\n")
			
			response.WriteString("4. Caching strategy:\n")
			response.WriteString("   - Use Valkey/KeyDB for caching\n")
			response.WriteString("   - Cache static assets\n")
			response.WriteString("   - Implement application-level caching\n")

		default:
			return ErrorResponse(
				"UNKNOWN_ISSUE_TYPE",
				fmt.Sprintf("Unknown issue type: %s", issueType),
				"Choose from: deployment, service, vpn, connection, performance",
			), nil
		}

		response.WriteString("\n\n💡 Additional resources:\n")
		response.WriteString("- Check service logs: service_logs\n")
		response.WriteString("- View project info: project_info\n")
		response.WriteString("- List all services: service_list\n")

		return SuccessResponse(map[string]interface{}{
			"message":    fmt.Sprintf("Completed %s diagnostics", issueType),
			"issue_type": issueType,
			"project_id": projectID,
			"service_id": serviceID,
			"diagnosis":  response.String(),
		}), nil
	})
}

// getServiceTypeForApp returns the appropriate service type for an application type
func getServiceTypeForApp(appType string) string {
	switch appType {
	case "nodejs":
		return "nodejs@20"
	case "php-laravel":
		return "php-apache@8.3"
	case "python-django":
		return "python@3.11"
	case "static-react":
		return "static"
	case "go":
		return "go@1.21"
	case "dotnet":
		return "dotnet@8.0"
	default:
		return "nodejs@20" // Default fallback
	}
}

// isRuntimeService checks if a service type is a runtime service (not database/cache)
func isRuntimeService(serviceType string) bool {
	nonRuntimePrefixes := []string{
		"postgresql", "mariadb", "mongodb", "mysql",
		"redis", "valkey", "keydb", "rabbitmq", 
		"elasticsearch", "opensearch",
	}
	
	typeLower := strings.ToLower(serviceType)
	for _, prefix := range nonRuntimePrefixes {
		if strings.HasPrefix(typeLower, prefix) {
			return false
		}
	}
	
	return true
}

// buildServiceConfig builds a YAML configuration snippet for a service
func buildServiceConfig(serviceType, hostname string) string {
	// For database services, use simple configuration
	switch {
	case strings.HasPrefix(serviceType, "postgresql"):
		return fmt.Sprintf(`  - hostname: %s
    type: %s
    mode: HA`, hostname, serviceType)
		
	case strings.HasPrefix(serviceType, "mariadb"):
		return fmt.Sprintf(`  - hostname: %s
    type: %s
    mode: HA`, hostname, serviceType)
		
	case strings.HasPrefix(serviceType, "mongodb"):
		return fmt.Sprintf(`  - hostname: %s
    type: %s
    mode: HA`, hostname, serviceType)
		
	case strings.HasPrefix(serviceType, "valkey"), strings.HasPrefix(serviceType, "keydb"):
		return fmt.Sprintf(`  - hostname: %s
    type: %s
    mode: NON_HA`, hostname, serviceType)
		
	default:
		// Unknown service type
		return ""
	}
}