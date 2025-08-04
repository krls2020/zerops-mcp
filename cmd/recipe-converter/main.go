package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// RecipeInfo holds information about a recipe
type RecipeInfo struct {
	Name              string
	Path              string
	Category          string // "tool" or "framework"
	HasZeropsYml      bool
	HasImportYml      bool
	ImportYmlPath     string
	ZeropsYmlPath     string
	HasBuildFromGit   bool
	Framework         string
	Language          string
	Tags              []string
}

// Pattern represents the knowledge pattern structure
type Pattern struct {
	PatternID         string                   `json:"patternId"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description"`
	Framework         string                   `json:"framework"`
	Language          string                   `json:"language"`
	Tags              []string                 `json:"tags"`
	Services          []map[string]interface{} `json:"services"`
	ZeropsYml         map[string]interface{}   `json:"zeropsYml"`
	DeploymentConfig  map[string]interface{}   `json:"deploymentConfig,omitempty"`
	SetupInstructions []string                 `json:"setupInstructions,omitempty"`
	CommonIssues      []map[string]string      `json:"commonIssues,omitempty"`
	BestPractices     []string                 `json:"bestPractices,omitempty"`
	SourceRecipe      string                   `json:"sourceRecipe"`
	DeploymentType    string                   `json:"deploymentType"` // "local" or "git"
}

// ImportYAML represents the structure of import YAML files
type ImportYAML struct {
	Project  map[string]interface{}   `yaml:"project"`
	Services []map[string]interface{} `yaml:"services"`
}

// ZeropsYAML represents the structure of zerops.yml files
type ZeropsYAML struct {
	Zerops []map[string]interface{} `yaml:"zerops"`
}

// Tool services that should keep buildFromGit
var toolServices = map[string]bool{
	"adminer":         true,
	"adminerevo":      true,
	"mailpit":         true,
	"metabase":        true,
	"s3browser":       true,
	"meilisearch":     true,
	"imgproxy":        true,
	"prerender":       true,
	"contember":       true,
	"mattermost":      true,
	"forgejo":         true,
	"chartbrew":       true,
	"directus":        true,
	"ghost":           true,
	"graylog":         true,
	"nextcloud":       true,
	"airflow":         true,
	"github-runner":   true,
	"minecraft":       true,
	"quake3":          true,
	"openpanel":       true,
	"payload":         true,
	"medama":          true,
	"filament":        true,
	"medusa":          true,
}

// Framework mapping
var frameworkMap = map[string]string{
	"laravel":    "Laravel",
	"django":     "Django",
	"flask":      "Flask",
	"fastapi":    "FastAPI",
	"express":    "Express",
	"nextjs":     "Next.js",
	"nuxt":       "Nuxt",
	"react":      "React",
	"vue":        "Vue",
	"angular":    "Angular",
	"svelte":     "Svelte",
	"astro":      "Astro",
	"remix":      "Remix",
	"gatsby":     "Gatsby",
	"rails":      "Rails",
	"phoenix":    "Phoenix",
	"symfony":    "Symfony",
	"wordpress":  "WordPress",
	"strapi":     "Strapi",
	"nestjs":     "NestJS",
	"adonis":     "AdonisJS",
	"hono":       "Hono",
	"elysia":     "Elysia",
}

// Language mapping
var languageMap = map[string]string{
	"nodejs":    "JavaScript",
	"python":    "Python",
	"php":       "PHP",
	"ruby":      "Ruby",
	"go":        "Go",
	"rust":      "Rust",
	"java":      "Java",
	"dotnet":    "C#",
	"elixir":    "Elixir",
	"deno":      "TypeScript",
	"bun":       "TypeScript",
}

func main() {
	recipesDir := "docs/zerops-recipes-library"
	outputDir := "internal/knowledge/data/patterns"

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Process recipes
	recipes, err := scanRecipes(recipesDir)
	if err != nil {
		log.Fatalf("Failed to scan recipes: %v", err)
	}

	fmt.Printf("Found %d recipes\n", len(recipes))

	// Process each recipe
	successCount := 0
	failCount := 0
	
	for _, recipe := range recipes {
		fmt.Printf("\nProcessing %s (%s)...\n", recipe.Name, recipe.Category)
		
		pattern, err := convertRecipeToPattern(recipe)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			failCount++
			continue
		}

		// Save pattern
		outputPath := filepath.Join(outputDir, recipe.Name+".json")
		if err := savePattern(pattern, outputPath); err != nil {
			fmt.Printf("  ❌ Failed to save: %v\n", err)
			failCount++
			continue
		}

		fmt.Printf("  ✅ Converted successfully\n")
		successCount++
	}

	fmt.Printf("\n=== Conversion Summary ===\n")
	fmt.Printf("Total recipes: %d\n", len(recipes))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failCount)

	// Create index
	if err := createPatternIndex(outputDir); err != nil {
		log.Printf("Failed to create pattern index: %v", err)
	}
}

func scanRecipes(dir string) ([]*RecipeInfo, error) {
	var recipes []*RecipeInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		recipe := &RecipeInfo{
			Name: strings.TrimSuffix(entry.Name(), "-main"),
			Path: filepath.Join(dir, entry.Name()),
		}

		// Check for files
		files, err := os.ReadDir(recipe.Path)
		if err != nil {
			continue
		}

		for _, file := range files {
			switch {
			case file.Name() == "zerops.yml":
				recipe.HasZeropsYml = true
				recipe.ZeropsYmlPath = filepath.Join(recipe.Path, file.Name())
			case strings.HasSuffix(file.Name(), "-import.yml") || strings.HasSuffix(file.Name(), "-import.yaml"):
				recipe.HasImportYml = true
				recipe.ImportYmlPath = filepath.Join(recipe.Path, file.Name())
			}
		}

		// Determine category
		recipeName := strings.ToLower(recipe.Name)
		if isToolService(recipeName) {
			recipe.Category = "tool"
		} else {
			recipe.Category = "framework"
		}

		// Extract framework and language
		recipe.Framework = extractFramework(recipeName)
		recipe.Language = extractLanguage(recipeName)
		recipe.Tags = extractTags(recipeName)

		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func isToolService(name string) bool {
	for tool := range toolServices {
		if strings.Contains(name, tool) {
			return true
		}
	}
	return false
}

func extractFramework(name string) string {
	for key, framework := range frameworkMap {
		if strings.Contains(name, key) {
			return framework
		}
	}
	
	// Special cases
	if strings.Contains(name, "hello-world") {
		return "Hello World"
	}
	
	return strings.Title(strings.ReplaceAll(name, "-", " "))
}

func extractLanguage(name string) string {
	// Check direct language references
	for key, language := range languageMap {
		if strings.Contains(name, key) {
			return language
		}
	}
	
	// Framework to language mapping
	frameworkLang := map[string]string{
		"laravel":   "PHP",
		"symfony":   "PHP",
		"django":    "Python",
		"flask":     "Python",
		"rails":     "Ruby",
		"express":   "JavaScript",
		"nextjs":    "JavaScript",
		"nuxt":      "JavaScript",
		"react":     "JavaScript",
		"vue":       "JavaScript",
		"angular":   "TypeScript",
		"phoenix":   "Elixir",
	}
	
	for framework, lang := range frameworkLang {
		if strings.Contains(name, framework) {
			return lang
		}
	}
	
	return "Unknown"
}

func extractTags(name string) []string {
	tags := []string{}
	
	// Add framework tags
	for key := range frameworkMap {
		if strings.Contains(name, key) {
			tags = append(tags, key)
		}
	}
	
	// Add language tags
	for key := range languageMap {
		if strings.Contains(name, key) {
			tags = append(tags, key)
		}
	}
	
	// Add category tags
	if strings.Contains(name, "static") {
		tags = append(tags, "static")
	}
	if strings.Contains(name, "minimal") {
		tags = append(tags, "minimal")
	}
	if strings.Contains(name, "full") || strings.Contains(name, "jetstream") {
		tags = append(tags, "full-stack")
	}
	
	return tags
}

func convertRecipeToPattern(recipe *RecipeInfo) (*Pattern, error) {
	pattern := &Pattern{
		PatternID:      recipe.Name,
		Name:           formatName(recipe.Name),
		Framework:      recipe.Framework,
		Language:       recipe.Language,
		Tags:           recipe.Tags,
		SourceRecipe:   recipe.Name + "-main",
		DeploymentType: "local",
	}

	// Set deployment type
	if recipe.Category == "tool" {
		pattern.DeploymentType = "git"
	}

	// Process import YAML if available
	if recipe.HasImportYml {
		importData, err := processImportYAML(recipe.ImportYmlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to process import YAML: %w", err)
		}
		
		pattern.Services = importData.Services
		
		// Check for buildFromGit
		for _, service := range pattern.Services {
			if _, hasBuildFromGit := service["buildFromGit"]; hasBuildFromGit {
				recipe.HasBuildFromGit = true
				// Remove buildFromGit for framework recipes
				if recipe.Category == "framework" {
					delete(service, "buildFromGit")
				}
			}
		}
	}

	// Process zerops.yml if available
	if recipe.HasZeropsYml {
		zeropsData, err := processZeropsYAML(recipe.ZeropsYmlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to process zerops.yml: %w", err)
		}
		pattern.ZeropsYml = map[string]interface{}{
			"zerops": zeropsData.Zerops,
		}
	}

	// Generate description
	pattern.Description = generateDescription(recipe)

	// Add best practices and common issues based on framework
	pattern.BestPractices = generateBestPractices(recipe.Framework)
	pattern.CommonIssues = generateCommonIssues(recipe.Framework)

	return pattern, nil
}

func processImportYAML(path string) (*ImportYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var importYAML ImportYAML
	if err := yaml.Unmarshal(data, &importYAML); err != nil {
		return nil, err
	}

	return &importYAML, nil
}

func processZeropsYAML(path string) (*ZeropsYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var zeropsYAML ZeropsYAML
	if err := yaml.Unmarshal(data, &zeropsYAML); err != nil {
		return nil, err
	}

	return &zeropsYAML, nil
}

func formatName(name string) string {
	// Remove "recipe-" prefix
	name = strings.TrimPrefix(name, "recipe-")
	
	// Replace hyphens with spaces and title case
	parts := strings.Split(name, "-")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	
	return strings.Join(parts, " ")
}

func generateDescription(recipe *RecipeInfo) string {
	if recipe.Category == "tool" {
		return fmt.Sprintf("%s service for Zerops", formatName(recipe.Name))
	}
	
	desc := fmt.Sprintf("%s application", recipe.Framework)
	if strings.Contains(recipe.Name, "minimal") {
		desc += " - minimal setup"
	} else if strings.Contains(recipe.Name, "full") {
		desc += " - full stack setup"
	}
	
	return desc
}

func generateBestPractices(framework string) []string {
	practices := map[string][]string{
		"Laravel": {
			"Use php-apache@8.3 for Laravel",
			"Cache all configurations in build",
			"Use Valkey for sessions and cache",
			"Enable OPcache for production",
		},
		"Django": {
			"Use gunicorn for production",
			"Collect static files during build",
			"Use PostgreSQL for database",
			"Set DEBUG=False for production",
		},
		"Next.js": {
			"Use standalone build for smaller images",
			"Enable ISR for dynamic content",
			"Use environment variables for API URLs",
			"Cache node_modules between builds",
		},
	}
	
	if p, ok := practices[framework]; ok {
		return p
	}
	
	return []string{}
}

func generateCommonIssues(framework string) []map[string]string {
	issues := map[string][]map[string]string{
		"Laravel": {
			{
				"issue":    "Migration fails",
				"solution": "Add --isolated flag to migration command",
			},
			{
				"issue":    "Storage not writable",
				"solution": "Storage directory is automatically writable in Zerops",
			},
		},
		"Django": {
			{
				"issue":    "Static files not found",
				"solution": "Run collectstatic in buildCommands",
			},
			{
				"issue":    "Database migrations fail",
				"solution": "Ensure database service is running first",
			},
		},
	}
	
	if i, ok := issues[framework]; ok {
		return i
	}
	
	return []map[string]string{}
}

func savePattern(pattern *Pattern, path string) error {
	data, err := json.MarshalIndent(pattern, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func createPatternIndex(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var patterns []map[string]string
	
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") && entry.Name() != "index.json" {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			
			var pattern Pattern
			if err := json.Unmarshal(data, &pattern); err != nil {
				continue
			}
			
			patterns = append(patterns, map[string]string{
				"id":          pattern.PatternID,
				"name":        pattern.Name,
				"framework":   pattern.Framework,
				"language":    pattern.Language,
				"category":    getCategory(pattern),
			})
		}
	}

	indexData, err := json.MarshalIndent(map[string]interface{}{
		"patterns": patterns,
		"total":    len(patterns),
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "index.json"), indexData, 0644)
}

func getCategory(pattern Pattern) string {
	if pattern.DeploymentType == "git" {
		return "tool"
	}
	return "framework"
}