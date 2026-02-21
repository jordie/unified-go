package repository

import (
	"context"
	"fmt"
	"time"

	"unified-go/internal/models"
	"unified-go/internal/storage"
)

// LockRepository defines the interface for lock data access
type LockRepository interface {
	// AcquireLock acquires a distributed lock
	AcquireLock(ctx context.Context, lockID, holderID string, duration time.Duration, priority int) (*models.Lock, error)

	// ReleaseLock releases a held lock
	ReleaseLock(ctx context.Context, lockID, holderID string) error

	// GetLock retrieves a lock by ID
	GetLock(ctx context.Context, lockID string) (*models.Lock, error)

	// GetLockStatus checks if a lock is held or expired
	GetLockStatus(ctx context.Context, lockID string) (*models.LockStatus, error)

	// IsLockAvailable checks if a lock can be acquired
	IsLockAvailable(ctx context.Context, lockID string) (bool, error)

	// WaitForLock waits for a lock to become available
	WaitForLock(ctx context.Context, lockID, holderID string, timeout time.Duration) (*models.LockWaitResult, error)

	// CleanupExpiredLocks removes expired locks
	CleanupExpiredLocks(ctx context.Context) (int64, error)

	// GetHeldLocks retrieves all locks held by a holder
	GetHeldLocks(ctx context.Context, holderID string) ([]*models.Lock, error)

	// GetLockMetrics retrieves lock metrics
	GetLockMetrics(ctx context.Context) (*models.LockMetrics, error)
}

// lockRepository implements LockRepository
type lockRepository struct {
	store *storage.SQLiteStore
}

// NewLockRepository creates a new lock repository
func NewLockRepository(store *storage.SQLiteStore) LockRepository {
	return &lockRepository{store: store}
}

func (r *lockRepository) AcquireLock(ctx context.Context, lockID, holderID string, duration time.Duration, priority int) (*models.Lock, error) {
	expiresAt := time.Now().Add(duration)

	err := r.store.Transaction(ctx, func(tx interface{}) error {
		// Check if lock exists and is still valid
		row := r.store.QueryRow(ctx,
			`SELECT id FROM locks WHERE lock_id = ? AND released_at IS NULL AND expires_at > datetime('now')`,
			lockID)

		var existingID int64
		err := row.Scan(&existingID)
		if err == nil {
			// Lock is held by someone else
			return fmt.Errorf("lock already held")
		}

		// Acquire or reacquire the lock
		_, err = r.store.Exec(ctx,
			`INSERT OR REPLACE INTO locks (lock_id, holder_id, expires_at, priority_level, acquired_at)
			 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			lockID, holderID, expiresAt, priority)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return the acquired lock
	row := r.store.QueryRow(ctx,
		`SELECT id, lock_id, holder_id, expires_at, priority_level, acquired_at, released_at FROM locks WHERE lock_id = ?`,
		lockID)

	lock := &models.Lock{}
	err = row.Scan(&lock.ID, &lock.LockID, &lock.HolderID, &lock.ExpiresAt, &lock.PriorityLevel, &lock.AcquiredAt, &lock.ReleasedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve acquired lock: %w", err)
	}

	return lock, nil
}

func (r *lockRepository) ReleaseLock(ctx context.Context, lockID, holderID string) error {
	_, err := r.store.Exec(ctx,
		`UPDATE locks SET released_at = CURRENT_TIMESTAMP WHERE lock_id = ? AND holder_id = ?`,
		lockID, holderID)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

func (r *lockRepository) GetLock(ctx context.Context, lockID string) (*models.Lock, error) {
	row := r.store.QueryRow(ctx,
		`SELECT id, lock_id, holder_id, expires_at, priority_level, acquired_at, released_at FROM locks WHERE lock_id = ?`,
		lockID)

	lock := &models.Lock{}
	err := row.Scan(&lock.ID, &lock.LockID, &lock.HolderID, &lock.ExpiresAt, &lock.PriorityLevel, &lock.AcquiredAt, &lock.ReleasedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get lock: %w", err)
	}

	return lock, nil
}

func (r *lockRepository) GetLockStatus(ctx context.Context, lockID string) (*models.LockStatus, error) {
	lock, err := r.GetLock(ctx, lockID)
	if err != nil {
		return nil, err
	}

	if lock == nil {
		return &models.LockStatus{
			LockID:   lockID,
			IsLocked: false,
		}, nil
	}

	return &models.LockStatus{
		LockID:        lock.LockID,
		HolderID:      lock.HolderID,
		IsLocked:      !lock.IsExpired(),
		ExpiresAt:     lock.ExpiresAt,
		TimeRemaining: lock.TimeRemaining(),
		AcquiredAt:    lock.AcquiredAt,
		ReleasedAt:    lock.ReleasedAt,
		PriorityLevel: lock.PriorityLevel,
	}, nil
}

func (r *lockRepository) IsLockAvailable(ctx context.Context, lockID string) (bool, error) {
	status, err := r.GetLockStatus(ctx, lockID)
	if err != nil {
		return false, err
	}

	return !status.IsLocked, nil
}

func (r *lockRepository) WaitForLock(ctx context.Context, lockID, holderID string, timeout time.Duration) (*models.LockWaitResult, error) {
	startTime := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return &models.LockWaitResult{
				LockID:   lockID,
				Acquired: false,
				WaitedMS: int64(time.Since(startTime).Milliseconds()),
			}, ctx.Err()

		case <-time.After(timeout):
			return &models.LockWaitResult{
				LockID:   lockID,
				Acquired: false,
				WaitedMS: int64(timeout.Milliseconds()),
			}, fmt.Errorf("lock acquisition timeout")

		case <-ticker.C:
			available, err := r.IsLockAvailable(ctx, lockID)
			if err != nil {
				return nil, err
			}

			if available {
				lock, err := r.AcquireLock(ctx, lockID, holderID, timeout, 5)
				if err == nil && lock != nil {
					return &models.LockWaitResult{
						LockID:    lockID,
						Acquired:  true,
						WaitedMS:  int64(time.Since(startTime).Milliseconds()),
						HolderID:  holderID,
						ExpiresAt: lock.ExpiresAt,
					}, nil
				}
			}
		}
	}
}

func (r *lockRepository) CleanupExpiredLocks(ctx context.Context) (int64, error) {
	result, err := r.store.Exec(ctx,
		`DELETE FROM locks WHERE released_at IS NOT NULL AND expires_at < datetime('now')`)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired locks: %w", err)
	}

	return result.RowsAffected()
}

func (r *lockRepository) GetHeldLocks(ctx context.Context, holderID string) ([]*models.Lock, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, lock_id, holder_id, expires_at, priority_level, acquired_at, released_at
		 FROM locks WHERE holder_id = ? AND released_at IS NULL AND expires_at > datetime('now')
		 ORDER BY acquired_at DESC`,
		holderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query held locks: %w", err)
	}
	defer rows.Close()

	var locks []*models.Lock
	for rows.Next() {
		lock := &models.Lock{}
		err := rows.Scan(&lock.ID, &lock.LockID, &lock.HolderID, &lock.ExpiresAt, &lock.PriorityLevel, &lock.AcquiredAt, &lock.ReleasedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lock: %w", err)
		}
		locks = append(locks, lock)
	}

	return locks, rows.Err()
}

func (r *lockRepository) GetLockMetrics(ctx context.Context) (*models.LockMetrics, error) {
	row := r.store.QueryRow(ctx, `
		SELECT
			COUNT(*) as total_locks,
			SUM(CASE WHEN released_at IS NULL AND expires_at > datetime('now') THEN 1 ELSE 0 END) as active_locks,
			SUM(CASE WHEN expires_at < datetime('now') THEN 1 ELSE 0 END) as expired_locks,
			SUM(CASE WHEN released_at IS NOT NULL THEN 1 ELSE 0 END) as released_locks,
			AVG(CAST((julianday(released_at) - julianday(acquired_at)) * 86400000 AS REAL)) as avg_hold_time_ms,
			MAX(CAST((julianday('now') - julianday(acquired_at)) * 86400000 AS REAL)) as max_wait_time_ms
		FROM locks`)

	metrics := &models.LockMetrics{}
	err := row.Scan(
		&metrics.TotalLocks,
		&metrics.ActiveLocks,
		&metrics.ExpiredLocks,
		&metrics.ReleasedLocks,
		&metrics.AverageHoldTimeMS,
		&metrics.AverageWaitTimeMS,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get lock metrics: %w", err)
	}

	// Calculate contention ratio
	if metrics.TotalLocks > 0 {
		metrics.ContentionRatio = float64(metrics.ReleasedLocks) / float64(metrics.TotalLocks)
	}

	return metrics, nil
}
