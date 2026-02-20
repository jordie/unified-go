# Architecture Overview

System design and component interactions for unified-go.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Web Browser / API Client                 │
│                   (HTTP/HTTPS Requests)                     │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│               Reverse Proxy (Nginx/Apache)                  │
│              (Load balancing, SSL/TLS, caching)             │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│          unified-go Server (Go + Chi v5 Router)             │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │            HTTP Router (Chi v5)                      │   │
│  │  • Request routing and handling                      │   │
│  │  • Middleware pipeline (logging, auth, CORS)        │   │
│  │  • Static file serving                              │   │
│  │  • Health checks and monitoring                      │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                         │
│  ┌──────────┬──────────┬──────────┬──────────────┐           │
│  ▼          ▼          ▼          ▼              ▼           │
│┌────────┐┌────────┐┌────────┐┌────────┐┌──────────────┐    │
││Reading││ Typing ││ Math   ││ Piano  ││   Static     │    │
││  App  ││  App   ││  App   ││  App   ││   Files      │    │
│└──┬──┬─┘└──┬──┬─┘└──┬──┬─┘└──┬──┬─┘└──────────────┘    │
│   │  │     │  │     │  │     │  │                        │
│   │  │ ┌───┴──┴─────┴──┴─────┴──┴──────────┐            │
│   │  │ │     Business Logic Services       │            │
│   │  │ │  (Score Calculation, Generation, │            │
│   │  │ │   Statistics Aggregation)         │            │
│   │  │ └───┬────────────────────────────────┘            │
│   │  │     │                                              │
│   │  │ ┌───┴────────────────────────────┐                │
│   │  │ │  Data Access Layer (DAL)       │                │
│   │  │ │  (Repository pattern)          │                │
│   │  │ │  • CRUD operations             │                │
│   │  │ │  • Query construction          │                │
│   │  │ │  • Error handling              │                │
│   │  │ └───┬────────────────────────────┘                │
│   │  │     │                                              │
└───┼──┼─────┼──────────────────────────────────────────────┘
    │  │     │
    │  └─────┼───────────────────────────────┐
    │        │                               │
    └────────┼───────────────────────────────┼────────────┐
             │                               │            │
             ▼                               ▼            ▼
    ┌─────────────────────┐      ┌──────────────────┐  ┌──────┐
    │   SQLite Database   │      │  Cache Layer     │  │ Logs │
    │  (data/unified.db)  │      │ (if configured)  │  │      │
    │                     │      │                  │  │      │
    │  • users            │      │  Redis/Memcached │  │ File │
    │  • reading_*        │      │  (sessions, hot  │  │      │
    │  • typing_*         │      │   data)          │  │      │
    │  • math_*           │      │                  │  │      │
    │  • piano_*          │      │                  │  │      │
    └─────────────────────┘      └──────────────────┘  └──────┘
```

## Component Diagram

```
unified-go/
│
├── cmd/
│   └── server/
│       └── main.go           ◄─── Entry point, config loading
│
├── internal/
│   ├── config/
│   │   └── config.go        ◄─── Configuration management
│   ├── database/
│   │   └── db.go            ◄─── Database connection pool
│   ├── middleware/
│   │   ├── auth.go          ◄─── Authentication
│   │   ├── cors.go          ◄─── CORS headers
│   │   ├── logging.go       ◄─── Request logging
│   │   └── recovery.go      ◄─── Panic recovery
│   └── router/
│       └── router.go        ◄─── Main router setup
│
├── pkg/
│   ├── reading/
│   │   ├── models.go        ◄─── Data structures
│   │   ├── service.go       ◄─── Business logic
│   │   ├── repository.go    ◄─── Database access
│   │   ├── router.go        ◄─── HTTP handlers
│   │   ├── handler.go       ◄─── Response formatting
│   │   ├── templates/       ◄─── HTML pages
│   │   ├── static/          ◄─── CSS, JS assets
│   │   ├── integration_test.go ◄─ Tests + benchmarks
│   │   └── README.md        ◄─── Documentation
│   │
│   ├── typing/
│   │   └── (same structure)
│   │
│   ├── math/
│   │   └── (same structure)
│   │
│   └── piano/
│       └── (same structure - Phase 5)
│
├── migrations/
│   ├── 001_init.sql         ◄─── Core schema
│   ├── 002_reading.sql      ◄─── Reading tables
│   ├── 003_typing.sql       ◄─── Typing tables
│   ├── 004_math.sql         ◄─── Math tables
│   └── 005_piano.sql        ◄─── Piano tables
│
├── data/
│   └── unified.db           ◄─── SQLite database
│
├── logs/
│   └── unified-go.log       ◄─── Application logs
│
└── docs/
    ├── ARCHITECTURE.md      ◄─── This file
    ├── DEPLOYMENT.md        ◄─── Production setup
    ├── API_ENDPOINTS_COMPLETE.md
    ├── DATABASE_SCHEMA.md
    └── TROUBLESHOOTING.md
```

## Layer Architecture

### 1. HTTP Router Layer (Chi v5)
- **Responsibility**: Route requests to handlers
- **Key Components**: 
  - Request routing
  - Middleware pipeline
  - Static file serving
  - CORS/Auth/Logging

### 2. Handler/Controller Layer
- **Responsibility**: Parse requests, call service, format responses
- **Files**: `handler.go` in each app
- **Key Tasks**:
  - Unmarshal JSON input
  - Extract user context
  - Call service methods
  - Marshal JSON response
  - Return appropriate HTTP status codes

### 3. Service/Business Logic Layer
- **Responsibility**: Core application logic
- **Files**: `service.go` in each app
- **Key Tasks**:
  - Problem/passage generation
  - Score/metric calculations
  - Statistics aggregation
  - Validation and error handling
  - No database direct access

### 4. Repository/Data Access Layer
- **Responsibility**: Database operations
- **Files**: `repository.go` in each app
- **Key Tasks**:
  - CRUD operations
  - SQL query construction
  - Result mapping to models
  - Index/constraint enforcement
  - Transaction management

### 5. Database Layer
- **Type**: SQLite3
- **Features**: 
  - ACID transactions
  - Foreign key constraints
  - Indexes for performance
  - Schema versioning via migrations

## Data Flow Example: Submit Math Session

```
1. Client Request
   POST /api/math/session/complete
   { "user_id": 1, "correct_answers": 8, ... }
        │
        ▼
2. Chi Router
   Matches POST route to handler
        │
        ▼
3. Handler (router.go)
   - Extract X-User-ID header
   - Parse JSON request body
   - Call service.CompleteSession(ctx, session)
        │
        ▼
4. Service (service.go)
   - Validate session data
   - Calculate score (correct/total * 100)
   - Call repository methods:
     * SaveSession(session)
     * UpdateUserStats(user_id, stats)
        │
        ▼
5. Repository (repository.go)
   - Execute SQL INSERT for math_sessions
   - Execute SQL UPDATE for math_user_stats
   - Commit transaction
        │
        ▼
6. Database (SQLite)
   - Write to disk with fsync
   - Update indexes
   - Return success
        │
        ▼
7. Response
   201 Created
   { "session_id": 301, "score": 90.0, "accuracy": 90.0 }
```

## Request Processing Pipeline

```
HTTP Request
    │
    ▼
Chi Router.ServeHTTP()
    │
    ├─► Middleware 1: RequestID (generate unique ID)
    │
    ├─► Middleware 2: RealIP (get client IP)
    │
    ├─► Middleware 3: Recovery (panic handler)
    │
    ├─► Middleware 4: Logging (log request details)
    │
    ├─► Middleware 5: Compress (gzip response)
    │
    ├─► Middleware 6: CORS (add CORS headers)
    │
    ├─► Middleware 7: Auth (validate session/token)
    │
    ▼
Route Matching
    │
    ▼
Handler Function
    │
    ├─► Parse request body
    ├─► Validate input
    ├─► Call service layer
    ├─► Handle errors
    ├─► Format response
    │
    ▼
Middleware Post-Processing (if any)
    │
    ▼
HTTP Response
```

## Concurrency Model

```
Server (1 main goroutine)
    │
    ├─► Listener (accepts connections)
    │
    ├─► Request Handler Pool
    │   ├─► Goroutine 1 (handling request A)
    │   ├─► Goroutine 2 (handling request B)
    │   ├─► Goroutine 3 (handling request C)
    │   └─► ...
    │
    ├─► Database Connection Pool
    │   ├─► Connection 1 (query from handler A)
    │   ├─► Connection 2 (query from handler B)
    │   └─► ...
    │
    └─► Context Cancellation
        (propagates to handlers and db queries on timeout/disconnect)
```

## Error Handling Strategy

```
User Input Error (400)
    └─► Bad JSON, invalid field
    └─► Return error message

Not Found Error (404)
    └─► Resource doesn't exist
    └─► Return not found message

Authorization Error (401)
    └─► User not authenticated
    └─► Return unauthorized message

Server Error (500)
    └─► Database error, logic error
    └─► Log full error, return generic message
    └─► Alert operations team
```

## Performance Optimizations

### 1. Database Indexes
```
- user_id (all result tables)
- created_at (for date range queries)
- type/difficulty (for filtering)
- Reduces query time from O(n) to O(log n)
```

### 2. Prepared Statements
```
- Handler calls Service
- Service builds SQL with parameterized queries
- Repository executes with parameters
- Prevents SQL injection, faster execution
```

### 3. Connection Pooling
```
- Reuse database connections
- Avoid connection overhead
- Configurable pool size
```

### 4. Caching (optional)
```
- Cache user stats (refresh on update)
- Cache leaderboard (refresh every 5 min)
- Cache problem generation (pre-generate problems)
```

### 5. Efficient Queries
```
- Single query for aggregations (SUM, COUNT in SQL)
- Batch operations in transactions
- Avoid N+1 queries (load related data in one query)
```

## Testing Strategy

### Unit Tests
- Test service layer logic in isolation
- Mock repository layer
- Quick, deterministic

### Integration Tests
- Test with real SQLite in-memory database
- Create schema, insert test data
- Verify end-to-end flow
- Currently: 100+ integration tests

### API Tests
- Use httptest for HTTP handlers
- Send real requests, verify responses
- Check status codes and JSON structure

### Benchmarks
- Measure performance of hot paths
- Problem generation: <1ms
- Database operations: <10ms
- Aggregations: <100µs

## Scalability Considerations

### Current (Single Server)
- ~1000 concurrent users
- ~1000 requests/second
- SQLite handles up to 5GB database

### For Growth to 10K+ Users

**Option 1: Database Replication**
```
Primary (writes)
    │
    ├─► Replica 1 (reads)
    ├─► Replica 2 (reads)
    └─► Replica 3 (reads)
```

**Option 2: Horizontal Scaling**
```
Load Balancer
    ├─► Server 1
    ├─► Server 2
    ├─► Server 3
    └─► Central PostgreSQL
```

**Option 3: Caching Layer**
```
Load Balancer
    ├─► Server 1 ──┐
    ├─► Server 2 ──┼─► Redis Cache
    ├─► Server 3 ──┤
    └─────────────┼─► PostgreSQL
                  └──
```

## Security Considerations

### Input Validation
- All user input validated at handler layer
- Parameterized queries prevent SQL injection
- JSON schema validation

### Authentication
- X-User-ID header or session cookie
- Can integrate with OAuth2, JWT

### CORS
- Configurable allowed origins
- Middleware enforces headers
- Prevents unauthorized cross-origin requests

### HTTPS
- Reverse proxy terminates SSL
- Secure cookies with HttpOnly flag
- Redirect HTTP to HTTPS

## Deployment Architecture

```
Development
    └─► Single server, SQLite

Staging
    └─► Single server, SQLite
    └─► With monitoring/alerting

Production
    ├─► Load Balancer
    ├─► 3x App Servers (unified-go)
    ├─► PostgreSQL (primary + replicas)
    ├─► Redis Cache
    ├─► Nginx Reverse Proxy
    └─► Monitoring (Prometheus, ELK)
```

## Dependency Graph

```
main.go
    ├─► config
    ├─► database (sqlite3)
    ├─► router
    │   ├─► reading router
    │   ├─► typing router
    │   ├─► math router
    │   └─► piano router
    │
    └─► middleware
        ├─► chi/middleware
        ├─► recovery
        ├─► logging
        ├─► auth
        └─► CORS

Each app router
    ├─► handler
    ├─► service
    └─► repository
        └─► database
```

---

See [DEPLOYMENT.md](DEPLOYMENT.md) for operational details or [API_ENDPOINTS_COMPLETE.md](API_ENDPOINTS_COMPLETE.md) for API documentation.
