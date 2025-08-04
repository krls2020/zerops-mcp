package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
)

// RegisterProcessTools registers process-related tools
func RegisterProcessTools(s *server.MCPServer, client *api.Client) {
	// Create process_status tool
	processStatusTool := mcp.NewTool(
		"process_status",
		mcp.WithDescription("Check the status of a Zerops process/operation"),
		mcp.WithString("process_id",
			mcp.Required(),
			mcp.Description("Process ID to check status for"),
		),
	)

	s.AddTool(processStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		processID, err := request.RequireString("process_id")
		if err != nil {
			return ErrorResponse(
				"INVALID_PROCESS_ID",
				"Process ID is required",
				"Provide a valid process ID from a previous operation",
			), nil
		}

		// Get process details
		process, err := client.GetProcess(ctx, processID)
		if err != nil {
			return HandleAPIError(err), nil
		}

		// Format dates
		var startedStr, finishedStr string
		if process.Started != nil {
			startedStr = process.Started.Format("2006-01-02 15:04:05")
		}
		if process.Finished != nil {
			finishedStr = process.Finished.Format("2006-01-02 15:04:05")
		}

		response := map[string]interface{}{
			"process_id":   process.ID,
			"action":       process.ActionName,
			"status":       process.Status,
			"project_id":   process.ProjectID,
			"service_id":   process.ServiceStackID,
			"created":      process.Created.Format("2006-01-02 15:04:05"),
			"started":      startedStr,
			"finished":     finishedStr,
		}

		// Add user info if available
		if process.CreatedByUser != nil {
			response["created_by"] = process.CreatedByUser.FullName
		}

		// Determine message based on status
		switch process.Status {
		case "PENDING":
			response["message"] = fmt.Sprintf("Process '%s' is pending and will start soon", process.ActionName)
		case "RUNNING":
			response["message"] = fmt.Sprintf("Process '%s' is currently running", process.ActionName)
		case "SUCCESS", "FINISHED":
			response["message"] = fmt.Sprintf("Process '%s' completed successfully", process.ActionName)
		case "FAILED":
			response["message"] = fmt.Sprintf("Process '%s' failed", process.ActionName)
			response["next_step"] = "Check service logs or contact support for more information"
		case "CANCELED":
			response["message"] = fmt.Sprintf("Process '%s' was canceled", process.ActionName)
			if process.CanceledByUser != nil {
				response["canceled_by"] = process.CanceledByUser.FullName
			}
		default:
			response["message"] = fmt.Sprintf("Process '%s' has status: %s", process.ActionName, process.Status)
		}

		return SuccessResponse(response), nil
	})
}