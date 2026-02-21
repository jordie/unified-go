package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"unified-go/internal/models"
	"unified-go/internal/storage"
)

// MetricsRepository defines the interface for metrics data access
type MetricsRepository interface {
	// Record records a metrics snapshot
	Record(ctx context.Context, metrics *models.MetricsCreate) error

	// GetLatest retrieves the most recent metrics
	GetLatest(ctx context.Context) (*models.Metrics, error)

	// GetByTimeRange retrieves metrics within a time range
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]*models.Metrics, error)

	// GetAggregated retrieves aggregated metrics over a time period
	GetAggregated(ctx context.Context, start, end time.Time) (*models.MetricsAggregate, error)

	// GetSystemHealth calculates current system health
	GetSystemHealth(ctx context.Context) (*models.SystemHealth, error)

	// GetPerformanceReport generates a performance report
	GetPerformanceReport(ctx context.Context, period time.Duration) (*models.PerformanceReport, error)

	// GetTimeSeries retrieves a metric as time series data
	GetTimeSeries(ctx context.Context, metricName string, start, end time.Time) (*models.MetricsTimeSeries, error)

	// CleanupOldMetrics removes metrics older than retention period
	CleanupOldMetrics(ctx context.Context, retentionDays int) (int64, error)

	// GetMetricsStats retrieves statistical summary
	GetMetricsStats(ctx context.Context, period time.Duration) (map[string]interface{}, error)
}

// metricsRepository implements MetricsRepository
type metricsRepository struct {
	store *storage.SQLiteStore
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(store *storage.SQLiteStore) MetricsRepository {
	return &metricsRepository{store: store}
}

func (r *metricsRepository) Record(ctx context.Context, metrics *models.MetricsCreate) error {
	metadataJSON := ""
	if metrics.Metadata != "" {
		metadataJSON = metrics.Metadata
	}

	_, err := r.store.Exec(ctx,
		`INSERT INTO metrics (task_throughput, avg_latency_ms, success_rate, memory_mb, cpu_percent, active_sessions, pending_tasks, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		metrics.TaskThroughput,
		metrics.AvgLatencyMS,
		metrics.SuccessRate,
		metrics.MemoryMB,
		metrics.CPUPercent,
		metrics.ActiveSessions,
		metrics.PendingTasks,
		sql.NullString{String: metadataJSON, Valid: metadataJSON != ""},
	)
	if err != nil {
		return fmt.Errorf("failed to record metrics: %w", err)
	}
	return nil
}

func (r *metricsRepository) GetLatest(ctx context.Context) (*models.Metrics, error) {
	row := r.store.QueryRow(ctx,
		`SELECT id, timestamp, task_throughput, avg_latency_ms, success_rate, memory_mb, cpu_percent, active_sessions, pending_tasks, metadata
		 FROM metrics ORDER BY timestamp DESC LIMIT 1`)

	metrics := &models.Metrics{}
	err := row.Scan(
		&metrics.ID, &metrics.Timestamp, &metrics.TaskThroughput, &metrics.AvgLatencyMS,
		&metrics.SuccessRate, &metrics.MemoryMB, &metrics.CPUPercent, &metrics.ActiveSessions,
		&metrics.PendingTasks, &metrics.Metadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}
	return metrics, nil
}

func (r *metricsRepository) GetByTimeRange(ctx context.Context, start, end time.Time) ([]*models.Metrics, error) {
	rows, err := r.store.Query(ctx,
		`SELECT id, timestamp, task_throughput, avg_latency_ms, success_rate, memory_mb, cpu_percent, active_sessions, pending_tasks, metadata
		 FROM metrics WHERE timestamp BETWEEN ? AND ? ORDER BY timestamp DESC`,
		start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metricsList []*models.Metrics
	for rows.Next() {
		metrics := &models.Metrics{}
		err := rows.Scan(
			&metrics.ID, &metrics.Timestamp, &metrics.TaskThroughput, &metrics.AvgLatencyMS,
			&metrics.SuccessRate, &metrics.MemoryMB, &metrics.CPUPercent, &metrics.ActiveSessions,
			&metrics.PendingTasks, &metrics.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metrics: %w", err)
		}
		metricsList = append(metricsList, metrics)
	}
	return metricsList, rows.Err()
}

func (r *metricsRepository) GetAggregated(ctx context.Context, start, end time.Time) (*models.MetricsAggregate, error) {
	metricsList, err := r.GetByTimeRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	if len(metricsList) == 0 {
		return &models.MetricsAggregate{
			StartTime:     start,
			EndTime:       end,
			PeriodSeconds: int64(end.Sub(start).Seconds()),
			SampleCount:   0,
		}, nil
	}

	agg := &models.MetricsAggregate{
		StartTime:     start,
		EndTime:       end,
		PeriodSeconds: int64(end.Sub(start).Seconds()),
		SampleCount:   int64(len(metricsList)),
	}

	// Aggregate each metric
	agg.TaskThroughput = aggregateNullInt64(metricsList, func(m *models.Metrics) sql.NullInt64 { return m.TaskThroughput })
	agg.AvgLatencyMS = aggregateNullFloat64(metricsList, func(m *models.Metrics) sql.NullFloat64 { return m.AvgLatencyMS })
	agg.SuccessRate = aggregateNullFloat64(metricsList, func(m *models.Metrics) sql.NullFloat64 { return m.SuccessRate })
	agg.MemoryMB = aggregateNullFloat64(metricsList, func(m *models.Metrics) sql.NullFloat64 { return m.MemoryMB })
	agg.CPUPercent = aggregateNullFloat64(metricsList, func(m *models.Metrics) sql.NullFloat64 { return m.CPUPercent })
	agg.ActiveSessions = aggregateNullInt64(metricsList, func(m *models.Metrics) sql.NullInt64 { return m.ActiveSessions })
	agg.PendingTasks = aggregateNullInt64(metricsList, func(m *models.Metrics) sql.NullInt64 { return m.PendingTasks })

	return agg, nil
}

func (r *metricsRepository) GetSystemHealth(ctx context.Context) (*models.SystemHealth, error) {
	latest, err := r.GetLatest(ctx)
	if err != nil {
		return nil, err
	}

	health := &models.SystemHealth{
		Status:        "healthy",
		HealthScore:   100.0,
		LastUpdateTime: time.Now(),
	}

	if latest == nil {
		return health, nil
	}

	if latest.TaskThroughput.Valid {
		health.ThroughputPerMin = float64(latest.TaskThroughput.Int64)
	}
	if latest.AvgLatencyMS.Valid {
		health.AverageLatencyMS = latest.AvgLatencyMS.Float64
	}
	if latest.SuccessRate.Valid {
		health.SuccessRate = latest.SuccessRate.Float64
	}
	if latest.MemoryMB.Valid {
		health.MemoryUsageMB = latest.MemoryMB.Float64
	}
	if latest.CPUPercent.Valid {
		health.CPUUsagePercent = latest.CPUPercent.Float64
	}
	if latest.ActiveSessions.Valid {
		health.ActiveSessions = latest.ActiveSessions.Int64
	}
	if latest.PendingTasks.Valid {
		health.PendingTasks = latest.PendingTasks.Int64
	}

	// Calculate health score based on metrics
	score := 100.0
	if health.CPUUsagePercent > 90 {
		score -= 20
		health.Status = "degraded"
	} else if health.CPUUsagePercent > 70 {
		score -= 10
	}

	if health.MemoryUsageMB > 8000 {
		score -= 20
		health.Status = "degraded"
	} else if health.MemoryUsageMB > 6000 {
		score -= 10
	}

	if health.SuccessRate < 0.9 {
		score -= 15
		health.Status = "degraded"
	}

	health.HealthScore = max(0.0, score)

	if health.Status == "degraded" && score < 50 {
		health.Status = "unhealthy"
	}

	return health, nil
}

func (r *metricsRepository) GetPerformanceReport(ctx context.Context, period time.Duration) (*models.PerformanceReport, error) {
	now := time.Now()
	start := now.Add(-period)

	health, err := r.GetSystemHealth(ctx)
	if err != nil {
		return nil, err
	}

	agg, err := r.GetAggregated(ctx, start, now)
	if err != nil {
		return nil, err
	}

	report := &models.PerformanceReport{
		GeneratedAt:      now,
		ReportPeriod:     period.String(),
		SystemHealth:     *health,
		MetricsAggregate: *agg,
		Recommendations:  []string{},
	}

	// Generate recommendations based on metrics
	if health.CPUUsagePercent > 80 {
		report.Recommendations = append(report.Recommendations, "High CPU usage - consider scaling up or optimizing queries")
	}
	if health.MemoryUsageMB > 7000 {
		report.Recommendations = append(report.Recommendations, "High memory usage - consider cleanup or resource optimization")
	}
	if health.SuccessRate < 0.95 {
		report.Recommendations = append(report.Recommendations, fmt.Sprintf("Low success rate (%.1f%%) - investigate failures", health.SuccessRate*100))
	}

	return report, nil
}

func (r *metricsRepository) GetTimeSeries(ctx context.Context, metricName string, start, end time.Time) (*models.MetricsTimeSeries, error) {
	metricsList, err := r.GetByTimeRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	ts := &models.MetricsTimeSeries{
		MetricName: metricName,
		StartTime:  start,
		EndTime:    end,
		DataPoints: []models.TimeSeriesPoint{},
	}

	for _, m := range metricsList {
		point := models.TimeSeriesPoint{
			Timestamp: m.Timestamp,
		}

		switch metricName {
		case "task_throughput":
			if m.TaskThroughput.Valid {
				point.Value = float64(m.TaskThroughput.Int64)
				ts.Unit = "tasks/min"
			}
		case "avg_latency_ms":
			if m.AvgLatencyMS.Valid {
				point.Value = m.AvgLatencyMS.Float64
				ts.Unit = "milliseconds"
			}
		case "success_rate":
			if m.SuccessRate.Valid {
				point.Value = m.SuccessRate.Float64
				ts.Unit = "percent"
			}
		case "memory_mb":
			if m.MemoryMB.Valid {
				point.Value = m.MemoryMB.Float64
				ts.Unit = "MB"
			}
		case "cpu_percent":
			if m.CPUPercent.Valid {
				point.Value = m.CPUPercent.Float64
				ts.Unit = "percent"
			}
		case "active_sessions":
			if m.ActiveSessions.Valid {
				point.Value = float64(m.ActiveSessions.Int64)
				ts.Unit = "sessions"
			}
		case "pending_tasks":
			if m.PendingTasks.Valid {
				point.Value = float64(m.PendingTasks.Int64)
				ts.Unit = "tasks"
			}
		}

		ts.DataPoints = append(ts.DataPoints, point)
	}

	return ts, nil
}

func (r *metricsRepository) CleanupOldMetrics(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := r.store.Exec(ctx,
		`DELETE FROM metrics WHERE timestamp < ?`,
		cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old metrics: %w", err)
	}
	return result.RowsAffected()
}

func (r *metricsRepository) GetMetricsStats(ctx context.Context, period time.Duration) (map[string]interface{}, error) {
	now := time.Now()
	start := now.Add(-period)

	agg, err := r.GetAggregated(ctx, start, now)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"period_seconds": agg.PeriodSeconds,
		"sample_count":   agg.SampleCount,
		"task_throughput": map[string]interface{}{
			"min":    agg.TaskThroughput.Min,
			"max":    agg.TaskThroughput.Max,
			"mean":   agg.TaskThroughput.Mean,
			"median": agg.TaskThroughput.Median,
			"p95":    agg.TaskThroughput.P95,
			"p99":    agg.TaskThroughput.P99,
		},
		"avg_latency_ms": map[string]interface{}{
			"min":    agg.AvgLatencyMS.Min,
			"max":    agg.AvgLatencyMS.Max,
			"mean":   agg.AvgLatencyMS.Mean,
			"median": agg.AvgLatencyMS.Median,
			"p95":    agg.AvgLatencyMS.P95,
			"p99":    agg.AvgLatencyMS.P99,
		},
	}

	return stats, nil
}

// Helper functions

func aggregateNullInt64(metrics []*models.Metrics, getter func(*models.Metrics) sql.NullInt64) models.AggregateStats {
	var values []float64
	for _, m := range metrics {
		v := getter(m)
		if v.Valid {
			values = append(values, float64(v.Int64))
		}
	}
	return aggregateValues(values)
}

func aggregateNullFloat64(metrics []*models.Metrics, getter func(*models.Metrics) sql.NullFloat64) models.AggregateStats {
	var values []float64
	for _, m := range metrics {
		v := getter(m)
		if v.Valid {
			values = append(values, v.Float64)
		}
	}
	return aggregateValues(values)
}

func aggregateValues(values []float64) models.AggregateStats {
	if len(values) == 0 {
		return models.AggregateStats{}
	}

	sort.Float64s(values)

	min := values[0]
	max := values[len(values)-1]

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate median
	var median float64
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	} else {
		median = values[len(values)/2]
	}

	// Calculate percentiles
	p95Idx := int(float64(len(values)) * 0.95)
	p99Idx := int(float64(len(values)) * 0.99)
	if p95Idx >= len(values) {
		p95Idx = len(values) - 1
	}
	if p99Idx >= len(values) {
		p99Idx = len(values) - 1
	}

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	stddev := math.Sqrt(variance)

	return models.AggregateStats{
		Min:    min,
		Max:    max,
		Mean:   mean,
		Median: median,
		P95:    values[p95Idx],
		P99:    values[p99Idx],
		StdDev: stddev,
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
