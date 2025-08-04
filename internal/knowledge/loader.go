package knowledge

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed data/runtimes/*.json data/patterns/*.json data/services/*.json data/*.md data/nginx/*.tmpl
var knowledgeFS embed.FS

func GetRuntime(name string) (*RuntimeKnowledge, error) {
	data, err := knowledgeFS.ReadFile(filepath.Join("data/runtimes", name+".json"))
	if err != nil {
		return nil, fmt.Errorf("runtime %s not found", name)
	}

	var runtime RuntimeKnowledge
	if err := json.Unmarshal(data, &runtime); err != nil {
		return nil, fmt.Errorf("failed to parse runtime data: %w", err)
	}

	return &runtime, nil
}

// ScoredPattern holds a pattern with its relevance score
type ScoredPattern struct {
	Pattern *Pattern
	Score   int
}

func SearchPatterns(requirements []string) ([]*Pattern, error) {
	// Get all patterns first
	allPatterns, err := GetAllPatterns()
	if err != nil {
		return nil, err
	}
	
	// If no requirements, return all patterns
	if len(requirements) == 0 {
		return allPatterns, nil
	}
	
	// Score patterns based on requirements
	scored := []ScoredPattern{}
	
	for _, pattern := range allPatterns {
		score := 0
		
		for _, req := range requirements {
			reqLower := strings.ToLower(req)
			
			// Exact framework match (highest priority)
			if strings.EqualFold(pattern.Framework, req) {
				score += 100
			}
			
			// Framework contains requirement
			if strings.Contains(strings.ToLower(pattern.Framework), reqLower) {
				score += 50
			}
			
			// Language match
			if strings.EqualFold(pattern.Language, req) {
				score += 40
			}
			
			// Pattern name match
			if strings.Contains(strings.ToLower(pattern.Name), reqLower) {
				score += 30
			}
			
			// Tag matches
			for _, tag := range pattern.Tags {
				if strings.EqualFold(tag, req) {
					score += 25
				} else if strings.Contains(strings.ToLower(tag), reqLower) {
					score += 15
				}
			}
			
			// Description match
			if strings.Contains(strings.ToLower(pattern.Description), reqLower) {
				score += 10
			}
		}
		
		if score > 0 {
			scored = append(scored, ScoredPattern{Pattern: pattern, Score: score})
		}
	}
	
	// If no patterns matched, return all patterns
	if len(scored) == 0 {
		return allPatterns, nil
	}
	
	// Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	
	// Return top matches
	results := []*Pattern{}
	for i, sp := range scored {
		if i < 10 { // Return top 10 matches
			results = append(results, sp.Pattern)
		}
	}
	
	return results, nil
}

// ValidateConfig performs basic validation on a zerops.yml configuration
func ValidateConfig(config string) ValidationResult {
	result := ValidationResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
	}

	// Basic checks
	if config == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "Configuration is empty")
		return result
	}

	if !strings.Contains(config, "zerops:") {
		result.Valid = false
		result.Errors = append(result.Errors, "Missing 'zerops:' root key")
	}

	if !strings.Contains(config, "setup:") {
		result.Valid = false
		result.Errors = append(result.Errors, "Missing 'setup:' key - service hostname required")
	}

	// Check for common issues
	if strings.Contains(config, "buildFromGit") {
		result.Warnings = append(result.Warnings, "buildFromGit is only for Zerops recipes, not user code")
		result.Suggestions = append(result.Suggestions, "Remove buildFromGit and use zcli push for deployment")
	}

	// Service naming
	setupRegex := regexp.MustCompile(`setup:\s*([a-zA-Z0-9-_]+)`)
	matches := setupRegex.FindAllStringSubmatch(config, -1)
	for _, match := range matches {
		serviceName := match[1]
		if strings.ContainsAny(serviceName, "-_") || strings.ToLower(serviceName) != serviceName {
			result.Errors = append(result.Errors, fmt.Sprintf("Service name '%s' invalid - use only lowercase letters and numbers", serviceName))
			result.Valid = false
		}
	}

	return result
}

// ResolveDependencies analyzes services and returns environment variable mappings
func ResolveDependencies(services []map[string]interface{}) DependencyResolution {
	resolution := DependencyResolution{
		EnvVariables:    make(map[string]map[string]string),
		DeploymentOrder: []string{},
	}

	// First pass: identify services
	serviceTypes := make(map[string]string)
	for _, service := range services {
		if name, ok := service["name"].(string); ok {
			if serviceType, ok := service["type"].(string); ok {
				serviceTypes[name] = serviceType
				
				// Databases should be deployed first
				if strings.Contains(serviceType, "postgresql") || strings.Contains(serviceType, "mariadb") {
					resolution.DeploymentOrder = append([]string{name}, resolution.DeploymentOrder...)
				} else {
					resolution.DeploymentOrder = append(resolution.DeploymentOrder, name)
				}
			}
		}
	}

	// Second pass: create environment variables
	for serviceName := range serviceTypes {
		envVars := make(map[string]string)
		
		// Check for database dependencies
		for depName, depType := range serviceTypes {
			if depName == serviceName {
				continue
			}
			
			if strings.Contains(depType, "postgresql") || strings.Contains(depType, "mariadb") {
				envVars["DB_HOST"] = fmt.Sprintf("${%s_hostname}", depName)
				envVars["DB_USER"] = fmt.Sprintf("${%s_user}", depName)
				envVars["DB_PASS"] = fmt.Sprintf("${%s_password}", depName)
				envVars["DB_NAME"] = fmt.Sprintf("${%s_dbName}", depName)
			}
			
			if strings.Contains(depType, "keydb") || strings.Contains(depType, "valkey") {
				envVars["CACHE_HOST"] = fmt.Sprintf("${%s_hostname}", depName)
				envVars["CACHE_PORT"] = fmt.Sprintf("${%s_port}", depName)
			}
		}
		
		if len(envVars) > 0 {
			resolution.EnvVariables[serviceName] = envVars
		}
	}

	return resolution
}

// GetService retrieves service configuration knowledge
func GetService(serviceType string) (*ServiceKnowledge, error) {
	// Extract base service name (e.g., "nodejs" from "nodejs@20")
	baseName := strings.Split(serviceType, "@")[0]
	
	// Handle special cases like php-apache -> php
	if strings.HasPrefix(baseName, "php-") && baseName != "php-apache" && baseName != "php-nginx" {
		baseName = "php"
	}
	
	// First try to find in native services
	data, err := knowledgeFS.ReadFile(filepath.Join("data/services", baseName+".json"))
	if err == nil {
		var service ServiceKnowledge
		if err := json.Unmarshal(data, &service); err != nil {
			return nil, fmt.Errorf("failed to parse service data: %w", err)
		}
		return &service, nil
	}
	
	// If not found in services, check if it's a recipe service
	// Recipe services are stored in patterns with specific naming
	recipePatterns := []string{
		fmt.Sprintf("recipe-%s.json", baseName),  // e.g., recipe-mailpit.json
		fmt.Sprintf("%s.json", baseName),          // e.g., s3browser.json
	}
	
	for _, pattern := range recipePatterns {
		data, err := knowledgeFS.ReadFile(filepath.Join("data/patterns", pattern))
		if err == nil {
			var pattern Pattern
			if err := json.Unmarshal(data, &pattern); err != nil {
				continue
			}
			
			// Convert pattern to ServiceKnowledge format
			return convertRecipePatternToService(&pattern), nil
		}
	}
	
	return nil, fmt.Errorf("service %s not found", serviceType)
}

// GetAllServices returns a list of all available services including recipe services
func GetAllServices() ([]*ServiceSummary, error) {
	// First get native services
	data, err := knowledgeFS.ReadFile("data/services/index.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read service index: %w", err)
	}

	var index ServiceIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse service index: %w", err)
	}

	services := index.Services
	
	// Add common recipe services
	recipeServices := []*ServiceSummary{
		{
			Type:        "mailpit",
			DisplayName: "Mailpit",
			Category:    "recipe",
			Tags:        []string{"email", "testing", "recipe", "buildFromGit"},
		},
		{
			Type:        "adminer",
			DisplayName: "Adminer",
			Category:    "recipe",
			Tags:        []string{"database", "admin", "recipe", "buildFromGit"},
		},
		{
			Type:        "s3browser",
			DisplayName: "S3 Browser",
			Category:    "recipe",
			Tags:        []string{"storage", "s3", "recipe", "buildFromGit"},
		},
	}
	
	services = append(services, recipeServices...)

	return services, nil
}

// GetKnowledgeBase returns the complete knowledge base documentation
func GetKnowledgeBase() (string, error) {
	data, err := knowledgeFS.ReadFile("data/ZEROPS_KNOWLEDGE_BASE.md")
	if err != nil {
		return "", fmt.Errorf("failed to read knowledge base: %w", err)
	}

	return string(data), nil
}

// GetNginxConfig returns the nginx configuration template for a framework
func GetNginxConfig(framework string) (string, error) {
	// Map framework names to config files
	configMap := map[string]string{
		"laravel":   "laravel.conf.tmpl",
		"symfony":   "symfony.conf.tmpl",
		"wordpress": "wordpress.conf.tmpl",
		"default":   "default-php.conf.tmpl",
	}
	
	// Get the config file name
	configFile, exists := configMap[strings.ToLower(framework)]
	if !exists {
		configFile = configMap["default"]
	}
	
	// Read the config file
	data, err := knowledgeFS.ReadFile(filepath.Join("data/nginx", configFile))
	if err != nil {
		// Fallback to default if specific config not found
		data, err = knowledgeFS.ReadFile(filepath.Join("data/nginx", "default-php.conf.tmpl"))
		if err != nil {
			return "", fmt.Errorf("failed to read nginx config: %w", err)
		}
	}
	
	return string(data), nil
}

// convertRecipePatternToService converts a recipe pattern to ServiceKnowledge format
func convertRecipePatternToService(pattern *Pattern) *ServiceKnowledge {
	// Extract the main service info from the pattern
	var serviceType string
	var gitRepo string
	
	// Look for the service that uses buildFromGit
	if pattern.Services != nil {
		for _, svc := range pattern.Services {
			if svc["buildFromGit"] != nil {
				gitRepo = svc["buildFromGit"].(string)
			}
			if svcType, ok := svc["type"].(string); ok {
				serviceType = svcType
			}
		}
	}
	
	// If no buildFromGit found in services and we have sourceRecipe, construct the URL
	if gitRepo == "" && pattern.SourceRecipe != "" {
		// Convert sourceRecipe to GitHub URL
		// sourceRecipe format: "recipe-name-main" -> "https://github.com/zeropsio/recipe-name"
		recipeName := strings.TrimSuffix(pattern.SourceRecipe, "-main")
		gitRepo = fmt.Sprintf("https://github.com/zeropsio/%s", recipeName)
	}
	
	// For patterns without services array (like adminer), extract from zeropsYml
	if serviceType == "" && pattern.ZeropsYml != nil {
		if zerops, ok := pattern.ZeropsYml["zerops"].([]interface{}); ok && len(zerops) > 0 {
			if svc, ok := zerops[0].(map[string]interface{}); ok {
				if run, ok := svc["run"].(map[string]interface{}); ok {
					if base, ok := run["base"].(string); ok {
						serviceType = base
					}
				}
			}
		}
	}
	
	// Build service knowledge emphasizing this is a recipe service
	var description string
	var deploymentType string
	var bestPractices []string
	
	if gitRepo != "" {
		description = fmt.Sprintf("%s (Recipe Service - deployed via buildFromGit)", pattern.Description)
		deploymentType = "buildFromGit"
		bestPractices = []string{
			"Use buildFromGit in the service definition",
			"Do not attempt to deploy with zcli push",
			"These services are maintained by Zerops and auto-update",
			fmt.Sprintf("Import using: buildFromGit: %s", gitRepo),
			"Recipe services are pre-built applications maintained by Zerops",
		}
	} else {
		description = fmt.Sprintf("%s (Recipe Service - manual deployment)", pattern.Description)
		deploymentType = "manual"
		bestPractices = []string{
			"This recipe requires manual deployment steps",
			"Check the zerops.yml for build commands",
			"Do not use buildFromGit for this service",
			"Deploy using zcli push with the provided zerops.yml",
		}
	}
	
	sk := &ServiceKnowledge{
		Type:        pattern.Name,  // Use pattern name as the service type
		DisplayName: pattern.Name,
		Category:    "recipe",
		Description: description,
		Versions:    []string{"latest"},
		Configuration: map[string]interface{}{
			"deploymentType": deploymentType,
			"gitRepository":  gitRepo,
			"baseService":    serviceType,
		},
		AutoGenVariables: []string{},
		BestPractices:    bestPractices,
		CommonIssues: func() []map[string]string {
			if gitRepo != "" {
				return []map[string]string{
					{
						"issue":    "Cannot deploy with zcli push",
						"solution": "Recipe services must use buildFromGit in project_import",
					},
					{
						"issue":    "Service not starting",
						"solution": "Check that buildFromGit URL is correct and accessible",
					},
				}
			}
			return []map[string]string{
				{
					"issue":    "Build commands failing",
					"solution": "Check the zerops.yml build section for required dependencies",
				},
				{
					"issue":    "Manual deployment required",
					"solution": "Use the pattern's zerops.yml with zcli push",
				},
			}
		}(),
	}
	
	// Extract ports and add to configuration
	var ports []int
	if pattern.ZeropsYml != nil {
		if zerops, ok := pattern.ZeropsYml["zerops"].([]interface{}); ok && len(zerops) > 0 {
			if svc, ok := zerops[0].(map[string]interface{}); ok {
				if run, ok := svc["run"].(map[string]interface{}); ok {
					if portsData, ok := run["ports"].([]interface{}); ok {
						for _, p := range portsData {
							if portMap, ok := p.(map[string]interface{}); ok {
								if port, ok := portMap["port"].(float64); ok {
									ports = append(ports, int(port))
								}
							}
						}
					}
				}
			}
		}
	}
	if len(ports) > 0 {
		sk.Configuration["defaultPorts"] = ports
	}
	
	// Add specific notes for common recipe services
	switch strings.ToLower(pattern.Name) {
	case "mailpit":
		sk.Configuration["description"] = "Email testing tool with web UI on port 8025 and SMTP on port 1025"
	case "adminer":
		sk.Configuration["description"] = "Database management UI on port 80"
	case "s3browser":
		sk.Configuration["description"] = "Object storage browser UI, requires object-storage service"
		sk.Configuration["requiredServices"] = []string{"object-storage"}
	}
	
	// Extract setup name from zeropsYml
	var setupName string
	if pattern.ZeropsYml != nil {
		if zerops, ok := pattern.ZeropsYml["zerops"].([]interface{}); ok && len(zerops) > 0 {
			if svc, ok := zerops[0].(map[string]interface{}); ok {
				if setup, ok := svc["setup"].(string); ok {
					setupName = setup
				}
			}
		}
	}
	
	// Add example YAML for recipe services
	if gitRepo != "" {
		exampleService := map[string]interface{}{
			"hostname":             strings.ToLower(pattern.Name),
			"type":                 serviceType,
			"buildFromGit":         gitRepo,
			"enableSubdomainAccess": true,
		}
		
		// Add zeropsSetup if the hostname differs from the setup name
		if setupName != "" && setupName != strings.ToLower(pattern.Name) {
			exampleService["zeropsSetup"] = setupName
		}
		
		sk.ExampleYAML = map[string]interface{}{
			"services": []interface{}{exampleService},
		}
		
		// Add note about zeropsSetup
		if setupName != "" {
			sk.BestPractices = append(sk.BestPractices, 
				fmt.Sprintf("The setup name in the recipe is '%s' - use 'zeropsSetup: %s' if you name your service differently", setupName, setupName))
			
			// Add an alternative example with different hostname
			sk.Configuration["alternativeExample"] = map[string]interface{}{
				"services": []interface{}{
					map[string]interface{}{
						"hostname":             "database-admin",  // Different name
						"type":                 serviceType,
						"buildFromGit":         gitRepo,
						"enableSubdomainAccess": true,
						"zeropsSetup":          setupName,  // Points to the correct setup
					},
				},
			}
		}
	} else {
		// For manual recipes, show the basic service definition
		sk.ExampleYAML = map[string]interface{}{
			"services": []map[string]interface{}{
				{
					"hostname":             strings.ToLower(pattern.Name),
					"type":                 serviceType,
					"enableSubdomainAccess": true,
				},
			},
			"note": "Deploy with zcli push using the pattern's zerops.yml",
		}
	}
	
	return sk
}