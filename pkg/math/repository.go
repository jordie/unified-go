package math

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository handles all database operations for the math app
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// === USER OPERATIONS ===

// SaveUser saves or updates a user
func (r *Repository) SaveUser(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	user.LastActive = time.Now()

	query := `
		INSERT INTO users (username, created_at, last_active)
		VALUES (?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET last_active = excluded.last_active
	`

	result, err := r.db.ExecContext(ctx, query, user.Username, time.Now(), user.LastActive)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		user.ID = uint(id)
	}

	return nil
}

// GetUser retrieves a user by ID
func (r *Repository) GetUser(ctx context.Context, userID uint) (*User, error) {
	user := &User{}
	query := `SELECT id, username, created_at, last_active FROM users WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.CreatedAt, &user.LastActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, created_at, last_active FROM users WHERE username = ?`

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.CreatedAt, &user.LastActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetAllUsers retrieves all users with pagination
func (r *Repository) GetAllUsers(ctx context.Context, limit int, offset int) ([]*User, error) {
	query := `SELECT id, username, created_at, last_active FROM users ORDER BY id LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// === MATH RESULT OPERATIONS ===

// SaveResult saves a practice session result
func (r *Repository) SaveResult(ctx context.Context, result *MathResult) error {
	if err := result.Validate(); err != nil {
		return fmt.Errorf("invalid result: %w", err)
	}

	result.CalculateAccuracy()
	result.CalculateAverageTime()
	result.Timestamp = time.Now()

	query := `
		INSERT INTO results (user_id, mode, difficulty, total_questions, correct_answers, total_time, average_time, accuracy, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(ctx, query,
		result.UserID, result.Mode, result.Difficulty, result.TotalQuestions,
		result.CorrectAnswers, result.TotalTime, result.AverageTime, result.Accuracy, result.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	id, _ := res.LastInsertId()
	result.ID = uint(id)

	return nil
}

// GetResult retrieves a result by ID
func (r *Repository) GetResult(ctx context.Context, resultID uint) (*MathResult, error) {
	result := &MathResult{}
	query := `
		SELECT id, user_id, mode, difficulty, total_questions, correct_answers, 
		       total_time, average_time, accuracy, timestamp
		FROM results WHERE id = ?
	`

	err := r.db.QueryRowContext(ctx, query, resultID).Scan(
		&result.ID, &result.UserID, &result.Mode, &result.Difficulty, &result.TotalQuestions,
		&result.CorrectAnswers, &result.TotalTime, &result.AverageTime, &result.Accuracy, &result.Timestamp,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("result not found")
		}
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	return result, nil
}

// GetResultsByUser retrieves results for a user with pagination
func (r *Repository) GetResultsByUser(ctx context.Context, userID uint, limit int, offset int) ([]*MathResult, error) {
	query := `
		SELECT id, user_id, mode, difficulty, total_questions, correct_answers,
		       total_time, average_time, accuracy, timestamp
		FROM results WHERE user_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get results: %w", err)
	}
	defer rows.Close()

	var results []*MathResult
	for rows.Next() {
		result := &MathResult{}
		if err := rows.Scan(&result.ID, &result.UserID, &result.Mode, &result.Difficulty,
			&result.TotalQuestions, &result.CorrectAnswers, &result.TotalTime, &result.AverageTime,
			&result.Accuracy, &result.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// GetResultCount returns the total count of results for a user
func (r *Repository) GetResultCount(ctx context.Context, userID uint) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM results WHERE user_id = ?`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get result count: %w", err)
	}

	return count, nil
}

// DeleteResult deletes a result
func (r *Repository) DeleteResult(ctx context.Context, resultID uint) error {
	query := `DELETE FROM results WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, resultID)
	if err != nil {
		return fmt.Errorf("failed to delete result: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("result not found")
	}

	return nil
}

// === QUESTION HISTORY OPERATIONS ===

// SaveQuestionHistory saves a question attempt
func (r *Repository) SaveQuestionHistory(ctx context.Context, history *QuestionHistory) error {
	if err := history.Validate(); err != nil {
		return fmt.Errorf("invalid question history: %w", err)
	}

	history.Timestamp = time.Now()

	query := `
		INSERT INTO question_history (user_id, question, user_answer, correct_answer, is_correct, time_taken, fact_family, mode, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(ctx, query,
		history.UserID, history.Question, history.UserAnswer, history.CorrectAnswer,
		history.IsCorrect, history.TimeTaken, history.FactFamily, history.Mode, history.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to save question history: %w", err)
	}

	id, _ := res.LastInsertId()
	history.ID = uint(id)

	return nil
}

// GetQuestionHistory retrieves question history by ID
func (r *Repository) GetQuestionHistory(ctx context.Context, historyID uint) (*QuestionHistory, error) {
	history := &QuestionHistory{}
	query := `
		SELECT id, user_id, question, user_answer, correct_answer, is_correct, time_taken, fact_family, mode, timestamp
		FROM question_history WHERE id = ?
	`

	err := r.db.QueryRowContext(ctx, query, historyID).Scan(
		&history.ID, &history.UserID, &history.Question, &history.UserAnswer, &history.CorrectAnswer,
		&history.IsCorrect, &history.TimeTaken, &history.FactFamily, &history.Mode, &history.Timestamp,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("question history not found")
		}
		return nil, fmt.Errorf("failed to get question history: %w", err)
	}

	return history, nil
}

// GetHistoryByUser retrieves question history for a user
func (r *Repository) GetHistoryByUser(ctx context.Context, userID uint, limit int, offset int) ([]*QuestionHistory, error) {
	query := `
		SELECT id, user_id, question, user_answer, correct_answer, is_correct, time_taken, fact_family, mode, timestamp
		FROM question_history WHERE user_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var histories []*QuestionHistory
	for rows.Next() {
		history := &QuestionHistory{}
		if err := rows.Scan(&history.ID, &history.UserID, &history.Question, &history.UserAnswer,
			&history.CorrectAnswer, &history.IsCorrect, &history.TimeTaken, &history.FactFamily,
			&history.Mode, &history.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		histories = append(histories, history)
	}

	return histories, rows.Err()
}

// === MISTAKE OPERATIONS ===

// SaveMistake records a mistake
func (r *Repository) SaveMistake(ctx context.Context, mistake *Mistake) error {
	if err := mistake.Validate(); err != nil {
		return fmt.Errorf("invalid mistake: %w", err)
	}

	mistake.LastError = time.Now()

	query := `
		INSERT INTO mistakes (user_id, question, correct_answer, user_answer, mode, fact_family, error_count, last_error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, question) DO UPDATE SET error_count = error_count + 1, last_error = excluded.last_error
	`

	_, err := r.db.ExecContext(ctx, query,
		mistake.UserID, mistake.Question, mistake.CorrectAnswer, mistake.UserAnswer,
		mistake.Mode, mistake.FactFamily, mistake.ErrorCount, mistake.LastError,
	)

	if err != nil {
		return fmt.Errorf("failed to save mistake: %w", err)
	}

	return nil
}

// GetMistake retrieves a specific mistake
func (r *Repository) GetMistake(ctx context.Context, userID uint, question string) (*Mistake, error) {
	mistake := &Mistake{}
	query := `
		SELECT id, user_id, question, correct_answer, user_answer, mode, fact_family, error_count, last_error
		FROM mistakes WHERE user_id = ? AND question = ?
	`

	err := r.db.QueryRowContext(ctx, query, userID, question).Scan(
		&mistake.ID, &mistake.UserID, &mistake.Question, &mistake.CorrectAnswer, &mistake.UserAnswer,
		&mistake.Mode, &mistake.FactFamily, &mistake.ErrorCount, &mistake.LastError,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mistake not found")
		}
		return nil, fmt.Errorf("failed to get mistake: %w", err)
	}

	return mistake, nil
}

// GetMistakesByUser retrieves mistakes for a user
func (r *Repository) GetMistakesByUser(ctx context.Context, userID uint, limit int) ([]*Mistake, error) {
	query := `
		SELECT id, user_id, question, correct_answer, user_answer, mode, fact_family, error_count, last_error
		FROM mistakes WHERE user_id = ? ORDER BY error_count DESC LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get mistakes: %w", err)
	}
	defer rows.Close()

	var mistakes []*Mistake
	for rows.Next() {
		mistake := &Mistake{}
		if err := rows.Scan(&mistake.ID, &mistake.UserID, &mistake.Question, &mistake.CorrectAnswer,
			&mistake.UserAnswer, &mistake.Mode, &mistake.FactFamily, &mistake.ErrorCount, &mistake.LastError); err != nil {
			return nil, fmt.Errorf("failed to scan mistake: %w", err)
		}
		mistakes = append(mistakes, mistake)
	}

	return mistakes, rows.Err()
}

// GetMistakesByFactFamily retrieves mistakes grouped by fact family
func (r *Repository) GetMistakesByFactFamily(ctx context.Context, userID uint, factFamily string) ([]*Mistake, error) {
	query := `
		SELECT id, user_id, question, correct_answer, user_answer, mode, fact_family, error_count, last_error
		FROM mistakes WHERE user_id = ? AND fact_family = ? ORDER BY error_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, factFamily)
	if err != nil {
		return nil, fmt.Errorf("failed to get mistakes: %w", err)
	}
	defer rows.Close()

	var mistakes []*Mistake
	for rows.Next() {
		mistake := &Mistake{}
		if err := rows.Scan(&mistake.ID, &mistake.UserID, &mistake.Question, &mistake.CorrectAnswer,
			&mistake.UserAnswer, &mistake.Mode, &mistake.FactFamily, &mistake.ErrorCount, &mistake.LastError); err != nil {
			return nil, fmt.Errorf("failed to scan mistake: %w", err)
		}
		mistakes = append(mistakes, mistake)
	}

	return mistakes, rows.Err()
}

// === MASTERY OPERATIONS ===

// SaveMastery saves or updates mastery for a fact
func (r *Repository) SaveMastery(ctx context.Context, mastery *Mastery) error {
	if err := mastery.Validate(); err != nil {
		return fmt.Errorf("invalid mastery: %w", err)
	}

	mastery.LastPracticed = time.Now()

	query := `
		INSERT INTO mastery (user_id, fact, mode, correct_streak, total_attempts, mastery_level, last_practiced, average_response_time, fastest_time, slowest_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, fact, mode) DO UPDATE SET
			correct_streak = excluded.correct_streak,
			total_attempts = excluded.total_attempts,
			mastery_level = excluded.mastery_level,
			last_practiced = excluded.last_practiced,
			average_response_time = excluded.average_response_time,
			fastest_time = excluded.fastest_time,
			slowest_time = excluded.slowest_time
	`

	_, err := r.db.ExecContext(ctx, query,
		mastery.UserID, mastery.Fact, mastery.Mode, mastery.CorrectStreak, mastery.TotalAttempts,
		mastery.MasteryLevel, mastery.LastPracticed, mastery.AverageResponseTime,
		mastery.FastestTime, mastery.SlowestTime,
	)

	if err != nil {
		return fmt.Errorf("failed to save mastery: %w", err)
	}

	return nil
}

// GetMastery retrieves mastery for a specific fact
func (r *Repository) GetMastery(ctx context.Context, userID uint, fact string, mode string) (*Mastery, error) {
	mastery := &Mastery{}
	query := `
		SELECT id, user_id, fact, mode, correct_streak, total_attempts, mastery_level, last_practiced, average_response_time, fastest_time, slowest_time
		FROM mastery WHERE user_id = ? AND fact = ? AND mode = ?
	`

	err := r.db.QueryRowContext(ctx, query, userID, fact, mode).Scan(
		&mastery.ID, &mastery.UserID, &mastery.Fact, &mastery.Mode, &mastery.CorrectStreak,
		&mastery.TotalAttempts, &mastery.MasteryLevel, &mastery.LastPracticed, &mastery.AverageResponseTime,
		&mastery.FastestTime, &mastery.SlowestTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mastery not found")
		}
		return nil, fmt.Errorf("failed to get mastery: %w", err)
	}

	return mastery, nil
}

// GetMasteryByUser retrieves all mastery records for a user
func (r *Repository) GetMasteryByUser(ctx context.Context, userID uint, mode string) ([]*Mastery, error) {
	query := `
		SELECT id, user_id, fact, mode, correct_streak, total_attempts, mastery_level, last_practiced, average_response_time, fastest_time, slowest_time
		FROM mastery WHERE user_id = ? AND mode = ? ORDER BY mastery_level DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to get mastery: %w", err)
	}
	defer rows.Close()

	var masteries []*Mastery
	for rows.Next() {
		mastery := &Mastery{}
		if err := rows.Scan(&mastery.ID, &mastery.UserID, &mastery.Fact, &mastery.Mode,
			&mastery.CorrectStreak, &mastery.TotalAttempts, &mastery.MasteryLevel, &mastery.LastPracticed,
			&mastery.AverageResponseTime, &mastery.FastestTime, &mastery.SlowestTime); err != nil {
			return nil, fmt.Errorf("failed to scan mastery: %w", err)
		}
		masteries = append(masteries, mastery)
	}

	return masteries, rows.Err()
}

// === LEARNING PROFILE OPERATIONS ===

// SaveLearningProfile saves or updates learning profile
func (r *Repository) SaveLearningProfile(ctx context.Context, profile *LearningProfile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid learning profile: %w", err)
	}

	profile.ProfileUpdated = time.Now()

	query := `
		INSERT INTO learning_profile (user_id, learning_style, preferred_time_of_day, attention_span_seconds, best_streak_time, weak_time_of_day, avg_session_length, total_practice_time, profile_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			learning_style = excluded.learning_style,
			preferred_time_of_day = excluded.preferred_time_of_day,
			attention_span_seconds = excluded.attention_span_seconds,
			best_streak_time = excluded.best_streak_time,
			weak_time_of_day = excluded.weak_time_of_day,
			avg_session_length = excluded.avg_session_length,
			total_practice_time = excluded.total_practice_time,
			profile_updated = excluded.profile_updated
	`

	_, err := r.db.ExecContext(ctx, query,
		profile.UserID, profile.LearningStyle, profile.PreferredTimeOfDay, profile.AttentionSpanSeconds,
		profile.BestStreakTime, profile.WeakTimeOfDay, profile.AvgSessionLength, profile.TotalPracticeTime,
		profile.ProfileUpdated,
	)

	if err != nil {
		return fmt.Errorf("failed to save learning profile: %w", err)
	}

	return nil
}

// GetLearningProfile retrieves learning profile for a user
func (r *Repository) GetLearningProfile(ctx context.Context, userID uint) (*LearningProfile, error) {
	profile := &LearningProfile{}
	query := `
		SELECT id, user_id, learning_style, preferred_time_of_day, attention_span_seconds, best_streak_time, weak_time_of_day, avg_session_length, total_practice_time, profile_updated
		FROM learning_profile WHERE user_id = ?
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.LearningStyle, &profile.PreferredTimeOfDay,
		&profile.AttentionSpanSeconds, &profile.BestStreakTime, &profile.WeakTimeOfDay,
		&profile.AvgSessionLength, &profile.TotalPracticeTime, &profile.ProfileUpdated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("learning profile not found")
		}
		return nil, fmt.Errorf("failed to get learning profile: %w", err)
	}

	return profile, nil
}

// === PERFORMANCE PATTERN OPERATIONS ===

// SavePerformancePattern saves or updates performance pattern
func (r *Repository) SavePerformancePattern(ctx context.Context, pattern *PerformancePattern) error {
	if err := pattern.Validate(); err != nil {
		return fmt.Errorf("invalid performance pattern: %w", err)
	}

	pattern.LastUpdated = time.Now()

	query := `
		INSERT INTO performance_patterns (user_id, hour_of_day, day_of_week, average_accuracy, average_speed, session_count, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, hour_of_day, day_of_week) DO UPDATE SET
			average_accuracy = excluded.average_accuracy,
			average_speed = excluded.average_speed,
			session_count = excluded.session_count,
			last_updated = excluded.last_updated
	`

	_, err := r.db.ExecContext(ctx, query,
		pattern.UserID, pattern.HourOfDay, pattern.DayOfWeek, pattern.AverageAccuracy,
		pattern.AverageSpeed, pattern.SessionCount, pattern.LastUpdated,
	)

	if err != nil {
		return fmt.Errorf("failed to save performance pattern: %w", err)
	}

	return nil
}

// GetPerformancePattern retrieves performance pattern for specific time
func (r *Repository) GetPerformancePattern(ctx context.Context, userID uint, hour int, dayOfWeek int) (*PerformancePattern, error) {
	pattern := &PerformancePattern{}
	query := `
		SELECT id, user_id, hour_of_day, day_of_week, average_accuracy, average_speed, session_count, last_updated
		FROM performance_patterns WHERE user_id = ? AND hour_of_day = ? AND day_of_week = ?
	`

	err := r.db.QueryRowContext(ctx, query, userID, hour, dayOfWeek).Scan(
		&pattern.ID, &pattern.UserID, &pattern.HourOfDay, &pattern.DayOfWeek, &pattern.AverageAccuracy,
		&pattern.AverageSpeed, &pattern.SessionCount, &pattern.LastUpdated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("performance pattern not found")
		}
		return nil, fmt.Errorf("failed to get performance pattern: %w", err)
	}

	return pattern, nil
}

// GetPatternsByUser retrieves all performance patterns for a user
func (r *Repository) GetPatternsByUser(ctx context.Context, userID uint) ([]*PerformancePattern, error) {
	query := `
		SELECT id, user_id, hour_of_day, day_of_week, average_accuracy, average_speed, session_count, last_updated
		FROM performance_patterns WHERE user_id = ? ORDER BY average_accuracy DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get patterns: %w", err)
	}
	defer rows.Close()

	var patterns []*PerformancePattern
	for rows.Next() {
		pattern := &PerformancePattern{}
		if err := rows.Scan(&pattern.ID, &pattern.UserID, &pattern.HourOfDay, &pattern.DayOfWeek,
			&pattern.AverageAccuracy, &pattern.AverageSpeed, &pattern.SessionCount, &pattern.LastUpdated); err != nil {
			return nil, fmt.Errorf("failed to scan pattern: %w", err)
		}
		patterns = append(patterns, pattern)
	}

	return patterns, rows.Err()
}

// === REPETITION SCHEDULE OPERATIONS ===

// SaveRepetitionSchedule saves or updates repetition schedule
func (r *Repository) SaveRepetitionSchedule(ctx context.Context, schedule *RepetitionSchedule) error {
	if err := schedule.Validate(); err != nil {
		return fmt.Errorf("invalid repetition schedule: %w", err)
	}

	query := `
		INSERT INTO repetition_schedule (user_id, fact, mode, next_review, interval_days, ease_factor, review_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, fact, mode) DO UPDATE SET
			next_review = excluded.next_review,
			interval_days = excluded.interval_days,
			ease_factor = excluded.ease_factor,
			review_count = excluded.review_count
	`

	res, err := r.db.ExecContext(ctx, query,
		schedule.UserID, schedule.Fact, schedule.Mode, schedule.NextReview,
		schedule.IntervalDays, schedule.EaseFactor, schedule.ReviewCount,
	)

	if err != nil {
		return fmt.Errorf("failed to save repetition schedule: %w", err)
	}

	id, _ := res.LastInsertId()
	schedule.ID = uint(id)

	return nil
}

// GetRepetitionSchedule retrieves repetition schedule for a fact
func (r *Repository) GetRepetitionSchedule(ctx context.Context, userID uint, fact string, mode string) (*RepetitionSchedule, error) {
	schedule := &RepetitionSchedule{}
	query := `
		SELECT id, user_id, fact, mode, next_review, interval_days, ease_factor, review_count
		FROM repetition_schedule WHERE user_id = ? AND fact = ? AND mode = ?
	`

	err := r.db.QueryRowContext(ctx, query, userID, fact, mode).Scan(
		&schedule.ID, &schedule.UserID, &schedule.Fact, &schedule.Mode, &schedule.NextReview,
		&schedule.IntervalDays, &schedule.EaseFactor, &schedule.ReviewCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("repetition schedule not found")
		}
		return nil, fmt.Errorf("failed to get repetition schedule: %w", err)
	}

	return schedule, nil
}

// GetDueRepetitions retrieves facts due for review
func (r *Repository) GetDueRepetitions(ctx context.Context, userID uint, limit int) ([]*RepetitionSchedule, error) {
	query := `
		SELECT id, user_id, fact, mode, next_review, interval_days, ease_factor, review_count
		FROM repetition_schedule WHERE user_id = ? AND next_review <= datetime('now') ORDER BY next_review ASC LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get due repetitions: %w", err)
	}
	defer rows.Close()

	var schedules []*RepetitionSchedule
	for rows.Next() {
		schedule := &RepetitionSchedule{}
		if err := rows.Scan(&schedule.ID, &schedule.UserID, &schedule.Fact, &schedule.Mode,
			&schedule.NextReview, &schedule.IntervalDays, &schedule.EaseFactor, &schedule.ReviewCount); err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

// GetAllRepetitions retrieves all repetition schedules for a user
func (r *Repository) GetAllRepetitions(ctx context.Context, userID uint) ([]*RepetitionSchedule, error) {
	query := `
		SELECT id, user_id, fact, mode, next_review, interval_days, ease_factor, review_count
		FROM repetition_schedule WHERE user_id = ? ORDER BY next_review ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get repetitions: %w", err)
	}
	defer rows.Close()

	var schedules []*RepetitionSchedule
	for rows.Next() {
		schedule := &RepetitionSchedule{}
		if err := rows.Scan(&schedule.ID, &schedule.UserID, &schedule.Fact, &schedule.Mode,
			&schedule.NextReview, &schedule.IntervalDays, &schedule.EaseFactor, &schedule.ReviewCount); err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

// === COMPLEX QUERY OPERATIONS ===

// GetWeakFactFamilies retrieves fact families with most errors
func (r *Repository) GetWeakFactFamilies(ctx context.Context, userID uint, minErrors int) (map[string]int, error) {
	query := `
		SELECT fact_family, COUNT(*) as error_count
		FROM mistakes
		WHERE user_id = ?
		GROUP BY fact_family
		HAVING COUNT(*) >= ?
		ORDER BY error_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, minErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to get weak fact families: %w", err)
	}
	defer rows.Close()

	weakFamilies := make(map[string]int)
	for rows.Next() {
		var family string
		var count int
		if err := rows.Scan(&family, &count); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		weakFamilies[family] = count
	}

	return weakFamilies, rows.Err()
}

// UserStats represents aggregated user statistics
type UserStats struct {
	UserID        uint
	TotalSessions int
	AverageAccuracy float64
	BestAccuracy  float64
	TotalQuestions int
	CorrectAnswers int
	TotalMastered  int
	TotalMistakes  int
	AverageWPM    float64
}

// GetUserStats retrieves aggregated statistics for a user
func (r *Repository) GetUserStats(ctx context.Context, userID uint) (*UserStats, error) {
	stats := &UserStats{UserID: userID}

	// Get session stats
	sessionQuery := `
		SELECT COUNT(*) as sessions, AVG(accuracy) as avg_acc, MAX(accuracy) as best_acc
		FROM results WHERE user_id = ?
	`
	err := r.db.QueryRowContext(ctx, sessionQuery, userID).Scan(
		&stats.TotalSessions, &stats.AverageAccuracy, &stats.BestAccuracy,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	// Get question stats
	questionQuery := `
		SELECT COUNT(*) as total, SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) as correct
		FROM question_history WHERE user_id = ?
	`
	err = r.db.QueryRowContext(ctx, questionQuery, userID).Scan(
		&stats.TotalQuestions, &stats.CorrectAnswers,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get question stats: %w", err)
	}

	// Get mastery count
	masteryQuery := `SELECT COUNT(*) FROM mastery WHERE user_id = ? AND mastery_level >= 80`
	err = r.db.QueryRowContext(ctx, masteryQuery, userID).Scan(&stats.TotalMastered)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get mastery count: %w", err)
	}

	// Get mistake count
	mistakeQuery := `SELECT COUNT(*) FROM mistakes WHERE user_id = ?`
	err = r.db.QueryRowContext(ctx, mistakeQuery, userID).Scan(&stats.TotalMistakes)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get mistake count: %w", err)
	}

	return stats, nil
}

// LeaderboardEntry represents a user's position on the leaderboard
type LeaderboardEntry struct {
	UserID    uint
	Username  string
	Value     float64
	Metric    string
}

// GetLeaderboard retrieves leaderboard rankings
func (r *Repository) GetLeaderboard(ctx context.Context, metric string, limit int) ([]*LeaderboardEntry, error) {
	var query string
	var columnName string

	switch metric {
	case "accuracy":
		columnName = "AVG(accuracy)"
	case "sessions":
		columnName = "COUNT(*)"
	case "speed":
		columnName = "AVG(average_time)"
	default:
		columnName = "COUNT(*)"
	}

	// Build query with the appropriate metric
	query = fmt.Sprintf(`
		SELECT u.id, u.username, %s as value
		FROM results r
		JOIN users u ON r.user_id = u.id
		GROUP BY r.user_id, u.id, u.username
		ORDER BY value DESC
		LIMIT ?
	`, columnName)

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*LeaderboardEntry
	for rows.Next() {
		entry := &LeaderboardEntry{Metric: metric}
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.Value); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// MistakeAnalysis represents mistake analysis for a user
type MistakeAnalysis struct {
	FactFamily string
	ErrorCount int
	Severity   string
}

// GetMistakeAnalysis retrieves mistake analysis grouped by fact family
func (r *Repository) GetMistakeAnalysis(ctx context.Context, userID uint) ([]*MistakeAnalysis, error) {
	query := `
		SELECT fact_family, SUM(error_count) as total_errors
		FROM mistakes
		WHERE user_id = ?
		GROUP BY fact_family
		ORDER BY total_errors DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mistake analysis: %w", err)
	}
	defer rows.Close()

	var analyses []*MistakeAnalysis
	for rows.Next() {
		analysis := &MistakeAnalysis{}
		if err := rows.Scan(&analysis.FactFamily, &analysis.ErrorCount); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}

		// Determine severity
		if analysis.ErrorCount >= 5 {
			analysis.Severity = "critical"
		} else if analysis.ErrorCount >= 4 {
			analysis.Severity = "high"
		} else if analysis.ErrorCount >= 3 {
			analysis.Severity = "medium"
		} else {
			analysis.Severity = "low"
		}

		analyses = append(analyses, analysis)
	}

	return analyses, rows.Err()
}

// GetRecentActivity retrieves recent activity for a user
func (r *Repository) GetRecentActivity(ctx context.Context, userID uint, hours int) (map[string]int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) as count
		FROM question_history
		WHERE user_id = ? AND timestamp >= datetime('now', '-%d hours')
	`, hours)

	result := make(map[string]int)
	var count int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	result["questions_answered"] = count

	// Get accuracy in recent period
	accuracyQuery := fmt.Sprintf(`
		SELECT AVG(CASE WHEN is_correct THEN 1 ELSE 0 END) * 100
		FROM question_history
		WHERE user_id = ? AND timestamp >= datetime('now', '-%d hours')
	`, hours)

	var accuracy sql.NullFloat64
	err = r.db.QueryRowContext(ctx, accuracyQuery, userID).Scan(&accuracy)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get recent accuracy: %w", err)
	}

	if accuracy.Valid {
		result["recent_accuracy_percent"] = int(accuracy.Float64)
	}

	return result, nil
}

// GetBestPerformanceTime retrieves the time of day with best performance
func (r *Repository) GetBestPerformanceTime(ctx context.Context, userID uint) (string, float64, error) {
	query := `
		SELECT hour_of_day, average_accuracy
		FROM performance_patterns
		WHERE user_id = ?
		ORDER BY average_accuracy DESC
		LIMIT 1
	`

	var hour int
	var accuracy float64

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&hour, &accuracy)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, fmt.Errorf("failed to get best performance time: %w", err)
	}

	timeOfDay := GetTimeOfDayFromHour(hour)
	return timeOfDay, accuracy, nil
}
