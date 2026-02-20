package typing

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jgirmay/unified-go/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Create tables
	createTablesSQL := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE typing_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		wpm REAL,
		raw_wpm REAL,
		accuracy REAL,
		errors INTEGER,
		time_taken REAL,
		test_mode TEXT,
		test_duration INTEGER,
		text_snippet TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE user_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE,
		total_tests INTEGER DEFAULT 0,
		average_wpm REAL DEFAULT 0,
		average_accuracy REAL DEFAULT 0,
		best_wpm INTEGER DEFAULT 0,
		total_time_typed INTEGER DEFAULT 0,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX idx_typing_results_user_id ON typing_results(user_id);
	`

	if _, err := db.Exec(createTablesSQL); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// Insert test user
	if _, err := db.Exec("INSERT INTO users (id, username) VALUES (1, 'testuser')"); err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	return db
}

// TestSaveResult tests saving a typing result
func TestSaveResult(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	result := &TypingResult{
		UserID:      1,
		Content:     "the quick brown fox",
		TimeSpent:   120.0,
		WPM:         60.5,
		RawWPM:      58.0,
		Accuracy:    95.5,
		ErrorsCount: 2,
		TestMode:    "paragraphs",
	}

	ctx := context.Background()
	id, err := repo.SaveResult(ctx, result)
	if err != nil {
		t.Fatalf("SaveResult() error = %v", err)
	}

	if id == 0 {
		t.Error("SaveResult() returned zero ID")
	}
}

// TestSaveResultInvalid tests saving an invalid result
func TestSaveResultInvalid(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	result := &TypingResult{
		UserID:    0, // Invalid: missing user_id
		Content:   "test",
		TimeSpent: 120.0,
		WPM:       60.5,
		Accuracy:  95.5,
	}

	ctx := context.Background()
	_, err := repo.SaveResult(ctx, result)
	if err == nil {
		t.Error("SaveResult() should return error for invalid result")
	}
}

// TestGetUserStats tests retrieving user statistics
func TestGetUserStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert stats
	stats := &UserStats{
		UserID:          1,
		TotalTests:      10,
		AverageWPM:      60.5,
		BestWPM:         75,
		AverageAccuracy: 95.5,
		TotalTimeTyped:  1200,
	}

	_, err := db.Exec(`
		INSERT INTO user_stats (
			user_id, total_tests, average_wpm, best_wpm,
			average_accuracy, total_time_typed, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, stats.UserID, stats.TotalTests, stats.AverageWPM, stats.BestWPM,
		stats.AverageAccuracy, stats.TotalTimeTyped, time.Now())

	if err != nil {
		t.Fatalf("failed to insert test stats: %v", err)
	}

	ctx := context.Background()
	retrieved, err := repo.GetUserStats(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if retrieved.UserID != 1 || retrieved.TotalTests != 10 {
		t.Errorf("GetUserStats() returned incorrect data: %+v", retrieved)
	}
}

// TestGetUserStatsNotFound tests retrieving stats for non-existent user
func TestGetUserStatsNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	ctx := context.Background()
	stats, err := repo.GetUserStats(ctx, 999)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if stats.UserID != 999 {
		t.Errorf("GetUserStats() should return empty stats for non-existent user")
	}
}

// TestGetLeaderboard tests retrieving leaderboard
func TestGetLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert multiple users and their stats
	for i := 2; i <= 5; i++ {
		// Insert user first
		username := fmt.Sprintf("user%d", i)
		_, err := db.Exec("INSERT INTO users (id, username) VALUES (?, ?)", i, username)
		if err != nil {
			t.Fatalf("failed to insert user: %v", err)
		}

		// Then insert stats
		_, err = db.Exec(`
			INSERT INTO user_stats (
				user_id, total_tests, average_wpm, best_wpm,
				average_accuracy, total_time_typed, last_updated
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, i, 10, float64(50+i*5), float64(60+i*5),
			float64(90+i), 1000, time.Now())

		if err != nil {
			t.Fatalf("failed to insert user stats: %v", err)
		}
	}

	// Also insert stats for user 1 (already exists)
	_, err := db.Exec(`
		INSERT INTO user_stats (
			user_id, total_tests, average_wpm, best_wpm,
			average_accuracy, total_time_typed, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, 1, 10, 50.0, 60.0, 90.0, 1000, time.Now())
	if err != nil {
		t.Fatalf("failed to insert user 1 stats: %v", err)
	}

	ctx := context.Background()
	leaderboard, err := repo.GetLeaderboard(ctx, 10)
	if err != nil {
		t.Fatalf("GetLeaderboard() error = %v", err)
	}

	if len(leaderboard) != 5 {
		t.Errorf("GetLeaderboard() returned %d entries, want 5", len(leaderboard))
	}

	// Verify ordering (best_wpm DESC)
	for i := 0; i < len(leaderboard)-1; i++ {
		if leaderboard[i].BestWPM < leaderboard[i+1].BestWPM {
			t.Error("GetLeaderboard() not properly ordered")
		}
	}
}

// TestGetLeaderboardWithLimit tests leaderboard limit
func TestGetLeaderboardWithLimit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert 20 users and their stats
	for i := 2; i <= 20; i++ {
		// Insert user first
		username := fmt.Sprintf("user%d", i)
		_, err := db.Exec("INSERT INTO users (id, username) VALUES (?, ?)", i, username)
		if err != nil {
			t.Fatalf("failed to insert user: %v", err)
		}

		// Insert stats
		_, err = db.Exec(`
			INSERT INTO user_stats (
				user_id, total_tests, average_wpm, best_wpm,
				average_accuracy, total_time_typed, last_updated
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, i, 10, float64(50+i), float64(60+i),
			float64(90), 1000, time.Now())

		if err != nil {
			t.Fatalf("failed to insert user stats: %v", err)
		}
	}

	// Also insert stats for user 1
	_, err := db.Exec(`
		INSERT INTO user_stats (
			user_id, total_tests, average_wpm, best_wpm,
			average_accuracy, total_time_typed, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, 1, 10, 50.0, 60.0, 90.0, 1000, time.Now())
	if err != nil {
		t.Fatalf("failed to insert user 1 stats: %v", err)
	}

	ctx := context.Background()
	leaderboard, err := repo.GetLeaderboard(ctx, 5)
	if err != nil {
		t.Fatalf("GetLeaderboard() error = %v", err)
	}

	if len(leaderboard) != 5 {
		t.Errorf("GetLeaderboard(5) returned %d entries, want 5", len(leaderboard))
	}
}

// TestGetUserTests tests retrieving user test history
func TestGetUserTests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert test results
	for i := 1; i <= 5; i++ {
		_, err := db.Exec(`
			INSERT INTO typing_results (
				user_id, wpm, raw_wpm, accuracy, errors,
				time_taken, test_mode, text_snippet, timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, 1, float64(50+i*5), float64(48+i*5), float64(90+i),
			i, 120.0, "paragraphs", "test snippet", time.Now())

		if err != nil {
			t.Fatalf("failed to insert test result: %v", err)
		}
	}

	ctx := context.Background()
	tests, err := repo.GetUserTests(ctx, 1, 10, 0)
	if err != nil {
		t.Fatalf("GetUserTests() error = %v", err)
	}

	if len(tests) != 5 {
		t.Errorf("GetUserTests() returned %d tests, want 5", len(tests))
	}
}

// TestGetUserTestsPagination tests pagination in test history
func TestGetUserTestsPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert 10 test results
	for i := 1; i <= 10; i++ {
		_, err := db.Exec(`
			INSERT INTO typing_results (
				user_id, wpm, raw_wpm, accuracy, errors,
				time_taken, test_mode, text_snippet, timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, 1, float64(50+i), float64(48+i), float64(90),
			1, 120.0, "paragraphs", "snippet", time.Now())

		if err != nil {
			t.Fatalf("failed to insert test result: %v", err)
		}
	}

	ctx := context.Background()

	// Test first page
	tests1, err := repo.GetUserTests(ctx, 1, 3, 0)
	if err != nil {
		t.Fatalf("GetUserTests() error = %v", err)
	}

	if len(tests1) != 3 {
		t.Errorf("GetUserTests(limit=3) returned %d tests, want 3", len(tests1))
	}

	// Test second page
	tests2, err := repo.GetUserTests(ctx, 1, 3, 3)
	if err != nil {
		t.Fatalf("GetUserTests() error = %v", err)
	}

	if len(tests2) != 3 {
		t.Errorf("GetUserTests(offset=3) returned %d tests, want 3", len(tests2))
	}
}

// TestGetTestHistory tests retrieving test history by date range
func TestGetTestHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert recent test
	_, err := db.Exec(`
		INSERT INTO typing_results (
			user_id, wpm, raw_wpm, accuracy, errors,
			time_taken, test_mode, text_snippet, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 1, 60.0, 58.0, 95.0, 2, 120.0, "paragraphs", "snippet", time.Now())

	if err != nil {
		t.Fatalf("failed to insert test result: %v", err)
	}

	ctx := context.Background()
	history, err := repo.GetTestHistory(ctx, 1, 30)
	if err != nil {
		t.Fatalf("GetTestHistory() error = %v", err)
	}

	if len(history) != 1 {
		t.Errorf("GetTestHistory() returned %d results, want 1", len(history))
	}
}

// TestGetTestCount tests retrieving test count for user
func TestGetTestCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert 5 tests
	for i := 1; i <= 5; i++ {
		_, err := db.Exec(`
			INSERT INTO typing_results (
				user_id, wpm, raw_wpm, accuracy, errors,
				time_taken, test_mode, text_snippet, timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, 1, 60.0, 58.0, 95.0, 2, 120.0, "paragraphs", "snippet", time.Now())

		if err != nil {
			t.Fatalf("failed to insert test result: %v", err)
		}
	}

	ctx := context.Background()
	count, err := repo.GetTestCount(ctx, 1)
	if err != nil {
		t.Fatalf("GetTestCount() error = %v", err)
	}

	if count != 5 {
		t.Errorf("GetTestCount() returned %d, want 5", count)
	}
}

// TestDeleteUserTests tests deleting all tests for a user
func TestDeleteUserTests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	pool := newPoolFromDB(db)
	repo := NewRepository(pool)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO typing_results (
			user_id, wpm, raw_wpm, accuracy, errors,
			time_taken, test_mode, text_snippet, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 1, 60.0, 58.0, 95.0, 2, 120.0, "paragraphs", "snippet", time.Now())

	if err != nil {
		t.Fatalf("failed to insert test result: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO user_stats (
			user_id, total_tests, average_wpm, best_wpm,
			average_accuracy, total_time_typed, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, 1, 5, 60.0, 70, 95.0, 600, time.Now())

	if err != nil {
		t.Fatalf("failed to insert user stats: %v", err)
	}

	ctx := context.Background()

	// Delete user tests
	err = repo.DeleteUserTests(ctx, 1)
	if err != nil {
		t.Fatalf("DeleteUserTests() error = %v", err)
	}

	// Verify deletion
	count, err := repo.GetTestCount(ctx, 1)
	if err != nil {
		t.Fatalf("GetTestCount() error = %v", err)
	}

	if count != 0 {
		t.Errorf("DeleteUserTests() did not delete tests, count = %d", count)
	}
}

// newPoolFromDB creates a Pool from a raw *sql.DB for testing
func newPoolFromDB(db *sql.DB) *database.Pool {
	return &database.Pool{DB: db}
}
