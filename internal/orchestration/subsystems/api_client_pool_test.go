package subsystems

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAPIClientPoolBasic(t *testing.T) {
	pool := NewAPIClientPool(10, 1000)
	defer pool.Close()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Make a request
	result, err := pool.MakeRequest(context.Background(), "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", result.StatusCode)
	}

	if string(result.Body) != "test response" {
		t.Fatalf("Expected 'test response', got %s", string(result.Body))
	}
}

func TestRateLimiting(t *testing.T) {
	limiter := NewRateLimiter(10) // 10 requests per second

	// Should allow first 10 requests immediately
	for i := 0; i < 10; i++ {
		if !limiter.Allow() {
			t.Fatalf("Expected request %d to be allowed", i)
		}
	}

	// 11th request should be denied
	if limiter.Allow() {
		t.Fatal("Expected 11th request to be rate limited")
	}

	// Wait for token refill
	time.Sleep(100 * time.Millisecond)
	if !limiter.Allow() {
		t.Fatal("Expected request to be allowed after refill")
	}
}

func TestConcurrentRequests(t *testing.T) {
	pool := NewAPIClientPool(10, 1000)
	defer pool.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))
	defer server.Close()

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	errorCount := 0

	// Make 50 concurrent requests
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := pool.MakeRequest(context.Background(), "GET", server.URL, nil)
			mu.Lock()
			if err != nil {
				errorCount++
			} else if result.StatusCode == http.StatusOK {
				successCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if successCount == 0 {
		t.Fatalf("Expected some successful requests, got %d", successCount)
	}

	metrics := pool.GetMetrics()
	if metrics["total_requests"].(int64) == 0 {
		t.Fatal("Expected non-zero total requests in metrics")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// Record 3 consecutive errors to open circuit
	for i := 0; i < 3; i++ {
		cb.RecordError("api.example.com")
	}

	// Circuit should be open now
	if cb.CanProceed("api.example.com") {
		t.Fatal("Expected circuit to be open after 3 errors")
	}

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)
	if !cb.CanProceed("api.example.com") {
		t.Fatal("Expected circuit to transition to half-open")
	}

	// Record successes to close circuit
	for i := 0; i < 3; i++ {
		cb.RecordSuccess("api.example.com")
	}

	// Circuit should be closed again
	if !cb.CanProceed("api.example.com") {
		t.Fatal("Expected circuit to be closed after successful requests")
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://api.example.com/users", "api.example.com"},
		{"http://example.com/path", "example.com"},
		{"https://api.example.com:8080/path", "api.example.com:8080"},
	}

	for _, test := range tests {
		result := extractDomain(test.url)
		if result != test.expected {
			t.Fatalf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestPerDomainRateLimit(t *testing.T) {
	pool := NewAPIClientPool(10, 0) // No global limit
	defer pool.Close()

	// Set different limits for different domains
	pool.SetRateLimit("api1.com", 5)
	pool.SetRateLimit("api2.com", 10)

	limiter1 := pool.getRateLimiter("api1.com")
	limiter2 := pool.getRateLimiter("api2.com")

	// Both should allow their max tokens
	for i := 0; i < 5; i++ {
		if !limiter1.Allow() {
			t.Fatalf("Expected api1 request %d to be allowed", i)
		}
	}

	// api1 should be rate limited
	if limiter1.Allow() {
		t.Fatal("Expected api1 to be rate limited after 5 requests")
	}

	// api2 should still have tokens
	for i := 0; i < 5; i++ {
		if !limiter2.Allow() {
			t.Fatalf("Expected api2 request %d to be allowed", i)
		}
	}
}

func TestMetrics(t *testing.T) {
	pool := NewAPIClientPool(10, 1000)
	defer pool.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	// Make some requests
	for i := 0; i < 10; i++ {
		pool.MakeRequest(context.Background(), "GET", server.URL, nil)
	}

	metrics := pool.GetMetrics()

	if metrics["total_requests"].(int64) != 10 {
		t.Fatalf("Expected 10 total requests, got %d", metrics["total_requests"].(int64))
	}

	if metrics["successful"].(int64) != 10 {
		t.Fatalf("Expected 10 successful requests, got %d", metrics["successful"].(int64))
	}

	successRate := metrics["success_rate"].(float64)
	if successRate != 100.0 {
		t.Fatalf("Expected 100%% success rate, got %f%%", successRate)
	}
}

func BenchmarkAPIClientPool(b *testing.B) {
	pool := NewAPIClientPool(100, 10000)
	defer pool.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))
	defer server.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.MakeRequest(context.Background(), "GET", server.URL, nil)
		}
	})

	metrics := pool.GetMetrics()
	b.Logf("Total requests: %d", metrics["total_requests"].(int64))
	b.Logf("Success rate: %.2f%%", metrics["success_rate"].(float64))
	b.Logf("Avg latency: %.2f ms", metrics["avg_latency_ms"].(float64))
}

func TestContextCancellation(t *testing.T) {
	pool := NewAPIClientPool(1, 100)
	defer pool.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := pool.MakeRequest(ctx, "GET", server.URL, nil)
	if err == nil {
		t.Fatal("Expected error for cancelled context")
	}

	if !strings.Contains(err.Error(), "cancelled") && !strings.Contains(err.Error(), "context") {
		t.Fatalf("Expected context cancellation error, got %v", err)
	}
}

func TestPoolExhaustion(t *testing.T) {
	pool := NewAPIClientPool(1, 0)
	defer pool.Close()

	blockingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block for 2 seconds to exhaust the pool
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer blockingServer.Close()

	// Start a long request that will hold the only client
	done := make(chan error)
	go func() {
		_, err := pool.MakeRequest(context.Background(), "GET", blockingServer.URL, nil)
		done <- err
	}()

	time.Sleep(100 * time.Millisecond) // Let first request start and hold client

	// Try to make another request with short timeout (should fail because pool is exhausted)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := pool.MakeRequest(ctx, "GET", blockingServer.URL, nil)
	if err == nil {
		t.Fatal("Expected error when pool is exhausted")
	}

	// Clean up the blocking request
	<-done
}

func TestErrorHandling(t *testing.T) {
	pool := NewAPIClientPool(5, 100)
	defer pool.Close()

	// Server that returns 500 errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	result, err := pool.MakeRequest(context.Background(), "GET", server.URL, nil)
	if err != nil {
		// Error responses might still return result
		if result == nil {
			t.Fatal("Expected result even with error status")
		}
	}

	if result != nil && result.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected 500 status, got %d", result.StatusCode)
	}
}

func TestRateLimitWithDomainExtraction(t *testing.T) {
	pool := NewAPIClientPool(20, 0)
	defer pool.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	// Get the actual domain
	domain := extractDomain(server.URL)
	pool.SetRateLimit(domain, 5) // 5 RPS for this domain

	// Try 10 requests in quick succession
	var successCount int
	for i := 0; i < 10; i++ {
		result, err := pool.MakeRequest(context.Background(), "GET", server.URL, nil)
		if err != nil {
			if strings.Contains(err.Error(), "rate limit") {
				continue
			}
			t.Fatalf("Unexpected error: %v", err)
		}
		if result != nil && result.StatusCode == http.StatusOK {
			successCount++
		}
	}

	// Should have some successes but not all 10
	if successCount == 0 {
		t.Fatal("Expected at least one success")
	}
	if successCount == 10 {
		t.Fatal("Expected some requests to be rate limited")
	}
}

func TestLoadPattern(t *testing.T) {
	pool := NewAPIClientPool(50, 0)
	defer pool.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))
	defer server.Close()

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]bool, 0)

	// Simulate 100 concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := pool.MakeRequest(context.Background(), "GET", server.URL, nil)
			mu.Lock()
			results = append(results, err == nil && result.StatusCode == http.StatusOK)
			mu.Unlock()
		}()
	}

	wg.Wait()

	successCount := 0
	for _, success := range results {
		if success {
			successCount++
		}
	}

	// Should have high success rate despite concurrency
	successRate := float64(successCount) / float64(len(results)) * 100
	if successRate < 95 {
		t.Fatalf("Expected high success rate, got %.2f%%", successRate)
	}
}
