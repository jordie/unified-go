package typing

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/middleware"
	"github.com/jgirmay/GAIA_GO/internal/models"
	"github.com/jgirmay/GAIA_GO/internal/session"
)

// RegisterHandlers registers all typing app handlers with the router
func RegisterHandlers(router *gin.RouterGroup, app *TypingApp, sessionMgr *session.Manager) {
	// User endpoints
	router.GET("/current-user", app.handleGetCurrentUser)
	router.GET("/users", app.handleGetUsers)
	router.POST("/users", app.handleCreateUser(sessionMgr))

	// Practice endpoints
	router.GET("/text", app.handleGetText)
	router.POST("/save-result", middleware.RequireAuth(), app.handleSaveResult)
	router.GET("/stats", middleware.RequireAuth(), app.handleGetStats)
	router.GET("/leaderboard", app.handleGetLeaderboard)

	// Race endpoints
	router.POST("/race/start", app.handleRaceStart)
	router.POST("/race/finish", middleware.RequireAuth(), app.handleRaceFinish(sessionMgr))
	router.GET("/race/stats", app.handleRaceStats)
	router.GET("/race/leaderboard", app.handleRaceLeaderboard)
}

// ============================================================================
// USER ENDPOINTS
// ============================================================================

func (app *TypingApp) handleGetCurrentUser(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	username, _ := middleware.GetUsername(c)

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
	})
}

func (app *TypingApp) handleGetUsers(c *gin.Context) {
	rows, err := app.db.Query(`SELECT id, username FROM users ORDER BY username`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []gin.H
	for rows.Next() {
		var id int64
		var username string
		if scanErr := rows.Scan(&id, &username); scanErr != nil {
			continue
		}
		users = append(users, gin.H{
			"id":       id,
			"username": username,
		})
	}

	c.JSON(http.StatusOK, users)
}

func (app *TypingApp) handleCreateUser(sessionMgr *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}

		username := req.Username
		if len(username) < 2 || len(username) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username must be between 2 and 20 characters"})
			return
		}

		// Check if user exists
		var existingID int64
		err := app.db.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&existingID)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
			return
		}

		// Create user
		user, err := sessionMgr.CreateUser(username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		// Initialize typing stats
		_, _ = app.db.Exec(
			`INSERT INTO typing_stats (user_id, total_tests) VALUES (?, 0)`,
			user.ID,
		)

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"user_id":  user.ID,
			"username": user.Username,
		})
	}
}

// ============================================================================
// PRACTICE ENDPOINTS
// ============================================================================

func (app *TypingApp) handleGetText(c *gin.Context) {
	testType := c.DefaultQuery("type", "common_words")
	category := c.DefaultQuery("category", "common_words")
	wordCountStr := c.DefaultQuery("word_count", "25")
	wordCount, _ := strconv.Atoi(wordCountStr)

	text := app.GetText(testType, category, wordCount)

	c.JSON(http.StatusOK, gin.H{
		"text": text,
	})
}

func (app *TypingApp) handleSaveResult(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		WPM                 int     `json:"wpm" binding:"required"`
		Accuracy            float64 `json:"accuracy" binding:"required"`
		TestType            string  `json:"test_type"`
		TestDuration        int     `json:"test_duration"`
		TotalCharacters     int     `json:"total_characters"`
		CorrectCharacters   int     `json:"correct_characters"`
		IncorrectCharacters int     `json:"incorrect_characters"`
		RawWPM              int     `json:"raw_wpm"`
		Errors              int     `json:"errors"`
		TextSnippet         string  `json:"text_snippet"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Save result
	result, err := app.db.Exec(
		`INSERT INTO typing_results
		(user_id, wpm, raw_wpm, accuracy, test_type, test_duration,
		 total_characters, correct_characters, incorrect_characters, errors, text_snippet)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.WPM, req.RawWPM, req.Accuracy, req.TestType, req.TestDuration,
		req.TotalCharacters, req.CorrectCharacters, req.IncorrectCharacters, req.Errors, req.TextSnippet,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save result"})
		return
	}

	// Update user stats
	var stats struct {
		ID              int64
		TotalTests      int
		AverageWPM      float64
		AverageAccuracy float64
		BestWPM         int
	}

	err = app.db.QueryRow(
		`SELECT id, total_tests, average_wpm, average_accuracy, best_wpm
		 FROM typing_stats WHERE user_id = ?`,
		userID,
	).Scan(&stats.ID, &stats.TotalTests, &stats.AverageWPM, &stats.AverageAccuracy, &stats.BestWPM)

	newTotalTests := stats.TotalTests + 1
	newAverageWPM := (stats.AverageWPM*float64(stats.TotalTests) + float64(req.WPM)) / float64(newTotalTests)
	newAverageAccuracy := (stats.AverageAccuracy*float64(stats.TotalTests) + req.Accuracy) / float64(newTotalTests)
	newBestWPM := stats.BestWPM
	if req.WPM > newBestWPM {
		newBestWPM = req.WPM
	}

	_, err = app.db.Exec(
		`UPDATE typing_stats
		 SET total_tests = ?, average_wpm = ?, average_accuracy = ?, best_wpm = ?,
		     last_updated = CURRENT_TIMESTAMP
		 WHERE user_id = ?`,
		newTotalTests, newAverageWPM, newAverageAccuracy, newBestWPM, userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update stats"})
		return
	}

	resultID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"result_id": resultID,
		"message":   "Result saved successfully",
	})
}

func (app *TypingApp) handleGetStats(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	// Get user stats
	var stats models.TypingStats
	err := app.db.QueryRow(
		`SELECT id, user_id, total_tests, average_wpm, average_accuracy, best_wpm, total_time_typed
		 FROM typing_stats WHERE user_id = ?`,
		userID,
	).Scan(&stats.ID, &stats.UserID, &stats.TotalSessions, &stats.AverageWPM, &stats.AverageAccuracy, &stats.BestWPM, &stats.TotalTime)

	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
		return
	}

	// Get recent results
	rows, err := app.db.Query(
		`SELECT wpm, accuracy, test_type, created_at FROM typing_results
		 WHERE user_id = ? ORDER BY created_at DESC LIMIT 10`,
		userID,
	)
	defer rows.Close()

	var recentResults []gin.H
	for rows.Next() {
		var wpm int
		var accuracy float64
		var testType string
		var createdAt time.Time
		if err := rows.Scan(&wpm, &accuracy, &testType, &createdAt); err != nil {
			continue
		}
		recentResults = append(recentResults, gin.H{
			"wpm":        wpm,
			"accuracy":   accuracy,
			"test_type":  testType,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_stats":     stats,
		"recent_results": recentResults,
	})
}

func (app *TypingApp) handleGetLeaderboard(c *gin.Context) {
	// Top WPM scores
	topWPM, _ := app.db.Query(
		`SELECT r.wpm, r.accuracy, r.test_type, r.created_at, u.username
		 FROM typing_results r
		 JOIN users u ON r.user_id = u.id
		 ORDER BY r.wpm DESC LIMIT 10`,
	)
	defer topWPM.Close()

	var topWPMList []gin.H
	for topWPM.Next() {
		var wpm int
		var accuracy float64
		var testType string
		var createdAt time.Time
		var username string
		if err := topWPM.Scan(&wpm, &accuracy, &testType, &createdAt, &username); err != nil {
			continue
		}
		topWPMList = append(topWPMList, gin.H{
			"wpm":        wpm,
			"accuracy":   accuracy,
			"test_type":  testType,
			"created_at": createdAt,
			"username":   username,
		})
	}

	// Top accuracy scores
	topAcc, _ := app.db.Query(
		`SELECT r.wpm, r.accuracy, r.test_type, r.created_at, u.username
		 FROM typing_results r
		 JOIN users u ON r.user_id = u.id
		 WHERE r.wpm >= 30
		 ORDER BY r.accuracy DESC, r.wpm DESC LIMIT 10`,
	)
	defer topAcc.Close()

	var topAccList []gin.H
	for topAcc.Next() {
		var wpm int
		var accuracy float64
		var testType string
		var createdAt time.Time
		var username string
		if err := topAcc.Scan(&wpm, &accuracy, &testType, &createdAt, &username); err != nil {
			continue
		}
		topAccList = append(topAccList, gin.H{
			"wpm":        wpm,
			"accuracy":   accuracy,
			"test_type":  testType,
			"created_at": createdAt,
			"username":   username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"top_wpm":      topWPMList,
		"top_accuracy": topAccList,
	})
}

// ============================================================================
// RACE ENDPOINTS
// ============================================================================

func (app *TypingApp) handleRaceStart(c *gin.Context) {
	var req struct {
		WordCount  int    `json:"word_count"`
		Difficulty string `json:"difficulty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.WordCount = 30
		req.Difficulty = "medium"
	}

	text, opponents := app.GenerateRaceOpponents(req.Difficulty, req.WordCount)

	c.JSON(http.StatusOK, gin.H{
		"text":        text,
		"word_count":  req.WordCount,
		"difficulty":  req.Difficulty,
		"opponents":   opponents,
	})
}

func (app *TypingApp) handleRaceFinish(sessionMgr *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c)

		var req struct {
			WPM        int     `json:"wpm" binding:"required"`
			Accuracy   float64 `json:"accuracy" binding:"required"`
			Placement  int     `json:"placement" binding:"required"`
			RaceTime   float64 `json:"race_time"`
			Difficulty string  `json:"difficulty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Calculate XP
		xpEarned := CalculateRaceXP(req.WPM, req.Accuracy, req.Placement, req.Difficulty)

		// Save race result
		_, err := app.db.Exec(
			`INSERT INTO races (user_id, difficulty, placement, wpm, accuracy, race_time, xp_earned)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			userID, req.Difficulty, req.Placement, req.WPM, req.Accuracy, req.RaceTime, xpEarned,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save race"})
			return
		}

		// Update racing stats
		var raceStats struct {
			ID       int64
			Wins     int
			Podiums  int
			TotalXP  int
		}

		err = app.db.QueryRow(
			`SELECT id, wins, podiums, total_xp FROM racing_stats WHERE user_id = ?`,
			userID,
		).Scan(&raceStats.ID, &raceStats.Wins, &raceStats.Podiums, &raceStats.TotalXP)

		winsDelta := 0
		if req.Placement == 1 {
			winsDelta = 1
		}
		podiumsDelta := 0
		if req.Placement <= 3 {
			podiumsDelta = 1
		}

		if err == nil {
			_, _ = app.db.Exec(
				`UPDATE racing_stats
				 SET wins = wins + ?, podiums = podiums + ?, total_xp = total_xp + ?
				 WHERE user_id = ?`,
				winsDelta, podiumsDelta, xpEarned, userID,
			)
		} else {
			wins := 0
			if req.Placement == 1 {
				wins = 1
			}
			podiums := 0
			if req.Placement <= 3 {
				podiums = 1
			}
			_, _ = app.db.Exec(
				`INSERT INTO racing_stats (user_id, total_races, wins, podiums, total_xp)
				 VALUES (?, 1, ?, ?, ?)`,
				userID, wins, podiums, xpEarned,
			)
		}

		// Add XP to user
		_ = sessionMgr.AddUserXP(userID, int64(xpEarned))

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"xp_earned":  xpEarned,
			"placement":  req.Placement,
		})
	}
}

func (app *TypingApp) handleRaceStats(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil || userID == 0 {
		c.JSON(http.StatusOK, gin.H{
			"total_races": 0,
			"wins":        0,
			"podiums":     0,
			"total_xp":    0,
			"avg_wpm":     0,
			"current_car": "🚗",
		})
		return
	}

	var stats struct {
		TotalRaces int
		Wins       int
		Podiums    int
		TotalXP    int
	}

	err = app.db.QueryRow(
		`SELECT total_races, wins, podiums, total_xp FROM racing_stats WHERE user_id = ?`,
		userID,
	).Scan(&stats.TotalRaces, &stats.Wins, &stats.Podiums, &stats.TotalXP)

	// Calculate average WPM from races
	var avgWPM float64
	_ = app.db.QueryRow(
		`SELECT AVG(wpm) FROM races WHERE user_id = ?`,
		userID,
	).Scan(&avgWPM)

	car := GetCarForXP(stats.TotalXP)

	c.JSON(http.StatusOK, gin.H{
		"total_races": stats.TotalRaces,
		"wins":        stats.Wins,
		"podiums":     stats.Podiums,
		"total_xp":    stats.TotalXP,
		"avg_wpm":     int(avgWPM),
		"current_car": car,
	})
}

func (app *TypingApp) handleRaceLeaderboard(c *gin.Context) {
	// Top racers by wins
	topWins, _ := app.db.Query(
		`SELECT u.username, s.wins, s.total_races, s.total_xp
		 FROM racing_stats s
		 JOIN users u ON s.user_id = u.id
		 WHERE s.total_races > 0
		 ORDER BY s.wins DESC LIMIT 10`,
	)
	defer topWins.Close()

	var topWinsList []gin.H
	for topWins.Next() {
		var username string
		var wins, totalRaces, totalXP int
		if err := topWins.Scan(&username, &wins, &totalRaces, &totalXP); err != nil {
			continue
		}
		topWinsList = append(topWinsList, gin.H{
			"username":    username,
			"wins":        wins,
			"total_races": totalRaces,
			"total_xp":    totalXP,
		})
	}

	// Top racers by XP
	topXP, _ := app.db.Query(
		`SELECT u.username, s.total_xp, s.wins, s.total_races
		 FROM racing_stats s
		 JOIN users u ON s.user_id = u.id
		 WHERE s.total_races > 0
		 ORDER BY s.total_xp DESC LIMIT 10`,
	)
	defer topXP.Close()

	var topXPList []gin.H
	for topXP.Next() {
		var username string
		var totalXP, wins, totalRaces int
		if err := topXP.Scan(&username, &totalXP, &wins, &totalRaces); err != nil {
			continue
		}
		topXPList = append(topXPList, gin.H{
			"username":    username,
			"total_xp":    totalXP,
			"wins":        wins,
			"total_races": totalRaces,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"top_wins": topWinsList,
		"top_xp":   topXPList,
	})
}
