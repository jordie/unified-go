// Package models defines domain models for the GAIA_HOME orchestrator
package models

// This file serves as the package documentation and re-exports key types

/*
models package provides:

1. Task Management (task.go)
   - Task: Core task queue item
   - TaskStatus: Enum for task states
   - TaskPriority: Enum for priority levels
   - TaskLog: State change audit trail

2. Session Management (session.go)
   - Session: GAIA-managed session representation
   - SessionState: Enum for session states
   - SessionType: Enum for session types
   - SessionMetrics: Performance metrics per session
   - RateLimiterState: Rate limiting state

3. Distributed Locking (lock.go)
   - Lock: Distributed lock for coordination
   - LockStatus: Current lock status
   - RateLimiterState: Rate limiting state per session

4. System Metrics (metrics.go)
   - Metrics: System performance metrics
   - MetricsAggregate: Aggregated metrics over time
   - SystemHealth: Overall system health status
   - PerformanceReport: Detailed performance analysis
   - AlertThreshold: Alert configuration

5. Common Types (types.go)
   - Response: Standard API response wrapper
   - ErrorResponse: Error response format
   - PaginatedResponse: Paginated data response
   - Event: System event representation
   - Notification: Notification format
   - Config: Configuration item

JSON Serialization:
All models implement JSON marshaling/unmarshaling for API integration.
Enums use custom MarshalJSON/UnmarshalJSON for proper serialization.

Example Usage:

    // Create a task
    task := &Task{
        Content: "Process data",
        Priority: TaskPriority(5),
        Status: TaskPending,
    }

    // Create a session
    session := &Session{
        SessionID: "session-1",
        SessionType: SessionTypeInteractive,
        Provider: "ollama",
        Status: SessionIdle,
    }

    // Marshal to JSON
    data, err := json.Marshal(task)

    // Unmarshal from JSON
    var task Task
    err := json.Unmarshal(data, &task)

    // Use response wrapper
    response := NewResponse(task)
*/
