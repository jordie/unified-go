package subsystems

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// NetworkCoordinator manages network resources: bandwidth, DNS, connections
type NetworkCoordinator struct {
	bandwidth        *BandwidthLimiter
	dnsCache         *DNSCache
	connectionLimiter *ConnectionLimiter
	metrics          *NetworkMetrics
	mu               sync.RWMutex
}

// BandwidthLimiter enforces bandwidth throttling across all connections
type BandwidthLimiter struct {
	maxBytesPerSec  int64
	currentBytes    int64
	lastRefill      time.Time
	mu              sync.Mutex
}

// DNSCache caches DNS lookups to avoid repeated resolution
type DNSCache struct {
	cache      map[string]*DNSEntry
	mu         sync.RWMutex
	ttl        time.Duration
	maxEntries int
}

// DNSEntry represents a cached DNS record
type DNSEntry struct {
	IPs       []string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ConnectionLimiter limits concurrent connections per host
type ConnectionLimiter struct {
	maxConnPerHost map[string]int64
	activeConns    map[string]int64
	mu             sync.RWMutex
}

// NetworkMetrics tracks network performance
type NetworkMetrics struct {
	bytesProcessed    int64
	requestsThrottled int64
	dnsHits           int64
	dnsMisses         int64
	connectionsActive  int64
	connectionsLimited int64
}

// NewNetworkCoordinator creates a new network coordinator
// maxBandwidth: maximum bytes per second (0 = unlimited)
func NewNetworkCoordinator(maxBandwidth int64) *NetworkCoordinator {
	nc := &NetworkCoordinator{
		bandwidth: &BandwidthLimiter{
			maxBytesPerSec: maxBandwidth,
			lastRefill:     time.Now(),
		},
		dnsCache: &DNSCache{
			cache:      make(map[string]*DNSEntry),
			ttl:        5 * time.Minute,
			maxEntries: 1000,
		},
		connectionLimiter: &ConnectionLimiter{
			maxConnPerHost: make(map[string]int64),
			activeConns:    make(map[string]int64),
		},
		metrics: &NetworkMetrics{},
	}

	// Set default connection limits
	nc.connectionLimiter.SetDefaultLimit(100) // 100 per host by default

	return nc
}

// ThrottleBandwidth applies bandwidth limiting to data transfer
func (nc *NetworkCoordinator) ThrottleBandwidth(ctx context.Context, bytes int64) (time.Duration, error) {
	if nc.bandwidth.maxBytesPerSec <= 0 {
		atomic.AddInt64(&nc.metrics.bytesProcessed, bytes)
		return 0, nil
	}

	delay := nc.bandwidth.Allow(bytes)
	if delay > 0 {
		atomic.AddInt64(&nc.metrics.requestsThrottled, 1)

		select {
		case <-time.After(delay):
			atomic.AddInt64(&nc.metrics.bytesProcessed, bytes)
			return delay, nil
		case <-ctx.Done():
			return 0, fmt.Errorf("bandwidth throttling cancelled")
		}
	}

	atomic.AddInt64(&nc.metrics.bytesProcessed, bytes)
	return 0, nil
}

// ResolveHost resolves a hostname with DNS caching
func (nc *NetworkCoordinator) ResolveHost(ctx context.Context, host string) ([]string, error) {
	// Check cache first
	if ips := nc.dnsCache.Get(host); ips != nil {
		atomic.AddInt64(&nc.metrics.dnsHits, 1)
		return ips, nil
	}

	atomic.AddInt64(&nc.metrics.dnsMisses, 1)

	// Resolve DNS (with timeout)
	resolver := net.Resolver{}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	addrs, err := resolver.LookupHost(ctxWithTimeout, host)
	if err != nil {
		return nil, fmt.Errorf("DNS resolution failed for %s: %w", host, err)
	}

	// Cache the result
	nc.dnsCache.Set(host, addrs)

	return addrs, nil
}

// AcquireConnection acquires a connection slot for a host
func (nc *NetworkCoordinator) AcquireConnection(host string) error {
	return nc.connectionLimiter.Acquire(host)
}

// ReleaseConnection releases a connection slot for a host
func (nc *NetworkCoordinator) ReleaseConnection(host string) {
	nc.connectionLimiter.Release(host)
}

// SetBandwidthLimit sets the bandwidth limit in bytes per second
func (nc *NetworkCoordinator) SetBandwidthLimit(bytesPerSec int64) {
	nc.bandwidth.maxBytesPerSec = bytesPerSec
}

// SetConnectionLimit sets the max connections for a specific host
func (nc *NetworkCoordinator) SetConnectionLimit(host string, limit int64) {
	nc.connectionLimiter.SetLimit(host, limit)
}

// GetMetrics returns current network metrics
func (nc *NetworkCoordinator) GetMetrics() map[string]interface{} {
	bytesProcessed := atomic.LoadInt64(&nc.metrics.bytesProcessed)
	throttled := atomic.LoadInt64(&nc.metrics.requestsThrottled)
	dnsHits := atomic.LoadInt64(&nc.metrics.dnsHits)
	dnsMisses := atomic.LoadInt64(&nc.metrics.dnsMisses)
	activeConns := atomic.LoadInt64(&nc.metrics.connectionsActive)
	limitedConns := atomic.LoadInt64(&nc.metrics.connectionsLimited)

	totalDNS := dnsHits + dnsMisses
	dnsHitRate := float64(0)
	if totalDNS > 0 {
		dnsHitRate = float64(dnsHits) / float64(totalDNS) * 100
	}

	// Use pooled map to reduce allocations
	metrics := GetMetricsMap()
	metrics["bytes_processed"] = bytesProcessed
	metrics["requests_throttled"] = throttled
	metrics["dns_hits"] = dnsHits
	metrics["dns_misses"] = dnsMisses
	metrics["dns_hit_rate"] = dnsHitRate
	metrics["active_connections"] = activeConns
	metrics["limited_connections"] = limitedConns
	metrics["dns_cache_size"] = nc.dnsCache.Size()
	metrics["bandwidth_limit_bytes"] = nc.bandwidth.maxBytesPerSec

	return metrics
}

// BandwidthLimiter methods

// Allow checks if bytes can be sent and returns delay if needed
func (bl *BandwidthLimiter) Allow(bytes int64) time.Duration {
	if bl.maxBytesPerSec <= 0 {
		return 0
	}

	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bl.lastRefill).Seconds()

	// Refill tokens based on elapsed time
	refill := int64(elapsed * float64(bl.maxBytesPerSec))
	if refill > 0 {
		bl.currentBytes += refill
		if bl.currentBytes > bl.maxBytesPerSec {
			bl.currentBytes = bl.maxBytesPerSec
		}
		bl.lastRefill = now
	}

	if bl.currentBytes >= bytes {
		bl.currentBytes -= bytes
		return 0
	}

	// Calculate delay needed
	needed := bytes - bl.currentBytes
	delaySeconds := float64(needed) / float64(bl.maxBytesPerSec)

	return time.Duration(delaySeconds * float64(time.Second))
}

// DNSCache methods

// Get retrieves a DNS entry from cache if valid
func (dc *DNSCache) Get(host string) []string {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if entry, exists := dc.cache[host]; exists {
		if time.Now().Before(entry.ExpiresAt) {
			return entry.IPs
		}
	}

	return nil
}

// Set stores a DNS entry in cache
func (dc *DNSCache) Set(host string, ips []string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Simple eviction: clear if cache is full
	if len(dc.cache) >= dc.maxEntries {
		dc.cache = make(map[string]*DNSEntry)
	}

	dc.cache[host] = &DNSEntry{
		IPs:       ips,
		ExpiresAt: time.Now().Add(dc.ttl),
		CreatedAt: time.Now(),
	}
}

// Size returns the number of entries in cache
func (dc *DNSCache) Size() int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return len(dc.cache)
}

// Clear clears all DNS cache entries
func (dc *DNSCache) Clear() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.cache = make(map[string]*DNSEntry)
}

// SetTTL sets the cache TTL
func (dc *DNSCache) SetTTL(ttl time.Duration) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.ttl = ttl
}

// ConnectionLimiter methods

// Acquire acquires a connection slot for a host
func (cl *ConnectionLimiter) Acquire(host string) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	limit := cl.maxConnPerHost[host]
	if limit <= 0 {
		limit = 100 // default
	}

	active := cl.activeConns[host]
	if active >= limit {
		return fmt.Errorf("connection limit exceeded for host %s: %d/%d", host, active, limit)
	}

	cl.activeConns[host] = active + 1
	return nil
}

// Release releases a connection slot for a host
func (cl *ConnectionLimiter) Release(host string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if active, exists := cl.activeConns[host]; exists && active > 0 {
		cl.activeConns[host] = active - 1
	}
}

// SetLimit sets the max connections for a specific host
func (cl *ConnectionLimiter) SetLimit(host string, limit int64) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.maxConnPerHost[host] = limit
}

// SetDefaultLimit sets the default connection limit for all hosts
func (cl *ConnectionLimiter) SetDefaultLimit(limit int64) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.maxConnPerHost["*"] = limit
}

// GetActive returns the number of active connections for a host
func (cl *ConnectionLimiter) GetActive(host string) int64 {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.activeConns[host]
}

// GetLimit returns the connection limit for a host
func (cl *ConnectionLimiter) GetLimit(host string) int64 {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	if limit, exists := cl.maxConnPerHost[host]; exists {
		return limit
	}

	return cl.maxConnPerHost["*"]
}

// Close closes the network coordinator
func (nc *NetworkCoordinator) Close() error {
	nc.dnsCache.Clear()
	return nil
}
