package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Example: Adding custom metrics beyond the built-in GAIA metrics
//
// This example demonstrates how to extend the monitoring system with custom metrics
// for application-specific tracking.
//
// Run with: go run examples/custom_metrics.go
// Then:     curl http://localhost:9091/metrics

// CustomMetrics holds application-specific metrics
type CustomMetrics struct {
	// Business logic metrics
	requestsProcessed prometheus.Counter
	requestsRejected  prometheus.Counter
	processingTime    prometheus.Histogram

	// Feature-specific metrics
	featureAUsage    prometheus.Counter
	featureBUsage    prometheus.Counter
	featureALatency  prometheus.Histogram

	// Custom gauges
	activeUsers   prometheus.Gauge
	queueDepth    prometheus.Gauge
	cacheHitRate  prometheus.Gauge
}

// NewCustomMetrics creates and registers custom metrics
func NewCustomMetrics() *CustomMetrics {
	registry := prometheus.NewRegistry()

	m := &CustomMetrics{
		// Counter: total requests processed
		requestsProcessed: promauto.NewCounterWithRegistry(
			prometheus.CounterOpts{
				Name: "app_requests_processed_total",
				Help: "Total number of requests processed",
			},
			registry,
		),

		// Counter: total requests rejected
		requestsRejected: promauto.NewCounterWithRegistry(
			prometheus.CounterOpts{
				Name: "app_requests_rejected_total",
				Help: "Total number of requests rejected",
			},
			registry,
		),

		// Histogram: request processing time
		processingTime: promauto.NewHistogramWithRegistry(
			prometheus.HistogramOpts{
				Name:    "app_processing_time_ms",
				Help:    "Request processing time in milliseconds",
				Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500},
			},
			registry,
		),

		// Counter: feature A usage
		featureAUsage: promauto.NewCounterWithRegistry(
			prometheus.CounterOpts{
				Name: "app_feature_a_usage_total",
				Help: "Number of times Feature A was used",
			},
			registry,
		),

		// Counter: feature B usage
		featureBUsage: promauto.NewCounterWithRegistry(
			prometheus.CounterOpts{
				Name: "app_feature_b_usage_total",
				Help: "Number of times Feature B was used",
			},
			registry,
		),

		// Histogram: feature A latency
		featureALatency: promauto.NewHistogramWithRegistry(
			prometheus.HistogramOpts{
				Name:    "app_feature_a_latency_ms",
				Help:    "Feature A operation latency in milliseconds",
				Buckets: []float64{5, 10, 25, 50, 100},
			},
			registry,
		),

		// Gauge: active users
		activeUsers: promauto.NewGaugeWithRegistry(
			prometheus.GaugeOpts{
				Name: "app_active_users",
				Help: "Number of currently active users",
			},
			registry,
		),

		// Gauge: queue depth
		queueDepth: promauto.NewGaugeWithRegistry(
			prometheus.GaugeOpts{
				Name: "app_queue_depth",
				Help: "Number of items in processing queue",
			},
			registry,
		),

		// Gauge: cache hit rate
		cacheHitRate: promauto.NewGaugeWithRegistry(
			prometheus.GaugeOpts{
				Name: "app_cache_hit_rate",
				Help: "Cache hit rate percentage (0-100)",
			},
			registry,
		),
	}

	// Register handler for custom metrics
	http.Handle("/custom-metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	return m
}

// SimulateWorkload demonstrates using custom metrics
func (m *CustomMetrics) SimulateWorkload() {
	// Simulate user activity
	for i := 0; i < 100; i++ {
		// Record active users
		m.activeUsers.Set(float64(50 + i%30))

		// Simulate request processing
		duration := float64(25 + i%75)
		m.processingTime.Observe(duration)
		m.requestsProcessed.Inc()

		// Simulate feature usage
		if i%2 == 0 {
			m.featureAUsage.Inc()
			m.featureALatency.Observe(float64(10 + i%40))
		} else {
			m.featureBUsage.Inc()
		}

		// Simulate queue activity
		m.queueDepth.Set(float64(i % 20))

		// Simulate cache performance
		hitRate := 75.0 + float64(i%20)
		m.cacheHitRate.Set(hitRate)
	}

	// Simulate some rejections
	m.requestsRejected.Add(5)
}

func main() {
	// Create custom metrics
	custom := NewCustomMetrics()

	// Simulate some activity
	go func() {
		for {
			custom.SimulateWorkload()
		}
	}()

	// Regular Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	// Custom metrics on different path
	fmt.Println("Custom metrics example - monitoring custom application metrics")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  Standard metrics (GAIA): http://localhost:9091/metrics")
	fmt.Println("  Custom metrics (app):    http://localhost:9091/custom-metrics")
	fmt.Println()
	fmt.Println("Example queries:")
	fmt.Println("  # Get request processing time distribution")
	fmt.Println("  curl http://localhost:9091/custom-metrics | grep app_processing_time")
	fmt.Println()
	fmt.Println("  # Get feature usage")
	fmt.Println("  curl http://localhost:9091/custom-metrics | grep app_feature")
	fmt.Println()
	fmt.Println("  # Get cache hit rate")
	fmt.Println("  curl http://localhost:9091/custom-metrics | grep app_cache_hit_rate")
	fmt.Println()
	fmt.Println("Listening on :9091...")

	if err := http.ListenAndServe(":9091", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Usage example in PromQL:
//
// Get 95th percentile request processing time:
//   histogram_quantile(0.95, rate(app_processing_time_ms_bucket[5m]))
//
// Get request rate:
//   rate(app_requests_processed_total[5m])
//
// Get rejection rate:
//   rate(app_requests_rejected_total[5m]) / rate(app_requests_processed_total[5m])
//
// Track feature usage trends:
//   rate(app_feature_a_usage_total[5m]) vs rate(app_feature_b_usage_total[5m])
//
// Alert on high queue depth:
//   app_queue_depth > 100
//
// Alert on low cache hit rate:
//   app_cache_hit_rate < 70
