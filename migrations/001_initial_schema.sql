-- GAIA_HOME Go Rewrite: Initial Schema
-- Migration: 001_initial_schema
-- Purpose: Create core tables for task management, sessions, locks, and metrics

-- Tasks Table: Core task queue storage
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

-- Sessions Table: Active sessions (GAIA-managed)
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

-- Locks Table: Distributed locking mechanism
CREATE TABLE IF NOT EXISTS locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lock_id TEXT NOT NULL UNIQUE,
    holder_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    priority_level INTEGER DEFAULT 5,
    acquired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP
);

-- Metrics Table: Performance metrics and statistics
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

-- Session State Log Table: Track session state changes
CREATE TABLE IF NOT EXISTS session_state_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    old_state TEXT,
    new_state TEXT,
    reason TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Task Log Table: Track task status changes
CREATE TABLE IF NOT EXISTS task_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    old_status TEXT,
    new_status TEXT,
    reason TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(task_id) REFERENCES tasks(id)
);

-- Rate Limiter State Table: Track assignment rate limiting
CREATE TABLE IF NOT EXISTS rate_limiter_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL UNIQUE,
    last_assignment_time TIMESTAMP,
    active_task_count INTEGER DEFAULT 0,
    total_assignments INTEGER DEFAULT 0,
    total_completions INTEGER DEFAULT 0,
    throttled_count INTEGER DEFAULT 0,
    last_throttled TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_target_session ON tasks(target_session);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_sessions_provider ON sessions(provider);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(active);
CREATE INDEX IF NOT EXISTS idx_locks_holder_id ON locks(holder_id);
CREATE INDEX IF NOT EXISTS idx_locks_expires_at ON locks(expires_at);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_task_log_task_id ON task_log(task_id);
CREATE INDEX IF NOT EXISTS idx_session_state_log_session_id ON session_state_log(session_id);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL UNIQUE,
    description TEXT,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Record this migration
INSERT OR IGNORE INTO schema_migrations (version, description)
VALUES ('001_initial_schema', 'Create core tables for tasks, sessions, locks, metrics');
