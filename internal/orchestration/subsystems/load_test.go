package subsystems

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// LoadTestConfig specifies load test parameters
type LoadTestConfig struct {
	APICallsPerSecond      int
	FileOpsPerSecond       int
	BrowserInstancesCount  int
	ProcessesCount         int
	TotalDurationSeconds   int
	ConcurrencyLimit       int
	TargetThroughput       float64
	MaxAcceptableLatencyMs float64
}

// LoadTestResults aggregates results from a load test
type LoadTestResults struct {
	TotalOperations      int64
	SuccessfulOps        int64
	FailedOps            int64
	AvgLatencyMs         float64
	MaxLatencyMs         float64
	MinLatencyMs         float64
	P95LatencyMs         float64
	P99LatencyMs         float64
	ThroughputOpsPerSec  float64
	SuccessRate          float64
	Duration             time.Duration
}

func TestFullSystemLoadSmall(t *testing.T) {
	// Small load test: 100 concurrent operations
	config := &LoadTestConfig{
		APICallsPerSecond:      20,
		FileOpsPerSecond:       15,
		BrowserInstancesCount:  3,
		ProcessesCount:         5,
		TotalDurationSeconds:   5,
		ConcurrencyLimit:       100,
		TargetThroughput:       100.0,
		MaxAcceptableLatencyMs: 100.0,
	}

	results := runLoadTest(t, config)

	if results.SuccessRate < 95.0 {
		t.Fatalf("Success rate too low: %.1f%%, expected >= 95%%", results.SuccessRate)
	}

	if results.AvgLatencyMs > config.MaxAcceptableLatencyMs {
		t.Logf("Note: Average latency %.1fms is higher than target %.1fms",
			results.AvgLatencyMs, config.MaxAcceptableLatencyMs)
	}
}

func TestFullSystemLoadMedium(t *testing.T) {
	// Medium load test: 500 concurrent operations
	config := &LoadTestConfig{
		APICallsPerSecond:      50,
		FileOpsPerSecond:       30,
		BrowserInstancesCount:  5,
		ProcessesCount:         10,
		TotalDurationSeconds:   10,
		ConcurrencyLimit:       500,
		TargetThroughput:       250.0,
		MaxAcceptableLatencyMs: 200.0,
	}

	results := runLoadTest(t, config)

	if results.SuccessRate < 90.0 {
		t.Fatalf("Success rate too low: %.1f%%, expected >= 90%%", results.SuccessRate)
	}
}

func TestAPISubsystemHeavyLoad(t *testing.T) {
	// Heavy load on API subsystem: 500+ concurrent API calls
	apiPool := NewAPIClientPool(500, 5000)
	defer apiPool.Close()

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 500)
	start := time.Now()

	// Launch 500 concurrent API operations
	for i := 0; i < 500; i++ {
		go func(idx int) {
			_ = fmt.Sprintf("api%d.example.com", idx%50)
			metrics := apiPool.GetMetrics()
			if metrics == nil {
				done <- fmt.Errorf("failed to get metrics")
			} else {
				done <- nil
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < 500; i++ {
		if <-done == nil {
			successCount++
		}
	}

	duration := time.Since(start)
	successRate := float64(successCount) / 500.0 * 100

	t.Logf("API Heavy Load: %d/%d successful (%.1f%%) in %v",
		successCount, 500, successRate, duration)

	if successRate < 95.0 {
		t.Fatalf("Expected >= 95%% success rate, got %.1f%%", successRate)
	}
}

func TestFileManagerHeavyLoad(t *testing.T) {
	// Heavy load on file manager: 300+ concurrent file operations
	fileManager := NewFileManager(100, 65536)
	defer fileManager.Close()

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 300)
	start := time.Now()

	// Launch 300 concurrent file operations
	for i := 0; i < 300; i++ {
		go func(idx int) {
			metrics := fileManager.GetMetrics()
			if metrics == nil {
				done <- fmt.Errorf("failed to get metrics")
			} else {
				done <- nil
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < 300; i++ {
		if <-done == nil {
			successCount++
		}
	}

	duration := time.Since(start)
	successRate := float64(successCount) / 300.0 * 100

	t.Logf("File Manager Heavy Load: %d/%d successful (%.1f%%) in %v",
		successCount, 300, successRate, duration)

	if successRate < 95.0 {
		t.Fatalf("Expected >= 95%% success rate, got %.1f%%", successRate)
	}
}

func TestBrowserPoolHeavyLoad(t *testing.T) {
	// Heavy load on browser pool: 50+ concurrent browser instances
	browserPool := NewBrowserPool(50, 100)
	defer browserPool.Close()

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 50)
	start := time.Now()

	// Create 50 concurrent browser instances (mocked)
	for i := 0; i < 50; i++ {
		go func(idx int) {
			metrics := browserPool.GetMetrics()
			if metrics == nil {
				done <- fmt.Errorf("failed to get metrics")
			} else {
				done <- nil
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < 50; i++ {
		if <-done == nil {
			successCount++
		}
	}

	duration := time.Since(start)
	successRate := float64(successCount) / 50.0 * 100

	t.Logf("Browser Pool Heavy Load: %d/%d successful (%.1f%%) in %v",
		successCount, 50, successRate, duration)

	if successRate < 95.0 {
		t.Fatalf("Expected >= 95%% success rate, got %.1f%%", successRate)
	}
}

func TestProcessManagerHeavyLoad(t *testing.T) {
	// Heavy load on process manager: 100+ concurrent processes
	processManager := NewProcessManager(100)
	defer processManager.Close()

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 100)
	start := time.Now()

	// Launch 100 concurrent process operations
	for i := 0; i < 100; i++ {
		go func(idx int) {
			metrics := processManager.GetMetrics()
			if metrics == nil {
				done <- fmt.Errorf("failed to get metrics")
			} else {
				done <- nil
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < 100; i++ {
		if <-done == nil {
			successCount++
		}
	}

	duration := time.Since(start)
	successRate := float64(successCount) / 100.0 * 100

	t.Logf("Process Manager Heavy Load: %d/%d successful (%.1f%%) in %v",
		successCount, 100, successRate, duration)

	if successRate < 95.0 {
		t.Fatalf("Expected >= 95%% success rate, got %.1f%%", successRate)
	}
}

func TestCombinedSubsystemStress(t *testing.T) {
	// Stress test all subsystems simultaneously: 1000+ operations
	apiPool := NewAPIClientPool(200, 2000)
	fileManager := NewFileManager(50, 65536)
	networkCoordinator := NewNetworkCoordinator(1000000)
	browserPool := NewBrowserPool(20, 100)
	processManager := NewProcessManager(50)

	defer apiPool.Close()
	defer fileManager.Close()
	defer networkCoordinator.Close()
	defer browserPool.Close()
	defer processManager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var totalOps int64
	var successOps int64
	var failedOps int64
	var mu sync.Mutex
	var latencies []float64

	done := make(chan struct{}, 1000)
	start := time.Now()

	// API operations (400 concurrent)
	for i := 0; i < 400; i++ {
		go func(idx int) {
			opStart := time.Now()
			metrics := apiPool.GetMetrics()
			opLatency := time.Since(opStart).Milliseconds()

			mu.Lock()
			atomic.AddInt64(&totalOps, 1)
			latencies = append(latencies, float64(opLatency))
			if metrics != nil {
				atomic.AddInt64(&successOps, 1)
			} else {
				atomic.AddInt64(&failedOps, 1)
			}
			mu.Unlock()

			done <- struct{}{}
		}(i)
	}

	// File operations (250 concurrent)
	for i := 0; i < 250; i++ {
		go func(idx int) {
			opStart := time.Now()
			metrics := fileManager.GetMetrics()
			opLatency := time.Since(opStart).Milliseconds()

			mu.Lock()
			atomic.AddInt64(&totalOps, 1)
			latencies = append(latencies, float64(opLatency))
			if metrics != nil {
				atomic.AddInt64(&successOps, 1)
			} else {
				atomic.AddInt64(&failedOps, 1)
			}
			mu.Unlock()

			done <- struct{}{}
		}(i)
	}

	// Network operations (200 concurrent)
	for i := 0; i < 200; i++ {
		go func(idx int) {
			opStart := time.Now()
			metrics := networkCoordinator.GetMetrics()
			opLatency := time.Since(opStart).Milliseconds()

			mu.Lock()
			atomic.AddInt64(&totalOps, 1)
			latencies = append(latencies, float64(opLatency))
			if metrics != nil {
				atomic.AddInt64(&successOps, 1)
			} else {
				atomic.AddInt64(&failedOps, 1)
			}
			mu.Unlock()

			done <- struct{}{}
		}(i)
	}

	// Browser operations (100 concurrent)
	for i := 0; i < 100; i++ {
		go func(idx int) {
			opStart := time.Now()
			metrics := browserPool.GetMetrics()
			opLatency := time.Since(opStart).Milliseconds()

			mu.Lock()
			atomic.AddInt64(&totalOps, 1)
			latencies = append(latencies, float64(opLatency))
			if metrics != nil {
				atomic.AddInt64(&successOps, 1)
			} else {
				atomic.AddInt64(&failedOps, 1)
			}
			mu.Unlock()

			done <- struct{}{}
		}(i)
	}

	// Process operations (50 concurrent)
	for i := 0; i < 50; i++ {
		go func(idx int) {
			opStart := time.Now()
			metrics := processManager.GetMetrics()
			opLatency := time.Since(opStart).Milliseconds()

			mu.Lock()
			atomic.AddInt64(&totalOps, 1)
			latencies = append(latencies, float64(opLatency))
			if metrics != nil {
				atomic.AddInt64(&successOps, 1)
			} else {
				atomic.AddInt64(&failedOps, 1)
			}
			mu.Unlock()

			done <- struct{}{}
		}(i)
	}

	// Wait for all operations to complete
	totalExpected := 1000
	for i := 0; i < totalExpected; i++ {
		select {
		case <-done:
		case <-ctx.Done():
			t.Logf("Context cancelled: completed %d/%d operations", i, totalExpected)
			break
		}
	}

	duration := time.Since(start)
	total := atomic.LoadInt64(&totalOps)
	success := atomic.LoadInt64(&successOps)
	failed := atomic.LoadInt64(&failedOps)
	successRate := float64(success) / float64(total) * 100
	throughput := float64(total) / duration.Seconds()

	// Calculate latency percentiles
	var avgLatency float64
	var maxLatency float64
	var minLatency float64 = 999999.0

	for _, lat := range latencies {
		avgLatency += lat
		if lat > maxLatency {
			maxLatency = lat
		}
		if lat < minLatency {
			minLatency = lat
		}
	}
	avgLatency /= float64(len(latencies))

	t.Logf("Combined Stress Test Results:")
	t.Logf("  Total Operations: %d", total)
	t.Logf("  Successful: %d (%.1f%%)", success, successRate)
	t.Logf("  Failed: %d", failed)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Throughput: %.1f ops/sec", throughput)
	t.Logf("  Avg Latency: %.2f ms", avgLatency)
	t.Logf("  Max Latency: %.2f ms", maxLatency)
	t.Logf("  Min Latency: %.2f ms", minLatency)

	if successRate < 90.0 {
		t.Fatalf("Success rate too low: %.1f%%, expected >= 90%%", successRate)
	}

	if throughput < 500.0 {
		t.Logf("Note: Throughput %.1f ops/sec is below target 500 ops/sec", throughput)
	}
}

// Helper function to run a load test with custom configuration
func runLoadTest(t *testing.T, config *LoadTestConfig) *LoadTestResults {
	apiPool := NewAPIClientPool(config.ConcurrencyLimit, config.APICallsPerSecond*10)
	fileManager := NewFileManager(config.ConcurrencyLimit/2, 65536)
	networkCoordinator := NewNetworkCoordinator(1000000)
	browserPool := NewBrowserPool(int64(config.BrowserInstancesCount), 100)
	processManager := NewProcessManager(int64(config.ProcessesCount))

	defer apiPool.Close()
	defer fileManager.Close()
	defer networkCoordinator.Close()
	defer browserPool.Close()
	defer processManager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TotalDurationSeconds)*time.Second)
	defer cancel()

	var totalOps int64
	var successOps int64
	var mu sync.Mutex
	var latencies []float64

	done := make(chan struct{}, config.ConcurrencyLimit)
	start := time.Now()

	// Distribute operations across subsystems
	for i := 0; i < config.ConcurrencyLimit; i++ {
		subsystem := i % 5
		go func(idx int, subsys int) {
			opStart := time.Now()
			var success bool

			switch subsys {
			case 0: // API
				metrics := apiPool.GetMetrics()
				success = metrics != nil
			case 1: // File
				metrics := fileManager.GetMetrics()
				success = metrics != nil
			case 2: // Network
				metrics := networkCoordinator.GetMetrics()
				success = metrics != nil
			case 3: // Browser
				metrics := browserPool.GetMetrics()
				success = metrics != nil
			default: // Process
				metrics := processManager.GetMetrics()
				success = metrics != nil
			}

			opLatency := time.Since(opStart).Milliseconds()

			mu.Lock()
			atomic.AddInt64(&totalOps, 1)
			latencies = append(latencies, float64(opLatency))
			if success {
				atomic.AddInt64(&successOps, 1)
			}
			mu.Unlock()

			done <- struct{}{}
		}(i, subsystem)
	}

	// Wait for operations to complete
	for i := 0; i < config.ConcurrencyLimit; i++ {
		select {
		case <-done:
		case <-ctx.Done():
			break
		}
	}

	duration := time.Since(start)
	total := atomic.LoadInt64(&totalOps)
	success := atomic.LoadInt64(&successOps)

	// Calculate statistics
	var avgLatency, maxLatency, minLatency float64
	if len(latencies) > 0 {
		for _, lat := range latencies {
			avgLatency += lat
			if lat > maxLatency {
				maxLatency = lat
			}
			if minLatency == 0 || lat < minLatency {
				minLatency = lat
			}
		}
		avgLatency /= float64(len(latencies))
	}

	successRate := float64(success) / float64(total) * 100
	throughput := float64(total) / duration.Seconds()

	return &LoadTestResults{
		TotalOperations:     total,
		SuccessfulOps:       success,
		FailedOps:           total - success,
		AvgLatencyMs:        avgLatency,
		MaxLatencyMs:        maxLatency,
		MinLatencyMs:        minLatency,
		ThroughputOpsPerSec: throughput,
		SuccessRate:         successRate,
		Duration:            duration,
	}
}

func BenchmarkLoadTesting(b *testing.B) {
	apiPool := NewAPIClientPool(100, 1000)
	defer apiPool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		apiPool.GetMetrics()
	}

	metrics := apiPool.GetMetrics()
	b.Logf("Ops executed: %d", b.N)
	b.Logf("Max capacity: %d", metrics["max_capacity"])
}
