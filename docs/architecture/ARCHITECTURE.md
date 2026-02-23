# GAIA_GO Architecture Documentation

## Overview

GAIA_GO is a Go microservices framework for educational applications featuring four main learning platforms: Math, Piano, Reading, and Typing. The framework is designed for scalability, monitoring, and performance optimization.

## Project Structure

```
GAIA_GO/
├── cmd/                          # Command-line applications
│   └── server/                   # Main application server
│       └── main.go               # Application entry point
│
├── internal/                     # Internal packages (not exported)
│   ├── api/                      # API request/response structures & helpers
│   │   ├── constants.go          # Error codes and standard messages
│   │   ├── errors.go             # APIError type and error definitions
│   │   ├── response.go           # Response builders and envelopes
│   │   └── dto.go                # Request/Response DTOs for all apps
│   │
│   ├── middleware/               # HTTP middleware
│   │   ├── auth.go               # Authentication middleware
│   │   └── unified.go            # Unified logging & error handling
│   │
│   ├── models/                   # Shared data structures
│   │   └── models.go             # Central data models for all apps
│   │
│   ├── monitoring/               # Health & metrics
│   │   ├── health.go             # Health check endpoints
│   │   ├── metrics.go            # Prometheus metrics registration
│   │   ├── prometheus.go         # Prometheus client integration
│   │   └── status.go             # System status tracking
│   │
│   ├── orchestration/            # Subsystem pooling & management
│   │   ├── subsystems/           # Pooled subsystem implementations
│   │   ├── factory.go            # Factory pattern for subsystems
│   │   ├── loader.go             # Configuration loader
│   │   ├── manager.go            # Lifecycle manager
│   │   ├── pool.go               # Connection pooling
│   │   └── registry.go           # Service registry
│   │
│   ├── session/                  # Session management
│   │   └── session.go            # Session handling
│   │
│   └── templates/                # Template rendering
│       └── renderer.go           # HTML template rendering
│
├── pkg/                          # Public packages
│   ├── apps/                     # Educational applications
│   │   ├── math/                 # Math learning app
│   │   │   ├── handlers.go       # HTTP request handlers
│   │   │   └── handlers_test.go  # Handler tests
│   │   │
│   │   ├── piano/                # Piano learning app
│   │   │   ├── handlers.go
│   │   │   └── handlers_test.go
│   │   │
│   │   ├── reading/              # Reading practice app
│   │   │   ├── handlers.go
│   │   │   └── handlers_test.go
│   │   │
│   │   └── typing/               # Typing speed/race app
│   │       ├── handlers.go
│   │       └── handlers_test.go
│   │
│   └── router/                   # HTTP routing
│       └── router.go             # Gin router setup
│
├── configs/                      # Configuration files
│   ├── prometheus.yml            # Prometheus metrics configuration
│   └── alerts.yml                # Alert rules configuration
│
├── migrations/                   # Database migrations
│   └── *.sql                     # Migration files
│
├── web/                          # Web assets
│   ├── static/                   # Static files (CSS, JS, images)
│   └── templates/                # HTML templates
│       ├── math/                 # Math app templates
│       ├── piano/                # Piano app templates
│       ├── reading/              # Reading app templates
│       └── typing/               # Typing app templates
│
├── examples/                     # Example files and documentation
│
├── build/                        # Build artifacts
│   ├── gaia-server              # Compiled binary
│   └── subsystems.test          # Test binary
│
├── profiling/                    # Performance profiling data
│   ├── cpu.prof                 # CPU profile
│   └── mem.prof                 # Memory profile
│
├── docs/                         # Documentation
│   ├── architecture/             # Architecture & design docs
│   ├── monitoring/               # Monitoring & observability docs
│   └── optimization/             # Performance & optimization docs
│
├── go.mod                        # Go module definition
└── go.sum                        # Go module checksums
```

## Core Packages

### `internal/api` - API Contract Layer
Provides standardized request/response handling across all applications.

**Purpose:** Ensure consistent API design, error handling, and data validation.

**Contents:**
- **constants.go** - Standardized error codes and messages
- **errors.go** - APIError type with structured error responses
- **response.go** - Response builders for success/error responses
- **dto.go** - Request/Response data transfer objects for all handlers

**Key Types:**
```go
// Standardized API Error
type APIError struct {
    Code       string                 // Error code (e.g., "INVALID_REQUEST")
    Message    string                 // Human-readable message
    StatusCode int                    // HTTP status code
    Details    map[string]interface{} // Additional error details
}

// Standard response envelope
type Response struct {
    Success bool        // Operation success
    Data    interface{} // Response data
    Error   *APIError   // Error details if failed
    Message string      // Optional message
}
```

**Usage Pattern:**
```go
// Error response
api.RespondWithError(c, api.ErrBadRequest)

// Success response
api.RespondWith(c, http.StatusOK, data)

// Creation response
api.RespondWithCreated(c, resourceID, data)
```

### `internal/models` - Data Models
Defines shared data structures used across all applications and the database layer.

**Key Models:**
- **BaseResult** - Common result structure with ID, UserID, AppName, CreatedAt
- **UserStats** - Aggregated statistics (sessions, scores, time)
- **App-specific Models:**
  - Math: MathProblem, MathResult, MathStats, MathWeakness
  - Piano: PianoSession, PianoStats, PianoBadge
  - Reading: ReadingPassage, ReadingResult, ReadingStats, WordMastery
  - Typing: TypingResult, TypingStats, Race, RacingStats
- **Gamification:** XPLog, LeaderboardEntry
- **Progression:** Goal, ProgressMilestone
- **User Journal:** UserJournal

### `internal/middleware` - Request Processing
Intercepts and processes HTTP requests before reaching handlers.

**Auth Middleware** (`auth.go`)
- Validates authentication tokens
- Extracts user context (UserID, session info)
- Blocks unauthorized requests

**Unified Middleware** (`unified.go`)
- Centralized logging
- Error handling wrapper
- Request/response correlation IDs
- Performance metrics collection

### `internal/orchestration` - Subsystem Management
Manages pooled connections and resources (database, cache, services).

**Components:**
- **pool.go** - Connection pooling for performance
- **factory.go** - Creates subsystem instances
- **manager.go** - Lifecycle management (init, health, shutdown)
- **registry.go** - Service discovery and lookup
- **loader.go** - Configuration loading

### `pkg/router` - HTTP Routing
Configures Gin web framework routes for all applications.

**Structure:**
```go
// Routes organized by app
r.POST("/math/problems", handlers.MathHandler)
r.POST("/piano/sessions", handlers.PianoHandler)
r.POST("/reading/words", handlers.ReadingHandler)
r.POST("/typing/results", handlers.TypingHandler)
```

### `pkg/apps/*` - Application Handlers
Each app implements request handlers for its specific endpoints.

**Handler Pattern:**
```go
// Parse request DTO
var req api.SaveSessionRequest
if err := c.BindJSON(&req); err != nil {
    api.RespondWithError(c, api.ErrBadRequest)
    return
}

// Get user context from middleware
userID := c.GetInt64("user_id")

// Process business logic
// Save to database, calculate metrics, award XP

// Respond with standardized response
api.RespondWithCreated(c, resultID, result)
```

## Data Flow

### Request Flow
```
HTTP Request
    ↓
Router (pkg/router/router.go)
    ↓
Middleware (internal/middleware)
  - Auth validation
  - Request logging
  - Error handling wrapper
    ↓
Handler (pkg/apps/*/handlers.go)
  - Parse request DTO (internal/api/dto.go)
  - Validate input
  - Call business logic
  - Build response using models (internal/models)
    ↓
Response Builder (internal/api/response.go)
  - Construct standard response envelope
  - Set appropriate HTTP status code
    ↓
HTTP Response
```

### Session Save Flow (Example)
```
POST /typing/results
    ↓
TypingHandler.SaveResult()
    ↓
Parse SaveResultRequest DTO
    ↓
Validate fields (WPM, Accuracy, TestType)
    ↓
Create TypingResult model
    ↓
Calculate metrics (average WPM, accuracy, time)
    ↓
Save to database via orchestration layer
    ↓
Calculate XP reward
    ↓
Update user stats
    ↓
Return SessionResultResponse with XP earned
```

## API Design

### Request/Response Pattern

**Standard Request:**
```json
{
  "difficulty": "hard",
  "accuracy": 95.5,
  "total_time": 300,
  "wpm": 120,
  "correct_characters": 500
}
```

**Success Response:**
```json
{
  "success": true,
  "data": {
    "id": 12345,
    "user_id": 1,
    "score": 95,
    "xp_earned": 250,
    "created_at": "2025-02-23T10:30:00Z"
  },
  "message": "Result saved successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request parameters",
    "details": {
      "field": "accuracy",
      "reason": "must be between 0 and 100"
    }
  }
}
```

### HTTP Status Codes
- **200 OK** - Successful GET/PUT request
- **201 Created** - Resource successfully created (POST)
- **204 No Content** - Successful DELETE request
- **400 Bad Request** - Invalid input parameters
- **401 Unauthorized** - Missing or invalid authentication
- **403 Forbidden** - User lacks permission
- **404 Not Found** - Resource not found
- **409 Conflict** - Resource already exists
- **500 Internal Server Error** - Server-side error
- **503 Service Unavailable** - Service down or maintenance

## Database Design

Models are persisted through the orchestration layer, which handles:
- Connection pooling
- Transaction management
- Query optimization
- Migration handling

**Key Tables:**
- `users` - User profiles
- `results` - Session results (polymorphic by app_name)
- `stats` - Cached user statistics
- `xp_logs` - XP transactions
- `leaderboards` - Rankings
- `word_mastery` - Reading app word progress
- `goals` - User learning goals
- `journals` - User learning journals

## Monitoring & Observability

### Health Checks
- Endpoint: `/health`
- Returns system status, database connectivity, service health
- Configured in `internal/monitoring/health.go`

### Prometheus Metrics
- Endpoint: `/metrics`
- Metrics tracked:
  - HTTP request count (by endpoint, method, status)
  - Request duration (by endpoint)
  - Database query performance
  - Error rates
  - Active connections
- Configuration: `configs/prometheus.yml`

### Logging
- Unified logging via middleware
- Request/response correlation IDs
- Performance metrics per request
- Error stack traces for debugging

## Performance Optimization

### Strategies Implemented
1. **Connection Pooling** - Reuse database connections
2. **Query Caching** - Cache frequently accessed stats
3. **Batch Operations** - Bulk insert/update for efficiency
4. **Profiling** - CPU and memory profiling available in `profiling/`
5. **Metrics** - Real-time performance monitoring

### Profiling
- Generate CPU profile: `go test -cpuprofile=cpu.prof`
- Generate memory profile: `go test -memprofile=mem.prof`
- Analyze: `go tool pprof cpu.prof`
- Profiles stored in `profiling/` directory

## Configuration

### Prometheus Configuration (`configs/prometheus.yml`)
Defines metrics collection endpoints and intervals.

### Alert Rules (`configs/alerts.yml`)
Defines alerting conditions for high error rates, slow responses, system issues.

### Environment Setup
- Database connection strings
- Service ports
- Authentication secrets
- Logging levels

## Development Workflow

### Adding a New Endpoint

1. **Define Request/Response DTOs** in `internal/api/dto.go`
   ```go
   type MyAppRequest struct {
       Field string `json:"field" binding:"required"`
   }
   ```

2. **Create Handler** in `pkg/apps/myapp/handlers.go`
   ```go
   func (h *Handler) MyEndpoint(c *gin.Context) {
       var req api.MyAppRequest
       if err := c.BindJSON(&req); err != nil {
           api.RespondWithError(c, api.ErrBadRequest)
           return
       }
       // Business logic
       api.RespondWith(c, http.StatusOK, result)
   }
   ```

3. **Register Route** in `pkg/router/router.go`
   ```go
   r.POST("/myapp/endpoint", handler.MyEndpoint)
   ```

4. **Add Tests** in `pkg/apps/myapp/handlers_test.go`

### Running the Application
```bash
go run cmd/server/main.go
```

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o build/gaia-server cmd/server/main.go
```

## Technology Stack

- **Framework:** Gin (Go web framework)
- **Database:** SQLite/SQL (via orchestration layer)
- **Monitoring:** Prometheus + Grafana
- **Language:** Go 1.23+
- **Testing:** Go testing package

## Migration from Ad-hoc API Design

### Current State (Before Consolidation)
- Inline request structs in handlers (50+ definitions)
- Inconsistent error responses (14+ variations)
- Ad-hoc response building with gin.H{}
- No centralized error codes

### Consolidated State (After)
- Centralized DTOs in `internal/api/dto.go`
- Unified error handling with `APIError`
- Standard response envelope with helper functions
- Consistent error codes in `constants.go`

**Benefits:**
- 37% reduction in handler code
- Improved type safety
- Consistent API contract
- Easier client integration
- Better error tracking and debugging

## Future Enhancements

1. **API Versioning** - Add `/v1/` prefix support
2. **Rate Limiting** - Implement per-user request limits
3. **Caching Layer** - Add Redis for distributed caching
4. **AsyncJobs** - Background task processing
5. **Webhooks** - Event notification system
6. **GraphQL** - Alternative query language support
7. **OpenAPI** - Automatic API documentation generation

## Related Documentation

- [Monitoring Guide](../monitoring/MONITORING.md)
- [Optimization Guide](../optimization/OPTIMIZATION.md)
- [Prometheus Quick Start](../monitoring/PROMETHEUS_QUICKSTART.md)
- [Grafana Dashboards](../monitoring/GRAFANA_DASHBOARDS_GUIDE.md)
