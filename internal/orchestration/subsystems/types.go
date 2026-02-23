package subsystems

import (
	"context"
	"time"
)

// WorkType defines the type of work to be orchestrated
type WorkType string

const (
	WorkTypeAPICall           WorkType = "api_call"
	WorkTypeFileOperation     WorkType = "file_operation"
	WorkTypeBrowserAutomation WorkType = "browser_automation"
	WorkTypeProcessExecution  WorkType = "process_execution"
	WorkTypeNetworkOperation  WorkType = "network_operation"
)

// WorkItem represents a unit of work to be executed
type WorkItem struct {
	ID             string
	Type           WorkType
	Priority       int
	Deadline       *time.Time
	Context        context.Context
	Cancel         context.CancelFunc
	Data           map[string]interface{}
	CreatedAt      time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
	Status         WorkStatus
	Result         interface{}
	Error          error
	RetryCount     int
	MaxRetries     int
	Dependencies   []string
	ResourceNeeds  *ResourceNeeds
}

// WorkStatus represents the status of a work item
type WorkStatus string

const (
	WorkStatusPending    WorkStatus = "pending"
	WorkStatusStarted    WorkStatus = "started"
	WorkStatusRunning    WorkStatus = "running"
	WorkStatusCompleted  WorkStatus = "completed"
	WorkStatusFailed     WorkStatus = "failed"
	WorkStatusRetrying   WorkStatus = "retrying"
	WorkStatusCancelled  WorkStatus = "cancelled"
)

// ResourceNeeds specifies resource requirements for a work item
type ResourceNeeds struct {
	MinMemoryMB    int64
	MaxMemoryMB    int64
	CPUPercent     float64
	MaxDurationSec int64
	RequiredAPIs   bool
	RequiredFiles  bool
	RequiredBrowser bool
	RequiredProcess bool
}

// OrchestrationMetrics aggregates metrics from all subsystems
type OrchestrationMetrics struct {
	APIMetrics              map[string]interface{}
	FileMetrics             map[string]interface{}
	NetworkMetrics          map[string]interface{}
	BrowserMetrics          map[string]interface{}
	ProcessMetrics          map[string]interface{}
	TotalWorkItems          int64
	CompletedWorkItems      int64
	FailedWorkItems         int64
	AverageCompletionTime   float64
	AverageCPUUsage         float64
	AverageMemoryUsageMB    float64
	PeakConcurrentOps       int64
	ThroughputOpsPerSecond  float64
}

// WorkflowDefinition defines a multi-step workflow
type WorkflowDefinition struct {
	ID          string
	Name        string
	Description string
	Steps       []*WorkflowStep
	Timeout     time.Duration
	Retries     int
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID          string
	Name        string
	Type        WorkType
	Inputs      map[string]interface{}
	Outputs     []string
	Dependencies []string
	Retry       int
	Timeout     time.Duration
	OnError     ErrorAction
}

// ErrorAction defines what to do when a step fails
type ErrorAction string

const (
	ErrorActionRetry    ErrorAction = "retry"
	ErrorActionContinue ErrorAction = "continue"
	ErrorActionFail     ErrorAction = "fail"
)

// WorkflowExecution tracks the execution of a workflow
type WorkflowExecution struct {
	ID            string
	WorkflowID    string
	Status        WorkStatus
	StepResults   map[string]interface{}
	StartedAt     time.Time
	CompletedAt   *time.Time
	TotalDuration time.Duration
	Errors        []WorkflowError
}

// WorkflowError represents an error in workflow execution
type WorkflowError struct {
	StepID    string
	Timestamp time.Time
	Error     string
	Retry     int
}

// PoolStatus provides status information about resource pools
type PoolStatus struct {
	PoolType             string
	MaxCapacity          int64
	CurrentUtilization   int64
	PeakUtilization      int64
	SuccessfulOperations int64
	FailedOperations     int64
	AverageLatencyMs     float64
	HealthStatus         HealthStatus
}

// HealthStatus represents the health of a subsystem
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// CircuitBreakerState tracks circuit breaker state
type CircuitBreakerState struct {
	State          CircuitState
	FailureCount   int64
	SuccessCount   int64
	LastFailureAt  *time.Time
	LastSuccessAt  *time.Time
	OpenedAt       *time.Time
	HalfOpenAt     *time.Time
	Threshold      int64
	ResetTimeout   time.Duration
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

// RateLimitConfig specifies rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond   int64
	BurstSize          int64
	TimeWindow         time.Duration
	EnableCircuitBreaker bool
	CircuitThreshold   int64
}

// ConcurrencyConfig specifies concurrency limits
type ConcurrencyConfig struct {
	MaxConcurrent      int64
	MaxPerDomain       int64
	MaxPerHost         int64
	MaxQueueSize       int64
	PriorityQueueSize  int64
}

// PerformanceTarget defines performance SLOs
type PerformanceTarget struct {
	MaxLatencyMs    float64
	P95LatencyMs    float64
	P99LatencyMs    float64
	MinThroughput   float64
	MaxErrorRate    float64
	AvailabilityTarget float64
}

// SystemConfig aggregates configuration for all subsystems
type SystemConfig struct {
	APIConfig       *ConcurrencyConfig
	FileConfig      *ConcurrencyConfig
	BrowserConfig   *ConcurrencyConfig
	ProcessConfig   *ConcurrencyConfig
	NetworkConfig   *ConcurrencyConfig
	RateLimitConfig *RateLimitConfig
	ResourceLimits  map[WorkType]*ResourceLimit
	PerformanceTargets *PerformanceTarget
}

// WorkQueue interface for queuing work items
type WorkQueue interface {
	Enqueue(item *WorkItem) error
	Dequeue(ctx context.Context) (*WorkItem, error)
	Size() int64
	Peek() *WorkItem
	Clear() error
	Close() error
}

// SubsystemExecutor interface for executing work items
type SubsystemExecutor interface {
	Execute(ctx context.Context, item *WorkItem) (interface{}, error)
	CanHandle(itemType WorkType) bool
	GetMetrics() map[string]interface{}
	Close() error
}

// Orchestrator interface for orchestrating all subsystems
type Orchestrator interface {
	SubmitWork(item *WorkItem) error
	ExecuteWorkflow(ctx context.Context, workflow *WorkflowDefinition) (*WorkflowExecution, error)
	GetMetrics() *OrchestrationMetrics
	GetPoolStatus() []*PoolStatus
	GetHealth() HealthStatus
	Close() error
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	RecordOperation(itemType WorkType, duration time.Duration, success bool)
	RecordThroughput(itemType WorkType, opsPerSec float64)
	GetMetrics() map[string]interface{}
	Reset()
}

// HealthChecker interface for health checks
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthStatus
	GetDetails() map[string]interface{}
}
