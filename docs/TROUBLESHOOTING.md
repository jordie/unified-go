# Troubleshooting Guide

Common issues and solutions for running unified-go.

## Server Issues

### Port Already in Use

**Error:** `listen tcp :8080: bind: address already in use`

**Solution:**
```bash
# Use different port
PORT=8081 ./server

# Or find and stop existing process
lsof -i :8080
kill -9 <PID>
```

### Server Won't Start

**Error:** `failed to initialize database` or `connection refused`

**Solution:**
```bash
# Verify database exists
ls -la data/unified.db

# Check permissions
chmod 644 data/unified.db

# Rebuild database if corrupted
rm data/unified.db
sqlite3 data/unified.db < migrations/001_init.sql
sqlite3 data/unified.db < migrations/002_reading.sql
sqlite3 data/unified.db < migrations/003_typing.sql
sqlite3 data/unified.db < migrations/004_math.sql
```

### Slow Startup

**Cause:** Database migrations or large dataset

**Solution:**
```bash
# Optimize database
sqlite3 data/unified.db "VACUUM;"
sqlite3 data/unified.db "ANALYZE;"

# Check for unfinished transactions
sqlite3 data/unified.db "PRAGMA integrity_check;"
```

---

## Database Issues

### Database Locked

**Error:** `database is locked`

**Cause:** Another process using database

**Solution:**
```bash
# Find locking process
lsof data/unified.db

# Kill if necessary
kill -9 <PID>

# If persistent, delete and recreate
rm data/unified.db-shm data/unified.db-wal 2>/dev/null
rm data/unified.db
# Reinitialize (see above)
```

### Corrupted Database

**Error:** `database disk image is malformed` or `file is encrypted or is not a database`

**Solution:**
```bash
# Restore from backup
sqlite3 data/unified.db < backup_20250220.sql

# Or rebuild empty
rm data/unified.db*
mkdir -p data
sqlite3 data/unified.db < migrations/001_init.sql
# ... run all migrations
```

### Foreign Key Constraint Violation

**Error:** `FOREIGN KEY constraint failed`

**Cause:** Inserting record with non-existent FK

**Solution:**
```bash
# Check referenced record exists
sqlite3 data/unified.db "SELECT * FROM users WHERE id = <user_id>;"

# Or disable FK check temporarily (not recommended for production)
sqlite3 data/unified.db "PRAGMA foreign_keys = OFF;"
```

### Connection Pool Exhausted

**Error:** `too many open connections` or connection timeouts

**Solution:**
```bash
# Increase connection limit in config
# Check connection count
sqlite3 data/unified.db "PRAGMA database_list;"

# Close idle connections
# Restart service to reset pool
```

---

## API Issues

### 404 Not Found

**Error:** `{"error": "Not found", "message": "Passage with id 999 not found"}`

**Solutions:**
- Verify resource ID is correct
- Check resource exists: `curl http://localhost:8080/api/reading/passages`
- For user endpoints, verify user exists

### 401 Unauthorized

**Error:** `{"error": "Unauthorized", "message": "User not found"}`

**Solutions:**
```bash
# Pass user ID in header
curl -H "X-User-ID: 1" http://localhost:8080/api/users/1/typing/stats

# Verify user exists in database
sqlite3 data/unified.db "SELECT * FROM users WHERE id = 1;"

# Create test user if needed
sqlite3 data/unified.db "INSERT INTO users (username) VALUES ('test');"
```

### 400 Bad Request

**Error:** `{"error": "Invalid request", "message": "..."}`

**Solutions:**
- Check request JSON is valid
- Verify required fields present
- Check field types match

Examples:
```bash
# Valid math problem request
curl -X POST http://localhost:8080/api/math/problem \
  -H "Content-Type: application/json" \
  -d '{"problem_type": "addition", "difficulty": "easy"}'

# Valid typing test request
curl -X POST http://localhost:8080/api/typing/test \
  -H "Content-Type: application/json" \
  -d '{
    "text": "The quick brown fox",
    "duration_seconds": 60,
    "user_input": "The quick brown fox"
  }'
```

### 500 Internal Server Error

**Error:** `{"error": "Internal server error"}`

**Solutions:**
1. Check server logs:
```bash
tail -f logs/unified-go.log
```

2. Common causes:
   - Database connection lost: Restart service
   - Invalid JSON in database: Check data integrity
   - Missing migration: Run all migrations

3. Restart service:
```bash
sudo systemctl restart unified-go
# Or if running locally
pkill server
./server &
```

---

## Build Issues

### Compilation Error: Undefined Package

**Error:** `cannot find package`

**Solution:**
```bash
# Download dependencies
go mod download

# Or with vendor directory
go mod vendor

# Then rebuild
go build -o server ./cmd/server
```

### Compilation Error: Type Mismatch

**Error:** `cannot use ... (type X) as type Y`

**Solution:**
```bash
# Update Go version
go version  # Should be 1.21+

# Or clean and rebuild
go clean -cache
go build -o server ./cmd/server
```

### Build Size Too Large

**Problem:** Binary is >20MB

**Solution:**
```bash
# Build with optimizations
go build -ldflags="-s -w" -o server ./cmd/server

# Size comparison
ls -lh server

# Should be ~13MB optimized
```

---

## Test Issues

### Tests Failing

**Error:** `FAIL: TestSomething`

**Solutions:**
```bash
# Run single test with verbose output
go test -v -run TestSomething ./pkg/reading

# Run with race detector
go test -race ./...

# Check test database initialization
# Ensure migrations run in test setup
```

### Tests Slow

**Cause:** Database operations or network calls

**Solution:**
```bash
# Run benchmarks to identify bottleneck
go test -bench=. -benchmem ./pkg/typing

# Output shows time per operation
```

### Tests Hang

**Cause:** Deadlock or infinite loop

**Solution:**
```bash
# Run with timeout
go test -timeout 30s ./...

# Or with context
timeout 30 go test ./...

# Check for database locks
lsof data/unified.db
```

---

## Performance Issues

### Slow API Responses

**Symptom:** API endpoints take >1 second

**Solutions:**
1. Check database indexes:
```bash
sqlite3 data/unified.db "PRAGMA index_list(reading_results);"
```

2. Optimize slow queries:
```bash
sqlite3 data/unified.db "EXPLAIN QUERY PLAN SELECT * FROM reading_results WHERE user_id = 1;"
```

3. Monitor database size:
```bash
sqlite3 data/unified.db "
SELECT 
    name,
    page_count * page_size / 1024.0 / 1024.0 as size_mb
FROM pragma_page_count(), pragma_page_size(),
    (SELECT name FROM sqlite_master WHERE type='table')
ORDER BY size_mb DESC;"
```

### High Memory Usage

**Symptom:** Server using >500MB RAM

**Solutions:**
1. Check for memory leaks:
```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

2. Limit Go memory:
```bash
GOMEMLIMIT=256MiB ./server
```

3. Reduce cache size if configured

### High CPU Usage

**Symptom:** Server CPU constantly >50%

**Solutions:**
1. Profile CPU:
```bash
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

2. Check for expensive queries:
```bash
sqlite3 data/unified.db "PRAGMA query_only = ON;"
```

3. Add database indexes for hot queries

---

## Deployment Issues

### SSL Certificate Error

**Error:** `certificate signed by unknown authority`

**Solutions:**
```bash
# If using self-signed cert
curl -k https://localhost:8443  # Insecure, dev only

# For production, get proper cert
sudo certbot certonly -d example.com

# Check cert validity
openssl x509 -in /etc/letsencrypt/live/example.com/fullchain.pem -text
```

### Nginx Proxy Issues

**Error:** `502 Bad Gateway` or `upstream timed out`

**Solutions:**
```bash
# Verify backend is running
curl http://localhost:8080/health

# Check Nginx logs
tail -f /var/log/nginx/error.log

# Increase timeout in nginx.conf
proxy_connect_timeout 30s;
proxy_send_timeout 30s;
proxy_read_timeout 30s;

# Reload Nginx
sudo nginx -t
sudo systemctl reload nginx
```

### Systemd Service Won't Start

**Error:** `Service failed with result: exit-code`

**Solutions:**
```bash
# Check service status
sudo systemctl status unified-go

# View detailed logs
sudo journalctl -u unified-go -n 50

# Verify config file syntax
cat /etc/systemd/system/unified-go.service

# Test manually
cd /app/unified-go
./server  # Run interactively to see errors
```

---

## Debugging

### Enable Debug Logging

```bash
# Set log level
LOG_LEVEL=debug ./server

# Or in systemd service
Environment="LOG_LEVEL=debug"
```

### Database Query Debugging

```bash
# Enable query logging
sqlite3 data/unified.db ".trace" | tee query.log

# Run queries
sqlite3 data/unified.db "SELECT * FROM reading_results LIMIT 1;"

# Examine query.log for performance insights
```

### API Request Debugging

```bash
# Use curl with verbose output
curl -v http://localhost:8080/api/reading/passages

# Or with debug headers
curl -H "X-Debug: true" http://localhost:8080/health
```

### Process Monitoring

```bash
# Watch server resource usage
watch -n 1 'ps aux | grep server'

# Or with top
top -p $(pgrep server)

# Network connections
netstat -an | grep 8080
```

---

## Getting Help

1. **Check logs first:**
```bash
tail -f logs/unified-go.log
tail -f /var/log/nginx/error.log  # If using Nginx
```

2. **Verify health:**
```bash
curl -s http://localhost:8080/health | jq
```

3. **Test connectivity:**
```bash
# Test database
sqlite3 data/unified.db ".tables"

# Test API
curl http://localhost:8080/api/reading/passages
```

4. **Collect diagnostics:**
```bash
# System info
uname -a
go version
sqlite3 --version

# Service status
systemctl status unified-go
journalctl -u unified-go -n 100

# Database info
sqlite3 data/unified.db "PRAGMA integrity_check;"
```

---

See [DEPLOYMENT.md](DEPLOYMENT.md) for deployment-specific issues or [API_ENDPOINTS_COMPLETE.md](API_ENDPOINTS_COMPLETE.md) for API documentation.
