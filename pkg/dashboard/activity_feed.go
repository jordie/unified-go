package dashboard

import (
	"fmt"
	"sync"
	"time"

	"github.com/jgirmay/unified-go/pkg/events"
	"github.com/jgirmay/unified-go/pkg/realtime"
)

// ActivityType represents the type of activity event
type ActivityType string

const (
	// Achievement activities
	ActivityAchievementUnlocked ActivityType = "achievement_unlocked"

	// Milestone activities
	ActivityMilestoneUnlocked ActivityType = "milestone_unlocked"

	// Leaderboard activities
	ActivityRankChanged  ActivityType = "rank_changed"
	ActivityRankMilestone ActivityType = "rank_milestone"

	// Session activities
	ActivitySessionStarted ActivityType = "session_started"
	ActivitySessionEnded   ActivityType = "session_ended"
	ActivitySessionBest    ActivityType = "session_best"

	// Social activities
	ActivityUserJoined ActivityType = "user_joined"
	ActivityFollowEvent ActivityType = "follow"
	ActivityCommentEvent ActivityType = "comment"

	// System activities
	ActivityMaintenanceNotice ActivityType = "maintenance_notice"
	ActivitySystemUpdate      ActivityType = "system_update"
)

// Activity represents a single activity event in the feed
type Activity struct {
	ID            uint
	Type          ActivityType
	UserID        uint
	Username      string
	App           string
	Title         string
	Description   string
	Icon          string
	Timestamp     time.Time
	Metadata      map[string]interface{}
	RelatedUserID *uint // For follow, mention events
	Score         float64
	Category      string
}

// ActivityFilter represents filtering options for activity queries
type ActivityFilter struct {
	ActivityType  []ActivityType
	UserID        *uint
	App           *string
	StartTime     *time.Time
	EndTime       *time.Time
	Limit         int
	Offset        int
}

// ActivityStats represents statistics about activities
type ActivityStats struct {
	TotalActivities     int
	TodayActivities     int
	ThisWeekActivities  int
	AchievementCount    int
	MilestoneCount      int
	RankChangeCount     int
	SessionCount        int
	MostActiveUser      string
	MostActiveApp       string
	LastActivityTime    time.Time
}

// ActivityFeed manages the activity feed for users
type ActivityFeed struct {
	hub            *realtime.Hub
	eventBus       *events.Bus
	userActivities map[uint][]*Activity // [userID][]activities
	globalActivity []*Activity           // Global activity timeline
	maxHistory     int
	mu             sync.RWMutex
}

// NewActivityFeed creates a new activity feed
func NewActivityFeed(hub *realtime.Hub, eventBus *events.Bus) *ActivityFeed {
	feed := &ActivityFeed{
		hub:            hub,
		eventBus:       eventBus,
		userActivities: make(map[uint][]*Activity),
		globalActivity: make([]*Activity, 0),
		maxHistory:     1000,
	}

	// Subscribe to relevant events
	feed.subscribeToEvents()

	return feed
}

// subscribeToEvents subscribes to event bus for activity tracking
func (af *ActivityFeed) subscribeToEvents() {
	if af.eventBus == nil {
		return
	}

	// Subscribe to achievement events
	af.eventBus.Subscribe(events.EventAchievementUnlocked, af.handleAchievementEvent)

	// Subscribe to streak milestone events
	af.eventBus.Subscribe(events.EventStreakMilestone, af.handleStreakEvent)

	// Subscribe to user milestone events
	af.eventBus.Subscribe(events.EventUserMilestone, af.handleUserMilestoneEvent)

	// Subscribe to rank change events
	af.eventBus.Subscribe(events.EventRankChanged, af.handleRankChangeEvent)

	// Subscribe to session events
	af.eventBus.Subscribe(events.EventSessionEnded, af.handleSessionEvent)

	// Subscribe to high score events
	af.eventBus.Subscribe(events.EventHighScore, af.handleHighScoreEvent)
}

// handleAchievementEvent handles achievement unlocked events
func (af *ActivityFeed) handleAchievementEvent(e *events.Event) error {
	if e == nil {
		return nil
	}

	activity := &Activity{
		Type:      ActivityAchievementUnlocked,
		UserID:    e.UserID,
		Timestamp: time.Now(),
		Metadata:  e.Data,
	}

	// Extract fields from event data
	if title, ok := e.Data["title"].(string); ok {
		activity.Title = title
	}
	if description, ok := e.Data["description"].(string); ok {
		activity.Description = description
	}
	if icon, ok := e.Data["icon"].(string); ok {
		activity.Icon = icon
	}

	af.recordActivity(activity)
	return nil
}

// handleStreakEvent handles streak milestone events
func (af *ActivityFeed) handleStreakEvent(e *events.Event) error {
	if e == nil {
		return nil
	}

	streakDays := 0
	if days, ok := e.Data["streak_days"].(int); ok {
		streakDays = days
	}

	activity := &Activity{
		Type:      ActivityMilestoneUnlocked,
		UserID:    e.UserID,
		App:       e.App,
		Title:     fmt.Sprintf("%d Day Streak!", streakDays),
		Timestamp: time.Now(),
		Metadata:  e.Data,
	}

	af.recordActivity(activity)
	return nil
}

// handleUserMilestoneEvent handles user milestone events
func (af *ActivityFeed) handleUserMilestoneEvent(e *events.Event) error {
	if e == nil {
		return nil
	}

	activity := &Activity{
		Type:      ActivityMilestoneUnlocked,
		UserID:    e.UserID,
		App:       e.App,
		Timestamp: time.Now(),
		Metadata:  e.Data,
	}

	if title, ok := e.Data["description"].(string); ok {
		activity.Title = title
	}

	af.recordActivity(activity)
	return nil
}

// handleRankChangeEvent handles rank change events
func (af *ActivityFeed) handleRankChangeEvent(e *events.Event) error {
	if e == nil {
		return nil
	}

	newRank := 0
	if rank, ok := e.Data["new_rank"].(int); ok {
		newRank = rank
	}

	category := ""
	if cat, ok := e.Data["category"].(string); ok {
		category = cat
	}

	activity := &Activity{
		Type:      ActivityRankChanged,
		UserID:    e.UserID,
		App:       e.App,
		Title:     fmt.Sprintf("Rank #%d in %s", newRank, category),
		Timestamp: time.Now(),
		Metadata:  e.Data,
	}

	af.recordActivity(activity)
	return nil
}

// handleSessionEvent handles session ended events
func (af *ActivityFeed) handleSessionEvent(e *events.Event) error {
	if e == nil {
		return nil
	}

	activity := &Activity{
		Type:      ActivitySessionEnded,
		UserID:    e.UserID,
		App:       e.App,
		Title:     fmt.Sprintf("Session completed in %s", e.App),
		Timestamp: time.Now(),
		Metadata:  e.Data,
	}

	if finalScore, ok := e.Data["final_score"].(float64); ok {
		activity.Score = finalScore
	}

	af.recordActivity(activity)
	return nil
}

// handleHighScoreEvent handles high score events
func (af *ActivityFeed) handleHighScoreEvent(e *events.Event) error {
	if e == nil {
		return nil
	}

	activity := &Activity{
		Type:      ActivitySessionBest,
		UserID:    e.UserID,
		App:       e.App,
		Title:     "New High Score!",
		Timestamp: time.Now(),
		Metadata:  e.Data,
	}

	if newScore, ok := e.Data["new_score"].(float64); ok {
		activity.Score = newScore
	}

	af.recordActivity(activity)
	return nil
}

// RecordActivity records a new activity
func (af *ActivityFeed) RecordActivity(activity *Activity) {
	if activity == nil {
		return
	}

	activity.ID = af.generateActivityID()
	activity.Timestamp = time.Now()

	af.recordActivity(activity)
}

// recordActivity records activity to feed (internal, mutex already held if needed)
func (af *ActivityFeed) recordActivity(activity *Activity) {
	af.mu.Lock()
	defer af.mu.Unlock()

	// Add to user-specific activities
	if af.userActivities[activity.UserID] == nil {
		af.userActivities[activity.UserID] = make([]*Activity, 0)
	}
	af.userActivities[activity.UserID] = append(af.userActivities[activity.UserID], activity)

	// Trim user history if needed
	if len(af.userActivities[activity.UserID]) > af.maxHistory {
		af.userActivities[activity.UserID] = af.userActivities[activity.UserID][1:]
	}

	// Add to global activity
	af.globalActivity = append(af.globalActivity, activity)

	// Trim global history if needed
	if len(af.globalActivity) > af.maxHistory {
		af.globalActivity = af.globalActivity[1:]
	}

	// Broadcast the activity
	af.broadcastActivity(activity)
}

// broadcastActivity broadcasts an activity to relevant channels (non-blocking)
func (af *ActivityFeed) broadcastActivity(activity *Activity) {
	if af.hub == nil {
		return
	}

	// Broadcast in a goroutine to avoid blocking on channel sends
	go func() {
		message := map[string]interface{}{
			"type":        "activity",
			"activity_id": activity.ID,
			"event_type":  string(activity.Type),
			"user_id":     activity.UserID,
			"username":    activity.Username,
			"app":         activity.App,
			"title":       activity.Title,
			"description": activity.Description,
			"icon":        activity.Icon,
			"timestamp":   activity.Timestamp,
			"metadata":    activity.Metadata,
		}

		// Broadcast to user's activity channel
		userChannel := fmt.Sprintf("user:%d:activity", activity.UserID)
		af.hub.BroadcastToUser(userChannel, activity.UserID, message)

		// Broadcast to global activity feed
		af.hub.Broadcast("activity:feed", message)

		// Broadcast to app-specific feed
		if activity.App != "" {
			appChannel := fmt.Sprintf("activity:app:%s", activity.App)
			af.hub.Broadcast(appChannel, message)
		}

		// Broadcast to activity type channel
		typeChannel := fmt.Sprintf("activity:type:%s", string(activity.Type))
		af.hub.Broadcast(typeChannel, message)
	}()
}

// GetUserActivity returns activity feed for a user
func (af *ActivityFeed) GetUserActivity(userID uint, filter *ActivityFilter) []*Activity {
	af.mu.RLock()
	defer af.mu.RUnlock()

	activities, exists := af.userActivities[userID]
	if !exists {
		return make([]*Activity, 0)
	}

	return af.filterActivities(activities, filter)
}

// GetGlobalActivity returns global activity feed
func (af *ActivityFeed) GetGlobalActivity(filter *ActivityFilter) []*Activity {
	af.mu.RLock()
	defer af.mu.RUnlock()

	return af.filterActivities(af.globalActivity, filter)
}

// filterActivities applies filters to activities
func (af *ActivityFeed) filterActivities(activities []*Activity, filter *ActivityFilter) []*Activity {
	if filter == nil {
		filter = &ActivityFilter{Limit: 50}
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	result := make([]*Activity, 0)

	for _, activity := range activities {
		// Type filter
		if len(filter.ActivityType) > 0 {
			found := false
			for _, t := range filter.ActivityType {
				if activity.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// App filter
		if filter.App != nil && activity.App != *filter.App {
			continue
		}

		// Time filters
		if filter.StartTime != nil && activity.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && activity.Timestamp.After(*filter.EndTime) {
			continue
		}

		result = append(result, activity)
	}

	// Apply pagination in reverse order (most recent first)
	if len(result) > 0 {
		// Reverse to get most recent first
		for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
			result[i], result[j] = result[j], result[i]
		}

		// Apply offset and limit
		start := filter.Offset
		end := start + filter.Limit

		if start >= len(result) {
			return make([]*Activity, 0)
		}

		if end > len(result) {
			end = len(result)
		}

		return result[start:end]
	}

	return result
}

// GetActivityCount returns count of activities matching filter
func (af *ActivityFeed) GetActivityCount(userID *uint, filter *ActivityFilter) int {
	af.mu.RLock()
	defer af.mu.RUnlock()

	var activities []*Activity
	if userID != nil {
		var exists bool
		activities, exists = af.userActivities[*userID]
		if !exists {
			return 0
		}
	} else {
		activities = af.globalActivity
	}

	count := 0
	for _, activity := range activities {
		// Type filter
		if filter != nil && len(filter.ActivityType) > 0 {
			found := false
			for _, t := range filter.ActivityType {
				if activity.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// App filter
		if filter != nil && filter.App != nil && activity.App != *filter.App {
			continue
		}

		count++
	}

	return count
}

// GetStats returns activity statistics
func (af *ActivityFeed) GetStats(userID *uint) ActivityStats {
	af.mu.RLock()
	defer af.mu.RUnlock()

	stats := ActivityStats{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)

	var activities []*Activity
	if userID != nil {
		var exists bool
		activities, exists = af.userActivities[*userID]
		if !exists {
			return stats
		}
	} else {
		activities = af.globalActivity
	}

	stats.TotalActivities = len(activities)

	appCounts := make(map[string]int)
	typeCounts := make(map[ActivityType]int)
	userCounts := make(map[string]int)

	for _, activity := range activities {
		typeCounts[activity.Type]++
		appCounts[activity.App]++
		userCounts[activity.Username]++

		if activity.Timestamp.After(todayStart) {
			stats.TodayActivities++
		}
		if activity.Timestamp.After(weekStart) {
			stats.ThisWeekActivities++
		}

		stats.LastActivityTime = activity.Timestamp
	}

	// Count by type
	stats.AchievementCount = typeCounts[ActivityAchievementUnlocked]
	stats.MilestoneCount = typeCounts[ActivityMilestoneUnlocked]
	stats.RankChangeCount = typeCounts[ActivityRankChanged]
	stats.SessionCount = typeCounts[ActivitySessionEnded]

	// Find most active app
	maxAppCount := 0
	for app, count := range appCounts {
		if count > maxAppCount {
			maxAppCount = count
			stats.MostActiveApp = app
		}
	}

	// Find most active user
	maxUserCount := 0
	for user, count := range userCounts {
		if count > maxUserCount {
			maxUserCount = count
			stats.MostActiveUser = user
		}
	}

	return stats
}

// ClearUserActivity clears activity history for a user
func (af *ActivityFeed) ClearUserActivity(userID uint) {
	af.mu.Lock()
	defer af.mu.Unlock()

	delete(af.userActivities, userID)
}

// ClearGlobalActivity clears all global activity (use with caution)
func (af *ActivityFeed) ClearGlobalActivity() {
	af.mu.Lock()
	defer af.mu.Unlock()

	af.globalActivity = make([]*Activity, 0)
}

// generateActivityID generates a unique activity ID
func (af *ActivityFeed) generateActivityID() uint {
	af.mu.RLock()
	defer af.mu.RUnlock()

	maxID := uint(0)
	for _, activities := range af.userActivities {
		for _, activity := range activities {
			if activity.ID > maxID {
				maxID = activity.ID
			}
		}
	}

	for _, activity := range af.globalActivity {
		if activity.ID > maxID {
			maxID = activity.ID
		}
	}

	return maxID + 1
}
