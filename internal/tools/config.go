package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
	"github.com/zeropsio/zerops-mcp-v3/internal/knowledge"
	"gopkg.in/yaml.v3"
)

// RegisterConfigTools registers all configuration-related tools
func RegisterConfigTools(s *server.MCPServer, client *api.Client) {
	// Register config_templates tool
	configTemplatesTool := mcp.NewTool(
		"config_templates",
		mcp.WithDescription("List all available configuration templates with descriptions"),
	)

	s.AddTool(configTemplatesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templates := []struct {
			Name        string
			Description string
			Runtime     string
			Features    []string
		}{
			{
				Name:        "nodejs",
				Description: "Node.js application template",
				Runtime:     "nodejs@20",
				Features:    []string{"npm build", "customizable port", "NODE_ENV support"},
			},
			{
				Name:        "php-laravel",
				Description: "PHP Laravel framework template",
				Runtime:     "php-apache@8.3",
				Features:    []string{"composer", "artisan commands", "Apache web server", "cache optimization"},
			},
			{
				Name:        "python-django",
				Description: "Python Django framework template",
				Runtime:     "python@3.11",
				Features:    []string{"pip requirements", "gunicorn server", "static files collection"},
			},
			{
				Name:        "static-react",
				Description: "Static React application template",
				Runtime:     "nodejs@20 (build only)",
				Features:    []string{"npm build", "static file serving", "SPA support"},
			},
			{
				Name:        "go",
				Description: "Go application template",
				Runtime:     "go@1.21",
				Features:    []string{"go build", "binary execution", "module caching"},
			},
			{
				Name:        "dotnet",
				Description: ".NET application template",
				Runtime:     "dotnet@8.0",
				Features:    []string{"dotnet publish", "ASP.NET Core support", "NuGet caching"},
			},
		}

		var response strings.Builder
		response.WriteString("Available configuration templates:\n\n")

		for _, tmpl := range templates {
			response.WriteString(fmt.Sprintf("📄 %s\n", tmpl.Name))
			response.WriteString(fmt.Sprintf("   Description: %s\n", tmpl.Description))
			response.WriteString(fmt.Sprintf("   Runtime: %s\n", tmpl.Runtime))
			response.WriteString(fmt.Sprintf("   Features:\n"))
			for _, feature := range tmpl.Features {
				response.WriteString(fmt.Sprintf("     - %s\n", feature))
			}
			response.WriteString("\n")
		}

		response.WriteString("Usage: Search for framework patterns to get deployment configurations\n")
		response.WriteString("Example: knowledge_search_patterns tags=[\"nodejs\"]\n\n")
		
		response.WriteString("Additional service templates available for project import:\n")
		response.WriteString("- postgresql (PostgreSQL database)\n")
		response.WriteString("- mariadb (MariaDB database, version 10.6 only)\n")
		response.WriteString("- mongodb (MongoDB database)\n")
		response.WriteString("- valkey (Redis-compatible cache)\n")
		response.WriteString("- keydb (Redis-compatible cache)\n")
		response.WriteString("- rabbitmq (Message queue)\n")
		response.WriteString("- elasticsearch (Search engine)\n")

		return SuccessResponse(map[string]interface{}{
			"message":         "Configuration templates listed",
			"template_count":  len(templates),
			"next_step":       "Use 'knowledge_search_patterns' to find deployment patterns for your framework",
		}), nil
	})

	// Register env_vars_show tool
	envVarsShowTool := mcp.NewTool(
		"env_vars_show",
		mcp.WithDescription("Show environment variables for a service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to show environment variables for"),
		),
		mcp.WithBoolean("show_values",
			mcp.Description("Show actual values (default: false, shows masked values)"),
		),
	)

	s.AddTool(envVarsShowTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		showValues := request.GetBool("show_values", false)

		// Get service details including environment variables
		service, err := client.GetService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		if len(service.EnvVariables) == 0 {
			return SuccessResponse(map[string]interface{}{
				"message":      fmt.Sprintf("Service '%s' has no environment variables configured", service.Name),
				"service_id":   serviceID,
				"service_name": service.Name,
				"env_count":    0,
				"next_step":    "Environment variables can only be set during service creation via import",
			}), nil
		}

		var response strings.Builder
		response.WriteString(fmt.Sprintf("Environment variables for service '%s':\n\n", service.Name))

		// Group variables by category
		type envVar struct {
			key   string
			value string
		}
		
		systemVars := []envVar{}
		userVars := []envVar{}
		
		for key, value := range service.EnvVariables {
			valueStr := fmt.Sprintf("%v", value)
			if strings.HasPrefix(key, "_") || 
			   strings.HasPrefix(key, "ZEROPS_") ||
			   strings.Contains(strings.ToLower(key), "hostname") ||
			   strings.Contains(strings.ToLower(key), "password") {
				systemVars = append(systemVars, envVar{key: key, value: valueStr})
			} else {
				userVars = append(userVars, envVar{key: key, value: valueStr})
			}
		}

		// Display system variables
		if len(systemVars) > 0 {
			response.WriteString("🔧 System Variables:\n")
			for _, env := range systemVars {
				value := env.value
				if !showValues && (strings.Contains(strings.ToLower(env.key), "password") || 
				                   strings.Contains(strings.ToLower(env.key), "secret") ||
				                   strings.Contains(strings.ToLower(env.key), "key")) {
					value = strings.Repeat("*", 8)
				}
				response.WriteString(fmt.Sprintf("  %s = %s\n", env.key, value))
			}
			response.WriteString("\n")
		}

		// Display user variables
		if len(userVars) > 0 {
			response.WriteString("👤 User Variables:\n")
			for _, env := range userVars {
				value := env.value
				if !showValues && (strings.Contains(strings.ToLower(env.key), "password") || 
				                   strings.Contains(strings.ToLower(env.key), "secret") ||
				                   strings.Contains(strings.ToLower(env.key), "key")) {
					value = strings.Repeat("*", 8)
				}
				response.WriteString(fmt.Sprintf("  %s = %s\n", env.key, value))
			}
			response.WriteString("\n")
		}

		// Add cross-service reference examples
		response.WriteString("💡 Cross-service references:\n")
		response.WriteString("Use ${servicename_variablename} to reference variables from other services\n")
		response.WriteString("Example: DB_HOST=${database_hostname}\n")

		if !showValues {
			response.WriteString("\nNote: Sensitive values are masked. Use show_values=true to reveal them.\n")
		}

		return SuccessResponse(map[string]interface{}{
			"message":       fmt.Sprintf("Found %d environment variables", len(service.EnvVariables)),
			"service_id":    serviceID,
			"service_name":  service.Name,
			"env_count":     len(service.EnvVariables),
			"system_count":  len(systemVars),
			"user_count":    len(userVars),
			"configuration": response.String(),
		}), nil
	})
	// Register config_validate tool
	configValidateTool := mcp.NewTool(
		"config_validate",
		mcp.WithDescription("Validate a zerops.yml configuration file"),
		mcp.WithString("config_path",
			mcp.Required(),
			mcp.Description("Path to the zerops.yml file to validate"),
		),
		mcp.WithBoolean("strict",
			mcp.Description("Enable strict validation (default: false)"),
		),
	)

	s.AddTool(configValidateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		configPath, err := request.RequireString("config_path")
		if err != nil {
			return ErrorResponse(
				"INVALID_CONFIG_PATH",
				"Configuration path is required",
				"Provide the path to your zerops.yml file",
			), nil
		}

		strict := request.GetBool("strict", false)

		// Check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return ErrorResponse(
				"CONFIG_NOT_FOUND",
				fmt.Sprintf("Configuration file not found: %s", configPath),
				"Check the file path and ensure the file exists",
			), nil
		}

		// Read the configuration file
		content, err := os.ReadFile(configPath)
		if err != nil {
			return ErrorResponse(
				"CONFIG_READ_ERROR",
				fmt.Sprintf("Failed to read configuration file: %v", err),
				"Ensure the file is readable and not corrupted",
			), nil
		}

		// Parse YAML
		var config map[string]interface{}
		if err := yaml.Unmarshal(content, &config); err != nil {
			return ErrorResponse(
				"CONFIG_PARSE_ERROR",
				fmt.Sprintf("Failed to parse YAML: %v", err),
				"Check YAML syntax - ensure proper indentation and structure",
			), nil
		}

		// Validate structure
		validationErrors := validateZeropsConfig(config, strict)
		if len(validationErrors) > 0 {
			errorMsg := "Configuration validation failed:\n"
			for _, err := range validationErrors {
				errorMsg += fmt.Sprintf("- %s\n", err)
			}
			return ErrorResponse(
				"CONFIG_INVALID",
				errorMsg,
				"Fix the validation errors in your zerops.yml file",
			), nil
		}

		// Extract key information
		serviceName := ""
		buildType := ""
		runtime := ""

		if configMap, ok := config["-"].(map[string]interface{}); ok {
			if setup, ok := configMap["setup"].(map[string]interface{}); ok {
				if name, ok := setup["name"].(string); ok {
					serviceName = name
				}
			}
			if build, ok := configMap["build"].(map[string]interface{}); ok {
				if base, ok := build["base"].(string); ok {
					runtime = base
				}
				if _, hasDeployFiles := build["deployFiles"]; hasDeployFiles {
					buildType = "static"
				} else {
					buildType = "runtime"
				}
			}
		}

		return SuccessResponse(map[string]interface{}{
			"message":      "Configuration is valid",
			"config_path":  configPath,
			"service_name": serviceName,
			"build_type":   buildType,
			"runtime":      runtime,
			"strict_mode":  strict,
			"next_step":    "Use 'deploy_push' to deploy with this configuration",
		}), nil
	})
	
	// Nginx config tool
	nginxTool := mcp.NewTool(
		"config_nginx",
		mcp.WithDescription("Get nginx configuration template for PHP frameworks (Laravel, Symfony, WordPress)"),
		mcp.WithString("framework",
			mcp.Required(),
			mcp.Description("Framework name (e.g., 'laravel', 'symfony', 'wordpress')"),
		),
	)

	s.AddTool(nginxTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		framework, err := request.RequireString("framework")
		if err != nil {
			return ErrorResponse(
				"INVALID_FRAMEWORK",
				"Framework name is required",
				"Provide a framework name like 'laravel', 'symfony', or 'wordpress'",
			), nil
		}

		// Get nginx config from knowledge base
		config, err := knowledge.GetNginxConfig(framework)
		if err != nil {
			return ErrorResponse(
				"CONFIG_NOT_FOUND",
				fmt.Sprintf("Failed to get nginx config: %v", err),
				"Try using 'laravel', 'symfony', 'wordpress', or 'default'",
			), nil
		}

		var response strings.Builder
		response.WriteString(fmt.Sprintf("# Nginx configuration for %s\n\n", framework))
		response.WriteString("**Instructions**:\n")
		response.WriteString("1. Save this configuration as `site.conf.tmpl` in your project root\n")
		response.WriteString("2. Add to your zerops.yml:\n")
		response.WriteString("   ```yaml\n")
		response.WriteString("   run:\n")
		response.WriteString("     siteConfigPath: site.conf.tmpl\n")
		response.WriteString("   ```\n")
		response.WriteString("3. Include `site.conf.tmpl` in your deployFiles\n\n")
		response.WriteString("**Configuration Template**:\n")
		response.WriteString("```nginx\n")
		response.WriteString(config)
		response.WriteString("\n```\n\n")
		response.WriteString("**Important Notes**:\n")
		response.WriteString("- {{.PhpSocket}} will be replaced with the PHP-FPM socket path\n")
		response.WriteString("- {{.DocumentRoot}} will be replaced with your document root (default: /var/www)\n")
		response.WriteString("- For Laravel: Document root should be /var/www/public\n")
		response.WriteString("- For WordPress/Symfony: Document root is usually /var/www\n")

		return mcp.NewToolResultText(response.String()), nil
	})
}

// validateZeropsConfig validates the zerops.yml configuration structure
func validateZeropsConfig(config map[string]interface{}, strict bool) []string {
	var errors []string

	// Check for root key "-"
	rootConfig, ok := config["-"].(map[string]interface{})
	if !ok {
		errors = append(errors, "Missing root key '-' in configuration")
		return errors
	}

	// Validate setup section
	if setup, ok := rootConfig["setup"].(map[string]interface{}); ok {
		if name, ok := setup["name"].(string); ok {
			if !isValidServiceName(name) {
				errors = append(errors, fmt.Sprintf("Invalid service name '%s': use only lowercase letters and numbers", name))
			}
		} else {
			errors = append(errors, "Missing or invalid 'setup.name'")
		}
	} else {
		errors = append(errors, "Missing 'setup' section")
	}

	// Validate build section
	if build, ok := rootConfig["build"].(map[string]interface{}); ok {
		// Check for either deployFiles (static) or base (runtime)
		hasDeployFiles := false
		hasBase := false

		if _, ok := build["deployFiles"]; ok {
			hasDeployFiles = true
		}
		if _, ok := build["base"]; ok {
			hasBase = true
		}

		if !hasDeployFiles && !hasBase {
			errors = append(errors, "Build section must have either 'base' (for runtime) or 'deployFiles' (for static)")
		}

		if hasDeployFiles && hasBase && strict {
			errors = append(errors, "Build section should not have both 'base' and 'deployFiles'")
		}

		// Validate build commands are arrays if present
		for _, cmd := range []string{"prepareCommands", "buildCommands"} {
			if val, exists := build[cmd]; exists {
				if _, ok := val.([]interface{}); !ok {
					errors = append(errors, fmt.Sprintf("'build.%s' must be an array", cmd))
				}
			}
		}
	} else {
		errors = append(errors, "Missing 'build' section")
	}

	// Validate run section (required for runtime services)
	if _, hasRun := rootConfig["run"]; !hasRun {
		if build, ok := rootConfig["build"].(map[string]interface{}); ok {
			if _, hasBase := build["base"]; hasBase {
				errors = append(errors, "Missing 'run' section (required for runtime services)")
			}
		}
	}

	return errors
}

// isValidServiceName checks if the service name contains only lowercase letters and numbers
// and starts with a letter
func isValidServiceName(name string) bool {
	if name == "" {
		return false
	}
	// Service name must start with a letter
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}