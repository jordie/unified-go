# Quick Start Guide

Get the unified-go educational platform running in under 60 seconds.

## Prerequisites

- Go 1.21+ installed
- SQLite3
- 100MB disk space

## 1. Clone & Build (30 seconds)

```bash
cd ~/Desktop/gitrepo/pyWork/unified-go

# Download dependencies
go mod download

# Build binary
go build -o server ./cmd/server

# Or build with size optimization
go build -ldflags="-s -w" -o server ./cmd/server
```

## 2. Initialize Database (15 seconds)

```bash
# Create data directory
mkdir -p data logs

# Run migrations
sqlite3 data/unified.db < migrations/001_init.sql
sqlite3 data/unified.db < migrations/002_reading.sql
sqlite3 data/unified.db < migrations/003_typing.sql
sqlite3 data/unified.db < migrations/004_math.sql
sqlite3 data/unified.db < migrations/005_piano.sql
```

## 3. Start Server (15 seconds)

```bash
# Run server
./server

# Or with custom config
PORT=8080 HOST=localhost ./server

# Or in background
./server &
```

## 4. Access Dashboard

Open browser: **http://localhost:8080**

Default users (if configured):
- Username: `student`
- Password: `student123`

Or access API directly:

```bash
# Health check
curl http://localhost:8080/health

# List all endpoints
curl http://localhost:8080/api/endpoints

# Get typing dashboard
curl http://localhost:8080/typing/dashboard
```

## 5. Run Tests (30 seconds)

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/reading -v

# With coverage
go test ./... -cover
```

---

## What's Running?

| App | Routes | Purpose |
|-----|--------|---------|
| **Reading** | `/reading/*` | Reading comprehension & vocabulary |
| **Typing** | `/typing/*` | Typing speed & accuracy practice |
| **Math** | `/math/*` | Math problems & score tracking |
| **Piano** | `/piano/*` | Piano practice & music theory |
| **API** | `/api/*` | JSON API endpoints |
| **Health** | `/health` | System health status |

## Common Tasks

### Create a Typing Test
```bash
curl -X POST http://localhost:8080/api/typing/test \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "text": "The quick brown fox", "duration_seconds": 60}'
```

### Get User Stats
```bash
curl http://localhost:8080/api/users/1/typing/stats
```

### Generate Math Problem
```bash
curl -X POST http://localhost:8080/api/math/problem \
  -H "Content-Type: application/json" \
  -d '{"problem_type": "addition", "difficulty": "easy"}'
```

### View Leaderboards
```bash
curl http://localhost:8080/api/typing/leaderboard?limit=10
curl http://localhost:8080/api/math/leaderboard?limit=10
```

---

## Environment Variables

```bash
PORT=8080              # Server port
HOST=0.0.0.0          # Server host
DB_PATH=data/unified.db # Database file
LOG_LEVEL=info        # Logging level: debug, info, warn, error
STATIC_DIR=static     # Static files directory
```

## Troubleshooting

**Port already in use:**
```bash
PORT=8081 ./server
```

**Database locked:**
```bash
# Close any existing connections
rm data/unified.db
# Reinitialize database (see step 2)
```

**Build fails:**
```bash
# Update dependencies
go mod tidy

# Clean build cache
go clean -cache

# Rebuild
go build -o server ./cmd/server
```

## Next Steps

- Read [DEPLOYMENT.md](docs/DEPLOYMENT.md) for production setup
- View [API endpoints](docs/API_ENDPOINTS_COMPLETE.md)
- Check [troubleshooting guide](docs/TROUBLESHOOTING.md)
- Review [architecture](docs/ARCHITECTURE.md)

---

**Need help?** See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) or check logs at `logs/unified-go.log`
