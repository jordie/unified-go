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

#### Typing App
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/typing/api/leaderboard` | GET | Top typing performers |
| `/typing/api/test` | POST | Create typing test |
| `/typing/api/users/{id}/stats` | GET | User typing statistics |

#### Math App
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/math/api/problem` | POST | Generate math problem |
| `/math/api/session/complete` | POST | Complete quiz session |
| `/math/api/users/{id}/stats` | GET | User math statistics |
| `/math/api/leaderboard` | GET | Math leaderboard |

#### Reading App
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/reading/api/passages` | GET | List reading passages |
| `/reading/api/answer` | POST | Submit reading answer |
| `/reading/api/users/{id}/stats` | GET | User reading statistics |
| `/reading/api/leaderboard` | GET | Reading leaderboard |

#### Piano App
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/piano/api/songs` | GET | List piano songs (20 pieces) |
| `/piano/api/songs/{id}` | GET | Get song details |
| `/piano/api/lessons` | POST | Start piano lesson |
| `/piano/api/users/{id}/progress` | GET | User piano progress |
| `/piano/api/leaderboard` | GET | Piano leaderboard |

#### Dashboard
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/dashboard` | GET | Main dashboard |
| `/health` | GET | Server health status |

## Database Schema

The application uses SQLite with WAL (Write-Ahead Logging) mode for concurrent access.

### Tables

**Core**
- `users` - User accounts and authentication
- `sessions` - Session storage
- `schema_migrations` - Database migration tracking

**Typing App**
- `typing_progress` - Typing lesson progress
- `typing_tests` - Individual typing tests
- `typing_results` - Test results with metrics

**Math App**
- `math_progress` - Math problem progress
- `math_problems` - Problem catalog
- `math_sessions` - Quiz sessions
- `math_user_stats` - User statistics

**Reading App**
- `reading_progress` - Reading book progress
- `reading_passages` - Reading comprehension passages
- `reading_questions` - Associated quiz questions
- `reading_results` - Answer results

**Piano App**
- `songs` - Piano song catalog (20 classical pieces)
- `piano_lessons` - Practice sessions
- `practice_sessions` - MIDI recordings
- `music_theory_quizzes` - Theory quiz attempts
- `user_music_metrics` - User statistics

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

### Phase 2: Typing App (COMPLETE)
- ✅ Typing practice with WPM and accuracy tracking
- ✅ Real-time metrics calculation
- ✅ User leaderboards
- ✅ Progress history

### Phase 3: Math App (COMPLETE)
- ✅ 6 problem types (addition, subtraction, multiplication, division, fractions, algebra)
- ✅ 4 difficulty levels (easy, medium, hard, very hard)
- ✅ Score calculation and tracking
- ✅ User statistics and recommendations

### Phase 4: Reading App (COMPLETE)
- ✅ Reading comprehension with passages
- ✅ Multiple-choice questions
- ✅ Performance metrics
- ✅ Category-based filtering

### Phase 5: Piano App (IN PROGRESS)
- ✅ Song catalog with 20+ classical pieces
- ✅ 4 difficulty levels (beginner to master)
- ✅ Practice session tracking
- ✅ Performance metrics (accuracy, tempo)
- 🔄 Handler optimization (Subtask 4: Documentation)
- ⏳ Music theory quizzes
- ⏳ MIDI recording and playback

### Phase 6: Optimization & Production
- Performance tuning
- Load testing
- Security audit
- Docker containerization
- Deploy to production

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
