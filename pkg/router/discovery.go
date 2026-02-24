package router

import (
	"database/sql"
	"fmt"
	"log"
	"sort"

	"github.com/gin-gonic/gin"
	appmodule "github.com/jgirmay/GAIA_GO/internal/app"
	"github.com/jgirmay/GAIA_GO/internal/session"
	mathapp "github.com/jgirmay/GAIA_GO/pkg/apps/math"
	pianoapp "github.com/jgirmay/GAIA_GO/pkg/apps/piano"
	readingapp "github.com/jgirmay/GAIA_GO/pkg/apps/reading"
	typingapp "github.com/jgirmay/GAIA_GO/pkg/apps/typing"
)

// DiscoveredApps holds all discovered applications and their metadata
type DiscoveredApps struct {
	Apps     []appmodule.AppRegistry
	Metadata map[string]*appmodule.AppMetadata
	LoadOrder []string // Topologically sorted app names
}

// DiscoverApps scans and loads all registered GAIA apps
func DiscoverApps(db *sql.DB, sessionManager *session.Manager) (*DiscoveredApps, error) {
	discovered := &DiscoveredApps{
		Apps:      make([]appmodule.AppRegistry, 0),
		Metadata:  make(map[string]*appmodule.AppMetadata),
		LoadOrder: make([]string, 0),
	}

	// Register all known apps
	appFactories := []func(db *sql.DB, sm *session.Manager) appmodule.AppRegistry{
		createMathApp,
		createTypingApp,
		createReadingApp,
		createPianoApp,
	}

	// Create app context for dependency injection
	appCtx := &appmodule.AppContext{
		DB:             db,
		SessionManager: sessionManager,
		Config: &appmodule.AppConfig{
			Environment: "development",
			Settings:    make(map[string]interface{}),
		},
	}

	// Load all apps
	appMap := make(map[string]appmodule.AppRegistry)
	for _, factory := range appFactories {
		app := factory(db, sessionManager)
		if app == nil {
			continue
		}

		// Initialize app
		if err := app.Initialize(appCtx); err != nil {
			log.Printf("Warning: Failed to initialize app %s: %v\n", app.Name(), err)
			continue
		}

		appMap[app.Name()] = app
		discovered.Metadata[app.Name()] = &appmodule.AppMetadata{
			Name:        app.Name(),
			Description: app.Description(),
			Version:     app.Version(),
			BasePath:    app.BasePath(),
			Routes:      app.RouteGroups(),
			Status:      "initialized",
		}
	}

	// Validate dependencies
	if err := ValidateDependencies(appMap); err != nil {
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	// Resolve load order using topological sort
	loadOrder, err := ResolveDependencies(appMap)
	if err != nil {
		return nil, fmt.Errorf("dependency resolution failed: %w", err)
	}

	discovered.LoadOrder = loadOrder
	discovered.Apps = make([]appmodule.AppRegistry, len(loadOrder))
	for i, appName := range loadOrder {
		discovered.Apps[i] = appMap[appName]
	}

	log.Printf("Discovered %d GAIA apps in load order: %v\n", len(discovered.Apps), loadOrder)
	return discovered, nil
}

// ValidateDependencies checks that all declared app-to-app dependencies exist
// System dependencies (database, session_manager, etc.) are provided by the framework
func ValidateDependencies(appMap map[string]appmodule.AppRegistry) error {
	// System dependencies provided by the framework
	systemDeps := map[string]bool{
		"database":        true,
		"session_manager": true,
	}

	for appName, app := range appMap {
		for _, dep := range app.Dependencies() {
			if dep.Required {
				// Only validate app-to-app dependencies, not system dependencies
				if !systemDeps[dep.Name] {
					if _, exists := appMap[dep.Name]; !exists {
						return fmt.Errorf("app %s requires missing app dependency: %s", appName, dep.Name)
					}
				}
			}
		}
	}
	return nil
}

// ResolveDependencies performs topological sort on app dependencies
// Returns apps in load order (dependencies before dependents)
func ResolveDependencies(appMap map[string]appmodule.AppRegistry) ([]string, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all apps
	for name := range appMap {
		graph[name] = make([]string, 0)
		inDegree[name] = 0
	}

	// Build edges
	for appName, app := range appMap {
		for _, dep := range app.Dependencies() {
			if _, exists := appMap[dep.Name]; exists {
				// Add edge from dependency to dependent
				graph[dep.Name] = append(graph[dep.Name], appName)
				inDegree[appName]++
			}
		}
	}

	// Kahn's algorithm for topological sort
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]string, 0)
	for len(queue) > 0 {
		// Process node with no dependencies
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree of dependent apps
		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(result) != len(appMap) {
		return nil, fmt.Errorf("circular dependency detected among apps")
	}

	// Sort alphabetically for deterministic ordering among same level
	sort.Strings(result)
	return result, nil
}

// GetAppMetadata returns metadata for a specific app
func (d *DiscoveredApps) GetAppMetadata(name string) *appmodule.AppMetadata {
	return d.Metadata[name]
}

// GetApp returns the app instance by name
func (d *DiscoveredApps) GetApp(name string) appmodule.AppRegistry {
	for _, app := range d.Apps {
		if app.Name() == name {
			return app
		}
	}
	return nil
}

// ============================================================================
// APP FACTORIES - Create app instances
// ============================================================================

// createMathApp creates a Math app registry implementation
func createMathApp(db *sql.DB, sm *session.Manager) appmodule.AppRegistry {
	return &mathAppRegistry{
		app: mathapp.NewMathApp(db),
		db:  db,
	}
}

// createTypingApp creates a Typing app registry implementation
func createTypingApp(db *sql.DB, sm *session.Manager) appmodule.AppRegistry {
	return &typingAppRegistry{
		app: typingapp.NewTypingApp(db),
		db:  db,
	}
}

// createReadingApp creates a Reading app registry implementation
func createReadingApp(db *sql.DB, sm *session.Manager) appmodule.AppRegistry {
	return &readingAppRegistry{
		app: readingapp.NewReadingApp(db),
		db:  db,
	}
}

// createPianoApp creates a Piano app registry implementation
func createPianoApp(db *sql.DB, sm *session.Manager) appmodule.AppRegistry {
	return &pianoAppRegistry{
		app: pianoapp.NewPianoApp(db),
		db:  db,
	}
}

// ============================================================================
// MATH APP REGISTRY
// ============================================================================

type mathAppRegistry struct {
	app *mathapp.MathApp
	db  *sql.DB
}

func (r *mathAppRegistry) Name() string {
	return "math"
}

func (r *mathAppRegistry) Description() string {
	return "Math practice application with problem generation and progress tracking"
}

func (r *mathAppRegistry) Version() string {
	return "1.0.0"
}

func (r *mathAppRegistry) BasePath() string {
	return "/api/math"
}

func (r *mathAppRegistry) RouteGroups() []appmodule.RouteGroup {
	return []appmodule.RouteGroup{
		{
			Path:        "/problems",
			Description: "Problem generation and answer checking",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/generate", Description: "Generate a new math problem"},
				{Method: "POST", Path: "/check", Description: "Check answer to a problem"},
				{Method: "POST", Path: "/check-speech", Description: "Check spoken answer"},
			},
		},
		{
			Path:        "/sessions",
			Description: "Session management",
			Routes: []appmodule.RouteInfo{
				{Method: "POST", Path: "/save", Description: "Save session results"},
			},
		},
		{
			Path:        "/stats",
			Description: "User statistics and leaderboards",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/stats", Description: "Get user math statistics"},
				{Method: "GET", Path: "/leaderboard", Description: "Get math leaderboard"},
			},
		},
		{
			Path:        "/audio",
			Description: "Audio transcription",
			Routes: []appmodule.RouteInfo{
				{Method: "POST", Path: "/transcribe", Description: "Transcribe audio to text"},
			},
		},
	}
}

func (r *mathAppRegistry) Dependencies() []appmodule.Dependency {
	return []appmodule.Dependency{
		{Name: "database", Type: "*sql.DB", Required: true},
	}
}

func (r *mathAppRegistry) Initialize(ctx *appmodule.AppContext) error {
	return r.app.InitDB()
}

func (r *mathAppRegistry) RegisterHandlers(group *gin.RouterGroup) error {
	// Handlers will be registered by the handler registration function
	// This is called by the auto-registry system
	return nil
}

// ============================================================================
// TYPING APP REGISTRY
// ============================================================================

type typingAppRegistry struct {
	app *typingapp.TypingApp
	db  *sql.DB
}

func (r *typingAppRegistry) Name() string {
	return "typing"
}

func (r *typingAppRegistry) Description() string {
	return "Typing practice application with WPM tracking and racing modes"
}

func (r *typingAppRegistry) Version() string {
	return "1.0.0"
}

func (r *typingAppRegistry) BasePath() string {
	return "/api/typing"
}

func (r *typingAppRegistry) RouteGroups() []appmodule.RouteGroup {
	return []appmodule.RouteGroup{
		{
			Path:        "/tests",
			Description: "Typing test management",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/text", Description: "Get text for typing test"},
				{Method: "POST", Path: "/save-result", Description: "Save typing test result"},
			},
		},
		{
			Path:        "/stats",
			Description: "Statistics and leaderboards",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/stats", Description: "Get user typing statistics"},
				{Method: "GET", Path: "/leaderboard", Description: "Get typing leaderboard"},
			},
		},
		{
			Path:        "/races",
			Description: "Racing mode",
			Routes: []appmodule.RouteInfo{
				{Method: "POST", Path: "/start", Description: "Start a new race"},
				{Method: "POST", Path: "/finish", Description: "Finish a race"},
				{Method: "GET", Path: "/leaderboard", Description: "Get race leaderboard"},
			},
		},
	}
}

func (r *typingAppRegistry) Dependencies() []appmodule.Dependency {
	return []appmodule.Dependency{
		{Name: "database", Type: "*sql.DB", Required: true},
	}
}

func (r *typingAppRegistry) Initialize(ctx *appmodule.AppContext) error {
	return r.app.InitDB()
}

func (r *typingAppRegistry) RegisterHandlers(group *gin.RouterGroup) error {
	// Handlers will be registered by the handler registration function
	return nil
}

// ============================================================================
// READING APP REGISTRY
// ============================================================================

type readingAppRegistry struct {
	app *readingapp.ReadingApp
	db  *sql.DB
}

func (r *readingAppRegistry) Name() string {
	return "reading"
}

func (r *readingAppRegistry) Description() string {
	return "Reading comprehension application with word mastery tracking"
}

func (r *readingAppRegistry) Version() string {
	return "1.0.0"
}

func (r *readingAppRegistry) BasePath() string {
	return "/api/reading"
}

func (r *readingAppRegistry) RouteGroups() []appmodule.RouteGroup {
	return []appmodule.RouteGroup{
		{
			Path:        "/words",
			Description: "Word management",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/", Description: "Get words for reading"},
				{Method: "GET", Path: "/level/:level", Description: "Get words by level"},
				{Method: "POST", Path: "/correct", Description: "Mark word as correct"},
				{Method: "POST", Path: "/error", Description: "Mark word as incorrect"},
			},
		},
		{
			Path:        "/stats",
			Description: "User progress and statistics",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/progress", Description: "Get user reading progress"},
				{Method: "GET", Path: "/mastery", Description: "Get word mastery data"},
				{Method: "GET", Path: "/leaderboard", Description: "Get reading leaderboard"},
			},
		},
	}
}

func (r *readingAppRegistry) Dependencies() []appmodule.Dependency {
	return []appmodule.Dependency{
		{Name: "database", Type: "*sql.DB", Required: true},
	}
}

func (r *readingAppRegistry) Initialize(ctx *appmodule.AppContext) error {
	return r.app.InitDB()
}

func (r *readingAppRegistry) RegisterHandlers(group *gin.RouterGroup) error {
	// Handlers will be registered by the handler registration function
	return nil
}

// ============================================================================
// PIANO APP REGISTRY
// ============================================================================

type pianoAppRegistry struct {
	app *pianoapp.PianoApp
	db  *sql.DB
}

func (r *pianoAppRegistry) Name() string {
	return "piano"
}

func (r *pianoAppRegistry) Description() string {
	return "Piano learning application with lessons, practice, and performance tracking"
}

func (r *pianoAppRegistry) Version() string {
	return "1.0.0"
}

func (r *pianoAppRegistry) BasePath() string {
	return "/api/piano"
}

func (r *pianoAppRegistry) RouteGroups() []appmodule.RouteGroup {
	return []appmodule.RouteGroup{
		{
			Path:        "/practice",
			Description: "Practice and lesson management",
			Routes: []appmodule.RouteInfo{
				{Method: "POST", Path: "/save-session", Description: "Save practice session"},
				{Method: "GET", Path: "/stats", Description: "Get practice statistics"},
			},
		},
		{
			Path:        "/warmups",
			Description: "Warm-up exercises",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/", Description: "Get available warmups"},
				{Method: "POST", Path: "/start", Description: "Start a warmup"},
				{Method: "POST", Path: "/complete", Description: "Complete a warmup"},
			},
		},
		{
			Path:        "/achievements",
			Description: "Badges and goals",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/badges", Description: "Get user badges"},
				{Method: "GET", Path: "/goals", Description: "Get user goals"},
			},
		},
		{
			Path:        "/leaderboard",
			Description: "Performance tracking",
			Routes: []appmodule.RouteInfo{
				{Method: "GET", Path: "/", Description: "Get piano leaderboard"},
				{Method: "GET", Path: "/streak", Description: "Get user streak"},
			},
		},
	}
}

func (r *pianoAppRegistry) Dependencies() []appmodule.Dependency {
	return []appmodule.Dependency{
		{Name: "database", Type: "*sql.DB", Required: true},
	}
}

func (r *pianoAppRegistry) Initialize(ctx *appmodule.AppContext) error {
	return r.app.InitDB()
}

func (r *pianoAppRegistry) RegisterHandlers(group *gin.RouterGroup) error {
	// Handlers will be registered by the handler registration function
	return nil
}
