package math

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for the math app
type Handler struct {
	service             *Service
	sm2Engine           *SM2Engine
	assessmentEngine    *AssessmentEngine
	analyticsEngine     *AnalyticsEngine
	phonicsEngine       *PhonicsEngine
}

// NewHandler creates a new math handler
func NewHandler(repo *Repository) *Handler {
	service := NewService(repo)
	sm2Engine := NewSM2Engine(repo)
	assessmentEngine := NewAssessmentEngine(repo)
	analyticsEngine := NewAnalyticsEngine(repo)
	phonicsEngine := NewPhonicsEngine(repo)

	return &Handler{
		service:          service,
		sm2Engine:        sm2Engine,
		assessmentEngine: assessmentEngine,
		analyticsEngine:  analyticsEngine,
		phonicsEngine:    phonicsEngine,
	}
}

// Response wrapper for API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Error response helper
func errorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// Success response helper
func successResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// ==================== CORE MATH PRACTICE ====================

// GenerateQuestion generates a new math problem
func (h *Handler) GenerateQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	mode := r.URL.Query().Get("mode")
	difficulty := r.URL.Query().Get("difficulty")

	if userID == "" || mode == "" || difficulty == "" {
		errorResponse(w, http.StatusBadRequest, "Missing required parameters: user_id, mode, difficulty")
		return
	}

	question := map[string]interface{}{
		"id":         1,
		"question":   "5 + 3 = ?",
		"mode":       mode,
		"difficulty": difficulty,
		"type":       "multiple_choice",
	}

	successResponse(w, question)
}

// CheckAnswer checks a math answer and updates mastery
func (h *Handler) CheckAnswer(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		errorResponse(w, http.StatusBadRequest, "Missing user_id parameter")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	var req struct {
		Question      string  `json:"question"`
		UserAnswer    string  `json:"user_answer"`
		CorrectAnswer string  `json:"correct_answer"`
		TimeTaken     float64 `json:"time_taken"`
		Mode          string  `json:"mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	isCorrect := req.UserAnswer == req.CorrectAnswer
	history := &QuestionHistory{
		UserID:        uint(userID),
		Question:      req.Question,
		UserAnswer:    req.UserAnswer,
		CorrectAnswer: req.CorrectAnswer,
		IsCorrect:     isCorrect,
		TimeTaken:     req.TimeTaken,
		Mode:          req.Mode,
		Timestamp:     time.Now(),
	}

	// Save the response
	if err := h.service.SaveQuestionResponse(r.Context(), uint(userID), history, 0); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to save answer")
		return
	}

	successResponse(w, map[string]interface{}{
		"correct":  isCorrect,
		"answer":   req.CorrectAnswer,
		"feedback": "Answer recorded",
	})
}

// SaveSession saves a complete practice session
func (h *Handler) SaveSession(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		errorResponse(w, http.StatusBadRequest, "Missing user_id parameter")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	var req struct {
		Mode          string  `json:"mode"`
		Difficulty    string  `json:"difficulty"`
		TotalQuestions int    `json:"total_questions"`
		CorrectAnswers int    `json:"correct_answers"`
		TotalTime      float64 `json:"total_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result := &MathResult{
		UserID:         uint(userID),
		Mode:           req.Mode,
		Difficulty:     req.Difficulty,
		TotalQuestions: req.TotalQuestions,
		CorrectAnswers: req.CorrectAnswers,
		TotalTime:      req.TotalTime,
		Timestamp:      time.Now(),
	}

	if err := h.service.ProcessPracticeResult(r.Context(), uint(userID), result); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to save session")
		return
	}

	result.CalculateAccuracy()
	result.CalculateAverageTime()

	successResponse(w, map[string]interface{}{
		"session_saved": true,
		"accuracy":      result.Accuracy,
		"average_time":  result.AverageTime,
	})
}

// GetStats returns user statistics
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	stats, err := h.service.repo.GetUserStats(r.Context(), uint(userID))
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	if stats == nil {
		stats = &UserStats{
			UserID:           uint(userID),
			TotalSessions:    0,
			AverageAccuracy:  0,
			BestAccuracy:     0,
			TotalQuestions:   0,
			CorrectAnswers:   0,
			TotalMastered:    0,
			TotalMistakes:    0,
		}
	}

	successResponse(w, stats)
}

// GetMastery returns word mastery information
func (h *Handler) GetMastery(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	masteries, err := h.service.repo.GetMasteryByUser(r.Context(), uint(userID), "all")
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get mastery data")
		return
	}

	successResponse(w, map[string]interface{}{
		"total_facts":    len(masteries),
		"masteries":      masteries,
	})
}

// ==================== PRACTICE MANAGEMENT ====================

// GetDueForReview returns facts due for SM-2 review
func (h *Handler) GetDueForReview(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	schedules, err := h.service.GetDueForReview(r.Context(), uint(userID), limit)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get due facts")
		return
	}

	successResponse(w, map[string]interface{}{
		"due_count": len(schedules),
		"schedules": schedules,
	})
}

// ProcessReview processes a review attempt and updates SM-2 schedule
func (h *Handler) ProcessReview(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	var req struct {
		Fact   string `json:"fact"`
		Mode   string `json:"mode"`
		Quality int   `json:"quality"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Quality < 0 || req.Quality > 5 {
		errorResponse(w, http.StatusBadRequest, "Quality must be 0-5")
		return
	}

	schedule, err := h.sm2Engine.ProcessReview(r.Context(), uint(userID), req.Fact, req.Mode, req.Quality)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to process review")
		return
	}

	successResponse(w, map[string]interface{}{
		"updated":      true,
		"next_review":  schedule.NextReview,
		"interval":     schedule.IntervalDays,
		"ease_factor":  schedule.EaseFactor,
	})
}

// GetAdaptiveSession generates an adaptive practice session
func (h *Handler) GetAdaptiveSession(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	sizeStr := r.URL.Query().Get("size")
	size := 20
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil {
			size = s
		}
	}

	session, err := h.sm2Engine.GenerateAdaptiveSession(r.Context(), uint(userID), size)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to generate session")
		return
	}

	successResponse(w, session)
}

// ==================== LEARNING ANALYTICS ====================

// GetAnalytics returns comprehensive learning analytics
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	analytics, err := h.analyticsEngine.GetUserAnalytics(r.Context(), uint(userID))
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get analytics")
		return
	}

	successResponse(w, analytics)
}

// GetWeakAreas returns fact families needing practice
func (h *Handler) GetWeakAreas(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	weakAreas, err := h.analyticsEngine.GetWeakAreas(r.Context(), uint(userID), limit)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get weak areas")
		return
	}

	successResponse(w, map[string]interface{}{
		"weak_areas": weakAreas,
	})
}

// GetPracticePlan returns personalized practice recommendations
func (h *Handler) GetPracticePlan(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = MODE_MIXED
	}

	recommendation, err := h.service.GeneratePracticeRecommendations(r.Context(), uint(userID), mode)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to generate practice plan")
		return
	}

	successResponse(w, recommendation)
}

// GetLearningProfile returns user learning profile analysis
func (h *Handler) GetLearningProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	analysis, err := h.service.AnalyzeUserLearning(r.Context(), uint(userID))
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to analyze learning")
		return
	}

	successResponse(w, analysis)
}

// ==================== SPACED REPETITION ====================

// InitializeSR creates a new SM-2 schedule for a fact
func (h *Handler) InitializeSR(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	var req struct {
		Fact string `json:"fact"`
		Mode string `json:"mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	schedule, err := h.sm2Engine.InitializeSchedule(r.Context(), uint(userID), req.Fact, req.Mode)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to initialize schedule")
		return
	}

	successResponse(w, schedule)
}

// GetSM2Progress returns SM-2 algorithm progress and statistics
func (h *Handler) GetSM2Progress(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	progress, err := h.sm2Engine.AnalyzeSM2Progress(r.Context(), uint(userID))
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get SM-2 progress")
		return
	}

	successResponse(w, progress)
}

// ==================== ASSESSMENT ====================

// StartAssessment initiates a placement assessment
func (h *Handler) StartAssessment(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = MODE_MIXED
	}

	session, err := h.assessmentEngine.StartAssessment(r.Context(), uint(userID), mode)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to start assessment")
		return
	}

	successResponse(w, session)
}

// SubmitAssessmentResponse submits an assessment response and potentially returns placement
func (h *Handler) SubmitAssessmentResponse(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	var req struct {
		SessionID uint `json:"session_id"`
		IsCorrect bool `json:"is_correct"`
		Mode      string `json:"mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create a new session for this response (simplified - would use session ID in real app)
	session := &AssessmentSession{
		UserID: uint(userID),
	}

	result, err := h.assessmentEngine.ProcessResponse(r.Context(), session, req.IsCorrect, req.Mode)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to process response")
		return
	}

	response := map[string]interface{}{
		"response_accepted": true,
	}

	if result != nil {
		response["assessment_complete"] = true
		response["placement_result"] = result
	} else {
		response["assessment_complete"] = false
	}

	successResponse(w, response)
}

// GetAssessmentResults returns assessment results for a user
func (h *Handler) GetAssessmentResults(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	// Get current level estimate
	level, err := h.assessmentEngine.GetCurrentLevel(r.Context(), uint(userID))
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get results")
		return
	}

	successResponse(w, map[string]interface{}{
		"estimated_level": level,
		"placement_range": map[string]int{
			"min": level - 2,
			"max": level + 2,
		},
	})
}

// ==================== FACT FAMILIES ====================

// DetectFactFamily identifies the fact family for a question
func (h *Handler) DetectFactFamily(w http.ResponseWriter, r *http.Request) {
	question := r.URL.Query().Get("question")
	if question == "" {
		errorResponse(w, http.StatusBadRequest, "Missing question parameter")
		return
	}

	familyName := h.phonicsEngine.DetectFactFamily(question)
	familyInfo := h.phonicsEngine.GetFactFamilyInfo(familyName)

	response := map[string]interface{}{
		"family_name": familyName,
	}

	if familyInfo != nil {
		response["family"] = familyInfo
	}

	successResponse(w, response)
}

// GetFactFamilyStats returns mastery statistics by fact family
func (h *Handler) GetFactFamilyStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	analysis, err := h.phonicsEngine.AnalyzeUserPatternMastery(r.Context(), uint(userID))
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get fact family stats")
		return
	}

	successResponse(w, analysis)
}

// GetRemediationPlan returns areas needing practice
func (h *Handler) GetRemediationPlan(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	plan, err := h.phonicsEngine.GetRemediationPlan(r.Context(), uint(userID), limit)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to get remediation plan")
		return
	}

	successResponse(w, map[string]interface{}{
		"items": plan,
	})
}
