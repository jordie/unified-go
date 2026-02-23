package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jgirmay/GAIA_GO/internal/orchestration/subsystems"
)

// HealthChecker implements health checking for all subsystems
type HealthChecker struct {
	apiPool        *subsystems.APIClientPool
	fileManager    *subsystems.FileManager
	browserPool    *subsystems.BrowserPool
	processManager *subsystems.ProcessManager
	networkCoord   *subsystems.NetworkCoordinator

	targets *subsystems.PerformanceTarget
	mu      sync.RWMutex
}

// HealthReport aggregates health status from all subsystems
type HealthReport struct {
	OverallStatus subsystems.HealthStatus      `json:"overall_status"`
	Timestamp     time.Time                    `json:"timestamp"`
	Subsystems    map[string]*SubsystemHealth  `json:"subsystems"`
	Issues        []HealthIssue                `json:"issues"`
}

// SubsystemHealth represents health of a single subsystem
type SubsystemHealth struct {
	Status       subsystems.HealthStatus `json:"status"`
	SuccessRate  float64                 `json:"success_rate"`
	AvgLatencyMs float64                 `json:"avg_latency_ms"`
	Utilization  float64                 `json:"utilization_percent"`
	Details      map[string]interface{}  `json:"details"`
}

// HealthIssue describes a health problem
type HealthIssue struct {
	Subsystem string  `json:"subsystem"`
	Severity  string  `json:"severity"` // "warning", "critical"
	Message   string  `json:"message"`
	Metric    string  `json:"metric"`
	Actual    float64 `json:"actual"`
	Expected  float64 `json:"expected"`
}

// NewHealthChecker creates health checker with performance targets
func NewHealthChecker(
	apiPool *subsystems.APIClientPool,
	fileManager *subsystems.FileManager,
	browserPool *subsystems.BrowserPool,
	processManager *subsystems.ProcessManager,
	networkCoord *subsystems.NetworkCoordinator,
	targets *subsystems.PerformanceTarget,
) *HealthChecker {
	return &HealthChecker{
		apiPool:        apiPool,
		fileManager:    fileManager,
		browserPool:    browserPool,
		processManager: processManager,
		networkCoord:   networkCoord,
		targets:        targets,
	}
}

// CheckHealth performs comprehensive health check across all subsystems
func (h *HealthChecker) CheckHealth(ctx context.Context) (*HealthReport, error) {
	// Check all subsystems in parallel
	type result struct {
		name   string
		health *SubsystemHealth
		issues []HealthIssue
	}

	resultsCh := make(chan result, 5)
	var wg sync.WaitGroup

	// API Health
	wg.Add(1)
	go func() {
		defer wg.Done()
		health := h.CheckAPIHealth(ctx)
		metrics := h.apiPool.GetMetrics()
		health.Details = metrics
		issues := h.evaluateMetricsWithIssues(metrics, "api")
		resultsCh <- result{"api", health, issues}
	}()

	// File Health
	wg.Add(1)
	go func() {
		defer wg.Done()
		health := h.CheckFileHealth(ctx)
		metrics := h.fileManager.GetMetrics()
		health.Details = metrics
		issues := h.evaluateMetricsWithIssues(metrics, "file")
		resultsCh <- result{"file", health, issues}
	}()

	// Browser Health
	wg.Add(1)
	go func() {
		defer wg.Done()
		health := h.CheckBrowserHealth(ctx)
		metrics := h.browserPool.GetMetrics()
		health.Details = metrics
		issues := h.evaluateMetricsWithIssues(metrics, "browser")
		resultsCh <- result{"browser", health, issues}
	}()

	// Process Health
	wg.Add(1)
	go func() {
		defer wg.Done()
		health := h.CheckProcessHealth(ctx)
		metrics := h.processManager.GetMetrics()
		health.Details = metrics
		issues := h.evaluateMetricsWithIssues(metrics, "process")
		resultsCh <- result{"process", health, issues}
	}()

	// Network Health
	wg.Add(1)
	go func() {
		defer wg.Done()
		health := h.CheckNetworkHealth(ctx)
		metrics := h.networkCoord.GetMetrics()
		health.Details = metrics
		issues := h.evaluateMetricsWithIssues(metrics, "network")
		resultsCh <- result{"network", health, issues}
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	subsystems := make(map[string]*SubsystemHealth)
	var allIssues []HealthIssue

	for result := range resultsCh {
		subsystems[result.name] = result.health
		allIssues = append(allIssues, result.issues...)
	}

	// Determine overall status
	overallStatus := h.determineOverallStatus(subsystems)

	return &HealthReport{
		OverallStatus: overallStatus,
		Timestamp:     time.Now(),
		Subsystems:    subsystems,
		Issues:        allIssues,
	}, nil
}

// CheckAPIHealth checks API client pool health
func (h *HealthChecker) CheckAPIHealth(ctx context.Context) *SubsystemHealth {
	metrics := h.apiPool.GetMetrics()

	// Extract success rate
	totalRequests, ok := metrics["total_requests"].(float64)
	if !ok {
		totalRequests = 0
	}
	successCount, ok := metrics["success_count"].(float64)
	if !ok {
		successCount = 0
	}

	var successRate float64
	if totalRequests > 0 {
		successRate = (successCount / totalRequests) * 100
	}

	// Extract latency
	avgLatency, ok := metrics["avg_latency_ms"].(float64)
	if !ok {
		avgLatency = 0
	}

	// Determine status
	status := h.determineStatus(successRate, avgLatency, 0)

	return &SubsystemHealth{
		Status:       status,
		SuccessRate:  successRate,
		AvgLatencyMs: avgLatency,
		Utilization:  0,
		Details:      metrics,
	}
}

// CheckFileHealth checks file manager health
func (h *HealthChecker) CheckFileHealth(ctx context.Context) *SubsystemHealth {
	metrics := h.fileManager.GetMetrics()

	// Extract success rate
	totalOps, ok := metrics["total_operations"].(float64)
	if !ok {
		totalOps = 0
	}
	successOps, ok := metrics["successful_operations"].(float64)
	if !ok {
		successOps = 0
	}

	var successRate float64
	if totalOps > 0 {
		successRate = (successOps / totalOps) * 100
	}

	// Extract latency
	avgLatency, ok := metrics["avg_latency_ms"].(float64)
	if !ok {
		avgLatency = 0
	}

	// Determine status
	status := h.determineStatus(successRate, avgLatency, 0)

	return &SubsystemHealth{
		Status:       status,
		SuccessRate:  successRate,
		AvgLatencyMs: avgLatency,
		Utilization:  0,
		Details:      metrics,
	}
}

// CheckBrowserHealth checks browser pool health
func (h *HealthChecker) CheckBrowserHealth(ctx context.Context) *SubsystemHealth {
	metrics := h.browserPool.GetMetrics()

	// Extract success rate
	totalOps, ok := metrics["total_operations"].(float64)
	if !ok {
		totalOps = 0
	}
	successOps, ok := metrics["successful_operations"].(float64)
	if !ok {
		successOps = 0
	}

	var successRate float64
	if totalOps > 0 {
		successRate = (successOps / totalOps) * 100
	}

	// Determine status
	status := h.determineStatus(successRate, 0, 0)

	return &SubsystemHealth{
		Status:       status,
		SuccessRate:  successRate,
		AvgLatencyMs: 0,
		Utilization:  0,
		Details:      metrics,
	}
}

// CheckProcessHealth checks process manager health
func (h *HealthChecker) CheckProcessHealth(ctx context.Context) *SubsystemHealth {
	metrics := h.processManager.GetMetrics()

	// Extract success rate
	totalProcs, ok := metrics["total_started"].(float64)
	if !ok {
		totalProcs = 0
	}
	completedProcs, ok := metrics["total_completed"].(float64)
	if !ok {
		completedProcs = 0
	}

	var successRate float64
	if totalProcs > 0 {
		successRate = (completedProcs / totalProcs) * 100
	}

	// Determine status
	status := h.determineStatus(successRate, 0, 0)

	return &SubsystemHealth{
		Status:       status,
		SuccessRate:  successRate,
		AvgLatencyMs: 0,
		Utilization:  0,
		Details:      metrics,
	}
}

// CheckNetworkHealth checks network coordinator health
func (h *HealthChecker) CheckNetworkHealth(ctx context.Context) *SubsystemHealth {
	metrics := h.networkCoord.GetMetrics()

	// Default: network is healthy if it exists
	successRate := 100.0
	status := subsystems.HealthStatusHealthy

	return &SubsystemHealth{
		Status:       status,
		SuccessRate:  successRate,
		AvgLatencyMs: 0,
		Utilization:  0,
		Details:      metrics,
	}
}

// determineStatus determines health status based on metrics
func (h *HealthChecker) determineStatus(successRate float64, latencyMs float64, utilization float64) subsystems.HealthStatus {
	// Success rate thresholds
	if successRate < 80 {
		return subsystems.HealthStatusUnhealthy
	}
	if successRate < 98 {
		return subsystems.HealthStatusDegraded
	}

	// Latency thresholds (if provided)
	if latencyMs > 0 && latencyMs > h.targets.P95LatencyMs {
		return subsystems.HealthStatusDegraded
	}

	// Utilization thresholds
	if utilization > 90 {
		return subsystems.HealthStatusDegraded
	}
	if utilization > 70 {
		// Check combined metrics
		if successRate < 99 {
			return subsystems.HealthStatusDegraded
		}
	}

	return subsystems.HealthStatusHealthy
}

// determineOverallStatus aggregates subsystem statuses
func (h *HealthChecker) determineOverallStatus(subsystemsHealth map[string]*SubsystemHealth) subsystems.HealthStatus {
	unhealthyCount := 0
	degradedCount := 0

	for _, health := range subsystemsHealth {
		if health.Status == subsystems.HealthStatusUnhealthy {
			unhealthyCount++
		} else if health.Status == subsystems.HealthStatusDegraded {
			degradedCount++
		}
	}

	// If any subsystem is unhealthy, overall is unhealthy
	if unhealthyCount > 0 {
		return subsystems.HealthStatusUnhealthy
	}

	// If more than one subsystem is degraded, overall is degraded
	if degradedCount > 1 {
		return subsystems.HealthStatusDegraded
	}

	// If any subsystem is degraded, overall is degraded
	if degradedCount > 0 {
		return subsystems.HealthStatusDegraded
	}

	return subsystems.HealthStatusHealthy
}

// evaluateMetricsWithIssues evaluates metrics and returns any issues found
func (h *HealthChecker) evaluateMetricsWithIssues(metrics map[string]interface{}, subsystemName string) []HealthIssue {
	var issues []HealthIssue

	// Check error rate if available
	if errorRate, ok := metrics["error_rate"].(float64); ok {
		if errorRate > h.targets.MaxErrorRate {
			issues = append(issues, HealthIssue{
				Subsystem: subsystemName,
				Severity:  "critical",
				Message:   fmt.Sprintf("Error rate %.2f%% exceeds target %.2f%%", errorRate*100, h.targets.MaxErrorRate*100),
				Metric:    "error_rate",
				Actual:    errorRate,
				Expected:  h.targets.MaxErrorRate,
			})
		}
	}

	return issues
}
