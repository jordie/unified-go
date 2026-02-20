package typing

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository handles data access for typing operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new typing repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveResult saves a typing test result to the database
func (r *Repository) SaveResult(ctx context.Context, result *TypingResult) (uint, error) {
	if err := result.Validate(); err != nil {
		return 0, fmt.Errorf("invalid result: %w", err)
	}

	query := `
		INSERT INTO typing_results (
			user_id, wpm, raw_wpm, accuracy, errors, time_taken,
			test_mode, text_snippet, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		result.UserID,
		result.WPM,
		result.RawWPM,
		result.Accuracy,
		result.ErrorsCount,
		result.TimeSpent,
		result.TestMode,
		result.Content,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save result: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Update user stats
	if err := r.updateUserStats(ctx, result.UserID); err != nil {
		return uint(id), fmt.Errorf("failed to update stats: %w", err)
	}

	return uint(id), nil
}

// GetUserStats retrieves aggregated statistics for a user
func (r *Repository) GetUserStats(ctx context.Context, userID uint) (*UserStats, error) {
	query := `
		SELECT
			user_id,
			total_tests,
			average_wpm,
			best_wpm,
			average_accuracy,
			total_time_typed,
			last_updated
		FROM user_stats
		WHERE user_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, userID)
	stats := &UserStats{}

	err := row.Scan(
		&stats.UserID,
		&stats.TotalTests,
		&stats.AverageWPM,
		&stats.BestWPM,
		&stats.AverageAccuracy,
		&stats.TotalTimeTyped,
		&stats.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return &UserStats{UserID: userID}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// GetLeaderboard retrieves top users by WPM
func (r *Repository) GetLeaderboard(ctx context.Context, limit int) ([]UserStats, error) {
	if limit <= 0 || limit > 1000 {
		limit = 10
	}

	query := `
		SELECT
			user_id,
			total_tests,
			average_wpm,
			best_wpm,
			average_accuracy,
			total_time_typed,
			last_updated
		FROM user_stats
		WHERE total_tests > 0
		ORDER BY best_wpm DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var stats []UserStats
	for rows.Next() {
		var s UserStats
		if err := rows.Scan(
			&s.UserID,
			&s.TotalTests,
			&s.AverageWPM,
			&s.BestWPM,
			&s.AverageAccuracy,
			&s.TotalTimeTyped,
			&s.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %w", err)
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("leaderboard query error: %w", err)
	}

	return stats, nil
}

// GetUserTests retrieves paginated test history for a user
func (r *Repository) GetUserTests(ctx context.Context, userID uint, limit, offset int) ([]TypingTest, error) {
	if limit <= 0 || limit > 1000 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id,
			user_id,
			wpm,
			raw_wpm,
			accuracy,
			errors,
			time_taken,
			test_mode,
			text_snippet,
			timestamp
		FROM typing_results
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query user tests: %w", err)
	}
	defer rows.Close()

	var tests []TypingTest
	for rows.Next() {
		var t TypingTest
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.WPM,
			&t.RawWPM,
			&t.Accuracy,
			&t.Errors,
			&t.Duration,
			&t.TestMode,
			&t.TextSnippet,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan test row: %w", err)
		}
		tests = append(tests, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user tests query error: %w", err)
	}

	return tests, nil
}

// GetTestHistory retrieves tests from the last N days for a user
func (r *Repository) GetTestHistory(ctx context.Context, userID uint, days int) ([]TypingResult, error) {
	if days <= 0 {
		days = 30
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT
			id,
			user_id,
			text_snippet,
			time_taken,
			wpm,
			raw_wpm,
			errors,
			accuracy,
			test_mode,
			timestamp
		FROM typing_results
		WHERE user_id = ? AND timestamp >= ?
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query test history: %w", err)
	}
	defer rows.Close()

	var results []TypingResult
	for rows.Next() {
		var r TypingResult
		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.Content,
			&r.TimeSpent,
			&r.WPM,
			&r.RawWPM,
			&r.ErrorsCount,
			&r.Accuracy,
			&r.TestMode,
			&r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("test history query error: %w", err)
	}

	return results, nil
}

// updateUserStats updates aggregated user statistics
func (r *Repository) updateUserStats(ctx context.Context, userID uint) error {
	// First, check if user stats record exists
	var count int
	checkQuery := "SELECT COUNT(*) FROM user_stats WHERE user_id = ?"
	if err := r.db.QueryRowContext(ctx, checkQuery, userID).Scan(&count); err != nil {
		return fmt.Errorf("failed to check user stats: %w", err)
	}

	// Calculate aggregate statistics
	statsQuery := `
		SELECT
			COUNT(*) as total_tests,
			AVG(wpm) as average_wpm,
			MAX(wpm) as best_wpm,
			AVG(accuracy) as average_accuracy,
			SUM(time_taken) as total_time_typed
		FROM typing_results
		WHERE user_id = ?
	`

	var totalTests int
	var averageWPM, bestWPM, averageAccuracy sql.NullFloat64
	var totalTimeTyped sql.NullFloat64

	if err := r.db.QueryRowContext(ctx, statsQuery, userID).Scan(
		&totalTests,
		&averageWPM,
		&bestWPM,
		&averageAccuracy,
		&totalTimeTyped,
	); err != nil {
		return fmt.Errorf("failed to calculate stats: %w", err)
	}

	// Insert or update
	if count == 0 {
		insertQuery := `
			INSERT INTO user_stats (
				user_id, total_tests, average_wpm, best_wpm,
				average_accuracy, total_time_typed, last_updated
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`

		_, err := r.db.ExecContext(
			ctx,
			insertQuery,
			userID,
			totalTests,
			averageWPM.Float64,
			bestWPM.Float64,
			averageAccuracy.Float64,
			int(totalTimeTyped.Float64),
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert user stats: %w", err)
		}
	} else {
		updateQuery := `
			UPDATE user_stats
			SET total_tests = ?,
				average_wpm = ?,
				best_wpm = ?,
				average_accuracy = ?,
				total_time_typed = ?,
				last_updated = ?
			WHERE user_id = ?
		`

		_, err := r.db.ExecContext(
			ctx,
			updateQuery,
			totalTests,
			averageWPM.Float64,
			bestWPM.Float64,
			averageAccuracy.Float64,
			int(totalTimeTyped.Float64),
			time.Now(),
			userID,
		)
		if err != nil {
			return fmt.Errorf("failed to update user stats: %w", err)
		}
	}

	return nil
}

// GetTestCount returns the total number of tests for a user
func (r *Repository) GetTestCount(ctx context.Context, userID uint) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM typing_results WHERE user_id = ?"
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get test count: %w", err)
	}
	return count, nil
}

// DeleteUserTests deletes all tests for a user (admin operation)
func (r *Repository) DeleteUserTests(ctx context.Context, userID uint) error {
	query := "DELETE FROM typing_results WHERE user_id = ?"
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user tests: %w", err)
	}

	// Also delete user stats
	statsQuery := "DELETE FROM user_stats WHERE user_id = ?"
	_, err = r.db.ExecContext(ctx, statsQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user stats: %w", err)
	}

	return nil
}
