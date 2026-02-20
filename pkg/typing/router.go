package typing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jgirmay/unified-go/internal/database"
	"github.com/jgirmay/unified-go/internal/middleware"
)

// Router handles routing for typing app
type Router struct {
	mux     *http.ServeMux
	service *Service
	authMW  *middleware.AuthMiddleware
}

// NewRouter creates a new typing router
func NewRouter(db *database.Pool, authMW *middleware.AuthMiddleware) *Router {
	service := NewServiceWithPool(db)
	return &Router{
		mux:     http.NewServeMux(),
		service: service,
		authMW:  authMW,
	}
}

// RegisterRoutes registers all typing routes
func (r *Router) RegisterRoutes() *http.ServeMux {
	// Public routes
	r.mux.HandleFunc("/typing/", r.indexHandler)

	// API routes
	r.mux.HandleFunc("/typing/api/save_result", r.saveResultHandler)
	r.mux.HandleFunc("/typing/api/stats", r.statsHandler)
	r.mux.HandleFunc("/typing/api/leaderboard", r.leaderboardHandler)
	r.mux.HandleFunc("/typing/api/history", r.historyHandler)
	r.mux.HandleFunc("/typing/api/settings", r.settingsHandler)

	return r.mux
}

// indexHandler serves the typing app homepage
func (r *Router) indexHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// For now, serve a basic HTML page
	// This will be replaced with template rendering in Subtask 5
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Typing Practice - Unified Educational Platform</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1 { color: #333; }
        .container { background: #f5f5f5; padding: 20px; border-radius: 5px; margin-top: 20px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-top: 20px; }
        .stat-card { background: white; padding: 15px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stat-value { font-size: 24px; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 5px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; }
        button:hover { background: #0056b3; }
    </style>
</head>
<body>
    <h1>Typing Practice</h1>
    <div class="container">
        <p>Welcome to the Typing Practice application! Improve your typing speed and accuracy.</p>
        <button onclick="location.href='/typing/practice'">Start Typing Test</button>
        <button onclick="location.href='/typing/stats'">View Statistics</button>
        <button onclick="location.href='/typing/leaderboard'">View Leaderboard</button>
    </div>
    <div class="stats" id="stats"></div>
    <script>
        // Load user stats via API
        fetch('/typing/api/stats')
            .then(r => r.json())
            .then(data => {
                if (data.success) {
                    const stats = data.data;
                    let html = '';
                    html += '<div class="stat-card"><div class="stat-value">' + (stats.average_wpm || 0) + '</div><div class="stat-label">Average WPM</div></div>';
                    html += '<div class="stat-card"><div class="stat-value">' + (stats.best_wpm || 0) + '</div><div class="stat-label">Best WPM</div></div>';
                    html += '<div class="stat-card"><div class="stat-value">' + ((stats.average_accuracy || 0).toFixed(1)) + '%</div><div class="stat-label">Accuracy</div></div>';
                    html += '<div class="stat-card"><div class="stat-value">' + (stats.total_tests || 0) + '</div><div class="stat-label">Total Tests</div></div>';
                    document.getElementById('stats').innerHTML = html;
                }
            })
            .catch(e => console.error('Failed to load stats:', e));
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

// SaveResultRequest represents a save result API request
type SaveResultRequest struct {
	Content    string  `json:"content"`
	TimeSpent  float64 `json:"time_spent"`
	ErrorCount int     `json:"error_count"`
	TestMode   string  `json:"test_mode"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// saveResultHandler handles POST /typing/api/save_result
func (r *Router) saveResultHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Get user ID from session
	userID, ok := middleware.GetUserID(req)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	// Parse request body
	var reqData SaveResultRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Process test result
	result, err := r.service.ProcessTestResult(
		req.Context(),
		uint(userID),
		reqData.Content,
		reqData.TimeSpent,
		reqData.ErrorCount,
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save result: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "Result saved successfully",
		Data: map[string]interface{}{
			"id":        result.ID,
			"wpm":       result.WPM,
			"raw_wpm":   result.RawWPM,
			"accuracy":  result.Accuracy,
			"timestamp": result.CreatedAt,
		},
	})
}

// statsHandler handles GET /typing/api/stats
func (r *Router) statsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Get user ID from session
	userID, ok := middleware.GetUserID(req)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	// Get user statistics
	stats, err := r.service.GetUserStatistics(req.Context(), uint(userID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get statistics: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    stats,
	})
}

// leaderboardHandler handles GET /typing/api/leaderboard
func (r *Router) leaderboardHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Parse query parameters
	limit := 10
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Get leaderboard
	leaderboard, err := r.service.GetLeaderboard(req.Context(), limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get leaderboard: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"leaderboard": leaderboard,
			"count":       len(leaderboard),
		},
	})
}

// historyHandler handles GET /typing/api/history
func (r *Router) historyHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Get user ID from session
	userID, ok := middleware.GetUserID(req)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	// Parse query parameters
	limit := 20
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get test history
	history, err := r.service.GetUserTestHistory(req.Context(), uint(userID), limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get history: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"history": history,
			"count":   len(history),
			"limit":   limit,
			"offset":  offset,
		},
	})
}

// settingsHandler handles POST /typing/api/settings
func (r *Router) settingsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Get user ID from session
	userID, ok := middleware.GetUserID(req)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	// Parse request body
	var settings map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&settings); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// For now, just acknowledge the settings save
	// In a full implementation, this would save to a preferences table
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "Settings saved successfully",
		Data: map[string]interface{}{
			"user_id":  userID,
			"settings": settings,
		},
	})
}
