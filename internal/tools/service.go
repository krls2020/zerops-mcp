package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
)

// RegisterServiceTools registers all service management tools
func RegisterServiceTools(s *server.MCPServer, client *api.Client) {
	// service_list
	serviceListTool := mcp.NewTool(
		"service_list",
		mcp.WithDescription("List all services in a project"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID to list services from"),
		),
		mcp.WithBoolean("include_system",
			mcp.Description("Include system services (default: false)"),
		),
	)

	s.AddTool(serviceListTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, err := request.RequireString("project_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROJECT_ID",
				"Project ID is required",
				"Provide a valid project ID from 'project_list' tool",
			), nil
		}

		includeSystem := request.GetBool("include_system", false)

		// List services
		services, err := client.ListServices(ctx, projectID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Build response
		var response strings.Builder
		response.WriteString(fmt.Sprintf("Services in project %s:\n\n", projectID))

		filteredCount := 0
		for _, svc := range services {
			// Skip system services if not requested
			if svc.ServiceStackTypeInfo.ServiceStackTypeCategory == "system" && !includeSystem {
				continue
			}
			filteredCount++

			response.WriteString(fmt.Sprintf("Service: %s\n", svc.Name))
			response.WriteString(fmt.Sprintf("  ID: %s\n", svc.ID))
			response.WriteString(fmt.Sprintf("  Type: %s\n", svc.ServiceStackTypeInfo.ServiceStackTypeVersionName))
			response.WriteString(fmt.Sprintf("  Category: %s\n", svc.ServiceStackTypeInfo.ServiceStackTypeCategory))
			response.WriteString(fmt.Sprintf("  Status: %s\n", svc.Status))
			response.WriteString(fmt.Sprintf("  Mode: %s\n", svc.Mode))
			response.WriteString(fmt.Sprintf("  Containers: %d-%d\n", svc.MinContainers, svc.MaxContainers))
			if len(svc.Ports) > 0 {
				response.WriteString("  Ports:\n")
				for _, port := range svc.Ports {
					portInfo := fmt.Sprintf("    - %d/%s", port.Port, port.Protocol)
					if port.HTTPRouting {
						portInfo += " [HTTP]"
					}
					response.WriteString(portInfo + "\n")
				}
			}
			// Show subdomain if available
			if svc.SubdomainAccess && svc.ZeropsSubdomainHost != nil && *svc.ZeropsSubdomainHost != "" {
				response.WriteString(fmt.Sprintf("  Subdomain: https://%s\n", *svc.ZeropsSubdomainHost))
			}
			response.WriteString("\n")
		}

		if filteredCount == 0 {
			if includeSystem {
				response.WriteString("No services found in this project.\n")
			} else {
				response.WriteString("No user services found. Use include_system=true to see system services.\n")
			}
		} else {
			response.WriteString(fmt.Sprintf("Total: %d services\n", filteredCount))
		}

		response.WriteString("\nNext steps:\n")
		response.WriteString("- Use 'service_info' to get detailed information about a service\n")
		response.WriteString("- Use 'service_logs' to view service logs\n")
		response.WriteString("- Use 'project_import' to add more services\n")

		return mcp.NewToolResultText(response.String()), nil
	})

	// service_info
	serviceInfoTool := mcp.NewTool(
		"service_info",
		mcp.WithDescription("Get detailed information about a service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to get information for"),
		),
		mcp.WithBoolean("include_env",
			mcp.Description("Include environment variables (default: true)"),
		),
	)

	s.AddTool(serviceInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		includeEnv := request.GetBool("include_env", true)

		// Get service details
		service, err := client.GetService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Build response
		var response strings.Builder
		response.WriteString(fmt.Sprintf("Service Details: %s\n\n", service.Name))
		response.WriteString(fmt.Sprintf("ID: %s\n", service.ID))
		response.WriteString(fmt.Sprintf("Project ID: %s\n", service.ProjectID))
		response.WriteString(fmt.Sprintf("Type: %s\n", service.ServiceStackTypeInfo.ServiceStackTypeVersionName))
		response.WriteString(fmt.Sprintf("Category: %s\n", service.ServiceStackTypeInfo.ServiceStackTypeCategory))
		response.WriteString(fmt.Sprintf("Status: %s\n", service.Status))
		response.WriteString(fmt.Sprintf("Mode: %s\n", service.Mode))
		response.WriteString(fmt.Sprintf("Created: %s\n", service.Created.Format(time.RFC3339)))
		response.WriteString(fmt.Sprintf("Last Update: %s\n", service.LastUpdate.Format(time.RFC3339)))
		response.WriteString(fmt.Sprintf("\nContainers: %d-%d\n", service.MinContainers, service.MaxContainers))

		// Resources
		if service.VerticalScaling != nil {
			response.WriteString("\nResources:\n")
			response.WriteString(fmt.Sprintf("  CPU: %d-%d cores\n", service.VerticalScaling.MinCPU, service.VerticalScaling.MaxCPU))
			response.WriteString(fmt.Sprintf("  RAM: %d-%d GB\n", service.VerticalScaling.MinRAM, service.VerticalScaling.MaxRAM))
			response.WriteString(fmt.Sprintf("  Disk: %d-%d GB\n", service.VerticalScaling.MinDisk, service.VerticalScaling.MaxDisk))
			response.WriteString(fmt.Sprintf("  CPU Mode: %s\n", service.VerticalScaling.CPUMode))
		}

		// Auto-scaling
		if service.AutoScaling != nil && service.AutoScaling.Enabled {
			response.WriteString("\nAuto-scaling: Enabled\n")
		}

		// Ports
		if len(service.Ports) > 0 {
			response.WriteString("\nPorts:\n")
			for _, port := range service.Ports {
				public := "private"
				if port.Public {
					public = "public"
				}
				portInfo := fmt.Sprintf("  - %d/%s (%s)", port.Port, port.Protocol, public)
				if port.HTTPRouting {
					portInfo += " [HTTP]"
				}
				response.WriteString(portInfo + "\n")
			}
		}

		// Subdomain Access
		if service.SubdomainAccess {
			response.WriteString("\nSubdomain Access: Enabled\n")
			if service.ZeropsSubdomainHost != nil && *service.ZeropsSubdomainHost != "" {
				response.WriteString(fmt.Sprintf("Subdomain URL: https://%s\n", *service.ZeropsSubdomainHost))
			}
		} else {
			// Check if service could have subdomain access
			hasHTTPPort := false
			for _, port := range service.Ports {
				if port.HTTPRouting || port.Scheme == "http" || port.Scheme == "https" {
					hasHTTPPort = true
					break
				}
			}
			if hasHTTPPort && service.ServiceStackTypeInfo.ServiceStackTypeCategory == "USER" {
				response.WriteString("\nSubdomain Access: Available (use 'subdomain_enable' to activate)\n")
			}
		}

		// Environment Variables
		if includeEnv && len(service.EnvVariables) > 0 {
			response.WriteString("\nEnvironment Variables:\n")
			for key, value := range service.EnvVariables {
				// Mask sensitive values
				displayValue := fmt.Sprintf("%v", value)
				if isSensitiveKey(key) {
					displayValue = "***masked***"
				}
				response.WriteString(fmt.Sprintf("  %s: %s\n", key, displayValue))
			}
		}

		// Connection info
		response.WriteString(fmt.Sprintf("\nInternal Hostname: %s\n", service.Name))
		response.WriteString(fmt.Sprintf("Connection URL: %s.prj\n", service.Name))

		response.WriteString("\nNext steps:\n")
		response.WriteString("- Use 'service_logs' to view logs\n")
		if service.Status == "RUNNING" {
			response.WriteString("- Use 'service_stop' to stop the service\n")
		} else {
			response.WriteString("- Use 'service_start' to start the service\n")
		}
		response.WriteString("- Use 'deploy_push' to deploy code\n")

		return mcp.NewToolResultText(response.String()), nil
	})

	// service_logs
	serviceLogsTool := mcp.NewTool(
		"service_logs",
		mcp.WithDescription("Get logs from a service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to get logs from"),
		),
		mcp.WithString("container",
			mcp.Description("Container name (runtime, prepare, init) - defaults to all"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of log lines to retrieve (default: 100, max: 1000)"),
		),
		mcp.WithString("since",
			mcp.Description("Show logs since timestamp (RFC3339 format) or duration (e.g., '5m', '1h')"),
		),
	)

	s.AddTool(serviceLogsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		container := request.GetString("container", "")
		limit := request.GetInt("limit", 100)
		if limit > 1000 {
			limit = 1000
		}
		since := request.GetString("since", "")

		// Get logs
		logs, err := client.GetServiceLogs(ctx, serviceID, container, limit, since)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Build response
		var response strings.Builder
		response.WriteString(fmt.Sprintf("Logs for service %s", serviceID))
		if container != "" {
			response.WriteString(fmt.Sprintf(" (container: %s)", container))
		}
		response.WriteString("\n\n")

		if len(logs) == 0 {
			response.WriteString("No logs found.\n")
			response.WriteString("\nPossible reasons:\n")
			response.WriteString("- Service has not started yet\n")
			response.WriteString("- Service crashed before producing logs\n")
			response.WriteString("- Logs may have been rotated\n")
			response.WriteString("\nNext steps:\n")
			response.WriteString("- Use 'service_info' to check service status\n")
			response.WriteString("- Use 'deploy_status' to check deployment status\n")
		} else {
			for _, log := range logs {
				response.WriteString(log)
				response.WriteString("\n")
			}
			
			// Analyze logs for common runtime errors
			errorAnalysis := AnalyzeRuntimeError(logs)
			if errorAnalysis != "" {
				response.WriteString(errorAnalysis)
			}
		}

		response.WriteString(fmt.Sprintf("\nShowing last %d lines\n", len(logs)))
		response.WriteString("\nTips:\n")
		response.WriteString("- Use 'container' parameter to filter by container type (runtime, prepare, init)\n")
		response.WriteString("- Use 'since' parameter to get recent logs (e.g., since='5m')\n")
		response.WriteString("- Use 'limit' parameter to get more logs (max 1000)\n")
		
		// Add specific tips based on service type
		if service, err := client.GetService(ctx, serviceID); err == nil {
			serviceType := service.ServiceStackTypeInfo.ServiceStackTypeVersionName
			if strings.Contains(serviceType, "python") || strings.Contains(serviceType, "nodejs") {
				response.WriteString("\nRuntime Tips:\n")
				response.WriteString("- Ensure dependencies are installed with 'prepareCommands' in zerops.yml\n")
				response.WriteString("- Check that your start command matches your application entry point\n")
			}
		}

		return mcp.NewToolResultText(response.String()), nil
	})

	// service_start
	serviceStartTool := mcp.NewTool(
		"service_start",
		mcp.WithDescription("Start a stopped service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to start"),
		),
	)

	s.AddTool(serviceStartTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		// Start the service
		process, err := client.StartService(ctx, serviceID)
		if err != nil {
			// Check if service is already running
			if strings.Contains(err.Error(), "already running") || strings.Contains(err.Error(), "invalidServiceStackStatus") {
				return ErrorResponse(
					"SERVICE_ALREADY_RUNNING",
					"Service is already running",
					"The service is already in a running state",
				), nil
			}
			return HandleAPIError(err), nil
		}

		return SuccessResponse(map[string]interface{}{
			"message":    fmt.Sprintf("Service start initiated"),
			"process_id": process.ID,
			"status":     process.Status,
			"next_step":  "Use 'service_info' to check status or 'service_logs' to monitor startup",
		}), nil
	})

	// service_stop
	serviceStopTool := mcp.NewTool(
		"service_stop",
		mcp.WithDescription("Stop a running service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to stop"),
		),
	)

	s.AddTool(serviceStopTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		// Stop the service
		process, err := client.StopService(ctx, serviceID)
		if err != nil {
			// Check if service is already stopped
			if strings.Contains(err.Error(), "already stopped") || strings.Contains(err.Error(), "invalidServiceStackStatus") {
				return ErrorResponse(
					"SERVICE_ALREADY_STOPPED",
					"Service is already stopped",
					"The service is already in a stopped state",
				), nil
			}
			return HandleAPIError(err), nil
		}

		return SuccessResponse(map[string]interface{}{
			"message":    fmt.Sprintf("Service stop initiated"),
			"process_id": process.ID,
			"status":     process.Status,
			"next_step":  "Use 'service_info' to check status",
		}), nil
	})

	// service_delete
	serviceDeleteTool := mcp.NewTool(
		"service_delete",
		mcp.WithDescription("Delete a service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to delete"),
		),
		mcp.WithBoolean("confirm",
			mcp.Required(),
			mcp.Description("Confirm deletion (must be true)"),
		),
	)

	s.AddTool(serviceDeleteTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		confirm := request.GetBool("confirm", false)
		if !confirm {
			return ErrorResponse(
				"CONFIRMATION_REQUIRED",
				"Service deletion requires confirmation",
				"Set confirm=true to delete the service",
			), nil
		}

		// Delete the service
		err = client.DeleteService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		return SuccessResponse(map[string]interface{}{
			"message":    "Service deletion initiated",
			"service_id": serviceID,
			"status":     "DELETING",
			"next_step":  "Use 'service_list' to verify deletion",
		}), nil
	})
}

// isSensitiveKey checks if an environment variable key contains sensitive data
func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	sensitivePatterns := []string{
		"password", "passwd", "pwd",
		"secret", "key", "token",
		"api_key", "apikey",
		"auth", "credential",
		"private", "cert",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(key, pattern) {
			return true
		}
	}
	return false
}