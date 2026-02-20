package unified

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository handles aggregation of data from all educational app repositories
type Repository struct {
	db *sql.DB

	// References to app repositories will be injected
	typingRepo  interface{} // *typing.Repository
	mathRepo    interface{} // *math.Repository
	readingRepo interface{} // *reading.Repository
	pianoRepo   interface{} // *piano.Repository
}

// NewRepository creates a new unified repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// SetAppRepositories registers the app-specific repositories
func (r *Repository) SetAppRepositories(typing, math, reading, piano interface{}) {
	r.typingRepo = typing
	r.mathRepo = math
	r.readingRepo = reading
	r.pianoRepo = piano
}

// GetUserProfile fetches and aggregates user data from all apps
func (r *Repository) GetUserProfile(ctx context.Context, userID uint) (*UnifiedUserProfile, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	profile := &UnifiedUserProfile{
		UserID:           userID,
		WeeklyActiveApps: make([]string, 0),
		TypingStats:      nil,
		MathStats:        nil,
		ReadingStats:     nil,
		PianoStats:       nil,
	}

	// Get user basic info
	var username *string
	var accountCreated *time.Time
	err := r.db.QueryRowContext(ctx, `SELECT username, created_at FROM users WHERE id = ?`, userID).
		Scan(&username, &accountCreated)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if username != nil {
		profile.Username = *username
	}
	if accountCreated != nil {
		profile.AccountCreated = *accountCreated
	} else {
		profile.AccountCreated = time.Now()
	}

	// Aggregate metrics from all apps
	// TODO: Fetch from each app's repository when they're registered
	// For now, initialize with defaults
	profile.TotalSessionsAll = 0
	profile.TotalPracticeMinutes = 0
	profile.TypingLevel = 0
	profile.MathLevel = 0
	profile.ReadingLevel = 0
	profile.PianoLevel = 0
	profile.OverallLevel = 0
	profile.DailyStreakDays = 0

	// Get last activity
	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(created_at), CURRENT_TIMESTAMP) FROM (
			SELECT created_at FROM typing_tests WHERE user_id = ?
			UNION ALL
			SELECT timestamp as created_at FROM math_results WHERE user_id = ?
			UNION ALL
			SELECT created_at FROM reading_sessions WHERE user_id = ?
			UNION ALL
			SELECT created_at FROM piano_lessons WHERE user_id = ?
		)`, userID, userID, userID, userID).Scan(&profile.LastActivityDate)
	if err != nil && err != sql.ErrNoRows {
		profile.LastActivityDate = time.Now()
	}

	return profile, nil
}

// GetCrossAppAnalytics calculates cross-app insights for a user
func (r *Repository) GetCrossAppAnalytics(ctx context.Context, userID uint) (*CrossAppAnalytics, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	analytics := &CrossAppAnalytics{
		UserID:              userID,
		WeeklyProgress:      make(map[string]float64),
		MonthlyProgress:     make(map[string]float64),
		AppMetrics:          make(map[string]*AppMetricsSummary),
	}

	// Count sessions per app
	typingCount, _ := r.countSessions(ctx, "typing_tests", userID, 7)
	mathCount, _ := r.countSessions(ctx, "math_results", userID, 7)
	readingCount, _ := r.countSessions(ctx, "reading_sessions", userID, 7)
	pianoCount, _ := r.countSessions(ctx, "piano_lessons", userID, 7)

	// Determine most practiced app
	appCounts := map[string]int{
		"typing":  typingCount,
		"math":    mathCount,
		"reading": readingCount,
		"piano":   pianoCount,
	}

	maxCount := 0
	var mostPracticedApp string
	for app, count := range appCounts {
		if count > maxCount {
			maxCount = count
			mostPracticedApp = app
		}
		if count > 0 {
			analytics.TotalAppsPracticed++
		}
	}

	analytics.MostPracticedApp = mostPracticedApp

	// Initialize app metrics
	for app := range appCounts {
		analytics.AppMetrics[app] = &AppMetricsSummary{
			App: app,
		}
	}

	// Get total practice time (in minutes)
	totalSeconds := 0.0
	r.getTotalDuration(ctx, "typing_tests", userID, &totalSeconds)
	r.getTotalDuration(ctx, "math_results", userID, &totalSeconds)
	r.getTotalDuration(ctx, "reading_sessions", userID, &totalSeconds)
	r.getTotalDuration(ctx, "piano_lessons", userID, &totalSeconds)
	analytics.TotalHoursPracticed = totalSeconds / 3600.0

	// Calculate trends
	analytics.WeeklyProgress["typing"] = 0
	analytics.WeeklyProgress["math"] = 0
	analytics.WeeklyProgress["reading"] = 0
	analytics.WeeklyProgress["piano"] = 0

	analytics.MonthlyProgress["typing"] = 0
	analytics.MonthlyProgress["math"] = 0
	analytics.MonthlyProgress["reading"] = 0
	analytics.MonthlyProgress["piano"] = 0

	return analytics, nil
}

// GetRecentSessions fetches recent sessions across all apps
func (r *Repository) GetRecentSessions(ctx context.Context, userID uint, limit int) ([]UnifiedSession, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sessions := make([]UnifiedSession, 0)

	// Fetch from each app and combine
	// This is a simplified version - in production, this would use UNION queries
	// or separate repository calls

	// Typing sessions
	typingSessions, _ := r.getTypingSessions(ctx, userID, limit)
	sessions = append(sessions, typingSessions...)

	// Math sessions
	mathSessions, _ := r.getMathSessions(ctx, userID, limit)
	sessions = append(sessions, mathSessions...)

	// Reading sessions
	readingSessions, _ := r.getReadingSessions(ctx, userID, limit)
	sessions = append(sessions, readingSessions...)

	// Piano sessions
	pianoSessions, _ := r.getPianoSessions(ctx, userID, limit)
	sessions = append(sessions, pianoSessions...)

	return sessions, nil
}

// GetUnifiedLeaderboard fetches cross-app rankings for a category
func (r *Repository) GetUnifiedLeaderboard(ctx context.Context, category string, limit int) (*UnifiedLeaderboard, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	lb := &UnifiedLeaderboard{
		Category:  category,
		Entries:   make([]LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}

	// Handle different category types
	switch category {
	case "typing_wpm":
		entries, _ := r.getTopTypingWPM(ctx, limit)
		lb.Entries = entries
	case "math_accuracy":
		entries, _ := r.getTopMathAccuracy(ctx, limit)
		lb.Entries = entries
	case "reading_comprehension":
		entries, _ := r.getTopReadingComprehension(ctx, limit)
		lb.Entries = entries
	case "piano_score":
		entries, _ := r.getTopPianoScore(ctx, limit)
		lb.Entries = entries
	case "overall":
		entries, _ := r.getOverallLeaderboard(ctx, limit)
		lb.Entries = entries
	default:
		return nil, fmt.Errorf("unknown leaderboard category: %s", category)
	}

	return lb, nil
}

// GetSystemStats returns platform-wide statistics
func (r *Repository) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{
		AppUsageCount:    make(map[string]int),
		AppAverageScore:  make(map[string]float64),
	}

	// Get total user count
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)

	// Get active users
	today := time.Now().AddDate(0, 0, -1)
	r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM typing_tests WHERE timestamp > ? UNION
		 SELECT COUNT(DISTINCT user_id) FROM math_results WHERE timestamp > ? UNION
		 SELECT COUNT(DISTINCT user_id) FROM reading_sessions WHERE created_at > ? UNION
		 SELECT COUNT(DISTINCT user_id) FROM piano_lessons WHERE created_at > ?`,
		today, today, today, today).Scan(&stats.ActiveUsersToday)

	// Initialize app usage counts
	stats.AppUsageCount["typing"] = 0
	stats.AppUsageCount["math"] = 0
	stats.AppUsageCount["reading"] = 0
	stats.AppUsageCount["piano"] = 0

	return stats, nil
}

// Helper methods

func (r *Repository) countSessions(ctx context.Context, table string, userID uint, days int) (int, error) {
	var count int
	since := time.Now().AddDate(0, 0, -days)

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE user_id = ? AND created_at > ?`, table)
	if table == "typing_tests" || table == "math_results" {
		query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE user_id = ? AND timestamp > ?`, table)
	}

	err := r.db.QueryRowContext(ctx, query, userID, since).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return count, nil
}

func (r *Repository) getTotalDuration(ctx context.Context, table string, userID uint, totalSeconds *float64) error {
	var duration float64
	query := fmt.Sprintf(`SELECT COALESCE(SUM(duration), 0) FROM %s WHERE user_id = ?`, table)
	if table == "math_results" {
		query = fmt.Sprintf(`SELECT COALESCE(SUM(total_time), 0) FROM %s WHERE user_id = ?`, table)
	}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&duration)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get duration: %w", err)
	}

	*totalSeconds += duration
	return nil
}

// Placeholder methods for fetching sessions from each app
func (r *Repository) getTypingSessions(ctx context.Context, userID uint, limit int) ([]UnifiedSession, error) {
	sessions := make([]UnifiedSession, 0)
	return sessions, nil
}

func (r *Repository) getMathSessions(ctx context.Context, userID uint, limit int) ([]UnifiedSession, error) {
	sessions := make([]UnifiedSession, 0)
	return sessions, nil
}

func (r *Repository) getReadingSessions(ctx context.Context, userID uint, limit int) ([]UnifiedSession, error) {
	sessions := make([]UnifiedSession, 0)
	return sessions, nil
}

func (r *Repository) getPianoSessions(ctx context.Context, userID uint, limit int) ([]UnifiedSession, error) {
	sessions := make([]UnifiedSession, 0)
	return sessions, nil
}

// Placeholder methods for leaderboards
func (r *Repository) getTopTypingWPM(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	entries := make([]LeaderboardEntry, 0)
	return entries, nil
}

func (r *Repository) getTopMathAccuracy(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	entries := make([]LeaderboardEntry, 0)
	return entries, nil
}

func (r *Repository) getTopReadingComprehension(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	entries := make([]LeaderboardEntry, 0)
	return entries, nil
}

func (r *Repository) getTopPianoScore(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	entries := make([]LeaderboardEntry, 0)
	return entries, nil
}

func (r *Repository) getOverallLeaderboard(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	entries := make([]LeaderboardEntry, 0)
	return entries, nil
}
