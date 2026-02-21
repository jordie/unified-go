package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a single database migration
type Migration struct {
	Version     string
	Description string
	UpSQL       string
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	store *SQLiteStore
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(store *SQLiteStore) *MigrationRunner {
	return &MigrationRunner{
		store: store,
	}
}

// Initialize runs all pending migrations
func (mr *MigrationRunner) Initialize(ctx context.Context, migrations []Migration) error {
	// Create schema_migrations table if it doesn't exist
	_, err := mr.store.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL UNIQUE,
			description TEXT,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Get applied migrations
	applied, err := mr.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; !ok {
			fmt.Printf("Applying migration: %s (%s)\n", migration.Version, migration.Description)

			if err := mr.runMigration(ctx, &migration); err != nil {
				return fmt.Errorf("migration %s failed: %w", migration.Version, err)
			}

			applied[migration.Version] = true
		}
	}

	return nil
}

// RunFromDirectory loads and runs migrations from a directory
func (mr *MigrationRunner) RunFromDirectory(ctx context.Context, dirPath string) error {
	// Read migration files
	entries, err := fs.ReadDir(nil, dirPath)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var migrations []Migration

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .sql files
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Extract version from filename (e.g., 001_initial_schema.sql)
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

		// Read migration SQL
		filePath := filepath.Join(dirPath, entry.Name())
		sqlBytes, err := fs.ReadFile(nil, filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
		}

		sql := string(sqlBytes)

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			UpSQL:       sql,
		})
	}

	// Run migrations
	return mr.Initialize(ctx, migrations)
}

// runMigration executes a single migration within a transaction
func (mr *MigrationRunner) runMigration(ctx context.Context, migration *Migration) error {
	return mr.store.Transaction(ctx, func(tx *sql.Tx) error {
		// Execute migration SQL
		// Split by semicolon to execute multiple statements
		statements := strings.Split(migration.UpSQL, ";")

		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("failed to execute statement: %w", err)
			}
		}

		// Record migration
		_, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, description) VALUES (?, ?)`,
			migration.Version, migration.Description)

		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		return nil
	})
}

// getAppliedMigrations returns a map of applied migration versions
func (mr *MigrationRunner) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	rows, err := mr.store.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		// Table might not exist yet
		if strings.Contains(err.Error(), "no such table") {
			return make(map[string]bool), nil
		}
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading migrations: %w", err)
	}

	return applied, nil
}

// GetMigrationStatus returns the status of all migrations
func (mr *MigrationRunner) GetMigrationStatus(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := mr.store.Query(ctx,
		`SELECT id, version, description, applied_at FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %w", err)
	}
	defer rows.Close()

	var status []map[string]interface{}

	for rows.Next() {
		var id int64
		var version string
		var description string
		var appliedAt string

		if err := rows.Scan(&id, &version, &description, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration status: %w", err)
		}

		status = append(status, map[string]interface{}{
			"id":         id,
			"version":    version,
			"description": description,
			"applied_at": appliedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading migration status: %w", err)
	}

	return status, nil
}

// Rollback would rollback a migration (not implemented for safety)
// For production use, rollbacks should be done with separate "down" migrations
func (mr *MigrationRunner) VerifySchema(ctx context.Context) error {
	// Check that all required tables exist
	requiredTables := []string{
		"tasks",
		"sessions",
		"locks",
		"metrics",
		"schema_migrations",
	}

	for _, table := range requiredTables {
		row := mr.store.QueryRow(ctx,
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table)

		var name string
		if err := row.Scan(&name); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("required table not found: %s", table)
			}
			return fmt.Errorf("failed to verify table %s: %w", table, err)
		}
	}

	return nil
}

// CreateDefaultMigrations returns the default set of migrations
func CreateDefaultMigrations() []Migration {
	return []Migration{
		{
			Version:     "001",
			Description: "initial_schema",
			UpSQL: `
-- Tasks Table
CREATE TABLE IF NOT EXISTS tasks (
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
);

-- Sessions Table
CREATE TABLE IF NOT EXISTS sessions (
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
);

-- Locks Table
CREATE TABLE IF NOT EXISTS locks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	lock_id TEXT NOT NULL UNIQUE,
	holder_id TEXT NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	priority_level INTEGER DEFAULT 5,
	acquired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	released_at TIMESTAMP
);

-- Metrics Table
CREATE TABLE IF NOT EXISTS metrics (
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
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_target_session ON tasks(target_session);
CREATE INDEX IF NOT EXISTS idx_sessions_provider ON sessions(provider);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_locks_holder_id ON locks(holder_id);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
			`,
		},
	}
}
