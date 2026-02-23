package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jgirmay/GAIA_GO/internal/monitoring"
	"github.com/jgirmay/GAIA_GO/internal/orchestration/subsystems"
)

func main() {
	// Example: Starting GAIA monitoring with all subsystems

	// 1. Initialize all subsystems
	fmt.Println("Initializing GAIA subsystems...")
	apiPool := subsystems.NewAPIClientPool(100, 1000)
	fileManager := subsystems.NewFileManager(50, 65536)
	browserPool := subsystems.NewBrowserPool(20, 100)
	processManager := subsystems.NewProcessManager(50)
	networkCoord := subsystems.NewNetworkCoordinator(10 * 1024 * 1024)

	defer apiPool.Close()
	defer fileManager.Close()
	defer browserPool.Close()
	defer processManager.Close()
	defer networkCoord.Close()

	fmt.Println("✓ All subsystems initialized")

	// 2. Create monitoring server
	fmt.Println("\nCreating monitoring server...")
	monServer := monitoring.NewMonitoringServer(
		apiPool,
		fileManager,
		browserPool,
		processManager,
		networkCoord,
	)
	fmt.Println("✓ Monitoring server created")

	// 3. Start monitoring server in background
	go func() {
		fmt.Println("\nStarting monitoring server on :9090...")
		if err := monServer.Start(":9090"); err != nil && err != http.ErrServerClosed {
			log.Printf("Monitoring server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// 4. Simulate some subsystem activity
	fmt.Println("\nSimulating subsystem activity...")

	// Simulate API calls
	for i := 0; i < 10; i++ {
		_ = apiPool.GetMetrics()
	}
	fmt.Println("✓ API activity simulated")

	// Simulate file operations
	for i := 0; i < 5; i++ {
		_ = fileManager.GetMetrics()
	}
	fmt.Println("✓ File operations simulated")

	// Simulate browser operations
	for i := 0; i < 3; i++ {
		_ = browserPool.GetMetrics()
	}
	fmt.Println("✓ Browser operations simulated")

	// Simulate process operations
	for i := 0; i < 7; i++ {
		_ = processManager.GetMetrics()
	}
	fmt.Println("✓ Process operations simulated")

	// Simulate network operations
	for i := 0; i < 4; i++ {
		_ = networkCoord.GetMetrics()
	}
	fmt.Println("✓ Network operations simulated")

	// 5. Check health endpoints
	fmt.Println("\n" + string([]byte{61, 61, 61}) + " HEALTH CHECK ENDPOINTS " + string([]byte{61, 61, 61}))

	healthEndpoints := []string{
		"/health",
		"/health/api",
		"/health/file",
		"/health/browser",
		"/health/process",
		"/health/network",
		"/readiness",
		"/liveness",
	}

	for _, endpoint := range healthEndpoints {
		resp, err := http.Get("http://localhost:9090" + endpoint)
		if err != nil {
			fmt.Printf("❌ %s - Error: %v\n", endpoint, err)
			continue
		}
		defer resp.Body.Close()

		status := "✓"
		if resp.StatusCode >= 400 {
			status = "❌"
		}
		fmt.Printf("%s %s - Status: %d\n", status, endpoint, resp.StatusCode)
	}

	// 6. Check Prometheus metrics endpoint
	fmt.Println("\n" + string([]byte{61, 61, 61}) + " PROMETHEUS METRICS " + string([]byte{61, 61, 61}))
	resp, err := http.Get("http://localhost:9090/metrics")
	if err != nil {
		fmt.Printf("❌ Failed to fetch metrics: %v\n", err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("✓ Metrics endpoint - Status: %d\n", resp.StatusCode)

		// Count metrics in response
		// Note: In production, use Prometheus client library for parsing
		fmt.Println("\nMetrics available on http://localhost:9090/metrics")
		fmt.Println("Import this URL into Prometheus with:")
		fmt.Println("  - static_configs: [targets: ['localhost:9090']]")
		fmt.Println("  - metrics_path: '/metrics'")
	}

	// 7. Check pprof endpoints (profiling)
	fmt.Println("\n" + string([]byte{61, 61, 61}) + " PROFILING ENDPOINTS " + string([]byte{61, 61, 61}))
	pprofEndpoints := []string{
		"/debug/pprof/",
		"/debug/pprof/heap",
		"/debug/pprof/goroutine",
		"/debug/pprof/threadcreate",
		"/debug/pprof/block",
		"/debug/pprof/mutex",
		"/debug/pprof/allocs",
	}

	for _, endpoint := range pprofEndpoints {
		resp, err := http.Get("http://localhost:9090" + endpoint)
		if err != nil {
			fmt.Printf("❌ %s - Error: %v\n", endpoint, err)
			continue
		}
		defer resp.Body.Close()

		status := "✓"
		if resp.StatusCode >= 400 {
			status = "❌"
		}
		fmt.Printf("%s %s - Status: %d\n", status, endpoint, resp.StatusCode)
	}

	// 8. Display usage instructions
	fmt.Println("\n" + string([]byte{61, 61, 61}) + " MONITORING SETUP INSTRUCTIONS " + string([]byte{61, 61, 61}))
	fmt.Println(`
1. Start Prometheus with the configuration file:
   prometheus --config.file=configs/prometheus.yml

2. View Prometheus UI:
   http://localhost:9090

3. View alerts configuration:
   Check configs/alerts.yml for 18 alert rules

4. Start Grafana:
   docker run -d -p 3000:3000 grafana/grafana

5. Add Prometheus as data source in Grafana:
   - URL: http://localhost:9090
   - Access: Browser

6. Import Grafana dashboards from configs/grafana_dashboards/

7. View profiling data (CPU, memory, goroutines):
   go tool pprof http://localhost:9090/debug/pprof/heap
   go tool pprof http://localhost:9090/debug/pprof/profile?seconds=30

8. Monitor system health:
   curl http://localhost:9090/health | jq .

Monitoring Infrastructure Endpoints:
├─ Health Checks: /health, /health/{subsystem}, /readiness, /liveness
├─ Prometheus Metrics: /metrics (scrape this in Prometheus)
└─ Profiling (pprof): /debug/pprof/*

Prometheus Metrics Available:
├─ API: gaia_api_requests_total, gaia_api_latency_ms, gaia_api_active_clients
├─ File: gaia_file_operations_total, gaia_file_latency_ms, gaia_file_concurrent_operations
├─ Browser: gaia_browser_instances_active, gaia_browser_active_tabs
├─ Process: gaia_process_active, gaia_process_memory_mb, gaia_process_cpu_percent
├─ Network: gaia_network_bytes_transferred_total, gaia_network_active_connections
└─ System: gaia_system_throughput_ops_per_sec, gaia_system_memory_mb, gaia_system_goroutines

Alert Rules (18 total):
├─ API: Error rate, latency, pool exhaustion, rate limiting
├─ File: Error rate, latency, concurrency
├─ Browser: Error rate, pool exhaustion, tab limits
├─ Process: Failure rate, memory, CPU
├─ Network: Connection limits, DNS cache hit rate
└─ System: Throughput, memory, goroutine leaks
`)

	// 9. Keep server running
	fmt.Println("Monitoring server running. Press Ctrl+C to exit.")
	fmt.Println("Visit http://localhost:9090 for health checks and metrics.")

	select {}
}
