package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// SessionState represents the current state of a session
type SessionState string

const (
	SessionIdle          SessionState = "idle"
	SessionBusy          SessionState = "busy"
	SessionWaitingInput  SessionState = "waiting_input"
	SessionBlocked       SessionState = "blocked"
	SessionUnhealthy     SessionState = "unhealthy"
	SessionTerminated    SessionState = "terminated"
)

// SessionType represents the type of GAIA session
type SessionType string

const (
	SessionTypeInteractive SessionType = "interactive"
	SessionTypeWorker      SessionType = "worker"
	SessionTypeBatch       SessionType = "batch"
)

// Session represents a GAIA-managed session
type Session struct {
	ID            int64          `json:"id"`
	SessionID     string         `json:"session_id"`
	SessionType   SessionType    `json:"session_type"`
	Provider      string         `json:"provider"`
	Status        SessionState   `json:"status"`
	CurrentTaskID sql.NullInt64  `json:"current_task_id"`
	LastHeartbeat time.Time      `json:"last_heartbeat"`
	HealthScore   float64        `json:"health_score"`
	MetricsJSON   sql.NullString `json:"metrics_json"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Active        bool           `json:"active"`
}

// SessionCreate represents input for creating a new session
type SessionCreate struct {
	SessionID   string      `json:"session_id"`
	SessionType SessionType `json:"session_type"`
	Provider    string      `json:"provider"`
	Status      SessionState `json:"status"`
}

// SessionUpdate represents updates to a session
type SessionUpdate struct {
	Status      SessionState `json:"status,omitempty"`
	HealthScore float64      `json:"health_score,omitempty"`
	Active      bool         `json:"active,omitempty"`
}

// SessionHealth represents health metrics for a session
type SessionHealth struct {
	SessionID     string    `json:"session_id"`
	HealthScore   float64   `json:"health_score"`
	Status        SessionState `json:"status"`
	IsHealthy     bool      `json:"is_healthy"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	TimeSinceLastHeartbeat int64 `json:"time_since_last_heartbeat_sec"`
	ActiveTaskID  *int64    `json:"active_task_id,omitempty"`
}

// SessionStateLog represents a session state change event
type SessionStateLog struct {
	ID        int64        `json:"id"`
	SessionID string       `json:"session_id"`
	OldState  SessionState `json:"old_state"`
	NewState  SessionState `json:"new_state"`
	Reason    string       `json:"reason"`
	Timestamp time.Time    `json:"timestamp"`
}

// SessionMetrics represents performance metrics for a session
type SessionMetrics struct {
	SessionID              string    `json:"session_id"`
	TotalTasks             int       `json:"total_tasks"`
	CompletedTasks         int       `json:"completed_tasks"`
	FailedTasks            int       `json:"failed_tasks"`
	ActiveTasks            int       `json:"active_tasks"`
	AverageCompletionTime  float64   `json:"average_completion_time_sec"`
	LastTaskCompletedAt    *time.Time `json:"last_task_completed_at"`
	ConsecutiveFailures    int       `json:"consecutive_failures"`
	SuccessRate            float64   `json:"success_rate"`
}

// IsHealthy determines if session is in a healthy state
func (s SessionState) IsHealthy() bool {
	switch s {
	case SessionIdle, SessionBusy, SessionWaitingInput:
		return true
	default:
		return false
	}
}

// IsTerminal checks if session is in a terminal state
func (s SessionState) IsTerminal() bool {
	return s == SessionTerminated
}

// String returns string representation
func (s SessionState) String() string {
	return string(s)
}

// String returns string representation
func (st SessionType) String() string {
	return string(st)
}

// MarshalJSON converts SessionState to JSON
func (s SessionState) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON converts JSON to SessionState
func (s *SessionState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = SessionState(str)
	return nil
}

// MarshalJSON converts SessionType to JSON
func (st SessionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(st))
}

// UnmarshalJSON converts JSON to SessionType
func (st *SessionType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*st = SessionType(str)
	return nil
}
