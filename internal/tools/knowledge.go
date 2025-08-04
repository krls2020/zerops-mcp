package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zeropsio/zerops-mcp-v3/internal/knowledge"
)

// RegisterKnowledgeTools registers knowledge retrieval tools
func RegisterKnowledgeTools(s *server.MCPServer) {
	// Runtime tool
	runtimeTool := mcp.NewTool(
		"knowledge_get_runtime",
		mcp.WithDescription("Get runtime configuration knowledge - USE THIS FIRST when creating any service configuration"),
		mcp.WithString("runtime",
			mcp.Required(),
			mcp.Description("Runtime name (e.g., 'nodejs', 'python', 'php')"),
		),
	)

	s.AddTool(runtimeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		runtimeName, err := request.RequireString("runtime")
		if err != nil {
			return mcp.NewToolResultError("runtime parameter required"), nil
		}

		runtime, err := knowledge.GetRuntime(runtimeName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return as formatted JSON
		data, err := json.MarshalIndent(runtime, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to format response"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})

	// Pattern search tool
	searchTool := mcp.NewTool(
		"knowledge_search_patterns",
		mcp.WithDescription("Search framework deployment patterns - ALWAYS USE THIS FIRST when user mentions a framework (Laravel, Django, Next.js, etc.)"),
		mcp.WithString("tags",
			mcp.Description("Optional tags to filter patterns (e.g., 'php', 'laravel', 'nodejs')"),
		),
	)

	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse tags parameter
		requirements := []string{}
		tagsParam := request.GetString("tags", "")
		if tagsParam != "" {
			// Handle different formats: "tag1,tag2" or "tag1 tag2"
			tagsParam = strings.ReplaceAll(tagsParam, ",", " ")
			for _, tag := range strings.Fields(tagsParam) {
				tag = strings.Trim(tag, `"'[]`)
				if tag != "" {
					requirements = append(requirements, tag)
				}
			}
		}

		patterns, err := knowledge.SearchPatterns(requirements)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Format response for better LLM consumption
		if len(patterns) == 0 {
			return mcp.NewToolResultText("No deployment patterns found. Try different search terms or use 'knowledge_list_services' to see available services."), nil
		}

		var response strings.Builder
		response.WriteString(fmt.Sprintf("Found %d deployment pattern(s):\n\n", len(patterns)))

		for i, pattern := range patterns {
			response.WriteString(fmt.Sprintf("## %d. %s (%s)\n", i+1, pattern.Name, pattern.Framework))
			response.WriteString(fmt.Sprintf("**Description**: %s\n", pattern.Description))
			response.WriteString(fmt.Sprintf("**Language**: %s\n", pattern.Language))
			response.WriteString(fmt.Sprintf("**Tags**: %s\n", strings.Join(pattern.Tags, ", ")))
			
			// Include the COMPLETE pattern data for Claude Code to use
			response.WriteString("\n**COMPLETE PATTERN DATA**:\n```json\n")
			patternData, err := json.MarshalIndent(pattern, "", "  ")
			if err == nil {
				response.WriteString(string(patternData))
			}
			response.WriteString("\n```\n\n")

			// Show how to use the pattern
			response.WriteString("**To use this pattern**:\n")
			response.WriteString("1. Extract the `services` field from the pattern above for project_import\n")
			response.WriteString("2. Extract the `zeropsYml` field and save as zerops.yml\n")
			response.WriteString("3. Update the `setup` field in zerops.yml to match your service name\n")
			response.WriteString("4. Use project_import with the services YAML (NOT manually created!)\n\n")

			// Add best practices if searching for specific framework
			if len(requirements) > 0 && i == 0 { // First match for specific search
				if len(pattern.BestPractices) > 0 {
					response.WriteString("**Best Practices**:\n")
					for _, practice := range pattern.BestPractices {
						response.WriteString(fmt.Sprintf("- %s\n", practice))
					}
					response.WriteString("\n")
				}

				if len(pattern.CommonIssues) > 0 {
					response.WriteString("**Common Issues to Avoid**:\n")
					for _, issue := range pattern.CommonIssues {
						if prob, ok := issue["issue"]; ok && prob != "" {
							if sol, ok := issue["solution"]; ok && sol != "" {
								response.WriteString(fmt.Sprintf("- **%s**: %s\n", prob, sol))
							}
						}
					}
					response.WriteString("\n")
				}
			}

			response.WriteString("---\n\n")
		}

		response.WriteString("**CRITICAL REMINDER**:\n")
		response.WriteString("- The pattern contains a COMPLETE 'services' field - USE IT!\n")
		response.WriteString("- DO NOT create service configurations manually\n")
		response.WriteString("- Extract the services array from the JSON above\n")
		response.WriteString("- Convert to YAML for project_import\n")

		return mcp.NewToolResultText(response.String()), nil
	})

	// Validation tool
	validateTool := mcp.NewTool(
		"knowledge_validate_config",
		mcp.WithDescription("Validate zerops.yml configuration - ALWAYS USE before project_import or deploy_push to prevent errors"),
		mcp.WithString("config",
			mcp.Required(),
			mcp.Description("Zerops YAML configuration to validate"),
		),
	)

	s.AddTool(validateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		config, err := request.RequireString("config")
		if err != nil {
			return mcp.NewToolResultError("config parameter required"), nil
		}

		result := knowledge.ValidateConfig(config)

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to format response"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})

	// Dependency resolution tool
	resolveTool := mcp.NewTool(
		"knowledge_resolve_dependencies",
		mcp.WithDescription("Resolve cross-service dependencies and environment variables"),
		mcp.WithObject("services",
			mcp.Required(),
			mcp.Description("List of services to resolve dependencies for"),
		),
	)

	s.AddTool(resolveTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// For now, create example services
		services := []map[string]interface{}{
			{"name": "api", "type": "nodejs@20"},
			{"name": "db", "type": "postgresql@16"},
		}

		resolution := knowledge.ResolveDependencies(services)

		data, err := json.MarshalIndent(resolution, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to format response"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})
	
	// Service knowledge tool
	serviceTool := mcp.NewTool(
		"knowledge_get_service",
		mcp.WithDescription("Get service configuration knowledge"),
		mcp.WithString("service",
			mcp.Required(),
			mcp.Description("Service type (e.g., 'nodejs@20', 'postgresql@16')"),
		),
	)

	s.AddTool(serviceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serviceType, err := request.RequireString("service")
		if err != nil {
			return mcp.NewToolResultError("service parameter required"), nil
		}

		service, err := knowledge.GetService(serviceType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := json.MarshalIndent(service, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to format response"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})
	
	// Service list tool
	serviceListTool := mcp.NewTool(
		"knowledge_list_services",
		mcp.WithDescription("List all available Zerops services"),
	)

	s.AddTool(serviceListTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		services, err := knowledge.GetAllServices()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := json.MarshalIndent(services, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to format response"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})
	
	// Knowledge base tool
	kbTool := mcp.NewTool(
		"knowledge_get_docs",
		mcp.WithDescription("Get comprehensive Zerops documentation"),
		mcp.WithString("section",
			mcp.Description("Optional section to retrieve (e.g., 'networking', 'deployment')"),
		),
	)

	s.AddTool(kbTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		section := request.GetString("section", "")
		
		kb, err := knowledge.GetKnowledgeBase()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		
		// If section requested, try to extract it
		if section != "" {
			// Simple section extraction based on headers
			lines := strings.Split(kb, "\n")
			var extractedSection strings.Builder
			inSection := false
			sectionHeader := fmt.Sprintf("## %s", strings.Title(section))
			
			for _, line := range lines {
				if strings.HasPrefix(line, "## ") {
					if strings.EqualFold(line, sectionHeader) {
						inSection = true
					} else if inSection {
						break // End of our section
					}
				}
				if inSection {
					extractedSection.WriteString(line + "\n")
				}
			}
			
			if extractedSection.Len() > 0 {
				return mcp.NewToolResultText(extractedSection.String()), nil
			}
		}

		return mcp.NewToolResultText(kb), nil
	})
}