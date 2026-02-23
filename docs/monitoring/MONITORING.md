# GAIA_GO Monitoring Guide

Complete monitoring solution for GAIA Phase 8 subsystems with Prometheus, Grafana, and health checks.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Architecture Overview](#architecture-overview)
3. [Health Checks](#health-checks)
4. [Metrics Reference](#metrics-reference)
5. [Alert Rules](#alert-rules)
6. [Grafana Dashboards](#grafana-dashboards)
7. [pprof Profiling](#pprof-profiling)
8. [Production Deployment](#production-deployment)
9. [Troubleshooting](#troubleshooting)

---

## Quick Start

### 1. Start GAIA Monitoring Server (5 minutes)

```bash
cd GAIA_GO

# Start monitoring server on port 9090
go run examples/monitoring_example.go &

# Verify health endpoint
curl http://localhost:9090/health | jq
```

**Expected Response**:
```json
{
  "status": "Healthy",
  "subsystems": {
    "api_pool": "Healthy",
    "file_manager": "Healthy",
    "browser_pool": "Healthy",
    "process_manager": "Healthy",
    "network_coordinator": "Healthy"
  },
  "timestamp": "2026-02-23T10:00:00Z"
}
```

### 2. Start Prometheus (5 minutes)

```bash
# Install Prometheus
brew install prometheus  # macOS
# or docker run -d -p 9090:9090 prom/prometheus

# Start with GAIA config
prometheus --config.file=GAIA_GO/configs/prometheus.yml &

# Access at http://localhost:9090
```

### 3. Start Grafana (5 minutes)

```bash
# Install and run Grafana
brew install grafana  # macOS
# or docker run -d -p 3000:3000 grafana/grafana

# Access at http://localhost:3000 (admin/admin)

# Add Prometheus data source:
# 1. Configuration → Data Sources → Add
# 2. URL: http://localhost:9090
# 3. Save & Test

# Import dashboards:
# 1. Dashboards → Import
# 2. Upload JSON from configs/grafana_dashboards/
# 3. Select Prometheus as data source
```

**Total Setup Time**: ~15 minutes

---

## Architecture Overview

### Monitoring Stack Components

```
Grafana (Port 3000)
    ↓ (displays)
Prometheus (Port 9090)
    ↓ (scrapes metrics every 15s)
GAIA Monitoring Server (Port 9090)
    ├─ /health - Health check endpoint
    ├─ /metrics - Prometheus metrics export
    └─ /debug/pprof/* - Runtime profiling

        ↓ (monitored)

5 GAIA Subsystems
    ├─ APIClientPool (1000+ concurrent HTTP calls)
    ├─ FileManager (500+ concurrent file ops)
    ├─ BrowserPool (100+ concurrent browser instances)
    ├─ ProcessManager (200+ concurrent processes)
    └─ NetworkCoordinator (bandwidth + DNS management)
```

### Data Flow

1. **Metric Collection** (every 15 seconds):
   - Prometheus scrapes `/metrics` endpoint
   - Exporter reads subsystem GetMetrics()
   - Metrics updated in Prometheus TSDB

2. **Alert Evaluation** (every 15 seconds):
   - Prometheus evaluates 18 alert rules
   - Alert conditions trigger notifications
   - Grafana displays alert status

3. **Visualization** (real-time):
   - Grafana queries Prometheus every 30 seconds
   - Dashboards refresh with latest data
   - Historical data available for trending

---

## Health Checks

### Health Endpoint Overview

The `/health` endpoint provides graduated status determination across all subsystems.

### Endpoints

#### **System Health**
```bash
curl http://localhost:9090/health | jq
```

**Response**:
```json
{
  "status": "Healthy|Degraded|Unhealthy",
  "message": "All subsystems healthy",
  "timestamp": "2026-02-23T10:00:00Z",
  "subsystems": {
    "api_pool": {
      "status": "Healthy",
      "metrics": {
        "success_rate": 99.5,
        "error_rate": 0.5,
        "avg_latency_ms": 45.2
      }
    },
    ...
  },
  "issues": []
}
```

#### **Per-Subsystem Health**
```bash
# API Client Pool
curl http://localhost:9090/health/api | jq

# File Manager
curl http://localhost:9090/health/file | jq

# Browser Pool
curl http://localhost:9090/health/browser | jq

# Process Manager
curl http://localhost:9090/health/process | jq

# Network Coordinator
curl http://localhost:9090/health/network | jq
```

#### **Kubernetes Probes**
```bash
# Liveness probe (is the server running?)
curl http://localhost:9090/liveness

# Readiness probe (can it handle requests?)
curl http://localhost:9090/readiness
```

### Health Status Thresholds

| Status | Success Rate | Condition |
|--------|-------------|-----------|
| Healthy | >98% | All metrics within SLO |
| Degraded | 80-98% | Some metrics approaching limits |
| Unhealthy | <80% | Critical metrics exceeded |

---

## Metrics Reference

### Metric Naming Convention

All metrics follow the pattern: `gaia_{subsystem}_{metric_name}`

### API Client Pool Metrics

**Counters**:
- `gaia_api_requests_total` - Total API requests made
- `gaia_api_success_total` - Successful API requests
- `gaia_api_errors_total` - Failed API requests
- `gaia_api_rate_limited_total` - Rate limited requests

**Gauges**:
- `gaia_api_active_clients` - Currently active HTTP clients
- `gaia_api_pool_size` - Total pool capacity

**Histograms**:
- `gaia_api_latency_ms` - Request latency in milliseconds
  - Buckets: 1, 5, 10, 50, 100, 500, 1000ms

**Calculated Metrics** (in PromQL):
- Success Rate: `100 * gaia_api_success_total / gaia_api_requests_total`
- Error Rate: `100 * gaia_api_errors_total / gaia_api_requests_total`
- P95 Latency: `histogram_quantile(0.95, gaia_api_latency_ms)`

### File Manager Metrics

**Counters**:
- `gaia_file_operations_total` - Total file operations
- `gaia_file_success_total` - Successful operations
- `gaia_file_errors_total` - Failed operations
- `gaia_file_bytes_processed_total` - Total bytes processed

**Gauges**:
- `gaia_file_concurrent_operations` - Currently active operations
- `gaia_file_max_concurrent` - Maximum concurrent limit

**Histograms**:
- `gaia_file_latency_ms` - Operation latency in milliseconds

### Browser Pool Metrics

**Counters**:
- `gaia_browser_operations_total` - Total browser operations
- `gaia_browser_errors_total` - Operation failures
- `gaia_browser_extensions_loaded_total` - Extensions loaded
- `gaia_browser_extensions_failed_total` - Extension load failures

**Gauges**:
- `gaia_browser_instances_active` - Active browser instances
- `gaia_browser_active_tabs` - Currently open tabs
- `gaia_browser_max_instances` - Maximum instance limit

### Process Manager Metrics

**Counters**:
- `gaia_process_started_total` - Processes started
- `gaia_process_completed_total` - Processes completed
- `gaia_process_failed_total` - Process failures

**Gauges**:
- `gaia_process_active` - Currently running processes
- `gaia_process_memory_mb` - Memory usage
- `gaia_process_cpu_percent` - CPU usage percentage
- `gaia_process_max_concurrent` - Maximum concurrent limit

### Network Coordinator Metrics

**Counters**:
- `gaia_network_bytes_transferred_total` - Total bytes transferred
- `gaia_network_dns_cache_hits_total` - DNS cache hits
- `gaia_network_dns_cache_misses_total` - DNS cache misses

**Gauges**:
- `gaia_network_active_connections` - Active connections
- `gaia_network_dns_cache_size` - DNS cache entries
- `gaia_network_bandwidth_limit_bytes` - Bandwidth limit

### System-Wide Metrics

**Gauges**:
- `gaia_system_throughput_ops_per_sec` - Operations per second
- `gaia_system_memory_mb` - Total memory usage
- `gaia_system_goroutines` - Active goroutine count

---

## Alert Rules

### Alert Configuration

Alerts are defined in `configs/alerts.yml` and evaluated every 15 seconds by Prometheus.

### Alert Rules by Subsystem

#### API Client Pool Alerts

1. **APIErrorRateHigh** (Critical)
   - Condition: Error rate > 1% for 5 minutes
   - Action: Page on-call engineer

2. **APILatencyHigh** (Warning)
   - Condition: P95 latency > 100ms for 10 minutes
   - Action: Investigate backend latency

3. **APIPoolExhausted** (Critical)
   - Condition: Available clients < 5
   - Action: Scale up connection pool

4. **APIRateLimitExceeded** (Warning)
   - Condition: Rate limited requests > 100 in 5 minutes
   - Action: Increase rate limit or reduce load

#### File Manager Alerts

1. **FileErrorRateHigh** (Critical)
   - Condition: Error rate > 1% for 5 minutes

2. **FileLatencyHigh** (Warning)
   - Condition: P95 latency > 500ms for 10 minutes

3. **FileConcurrencyHigh** (Warning)
   - Condition: Concurrent ops > 100 for 10 minutes

#### Browser Pool Alerts

1. **BrowserErrorRateHigh** (Critical)
   - Condition: Error rate > 5% for 5 minutes

2. **BrowserPoolExhausted** (Critical)
   - Condition: Active instances >= max instances

3. **BrowserTabsHigh** (Warning)
   - Condition: Active tabs > 1000 for 10 minutes

#### Process Manager Alerts

1. **ProcessFailureRateHigh** (Critical)
   - Condition: Failure rate > 5% for 5 minutes

2. **ProcessMemoryHigh** (Warning)
   - Condition: Memory > 500MB for 10 minutes

3. **ProcessCPUHigh** (Warning)
   - Condition: CPU > 80% for 10 minutes

#### Network Coordinator Alerts

1. **NetworkConnectionsHigh** (Warning)
   - Condition: Connections > 1000 for 10 minutes

2. **DNSCacheHitRateLow** (Warning)
   - Condition: Hit rate < 80% for 15 minutes

#### System Alerts

1. **SystemThroughputLow** (Warning)
   - Condition: Throughput < 100K ops/sec for 10 minutes

2. **SystemMemoryHigh** (Critical)
   - Condition: Memory > 2GB for 10 minutes

3. **GoroutineLeakDetected** (Warning)
   - Condition: Goroutines > 2000 for 10 minutes

#### Availability Alert

1. **SLAViolationDetected** (Critical)
   - Condition: Availability < 99.9% for 5 minutes
   - Action: Page on-call immediately

---

## Grafana Dashboards

### Available Dashboards

7 production-grade dashboards are provided:

1. **System Overview** - High-level health and trends
2. **API Client Pool** - Detailed API performance
3. **File Manager** - File operation metrics
4. **Browser Pool** - Browser instance management
5. **Process Manager** - Process execution and resources
6. **Network Coordinator** - Network and DNS metrics
7. **SLA Compliance** - Service level tracking

### Importing Dashboards

**Via UI**:
1. Go to http://localhost:3000
2. Dashboards → Import
3. Upload JSON from `configs/grafana_dashboards/`
4. Select Prometheus as data source
5. Click Import

**Via API**:
```bash
for dashboard in configs/grafana_dashboards/*.json; do
  curl -X POST http://localhost:3000/api/dashboards/db \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${GRAFANA_TOKEN}" \
    -d @"$dashboard"
done
```

### Dashboard Features

- Real-time metrics (30-second refresh)
- 24-hour historical view
- SLO-aligned color coding (green/yellow/red)
- Alert integration
- Drill-down navigation between dashboards
- Exportable to CSV, JSON, PNG

---

## pprof Profiling

### Available Profiles

The monitoring server exposes standard Go profiling endpoints at `/debug/pprof/`:

#### CPU Profiling
```bash
# Capture 30-second CPU profile
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze profile
go tool pprof http://localhost:9090/debug/pprof/profile?seconds=30
```

#### Memory Profiling
```bash
# Capture heap snapshot
curl http://localhost:9090/debug/pprof/heap > mem.prof

# Analyze allocations
go tool pprof -alloc_space mem.prof
go tool pprof -alloc_count mem.prof
go tool pprof -inuse_space mem.prof
```

#### Goroutine Profiling
```bash
# List current goroutines
curl http://localhost:9090/debug/pprof/goroutine?debug=1

# Detailed goroutine trace
curl http://localhost:9090/debug/pprof/goroutine?debug=2 > goroutines.prof
```

#### Block Profiling (Contention)
```bash
# Profile lock contention
curl http://localhost:9090/debug/pprof/block > block.prof
go tool pprof block.prof
```

#### Mutex Profiling
```bash
# Profile mutex contention
curl http://localhost:9090/debug/pprof/mutex > mutex.prof
go tool pprof mutex.prof
```

#### All Metrics
```bash
# View all available profiles
curl http://localhost:9090/debug/pprof/ | head -20
```

### Profiling Examples

**Find memory leaks**:
```bash
# Capture before and after
curl http://localhost:9090/debug/pprof/heap > before.prof
sleep 300  # Wait 5 minutes
curl http://localhost:9090/debug/pprof/heap > after.prof

# Compare growth
go tool pprof -base before.prof after.prof
```

**Identify hot functions**:
```bash
# CPU profile with top functions
go tool pprof http://localhost:9090/debug/pprof/profile?seconds=60
# In pprof: type "top"
```

---

## Production Deployment

### Security Considerations

1. **Bind to localhost only** (development):
```bash
# monitoring_example.go binds to 127.0.0.1:9090
```

2. **Use authentication** (production):
```go
// Add authentication middleware
server.Use(authMiddleware())
```

3. **Restrict pprof access**:
```go
// Only allow from internal IPs
if !isInternalIP(req.RemoteAddr) {
    return http.StatusForbidden
}
```

4. **Use HTTPS**:
```go
// Generate self-signed cert for production
listener, _ := tls.Listen("tcp", ":9090", &tls.Config{...})
server.Listener = listener
```

### High Availability Setup

**Multi-instance monitoring**:
```yaml
# prometheus.yml - scrape multiple GAIA instances
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'gaia-primary'
    static_configs:
      - targets: ['192.168.1.100:9090']

  - job_name: 'gaia-secondary'
    static_configs:
      - targets: ['192.168.1.101:9090']
```

### Backup & Retention

```yaml
# prometheus.yml - retention policy
global:
  retention: 30d
  retention_size: "10GB"
```

---

## Troubleshooting

### Issue: No data in Grafana panels

**Solution**:
1. Verify Prometheus scraping: http://localhost:9090/targets
2. Verify metrics are being collected: http://localhost:9090/metrics
3. Check data source configuration in Grafana
4. Wait 15+ seconds for Prometheus to scrape

### Issue: Alerts not firing

**Solution**:
1. Verify alerts.yml is loaded: http://localhost:9090/alerts
2. Check Prometheus alert status page
3. Verify alert conditions are being met
4. Check for YAML syntax errors in alerts.yml

### Issue: High memory usage in monitoring server

**Solution**:
1. Check goroutine count: http://localhost:9090/debug/pprof/goroutine?debug=1
2. Profile memory allocations
3. Look for allocation leaks in GetMetrics()
4. Consider reducing Prometheus scrape frequency

### Issue: Slow Grafana dashboard loading

**Solution**:
1. Reduce time range (1h instead of 24h)
2. Disable auto-refresh temporarily
3. Check Prometheus query performance
4. Upgrade to latest Grafana version

### Issue: pprof endpoints returning errors

**Solution**:
1. Verify monitoring server is running
2. Check server logs for errors
3. Ensure ports are not in use
4. Verify firewall allows connections

---

## Best Practices

### Monitoring Strategy

1. **Set up alerts first** - Define what "bad" looks like
2. **Use SLOs** - Set explicit service level objectives
3. **Dashboard per team** - Each subsystem has dedicated dashboard
4. **Regular reviews** - Weekly review of alert patterns
5. **Test incident response** - Practice responding to alerts

### Performance Tips

1. **Keep retention reasonable** - 30 days is typical
2. **Scrape less frequently if needed** - 30s or 1m for less load
3. **Aggregate long-term data** - Archive after 30 days
4. **Monitor the monitors** - Ensure monitoring isn't a bottleneck

### Documentation

1. **Document SLOs** - Link to each metric
2. **Runbooks for alerts** - Step-by-step resolution
3. **Escalation procedures** - Who to page at each severity
4. **Maintenance windows** - Plan downtime in advance

---

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Go pprof Guide](https://golang.org/blog/profiling-go-programs)
- [SLO Best Practices](https://sre.google/workbook/error-budgets/)

---

*Generated: February 23, 2026*
*Phase 8.8 Day 5: Documentation*
*Complete monitoring solution for GAIA Phase 8 subsystems*
