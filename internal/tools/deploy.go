package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
	"github.com/zeropsio/zerops-mcp-v3/internal/zcli"
)

// RegisterDeployTools registers all deployment tools
func RegisterDeployTools(s *server.MCPServer, client *api.Client, zcliWrapper *zcli.ZCLIWrapper) {
	// vpn_status
	vpnStatusTool := mcp.NewTool(
		"vpn_status",
		mcp.WithDescription("Check VPN connection status"),
		mcp.WithBoolean("detailed",
			mcp.Description("Show detailed status information (default: false)"),
		),
	)

	s.AddTool(vpnStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		detailed := request.GetBool("detailed", false)

		// Check if zcli is installed
		if !zcliWrapper.IsInstalled() {
			return ErrorResponse(
				"ZCLI_NOT_INSTALLED",
				"zcli is not installed",
				"Install zcli from https://docs.zerops.io/cli/installation/",
			), nil
		}

		// Get VPN status
		connected, projectID, message := zcliWrapper.VPNStatus(ctx)

		// Build response
		var response strings.Builder
		response.WriteString("VPN Status:\n\n")

		if connected {
			response.WriteString("✅ Connected\n")
			if projectID != "" {
				response.WriteString(fmt.Sprintf("Project ID: %s\n", projectID))
			}
		} else {
			response.WriteString("❌ Not connected\n")
		}

		if detailed {
			response.WriteString(fmt.Sprintf("\nDetails: %s\n", message))
			
			// Check zcli version
			version, err := zcliWrapper.Version(ctx)
			if err == nil {
				response.WriteString(fmt.Sprintf("\nzcli version: %s\n", version))
			}
		}

		response.WriteString("\nNext steps:\n")
		if connected {
			response.WriteString("- Use 'deploy_push' to deploy your application\n")
			response.WriteString("- Use 'vpn_disconnect' to disconnect VPN\n")
		} else {
			response.WriteString("- Use 'vpn_connect' with a project ID to connect\n")
			response.WriteString("- Use 'project_list' to see available projects\n")
		}

		return mcp.NewToolResultText(response.String()), nil
	})

	// vpn_connect
	vpnConnectTool := mcp.NewTool(
		"vpn_connect",
		mcp.WithDescription("Connect to Zerops VPN (requires sudo password)"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID to connect to"),
		),
	)

	s.AddTool(vpnConnectTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, err := request.RequireString("project_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROJECT_ID",
				"Project ID is required",
				"Provide a valid project ID from 'project_list' tool",
			), nil
		}

		// Check if zcli is installed
		if !zcliWrapper.IsInstalled() {
			return ErrorResponse(
				"ZCLI_NOT_INSTALLED",
				"zcli is not installed",
				"Install zcli from https://docs.zerops.io/cli/installation/",
			), nil
		}

		// Track if we're doing an automatic reconnection
		var previousProjectInfo string
		
		// Check if already connected
		if zcliWrapper.IsVPNConnected(ctx) {
			connectedID, _ := zcliWrapper.GetConnectedProjectID(ctx)
			if connectedID == projectID {
				// Already connected to the same project - this is success, not error
				return SuccessResponse(map[string]interface{}{
					"message":     fmt.Sprintf("Already connected to project %s", projectID),
					"status":      "connected",
					"project_id":  projectID,
					"info":        "VPN connection is active",
					"next_step":   "Use 'deploy_push' to deploy your application",
				}), nil
			}
			
			// Connected to different project - need to reconnect
			// Get the previous project info for logging
			previousProjectInfo = fmt.Sprintf("(was connected to %s)", connectedID)
			
			// Automatically disconnect and connect to new project
			if err := zcliWrapper.VPNDisconnect(ctx); err != nil {
				return ErrorResponse(
					"VPN_DISCONNECT_FAILED",
					fmt.Sprintf("Failed to disconnect from current project: %v", err),
					"Try manually running 'vpn_disconnect' first",
				), nil
			}
			
			// Small delay to ensure clean disconnect
			time.Sleep(1 * time.Second)
			
			// Note: We'll mention the automatic reconnection in the success message below
			_ = previousProjectInfo // Will be used in success message
		}

		// Verify project exists
		project, err := client.GetProject(ctx, projectID)
		if err != nil {
			return ErrorResponse(
				"PROJECT_NOT_FOUND",
				fmt.Sprintf("Project %s not found", projectID),
				"Verify the project ID with 'project_list' tool",
			), nil
		}

		// Connect VPN
		if err := zcliWrapper.VPNConnect(ctx, projectID); err != nil {
			// Check for common errors
			if strings.Contains(err.Error(), "sudo") || strings.Contains(err.Error(), "password") {
				return ErrorResponse(
					"SUDO_REQUIRED",
					"VPN connection requires sudo password",
					"Run this tool from a terminal with sudo access, or configure passwordless sudo for zcli",
				), nil
			}
			if strings.Contains(err.Error(), "already connected") {
				return ErrorResponseWithNext(
					"VPN_ALREADY_CONNECTED",
					"VPN is already connected",
					"Use 'vpn_disconnect' first to disconnect",
					"vpn_disconnect",
				), nil
			}
			return ErrorResponse(
				"VPN_CONNECTION_FAILED",
				fmt.Sprintf("Failed to connect VPN: %v", err),
				"Check your network connection and try again. Ensure zcli is properly configured",
			), nil
		}

		// Check if we did an automatic reconnection
		successData := map[string]interface{}{
			"status":      "connected",
			"project_id":  projectID,
			"project_name": project.Name,
			"next_step":   "Use 'deploy_push' to deploy your application",
		}
		
		// Customize message based on whether we auto-reconnected
		if previousProjectInfo != "" {
			successData["message"] = fmt.Sprintf("Automatically reconnected to project '%s' %s", project.Name, previousProjectInfo)
			successData["info"] = "Previous VPN connection was automatically closed"
		} else {
			successData["message"] = fmt.Sprintf("Connected to VPN for project '%s'", project.Name)
		}
		
		return SuccessResponse(successData), nil
	})

	// vpn_disconnect
	vpnDisconnectTool := mcp.NewTool(
		"vpn_disconnect",
		mcp.WithDescription("Disconnect from Zerops VPN (requires sudo password)"),
	)

	s.AddTool(vpnDisconnectTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if zcli is installed
		if !zcliWrapper.IsInstalled() {
			return ErrorResponse(
				"ZCLI_NOT_INSTALLED",
				"zcli is not installed",
				"Install zcli from https://docs.zerops.io/cli/installation/",
			), nil
		}

		// Check if connected
		if !zcliWrapper.IsVPNConnected(ctx) {
			return InfoResponse(
				"VPN Status",
				"VPN is not connected",
				"Use 'vpn_connect' to connect to a project",
			), nil
		}

		// Get connected project ID before disconnecting
		connectedID, _ := zcliWrapper.GetConnectedProjectID(ctx)

		// Disconnect VPN
		if err := zcliWrapper.VPNDisconnect(ctx); err != nil {
			if strings.Contains(err.Error(), "sudo") || strings.Contains(err.Error(), "password") {
				return ErrorResponse(
					"SUDO_REQUIRED",
					"VPN disconnection requires sudo password",
					"Run this tool from a terminal with sudo access",
				), nil
			}
			return ErrorResponse(
				"VPN_DISCONNECT_FAILED",
				fmt.Sprintf("Failed to disconnect VPN: %v", err),
				"Try running 'sudo zcli vpn down' manually",
			), nil
		}

		response := map[string]interface{}{
			"message":    "Disconnected from VPN",
			"status":     "disconnected",
			"next_step":  "Use 'vpn_connect' to connect to a project",
		}
		
		if connectedID != "" {
			response["previous_project"] = connectedID
		}

		return SuccessResponse(response), nil
	})

	// deploy_validate
	deployValidateTool := mcp.NewTool(
		"deploy_validate",
		mcp.WithDescription("Validate deployment prerequisites"),
		mcp.WithString("working_dir",
			mcp.Description("Working directory containing the code (default: current directory)"),
		),
		mcp.WithString("config_path",
			mcp.Description("Path to zerops.yml configuration file (default: zerops.yml in working directory)"),
		),
	)

	s.AddTool(deployValidateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workDir := request.GetString("working_dir", ".")
		configPath := request.GetString("config_path", "")

		// Default config path
		if configPath == "" {
			configPath = filepath.Join(workDir, "zerops.yml")
		}

		// Check if zcli is installed
		if !zcliWrapper.IsInstalled() {
			return ErrorResponse(
				"ZCLI_NOT_INSTALLED",
				"zcli is not installed",
				"Install zcli from https://docs.zerops.io/cli/installation/",
			), nil
		}

		var issues []string
		var warnings []string

		// Check working directory exists
		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Working directory '%s' does not exist", workDir))
		}

		// Check for git repository
		gitDir := filepath.Join(workDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			warnings = append(warnings, "Working directory is not a git repository (deployment works better with git)")
		}

		// Check zerops.yml exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Configuration file '%s' not found", configPath))
		} else {
			// Validate configuration
			if err := zcliWrapper.ValidateConfig(ctx, configPath); err != nil {
				issues = append(issues, fmt.Sprintf("Configuration validation failed: %v", err))
			}
		}

		// Check VPN connection
		if !zcliWrapper.IsVPNConnected(ctx) {
			issues = append(issues, "VPN is not connected")
		}

		// Build response
		var response strings.Builder
		response.WriteString("Deployment Validation:\n\n")

		if len(issues) == 0 {
			response.WriteString("✅ All checks passed!\n\n")
			response.WriteString("Ready to deploy:\n")
			response.WriteString(fmt.Sprintf("- Working directory: %s\n", workDir))
			response.WriteString(fmt.Sprintf("- Config file: %s\n", configPath))
			response.WriteString("- VPN: Connected\n")
			
			if len(warnings) > 0 {
				response.WriteString("\n⚠️ Warnings:\n")
				for _, warning := range warnings {
					response.WriteString(fmt.Sprintf("- %s\n", warning))
				}
			}
			
			response.WriteString("\nNext step: Use 'deploy_push' to deploy your application\n")
		} else {
			response.WriteString("❌ Validation failed:\n\n")
			for i, issue := range issues {
				response.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
			}
			
			response.WriteString("\nResolution steps:\n")
			for _, issue := range issues {
				if strings.Contains(issue, "VPN") {
					response.WriteString("- Use 'vpn_connect' to connect to VPN\n")
				}
				if strings.Contains(issue, "zerops.yml") {
					response.WriteString("- Create a zerops.yml configuration file\n")
					response.WriteString("- Use 'config_generate' to create from template\n")
				}
				if strings.Contains(issue, "directory") {
					response.WriteString("- Ensure you're in the correct directory\n")
					response.WriteString("- Provide the correct working_dir parameter\n")
				}
			}
		}

		if len(issues) > 0 {
			return mcp.NewToolResultError(response.String()), nil
		}
		return mcp.NewToolResultText(response.String()), nil
	})

	// deploy_push
	deployPushTool := mcp.NewTool(
		"deploy_push",
		mcp.WithDescription("Deploy application to Zerops. IMPORTANT: The service_name parameter is usually required unless your zerops.yml has only one service"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID to deploy to (get from project_list)"),
		),
		mcp.WithString("service_name",
			mcp.Description("Service name/hostname to deploy to (IMPORTANT: Usually required! Must match the 'hostname' field in your zerops.yml). If your zerops.yml has multiple services, this is REQUIRED. Use service_list to find the correct hostname"),
		),
		mcp.WithString("working_dir",
			mcp.Description("Working directory containing the code (default: current directory)"),
		),
		mcp.WithString("config_path",
			mcp.Description("Path to zerops.yml configuration file (default: zerops.yml in working directory)"),
		),
	)

	s.AddTool(deployPushTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID := request.GetString("project_id", "")
		serviceName := request.GetString("service_name", "")
		workDir := request.GetString("working_dir", ".")
		configPath := request.GetString("config_path", "")
		
		// Check prerequisites
		if !zcliWrapper.IsInstalled() {
			return ErrorResponse(
				"ZCLI_NOT_INSTALLED",
				"zcli is not installed",
				"Install zcli from https://docs.zerops.io/cli/installation/",
			), nil
		}

		if !zcliWrapper.IsVPNConnected(ctx) {
			return ErrorResponseWithNext(
				"VPN_NOT_CONNECTED",
				"VPN is not connected",
				"Connect to VPN first before deploying",
				"vpn_connect",
			), nil
		}

		// If project ID is not provided, we need to ask for it
		if projectID == "" {
			return ErrorResponse(
				"PROJECT_ID_REQUIRED",
				"Project ID is required for deployment",
				"Provide the project ID using the 'project_id' parameter. You can get it from 'project_list' tool",
			), nil
		}

		// Execute deployment
		output, err := zcliWrapper.Push(ctx, projectID, serviceName, workDir, configPath)
		if err != nil {
			// Parse common errors
			if strings.Contains(output, "Please, select a service") || strings.Contains(output, "Interactive selection can be used only in terminal mode") {
				// Get the services from the project to help the user
				services, listErr := client.ListServices(ctx, projectID)
				var serviceHelp strings.Builder
				serviceHelp.WriteString("Service selection is required for deployment.\n\n")
				
				if listErr == nil && len(services) > 0 {
					serviceHelp.WriteString("Available services in your project:\n")
					for _, svc := range services {
						serviceHelp.WriteString(fmt.Sprintf("- %s (type: %s@%s)\n", svc.Name, svc.ServiceStackTypeInfo.ServiceStackTypeName, svc.ServiceStackTypeInfo.ServiceStackTypeVersionName))
					}
					serviceHelp.WriteString("\nUse the service_name parameter with one of the above service names (hostnames).\n")
					serviceHelp.WriteString("Example: deploy_push(project_id=\"...\", service_name=\"app\", ...)")
				} else {
					serviceHelp.WriteString("Use 'service_list' tool to see available services in your project.\n")
					serviceHelp.WriteString("Then provide the service_name parameter matching a service hostname from your zerops.yml.")
				}
				
				return ErrorResponse(
					"SERVICE_SELECTION_REQUIRED",
					"Service name must be specified for deployment",
					serviceHelp.String(),
				), nil
			}
			if strings.Contains(err.Error(), "exit status 128") || strings.Contains(output, "exit status 128") {
				return ErrorResponse(
					"GIT_ERROR",
					"Git repository error (exit status 128)",
					"Ensure your git repository has at least one commit:\n1. git add .\n2. git commit -m \"Initial commit\"\n\nNote: zcli requires a git repository with commits, not just 'git init'",
				), nil
			}
			if strings.Contains(output, "websocket: bad handshake") {
				// This might not be a real failure - check if deployment actually succeeded
				return ErrorResponse(
					"LOG_STREAMING_ERROR", 
					"Log streaming error during deployment",
					"This is usually a temporary issue with log streaming. The deployment may still succeed.\nCheck deployment status with 'deploy_status' tool or check service logs",
				), nil
			}
			if strings.Contains(output, "projectWillBeDeleted") || strings.Contains(output, "No action allowed") {
				return ErrorResponse(
					"PROJECT_NOT_READY",
					"Project not ready for deployment",
					"The project might be initializing or marked for deletion. Wait a moment and try again.\nEnsure services are fully initialized after import (wait 30+ seconds)",
				), nil
			}
			if strings.Contains(err.Error(), "zerops.yml") || strings.Contains(err.Error(), "config") {
				return ErrorResponse(
					"CONFIG_ERROR",
					"Configuration error",
					"Check your zerops.yml file for syntax errors or missing required fields",
				), nil
			}
			if strings.Contains(err.Error(), "VPN") || strings.Contains(err.Error(), "connection") {
				return ErrorResponseWithNext(
					"CONNECTION_ERROR",
					"Connection error during deployment",
					"Ensure VPN is connected and stable",
					"vpn_status",
				), nil
			}
			if strings.Contains(err.Error(), "service") {
				return ErrorResponse(
					"SERVICE_ERROR",
					"Service configuration error",
					"Ensure the service name in zerops.yml matches an existing service in your project",
				), nil
			}

			return ErrorResponse(
				"DEPLOY_FAILED",
				fmt.Sprintf("Deployment failed: %v", err),
				fmt.Sprintf("Check the error details:\n%s\n\nUse 'deploy_troubleshoot' for help", output),
			), nil
		}

		// Parse output for success indicators
		var response strings.Builder
		response.WriteString("Deployment initiated successfully!\n\n")
		response.WriteString("Deployment progress:\n")
		
		// Include relevant output
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "Using config file") {
				response.WriteString(fmt.Sprintf("- %s\n", line))
			}
		}

		response.WriteString("\nNext steps:\n")
		response.WriteString("- Use 'deploy_status' to check deployment progress\n")
		response.WriteString("- Use 'deploy_logs' to view build logs\n")
		response.WriteString("- Use 'service_logs' to view runtime logs\n")

		return mcp.NewToolResultText(response.String()), nil
	})

	// deploy_status
	deployStatusTool := mcp.NewTool(
		"deploy_status",
		mcp.WithDescription("Get deployment status for a service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to check deployment status for"),
		),
		mcp.WithBoolean("detailed",
			mcp.Description("Show detailed deployment information (default: false)"),
		),
	)

	s.AddTool(deployStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		detailed := request.GetBool("detailed", false)

		// Get service info to check current deployment
		service, err := client.GetService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Build response
		var response strings.Builder
		response.WriteString(fmt.Sprintf("Deployment Status for %s:\n\n", service.Name))
		
		// Service status indicates deployment state
		response.WriteString(fmt.Sprintf("Service Status: %s\n", service.Status))
		
		switch service.Status {
		case "RUNNING":
			response.WriteString("✅ Service is running\n")
		case "BUILDING", "DEPLOYING":
			response.WriteString("🔄 Deployment in progress\n")
		case "FAILED":
			response.WriteString("❌ Deployment failed\n")
		case "STOPPED":
			response.WriteString("⏸️ Service is stopped\n")
		default:
			response.WriteString(fmt.Sprintf("ℹ️ Status: %s\n", service.Status))
		}

		if detailed {
			response.WriteString(fmt.Sprintf("\nService Details:\n"))
			response.WriteString(fmt.Sprintf("- Type: %s\n", service.ServiceStackTypeInfo.ServiceStackTypeVersionName))
			response.WriteString(fmt.Sprintf("- Mode: %s\n", service.Mode))
			response.WriteString(fmt.Sprintf("- Created: %s\n", service.Created.Format("2006-01-02 15:04:05")))
			response.WriteString(fmt.Sprintf("- Last Update: %s\n", service.LastUpdate.Format("2006-01-02 15:04:05")))
		}

		response.WriteString("\nNext steps:\n")
		if service.Status == "BUILDING" || service.Status == "DEPLOYING" {
			response.WriteString("- Use 'deploy_logs' to view build progress\n")
			response.WriteString("- Wait for deployment to complete\n")
		} else if service.Status == "FAILED" {
			response.WriteString("- Use 'deploy_logs' to see what went wrong\n")
			response.WriteString("- Use 'deploy_troubleshoot' for help\n")
		} else if service.Status == "RUNNING" {
			response.WriteString("- Use 'service_logs' to view runtime logs\n")
			response.WriteString("- Use 'service_info' for detailed information\n")
		}

		return mcp.NewToolResultText(response.String()), nil
	})

	// deploy_logs
	deployLogsTool := mcp.NewTool(
		"deploy_logs",
		mcp.WithDescription("Get deployment/build logs for a service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to get deployment logs for"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of log lines to retrieve (default: 100, max: 1000)"),
		),
	)

	s.AddTool(deployLogsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		limit := request.GetInt("limit", 100)
		if limit > 1000 {
			limit = 1000
		}

		// Get build/deployment logs
		// Using "prepare" container which typically contains build logs
		logs, err := client.GetServiceLogs(ctx, serviceID, "prepare", limit, "")
		if err != nil {
			// Try runtime logs as fallback
			logs, err = client.GetServiceLogs(ctx, serviceID, "", limit, "")
			if err != nil {
				return ErrorResponse(
					"LOGS_NOT_AVAILABLE",
					"Deployment logs not available",
					"Logs might not be available yet for new deployments. Try again in a few moments",
				), nil
			}
		}

		// Build response
		var response strings.Builder
		response.WriteString(fmt.Sprintf("Deployment Logs for service %s:\n\n", serviceID))

		if len(logs) == 0 {
			response.WriteString("No deployment logs found.\n")
			response.WriteString("\nPossible reasons:\n")
			response.WriteString("- Service has not been deployed yet\n")
			response.WriteString("- Logs have been rotated\n")
			response.WriteString("- Service is still initializing\n")
		} else {
			for _, log := range logs {
				response.WriteString(log)
				response.WriteString("\n")
			}
			response.WriteString(fmt.Sprintf("\nShowing last %d lines\n", len(logs)))
		}

		response.WriteString("\nNext steps:\n")
		response.WriteString("- Use 'service_logs' for runtime logs\n")
		response.WriteString("- Use 'deploy_status' to check deployment state\n")
		response.WriteString("- Use 'deploy_troubleshoot' if deployment failed\n")

		return mcp.NewToolResultText(response.String()), nil
	})

	// deploy_troubleshoot
	deployTroubleshootTool := mcp.NewTool(
		"deploy_troubleshoot",
		mcp.WithDescription("Analyze and troubleshoot deployment issues"),
		mcp.WithString("service_id",
			mcp.Description("Service ID to troubleshoot (optional)"),
		),
		mcp.WithString("error_message",
			mcp.Description("Error message to analyze (optional)"),
		),
	)

	s.AddTool(deployTroubleshootTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID := request.GetString("service_id", "")
		errorMsg := request.GetString("error_message", "")

		var response strings.Builder
		response.WriteString("Deployment Troubleshooting Guide:\n\n")

		// Check common issues
		issues := []struct {
			title       string
			check       func() (bool, string)
			resolution  string
		}{
			{
				title: "VPN Connection",
				check: func() (bool, string) {
					if !zcliWrapper.IsVPNConnected(ctx) {
						return false, "VPN is not connected"
					}
					return true, "VPN is connected"
				},
				resolution: "Use 'vpn_connect' to establish VPN connection",
			},
			{
				title: "zcli Installation",
				check: func() (bool, string) {
					if !zcliWrapper.IsInstalled() {
						return false, "zcli is not installed"
					}
					version, _ := zcliWrapper.Version(ctx)
					return true, fmt.Sprintf("zcli version: %s", version)
				},
				resolution: "Install zcli from https://docs.zerops.io/cli/installation/",
			},
		}

		// Run checks
		response.WriteString("System Checks:\n")
		for _, issue := range issues {
			ok, status := issue.check()
			if ok {
				response.WriteString(fmt.Sprintf("✅ %s: %s\n", issue.title, status))
			} else {
				response.WriteString(fmt.Sprintf("❌ %s: %s\n", issue.title, status))
				response.WriteString(fmt.Sprintf("   Resolution: %s\n", issue.resolution))
			}
		}

		// Check specific service if provided
		if serviceID != "" {
			response.WriteString(fmt.Sprintf("\nService Analysis (%s):\n", serviceID))
			
			service, err := client.GetService(ctx, serviceID)
			if err != nil {
				response.WriteString("❌ Could not retrieve service information\n")
			} else {
				response.WriteString(fmt.Sprintf("- Status: %s\n", service.Status))
				response.WriteString(fmt.Sprintf("- Type: %s\n", service.ServiceStackTypeInfo.ServiceStackTypeVersionName))
				
				if service.Status == "FAILED" {
					response.WriteString("\n⚠️ Service is in FAILED state\n")
					response.WriteString("Recommended actions:\n")
					response.WriteString("1. Check deployment logs: deploy_logs\n")
					response.WriteString("2. Verify zerops.yml configuration\n")
					response.WriteString("3. Ensure service name matches in config\n")
				}
			}
		}

		// Analyze error message if provided
		if errorMsg != "" {
			response.WriteString(fmt.Sprintf("\nError Analysis:\n"))
			response.WriteString(fmt.Sprintf("Error: %s\n\n", errorMsg))
			
			// Common error patterns
			if strings.Contains(strings.ToLower(errorMsg), "config") || strings.Contains(errorMsg, "zerops.yml") {
				response.WriteString("📋 Configuration Issue Detected:\n")
				response.WriteString("- Check zerops.yml syntax (YAML format)\n")
				response.WriteString("- Ensure service names match existing services\n")
				response.WriteString("- Verify all required fields are present\n")
				response.WriteString("- Use 'config_validate' to check configuration\n")
			}
			
			if strings.Contains(strings.ToLower(errorMsg), "connection") || strings.Contains(strings.ToLower(errorMsg), "vpn") {
				response.WriteString("🌐 Network Issue Detected:\n")
				response.WriteString("- Check VPN connection: vpn_status\n")
				response.WriteString("- Reconnect if needed: vpn_connect\n")
				response.WriteString("- Ensure stable internet connection\n")
			}
			
			if strings.Contains(strings.ToLower(errorMsg), "permission") || strings.Contains(strings.ToLower(errorMsg), "denied") {
				response.WriteString("🔒 Permission Issue Detected:\n")
				response.WriteString("- Ensure you have access to the project\n")
				response.WriteString("- Check API key permissions\n")
				response.WriteString("- Verify project ownership\n")
			}
		}

		response.WriteString("\nGeneral Troubleshooting Steps:\n")
		response.WriteString("1. Validate prerequisites: deploy_validate\n")
		response.WriteString("2. Check service status: service_info\n")
		response.WriteString("3. Review deployment logs: deploy_logs\n")
		response.WriteString("4. Verify configuration: config_validate\n")
		response.WriteString("5. Test with minimal config first\n")
		
		response.WriteString("\nCommon Issues:\n")
		response.WriteString("- Service name mismatch between zerops.yml and project\n")
		response.WriteString("- Missing or invalid zerops.yml configuration\n")
		response.WriteString("- VPN connection dropped during deployment\n")
		response.WriteString("- Incorrect working directory\n")
		response.WriteString("- Build command failures\n")

		return mcp.NewToolResultText(response.String()), nil
	})
}