package dashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jgirmay/unified-go/pkg/realtime"
)

// AchievementType defines types of achievements
type AchievementType string

const (
	// Streak achievements
	AchievementStreak7Days   AchievementType = "streak_7_days"
	AchievementStreak30Days  AchievementType = "streak_30_days"
	AchievementStreak100Days AchievementType = "streak_100_days"

	// Score achievements
	AchievementScore100    AchievementType = "score_100"
	AchievementScore500    AchievementType = "score_500"
	AchievementScore1000   AchievementType = "score_1000"
	AchievementScore5000   AchievementType = "score_5000"
	AchievementScore10000  AchievementType = "score_10000"

	// Rank achievements
	AchievementRankTop10  AchievementType = "rank_top_10"
	AchievementRankTop5   AchievementType = "rank_top_5"
	AchievementRankFirst  AchievementType = "rank_first_place"

	// Skill achievements
	AchievementSkillLevel5   AchievementType = "skill_level_5"
	AchievementSkillLevel10  AchievementType = "skill_level_10"
	AchievementSkillLevel25  AchievementType = "skill_level_25"

	// Consistency achievements
	AchievementPerfectAccuracy AchievementType = "perfect_accuracy"
	AchievementHighAccuracy    AchievementType = "high_accuracy_95"
	AchievementConsistency     AchievementType = "consistency_10_sessions"
)

// Achievement represents an unlocked achievement
type Achievement struct {
	Type        AchievementType
	Title       string
	Description string
	Icon        string
	Points      int
	Timestamp   time.Time
	Category    string // "streak", "score", "rank", "skill", "consistency"
}

// AchievementUnlock represents an achievement unlock event
type AchievementUnlock struct {
	UserID           uint
	Username         string
	Achievement      *Achievement
	App              string
	UnlockedAt       time.Time
	NotificationSent bool
}

// AchievementNotifier detects and broadcasts achievements
type AchievementNotifier struct {
	hub                  *realtime.Hub
	unlockedAchievements map[uint]map[AchievementType]bool // [userID][achievementType]unlocked
	recentUnlocks        map[uint][]*AchievementUnlock     // [userID]recent unlocks
	maxRecentUnlocks     int
	mu                   sync.RWMutex
}

// NewAchievementNotifier creates a new achievement notifier
func NewAchievementNotifier(hub *realtime.Hub) *AchievementNotifier {
	return &AchievementNotifier{
		hub:                  hub,
		unlockedAchievements: make(map[uint]map[AchievementType]bool),
		recentUnlocks:        make(map[uint][]*AchievementUnlock),
		maxRecentUnlocks:     100,
	}
}

// CheckStreakMilestone checks for streak achievements
func (an *AchievementNotifier) CheckStreakMilestone(userID uint, username string, streakDays int) []*AchievementUnlock {
	an.mu.Lock()
	defer an.mu.Unlock()

	if an.unlockedAchievements[userID] == nil {
		an.unlockedAchievements[userID] = make(map[AchievementType]bool)
	}

	var unlocks []*AchievementUnlock

	if streakDays >= 7 && !an.unlockedAchievements[userID][AchievementStreak7Days] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementStreak7Days,
				Title:       "Week Warrior",
				Description: "Maintain a 7-day practice streak",
				Icon:        "🔥",
				Points:      50,
				Category:    "streak",
			},
		))
		an.unlockedAchievements[userID][AchievementStreak7Days] = true
	}

	if streakDays >= 30 && !an.unlockedAchievements[userID][AchievementStreak30Days] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementStreak30Days,
				Title:       "Month Master",
				Description: "Maintain a 30-day practice streak",
				Icon:        "🌟",
				Points:      100,
				Category:    "streak",
			},
		))
		an.unlockedAchievements[userID][AchievementStreak30Days] = true
	}

	if streakDays >= 100 && !an.unlockedAchievements[userID][AchievementStreak100Days] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementStreak100Days,
				Title:       "Century Collector",
				Description: "Maintain a 100-day practice streak",
				Icon:        "💯",
				Points:      500,
				Category:    "streak",
			},
		))
		an.unlockedAchievements[userID][AchievementStreak100Days] = true
	}

	return unlocks
}

// CheckScoreMilestone checks for score achievements
func (an *AchievementNotifier) CheckScoreMilestone(userID uint, username string, score float64, app string) []*AchievementUnlock {
	an.mu.Lock()
	defer an.mu.Unlock()

	if an.unlockedAchievements[userID] == nil {
		an.unlockedAchievements[userID] = make(map[AchievementType]bool)
	}

	var unlocks []*AchievementUnlock

	if score >= 100 && !an.unlockedAchievements[userID][AchievementScore100] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementScore100,
				Title:       "Hundred Point Club",
				Description: "Reach 100 points in " + app,
				Icon:        "🎯",
				Points:      25,
				Category:    "score",
			},
		))
		an.unlockedAchievements[userID][AchievementScore100] = true
	}

	if score >= 500 && !an.unlockedAchievements[userID][AchievementScore500] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementScore500,
				Title:       "Five Hundred Specialist",
				Description: "Reach 500 points in " + app,
				Icon:        "🏆",
				Points:      75,
				Category:    "score",
			},
		))
		an.unlockedAchievements[userID][AchievementScore500] = true
	}

	if score >= 1000 && !an.unlockedAchievements[userID][AchievementScore1000] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementScore1000,
				Title:       "Thousand Triumph",
				Description: "Reach 1000 points in " + app,
				Icon:        "👑",
				Points:      150,
				Category:    "score",
			},
		))
		an.unlockedAchievements[userID][AchievementScore1000] = true
	}

	if score >= 5000 && !an.unlockedAchievements[userID][AchievementScore5000] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementScore5000,
				Title:       "Master Achiever",
				Description: "Reach 5000 points in " + app,
				Icon:        "🥇",
				Points:      300,
				Category:    "score",
			},
		))
		an.unlockedAchievements[userID][AchievementScore5000] = true
	}

	if score >= 10000 && !an.unlockedAchievements[userID][AchievementScore10000] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementScore10000,
				Title:       "Legendary Performer",
				Description: "Reach 10,000 points in " + app,
				Icon:        "⭐",
				Points:      500,
				Category:    "score",
			},
		))
		an.unlockedAchievements[userID][AchievementScore10000] = true
	}

	return unlocks
}

// CheckRankMilestone checks for rank achievements
func (an *AchievementNotifier) CheckRankMilestone(userID uint, username string, rank int, category string) []*AchievementUnlock {
	an.mu.Lock()
	defer an.mu.Unlock()

	if an.unlockedAchievements[userID] == nil {
		an.unlockedAchievements[userID] = make(map[AchievementType]bool)
	}

	var unlocks []*AchievementUnlock

	if rank <= 10 && !an.unlockedAchievements[userID][AchievementRankTop10] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementRankTop10,
				Title:       "Top 10 Contender",
				Description: "Reach top 10 in " + category,
				Icon:        "🏅",
				Points:      100,
				Category:    "rank",
			},
		))
		an.unlockedAchievements[userID][AchievementRankTop10] = true
	}

	if rank <= 5 && !an.unlockedAchievements[userID][AchievementRankTop5] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementRankTop5,
				Title:       "Elite Five",
				Description: "Reach top 5 in " + category,
				Icon:        "🌟",
				Points:      200,
				Category:    "rank",
			},
		))
		an.unlockedAchievements[userID][AchievementRankTop5] = true
	}

	if rank == 1 && !an.unlockedAchievements[userID][AchievementRankFirst] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementRankFirst,
				Title:       "The Champion",
				Description: "Reach #1 in " + category,
				Icon:        "🥇",
				Points:      500,
				Category:    "rank",
			},
		))
		an.unlockedAchievements[userID][AchievementRankFirst] = true
	}

	return unlocks
}

// CheckAccuracyMilestone checks for consistency achievements
func (an *AchievementNotifier) CheckAccuracyMilestone(userID uint, username string, accuracy float64) []*AchievementUnlock {
	an.mu.Lock()
	defer an.mu.Unlock()

	if an.unlockedAchievements[userID] == nil {
		an.unlockedAchievements[userID] = make(map[AchievementType]bool)
	}

	var unlocks []*AchievementUnlock

	if accuracy >= 100 && !an.unlockedAchievements[userID][AchievementPerfectAccuracy] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementPerfectAccuracy,
				Title:       "Perfection Achieved",
				Description: "Achieve 100% accuracy in a session",
				Icon:        "💎",
				Points:      250,
				Category:    "consistency",
			},
		))
		an.unlockedAchievements[userID][AchievementPerfectAccuracy] = true
	}

	if accuracy >= 95 && !an.unlockedAchievements[userID][AchievementHighAccuracy] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementHighAccuracy,
				Title:       "Accuracy Master",
				Description: "Achieve 95%+ accuracy in a session",
				Icon:        "🎯",
				Points:      100,
				Category:    "consistency",
			},
		))
		an.unlockedAchievements[userID][AchievementHighAccuracy] = true
	}

	return unlocks
}

// CheckConsistencyMilestone checks for consistency achievements
func (an *AchievementNotifier) CheckConsistencyMilestone(userID uint, username string, sessionCount int) []*AchievementUnlock {
	an.mu.Lock()
	defer an.mu.Unlock()

	if an.unlockedAchievements[userID] == nil {
		an.unlockedAchievements[userID] = make(map[AchievementType]bool)
	}

	var unlocks []*AchievementUnlock

	if sessionCount >= 10 && !an.unlockedAchievements[userID][AchievementConsistency] {
		unlocks = append(unlocks, an.createAchievementUnlock(
			userID, username,
			&Achievement{
				Type:        AchievementConsistency,
				Title:       "Consistency King",
				Description: "Complete 10 sessions",
				Icon:        "💪",
				Points:      75,
				Category:    "consistency",
			},
		))
		an.unlockedAchievements[userID][AchievementConsistency] = true
	}

	return unlocks
}

// createAchievementUnlock creates an achievement unlock
func (an *AchievementNotifier) createAchievementUnlock(userID uint, username string, achievement *Achievement) *AchievementUnlock {
	unlock := &AchievementUnlock{
		UserID:      userID,
		Username:    username,
		Achievement: achievement,
		UnlockedAt:  time.Now(),
	}

	// Add to recent unlocks
	if an.recentUnlocks[userID] == nil {
		an.recentUnlocks[userID] = make([]*AchievementUnlock, 0)
	}

	an.recentUnlocks[userID] = append(an.recentUnlocks[userID], unlock)

	// Trim if too large
	if len(an.recentUnlocks[userID]) > an.maxRecentUnlocks {
		an.recentUnlocks[userID] = an.recentUnlocks[userID][1:]
	}

	return unlock
}

// BroadcastAchievement broadcasts an achievement unlock
func (an *AchievementNotifier) BroadcastAchievement(ctx context.Context, unlock *AchievementUnlock) {
	if unlock == nil {
		return
	}

	// Broadcast to user's achievements channel
	achievementChannel := fmt.Sprintf("user:%d:achievements", unlock.UserID)
	message := map[string]interface{}{
		"type":        "achievement_unlocked",
		"achievement": string(unlock.Achievement.Type),
		"title":       unlock.Achievement.Title,
		"description": unlock.Achievement.Description,
		"icon":        unlock.Achievement.Icon,
		"points":      unlock.Achievement.Points,
		"category":    unlock.Achievement.Category,
		"timestamp":   unlock.UnlockedAt,
	}

	an.hub.BroadcastToUser(achievementChannel, unlock.UserID, message)

	// Also broadcast to global activity
	activityMessage := map[string]interface{}{
		"type":        "achievement_unlocked",
		"user_id":     unlock.UserID,
		"username":    unlock.Username,
		"achievement": string(unlock.Achievement.Type),
		"title":       unlock.Achievement.Title,
		"icon":        unlock.Achievement.Icon,
		"timestamp":   unlock.UnlockedAt,
	}

	an.hub.Broadcast("activity:achievements", activityMessage)

	unlock.NotificationSent = true
}

// BroadcastMultiple broadcasts multiple achievements
func (an *AchievementNotifier) BroadcastMultiple(ctx context.Context, unlocks []*AchievementUnlock) {
	for _, unlock := range unlocks {
		an.BroadcastAchievement(ctx, unlock)
	}
}

// GetUnlockedAchievements returns all unlocked achievements for a user
func (an *AchievementNotifier) GetUnlockedAchievements(userID uint) []AchievementType {
	an.mu.RLock()
	defer an.mu.RUnlock()

	unlocked := make([]AchievementType, 0)
	for achievementType := range an.unlockedAchievements[userID] {
		unlocked = append(unlocked, achievementType)
	}

	return unlocked
}

// GetRecentUnlocks returns recent achievement unlocks for a user
func (an *AchievementNotifier) GetRecentUnlocks(userID uint, limit int) []*AchievementUnlock {
	an.mu.RLock()
	defer an.mu.RUnlock()

	unlocks, exists := an.recentUnlocks[userID]
	if !exists {
		return make([]*AchievementUnlock, 0)
	}

	if limit <= 0 || limit > len(unlocks) {
		limit = len(unlocks)
	}

	result := make([]*AchievementUnlock, limit)
	copy(result, unlocks[len(unlocks)-limit:])
	return result
}

// IsAchievementUnlocked checks if an achievement has been unlocked
func (an *AchievementNotifier) IsAchievementUnlocked(userID uint, achievementType AchievementType) bool {
	an.mu.RLock()
	defer an.mu.RUnlock()

	return an.unlockedAchievements[userID][achievementType]
}

// GetAchievementStats returns statistics about achievements
type AchievementStats struct {
	TotalUnlocked   int
	StreakCount     int
	ScoreCount      int
	RankCount       int
	ConsistencyCount int
	TotalPoints     int
}

// GetStats returns statistics about user's achievements
func (an *AchievementNotifier) GetStats(userID uint) AchievementStats {
	an.mu.RLock()
	defer an.mu.RUnlock()

	stats := AchievementStats{
		TotalUnlocked: len(an.unlockedAchievements[userID]),
	}

	// Count by category and calculate points
	for achievementType := range an.unlockedAchievements[userID] {
		switch achievementType {
		case AchievementStreak7Days, AchievementStreak30Days, AchievementStreak100Days:
			stats.StreakCount++
		case AchievementScore100, AchievementScore500, AchievementScore1000, AchievementScore5000, AchievementScore10000:
			stats.ScoreCount++
		case AchievementRankTop10, AchievementRankTop5, AchievementRankFirst:
			stats.RankCount++
		case AchievementPerfectAccuracy, AchievementHighAccuracy, AchievementConsistency:
			stats.ConsistencyCount++
		}
	}

	// Calculate total points
	for achievementType := range an.unlockedAchievements[userID] {
		achievement := an.getAchievementDefinition(achievementType)
		if achievement != nil {
			stats.TotalPoints += achievement.Points
		}
	}

	return stats
}

// getAchievementDefinition returns the achievement definition
func (an *AchievementNotifier) getAchievementDefinition(achievementType AchievementType) *Achievement {
	definitions := map[AchievementType]*Achievement{
		AchievementStreak7Days:     {Type: AchievementStreak7Days, Points: 50},
		AchievementStreak30Days:    {Type: AchievementStreak30Days, Points: 100},
		AchievementStreak100Days:   {Type: AchievementStreak100Days, Points: 500},
		AchievementScore100:        {Type: AchievementScore100, Points: 25},
		AchievementScore500:        {Type: AchievementScore500, Points: 75},
		AchievementScore1000:       {Type: AchievementScore1000, Points: 150},
		AchievementScore5000:       {Type: AchievementScore5000, Points: 300},
		AchievementScore10000:      {Type: AchievementScore10000, Points: 500},
		AchievementRankTop10:       {Type: AchievementRankTop10, Points: 100},
		AchievementRankTop5:        {Type: AchievementRankTop5, Points: 200},
		AchievementRankFirst:       {Type: AchievementRankFirst, Points: 500},
		AchievementPerfectAccuracy: {Type: AchievementPerfectAccuracy, Points: 250},
		AchievementHighAccuracy:    {Type: AchievementHighAccuracy, Points: 100},
		AchievementConsistency:     {Type: AchievementConsistency, Points: 75},
	}

	return definitions[achievementType]
}
