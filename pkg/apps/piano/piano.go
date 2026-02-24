package piano

import (
	"database/sql"
	"time"
)

// PianoApp manages the piano learning application
type PianoApp struct {
	db *sql.DB
}

// NewPianoApp creates a new piano app instance
func NewPianoApp(db *sql.DB) *PianoApp {
	return &PianoApp{db: db}
}

// ============================================================================
// DATA MODELS
// ============================================================================

// Hand represents left or right hand
type Hand string

const (
	LeftHand  Hand = "left"
	RightHand Hand = "right"
	BothHands Hand = "both"
)

// PracticeSession represents a single piano practice session
type PracticeSession struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Level           int       `json:"level"`
	Hand            Hand      `json:"hand"`
	Score           int       `json:"score"`
	Accuracy        float64   `json:"accuracy"`
	TotalNotes      int       `json:"total_notes"`
	CorrectNotes    int       `json:"correct_notes"`
	AvgResponseTime float64   `json:"avg_response_time"`
	Duration        int       `json:"duration"` // seconds
	Timestamp       time.Time `json:"timestamp"`
}

// NotePerformance tracks performance on individual notes
type NotePerformance struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Note            string    `json:"note"`
	Hand            Hand      `json:"hand"`
	CorrectCount    int       `json:"correct_count"`
	IncorrectCount  int       `json:"incorrect_count"`
	AvgResponseTime float64   `json:"avg_response_time"`
	Accuracy        float64   `json:"accuracy"`
	LastPracticed   time.Time `json:"last_practiced"`
}

// UserLevel tracks user's piano level and progress
type UserLevel struct {
	UserID           int64     `json:"user_id"`
	CurrentLevel     int       `json:"current_level"`
	PracticeSessions int       `json:"practice_sessions"`
	TotalScore       int       `json:"total_score"`
	LastPracticed    time.Time `json:"last_practiced"`
}

// StreakData tracks practice streaks
type StreakData struct {
	CurrentStreak    int       `json:"current_streak"`
	LongestStreak    int       `json:"longest_streak"`
	LastPracticeDate string    `json:"last_practice_date"`
	ConsecutiveDays  int       `json:"consecutive_days"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Goal represents a user's practice goal
type Goal struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	GoalType        string    `json:"goal_type"` // daily_practice, weekly_sessions, accuracy, etc.
	TargetValue     int       `json:"target_value"`
	CurrentProgress int       `json:"current_progress"`
	DueDate         string    `json:"due_date"`
	Completed       bool      `json:"completed"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Achievement/Badge represents earned badge
type Achievement struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	BadgeType  string    `json:"badge_type"` // note_master, scale_virtuoso, consistency_champion, etc.
	Level      int       `json:"level"`
	UnlockedAt time.Time `json:"unlocked_at"`
}

// ============================================================================
// PRACTICE SESSION MANAGEMENT
// ============================================================================

// SavePracticeSession saves a completed practice session
func (app *PianoApp) SavePracticeSession(userID int64, session *PracticeSession) (int64, error) {
	now := time.Now()
	session.UserID = userID
	session.Timestamp = now

	// Validate accuracy
	if session.TotalNotes > 0 {
		session.Accuracy = float64(session.CorrectNotes*100) / float64(session.TotalNotes)
	}

	result, err := app.db.Exec(`
		INSERT INTO practice_sessions
		(user_id, level, hand, score, accuracy, total_notes, correct_notes, avg_response_time, duration, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		userID, session.Level, session.Hand, session.Score, session.Accuracy,
		session.TotalNotes, session.CorrectNotes, session.AvgResponseTime, session.Duration, now)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	session.ID = id
	return id, err
}

// GetUserStats retrieves comprehensive user statistics
func (app *PianoApp) GetUserStats(userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total sessions
	var sessionCount int
	err := app.db.QueryRow(`
		SELECT COUNT(*) FROM practice_sessions WHERE user_id = ?
	`, userID).Scan(&sessionCount)
	if err != nil {
		sessionCount = 0
	}
	stats["total_sessions"] = sessionCount

	// Average accuracy
	var avgAccuracy float64
	err = app.db.QueryRow(`
		SELECT COALESCE(AVG(accuracy), 0) FROM practice_sessions WHERE user_id = ?
	`, userID).Scan(&avgAccuracy)
	if err != nil {
		avgAccuracy = 0
	}
	stats["average_accuracy"] = avgAccuracy

	// Highest score
	var highScore int
	err = app.db.QueryRow(`
		SELECT COALESCE(MAX(score), 0) FROM practice_sessions WHERE user_id = ?
	`, userID).Scan(&highScore)
	if err != nil {
		highScore = 0
	}
	stats["highest_score"] = highScore

	// Total practice time
	var totalTime int
	err = app.db.QueryRow(`
		SELECT COALESCE(SUM(duration), 0) FROM practice_sessions WHERE user_id = ?
	`, userID).Scan(&totalTime)
	if err != nil {
		totalTime = 0
	}
	stats["total_time_minutes"] = totalTime / 60

	// Current level
	var currentLevel int
	err = app.db.QueryRow(`
		SELECT COALESCE(current_level, 1) FROM user_levels WHERE user_id = ?
	`, userID).Scan(&currentLevel)
	if err != nil || currentLevel == 0 {
		currentLevel = 1
	}
	stats["current_level"] = currentLevel

	return stats, nil
}

// ============================================================================
// STREAK TRACKING
// ============================================================================

// GetStreak retrieves user's practice streak
func (app *PianoApp) GetStreak(userID int64) (*StreakData, error) {
	streak := &StreakData{}

	rows, err := app.db.Query(`
		SELECT current_streak, longest_streak, last_practice_date, consecutive_days
		FROM streaks WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var lastDateStr sql.NullString
		err := rows.Scan(&streak.CurrentStreak, &streak.LongestStreak, &lastDateStr, &streak.ConsecutiveDays)
		if err != nil {
			return nil, err
		}
		if lastDateStr.Valid {
			streak.LastPracticeDate = lastDateStr.String
		}
		streak.UpdatedAt = time.Now()
	} else {
		// First time practicing
		streak.CurrentStreak = 1
		streak.LongestStreak = 1
		streak.ConsecutiveDays = 1
		streak.LastPracticeDate = time.Now().Format("2006-01-02")
		streak.UpdatedAt = time.Now()

		// Create initial record
		app.db.Exec(`
			INSERT INTO streaks (user_id, current_streak, longest_streak, last_practice_date, consecutive_days)
			VALUES (?, 1, 1, ?, 1)
		`, userID, streak.LastPracticeDate)
	}

	return streak, nil
}

// UpdateStreak updates the user's practice streak
func (app *PianoApp) UpdateStreak(userID int64) error {
	today := time.Now().Format("2006-01-02")

	// Get current streak
	streak, err := app.GetStreak(userID)
	if err != nil {
		return err
	}

	// Check if already practiced today
	lastDate := streak.LastPracticeDate
	if lastDate == today {
		return nil // Already counted today
	}

	// Parse dates
	lastTime, _ := time.Parse("2006-01-02", lastDate)
	todayTime, _ := time.Parse("2006-01-02", today)
	daysDiff := int(todayTime.Sub(lastTime).Hours() / 24)

	newStreak := streak.CurrentStreak
	if daysDiff == 1 {
		// Consecutive day
		newStreak++
	} else if daysDiff > 1 {
		// Streak broken, restart
		newStreak = 1
	}

	// Update longest streak
	longestStreak := streak.LongestStreak
	if newStreak > longestStreak {
		longestStreak = newStreak
	}

	_, err = app.db.Exec(`
		UPDATE streaks
		SET current_streak = ?, longest_streak = ?, last_practice_date = ?, consecutive_days = ?
		WHERE user_id = ?
	`, newStreak, longestStreak, today, newStreak, userID)

	return err
}

// ============================================================================
// GOALS SYSTEM
// ============================================================================

// CreateGoal creates a new practice goal
func (app *PianoApp) CreateGoal(userID int64, goal *Goal) (int64, error) {
	now := time.Now()
	goal.UserID = userID
	goal.CreatedAt = now
	goal.UpdatedAt = now

	result, err := app.db.Exec(`
		INSERT INTO goals (user_id, goal_type, target_value, current_progress, due_date, completed)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		userID, goal.GoalType, goal.TargetValue, goal.CurrentProgress, goal.DueDate, goal.Completed)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetGoals retrieves user's goals
func (app *PianoApp) GetGoals(userID int64) ([]Goal, error) {
	var goals []Goal

	rows, err := app.db.Query(`
		SELECT id, user_id, goal_type, target_value, current_progress, due_date, completed, created_at, updated_at
		FROM goals WHERE user_id = ? AND completed = 0
		ORDER BY due_date ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var goal Goal
		var createdAt, updatedAt string
		err := rows.Scan(&goal.ID, &goal.UserID, &goal.GoalType, &goal.TargetValue,
			&goal.CurrentProgress, &goal.DueDate, &goal.Completed, &createdAt, &updatedAt)
		if err != nil {
			continue
		}
		goal.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		goal.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		goals = append(goals, goal)
	}

	return goals, nil
}

// UpdateGoalProgress updates a goal's progress
func (app *PianoApp) UpdateGoalProgress(goalID int64, progress int) error {
	_, err := app.db.Exec(`
		UPDATE goals
		SET current_progress = ?
		WHERE id = ?
	`, progress, goalID)
	return err
}

// ============================================================================
// ACHIEVEMENTS/BADGES
// ============================================================================

// BadgeDefinitions map badge types to their requirements
var BadgeDefinitions = map[string]map[string]int{
	"note_master": {
		"description": 100,
		"requirement": 50, // 50 notes mastered
	},
	"scale_virtuoso": {
		"description": 101,
		"requirement": 100, // 100 consecutive correct notes
	},
	"consistency_champion": {
		"description": 102,
		"requirement": 7, // 7 day streak
	},
	"accuracy_expert": {
		"description": 103,
		"requirement": 95, // 95% accuracy
	},
	"speed_demon": {
		"description": 104,
		"requirement": 500, // 500 milliseconds avg response
	},
	"level_5_reached": {
		"description": 105,
		"requirement": 5, // Level 5
	},
}

// AwardAchievement awards a badge to the user
func (app *PianoApp) AwardAchievement(userID int64, badgeType string, level int) (bool, error) {
	// Check if already awarded
	var exists int
	err := app.db.QueryRow(`
		SELECT COUNT(*) FROM achievements WHERE user_id = ? AND badge_type = ?
	`, userID, badgeType).Scan(&exists)

	if err != nil || exists > 0 {
		return false, err
	}

	// Award the badge
	_, err = app.db.Exec(`
		INSERT INTO achievements (user_id, badge_type, level, unlocked_at)
		VALUES (?, ?, ?, ?)
	`, userID, badgeType, level, time.Now())

	return err == nil, err
}

// GetAchievements retrieves user's badges
func (app *PianoApp) GetAchievements(userID int64) ([]Achievement, error) {
	var achievements []Achievement

	rows, err := app.db.Query(`
		SELECT id, user_id, badge_type, level, unlocked_at
		FROM achievements WHERE user_id = ?
		ORDER BY unlocked_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ach Achievement
		var unlockedAtStr string
		err := rows.Scan(&ach.ID, &ach.UserID, &ach.BadgeType, &ach.Level, &unlockedAtStr)
		if err != nil {
			continue
		}
		ach.UnlockedAt, _ = time.Parse("2006-01-02 15:04:05", unlockedAtStr)
		achievements = append(achievements, ach)
	}

	return achievements, nil
}

// ============================================================================
// NOTE PERFORMANCE TRACKING
// ============================================================================

// RecordNoteAttempt records a note practice attempt
func (app *PianoApp) RecordNoteAttempt(userID int64, note string, hand Hand, correct bool) error {
	// Check if record exists
	var exists int
	err := app.db.QueryRow(`
		SELECT COUNT(*) FROM note_performance WHERE user_id = ? AND note = ? AND hand = ?
	`, userID, note, hand).Scan(&exists)

	if err != nil {
		return err
	}

	now := time.Now()

	if exists == 0 {
		// Create new record
		_, err = app.db.Exec(`
			INSERT INTO note_performance (user_id, note, hand, correct_count, incorrect_count, last_practiced)
			VALUES (?, ?, ?, ?, ?, ?)
		`,
			userID, note, hand,
			boolToInt(correct), boolToInt(!correct), now)
	} else {
		// Update existing record
		correctDelta := boolToInt(correct)
		incorrectDelta := boolToInt(!correct)

		_, err = app.db.Exec(`
			UPDATE note_performance
			SET correct_count = correct_count + ?,
				incorrect_count = incorrect_count + ?,
				last_practiced = ?
			WHERE user_id = ? AND note = ? AND hand = ?
		`,
			correctDelta, incorrectDelta, now,
			userID, note, hand)
	}

	return err
}

// GetNoteAnalytics retrieves performance data for all notes
func (app *PianoApp) GetNoteAnalytics(userID int64) ([]NotePerformance, error) {
	var analytics []NotePerformance

	rows, err := app.db.Query(`
		SELECT id, user_id, note, hand, correct_count, incorrect_count, avg_response_time, last_practiced
		FROM note_performance WHERE user_id = ?
		ORDER BY (correct_count + incorrect_count) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var np NotePerformance
		var lastPracticedStr string
		err := rows.Scan(&np.ID, &np.UserID, &np.Note, &np.Hand, &np.CorrectCount,
			&np.IncorrectCount, &np.AvgResponseTime, &lastPracticedStr)
		if err != nil {
			continue
		}

		// Calculate accuracy
		total := np.CorrectCount + np.IncorrectCount
		if total > 0 {
			np.Accuracy = float64(np.CorrectCount*100) / float64(total)
		}

		np.LastPracticed, _ = time.Parse("2006-01-02 15:04:05", lastPracticedStr)
		analytics = append(analytics, np)
	}

	return analytics, nil
}

// ============================================================================
// DATABASE INITIALIZATION
// ============================================================================

// InitDB initializes the piano app database tables
func (app *PianoApp) InitDB() error {
	_, err := app.db.Exec(`
		CREATE TABLE IF NOT EXISTS practice_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			hand TEXT NOT NULL,
			score INTEGER NOT NULL,
			accuracy REAL NOT NULL,
			total_notes INTEGER NOT NULL,
			correct_notes INTEGER NOT NULL,
			avg_response_time REAL,
			duration INTEGER DEFAULT 0,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS note_performance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			note TEXT NOT NULL,
			hand TEXT NOT NULL,
			correct_count INTEGER DEFAULT 0,
			incorrect_count INTEGER DEFAULT 0,
			avg_response_time REAL,
			last_practiced TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, note, hand)
		);

		CREATE TABLE IF NOT EXISTS user_levels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			current_level INTEGER DEFAULT 1,
			practice_sessions INTEGER DEFAULT 0,
			total_score INTEGER DEFAULT 0,
			last_practiced TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS streaks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			current_streak INTEGER DEFAULT 0,
			longest_streak INTEGER DEFAULT 0,
			last_practice_date TEXT,
			consecutive_days INTEGER DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS goals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			goal_type TEXT NOT NULL,
			target_value INTEGER NOT NULL,
			current_progress INTEGER DEFAULT 0,
			due_date TEXT,
			completed BOOLEAN DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS achievements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			badge_type TEXT NOT NULL,
			level INTEGER DEFAULT 1,
			unlocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, badge_type)
		);

		CREATE INDEX IF NOT EXISTS idx_practice_sessions_user ON practice_sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_note_performance_user ON note_performance(user_id);
		CREATE INDEX IF NOT EXISTS idx_goals_user ON goals(user_id);
		CREATE INDEX IF NOT EXISTS idx_achievements_user ON achievements(user_id);
	`)
	return err
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
