package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: extract-knowledge <runtime>")
		os.Exit(1)
	}

	runtime := os.Args[1]

	// Read overview.mdx
	docPath := fmt.Sprintf("docs/zerops-documentation/content/%s/overview.mdx", runtime)
	content, err := ioutil.ReadFile(docPath)
	if err != nil {
		fmt.Printf("Error reading doc: %v\n", err)
		os.Exit(1)
	}

	// Extract versions (handle @version patterns and type: go@version)
	versionRegex := regexp.MustCompile(`(?:@|go@)(\d+(?:\.\d+)?|latest)`)
	matches := versionRegex.FindAllStringSubmatch(string(content), -1)

	versions := []string{}
	seenVersions := make(map[string]bool)
	for _, match := range matches {
		version := match[1]
		if !seenVersions[version] {
			versions = append(versions, version)
			seenVersions[version] = true
		}
	}

	// Extract build commands from code blocks
	buildCommands := extractBuildCommands(string(content))
	
	// Extract deploy files patterns
	deployFiles := extractDeployFiles(string(content))

	// Create knowledge structure
	knowledge := map[string]interface{}{
		"runtime":     runtime,
		"displayName": getDisplayName(runtime),
		"versions": map[string]interface{}{
			"available": versions,
			"default":   getDefaultVersion(versions),
		},
		"buildConfig": map[string]interface{}{
			"buildCommands": buildCommands,
			"deployFiles":   deployFiles,
		},
		"runConfig": map[string]interface{}{
			"startCommand": getStartCommand(runtime),
			"port":         getDefaultPort(runtime),
		},
	}

	// Save to file
	output, _ := json.MarshalIndent(knowledge, "", "  ")
	outputPath := fmt.Sprintf("internal/knowledge/data/runtimes/%s.json", runtime)
	
	// Create directory if needed
	os.MkdirAll(filepath.Dir(outputPath), 0755)
	
	err = ioutil.WriteFile(outputPath, output, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Extracted knowledge for %s\n", runtime)
}

func getDisplayName(runtime string) string {
	names := map[string]string{
		"nodejs": "Node.js",
		"python": "Python",
		"php":    "PHP",
		"go":     "Go",
		"java":   "Java",
		"rust":   "Rust",
		"dotnet": ".NET",
		"bun":    "Bun",
		"deno":   "Deno",
		"elixir": "Elixir",
		"gleam":  "Gleam",
	}
	if name, ok := names[runtime]; ok {
		return name
	}
	return strings.Title(runtime)
}

func getDefaultVersion(versions []string) string {
	if len(versions) > 0 {
		return versions[0]
	}
	return ""
}

func extractBuildCommands(content string) []string {
	// Look for buildCommands in code blocks
	buildCmdRegex := regexp.MustCompile(`buildCommands:\s*\n\s*-\s*(.+)`)
	matches := buildCmdRegex.FindAllStringSubmatch(content, -1)
	
	commands := []string{}
	for _, match := range matches {
		cmd := strings.TrimSpace(match[1])
		if !contains(commands, cmd) {
			commands = append(commands, cmd)
		}
	}
	
	// Add runtime-specific defaults if none found
	if len(commands) == 0 {
		commands = getDefaultBuildCommands(filepath.Base(content))
	}
	
	return commands
}

func extractDeployFiles(content string) []string {
	// Look for deployFiles in code blocks
	deployRegex := regexp.MustCompile(`deployFiles:\s*\n(?:\s*-\s*(.+)\n)*`)
	matches := deployRegex.FindAllStringSubmatch(content, -1)
	
	files := []string{}
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			file := strings.TrimSpace(match[1])
			if !contains(files, file) {
				files = append(files, file)
			}
		}
	}
	
	if len(files) == 0 {
		files = []string{"."}
	}
	
	return files
}

func getDefaultBuildCommands(runtime string) []string {
	defaults := map[string][]string{
		"nodejs": {"npm install", "npm run build"},
		"python": {"pip install -r requirements.txt"},
		"php":    {"composer install --optimize-autoloader --no-dev"},
		"go":     {"go mod download", "go build"},
		"java":   {"mvn clean package"},
		"rust":   {"cargo build --release"},
	}
	if cmds, ok := defaults[runtime]; ok {
		return cmds
	}
	return []string{}
}

func getStartCommand(runtime string) string {
	defaults := map[string]string{
		"nodejs": "npm start",
		"python": "python app.py",
		"php":    "php-fpm",
		"go":     "./app",
		"java":   "java -jar app.jar",
		"rust":   "./target/release/app",
	}
	if cmd, ok := defaults[runtime]; ok {
		return cmd
	}
	return ""
}

func getDefaultPort(runtime string) int {
	ports := map[string]int{
		"nodejs": 3000,
		"python": 8000,
		"php":    9000,
		"go":     8080,
		"java":   8080,
		"rust":   8080,
	}
	if port, ok := ports[runtime]; ok {
		return port
	}
	return 8080
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}