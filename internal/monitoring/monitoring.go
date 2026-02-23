package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/jgirmay/GAIA_GO/internal/orchestration/subsystems"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MonitoringServer provides observability endpoints including health checks, profiling, and metrics
type MonitoringServer struct {
	apiPool       *subsystems.APIClientPool
	fileManager   *subsystems.FileManager
	browserPool   *subsystems.BrowserPool
	processManager *subsystems.ProcessManager
	networkCoord  *subsystems.NetworkCoordinator

	healthChecker *HealthChecker
	exporter      *PrometheusExporter
	mux           *http.ServeMux
	server        *http.Server
	mu            sync.RWMutex
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Subsystems map[string]interface{} `json:"subsystems"`
	Issues     []string               `json:"issues,omitempty"`
}

// NewMonitoringServer creates a monitoring server with all subsystems
func NewMonitoringServer(
	apiPool *subsystems.APIClientPool,
	fileManager *subsystems.FileManager,
	browserPool *subsystems.BrowserPool,
	processManager *subsystems.ProcessManager,
	networkCoord *subsystems.NetworkCoordinator,
) *MonitoringServer {
	m := &MonitoringServer{
		apiPool:        apiPool,
		fileManager:    fileManager,
		browserPool:    browserPool,
		processManager: processManager,
		networkCoord:   networkCoord,
		mux:            http.NewServeMux(),
	}

	// Create health checker with default targets
	targets := &subsystems.PerformanceTarget{
		MaxLatencyMs:       100,
		P95LatencyMs:       50,
		P99LatencyMs:       150,
		MinThroughput:      500000,
		MaxErrorRate:       0.001,
		AvailabilityTarget: 0.999,
	}
	m.healthChecker = NewHealthChecker(apiPool, fileManager, browserPool, processManager, networkCoord, targets)

	// Create Prometheus exporter
	m.exporter = NewPrometheusExporter()

	return m
}

// RegisterHandlers registers all monitoring HTTP endpoints
func (m *MonitoringServer) RegisterHandlers() {
	// Prometheus metrics endpoint
	m.mux.Handle("/metrics", promhttp.Handler())

	// Per-subsystem metrics (JSON for debugging)
	m.mux.HandleFunc("/metrics/api", m.handleAPIMetrics)
	m.mux.HandleFunc("/metrics/file", m.handleFileMetrics)
	m.mux.HandleFunc("/metrics/browser", m.handleBrowserMetrics)
	m.mux.HandleFunc("/metrics/process", m.handleProcessMetrics)
	m.mux.HandleFunc("/metrics/network", m.handleNetworkMetrics)
	m.mux.HandleFunc("/metrics/aggregate", m.handleAggregateMetrics)

	// Health check endpoints
	m.mux.HandleFunc("/health", m.handleHealth)
	m.mux.HandleFunc("/health/api", m.handleAPIHealth)
	m.mux.HandleFunc("/health/file", m.handleFileHealth)
	m.mux.HandleFunc("/health/browser", m.handleBrowserHealth)
	m.mux.HandleFunc("/health/process", m.handleProcessHealth)
	m.mux.HandleFunc("/health/network", m.handleNetworkHealth)
	m.mux.HandleFunc("/readiness", m.handleReadiness)
	m.mux.HandleFunc("/liveness", m.handleLiveness)

	// pprof endpoints
	m.mux.HandleFunc("/debug/pprof/", pprof.Index)
	m.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	m.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Additional profiling
	m.mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	m.mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	m.mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	m.mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	m.mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	m.mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
}

// Start starts the monitoring server on the specified port
func (m *MonitoringServer) Start(port string) error {
	m.RegisterHandlers()

	// Start periodic metrics collection
	stopCh := make(chan struct{})
	go m.exporter.StartPeriodicCollection(15*time.Second, stopCh)

	m.server = &http.Server{
		Addr:         port,
		Handler:      m.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting monitoring server on %s", port)
	return m.server.ListenAndServe()
}

// Close gracefully closes the monitoring server
func (m *MonitoringServer) Close() error {
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return m.server.Shutdown(ctx)
	}
	return nil
}

// Handler methods

func (m *MonitoringServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health, err := m.healthChecker.CheckHealth(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := HealthResponse{
		Status:     string(health.OverallStatus),
		Timestamp:  health.Timestamp,
		Subsystems: make(map[string]interface{}),
	}

	// Add subsystem health data
	for name, subsystemHealth := range health.Subsystems {
		resp.Subsystems[name] = subsystemHealth
	}

	if len(health.Issues) > 0 {
		for _, issue := range health.Issues {
			resp.Issues = append(resp.Issues, issue.Message)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (m *MonitoringServer) handleAPIHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := m.healthChecker.CheckAPIHealth(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (m *MonitoringServer) handleFileHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := m.healthChecker.CheckFileHealth(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (m *MonitoringServer) handleBrowserHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := m.healthChecker.CheckBrowserHealth(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (m *MonitoringServer) handleProcessHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := m.healthChecker.CheckProcessHealth(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (m *MonitoringServer) handleNetworkHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := m.healthChecker.CheckNetworkHealth(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (m *MonitoringServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Readiness check: all subsystems must be available
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health, err := m.healthChecker.CheckHealth(ctx)
	if err != nil {
		http.Error(w, "health check failed", http.StatusServiceUnavailable)
		return
	}

	// Check if any subsystem is unhealthy
	for _, subsystemHealth := range health.Subsystems {
		if subsystemHealth.Status == subsystems.HealthStatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "subsystem unhealthy")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ready")
}

func (m *MonitoringServer) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Liveness check: server is still running
	// Check if we can reach subsystems
	if m.apiPool == nil {
		http.Error(w, "api pool not initialized", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "alive")
}

// Metrics handlers

func (m *MonitoringServer) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.apiPool.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (m *MonitoringServer) handleFileMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.fileManager.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (m *MonitoringServer) handleBrowserMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.browserPool.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (m *MonitoringServer) handleProcessMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.processManager.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (m *MonitoringServer) handleNetworkMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.networkCoord.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (m *MonitoringServer) handleAggregateMetrics(w http.ResponseWriter, r *http.Request) {
	aggregate := map[string]interface{}{
		"api":     m.apiPool.GetMetrics(),
		"file":    m.fileManager.GetMetrics(),
		"browser": m.browserPool.GetMetrics(),
		"process": m.processManager.GetMetrics(),
		"network": m.networkCoord.GetMetrics(),
		"timestamp": time.Now().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aggregate)
}

// IsPortAvailable checks if a port is available
func IsPortAvailable(port string) bool {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
