package tools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
)

// ParamValidator defines validation rules for a parameter
type ParamValidator struct {
	ParamName  string
	ErrorCode  string
	ErrorMsg   string
	Resolution string
}

// CommonValidators contains pre-defined validators for common parameters
var CommonValidators = struct {
	ProjectID ParamValidator
	ServiceID ParamValidator
	Region    ParamValidator
	Name      ParamValidator
}{
	ProjectID: ParamValidator{
		ParamName:  "project_id",
		ErrorCode:  "INVALID_PROJECT_ID",
		ErrorMsg:   "Project ID is required",
		Resolution: "Provide a valid project ID from 'project_list' tool",
	},
	ServiceID: ParamValidator{
		ParamName:  "service_id",
		ErrorCode:  "INVALID_SERVICE_ID",
		ErrorMsg:   "Service ID is required",
		Resolution: "Provide a valid service ID from 'service_list' tool",
	},
	Region: ParamValidator{
		ParamName:  "region",
		ErrorCode:  "INVALID_REGION",
		ErrorMsg:   "Region is required",
		Resolution: "Use 'region_list' tool to see available regions",
	},
	Name: ParamValidator{
		ParamName:  "name",
		ErrorCode:  "INVALID_NAME",
		ErrorMsg:   "Name is required",
		Resolution: "Provide a valid name (3-50 characters)",
	},
}

// RequireParam validates and retrieves a required parameter
func RequireParam(request mcp.CallToolRequest, validator ParamValidator) (string, *mcp.CallToolResult) {
	value, err := request.RequireString(validator.ParamName)
	if err != nil {
		return "", ErrorResponse(validator.ErrorCode, validator.ErrorMsg, validator.Resolution)
	}
	return value, nil
}

// GetClientID retrieves the client ID from the current user
func GetClientID(ctx context.Context, client *api.Client) (string, *mcp.CallToolResult) {
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return "", HandleAPIError(err)
	}

	if len(user.ClientUserList) == 0 {
		return "", ErrorResponse(
			"NO_CLIENT",
			"No client associations found for user",
			"Contact Zerops support to resolve account issues",
		)
	}

	return user.ClientUserList[0].ClientID, nil
}

// ValidateProjectAccess checks if a project exists and is accessible
func ValidateProjectAccess(ctx context.Context, client *api.Client, projectID string) (*api.Project, *mcp.CallToolResult) {
	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrorResponse(
				"PROJECT_NOT_FOUND",
				fmt.Sprintf("Project '%s' not found", projectID),
				"Check the project ID or use 'project_list' to see available projects",
			)
		}
		return nil, HandleAPIError(err)
	}
	return project, nil
}

// ValidateServiceAccess checks if a service exists and is accessible
func ValidateServiceAccess(ctx context.Context, client *api.Client, projectID, serviceID string) (*api.ServiceDetails, *mcp.CallToolResult) {
	service, err := client.GetService(ctx, serviceID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrorResponse(
				"SERVICE_NOT_FOUND",
				fmt.Sprintf("Service '%s' not found in project", serviceID),
				"Check the service ID or use 'service_list' to see available services",
			)
		}
		return nil, HandleAPIError(err)
	}
	return service, nil
}

// ValidateServiceName checks if a service name is valid
func ValidateServiceName(name string) *mcp.CallToolResult {
	if len(name) < 3 || len(name) > 30 {
		return ErrorResponse(
			"INVALID_SERVICE_NAME",
			"Service name must be 3-30 characters",
			"Use a name between 3 and 30 characters",
		)
	}

	// Check for valid characters (only lowercase letters and numbers)
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			return ErrorResponse(
				"INVALID_SERVICE_NAME",
				"Service name can only contain lowercase letters and numbers",
				"Use only lowercase letters (a-z) and numbers (0-9) in the service name",
			)
		}
	}

	return nil
}

// isNotFoundError checks if an error is a 404 not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "404") || contains(errStr, "not found")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && len(substr) > 0 && 
		(s == substr || (len(s) > len(substr) && containsAt(s, substr, 0)))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if s[start+i] != substr[i] {
			return false
		}
	}
	return true
}