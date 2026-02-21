package models

import (
	"database/sql"
	"time"
)

// Lock represents a distributed lock for coordination
type Lock struct {
	ID            int64     `json:"id"`
	LockID        string    `json:"lock_id"`
	HolderID      string    `json:"holder_id"`
	ExpiresAt     time.Time `json:"expires_at"`
	PriorityLevel int       `json:"priority_level"`
	AcquiredAt    time.Time `json:"acquired_at"`
	ReleasedAt    *time.Time `json:"released_at"`
}

// LockCreate represents input for acquiring a lock
type LockCreate struct {
	LockID        string    `json:"lock_id"`
	HolderID      string    `json:"holder_id"`
	ExpiresAt     time.Time `json:"expires_at"`
	PriorityLevel int       `json:"priority_level"`
}

// LockStatus represents the status of a lock
type LockStatus struct {
	LockID        string     `json:"lock_id"`
	HolderID      string     `json:"holder_id"`
	IsLocked      bool       `json:"is_locked"`
	ExpiresAt     time.Time  `json:"expires_at"`
	TimeRemaining int64      `json:"time_remaining_sec"`
	AcquiredAt    time.Time  `json:"acquired_at"`
	ReleasedAt    *time.Time `json:"released_at"`
	PriorityLevel int        `json:"priority_level"`
}

// IsExpired checks if the lock has expired
func (l *Lock) IsExpired() bool {
	if l.ReleasedAt != nil {
		return true
	}
	return time.Now().After(l.ExpiresAt)
}

// TimeRemaining returns seconds until lock expiration
func (l *Lock) TimeRemaining() int64 {
	if l.IsExpired() {
		return 0
	}
	return int64(l.ExpiresAt.Sub(time.Now()).Seconds())
}

// ToLockStatus converts Lock to LockStatus
func (l *Lock) ToLockStatus() LockStatus {
	return LockStatus{
		LockID:        l.LockID,
		HolderID:      l.HolderID,
		IsLocked:      !l.IsExpired(),
		ExpiresAt:     l.ExpiresAt,
		TimeRemaining: l.TimeRemaining(),
		AcquiredAt:    l.AcquiredAt,
		ReleasedAt:    l.ReleasedAt,
		PriorityLevel: l.PriorityLevel,
	}
}

// LockWaitResult represents the result of waiting for a lock
type LockWaitResult struct {
	LockID      string `json:"lock_id"`
	Acquired    bool   `json:"acquired"`
	WaitedMS    int64  `json:"waited_ms"`
	HolderID    string `json:"holder_id"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// LockMetrics represents metrics about locks
type LockMetrics struct {
	TotalLocks           int64   `json:"total_locks"`
	ActiveLocks          int64   `json:"active_locks"`
	ExpiredLocks         int64   `json:"expired_locks"`
	ReleasedLocks        int64   `json:"released_locks"`
	AverageHoldTimeMS    float64 `json:"average_hold_time_ms"`
	AverageWaitTimeMS    float64 `json:"average_wait_time_ms"`
	ContentionRatio      float64 `json:"contention_ratio"`
}

// RateLimiterState represents rate limiting state per session
type RateLimiterState struct {
	ID                   int64     `json:"id"`
	SessionID            string    `json:"session_id"`
	LastAssignmentTime   *time.Time `json:"last_assignment_time"`
	ActiveTaskCount      int       `json:"active_task_count"`
	TotalAssignments     int       `json:"total_assignments"`
	TotalCompletions     int       `json:"total_completions"`
	ThrottledCount       int       `json:"throttled_count"`
	LastThrottled        *time.Time `json:"last_throttled"`
}

// RateLimitStatus represents current rate limit status
type RateLimitStatus struct {
	SessionID              string    `json:"session_id"`
	CanAssign              bool      `json:"can_assign"`
	TimeSinceLastAssignment int64    `json:"time_since_last_assignment_sec"`
	SecondsSinceThrottle   int64     `json:"seconds_since_throttle"`
	ActiveTasks            int       `json:"active_tasks"`
	TotalAssignments       int       `json:"total_assignments"`
	ThrottleEvents         int       `json:"throttle_events"`
}
