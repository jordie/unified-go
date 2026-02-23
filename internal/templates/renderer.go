package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Renderer handles template rendering for all education apps
type Renderer struct {
	templates map[string]*template.Template
	funcs     template.FuncMap
	mutex     sync.RWMutex
	baseDir   string
}

// NewRenderer creates a new template renderer
func NewRenderer(baseDir string) *Renderer {
	r := &Renderer{
		templates: make(map[string]*template.Template),
		baseDir:   baseDir,
		funcs:     createFuncMap(),
	}
	return r
}

// LoadTemplate loads a template by app and name
func (r *Renderer) LoadTemplate(app, name string) (*template.Template, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := fmt.Sprintf("%s/%s", app, name)

	// Check cache
	if tmpl, exists := r.templates[key]; exists {
		return tmpl, nil
	}

	// Load from filesystem
	templatePath := filepath.Join(r.baseDir, "templates", app, name+".html")

	tmpl, err := template.New(name).Funcs(r.funcs).ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Cache the template
	r.templates[key] = tmpl
	return tmpl, nil
}

// LoadAppTemplates loads all templates for a specific app
func (r *Renderer) LoadAppTemplates(app string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	appDir := filepath.Join(r.baseDir, "templates", app)
	entries, err := os.ReadDir(appDir)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".html" {
			name := entry.Name()[:len(entry.Name())-5] // Remove .html

			templatePath := filepath.Join(appDir, entry.Name())
			tmpl, err := template.New(name).Funcs(r.funcs).ParseFiles(templatePath)
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
			}

			key := fmt.Sprintf("%s/%s", app, name)
			r.templates[key] = tmpl
		}
	}

	return nil
}

// Render renders a template with the given data
func (r *Renderer) Render(app, templateName string, data interface{}) (string, error) {
	tmpl, err := r.LoadTemplate(app, templateName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ClearCache clears the template cache
func (r *Renderer) ClearCache() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.templates = make(map[string]*template.Template)
}

// createFuncMap returns a map of template functions available to all templates
func createFuncMap() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,

		// Math functions
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Formatting functions
		"percent": func(value float64) string {
			return fmt.Sprintf("%.1f%%", value)
		},
		"formatTime": func(seconds int) string {
			hours := seconds / 3600
			minutes := (seconds % 3600) / 60
			secs := seconds % 60
			return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
		},
		"formatDate": func(t interface{}) string {
			// Handle different types
			if str, ok := t.(string); ok {
				return str
			}
			return fmt.Sprintf("%v", t)
		},

		// Data functions
		"isEmpty": func(val interface{}) bool {
			if val == nil {
				return true
			}
			switch v := val.(type) {
			case string:
				return v == ""
			case []interface{}:
				return len(v) == 0
			case map[string]interface{}:
				return len(v) == 0
			default:
				return false
			}
		},
		"getIndex": func(arr interface{}, idx int) interface{} {
			// Helper to get array element
			return nil // Simplified
		},

		// Status functions
		"statusClass": func(status string) string {
			switch status {
			case "success":
				return "status-success"
			case "error":
				return "status-error"
			case "warning":
				return "status-warning"
			default:
				return "status-info"
			}
		},

		// Badge/Level functions
		"getLevelColor": func(level int) string {
			switch {
			case level <= 1:
				return "bronze"
			case level <= 5:
				return "silver"
			case level <= 10:
				return "gold"
			default:
				return "platinum"
			}
		},
		"getLevelBadge": func(level int) string {
			badges := map[int]string{
				1: "🥉",
				5: "🥈",
				10: "🥇",
				20: "👑",
			}
			if badge, ok := badges[level]; ok {
				return badge
			}
			return "⭐"
		},
	}
}
