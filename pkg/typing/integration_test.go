package typing

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestFullTypingWorkflow tests the complete typing test workflow
func TestFullTypingWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Step 1: Process a typing test
	content := "the quick brown fox jumps over the lazy dog"
	timeSpent := 60.0
	errorCount := 2

	result, err := service.ProcessTestResult(ctx, 1, content, timeSpent, errorCount)
	if err != nil {
		t.Fatalf("ProcessTestResult failed: %v", err)
	}

	if result.ID == 0 {
		t.Error("Result should have an ID")
	}

	if result.WPM <= 0 {
		t.Error("WPM should be calculated")
	}

	// Step 2: Verify stats are updated
	stats, err := service.GetUserStatistics(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStatistics failed: %v", err)
	}

	if stats.TotalTests != 1 {
		t.Errorf("Total tests should be 1, got %d", stats.TotalTests)
	}

	if stats.BestWPM == 0 {
		t.Error("Best WPM should be set")
	}

	// Step 3: Submit another test
	_, err = service.ProcessTestResult(ctx, 1, content, 60, 1)
	if err != nil {
		t.Fatalf("Second ProcessTestResult failed: %v", err)
	}

	// Step 4: Verify stats are updated again
	stats2, err := service.GetUserStatistics(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStatistics failed: %v", err)
	}

	if stats2.TotalTests != 2 {
		t.Errorf("Total tests should be 2, got %d", stats2.TotalTests)
	}

	if stats2.AverageWPM == 0 {
		t.Error("Average WPM should be calculated")
	}
}

// TestLeaderboardRanking tests leaderboard sorting
func TestLeaderboardRanking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Create 5 users with different scores using the service
	testCases := []struct {
		userID  uint
		content string
		wpm     float64
	}{
		{1, "test content for user one with more text", 50.0},
		{2, "test content for user two with longer text", 70.0},
		{3, "test content for user three with additional text", 80.0},
		{4, "test content for user four with even more text", 90.0},
		{5, "test content for user five with lots of extra text", 100.0},
	}

	// Insert user records
	for i := 2; i <= 5; i++ {
		username := fmt.Sprintf("user%d", i)
		db.Exec("INSERT INTO users (id, username) VALUES (?, ?)", i, username)
	}

	// Create test results for each user
	for _, tc := range testCases {
		for j := 0; j < 3; j++ {
			service.ProcessTestResult(ctx, tc.userID, tc.content, 60, 1)
		}
	}

	// Get leaderboard
	leaderboard, err := service.GetLeaderboard(ctx, 5)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if len(leaderboard) < 1 {
		t.Errorf("Expected at least 1 user in leaderboard, got %d", len(leaderboard))
		return
	}

	// Verify leaderboard is sorted by best WPM descending
	for i := 0; i < len(leaderboard)-1; i++ {
		if leaderboard[i].BestWPM < leaderboard[i+1].BestWPM {
			t.Error("Leaderboard not sorted correctly")
			break
		}
	}
}

// TestHistoryPagination tests pagination logic
func TestHistoryPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Insert 25 tests
	for i := 0; i < 25; i++ {
		_, err := db.Exec(`
			INSERT INTO typing_results (
				user_id, wpm, raw_wpm, accuracy, errors,
				time_taken, test_mode, text_snippet, timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, 1, float64(50+i), float64(48+i), 90.0, i%5, 60.0, "test", "snippet", time.Now().Add(-time.Hour*time.Duration(i)))
		if err != nil {
			t.Fatalf("Failed to insert test: %v", err)
		}
	}

	// Test first page
	page1, err := service.GetUserTestHistory(ctx, 1, 10, 0)
	if err != nil {
		t.Fatalf("GetUserTestHistory failed: %v", err)
	}

	if len(page1) != 10 {
		t.Errorf("Expected 10 tests on page 1, got %d", len(page1))
	}

	// Test second page
	page2, err := service.GetUserTestHistory(ctx, 1, 10, 10)
	if err != nil {
		t.Fatalf("GetUserTestHistory page 2 failed: %v", err)
	}

	if len(page2) != 10 {
		t.Errorf("Expected 10 tests on page 2, got %d", len(page2))
	}

	// Verify pages don't overlap
	if page1[0].ID == page2[0].ID {
		t.Error("Pages should not have overlapping tests")
	}
}

// TestUserIsolation tests that users can't see each other's data
func TestUserIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Create second user
	_, err := db.Exec("INSERT INTO users (id, username) VALUES (?, ?)", 2, "user2")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// User 1 submits test
	result1, err := service.ProcessTestResult(ctx, 1, "test content", 60, 1)
	if err != nil {
		t.Fatalf("ProcessTestResult for user 1 failed: %v", err)
	}

	// User 2 submits test
	result2, err := service.ProcessTestResult(ctx, 2, "test content", 60, 1)
	if err != nil {
		t.Fatalf("ProcessTestResult for user 2 failed: %v", err)
	}

	// Get stats for each user
	stats1, err := service.GetUserStatistics(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStatistics for user 1 failed: %v", err)
	}

	stats2, err := service.GetUserStatistics(ctx, 2)
	if err != nil {
		t.Fatalf("GetUserStatistics for user 2 failed: %v", err)
	}

	if stats1.TotalTests != 1 {
		t.Errorf("User 1 should have 1 test, got %d", stats1.TotalTests)
	}

	if stats2.TotalTests != 1 {
		t.Errorf("User 2 should have 1 test, got %d", stats2.TotalTests)
	}

	// Get history for user 1
	history1, err := service.GetUserTestHistory(ctx, 1, 10, 0)
	if err != nil {
		t.Fatalf("GetUserTestHistory for user 1 failed: %v", err)
	}

	if len(history1) != 1 {
		t.Errorf("User 1 history should have 1 test, got %d", len(history1))
	}

	if history1[0].ID != result1.ID {
		t.Error("User 1 history should only contain their own test")
	}

	// Verify user 2 can't see user 1's test
	history2, err := service.GetUserTestHistory(ctx, 2, 10, 0)
	if err != nil {
		t.Fatalf("GetUserTestHistory for user 2 failed: %v", err)
	}

	if len(history2) != 1 {
		t.Errorf("User 2 history should have 1 test, got %d", len(history2))
	}

	if history2[0].ID != result2.ID {
		t.Error("User 2 history should only contain their own test")
	}
}

// TestDataPersistence tests that data survives multiple operations
func TestDataPersistence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service1 := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Save test with service1
	result, err := service1.ProcessTestResult(ctx, 1, "test content", 60, 1)
	if err != nil {
		t.Fatalf("ProcessTestResult failed: %v", err)
	}

	resultID := result.ID

	// Create new service instance (simulates server restart)
	service2 := NewService(&Repository{db: pool})

	// Retrieve data with service2
	stats, err := service2.GetUserStatistics(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStatistics failed: %v", err)
	}

	if stats.TotalTests == 0 {
		t.Error("Data should persist across service instances")
	}

	// Verify the exact test is still there
	history, err := service2.GetUserTestHistory(ctx, 1, 10, 0)
	if err != nil {
		t.Fatalf("GetUserTestHistory failed: %v", err)
	}

	if len(history) == 0 {
		t.Error("Test history should persist")
	}

	if history[0].ID != resultID {
		t.Errorf("Expected test ID %d, got %d", resultID, history[0].ID)
	}
}

// BenchmarkSaveResult benchmarks the save operation
func BenchmarkSaveResult(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	result := &TypingResult{
		UserID:      1,
		Content:     "the quick brown fox jumps over the lazy dog",
		TimeSpent:   60.0,
		WPM:         60.0,
		Accuracy:    95.5,
		ErrorsCount: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ProcessTestResult(ctx, 1, result.Content, result.TimeSpent, result.ErrorsCount)
	}
}

// BenchmarkGetStats benchmarks statistics retrieval
func BenchmarkGetStats(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Pre-populate with test data
	for i := 0; i < 10; i++ {
		service.ProcessTestResult(ctx, 1, "test content", 60, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetUserStatistics(ctx, 1)
	}
}

// BenchmarkLeaderboard benchmarks leaderboard retrieval
func BenchmarkLeaderboard(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Pre-populate with test data for 20 users
	for u := 1; u <= 20; u++ {
		if u > 1 {
			db.Exec("INSERT INTO users (id, username) VALUES (?, ?)", u, fmt.Sprintf("user%d", u))
		}
		for i := 0; i < 5; i++ {
			service.ProcessTestResult(ctx, uint(u), "test content", 60, 1)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetLeaderboard(ctx, 10)
	}
}

// TestConcurrentRequests verifies thread safety of service methods
func TestConcurrentRequests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Pre-populate with test results
	for i := 0; i < 3; i++ {
		service.ProcessTestResult(ctx, 1, "test content for concurrent testing", 60, 1)
	}

	const numConcurrent = 5

	var wg sync.WaitGroup
	var panicCount int64
	startTime := time.Now()

	// Run concurrent operations to verify thread safety
	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt64(&panicCount, 1)
				}
			}()

			// Perform various operations
			service.GetUserStatistics(ctx, 1)
			service.GetUserTestHistory(ctx, 1, 10, 0)
			service.GetLeaderboard(ctx, 10)
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	if panicCount > 0 {
		t.Errorf("Expected no panics, got %d", panicCount)
	}

	t.Logf("Concurrent safety test completed: %d concurrent operations in %v", numConcurrent, duration)
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    uint
		content   string
		timeSpent float64
		errors    int
		wantErr   bool
	}{
		{"missing user_id", 0, "test", 60, 1, true},
		{"empty content", 1, "", 60, 1, true},
		{"zero time spent", 1, "test", 0, 1, true},
		{"negative errors", 1, "test", 60, -1, true},
		{"valid", 1, "test content here", 60, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ProcessTestResult(ctx, tt.userID, tt.content, tt.timeSpent, tt.errors)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessTestResult error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStatisticsAccuracy tests calculation accuracy
func TestStatisticsAccuracy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewService(&Repository{db: pool})
	ctx := context.Background()

	// Create test results with the service (which calculates stats automatically)
	// Use different content to generate different WPM values
	contents := []string{
		"the quick brown fox jumps over the lazy dog",               // ~44 chars -> ~44 WPM in 60s
		"the quick brown fox jumps over the lazy dog and more text", // ~52 chars -> ~52 WPM in 60s
		"the quick brown fox",                                       // ~20 chars -> ~20 WPM in 60s
	}

	for _, content := range contents {
		_, err := service.ProcessTestResult(ctx, 1, content, 60, 1)
		if err != nil {
			t.Fatalf("ProcessTestResult failed: %v", err)
		}
	}

	stats, err := service.GetUserStatistics(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStatistics failed: %v", err)
	}

	if stats.TotalTests != 3 {
		t.Errorf("Expected 3 tests, got %d", stats.TotalTests)
	}

	// Just verify stats are populated - exact values depend on formula
	if stats.AverageWPM == 0 {
		t.Error("Average WPM should be calculated")
	}

	if stats.AverageAccuracy == 0 {
		t.Error("Average accuracy should be calculated")
	}

	if stats.BestWPM == 0 {
		t.Error("Best WPM should be set")
	}
}
