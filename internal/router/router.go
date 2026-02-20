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

	// Initialize app handlers with database connection
	initializeAppHandlers(db)

	// Health check endpoint (public)
	r.Get("/health", healthHandler)

	// Static file serving
	fileServer := http.FileServer(http.Dir(cfg.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Typing app routes
		r.Route("/typing", func(r chi.Router) {
			r.Get("/", typing.ListLessons)
			r.Post("/progress", typing.SaveProgress)
		})

		// Math app routes
		r.Route("/math", func(r chi.Router) {
			r.Get("/", math.ListProblems)
			r.Post("/progress", math.SaveProgress)
		})

		// Reading app routes (will be updated after handler setup)
		r.Route("/reading", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"status": "reading api"})
			})
		})

		// Piano app routes
		r.Route("/piano", func(r chi.Router) {
			r.Get("/", piano.ListSongs)
			r.Post("/progress", piano.SaveProgress)
		})

		// Dashboard routes
		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/stats", dashboard.GetStats)
		})
	})

	// App-specific routes (with templates)
	r.Route("/typing", func(r chi.Router) {
		r.Get("/", typing.IndexHandler)
	})

	r.Route("/math", func(r chi.Router) {
		r.Get("/", math.IndexHandler)
	})

	r.Route("/reading", func(r chi.Router) {
		r.Get("/", reading.IndexHandler)
	})

	r.Route("/piano", func(r chi.Router) {
		r.Get("/", piano.IndexHandler)
	})

	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/", dashboard.IndexHandler)
	})

	// Root redirect
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	return r
}

// initializeAppHandlers sets up handlers with database connections
func initializeAppHandlers(db *database.Pool) {
	// Initialize math app handler
	mathRepo := math.NewRepository(db.DB)
	mathService := math.NewService(mathRepo)
	mathHandler := math.NewHandler(mathService)
	math.SetGlobalHandler(mathHandler)

	// TODO: Initialize other app handlers similarly
	// - typing
	// - reading
	// - piano
	// - dashboard
}

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
