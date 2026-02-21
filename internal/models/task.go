package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// TaskStatus represents the state of a task in the queue
type TaskStatus string

const (
	TaskPending           TaskStatus = "pending"
	TaskAssigned          TaskStatus = "assigned"
	TaskInProgress        TaskStatus = "in_progress"
	TaskWaitingCompletion TaskStatus = "waiting_completion"
	TaskCompleted         TaskStatus = "completed"
	TaskFailed            TaskStatus = "failed"
	TaskCancelled         TaskStatus = "cancelled"
	TaskStuck             TaskStatus = "stuck"
)

// TaskPriority represents task priority level
type TaskPriority int

const (
	PriorityLowest  TaskPriority = 1
	PriorityLow     TaskPriority = 3
	PriorityNormal  TaskPriority = 5
	PriorityHigh    TaskPriority = 7
	PriorityUrgent  TaskPriority = 10
)

// Task represents a work item in the queue
type Task struct {
	ID              int64          `json:"id"`
	Content         string         `json:"content"`
	Priority        TaskPriority   `json:"priority"`
	Status          TaskStatus     `json:"status"`
	TargetSession   sql.NullString `json:"target_session"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CompletedAt     *time.Time     `json:"completed_at"`
	AssignedAt      *time.Time     `json:"assigned_at"`
	RetryCount      int            `json:"retry_count"`
	MaxRetries      int            `json:"max_retries"`
	TimeoutMinutes  int            `json:"timeout_minutes"`
	ErrorMessage    sql.NullString `json:"error_message"`
	Metadata        sql.NullString `json:"metadata"`
}

// TaskCreate represents the input for creating a new task
type TaskCreate struct {
	Content        string        `json:"content"`
	Priority       TaskPriority  `json:"priority"`
	TargetSession  string        `json:"target_session,omitempty"`
	TimeoutMinutes int           `json:"timeout_minutes"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// TaskUpdate represents updates to a task
type TaskUpdate struct {
	Status       TaskStatus `json:"status,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// TaskLog represents a task state change event
type TaskLog struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// IsTerminal checks if the task status is terminal
func (s TaskStatus) IsTerminal() bool {
	switch s {
	case TaskCompleted, TaskFailed, TaskCancelled, TaskStuck:
		return true
	default:
		return false
	}
}

// String returns string representation
func (p TaskPriority) String() string {
	switch p {
	case PriorityLowest:
		return "lowest"
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return "unknown"
	}
}

// MarshalJSON converts TaskStatus to JSON
func (s TaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON converts JSON to TaskStatus
func (s *TaskStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = TaskStatus(str)
	return nil
}

// MarshalJSON converts TaskPriority to JSON
func (p TaskPriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(p))
}

// UnmarshalJSON converts JSON to TaskPriority
func (p *TaskPriority) UnmarshalJSON(data []byte) error {
	var num int
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*p = TaskPriority(num)
	return nil
}
