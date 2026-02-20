package math

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
)

// Router handles HTTP routes for math app
type Router struct {
	db      *sql.DB
	handler *Handler
	router  chi.Router
}

// NewRouter creates a new math router
func NewRouter(db *sql.DB) *Router {
	repo := NewRepository(db)
	handler := NewHandler(repo)

	return &Router{
		db:      db,
		handler: handler,
		router:  chi.NewRouter(),
	}
}

// Routes configures all math app routes and returns the chi.Router
func (r *Router) Routes() chi.Router {
	// ==================== CORE MATH PRACTICE ====================

	// Generate a new math question
	r.router.Get("/api/math/question", r.handler.GenerateQuestion)

	// Check answer and get feedback
	r.router.Post("/api/math/check-answer", r.handler.CheckAnswer)

	// Save a complete practice session
	r.router.Post("/api/math/save-session", r.handler.SaveSession)

	// Get user statistics
	r.router.Get("/api/users/{userId}/math/stats", r.handler.GetStats)

	// Get word/fact mastery information
	r.router.Get("/api/users/{userId}/math/mastery", r.handler.GetMastery)

	// ==================== PRACTICE MANAGEMENT ====================

	// Get facts due for SM-2 review
	r.router.Get("/api/users/{userId}/math/due-review", r.handler.GetDueForReview)

	// Process a review attempt
	r.router.Post("/api/users/{userId}/math/process-review", r.handler.ProcessReview)

	// Get an adaptive practice session (40% due, 60% new)
	r.router.Get("/api/users/{userId}/math/adaptive-session", r.handler.GetAdaptiveSession)

	// ==================== LEARNING ANALYTICS ====================

	// Get comprehensive learning analytics
	r.router.Get("/api/users/{userId}/math/analytics", r.handler.GetAnalytics)

	// Get weak fact families needing practice
	r.router.Get("/api/users/{userId}/math/weak-areas", r.handler.GetWeakAreas)

	// Get personalized practice plan
	r.router.Get("/api/users/{userId}/math/practice-plan", r.handler.GetPracticePlan)

	// Get learning profile analysis
	r.router.Get("/api/users/{userId}/math/learning-profile", r.handler.GetLearningProfile)

	// ==================== SPACED REPETITION ====================

	// Initialize SM-2 schedule for a fact
	r.router.Post("/api/users/{userId}/math/sr-initialize", r.handler.InitializeSR)

	// Get SM-2 progress and statistics
	r.router.Get("/api/users/{userId}/math/sr-progress", r.handler.GetSM2Progress)

	// ==================== ASSESSMENT ====================

	// Start a placement assessment (binary search, levels 1-15)
	r.router.Post("/api/users/{userId}/math/assessment-start", r.handler.StartAssessment)

	// Submit an assessment response
	r.router.Post("/api/users/{userId}/math/assessment-response", r.handler.SubmitAssessmentResponse)

	// Get assessment results and placement
	r.router.Get("/api/users/{userId}/math/assessment-results", r.handler.GetAssessmentResults)

	// ==================== FACT FAMILIES ====================

	// Detect which fact family a question belongs to
	r.router.Get("/api/math/detect-family", r.handler.DetectFactFamily)

	// Get fact family mastery statistics
	r.router.Get("/api/users/{userId}/math/family-stats", r.handler.GetFactFamilyStats)

	// Get remediation plan for weak areas
	r.router.Get("/api/users/{userId}/math/remediation-plan", r.handler.GetRemediationPlan)

	return r.router
}

// Handler returns the underlying handler
func (r *Router) Handler() chi.Router {
	return r.Routes()
}
