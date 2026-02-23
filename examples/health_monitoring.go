package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Example: Using GAIA health check endpoints for application monitoring
//
// This example demonstrates how to:
// 1. Check overall system health
// 2. Monitor per-subsystem health
// 3. Use Kubernetes liveness/readiness probes
// 4. Implement automated health dashboard
//
// Run with: go run examples/health_monitoring.go
// Then visit: http://localhost:9092/health-dashboard

// HealthStatus represents the health check response
type HealthStatus struct {
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	Subsystems map[string]interface{} `json:"subsystems,omitempty"`
	Issues     []string               `json:"issues,omitempty"`
}

// SubsystemHealth represents per-subsystem health info
type SubsystemHealth struct {
	Status  string                 `json:"status"`
	Metrics map[string]interface{} `json:"metrics,omitempty"`
}

// HealthMonitor fetches and displays health information
type HealthMonitor struct {
	gaia_url string
	ticker   *time.Ticker
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(gaiaURL string) *HealthMonitor {
	return &HealthMonitor{
		gaia_url: gaiaURL,
		ticker:   time.NewTicker(10 * time.Second),
	}
}

// CheckSystemHealth gets overall system health
func (hm *HealthMonitor) CheckSystemHealth() (*HealthStatus, error) {
	resp, err := http.Get(hm.gaia_url + "/health")
	if err != nil {
		return nil, fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	var health HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health: %w", err)
	}

	return &health, nil
}

// CheckSubsystemHealth gets health for a specific subsystem
func (hm *HealthMonitor) CheckSubsystemHealth(subsystem string) (*SubsystemHealth, error) {
	resp, err := http.Get(fmt.Sprintf("%s/health/%s", hm.gaia_url, subsystem))
	if err != nil {
		return nil, fmt.Errorf("failed to check %s health: %w", subsystem, err)
	}
	defer resp.Body.Close()

	var health SubsystemHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode %s health: %w", subsystem, err)
	}

	return &health, nil
}

// CheckReadiness checks if the service is ready to handle requests
func (hm *HealthMonitor) CheckReadiness() (bool, error) {
	resp, err := http.Get(hm.gaia_url + "/readiness")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// CheckLiveness checks if the service is alive
func (hm *HealthMonitor) CheckLiveness() (bool, error) {
	resp, err := http.Get(hm.gaia_url + "/liveness")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// PrintHealthDashboard prints a formatted health dashboard
func (hm *HealthMonitor) PrintHealthDashboard() {
	health, err := hm.CheckSystemHealth()
	if err != nil {
		log.Printf("Error checking health: %v", err)
		return
	}

	// Clear screen (Unix only)
	fmt.Print("\033[2J\033[H")

	// Header
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           GAIA Phase 8 - System Health Dashboard               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Overall status with color
	statusIcon := "✓"
	statusColor := "\033[32m"  // Green
	if health.Status != "Healthy" {
		statusIcon = "✗"
		statusColor = "\033[31m"  // Red
	}

	fmt.Printf("Overall Status: %s%s %s\033[0m\n", statusColor, statusIcon, health.Status)
	fmt.Printf("Message: %s\n", health.Message)
	fmt.Printf("Timestamp: %s\n", health.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Per-subsystem status
	fmt.Println("Subsystems:")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	subsystems := []string{"api_pool", "file_manager", "browser_pool", "process_manager", "network_coordinator"}

	for _, subsystem := range subsystems {
		sh, err := hm.CheckSubsystemHealth(subsystem)
		if err != nil {
			fmt.Printf("  ✗ %s: ERROR - %v\n", subsystem, err)
			continue
		}

		icon := "✓"
		color := "\033[32m"  // Green
		if sh.Status != "Healthy" {
			icon = "✗"
			color = "\033[31m"  // Red
		}

		fmt.Printf("  %s%s %s %-20s %s\033[0m\n",
			color, icon, subsystem, sh.Status,
			formatMetrics(sh.Metrics))
	}
	fmt.Println()

	// Kubernetes probes
	fmt.Println("Kubernetes Probes:")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	readiness, _ := hm.CheckReadiness()
	liveness, _ := hm.CheckLiveness()

	readinessIcon := "✓"
	readinessColor := "\033[32m"
	if !readiness {
		readinessIcon = "✗"
		readinessColor = "\033[31m"
	}

	livenessIcon := "✓"
	livenessColor := "\033[32m"
	if !liveness {
		livenessIcon = "✗"
		livenessColor = "\033[31m"
	}

	fmt.Printf("  %sReadiness: %s (/readiness)\033[0m\n", readinessColor, readinessIcon)
	fmt.Printf("  %sLiveness:  %s (/liveness)\033[0m\n", livenessColor, livenessIcon)
	fmt.Println()

	// Issues if any
	if len(health.Issues) > 0 {
		fmt.Println("Issues Detected:")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for i, issue := range health.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
		fmt.Println()
	}

	fmt.Println("Last updated: " + time.Now().Format("2006-01-02 15:04:05"))
}

// ServeHealthDashboard serves an HTML health dashboard
func (hm *HealthMonitor) ServeHealthDashboard(w http.ResponseWriter, r *http.Request) {
	health, err := hm.CheckSystemHealth()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>GAIA Health Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; }
        .header { background: #333; color: white; padding: 20px; border-radius: 5px; }
        .status { font-size: 24px; font-weight: bold; margin: 10px 0; }
        .healthy { color: #4CAF50; }
        .degraded { color: #FFC107; }
        .unhealthy { color: #f44336; }
        .subsystem {
            background: white;
            padding: 15px;
            margin: 10px 0;
            border-left: 4px solid #ddd;
            border-radius: 3px;
        }
        .subsystem.healthy { border-left-color: #4CAF50; }
        .subsystem.degraded { border-left-color: #FFC107; }
        .subsystem.unhealthy { border-left-color: #f44336; }
        .metrics { font-size: 12px; color: #666; margin-top: 10px; }
        .refresh-btn {
            padding: 10px 20px;
            background: #2196F3;
            color: white;
            border: none;
            border-radius: 3px;
            cursor: pointer;
        }
        .refresh-btn:hover { background: #0b7dda; }
        .timestamp { color: #999; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>GAIA Phase 8 - System Health</h1>
            <div class="status %s">%s</div>
            <p>%s</p>
        </div>

        <button class="refresh-btn" onclick="location.reload()">Refresh</button>

        <h2>Subsystems</h2>
`,
		health.Status,
		health.Status,
		health.Message,
	)

	// Add subsystems
	subsystems := []string{"api_pool", "file_manager", "browser_pool", "process_manager", "network_coordinator"}
	for _, subsystem := range subsystems {
		sh, err := hm.CheckSubsystemHealth(subsystem)
		if err != nil {
			html += fmt.Sprintf(`<div class="subsystem unhealthy"><strong>%s</strong>: ERROR</div>`, subsystem)
			continue
		}

		statusClass := "healthy"
		if sh.Status != "Healthy" {
			statusClass = "degraded"
		}

		html += fmt.Sprintf(`<div class="subsystem %s"><strong>%s</strong>: %s</div>`, statusClass, subsystem, sh.Status)
	}

	html += fmt.Sprintf(`
        <p class="timestamp">Last updated: %s</p>
    </div>
    <script>
        // Auto-refresh every 10 seconds
        setTimeout(() => location.reload(), 10000);
    </script>
</body>
</html>
`, time.Now().Format("2006-01-02 15:04:05"))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// MonitorContinuously runs continuous health monitoring
func (hm *HealthMonitor) MonitorContinuously() {
	for range hm.ticker.C {
		hm.PrintHealthDashboard()
	}
}

func main() {
	monitor := NewHealthMonitor("http://localhost:9090")

	// Serve web dashboard
	http.HandleFunc("/health-dashboard", monitor.ServeHealthDashboard)

	fmt.Println("Health Monitoring Example")
	fmt.Println()
	fmt.Println("This example demonstrates how to use GAIA health check endpoints.")
	fmt.Println()
	fmt.Println("Endpoints available:")
	fmt.Println("  • Web Dashboard:  http://localhost:9092/health-dashboard")
	fmt.Println("  • GAIA Health:    http://localhost:9090/health")
	fmt.Println("  • Liveness:       http://localhost:9090/liveness")
	fmt.Println("  • Readiness:      http://localhost:9090/readiness")
	fmt.Println()
	fmt.Println("Make sure GAIA monitoring server is running:")
	fmt.Println("  go run examples/monitoring_example.go")
	fmt.Println()
	fmt.Println("Starting health monitoring dashboard on :9092...")
	fmt.Println("Open browser: http://localhost:9092/health-dashboard")
	fmt.Println()

	// Start terminal dashboard
	go monitor.MonitorContinuously()

	// Start web server
	if err := http.ListenAndServe(":9092", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Helper function to format metrics as a string
func formatMetrics(metrics map[string]interface{}) string {
	if len(metrics) == 0 {
		return ""
	}

	// Extract key metrics
	var parts []string
	if sr, ok := metrics["success_rate"].(float64); ok {
		parts = append(parts, fmt.Sprintf("SR:%.1f%%", sr))
	}
	if lat, ok := metrics["avg_latency_ms"].(float64); ok {
		parts = append(parts, fmt.Sprintf("Lat:%.1fms", lat))
	}

	if len(parts) > 0 {
		return "(" + fmt.Sprintf("%v", parts[0]) + ")"
	}
	return ""
}
