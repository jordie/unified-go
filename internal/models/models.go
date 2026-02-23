package models

import "time"

// ============================================================================
// SHARED MODELS - Used across all education apps
// ============================================================================

// BaseResult represents the common result structure for all apps
type BaseResult struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	AppName   string    `json:"app_name"`
	CreatedAt time.Time `json:"created_at"`
}

// UserStats represents aggregated stats for a user across one app
type UserStats struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	AppName      string    `json:"app_name"`
	TotalSessions int      `json:"total_sessions"`
	AverageScore float64   `json:"average_score"`
	BestScore    float64   `json:"best_score"`
	TotalTime    int       `json:"total_time"` // in seconds
	LastUpdated  time.Time `json:"last_updated"`
}

// ============================================================================
// TYPING APP MODELS
// ============================================================================

// TypingResult represents a single typing test result
type TypingResult struct {
	BaseResult
	WPM                 int     `json:"wpm"`
	RawWPM              int     `json:"raw_wpm"`
	Accuracy            float64 `json:"accuracy"`
	TestType            string  `json:"test_type"` // "words", "time", "race"
	TestMode            string  `json:"test_mode"`
	TestDuration        int     `json:"test_duration"` // in seconds
	TotalCharacters     int     `json:"total_characters"`
	CorrectCharacters   int     `json:"correct_characters"`
	IncorrectCharacters int     `json:"incorrect_characters"`
	Errors              int     `json:"errors"`
	TimeTaken           float64 `json:"time_taken"`
	TextSnippet         string  `json:"text_snippet"`
}

// TypingStats represents typing statistics for a user
type TypingStats struct {
	UserStats
	BestWPM        int     `json:"best_wpm"`
	AverageWPM     float64 `json:"average_wpm"`
	AverageAccuracy float64 `json:"average_accuracy"`
	TotalWordsTyped int     `json:"total_words_typed"`
}

// Race represents a typing race result
type Race struct {
	BaseResult
	Difficulty string  `json:"difficulty"` // "easy", "medium", "hard"
	Placement  int     `json:"placement"`
	WPM        int     `json:"wpm"`
	Accuracy   float64 `json:"accuracy"`
	RaceTime   float64 `json:"race_time"`
	XPEarned   int     `json:"xp_earned"`
}

// RacingStats represents racing statistics for a user
type RacingStats struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TotalRaces int      `json:"total_races"`
	Wins      int       `json:"wins"`
	Podiums   int       `json:"podiums"`
	TotalXP   int       `json:"total_xp"`
	CurrentCar string    `json:"current_car"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// MATH APP MODELS
// ============================================================================

// MathProblem represents a generated math problem
type MathProblem struct {
	ID           string  `json:"id"`
	Operation    string  `json:"operation"` // "+", "-", "*", "/"
	Operand1     int     `json:"operand1"`
	Operand2     int     `json:"operand2"`
	Difficulty   string  `json:"difficulty"` // "easy", "medium", "hard"
	CorrectAnswer float64 `json:"correct_answer"`
}

// MathResult represents a math problem attempt
type MathResult struct {
	BaseResult
	ProblemID     string    `json:"problem_id"`
	Operation     string    `json:"operation"`
	Operand1      int       `json:"operand1"`
	Operand2      int       `json:"operand2"`
	Difficulty    string    `json:"difficulty"`
	UserAnswer    float64   `json:"user_answer"`
	CorrectAnswer float64   `json:"correct_answer"`
	IsCorrect     bool      `json:"is_correct"`
	TimeTaken     float64   `json:"time_taken"` // in seconds
	SessionID     string    `json:"session_id"`
}

// MathStats represents math learning statistics
type MathStats struct {
	UserStats
	Accuracy       float64 `json:"accuracy"`
	CorrectCount   int     `json:"correct_count"`
	IncorrectCount int     `json:"incorrect_count"`
	TimePerProblem float64 `json:"time_per_problem"`
}

// MathWeakness represents an area where a user needs improvement
type MathWeakness struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	Operation   string  `json:"operation"`
	Difficulty  string  `json:"difficulty"`
	ErrorRate   float64 `json:"error_rate"`
	Priority    int     `json:"priority"`
	RecommendedAction string `json:"recommended_action"`
}

// ============================================================================
// READING APP MODELS
// ============================================================================

// ReadingPassage represents a passage for reading practice
type ReadingPassage struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	DifficultyLevel int    `json:"difficulty_level"` // 1-10
	WordCount    int    `json:"word_count"`
	GradeLevel   int    `json:"grade_level"`
	Category     string `json:"category"`
}

// ReadingResult represents a reading session result
type ReadingResult struct {
	BaseResult
	PassageID      int64   `json:"passage_id"`
	WordsRead      int     `json:"words_read"`
	CorrectWords   int     `json:"correct_words"`
	IncorrectWords int     `json:"incorrect_words"`
	Accuracy       float64 `json:"accuracy"`
	TimeSpent      int     `json:"time_spent"` // in seconds
	ReadingSpeed   float64 `json:"reading_speed"` // words per minute
	ComprehensionScore float64 `json:"comprehension_score"`
}

// ReadingStats represents reading statistics for a user
type ReadingStats struct {
	UserStats
	TotalWordsRead     int     `json:"total_words_read"`
	AverageAccuracy   float64 `json:"average_accuracy"`
	AverageSpeed      float64 `json:"average_speed"`
	AverageComprehension float64 `json:"average_comprehension"`
}

// WordMastery represents a user's mastery of a specific word
type WordMastery struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Word         string    `json:"word"`
	MasteryLevel int       `json:"mastery_level"` // 0-5
	AttemptCount int       `json:"attempt_count"`
	CorrectCount int       `json:"correct_count"`
	LastAttempt  time.Time `json:"last_attempt"`
}

// ============================================================================
// PIANO APP MODELS
// ============================================================================

// PianoSession represents a piano practice session
type PianoSession struct {
	BaseResult
	Duration       int    `json:"duration"` // in seconds
	NotesPlayed    int    `json:"notes_played"`
	NotesCorrect   int    `json:"notes_correct"`
	Accuracy       float64 `json:"accuracy"`
	Level          int    `json:"level"`
	Difficulty     string  `json:"difficulty"`
	PieceID        string  `json:"piece_id"`
	ExerciseType   string  `json:"exercise_type"` // "warmup", "exercise", "piece"
}

// PianoStats represents piano statistics for a user
type PianoStats struct {
	UserStats
	CurrentLevel   int     `json:"current_level"`
	Accuracy       float64 `json:"average_accuracy"`
	TotalNotes     int     `json:"total_notes"`
	CorrectNotes   int     `json:"correct_notes"`
	StreakDays     int     `json:"streak_days"`
	TotalPracticeTime int   `json:"total_practice_time"`
}

// PianoBadge represents an achievement badge earned by the user
type PianoBadge struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	BadgeName   string    `json:"badge_name"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	EarnedAt    time.Time `json:"earned_at"`
}

// Goal represents a learning goal for a user
type Goal struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	AppName      string    `json:"app_name"`
	GoalType     string    `json:"goal_type"` // "accuracy", "speed", "completion", "custom"
	TargetValue  int       `json:"target_value"`
	CurrentValue int       `json:"current_value"`
	DueDate      time.Time `json:"due_date"`
	Completed    bool      `json:"completed"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// ============================================================================
// GAMIFICATION MODELS
// ============================================================================

// XPLog represents an XP earning event
type XPLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	AppName   string    `json:"app_name"`
	Amount    int       `json:"amount"`
	Source    string    `json:"source"` // "achievement", "completion", "bonus"
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// Leaderboard represents a user's position in a leaderboard
type LeaderboardEntry struct {
	Rank      int     `json:"rank"`
	UserID    int64   `json:"user_id"`
	Username  string  `json:"username"`
	Score     int     `json:"score"`
	Value     float64 `json:"value"`
	Metric    string  `json:"metric"` // "xp", "accuracy", "speed"
	AppName   string  `json:"app_name"`
}

// ============================================================================
// JOURNAL & PROGRESS MODELS
// ============================================================================

// UserJournal represents a user's learning journal entry
type UserJournal struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	AppName     string    `json:"app_name"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Reflection  string    `json:"reflection"`
	MoodRating  int       `json:"mood_rating"` // 1-5
	CreatedAt   time.Time `json:"created_at"`
}

// ProgressMilestone represents a milestone achievement
type ProgressMilestone struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	AppName    string    `json:"app_name"`
	Title      string    `json:"title"`
	Description string   `json:"description"`
	AchievedAt time.Time `json:"achieved_at"`
}
