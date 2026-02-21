package models

import (
	"database/sql"
	"time"
)

// Metrics represents system performance metrics
type Metrics struct {
	ID              int64          `json:"id"`
	Timestamp       time.Time      `json:"timestamp"`
	TaskThroughput  sql.NullInt64  `json:"task_throughput"`
	AvgLatencyMS    sql.NullFloat64 `json:"avg_latency_ms"`
	SuccessRate     sql.NullFloat64 `json:"success_rate"`
	MemoryMB        sql.NullFloat64 `json:"memory_mb"`
	CPUPercent      sql.NullFloat64 `json:"cpu_percent"`
	ActiveSessions  sql.NullInt64  `json:"active_sessions"`
	PendingTasks    sql.NullInt64  `json:"pending_tasks"`
	Metadata        sql.NullString `json:"metadata"`
}

// MetricsCreate represents input for recording metrics
type MetricsCreate struct {
	TaskThroughput *int    `json:"task_throughput"`
	AvgLatencyMS   *float64 `json:"avg_latency_ms"`
	SuccessRate    *float64 `json:"success_rate"`
	MemoryMB       *float64 `json:"memory_mb"`
	CPUPercent     *float64 `json:"cpu_percent"`
	ActiveSessions *int    `json:"active_sessions"`
	PendingTasks   *int    `json:"pending_tasks"`
	Metadata       string  `json:"metadata,omitempty"`
}

// MetricsAggregate represents aggregated metrics over a time period
type MetricsAggregate struct {
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	PeriodSeconds       int64     `json:"period_seconds"`
	TaskThroughput      AggregateStats `json:"task_throughput"`
	AvgLatencyMS        AggregateStats `json:"avg_latency_ms"`
	SuccessRate         AggregateStats `json:"success_rate"`
	MemoryMB            AggregateStats `json:"memory_mb"`
	CPUPercent          AggregateStats `json:"cpu_percent"`
	ActiveSessions      AggregateStats `json:"active_sessions"`
	PendingTasks        AggregateStats `json:"pending_tasks"`
	SampleCount         int64     `json:"sample_count"`
}

// AggregateStats represents statistical aggregation
type AggregateStats struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	P95    float64 `json:"p95"`
	P99    float64 `json:"p99"`
	StdDev float64 `json:"stddev"`
}

// SystemHealth represents overall system health
type SystemHealth struct {
	Status            string  `json:"status"`
	HealthScore       float64 `json:"health_score"`
	ActiveSessions    int64   `json:"active_sessions"`
	PendingTasks      int64   `json:"pending_tasks"`
	ThroughputPerMin  float64 `json:"throughput_per_min"`
	AverageLatencyMS  float64 `json:"average_latency_ms"`
	SuccessRate       float64 `json:"success_rate"`
	MemoryUsageMB     float64 `json:"memory_usage_mb"`
	CPUUsagePercent   float64 `json:"cpu_usage_percent"`
	LastUpdateTime    time.Time `json:"last_update_time"`
}

// PerformanceReport represents a detailed performance report
type PerformanceReport struct {
	GeneratedAt        time.Time        `json:"generated_at"`
	ReportPeriod       string           `json:"report_period"`
	SystemHealth       SystemHealth     `json:"system_health"`
	MetricsAggregate   MetricsAggregate `json:"metrics_aggregate"`
	TopSessions        []SessionMetrics `json:"top_sessions"`
	SlowTasks          []TaskMetrics    `json:"slow_tasks"`
	FailedTasks        []TaskMetrics    `json:"failed_tasks"`
	Recommendations    []string         `json:"recommendations"`
}

// TaskMetrics represents metrics for a specific task
type TaskMetrics struct {
	TaskID            int64     `json:"task_id"`
	Content           string    `json:"content"`
	Priority          TaskPriority `json:"priority"`
	Status            TaskStatus   `json:"status"`
	ExecutionTimeMS   int64     `json:"execution_time_ms"`
	CompletedAt       *time.Time `json:"completed_at"`
	FailureReason     string    `json:"failure_reason,omitempty"`
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Healthy           bool              `json:"healthy"`
	Timestamp         time.Time         `json:"timestamp"`
	Checks            map[string]bool   `json:"checks"`
	Warnings          []string          `json:"warnings"`
	Errors            []string          `json:"errors"`
}

// MetricsTimeSeries represents a time series of metric data
type MetricsTimeSeries struct {
	MetricName string        `json:"metric_name"`
	Unit       string        `json:"unit"`
	DataPoints []TimeSeriesPoint `json:"data_points"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
}

// TimeSeriesPoint represents a single point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// AlertThreshold represents thresholds for alerting
type AlertThreshold struct {
	MetricName  string  `json:"metric_name"`
	WarningMin  *float64 `json:"warning_min,omitempty"`
	WarningMax  *float64 `json:"warning_max,omitempty"`
	CriticalMin *float64 `json:"critical_min,omitempty"`
	CriticalMax *float64 `json:"critical_max,omitempty"`
}

// MetricsAlert represents an alert based on metric thresholds
type MetricsAlert struct {
	ID          string    `json:"id"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Severity    string    `json:"severity"` // "warning" or "critical"
	Threshold   float64   `json:"threshold"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	Message     string    `json:"message"`
}
