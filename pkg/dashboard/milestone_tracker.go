package dashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jgirmay/unified-go/pkg/realtime"
)

// MilestoneType represents types of milestones
type MilestoneType string

const (
	// Practice milestones
	MilestoneFirstSession   MilestoneType = "first_session"
	Milestone10Sessions     MilestoneType = "10_sessions"
	Milestone50Sessions     MilestoneType = "50_sessions"
	Milestone100Sessions    MilestoneType = "100_sessions"
	Milestone500Sessions    MilestoneType = "500_sessions"

	// Time milestones
	MilestoneFirstHour     MilestoneType = "first_hour"
	MilestoneOneDay        MilestoneType = "one_day"
	MilestoneTotalHours10  MilestoneType = "total_hours_10"
	MilestoneTotalHours50  MilestoneType = "total_hours_50"
	MilestoneTotalHours100 MilestoneType = "total_hours_100"

	// Progress milestones
	MilestoneFirstImprovement MilestoneType = "first_improvement"
	MilestoneDoubleScore      MilestoneType = "double_score"
	Milestone10Streak         MilestoneType = "10_day_streak"
)

// Milestone represents a milestone event
type Milestone struct {
	Type           MilestoneType
	Title          string
	Description    string
	Icon           string
	Reward         int // Points or badge value
	Category       string
	Timestamp      time.Time
	UnlockedAt     time.Time
	UserID         uint
	Username       string
	Metadata       map[string]interface{}
}

// MilestoneTracker tracks user milestones and patterns
type MilestoneTracker struct {
	hub              *realtime.Hub
	userMilestones   map[uint]map[MilestoneType]bool // [userID][milestoneType]unlocked
	milestoneHistory map[uint][]*Milestone           // [userID]history
	maxHistory       int
	mu               sync.RWMutex
}

// NewMilestoneTracker creates a new milestone tracker
func NewMilestoneTracker(hub *realtime.Hub) *MilestoneTracker {
	return &MilestoneTracker{
		hub:              hub,
		userMilestones:   make(map[uint]map[MilestoneType]bool),
		milestoneHistory: make(map[uint][]*Milestone),
		maxHistory:       500,
	}
}

// CheckSessionMilestone checks for session count milestones
func (mt *MilestoneTracker) CheckSessionMilestone(
	userID uint,
	username string,
	sessionCount int,
) []*Milestone {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.userMilestones[userID] == nil {
		mt.userMilestones[userID] = make(map[MilestoneType]bool)
	}

	var milestones []*Milestone

	if sessionCount == 1 && !mt.userMilestones[userID][MilestoneFirstSession] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        MilestoneFirstSession,
				Title:       "First Steps",
				Description: "Complete your first session",
				Icon:        "🎬",
				Reward:      10,
				Category:    "practice",
			},
		))
		mt.userMilestones[userID][MilestoneFirstSession] = true
	}

	if sessionCount == 10 && !mt.userMilestones[userID][Milestone10Sessions] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        Milestone10Sessions,
				Title:       "Getting Started",
				Description: "Complete 10 sessions",
				Icon:        "🚀",
				Reward:      50,
				Category:    "practice",
			},
		))
		mt.userMilestones[userID][Milestone10Sessions] = true
	}

	if sessionCount == 50 && !mt.userMilestones[userID][Milestone50Sessions] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        Milestone50Sessions,
				Title:       "Halfway There",
				Description: "Complete 50 sessions",
				Icon:        "⚡",
				Reward:      100,
				Category:    "practice",
			},
		))
		mt.userMilestones[userID][Milestone50Sessions] = true
	}

	if sessionCount == 100 && !mt.userMilestones[userID][Milestone100Sessions] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        Milestone100Sessions,
				Title:       "Century Reached",
				Description: "Complete 100 sessions",
				Icon:        "💯",
				Reward:      250,
				Category:    "practice",
			},
		))
		mt.userMilestones[userID][Milestone100Sessions] = true
	}

	if sessionCount == 500 && !mt.userMilestones[userID][Milestone500Sessions] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        Milestone500Sessions,
				Title:       "The Grinder",
				Description: "Complete 500 sessions",
				Icon:        "🏆",
				Reward:      500,
				Category:    "practice",
			},
		))
		mt.userMilestones[userID][Milestone500Sessions] = true
	}

	return milestones
}

// CheckTimeMilestone checks for time-based milestones
func (mt *MilestoneTracker) CheckTimeMilestone(
	userID uint,
	username string,
	totalMinutes int,
	isContinuous bool,
) []*Milestone {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.userMilestones[userID] == nil {
		mt.userMilestones[userID] = make(map[MilestoneType]bool)
	}

	var milestones []*Milestone
	totalHours := totalMinutes / 60

	// Time-based milestones
	if totalHours >= 10 && !mt.userMilestones[userID][MilestoneTotalHours10] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        MilestoneTotalHours10,
				Title:       "10 Hour Commitment",
				Description: "Practice for 10 hours total",
				Icon:        "⏱️",
				Reward:      75,
				Category:    "time",
			},
		))
		mt.userMilestones[userID][MilestoneTotalHours10] = true
	}

	if totalHours >= 50 && !mt.userMilestones[userID][MilestoneTotalHours50] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        MilestoneTotalHours50,
				Title:       "50 Hour Master",
				Description: "Practice for 50 hours total",
				Icon:        "🎓",
				Reward:      150,
				Category:    "time",
			},
		))
		mt.userMilestones[userID][MilestoneTotalHours50] = true
	}

	if totalHours >= 100 && !mt.userMilestones[userID][MilestoneTotalHours100] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        MilestoneTotalHours100,
				Title:       "Centennial Practitioner",
				Description: "Practice for 100 hours total",
				Icon:        "👑",
				Reward:      300,
				Category:    "time",
			},
		))
		mt.userMilestones[userID][MilestoneTotalHours100] = true
	}

	// Continuous session milestone
	if totalMinutes >= 60 && isContinuous && !mt.userMilestones[userID][MilestoneFirstHour] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        MilestoneFirstHour,
				Title:       "Hour of Power",
				Description: "Complete a 1-hour session",
				Icon:        "⚙️",
				Reward:      50,
				Category:    "time",
			},
		))
		mt.userMilestones[userID][MilestoneFirstHour] = true
	}

	return milestones
}

// CheckProgressMilestone checks for progress-based milestones
func (mt *MilestoneTracker) CheckProgressMilestone(
	userID uint,
	username string,
	previousBest float64,
	currentScore float64,
) []*Milestone {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.userMilestones[userID] == nil {
		mt.userMilestones[userID] = make(map[MilestoneType]bool)
	}

	var milestones []*Milestone

	// First improvement
	if previousBest > 0 && currentScore > previousBest && !mt.userMilestones[userID][MilestoneFirstImprovement] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        MilestoneFirstImprovement,
				Title:       "Personal Best",
				Description: "Score higher than ever before",
				Icon:        "📈",
				Reward:      50,
				Category:    "progress",
			},
		))
		mt.userMilestones[userID][MilestoneFirstImprovement] = true
	}

	// Double score
	if previousBest > 0 && currentScore >= previousBest*2 && !mt.userMilestones[userID][MilestoneDoubleScore] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        MilestoneDoubleScore,
				Title:       "Double Trouble",
				Description: "Score double your previous best",
				Icon:        "2️⃣",
				Reward:      150,
				Category:    "progress",
			},
		))
		mt.userMilestones[userID][MilestoneDoubleScore] = true
	}

	return milestones
}

// CheckStreakMilestone checks for streak-based milestones
func (mt *MilestoneTracker) CheckStreakMilestone(
	userID uint,
	username string,
	streakDays int,
) []*Milestone {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.userMilestones[userID] == nil {
		mt.userMilestones[userID] = make(map[MilestoneType]bool)
	}

	var milestones []*Milestone

	if streakDays >= 10 && !mt.userMilestones[userID][Milestone10Streak] {
		milestones = append(milestones, mt.createMilestone(
			userID, username,
			&Milestone{
				Type:        Milestone10Streak,
				Title:       "10 Day Fire 🔥",
				Description: "Maintain a 10-day streak",
				Icon:        "🔥",
				Reward:      100,
				Category:    "streak",
			},
		))
		mt.userMilestones[userID][Milestone10Streak] = true
	}

	return milestones
}

// createMilestone creates a milestone event
func (mt *MilestoneTracker) createMilestone(userID uint, username string, milestone *Milestone) *Milestone {
	milestone.UserID = userID
	milestone.Username = username
	milestone.UnlockedAt = time.Now()
	milestone.Timestamp = time.Now()

	// Add to history
	if mt.milestoneHistory[userID] == nil {
		mt.milestoneHistory[userID] = make([]*Milestone, 0)
	}

	mt.milestoneHistory[userID] = append(mt.milestoneHistory[userID], milestone)

	// Trim if too large
	if len(mt.milestoneHistory[userID]) > mt.maxHistory {
		mt.milestoneHistory[userID] = mt.milestoneHistory[userID][1:]
	}

	return milestone
}

// BroadcastMilestone broadcasts a milestone event
func (mt *MilestoneTracker) BroadcastMilestone(ctx context.Context, milestone *Milestone) {
	if milestone == nil {
		return
	}

	// Broadcast to user's achievements channel
	achievementChannel := fmt.Sprintf("user:%d:achievements", milestone.UserID)
	message := map[string]interface{}{
		"type":        "milestone_unlocked",
		"milestone":   string(milestone.Type),
		"title":       milestone.Title,
		"description": milestone.Description,
		"icon":        milestone.Icon,
		"reward":      milestone.Reward,
		"category":    milestone.Category,
		"timestamp":   milestone.UnlockedAt,
	}

	mt.hub.BroadcastToUser(achievementChannel, milestone.UserID, message)

	// Also broadcast to global activity
	activityMessage := map[string]interface{}{
		"type":      "milestone_unlocked",
		"user_id":   milestone.UserID,
		"username":  milestone.Username,
		"milestone": string(milestone.Type),
		"title":     milestone.Title,
		"icon":      milestone.Icon,
		"timestamp": milestone.UnlockedAt,
	}

	mt.hub.Broadcast("activity:achievements", activityMessage)
}

// BroadcastMultiple broadcasts multiple milestones
func (mt *MilestoneTracker) BroadcastMultiple(ctx context.Context, milestones []*Milestone) {
	for _, milestone := range milestones {
		mt.BroadcastMilestone(ctx, milestone)
	}
}

// GetUnlockedMilestones returns all unlocked milestones for a user
func (mt *MilestoneTracker) GetUnlockedMilestones(userID uint) []MilestoneType {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	unlocked := make([]MilestoneType, 0)
	for milestoneType := range mt.userMilestones[userID] {
		unlocked = append(unlocked, milestoneType)
	}

	return unlocked
}

// GetMilestoneHistory returns milestone history for a user
func (mt *MilestoneTracker) GetMilestoneHistory(userID uint, limit int) []*Milestone {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	history, exists := mt.milestoneHistory[userID]
	if !exists {
		return make([]*Milestone, 0)
	}

	if limit <= 0 || limit > len(history) {
		limit = len(history)
	}

	result := make([]*Milestone, limit)
	copy(result, history[len(history)-limit:])
	return result
}

// IsMilestoneUnlocked checks if a milestone has been unlocked
func (mt *MilestoneTracker) IsMilestoneUnlocked(userID uint, milestoneType MilestoneType) bool {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	return mt.userMilestones[userID][milestoneType]
}

// GetMilestoneStats returns statistics about milestones
type MilestoneStats struct {
	TotalUnlocked      int
	PracticeCount      int
	TimeCount          int
	ProgressCount      int
	StreakCount        int
	TotalRewardPoints  int
}

// GetStats returns statistics about user's milestones
func (mt *MilestoneTracker) GetStats(userID uint) MilestoneStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	stats := MilestoneStats{
		TotalUnlocked: len(mt.userMilestones[userID]),
	}

	// Count by category and calculate rewards
	for milestoneType := range mt.userMilestones[userID] {
		switch milestoneType {
		case MilestoneFirstSession, Milestone10Sessions, Milestone50Sessions, Milestone100Sessions, Milestone500Sessions:
			stats.PracticeCount++
		case MilestoneFirstHour, MilestoneOneDay, MilestoneTotalHours10, MilestoneTotalHours50, MilestoneTotalHours100:
			stats.TimeCount++
		case MilestoneFirstImprovement, MilestoneDoubleScore:
			stats.ProgressCount++
		case Milestone10Streak:
			stats.StreakCount++
		}
	}

	// Calculate total reward points
	rewards := map[MilestoneType]int{
		MilestoneFirstSession:     10,
		Milestone10Sessions:       50,
		Milestone50Sessions:       100,
		Milestone100Sessions:      250,
		Milestone500Sessions:      500,
		MilestoneFirstHour:        50,
		MilestoneOneDay:           75,
		MilestoneTotalHours10:     75,
		MilestoneTotalHours50:     150,
		MilestoneTotalHours100:    300,
		MilestoneFirstImprovement: 50,
		MilestoneDoubleScore:      150,
		Milestone10Streak:         100,
	}

	for milestoneType := range mt.userMilestones[userID] {
		if reward, exists := rewards[milestoneType]; exists {
			stats.TotalRewardPoints += reward
		}
	}

	return stats
}
