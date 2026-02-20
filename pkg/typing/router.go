package typing

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Router handles HTTP routes for typing app
type Router struct {
	db      *sql.DB
	service *Service
	router  chi.Router
}

// NewRouter creates a new typing router
func NewRouter(db *sql.DB) *Router {
	repo := NewRepository(db)
	service := NewService(repo)

	return &Router{
		db:      db,
		service: service,
		router:  chi.NewRouter(),
	}
}

// Routes configures all typing app routes
func (r *Router) Routes() chi.Router {
	// Test endpoints
	r.router.Post("/api/typing/test", r.CreateTest)
	r.router.Get("/api/typing/test/{testId}", r.GetTest)
	r.router.Get("/api/users/{userId}/typing/tests", r.GetUserTests)

	// Statistics endpoints
	r.router.Get("/api/users/{userId}/typing/stats", r.GetUserStats)
	r.router.Get("/api/typing/leaderboard", r.GetLeaderboard)
	r.router.Get("/api/users/{userId}/typing/history", r.GetHistory)

	// Dashboard
	r.router.Get("/api/typing/dashboard/{userId}", r.GetDashboard)

	// Lessons endpoints
	r.router.Get("/api/typing/lessons", r.GetLessons)
	r.router.Get("/api/typing/lessons/{lessonId}", r.GetLesson)

	return r.router
}

// CreateTest handles creating a new typing test
func (r *Router) CreateTest(w http.ResponseWriter, req *http.Request) {
	var testData struct {
		UserID   uint    `json:"user_id"`
		Content  string  `json:"content"`
		Duration float64 `json:"duration"`
		Errors   int     `json:"errors"`
	}

	if err := json.NewDecoder(req.Body).Decode(&testData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := r.service.ProcessTypingTest(req.Context(), testData.UserID, testData.Content, testData.Duration, testData.Errors)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

// GetTest retrieves a specific typing test
func (r *Router) GetTest(w http.ResponseWriter, req *http.Request) {
	testIdStr := chi.URLParam(req, "testId")
	testID, err := strconv.ParseUint(testIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid test ID")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id": testID,
		"status": "placeholder - GetTest",
	})
}

// GetUserTests retrieves user's typing tests
func (r *Router) GetUserTests(w http.ResponseWriter, req *http.Request) {
	userIdStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	limit := 20
	offset := 0

	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	tests, err := r.service.repo.GetUserTests(req.Context(), uint(userID), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get tests")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tests": tests,
		"user_id": userID,
		"limit": limit,
		"offset": offset,
	})
}

// GetUserStats retrieves user typing statistics
func (r *Router) GetUserStats(w http.ResponseWriter, req *http.Request) {
	userIdStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	stats, err := r.service.GetUserProgress(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetLeaderboard retrieves typing leaderboard
func (r *Router) GetLeaderboard(w http.ResponseWriter, req *http.Request) {
	limit := 100
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	leaderboard, err := r.service.GetLeaderboard(req.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get leaderboard")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"leaderboard": leaderboard,
		"limit": limit,
	})
}

// GetHistory retrieves user's typing history
func (r *Router) GetHistory(w http.ResponseWriter, req *http.Request) {
	userIdStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	days := 30
	if daysStr := req.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	history, err := r.service.GetUserHistory(req.Context(), uint(userID), days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get history")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
		"user_id": userID,
		"days": days,
	})
}

// GetDashboard retrieves user's typing dashboard
func (r *Router) GetDashboard(w http.ResponseWriter, req *http.Request) {
	userIdStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	stats, err := r.service.GetUserProgress(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get dashboard")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
		"skill_level": EstimateTypingLevel(stats.AverageWPM),
	})
}

// GetLessons retrieves available typing lessons
func (r *Router) GetLessons(w http.ResponseWriter, req *http.Request) {
	lessons := []map[string]interface{}{
		{"id": 1, "title": "Home Row Keys", "difficulty": "beginner", "duration_minutes": 5},
		{"id": 2, "title": "Top Row Keys", "difficulty": "beginner", "duration_minutes": 5},
		{"id": 3, "title": "Bottom Row Keys", "difficulty": "beginner", "duration_minutes": 5},
		{"id": 4, "title": "Number Keys", "difficulty": "intermediate", "duration_minutes": 10},
		{"id": 5, "title": "Symbol Keys", "difficulty": "intermediate", "duration_minutes": 10},
		{"id": 6, "title": "Speed Test 1", "difficulty": "advanced", "duration_minutes": 1},
		{"id": 7, "title": "Speed Test 2", "difficulty": "advanced", "duration_minutes": 5},
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"lessons": lessons,
		"total": len(lessons),
	})
}

// GetLesson retrieves a specific typing lesson
func (r *Router) GetLesson(w http.ResponseWriter, req *http.Request) {
	lessonIdStr := chi.URLParam(req, "lessonId")
	lessonID64, err := strconv.ParseUint(lessonIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	lessonID := uint(lessonID64)

	lessons := map[uint]map[string]interface{}{
		1: {"id": 1, "title": "Home Row Keys", "difficulty": "beginner", "content": "asdfghjkl;"},
		2: {"id": 2, "title": "Top Row Keys", "difficulty": "beginner", "content": "qwertyuiop"},
		3: {"id": 3, "title": "Bottom Row Keys", "difficulty": "beginner", "content": "zxcvbnm,./"},
	}

	if lesson, exists := lessons[lessonID]; exists {
		respondJSON(w, http.StatusOK, lesson)
	} else {
		respondError(w, http.StatusNotFound, "Lesson not found")
	}
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"status": status,
	})
}
