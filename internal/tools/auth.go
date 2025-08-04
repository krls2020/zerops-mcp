package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
)

// RegisterAuthTools registers all authentication tools
func RegisterAuthTools(s *server.MCPServer, client *api.Client) {
	// Register auth_validate tool
	authValidateTool := mcp.NewTool(
		"auth_validate",
		mcp.WithDescription("Validate Zerops API key and check authentication"),
	)

	s.AddTool(authValidateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Validate API key by making a simple API call
		user, err := client.GetCurrentUser(ctx)
		if err != nil {
			return HandleAPIError(err), nil
		}

		return SuccessResponse(map[string]interface{}{
			"message":    "Authentication successful",
			"userId":     user.ID,
			"email":      user.Email,
			"fullName":   user.FullName,
			"status":     user.Status,
			"clients":    len(user.ClientUserList),
			"nextStep":   "Use 'project_list' to see your projects",
		}), nil
	})

	// Register platform_info tool
	platformInfoTool := mcp.NewTool(
		"platform_info",
		mcp.WithDescription("Get Zerops platform information and capabilities"),
	)

	s.AddTool(platformInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get user info to check platform access
		user, err := client.GetCurrentUser(ctx)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Get regions to show available regions
		regions, err := client.ListRegions(ctx)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Prepare region list
		regionNames := make([]string, len(regions))
		defaultRegion := ""
		for i, region := range regions {
			regionNames[i] = region.Name
			if region.IsDefault {
				defaultRegion = region.Name
			}
		}

		return SuccessResponse(map[string]interface{}{
			"message":        "Platform information retrieved",
			"apiUrl":         client.GetBaseURL(),
			"userEmail":      user.Email,
			"availableRegions": regionNames,
			"defaultRegion":  defaultRegion,
			"features": map[string]bool{
				"projects":    true,
				"services":    true,
				"deployment":  true,
				"vpn":         true,
				"logs":        true,
				"metrics":     true,
			},
			"nextStep": "Use 'region_list' for detailed region info or 'project_create' to create a project",
		}), nil
	})

	// Register region_list tool
	regionListTool := mcp.NewTool(
		"region_list",
		mcp.WithDescription("List all available Zerops regions"),
	)

	s.AddTool(regionListTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		regions, err := client.ListRegions(ctx)
		if err != nil {
			return HandleAPIError(err), nil
		}

		if len(regions) == 0 {
			return ErrorResponse(
				"NO_REGIONS",
				"No regions available",
				"Contact Zerops support if this persists",
			), nil
		}

		// Format region information
		response := fmt.Sprintf("Available Zerops regions (%d total):\n\n", len(regions))
		
		for _, region := range regions {
			response += fmt.Sprintf("• %s", region.Name)
			if region.IsDefault {
				response += " (default)"
			}
			response += fmt.Sprintf("\n  API endpoint: %s\n", region.Address)
		}

		response += "\nNext step: Use 'project_create' with one of these regions"

		return mcp.NewToolResultText(response), nil
	})
}