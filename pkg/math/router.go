package math

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Router handles HTTP routes for math app
type Router struct {
	db      *sql.DB
	service *Service
	router  chi.Router
}

// NewRouter creates a new math router
func NewRouter(db *sql.DB) *Router {
	repo := NewRepository(db)
	service := NewService(repo)

	return &Router{
		db:      db,
		service: service,
		router:  chi.NewRouter(),
	}
}

// Routes configures all math app routes
func (r *Router) Routes() chi.Router {
	// Problem generation endpoints
	r.router.Post("/api/math/problem", r.GenerateProblem)
	r.router.Get("/api/math/problem/types", r.GetProblemTypes)

	// Session endpoints
	r.router.Post("/api/math/session/start", r.StartSession)
	r.router.Post("/api/math/session/complete", r.CompleteSession)
	r.router.Get("/api/users/{userId}/math/sessions", r.GetUserSessions)

	// Statistics endpoints
	r.router.Get("/api/users/{userId}/math/stats", r.GetUserStats)
	r.router.Get("/api/users/{userId}/math/problem-type/{problemType}", r.GetProblemTypeStats)
	r.router.Get("/api/math/leaderboard", r.GetLeaderboard)

	return r.router
}

// GenerateProblem generates a new math problem
func (r *Router) GenerateProblem(w http.ResponseWriter, req *http.Request) {
	var requestData struct {
		ProblemType ProblemType   `json:"problem_type"`
		Difficulty  DifficultyLevel `json:"difficulty"`
	}

	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	question, answer, err := r.service.GenerateProblem(requestData.ProblemType, requestData.Difficulty)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"question": question,
		"answer":   answer,
	})
}

// GetProblemTypes returns available problem types
func (r *Router) GetProblemTypes(w http.ResponseWriter, req *http.Request) {
	types := []map[string]string{
		{"type": "addition", "description": "Basic addition"},
		{"type": "subtraction", "description": "Basic subtraction"},
		{"type": "multiplication", "description": "Multiplication"},
		{"type": "division", "description": "Division"},
		{"type": "fractions", "description": "Fraction operations"},
		{"type": "algebra", "description": "Simple algebra"},
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"problem_types": types,
		"total":         len(types),
	})
}

// StartSession starts a new quiz session
func (r *Router) StartSession(w http.ResponseWriter, req *http.Request) {
	var sessionData struct {
		UserID      uint            `json:"user_id"`
		ProblemType ProblemType     `json:"problem_type"`
		Difficulty  DifficultyLevel `json:"difficulty"`
		TotalProblems int           `json:"total_problems"`
	}

	if err := json.NewDecoder(req.Body).Decode(&sessionData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session := &QuizSession{
		UserID:        sessionData.UserID,
		ProblemType:   sessionData.ProblemType,
		Difficulty:    sessionData.Difficulty,
		TotalProblems: sessionData.TotalProblems,
		StartedAt:     time.Now(),
	}

	respondJSON(w, http.StatusCreated, session)
}

// CompleteSession completes a quiz session
func (r *Router) CompleteSession(w http.ResponseWriter, req *http.Request) {
	var sessionData struct {
		UserID         uint            `json:"user_id"`
		ProblemType    ProblemType     `json:"problem_type"`
		Difficulty     DifficultyLevel `json:"difficulty"`
		TotalProblems  int             `json:"total_problems"`
		CorrectAnswers int             `json:"correct_answers"`
		TimeSpent      float64         `json:"time_spent"`
		StartedAt      time.Time       `json:"started_at"`
	}

	if err := json.NewDecoder(req.Body).Decode(&sessionData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session := &QuizSession{
		UserID:                sessionData.UserID,
		ProblemType:           sessionData.ProblemType,
		Difficulty:            sessionData.Difficulty,
		TotalProblems:         sessionData.TotalProblems,
		CorrectAnswers:        sessionData.CorrectAnswers,
		TimeSpent:             sessionData.TimeSpent,
		StartedAt:             sessionData.StartedAt,
		AverageTimePerProblem: sessionData.TimeSpent / float64(sessionData.TotalProblems),
	}

	if err := r.service.CompleteSession(req.Context(), session); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"score":                session.Score,
		"correct_answers":      session.CorrectAnswers,
		"total_problems":       session.TotalProblems,
		"time_spent":           session.TimeSpent,
		"average_time_per_problem": session.AverageTimePerProblem,
	})
}

// GetUserStats retrieves user math statistics
func (r *Router) GetUserStats(w http.ResponseWriter, req *http.Request) {
	userIdStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	stats, err := r.service.GetUserStats(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetProblemTypeStats retrieves stats for a problem type
func (r *Router) GetProblemTypeStats(w http.ResponseWriter, req *http.Request) {
	userIdStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	problemTypeStr := chi.URLParam(req, "problemType")
	problemType := ProblemType(problemTypeStr)

	result, err := r.service.GetProblemTypeStats(req.Context(), uint(userID), problemType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get problem type stats")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetLeaderboard retrieves math leaderboard
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
		"limit":       limit,
	})
}

// GetUserSessions retrieves user's quiz sessions
func (r *Router) GetUserSessions(w http.ResponseWriter, req *http.Request) {
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

	sessions, err := r.service.GetUserSessions(req.Context(), uint(userID), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sessions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"user_id":  userID,
		"limit":    limit,
		"offset":   offset,
	})
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
		"error":  message,
		"status": status,
	})
}
