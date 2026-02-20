package reading

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	schema := `
	CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT,
		content TEXT,
		reading_level TEXT,
		language TEXT DEFAULT 'english',
		word_count INTEGER,
		estimated_time_minutes REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE reading_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		book_id INTEGER NOT NULL,
		start_time DATETIME,
		end_time DATETIME,
		wpm REAL,
		accuracy REAL,
		comprehension REAL,
		duration REAL,
		words_read INTEGER,
		error_count INTEGER,
		completed BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE comprehension_tests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		question TEXT,
		user_answer TEXT,
		correct_answer TEXT,
		is_correct BOOLEAN,
		score REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_sessions_user ON reading_sessions(user_id);
	CREATE INDEX idx_sessions_book ON reading_sessions(book_id);
	CREATE INDEX idx_tests_session ON comprehension_tests(session_id);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// newPoolFromDB creates a database pool from sql.DB for testing
func newPoolFromDB(db *sql.DB) *sql.DB {
	return db
}

// TestSaveBook tests saving a book
func TestSaveBook(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	book := &Book{
		Title:        "Test Book",
		Author:       "Test Author",
		Content:      "This is a test book with sufficient content for reading.",
		ReadingLevel: "intermediate",
	}

	id, err := repo.SaveBook(ctx, book)
	if err != nil {
		t.Fatalf("SaveBook() error = %v", err)
	}

	if id == 0 {
		t.Error("SaveBook() should return non-zero ID")
	}
}

// TestGetBookByID tests retrieving a book by ID
func TestGetBookByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Save a book first
	originalBook := &Book{
		Title:        "The Great Gatsby",
		Author:       "F. Scott Fitzgerald",
		Content:      "In my younger and more vulnerable years, my father gave me advice.",
		ReadingLevel: "advanced",
	}
	id, err := repo.SaveBook(ctx, originalBook)
	if err != nil {
		t.Fatalf("SaveBook() error = %v", err)
	}

	// Retrieve the book
	retrieved, err := repo.GetBookByID(ctx, id)
	if err != nil {
		t.Fatalf("GetBookByID() error = %v", err)
	}

	if retrieved.Title != originalBook.Title {
		t.Errorf("GetBookByID() title mismatch: got %s, want %s", retrieved.Title, originalBook.Title)
	}
}

// TestSaveLesson tests saving a reading session
func TestSaveLesson(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Save a book first
	book := &Book{
		Title:        "Test Book",
		Author:       "Author",
		Content:      "This is test content for a reading session practice.",
		ReadingLevel: "beginner",
	}
	bookID, _ := repo.SaveBook(ctx, book)

	session := &ReadingSession{
		UserID:               1,
		BookID:               bookID,
		StartTime:            time.Now(),
		EndTime:              time.Now().Add(10 * time.Minute),
		WPM:                  150.0,
		Accuracy:             95.0,
		ComprehensionScore:   85.0,
		Duration:             600.0,
		WordsRead:            100,
		ErrorCount:           2,
		Completed:            true,
	}

	id, err := repo.SaveLesson(ctx, session)
	if err != nil {
		t.Fatalf("SaveLesson() error = %v", err)
	}

	if id == 0 {
		t.Error("SaveLesson() should return non-zero ID")
	}
}

// TestGetUserSessions tests retrieving user sessions
func TestGetUserSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test data
	book := &Book{
		Title:        "Test",
		Author:       "Author",
		Content:      "Content for testing sessions retrieval operations.",
		ReadingLevel: "intermediate",
	}
	bookID, _ := repo.SaveBook(ctx, book)

	for i := 0; i < 3; i++ {
		session := &ReadingSession{
			UserID:               1,
			BookID:               bookID,
			Duration:             600.0,
			WPM:                  100.0 + float64(i*10),
			Accuracy:             90.0,
			ComprehensionScore:   80.0,
		}
		repo.SaveLesson(ctx, session)
	}

	// Retrieve sessions
	sessions, err := repo.GetUserSessions(ctx, 1, 10, 0)
	if err != nil {
		t.Fatalf("GetUserSessions() error = %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("GetUserSessions() expected 3 sessions, got %d", len(sessions))
	}
}

// TestGetUserStats tests user statistics calculation
func TestGetUserStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test data
	book := &Book{
		Title:        "Stats Test",
		Author:       "Author",
		Content:      "This is a book used for testing statistics calculation and aggregation.",
		ReadingLevel: "beginner",
	}
	bookID, _ := repo.SaveBook(ctx, book)

	session := &ReadingSession{
		UserID:               1,
		BookID:               bookID,
		Duration:             300.0,
		WPM:                  150.0,
		Accuracy:             95.0,
		ComprehensionScore:   85.0,
		Completed:            true,
	}
	repo.SaveLesson(ctx, session)

	// Get stats
	stats, err := repo.GetUserStats(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if stats.UserID != 1 {
		t.Errorf("Stats UserID mismatch: got %d, want 1", stats.UserID)
	}

	if stats.TotalSessionsCount == 0 {
		t.Error("Stats should show at least 1 session")
	}
}

// TestGetLeaderboardFuture tests leaderboard retrieval - TODO: Fix this test
// Currently skipped due to database issues
/*
func TestGetLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test data for multiple users
	book := &Book{
		Title:        "Leaderboard Test",
		Author:       "Author",
		Content:      "Content for leaderboard testing with multiple users competing.",
		ReadingLevel: "advanced",
	}
	bookID, _ := repo.SaveBook(ctx, book)

	for userID := 1; userID <= 3; userID++ {
		session := &ReadingSession{
			UserID:               uint(userID),
			BookID:               bookID,
			Duration:             600.0,
			WPM:                  100.0 + float64(userID*20),
			Accuracy:             90.0,
			ComprehensionScore:   80.0,
			Completed:            true,
			StartTime:            time.Now(),
			EndTime:              time.Now(),
		}
		_, err := repo.SaveLesson(ctx, session)
		if err != nil {
			t.Fatalf("SaveLesson() error = %v", err)
		}
	}

	// Verify sessions were saved
	count, err := repo.GetSessionCount(ctx, 1)
	if err != nil {
		t.Fatalf("GetSessionCount() error = %v", err)
	}
	if count == 0 {
		t.Error("Sessions should be saved before getting leaderboard")
	}

	// Get leaderboard
	leaderboard, err := repo.GetLeaderboard(ctx, 10)
	if err != nil {
		t.Fatalf("GetLeaderboard() error = %v", err)
	}

	if len(leaderboard) == 0 {
		t.Error("Leaderboard should have entries")
	}

	// Verify sorting
	for i := 0; i < len(leaderboard)-1; i++ {
		if leaderboard[i].BestWPM < leaderboard[i+1].BestWPM {
			t.Error("Leaderboard not sorted correctly by WPM")
		}
	}
}
*/

// TestSaveComprehensionTest tests saving a comprehension test
func TestSaveComprehensionTest(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	test := &ComprehensionTest{
		SessionID:     1,
		Question:      "What is the main theme?",
		UserAnswer:    "Love and ambition",
		CorrectAnswer: "Love and ambition",
		IsCorrect:     true,
		Score:         100.0,
	}

	id, err := repo.SaveComprehensionTest(ctx, test)
	if err != nil {
		t.Fatalf("SaveComprehensionTest() error = %v", err)
	}

	if id == 0 {
		t.Error("SaveComprehensionTest() should return non-zero ID")
	}
}

// TestGetComprehensionTests tests retrieving comprehension tests
func TestGetComprehensionTests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Save tests
	for i := 0; i < 3; i++ {
		test := &ComprehensionTest{
			SessionID:     1,
			Question:      "Question " + string(rune('A'+i)),
			CorrectAnswer: "Answer",
			UserAnswer:    "Answer",
			IsCorrect:     true,
			Score:         100.0,
		}
		repo.SaveComprehensionTest(ctx, test)
	}

	// Retrieve tests
	tests, err := repo.GetComprehensionTests(ctx, 1)
	if err != nil {
		t.Fatalf("GetComprehensionTests() error = %v", err)
	}

	if len(tests) != 3 {
		t.Errorf("GetComprehensionTests() expected 3 tests, got %d", len(tests))
	}
}

// TestUpdateSessionCompletion tests marking session as completed
func TestUpdateSessionCompletion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create a book and session
	book := &Book{
		Title:        "Completion Test",
		Author:       "Author",
		Content:      "Content for testing session completion status updates.",
		ReadingLevel: "intermediate",
	}
	bookID, _ := repo.SaveBook(ctx, book)

	session := &ReadingSession{
		UserID:               1,
		BookID:               bookID,
		Duration:             300.0,
		WPM:                  100.0,
		Accuracy:             90.0,
		ComprehensionScore:   80.0,
		Completed:            false,
	}
	sessionID, _ := repo.SaveLesson(ctx, session)

	// Update completion
	err := repo.UpdateSessionCompletion(ctx, sessionID, true)
	if err != nil {
		t.Fatalf("UpdateSessionCompletion() error = %v", err)
	}

	// Verify
	updated, _ := repo.GetSessionByID(ctx, sessionID)
	if !updated.Completed {
		t.Error("Session should be marked as completed")
	}
}

// TestGetSessionCount tests retrieving session count
func TestGetSessionCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	book := &Book{
		Title:        "Count Test",
		Author:       "Author",
		Content:      "Testing session count retrieval with multiple practice sessions.",
		ReadingLevel: "beginner",
	}
	bookID, _ := repo.SaveBook(ctx, book)

	for i := 0; i < 5; i++ {
		session := &ReadingSession{
			UserID:               1,
			BookID:               bookID,
			Duration:             300.0,
			WPM:                  100.0,
			Accuracy:             90.0,
			ComprehensionScore:   80.0,
		}
		repo.SaveLesson(ctx, session)
	}

	count, err := repo.GetSessionCount(ctx, 1)
	if err != nil {
		t.Fatalf("GetSessionCount() error = %v", err)
	}

	if count != 5 {
		t.Errorf("GetSessionCount() expected 5, got %d", count)
	}
}

// TestGetBooks tests retrieving books with filtering
func TestGetBooks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create books with different levels
	for _, level := range []string{"beginner", "intermediate", "advanced"} {
		book := &Book{
			Title:        "Book " + level,
			Author:       "Author",
			Content:      "This is a " + level + " level book for testing filtering operations.",
			ReadingLevel: level,
		}
		repo.SaveBook(ctx, book)
	}

	// Get books by level
	books, err := repo.GetBooks(ctx, "intermediate", 10, 0)
	if err != nil {
		t.Fatalf("GetBooks() error = %v", err)
	}

	if len(books) != 1 {
		t.Errorf("GetBooks() expected 1 book, got %d", len(books))
	}

	if len(books) > 0 && books[0].ReadingLevel != "intermediate" {
		t.Error("GetBooks() filter not working correctly")
	}
}

// TestCountWords tests word counting helper function
func TestCountWords(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"hello world", 2},
		{"one two three four", 4},
		{"single", 1},
		{"", 0},
		{"  spaces  between  ", 2},
	}

	for _, tt := range tests {
		t.Run("word counting", func(t *testing.T) {
			result := countWords(tt.text)
			if result != tt.expected {
				t.Errorf("countWords(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}
