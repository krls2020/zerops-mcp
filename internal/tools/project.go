package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
	"gopkg.in/yaml.v3"
)

// RegisterProjectTools registers all project-related tools
func RegisterProjectTools(s *server.MCPServer, client *api.Client) {
	// Register project_list tool
	projectListTool := mcp.NewTool(
		"project_list",
		mcp.WithDescription("List all projects in your Zerops account"),
	)

	s.AddTool(projectListTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get current user to determine client ID
		user, err := client.GetCurrentUser(ctx)
		if err != nil {
			return HandleAPIError(err), nil
		}

		if len(user.ClientUserList) == 0 {
			return ErrorResponse(
				"NO_CLIENT",
				"No client associations found for user",
				"Contact Zerops support to resolve account issues",
			), nil
		}

		// Use the first client ID
		clientID := user.ClientUserList[0].ClientID

		// List projects
		projects, err := client.ListProjects(ctx, clientID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		if len(projects) == 0 {
			return SuccessResponse(map[string]interface{}{
				"message":  "No projects found",
				"count":    0,
				"nextStep": "Use 'project_create' to create your first project",
			}), nil
		}

		// Format project list
		response := fmt.Sprintf("Found %d project(s):\n\n", len(projects))
		
		for i, project := range projects {
			response += fmt.Sprintf("%d. %s\n", i+1, project.Name)
			response += fmt.Sprintf("   ID: %s\n", project.ID)
			response += fmt.Sprintf("   Status: %s\n", project.Status)
			if project.Description != "" {
				response += fmt.Sprintf("   Description: %s\n", project.Description)
			}
			response += fmt.Sprintf("   Created: %s\n", project.Created.Format("2006-01-02 15:04:05"))
			response += "\n"
		}

		response += "Next step: Use 'project_info' for details or 'service_list' to see services"

		return mcp.NewToolResultText(response), nil
	})

	// Register project_create tool
	projectCreateTool := mcp.NewTool(
		"project_create",
		mcp.WithDescription("Create a new Zerops project"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Project name (any characters allowed)"),
		),
		mcp.WithString("region",
			mcp.Required(),
			mcp.Description("Region ID (use region_list to see available regions)"),
		),
		mcp.WithString("description",
			mcp.Description("Optional project description"),
		),
	)

	s.AddTool(projectCreateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		name, err := request.RequireString("name")
		if err != nil {
			return ErrorResponse(
				"INVALID_NAME",
				"Project name is required",
				"Provide a project name",
			), nil
		}

		region, err := request.RequireString("region")
		if err != nil {
			return ErrorResponse(
				"INVALID_REGION",
				"Region is required",
				"Use 'region_list' tool to see available regions",
			), nil
		}

		description := request.GetString("description", "")

		// No validation needed for project names - Zerops accepts any characters

		// Get current user to determine client ID
		user, err := client.GetCurrentUser(ctx)
		if err != nil {
			return HandleAPIError(err), nil
		}

		if len(user.ClientUserList) == 0 {
			return ErrorResponse(
				"NO_CLIENT",
				"No client associations found for user",
				"Contact Zerops support to resolve account issues",
			), nil
		}

		// Use the first client ID
		clientID := user.ClientUserList[0].ClientID

		// Create project
		project, err := client.CreateProject(ctx, api.CreateProjectRequest{
			Name:        name,
			RegionID:    region,
			ClientID:    clientID,
			Description: description,
			TagList:     []string{},
		})
		if err != nil {
			// Handle specific errors
			if strings.Contains(err.Error(), "already exists") {
				return ErrorResponse(
					"PROJECT_EXISTS",
					fmt.Sprintf("Project with name '%s' already exists", name),
					"Choose a different project name",
				), nil
			}
			if strings.Contains(err.Error(), "invalid region") {
				return ErrorResponseWithNext(
					"INVALID_REGION",
					fmt.Sprintf("Region '%s' is not valid", region),
					"Use a valid region ID from the list",
					"region_list",
				), nil
			}
			return HandleAPIError(err), nil
		}

		return SuccessResponse(map[string]interface{}{
			"message":    fmt.Sprintf("Project '%s' created successfully", name),
			"projectId":  project.ID,
			"name":       project.Name,
			"region":     region,
			"status":     project.Status,
			"nextStep":   "Use 'project_import' to add services to your project",
		}), nil
	})

	// Register project_info tool
	projectInfoTool := mcp.NewTool(
		"project_info",
		mcp.WithDescription("Get detailed information about a specific project"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID (use project_list to find IDs)"),
		),
	)

	s.AddTool(projectInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, err := request.RequireString("project_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROJECT_ID",
				"Project ID is required",
				"Provide a valid project ID from 'project_list'",
			), nil
		}

		// Get project details
		project, err := client.GetProject(ctx, projectID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				return ErrorResponseWithNext(
					"PROJECT_NOT_FOUND",
					fmt.Sprintf("Project with ID '%s' not found", projectID),
					"Check the project ID or use 'project_list' to find valid projects",
					"project_list",
				), nil
			}
			return HandleAPIError(err), nil
		}

		// Get services count
		services, err := client.ListServices(ctx, projectID)
		serviceCount := 0
		if err == nil {
			serviceCount = len(services)
		}

		return SuccessResponse(map[string]interface{}{
			"projectId":    project.ID,
			"name":         project.Name,
			"description":  project.Description,
			"status":       project.Status,
			"mode":         project.Mode,
			"created":      project.Created.Format("2006-01-02 15:04:05"),
			"lastUpdate":   project.LastUpdate.Format("2006-01-02 15:04:05"),
			"serviceCount": serviceCount,
			"tags":         project.TagList,
			"nextStep":     "Use 'service_list' to see services or 'project_import' to add services",
		}), nil
	})

	// Register project_import tool
	projectImportTool := mcp.NewTool(
		"project_import",
		mcp.WithDescription("Import services to a project using YAML configuration. Supports Zerops preprocessing syntax like <@generateRandomString(<32>)> for generating secrets"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID to import services into"),
		),
		mcp.WithString("yaml",
			mcp.Required(),
			mcp.Description("YAML configuration defining services to import. Supports preprocessing functions in envSecrets"),
		),
	)

	s.AddTool(projectImportTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, err := request.RequireString("project_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROJECT_ID",
				"Project ID is required",
				"Provide a valid project ID from 'project_list'",
			), nil
		}

		yamlConfig, err := request.RequireString("yaml")
		if err != nil {
			return ErrorResponse(
				"INVALID_YAML",
				"YAML configuration is required",
				"Provide a valid YAML configuration for services",
			), nil
		}
		
		// Debug: Log received YAML
		if strings.Contains(yamlConfig, "<@") {
			// Log first 200 chars to see what we received
			preview := yamlConfig
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			// Note: In MCP server, we can't use stdout logging
			// The debugging will show in the preprocessing message
		}

		// Check if YAML contains projectConfig (from patterns)
		var projectConfig map[string]interface{}
		var servicesYAML string
		
		// Try to parse YAML to extract projectConfig if present
		var yamlData map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlConfig), &yamlData); err == nil {
			// Check if this is a pattern-style YAML with projectConfig
			if pc, ok := yamlData["projectConfig"].(map[string]interface{}); ok {
				projectConfig = pc
				// Extract just the services part
				if services, ok := yamlData["services"]; ok {
					servicesData := map[string]interface{}{"services": services}
					if servicesBytes, err := yaml.Marshal(servicesData); err == nil {
						servicesYAML = string(servicesBytes)
					}
				}
			}
		}
		
		// If no extraction happened, use the original YAML
		if servicesYAML == "" {
			servicesYAML = yamlConfig
		}
		
		// Validate YAML has required structure
		if !strings.Contains(servicesYAML, "services:") {
			return ErrorResponse(
				"INVALID_YAML_STRUCTURE",
				"YAML must contain 'services:' section",
				"Ensure your YAML starts with 'services:' and defines service configurations",
			), nil
		}

		// Check if YAML contains preprocessing functions
		hasPreprocessing := strings.Contains(servicesYAML, "<@") && strings.Contains(servicesYAML, ">")
		
		// Ensure preprocessing is enabled if needed
		// The PreprocessYAML function adds #yamlPreprocessor=on if not present
		originalYAML := servicesYAML
		servicesYAML = PreprocessYAML(servicesYAML)
		
		// Debug: Check if preprocessing was added
		preprocessingAdded := !strings.Contains(originalYAML, "#yamlPreprocessor=on") && strings.Contains(servicesYAML, "#yamlPreprocessor=on")

		// Get current user to determine client ID
		user, err := client.GetCurrentUser(ctx)
		if err != nil {
			return HandleAPIError(err), nil
		}

		if len(user.ClientUserList) == 0 {
			return ErrorResponse(
				"NO_CLIENT",
				"No client associations found for user",
				"Contact Zerops support to resolve account issues",
			), nil
		}

		// Use the first client ID
		clientID := user.ClientUserList[0].ClientID

		// Import services
		err = client.ImportProjectServices(ctx, projectID, clientID, servicesYAML)
		if err != nil {
			// Handle specific errors
			if strings.Contains(err.Error(), "invalid service name") {
				return ErrorResponse(
					"INVALID_SERVICE_NAME",
					"Service name contains invalid characters",
					"Use only lowercase letters and numbers (no hyphens) for service names",
				), nil
			}
			if strings.Contains(err.Error(), "unknown type") {
				return ErrorResponse(
					"INVALID_SERVICE_TYPE",
					"Unknown service type specified",
					"Use valid service types like nodejs@20, postgresql@16, etc.",
				), nil
			}
			if strings.Contains(err.Error(), "404") {
				return ErrorResponseWithNext(
					"PROJECT_NOT_FOUND",
					fmt.Sprintf("Project with ID '%s' not found", projectID),
					"Check the project ID or use 'project_list' to find valid projects",
					"project_list",
				), nil
			}
			return HandleAPIError(err), nil
		}

		// Handle projectConfig if present
		var projectEnvErrors []string
		if projectConfig != nil {
			// Check for envSecrets in projectConfig
			if envSecrets, ok := projectConfig["envSecrets"].(map[string]interface{}); ok {
				for key, value := range envSecrets {
					// Convert value to string
					var content string
					switch v := value.(type) {
					case string:
						content = v
					default:
						content = fmt.Sprintf("%v", v)
					}
					
					// Create project environment variable
					_, err := client.CreateProjectEnv(ctx, projectID, key, content, true)
					if err != nil {
						projectEnvErrors = append(projectEnvErrors, fmt.Sprintf("%s: %v", key, err))
					}
				}
			}
		}
		
		successMsg := "Services imported successfully"
		if hasPreprocessing && preprocessingAdded {
			successMsg += " (preprocessing enabled for secret generation)"
		} else if hasPreprocessing && !preprocessingAdded {
			successMsg += " (preprocessing already enabled)"
		}
		
		if len(projectEnvErrors) > 0 {
			successMsg += fmt.Sprintf("\n\nWarning: Some project environment variables could not be created:\n- %s", strings.Join(projectEnvErrors, "\n- "))
		} else if projectConfig != nil && projectConfig["envSecrets"] != nil {
			successMsg += "\n\nProject environment secrets created successfully"
		}
		
		response := map[string]interface{}{
			"message":   successMsg,
			"projectId": projectID,
			"nextStep":  "Use 'service_list' to see imported services or 'vpn_connect' to prepare for deployment",
		}
		
		// Add debug info if preprocessing was involved
		if hasPreprocessing {
			previewLen := 50
			if len(yamlConfig) < previewLen {
				previewLen = len(yamlConfig)
			}
			response["preprocessingInfo"] = fmt.Sprintf("Directive added: %v, YAML starts with: %.50s", 
				preprocessingAdded, 
				strings.ReplaceAll(yamlConfig[:previewLen], "\n", "\\n"))
		}
		
		return SuccessResponse(response), nil
	})

	// Register project_delete tool
	projectDeleteTool := mcp.NewTool(
		"project_delete",
		mcp.WithDescription("Delete a Zerops project (WARNING: This will delete all services and data)"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID to delete"),
		),
		mcp.WithBoolean("confirm",
			mcp.Description("Set to true to confirm deletion"),
		),
	)

	s.AddTool(projectDeleteTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, err := request.RequireString("project_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROJECT_ID",
				"Project ID is required",
				"Provide a valid project ID from 'project_list'",
			), nil
		}

		confirm := request.GetBool("confirm", false)
		if !confirm {
			return ErrorResponse(
				"CONFIRMATION_REQUIRED",
				"Project deletion requires confirmation",
				"Set 'confirm' parameter to true to delete the project. WARNING: This action cannot be undone!",
			), nil
		}

		// Get project info first to show what's being deleted
		project, err := client.GetProject(ctx, projectID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				return ErrorResponse(
					"PROJECT_NOT_FOUND",
					fmt.Sprintf("Project with ID '%s' not found", projectID),
					"The project may have already been deleted",
				), nil
			}
			return HandleAPIError(err), nil
		}

		projectName := project.Name

		// Delete project
		_, err = client.DeleteProject(ctx, projectID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				return ErrorResponse(
					"PROJECT_NOT_FOUND",
					"Project not found",
					"The project may have already been deleted",
				), nil
			}
			return HandleAPIError(err), nil
		}

		return SuccessResponse(map[string]interface{}{
			"message":   fmt.Sprintf("Project '%s' has been deleted", projectName),
			"projectId": projectID,
			"warning":   "All services and data have been permanently removed",
			"nextStep":  "Use 'project_list' to see remaining projects or 'project_create' to create a new one",
		}), nil
	})
}