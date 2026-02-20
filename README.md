# Unified Educational Platform - Go Edition

**Phase 1: Foundation Layer** - Complete Go-based foundation for migrating 5 educational apps from Python/Flask to Go.

## Overview

This is the unified backend that will power multiple educational applications:
- **Typing App** - Interactive typing practice with lessons
- **Math App** - Math problem solving and practice
- **Reading App** - Reading comprehension with books and quizzes
- **Piano App** - Piano learning with guided lessons
- **Dashboard** - Central hub for all applications

## Project Structure

```
unified-go/
├── cmd/server/main.go           # Application entry point
├── internal/                    # Private application code
│   ├── router/router.go         # HTTP routing with chi
│   ├── middleware/
│   │   ├── auth.go              # Session management
│   │   ├── logging.go           # Request logging & recovery
│   │   └── cors.go              # CORS configuration
│   ├── database/
│   │   ├── pool.go              # SQLite connection pool with WAL
│   │   └── migrations.go        # Database schema migrations
│   └── config/config.go         # Environment configuration
├── pkg/                         # Public reusable packages
│   ├── typing/handler.go        # Typing app handlers
│   ├── math/handler.go          # Math app handlers
│   ├── reading/handler.go       # Reading app handlers
│   ├── piano/handler.go         # Piano app handlers
│   └── dashboard/handler.go     # Dashboard handlers
├── templates/                   # HTML templates (go html/template)
├── static/                      # Static assets (CSS, JS, images)
├── data/                        # SQLite database files (gitignored)
├── go.mod                       # Go module definition
└── go.sum                       # Dependency checksums
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- SQLite3 (included via go-sqlite3)

### Installation

```bash
# Navigate to project directory
cd /Users/jgirmay/Desktop/gitrepo/pyWork/unified-go

# Download dependencies
go mod download

# Verify dependencies
go mod tidy
```

### Running the Server

```bash
# Development mode (default)
go run cmd/server/main.go

# Build binary
go build -o unified-go cmd/server/main.go

# Run binary
./unified-go
```

The server will start on `http://localhost:5000`

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Health check
curl http://localhost:5000/health
```

## Environment Variables

Configure the application using these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `5000` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `ENVIRONMENT` | `development` | Environment (development, staging, production) |
| `DATABASE_URL` | `./data/unified.db` | SQLite database path |
| `SESSION_SECRET` | (auto-generated) | Session encryption key (CHANGE IN PRODUCTION) |
| `SESSION_NAME` | `unified_session` | Session cookie name |
| `CORS_ORIGIN` | `*` | Allowed CORS origin |
| `STATIC_DIR` | `./static` | Static files directory |
| `TEMPLATE_DIR` | `./templates` | Templates directory |

### Example Configuration

```bash
# .env file
PORT=8080
ENVIRONMENT=production
DATABASE_URL=/var/lib/unified/production.db
SESSION_SECRET=your-super-secret-key-change-this
CORS_ORIGIN=https://yourdomain.com
```

## API Endpoints

### Health Check

```bash
GET /health
```

Returns server health status with Go version, uptime, and resource usage.

### Applications

| Endpoint | Description |
|----------|-------------|
| `GET /` | Redirects to dashboard |
| `GET /dashboard` | Main dashboard homepage |
| `GET /typing` | Typing app homepage |
| `GET /math` | Math app homepage |
| `GET /reading` | Reading app homepage |
| `GET /piano` | Piano app homepage |

### API Routes (JSON)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/typing` | GET | List typing lessons |
| `/api/typing/progress` | POST | Save typing progress |
| `/api/math` | GET | List math problems |
| `/api/math/progress` | POST | Save math progress |
| `/api/reading` | GET | List reading books |
| `/api/reading/progress` | POST | Save reading progress |
| `/api/piano` | GET | List piano songs |
| `/api/piano/progress` | POST | Save piano progress |
| `/api/dashboard/stats` | GET | Get dashboard statistics |

## Database Schema

The application uses SQLite with WAL (Write-Ahead Logging) mode for concurrent access.

### Tables

- `users` - User accounts and authentication
- `sessions` - Session storage
- `typing_progress` - Typing lesson progress
- `math_progress` - Math problem progress
- `reading_progress` - Reading book progress
- `piano_progress` - Piano song progress
- `schema_migrations` - Database migration tracking

### Migrations

Migrations run automatically on server startup. Current version is tracked in the `schema_migrations` table.

## Architecture Decisions

### Why Go?

- **Performance**: 10-50x faster than Python for web requests
- **Concurrency**: Built-in goroutines for handling multiple connections
- **Type Safety**: Compile-time error detection
- **Single Binary**: No dependency management nightmares
- **Memory Efficiency**: Lower memory footprint than Python

### Why Chi Router?

- Lightweight and composable
- Standard library compatible
- Excellent middleware support
- Context-based request handling

### Why SQLite with WAL?

- Zero configuration database
- Excellent for read-heavy workloads
- WAL mode enables concurrent reads and writes
- Perfect for local/embedded applications

### Session Compatibility

The session system uses `gorilla/sessions` configured to be compatible with Python Flask sessions:
- Same cookie-based storage
- Compatible encryption (can migrate existing sessions)
- Same session timeout behavior

## Development Workflow

### Adding a New Feature

1. Create handler in appropriate `pkg/` directory
2. Add route in `internal/router/router.go`
3. Add database migration if needed in `internal/database/migrations.go`
4. Test with `go test`
5. Deploy

### Adding a New App

1. Create new package: `pkg/newapp/handler.go`
2. Add routes in router
3. Add database tables via migration
4. Create templates/static assets
5. Update dashboard to link to new app

## Deployment

### Development

```bash
go run cmd/server/main.go
```

### Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o unified-go cmd/server/main.go

# Run with production config
ENVIRONMENT=production \
SESSION_SECRET=your-secret-key \
DATABASE_URL=/var/lib/unified/production.db \
./unified-go
```

### Docker (Coming in Phase 3)

```bash
docker build -t unified-go .
docker run -p 5000:5000 unified-go
```

## Migration Plan

### Phase 1: Foundation (COMPLETE)
- ✅ Go project structure
- ✅ HTTP router with middleware
- ✅ Database connection pool
- ✅ Session management
- ✅ Placeholder handlers for all apps
- ✅ Health check endpoint

### Phase 2: App Migration (Next)
- Migrate typing app Python code to Go
- Migrate math app Python code to Go
- Migrate reading app Python code to Go
- Migrate piano app Python code to Go
- Port all templates and static assets

### Phase 3: Optimization
- Implement caching
- Add WebSocket support
- Optimize database queries
- Add monitoring/metrics
- Docker containerization

### Phase 4: Production
- Load testing
- Security audit
- Deploy to production
- Monitor and iterate

## Dependencies

- **github.com/go-chi/chi/v5** - HTTP router and middleware
- **github.com/mattn/go-sqlite3** - SQLite3 driver (CGO required)
- **github.com/gorilla/sessions** - Session management
- **github.com/gorilla/securecookie** - Secure cookie encoding

## Performance Benchmarks

(To be filled after Phase 2 migration)

Expected improvements over Python/Flask:
- Request latency: 10-50x faster
- Memory usage: 50-70% reduction
- Concurrent requests: 10x+ improvement

## Contributing

This is a private educational project. For questions or suggestions, contact the development team.

## License

Proprietary - All rights reserved.

## Support

For issues or questions:
- Check the `/health` endpoint for server status
- Review logs at startup for configuration issues
- Ensure environment variables are set correctly
- Verify SQLite database directory has write permissions

## Next Steps

1. Test all endpoints: `curl http://localhost:5000/health`
2. Visit dashboard: `http://localhost:5000/dashboard`
3. Explore each app placeholder
4. Begin Phase 2 migration with typing app
5. Monitor performance and optimize as needed
