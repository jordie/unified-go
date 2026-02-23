package math

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/middleware"
	"github.com/jgirmay/GAIA_GO/internal/session"
)

// RegisterHandlers registers all math app routes
func RegisterHandlers(router *gin.RouterGroup, app *MathApp, sessionMgr *session.Manager) {
	// Problem generation endpoints
	router.POST("/api/generate_problem", func(c *gin.Context) {
		handleGenerateProblem(c, app)
	})

	// Speech-to-text endpoints (CRITICAL)
	router.POST("/api/transcribe", func(c *gin.Context) {
		handleTranscribeAudio(c)
	})

	router.POST("/api/check_speech_answer", func(c *gin.Context) {
		handleCheckSpeechAnswer(c)
	})

	// Answer checking
	router.POST("/api/check_answer", func(c *gin.Context) {
		handleCheckAnswer(c, app, sessionMgr)
	})

	// Practice session endpoints
	router.POST("/api/save_session", func(c *gin.Context) {
		handleSaveSession(c, app, sessionMgr)
	})

	// Stats endpoints
	router.GET("/api/stats", func(c *gin.Context) {
		handleGetStats(c, app, sessionMgr)
	})

	router.GET("/api/leaderboard", func(c *gin.Context) {
		handleGetLeaderboard(c, app)
	})
}

// handleGenerateProblem generates a new math problem
func handleGenerateProblem(c *gin.Context, app *MathApp) {
	var req struct {
		Operation  string `json:"operation" binding:"required"`
		Difficulty string `json:"difficulty" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	problem := app.GenerateProblem(req.Operation, req.Difficulty)
	c.JSON(http.StatusOK, problem)
}

// ============================================================================
// SPEECH-TO-TEXT HANDLERS (CRITICAL)
// ============================================================================

// handleTranscribeAudio transcribes audio using Whisper API
// This is the critical endpoint for speech-to-text
func handleTranscribeAudio(c *gin.Context) {
	// Get audio file from multipart form
	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No audio file provided",
		})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read audio file",
		})
		return
	}
	defer src.Close()

	// Read file content
	audioData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read audio data",
		})
		return
	}

	// Transcribe using Whisper API
	transcribedText, confidence, err := TranscribeAudio(audioData, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Transcription failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"text":       transcribedText,
		"confidence": confidence,
	})
}

// handleCheckSpeechAnswer checks if a spoken answer matches the expected answer
func handleCheckSpeechAnswer(c *gin.Context) {
	var req struct {
		SpokenText      string  `json:"spoken_text" binding:"required"`
		ExpectedAnswer  float64 `json:"expected_answer" binding:"required"`
		Tolerance       float64 `json:"tolerance"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	// Default tolerance to 0.01 if not provided
	tolerance := req.Tolerance
	if tolerance == 0 {
		tolerance = 0.01
	}

	// Check the answer
	result := CheckSpeechAnswer(req.SpokenText, req.ExpectedAnswer, tolerance)

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"match":             result.IsMatch,
		"spoken_number":     result.SpokenNumber,
		"expected_number":   result.ExpectedNumber,
		"match_type":        result.MatchType,
		"feedback":          result.Feedback,
		"score":             result.Score,
	})
}

// ============================================================================
// ANSWER AND SESSION HANDLERS
// ============================================================================

// handleCheckAnswer checks a regular typed answer
func handleCheckAnswer(c *gin.Context, app *MathApp, sessionMgr *session.Manager) {
	var req struct {
		ProblemID      string  `json:"problem_id" binding:"required"`
		UserAnswer     float64 `json:"user_answer" binding:"required"`
		ExpectedAnswer float64 `json:"expected_answer" binding:"required"`
		TimeTaken      int     `json:"time_taken"` // seconds
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check if answer is correct (with small tolerance for floating point)
	isCorrect := CheckAnswerCorrect(req.UserAnswer, req.ExpectedAnswer, 0.01)

	c.JSON(http.StatusOK, gin.H{
		"correct":          isCorrect,
		"expected_answer":  req.ExpectedAnswer,
		"user_answer":      req.UserAnswer,
		"time_taken":       req.TimeTaken,
	})
}

// handleSaveSession saves a practice session
func handleSaveSession(c *gin.Context, app *MathApp, sessionMgr *session.Manager) {
	userID, err := middleware.GetUserID(c)
	if err != nil || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Operation       string `json:"operation" binding:"required"`
		Difficulty      string `json:"difficulty" binding:"required"`
		TotalQuestions  int    `json:"total_questions" binding:"required"`
		CorrectAnswers  int    `json:"correct_answers" binding:"required"`
		TotalTime       int    `json:"total_time"` // seconds
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	accuracy := CalculateAccuracy(req.CorrectAnswers, req.TotalQuestions)
	averageTime := CalculateAverageTime(req.TotalTime, req.TotalQuestions)

	// Save to database
	result, err := app.db.Exec(`
		INSERT INTO math_results (
			user_id, operation, difficulty, total_questions, correct_answers,
			accuracy, average_time, total_time, mode
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'practice')
	`,
		userID, req.Operation, req.Difficulty, req.TotalQuestions,
		req.CorrectAnswers, accuracy, averageTime, req.TotalTime,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	resultID, _ := result.LastInsertId()

	// Calculate and award XP
	xpEarned := int64(CalculateMathXP(fmt.Sprintf("%.1f%%", accuracy), req.Difficulty))
	_ = sessionMgr.AddUserXP(userID, xpEarned)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"result_id":     resultID,
		"accuracy":      accuracy,
		"average_time":  averageTime,
		"xp_earned":     xpEarned,
	})
}

// handleGetStats retrieves user statistics
func handleGetStats(c *gin.Context, app *MathApp, sessionMgr *session.Manager) {
	userID, err := middleware.GetUserID(c)
	if err != nil || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var stats struct {
		TotalProblemsSolved int     `db:"total_problems_solved"`
		AverageAccuracy     float64 `db:"average_accuracy"`
		BestAccuracy        float64 `db:"best_accuracy"`
		TotalTimeSpent      int     `db:"total_time_spent"`
	}

	dbErr := app.db.QueryRow(`
		SELECT COALESCE(total_problems_solved, 0),
		       COALESCE(average_accuracy, 0),
		       COALESCE(best_accuracy, 0),
		       COALESCE(total_time_spent, 0)
		FROM math_stats
		WHERE user_id = ?
	`, userID).Scan(&stats.TotalProblemsSolved, &stats.AverageAccuracy,
		&stats.BestAccuracy, &stats.TotalTimeSpent)

	if dbErr != nil {
		// User has no stats yet
		stats.TotalProblemsSolved = 0
		stats.AverageAccuracy = 0
		stats.BestAccuracy = 0
		stats.TotalTimeSpent = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"total_problems_solved": stats.TotalProblemsSolved,
		"average_accuracy":      stats.AverageAccuracy,
		"best_accuracy":         stats.BestAccuracy,
		"total_time_spent":      stats.TotalTimeSpent,
	})
}

// handleGetLeaderboard retrieves the math leaderboard
func handleGetLeaderboard(c *gin.Context, app *MathApp) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 {
		limit = 10
	}

	rows, err := app.db.Query(`
		SELECT u.username, ms.average_accuracy, ms.total_problems_solved, ms.total_time_spent
		FROM math_stats ms
		JOIN users u ON ms.user_id = u.id
		ORDER BY ms.average_accuracy DESC, ms.total_problems_solved DESC
		LIMIT ?
	`, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	defer rows.Close()

	type LeaderboardEntry struct {
		Username             string  `json:"username"`
		AverageAccuracy      float64 `json:"average_accuracy"`
		TotalProblemsSolved  int     `json:"total_problems_solved"`
		TotalTimeSpent       int     `json:"total_time_spent"`
	}

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.AverageAccuracy,
			&entry.TotalProblemsSolved, &entry.TotalTimeSpent); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"leaderboard":  entries,
		"entry_count":  len(entries),
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// CheckAnswerCorrect checks if a user's answer matches the expected answer
func CheckAnswerCorrect(userAnswer, expectedAnswer, tolerance float64) bool {
	diff := userAnswer - expectedAnswer
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
