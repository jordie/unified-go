package dashboard

import (
	"testing"
	"time"
)

func TestNewRankTracker(t *testing.T) {
	rt := NewRankTracker()

	if rt == nil {
		t.Fatal("NewRankTracker returned nil")
	}
	if rt.snapshots == nil {
		t.Fatal("snapshots not initialized")
	}
	if rt.history == nil {
		t.Fatal("history not initialized")
	}
	if rt.maxHistory != 100 {
		t.Errorf("expected maxHistory 100, got %d", rt.maxHistory)
	}
}

func TestRecordSnapshot(t *testing.T) {
	rt := NewRankTracker()

	rt.RecordSnapshot(123, "typing_wpm", 5, 150.5)

	// Verify snapshot was recorded
	rank, exists := rt.GetCurrentRank(123, "typing_wpm")
	if !exists {
		t.Fatal("snapshot not recorded")
	}
	if rank != 5 {
		t.Errorf("expected rank 5, got %d", rank)
	}
}

func TestDetectRankChange(t *testing.T) {
	rt := NewRankTracker()
	userID := uint(123)
	category := "typing_wpm"

	// Record first snapshot
	rt.RecordSnapshot(userID, category, 10, 120.0)
	time.Sleep(10 * time.Millisecond)

	// Record second snapshot with improved rank
	rt.RecordSnapshot(userID, category, 7, 135.0)

	// Detect change
	change := rt.DetectRankChange(userID, category)

	if change == nil {
		t.Fatal("rank change not detected")
	}
	if change.UserID != userID {
		t.Errorf("expected userID %d, got %d", userID, change.UserID)
	}
	if change.Category != category {
		t.Errorf("expected category %s, got %s", category, change.Category)
	}
	if change.PreviousRank != 10 {
		t.Errorf("expected previous rank 10, got %d", change.PreviousRank)
	}
	if change.CurrentRank != 7 {
		t.Errorf("expected current rank 7, got %d", change.CurrentRank)
	}
	if change.RankDelta != 3 {
		t.Errorf("expected rank delta 3 (improved), got %d", change.RankDelta)
	}
	if !change.IsPromotion {
		t.Error("expected IsPromotion to be true")
	}
}

func TestDetectRankChangeNoChange(t *testing.T) {
	rt := NewRankTracker()
	userID := uint(456)
	category := "math_accuracy"

	// Record first snapshot
	rt.RecordSnapshot(userID, category, 5, 95.0)
	time.Sleep(10 * time.Millisecond)

	// Record second snapshot with same rank
	rt.RecordSnapshot(userID, category, 5, 95.5)

	// Detect change
	change := rt.DetectRankChange(userID, category)

	if change != nil {
		t.Error("rank change should not be detected for same rank")
	}
}

func TestDetectRankChangeDemotion(t *testing.T) {
	rt := NewRankTracker()
	userID := uint(789)
	category := "piano_score"

	// Record first snapshot
	rt.RecordSnapshot(userID, category, 3, 9500.0)
	time.Sleep(10 * time.Millisecond)

	// Record second snapshot with worsened rank
	rt.RecordSnapshot(userID, category, 8, 9200.0)

	// Detect change
	change := rt.DetectRankChange(userID, category)

	if change == nil {
		t.Fatal("rank change not detected")
	}
	if change.RankDelta != -5 {
		t.Errorf("expected rank delta -5 (demoted), got %d", change.RankDelta)
	}
	if change.IsPromotion {
		t.Error("expected IsPromotion to be false")
	}
}

func TestGetCurrentRank(t *testing.T) {
	rt := NewRankTracker()

	// Non-existent user
	rank, exists := rt.GetCurrentRank(999, "typing_wpm")
	if exists {
		t.Error("expected rank not to exist")
	}
	if rank != 0 {
		t.Errorf("expected rank 0, got %d", rank)
	}

	// Record snapshot
	rt.RecordSnapshot(123, "typing_wpm", 5, 150.0)

	// Get rank
	rank, exists = rt.GetCurrentRank(123, "typing_wpm")
	if !exists {
		t.Fatal("expected rank to exist")
	}
	if rank != 5 {
		t.Errorf("expected rank 5, got %d", rank)
	}
}

func TestGetRankHistory(t *testing.T) {
	rt := NewRankTracker()
	userID := uint(123)
	category := "typing_wpm"

	// Record multiple snapshots
	for i := 1; i <= 5; i++ {
		rt.RecordSnapshot(userID, category, i, float64(100+i*10))
		time.Sleep(5 * time.Millisecond)
	}

	// Get full history
	history := rt.GetRankHistory(userID, category, 0)
	if len(history) != 5 {
		t.Errorf("expected 5 snapshots, got %d", len(history))
	}

	// Get limited history
	limited := rt.GetRankHistory(userID, category, 2)
	if len(limited) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(limited))
	}

	// Verify last 2 snapshots are returned
	if limited[0].Rank != 4 {
		t.Errorf("expected rank 4, got %d", limited[0].Rank)
	}
	if limited[1].Rank != 5 {
		t.Errorf("expected rank 5, got %d", limited[1].Rank)
	}
}

func TestGetCategoryRanks(t *testing.T) {
	rt := NewRankTracker()
	category := "typing_wpm"

	// Record snapshots for multiple users
	rt.RecordSnapshot(1, category, 5, 150.0)
	rt.RecordSnapshot(2, category, 3, 160.0)
	rt.RecordSnapshot(3, category, 8, 140.0)

	// Get all ranks for category
	ranks := rt.GetCategoryRanks(category)

	if len(ranks) != 3 {
		t.Errorf("expected 3 ranks, got %d", len(ranks))
	}
	if ranks[1] != 5 {
		t.Errorf("expected user 1 rank 5, got %d", ranks[1])
	}
	if ranks[2] != 3 {
		t.Errorf("expected user 2 rank 3, got %d", ranks[2])
	}
	if ranks[3] != 8 {
		t.Errorf("expected user 3 rank 8, got %d", ranks[3])
	}
}

func TestGetAllCategories(t *testing.T) {
	rt := NewRankTracker()

	// Record snapshots in different categories
	rt.RecordSnapshot(1, "typing_wpm", 5, 150.0)
	rt.RecordSnapshot(1, "math_accuracy", 3, 95.0)
	rt.RecordSnapshot(1, "piano_score", 10, 8500.0)

	// Get all categories
	categories := rt.GetAllCategories()

	if len(categories) != 3 {
		t.Errorf("expected 3 categories, got %d", len(categories))
	}
}

func TestClearCategory(t *testing.T) {
	rt := NewRankTracker()

	// Record snapshots
	rt.RecordSnapshot(1, "typing_wpm", 5, 150.0)
	rt.RecordSnapshot(1, "math_accuracy", 3, 95.0)

	// Clear one category
	rt.ClearCategory("typing_wpm")

	// Verify typing_wpm is cleared
	rank, exists := rt.GetCurrentRank(1, "typing_wpm")
	if exists {
		t.Error("expected typing_wpm to be cleared")
	}

	// Verify math_accuracy still exists
	rank, exists = rt.GetCurrentRank(1, "math_accuracy")
	if !exists {
		t.Fatal("expected math_accuracy to still exist")
	}
	if rank != 3 {
		t.Errorf("expected rank 3, got %d", rank)
	}
}

func TestClearAll(t *testing.T) {
	rt := NewRankTracker()

	// Record snapshots
	rt.RecordSnapshot(1, "typing_wpm", 5, 150.0)
	rt.RecordSnapshot(1, "math_accuracy", 3, 95.0)

	// Clear all
	rt.ClearAll()

	// Verify all are cleared
	categories := rt.GetAllCategories()
	if len(categories) != 0 {
		t.Errorf("expected 0 categories after clear, got %d", len(categories))
	}
}

func TestTrackerGetStats(t *testing.T) {
	rt := NewRankTracker()

	// Record snapshots
	rt.RecordSnapshot(1, "typing_wpm", 5, 150.0)
	rt.RecordSnapshot(2, "typing_wpm", 3, 160.0)
	rt.RecordSnapshot(1, "math_accuracy", 8, 85.0)

	// Get stats
	stats := rt.GetStats()

	if stats.CategoriesTracked != 2 {
		t.Errorf("expected 2 categories, got %d", stats.CategoriesTracked)
	}
	if stats.TotalSnapshots != 3 {
		t.Errorf("expected 3 snapshots, got %d", stats.TotalSnapshots)
	}
}

func TestNewRankVelocityAnalyzer(t *testing.T) {
	rt := NewRankTracker()
	analyzer := NewRankVelocityAnalyzer(rt)

	if analyzer == nil {
		t.Fatal("NewRankVelocityAnalyzer returned nil")
	}
	if analyzer.tracker != rt {
		t.Error("tracker not set correctly")
	}
}

func TestRecordChange(t *testing.T) {
	rt := NewRankTracker()
	analyzer := NewRankVelocityAnalyzer(rt)

	change := &RankChange{
		UserID:      123,
		Category:    "typing_wpm",
		PreviousRank: 10,
		CurrentRank: 7,
		RankDelta:   3,
		Velocity:    5.0,
		Timestamp:   time.Now(),
		IsPromotion: true,
	}

	analyzer.RecordChange(change)

	// Verify change was recorded
	avg := analyzer.GetAverageVelocity(123, "typing_wpm", 1)
	if avg != 5.0 {
		t.Errorf("expected velocity 5.0, got %f", avg)
	}
}

func TestGetAverageVelocity(t *testing.T) {
	rt := NewRankTracker()
	analyzer := NewRankVelocityAnalyzer(rt)

	// Record multiple changes
	for i := 1; i <= 3; i++ {
		change := &RankChange{
			UserID:      123,
			Category:    "typing_wpm",
			Velocity:    float64(i) * 2.0, // 2.0, 4.0, 6.0
			Timestamp:   time.Now(),
			IsPromotion: true,
		}
		analyzer.RecordChange(change)
	}

	// Get average velocity
	avg := analyzer.GetAverageVelocity(123, "typing_wpm", 0)
	expected := (2.0 + 4.0 + 6.0) / 3.0
	if avg != expected {
		t.Errorf("expected average %f, got %f", expected, avg)
	}
}

func TestIsMomentum(t *testing.T) {
	rt := NewRankTracker()
	analyzer := NewRankVelocityAnalyzer(rt)

	// Record positive momentum
	for i := 1; i <= 3; i++ {
		change := &RankChange{
			UserID:      123,
			Category:    "typing_wpm",
			Velocity:    2.0, // Positive velocity
			Timestamp:   time.Now(),
			IsPromotion: true,
		}
		analyzer.RecordChange(change)
	}

	// Check momentum
	hasMomentum := analyzer.IsMomentum(123, "typing_wpm", 3)
	if !hasMomentum {
		t.Error("expected user to have momentum")
	}
}

func TestGetRankStreak(t *testing.T) {
	rt := NewRankTracker()
	analyzer := NewRankVelocityAnalyzer(rt)

	// Record upward streak
	for i := 1; i <= 3; i++ {
		change := &RankChange{
			UserID:      123,
			Category:    "typing_wpm",
			RankDelta:   2, // Positive = improvement
			Velocity:    2.0,
			Timestamp:   time.Now(),
			IsPromotion: true,
		}
		analyzer.RecordChange(change)
	}

	// Get streak info
	streak := analyzer.GetRankStreak(123, "typing_wpm")

	if streak == nil {
		t.Fatal("GetRankStreak returned nil")
	}
	if streak.Direction != "up" {
		t.Errorf("expected direction 'up', got %s", streak.Direction)
	}
	if streak.TotalDelta != 6 {
		t.Errorf("expected total delta 6, got %d", streak.TotalDelta)
	}
	if streak.ChangeCount != 3 {
		t.Errorf("expected 3 changes, got %d", streak.ChangeCount)
	}
}

// BenchmarkRecordSnapshot benchmarks snapshot recording
func BenchmarkRecordSnapshot(b *testing.B) {
	rt := NewRankTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt.RecordSnapshot(uint(i%1000), "typing_wpm", i%100, float64(100+i%50))
	}
}

// BenchmarkDetectRankChange benchmarks rank change detection
func BenchmarkDetectRankChange(b *testing.B) {
	rt := NewRankTracker()

	// Pre-populate some data
	for i := 1; i <= 100; i++ {
		rt.RecordSnapshot(uint(i), "typing_wpm", i, float64(100+i))
		time.Sleep(5 * time.Millisecond)
		// Record a second snapshot for each user to enable change detection
		rt.RecordSnapshot(uint(i), "typing_wpm", i-1, float64(100+i+5))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := uint((i % 100) + 1)
		_ = rt.DetectRankChange(userID, "typing_wpm")
	}
}

// BenchmarkGetCategoryRanks benchmarks retrieving all ranks for a category
func BenchmarkGetCategoryRanks(b *testing.B) {
	rt := NewRankTracker()

	// Pre-populate data
	for i := 1; i <= 1000; i++ {
		rt.RecordSnapshot(uint(i), "typing_wpm", i, float64(100+i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rt.GetCategoryRanks("typing_wpm")
	}
}

// TestRankTrackerConcurrency verifies thread-safe access
func TestRankTrackerConcurrency(t *testing.T) {
	rt := NewRankTracker()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				rt.RecordSnapshot(uint(id), "typing_wpm", j, float64(100+j))
				_, _ = rt.GetCurrentRank(uint(id), "typing_wpm")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// If no panic, test passes
}
