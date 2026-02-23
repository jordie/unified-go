package subsystems

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// APIClientPool manages a pool of HTTP clients with rate limiting and circuit breaker pattern
type APIClientPool struct {
	clients      chan *http.Client
	maxClients   int
	globalRPS    int
	rateLimiters map[string]*RateLimiter
	circuitBreaker *CircuitBreaker
	metrics      *PoolMetrics
	mu           sync.RWMutex
}

// RateLimiter implements token bucket algorithm for per-domain rate limiting
type RateLimiter struct {
	maxTokens   float64
	currentTokens float64
	lastRefill  time.Time
	refillRate  float64 // tokens per second
	mu          sync.Mutex
}

// CircuitBreaker tracks endpoint health and prevents cascading failures
type CircuitBreaker struct {
	endpoints map[string]*EndpointState
	mu        sync.RWMutex
	threshold int // consecutive failures before opening circuit
	timeout   time.Duration
}

// EndpointState tracks the state of a single endpoint
type EndpointState struct {
	state            string // "closed", "open", "half-open"
	consecutiveErrors int
	lastError        time.Time
	successCount     int
}

// PoolMetrics tracks performance and health metrics
type PoolMetrics struct {
	successCount  int64
	errorCount    int64
	totalRequests int64
	totalLatency  int64 // nanoseconds
	rateLimited   int64
}

// RequestResult represents the result of an API call
type RequestResult struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
	Error      error
}

// NewAPIClientPool creates a new HTTP client pool with rate limiting
// maxConnections: size of the connection pool
// globalRPS: global requests per second limit (0 for unlimited)
func NewAPIClientPool(maxConnections, globalRPS int) *APIClientPool {
	pool := &APIClientPool{
		clients:        make(chan *http.Client, maxConnections),
		maxClients:     maxConnections,
		globalRPS:      globalRPS,
		rateLimiters:   make(map[string]*RateLimiter),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
		metrics:        &PoolMetrics{},
	}

	// Pre-create HTTP clients
	for i := 0; i < maxConnections; i++ {
		client := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		}
		pool.clients <- client
	}

	// Start global rate limiter ticker if needed
	if globalRPS > 0 {
		go pool.globalRateLimiterTicker()
	}

	return pool
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		endpoints: make(map[string]*EndpointState),
		threshold: threshold,
		timeout:   timeout,
	}
}

// MakeRequest makes an HTTP request with rate limiting and circuit breaker protection
func (p *APIClientPool) MakeRequest(ctx context.Context, method, url string, body io.Reader) (*RequestResult, error) {
	atomic.AddInt64(&p.metrics.totalRequests, 1)

	// Extract domain for rate limiting
	domain := extractDomain(url)

	// Check circuit breaker
	if !p.circuitBreaker.CanProceed(domain) {
		atomic.AddInt64(&p.metrics.errorCount, 1)
		return nil, fmt.Errorf("circuit breaker open for domain: %s", domain)
	}

	// Apply rate limiting
	rateLimiter := p.getRateLimiter(domain)
	if !rateLimiter.Allow() {
		atomic.AddInt64(&p.metrics.rateLimited, 1)
		return nil, fmt.Errorf("rate limit exceeded for domain: %s", domain)
	}

	// Get a client from pool (with timeout)
	select {
	case client := <-p.clients:
		defer func() { p.clients <- client }()

		// Create request with context
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			p.circuitBreaker.RecordError(domain)
			atomic.AddInt64(&p.metrics.errorCount, 1)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Execute request and measure latency
		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)
		atomic.AddInt64(&p.metrics.totalLatency, duration.Nanoseconds())

		if err != nil {
			p.circuitBreaker.RecordError(domain)
			atomic.AddInt64(&p.metrics.errorCount, 1)
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// Read response body
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			p.circuitBreaker.RecordError(domain)
			atomic.AddInt64(&p.metrics.errorCount, 1)
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check HTTP status for errors
		if resp.StatusCode >= 500 {
			p.circuitBreaker.RecordError(domain)
			atomic.AddInt64(&p.metrics.errorCount, 1)
		} else if resp.StatusCode >= 400 {
			// Client errors don't count as endpoint failures
			p.circuitBreaker.RecordSuccess(domain)
		} else {
			p.circuitBreaker.RecordSuccess(domain)
			atomic.AddInt64(&p.metrics.successCount, 1)
		}

		return &RequestResult{
			StatusCode: resp.StatusCode,
			Body:       respBody,
			Headers:    resp.Header,
			Duration:   duration,
			Error:      nil,
		}, nil

	case <-ctx.Done():
		atomic.AddInt64(&p.metrics.errorCount, 1)
		return nil, fmt.Errorf("context cancelled")
	case <-time.After(5 * time.Second):
		atomic.AddInt64(&p.metrics.errorCount, 1)
		return nil, fmt.Errorf("timeout waiting for available client")
	}
}

// getRateLimiter returns or creates a rate limiter for a domain
func (p *APIClientPool) getRateLimiter(domain string) *RateLimiter {
	p.mu.RLock()
	if limiter, exists := p.rateLimiters[domain]; exists {
		p.mu.RUnlock()
		return limiter
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check again after acquiring write lock
	if limiter, exists := p.rateLimiters[domain]; exists {
		return limiter
	}

	// Default to 100 RPS per domain if global limit is set, otherwise unlimited
	maxRPS := 100
	if p.globalRPS > 0 {
		maxRPS = p.globalRPS / 10 // Distribute global limit across domains
	}

	limiter := NewRateLimiter(maxRPS)
	p.rateLimiters[domain] = limiter
	return limiter
}

// SetRateLimit sets the rate limit for a specific domain
func (p *APIClientPool) SetRateLimit(domain string, rps int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rateLimiters[domain] = NewRateLimiter(rps)
}

// NewRateLimiter creates a new rate limiter with token bucket algorithm
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		maxTokens:     float64(rps),
		currentTokens: float64(rps),
		lastRefill:    time.Now(),
		refillRate:    float64(rps),
	}
}

// Allow checks if a request is allowed under rate limit
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.currentTokens += elapsed * r.refillRate

	if r.currentTokens > r.maxTokens {
		r.currentTokens = r.maxTokens
	}

	r.lastRefill = now

	if r.currentTokens >= 1 {
		r.currentTokens--
		return true
	}

	return false
}

// globalRateLimiterTicker applies global rate limiting across all domains
func (p *APIClientPool) globalRateLimiterTicker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.RLock()
		for _, limiter := range p.rateLimiters {
			limiter.mu.Lock()
			limiter.currentTokens = limiter.maxTokens // Refill each second
			limiter.mu.Unlock()
		}
		p.mu.RUnlock()
	}
}

// extractDomain extracts domain from URL
func extractDomain(url string) string {
	// Simple domain extraction - in production would use net.url.URL
	if start := len("https://"); len(url) > start {
		if i := len("https://"); len(url) > i && url[:i] == "https://" {
			url = url[i:]
		} else if i := len("http://"); len(url) > i && url[:i] == "http://" {
			url = url[i:]
		}
	}
	if i := len(url); i > 0 {
		if slash := -1; slash < len(url) {
			for j := 0; j < len(url); j++ {
				if url[j] == '/' {
					slash = j
					break
				}
			}
			if slash > 0 {
				return url[:slash]
			}
		}
	}
	return url
}

// CircuitBreaker methods

// CanProceed checks if a request can proceed based on circuit breaker state
func (cb *CircuitBreaker) CanProceed(endpoint string) bool {
	cb.mu.RLock()
	state, exists := cb.endpoints[endpoint]
	cb.mu.RUnlock()

	if !exists {
		return true // Unknown endpoints are allowed
	}

	if state.state == "closed" {
		return true
	}

	if state.state == "open" {
		// Check if timeout has passed to transition to half-open
		if time.Since(state.lastError) > cb.timeout {
			cb.mu.Lock()
			state.state = "half-open"
			state.successCount = 0
			cb.mu.Unlock()
			return true
		}
		return false
	}

	// half-open state
	return true
}

// RecordError records a failure for an endpoint
func (cb *CircuitBreaker) RecordError(endpoint string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if _, exists := cb.endpoints[endpoint]; !exists {
		cb.endpoints[endpoint] = &EndpointState{state: "closed"}
	}

	state := cb.endpoints[endpoint]
	state.consecutiveErrors++
	state.lastError = time.Now()

	if state.consecutiveErrors >= cb.threshold {
		state.state = "open"
	}
}

// RecordSuccess records a successful request for an endpoint
func (cb *CircuitBreaker) RecordSuccess(endpoint string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if _, exists := cb.endpoints[endpoint]; !exists {
		cb.endpoints[endpoint] = &EndpointState{state: "closed"}
	}

	state := cb.endpoints[endpoint]
	state.consecutiveErrors = 0

	if state.state == "half-open" {
		state.successCount++
		if state.successCount >= 3 { // Need 3 successes to close circuit
			state.state = "closed"
		}
	}
}

// GetMetrics returns current pool metrics
func (p *APIClientPool) GetMetrics() map[string]interface{} {
	total := atomic.LoadInt64(&p.metrics.totalRequests)
	success := atomic.LoadInt64(&p.metrics.successCount)
	errors := atomic.LoadInt64(&p.metrics.errorCount)
	rateLimited := atomic.LoadInt64(&p.metrics.rateLimited)
	totalLatency := atomic.LoadInt64(&p.metrics.totalLatency)

	var avgLatency float64
	if total > 0 {
		avgLatency = float64(totalLatency) / float64(total)
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	// Use pooled map to reduce allocations
	metrics := GetMetricsMap()
	metrics["total_requests"] = total
	metrics["successful"] = success
	metrics["errors"] = errors
	metrics["rate_limited"] = rateLimited
	metrics["success_rate"] = successRate
	metrics["avg_latency_ns"] = avgLatency
	metrics["avg_latency_ms"] = avgLatency / 1_000_000
	metrics["available_clients"] = len(p.clients)
	metrics["pool_size"] = p.maxClients

	return metrics
}

// Close gracefully closes the pool
func (p *APIClientPool) Close() error {
	close(p.clients)
	return nil
}
