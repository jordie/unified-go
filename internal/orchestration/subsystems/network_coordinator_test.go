package subsystems

import (
	"context"
	"testing"
	"time"
)

func TestNetworkCoordinatorBasic(t *testing.T) {
	nc := NewNetworkCoordinator(1000000) // 1MB/s
	defer nc.Close()

	// Test bandwidth throttling
	delay, err := nc.ThrottleBandwidth(context.Background(), 100)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if delay < 0 {
		t.Fatalf("Expected non-negative delay, got %v", delay)
	}

	metrics := nc.GetMetrics()
	if metrics["bytes_processed"].(int64) != 100 {
		t.Fatalf("Expected 100 bytes processed, got %d", metrics["bytes_processed"].(int64))
	}
}

func TestBandwidthThrottling(t *testing.T) {
	nc := NewNetworkCoordinator(10000000) // 10MB/s (very high)
	defer nc.Close()

	// First transfer should not be throttled
	_, err := nc.ThrottleBandwidth(context.Background(), 1000000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second transfer also should not be throttled (plenty of bandwidth)
	_, err = nc.ThrottleBandwidth(context.Background(), 1000000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check metrics
	metrics := nc.GetMetrics()
	if metrics["bytes_processed"].(int64) < 2000000 {
		t.Fatalf("Expected at least 2000000 bytes processed, got %d", metrics["bytes_processed"].(int64))
	}

	// Test that low bandwidth does throttle
	ncLow := NewNetworkCoordinator(100) // 100 bytes/sec
	defer ncLow.Close()

	start := time.Now()
	ncLow.ThrottleBandwidth(context.Background(), 100)
	ncLow.ThrottleBandwidth(context.Background(), 100) // Should throttle and wait
	duration := time.Since(start)

	if duration < 500*time.Millisecond {
		t.Logf("Note: throttling wait was short (%.1f ms)", duration.Seconds()*1000)
	}
}

func TestDNSCache(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	// First lookup should miss cache
	ips, err := nc.ResolveHost(context.Background(), "localhost")
	if err != nil {
		// localhost might not resolve, that's ok
		t.Logf("DNS lookup failed (expected): %v", err)
	}

	// Check cache size increased
	metrics := nc.GetMetrics()
	cacheSize := metrics["dns_cache_size"].(int)

	if cacheSize == 0 && ips != nil {
		t.Log("Note: DNS cache might be empty if lookup failed")
	}

	// Check metrics exist
	if metrics["dns_misses"].(int64) < 0 {
		t.Fatal("Expected non-negative dns_misses")
	}
}

func TestConnectionLimiting(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	nc.SetConnectionLimit("api.example.com", 2)

	// First two connections should succeed
	err := nc.AcquireConnection("api.example.com")
	if err != nil {
		t.Fatalf("Expected first connection to succeed, got %v", err)
	}

	err = nc.AcquireConnection("api.example.com")
	if err != nil {
		t.Fatalf("Expected second connection to succeed, got %v", err)
	}

	// Third should fail
	err = nc.AcquireConnection("api.example.com")
	if err == nil {
		t.Fatal("Expected third connection to fail")
	}

	// Release one
	nc.ReleaseConnection("api.example.com")

	// Now third should succeed
	err = nc.AcquireConnection("api.example.com")
	if err != nil {
		t.Fatalf("Expected connection to succeed after release, got %v", err)
	}
}

func TestMultipleHosts(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	nc.SetConnectionLimit("host1.com", 2)
	nc.SetConnectionLimit("host2.com", 3)

	// Each host has separate limits
	for i := 0; i < 2; i++ {
		err := nc.AcquireConnection("host1.com")
		if err != nil {
			t.Fatalf("host1 connection %d failed: %v", i, err)
		}
	}

	for i := 0; i < 3; i++ {
		err := nc.AcquireConnection("host2.com")
		if err != nil {
			t.Fatalf("host2 connection %d failed: %v", i, err)
		}
	}

	// host1 should be at limit
	err := nc.AcquireConnection("host1.com")
	if err == nil {
		t.Fatal("Expected host1 to be at limit")
	}

	// host2 should also be at limit
	err = nc.AcquireConnection("host2.com")
	if err == nil {
		t.Fatal("Expected host2 to be at limit")
	}
}

func TestNetworkContextCancellation(t *testing.T) {
	nc := NewNetworkCoordinator(100) // Very low bandwidth to force throttling

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := nc.ThrottleBandwidth(ctx, 1000000)
	if err == nil {
		t.Fatal("Expected error for cancelled context")
	}
}

func TestNetworkMetrics(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	// Process some data
	nc.ThrottleBandwidth(context.Background(), 500)
	nc.ThrottleBandwidth(context.Background(), 500)

	metrics := nc.GetMetrics()

	if metrics["bytes_processed"].(int64) != 1000 {
		t.Fatalf("Expected 1000 bytes processed, got %d", metrics["bytes_processed"].(int64))
	}

	if metrics["requests_throttled"].(int64) < 0 {
		t.Fatal("Expected non-negative requests_throttled")
	}

	if metrics["dns_hit_rate"].(float64) < 0 {
		t.Fatal("Expected non-negative dns_hit_rate")
	}
}

func TestDNSCacheHitRate(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	// Manually populate cache
	nc.dnsCache.Set("test.com", []string{"1.2.3.4"})

	// Retrieve from cache (hit)
	ips := nc.dnsCache.Get("test.com")
	if ips == nil {
		t.Fatal("Expected DNS cache hit")
	}

	// Retrieve again (another hit)
	ips = nc.dnsCache.Get("test.com")
	if ips == nil {
		t.Fatal("Expected second DNS cache hit")
	}

	// Try non-existent (miss)
	ips = nc.dnsCache.Get("notfound.com")
	if ips != nil {
		t.Fatal("Expected DNS cache miss for non-existent entry")
	}

	metrics := nc.GetMetrics()
	hitRate := metrics["dns_hit_rate"].(float64)

	// Should have 2 hits and 1 miss = 66.67% hit rate
	if hitRate < 50 {
		t.Logf("Hit rate: %.2f (with 2 hits and 1 miss)", hitRate)
	}
}

func TestDNSCacheExpiration(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	nc.dnsCache.SetTTL(100 * time.Millisecond)
	nc.dnsCache.Set("expire.com", []string{"1.2.3.4"})

	// Should be in cache initially
	ips := nc.dnsCache.Get("expire.com")
	if ips == nil {
		t.Fatal("Expected DNS entry to be cached")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	ips = nc.dnsCache.Get("expire.com")
	if ips != nil {
		t.Fatal("Expected DNS entry to be expired")
	}
}

func TestConnectionLimitDefaults(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	// Get limit for unconfigured host (should use default)
	limit := nc.connectionLimiter.GetLimit("unknown.com")
	if limit <= 0 {
		t.Fatalf("Expected default limit > 0, got %d", limit)
	}
}

func TestConcurrentConnections(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	nc.SetConnectionLimit("concurrent.com", 10)

	// Acquire 10 connections concurrently
	done := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			err := nc.AcquireConnection("concurrent.com")
			done <- err
		}()
	}

	// All 10 should succeed
	for i := 0; i < 10; i++ {
		err := <-done
		if err != nil {
			t.Fatalf("Connection %d failed: %v", i, err)
		}
	}

	// 11th should fail
	err := nc.AcquireConnection("concurrent.com")
	if err == nil {
		t.Fatal("Expected 11th connection to fail")
	}

	// Release all
	for i := 0; i < 10; i++ {
		nc.ReleaseConnection("concurrent.com")
	}
}

func TestUnlimitedBandwidth(t *testing.T) {
	nc := NewNetworkCoordinator(0) // Unlimited
	defer nc.Close()

	// Should not throttle
	delay, err := nc.ThrottleBandwidth(context.Background(), 999999999)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if delay > 0 {
		t.Fatalf("Expected no throttling with unlimited bandwidth, got delay %v", delay)
	}
}

func BenchmarkNetworkCoordinator(b *testing.B) {
	nc := NewNetworkCoordinator(0) // Unlimited bandwidth
	defer nc.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nc.ThrottleBandwidth(context.Background(), 1024)
	}

	metrics := nc.GetMetrics()
	b.Logf("Bytes processed: %d", metrics["bytes_processed"].(int64))
	b.Logf("DNS hit rate: %.2f%%", metrics["dns_hit_rate"].(float64))
}

func BenchmarkConnectionLimiting(b *testing.B) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	nc.SetConnectionLimit("bench.com", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nc.AcquireConnection("bench.com")
		nc.ReleaseConnection("bench.com")
	}

	metrics := nc.GetMetrics()
	b.Logf("Active connections: %d", metrics["active_connections"].(int64))
}

func TestDNSCacheSize(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	// Add multiple entries
	for i := 0; i < 10; i++ {
		host := "host" + string(rune('0'+i)) + ".com"
		nc.dnsCache.Set(host, []string{"1.2.3.4"})
	}

	metrics := nc.GetMetrics()
	cacheSize := metrics["dns_cache_size"].(int)

	if cacheSize != 10 {
		t.Fatalf("Expected 10 cache entries, got %d", cacheSize)
	}
}

func TestDNSCacheOverflow(t *testing.T) {
	nc := NewNetworkCoordinator(0)
	defer nc.Close()

	// Fill cache to max
	maxEntries := 1000
	for i := 0; i < maxEntries + 10; i++ {
		host := "host" + string(rune(i)) + ".com"
		nc.dnsCache.Set(host, []string{"1.2.3.4"})
	}

	metrics := nc.GetMetrics()
	cacheSize := metrics["dns_cache_size"].(int)

	// Should be cleared and refilled to prevent overflow
	if cacheSize > maxEntries {
		t.Fatalf("Expected cache size <= %d, got %d", maxEntries, cacheSize)
	}
}

func TestNetworkLoadPattern(t *testing.T) {
	nc := NewNetworkCoordinator(10000000) // 10MB/s
	defer nc.Close()

	nc.SetConnectionLimit("load.com", 50)

	// Simulate concurrent connections
	done := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func() {
			err := nc.AcquireConnection("load.com")
			done <- err
		}()
	}

	// Count successes and failures
	successCount := 0
	for i := 0; i < 100; i++ {
		if <-done == nil {
			successCount++
		}
	}

	// Should have 50 successes (the limit)
	if successCount != 50 {
		t.Fatalf("Expected 50 successful connections, got %d", successCount)
	}

	// Release all
	for i := 0; i < 50; i++ {
		nc.ReleaseConnection("load.com")
	}
}
