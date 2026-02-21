package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"unified-go/internal/models"
	"unified-go/internal/storage"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	// Create adds a new session
	Create(ctx context.Context, session *models.SessionCreate) (int64, error)

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id int64) (*models.Session, error)

	// GetBySessionID retrieves a session by session ID string
	GetBySessionID(ctx context.Context, sessionID string) (*models.Session, error)

	// GetByProvider retrieves all sessions for a provider
	GetByProvider(ctx context.Context, provider string) ([]*models.Session, error)

	// GetByStatus retrieves all sessions with a specific status
	GetByStatus(ctx context.Context, status models.SessionState) ([]*models.Session, error)

	// GetActive retrieves all active sessions
	GetActive(ctx context.Context) ([]*models.Session, error)

	// UpdateStatus updates session status and reason
	UpdateStatus(ctx context.Context, sessionID string, status models.SessionState) error

	// UpdateHealth updates session health score
	UpdateHealth(ctx context.Context, sessionID string, healthScore float64) error

	// UpdateMetrics updates session metrics
	UpdateMetrics(ctx context.Context, sessionID string, metrics *models.SessionMetrics) error

	// RecordHeartbeat records a session heartbeat
	RecordHeartbeat(ctx context.Context, sessionID string) error

	// GetHealth retrieves health information for a session
	GetHealth(ctx context.Context, sessionID string) (*models.SessionHealth, error)

	// GetMetrics retrieves metrics for a session
	GetMetrics(ctx context.Context, sessionID string) (*models.SessionMetrics, error)

	// LogStateChange records a session state change
	LogStateChange(ctx context.Context, sessionID string, oldState, newState models.SessionState, reason string) error

	// GetStateChangeLogs retrieves state change history for a session
	GetStateChangeLogs(ctx context.Context, sessionID string) ([]*models.SessionStateLog, error)

	// Deactivate marks a session as inactive
	Deactivate(ctx context.Context, sessionID string) error

	// Delete removes a session
	Delete(ctx context.Context, sessionID string) error

	// GetIdleSessions retrieves idle sessions available for assignment
	GetIdleSessions(ctx context.Context) ([]*models.Session, error)

	// SetCurrentTask sets the current task for a session
	SetCurrentTask(ctx context.Context, sessionID string, taskID int64) error

	// ClearCurrentTask clears the current task for a session
	ClearCurrentTask(ctx context.Context, sessionID string) error
}

// sessionRepository implements SessionRepository
type sessionRepository struct {
	store *storage.SQLiteStore
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(store *storage.SQLiteStore) SessionRepository {
	return &sessionRepository{store: store}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.SessionCreate) (int64, error) {
	result, err := r.store.Exec(ctx,
		`INSERT INTO sessions (session_id, session_type, provider, status, health_score, active)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		session.SessionID,
		session.SessionType,
		session.Provider,
		session.Status,
		100.0, // Initial health score
		1,     // Active by default
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}
	return result.LastInsertId()
}

func (r *sessionRepository) GetByID(ctx context.Context, id int64) (*models.Session, error) {
	row := r.store.QueryRow(ctx,
		`SELECT id, session_id, session_type, provider, status, current_task_id, last_heartbeat,
		        health_score, metrics_json, created_at, updated_at, active
		 FROM sessions WHERE id = ?`,
		id)

	session := &models.Session{}
	err := row.Scan(
		&session.ID, &session.SessionID, &session.SessionType, &session.Provider,
		&session.Status, &session.CurrentTaskID, &session.LastHeartbeat,
		&session.HealthScore, &session.MetricsJSON, &session.CreatedAt, &session.UpdatedAt,
		&session.Active,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

func (r *sessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*models.Session, error) {
	row := r.store.QueryRow(ctx,
		`SELECT id, session_id, session_type, provider, status, current_task_id, last_heartbeat,
		        health_score, metrics_json, created_at, updated_at, active
		 FROM sessions WHERE session_id = ?`,
		sessionID)

	session := &models.Session{}
	err := row.Scan(
		&session.ID, &session.SessionID, &session.SessionType, &session.Provider,
		&session.Status, &session.CurrentTaskID, &session.LastHeartbeat,
		&session.HealthScore, &session.MetricsJSON, &session.CreatedAt, &session.UpdatedAt,
		&session.Active,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

func (r *sessionRepository) GetByProvider(ctx context.Context, provider string) ([]*models.Session, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, session_id, session_type, provider, status, current_task_id, last_heartbeat,
		        health_score, metrics_json, created_at, updated_at, active
		 FROM sessions WHERE provider = ? AND active = 1 ORDER BY health_score DESC`,
		provider)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions by provider: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID, &session.SessionID, &session.SessionType, &session.Provider,
			&session.Status, &session.CurrentTaskID, &session.LastHeartbeat,
			&session.HealthScore, &session.MetricsJSON, &session.CreatedAt, &session.UpdatedAt,
			&session.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *sessionRepository) GetByStatus(ctx context.Context, status models.SessionState) ([]*models.Session, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, session_id, session_type, provider, status, current_task_id, last_heartbeat,
		        health_score, metrics_json, created_at, updated_at, active
		 FROM sessions WHERE status = ? AND active = 1 ORDER BY health_score DESC`,
		status)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions by status: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID, &session.SessionID, &session.SessionType, &session.Provider,
			&session.Status, &session.CurrentTaskID, &session.LastHeartbeat,
			&session.HealthScore, &session.MetricsJSON, &session.CreatedAt, &session.UpdatedAt,
			&session.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *sessionRepository) GetActive(ctx context.Context) ([]*models.Session, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, session_id, session_type, provider, status, current_task_id, last_heartbeat,
		        health_score, metrics_json, created_at, updated_at, active
		 FROM sessions WHERE active = 1 ORDER BY last_heartbeat DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID, &session.SessionID, &session.SessionType, &session.Provider,
			&session.Status, &session.CurrentTaskID, &session.LastHeartbeat,
			&session.HealthScore, &session.MetricsJSON, &session.CreatedAt, &session.UpdatedAt,
			&session.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *sessionRepository) UpdateStatus(ctx context.Context, sessionID string, status models.SessionState) error {
	oldSession, err := r.GetBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}
	if oldSession == nil {
		return fmt.Errorf("session not found")
	}

	_, err = r.store.Exec(ctx,
		`UPDATE sessions SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		status, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	return r.LogStateChange(ctx, sessionID, oldSession.Status, status, "Status update")
}

func (r *sessionRepository) UpdateHealth(ctx context.Context, sessionID string, healthScore float64) error {
	if healthScore < 0 {
		healthScore = 0
	}
	if healthScore > 100 {
		healthScore = 100
	}

	_, err := r.store.Exec(ctx,
		`UPDATE sessions SET health_score = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		healthScore, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session health: %w", err)
	}
	return nil
}

func (r *sessionRepository) UpdateMetrics(ctx context.Context, sessionID string, metrics *models.SessionMetrics) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	_, err = r.store.Exec(ctx,
		`UPDATE sessions SET metrics_json = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		string(metricsJSON), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session metrics: %w", err)
	}
	return nil
}

func (r *sessionRepository) RecordHeartbeat(ctx context.Context, sessionID string) error {
	_, err := r.store.Exec(ctx,
		`UPDATE sessions SET last_heartbeat = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		sessionID)
	if err != nil {
		return fmt.Errorf("failed to record heartbeat: %w", err)
	}
	return nil
}

func (r *sessionRepository) GetHealth(ctx context.Context, sessionID string) (*models.SessionHealth, error) {
	row := r.store.QueryRow(ctx,
		`SELECT id, session_id, status, health_score, last_heartbeat, current_task_id FROM sessions WHERE session_id = ?`,
		sessionID)

	var id int64
	var session_id string
	var status models.SessionState
	var healthScore float64
	var lastHeartbeat time.Time
	var currentTaskID sql.NullInt64

	err := row.Scan(&id, &session_id, &status, &healthScore, &lastHeartbeat, &currentTaskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session health: %w", err)
	}

	timeSinceLast := int64(time.Since(lastHeartbeat).Seconds())

	health := &models.SessionHealth{
		SessionID:                  session_id,
		HealthScore:                healthScore,
		Status:                     status,
		IsHealthy:                  healthScore >= 50 && status.IsHealthy(),
		LastHeartbeat:              lastHeartbeat,
		TimeSinceLastHeartbeat:     timeSinceLast,
	}

	if currentTaskID.Valid {
		health.ActiveTaskID = &currentTaskID.Int64
	}

	return health, nil
}

func (r *sessionRepository) GetMetrics(ctx context.Context, sessionID string) (*models.SessionMetrics, error) {
	row := r.store.QueryRow(ctx,
		`SELECT metrics_json FROM sessions WHERE session_id = ?`,
		sessionID)

	var metricsJSON sql.NullString
	err := row.Scan(&metricsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session metrics: %w", err)
	}

	if !metricsJSON.Valid {
		return &models.SessionMetrics{SessionID: sessionID}, nil
	}

	var metrics models.SessionMetrics
	err = json.Unmarshal([]byte(metricsJSON.String), &metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &metrics, nil
}

func (r *sessionRepository) LogStateChange(ctx context.Context, sessionID string, oldState, newState models.SessionState, reason string) error {
	_, err := r.store.Exec(ctx,
		`INSERT INTO session_state_log (session_id, old_state, new_state, reason) VALUES (?, ?, ?, ?)`,
		sessionID, oldState, newState, reason)
	if err != nil {
		return fmt.Errorf("failed to log state change: %w", err)
	}
	return nil
}

func (r *sessionRepository) GetStateChangeLogs(ctx context.Context, sessionID string) ([]*models.SessionStateLog, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, session_id, old_state, new_state, reason, timestamp FROM session_state_log WHERE session_id = ? ORDER BY timestamp DESC`,
		sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query state logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.SessionStateLog
	for rows.Next() {
		log := &models.SessionStateLog{}
		err := rows.Scan(&log.ID, &log.SessionID, &log.OldState, &log.NewState, &log.Reason, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *sessionRepository) Deactivate(ctx context.Context, sessionID string) error {
	_, err := r.store.Exec(ctx,
		`UPDATE sessions SET active = 0, status = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		models.SessionTerminated, sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}
	return r.LogStateChange(ctx, sessionID, models.SessionBusy, models.SessionTerminated, "Session deactivated")
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	_, err := r.store.Exec(ctx, `DELETE FROM sessions WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *sessionRepository) GetIdleSessions(ctx context.Context) ([]*models.Session, error) {
	return r.GetByStatus(ctx, models.SessionIdle)
}

func (r *sessionRepository) SetCurrentTask(ctx context.Context, sessionID string, taskID int64) error {
	_, err := r.store.Exec(ctx,
		`UPDATE sessions SET current_task_id = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		taskID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to set current task: %w", err)
	}
	return nil
}

func (r *sessionRepository) ClearCurrentTask(ctx context.Context, sessionID string) error {
	_, err := r.store.Exec(ctx,
		`UPDATE sessions SET current_task_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		sessionID)
	if err != nil {
		return fmt.Errorf("failed to clear current task: %w", err)
	}
	return nil
}
