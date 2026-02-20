# Quick Start Guide

Get the Unified Educational Platform running in 60 seconds.

## Prerequisites

- Go 1.21+ installed
- SQLite3 (comes with macOS)

## Installation

```bash
# Navigate to project
cd /Users/jgirmay/Desktop/gitrepo/pyWork/unified-go

# Install dependencies
make install

# Build the project
make build
```

## Running the Server

### Option 1: Using Make (Recommended)

```bash
make run
```

### Option 2: Direct Execution

```bash
./unified-go
```

### Option 3: Development Mode (with Go run)

```bash
go run cmd/server/main.go
```

### Option 4: Custom Port

```bash
PORT=8080 ./unified-go
```

## Verify Installation

Once the server is running, you should see:

```
2026/02/20 08:52:13 Starting Unified Educational App Server...
2026/02/20 08:52:13 Environment: development
2026/02/20 08:52:13 Port: 5000
2026/02/20 08:52:13 Database initialized successfully
2026/02/20 08:52:13 Server listening on http://localhost:5000
```

## Test the Server

### 1. Health Check

```bash
curl http://localhost:5000/health | jq .
```

Expected output:
```json
{
  "status": "healthy",
  "go_version": "go1.21.x",
  "uptime": "5.2s",
  "timestamp": "2026-02-20T16:52:18Z",
  "goroutines": 8,
  "environment": "development"
}
```

### 2. Visit the Dashboard

Open your browser to: http://localhost:5000/dashboard

You should see the main dashboard with links to all 5 educational apps.

### 3. Test Each App

- **Typing**: http://localhost:5000/typing
- **Math**: http://localhost:5000/math
- **Reading**: http://localhost:5000/reading
- **Piano**: http://localhost:5000/piano

### 4. Test API Endpoints

```bash
# Dashboard stats
curl http://localhost:5000/api/dashboard/stats | jq .

# Typing lessons
curl http://localhost:5000/api/typing | jq .

# Math problems
curl http://localhost:5000/api/math | jq .

# Reading books
curl http://localhost:5000/api/reading | jq .

# Piano songs
curl http://localhost:5000/api/piano | jq .
```

## Database

The database is automatically created at `./data/unified.db` with WAL mode enabled.

### Check Database Tables

```bash
make db-tables
```

Output:
```
math_progress      reading_progress   sessions           users
piano_progress     schema_migrations  typing_progress
```

### View Database Schema

```bash
make db-schema
```

### Reset Database (WARNING: deletes all data)

```bash
make db-reset
```

## Environment Variables

Create a `.env` file (optional):

```bash
cp .env.example .env
```

Edit `.env` to customize:
- `PORT` - Server port (default: 5000)
- `ENVIRONMENT` - dev/staging/production (default: development)
- `DATABASE_URL` - Database path (default: ./data/unified.db)
- `SESSION_SECRET` - Session encryption key (CHANGE IN PRODUCTION)

## Common Tasks

```bash
# Build the project
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Clean build artifacts
make clean

# Show help
make help
```

## Troubleshooting

### Port Already in Use

If port 5000 is already in use:

```bash
PORT=8080 ./unified-go
```

### Database Locked

If you get "database locked" errors, ensure only one instance is running:

```bash
pkill unified-go
```

### Permission Denied

If the binary won't execute:

```bash
chmod +x unified-go
```

### Missing Dependencies

If you see "package not found" errors:

```bash
go mod download
go mod tidy
```

## What's Next?

This is **Phase 1** - the foundation layer. All apps currently return placeholder content.

**Phase 2** will migrate the actual application logic from Python/Flask:
- Real typing lessons and progress tracking
- Actual math problems and scoring
- Reading comprehension with books
- Piano lessons with audio
- User authentication and profiles

See [README.md](README.md) for the complete migration roadmap.

## File Locations

| Item | Location |
|------|----------|
| Binary | `./unified-go` |
| Database | `./data/unified.db` |
| Logs | stdout/stderr |
| Source Code | `./cmd`, `./internal`, `./pkg` |
| Tests | `./internal/*/\*_test.go` |
| Static Assets | `./static/` |
| Templates | `./templates/` |

## Support

For issues:
1. Check server logs for errors
2. Verify `/health` endpoint is responding
3. Ensure environment variables are set correctly
4. Check that database directory has write permissions

---

**Ready to start?** Run `make run` and visit http://localhost:5000/dashboard
