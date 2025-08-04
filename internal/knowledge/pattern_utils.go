package knowledge

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// GetAllPatterns returns all available patterns
func GetAllPatterns() ([]*Pattern, error) {
	entries, err := knowledgeFS.ReadDir("data/patterns")
	if err != nil {
		return nil, err
	}

	var patterns []*Pattern
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := knowledgeFS.ReadFile(filepath.Join("data/patterns", entry.Name()))
			if err != nil {
				continue
			}

			var pattern Pattern
			if err := json.Unmarshal(data, &pattern); err != nil {
				continue
			}

			// Validate pattern before adding
			if errs := validatePattern(&pattern); len(errs) > 0 {
				// Log validation errors but still include pattern
				// In production, you might want to exclude invalid patterns
				// Note: Don't print to stdout in MCP server - it interferes with JSON protocol
				// TODO: Use proper logging framework that writes to stderr or log file
			}

			patterns = append(patterns, &pattern)
		}
	}

	return patterns, nil
}

// validatePattern validates a pattern's structure and content
func validatePattern(pattern *Pattern) []string {
	errors := []string{}
	
	// Check required fields
	if pattern.PatternID == "" {
		errors = append(errors, "Missing patternId")
	}
	if pattern.Name == "" {
		errors = append(errors, "Missing name")
	}
	if pattern.Framework == "" {
		errors = append(errors, "Missing framework")
	}
	if pattern.ZeropsYml == nil {
		errors = append(errors, "Missing zeropsYml configuration")
	}
	
	// Validate zerops.yml structure if present
	if pattern.ZeropsYml != nil {
		if zerops, ok := pattern.ZeropsYml["zerops"].([]interface{}); ok && len(zerops) > 0 {
			for i, service := range zerops {
				if svc, ok := service.(map[string]interface{}); ok {
					// Check for setup field
					if _, hasSetup := svc["setup"]; !hasSetup {
						errors = append(errors, fmt.Sprintf("Service %d missing 'setup' field", i))
					}
					
					// Check build section
					if build, ok := svc["build"].(map[string]interface{}); ok {
						if _, hasBase := build["base"]; !hasBase {
							if _, hasDeployFiles := build["deployFiles"]; !hasDeployFiles {
								errors = append(errors, fmt.Sprintf("Service %d build missing 'base' or 'deployFiles'", i))
							}
						}
					} else {
						errors = append(errors, fmt.Sprintf("Service %d missing 'build' section", i))
					}
					
					// Check run section for runtime services
					if run, ok := svc["run"].(map[string]interface{}); ok {
						// Check for prepareCommands in runtime services
						if _, hasPrepare := run["prepareCommands"]; !hasPrepare {
							// Check if this is a runtime service (has base in build)
							if build, ok := svc["build"].(map[string]interface{}); ok {
								if _, hasBase := build["base"]; hasBase {
									// It's a runtime service, should have prepareCommands
									errors = append(errors, fmt.Sprintf("Service %d missing 'prepareCommands' in run section", i))
								}
							}
						}
						
						// Check for base in run section
						if _, hasBase := run["base"]; !hasBase {
							// Only required for runtime services
							if build, ok := svc["build"].(map[string]interface{}); ok {
								if _, hasBase := build["base"]; hasBase {
									errors = append(errors, fmt.Sprintf("Service %d missing 'base' in run section", i))
								}
							}
						}
					}
				}
			}
		} else {
			errors = append(errors, "Invalid zeropsYml structure")
		}
	}
	
	// Validate services array
	if len(pattern.Services) == 0 {
		errors = append(errors, "No services defined")
	}
	
	return errors
}