package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/jgirmay/unified-go/pkg/unified"
)

// TestNewLeaderboardService tests service creation
func TestNewLeaderboardService(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)

	if lbs == nil {
		t.Fatal("NewLeaderboardService returned nil")
	}
}

// TestGetTypingLeaderboard tests typing leaderboard retrieval
func TestGetTypingLeaderboard(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	lb, err := lbs.GetTypingLeaderboard(ctx, 10)

	if err != nil {
		t.Fatalf("GetTypingLeaderboard failed: %v", err)
	}

	if lb == nil {
		t.Fatal("Leaderboard is nil")
	}

	if lb.Category != "typing_wpm" {
		t.Errorf("Expected category 'typing_wpm', got '%s'", lb.Category)
	}

	if len(lb.Entries) == 0 {
		t.Fatal("Leaderboard has no entries")
	}

	if lb.Entries[0].Rank != 1 {
		t.Errorf("Expected first entry rank 1, got %d", lb.Entries[0].Rank)
	}
}

// TestGetMathLeaderboard tests math leaderboard retrieval
func TestGetMathLeaderboard(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	lb, err := lbs.GetMathLeaderboard(ctx, 10)

	if err != nil {
		t.Fatalf("GetMathLeaderboard failed: %v", err)
	}

	if lb.Category != "math_accuracy" {
		t.Errorf("Expected category 'math_accuracy', got '%s'", lb.Category)
	}

	if len(lb.Entries) == 0 {
		t.Fatal("Leaderboard has no entries")
	}
}

// TestGetReadingLeaderboard tests reading leaderboard retrieval
func TestGetReadingLeaderboard(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	lb, err := lbs.GetReadingLeaderboard(ctx, 10)

	if err != nil {
		t.Fatalf("GetReadingLeaderboard failed: %v", err)
	}

	if lb.Category != "reading_comprehension" {
		t.Errorf("Expected category 'reading_comprehension', got '%s'", lb.Category)
	}
}

// TestGetPianoLeaderboard tests piano leaderboard retrieval
func TestGetPianoLeaderboard(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	lb, err := lbs.GetPianoLeaderboard(ctx, 10)

	if err != nil {
		t.Fatalf("GetPianoLeaderboard failed: %v", err)
	}

	if lb.Category != "piano_score" {
		t.Errorf("Expected category 'piano_score', got '%s'", lb.Category)
	}
}

// TestGetOverallLeaderboard tests overall leaderboard retrieval
func TestGetOverallLeaderboard(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	lb, err := lbs.GetOverallLeaderboard(ctx, 10)

	if err != nil {
		t.Fatalf("GetOverallLeaderboard failed: %v", err)
	}

	if lb.Category != "overall" {
		t.Errorf("Expected category 'overall', got '%s'", lb.Category)
	}
}

// TestGetLeaderboardByCategory tests dispatcher method
func TestGetLeaderboardByCategory(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	categories := []string{"typing_wpm", "math_accuracy", "reading_comprehension", "piano_score", "overall"}

	for _, category := range categories {
		lb, err := lbs.GetLeaderboardByCategory(ctx, category, 10)
		if err != nil {
			t.Errorf("Failed to get %s leaderboard: %v", category, err)
		}

		if lb.Category != category {
			t.Errorf("Expected category %s, got %s", category, lb.Category)
		}
	}
}

// TestGetLeaderboardByCategoryInvalid tests error handling
func TestGetLeaderboardByCategoryInvalid(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	_, err := lbs.GetLeaderboardByCategory(ctx, "invalid_category", 10)
	if err == nil {
		t.Error("Expected error for invalid category")
	}
}

// TestGetMultipleLeaderboards tests fetching multiple leaderboards
func TestGetMultipleLeaderboards(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	categories := []string{"typing_wpm", "math_accuracy", "overall"}
	result, err := lbs.GetMultipleLeaderboards(ctx, categories, 10)

	if err != nil {
		t.Fatalf("GetMultipleLeaderboards failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 leaderboards, got %d", len(result))
	}

	for _, category := range categories {
		if _, ok := result[category]; !ok {
			t.Errorf("Missing leaderboard for category %s", category)
		}
	}
}

// TestGetUserRank tests user rank lookup
func TestGetUserRank(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	// User 1 is in typing leaderboard at rank 1
	rank, err := lbs.GetUserRank(ctx, 1, "typing_wpm")
	if err != nil {
		t.Fatalf("GetUserRank failed: %v", err)
	}

	if rank != 1 {
		t.Errorf("Expected rank 1, got %d", rank)
	}
}

// TestGetUserRankNotFound tests user not in leaderboard
func TestGetUserRankNotFound(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	// User 999 is not in leaderboard
	_, err := lbs.GetUserRank(ctx, 999, "typing_wpm")
	if err == nil {
		t.Error("Expected error for user not in leaderboard")
	}
}

// TestGetUserRanks tests multiple rank lookups
func TestGetUserRanks(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	ranks, err := lbs.GetUserRanks(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserRanks failed: %v", err)
	}

	if len(ranks) == 0 {
		t.Fatal("No ranks returned")
	}
}

// TestCalculateRankChange tests rank change calculation
func TestCalculateRankChange(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)

	// Improved from rank 10 to rank 5
	change := lbs.CalculateRankChange(10, 5)
	if change != 5 {
		t.Errorf("Expected change 5, got %d", change)
	}

	// Declined from rank 5 to rank 10
	change = lbs.CalculateRankChange(5, 10)
	if change != -5 {
		t.Errorf("Expected change -5, got %d", change)
	}

	// No change
	change = lbs.CalculateRankChange(10, 10)
	if change != 0 {
		t.Errorf("Expected change 0, got %d", change)
	}
}

// TestIsImprovement tests improvement detection
func TestIsImprovement(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)

	if !lbs.IsImprovement(5) {
		t.Error("Expected positive rank change to be improvement")
	}

	if lbs.IsImprovement(-5) {
		t.Error("Expected negative rank change to not be improvement")
	}

	if lbs.IsImprovement(0) {
		t.Error("Expected zero rank change to not be improvement")
	}
}

// TestSortLeaderboardEntries tests sorting functionality
func TestSortLeaderboardEntries(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)

	entries := []unified.LeaderboardEntry{
		{Rank: 3, UserID: 3, Username: "third"},
		{Rank: 1, UserID: 1, Username: "first"},
		{Rank: 2, UserID: 2, Username: "second"},
	}

	sorted := lbs.SortLeaderboardEntries(entries)

	if sorted[0].Rank != 1 {
		t.Errorf("Expected first entry rank 1, got %d", sorted[0].Rank)
	}

	if sorted[1].Rank != 2 {
		t.Errorf("Expected second entry rank 2, got %d", sorted[1].Rank)
	}

	if sorted[2].Rank != 3 {
		t.Errorf("Expected third entry rank 3, got %d", sorted[2].Rank)
	}
}

// TestFilterLeaderboardByMetric tests filtering functionality
func TestFilterLeaderboardByMetric(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)

	entries := []unified.LeaderboardEntry{
		{MetricValue: 100.0},
		{MetricValue: 80.0},
		{MetricValue: 60.0},
		{MetricValue: 40.0},
	}

	filtered := lbs.FilterLeaderboardByMetric(entries, 70.0)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered entries, got %d", len(filtered))
	}

	for _, entry := range filtered {
		if entry.MetricValue < 70.0 {
			t.Errorf("Filtered entry should have metric >= 70.0, got %f", entry.MetricValue)
		}
	}
}

// TestGetLeaderboardStats tests stats calculation
func TestGetLeaderboardStats(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	lb, _ := lbs.GetTypingLeaderboard(ctx, 10)
	stats := lbs.GetLeaderboardStats(lb)

	if stats["entry_count"].(int) != len(lb.Entries) {
		t.Errorf("Stat entry_count mismatch")
	}

	if avg, ok := stats["avg_metric"].(float64); !ok || avg <= 0 {
		t.Error("Average metric should be positive")
	}

	if max, ok := stats["max_metric"].(float64); !ok || max <= 0 {
		t.Error("Max metric should be positive")
	}
}

// TestGetLeaderboardStatsEmpty tests stats with empty leaderboard
func TestGetLeaderboardStatsEmpty(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)

	lb := &unified.UnifiedLeaderboard{
		Category:  "test",
		Entries:   make([]unified.LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}

	stats := lbs.GetLeaderboardStats(lb)

	if stats["entry_count"].(int) != 0 {
		t.Error("Expected 0 entries for empty leaderboard")
	}
}

// TestGetLeaderboardDistribution tests distribution calculation
func TestGetLeaderboardDistribution(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	lb, _ := lbs.GetTypingLeaderboard(ctx, 10)
	distribution := lbs.GetLeaderboardDistribution(lb, 5)

	if len(distribution) != 5 {
		t.Errorf("Expected 5 distribution buckets, got %d", len(distribution))
	}

	// Sum of distribution should equal number of entries
	total := 0
	for _, count := range distribution {
		total += count
	}

	if total != len(lb.Entries) {
		t.Errorf("Distribution total %d != entry count %d", total, len(lb.Entries))
	}
}

// TestGetLeaderboardDistributionEdgeCases tests edge cases
func TestGetLeaderboardDistributionEdgeCases(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)

	// Empty leaderboard
	lb := &unified.UnifiedLeaderboard{
		Category:  "test",
		Entries:   make([]unified.LeaderboardEntry, 0),
		UpdatedAt: time.Now(),
	}
	distribution := lbs.GetLeaderboardDistribution(lb, 5)
	if len(distribution) != 0 {
		t.Error("Expected empty distribution for empty leaderboard")
	}

	// Zero buckets
	lb.Entries = append(lb.Entries, unified.LeaderboardEntry{MetricValue: 100})
	distribution = lbs.GetLeaderboardDistribution(lb, 0)
	if len(distribution) != 0 {
		t.Error("Expected empty distribution for zero buckets")
	}
}

// TestLeaderboardLimitValidation tests limit handling
func TestLeaderboardLimitValidation(t *testing.T) {
	service := NewService(nil)
	lbs := NewLeaderboardService(service)
	ctx := context.Background()

	// Negative limit should default to 20
	lb, _ := lbs.GetTypingLeaderboard(ctx, -10)
	if len(lb.Entries) > 20 {
		t.Error("Negative limit should be adjusted")
	}

	// Zero limit should default to 20
	lb, _ = lbs.GetTypingLeaderboard(ctx, 0)
	if len(lb.Entries) > 20 {
		t.Error("Zero limit should be adjusted")
	}

	// Limit > 100 should be capped
	lb, _ = lbs.GetTypingLeaderboard(ctx, 200)
	if len(lb.Entries) > 20 {
		t.Error("Large limit should be capped at 100")
	}
}
