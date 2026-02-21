package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds SQLite configuration
type Config struct {
	DatabasePath      string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	BusyTimeout       time.Duration
	LogQueries        bool
}

// SQLiteStore manages SQLite database connections with pooling
type SQLiteStore struct {
	db     *sql.DB
	config Config
	mu     sync.RWMutex
}

// NewSQLiteStore creates a new SQLite store with connection pooling
func NewSQLiteStore(config Config) (*SQLiteStore, error) {
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 25 // Default max connections
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Hour
	}
	if config.BusyTimeout == 0 {
		config.BusyTimeout = 30 * time.Second
	}

	// Open database with SQLite-specific pragmas
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=%d",
		config.DatabasePath,
		int(config.BusyTimeout.Milliseconds()))

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable pragmas for performance
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000",
		"PRAGMA temp_store = MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	return &SQLiteStore{
		db:     db,
		config: config,
	}, nil
}

// Initialize creates tables from migration schema
func (s *SQLiteStore) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create tables using schema
	queries := []string{
		// Tasks Table
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			priority INTEGER NOT NULL DEFAULT 5,
			status TEXT NOT NULL DEFAULT 'pending',
			target_session TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			assigned_at TIMESTAMP,
			retry_count INTEGER DEFAULT 0,
			max_retries INTEGER DEFAULT 3,
			timeout_minutes INTEGER DEFAULT 30,
			error_message TEXT,
			metadata TEXT
		)`,

		// Sessions Table
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL UNIQUE,
			session_type TEXT NOT NULL,
			provider TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'idle',
			current_task_id INTEGER,
			last_heartbeat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			health_score REAL DEFAULT 100.0,
			metrics_json TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			active INTEGER DEFAULT 1,
			FOREIGN KEY(current_task_id) REFERENCES tasks(id)
		)`,

		// Locks Table
		`CREATE TABLE IF NOT EXISTS locks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			lock_id TEXT NOT NULL UNIQUE,
			holder_id TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			priority_level INTEGER DEFAULT 5,
			acquired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			released_at TIMESTAMP
		)`,

		// Metrics Table
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			task_throughput INTEGER,
			avg_latency_ms REAL,
			success_rate REAL,
			memory_mb REAL,
			cpu_percent REAL,
			active_sessions INTEGER,
			pending_tasks INTEGER,
			metadata TEXT
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_target_session ON tasks(target_session)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_provider ON sessions(provider)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_locks_holder_id ON locks(holder_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp)`,
	}

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// Query executes a SELECT query and returns rows
func (s *SQLiteStore) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config.LogQueries {
		fmt.Printf("[QUERY] %s (args: %v)\n", query, args)
	}

	return s.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a SELECT query and returns a single row
func (s *SQLiteStore) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config.LogQueries {
		fmt.Printf("[QUERY] %s (args: %v)\n", query, args)
	}

	return s.db.QueryRowContext(ctx, query, args...)
}

// Exec executes an INSERT/UPDATE/DELETE query
func (s *SQLiteStore) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config.LogQueries {
		fmt.Printf("[EXEC] %s (args: %v)\n", query, args)
	}

	return s.db.ExecContext(ctx, query, args...)
}

// Transaction executes a function within a database transaction
func (s *SQLiteStore) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a wrapper that allows operations within the transaction
	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed with error %v and rollback failed with %v", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection pool
func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Health checks the health of the database connection
func (s *SQLiteStore) Health(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db.PingContext(ctx)
}

// Stats returns connection pool statistics
func (s *SQLiteStore) Stats() sql.DBStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db.Stats()
}

// TaskRow represents a task in the database
type TaskRow struct {
	ID             int64
	Content        string
	Priority       int
	Status         string
	TargetSession  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CompletedAt    *time.Time
	AssignedAt     *time.Time
	RetryCount     int
	MaxRetries     int
	TimeoutMinutes int
	ErrorMessage   string
	Metadata       string
}

// SessionRow represents a session in the database
type SessionRow struct {
	ID            int64
	SessionID     string
	SessionType   string
	Provider      string
	Status        string
	CurrentTaskID *int64
	LastHeartbeat time.Time
	HealthScore   float64
	MetricsJSON   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Active        bool
}

// LockRow represents a lock in the database
type LockRow struct {
	ID            int64
	LockID        string
	HolderID      string
	ExpiresAt     time.Time
	PriorityLevel int
	AcquiredAt    time.Time
	ReleasedAt    *time.Time
}

// MetricsRow represents metrics in the database
type MetricsRow struct {
	ID               int64
	Timestamp        time.Time
	TaskThroughput   *int
	AvgLatencyMS     *float64
	SuccessRate      *float64
	MemoryMB         *float64
	CPUPercent       *float64
	ActiveSessions   *int
	PendingTasks     *int
	Metadata         string
}

// Helper methods for common operations

// InsertTask inserts a new task
func (s *SQLiteStore) InsertTask(ctx context.Context, task *TaskRow) (int64, error) {
	result, err := s.Exec(ctx,
		`INSERT INTO tasks (content, priority, status, target_session, timeout_minutes, metadata)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		task.Content, task.Priority, task.Status, task.TargetSession, task.TimeoutMinutes, task.Metadata)

	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}

	return result.LastInsertId()
}

// UpdateTaskStatus updates a task's status
func (s *SQLiteStore) UpdateTaskStatus(ctx context.Context, taskID int64, status string) error {
	_, err := s.Exec(ctx,
		`UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, taskID)

	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// GetTask retrieves a task by ID
func (s *SQLiteStore) GetTask(ctx context.Context, taskID int64) (*TaskRow, error) {
	row := s.QueryRow(ctx,
		`SELECT id, content, priority, status, target_session, created_at, updated_at,
		        completed_at, assigned_at, retry_count, max_retries, timeout_minutes, error_message, metadata
		 FROM tasks WHERE id = ?`,
		taskID)

	task := &TaskRow{}
	err := row.Scan(&task.ID, &task.Content, &task.Priority, &task.Status, &task.TargetSession,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &task.AssignedAt,
		&task.RetryCount, &task.MaxRetries, &task.TimeoutMinutes, &task.ErrorMessage, &task.Metadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// InsertSession inserts a new session
func (s *SQLiteStore) InsertSession(ctx context.Context, session *SessionRow) (int64, error) {
	result, err := s.Exec(ctx,
		`INSERT INTO sessions (session_id, session_type, provider, status, health_score, metrics_json)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		session.SessionID, session.SessionType, session.Provider, session.Status, session.HealthScore, session.MetricsJSON)

	if err != nil {
		return 0, fmt.Errorf("failed to insert session: %w", err)
	}

	return result.LastInsertId()
}

// GetSession retrieves a session by ID
func (s *SQLiteStore) GetSession(ctx context.Context, sessionID string) (*SessionRow, error) {
	row := s.QueryRow(ctx,
		`SELECT id, session_id, session_type, provider, status, current_task_id, last_heartbeat,
		        health_score, metrics_json, created_at, updated_at, active
		 FROM sessions WHERE session_id = ?`,
		sessionID)

	session := &SessionRow{}
	err := row.Scan(&session.ID, &session.SessionID, &session.SessionType, &session.Provider,
		&session.Status, &session.CurrentTaskID, &session.LastHeartbeat, &session.HealthScore,
		&session.MetricsJSON, &session.CreatedAt, &session.UpdatedAt, &session.Active)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// InsertLock inserts a new lock
func (s *SQLiteStore) InsertLock(ctx context.Context, lock *LockRow) (int64, error) {
	result, err := s.Exec(ctx,
		`INSERT INTO locks (lock_id, holder_id, expires_at, priority_level)
		 VALUES (?, ?, ?, ?)`,
		lock.LockID, lock.HolderID, lock.ExpiresAt, lock.PriorityLevel)

	if err != nil {
		return 0, fmt.Errorf("failed to insert lock: %w", err)
	}

	return result.LastInsertId()
}

// InsertMetrics inserts a metrics record
func (s *SQLiteStore) InsertMetrics(ctx context.Context, metrics *MetricsRow) (int64, error) {
	result, err := s.Exec(ctx,
		`INSERT INTO metrics (task_throughput, avg_latency_ms, success_rate, memory_mb, cpu_percent, active_sessions, pending_tasks, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		metrics.TaskThroughput, metrics.AvgLatencyMS, metrics.SuccessRate, metrics.MemoryMB,
		metrics.CPUPercent, metrics.ActiveSessions, metrics.PendingTasks, metrics.Metadata)

	if err != nil {
		return 0, fmt.Errorf("failed to insert metrics: %w", err)
	}

	return result.LastInsertId()
}
