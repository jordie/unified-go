package math

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// Handler encapsulates HTTP handlers for math app
type Handler struct {
	service *Service
}

// NewHandler creates a new math HTTP handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Global handler instance for package-level functions
var globalHandler *Handler

// SetGlobalHandler sets the global handler instance
func SetGlobalHandler(handler *Handler) {
	globalHandler = handler
}

// IndexHandler serves the math app homepage
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Math Practice - Unified Educational Platform</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1000px; margin: 20px auto; padding: 20px; background: #f5f5f5; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        h1 { margin: 0; font-size: 28px; }
        .content { background: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .button { background: #3498db; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; margin: 5px 0; }
        .button:hover { background: #2980b9; }
        .stats { margin-top: 20px; padding: 15px; background: #ecf0f1; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧮 Math Practice</h1>
        <p>Improve your math skills with adaptive difficulty levels</p>
    </div>
    <div class="content">
        <h2>Start a Practice Session</h2>
        <p>Choose a problem type and difficulty level:</p>
        <form id="mathForm">
            <div>
                <label>Problem Type:</label>
                <select id="problemType" required>
                    <option value="addition">Addition</option>
                    <option value="subtraction">Subtraction</option>
                    <option value="multiplication">Multiplication</option>
                    <option value="division">Division</option>
                    <option value="fractions">Fractions</option>
                    <option value="algebra">Algebra</option>
                </select>
            </div>
            <div>
                <label>Difficulty:</label>
                <select id="difficulty" required>
                    <option value="easy">Easy</option>
                    <option value="medium">Medium</option>
                    <option value="hard">Hard</option>
                    <option value="very_hard">Very Hard</option>
                </select>
            </div>
            <button class="button" type="submit">Start Practice</button>
            <a href="/dashboard" class="button" style="text-decoration: none;">← Back to Dashboard</a>
        </form>
        <div class="stats">
            <h3>Your Stats</h3>
            <p>View your progress and performance metrics.</p>
            <a href="/math/api/stats" class="button">View Statistics</a>
        </div>
    </div>
</body>
</html>
	`))
}

// RequestTypes for API
type GenerateProblemRequest struct {
	ProblemType ProblemType   `json:"problem_type"`
	Difficulty  DifficultyLevel `json:"difficulty"`
}

type RecordSolutionRequest struct {
	ProblemID uint    `json:"problem_id"`
	Attempt   float64 `json:"attempt"`
	Correct   bool    `json:"correct"`
	TimeSpent float64 `json:"time_spent"`
}

type CompleteSessionRequest struct {
	ProblemType     ProblemType   `json:"problem_type"`
	Difficulty      DifficultyLevel `json:"difficulty"`
	TotalProblems   int           `json:"total_problems"`
	CorrectAnswers  int           `json:"correct_answers"`
	TimeSpent       float64       `json:"time_spent"`
}

// ResponseTypes for API
type GenerateProblemResponse struct {
	Question string  `json:"question"`
	Answer   float64 `json:"answer"`
}

type SessionResponse struct {
	SessionID     uint    `json:"session_id"`
	Score         float64 `json:"score"`
	Accuracy      float64 `json:"accuracy"`
	Message       string  `json:"message"`
}

type StatsResponse struct {
	Stats            UserMathStats `json:"stats"`
	MathLevel        string        `json:"math_level"`
	NextRecommendation string      `json:"next_recommendation"`
}

// GenerateProblem generates a new math problem
func (h *Handler) GenerateProblem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	question, answer, err := h.service.GenerateProblem(req.ProblemType, req.Difficulty)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateProblemResponse{
		Question: question,
		Answer:   answer,
	})
}

// RecordSolution records a user's solution
func (h *Handler) RecordSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := parseUserID(r)
	if userID == 0 {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	var req RecordSolutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	solution := &ProblemSolution{
		UserID:    userID,
		ProblemID: req.ProblemID,
		Attempt:   req.Attempt,
		Correct:   req.Correct,
		TimeSpent: req.TimeSpent,
	}

	if err := h.service.RecordSolution(r.Context(), solution); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// CompleteSession completes a quiz session
func (h *Handler) CompleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := parseUserID(r)
	if userID == 0 {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	var req CompleteSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session := &QuizSession{
		UserID:         userID,
		ProblemType:    req.ProblemType,
		Difficulty:     req.Difficulty,
		TotalProblems:  req.TotalProblems,
		CorrectAnswers: req.CorrectAnswers,
		TimeSpent:      req.TimeSpent,
		StartedAt:      getSessionStartTime(r),
	}

	if err := h.service.CompleteSession(r.Context(), session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accuracy := CalculateAccuracy(req.CorrectAnswers, req.TotalProblems)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SessionResponse{
		SessionID: session.ID,
		Score:     session.Score,
		Accuracy:  accuracy,
		Message:   "Session completed successfully",
	})
}

// GetUserStats retrieves user statistics
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := parseUserID(r)
	if userID == 0 {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	stats, err := h.service.GetUserStats(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mathLevel := EstimateMathLevel(stats.Accuracy)
	recommendation := getNextRecommendation(stats, mathLevel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatsResponse{
		Stats:              *stats,
		MathLevel:          mathLevel,
		NextRecommendation: recommendation,
	})
}

// GetProblemTypeStats retrieves stats for a problem type
func (h *Handler) GetProblemTypeStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := parseUserID(r)
	if userID == 0 {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	problemType := ProblemType(r.URL.Query().Get("type"))
	if problemType == "" {
		http.Error(w, "Problem type is required", http.StatusBadRequest)
		return
	}

	stats, err := h.service.GetProblemTypeStats(r.Context(), userID, problemType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetLeaderboard retrieves top performers
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	leaderboard, err := h.service.GetLeaderboard(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaderboard": leaderboard,
		"limit":       limit,
	})
}

// GetUserSessions retrieves user's session history
func (h *Handler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := parseUserID(r)
	if userID == 0 {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	sessions, err := h.service.GetUserSessions(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"limit":    limit,
		"offset":   offset,
	})
}

// Helper functions
func parseUserID(r *http.Request) uint {
	// In a real app, this would extract from session/JWT
	// For now, we'll use a query parameter or header
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}
	if userIDStr == "" {
		return 0
	}
	id, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}

func getSessionStartTime(r *http.Request) time.Time {
	// In a real app, this would be tracked properly
	// For now, return current time
	return time.Now()
}

func getNextRecommendation(stats *UserMathStats, level string) string {
	if stats.SessionsCompleted == 0 {
		return "Complete your first session to get recommendations"
	}

	switch level {
	case "beginner":
		return "Practice more easy problems to build confidence"
	case "intermediate":
		return "Try medium difficulty problems to improve"
	case "advanced":
		return "Challenge yourself with hard problems"
	default:
		return "You're an expert! Try teaching others or creating your own problems"
	}
}

// ============================================================================
// Package-level adapter functions for backward compatibility with router
// ============================================================================

// IndexHandler serves the math app homepage (package-level function)
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.IndexHandler(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// GenerateProblemAPI generates a new problem (package-level)
func GenerateProblemAPI(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.GenerateProblem(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// RecordSolutionAPI records a solution (package-level)
func RecordSolutionAPI(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.RecordSolution(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// CompleteSessionAPI completes a session (package-level)
func CompleteSessionAPI(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.CompleteSession(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// GetUserStatsAPI retrieves user stats (package-level)
func GetUserStatsAPI(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.GetUserStats(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// GetProblemTypeStatsAPI retrieves problem type stats (package-level)
func GetProblemTypeStatsAPI(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.GetProblemTypeStats(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// GetLeaderboardAPI retrieves leaderboard (package-level)
func GetLeaderboardAPI(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.GetLeaderboard(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// GetUserSessionsAPI retrieves user sessions (package-level)
func GetUserSessionsAPI(w http.ResponseWriter, r *http.Request) {
	if globalHandler != nil {
		globalHandler.GetUserSessions(w, r)
		return
	}
	http.Error(w, "Handler not initialized", http.StatusInternalServerError)
}

// Legacy placeholder functions (for backward compatibility)
func ListProblems(w http.ResponseWriter, r *http.Request) {
	GetLeaderboardAPI(w, r)
}

func SaveProgress(w http.ResponseWriter, r *http.Request) {
	CompleteSessionAPI(w, r)
}
