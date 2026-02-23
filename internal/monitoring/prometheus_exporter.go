package monitoring

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// PrometheusExporter converts subsystem metrics to Prometheus format
type PrometheusExporter struct {
	// API Client Pool Metrics
	apiTotalRequests   prometheus.Counter
	apiSuccessCount    prometheus.Counter
	apiErrorCount      prometheus.Counter
	apiLatencyMs       prometheus.Histogram
	apiActiveClients   prometheus.Gauge
	apiRateLimited     prometheus.Counter

	// File Manager Metrics
	fileOperations      prometheus.Counter
	fileSuccessCount    prometheus.Counter
	fileErrorCount      prometheus.Counter
	fileBytesProcessed  prometheus.Counter
	fileLatencyMs       prometheus.Histogram
	fileConcurrent      prometheus.Gauge

	// Browser Pool Metrics
	browserInstances    prometheus.Gauge
	browserActiveTabs   prometheus.Gauge
	browserOperations   prometheus.Counter
	browserErrors       prometheus.Counter

	// Process Manager Metrics
	processActive       prometheus.Gauge
	processStarted      prometheus.Counter
	processCompleted    prometheus.Counter
	processFailed       prometheus.Counter
	processMemoryMB     prometheus.Gauge
	processCPUPercent   prometheus.Gauge

	// Network Coordinator Metrics
	networkBytesTransferred prometheus.Counter
	networkActiveConnections prometheus.Gauge
	networkDNSCacheHits     prometheus.Counter
	networkDNSCacheMisses   prometheus.Counter

	// System-wide Metrics
	systemThroughput    prometheus.Gauge
	systemMemoryMB      prometheus.Gauge
	systemGoroutines    prometheus.Gauge

	registry   *prometheus.Registry
	mu         sync.RWMutex
}

// NewPrometheusExporter creates exporter and registers all metrics
func NewPrometheusExporter() *PrometheusExporter {
	registry := prometheus.NewRegistry()

	// Register Go runtime metrics
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	e := &PrometheusExporter{
		registry:   registry,
	}

	e.registerMetrics()
	return e
}

// registerMetrics registers all Prometheus collectors
func (e *PrometheusExporter) registerMetrics() {
	// API Client Pool Counters
	e.apiTotalRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_api_requests_total",
		Help: "Total API requests made",
	})
	e.registry.MustRegister(e.apiTotalRequests)

	e.apiSuccessCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_api_success_total",
		Help: "Total successful API requests",
	})
	e.registry.MustRegister(e.apiSuccessCount)

	e.apiErrorCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_api_errors_total",
		Help: "Total API errors",
	})
	e.registry.MustRegister(e.apiErrorCount)

	e.apiRateLimited = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_api_rate_limited_total",
		Help: "Total rate limited requests",
	})
	e.registry.MustRegister(e.apiRateLimited)

	// API Latency Histogram
	e.apiLatencyMs = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "gaia_api_latency_ms",
		Help:    "API request latency in milliseconds",
		Buckets: []float64{1, 5, 10, 50, 100, 500, 1000},
	})
	e.registry.MustRegister(e.apiLatencyMs)

	// API Active Clients Gauge
	e.apiActiveClients = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_api_active_clients",
		Help: "Number of active API clients",
	})
	e.registry.MustRegister(e.apiActiveClients)

	// File Manager Counters
	e.fileOperations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_file_operations_total",
		Help: "Total file operations",
	})
	e.registry.MustRegister(e.fileOperations)

	e.fileSuccessCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_file_success_total",
		Help: "Total successful file operations",
	})
	e.registry.MustRegister(e.fileSuccessCount)

	e.fileErrorCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_file_errors_total",
		Help: "Total file operation errors",
	})
	e.registry.MustRegister(e.fileErrorCount)

	e.fileBytesProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_file_bytes_processed_total",
		Help: "Total bytes processed by file manager",
	})
	e.registry.MustRegister(e.fileBytesProcessed)

	// File Latency Histogram
	e.fileLatencyMs = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "gaia_file_latency_ms",
		Help:    "File operation latency in milliseconds",
		Buckets: []float64{1, 5, 10, 50, 100, 500, 1000},
	})
	e.registry.MustRegister(e.fileLatencyMs)

	// File Concurrent Gauge
	e.fileConcurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_file_concurrent_operations",
		Help: "Number of concurrent file operations",
	})
	e.registry.MustRegister(e.fileConcurrent)

	// Browser Pool Gauges
	e.browserInstances = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_browser_instances_active",
		Help: "Number of active browser instances",
	})
	e.registry.MustRegister(e.browserInstances)

	e.browserActiveTabs = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_browser_active_tabs",
		Help: "Number of active browser tabs",
	})
	e.registry.MustRegister(e.browserActiveTabs)

	// Browser Operations
	e.browserOperations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_browser_operations_total",
		Help: "Total browser operations",
	})
	e.registry.MustRegister(e.browserOperations)

	e.browserErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_browser_errors_total",
		Help: "Total browser operation errors",
	})
	e.registry.MustRegister(e.browserErrors)

	// Process Manager Gauges
	e.processActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_process_active",
		Help: "Number of active processes",
	})
	e.registry.MustRegister(e.processActive)

	e.processMemoryMB = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_process_memory_mb",
		Help: "Total process manager memory in MB",
	})
	e.registry.MustRegister(e.processMemoryMB)

	e.processCPUPercent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_process_cpu_percent",
		Help: "Process manager CPU usage percentage",
	})
	e.registry.MustRegister(e.processCPUPercent)

	// Process Operations
	e.processStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_process_started_total",
		Help: "Total processes started",
	})
	e.registry.MustRegister(e.processStarted)

	e.processCompleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_process_completed_total",
		Help: "Total processes completed",
	})
	e.registry.MustRegister(e.processCompleted)

	e.processFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_process_failed_total",
		Help: "Total process failures",
	})
	e.registry.MustRegister(e.processFailed)

	// Network Coordinator Metrics
	e.networkBytesTransferred = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_network_bytes_transferred_total",
		Help: "Total bytes transferred over network",
	})
	e.registry.MustRegister(e.networkBytesTransferred)

	e.networkActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_network_active_connections",
		Help: "Number of active network connections",
	})
	e.registry.MustRegister(e.networkActiveConnections)

	e.networkDNSCacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_network_dns_cache_hits_total",
		Help: "Total DNS cache hits",
	})
	e.registry.MustRegister(e.networkDNSCacheHits)

	e.networkDNSCacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gaia_network_dns_cache_misses_total",
		Help: "Total DNS cache misses",
	})
	e.registry.MustRegister(e.networkDNSCacheMisses)

	// System-wide Metrics
	e.systemThroughput = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_system_throughput_ops_per_sec",
		Help: "System throughput in operations per second",
	})
	e.registry.MustRegister(e.systemThroughput)

	e.systemMemoryMB = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_system_memory_mb",
		Help: "System memory usage in MB",
	})
	e.registry.MustRegister(e.systemMemoryMB)

	e.systemGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gaia_system_goroutines",
		Help: "Number of active goroutines",
	})
	e.registry.MustRegister(e.systemGoroutines)
}

// CollectMetrics is called by the monitoring server to update metrics
// Note: Metrics are updated by the monitoring server when it calls updateAPIMetrics, etc.
func (e *PrometheusExporter) CollectMetrics() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Metrics are updated via UpdateAPIMetrics, UpdateFileMetrics, etc. methods
	return nil
}

// StartPeriodicCollection runs collection every interval
func (e *PrometheusExporter) StartPeriodicCollection(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := e.CollectMetrics(); err != nil {
				log.Printf("Error collecting metrics: %v", err)
			}
		case <-stopCh:
			return
		}
	}
}

// GetRegistry returns the Prometheus registry for custom handlers
func (e *PrometheusExporter) GetRegistry() *prometheus.Registry {
	return e.registry
}

// Metric update methods

func (e *PrometheusExporter) updateAPIMetrics(metrics map[string]interface{}) {
	// Extract and update API metrics
	if activeClients, ok := metrics["active_clients"].(float64); ok {
		e.apiActiveClients.Set(activeClients)
	}

	if avgLatency, ok := metrics["avg_latency_ms"].(float64); ok && avgLatency > 0 {
		e.apiLatencyMs.Observe(avgLatency)
	}
}

func (e *PrometheusExporter) updateFileMetrics(metrics map[string]interface{}) {
	// Extract and update file metrics
	if concurrent, ok := metrics["concurrent_operations"].(float64); ok {
		e.fileConcurrent.Set(concurrent)
	}

	if avgLatency, ok := metrics["avg_latency_ms"].(float64); ok && avgLatency > 0 {
		e.fileLatencyMs.Observe(avgLatency)
	}
}

func (e *PrometheusExporter) updateBrowserMetrics(metrics map[string]interface{}) {
	// Extract and update browser metrics
	if instances, ok := metrics["active_instances"].(float64); ok {
		e.browserInstances.Set(instances)
	}

	if tabs, ok := metrics["active_tabs"].(float64); ok {
		e.browserActiveTabs.Set(tabs)
	}
}

func (e *PrometheusExporter) updateProcessMetrics(metrics map[string]interface{}) {
	// Extract and update process metrics
	if active, ok := metrics["active_processes"].(float64); ok {
		e.processActive.Set(active)
	}

	if memory, ok := metrics["memory_mb"].(float64); ok {
		e.processMemoryMB.Set(memory)
	}

	if cpu, ok := metrics["cpu_percent"].(float64); ok {
		e.processCPUPercent.Set(cpu)
	}
}

func (e *PrometheusExporter) updateNetworkMetrics(metrics map[string]interface{}) {
	// Extract and update network metrics
	if conns, ok := metrics["active_connections"].(float64); ok {
		e.networkActiveConnections.Set(conns)
	}
}
