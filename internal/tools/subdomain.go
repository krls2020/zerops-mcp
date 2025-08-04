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

// RegisterSubdomainTools registers all subdomain-related tools
func RegisterSubdomainTools(s *server.MCPServer, client *api.Client) {
	// Create subdomain_enable tool
	subdomainEnableTool := mcp.NewTool(
		"subdomain_enable",
		mcp.WithDescription("Enable subdomain access for a service. Requires service to have HTTP/HTTPS ports configured. By default waits up to 30 seconds for completion. Set wait=false to return immediately."),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to enable subdomain for"),
		),
	)

	s.AddTool(subdomainEnableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		wait := request.GetBool("wait", true)

		// Get service details first to check current state
		service, err := client.GetService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Check if subdomain is already enabled
		if service.SubdomainAccess {
			// Get project details for zeropsSubdomainHost
			project, err := client.GetProject(ctx, service.ProjectID)
			if err != nil {
				return HandleAPIError(err), nil
			}

			var url string
			if project.ZeropsSubdomainHost != nil && *project.ZeropsSubdomainHost != "" {
				// Find the HTTP port
				for _, port := range service.Ports {
					if port.HTTPRouting || port.Scheme == "http" || port.Scheme == "https" {
						url = GenerateSubdomainURL(service.Name, *project.ZeropsSubdomainHost, port.Port)
						break
					}
				}
			}
			return SuccessResponse(map[string]interface{}{
				"message":         "Subdomain access is already enabled",
				"service_id":      serviceID,
				"service_name":    service.Name,
				"subdomain_url":   url,
				"already_enabled": true,
			}), nil
		}

		// Check if service has HTTP ports
		hasHTTPPort := false
		for _, port := range service.Ports {
			if port.HTTPRouting || port.Scheme == "http" || port.Scheme == "https" {
				hasHTTPPort = true
				break
			}
		}

		if !hasHTTPPort {
			return ErrorResponse(
				"NO_HTTP_PORTS",
				fmt.Sprintf("Service '%s' does not have HTTP/HTTPS ports configured", service.Name),
				"Only services with HTTP/HTTPS ports can have subdomain access. Add ports with httpSupport: true in your service configuration.",
			), nil
		}

		// Check service status
		if service.Status != "ACTIVE" && service.Status != "RUNNING" {
			return ErrorResponseWithNext(
				"SERVICE_NOT_ACTIVE",
				fmt.Sprintf("Service '%s' is in %s state", service.Name, service.Status),
				"Service must be in ACTIVE or RUNNING state. Deploy your application or start the service first.",
				"service_start",
			), nil
		}

		// Enable subdomain access
		process, err := client.EnableSubdomainAccess(ctx, serviceID)
		if err != nil {
			if strings.Contains(err.Error(), "serviceStackIsNotHttp") {
				return ErrorResponse(
					"NOT_HTTP_SERVICE",
					"Service is not configured as HTTP/HTTPS service",
					"Ensure your service has ports configured with 'httpSupport: true' in the import YAML",
				), nil
			}
			return HandleAPIError(err), nil
		}

		// If not waiting, return immediately
		if !wait {
			return SuccessResponse(map[string]interface{}{
				"message":     fmt.Sprintf("Enabling subdomain access for service '%s'", service.Name),
				"process_id":  process.ID,
				"service_id":  serviceID,
				"status":      "PENDING",
				"next_step":   "Use 'process_status' to check progress",
			}), nil
		}

		// Wait for process to complete (subdomain operations are typically fast)
		completedProcess, err := client.WaitForProcess(ctx, process.ID, 30*time.Second)
		if err != nil {
			// If timeout, try to get current process status
			currentProcess, statusErr := client.GetProcess(ctx, process.ID)
			if statusErr == nil && currentProcess != nil {
				// Check if process is still running
				if currentProcess.Status == "PENDING" || currentProcess.Status == "RUNNING" {
					return SuccessResponse(map[string]interface{}{
						"message":      fmt.Sprintf("Subdomain enable operation is still in progress for service '%s'", service.Name),
						"process_id":   process.ID,
						"status":       currentProcess.Status,
						"elapsed_time": fmt.Sprintf("%.0f seconds", time.Since(currentProcess.Created).Seconds()),
						"service_id":   serviceID,
						"next_step":    "Operation is taking longer than expected. Use 'process_status' to check progress or 'subdomain_status' to verify if enabled",
						"tip":          "For long-running operations, consider using wait=false parameter",
					}), nil
				}
				// Process finished with some status
				if currentProcess.Status == "SUCCESS" || currentProcess.Status == "FINISHED" {
					// Even though we timed out waiting, the process succeeded
					// Continue to get service details below
					completedProcess = currentProcess
				} else {
					return ErrorResponse(
						"SUBDOMAIN_ENABLE_FAILED",
						fmt.Sprintf("Subdomain enable process failed with status: %s", currentProcess.Status),
						"Check service logs for more details or try again",
					), nil
				}
			} else {
				// Couldn't get process status, return original error
				return ErrorResponse(
					"PROCESS_WAIT_TIMEOUT",
					fmt.Sprintf("Timed out waiting for subdomain enable process after 30 seconds"),
					fmt.Sprintf("Use 'process_status' with process_id=%s to check current status, or 'subdomain_status' to verify if subdomain was enabled", process.ID),
				), nil
			}
		}

		// Check if process succeeded (if we didn't already handle it above)
		if completedProcess != nil && completedProcess.Status != "SUCCESS" && completedProcess.Status != "FINISHED" {
			return ErrorResponse(
				"SUBDOMAIN_ENABLE_FAILED",
				fmt.Sprintf("Failed to enable subdomain access: process status %s", completedProcess.Status),
				"Check service configuration and ensure it has HTTP ports configured",
			), nil
		}

		// Get updated service details
		updatedService, err := client.GetService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Get project details for zeropsSubdomainHost
		project, err := client.GetProject(ctx, updatedService.ProjectID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		var subdomainURL string
		if updatedService.SubdomainAccess && project.ZeropsSubdomainHost != nil && *project.ZeropsSubdomainHost != "" {
			// Find the HTTP port
			for _, port := range updatedService.Ports {
				if port.HTTPRouting || port.Scheme == "http" || port.Scheme == "https" {
					subdomainURL = GenerateSubdomainURL(updatedService.Name, *project.ZeropsSubdomainHost, port.Port)
					break
				}
			}
		}

		return SuccessResponse(map[string]interface{}{
			"message":       fmt.Sprintf("Successfully enabled subdomain access for service '%s'", service.Name),
			"service_id":    serviceID,
			"service_name":  service.Name,
			"subdomain_url": subdomainURL,
			"process_id":    process.ID,
			"next_step":     "Your service is now accessible via the subdomain URL",
		}), nil
	})

	// Create subdomain_disable tool
	subdomainDisableTool := mcp.NewTool(
		"subdomain_disable",
		mcp.WithDescription("Disable subdomain access for a service. By default waits up to 20 seconds for completion. Set wait=false to return immediately."),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to disable subdomain for"),
		),
	)

	s.AddTool(subdomainDisableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		wait := request.GetBool("wait", true)

		// Get service details first
		service, err := client.GetService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Check if subdomain is already disabled
		if !service.SubdomainAccess {
			return SuccessResponse(map[string]interface{}{
				"message":          "Subdomain access is already disabled",
				"service_id":       serviceID,
				"service_name":     service.Name,
				"already_disabled": true,
			}), nil
		}

		// Disable subdomain access
		process, err := client.DisableSubdomainAccess(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// If not waiting, return immediately
		if !wait {
			return SuccessResponse(map[string]interface{}{
				"message":    fmt.Sprintf("Disabling subdomain access for service '%s'", service.Name),
				"process_id": process.ID,
				"service_id": serviceID,
				"status":     "PENDING",
				"next_step":  "Use 'process_status' to check progress",
			}), nil
		}

		// Wait for process to complete (subdomain operations are typically fast)
		completedProcess, err := client.WaitForProcess(ctx, process.ID, 20*time.Second)
		if err != nil {
			// If timeout, try to get current process status
			currentProcess, statusErr := client.GetProcess(ctx, process.ID)
			if statusErr == nil && currentProcess != nil {
				// Check if process is still running
				if currentProcess.Status == "PENDING" || currentProcess.Status == "RUNNING" {
					return SuccessResponse(map[string]interface{}{
						"message":      fmt.Sprintf("Subdomain disable operation is still in progress for service '%s'", service.Name),
						"process_id":   process.ID,
						"status":       currentProcess.Status,
						"elapsed_time": fmt.Sprintf("%.0f seconds", time.Since(currentProcess.Created).Seconds()),
						"service_id":   serviceID,
						"next_step":    "Operation is taking longer than expected. Use 'process_status' to check progress",
						"tip":          "For long-running operations, consider using wait=false parameter",
					}), nil
				}
				// Process finished with some status
				if currentProcess.Status == "SUCCESS" || currentProcess.Status == "FINISHED" {
					// Even though we timed out waiting, the process succeeded
					// Continue to return success below
					completedProcess = currentProcess
				} else {
					return ErrorResponse(
						"SUBDOMAIN_DISABLE_FAILED",
						fmt.Sprintf("Subdomain disable process failed with status: %s", currentProcess.Status),
						"Check service logs for more details or try again",
					), nil
				}
			} else {
				// Couldn't get process status, return timeout error
				return ErrorResponse(
					"PROCESS_WAIT_TIMEOUT",
					fmt.Sprintf("Timed out waiting for subdomain disable process after 20 seconds"),
					fmt.Sprintf("Use 'process_status' with process_id=%s to check current status", process.ID),
				), nil
			}
		}

		// Check if process succeeded (if we didn't already handle it above)
		if completedProcess != nil && completedProcess.Status != "SUCCESS" && completedProcess.Status != "FINISHED" {
			return ErrorResponse(
				"SUBDOMAIN_DISABLE_FAILED",
				fmt.Sprintf("Failed to disable subdomain access: process status %s", completedProcess.Status),
				"Try again or contact support if the issue persists",
			), nil
		}

		return SuccessResponse(map[string]interface{}{
			"message":      fmt.Sprintf("Successfully disabled subdomain access for service '%s'", service.Name),
			"service_id":   serviceID,
			"service_name": service.Name,
			"process_id":   process.ID,
		}), nil
	})

	// Create subdomain_status tool
	subdomainStatusTool := mcp.NewTool(
		"subdomain_status",
		mcp.WithDescription("Check subdomain access status for a service"),
		mcp.WithString("service_id",
			mcp.Required(),
			mcp.Description("Service ID to check subdomain status for"),
		),
	)

	s.AddTool(subdomainStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceID, err := request.RequireString("service_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_SERVICE_ID",
				"Service ID is required",
				"Provide a valid service ID from 'service_list' tool",
			), nil
		}

		// Get service details
		service, err := client.GetService(ctx, serviceID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Check if service can have subdomain access
		canHaveSubdomain := false
		hasHTTPPort := false
		var portInfo []string

		for _, port := range service.Ports {
			portStr := fmt.Sprintf("Port %d (%s)", port.Port, port.Protocol)
			if port.HTTPRouting {
				portStr += " [HTTP routing enabled]"
				hasHTTPPort = true
			}
			if port.Scheme == "http" || port.Scheme == "https" {
				hasHTTPPort = true
			}
			portInfo = append(portInfo, portStr)
		}

		// Check service type
		serviceCategory := service.ServiceStackTypeInfo.ServiceStackTypeCategory
		if serviceCategory == "USER" && hasHTTPPort {
			canHaveSubdomain = true
		}

		// Get project details for zeropsSubdomainHost
		project, err := client.GetProject(ctx, service.ProjectID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		var subdomainURL string
		if service.SubdomainAccess && project.ZeropsSubdomainHost != nil && *project.ZeropsSubdomainHost != "" {
			// Find the HTTP port
			for _, port := range service.Ports {
				if port.HTTPRouting || port.Scheme == "http" || port.Scheme == "https" {
					subdomainURL = GenerateSubdomainURL(service.Name, *project.ZeropsSubdomainHost, port.Port)
					break
				}
			}
		}

		response := map[string]interface{}{
			"service_id":          serviceID,
			"service_name":        service.Name,
			"service_type":        service.ServiceStackTypeInfo.ServiceStackTypeVersionName,
			"service_status":      service.Status,
			"subdomain_enabled":   service.SubdomainAccess,
			"subdomain_url":       subdomainURL,
			"can_have_subdomain":  canHaveSubdomain,
			"has_http_ports":      hasHTTPPort,
			"ports":               portInfo,
		}

		if !canHaveSubdomain {
			reasons := []string{}
			if serviceCategory != "USER" {
				reasons = append(reasons, fmt.Sprintf("Service category is %s (only USER services can have subdomains)", serviceCategory))
			}
			if !hasHTTPPort {
				reasons = append(reasons, "No HTTP/HTTPS ports configured")
			}
			response["cannot_enable_reason"] = strings.Join(reasons, "; ")
		}

		if service.SubdomainAccess && subdomainURL != "" {
			response["message"] = fmt.Sprintf("Subdomain access is enabled. Service accessible at: %s", subdomainURL)
		} else if service.SubdomainAccess && subdomainURL == "" {
			response["message"] = "Subdomain access is enabled but URL not yet available. Try again in a few seconds."
			response["next_step"] = "Use 'service_info' to check for subdomain URL"
		} else if canHaveSubdomain {
			response["message"] = "Subdomain access is disabled but can be enabled"
			response["next_step"] = "Use 'subdomain_enable' to enable subdomain access"
		} else {
			response["message"] = "This service cannot have subdomain access"
		}

		return SuccessResponse(response), nil
	})
}