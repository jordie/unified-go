// Package repository defines data access interfaces and implementations
package repository

import (
	"unified-go/internal/storage"
)

// Manager provides access to all repositories
type Manager interface {
	Tasks() TaskRepository
	Sessions() SessionRepository
	Locks() LockRepository
	Metrics() MetricsRepository
}

// manager implements Manager
type manager struct {
	tasks    TaskRepository
	sessions SessionRepository
	locks    LockRepository
	metrics  MetricsRepository
}

// NewManager creates a new repository manager
func NewManager(store *storage.SQLiteStore) Manager {
	return &manager{
		tasks:    NewTaskRepository(store),
		sessions: NewSessionRepository(store),
		locks:    NewLockRepository(store),
		metrics:  NewMetricsRepository(store),
	}
}

func (m *manager) Tasks() TaskRepository {
	return m.tasks
}

func (m *manager) Sessions() SessionRepository {
	return m.sessions
}

func (m *manager) Locks() LockRepository {
	return m.locks
}

func (m *manager) Metrics() MetricsRepository {
	return m.metrics
}

/*
Repository Package Overview:

This package provides a data access layer for the GAIA_HOME orchestrator.
It implements the repository pattern to abstract database operations.

Repositories:

1. TaskRepository
   - Interface for task queue operations
   - CRUD operations on tasks
   - Status transitions and retry logic
   - Task logging and audit trail

2. SessionRepository
   - Interface for session management
   - Session lifecycle (create, update, deactivate)
   - Health monitoring and metrics
   - Session state change logging

3. LockRepository
   - Interface for distributed locking
   - Lock acquisition and release
   - Lock expiration and cleanup
   - Lock contention metrics

4. MetricsRepository
   - Interface for metrics storage and retrieval
   - Time series data collection
   - Aggregated statistics
   - System health calculation

Manager:
The Manager interface provides unified access to all repositories,
enabling dependency injection and simplified client code.

Usage Example:

	// Create repository manager
	repos := repository.NewManager(store)

	// Access specific repository
	taskRepo := repos.Tasks()

	// Perform operations
	taskID, err := taskRepo.Create(ctx, &models.TaskCreate{...})
	task, err := taskRepo.GetByID(ctx, taskID)
	err := taskRepo.UpdateStatus(ctx, taskID, models.TaskCompleted, "finished")

All repository methods take a context parameter for cancellation support
and follow Go best practices for error handling.
*/
