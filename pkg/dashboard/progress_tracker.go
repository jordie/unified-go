package dashboard

import (
	"sync"
	"time"
)

// ProgressMetric represents a single progress measurement
type ProgressMetric struct {
	Timestamp   time.Time
	MetricType  string  // "wpm", "accuracy", "score", "time_elapsed"
	Value       float64
	Label       string
	IsImproving bool
}

// SessionProgress represents current session progress
type SessionProgress struct {
	SessionID       string
	UserID          uint
	App             string
	StartTime       time.Time
	LastUpdateTime  time.Time
	Duration        time.Duration
	CurrentMetric   float64
	MetricType      string
	CurrentAccuracy float64
	MetricHistory   []*ProgressMetric
	IsActive        bool
	mu              sync.RWMutex
}

// ProgressTracker tracks progress metrics during a session
type ProgressTracker struct {
	mu              sync.RWMutex
	activeSessions  map[string]*SessionProgress // [sessionID]progress
	sessionMetrics  map[string][]*ProgressMetric // [sessionID]metrics
	maxHistorySize  int
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		activeSessions:  make(map[string]*SessionProgress),
		sessionMetrics:  make(map[string][]*ProgressMetric),
		maxHistorySize:  1000, // Keep up to 1000 metric samples per session
	}
}

// StartSession starts tracking a new session
func (pt *ProgressTracker) StartSession(sessionID string, userID uint, app string) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if _, exists := pt.activeSessions[sessionID]; exists {
		return ErrSessionAlreadyExists
	}

	now := time.Now()
	session := &SessionProgress{
		SessionID:      sessionID,
		UserID:         userID,
		App:            app,
		StartTime:      now,
		LastUpdateTime: now,
		MetricHistory:  make([]*ProgressMetric, 0),
		IsActive:       true,
	}

	pt.activeSessions[sessionID] = session
	pt.sessionMetrics[sessionID] = make([]*ProgressMetric, 0)

	return nil
}

// EndSession marks a session as complete
func (pt *ProgressTracker) EndSession(sessionID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if session, exists := pt.activeSessions[sessionID]; exists {
		session.mu.Lock()
		session.IsActive = false
		session.mu.Unlock()
	}
}

// RecordMetric records a progress metric
func (pt *ProgressTracker) RecordMetric(sessionID string, metricType string, value float64, label string) error {
	pt.mu.Lock()
	session, exists := pt.activeSessions[sessionID]
	pt.mu.Unlock()

	if !exists {
		return ErrSessionNotFound
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	now := time.Now()

	// Detect if improving (higher is better for most metrics)
	isImproving := false
	if len(session.MetricHistory) > 0 {
		lastMetric := session.MetricHistory[len(session.MetricHistory)-1]
		if lastMetric.MetricType == metricType {
			isImproving = value > lastMetric.Value
		}
	}

	metric := &ProgressMetric{
		Timestamp:   now,
		MetricType:  metricType,
		Value:       value,
		Label:       label,
		IsImproving: isImproving,
	}

	session.MetricHistory = append(session.MetricHistory, metric)

	// Trim history if too large
	if len(session.MetricHistory) > pt.maxHistorySize {
		session.MetricHistory = session.MetricHistory[1:]
	}

	// Update session metrics
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.sessionMetrics[sessionID] = append(pt.sessionMetrics[sessionID], metric)
	if len(pt.sessionMetrics[sessionID]) > pt.maxHistorySize {
		pt.sessionMetrics[sessionID] = pt.sessionMetrics[sessionID][1:]
	}

	// Update current values
	switch metricType {
	case "wpm", "score":
		session.CurrentMetric = value
		session.MetricType = metricType
	case "accuracy":
		session.CurrentAccuracy = value
	}

	session.LastUpdateTime = now
	session.Duration = now.Sub(session.StartTime)

	return nil
}

// GetSessionProgress returns current progress for a session
func (pt *ProgressTracker) GetSessionProgress(sessionID string) *SessionProgress {
	pt.mu.RLock()
	session, exists := pt.activeSessions[sessionID]
	pt.mu.RUnlock()

	if !exists {
		return nil
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// Return a copy
	return &SessionProgress{
		SessionID:       session.SessionID,
		UserID:          session.UserID,
		App:             session.App,
		StartTime:       session.StartTime,
		LastUpdateTime:  session.LastUpdateTime,
		Duration:        session.Duration,
		CurrentMetric:   session.CurrentMetric,
		MetricType:      session.MetricType,
		CurrentAccuracy: session.CurrentAccuracy,
		IsActive:        session.IsActive,
	}
}

// GetMetricHistory returns the metric history for a session
func (pt *ProgressTracker) GetMetricHistory(sessionID string, limit int) []*ProgressMetric {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	metrics, exists := pt.sessionMetrics[sessionID]
	if !exists {
		return make([]*ProgressMetric, 0)
	}

	if limit <= 0 || limit > len(metrics) {
		limit = len(metrics)
	}

	// Return the last `limit` metrics
	result := make([]*ProgressMetric, limit)
	copy(result, metrics[len(metrics)-limit:])
	return result
}

// GetActiveSessions returns all active sessions
func (pt *ProgressTracker) GetActiveSessions() []string {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	sessions := make([]string, 0, len(pt.activeSessions))
	for sessionID, session := range pt.activeSessions {
		session.mu.RLock()
		isActive := session.IsActive
		session.mu.RUnlock()

		if isActive {
			sessions = append(sessions, sessionID)
		}
	}

	return sessions
}

// MetricsAnalyzer analyzes session metrics for trends and insights
type MetricsAnalyzer struct {
	tracker *ProgressTracker
	mu      sync.RWMutex
}

// NewMetricsAnalyzer creates a new metrics analyzer
func NewMetricsAnalyzer(tracker *ProgressTracker) *MetricsAnalyzer {
	return &MetricsAnalyzer{
		tracker: tracker,
	}
}

// MetricStats represents statistics about a metric
type MetricStats struct {
	Current    float64
	Average    float64
	Min        float64
	Max        float64
	Trend      string // "improving", "declining", "stable"
	Velocity   float64 // Change per minute
	SampleSize int
}

// AnalyzeMetric analyzes a specific metric for a session
func (ma *MetricsAnalyzer) AnalyzeMetric(sessionID string, metricType string) *MetricStats {
	metrics := ma.tracker.GetMetricHistory(sessionID, 0)

	if len(metrics) == 0 {
		return nil
	}

	// Filter metrics by type
	typeMetrics := make([]*ProgressMetric, 0)
	for _, m := range metrics {
		if m.MetricType == metricType {
			typeMetrics = append(typeMetrics, m)
		}
	}

	if len(typeMetrics) == 0 {
		return nil
	}

	// Calculate statistics
	stats := &MetricStats{
		Current:    typeMetrics[len(typeMetrics)-1].Value,
		Min:        typeMetrics[0].Value,
		Max:        typeMetrics[0].Value,
		SampleSize: len(typeMetrics),
	}

	sum := 0.0
	for i, m := range typeMetrics {
		if m.Value < stats.Min {
			stats.Min = m.Value
		}
		if m.Value > stats.Max {
			stats.Max = m.Value
		}
		sum += m.Value

		// Calculate trend
		if i > 0 {
			if m.Value > typeMetrics[i-1].Value {
				stats.Trend = "improving"
			} else if m.Value < typeMetrics[i-1].Value {
				stats.Trend = "declining"
			} else {
				stats.Trend = "stable"
			}
		}
	}

	stats.Average = sum / float64(len(typeMetrics))

	// Calculate velocity (change per minute)
	if len(typeMetrics) > 1 {
		firstMetric := typeMetrics[0]
		lastMetric := typeMetrics[len(typeMetrics)-1]
		timeDiff := lastMetric.Timestamp.Sub(firstMetric.Timestamp).Minutes()
		if timeDiff > 0 {
			stats.Velocity = (lastMetric.Value - firstMetric.Value) / timeDiff
		}
	}

	return stats
}

// AnalyzeAllMetrics analyzes all metric types for a session
func (ma *MetricsAnalyzer) AnalyzeAllMetrics(sessionID string) map[string]*MetricStats {
	metrics := ma.tracker.GetMetricHistory(sessionID, 0)

	if len(metrics) == 0 {
		return make(map[string]*MetricStats)
	}

	// Collect all metric types
	metricTypes := make(map[string]bool)
	for _, m := range metrics {
		metricTypes[m.MetricType] = true
	}

	// Analyze each type
	result := make(map[string]*MetricStats)
	for metricType := range metricTypes {
		if stats := ma.AnalyzeMetric(sessionID, metricType); stats != nil {
			result[metricType] = stats
		}
	}

	return result
}

// SessionComparison represents comparison between sessions
type SessionComparison struct {
	CurrentSessionID  string
	PreviousSessionID string
	MetricType        string
	CurrentValue      float64
	PreviousValue     float64
	Improvement       float64 // Percentage improvement
	IsBetter          bool
}

// CompareWithPrevious compares current session with previous session
func (ma *MetricsAnalyzer) CompareWithPrevious(currentSessionID string, previousSessionID string, metricType string) *SessionComparison {
	currentMetrics := ma.tracker.GetMetricHistory(currentSessionID, 1)
	previousMetrics := ma.tracker.GetMetricHistory(previousSessionID, 1)

	if len(currentMetrics) == 0 || len(previousMetrics) == 0 {
		return nil
	}

	currentValue := currentMetrics[0].Value
	previousValue := previousMetrics[0].Value

	comparison := &SessionComparison{
		CurrentSessionID:  currentSessionID,
		PreviousSessionID: previousSessionID,
		MetricType:        metricType,
		CurrentValue:      currentValue,
		PreviousValue:     previousValue,
	}

	if previousValue != 0 {
		improvement := ((currentValue - previousValue) / previousValue) * 100
		comparison.Improvement = improvement
		comparison.IsBetter = currentValue > previousValue
	}

	return comparison
}

// SessionSummary provides a summary of session performance
type SessionSummary struct {
	SessionID        string
	UserID           uint
	App              string
	Duration         time.Duration
	StartTime        time.Time
	EndTime          time.Time
	MetricsCount     int
	BestMetric       string
	WorstMetric      string
	AverageAccuracy  float64
	FinalScore       float64
	SessionRating    string // "excellent", "good", "average", "needs_improvement"
}

// SummarizeSession creates a summary of a session
func (ma *MetricsAnalyzer) SummarizeSession(sessionID string) *SessionSummary {
	progress := ma.tracker.GetSessionProgress(sessionID)
	if progress == nil {
		return nil
	}

	allMetrics := ma.AnalyzeAllMetrics(sessionID)

	summary := &SessionSummary{
		SessionID:       sessionID,
		UserID:          progress.UserID,
		App:             progress.App,
		Duration:        progress.Duration,
		StartTime:       progress.StartTime,
		EndTime:         progress.LastUpdateTime,
		MetricsCount:    len(allMetrics),
		AverageAccuracy: progress.CurrentAccuracy,
		FinalScore:      progress.CurrentMetric,
	}

	// Find best and worst metrics
	var bestStats, worstStats *MetricStats
	for metricType, stats := range allMetrics {
		if bestStats == nil || stats.Max > bestStats.Max {
			bestStats = stats
			summary.BestMetric = metricType
		}
		if worstStats == nil || stats.Min < worstStats.Min {
			worstStats = stats
			summary.WorstMetric = metricType
		}
	}

	// Rate the session
	if progress.CurrentAccuracy >= 95 && progress.CurrentMetric > 500 {
		summary.SessionRating = "excellent"
	} else if progress.CurrentAccuracy >= 85 && progress.CurrentMetric > 300 {
		summary.SessionRating = "good"
	} else if progress.CurrentAccuracy >= 75 && progress.CurrentMetric > 100 {
		summary.SessionRating = "average"
	} else {
		summary.SessionRating = "needs_improvement"
	}

	return summary
}

// Errors
var (
	ErrSessionNotFound      = &Error{"SESSION_NOT_FOUND", "session not found"}
	ErrSessionAlreadyExists = &Error{"SESSION_ALREADY_EXISTS", "session already exists"}
)

// Error represents a progress tracker error
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}
