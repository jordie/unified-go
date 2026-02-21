package dashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jgirmay/unified-go/pkg/events"
	"github.com/jgirmay/unified-go/pkg/realtime"
)

// SessionStreamingManager handles real-time session progress streaming
type SessionStreamingManager struct {
	hub                *realtime.Hub
	progressTracker    *ProgressTracker
	metricsAnalyzer    *MetricsAnalyzer
	mu                 sync.RWMutex
	streamingSessions  map[string]*ActiveSession // [sessionID]session
	subscriberCounts   map[string]int             // [sessionID]count
}

// ActiveSession represents an active streaming session
type ActiveSession struct {
	SessionID         string
	UserID            uint
	App               string
	StartTime         time.Time
	LastUpdateTime    time.Time
	TotalUpdates      int
	CurrentMetric     float64
	CurrentAccuracy   float64
	TimeElapsed       time.Duration
	ActiveSubscribers int
	mu                sync.RWMutex
}

// NewSessionStreamingManager creates a new session streaming manager
func NewSessionStreamingManager(hub *realtime.Hub) *SessionStreamingManager {
	tracker := NewProgressTracker()
	analyzer := NewMetricsAnalyzer(tracker)

	return &SessionStreamingManager{
		hub:               hub,
		progressTracker:   tracker,
		metricsAnalyzer:   analyzer,
		streamingSessions: make(map[string]*ActiveSession),
		subscriberCounts:  make(map[string]int),
	}
}

// StartSessionStream starts streaming a session
func (ssm *SessionStreamingManager) StartSessionStream(
	ctx context.Context,
	sessionID string,
	userID uint,
	app string,
) error {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	if _, exists := ssm.streamingSessions[sessionID]; exists {
		return fmt.Errorf("session stream already active: %s", sessionID)
	}

	// Start tracking the session
	if err := ssm.progressTracker.StartSession(sessionID, userID, app); err != nil {
		return fmt.Errorf("failed to start tracking: %w", err)
	}

	session := &ActiveSession{
		SessionID:      sessionID,
		UserID:         userID,
		App:            app,
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
	}

	ssm.streamingSessions[sessionID] = session
	ssm.subscriberCounts[sessionID] = 0

	// Broadcast session start
	ssm.broadcastSessionStart(sessionID, userID, app)

	return nil
}

// EndSessionStream ends a session stream
func (ssm *SessionStreamingManager) EndSessionStream(sessionID string) error {
	ssm.mu.Lock()
	session, exists := ssm.streamingSessions[sessionID]
	ssm.mu.Unlock()

	if !exists {
		return fmt.Errorf("session stream not found: %s", sessionID)
	}

	// Mark session as complete in tracker
	ssm.progressTracker.EndSession(sessionID)

	// Generate summary
	summary := ssm.metricsAnalyzer.SummarizeSession(sessionID)

	// Broadcast session end
	ssm.broadcastSessionEnd(sessionID, session.UserID, summary)

	// Clean up
	ssm.mu.Lock()
	delete(ssm.streamingSessions, sessionID)
	delete(ssm.subscriberCounts, sessionID)
	ssm.mu.Unlock()

	return nil
}

// UpdateProgress records a progress update for a session
func (ssm *SessionStreamingManager) UpdateProgress(
	ctx context.Context,
	sessionID string,
	metricType string,
	value float64,
	label string,
) error {
	ssm.mu.RLock()
	session, exists := ssm.streamingSessions[sessionID]
	ssm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Record the metric
	if err := ssm.progressTracker.RecordMetric(sessionID, metricType, value, label); err != nil {
		return fmt.Errorf("failed to record metric: %w", err)
	}

	// Update session
	progress := ssm.progressTracker.GetSessionProgress(sessionID)
	if progress != nil {
		session.mu.Lock()
		session.LastUpdateTime = progress.LastUpdateTime
		session.TotalUpdates++
		session.CurrentMetric = progress.CurrentMetric
		session.CurrentAccuracy = progress.CurrentAccuracy
		session.TimeElapsed = progress.Duration
		session.mu.Unlock()
	}

	// Broadcast the update
	ssm.broadcastProgressUpdate(sessionID, metricType, value, label)

	// Check for milestones
	ssm.checkAndBroadcastMilestones(sessionID, metricType, value)

	return nil
}

// broadcastSessionStart broadcasts a session start event
func (ssm *SessionStreamingManager) broadcastSessionStart(sessionID string, userID uint, app string) {
	// Broadcast to user's progress channel
	progressChannel := fmt.Sprintf("user:%d:progress", userID)
	message := map[string]interface{}{
		"type":       "session_started",
		"session_id": sessionID,
		"app":        app,
		"timestamp":  time.Now(),
	}

	ssm.hub.BroadcastToUser(progressChannel, userID, message)

	// Also broadcast to session-specific channel
	sessionChannel := fmt.Sprintf("session:%s:live", sessionID)
	ssm.hub.Broadcast(sessionChannel, message)
}

// broadcastSessionEnd broadcasts a session end event with summary
func (ssm *SessionStreamingManager) broadcastSessionEnd(
	sessionID string,
	userID uint,
	summary *SessionSummary,
) {
	// Broadcast to user's progress channel
	progressChannel := fmt.Sprintf("user:%d:progress", userID)

	message := map[string]interface{}{
		"type":           "session_ended",
		"session_id":     sessionID,
		"duration":       summary.Duration.Seconds(),
		"metrics_count":  summary.MetricsCount,
		"final_score":    summary.FinalScore,
		"accuracy":       summary.AverageAccuracy,
		"session_rating": summary.SessionRating,
		"timestamp":      time.Now(),
	}

	ssm.hub.BroadcastToUser(progressChannel, userID, message)

	// Also broadcast to activity feed
	activityMessage := map[string]interface{}{
		"type":           "session_completed",
		"user_id":        userID,
		"app":            summary.App,
		"duration":       summary.Duration.Seconds(),
		"session_rating": summary.SessionRating,
		"timestamp":      time.Now(),
	}

	ssm.hub.Broadcast("activity:feed", activityMessage)
}

// broadcastProgressUpdate broadcasts a real-time progress update
func (ssm *SessionStreamingManager) broadcastProgressUpdate(
	sessionID string,
	metricType string,
	value float64,
	label string,
) {
	ssm.mu.RLock()
	session, exists := ssm.streamingSessions[sessionID]
	ssm.mu.RUnlock()

	if !exists {
		return
	}

	// Broadcast to session-specific channel
	sessionChannel := fmt.Sprintf("session:%s:live", sessionID)
	message := map[string]interface{}{
		"type":        "progress_update",
		"metric_type": metricType,
		"value":       value,
		"label":       label,
		"timestamp":   time.Now(),
	}

	ssm.hub.Broadcast(sessionChannel, message)

	// Broadcast to user's progress channel
	progressChannel := fmt.Sprintf("user:%d:progress", session.UserID)
	userMessage := map[string]interface{}{
		"type":        "progress_update",
		"session_id":  sessionID,
		"metric_type": metricType,
		"value":       value,
		"label":       label,
		"timestamp":   time.Now(),
	}

	ssm.hub.BroadcastToUser(progressChannel, session.UserID, userMessage)
}

// checkAndBroadcastMilestones checks for achievement milestones
func (ssm *SessionStreamingManager) checkAndBroadcastMilestones(
	sessionID string,
	metricType string,
	value float64,
) {
	ssm.mu.RLock()
	session, exists := ssm.streamingSessions[sessionID]
	ssm.mu.RUnlock()

	if !exists {
		return
	}

	// Check for specific milestones based on metric type and app
	var milestone string
	var data map[string]interface{}

	switch session.App {
	case "typing":
		if metricType == "wpm" {
			if value >= 100 && session.CurrentMetric < 100 {
				milestone = "typing_100wpm"
				data = map[string]interface{}{"wpm": value}
			} else if value >= 150 && session.CurrentMetric < 150 {
				milestone = "typing_150wpm"
				data = map[string]interface{}{"wpm": value}
			} else if value >= 200 && session.CurrentMetric < 200 {
				milestone = "typing_200wpm"
				data = map[string]interface{}{"wpm": value}
			}
		}
	case "math":
		if metricType == "accuracy" {
			if value >= 90 && session.CurrentAccuracy < 90 {
				milestone = "math_90accuracy"
				data = map[string]interface{}{"accuracy": value}
			} else if value >= 95 && session.CurrentAccuracy < 95 {
				milestone = "math_95accuracy"
				data = map[string]interface{}{"accuracy": value}
			}
		}
	case "piano":
		if metricType == "score" {
			if value >= 500 && session.CurrentMetric < 500 {
				milestone = "piano_500score"
				data = map[string]interface{}{"score": value}
			} else if value >= 1000 && session.CurrentMetric < 1000 {
				milestone = "piano_1000score"
				data = map[string]interface{}{"score": value}
			}
		}
	case "reading":
		if metricType == "accuracy" {
			if value >= 90 && session.CurrentAccuracy < 90 {
				milestone = "reading_90accuracy"
				data = map[string]interface{}{"accuracy": value}
			}
		}
	}

	if milestone != "" {
		ssm.broadcastMilestone(sessionID, session.UserID, milestone, data)
	}
}

// broadcastMilestone broadcasts a milestone achievement
func (ssm *SessionStreamingManager) broadcastMilestone(
	sessionID string,
	userID uint,
	milestone string,
	data map[string]interface{},
) {
	// Broadcast to user's achievements channel
	achievementChannel := fmt.Sprintf("user:%d:achievements", userID)
	message := map[string]interface{}{
		"type":       "milestone_achieved",
		"session_id": sessionID,
		"milestone":  milestone,
		"timestamp":  time.Now(),
	}

	for k, v := range data {
		message[k] = v
	}

	ssm.hub.BroadcastToUser(achievementChannel, userID, message)

	// Also broadcast to global activity
	activityMessage := map[string]interface{}{
		"type":       "milestone_achieved",
		"user_id":    userID,
		"milestone":  milestone,
		"timestamp":  time.Now(),
	}

	for k, v := range data {
		activityMessage[k] = v
	}

	ssm.hub.Broadcast("activity:achievements", activityMessage)
}

// GetSessionProgress returns current progress for a session
func (ssm *SessionStreamingManager) GetSessionProgress(sessionID string) *SessionProgress {
	return ssm.progressTracker.GetSessionProgress(sessionID)
}

// GetMetricsAnalysis returns analysis of session metrics
func (ssm *SessionStreamingManager) GetMetricsAnalysis(sessionID string) map[string]*MetricStats {
	return ssm.metricsAnalyzer.AnalyzeAllMetrics(sessionID)
}

// GetSessionSummary returns a summary of a completed session
func (ssm *SessionStreamingManager) GetSessionSummary(sessionID string) *SessionSummary {
	return ssm.metricsAnalyzer.SummarizeSession(sessionID)
}

// UpdateSubscriberCount updates the active subscriber count for a session
func (ssm *SessionStreamingManager) UpdateSubscriberCount(sessionID string, count int) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	if session, exists := ssm.streamingSessions[sessionID]; exists {
		session.mu.Lock()
		session.ActiveSubscribers = count
		session.mu.Unlock()
	}
	ssm.subscriberCounts[sessionID] = count
}

// GetActiveSessions returns all active session streams
func (ssm *SessionStreamingManager) GetActiveSessions() []string {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	sessions := make([]string, 0, len(ssm.streamingSessions))
	for sessionID := range ssm.streamingSessions {
		sessions = append(sessions, sessionID)
	}

	return sessions
}

// HandleSessionEvent processes session events from the event bus
func (ssm *SessionStreamingManager) HandleSessionEvent(ctx context.Context, event *events.Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	switch event.Type {
	case events.EventSessionStarted:
		return ssm.handleSessionStarted(ctx, event)
	case events.EventSessionEnded:
		return ssm.handleSessionEnded(ctx, event)
	case events.EventScoreUpdated:
		return ssm.handleScoreUpdated(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// handleSessionStarted handles a session started event
func (ssm *SessionStreamingManager) handleSessionStarted(ctx context.Context, event *events.Event) error {
	sessionID, ok := event.Data["session_id"].(string)
	if !ok {
		return fmt.Errorf("session_id not found in event data")
	}

	app, ok := event.Data["app"].(string)
	if !ok {
		return fmt.Errorf("app not found in event data")
	}

	return ssm.StartSessionStream(ctx, sessionID, event.UserID, app)
}

// handleSessionEnded handles a session ended event
func (ssm *SessionStreamingManager) handleSessionEnded(ctx context.Context, event *events.Event) error {
	sessionID, ok := event.Data["session_id"].(string)
	if !ok {
		return fmt.Errorf("session_id not found in event data")
	}

	return ssm.EndSessionStream(sessionID)
}

// handleScoreUpdated handles a score updated event
func (ssm *SessionStreamingManager) handleScoreUpdated(ctx context.Context, event *events.Event) error {
	sessionID, ok := event.Data["session_id"].(string)
	if !ok {
		return fmt.Errorf("session_id not found in event data")
	}

	metricType, ok := event.Data["metric_type"].(string)
	if !ok {
		return fmt.Errorf("metric_type not found in event data")
	}

	score, ok := event.Data["score"].(float64)
	if !ok {
		return fmt.Errorf("score not found in event data")
	}

	label, ok := event.Data["label"].(string)
	if !ok {
		label = ""
	}

	return ssm.UpdateProgress(ctx, sessionID, metricType, score, label)
}

// SessionStreamingStats represents statistics about active session streaming
type SessionStreamingStats struct {
	ActiveSessions   int
	TotalUpdates     int
	TotalSubscribers int
	AverageMetrics   int
}

// GetStreamingStats returns statistics about active session streaming
func (ssm *SessionStreamingManager) GetStreamingStats() SessionStreamingStats {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	totalUpdates := 0
	totalSubscribers := 0
	totalMetrics := 0

	for _, session := range ssm.streamingSessions {
		session.mu.RLock()
		totalUpdates += session.TotalUpdates
		totalSubscribers += session.ActiveSubscribers
		session.mu.RUnlock()

		metrics := ssm.progressTracker.GetMetricHistory(session.SessionID, 0)
		totalMetrics += len(metrics)
	}

	avgMetrics := 0
	if len(ssm.streamingSessions) > 0 {
		avgMetrics = totalMetrics / len(ssm.streamingSessions)
	}

	return SessionStreamingStats{
		ActiveSessions:   len(ssm.streamingSessions),
		TotalUpdates:     totalUpdates,
		TotalSubscribers: totalSubscribers,
		AverageMetrics:   avgMetrics,
	}
}
