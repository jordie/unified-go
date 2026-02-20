package typing

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SaveRace saves a racing session to the database
func (r *Repository) SaveRace(ctx context.Context, race *Race) (uint, error) {
	if err := race.Validate(); err != nil {
		return 0, fmt.Errorf("invalid race: %w", err)
	}

	query := `
		INSERT INTO races (
			user_id, mode, placement, wpm, accuracy, race_time, xp_earned, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		race.UserID,
		race.Mode,
		race.Placement,
		race.WPM,
		race.Accuracy,
		race.RaceTime,
		race.XPEarned,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save race: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Update racing stats
	if err := r.updateRacingStats(ctx, race); err != nil {
		return uint(id), fmt.Errorf("failed to update racing stats: %w", err)
	}

	return uint(id), nil
}

// GetRace retrieves a specific race by ID
func (r *Repository) GetRace(ctx context.Context, raceID uint) (*Race, error) {
	query := `
		SELECT
			id, user_id, mode, placement, wpm, accuracy, race_time, xp_earned, created_at
		FROM races
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, raceID)
	race := &Race{}

	if err := row.Scan(
		&race.ID,
		&race.UserID,
		&race.Mode,
		&race.Placement,
		&race.WPM,
		&race.Accuracy,
		&race.RaceTime,
		&race.XPEarned,
		&race.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("race not found")
		}
		return nil, fmt.Errorf("failed to query race: %w", err)
	}

	return race, nil
}

// GetRacesByUser retrieves all races for a user with pagination
func (r *Repository) GetRacesByUser(ctx context.Context, userID uint, limit, offset int) ([]Race, error) {
	query := `
		SELECT
			id, user_id, mode, placement, wpm, accuracy, race_time, xp_earned, created_at
		FROM races
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query races: %w", err)
	}
	defer rows.Close()

	var races []Race
	for rows.Next() {
		race := Race{}
		if err := race.ScanRow(rows); err != nil {
			return nil, fmt.Errorf("failed to scan race: %w", err)
		}
		races = append(races, race)
	}

	return races, rows.Err()
}

// GetUserRacingStats retrieves aggregated racing statistics for a user
func (r *Repository) GetUserRacingStats(ctx context.Context, userID uint) (*UserRacingStats, error) {
	query := `
		SELECT
			id, user_id, total_races, wins, podiums, total_xp, current_car, last_updated
		FROM user_racing_stats
		WHERE user_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, userID)
	stats := &UserRacingStats{}

	if err := row.Scan(
		&stats.ID,
		&stats.UserID,
		&stats.TotalRaces,
		&stats.Wins,
		&stats.Podiums,
		&stats.TotalXP,
		&stats.CurrentCar,
		&stats.LastUpdated,
	); err != nil {
		if err == sql.ErrNoRows {
			// Create default racing stats for new user
			return r.createDefaultRacingStats(ctx, userID)
		}
		return nil, fmt.Errorf("failed to query racing stats: %w", err)
	}

	return stats, nil
}

// createDefaultRacingStats creates default racing stats for a new user
func (r *Repository) createDefaultRacingStats(ctx context.Context, userID uint) (*UserRacingStats, error) {
	stats := &UserRacingStats{
		UserID:      userID,
		TotalRaces:  0,
		Wins:        0,
		Podiums:     0,
		TotalXP:     0,
		CurrentCar:  "🚗",
		LastUpdated: time.Now(),
	}

	query := `
		INSERT INTO user_racing_stats (
			user_id, total_races, wins, podiums, total_xp, current_car, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		stats.UserID,
		stats.TotalRaces,
		stats.Wins,
		stats.Podiums,
		stats.TotalXP,
		stats.CurrentCar,
		stats.LastUpdated,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create racing stats: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	stats.ID = uint(id)
	return stats, nil
}

// updateRacingStats updates racing statistics after a race
func (r *Repository) updateRacingStats(ctx context.Context, race *Race) error {
	// Get current stats
	stats, err := r.GetUserRacingStats(ctx, race.UserID)
	if err != nil && err.Error() != "race not found" {
		return err
	}

	// Calculate updates
	totalRaces := 1
	wins := 0
	podiums := 0

	if stats != nil {
		totalRaces = stats.TotalRaces + 1
		wins = stats.Wins
		podiums = stats.Podiums
	}

	if race.Placement == 1 {
		wins++
		podiums++
	} else if race.Placement <= 3 {
		podiums++
	}

	totalXP := race.XPEarned
	if stats != nil {
		totalXP += stats.TotalXP
	}

	// Determine current car based on XP
	currentCar := "🚗"
	for _, prog := range CarProgressions {
		if totalXP >= prog.XPRequired {
			currentCar = prog.Car
		}
	}

	// Update or insert
	if stats != nil {
		updateQuery := `
			UPDATE user_racing_stats
			SET total_races = ?,
				wins = ?,
				podiums = ?,
				total_xp = ?,
				current_car = ?,
				last_updated = ?
			WHERE user_id = ?
		`

		_, err := r.db.ExecContext(
			ctx,
			updateQuery,
			totalRaces,
			wins,
			podiums,
			totalXP,
			currentCar,
			time.Now(),
			race.UserID,
		)
		if err != nil {
			return fmt.Errorf("failed to update racing stats: %w", err)
		}
	} else {
		insertQuery := `
			INSERT INTO user_racing_stats (
				user_id, total_races, wins, podiums, total_xp, current_car, last_updated
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`

		_, err := r.db.ExecContext(
			ctx,
			insertQuery,
			race.UserID,
			totalRaces,
			wins,
			podiums,
			totalXP,
			currentCar,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert racing stats: %w", err)
		}
	}

	return nil
}

// GetRacingLeaderboard retrieves top racers ranked by metric
func (r *Repository) GetRacingLeaderboard(ctx context.Context, metric string, limit int) ([]UserRacingStats, error) {
	orderBy := "total_xp DESC"
	switch metric {
	case "wins":
		orderBy = "wins DESC"
	case "wpm":
		orderBy = "total_xp DESC" // Can be enhanced with average WPM
	case "races":
		orderBy = "total_races DESC"
	}

	query := fmt.Sprintf(`
		SELECT
			id, user_id, total_races, wins, podiums, total_xp, current_car, last_updated
		FROM user_racing_stats
		ORDER BY %s
		LIMIT ?
	`, orderBy)

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var leaderboard []UserRacingStats
	for rows.Next() {
		stats := UserRacingStats{}
		if err := stats.ScanRow(rows); err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		leaderboard = append(leaderboard, stats)
	}

	return leaderboard, rows.Err()
}

// GetRacesByMode retrieves races filtered by mode
func (r *Repository) GetRacesByMode(ctx context.Context, userID uint, mode string, limit, offset int) ([]Race, error) {
	query := `
		SELECT
			id, user_id, mode, placement, wpm, accuracy, race_time, xp_earned, created_at
		FROM races
		WHERE user_id = ? AND mode = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, mode, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query races: %w", err)
	}
	defer rows.Close()

	var races []Race
	for rows.Next() {
		race := Race{}
		if err := race.ScanRow(rows); err != nil {
			return nil, fmt.Errorf("failed to scan race: %w", err)
		}
		races = append(races, race)
	}

	return races, rows.Err()
}

// GetRaceCount returns total number of races for a user
func (r *Repository) GetRaceCount(ctx context.Context, userID uint) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM races WHERE user_id = ?"
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get race count: %w", err)
	}
	return count, nil
}

// DeleteRace deletes a specific race (admin operation)
func (r *Repository) DeleteRace(ctx context.Context, raceID uint) error {
	query := "DELETE FROM races WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, raceID)
	if err != nil {
		return fmt.Errorf("failed to delete race: %w", err)
	}
	return nil
}

// DeleteUserRaces deletes all races for a user (admin operation)
func (r *Repository) DeleteUserRaces(ctx context.Context, userID uint) error {
	query := "DELETE FROM races WHERE user_id = ?"
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user races: %w", err)
	}

	// Also delete racing stats
	statsQuery := "DELETE FROM user_racing_stats WHERE user_id = ?"
	_, err = r.db.ExecContext(ctx, statsQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to delete racing stats: %w", err)
	}

	return nil
}
