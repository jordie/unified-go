package dashboard

import (
	"sync"
	"time"
)

// RankSnapshot represents a user's rank at a specific point in time
type RankSnapshot struct {
	UserID       uint
	Category     string
	Rank         int
	MetricValue  float64
	Timestamp    time.Time
}

// RankChange represents a detected rank change for a user
type RankChange struct {
	UserID        uint
	Category      string
	PreviousRank  int
	CurrentRank   int
	RankDelta     int         // Positive = rank improved (number went down), Negative = rank worsened
	MetricDelta   float64     // Change in metric value
	Velocity      float64     // Ranks per hour (positive = climbing)
	TimeDuration  time.Duration
	Timestamp     time.Time
	IsPromotion   bool // True if rank improved
}

// RankTracker monitors rank changes across all leaderboards
type RankTracker struct {
	mu        sync.RWMutex
	snapshots map[string]map[uint]*RankSnapshot // [category][userID]snapshot
	history   map[string]map[uint][]*RankSnapshot // [category][userID]history (limited size)
	maxHistory int
}

// NewRankTracker creates a new rank tracker
func NewRankTracker() *RankTracker {
	return &RankTracker{
		snapshots:  make(map[string]map[uint]*RankSnapshot),
		history:    make(map[string]map[uint][]*RankSnapshot),
		maxHistory: 100, // Keep last 100 snapshots per user per category
	}
}

// RecordSnapshot records a rank snapshot for a user
func (rt *RankTracker) RecordSnapshot(userID uint, category string, rank int, metricValue float64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	now := time.Now()
	snapshot := &RankSnapshot{
		UserID:      userID,
		Category:    category,
		Rank:        rank,
		MetricValue: metricValue,
		Timestamp:   now,
	}

	// Initialize category map if needed
	if rt.snapshots[category] == nil {
		rt.snapshots[category] = make(map[uint]*RankSnapshot)
	}
	if rt.history[category] == nil {
		rt.history[category] = make(map[uint][]*RankSnapshot)
	}

	// Store current snapshot
	rt.snapshots[category][userID] = snapshot

	// Add to history
	rt.history[category][userID] = append(rt.history[category][userID], snapshot)

	// Trim history if too large
	if len(rt.history[category][userID]) > rt.maxHistory {
		rt.history[category][userID] = rt.history[category][userID][1:]
	}
}

// DetectRankChange detects and returns a rank change for a user
func (rt *RankTracker) DetectRankChange(userID uint, category string) *RankChange {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	// Get snapshots for this category
	categorySnapshots, exists := rt.snapshots[category]
	if !exists {
		return nil
	}

	currentSnapshot, exists := categorySnapshots[userID]
	if !exists {
		return nil
	}

	// Get history for this user
	userHistory, exists := rt.history[category][userID]
	if !exists || len(userHistory) < 2 {
		// Need at least 2 snapshots to detect change
		return nil
	}

	// Get previous snapshot (at least 1 before current)
	previousSnapshot := userHistory[len(userHistory)-2]

	// Only report if rank actually changed
	if previousSnapshot.Rank == currentSnapshot.Rank {
		return nil
	}

	timeDuration := currentSnapshot.Timestamp.Sub(previousSnapshot.Timestamp)
	if timeDuration == 0 {
		timeDuration = 1 * time.Millisecond // Avoid division by zero
	}

	rankDelta := previousSnapshot.Rank - currentSnapshot.Rank // Positive = improved
	metricDelta := currentSnapshot.MetricValue - previousSnapshot.MetricValue
	velocity := float64(rankDelta) / timeDuration.Hours() // Ranks per hour

	return &RankChange{
		UserID:       userID,
		Category:     category,
		PreviousRank: previousSnapshot.Rank,
		CurrentRank:  currentSnapshot.Rank,
		RankDelta:    rankDelta,
		MetricDelta:  metricDelta,
		Velocity:     velocity,
		TimeDuration: timeDuration,
		Timestamp:    currentSnapshot.Timestamp,
		IsPromotion:  rankDelta > 0, // Positive delta = promotion
	}
}

// GetCurrentRank returns the current rank for a user in a category
func (rt *RankTracker) GetCurrentRank(userID uint, category string) (int, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	categorySnapshots, exists := rt.snapshots[category]
	if !exists {
		return 0, false
	}

	snapshot, exists := categorySnapshots[userID]
	if !exists {
		return 0, false
	}

	return snapshot.Rank, true
}

// GetRankHistory returns the rank history for a user in a category
func (rt *RankTracker) GetRankHistory(userID uint, category string, limit int) []*RankSnapshot {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	userHistory, exists := rt.history[category][userID]
	if !exists {
		return []*RankSnapshot{}
	}

	if limit <= 0 || limit > len(userHistory) {
		limit = len(userHistory)
	}

	// Return the last `limit` snapshots
	result := make([]*RankSnapshot, limit)
	copy(result, userHistory[len(userHistory)-limit:])
	return result
}

// GetCategoryRanks returns all current ranks for a category
func (rt *RankTracker) GetCategoryRanks(category string) map[uint]int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	categorySnapshots, exists := rt.snapshots[category]
	if !exists {
		return make(map[uint]int)
	}

	ranks := make(map[uint]int)
	for userID, snapshot := range categorySnapshots {
		ranks[userID] = snapshot.Rank
	}

	return ranks
}

// GetAllCategories returns all categories being tracked
func (rt *RankTracker) GetAllCategories() []string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	categories := make([]string, 0, len(rt.snapshots))
	for category := range rt.snapshots {
		categories = append(categories, category)
	}

	return categories
}

// ClearCategory clears all data for a category
func (rt *RankTracker) ClearCategory(category string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	delete(rt.snapshots, category)
	delete(rt.history, category)
}

// ClearAll clears all tracked data
func (rt *RankTracker) ClearAll() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.snapshots = make(map[string]map[uint]*RankSnapshot)
	rt.history = make(map[string]map[uint][]*RankSnapshot)
}

// Statistics returns statistics about the tracker
type TrackerStats struct {
	CategoriesTracked int
	TotalSnapshots    int
	AverageHistoryLen float64
}

// GetStats returns statistics about the tracker
func (rt *RankTracker) GetStats() TrackerStats {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	totalSnapshots := 0
	totalHistoryLen := 0
	historyCount := 0

	for _, categoryHistory := range rt.history {
		for _, userHistory := range categoryHistory {
			totalHistoryLen += len(userHistory)
			historyCount++
		}
	}

	for _, categorySnapshots := range rt.snapshots {
		totalSnapshots += len(categorySnapshots)
	}

	avgHistoryLen := 0.0
	if historyCount > 0 {
		avgHistoryLen = float64(totalHistoryLen) / float64(historyCount)
	}

	return TrackerStats{
		CategoriesTracked: len(rt.snapshots),
		TotalSnapshots:    totalSnapshots,
		AverageHistoryLen: avgHistoryLen,
	}
}

// RankVelocityAnalyzer analyzes rank velocity over time
type RankVelocityAnalyzer struct {
	tracker *RankTracker
	mu      sync.RWMutex
	changes map[string]map[uint][]*RankChange // [category][userID]changes
}

// NewRankVelocityAnalyzer creates a new velocity analyzer
func NewRankVelocityAnalyzer(tracker *RankTracker) *RankVelocityAnalyzer {
	return &RankVelocityAnalyzer{
		tracker: tracker,
		changes: make(map[string]map[uint][]*RankChange),
	}
}

// RecordChange records a rank change for velocity analysis
func (rva *RankVelocityAnalyzer) RecordChange(change *RankChange) {
	rva.mu.Lock()
	defer rva.mu.Unlock()

	if rva.changes[change.Category] == nil {
		rva.changes[change.Category] = make(map[uint][]*RankChange)
	}

	rva.changes[change.Category][change.UserID] = append(
		rva.changes[change.Category][change.UserID],
		change,
	)
}

// GetAverageVelocity returns average rank velocity for a user
func (rva *RankVelocityAnalyzer) GetAverageVelocity(userID uint, category string, limit int) float64 {
	rva.mu.RLock()
	defer rva.mu.RUnlock()

	changes, exists := rva.changes[category][userID]
	if !exists || len(changes) == 0 {
		return 0
	}

	if limit <= 0 || limit > len(changes) {
		limit = len(changes)
	}

	totalVelocity := 0.0
	for i := len(changes) - limit; i < len(changes); i++ {
		totalVelocity += changes[i].Velocity
	}

	return totalVelocity / float64(limit)
}

// IsMomentum checks if a user has positive momentum (climbing ranks)
func (rva *RankVelocityAnalyzer) IsMomentum(userID uint, category string, recentChanges int) bool {
	if recentChanges <= 0 {
		recentChanges = 5
	}

	velocity := rva.GetAverageVelocity(userID, category, recentChanges)
	return velocity > 0.5 // At least 0.5 ranks per hour
}

// GetStreakInfo returns information about a user's rank streak
type StreakInfo struct {
	Direction     string    // "up", "down", or "steady"
	ChangeCount   int
	TotalDelta    int
	StartTime     time.Time
	LastChangeTime time.Time
}

// GetRankStreak analyzes rank movement patterns
func (rva *RankVelocityAnalyzer) GetRankStreak(userID uint, category string) *StreakInfo {
	rva.mu.RLock()
	defer rva.mu.RUnlock()

	changes, exists := rva.changes[category][userID]
	if !exists || len(changes) == 0 {
		return nil
	}

	if len(changes) == 0 {
		return nil
	}

	totalDelta := 0
	direction := "steady"

	for _, change := range changes {
		totalDelta += change.RankDelta
	}

	if totalDelta > 0 {
		direction = "up"
	} else if totalDelta < 0 {
		direction = "down"
	}

	return &StreakInfo{
		Direction:     direction,
		ChangeCount:   len(changes),
		TotalDelta:    totalDelta,
		StartTime:     changes[0].Timestamp,
		LastChangeTime: changes[len(changes)-1].Timestamp,
	}
}
