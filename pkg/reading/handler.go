package reading

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler handles HTTP requests for reading app
type Handler struct {
	service *Service
}

// NewHandler creates a new reading handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Request/Response types

type SaveReadingResultRequest struct {
	UserID      uint    `json:"user_id"`
	BookID      uint    `json:"book_id"`
	Content     string  `json:"content"`
	Duration    float64 `json:"duration_seconds"`
	ErrorCount  int     `json:"error_count"`
}

type SaveReadingResultResponse struct {
	Success bool              `json:"success"`
	Data    *ReadingSession   `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type GetMasteryStatsResponse struct {
	Success bool              `json:"success"`
	Data    *ReadingStats     `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type GetReadingStatsResponse struct {
	Success bool              `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type GetBooksResponse struct {
	Success bool              `json:"success"`
	Data    []Book            `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type SubmitComprehensionRequest struct {
	SessionID uint              `json:"session_id"`
	Answers   []ComprehensionAnswer `json:"answers"`
}

type ComprehensionAnswer struct {
	Question      string `json:"question"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
}

type SubmitComprehensionResponse struct {
	Success bool              `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type ComprehensionStatsResponse struct {
	Success bool              `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type UserSkillsResponse struct {
	Success bool              `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type GetLeaderboardResponse struct {
	Success bool              `json:"success"`
	Data    []ReadingStats    `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// Helper function to respond with JSON
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Helper function to respond with error
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// SaveReadingResult saves a reading session result
// POST /api/save_reading_result
func (h *Handler) SaveReadingResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SaveReadingResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.ProcessTestResult(r.Context(), req.UserID, req.BookID, req.Content, req.Duration, req.ErrorCount)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, SaveReadingResultResponse{
		Success: true,
		Data:    result,
	})
}

// GetMasteryStats gets word mastery statistics
// GET /api/get_mastery_stats?user_id=1
func (h *Handler) GetMasteryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "user_id parameter required")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	stats, err := h.service.GetUserStatistics(r.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GetMasteryStatsResponse{
		Success: true,
		Data:    stats,
	})
}

// GetReadingStats gets overall reading statistics
// GET /api/reading_stats?user_id=1
func (h *Handler) GetReadingStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "user_id parameter required")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	progress, err := h.service.CalculateUserProgress(r.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GetReadingStatsResponse{
		Success: true,
		Data:    progress,
	})
}

// GetBooks gets available books
// GET /api/passages?difficulty=intermediate&limit=10&offset=0
func (h *Handler) GetBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	difficulty := r.URL.Query().Get("difficulty")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	books, err := h.service.repo.GetBooks(r.Context(), difficulty, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GetBooksResponse{
		Success: true,
		Data:    books,
	})
}

// GetBook gets a specific book
// GET /api/passages/<id>
func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract book ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/reading/api/passages/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		respondError(w, http.StatusBadRequest, "Book ID required")
		return
	}

	bookID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	book, err := h.service.repo.GetBookByID(r.Context(), uint(bookID))
	if err != nil {
		if errors.Is(err, errors.New("book not found")) {
			respondError(w, http.StatusNotFound, "Book not found")
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    book,
	})
}

// SubmitComprehension submits comprehension answers
// POST /api/submit_comprehension
func (h *Handler) SubmitComprehension(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SubmitComprehensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Save all comprehension answers
	for _, answer := range req.Answers {
		test := &ComprehensionTest{
			SessionID:     req.SessionID,
			Question:      answer.Question,
			UserAnswer:    answer.UserAnswer,
			CorrectAnswer: answer.CorrectAnswer,
			IsCorrect:     answer.UserAnswer == answer.CorrectAnswer,
			Score:         0,
		}
		if test.IsCorrect {
			test.Score = 100.0
		}
		h.service.repo.SaveComprehensionTest(r.Context(), test)
	}

	// Get analysis
	analysis, err := h.service.GetComprehensionAnalysis(r.Context(), req.SessionID)
	if err != nil {
		analysis = map[string]interface{}{
			"total_questions": len(req.Answers),
			"correct_answers": 0,
			"score":           0.0,
		}
	}

	respondJSON(w, http.StatusOK, SubmitComprehensionResponse{
		Success: true,
		Data:    analysis,
	})
}

// GetComprehensionStats gets comprehension statistics
// GET /api/comprehension_stats/<session_id>
func (h *Handler) GetComprehensionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/reading/api/comprehension_stats/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		respondError(w, http.StatusBadRequest, "Session ID required")
		return
	}

	sessionID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	analysis, err := h.service.GetComprehensionAnalysis(r.Context(), uint(sessionID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ComprehensionStatsResponse{
		Success: true,
		Data:    analysis,
	})
}

// GetUserSkills gets user skill levels
// GET /api/user_skills/<user_id>
func (h *Handler) GetUserSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/reading/api/user_skills/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		respondError(w, http.StatusBadRequest, "User ID required")
		return
	}

	userID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	analysis, err := h.service.AnalyzeReadingPerformance(r.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, UserSkillsResponse{
		Success: true,
		Data:    analysis,
	})
}

// GetLeaderboard gets the reading leaderboard
// GET /api/get_leaderboard?limit=10
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	leaderboard, err := h.service.GetLeaderboard(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GetLeaderboardResponse{
		Success: true,
		Data:    leaderboard,
	})
}

// RecommendBooks gets personalized book recommendations
// GET /api/recommend_books?user_id=1&limit=5
func (h *Handler) RecommendBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "user_id parameter required")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	recommendations, err := h.service.GetBookRecommendations(r.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    recommendations,
	})
}

// IndexHandler serves the reading app homepage
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Reading App - Unified Educational Platform</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .api-docs { background: #f0f0f0; padding: 20px; border-radius: 5px; margin-top: 20px; }
        .endpoint { margin: 10px 0; padding: 10px; background: #fff; border-left: 4px solid #007bff; }
    </style>
</head>
<body>
    <h1>📚 Reading App (Go)</h1>
    <div class="api-docs">
        <h2>API Endpoints</h2>
        <div class="endpoint"><strong>POST</strong> /api/save_reading_result - Save reading session</div>
        <div class="endpoint"><strong>GET</strong> /api/get_mastery_stats - Get word mastery stats</div>
        <div class="endpoint"><strong>GET</strong> /api/reading_stats - Get reading statistics</div>
        <div class="endpoint"><strong>GET</strong> /api/passages - List books</div>
        <div class="endpoint"><strong>GET</strong> /api/passages/{id} - Get specific book</div>
        <div class="endpoint"><strong>POST</strong> /api/submit_comprehension - Submit comprehension answers</div>
        <div class="endpoint"><strong>GET</strong> /api/comprehension_stats/{id} - Get comprehension stats</div>
        <div class="endpoint"><strong>GET</strong> /api/user_skills/{id} - Get user skill analysis</div>
        <div class="endpoint"><strong>GET</strong> /api/get_leaderboard - Get leaderboard</div>
        <div class="endpoint"><strong>GET</strong> /api/recommend_books - Get book recommendations</div>
        <p><a href="/dashboard">← Back to Dashboard</a></p>
    </div>
</body>
</html>
	`))
}
