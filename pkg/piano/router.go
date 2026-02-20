package piano

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Router configures piano app routes
type Router struct {
	service *Service
	auth    *PianoAuthMiddleware
}

// NewRouter creates a new piano router
func NewRouter(db *sql.DB) *Router {
	repo := NewRepository(db)
	service := NewService(repo)
	return &Router{
		service: service,
		auth:    nil, // Auth middleware will be set if available
	}
}

// SetAuthMiddleware sets the auth middleware for this router
func (r *Router) SetAuthMiddleware(auth *PianoAuthMiddleware) {
	r.auth = auth
}

// Routes returns the piano router with all configured routes
func (r *Router) Routes() chi.Router {
	router := chi.NewRouter()

	// Public UI Routes - no authentication required
	router.Get("/", IndexHandler)
	router.Get("/songs", r.SongsHandler)      // Public songs listing

	// Public API Routes - no authentication required
	router.Get("/api/songs", r.GetSongs)      // List songs (no auth)
	router.Get("/api/songs/{id}", r.GetSong)  // Get song details (no auth)
	router.Get("/api/leaderboard", r.GetLeaderboard) // Public leaderboard

	// Protected UI Routes - require authentication
	router.With(r.requireAuth).Get("/practice/{id}", r.PracticeHandler)
	router.With(r.requireAuth).Get("/dashboard", r.DashboardHandler)

	// Protected API Routes - require authentication
	// Song operations (create requires auth)
	router.With(r.requireAuthJSON).Post("/api/songs", r.CreateSong)

	// Lesson operations (all require auth)
	router.With(r.requireAuthJSON).Post("/api/lessons", r.StartLesson)
	router.With(r.requireAuthJSON).Get("/api/lessons/{id}", r.GetLesson)
	router.With(r.requireAuthJSON).Get("/api/users/{userId}/lessons", r.GetUserLessons)

	// Practice session operations (require auth)
	router.With(r.requireAuthJSON).Post("/api/practice", r.SavePracticeSession)
	router.With(r.requireAuthJSON).Get("/api/practice/{id}", r.GetPracticeSession)

	// User progress and metrics (require auth)
	router.With(r.requireAuthJSON).Get("/api/users/{userId}/progress", r.GetUserProgress)
	router.With(r.requireAuthJSON).Get("/api/users/{userId}/metrics", r.GetUserMetrics)
	router.With(r.requireAuthJSON).Get("/api/users/{userId}/evaluation", r.EvaluatePerformance)

	// Music theory (require auth)
	router.With(r.requireAuthJSON).Post("/api/theory-quiz", r.GenerateQuiz)
	router.With(r.requireAuthJSON).Get("/api/sessions/{sessionId}/analysis", r.AnalyzeTheory)

	// MIDI operations (require auth)
	router.With(r.requireAuthJSON).Post("/api/midi/upload", r.UploadMIDI)
	router.With(r.requireAuthJSON).Get("/api/midi/{sessionId}", r.DownloadMIDI)

	// Lesson recommendations (require auth)
	router.With(r.requireAuthJSON).Get("/api/recommend/{userId}", r.RecommendLesson)
	router.With(r.requireAuthJSON).Get("/api/progression-path/{userId}", r.GetProgressionPath)

	return router
}

// requireAuth is a middleware that requires authentication (redirects on failure)
func (r *Router) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userID := GetUserIDFromRequest(req)
		if userID == 0 {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// requireAuthJSON is a middleware that requires authentication (returns JSON error)
func (r *Router) requireAuthJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userID := GetUserIDFromRequest(req)
		if userID == 0 {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"error":  "authentication required",
				"status": 401,
			})
			return
		}
		next.ServeHTTP(w, req)
	})
}

// GetSongs retrieves available songs with filtering
func (r *Router) GetSongs(w http.ResponseWriter, req *http.Request) {
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

	songs, err := r.service.repo.GetSongs(req.Context(), difficulty, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get songs: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"songs":  songs,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateSong creates a new song
func (r *Router) CreateSong(w http.ResponseWriter, req *http.Request) {
	var song Song
	if err := json.NewDecoder(req.Body).Decode(&song); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := r.service.repo.SaveSong(req.Context(), &song)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to save song: "+err.Error())
		return
	}

	song.ID = id
	respondJSON(w, http.StatusCreated, song)
}

// GetSong retrieves a single song by ID
func (r *Router) GetSong(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid song ID")
		return
	}

	song, err := r.service.repo.GetSongByID(req.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get song")
		return
	}

	if song == nil {
		respondError(w, http.StatusNotFound, "Song not found")
		return
	}

	respondJSON(w, http.StatusOK, song)
}

// StartLesson starts a new piano lesson
func (r *Router) StartLesson(w http.ResponseWriter, req *http.Request) {
	var lesson PianoLesson
	if err := json.NewDecoder(req.Body).Decode(&lesson); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := r.service.repo.SaveLesson(req.Context(), &lesson)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to create lesson: "+err.Error())
		return
	}

	lesson.ID = id
	respondJSON(w, http.StatusCreated, lesson)
}

// GetLesson retrieves a lesson by ID
func (r *Router) GetLesson(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	lesson, err := r.service.repo.GetLessonByID(req.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get lesson")
		return
	}

	if lesson == nil {
		respondError(w, http.StatusNotFound, "Lesson not found")
		return
	}

	respondJSON(w, http.StatusOK, lesson)
}

// GetUserLessons retrieves a user's lessons
func (r *Router) GetUserLessons(w http.ResponseWriter, req *http.Request) {
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

	lessons, err := r.service.repo.GetUserLessons(req.Context(), uint(userID), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get lessons")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"lessons": lessons,
		"limit":   limit,
		"offset":  offset,
	})
}

// SavePracticeSession saves a practice session with MIDI recording
func (r *Router) SavePracticeSession(w http.ResponseWriter, req *http.Request) {
	var reqData struct {
		UserID       uint    `json:"user_id"`
		SongID       uint    `json:"song_id"`
		RecordedBPM  float64 `json:"recorded_bpm"`
		Duration     float64 `json:"duration"`
		NotesCorrect int     `json:"notes_correct"`
		NotesTotal   int     `json:"notes_total"`
	}

	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := r.service.ProcessLesson(req.Context(), reqData.UserID, reqData.SongID,
		reqData.RecordedBPM, reqData.Duration, reqData.NotesCorrect, reqData.NotesTotal)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to save session: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, session)
}

// GetPracticeSession retrieves a practice session
func (r *Router) GetPracticeSession(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Placeholder - would need GetSessionByID in repository
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":     id,
		"status": "retrieved",
	})
}

// GetUserProgress retrieves user piano progress
func (r *Router) GetUserProgress(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	progress, err := r.service.repo.GetUserProgress(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get progress")
		return
	}

	respondJSON(w, http.StatusOK, progress)
}

// GetUserMetrics retrieves comprehensive user metrics
func (r *Router) GetUserMetrics(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	metrics, err := r.service.GetUserMetrics(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get metrics")
		return
	}

	respondJSON(w, http.StatusOK, metrics)
}

// EvaluatePerformance evaluates user performance with trends
func (r *Router) EvaluatePerformance(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	evaluation, err := r.service.EvaluatePerformance(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to evaluate performance")
		return
	}

	respondJSON(w, http.StatusOK, evaluation)
}

// GenerateQuiz generates a music theory quiz
func (r *Router) GenerateQuiz(w http.ResponseWriter, req *http.Request) {
	var reqData struct {
		Difficulty string `json:"difficulty"`
		Count      int    `json:"count"`
	}

	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if reqData.Count == 0 {
		reqData.Count = 5
	}

	questions, err := r.service.GenerateMusicTheoryQuiz(req.Context(), reqData.Difficulty, reqData.Count)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to generate quiz: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"questions": questions,
		"count":     len(questions),
	})
}

// AnalyzeTheory analyzes theory quiz results
func (r *Router) AnalyzeTheory(w http.ResponseWriter, req *http.Request) {
	sessionIDStr := chi.URLParam(req, "sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	analysis, err := r.service.AnalyzeMusicTheory(req.Context(), uint(sessionID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to analyze theory")
		return
	}

	respondJSON(w, http.StatusOK, analysis)
}

// UploadMIDI handles MIDI file uploads
func (r *Router) UploadMIDI(w http.ResponseWriter, req *http.Request) {
	// Parse multipart form
	if err := req.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse upload")
		return
	}

	file, _, err := req.FormFile("midi")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No MIDI file provided")
		return
	}
	defer file.Close()

	// Read MIDI data
	midiData := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			midiData = append(midiData, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"uploaded": true,
		"size":     len(midiData),
		"message":  "MIDI file uploaded successfully",
	})
}

// DownloadMIDI downloads a MIDI recording
func (r *Router) DownloadMIDI(w http.ResponseWriter, req *http.Request) {
	sessionIDStr := chi.URLParam(req, "sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	midiData, err := r.service.repo.GetMIDIRecording(req.Context(), uint(sessionID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get MIDI recording")
		return
	}

	w.Header().Set("Content-Type", "audio/midi")
	w.Header().Set("Content-Disposition", "attachment; filename=recording.mid")
	w.WriteHeader(http.StatusOK)
	w.Write(midiData)
}

// RecommendLesson recommends a lesson based on user level
func (r *Router) RecommendLesson(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get user progress to determine difficulty
	progress, err := r.service.repo.GetUserProgress(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user progress")
		return
	}

	// Recommend lesson based on current level
	difficulty := progress.EstimatedLevel
	if difficulty == "" {
		difficulty = "beginner"
	}

	song, err := r.service.GenerateLesson(req.Context(), uint(userID), difficulty)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to recommend lesson: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"recommended": song,
		"difficulty":  difficulty,
	})
}

// GetProgressionPath returns the learning progression path
func (r *Router) GetProgressionPath(w http.ResponseWriter, req *http.Request) {
	userIDStr := chi.URLParam(req, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	path, err := r.service.GetProgressionPath(req.Context(), uint(userID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get progression path")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"path": path,
	})
}

// GetLeaderboard retrieves the piano leaderboard
func (r *Router) GetLeaderboard(w http.ResponseWriter, req *http.Request) {
	limitStr := req.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"leaderboard": []interface{}{},
		"limit":       limit,
		"message":     "Leaderboard coming soon",
	})
}
