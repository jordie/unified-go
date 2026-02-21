package realtime

import (
	"encoding/json"
	"time"
)

// MessageType defines the type of message being sent
type MessageType string

const (
	// Connection messages
	MessageTypeSubscribe   MessageType = "subscribe"
	MessageTypeUnsubscribe MessageType = "unsubscribe"
	MessageTypeConnect     MessageType = "connect"
	MessageTypeDisconnect  MessageType = "disconnect"

	// Leaderboard messages
	MessageTypeLeaderboardUpdate MessageType = "leaderboard.update"
	MessageTypeRankChange        MessageType = "rank.change"

	// Progress messages
	MessageTypeProgressUpdate MessageType = "progress.update"
	MessageTypeSessionStart   MessageType = "session.start"
	MessageTypeSessionEnd     MessageType = "session.end"

	// Achievement messages
	MessageTypeAchievementUnlocked MessageType = "achievement.unlocked"
	MessageTypeStreakMilestone     MessageType = "streak.milestone"
	MessageTypeHighScore           MessageType = "high.score"

	// Activity messages
	MessageTypeActivityFeed      MessageType = "activity.feed"
	MessageTypeUserActivity      MessageType = "user.activity"
	MessageTypeLeaderboardChange MessageType = "leaderboard.change"

	// Control messages
	MessageTypePing  MessageType = "ping"
	MessageTypePong  MessageType = "pong"
	MessageTypeError MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Channel   string                 `json:"channel"`
	UserID    uint                   `json:"user_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// SubscribeMessage represents a subscription request
type SubscribeMessage struct {
	Channels []string `json:"channels"`
	UserID   uint     `json:"user_id"`
}

// UnsubscribeMessage represents an unsubscription request
type UnsubscribeMessage struct {
	Channels []string `json:"channels"`
}

// LeaderboardUpdateMessage represents a leaderboard rank update
type LeaderboardUpdateMessage struct {
	Category      string    `json:"category"`
	Rank          int       `json:"rank"`
	UserID        uint      `json:"user_id"`
	Username      string    `json:"username"`
	MetricValue   float64   `json:"metric_value"`
	MetricLabel   string    `json:"metric_label"`
	PreviousRank  int       `json:"previous_rank"`
	RankChange    int       `json:"rank_change"`
	Timestamp     time.Time `json:"timestamp"`
}

// ProgressUpdateMessage represents real-time session progress
type ProgressUpdateMessage struct {
	SessionID       string    `json:"session_id"`
	UserID          uint      `json:"user_id"`
	App             string    `json:"app"`
	CurrentMetric   float64   `json:"current_metric"`
	MetricLabel     string    `json:"metric_label"`
	CurrentAccuracy float64   `json:"current_accuracy"`
	TimeElapsed     int       `json:"time_elapsed"`
	IsImproving     bool      `json:"is_improving"`
	Timestamp       time.Time `json:"timestamp"`
}

// AchievementMessage represents a newly unlocked achievement
type AchievementMessage struct {
	Achievement string    `json:"achievement"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Points      int       `json:"points"`
	Timestamp   time.Time `json:"timestamp"`
}

// RankChangeMessage represents a rank change event
type RankChangeMessage struct {
	Category     string    `json:"category"`
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	NewRank      int       `json:"new_rank"`
	OldRank      int       `json:"old_rank"`
	RankChange   int       `json:"rank_change"` // Positive = improved
	Improvement  bool      `json:"improvement"`
	Timestamp    time.Time `json:"timestamp"`
}

// ActivityFeedMessage represents a global activity event
type ActivityFeedMessage struct {
	EventType   string    `json:"event_type"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	Description string    `json:"description"`
	App         string    `json:"app"`
	MetricValue string    `json:"metric_value"`
	Timestamp   time.Time `json:"timestamp"`
}

// ErrorMessage represents an error response
type ErrorMessage struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Time    time.Time `json:"timestamp"`
}

// LeaderboardEntry represents a single entry in a leaderboard
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	App         string    `json:"app"`
	MetricValue float64   `json:"metric_value"`
	MetricLabel string    `json:"metric_label"`
	Timestamp   time.Time `json:"timestamp"`
}

// LeaderboardMessage represents a complete leaderboard snapshot
type LeaderboardMessage struct {
	Category  string              `json:"category"`
	Entries   []LeaderboardEntry  `json:"entries"`
	Timestamp time.Time           `json:"timestamp"`
}

// ToJSON marshals the message to JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// NewMessage creates a new message
func NewMessage(t MessageType, channel string, data map[string]interface{}) *Message {
	return &Message{
		Type:      t,
		Channel:   channel,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewLeaderboardUpdateMessage creates a leaderboard update message
func NewLeaderboardUpdateMessage(update LeaderboardUpdateMessage) *Message {
	data := map[string]interface{}{
		"category":      update.Category,
		"rank":          update.Rank,
		"user_id":       update.UserID,
		"username":      update.Username,
		"metric_value":  update.MetricValue,
		"metric_label":  update.MetricLabel,
		"previous_rank": update.PreviousRank,
		"rank_change":   update.RankChange,
	}
	return &Message{
		Type:      MessageTypeLeaderboardUpdate,
		Channel:   "leaderboard:" + update.Category,
		UserID:    update.UserID,
		Timestamp: update.Timestamp,
		Data:      data,
	}
}

// NewProgressUpdateMessage creates a progress update message
func NewProgressUpdateMessage(update ProgressUpdateMessage) *Message {
	data := map[string]interface{}{
		"session_id":       update.SessionID,
		"app":              update.App,
		"current_metric":   update.CurrentMetric,
		"metric_label":     update.MetricLabel,
		"current_accuracy": update.CurrentAccuracy,
		"time_elapsed":     update.TimeElapsed,
		"is_improving":     update.IsImproving,
	}
	return &Message{
		Type:      MessageTypeProgressUpdate,
		Channel:   "user:" + string(rune(update.UserID)) + ":progress",
		UserID:    update.UserID,
		Timestamp: update.Timestamp,
		Data:      data,
	}
}

// NewAchievementMessage creates an achievement notification message
func NewAchievementMessage(userID uint, achievement AchievementMessage) *Message {
	data := map[string]interface{}{
		"achievement": achievement.Achievement,
		"title":       achievement.Title,
		"description": achievement.Description,
		"icon":        achievement.Icon,
		"points":      achievement.Points,
	}
	return &Message{
		Type:      MessageTypeAchievementUnlocked,
		Channel:   "user:" + string(rune(userID)) + ":achievements",
		UserID:    userID,
		Timestamp: achievement.Timestamp,
		Data:      data,
	}
}

// NewRankChangeMessage creates a rank change notification message
func NewRankChangeMessage(rankChange RankChangeMessage) *Message {
	data := map[string]interface{}{
		"category":      rankChange.Category,
		"new_rank":      rankChange.NewRank,
		"old_rank":      rankChange.OldRank,
		"rank_change":   rankChange.RankChange,
		"improvement":   rankChange.Improvement,
		"username":      rankChange.Username,
	}
	return &Message{
		Type:      MessageTypeRankChange,
		Channel:   "user:" + string(rune(rankChange.UserID)) + ":rank-changes",
		UserID:    rankChange.UserID,
		Timestamp: rankChange.Timestamp,
		Data:      data,
	}
}

// NewActivityFeedMessage creates an activity feed message
func NewActivityFeedMessage(activity ActivityFeedMessage) *Message {
	data := map[string]interface{}{
		"event_type":    activity.EventType,
		"username":      activity.Username,
		"description":   activity.Description,
		"app":           activity.App,
		"metric_value":  activity.MetricValue,
	}
	return &Message{
		Type:      MessageTypeActivityFeed,
		Channel:   "activity:feed",
		UserID:    activity.UserID,
		Timestamp: activity.Timestamp,
		Data:      data,
	}
}

// NewErrorMessage creates an error message
func NewErrorMessage(code, message, details string) *Message {
	data := map[string]interface{}{
		"code":    code,
		"message": message,
		"details": details,
	}
	return &Message{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		Data:      data,
	}
}
