package math

import (
	"fmt"
	"strings"
	"time"
)

// User represents a math practice user
type User struct {
	ID        uint      `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	LastActive time.Time `json:"last_active" db:"last_active"`
}

// Validate checks if user data is valid
func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if len(u.Username) > 100 {
		return fmt.Errorf("username exceeds 100 characters")
	}
	return nil
}

// MathResult represents a practice session summary
type MathResult struct {
	ID             uint      `json:"id" db:"id"`
	UserID         uint      `json:"user_id" db:"user_id"`
	Mode           string    `json:"mode" db:"mode"` // addition, subtraction, multiplication, division, mixed
	Difficulty     string    `json:"difficulty" db:"difficulty"` // easy, medium, hard, expert
	TotalQuestions int       `json:"total_questions" db:"total_questions"`
	CorrectAnswers int       `json:"correct_answers" db:"correct_answers"`
	TotalTime      float64   `json:"total_time" db:"total_time"`
	AverageTime    float64   `json:"average_time" db:"average_time"`
	Accuracy       float64   `json:"accuracy" db:"accuracy"` // 0-100
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
}

// Validate checks if result data is valid
func (r *MathResult) Validate() error {
	validModes := map[string]bool{"addition": true, "subtraction": true, "multiplication": true, "division": true, "mixed": true}
	if !validModes[r.Mode] {
		return fmt.Errorf("invalid mode: %s", r.Mode)
	}

	validDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true, "expert": true}
	if !validDifficulties[r.Difficulty] {
		return fmt.Errorf("invalid difficulty: %s", r.Difficulty)
	}

	if r.TotalQuestions <= 0 {
		return fmt.Errorf("total_questions must be positive")
	}

	if r.CorrectAnswers < 0 || r.CorrectAnswers > r.TotalQuestions {
		return fmt.Errorf("correct_answers out of range")
	}

	if r.Accuracy < 0 || r.Accuracy > 100 {
		return fmt.Errorf("accuracy must be 0-100")
	}

	if r.TotalTime <= 0 || r.AverageTime <= 0 {
		return fmt.Errorf("time values must be positive")
	}

	return nil
}

// CalculateAccuracy calculates accuracy percentage
func (r *MathResult) CalculateAccuracy() {
	if r.TotalQuestions > 0 {
		r.Accuracy = (float64(r.CorrectAnswers) / float64(r.TotalQuestions)) * 100
	} else {
		r.Accuracy = 0
	}
}

// CalculateAverageTime calculates average time per question
func (r *MathResult) CalculateAverageTime() {
	if r.TotalQuestions > 0 {
		r.AverageTime = r.TotalTime / float64(r.TotalQuestions)
	} else {
		r.AverageTime = 0
	}
}

// QuestionHistory represents an individual question attempt
type QuestionHistory struct {
	ID            uint      `json:"id" db:"id"`
	UserID        uint      `json:"user_id" db:"user_id"`
	Question      string    `json:"question" db:"question"`
	UserAnswer    string    `json:"user_answer" db:"user_answer"`
	CorrectAnswer string    `json:"correct_answer" db:"correct_answer"`
	IsCorrect     bool      `json:"is_correct" db:"is_correct"`
	TimeTaken     float64   `json:"time_taken" db:"time_taken"` // seconds
	FactFamily    string    `json:"fact_family" db:"fact_family"`
	Mode          string    `json:"mode" db:"mode"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
}

// Validate checks if question history is valid
func (qh *QuestionHistory) Validate() error {
	if strings.TrimSpace(qh.Question) == "" {
		return fmt.Errorf("question cannot be empty")
	}

	if strings.TrimSpace(qh.CorrectAnswer) == "" {
		return fmt.Errorf("correct_answer cannot be empty")
	}

	if qh.TimeTaken < 0 {
		return fmt.Errorf("time_taken cannot be negative")
	}

	return nil
}

// Mistake represents a recurring error
type Mistake struct {
	ID            uint      `json:"id" db:"id"`
	UserID        uint      `json:"user_id" db:"user_id"`
	Question      string    `json:"question" db:"question"`
	CorrectAnswer string    `json:"correct_answer" db:"correct_answer"`
	UserAnswer    string    `json:"user_answer" db:"user_answer"`
	Mode          string    `json:"mode" db:"mode"`
	FactFamily    string    `json:"fact_family" db:"fact_family"`
	ErrorCount    int       `json:"error_count" db:"error_count"`
	LastError     time.Time `json:"last_error" db:"last_error"`
}

// Validate checks if mistake data is valid
func (m *Mistake) Validate() error {
	if strings.TrimSpace(m.Question) == "" {
		return fmt.Errorf("question cannot be empty")
	}

	if m.ErrorCount <= 0 {
		return fmt.Errorf("error_count must be positive")
	}

	return nil
}

// IncrementError increments error count and updates timestamp
func (m *Mistake) IncrementError() {
	m.ErrorCount++
	m.LastError = time.Now()
}

// Mastery represents individual fact mastery tracking
type Mastery struct {
	ID                  uint      `json:"id" db:"id"`
	UserID              uint      `json:"user_id" db:"user_id"`
	Fact                string    `json:"fact" db:"fact"` // e.g., "3+5"
	Mode                string    `json:"mode" db:"mode"`
	CorrectStreak       int       `json:"correct_streak" db:"correct_streak"`
	TotalAttempts       int       `json:"total_attempts" db:"total_attempts"`
	MasteryLevel        float64   `json:"mastery_level" db:"mastery_level"` // 0-100
	LastPracticed       time.Time `json:"last_practiced" db:"last_practiced"`
	AverageResponseTime float64   `json:"average_response_time" db:"average_response_time"`
	FastestTime         float64   `json:"fastest_time" db:"fastest_time"`
	SlowestTime         float64   `json:"slowest_time" db:"slowest_time"`
}

// Validate checks if mastery data is valid
func (m *Mastery) Validate() error {
	if strings.TrimSpace(m.Fact) == "" {
		return fmt.Errorf("fact cannot be empty")
	}

	if m.CorrectStreak < 0 {
		return fmt.Errorf("correct_streak cannot be negative")
	}

	if m.TotalAttempts < 0 {
		return fmt.Errorf("total_attempts cannot be negative")
	}

	if m.MasteryLevel < 0 || m.MasteryLevel > 100 {
		return fmt.Errorf("mastery_level must be 0-100")
	}

	if m.AverageResponseTime < 0 || m.FastestTime < 0 || m.SlowestTime < 0 {
		return fmt.Errorf("response times cannot be negative")
	}

	return nil
}

// CalculateMasteryLevel calculates mastery based on performance
// Formula: base_accuracy * 80 + streak * 4 + speed_bonus * 100
func (m *Mastery) CalculateMasteryLevel(baseAccuracy float64, speedBonusApplied bool) {
	masteryLevel := baseAccuracy * 80
	masteryLevel += float64(m.CorrectStreak) * 4

	if speedBonusApplied {
		masteryLevel += 100
	}

	if masteryLevel > 100 {
		masteryLevel = 100
	} else if masteryLevel < 0 {
		masteryLevel = 0
	}

	m.MasteryLevel = masteryLevel
}

// UpdateResponseTime updates timing statistics
func (m *Mastery) UpdateResponseTime(newTime float64) {
	m.TotalAttempts++

	// Update average
	if m.AverageResponseTime == 0 {
		m.AverageResponseTime = newTime
	} else {
		m.AverageResponseTime = (m.AverageResponseTime*float64(m.TotalAttempts-1) + newTime) / float64(m.TotalAttempts)
	}

	// Update fastest
	if m.FastestTime == 0 || newTime < m.FastestTime {
		m.FastestTime = newTime
	}

	// Update slowest
	if newTime > m.SlowestTime {
		m.SlowestTime = newTime
	}

	m.LastPracticed = time.Now()
}

// LearningProfile represents user learning characteristics
type LearningProfile struct {
	ID                   uint      `json:"id" db:"id"`
	UserID               uint      `json:"user_id" db:"user_id"`
	LearningStyle        string    `json:"learning_style" db:"learning_style"` // visual, sequential, global
	PreferredTimeOfDay   string    `json:"preferred_time_of_day" db:"preferred_time_of_day"` // morning, afternoon, evening
	AttentionSpanSeconds int       `json:"attention_span_seconds" db:"attention_span_seconds"`
	BestStreakTime       string    `json:"best_streak_time" db:"best_streak_time"`
	WeakTimeOfDay        string    `json:"weak_time_of_day" db:"weak_time_of_day"`
	AvgSessionLength     int       `json:"avg_session_length" db:"avg_session_length"` // seconds
	TotalPracticeTime    int       `json:"total_practice_time" db:"total_practice_time"` // seconds
	ProfileUpdated       time.Time `json:"profile_updated" db:"profile_updated"`
}

// Validate checks if learning profile is valid
func (lp *LearningProfile) Validate() error {
	validStyles := map[string]bool{"visual": true, "sequential": true, "global": true}
	if !validStyles[lp.LearningStyle] {
		return fmt.Errorf("invalid learning_style: %s", lp.LearningStyle)
	}

	validTimes := map[string]bool{"morning": true, "afternoon": true, "evening": true}
	if lp.PreferredTimeOfDay != "" && !validTimes[lp.PreferredTimeOfDay] {
		return fmt.Errorf("invalid preferred_time_of_day: %s", lp.PreferredTimeOfDay)
	}

	if lp.AttentionSpanSeconds <= 0 {
		return fmt.Errorf("attention_span_seconds must be positive")
	}

	if lp.AvgSessionLength < 0 || lp.TotalPracticeTime < 0 {
		return fmt.Errorf("session times cannot be negative")
	}

	return nil
}

// GetTimeOfDayFromHour determines time of day from hour (0-23)
func GetTimeOfDayFromHour(hour int) string {
	if hour >= 5 && hour < 12 {
		return "morning"
	} else if hour >= 12 && hour < 17 {
		return "afternoon"
	}
	return "evening"
}

// PerformancePattern represents time-of-day performance analysis
type PerformancePattern struct {
	ID              uint      `json:"id" db:"id"`
	UserID          uint      `json:"user_id" db:"user_id"`
	HourOfDay       int       `json:"hour_of_day" db:"hour_of_day"` // 0-23
	DayOfWeek       int       `json:"day_of_week" db:"day_of_week"` // 0-6
	AverageAccuracy float64   `json:"average_accuracy" db:"average_accuracy"`
	AverageSpeed    float64   `json:"average_speed" db:"average_speed"` // questions/minute
	SessionCount    int       `json:"session_count" db:"session_count"`
	LastUpdated     time.Time `json:"last_updated" db:"last_updated"`
}

// Validate checks if performance pattern is valid
func (pp *PerformancePattern) Validate() error {
	if pp.HourOfDay < 0 || pp.HourOfDay > 23 {
		return fmt.Errorf("hour_of_day must be 0-23")
	}

	if pp.DayOfWeek < 0 || pp.DayOfWeek > 6 {
		return fmt.Errorf("day_of_week must be 0-6")
	}

	if pp.AverageAccuracy < 0 || pp.AverageAccuracy > 100 {
		return fmt.Errorf("average_accuracy must be 0-100")
	}

	if pp.AverageSpeed < 0 {
		return fmt.Errorf("average_speed cannot be negative")
	}

	if pp.SessionCount < 0 {
		return fmt.Errorf("session_count cannot be negative")
	}

	return nil
}

// RepetitionSchedule represents SM-2 spaced repetition scheduling
type RepetitionSchedule struct {
	ID           uint      `json:"id" db:"id"`
	UserID       uint      `json:"user_id" db:"user_id"`
	Fact         string    `json:"fact" db:"fact"`
	Mode         string    `json:"mode" db:"mode"`
	NextReview   time.Time `json:"next_review" db:"next_review"`
	IntervalDays int       `json:"interval_days" db:"interval_days"`
	EaseFactor   float64   `json:"ease_factor" db:"ease_factor"` // 1.3-3.5
	ReviewCount  int       `json:"review_count" db:"review_count"`
}

// Validate checks if repetition schedule is valid
func (rs *RepetitionSchedule) Validate() error {
	if strings.TrimSpace(rs.Fact) == "" {
		return fmt.Errorf("fact cannot be empty")
	}

	if rs.EaseFactor < 1.3 || rs.EaseFactor > 3.5 {
		return fmt.Errorf("ease_factor must be 1.3-3.5")
	}

	if rs.IntervalDays < 1 {
		return fmt.Errorf("interval_days must be at least 1")
	}

	if rs.ReviewCount < 0 {
		return fmt.Errorf("review_count cannot be negative")
	}

	return nil
}

// UpdateEaseFactor updates ease factor using SM-2 algorithm
// Formula: ease_factor + (0.1 - (5-quality) * (0.08 + (5-quality)*0.02))
func (rs *RepetitionSchedule) UpdateEaseFactor(quality int) {
	// Quality should be 0-5
	if quality < 0 {
		quality = 0
	} else if quality > 5 {
		quality = 5
	}

	adjustment := 0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02)
	newEase := rs.EaseFactor + adjustment

	// Clamp to valid range
	if newEase < 1.3 {
		newEase = 1.3
	} else if newEase > 3.5 {
		newEase = 3.5
	}

	rs.EaseFactor = newEase
}

// CalculateNextInterval calculates next review interval using SM-2
func (rs *RepetitionSchedule) CalculateNextInterval(quality int) int {
	const (
		INITIAL_INTERVAL = 1
		SECOND_INTERVAL  = 6
	)

	// Quality < 3 means failed review
	if quality < 3 {
		return INITIAL_INTERVAL
	}

	// First two reviews have fixed intervals
	if rs.ReviewCount == 0 {
		return INITIAL_INTERVAL
	}
	if rs.ReviewCount == 1 {
		return SECOND_INTERVAL
	}

	// Subsequent reviews follow exponential growth
	nextInterval := float64(rs.IntervalDays) * rs.EaseFactor
	return int(nextInterval)
}

// IsDueForReview checks if fact is due for review
func (rs *RepetitionSchedule) IsDueForReview() bool {
	return time.Now().After(rs.NextReview) || time.Now().Equal(rs.NextReview)
}

// ScheduleNextReview schedules the next review date
func (rs *RepetitionSchedule) ScheduleNextReview(quality int) {
	rs.UpdateEaseFactor(quality)
	nextInterval := rs.CalculateNextInterval(quality)
	rs.IntervalDays = nextInterval
	rs.NextReview = time.Now().AddDate(0, 0, nextInterval)
	rs.ReviewCount++
}

// === Supporting Types ===

// FactFamily represents a family of related math facts for learning
type FactFamily struct {
	Name     string
	Category string // addition, subtraction, multiplication, division
	Examples []string
	Hint     string
	Strategy string
}

// Quality ratings for SM-2 algorithm
const (
	QUALITY_BLACKOUT             = 0 // Complete failure
	QUALITY_WRONG                = 1 // Completely wrong
	QUALITY_WRONG_REMEMBERED     = 2 // Wrong but recalled
	QUALITY_DIFFICULT            = 3 // Slower than average
	QUALITY_CORRECT              = 4 // Close to average time
	QUALITY_PERFECT              = 5 // Much faster than average
)

// SM-2 Algorithm Constants
const (
	MIN_EASE_FACTOR     = 1.3
	MAX_EASE_FACTOR     = 3.5
	INITIAL_EASE_FACTOR = 2.5
)

// Practice modes
const (
	MODE_ADDITION       = "addition"
	MODE_SUBTRACTION    = "subtraction"
	MODE_MULTIPLICATION = "multiplication"
	MODE_DIVISION       = "division"
	MODE_MIXED          = "mixed"
)

// Difficulty levels
const (
	DIFFICULTY_EASY   = "easy"
	DIFFICULTY_MEDIUM = "medium"
	DIFFICULTY_HARD   = "hard"
	DIFFICULTY_EXPERT = "expert"
)

// Learning styles
const (
	STYLE_VISUAL     = "visual"
	STYLE_SEQUENTIAL = "sequential"
	STYLE_GLOBAL     = "global"
)

// Times of day
const (
	TIME_MORNING   = "morning"
	TIME_AFTERNOON = "afternoon"
	TIME_EVENING   = "evening"
)
