# Grafana Dashboards Guide

Complete guide to importing and using the 7 GAIA monitoring dashboards.

## Quick Start (5 minutes)

### 1. Start Grafana

**Option A: Standalone**
```bash
brew install grafana
grafana-server
# Open http://localhost:3000 (admin/admin)
```

**Option B: Docker**
```bash
docker run -d -p 3000:3000 grafana/grafana:latest
# Open http://localhost:3000 (admin/admin/admin)
```

### 2. Add Prometheus Data Source

In Grafana UI:
1. Navigate to **Configuration** → **Data Sources**
2. Click **Add data source**
3. Choose **Prometheus**
4. Set URL to `http://localhost:9090`
5. Click **Save & Test** (should show "Data source is working")

### 3. Import Dashboards

**Option A: Via UI**
1. Navigate to **Dashboards** → **Import**
2. Upload JSON file from `configs/grafana_dashboards/`
3. Select Prometheus as data source
4. Click **Import**
5. Repeat for all 7 dashboards

**Option B: Via API**
```bash
for dashboard in configs/grafana_dashboards/*.json; do
  curl -X POST http://localhost:3000/api/dashboards/db \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <GRAFANA_API_TOKEN>" \
    -d @"$dashboard"
done
```

**Option C: Via Command Line**
```bash
# Using grafana-cli (if installed)
grafana-cli dashboard import configs/grafana_dashboards/*.json
```

## Dashboard Overview

### 1. System Overview Dashboard
**File**: `1_system_overview_dashboard.json`

High-level view of all subsystems in one place.

**Panels**:
- System Health Status (UP/DOWN)
- System Throughput (ops/sec)
- System Memory (MB)
- Active Goroutines
- API/File/Browser/Process/Network health gauges (success rates)
- 24h trends for throughput and error rates
- Active connections over time
- Memory utilization
- Goroutine leak detection

**Use Cases**:
- Quick health check of entire system
- Identify which subsystem needs investigation
- Monitor overall throughput and memory trends
- Detect goroutine leaks

**Key Metrics**:
- Throughput: 100K-2M ops/sec (target: >100K)
- Memory: 0-2GB (target: <2GB)
- Goroutines: 0-2000 (target: <1000)

### 2. API Client Pool Dashboard
**File**: `2_api_subsystem_dashboard.json`

Deep dive into API performance and diagnostics.

**Panels**:
- Success Rate (%)
- Error Rate (%)
- P95 Latency (ms)
- Active Clients Count
- Request rate trends
- Success vs failure trends
- Rate limiting events
- Latency percentiles (P50, P95, P99)
- Pool utilization (%)
- Alert status

**Use Cases**:
- Monitor API client pool health
- Identify rate limiting issues
- Track latency degradation
- Diagnose connection pool exhaustion
- Monitor request patterns

**SLOs**:
- Success Rate: >99%
- Error Rate: <1%
- P95 Latency: <100ms
- Pool Utilization: <90% (warning), <100% (critical)

### 3. File Manager Dashboard
**File**: `3_file_subsystem_dashboard.json`

File operation performance and throughput monitoring.

**Panels**:
- Success Rate (%)
- Error Rate (%)
- P95 Latency (ms)
- Concurrent Operations Count
- Operations rate trends
- Success vs failure trends
- Bytes processed
- Throughput (MB/s)
- Latency percentiles
- Concurrent operations trend
- 24h summaries (total, success, failed)

**Use Cases**:
- Monitor file operation performance
- Track I/O throughput
- Identify concurrent operation bottlenecks
- Monitor data processing volume
- Detect file operation errors

**SLOs**:
- Success Rate: >99%
- Error Rate: <1%
- P95 Latency: <500ms
- Concurrent Operations: <100 (warning), <150 (critical)
- Throughput: Varies by workload

### 4. Browser Pool Dashboard
**File**: `4_browser_subsystem_dashboard.json`

Browser instance and tab management monitoring.

**Panels**:
- Active Instances Count
- Active Tabs Count
- Success Rate (%)
- Error Rate (%)
- Instance trend
- Tabs trend
- Operations rate
- Success vs failure rate
- Pool capacity utilization
- 24h summaries (total ops, failures)

**Use Cases**:
- Monitor browser pool capacity
- Track tab proliferation
- Identify browser operation failures
- Monitor instance allocation
- Detect memory leaks in browser pool

**SLOs**:
- Success Rate: >95%
- Error Rate: <5%
- Active Instances: <50 (warning), max 50
- Active Tabs: <1000 (warning), <10000 (critical)

### 5. Process Manager Dashboard
**File**: `5_process_subsystem_dashboard.json`

Process execution and resource utilization monitoring.

**Panels**:
- Active Processes Count
- Memory Usage (MB)
- CPU Usage (%)
- Success Rate (%)
- Active processes trend
- Completion rate trend
- Memory trend
- CPU trend
- 24h summaries (started, completed, failed)

**Use Cases**:
- Monitor process execution health
- Track resource utilization
- Identify CPU/memory bottlenecks
- Detect process failures
- Monitor process completion rates

**SLOs**:
- Success Rate: >95%
- Failure Rate: <5%
- Memory: <500MB (warning), <1GB (critical)
- CPU: <50% (normal), <80% (warning)
- Active Processes: <200

### 6. Network Coordinator Dashboard
**File**: `6_network_subsystem_dashboard.json`

Network connectivity, bandwidth, and DNS cache performance.

**Panels**:
- Active Connections Count
- DNS Cache Hit Rate (%)
- Bytes Transferred (24h)
- Current Throughput (Mbps)
- Connections trend
- DNS cache hit/miss rates
- Network throughput trend
- Cumulative data transferred
- 24h summaries (DNS hits, misses, data)

**Use Cases**:
- Monitor network connectivity
- Track DNS cache efficiency
- Monitor bandwidth utilization
- Identify connection exhaustion
- Analyze data transfer patterns

**SLOs**:
- Active Connections: <1000 (warning), <2000 (critical)
- DNS Cache Hit Rate: >80%
- Throughput: Varies by workload

### 7. SLA Compliance Dashboard
**File**: `7_sla_compliance_dashboard.json`

Service Level Agreement compliance tracking.

**Panels**:
- Overall Availability (%)
- API Availability (target: 99.9%)
- File Manager Availability (target: 99%)
- System Uptime (UP/DOWN)
- API Latency vs Target (P95 < 100ms)
- File Latency vs Target (P95 < 500ms)
- Error rate compliance trends
- SLA violations count
- Throughput vs Target (>100K ops/sec)
- Memory vs Limit (<2GB)
- Individual SLO status indicators:
  - API SLO Status
  - File SLO Status
  - Availability SLO Status
  - Throughput SLO Status
  - Memory SLO Status
  - Overall Compliance Status

**Use Cases**:
- Track SLA compliance over time
- Identify SLA violations
- Generate compliance reports
- Monitor approach to SLA limits
- Plan capacity based on headroom

**SLOs Tracked**:
- Overall Availability: 99.9%
- API Availability: 99.9%
- File Manager Availability: 99%
- API Latency (P95): <100ms
- File Latency (P95): <500ms
- System Throughput: >100K ops/sec
- System Memory: <2GB

## Dashboard Features

### Real-time Monitoring
- 30-second refresh rate by default
- Customizable time ranges (1h, 6h, 24h, 7d)
- Auto-refresh capability

### Visualization Types
- **Stat Panels**: Current values with color-coded status
- **Gauge Panels**: Visual progress indicators with thresholds
- **Time Series**: Trend analysis with multiple series
- **Table Panels**: Detailed data with sorting/filtering

### Color Coding
- **Green**: Healthy, within SLO
- **Yellow**: Warning, approaching SLO limit
- **Red**: Critical, violating SLO

### Drill-Down Navigation
Each dashboard references the system overview, allowing easy navigation:
- Click on subsystem name → Go to detailed subsystem dashboard
- Click on metric alert → See related SLA dashboard
- Use time picker to zoom into specific time ranges

## Customization

### Add Custom Metrics

Edit dashboard JSON to add new panels:

```json
{
  "title": "Custom Metric Panel",
  "type": "timeseries",
  "gridPos": {"x": 0, "y": 0, "w": 12, "h": 8},
  "targets": [
    {
      "expr": "your_metric_name"
    }
  ]
}
```

### Adjust Thresholds

Modify `thresholds.steps` in dashboard JSON:

```json
"thresholds": {
  "mode": "absolute",
  "steps": [
    {"color": "green", "value": 0},
    {"color": "yellow", "value": 50},
    {"color": "red", "value": 100}
  ]
}
```

### Change Time Range

Modify dashboard time settings:

```json
"time": {
  "from": "now-7d",
  "to": "now"
}
```

## Integration with Alerts

Dashboards complement Prometheus alert rules defined in `configs/alerts.yml`.

**Alert → Dashboard Workflow**:
1. Prometheus evaluates 18 alert rules (15-second intervals)
2. Alert fires when SLO violated
3. Alert appears in Grafana dashboard
4. Click alert → Navigate to relevant dashboard for investigation
5. Use dashboard to diagnose root cause
6. Resolution reflected in dashboard in real-time

## Performance Dashboards

### Dashboard Load Times
- System Overview: <1s (retrieves 13 metrics)
- Subsystem Dashboards: <1s (retrieves 8-12 metrics each)
- SLA Dashboard: <2s (retrieves 20+ metrics with aggregations)

### Query Optimization
Dashboards use efficient PromQL queries:
- Histogram quantiles for latency
- Rate() for trend analysis
- Increase() for time-windowed totals
- Simple arithmetic for derived metrics

## Exporting & Sharing

### Export Dashboard as JSON
1. Click **Settings** (gear icon)
2. Click **Export**
3. Choose **Export as JSON**
4. Save file

### Export Metrics as CSV
1. Right-click panel
2. Click **Export CSV**
3. Choose time range
4. Download file

### Share Dashboard Link
1. Click **Share** (arrow icon)
2. Copy link (includes time range, variables)
3. Paste in chat/email/docs

## Troubleshooting

### No Data in Panels
1. Verify Prometheus is running: http://localhost:9090
2. Verify GAIA monitoring server is running: http://localhost:9090/metrics
3. Check data source configuration in Grafana
4. Wait 15+ seconds for Prometheus to scrape metrics
5. Verify metrics in Prometheus: http://localhost:9090 → Graph tab

### Panels Show "Loading"
1. Check Prometheus query execution time
2. Verify no query syntax errors
3. Try shorter time range (e.g., 1h instead of 7d)
4. Check Grafana logs: `docker logs <grafana-container>`

### Alert Status Not Showing
1. Verify alerts.yml is loaded in Prometheus
2. Check Prometheus Alerts page: http://localhost:9090/alerts
3. Verify alert conditions are met (wait 15+ seconds)
4. Check for syntax errors in alerts.yml

### Slow Dashboard Loading
1. Reduce time range (1h instead of 24h)
2. Disable auto-refresh temporarily
3. Check Prometheus performance
4. Upgrade to latest Grafana version

## Best Practices

### Daily Monitoring
- Check System Overview dashboard in morning
- Review overnight errors and performance
- Monitor SLA Compliance for violations

### Incident Investigation
1. Start with System Overview
2. Identify affected subsystem
3. Navigate to subsystem dashboard
4. Check SLA dashboard for impact
5. Review Prometheus alerts

### Capacity Planning
- Monitor average throughput over weeks
- Track memory and CPU trends
- Plan for 2x peak load
- Use SLA dashboard to find headroom

### Performance Tuning
- Use subsystem dashboards to identify bottlenecks
- Profile hot functions using pprof endpoints
- Monitor improvements in real-time
- Target 20%+ improvement (Day 4 of Phase 8.8)

## Summary

**7 Production-Grade Dashboards**:
1. ✅ System Overview - High-level health and trends
2. ✅ API Client Pool - Detailed API performance
3. ✅ File Manager - File operation metrics
4. ✅ Browser Pool - Browser instance management
5. ✅ Process Manager - Process execution and resources
6. ✅ Network Coordinator - Network and DNS metrics
7. ✅ SLA Compliance - Agreement tracking

**Features**:
- 95+ individual panels
- Real-time 30-second refresh
- 18 alert rule integrations
- Color-coded SLO status
- Drill-down navigation
- Customizable thresholds
- Exportable to JSON/CSV

**Quick Start**: Import all 7 dashboards in Grafana, add Prometheus data source, start monitoring!

---

**Phase 8.8 Progress**: Days 1-3 complete (60%)
**Next**: Day 4 - Performance Profiling & Optimization
