package app

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/session"
)

// AppRegistry defines the contract all GAIA apps must implement
// This enables automatic discovery, validation, and registration
type AppRegistry interface {
	// Identity methods
	Name() string        // Unique app identifier (e.g., "math", "typing")
	Description() string // Human-readable description
	Version() string     // Semantic version (e.g., "1.0.0")

	// Route information
	BasePath() string       // Root path for this app (e.g., "/api/math")
	RouteGroups() []RouteGroup // Organized route information

	// Dependency declaration
	Dependencies() []Dependency // What this app requires

	// Initialization
	Initialize(ctx *AppContext) error // Setup phase

	// Handler registration
	RegisterHandlers(router *gin.RouterGroup) error // Bind HTTP handlers
}

// Dependency describes a dependency this app requires
type Dependency struct {
	Name     string // Unique name (e.g., "database", "session_manager")
	Type     string // Go type name for documentation
	Required bool   // Is this dependency required?
}

// RouteGroup organizes related routes for documentation
type RouteGroup struct {
	Path        string      // Base path for this group (e.g., "/problem")
	Description string      // What this group does
	Routes      []RouteInfo // Individual routes in this group
}

// RouteInfo describes a single HTTP route
type RouteInfo struct {
	Method      string // HTTP method (GET, POST, PUT, DELETE)
	Path        string // Route path relative to group
	Description string // What this endpoint does
}

// AppContext provides dependency-injected services to apps
type AppContext struct {
	DB             *sql.DB
	SessionManager *session.Manager
	Config         *AppConfig
}

// AppConfig holds configuration accessible to all apps
type AppConfig struct {
	Environment string                 // "development", "staging", "production"
	Settings    map[string]interface{} // App-specific configuration
}

// AppMetadata represents discovered app information
type AppMetadata struct {
	Name        string
	Description string
	Version     string
	BasePath    string
	Routes      []RouteGroup
	Status      string // "initialized", "registered", "error"
}
