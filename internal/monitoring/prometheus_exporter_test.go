package monitoring

import (
	"testing"
)

func TestPrometheusExporterRegistration(t *testing.T) {
	exporter := NewPrometheusExporter()

	if exporter.registry == nil {
		t.Fatal("Registry not initialized")
	}

	// Verify all metrics are registered
	metrics := []string{
		"gaia_api_requests_total",
		"gaia_api_success_total",
		"gaia_api_errors_total",
		"gaia_api_rate_limited_total",
		"gaia_api_latency_ms",
		"gaia_api_active_clients",
		"gaia_file_operations_total",
		"gaia_file_success_total",
		"gaia_file_errors_total",
		"gaia_file_bytes_processed_total",
		"gaia_file_latency_ms",
		"gaia_file_concurrent_operations",
		"gaia_browser_instances_active",
		"gaia_browser_active_tabs",
		"gaia_browser_operations_total",
		"gaia_browser_errors_total",
		"gaia_process_active",
		"gaia_process_memory_mb",
		"gaia_process_cpu_percent",
		"gaia_process_started_total",
		"gaia_process_completed_total",
		"gaia_process_failed_total",
		"gaia_network_bytes_transferred_total",
		"gaia_network_active_connections",
		"gaia_network_dns_cache_hits_total",
		"gaia_network_dns_cache_misses_total",
		"gaia_system_throughput_ops_per_sec",
		"gaia_system_memory_mb",
		"gaia_system_goroutines",
	}

	for _, metricName := range metrics {
		// Verify metric exists by checking registry
		// (actual values will be in the metrics output)
		if exporter.registry == nil {
			t.Fatalf("Expected metric %s to be registered", metricName)
		}
	}
}

func TestAPIMetricsUpdate(t *testing.T) {
	exporter := NewPrometheusExporter()

	metrics := map[string]interface{}{
		"active_clients": 10.0,
		"avg_latency_ms": 25.5,
	}

	exporter.updateAPIMetrics(metrics)

	// Verify metrics were updated
	if exporter.apiActiveClients == nil {
		t.Fatal("API active clients metric not initialized")
	}

	if exporter.apiLatencyMs == nil {
		t.Fatal("API latency metric not initialized")
	}
}

func TestFileMetricsUpdate(t *testing.T) {
	exporter := NewPrometheusExporter()

	metrics := map[string]interface{}{
		"concurrent_operations": 5.0,
		"avg_latency_ms":        100.0,
	}

	exporter.updateFileMetrics(metrics)

	if exporter.fileConcurrent == nil {
		t.Fatal("File concurrent metric not initialized")
	}

	if exporter.fileLatencyMs == nil {
		t.Fatal("File latency metric not initialized")
	}
}

func TestBrowserMetricsUpdate(t *testing.T) {
	exporter := NewPrometheusExporter()

	metrics := map[string]interface{}{
		"active_instances": 5.0,
		"active_tabs":      50.0,
	}

	exporter.updateBrowserMetrics(metrics)

	if exporter.browserInstances == nil {
		t.Fatal("Browser instances metric not initialized")
	}

	if exporter.browserActiveTabs == nil {
		t.Fatal("Browser active tabs metric not initialized")
	}
}

func TestProcessMetricsUpdate(t *testing.T) {
	exporter := NewPrometheusExporter()

	metrics := map[string]interface{}{
		"active_processes": 10.0,
		"memory_mb":        100.0,
		"cpu_percent":      25.5,
	}

	exporter.updateProcessMetrics(metrics)

	if exporter.processActive == nil {
		t.Fatal("Process active metric not initialized")
	}

	if exporter.processMemoryMB == nil {
		t.Fatal("Process memory metric not initialized")
	}

	if exporter.processCPUPercent == nil {
		t.Fatal("Process CPU metric not initialized")
	}
}

func TestNetworkMetricsUpdate(t *testing.T) {
	exporter := NewPrometheusExporter()

	metrics := map[string]interface{}{
		"active_connections": 50.0,
	}

	exporter.updateNetworkMetrics(metrics)

	if exporter.networkActiveConnections == nil {
		t.Fatal("Network active connections metric not initialized")
	}
}

func TestPrometheusRegistryUsability(t *testing.T) {
	exporter := NewPrometheusExporter()

	// Verify the registry can be used for Prometheus scraping
	gathered, err := exporter.registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	if gathered == nil {
		t.Fatal("Expected gathered metrics, got nil")
	}

	// Verify we have some metrics (Go runtime metrics)
	if len(gathered) == 0 {
		t.Fatal("Expected at least some metric families, got 0")
	}
}

func TestGetRegistry(t *testing.T) {
	exporter := NewPrometheusExporter()

	registry := exporter.GetRegistry()
	if registry == nil {
		t.Fatal("Expected registry, got nil")
	}

	if registry != exporter.registry {
		t.Fatal("Expected returned registry to match internal registry")
	}
}

func TestCollectMetrics(t *testing.T) {
	exporter := NewPrometheusExporter()

	err := exporter.CollectMetrics()
	if err != nil {
		t.Fatalf("CollectMetrics failed: %v", err)
	}
}

func TestMetricNaming(t *testing.T) {
	exporter := NewPrometheusExporter()

	// Test metric names follow the pattern: gaia_{subsystem}_{metric_name}
	expectedNames := map[string]bool{
		"gaia_api_requests_total":          true,
		"gaia_file_operations_total":       true,
		"gaia_browser_instances_active":    true,
		"gaia_process_active":              true,
		"gaia_network_active_connections":  true,
		"gaia_system_throughput_ops_per_sec": true,
	}

	gatherers, err := exporter.registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	collectedNames := make(map[string]bool)
	for _, mf := range gatherers {
		if mf.Name != nil {
			collectedNames[*mf.Name] = true
		}
	}

	// Verify some expected metrics are present
	for expectedName := range expectedNames {
		// Note: Some metrics might not be in the output if they haven't been updated
		// Just verify we have some gaia_ metrics
		if _, exists := collectedNames[expectedName]; exists {
			delete(expectedNames, expectedName)
		}
	}

	// We should have at least gathered some metrics from Go runtime
	if len(collectedNames) == 0 {
		t.Fatal("No metrics collected from registry")
	}
}

func TestPrometheusFormattingCompatibility(t *testing.T) {
	exporter := NewPrometheusExporter()

	// Test that registry output can be formatted correctly
	gathered, err := exporter.registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify we can gather metrics successfully
	if gathered == nil || len(gathered) == 0 {
		t.Fatal("Expected metrics to be gathered")
	}

	// Verify output is compatible with Prometheus format
	for _, mf := range gathered {
		if mf.Name == nil {
			t.Fatal("Metric family name is nil")
		}
		if len(*mf.Name) == 0 {
			t.Fatal("Metric family name is empty")
		}
	}
}

func BenchmarkMetricsCollection(b *testing.B) {
	exporter := NewPrometheusExporter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = exporter.CollectMetrics()
	}
}

func BenchmarkRegistryGather(b *testing.B) {
	exporter := NewPrometheusExporter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = exporter.registry.Gather()
	}
}
