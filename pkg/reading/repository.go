package reading

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository handles database operations for reading app
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new reading repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveLesson saves a reading session to the database
func (r *Repository) SaveLesson(ctx context.Context, session *ReadingSession) (uint, error) {
	if session == nil {
		return 0, errors.New("session cannot be nil")
	}

	if err := session.Validate(); err != nil {
		return 0, fmt.Errorf("invalid session: %w", err)
	}

	stmt := `INSERT INTO reading_sessions (user_id, book_id, start_time, end_time, wpm, accuracy, comprehension, duration, words_read, error_count, completed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.ExecContext(ctx, stmt,
		session.UserID, session.BookID, session.StartTime, session.EndTime,
		session.WPM, session.Accuracy, session.ComprehensionScore, session.Duration,
		session.WordsRead, session.ErrorCount, session.Completed)

	if err != nil {
		return 0, fmt.Errorf("failed to save lesson: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return uint(id), nil
}

// SaveBook saves a book to the database
func (r *Repository) SaveBook(ctx context.Context, book *Book) (uint, error) {
	if book == nil {
		return 0, errors.New("book cannot be nil")
	}

	if err := book.Validate(); err != nil {
		return 0, fmt.Errorf("invalid book: %w", err)
	}

	if book.WordCount == 0 && len(book.Content) > 0 {
		book.WordCount = countWords(book.Content)
	}

	stmt := `INSERT INTO books (title, author, content, reading_level, language, word_count, estimated_time_minutes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.ExecContext(ctx, stmt,
		book.Title, book.Author, book.Content, book.ReadingLevel, book.Language,
		book.WordCount, book.EstimatedTime)

	if err != nil {
		return 0, fmt.Errorf("failed to save book: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return uint(id), nil
}

// GetBookByID retrieves a book by ID
func (r *Repository) GetBookByID(ctx context.Context, bookID uint) (*Book, error) {
	if bookID == 0 {
		return nil, errors.New("book_id is required")
	}

	var book Book
	stmt := `SELECT id, title, author, content, reading_level, language, word_count, estimated_time_minutes, created_at, updated_at FROM books WHERE id = ?`

	err := r.db.QueryRowContext(ctx, stmt, bookID).Scan(
		&book.ID, &book.Title, &book.Author, &book.Content, &book.ReadingLevel,
		&book.Language, &book.WordCount, &book.EstimatedTime, &book.CreatedAt, &book.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("book not found")
		}
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	return &book, nil
}

// GetBooks retrieves books with optional filtering
func (r *Repository) GetBooks(ctx context.Context, difficulty string, limit, offset int) ([]Book, error) {
	if limit <= 0 || limit > 1000 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	stmt := `SELECT id, title, author, content, reading_level, language, word_count, estimated_time_minutes, created_at, updated_at FROM books`

	if difficulty != "" {
		stmt += ` WHERE reading_level = ?`
	}

	stmt += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var rows *sql.Rows
	var err error

	if difficulty != "" {
		rows, err = r.db.QueryContext(ctx, stmt, difficulty, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, stmt, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Content, &book.ReadingLevel,
			&book.Language, &book.WordCount, &book.EstimatedTime, &book.CreatedAt, &book.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}
		books = append(books, book)
	}

	return books, rows.Err()
}

// SaveComprehensionTest saves a comprehension test result
func (r *Repository) SaveComprehensionTest(ctx context.Context, test *ComprehensionTest) (uint, error) {
	if test == nil {
		return 0, errors.New("test cannot be nil")
	}

	if err := test.Validate(); err != nil {
		return 0, fmt.Errorf("invalid test: %w", err)
	}

	stmt := `INSERT INTO comprehension_tests (session_id, question, user_answer, correct_answer, is_correct, score, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	result, err := r.db.ExecContext(ctx, stmt,
		test.SessionID, test.Question, test.UserAnswer, test.CorrectAnswer, test.IsCorrect, test.Score)

	if err != nil {
		return 0, fmt.Errorf("failed to save comprehension test: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return uint(id), nil
}

// GetUserSessions retrieves all reading sessions for a user
func (r *Repository) GetUserSessions(ctx context.Context, userID uint, limit, offset int) ([]ReadingSession, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if limit <= 0 || limit > 1000 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	stmt := `SELECT id, user_id, book_id, start_time, end_time, wpm, accuracy, comprehension, duration, words_read, error_count, completed, created_at, updated_at
		FROM reading_sessions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, stmt, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []ReadingSession
	for rows.Next() {
		var session ReadingSession
		if err := rows.Scan(&session.ID, &session.UserID, &session.BookID, &session.StartTime, &session.EndTime,
			&session.WPM, &session.Accuracy, &session.ComprehensionScore, &session.Duration,
			&session.WordsRead, &session.ErrorCount, &session.Completed, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// GetUserStats calculates aggregated statistics for a user
func (r *Repository) GetUserStats(ctx context.Context, userID uint) (*ReadingStats, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	stats := &ReadingStats{
		UserID: userID,
	}

	// Count total books read
	var bookCount int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT book_id) FROM reading_sessions WHERE user_id = ?`, userID).Scan(&bookCount)
	stats.TotalBooksRead = int(bookCount)

	// Get completed sessions
	stmt := `SELECT id, user_id, book_id, start_time, end_time, wpm, accuracy, comprehension, duration, words_read, error_count, completed, created_at, updated_at
		FROM reading_sessions WHERE user_id = ? AND completed = 1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	defer rows.Close()

	var sessions []ReadingSession
	var totalWPM, totalAccuracy, totalComprehension, totalTime float64
	var maxWPM float64

	for rows.Next() {
		var session ReadingSession
		if err := rows.Scan(&session.ID, &session.UserID, &session.BookID, &session.StartTime, &session.EndTime,
			&session.WPM, &session.Accuracy, &session.ComprehensionScore, &session.Duration,
			&session.WordsRead, &session.ErrorCount, &session.Completed, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
		totalWPM += session.WPM
		totalAccuracy += session.Accuracy
		totalComprehension += session.ComprehensionScore
		totalTime += session.Duration

		if session.WPM > maxWPM {
			maxWPM = session.WPM
		}
	}

	if len(sessions) == 0 {
		return stats, nil
	}

	count := float64(len(sessions))
	stats.TotalSessionsCount = len(sessions)
	stats.AverageWPM = totalWPM / count
	stats.BestWPM = maxWPM
	stats.AverageAccuracy = totalAccuracy / count
	stats.AverageComprehension = totalComprehension / count
	stats.TotalReadingTime = totalTime

	// Get last session time
	lastTime := sessions[0].CreatedAt
	stats.LastSessionTime = &lastTime

	// Find favorite reading level
	var favoriteLevel *string
	r.db.QueryRowContext(ctx, `SELECT b.reading_level FROM reading_sessions rs
		JOIN books b ON rs.book_id = b.id
		WHERE rs.user_id = ?
		GROUP BY b.reading_level
		ORDER BY COUNT(*) DESC LIMIT 1`, userID).Scan(&favoriteLevel)
	if favoriteLevel != nil {
		stats.FavoriteReadingLevel = *favoriteLevel
	}

	return stats, nil
}

// GetLeaderboard retrieves top readers by best WPM
func (r *Repository) GetLeaderboard(ctx context.Context, limit int) ([]ReadingStats, error) {
	if limit <= 0 || limit > 1000 {
		limit = 10
	}

	// Get all sessions to find unique users
	stmt := `SELECT id, user_id, book_id, start_time, end_time, wpm, accuracy, comprehension, duration, words_read, error_count, completed, created_at, updated_at FROM reading_sessions ORDER BY wpm DESC`

	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for leaderboard: %w", err)
	}
	defer rows.Close()

	seenUsers := make(map[uint]bool)
	var leaderboard []ReadingStats

	for rows.Next() {
		var session ReadingSession
		if err := rows.Scan(&session.ID, &session.UserID, &session.BookID, &session.StartTime, &session.EndTime,
			&session.WPM, &session.Accuracy, &session.ComprehensionScore, &session.Duration,
			&session.WordsRead, &session.ErrorCount, &session.Completed, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		if !seenUsers[session.UserID] {
			// Get full stats for this user
			stats, err := r.GetUserStats(ctx, session.UserID)
			if err == nil && stats != nil {
				leaderboard = append(leaderboard, *stats)
				seenUsers[session.UserID] = true

				if len(leaderboard) >= limit {
					break
				}
			}
		}
	}

	return leaderboard, rows.Err()
}

// GetSessionCount returns the number of sessions for a user
func (r *Repository) GetSessionCount(ctx context.Context, userID uint) (int, error) {
	if userID == 0 {
		return 0, errors.New("user_id is required")
	}

	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reading_sessions WHERE user_id = ?`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get session count: %w", err)
	}

	return count, nil
}

// GetComprehensionTests retrieves comprehension tests for a session
func (r *Repository) GetComprehensionTests(ctx context.Context, sessionID uint) ([]ComprehensionTest, error) {
	if sessionID == 0 {
		return nil, errors.New("session_id is required")
	}

	stmt := `SELECT id, session_id, question, user_answer, correct_answer, is_correct, score, created_at
		FROM comprehension_tests WHERE session_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, stmt, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comprehension tests: %w", err)
	}
	defer rows.Close()

	var tests []ComprehensionTest
	for rows.Next() {
		var test ComprehensionTest
		if err := rows.Scan(&test.ID, &test.SessionID, &test.Question, &test.UserAnswer, &test.CorrectAnswer,
			&test.IsCorrect, &test.Score, &test.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan test: %w", err)
		}
		tests = append(tests, test)
	}

	return tests, rows.Err()
}

// UpdateSessionCompletion marks a session as completed
func (r *Repository) UpdateSessionCompletion(ctx context.Context, sessionID uint, completed bool) error {
	if sessionID == 0 {
		return errors.New("session_id is required")
	}

	stmt := `UPDATE reading_sessions SET completed = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, stmt, completed, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// GetSessionByID retrieves a specific reading session
func (r *Repository) GetSessionByID(ctx context.Context, sessionID uint) (*ReadingSession, error) {
	if sessionID == 0 {
		return nil, errors.New("session_id is required")
	}

	stmt := `SELECT id, user_id, book_id, start_time, end_time, wpm, accuracy, comprehension, duration, words_read, error_count, completed, created_at, updated_at FROM reading_sessions WHERE id = ?`

	var session ReadingSession
	err := r.db.QueryRowContext(ctx, stmt, sessionID).Scan(
		&session.ID, &session.UserID, &session.BookID, &session.StartTime, &session.EndTime,
		&session.WPM, &session.Accuracy, &session.ComprehensionScore, &session.Duration,
		&session.WordsRead, &session.ErrorCount, &session.Completed, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// Helper function to count words in text
func countWords(text string) int {
	if text == "" {
		return 0
	}
	wordCount := 0
	inWord := false
	for _, char := range text {
		if char == ' ' || char == '\n' || char == '\t' || char == '\r' {
			inWord = false
		} else if !inWord {
			wordCount++
			inWord = true
		}
	}
	return wordCount
}
