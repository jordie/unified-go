package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jgirmay/unified-go/pkg/unified"
)

// LeaderboardService provides leaderboard functionality
type LeaderboardService struct {
	service *Service
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(service *Service) *LeaderboardService {
	return &LeaderboardService{
		service: service,
	}
}

// GetTypingLeaderboard returns top typing performers (WPM)
func (ls *LeaderboardService) GetTypingLeaderboard(ctx context.Context, limit int) (*unified.UnifiedLeaderboard, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	lb := &unified.UnifiedLeaderboard{
		Category:  "typing_wpm",
		Entries:   make([]unified.LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}

	// Placeholder entries for demonstration
	// In production, this would query the typing repository
	entries := []unified.LeaderboardEntry{
		{
			Rank:        1,
			UserID:      1,
			Username:    "speedster",
			App:         "typing",
			MetricValue: 145.5,
			MetricLabel: "145.5 WPM",
			Timestamp:   time.Now(),
		},
		{
			Rank:        2,
			UserID:      2,
			Username:    "typist",
			App:         "typing",
			MetricValue: 132.0,
			MetricLabel: "132.0 WPM",
			Timestamp:   time.Now(),
		},
		{
			Rank:        3,
			UserID:      3,
			Username:    "keymaster",
			App:         "typing",
			MetricValue: 125.8,
			MetricLabel: "125.8 WPM",
			Timestamp:   time.Now(),
		},
	}

	for i, entry := range entries {
		if i >= limit {
			break
		}
		lb.Entries = append(lb.Entries, entry)
	}

	return lb, nil
}

// GetMathLeaderboard returns top math performers (Accuracy)
func (ls *LeaderboardService) GetMathLeaderboard(ctx context.Context, limit int) (*unified.UnifiedLeaderboard, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	lb := &unified.UnifiedLeaderboard{
		Category:  "math_accuracy",
		Entries:   make([]unified.LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}

	entries := []unified.LeaderboardEntry{
		{
			Rank:        1,
			UserID:      4,
			Username:    "mathgeni us",
			App:         "math",
			MetricValue: 98.5,
			MetricLabel: "98.5% Accuracy",
			Timestamp:   time.Now(),
		},
		{
			Rank:        2,
			UserID:      5,
			Username:    "calculator",
			App:         "math",
			MetricValue: 97.2,
			MetricLabel: "97.2% Accuracy",
			Timestamp:   time.Now(),
		},
		{
			Rank:        3,
			UserID:      6,
			Username:    "numbers",
			App:         "math",
			MetricValue: 96.8,
			MetricLabel: "96.8% Accuracy",
			Timestamp:   time.Now(),
		},
	}

	for i, entry := range entries {
		if i >= limit {
			break
		}
		lb.Entries = append(lb.Entries, entry)
	}

	return lb, nil
}

// GetReadingLeaderboard returns top reading performers (Comprehension)
func (ls *LeaderboardService) GetReadingLeaderboard(ctx context.Context, limit int) (*unified.UnifiedLeaderboard, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	lb := &unified.UnifiedLeaderboard{
		Category:  "reading_comprehension",
		Entries:   make([]unified.LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}

	entries := []unified.LeaderboardEntry{
		{
			Rank:        1,
			UserID:      7,
			Username:    "bookworm",
			App:         "reading",
			MetricValue: 94.3,
			MetricLabel: "94.3% Comprehension",
			Timestamp:   time.Now(),
		},
		{
			Rank:        2,
			UserID:      8,
			Username:    "reader",
			App:         "reading",
			MetricValue: 92.1,
			MetricLabel: "92.1% Comprehension",
			Timestamp:   time.Now(),
		},
		{
			Rank:        3,
			UserID:      9,
			Username:    "scholar",
			App:         "reading",
			MetricValue: 90.7,
			MetricLabel: "90.7% Comprehension",
			Timestamp:   time.Now(),
		},
	}

	for i, entry := range entries {
		if i >= limit {
			break
		}
		lb.Entries = append(lb.Entries, entry)
	}

	return lb, nil
}

// GetPianoLeaderboard returns top piano performers (Score)
func (ls *LeaderboardService) GetPianoLeaderboard(ctx context.Context, limit int) (*unified.UnifiedLeaderboard, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	lb := &unified.UnifiedLeaderboard{
		Category:  "piano_score",
		Entries:   make([]unified.LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}

	entries := []unified.LeaderboardEntry{
		{
			Rank:        1,
			UserID:      10,
			Username:    "pianist",
			App:         "piano",
			MetricValue: 92.8,
			MetricLabel: "92.8 Score",
			Timestamp:   time.Now(),
		},
		{
			Rank:        2,
			UserID:      11,
			Username:    "maestro",
			App:         "piano",
			MetricValue: 89.5,
			MetricLabel: "89.5 Score",
			Timestamp:   time.Now(),
		},
		{
			Rank:        3,
			UserID:      12,
			Username:    "musiclover",
			App:         "piano",
			MetricValue: 86.2,
			MetricLabel: "86.2 Score",
			Timestamp:   time.Now(),
		},
	}

	for i, entry := range entries {
		if i >= limit {
			break
		}
		lb.Entries = append(lb.Entries, entry)
	}

	return lb, nil
}

// GetOverallLeaderboard returns top performers across all apps (combined ranking)
func (ls *LeaderboardService) GetOverallLeaderboard(ctx context.Context, limit int) (*unified.UnifiedLeaderboard, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	lb := &unified.UnifiedLeaderboard{
		Category:  "overall",
		Entries:   make([]unified.LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}

	entries := []unified.LeaderboardEntry{
		{
			Rank:        1,
			UserID:      1,
			Username:    "allstar",
			App:         "combined",
			MetricValue: 92.5,
			MetricLabel: "92.5 Overall Score",
			Timestamp:   time.Now(),
		},
		{
			Rank:        2,
			UserID:      4,
			Username:    "mathgenius",
			App:         "combined",
			MetricValue: 91.2,
			MetricLabel: "91.2 Overall Score",
			Timestamp:   time.Now(),
		},
		{
			Rank:        3,
			UserID:      7,
			Username:    "balanced",
			App:         "combined",
			MetricValue: 89.8,
			MetricLabel: "89.8 Overall Score",
			Timestamp:   time.Now(),
		},
	}

	for i, entry := range entries {
		if i >= limit {
			break
		}
		lb.Entries = append(lb.Entries, entry)
	}

	return lb, nil
}

// GetLeaderboardByCategory returns leaderboard for any category
func (ls *LeaderboardService) GetLeaderboardByCategory(ctx context.Context, category string, limit int) (*unified.UnifiedLeaderboard, error) {
	switch category {
	case "typing_wpm":
		return ls.GetTypingLeaderboard(ctx, limit)
	case "math_accuracy":
		return ls.GetMathLeaderboard(ctx, limit)
	case "reading_comprehension":
		return ls.GetReadingLeaderboard(ctx, limit)
	case "piano_score":
		return ls.GetPianoLeaderboard(ctx, limit)
	case "overall":
		return ls.GetOverallLeaderboard(ctx, limit)
	default:
		return nil, fmt.Errorf("unknown leaderboard category: %s", category)
	}
}

// GetMultipleLeaderboards returns multiple leaderboards at once
func (ls *LeaderboardService) GetMultipleLeaderboards(ctx context.Context, categories []string, limit int) (map[string]*unified.UnifiedLeaderboard, error) {
	result := make(map[string]*unified.UnifiedLeaderboard)

	for _, category := range categories {
		lb, err := ls.GetLeaderboardByCategory(ctx, category, limit)
		if err != nil {
			continue
		}
		result[category] = lb
	}

	return result, nil
}

// GetUserRank returns a user's rank in a specific category
func (ls *LeaderboardService) GetUserRank(ctx context.Context, userID uint, category string) (int, error) {
	lb, err := ls.GetLeaderboardByCategory(ctx, category, 100)
	if err != nil {
		return 0, err
	}

	for _, entry := range lb.Entries {
		if entry.UserID == userID {
			return entry.Rank, nil
		}
	}

	return 0, fmt.Errorf("user not found in leaderboard")
}

// GetUserRanks returns a user's ranks across all categories
func (ls *LeaderboardService) GetUserRanks(ctx context.Context, userID uint) (map[string]int, error) {
	ranks := make(map[string]int)
	categories := []string{"typing_wpm", "math_accuracy", "reading_comprehension", "piano_score", "overall"}

	for _, category := range categories {
		rank, _ := ls.GetUserRank(ctx, userID, category)
		if rank > 0 {
			ranks[category] = rank
		}
	}

	return ranks, nil
}

// CalculateRankChange calculates change in rank between two snapshots
func (ls *LeaderboardService) CalculateRankChange(oldRank, newRank int) int {
	return oldRank - newRank
}

// IsImprovement determines if rank change is an improvement
func (ls *LeaderboardService) IsImprovement(rankChange int) bool {
	return rankChange > 0
}

// SortLeaderboardEntries sorts entries by rank
func (ls *LeaderboardService) SortLeaderboardEntries(entries []unified.LeaderboardEntry) []unified.LeaderboardEntry {
	sortedEntries := make([]unified.LeaderboardEntry, len(entries))
	copy(sortedEntries, entries)

	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].Rank < sortedEntries[j].Rank
	})

	return sortedEntries
}

// FilterLeaderboardByMetric filters entries by minimum metric value
func (ls *LeaderboardService) FilterLeaderboardByMetric(entries []unified.LeaderboardEntry, minValue float64) []unified.LeaderboardEntry {
	filtered := make([]unified.LeaderboardEntry, 0)

	for _, entry := range entries {
		if entry.MetricValue >= minValue {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// GetLeaderboardStats returns statistics about a leaderboard
func (ls *LeaderboardService) GetLeaderboardStats(lb *unified.UnifiedLeaderboard) map[string]interface{} {
	if len(lb.Entries) == 0 {
		return map[string]interface{}{
			"category":     lb.Category,
			"entry_count":  0,
			"avg_metric":   0.0,
			"min_metric":   0.0,
			"max_metric":   0.0,
			"updated_at":   lb.UpdatedAt,
		}
	}

	totalMetric := 0.0
	minMetric := lb.Entries[0].MetricValue
	maxMetric := lb.Entries[0].MetricValue

	for _, entry := range lb.Entries {
		totalMetric += entry.MetricValue
		if entry.MetricValue < minMetric {
			minMetric = entry.MetricValue
		}
		if entry.MetricValue > maxMetric {
			maxMetric = entry.MetricValue
		}
	}

	avgMetric := totalMetric / float64(len(lb.Entries))

	return map[string]interface{}{
		"category":     lb.Category,
		"entry_count":  len(lb.Entries),
		"avg_metric":   avgMetric,
		"min_metric":   minMetric,
		"max_metric":   maxMetric,
		"updated_at":   lb.UpdatedAt,
	}
}

// GetLeaderboardDistribution returns distribution of metrics
func (ls *LeaderboardService) GetLeaderboardDistribution(lb *unified.UnifiedLeaderboard, buckets int) map[string]int {
	if len(lb.Entries) == 0 || buckets <= 0 {
		return make(map[string]int)
	}

	// Find min/max
	minVal := lb.Entries[0].MetricValue
	maxVal := lb.Entries[0].MetricValue

	for _, entry := range lb.Entries {
		if entry.MetricValue < minVal {
			minVal = entry.MetricValue
		}
		if entry.MetricValue > maxVal {
			maxVal = entry.MetricValue
		}
	}

	// Create buckets
	bucketRange := (maxVal - minVal) / float64(buckets)
	distribution := make(map[string]int)

	for i := 0; i < buckets; i++ {
		bucketStart := minVal + float64(i)*bucketRange
		bucketEnd := bucketStart + bucketRange
		bucketLabel := fmt.Sprintf("%.0f-%.0f", bucketStart, bucketEnd)
		distribution[bucketLabel] = 0
	}

	// Populate buckets
	for _, entry := range lb.Entries {
		bucketIndex := int((entry.MetricValue - minVal) / bucketRange)
		if bucketIndex >= buckets {
			bucketIndex = buckets - 1
		}

		bucketStart := minVal + float64(bucketIndex)*bucketRange
		bucketEnd := bucketStart + bucketRange
		bucketLabel := fmt.Sprintf("%.0f-%.0f", bucketStart, bucketEnd)
		distribution[bucketLabel]++
	}

	return distribution
}
