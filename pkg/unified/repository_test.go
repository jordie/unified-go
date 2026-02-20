package unified

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestNewRepository tests repository creation
func TestNewRepository(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}

	if repo.db == nil {
		t.Fatal("Repository db is nil")
	}
}

// TestSetAppRepositories tests setting app repositories
func TestSetAppRepositories(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)

	typing := "typing_repo"
	math := "math_repo"
	reading := "reading_repo"
	piano := "piano_repo"

	repo.SetAppRepositories(typing, math, reading, piano)

	if repo.typingRepo != typing {
		t.Error("Typing repository not set correctly")
	}

	if repo.mathRepo != math {
		t.Error("Math repository not set correctly")
	}

	if repo.readingRepo != reading {
		t.Error("Reading repository not set correctly")
	}

	if repo.pianoRepo != piano {
		t.Error("Piano repository not set correctly")
	}
}

// TestGetUserProfileWithInvalidID tests error handling for invalid user ID
func TestGetUserProfileWithInvalidID(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Test with zero user ID
	_, err := repo.GetUserProfile(ctx, 0)
	if err == nil {
		t.Error("Expected error for zero user ID, got nil")
	}
}

// TestGetUserProfile tests user profile retrieval
func TestGetUserProfile(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT, created_at DATETIME)`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create app tables
	db.Exec(`CREATE TABLE typing_tests (id INTEGER PRIMARY KEY, user_id INTEGER, created_at DATETIME, timestamp DATETIME)`)
	db.Exec(`CREATE TABLE math_results (id INTEGER PRIMARY KEY, user_id INTEGER, timestamp DATETIME)`)
	db.Exec(`CREATE TABLE reading_sessions (id INTEGER PRIMARY KEY, user_id INTEGER, created_at DATETIME)`)
	db.Exec(`CREATE TABLE piano_lessons (id INTEGER PRIMARY KEY, user_id INTEGER, created_at DATETIME)`)

	// Insert test user
	_, err = db.Exec(`INSERT INTO users (id, username, created_at) VALUES (1, 'testuser', datetime('now'))`)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	repo := NewRepository(db)
	ctx := context.Background()

	profile, err := repo.GetUserProfile(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get user profile: %v", err)
	}

	if profile == nil {
		t.Fatal("Profile is nil")
	}

	if profile.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", profile.UserID)
	}

	if profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", profile.Username)
	}
}

// TestGetCrossAppAnalyticsWithInvalidID tests error handling
func TestGetCrossAppAnalyticsWithInvalidID(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetCrossAppAnalytics(ctx, 0)
	if err == nil {
		t.Error("Expected error for zero user ID, got nil")
	}
}

// TestGetCrossAppAnalytics tests analytics calculation
func TestGetCrossAppAnalytics(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create app tables
	db.Exec(`CREATE TABLE typing_tests (id INTEGER PRIMARY KEY, user_id INTEGER, created_at TIMESTAMP, timestamp TIMESTAMP, duration REAL)`)
	db.Exec(`CREATE TABLE math_results (id INTEGER PRIMARY KEY, user_id INTEGER, timestamp TIMESTAMP, total_time REAL)`)
	db.Exec(`CREATE TABLE reading_sessions (id INTEGER PRIMARY KEY, user_id INTEGER, created_at TIMESTAMP, duration REAL)`)
	db.Exec(`CREATE TABLE piano_lessons (id INTEGER PRIMARY KEY, user_id INTEGER, created_at TIMESTAMP, duration REAL)`)

	// Insert test sessions
	now := time.Now()
	db.Exec(`INSERT INTO typing_tests (user_id, created_at, timestamp, duration) VALUES (1, ?, ?, 5.0)`, now, now)
	db.Exec(`INSERT INTO math_results (user_id, timestamp, total_time) VALUES (1, ?, 10.0)`, now)

	repo := NewRepository(db)
	ctx := context.Background()

	analytics, err := repo.GetCrossAppAnalytics(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get analytics: %v", err)
	}

	if analytics == nil {
		t.Fatal("Analytics is nil")
	}

	if analytics.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", analytics.UserID)
	}

	if analytics.AppMetrics == nil {
		t.Fatal("AppMetrics is nil")
	}

	if len(analytics.AppMetrics) == 0 {
		t.Error("AppMetrics is empty")
	}
}

// TestGetRecentSessionsWithInvalidID tests error handling
func TestGetRecentSessionsWithInvalidID(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetRecentSessions(ctx, 0, 10)
	if err == nil {
		t.Error("Expected error for zero user ID, got nil")
	}
}

// TestGetRecentSessions tests recent sessions retrieval
func TestGetRecentSessions(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	sessions, err := repo.GetRecentSessions(ctx, 1, 10)
	if err != nil {
		t.Fatalf("Failed to get recent sessions: %v", err)
	}

	if sessions == nil {
		t.Fatal("Sessions is nil")
	}

	// Should be empty initially
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}

// TestGetRecentSessionsWithLimitValidation tests limit validation
func TestGetRecentSessionsWithLimitValidation(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Test with negative limit (should default to 20)
	sessions, err := repo.GetRecentSessions(ctx, 1, -5)
	if err != nil {
		t.Fatalf("Failed to get recent sessions: %v", err)
	}

	if sessions == nil {
		t.Fatal("Sessions is nil")
	}

	// Test with zero limit (should default to 20)
	sessions, err = repo.GetRecentSessions(ctx, 1, 0)
	if err != nil {
		t.Fatalf("Failed to get recent sessions: %v", err)
	}

	if sessions == nil {
		t.Fatal("Sessions is nil")
	}
}

// TestGetUnifiedLeaderboard tests leaderboard retrieval
func TestGetUnifiedLeaderboard(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Test valid categories
	validCategories := []string{"typing_wpm", "math_accuracy", "reading_comprehension", "piano_score", "overall"}

	for _, category := range validCategories {
		lb, err := repo.GetUnifiedLeaderboard(ctx, category, 10)
		if err != nil {
			t.Fatalf("Failed to get leaderboard for %s: %v", category, err)
		}

		if lb == nil {
			t.Fatalf("Leaderboard for %s is nil", category)
		}

		if lb.Category != category {
			t.Errorf("Expected category %s, got %s", category, lb.Category)
		}

		if lb.Entries == nil {
			t.Errorf("Entries for %s is nil", category)
		}
	}
}

// TestGetUnifiedLeaderboardWithInvalidCategory tests error handling
func TestGetUnifiedLeaderboardWithInvalidCategory(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetUnifiedLeaderboard(ctx, "invalid_category", 10)
	if err == nil {
		t.Error("Expected error for invalid category, got nil")
	}
}

// TestGetUnifiedLeaderboardWithLimitValidation tests limit validation
func TestGetUnifiedLeaderboardWithLimitValidation(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Test with zero limit (should default to 20)
	lb, err := repo.GetUnifiedLeaderboard(ctx, "typing_wpm", 0)
	if err != nil {
		t.Fatalf("Failed to get leaderboard: %v", err)
	}

	if lb == nil {
		t.Fatal("Leaderboard is nil")
	}

	// Test with negative limit (should default to 20)
	lb, err = repo.GetUnifiedLeaderboard(ctx, "typing_wpm", -10)
	if err != nil {
		t.Fatalf("Failed to get leaderboard: %v", err)
	}

	if lb == nil {
		t.Fatal("Leaderboard is nil")
	}
}

// TestGetSystemStats tests system statistics retrieval
func TestGetSystemStats(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create users table
	db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY)`)
	db.Exec(`INSERT INTO users (id) VALUES (1), (2), (3)`)

	// Create app tables
	db.Exec(`CREATE TABLE typing_tests (id INTEGER PRIMARY KEY, user_id INTEGER, timestamp TIMESTAMP)`)
	db.Exec(`CREATE TABLE math_results (id INTEGER PRIMARY KEY, user_id INTEGER, timestamp TIMESTAMP)`)
	db.Exec(`CREATE TABLE reading_sessions (id INTEGER PRIMARY KEY, user_id INTEGER, created_at TIMESTAMP)`)
	db.Exec(`CREATE TABLE piano_lessons (id INTEGER PRIMARY KEY, user_id INTEGER, created_at TIMESTAMP)`)

	repo := NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetSystemStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get system stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats is nil")
	}

	if stats.TotalUsers != 3 {
		t.Errorf("Expected 3 users, got %d", stats.TotalUsers)
	}

	if stats.AppUsageCount == nil {
		t.Fatal("AppUsageCount is nil")
	}
}
