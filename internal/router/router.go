package router

import (
	"encoding/json"
	"net/http"
	"path/filepath"
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

	// Global static file serving
	fileServer := http.FileServer(http.Dir(cfg.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// ============================================================
	// Math App Routes
	// ============================================================
	r.Route("/math", func(r chi.Router) {
		// Math app static files (CSS, JS, service worker, etc.)
		mathStaticDir := filepath.Join(cfg.StaticDir, "math")
		mathFileServer := http.FileServer(http.Dir(mathStaticDir))
		r.Handle("/static/*", http.StripPrefix("/math/static", mathFileServer))

		// Math app service worker (explicit route for offline support)
		r.Get("/service-worker.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Header().Set("Service-Worker-Allowed", "/")
			http.ServeFile(w, r, filepath.Join(mathStaticDir, "service-worker.js"))
		})

		// Math app index.html (main page)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, filepath.Join(cfg.StaticDir, "math", "index.html"))
		})

		// Mount math API routes
		r.Mount("/api", math.NewRouter(db.DB).Routes())
	})

	// ============================================================
	// Reading App Routes
	// ============================================================
	r.Route("/reading", func(r chi.Router) {
		// Reading app static files (CSS, JS, service worker, etc.)
		readingStaticDir := filepath.Join(cfg.StaticDir, "reading")
		readingFileServer := http.FileServer(http.Dir(readingStaticDir))
		r.Handle("/static/*", http.StripPrefix("/reading/static", readingFileServer))

		// Reading app service worker (explicit route for offline support)
		r.Get("/service-worker.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Header().Set("Service-Worker-Allowed", "/")
			http.ServeFile(w, r, filepath.Join(readingStaticDir, "service-worker.js"))
		})

		// Reading app index.html (main page)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, filepath.Join(cfg.StaticDir, "reading", "index.html"))
		})

		// Mount reading API routes
		r.Mount("/api", reading.NewRouter(db.DB).Routes())
	})

	// ============================================================
	// Piano App Routes
	// ============================================================
	r.Mount("/piano", piano.NewRouter(db.DB).Routes())

	// ============================================================
	// Typing App Routes
	// ============================================================
	r.Mount("/typing", typing.NewRouter(db.DB).Routes())

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
