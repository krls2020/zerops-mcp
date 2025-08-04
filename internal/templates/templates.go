package templates

// Template represents a configuration template
type Template struct {
	Name        string
	Description string
	Config      map[string]interface{}
}

var templates = map[string]Template{
	"python": {
		Name:        "Python",
		Description: "Basic Python application template",
		Config: map[string]interface{}{
			"zerops": []interface{}{
				map[string]interface{}{
					"setup": "app",
					"build": map[string]interface{}{
						"base": "python@3.12",
						"buildCommands": []string{
							"python3 -m pip install --upgrade pip",
							"python3 -m pip install -r requirements.txt",
						},
						"deployFiles": []string{"."},
						"addToRunPrepare": []string{"requirements.txt"},
					},
					"run": map[string]interface{}{
						"base": "python@3.12",
						"prepareCommands": []string{
							"python3 -m pip install --ignore-installed -r requirements.txt",
						},
						"start": "python3 app.py",
						"ports": []map[string]interface{}{
							{"port": 8000, "httpSupport": true},
						},
					},
				},
			},
		},
	},
	"nodejs": {
		Name:        "Node.js",
		Description: "Node.js application template",
		Config: map[string]interface{}{
			"zerops": []interface{}{
				map[string]interface{}{
					"setup": "app",
					"build": map[string]interface{}{
						"base": "nodejs@20",
						"buildCommands": []string{
							"npm ci",
						},
						"deployFiles": []string{"."},
						"addToRunPrepare": []string{"package.json", "package-lock.json"},
					},
					"run": map[string]interface{}{
						"base": "nodejs@20",
						"prepareCommands": []string{
							"npm ci --production",
						},
						"start": "npm start",
						"ports": []map[string]interface{}{
							{"port": 3000, "httpSupport": true},
						},
					},
				},
			},
		},
	},
	"php": {
		Name:        "PHP",
		Description: "PHP application template",
		Config: map[string]interface{}{
			"zerops": []interface{}{
				map[string]interface{}{
					"setup": "app",
					"build": map[string]interface{}{
						"base": "php-apache@8.3",
						"buildCommands": []string{
							"composer install --optimize-autoloader",
						},
						"deployFiles": []string{"."},
						"addToRunPrepare": []string{"composer.json", "composer.lock"},
					},
					"run": map[string]interface{}{
						"base": "php-apache@8.3",
						"prepareCommands": []string{
							"composer install --no-dev --optimize-autoloader",
						},
						"documentRoot": "public",
					},
				},
			},
		},
	},
}

// GetTemplate returns a template by name
func GetTemplate(name string) (Template, bool) {
	tmpl, exists := templates[name]
	return tmpl, exists
}

// ListTemplates returns all available templates
func ListTemplates() map[string]Template {
	return templates
}