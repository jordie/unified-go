# GAIA_GO Prometheus Integration Quick Start

## Overview

Day 2 of Phase 8.8 (Optimization & Monitoring) establishes Prometheus metrics collection and alert rules for all GAIA subsystems. This guide helps you get Prometheus and Grafana running with GAIA metrics.

## Prerequisites

- GAIA_GO running on port :9090 (monitoring server)
- Prometheus (latest stable)
- Grafana (latest stable)

## Quick Start (5 minutes)

### 1. Start GAIA with Monitoring

```bash
cd GAIA_GO
go run examples/monitoring_example.go
```

The monitoring server will start on port :9090.

### 2. Verify Metrics Endpoint

```bash
# Check health
curl http://localhost:9090/health | jq .

# Check metrics (should return Prometheus format)
curl http://localhost:9090/metrics | head -20
```

You should see metrics like:
```
# HELP gaia_api_requests_total Total API requests made
# TYPE gaia_api_requests_total counter
gaia_api_requests_total 0

# HELP gaia_api_latency_ms API request latency in milliseconds
# TYPE gaia_api_latency_ms histogram
...
```

### 3. Start Prometheus

**Option A: Standalone (macOS/Linux)**

```bash
# Install if needed (macOS)
brew install prometheus

# Start with GAIA config
prometheus --config.file=GAIA_GO/configs/prometheus.yml
```

**Option B: Docker**

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v "$(pwd)/GAIA_GO/configs/prometheus.yml:/etc/prometheus/prometheus.yml" \
  -v "$(pwd)/GAIA_GO/configs/alerts.yml:/etc/prometheus/alerts.yml" \
  prom/prometheus:latest
```

### 4. Access Prometheus

Open http://localhost:9090

Navigate to:
- **Graph**: Query metrics (e.g., `gaia_api_requests_total`)
- **Alerts**: View configured alert rules (18 total)
- **Targets**: Verify GAIA endpoints are scraped (should show "UP")

### 5. Start Grafana (Optional)

```bash
# Install if needed (macOS)
brew install grafana

# Start
grafana-server

# Or Docker
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  grafana/grafana:latest
```

Access Grafana at http://localhost:3000 (admin/admin)

Add Prometheus data source:
- URL: `http://localhost:9090`
- Access: Browser
- Skip TLS Verify

## Configuration Files

### prometheus.yml

Configuration for scraping GAIA metrics. Features:

- **Global**: 15s scrape interval, 15s evaluation interval
- **Alerting**: Reference to alerts.yml
- **Scrape Configs**: 7 jobs for different subsystems
- **Metric Filters**: Each job filters for its subsystem metrics

Key jobs:
- `gaia-monitoring` - Main metrics endpoint
- `gaia-api-subsystem` - API Client Pool metrics
- `gaia-file-subsystem` - File Manager metrics
- `gaia-network-subsystem` - Network Coordinator metrics
- `gaia-browser-subsystem` - Browser Pool metrics
- `gaia-process-subsystem` - Process Manager metrics
- `gaia-system-metrics` - System-wide metrics

### alerts.yml

Alert rules organized by subsystem:

**API Client Pool (4 alerts)**
- `APIHighErrorRate` - Error rate > 1% (critical)
- `APIHighLatency` - P95 latency > 100ms (warning)
- `APIPoolExhausted` - All clients in use (critical)
- `APIRateLimited` - Rate limiting active (warning)

**File Manager (3 alerts)**
- `FileHighErrorRate` - Error rate > 1% (critical)
- `FileHighLatency` - P95 latency > 500ms (warning)
- `FileConcurrencyHigh` - >100 concurrent ops (warning)

**Browser Pool (3 alerts)**
- `BrowserHighErrorRate` - Error rate > 5% (critical)
- `BrowserPoolExhausted` - All instances in use (critical)
- `BrowserTabsExcessive` - >1000 open tabs (warning)

**Process Manager (3 alerts)**
- `ProcessHighFailureRate` - Failure rate > 5% (critical)
- `ProcessHighMemory` - >500MB used (warning)
- `ProcessHighCPU` - >80% CPU (warning)

**Network Coordinator (2 alerts)**
- `NetworkConnectionLimitHigh` - >1000 active (warning)
- `NetworkDNSCacheLowHitRate` - <80% hit rate (info)

**System-wide (2 alerts)**
- `SystemLowThroughput` - <100K ops/sec (warning)
- `SystemHighMemory` - >2GB used (warning)
- `SystemGoroutineLeakDetected` - >1000 goroutine increase in 5m (critical)

**Availability (1 alert)**
- `AvailabilitySLAViolation` - <99.9% availability (critical)

**Total: 18 alert rules**

## Metrics Overview

### API Client Pool (6 metrics)

```prometheus
# Counters
gaia_api_requests_total           # Total API requests
gaia_api_success_total            # Successful requests
gaia_api_errors_total             # Failed requests
gaia_api_rate_limited_total       # Rate-limited requests

# Histogram
gaia_api_latency_ms               # Request latency (buckets: 1, 5, 10, 50, 100, 500, 1000ms)

# Gauge
gaia_api_active_clients           # Currently active clients
```

### File Manager (6 metrics)

```prometheus
# Counters
gaia_file_operations_total        # Total file operations
gaia_file_success_total           # Successful operations
gaia_file_errors_total            # Failed operations
gaia_file_bytes_processed_total   # Total bytes processed

# Histogram
gaia_file_latency_ms              # Operation latency

# Gauge
gaia_file_concurrent_operations   # Currently concurrent operations
```

### Browser Pool (4 metrics)

```prometheus
# Gauges
gaia_browser_instances_active     # Active browser instances
gaia_browser_active_tabs          # Open tabs

# Counters
gaia_browser_operations_total     # Total operations
gaia_browser_errors_total         # Failed operations
```

### Process Manager (6 metrics)

```prometheus
# Gauges
gaia_process_active               # Active processes
gaia_process_memory_mb            # Memory usage (MB)
gaia_process_cpu_percent          # CPU usage (%)

# Counters
gaia_process_started_total        # Processes started
gaia_process_completed_total      # Processes completed
gaia_process_failed_total         # Process failures
```

### Network Coordinator (4 metrics)

```prometheus
# Counters
gaia_network_bytes_transferred_total  # Total bytes
gaia_network_dns_cache_hits_total     # DNS hits
gaia_network_dns_cache_misses_total   # DNS misses

# Gauge
gaia_network_active_connections   # Active connections
```

### System-wide (3 metrics)

```prometheus
# Gauges
gaia_system_throughput_ops_per_sec    # Operations/sec
gaia_system_memory_mb                 # Memory usage (MB)
gaia_system_goroutines                # Active goroutines
```

**Total: 36 custom metrics + Go runtime metrics**

## Prometheus Queries (PromQL Examples)

### API Health

```promql
# Success rate
(gaia_api_success_total / gaia_api_requests_total) * 100

# Error rate
(gaia_api_errors_total / gaia_api_requests_total) * 100

# P95 latency
histogram_quantile(0.95, gaia_api_latency_ms)

# Request rate
rate(gaia_api_requests_total[5m])
```

### File Manager Health

```promql
# Success rate
(gaia_file_success_total / gaia_file_operations_total) * 100

# Throughput (operations/sec)
rate(gaia_file_operations_total[1m])

# Concurrent operations
gaia_file_concurrent_operations
```

### System Performance

```promql
# Combined throughput
gaia_system_throughput_ops_per_sec

# Memory usage
gaia_system_memory_mb

# Goroutine count
gaia_system_goroutines
```

## Troubleshooting

### Prometheus won't scrape targets

1. Check GAIA is running: `curl http://localhost:9090/metrics`
2. Check prometheus.yml syntax: `promtool check config prometheus.yml`
3. Check targets in Prometheus UI: http://localhost:9090/targets
4. Look for error messages in Prometheus logs

### No metrics showing up

1. Wait 15+ seconds (scrape interval) for metrics to appear
2. Verify metrics endpoint: `curl http://localhost:9090/metrics | head`
3. Check if GAIA subsystems are generating activity
4. Run `go run examples/monitoring_example.go` to simulate activity

### Alerts not triggering

1. Check alerts.yml syntax: `promtool check rules alerts.yml`
2. Verify alert conditions in Prometheus UI: http://localhost:9090/alerts
3. Check if metrics meet threshold conditions
4. Look at alert evaluation interval (15s default)

### Grafana can't reach Prometheus

1. Ensure Prometheus is running: `curl http://localhost:9090`
2. Check data source URL in Grafana settings
3. If using containers, use container hostnames instead of localhost
4. Check Docker network connectivity

## Next Steps (Day 3)

Phase 8.8 Day 3 focuses on **Grafana Dashboards**:
- Create 7 dashboard JSON files
- System overview dashboard
- Per-subsystem dashboards
- SLA compliance dashboard
- Import and configure in Grafana

## Files in This Integration

- `GAIA_GO/configs/prometheus.yml` - Prometheus configuration (7 jobs, 15s scrape)
- `GAIA_GO/configs/alerts.yml` - Alert rules (18 alerts across all subsystems)
- `GAIA_GO/internal/monitoring/prometheus_exporter.go` - Prometheus exporter (60+ metrics)
- `GAIA_GO/examples/monitoring_example.go` - Example usage

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus PromQL Queries](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Prometheus Alerting Rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
- [GAIA_GO Documentation](../README.md)

## Summary

✅ **Day 1 Complete**: Core Monitoring Infrastructure (HTTP server, health checks, pprof)
✅ **Day 2 Complete**: Prometheus Integration (configuration, alerts, metrics)

**Next**: Day 3 - Grafana Dashboards (visualizations, SLA tracking)
