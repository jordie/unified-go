package reading

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Router configures reading app routes
type Router struct {
	service *Service
}

// NewRouter creates a new reading router
func NewRouter(db *sql.DB) *Router {
	repo := NewRepository(db)
	service := NewService(repo)
	return &Router{service: service}
}

// Routes returns the reading router with all configured routes
func (r *Router) Routes() chi.Router {
	router := chi.NewRouter()

	// Index page
	router.Get("/", IndexHandler)

	// Book operations
	router.Get("/api/books", r.GetBooks)
	router.Post("/api/books", r.CreateBook)
	router.Get("/api/books/{id}", r.GetBook)

	// Reading session operations
	router.Post("/api/sessions", r.ProcessSession)
	router.Get("/api/sessions/{id}", r.GetSession)
	router.Get("/api/users/{userId}/sessions", r.GetUserSessions)

	// User statistics
	router.Get("/api/users/{userId}/stats", r.GetUserStats)
	router.Get("/api/users/{userId}/progress", r.GetUserProgress)
	router.Get("/api/leaderboard", r.GetLeaderboard)

	// Comprehension tests
	router.Post("/api/comprehension", r.SaveComprehensionTest)
	router.Get("/api/sessions/{sessionId}/comprehension", r.GetComprehensionTests)
	router.Get("/api/sessions/{sessionId}/analysis", r.AnalyzeComprehension)

	// Content validation
	router.Post("/api/validate", r.ValidateContent)

	return router
}

// GetBooks retrieves available books with filtering
func (r *Router) GetBooks(w http.ResponseWriter, req *http.Request) {
	difficulty := req.URL.Query().Get("difficulty")
	limitStr := req.URL.Query().Get("limit")
	offsetStr := req.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	books, err := r.service.repo.GetBooks(req.Context(), difficulty, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get books: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"books":  books,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateBook creates a new book
func (r *Router) CreateBook(w http.ResponseWriter, req *http.Request) {
	var book Book
	if err := json.NewDecoder(req.Body).Decode(&book); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := r.service.repo.SaveBook(req.Context(), &book)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to save book: "+err.Error())
		return
	}

	book.ID = id
	respondJSON(w, http.StatusCreated, book)
}

// GetBook retrieves a single book by ID
func (r *Router) GetBook(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	book, err := r.service.repo.GetBookByID(req.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get book")
		return
	}

	if book == nil {
		respondError(w, http.StatusNotFound, "Book not found")
		return
	}

	respondJSON(w, http.StatusOK, book)
}

// ProcessSession processes a completed reading session
func (r *Router) ProcessSession(w http.ResponseWriter, req *http.Request) {
	var reqData struct {
		UserID    uint    `json:"user_id"`
		BookID    uint    `json:"book_id"`
		Content   string  `json:"content"`
		TimeSpent float64 `json:"time_spent"`
		Errors    int     `json:"errors"`
	}

	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := r.service.ProcessTestResult(req.Context(), reqData.UserID, reqData.BookID,
		reqData.Content, reqData.TimeSpent, reqData.Errors)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to process session: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, session)
}

// GetSession retrieves a reading session by ID
func (r *Router) GetSession(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	session, err := r.service.repo.GetSessionByID(req.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get session")
		return
	}

	if session == nil {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// GetUserSessions retrieves a user's reading sessions
func (r *Router) GetUserSessions(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	limitStr := req.URL.Query().Get("limit")
	offsetStr := req.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	sessions, err := r.service.GetUserTestHistory(req.Context(), uint(userID), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sessions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetUserStats retrieves aggregated user statistics
func (r *Router) GetUserStats(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	stats, err := r.service.GetUserStatistics(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetUserProgress retrieves user progress with trend analysis
func (r *Router) GetUserProgress(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	progress, err := r.service.CalculateUserProgress(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate progress")
		return
	}

	respondJSON(w, http.StatusOK, progress)
}

// GetLeaderboard retrieves the reading leaderboard
func (r *Router) GetLeaderboard(w http.ResponseWriter, req *http.Request) {
	limitStr := req.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	stats, err := r.service.GetLeaderboard(req.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get leaderboard")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"leaderboard": stats,
		"limit":       limit,
	})
}

// SaveComprehensionTest saves a comprehension quiz result
func (r *Router) SaveComprehensionTest(w http.ResponseWriter, req *http.Request) {
	var test ComprehensionTest
	if err := json.NewDecoder(req.Body).Decode(&test); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := r.service.repo.SaveComprehensionTest(req.Context(), &test)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to save test: "+err.Error())
		return
	}

	test.ID = id
	respondJSON(w, http.StatusCreated, test)
}

// GetComprehensionTests retrieves comprehension tests for a session
func (r *Router) GetComprehensionTests(w http.ResponseWriter, req *http.Request) {
	sessionIDStr := chi.URLParam(req, "sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	tests, err := r.service.repo.GetComprehensionTests(req.Context(), uint(sessionID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get tests")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tests": tests,
	})
}

// AnalyzeComprehension analyzes comprehension test results
func (r *Router) AnalyzeComprehension(w http.ResponseWriter, req *http.Request) {
	sessionIDStr := chi.URLParam(req, "sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	analysis, err := r.service.GetComprehensionAnalysis(req.Context(), uint(sessionID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to analyze comprehension")
		return
	}

	respondJSON(w, http.StatusOK, analysis)
}

// ValidateContent validates reading content
func (r *Router) ValidateContent(w http.ResponseWriter, req *http.Request) {
	var reqData struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := r.service.ValidateTestContent(reqData.Content)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
	})
}
