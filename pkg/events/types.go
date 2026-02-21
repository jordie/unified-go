package events

import (
	"time"
)

// EventType defines the type of event
type EventType string

// Event types
const (
	// Session events
	EventSessionStarted EventType = "session.started"
	EventSessionEnded   EventType = "session.ended"

	// Score/Progress events
	EventScoreUpdated    EventType = "score.updated"
	EventAccuracyUpdate  EventType = "accuracy.update"
	EventWPMUpdate       EventType = "wpm.update"
	EventMetricUpdate    EventType = "metric.update"

	// Ranking events
	EventRankChanged         EventType = "rank.changed"
	EventLeaderboardUpdate   EventType = "leaderboard.update"
	EventHighScore           EventType = "high.score"

	// Achievement events
	EventAchievementUnlocked EventType = "achievement.unlocked"
	EventStreakMilestone     EventType = "streak.milestone"
	EventLevelUp             EventType = "level.up"

	// User events
	EventUserGoalReached   EventType = "user.goal.reached"
	EventUserMilestone     EventType = "user.milestone"
	EventConsecutiveDays   EventType = "consecutive.days"

	// System events
	EventLeaderboardRefresh EventType = "leaderboard.refresh"
	EventDailyReportReady   EventType = "daily.report.ready"
	EventWeeklyReportReady  EventType = "weekly.report.ready"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	UserID    uint                   `json:"user_id"`
	App       string                 `json:"app"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// SessionStartedData contains data for a session start event
type SessionStartedData struct {
	SessionID string    `json:"session_id"`
	App       string    `json:"app"`
	StartTime time.Time `json:"start_time"`
}

// SessionEndedData contains data for a session end event
type SessionEndedData struct {
	SessionID    string        `json:"session_id"`
	App          string        `json:"app"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	FinalScore   float64       `json:"final_score"`
	FinalRank    int           `json:"final_rank"`
}

// ScoreUpdatedData contains data for a score update event
type ScoreUpdatedData struct {
	SessionID   string    `json:"session_id"`
	App         string    `json:"app"`
	PreviousWPM float64   `json:"previous_wpm"`
	CurrentWPM  float64   `json:"current_wpm"`
	Improvement float64   `json:"improvement"`
	Timestamp   time.Time `json:"timestamp"`
}

// RankChangedData contains data for a rank change event
type RankChangedData struct {
	Category    string `json:"category"`
	PreviousRank int    `json:"previous_rank"`
	NewRank     int    `json:"new_rank"`
	RankChange  int    `json:"rank_change"`
	Improvement bool   `json:"improvement"`
}

// AchievementUnlockedData contains data for an achievement event
type AchievementUnlockedData struct {
	AchievementID string `json:"achievement_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Icon          string `json:"icon"`
	Points        int    `json:"points"`
}

// StreakMilestoneData contains data for a streak milestone event
type StreakMilestoneData struct {
	StreakDays int    `json:"streak_days"`
	MilestoneType string `json:"milestone_type"` // "7_day", "30_day", "100_day"
	Points     int    `json:"points"`
}

// LevelUpData contains data for a level up event
type LevelUpData struct {
	App          string `json:"app"`
	PreviousLevel int    `json:"previous_level"`
	NewLevel     int    `json:"new_level"`
	TotalXP      int    `json:"total_xp"`
}

// HighScoreData contains data for a high score event
type HighScoreData struct {
	Category      string  `json:"category"`
	PreviousScore float64 `json:"previous_score"`
	NewScore      float64 `json:"new_score"`
	Improvement   float64 `json:"improvement"`
}

// LeaderboardUpdateData contains data for leaderboard updates
type LeaderboardUpdateData struct {
	Category       string `json:"category"`
	UpdateType     string `json:"update_type"` // "rank_change", "new_entry", "score_update"
	AffectedCount  int    `json:"affected_count"`
	TopEntriesOnly bool   `json:"top_entries_only"`
}

// GoalReachedData contains data for goal achievement
type GoalReachedData struct {
	GoalID      string    `json:"goal_id"`
	GoalType    string    `json:"goal_type"` // "daily", "weekly", "monthly"
	Description string    `json:"description"`
	RewardPoints int      `json:"reward_points"`
}

// MetricUpdateData contains generic metric updates
type MetricUpdateData struct {
	MetricName  string      `json:"metric_name"`
	PreviousValue float64   `json:"previous_value"`
	CurrentValue  float64   `json:"current_value"`
	ChangePercent float64   `json:"change_percent"`
}

// EventHandler is a function that handles an event
type EventHandler func(*Event) error

// EventFilter allows filtering events by criteria
type EventFilter struct {
	EventTypes []EventType
	UserID     *uint
	App        *string
	TimeRange  *TimeRange
}

// TimeRange defines a time range for filtering
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewEvent creates a new event
func NewEvent(eventType EventType, userID uint, app string, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		UserID:    userID,
		App:       app,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewSessionStartedEvent creates a session started event
func NewSessionStartedEvent(userID uint, sessionID, app string) *Event {
	return &Event{
		ID:     generateEventID(),
		Type:   EventSessionStarted,
		UserID: userID,
		App:    app,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"session_id": sessionID,
			"app":        app,
			"start_time": time.Now(),
		},
	}
}

// NewSessionEndedEvent creates a session ended event
func NewSessionEndedEvent(userID uint, sessionID, app string, duration time.Duration, finalScore float64, finalRank int) *Event {
	return &Event{
		ID:     generateEventID(),
		Type:   EventSessionEnded,
		UserID: userID,
		App:    app,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"session_id":  sessionID,
			"app":         app,
			"end_time":    time.Now(),
			"duration":    duration,
			"final_score": finalScore,
			"final_rank":  finalRank,
		},
	}
}

// NewScoreUpdatedEvent creates a score updated event
func NewScoreUpdatedEvent(userID uint, app, sessionID string, previousScore, currentScore float64) *Event {
	improvement := currentScore - previousScore
	if improvement < 0 {
		improvement = 0
	}

	return &Event{
		ID:     generateEventID(),
		Type:   EventScoreUpdated,
		UserID: userID,
		App:    app,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"session_id":      sessionID,
			"app":             app,
			"previous_score":  previousScore,
			"current_score":   currentScore,
			"improvement":     improvement,
		},
	}
}

// NewRankChangedEvent creates a rank changed event
func NewRankChangedEvent(userID uint, category string, previousRank, newRank int) *Event {
	rankChange := previousRank - newRank
	improvement := rankChange > 0

	return &Event{
		ID:     generateEventID(),
		Type:   EventRankChanged,
		UserID: userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"category":      category,
			"previous_rank": previousRank,
			"new_rank":      newRank,
			"rank_change":   rankChange,
			"improvement":   improvement,
		},
	}
}

// NewAchievementUnlockedEvent creates an achievement unlocked event
func NewAchievementUnlockedEvent(userID uint, achievementID, title, description, icon string, points int) *Event {
	return &Event{
		ID:     generateEventID(),
		Type:   EventAchievementUnlocked,
		UserID: userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"achievement_id": achievementID,
			"title":          title,
			"description":    description,
			"icon":           icon,
			"points":         points,
		},
	}
}

// NewStreakMilestoneEvent creates a streak milestone event
func NewStreakMilestoneEvent(userID uint, streakDays int, milestoneType string, points int) *Event {
	return &Event{
		ID:     generateEventID(),
		Type:   EventStreakMilestone,
		UserID: userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"streak_days":    streakDays,
			"milestone_type": milestoneType,
			"points":         points,
		},
	}
}

// NewHighScoreEvent creates a high score event
func NewHighScoreEvent(userID uint, category string, previousScore, newScore float64) *Event {
	improvement := newScore - previousScore

	return &Event{
		ID:     generateEventID(),
		Type:   EventHighScore,
		UserID: userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"category":        category,
			"previous_score":  previousScore,
			"new_score":       newScore,
			"improvement":     improvement,
		},
	}
}

// NewLeaderboardUpdateEvent creates a leaderboard update event
func NewLeaderboardUpdateEvent(category, updateType string, affectedCount int, topEntriesOnly bool) *Event {
	return &Event{
		ID:     generateEventID(),
		Type:   EventLeaderboardUpdate,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"category":        category,
			"update_type":     updateType,
			"affected_count":  affectedCount,
			"top_entries_only": topEntriesOnly,
		},
	}
}

// Matches checks if the event matches the given filter
func (e *Event) Matches(filter *EventFilter) bool {
	if filter == nil {
		return true
	}

	// Check event type
	if len(filter.EventTypes) > 0 {
		typeMatch := false
		for _, et := range filter.EventTypes {
			if e.Type == et {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}

	// Check user ID
	if filter.UserID != nil && e.UserID != *filter.UserID {
		return false
	}

	// Check app
	if filter.App != nil && e.App != *filter.App {
		return false
	}

	// Check time range
	if filter.TimeRange != nil {
		if e.Timestamp.Before(filter.TimeRange.Start) || e.Timestamp.After(filter.TimeRange.End) {
			return false
		}
	}

	return true
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[(int(time.Now().UnixNano()) + i) % len(charset)]
	}
	return string(b)
}
