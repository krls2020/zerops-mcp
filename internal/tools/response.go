package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
	
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zeropsio/zerops-mcp-v3/internal/api"
)

// ResponseBuilder helps construct structured responses
type ResponseBuilder struct {
	sections []section
}

type section struct {
	title   string
	content string
	isList  bool
	items   []string
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		sections: make([]section, 0),
	}
}

// AddSection adds a text section to the response
func (rb *ResponseBuilder) AddSection(title, content string) *ResponseBuilder {
	rb.sections = append(rb.sections, section{
		title:   title,
		content: content,
	})
	return rb
}

// AddList adds a list section to the response
func (rb *ResponseBuilder) AddList(title string, items []string) *ResponseBuilder {
	rb.sections = append(rb.sections, section{
		title:  title,
		isList: true,
		items:  items,
	})
	return rb
}

// AddKeyValue adds a key-value pair
func (rb *ResponseBuilder) AddKeyValue(key string, value interface{}) *ResponseBuilder {
	content := fmt.Sprintf("%v", value)
	return rb.AddSection(key, content)
}

// AddNextSteps adds a next steps section
func (rb *ResponseBuilder) AddNextSteps(steps ...string) *ResponseBuilder {
	return rb.AddList("Next steps", steps)
}

// Build constructs the final response string
func (rb *ResponseBuilder) Build() string {
	var result strings.Builder
	
	for i, s := range rb.sections {
		if i > 0 {
			result.WriteString("\n")
		}
		
		if s.title != "" {
			result.WriteString(fmt.Sprintf("%s:\n", s.title))
		}
		
		if s.isList {
			for _, item := range s.items {
				result.WriteString(fmt.Sprintf("- %s\n", item))
			}
		} else {
			result.WriteString(s.content)
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// ListFormatter provides consistent list formatting
type ListFormatter struct {
	Title      string
	NoItemsMsg string
	NextSteps  []string
}

// FormatListResponse formats a list of items consistently
func FormatListResponse[T ListItem](formatter ListFormatter, items []T) string {
	rb := NewResponseBuilder()
	
	if len(items) == 0 {
		rb.AddSection("", formatter.NoItemsMsg)
	} else {
		var formatted []string
		for i, item := range items {
			formatted = append(formatted, item.Format(i+1))
		}
		rb.AddSection(fmt.Sprintf("%s (%d total)", formatter.Title, len(items)), 
			strings.Join(formatted, "\n"))
	}
	
	if len(formatter.NextSteps) > 0 {
		rb.AddNextSteps(formatter.NextSteps...)
	}
	
	return rb.Build()
}

// ListItem interface for items that can be formatted in lists
type ListItem interface {
	Format(index int) string
}

// ProcessWaitResult handles async process waiting with consistent responses
type ProcessWaitConfig struct {
	Wait          bool
	Timeout       time.Duration
	OperationName string
	EntityName    string
}

// HandleAsyncProcess manages async operations with optional waiting
func HandleAsyncProcess(ctx context.Context, client *api.Client, process *api.Process, config ProcessWaitConfig) *mcp.CallToolResult {
	if !config.Wait {
		return SuccessResponse(map[string]interface{}{
			"message":    fmt.Sprintf("%s initiated for %s", config.OperationName, config.EntityName),
			"process_id": process.ID,
			"status":     "PENDING",
			"next_step":  "Use 'process_status' to check progress",
		})
	}
	
	completedProcess, err := client.WaitForProcess(ctx, process.ID, config.Timeout)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return HandleProcessTimeout(ctx, client, process, config)
		}
		return HandleAPIError(err)
	}
	
	return HandleProcessCompletion(completedProcess, config)
}

// HandleProcessTimeout handles process timeout scenarios
func HandleProcessTimeout(ctx context.Context, client *api.Client, process *api.Process, config ProcessWaitConfig) *mcp.CallToolResult {
	// Try to get current status
	currentProcess, _ := client.GetProcess(ctx, process.ID)
	
	rb := NewResponseBuilder()
	rb.AddSection("Operation timed out", 
		fmt.Sprintf("The %s for %s is still in progress after %v",
			config.OperationName, config.EntityName, config.Timeout))
	
	if currentProcess != nil {
		rb.AddKeyValue("Process ID", currentProcess.ID)
		rb.AddKeyValue("Current Status", currentProcess.Status)
	}
	
	rb.AddNextSteps(
		fmt.Sprintf("Use 'process_status --process_id %s' to check progress", process.ID),
		"The operation may still complete successfully",
	)
	
	return ErrorResponse(
		"OPERATION_TIMEOUT",
		rb.Build(),
		"Check process status or wait longer",
	)
}

// HandleProcessCompletion handles completed process results
func HandleProcessCompletion(process *api.Process, config ProcessWaitConfig) *mcp.CallToolResult {
	if process.Status == "SUCCESS" || process.Status == "FINISHED" {
		return SuccessResponse(map[string]interface{}{
			"message":    fmt.Sprintf("%s completed successfully for %s", config.OperationName, config.EntityName),
			"process_id": process.ID,
			"status":     process.Status,
			"duration":   process.LastUpdate.Sub(process.Created).String(),
		})
	}
	
	return ErrorResponse(
		"OPERATION_FAILED",
		fmt.Sprintf("%s failed for %s: %s", config.OperationName, config.EntityName, process.Status),
		"Check logs for more details about the failure",
	)
}