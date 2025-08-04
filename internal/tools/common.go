package tools

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ErrorResponse creates a standardized error response
func ErrorResponse(code, message, resolution string) *mcp.CallToolResult {
	errorText := fmt.Sprintf("Error: %s\n%s\n\nResolution: %s", code, message, resolution)
	return mcp.NewToolResultError(errorText)
}

// ErrorResponseWithNext includes a suggestion for the next tool to use
func ErrorResponseWithNext(code, message, resolution, nextTool string) *mcp.CallToolResult {
	errorText := fmt.Sprintf("Error: %s\n%s\n\nResolution: %s\n\nNext step: Use '%s' tool",
		code, message, resolution, nextTool)
	return mcp.NewToolResultError(errorText)
}

// SuccessResponse creates a standardized success response
func SuccessResponse(data map[string]interface{}) *mcp.CallToolResult {
	var response string
	if msg, ok := data["message"].(string); ok {
		response = msg + "\n\n"
		delete(data, "message") // Remove message from data to avoid duplication
	}
	
	// Format the remaining data
	for key, value := range data {
		// Convert key from camelCase to human readable
		humanKey := camelCaseToHuman(key)
		response += fmt.Sprintf("%s: %v\n", humanKey, value)
	}
	
	return mcp.NewToolResultText(response)
}

// InfoResponse creates a standardized informational response (not an error)
func InfoResponse(title, message, nextStep string) *mcp.CallToolResult {
	infoText := fmt.Sprintf("ℹ️  %s\n\n%s", title, message)
	if nextStep != "" {
		infoText += fmt.Sprintf("\n\nNext step: %s", nextStep)
	}
	return mcp.NewToolResultText(infoText)
}

// InfoResponseWithAction creates an info response with an action taken
func InfoResponseWithAction(title, message, action, nextStep string) *mcp.CallToolResult {
	infoText := fmt.Sprintf("ℹ️  %s\n\n%s\n\nAction taken: %s", title, message, action)
	if nextStep != "" {
		infoText += fmt.Sprintf("\n\nNext step: %s", nextStep)
	}
	return mcp.NewToolResultText(infoText)
}

// HandleAPIError converts API errors to user-friendly messages
func HandleAPIError(err error) *mcp.CallToolResult {
	errStr := err.Error()
	
	// Check for specific error types
	switch {
	case strings.Contains(errStr, "401"):
		return ErrorResponse(
			"AUTH_FAILED",
			"Authentication failed - invalid API key",
			"Set ZEROPS_API_KEY environment variable with a valid key",
		)
	case strings.Contains(errStr, "403"):
		return ErrorResponse(
			"FORBIDDEN",
			"Access denied - insufficient permissions",
			"Check that your API key has the required permissions for this operation",
		)
	case strings.Contains(errStr, "404"):
		// Special handling for logs endpoint
		if strings.Contains(errStr, "logs endpoint not available") {
			return ErrorResponseWithNext(
				"LOGS_NOT_AVAILABLE",
				"Service logs are not available through the API",
				"The service may be restarting or logs may not be enabled. Check service status first",
				"service_info",
			)
		}
		return ErrorResponse(
			"NOT_FOUND",
			"Resource not found",
			"Check the ID and try again, or use list tools to find valid resources",
		)
	case strings.Contains(errStr, "409"):
		return ErrorResponse(
			"CONFLICT",
			"Resource already exists or conflict detected",
			"Use a different name or check existing resources with list tools",
		)
	case strings.Contains(errStr, "422"):
		return ErrorResponse(
			"INVALID_INPUT",
			"Invalid input data",
			"Check your input parameters and ensure they meet the requirements",
		)
	case strings.Contains(errStr, "400"):
		// Parse common 400 errors
		if strings.Contains(errStr, "processNotFound") {
			return ErrorResponse(
				"PROCESS_NOT_FOUND",
				"Process not found or has expired",
				"The process may have completed and been removed from the system. Processes are only tracked for a limited time after completion",
			)
		}
		if strings.Contains(errStr, "projectImportInvalidParameter") {
			return ErrorResponse(
				"INVALID_YAML_STRUCTURE",
				"YAML must contain 'services:' section with proper structure",
				"Ensure your YAML has 'services:' section and uses 'hostname' (not 'name') for service names",
			)
		}
		if strings.Contains(errStr, "projectImportProjectIncluded") {
			return ErrorResponse(
				"PROJECT_CONFIG_NOT_ALLOWED",
				"Project configuration is not allowed in import YAML",
				"Remove the 'project:' section from your YAML - only 'services:' section is needed",
			)
		}
		return ErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
			"Check your input parameters and try again",
		)
	case strings.Contains(errStr, "500") || strings.Contains(errStr, "502") || strings.Contains(errStr, "503"):
		return ErrorResponse(
			"SERVER_ERROR",
			"Zerops API is experiencing issues",
			"Wait a moment and try again. If the problem persists, check Zerops status page",
		)
	case strings.Contains(errStr, "connection refused"):
		return ErrorResponse(
			"CONNECTION_FAILED",
			"Failed to connect to Zerops API",
			"Check your internet connection and try again",
		)
	case strings.Contains(errStr, "timeout"):
		return ErrorResponse(
			"TIMEOUT",
			"Request timed out",
			"The operation took too long. Try again or check if the service is responding",
		)
	default:
		return ErrorResponse(
			"API_ERROR",
			fmt.Sprintf("API request failed: %v", err),
			"Check your inputs and try again. If the problem persists, check API status",
		)
	}
}

// camelCaseToHuman converts camelCase to human readable format
func camelCaseToHuman(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, ' ')
		}
		if i == 0 {
			result = append(result, r)
		} else {
			result = append(result, r)
		}
	}
	
	// Capitalize first letter
	str := string(result)
	if len(str) > 0 {
		return strings.ToUpper(str[:1]) + str[1:]
	}
	return str
}

// AnalyzeRuntimeError analyzes common runtime errors and provides solutions
func AnalyzeRuntimeError(logs []string) string {
	var issues []string
	
	// Check each log line for common errors
	for _, line := range logs {
		lowerLine := strings.ToLower(line)
		
		// Python module errors
		if strings.Contains(line, "ModuleNotFoundError") || strings.Contains(line, "ImportError") {
			if strings.Contains(line, "flask") || strings.Contains(line, "django") || 
			   strings.Contains(line, "fastapi") || strings.Contains(line, "requests") {
				issues = append(issues, "Python dependencies not installed in runtime container")
			}
		}
		
		// Node.js module errors
		if strings.Contains(line, "Cannot find module") || strings.Contains(line, "MODULE_NOT_FOUND") {
			issues = append(issues, "Node.js dependencies not installed in runtime container")
		}
		
		// Port binding errors
		if strings.Contains(lowerLine, "address already in use") || strings.Contains(lowerLine, "eaddrinuse") {
			issues = append(issues, "Port conflict - the specified port is already in use")
		}
		
		// Permission errors
		if strings.Contains(lowerLine, "permission denied") || strings.Contains(lowerLine, "eacces") {
			issues = append(issues, "Permission denied - check file permissions or port numbers")
		}
		
		// Database connection errors
		if strings.Contains(lowerLine, "connection refused") || strings.Contains(lowerLine, "econnrefused") {
			if strings.Contains(lowerLine, "postgres") || strings.Contains(lowerLine, "mysql") ||
			   strings.Contains(lowerLine, "mongodb") {
				issues = append(issues, "Database connection failed - check environment variables")
			}
		}
		
		// Memory errors
		if strings.Contains(lowerLine, "out of memory") || strings.Contains(lowerLine, "heap out of memory") {
			issues = append(issues, "Out of memory - increase container resources")
		}
	}
	
	// Remove duplicates and create resolution text
	seen := make(map[string]bool)
	var resolutions []string
	
	for _, issue := range issues {
		if !seen[issue] {
			seen[issue] = true
			resolution := getResolutionForIssue(issue)
			resolutions = append(resolutions, fmt.Sprintf("- %s: %s", issue, resolution))
		}
	}
	
	if len(resolutions) > 0 {
		return "\n\nDetected Issues:\n" + strings.Join(resolutions, "\n")
	}
	
	return ""
}

// getResolutionForIssue provides specific resolution for each issue type
func getResolutionForIssue(issue string) string {
	switch issue {
	case "Python dependencies not installed in runtime container":
		return "Add 'prepareCommands: [\"pip install -r requirements.txt\"]' to the 'run' section of zerops.yml"
	case "Node.js dependencies not installed in runtime container":
		return "Add 'prepareCommands: [\"npm ci --production\"]' to the 'run' section of zerops.yml"
	case "Port conflict - the specified port is already in use":
		return "Change the port in your application or update the port configuration in zerops.yml"
	case "Permission denied - check file permissions or port numbers":
		return "Use ports above 1024 or check file permissions in your deployment"
	case "Database connection failed - check environment variables":
		return "Verify environment variables are set correctly (DB_HOST, DB_USER, etc.) and database service is running"
	case "Out of memory - increase container resources":
		return "Increase minRAM/maxRAM in service configuration or optimize application memory usage"
	default:
		return "Check application configuration and logs for more details"
	}
}

// GenerateSubdomainURL generates the subdomain URL for a service
// The pattern is: 
// - For port 80: https://{service-name}-{zeropsSubdomainHost}.prg1.zerops.app
// - For other ports: https://{service-name}-{zeropsSubdomainHost}-{port}.prg1.zerops.app
// Examples: 
// - https://app-15e3.prg1.zerops.app (port 80)
// - https://mailpit-15e3-8025.prg1.zerops.app (port 8025)
func GenerateSubdomainURL(serviceName, zeropsSubdomainHost string, port int) string {
	if port == 80 {
		return fmt.Sprintf("https://%s-%s.prg1.zerops.app", serviceName, zeropsSubdomainHost)
	}
	return fmt.Sprintf("https://%s-%s-%d.prg1.zerops.app", serviceName, zeropsSubdomainHost, port)
}

// PreprocessYAML preprocesses Zerops YAML to handle special syntax
// It adds the required #yamlPreprocessor=on directive if not present
func PreprocessYAML(yamlContent string) string {
	// The Zerops API requires #yamlPreprocessor=on to enable preprocessing for:
	// - <@generateRandomString(<length>)> - generates random string
	// - <@generateRandomBytes(<length>) | toString> - generates random bytes as string
	// - <@sha256(<value>)> - generates SHA256 hash
	// - And other preprocessing functions
	
	// Check if preprocessing is already enabled
	if strings.Contains(yamlContent, "#yamlPreprocessor=on") {
		return yamlContent
	}
	
	// Add the preprocessor directive at the beginning
	return "#yamlPreprocessor=on\n" + yamlContent
}
