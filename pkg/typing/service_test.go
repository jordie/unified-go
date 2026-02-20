package typing

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupServiceTestDB creates a test database with sample data
func setupServiceTestDB(t *testing.T) (*sql.DB, *Service) {
	db := setupTestDB(t)

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)
	service := NewService(repo)

	return db, service
}

// TestCalculateWPM tests WPM calculation
func TestCalculateWPM(t *testing.T) {
	_, service := setupServiceTestDB(t)

	tests := []struct {
		name            string
		content         string
		timeSpent       float64
		expectedMinWPM  float64
		expectedMaxWPM  float64
	}{
		{
			name:           "typical typing test",
			content:        "the quick brown fox jumps over the lazy dog", // 44 chars
			timeSpent:      60,                                            // 1 minute
			expectedMinWPM: 8.5,
			expectedMaxWPM: 9.0,
		},
		{
			name:           "fast typing",
			content:        "abcdefghijklmnopqrstuvwxyz", // 26 chars
			timeSpent:      30,                            // 30 seconds
			expectedMinWPM: 10.0,
			expectedMaxWPM: 11.0,
		},
		{
			name:           "slow typing",
			content:        "hello world", // 11 chars
			timeSpent:      120,           // 2 minutes
			expectedMinWPM: 0.9,
			expectedMaxWPM: 1.1,
		},
		{
			name:           "zero time spent",
			content:        "test",
			timeSpent:      0,
			expectedMinWPM: 0,
			expectedMaxWPM: 0,
		},
		{
			name:           "very short content",
			content:        "a",
			timeSpent:      1,
			expectedMinWPM: 10.0,
			expectedMaxWPM: 14.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wpm := service.CalculateWPM(tt.content, tt.timeSpent)
			if tt.expectedMinWPM == 0 && tt.expectedMaxWPM == 0 {
				if wpm != 0 {
					t.Errorf("CalculateWPM() = %v, want 0", wpm)
				}
			} else if wpm < tt.expectedMinWPM || wpm > tt.expectedMaxWPM {
				t.Errorf("CalculateWPM() = %v, expected between %v and %v", wpm, tt.expectedMinWPM, tt.expectedMaxWPM)
			}
		})
	}
}

// TestCalculateAccuracy tests accuracy calculation
func TestCalculateAccuracy(t *testing.T) {
	_, service := setupServiceTestDB(t)

	tests := []struct {
		name            string
		typed           string
		expected        string
		expectedMinAcc  float64
		expectedMaxAcc  float64
	}{
		{
			name:           "perfect match",
			typed:          "hello world",
			expected:       "hello world",
			expectedMinAcc: 99.9,
			expectedMaxAcc: 100.0,
		},
		{
			name:           "one character wrong",
			typed:          "hallo world",
			expected:       "hello world",
			expectedMinAcc: 90.0,
			expectedMaxAcc: 100.0,
		},
		{
			name:           "typed less than expected",
			typed:          "hello world test",
			expected:       "hello world",
			expectedMinAcc: 80.0,
			expectedMaxAcc: 100.0,
		},
		{
			name:           "typed more than expected",
			typed:          "hello world extra",
			expected:       "hello world",
			expectedMinAcc: 50.0,
			expectedMaxAcc: 100.0,
		},
		{
			name:           "completely different",
			typed:          "abcde",
			expected:       "hello",
			expectedMinAcc: 0.0,
			expectedMaxAcc: 20.0,
		},
		{
			name:           "whitespace differences",
			typed:          "  hello world  ",
			expected:       "hello world",
			expectedMinAcc: 99.9,
			expectedMaxAcc: 100.0,
		},
		{
			name:           "empty typed",
			typed:          "",
			expected:       "hello",
			expectedMinAcc: 0.0,
			expectedMaxAcc: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accuracy := service.CalculateAccuracy(tt.typed, tt.expected)
			if accuracy < tt.expectedMinAcc || accuracy > tt.expectedMaxAcc {
				t.Errorf("CalculateAccuracy() = %v, expected between %v and %v", accuracy, tt.expectedMinAcc, tt.expectedMaxAcc)
			}
		})
	}
}

// TestProcessTestResult tests complete test result processing
func TestProcessTestResult(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	result, err := service.ProcessTestResult(
		ctx,
		1,
		"the quick brown fox jumps over the lazy dog",
		60,
		2,
	)

	if err != nil {
		t.Fatalf("ProcessTestResult() error = %v", err)
	}

	if result.ID == 0 {
		t.Error("ProcessTestResult() returned zero ID")
	}

	if result.UserID != 1 {
		t.Errorf("ProcessTestResult() UserID = %v, want 1", result.UserID)
	}

	if result.WPM <= 0 {
		t.Errorf("ProcessTestResult() WPM = %v, want > 0", result.WPM)
	}

	if result.Accuracy <= 0 || result.Accuracy > 100 {
		t.Errorf("ProcessTestResult() Accuracy = %v, want 0-100", result.Accuracy)
	}
}

// TestProcessTestResultInvalid tests invalid test result processing
func TestProcessTestResultInvalid(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		userID  uint
		content string
		time    float64
		errors  int
		wantErr bool
	}{
		{
			name:    "missing user_id",
			userID:  0,
			content: "test",
			time:    60,
			errors:  0,
			wantErr: true,
		},
		{
			name:    "empty content",
			userID:  1,
			content: "",
			time:    60,
			errors:  0,
			wantErr: true,
		},
		{
			name:    "zero time",
			userID:  1,
			content: "test",
			time:    0,
			errors:  0,
			wantErr: true,
		},
		{
			name:    "negative errors",
			userID:  1,
			content: "test",
			time:    60,
			errors:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ProcessTestResult(ctx, tt.userID, tt.content, tt.time, tt.errors)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessTestResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetUserStatistics tests retrieving user statistics
func TestGetUserStatistics(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Process a test to generate statistics
	_, err := service.ProcessTestResult(ctx, 1, "test content for statistics", 60, 1)
	if err != nil {
		t.Fatalf("ProcessTestResult() error = %v", err)
	}

	stats, err := service.GetUserStatistics(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStatistics() error = %v", err)
	}

	if stats.UserID != 1 {
		t.Errorf("GetUserStatistics() UserID = %v, want 1", stats.UserID)
	}

	if stats.TotalTests == 0 {
		t.Error("GetUserStatistics() TotalTests = 0, want > 0")
	}
}

// TestServiceGetLeaderboard tests retrieving leaderboard through service
func TestServiceGetLeaderboard(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Insert test data for multiple users
	for i := 2; i <= 5; i++ {
		username := ("user" + string(rune(i)))
		_, err := db.Exec("INSERT INTO users (id, username) VALUES (?, ?)", i, username)
		if err != nil {
			t.Fatalf("failed to insert user: %v", err)
		}

		// Process tests for each user
		for j := 0; j < 3; j++ {
			_, err = service.ProcessTestResult(
				ctx,
				uint(i),
				"the quick brown fox jumps over the lazy dog",
				60,
				1,
			)
			if err != nil {
				t.Fatalf("ProcessTestResult() error = %v", err)
			}
		}
	}

	// Get leaderboard
	leaderboard, err := service.GetLeaderboard(ctx, 10)
	if err != nil {
		t.Fatalf("GetLeaderboard() error = %v", err)
	}

	if len(leaderboard) == 0 {
		t.Error("GetLeaderboard() returned empty leaderboard")
	}
}

// TestValidateTestContent tests content validation
func TestValidateTestContent(t *testing.T) {
	_, service := setupServiceTestDB(t)

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid content",
			content: "the quick brown fox jumps over the lazy dog",
			wantErr: false,
		},
		{
			name:    "too short",
			content: "hi",
			wantErr: true,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			content: "   ",
			wantErr: true,
		},
		{
			name:    "no letters",
			content: "123456789012345",
			wantErr: true,
		},
		{
			name:    "valid with numbers",
			content: "the quick 123 fox jumps",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateTestContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTestContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetUserTestCount tests getting test count
func TestGetUserTestCount(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Process multiple tests
	for i := 0; i < 5; i++ {
		_, err := service.ProcessTestResult(ctx, 1, "test content", 60, 1)
		if err != nil {
			t.Fatalf("ProcessTestResult() error = %v", err)
		}
	}

	count, err := service.GetUserTestCount(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserTestCount() error = %v", err)
	}

	if count != 5 {
		t.Errorf("GetUserTestCount() = %d, want 5", count)
	}
}

// TestGetUserTestHistory tests retrieving test history
func TestGetUserTestHistory(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Process multiple tests
	for i := 0; i < 10; i++ {
		_, err := service.ProcessTestResult(ctx, 1, "test content", 60, 1)
		if err != nil {
			t.Fatalf("ProcessTestResult() error = %v", err)
		}
	}

	// Get history with pagination
	tests, err := service.GetUserTestHistory(ctx, 1, 5, 0)
	if err != nil {
		t.Fatalf("GetUserTestHistory() error = %v", err)
	}

	if len(tests) != 5 {
		t.Errorf("GetUserTestHistory() returned %d tests, want 5", len(tests))
	}
}

// TestEstimateUserLevel tests user level estimation
func TestEstimateUserLevel(t *testing.T) {
	tests := []struct {
		wpm           float64
		expectedLevel string
	}{
		{wpm: 25, expectedLevel: "beginner"},
		{wpm: 50, expectedLevel: "intermediate"},
		{wpm: 70, expectedLevel: "advanced"},
		{wpm: 100, expectedLevel: "expert"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedLevel, func(t *testing.T) {
			level := estimateUserLevel(tt.wpm)
			if level != tt.expectedLevel {
				t.Errorf("estimateUserLevel(%v) = %s, want %s", tt.wpm, level, tt.expectedLevel)
			}
		})
	}
}

// TestCalculateUserProgress tests progress calculation
func TestCalculateUserProgress(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Process a test
	_, err := service.ProcessTestResult(ctx, 1, "test content for progress", 60, 1)
	if err != nil {
		t.Fatalf("ProcessTestResult() error = %v", err)
	}

	progress, err := service.CalculateUserProgress(ctx, 1)
	if err != nil {
		t.Fatalf("CalculateUserProgress() error = %v", err)
	}

	if progress["user_id"] != uint(1) {
		t.Errorf("CalculateUserProgress() user_id = %v, want 1", progress["user_id"])
	}

	if progress["total_tests"] == nil {
		t.Error("CalculateUserProgress() missing total_tests")
	}

	if progress["estimated_level"] == nil {
		t.Error("CalculateUserProgress() missing estimated_level")
	}
}

// TestNewServiceWithPool tests creating service with pool
func TestNewServiceWithPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	service := NewServiceWithPool(pool)

	if service == nil {
		t.Error("NewServiceWithPool() returned nil")
	}

	if service.repo == nil {
		t.Error("NewServiceWithPool() service has nil repo")
	}
}

// BenchmarkCalculateWPM benchmarks WPM calculation
func BenchmarkCalculateWPM(b *testing.B) {
	_, service := setupServiceTestDB(&testing.T{})

	content := "the quick brown fox jumps over the lazy dog the quick brown fox jumps over the lazy dog"
	timeSpent := 120.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateWPM(content, timeSpent)
	}
}

// BenchmarkCalculateAccuracy benchmarks accuracy calculation
func BenchmarkCalculateAccuracy(b *testing.B) {
	_, service := setupServiceTestDB(&testing.T{})

	typed := "the quick brown fox jumps over the lazy dog"
	expected := "the quick brown fox jumps over the lazy dog"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateAccuracy(typed, expected)
	}
}
