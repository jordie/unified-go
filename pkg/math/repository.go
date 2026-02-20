package math

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository handles data access for math operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new math repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveSolution saves a problem solution to the database
func (r *Repository) SaveSolution(ctx context.Context, solution *ProblemSolution) (uint, error) {
	if err := solution.Validate(); err != nil {
		return 0, fmt.Errorf("invalid solution: %w", err)
	}

	query := `
		INSERT INTO math_solutions (
			user_id, problem_id, attempt, correct, time_spent, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		solution.UserID,
		solution.ProblemID,
		solution.Attempt,
		solution.Correct,
		solution.TimeSpent,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save solution: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint(id), nil
}

// SaveSession saves a completed quiz session
func (r *Repository) SaveSession(ctx context.Context, session *QuizSession) (uint, error) {
	if err := session.Validate(); err != nil {
		return 0, fmt.Errorf("invalid session: %w", err)
	}

	query := `
		INSERT INTO math_sessions (
			user_id, problem_type, difficulty, total_problems, correct_answers,
			score, time_spent, started_at, completed_at, average_time_per_problem
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		session.UserID,
		session.ProblemType,
		session.Difficulty,
		session.TotalProblems,
		session.CorrectAnswers,
		session.Score,
		session.TimeSpent,
		session.StartedAt,
		session.CompletedAt,
		session.AverageTimePerProblem,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save session: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint(id), nil
}

// GetUserStats retrieves aggregated statistics for a user
func (r *Repository) GetUserStats(ctx context.Context, userID uint) (*UserMathStats, error) {
	query := `
		SELECT
			user_id,
			total_problems,
			correct_answers,
			accuracy,
			average_time_per_problem,
			best_score,
			total_time_spent,
			sessions_completed,
			last_updated
		FROM math_user_stats
		WHERE user_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, userID)
	stats := &UserMathStats{}

	err := row.Scan(
		&stats.UserID,
		&stats.TotalProblems,
		&stats.CorrectAnswers,
		&stats.Accuracy,
		&stats.AverageTimePerProblem,
		&stats.BestScore,
		&stats.TotalTimeSpent,
		&stats.SessionsCompleted,
		&stats.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return &UserMathStats{UserID: userID}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// GetProblemTypeStats retrieves stats for a specific problem type
func (r *Repository) GetProblemTypeStats(ctx context.Context, userID uint, problemType ProblemType) (*MathResult, error) {
	query := `
		SELECT
			CAST(? AS TEXT) as problem_type,
			CAST('' AS TEXT) as difficulty,
			COUNT(*) as total_attempts,
			SUM(CASE WHEN correct = 1 THEN 1 ELSE 0 END) as correct_answers,
			(SUM(CASE WHEN correct = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as accuracy,
			AVG(time_spent) as average_time_per_problem
		FROM math_solutions
		WHERE user_id = ? AND problem_id IN (
			SELECT id FROM math_problems WHERE type = ?
		)
	`

	row := r.db.QueryRowContext(ctx, query, problemType, userID, problemType)
	result := &MathResult{}

	var difficulty string
	err := row.Scan(
		&result.ProblemType,
		&difficulty,
		&result.TotalAttempts,
		&result.CorrectAnswers,
		&result.Accuracy,
		&result.AverageTimePerProblem,
	)

	if err == sql.ErrNoRows {
		return &MathResult{ProblemType: problemType}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get problem type stats: %w", err)
	}

	return result, nil
}

// GetLeaderboard retrieves top users by accuracy
func (r *Repository) GetLeaderboard(ctx context.Context, limit int) ([]UserMathStats, error) {
	if limit <= 0 || limit > 1000 {
		limit = 10
	}

	query := `
		SELECT
			user_id,
			total_problems,
			correct_answers,
			accuracy,
			average_time_per_problem,
			best_score,
			total_time_spent,
			sessions_completed,
			last_updated
		FROM math_user_stats
		WHERE sessions_completed > 0
		ORDER BY accuracy DESC, best_score DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var stats []UserMathStats
	for rows.Next() {
		var s UserMathStats
		if err := rows.Scan(
			&s.UserID,
			&s.TotalProblems,
			&s.CorrectAnswers,
			&s.Accuracy,
			&s.AverageTimePerProblem,
			&s.BestScore,
			&s.TotalTimeSpent,
			&s.SessionsCompleted,
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

// GetUserSessions retrieves paginated session history for a user
func (r *Repository) GetUserSessions(ctx context.Context, userID uint, limit, offset int) ([]QuizSession, error) {
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
			problem_type,
			difficulty,
			total_problems,
			correct_answers,
			score,
			time_spent,
			started_at,
			completed_at,
			average_time_per_problem
		FROM math_sessions
		WHERE user_id = ?
		ORDER BY completed_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []QuizSession
	for rows.Next() {
		var s QuizSession
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.ProblemType,
			&s.Difficulty,
			&s.TotalProblems,
			&s.CorrectAnswers,
			&s.Score,
			&s.TimeSpent,
			&s.StartedAt,
			&s.CompletedAt,
			&s.AverageTimePerProblem,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user sessions query error: %w", err)
	}

	return sessions, nil
}

// updateUserStats updates aggregated user statistics
func (r *Repository) updateUserStats(ctx context.Context, userID uint) error {
	// Check if stats record exists
	var count int
	checkQuery := "SELECT COUNT(*) FROM math_user_stats WHERE user_id = ?"
	if err := r.db.QueryRowContext(ctx, checkQuery, userID).Scan(&count); err != nil {
		return fmt.Errorf("failed to check user stats: %w", err)
	}

	// Calculate aggregate statistics from sessions
	statsQuery := `
		SELECT
			COUNT(*) as total_sessions,
			SUM(total_problems) as total_problems,
			SUM(correct_answers) as correct_answers,
			MAX(score) as best_score,
			SUM(time_spent) as total_time_spent,
			AVG(average_time_per_problem) as avg_time_per_problem
		FROM math_sessions
		WHERE user_id = ? AND completed_at IS NOT NULL
	`

	var totalSessions int
	var totalProblems sql.NullInt64
	var correctAnswers sql.NullInt64
	var bestScore sql.NullFloat64
	var totalTimeSpent sql.NullFloat64
	var avgTimePerProblem sql.NullFloat64

	if err := r.db.QueryRowContext(ctx, statsQuery, userID).Scan(
		&totalSessions,
		&totalProblems,
		&correctAnswers,
		&bestScore,
		&totalTimeSpent,
		&avgTimePerProblem,
	); err != nil {
		return fmt.Errorf("failed to calculate stats: %w", err)
	}

	// Calculate accuracy
	var accuracy float64
	if totalProblems.Valid && totalProblems.Int64 > 0 && correctAnswers.Valid {
		accuracy = (float64(correctAnswers.Int64) / float64(totalProblems.Int64)) * 100.0
	}

	// Insert or update
	if count == 0 {
		insertQuery := `
			INSERT INTO math_user_stats (
				user_id, total_problems, correct_answers, accuracy,
				average_time_per_problem, best_score, total_time_spent,
				sessions_completed, last_updated
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err := r.db.ExecContext(
			ctx,
			insertQuery,
			userID,
			totalProblems.Int64,
			correctAnswers.Int64,
			accuracy,
			avgTimePerProblem.Float64,
			bestScore.Float64,
			int(totalTimeSpent.Float64),
			totalSessions,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert user stats: %w", err)
		}
	} else {
		updateQuery := `
			UPDATE math_user_stats
			SET total_problems = ?,
				correct_answers = ?,
				accuracy = ?,
				average_time_per_problem = ?,
				best_score = ?,
				total_time_spent = ?,
				sessions_completed = ?,
				last_updated = ?
			WHERE user_id = ?
		`

		_, err := r.db.ExecContext(
			ctx,
			updateQuery,
			totalProblems.Int64,
			correctAnswers.Int64,
			accuracy,
			avgTimePerProblem.Float64,
			bestScore.Float64,
			int(totalTimeSpent.Float64),
			totalSessions,
			time.Now(),
			userID,
		)
		if err != nil {
			return fmt.Errorf("failed to update user stats: %w", err)
		}
	}

	return nil
}
