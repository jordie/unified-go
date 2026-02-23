package subsystems

import (
	"sync"
	"sync/atomic"
)

// MetricsPool provides a sync.Pool for reusable metric maps to reduce allocations.
// Each subsystem's GetMetrics() can use this pool instead of creating new maps.
type MetricsPool struct {
	pool      *sync.Pool
	gets      int64
	puts      int64
	hitCount  int64
	missCount int64
}

// PoolStats returns statistics about pool usage
func (mp *MetricsPool) PoolStats() (gets, puts, hits, misses int64) {
	return atomic.LoadInt64(&mp.gets), atomic.LoadInt64(&mp.puts),
		atomic.LoadInt64(&mp.hitCount), atomic.LoadInt64(&mp.missCount)
}

// NewMetricsPool creates a new metrics pool with pre-allocated maps.
func NewMetricsPool() *MetricsPool {
	return &MetricsPool{
		pool: &sync.Pool{
			New: func() interface{} {
				// Pre-allocate with expected capacity (20-30 fields per subsystem)
				return make(map[string]interface{}, 32)
			},
		},
	}
}

// Get retrieves a metric map from the pool, creating if necessary.
func (mp *MetricsPool) Get() map[string]interface{} {
	return mp.pool.Get().(map[string]interface{})
}

// Put returns a metric map to the pool after clearing it.
func (mp *MetricsPool) Put(m map[string]interface{}) {
	// Clear the map for reuse
	for k := range m {
		delete(m, k)
	}
	mp.pool.Put(m)
}

// GlobalMetricsPool is a shared pool for all subsystems.
var GlobalMetricsPool = NewMetricsPool()

// GetMetricsMap gets a metrics map from the global pool.
func GetMetricsMap() map[string]interface{} {
	return GlobalMetricsPool.Get()
}

// PutMetricsMap returns a metrics map to the global pool.
func PutMetricsMap(m map[string]interface{}) {
	GlobalMetricsPool.Put(m)
}
