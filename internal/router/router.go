package router

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/jgirmay/unified-go/internal/config"
	"github.com/jgirmay/unified-go/internal/database"
	"github.com/jgirmay/unified-go/internal/middleware"
	"github.com/jgirmay/unified-go/pkg/dashboard"
	"github.com/jgirmay/unified-go/pkg/math"
	"github.com/jgirmay/unified-go/pkg/piano"
	"github.com/jgirmay/unified-go/pkg/reading"
	"github.com/jgirmay/unified-go/pkg/typing"
)

var serverStartTime = time.Now()

// Setup configures and returns the HTTP router
func Setup(cfg *config.Config, db *database.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Apply global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Compress(5))

	// CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware(cfg.CORSOrigins)
	r.Use(corsMiddleware.Handler)

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.SessionSecret, cfg.SessionName)
	r.Use(authMiddleware.Handler)

	// Health check endpoint (public)
	r.Get("/health", healthHandler)

	// Static file serving
	fileServer := http.FileServer(http.Dir(cfg.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Mount educational app routers
	r.Mount("/reading", reading.NewRouter(db.DB).Routes())
	r.Mount("/piano", piano.NewRouter(db.DB).Routes())
	r.Mount("/typing", typing.NewRouter(db.DB).Routes())
	r.Mount("/math", math.NewRouter(db.DB).Routes())

	// Dashboard routes
	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/", dashboard.IndexHandler)
	})

	// Root redirect
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	return r
}

// Note: App routers are now initialized directly in Setup() via NewRouter calls
// Legacy initializeAppHandlers function removed - all apps use router pattern

// healthHandler returns server health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":      "healthy",
		"go_version":  runtime.Version(),
		"uptime":      time.Since(serverStartTime).String(),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"goroutines":  runtime.NumGoroutine(),
		"environment": "development",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
