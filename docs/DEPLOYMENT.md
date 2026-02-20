# Deployment Guide

Complete instructions for deploying unified-go to development, staging, and production environments.

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                 Web Browser / API Client            │
└────────────────────┬────────────────────────────────┘
                     │ HTTP/HTTPS
┌────────────────────▼────────────────────────────────┐
│         unified-go Server (Go + Chi Router)         │
│  ┌──────────────────────────────────────────────┐  │
│  │  Reading App  │ Typing App │ Math │ Piano   │  │
│  │  (31 tests)   │ (33 tests) │(35+) │(pending)│  │
│  └──────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────┐  │
│  │      HTTP Router - 33+ API Endpoints        │  │
│  │  - /reading/*     (comprehension tests)     │  │
│  │  - /typing/*      (speed/accuracy metrics)  │  │
│  │  - /math/*        (problem generation)      │  │
│  │  - /piano/*       (music practice)          │  │
│  │  - /api/*         (JSON endpoints)          │  │
│  │  - /health        (status check)            │  │
│  └──────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────┐
│    SQLite Database (data/unified.db)                │
│  ┌──────────────────────────────────────────────┐  │
│  │ reading_passages  │ reading_questions        │  │
│  │ reading_answers   │ reading_results          │  │
│  │ typing_tests      │ typing_results           │  │
│  │ math_problems     │ math_solutions           │  │
│  │ math_sessions     │ math_user_stats          │  │
│  │ piano_lessons     │ piano_progress           │  │
│  │ users             │ user_stats               │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Development Environment

### Setup

```bash
# 1. Clone repository
git clone <repo>
cd unified-go

# 2. Install dependencies
go mod download

# 3. Build
go build -o server ./cmd/server

# 4. Initialize database
mkdir -p data logs
sqlite3 data/unified.db < migrations/001_init.sql
sqlite3 data/unified.db < migrations/002_reading.sql
sqlite3 data/unified.db < migrations/003_typing.sql
sqlite3 data/unified.db < migrations/004_math.sql
sqlite3 data/unified.db < migrations/005_piano.sql

# 5. Run tests
go test ./...

# 6. Start server
./server
```

### Environment Variables

```bash
export PORT=8080
export HOST=0.0.0.0
export LOG_LEVEL=debug
export STATIC_DIR=static
```

### Testing

```bash
# All tests with verbose output
go test -v ./...

# Specific package
go test -v ./pkg/reading

# With coverage report
go test -cover ./...

# Benchmark tests
go test -bench=. ./pkg/typing

# Single test
go test -v -run TestGenerateProblem ./pkg/math
```

## Staging Environment

### Deployment Steps

```bash
# 1. Build with optimizations
go build -ldflags="-s -w" -o server ./cmd/server

# 2. Copy to staging server
scp server user@staging.example.com:/app/unified-go/

# 3. SSH into staging
ssh user@staging.example.com

# 4. Verify database
sqlite3 /app/unified-go/data/unified.db ".tables"

# 5. Run tests on staging
cd /app/unified-go
go test ./...

# 6. Start with systemd
sudo systemctl restart unified-go
```

### Systemd Service File

Create `/etc/systemd/system/unified-go.service`:

```ini
[Unit]
Description=Unified-Go Educational Platform
After=network.target

[Service]
Type=simple
User=app
WorkingDirectory=/app/unified-go
Environment="PORT=8080"
Environment="HOST=0.0.0.0"
Environment="LOG_LEVEL=info"
Environment="STATIC_DIR=/app/unified-go/static"
ExecStart=/app/unified-go/server
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable unified-go
sudo systemctl start unified-go
sudo systemctl status unified-go
```

### Nginx Reverse Proxy

```nginx
upstream unified_go {
    server localhost:8080;
}

server {
    listen 80;
    server_name staging.example.com;

    location / {
        proxy_pass http://unified_go;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 90s;
    }

    location /static/ {
        alias /app/unified-go/static/;
        expires 30d;
    }
}
```

Reload Nginx:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Production Environment

### Pre-Deployment Checklist

- [ ] All tests passing (100+ tests)
- [ ] Code reviewed and approved
- [ ] Database backups configured
- [ ] Monitoring/alerting set up
- [ ] SSL/TLS certificate obtained
- [ ] Load testing completed
- [ ] Documentation updated
- [ ] Rollback plan documented

### Production Build

```bash
# Build optimized binary
CGO_ENABLED=1 go build \
  -ldflags="-s -w -X main.Version=$(git describe --tags)" \
  -o server ./cmd/server

# Verify binary
./server --version
file server  # Should show "statically linked"
```

### Database Backup

```bash
# Full backup
sqlite3 data/unified.db ".dump" > backup_$(date +%Y%m%d).sql

# Backup script (run daily)
#!/bin/bash
BACKUP_DIR="/backups/unified-go"
mkdir -p $BACKUP_DIR
sqlite3 /app/unified-go/data/unified.db ".dump" > \
  $BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).sql
# Keep last 30 days
find $BACKUP_DIR -mtime +30 -delete
```

### SSL/TLS Configuration

```bash
# Obtain certificate (Let's Encrypt)
sudo certbot certonly -d example.com -d www.example.com

# Update Nginx configuration
server {
    listen 443 ssl http2;
    server_name example.com www.example.com;

    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # ... rest of config
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name example.com www.example.com;
    return 301 https://$server_name$request_uri;
}
```

### Monitoring Setup

```bash
# Health check endpoint
curl -s http://localhost:8080/health | jq

# Monitor logs
tail -f logs/unified-go.log

# Check database size
sqlite3 data/unified.db "SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size();"

# Monitor process
ps aux | grep server
```

### Performance Tuning

```bash
# SQLite optimizations (in config/config.go)
PRAGMA journal_mode = WAL;      # Write-Ahead Logging
PRAGMA synchronous = NORMAL;    # Faster writes
PRAGMA cache_size = 10000;      # Larger cache
PRAGMA temp_store = MEMORY;     # Temp in memory

# Go runtime tuning
GOMAXPROCS=4                    # CPU cores
GOMEMLIMIT=512MiB              # Memory limit
```

### Load Balancing

For multiple servers:

```nginx
upstream unified_go_cluster {
    least_conn;
    server server1.example.com:8080 weight=5;
    server server2.example.com:8080 weight=5;
    server server3.example.com:8080 weight=3;
    
    keepalive 32;
}

server {
    location / {
        proxy_pass http://unified_go_cluster;
        # ... rest of proxy settings
    }
}
```

## Rollback Procedure

If deployment fails:

```bash
# 1. Stop current version
sudo systemctl stop unified-go

# 2. Restore previous binary
cp /app/unified-go/server.backup /app/unified-go/server

# 3. Restore database from backup (if corrupted)
sqlite3 /app/unified-go/data/unified.db < /backups/latest.sql

# 4. Start service
sudo systemctl start unified-go

# 5. Verify health
curl -s http://localhost:8080/health

# 6. Check logs
tail -f /var/log/unified-go.log
```

## Monitoring & Alerting

### Health Checks

```bash
#!/bin/bash
# Check server health every minute
while true; do
    STATUS=$(curl -s http://localhost:8080/health | jq -r .status)
    if [ "$STATUS" != "healthy" ]; then
        echo "ALERT: Server unhealthy" | mail -s "Alert" ops@example.com
        systemctl restart unified-go
    fi
    sleep 60
done
```

### Log Aggregation

Configure with ELK Stack or similar:

```bash
# Shipper config (Filebeat)
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /app/unified-go/logs/*.log
  multiline.pattern: '^\['
  multiline.negate: true
  multiline.match: after

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

## Performance Benchmarks

Expected metrics with proper deployment:

- **Response Time**: <100ms for API endpoints
- **Throughput**: 1000+ requests/second
- **Database Query**: <10ms for user stats
- **Memory**: <100MB base + 50MB per concurrent user
- **CPU**: <20% on single core at 100 RPS

---

## Troubleshooting Deployment

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and solutions.
