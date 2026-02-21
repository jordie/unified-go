package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"unified-go/internal/models"
	"unified-go/internal/storage"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	// Create adds a new task and returns its ID
	Create(ctx context.Context, task *models.TaskCreate) (int64, error)

	// GetByID retrieves a task by ID
	GetByID(ctx context.Context, id int64) (*models.Task, error)

	// GetPending retrieves the next pending task for assignment
	GetPending(ctx context.Context, limit int) ([]*models.Task, error)

	// GetByStatus retrieves all tasks with a specific status
	GetByStatus(ctx context.Context, status models.TaskStatus, limit int, offset int) ([]*models.Task, error)

	// GetBySession retrieves tasks assigned to a session
	GetBySession(ctx context.Context, sessionID string, status models.TaskStatus) ([]*models.Task, error)

	// UpdateStatus updates a task's status
	UpdateStatus(ctx context.Context, id int64, status models.TaskStatus, reason string) error

	// UpdateWithError updates a task with error information
	UpdateWithError(ctx context.Context, id int64, status models.TaskStatus, errMsg string) error

	// MarkCompleted marks a task as completed
	MarkCompleted(ctx context.Context, id int64) error

	// MarkFailed marks a task as failed with retry logic
	MarkFailed(ctx context.Context, id int64, errMsg string, shouldRetry bool) error

	// IncrementRetry increments the retry count
	IncrementRetry(ctx context.Context, id int64) error

	// Assign assigns a task to a session
	Assign(ctx context.Context, id int64, sessionID string) error

	// GetStuckTasks retrieves tasks stuck beyond timeout
	GetStuckTasks(ctx context.Context, timeoutMinutes int) ([]*models.Task, error)

	// GetMetrics retrieves task metrics
	GetMetrics(ctx context.Context, period time.Duration) (*models.TaskMetrics, error)

	// Delete removes a task (for cleanup)
	Delete(ctx context.Context, id int64) error

	// LogStatusChange records a task status change
	LogStatusChange(ctx context.Context, taskID int64, oldStatus, newStatus models.TaskStatus, reason string) error

	// GetStatusChangeLogs retrieves status change history for a task
	GetStatusChangeLogs(ctx context.Context, taskID int64) ([]*models.TaskLog, error)
}

// taskRepository implements TaskRepository
type taskRepository struct {
	store *storage.SQLiteStore
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(store *storage.SQLiteStore) TaskRepository {
	return &taskRepository{store: store}
}

func (r *taskRepository) Create(ctx context.Context, task *models.TaskCreate) (int64, error) {
	result, err := r.store.Exec(ctx,
		`INSERT INTO tasks (content, priority, status, target_session, timeout_minutes, metadata)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		task.Content,
		task.Priority,
		models.TaskPending,
		sql.NullString{String: task.TargetSession, Valid: task.TargetSession != ""},
		task.TimeoutMinutes,
		sql.NullString{String: string(task.Metadata), Valid: len(task.Metadata) > 0},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}
	return result.LastInsertId()
}

func (r *taskRepository) GetByID(ctx context.Context, id int64) (*models.Task, error) {
	row := r.store.QueryRow(ctx,
		`SELECT id, content, priority, status, target_session, created_at, updated_at,
		        completed_at, assigned_at, retry_count, max_retries, timeout_minutes, error_message, metadata
		 FROM tasks WHERE id = ?`,
		id)

	task := &models.Task{}
	err := row.Scan(
		&task.ID, &task.Content, &task.Priority, &task.Status, &task.TargetSession,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &task.AssignedAt,
		&task.RetryCount, &task.MaxRetries, &task.TimeoutMinutes, &task.ErrorMessage, &task.Metadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (r *taskRepository) GetPending(ctx context.Context, limit int) ([]*models.Task, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, content, priority, status, target_session, created_at, updated_at,
		        completed_at, assigned_at, retry_count, max_retries, timeout_minutes, error_message, metadata
		 FROM tasks
		 WHERE status = ? AND retry_count < max_retries
		 ORDER BY priority DESC, created_at ASC
		 LIMIT ?`,
		models.TaskPending, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.Content, &task.Priority, &task.Status, &task.TargetSession,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &task.AssignedAt,
			&task.RetryCount, &task.MaxRetries, &task.TimeoutMinutes, &task.ErrorMessage, &task.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *taskRepository) GetByStatus(ctx context.Context, status models.TaskStatus, limit int, offset int) ([]*models.Task, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, content, priority, status, target_session, created_at, updated_at,
		        completed_at, assigned_at, retry_count, max_retries, timeout_minutes, error_message, metadata
		 FROM tasks
		 WHERE status = ?
		 ORDER BY priority DESC, created_at DESC
		 LIMIT ? OFFSET ?`,
		status, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by status: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.Content, &task.Priority, &task.Status, &task.TargetSession,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &task.AssignedAt,
			&task.RetryCount, &task.MaxRetries, &task.TimeoutMinutes, &task.ErrorMessage, &task.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *taskRepository) GetBySession(ctx context.Context, sessionID string, status models.TaskStatus) ([]*models.Task, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, content, priority, status, target_session, created_at, updated_at,
		        completed_at, assigned_at, retry_count, max_retries, timeout_minutes, error_message, metadata
		 FROM tasks
		 WHERE target_session = ? AND status = ?
		 ORDER BY priority DESC`,
		sessionID, status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by session: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.Content, &task.Priority, &task.Status, &task.TargetSession,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &task.AssignedAt,
			&task.RetryCount, &task.MaxRetries, &task.TimeoutMinutes, &task.ErrorMessage, &task.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *taskRepository) UpdateStatus(ctx context.Context, id int64, status models.TaskStatus, reason string) error {
	// Get old status for audit log
	oldTask, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if oldTask == nil {
		return fmt.Errorf("task not found")
	}

	// Update status
	_, err = r.store.Exec(ctx,
		`UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Log the change
	return r.LogStatusChange(ctx, id, oldTask.Status, status, reason)
}

func (r *taskRepository) UpdateWithError(ctx context.Context, id int64, status models.TaskStatus, errMsg string) error {
	_, err := r.store.Exec(ctx,
		`UPDATE tasks SET status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, sql.NullString{String: errMsg, Valid: errMsg != ""}, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task with error: %w", err)
	}
	return nil
}

func (r *taskRepository) MarkCompleted(ctx context.Context, id int64) error {
	_, err := r.store.Exec(ctx,
		`UPDATE tasks SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		models.TaskCompleted, id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark task completed: %w", err)
	}
	return r.LogStatusChange(ctx, id, models.TaskInProgress, models.TaskCompleted, "Task completed successfully")
}

func (r *taskRepository) MarkFailed(ctx context.Context, id int64, errMsg string, shouldRetry bool) error {
	if !shouldRetry {
		return r.UpdateWithError(ctx, id, models.TaskFailed, errMsg)
	}

	// If should retry, mark as pending and increment retry count
	_, err := r.store.Exec(ctx,
		`UPDATE tasks SET status = ?, error_message = ?, retry_count = retry_count + 1,
		 updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		models.TaskPending,
		sql.NullString{String: errMsg, Valid: errMsg != ""},
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark task failed: %w", err)
	}
	return nil
}

func (r *taskRepository) IncrementRetry(ctx context.Context, id int64) error {
	_, err := r.store.Exec(ctx,
		`UPDATE tasks SET retry_count = retry_count + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}
	return nil
}

func (r *taskRepository) Assign(ctx context.Context, id int64, sessionID string) error {
	_, err := r.store.Exec(ctx,
		`UPDATE tasks SET target_session = ?, status = ?, assigned_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		sessionID, models.TaskAssigned, id,
	)
	if err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}
	return r.LogStatusChange(ctx, id, models.TaskPending, models.TaskAssigned, fmt.Sprintf("Assigned to session %s", sessionID))
}

func (r *taskRepository) GetStuckTasks(ctx context.Context, timeoutMinutes int) ([]*models.Task, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, content, priority, status, target_session, created_at, updated_at,
		        completed_at, assigned_at, retry_count, max_retries, timeout_minutes, error_message, metadata
		 FROM tasks
		 WHERE status IN (?, ?)
		 AND datetime(assigned_at, '+' || timeout_minutes || ' minutes') < datetime('now')
		 ORDER BY assigned_at ASC`,
		models.TaskAssigned, models.TaskInProgress,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query stuck tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.Content, &task.Priority, &task.Status, &task.TargetSession,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &task.AssignedAt,
			&task.RetryCount, &task.MaxRetries, &task.TimeoutMinutes, &task.ErrorMessage, &task.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *taskRepository) GetMetrics(ctx context.Context, period time.Duration) (*models.TaskMetrics, error) {
	// This is a placeholder - actual implementation would aggregate task data
	// For now, return a zero-value struct
	return &models.TaskMetrics{}, nil
}

func (r *taskRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.store.Exec(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *taskRepository) LogStatusChange(ctx context.Context, taskID int64, oldStatus, newStatus models.TaskStatus, reason string) error {
	_, err := r.store.Exec(ctx,
		`INSERT INTO task_log (task_id, old_status, new_status, reason) VALUES (?, ?, ?, ?)`,
		taskID, oldStatus, newStatus, reason,
	)
	if err != nil {
		return fmt.Errorf("failed to log status change: %w", err)
	}
	return nil
}

func (r *taskRepository) GetStatusChangeLogs(ctx context.Context, taskID int64) ([]*models.TaskLog, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, task_id, old_status, new_status, reason, timestamp FROM task_log WHERE task_id = ? ORDER BY timestamp DESC`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query task logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.TaskLog
	for rows.Next() {
		log := &models.TaskLog{}
		err := rows.Scan(&log.ID, &log.TaskID, &log.OldStatus, &log.NewStatus, &log.Reason, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
