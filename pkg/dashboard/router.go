package dashboard

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router handles dashboard routes and API endpoints
type Router struct {
	router  *chi.Mux
	service *Service
}

// NewRouter creates and configures the dashboard router
func NewRouter(db *sql.DB) *Router {
	// Create service with unified repository
	service := NewService(nil)

	r := &Router{
		router:  chi.NewRouter(),
		service: service,
	}

	// Setup middleware
	r.router.Use(middleware.Logger)
	r.router.Use(middleware.Recoverer)

	// Setup routes
	r.setupRoutes()

	return r
}

// setupRoutes configures all dashboard routes
func (r *Router) setupRoutes() {
	// UI Routes
	r.router.Get("/", r.indexHandler)
	r.router.Get("/unified", r.unifiedDashboardHandler)

	// API Routes
	r.router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Get("/stats", r.getSystemStats)
		apiRouter.Route("/users/{userID}", func(userRouter chi.Router) {
			userRouter.Get("/profile", r.getUserProfile)
			userRouter.Get("/analytics", r.getUserAnalytics)
			userRouter.Get("/sessions", r.getUserSessions)
		})
		apiRouter.Route("/leaderboard", func(lbRouter chi.Router) {
			lbRouter.Get("/{category}", r.getLeaderboard)
			lbRouter.Get("/", r.listLeaderboards)
		})
		apiRouter.Get("/trends/{userID}", r.getTrends)
		apiRouter.Get("/overview/{userID}", r.getDashboardOverview)
		apiRouter.Get("/recommendations/{userID}", r.getRecommendations)
	})
}

// Handler returns the configured router
func (r *Router) Handler() http.Handler {
	return r.router
}

// UI Handlers

// indexHandler serves the main dashboard landing page
func (r *Router) indexHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Unified Educational Platform - Dashboard</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 50px auto;
            padding: 20px;
            background: #f5f5f5;
        }
        h1 { color: #333; text-align: center; }
        .apps-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .app-card {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
            transition: transform 0.2s;
        }
        .app-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }
        .app-card h2 { color: #2c3e50; margin-bottom: 10px; }
        .app-card p { color: #7f8c8d; margin-bottom: 20px; }
        .app-card a {
            display: inline-block;
            padding: 10px 20px;
            background: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            transition: background 0.2s;
            margin: 5px;
        }
        .app-card a:hover { background: #2980b9; }
        .status {
            background: #e8f5e9;
            color: #2e7d32;
            padding: 15px;
            border-radius: 5px;
            text-align: center;
            margin-bottom: 30px;
        }
        .nav-links {
            text-align: center;
            margin-top: 30px;
        }
        .nav-links a {
            margin: 0 10px;
            color: #3498db;
            text-decoration: none;
        }
        .nav-links a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>Unified Educational Platform</h1>
    <div class="status">
        <strong>Phase 7 In Progress:</strong> Dashboard aggregation and analytics!
    </div>
    <div class="apps-grid">
        <div class="app-card">
            <h2>📝 Typing</h2>
            <p>Practice typing skills with interactive lessons</p>
            <a href="/typing">Launch App</a>
        </div>
        <div class="app-card">
            <h2>🔢 Math</h2>
            <p>Solve math problems and improve skills</p>
            <a href="/math">Launch App</a>
        </div>
        <div class="app-card">
            <h2>📚 Reading</h2>
            <p>Read books and test comprehension</p>
            <a href="/reading">Launch App</a>
        </div>
        <div class="app-card">
            <h2>🎹 Piano</h2>
            <p>Learn to play piano with guided lessons</p>
            <a href="/piano">Launch App</a>
        </div>
    </div>
    <div class="nav-links">
        <a href="/dashboard/unified">📊 Unified Dashboard</a>
        <a href="/api/stats">📈 API Stats</a>
        <a href="/health">🏥 Health Check</a>
    </div>
</body>
</html>
	`))
}

// unifiedDashboardHandler serves the unified dashboard page
func (r *Router) unifiedDashboardHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Unified Dashboard - Educational Platform</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            overflow: hidden;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        header h1 { font-size: 2em; margin-bottom: 10px; }
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            padding: 30px;
        }
        .card {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .card h3 {
            color: #333;
            margin-bottom: 15px;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        .stat-value {
            font-size: 2em;
            color: #667eea;
            font-weight: bold;
            margin: 15px 0;
        }
        .stat-label {
            color: #666;
            font-size: 0.9em;
        }
        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e0e0e0;
            border-radius: 4px;
            overflow: hidden;
            margin: 10px 0;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
            transition: width 0.3s ease;
        }
        .api-link {
            display: inline-block;
            padding: 10px 20px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin-top: 10px;
            transition: background 0.2s;
        }
        .api-link:hover { background: #764ba2; }
        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>📊 Unified Dashboard</h1>
            <p>Your aggregated learning analytics and progress</p>
        </header>
        <div class="dashboard-grid">
            <div class="card">
                <h3>📈 Overall Level</h3>
                <div class="stat-value">--</div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: 0%"></div>
                </div>
                <p class="stat-label">Loading profile data...</p>
            </div>
            <div class="card">
                <h3>📝 Typing</h3>
                <div class="stat-value">--</div>
                <p class="stat-label">WPM Score</p>
                <a href="/typing" class="api-link">Launch App</a>
            </div>
            <div class="card">
                <h3>🔢 Math</h3>
                <div class="stat-value">--</div>
                <p class="stat-label">Accuracy Score</p>
                <a href="/math" class="api-link">Launch App</a>
            </div>
            <div class="card">
                <h3>📚 Reading</h3>
                <div class="stat-value">--</div>
                <p class="stat-label">Comprehension Score</p>
                <a href="/reading" class="api-link">Launch App</a>
            </div>
            <div class="card">
                <h3>🎹 Piano</h3>
                <div class="stat-value">--</div>
                <p class="stat-label">Performance Score</p>
                <a href="/piano" class="api-link">Launch App</a>
            </div>
            <div class="card">
                <h3>🔥 Daily Streak</h3>
                <div class="stat-value">--</div>
                <p class="stat-label">consecutive days</p>
            </div>
        </div>
    </div>
    <script>
        // Placeholder for dashboard data loading
        console.log('Unified Dashboard loaded. API endpoints available at /api/');
    </script>
</body>
</html>
	`))
}

// API Handlers

// getSystemStats returns platform-wide statistics
func (r *Router) getSystemStats(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	stats, err := r.service.GetSystemStats(ctx)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// getUserProfile returns user profile data
func (r *Router) getUserProfile(w http.ResponseWriter, req *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(req, "userID"), 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}

	ctx := req.Context()
	profile, err := r.service.GetUserProfile(ctx, uint(userID))
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

// getUserAnalytics returns cross-app analytics for a user
func (r *Router) getUserAnalytics(w http.ResponseWriter, req *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(req, "userID"), 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}

	ctx := req.Context()
	analytics, err := r.service.GetUserAnalytics(ctx, uint(userID))
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, analytics)
}

// getUserSessions returns recent sessions for a user
func (r *Router) getUserSessions(w http.ResponseWriter, req *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(req, "userID"), 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}

	// Get limit from query parameter
	limitStr := req.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := req.Context()
	sessions, err := r.service.GetRecentActivity(ctx, uint(userID), limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, sessions)
}

// getLeaderboard returns leaderboard for a category
func (r *Router) getLeaderboard(w http.ResponseWriter, req *http.Request) {
	category := chi.URLParam(req, "category")

	if !r.service.ValidateCategory(category) {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":              "invalid category",
			"available_categories": r.service.GetAvailableCategories(),
		})
		return
	}

	// Get limit from query parameter
	limitStr := req.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := req.Context()
	leaderboard, err := r.service.GetLeaderboard(ctx, category, limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, leaderboard)
}

// listLeaderboards returns all available leaderboard categories
func (r *Router) listLeaderboards(w http.ResponseWriter, req *http.Request) {
	categories := r.service.GetAvailableCategories()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	})
}

// getTrends returns performance trends for a user
func (r *Router) getTrends(w http.ResponseWriter, req *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(req, "userID"), 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}

	ctx := req.Context()
	trends, err := r.service.GetTrends(ctx, uint(userID))
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, trends)
}

// getDashboardOverview returns comprehensive dashboard overview
func (r *Router) getDashboardOverview(w http.ResponseWriter, req *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(req, "userID"), 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}

	ctx := req.Context()
	overview, err := r.service.GetDashboardOverview(ctx, uint(userID))
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, overview)
}

// getRecommendations returns personalized recommendations for a user
func (r *Router) getRecommendations(w http.ResponseWriter, req *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(req, "userID"), 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}

	ctx := req.Context()
	recommendations, err := r.service.GetRecommendations(ctx, uint(userID))
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, recommendations)
}

// Helper function to respond with JSON
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
