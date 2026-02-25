package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPMetricsRegistry tracks HTTP API metrics for GAIA_GO main server
type HTTPMetricsRegistry struct {
	// Request tracking
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestSize      *prometheus.HistogramVec
	responseSize     *prometheus.HistogramVec

	// Error tracking
	errorsTotal      *prometheus.CounterVec
	errors4xxTotal   *prometheus.CounterVec
	errors5xxTotal   *prometheus.CounterVec

	// Active connections
	activeRequests   prometheus.Gauge

	registry *prometheus.Registry
	mu       sync.RWMutex
}

// NewHTTPMetricsRegistry creates and registers all HTTP metrics
func NewHTTPMetricsRegistry() *HTTPMetricsRegistry {
	registry := prometheus.NewRegistry()

	h := &HTTPMetricsRegistry{
		registry: registry,
	}

	h.registerMetrics()
	return h
}

// registerMetrics registers all HTTP metric collectors
func (h *HTTPMetricsRegistry) registerMetrics() {
	// Request counter: tracks total HTTP requests by method, path, status, and app
	h.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_http_requests_total",
			Help: "Total HTTP requests by method, path, status code, and app",
		},
		[]string{"method", "path", "status_code", "app"},
	)
	h.registry.MustRegister(h.requestsTotal)

	// Request duration histogram: tracks HTTP request latency by method, path, and app
	// Buckets: 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
	h.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gaia_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds by method, path, and app",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "app"},
	)
	h.registry.MustRegister(h.requestDuration)

	// Request size histogram: tracks request body size by method and app
	h.requestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gaia_http_request_size_bytes",
			Help:    "HTTP request body size in bytes by method and app",
			Buckets: []float64{100, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
		},
		[]string{"method", "app"},
	)
	h.registry.MustRegister(h.requestSize)

	// Response size histogram: tracks response body size by method, status, and app
	h.responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gaia_http_response_size_bytes",
			Help:    "HTTP response body size in bytes by method, status code, and app",
			Buckets: []float64{100, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
		},
		[]string{"method", "status_code", "app"},
	)
	h.registry.MustRegister(h.responseSize)

	// Error counter: tracks total errors by method, path, and error type
	h.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_http_errors_total",
			Help: "Total HTTP errors by method, path, and error type",
		},
		[]string{"method", "path", "error_type"},
	)
	h.registry.MustRegister(h.errorsTotal)

	// 4xx error counter: tracks client errors by method and path
	h.errors4xxTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_http_4xx_responses_total",
			Help: "Total HTTP 4xx responses by method and path",
		},
		[]string{"method", "path"},
	)
	h.registry.MustRegister(h.errors4xxTotal)

	// 5xx error counter: tracks server errors by method and path
	h.errors5xxTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gaia_http_5xx_responses_total",
			Help: "Total HTTP 5xx responses by method and path",
		},
		[]string{"method", "path"},
	)
	h.registry.MustRegister(h.errors5xxTotal)

	// Active requests gauge: tracks number of in-flight HTTP requests
	h.activeRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gaia_http_active_requests",
			Help: "Number of active HTTP requests",
		},
	)
	h.registry.MustRegister(h.activeRequests)
}

// RecordRequest records HTTP request metrics
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - path: Request path template (e.g., /api/typing/users/:id)
//   - statusCode: HTTP response status code
//   - app: Application name (math, typing, reading, piano)
//   - duration: Request duration in seconds
//   - reqSize: Request body size in bytes (-1 if unknown)
//   - respSize: Response body size in bytes
func (h *HTTPMetricsRegistry) RecordRequest(method, path string, statusCode int, app string, duration float64, reqSize, respSize int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Record request counter
	statusStr := string(rune(statusCode/100)*100) // e.g., 200, 400, 500
	h.requestsTotal.WithLabelValues(method, path, statusStr, app).Inc()

	// Record request duration
	h.requestDuration.WithLabelValues(method, path, app).Observe(duration)

	// Record request size if provided
	if reqSize > 0 {
		h.requestSize.WithLabelValues(method, app).Observe(float64(reqSize))
	}

	// Record response size
	h.responseSize.WithLabelValues(method, statusStr, app).Observe(float64(respSize))

	// Record error metrics for non-2xx responses
	if statusCode >= 400 {
		errorType := "unknown"
		if statusCode >= 500 {
			errorType = "server_error"
			h.errors5xxTotal.WithLabelValues(method, path).Inc()
		} else {
			errorType = "client_error"
			h.errors4xxTotal.WithLabelValues(method, path).Inc()
		}
		h.errorsTotal.WithLabelValues(method, path, errorType).Inc()
	}
}

// IncrementActiveRequests increments the active request counter
func (h *HTTPMetricsRegistry) IncrementActiveRequests() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.activeRequests.Inc()
}

// DecrementActiveRequests decrements the active request counter
func (h *HTTPMetricsRegistry) DecrementActiveRequests() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.activeRequests.Dec()
}

// GetPrometheusRegistry returns the underlying prometheus.Registry for HTTP handler
func (h *HTTPMetricsRegistry) GetPrometheusRegistry() *prometheus.Registry {
	return h.registry
}

// GetHTTPHandler returns the Prometheus HTTP handler for the /metrics endpoint
func (h *HTTPMetricsRegistry) GetHTTPHandler() http.HandlerFunc {
	return promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}).ServeHTTP
}

// Helper function to convert status code to label
func statusCodeLabel(statusCode int) string {
	return string(rune((statusCode / 100) * 100))
}
