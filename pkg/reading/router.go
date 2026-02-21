package reading

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

	// Audio recording and transcription
	router.Post("/api/audio/record", r.RecordAudio)
	router.Post("/api/audio/transcribe", r.TranscribeAudio)

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

// RecordAudio handles POST /api/reading/audio/record
func (r *Router) RecordAudio(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse multipart form with max 10MB file size
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse form: %v", err),
		})
		return
	}

	// Get audio file from request
	file, fileHeader, err := req.FormFile("audio")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   "No audio file provided",
		})
		return
	}
	defer file.Close()

	// Get user ID from form
	userIDStr := req.FormValue("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   "Invalid user_id",
		})
		return
	}

	// Create audio storage directory if it doesn't exist
	audioDir := "data/audio/reading"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create audio directory: %v", err),
		})
		return
	}

	// Generate unique audio ID with timestamp
	audioID := fmt.Sprintf("reading_%d_%d", userID, time.Now().UnixNano())
	audioPath := filepath.Join(audioDir, audioID+filepath.Ext(fileHeader.Filename))

	// Save audio file
	dst, err := os.Create(audioPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save audio file: %v", err),
		})
		return
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write audio file: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AudioRecordResponse{
		Success: true,
		AudioID: audioID,
		Message: fmt.Sprintf("Audio file recorded successfully (%d bytes)", fileHeader.Size),
	})
}

// TranscribeAudio handles POST /api/reading/audio/transcribe
func (r *Router) TranscribeAudio(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqData AudioTranscribeRequest
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioTranscribeResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if reqData.AudioID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioTranscribeResponse{
			Success: false,
			Error:   "audio_id is required",
		})
		return
	}

	// For now, return a placeholder transcription
	// In production, this would call Google Speech-to-Text API, Whisper, or similar service
	transcript := "This is a placeholder transcription. Audio processing service not yet integrated."

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AudioTranscribeResponse{
		Success:    true,
		AudioID:    reqData.AudioID,
		Transcript: transcript,
		Confidence: 0.0, // Placeholder confidence
	})
}
