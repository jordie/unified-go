package dashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jgirmay/unified-go/pkg/events"
	"github.com/jgirmay/unified-go/pkg/realtime"
)

// LeaderboardStreamingManager handles real-time leaderboard updates
type LeaderboardStreamingManager struct {
	hub                *realtime.Hub
	rankTracker        *RankTracker
	velocityAnalyzer   *RankVelocityAnalyzer
	leaderboardService *LeaderboardService
	mu                 sync.RWMutex
	streamingSessions  map[string]*StreamingSession // [category]session
}

// StreamingSession represents an active leaderboard streaming session
type StreamingSession struct {
	Category          string
	StartTime         time.Time
	UpdateCount       int
	RankChanges       map[uint]*RankChange // [userID]change
	ActiveSubscribers int
	mu                sync.RWMutex
}

// NewLeaderboardStreamingManager creates a new leaderboard streaming manager
func NewLeaderboardStreamingManager(
	hub *realtime.Hub,
	leaderboardService *LeaderboardService,
) *LeaderboardStreamingManager {
	tracker := NewRankTracker()
	analyzer := NewRankVelocityAnalyzer(tracker)

	return &LeaderboardStreamingManager{
		hub:                hub,
		rankTracker:        tracker,
		velocityAnalyzer:   analyzer,
		leaderboardService: leaderboardService,
		streamingSessions:  make(map[string]*StreamingSession),
	}
}

// StartLeaderboardStream starts streaming a specific leaderboard category
func (lsm *LeaderboardStreamingManager) StartLeaderboardStream(category string) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	if _, exists := lsm.streamingSessions[category]; exists {
		return fmt.Errorf("stream already active for category: %s", category)
	}

	session := &StreamingSession{
		Category:    category,
		StartTime:   time.Now(),
		RankChanges: make(map[uint]*RankChange),
	}

	lsm.streamingSessions[category] = session
	return nil
}

// StopLeaderboardStream stops streaming a leaderboard category
func (lsm *LeaderboardStreamingManager) StopLeaderboardStream(category string) {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	delete(lsm.streamingSessions, category)
	lsm.rankTracker.ClearCategory(category)
}

// ProcessScoreUpdate processes a score update and broadcasts rank changes
func (lsm *LeaderboardStreamingManager) ProcessScoreUpdate(
	ctx context.Context,
	userID uint,
	username string,
	app string,
	category string,
	newScore float64,
) error {
	lsm.mu.RLock()
	session, sessionExists := lsm.streamingSessions[category]
	lsm.mu.RUnlock()

	if !sessionExists {
		return fmt.Errorf("no active stream for category: %s", category)
	}

	// Get current leaderboard
	leaderboard, err := lsm.GetLeaderboardByApp(ctx, app, 100)
	if err != nil {
		return fmt.Errorf("failed to get leaderboard: %w", err)
	}

	// Find user's new rank
	newRank := lsm.findUserRank(leaderboard, userID)
	if newRank == 0 {
		return fmt.Errorf("user not found in leaderboard")
	}

	// Record the snapshot
	lsm.rankTracker.RecordSnapshot(userID, category, newRank, newScore)

	// Detect rank change
	rankChange := lsm.rankTracker.DetectRankChange(userID, category)

	// Broadcast the score update
	lsm.broadcastScoreUpdate(category, userID, username, newScore, newRank)

	// If rank changed, broadcast the rank change
	if rankChange != nil {
		lsm.velocityAnalyzer.RecordChange(rankChange)
		lsm.broadcastRankChange(rankChange)

		// Update session stats
		session.mu.Lock()
		session.RankChanges[userID] = rankChange
		session.UpdateCount++
		session.mu.Unlock()

		// Check for special milestones
		lsm.checkAndBroadcastMilestones(rankChange, username)
	}

	return nil
}

// broadcastScoreUpdate broadcasts a score update to leaderboard subscribers
func (lsm *LeaderboardStreamingManager) broadcastScoreUpdate(
	category string,
	userID uint,
	username string,
	score float64,
	rank int,
) {
	channel := fmt.Sprintf("leaderboard:%s", category)

	message := map[string]interface{}{
		"type":       "score_update",
		"user_id":    userID,
		"username":   username,
		"score":      score,
		"rank":       rank,
		"timestamp":  time.Now(),
		"category":   category,
	}

	lsm.hub.Broadcast(channel, message)
}

// broadcastRankChange broadcasts a rank change to leaderboard and user-specific subscribers
func (lsm *LeaderboardStreamingManager) broadcastRankChange(change *RankChange) {
	// Broadcast to leaderboard channel
	leaderboardChannel := fmt.Sprintf("leaderboard:%s", change.Category)
	leaderboardMessage := map[string]interface{}{
		"type":          "rank_change",
		"user_id":       change.UserID,
		"previous_rank": change.PreviousRank,
		"current_rank":  change.CurrentRank,
		"rank_delta":    change.RankDelta,
		"metric_delta":  change.MetricDelta,
		"velocity":      change.Velocity,
		"is_promotion":  change.IsPromotion,
		"timestamp":     change.Timestamp,
		"category":      change.Category,
	}

	lsm.hub.Broadcast(leaderboardChannel, leaderboardMessage)

	// Broadcast to user-specific rank change channel
	rankChangeChannel := fmt.Sprintf("user:%d:rank-changes", change.UserID)
	userRankMessage := map[string]interface{}{
		"type":          "rank_changed",
		"category":      change.Category,
		"previous_rank": change.PreviousRank,
		"current_rank":  change.CurrentRank,
		"metric_delta":  change.MetricDelta,
		"velocity":      change.Velocity,
		"is_promotion":  change.IsPromotion,
		"timestamp":     change.Timestamp,
	}

	lsm.hub.BroadcastToUser(rankChangeChannel, change.UserID, userRankMessage)
}

// checkAndBroadcastMilestones checks for and broadcasts rank milestones
func (lsm *LeaderboardStreamingManager) checkAndBroadcastMilestones(
	change *RankChange,
	username string,
) {
	// Top 10 promotion
	if change.PreviousRank > 10 && change.CurrentRank <= 10 {
		lsm.broadcastMilestone(
			change.UserID,
			username,
			"rank_milestone_top10",
			map[string]interface{}{
				"rank":     change.CurrentRank,
				"category": change.Category,
			},
		)
	}

	// Top 5 promotion
	if change.PreviousRank > 5 && change.CurrentRank <= 5 {
		lsm.broadcastMilestone(
			change.UserID,
			username,
			"rank_milestone_top5",
			map[string]interface{}{
				"rank":     change.CurrentRank,
				"category": change.Category,
			},
		)
	}

	// #1 rank achieved
	if change.CurrentRank == 1 {
		lsm.broadcastMilestone(
			change.UserID,
			username,
			"rank_milestone_first",
			map[string]interface{}{
				"category": change.Category,
			},
		)
	}
}

// broadcastMilestone broadcasts a milestone achievement
func (lsm *LeaderboardStreamingManager) broadcastMilestone(
	userID uint,
	username string,
	milestoneType string,
	data map[string]interface{},
) {
	// Broadcast to user's achievement channel
	achievementChannel := fmt.Sprintf("user:%d:achievements", userID)
	message := map[string]interface{}{
		"type":       "rank_milestone",
		"milestone":  milestoneType,
		"username":   username,
		"timestamp":  time.Now(),
	}

	// Merge additional data
	for k, v := range data {
		message[k] = v
	}

	lsm.hub.BroadcastToUser(achievementChannel, userID, message)

	// Also broadcast to activity feed for global visibility
	activityMessage := map[string]interface{}{
		"type":       "rank_milestone",
		"milestone":  milestoneType,
		"user_id":    userID,
		"username":   username,
		"timestamp":  time.Now(),
	}

	for k, v := range data {
		activityMessage[k] = v
	}

	lsm.hub.Broadcast("activity:achievements", activityMessage)
}

// GetLeaderboardByApp returns leaderboard for an app
func (lsm *LeaderboardStreamingManager) GetLeaderboardByApp(
	ctx context.Context,
	app string,
	limit int,
) (*realtime.LeaderboardMessage, error) {
	// This would integrate with the actual app repositories
	// For now, return a placeholder structure
	return &realtime.LeaderboardMessage{
		Category: app,
		Entries:  make([]realtime.LeaderboardEntry, 0),
	}, nil
}

// findUserRank finds a user's rank in the leaderboard
func (lsm *LeaderboardStreamingManager) findUserRank(
	lb *realtime.LeaderboardMessage,
	userID uint,
) int {
	for _, entry := range lb.Entries {
		if entry.UserID == userID {
			return entry.Rank
		}
	}
	return 0
}

// GetStreamingSession returns information about a streaming session
func (lsm *LeaderboardStreamingManager) GetStreamingSession(category string) *StreamingSession {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	session, exists := lsm.streamingSessions[category]
	if !exists {
		return nil
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	return &StreamingSession{
		Category:          session.Category,
		StartTime:         session.StartTime,
		UpdateCount:       session.UpdateCount,
		RankChanges:       session.RankChanges,
		ActiveSubscribers: session.ActiveSubscribers,
	}
}

// GetActiveStreams returns all active streaming categories
func (lsm *LeaderboardStreamingManager) GetActiveStreams() []string {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	streams := make([]string, 0, len(lsm.streamingSessions))
	for category := range lsm.streamingSessions {
		streams = append(streams, category)
	}

	return streams
}

// UpdateSubscriberCount updates the subscriber count for a category
func (lsm *LeaderboardStreamingManager) UpdateSubscriberCount(category string, count int) {
	lsm.mu.RLock()
	session, exists := lsm.streamingSessions[category]
	lsm.mu.RUnlock()

	if exists {
		session.mu.Lock()
		session.ActiveSubscribers = count
		session.mu.Unlock()
	}
}

// HandleLeaderboardEvent processes leaderboard events from the event bus
func (lsm *LeaderboardStreamingManager) HandleLeaderboardEvent(ctx context.Context, event *events.Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	switch event.Type {
	case events.EventLeaderboardUpdate:
		return lsm.handleLeaderboardUpdate(ctx, event)
	case events.EventScoreUpdated:
		return lsm.handleScoreUpdate(ctx, event)
	case events.EventRankChanged:
		return lsm.handleRankChanged(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// handleLeaderboardUpdate handles a leaderboard update event
func (lsm *LeaderboardStreamingManager) handleLeaderboardUpdate(ctx context.Context, event *events.Event) error {
	if event.Data == nil {
		return fmt.Errorf("event data is nil")
	}

	category, ok := event.Data["category"].(string)
	if !ok {
		return fmt.Errorf("category not found in event data")
	}

	// Broadcast leaderboard update to all subscribers
	channel := fmt.Sprintf("leaderboard:%s", category)
	message := map[string]interface{}{
		"type":      "leaderboard_update",
		"category":  category,
		"timestamp": event.Timestamp,
	}

	lsm.hub.Broadcast(channel, message)
	return nil
}

// handleScoreUpdate handles a score updated event
func (lsm *LeaderboardStreamingManager) handleScoreUpdate(ctx context.Context, event *events.Event) error {
	if event.Data == nil {
		return fmt.Errorf("event data is nil")
	}

	userID := event.UserID
	app := event.App

	category, ok := event.Data["category"].(string)
	if !ok {
		return fmt.Errorf("category not found in event data")
	}

	newScore, ok := event.Data["score"].(float64)
	if !ok {
		return fmt.Errorf("score not found in event data")
	}

	username, ok := event.Data["username"].(string)
	if !ok {
		username = fmt.Sprintf("User%d", userID)
	}

	return lsm.ProcessScoreUpdate(ctx, userID, username, app, category, newScore)
}

// handleRankChanged handles a rank changed event
func (lsm *LeaderboardStreamingManager) handleRankChanged(ctx context.Context, event *events.Event) error {
	if event.Data == nil {
		return fmt.Errorf("event data is nil")
	}

	category, ok := event.Data["category"].(string)
	if !ok {
		return fmt.Errorf("category not found in event data")
	}

	newRank, ok := event.Data["new_rank"].(float64)
	if !ok {
		return fmt.Errorf("new_rank not found in event data")
	}

	// Record the rank snapshot
	lsm.rankTracker.RecordSnapshot(event.UserID, category, int(newRank), 0)

	// Detect and broadcast change
	rankChange := lsm.rankTracker.DetectRankChange(event.UserID, category)
	if rankChange != nil {
		username, ok := event.Data["username"].(string)
		if !ok {
			username = fmt.Sprintf("User%d", event.UserID)
		}

		lsm.velocityAnalyzer.RecordChange(rankChange)
		lsm.broadcastRankChange(rankChange)
		lsm.checkAndBroadcastMilestones(rankChange, username)
	}

	return nil
}

// StreamingStats returns statistics about the streaming manager
type StreamingStats struct {
	ActiveStreams        int
	TotalRankChanges     int
	RankedUsers          int
	CategoriesTracking   int
}

// GetStreamingStats returns statistics about active streaming
func (lsm *LeaderboardStreamingManager) GetStreamingStats() StreamingStats {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	totalRankChanges := 0
	for _, session := range lsm.streamingSessions {
		session.mu.RLock()
		totalRankChanges += len(session.RankChanges)
		session.mu.RUnlock()
	}

	trackerStats := lsm.rankTracker.GetStats()

	return StreamingStats{
		ActiveStreams:      len(lsm.streamingSessions),
		TotalRankChanges:   totalRankChanges,
		RankedUsers:        trackerStats.TotalSnapshots,
		CategoriesTracking: trackerStats.CategoriesTracked,
	}
}
