# Development Guide - Reading & Piano Apps

Comprehensive guide for developing, testing, and contributing to the Phase 4 (Reading) and Phase 5 (Piano) educational applications.

## Table of Contents

1. [Setup & Installation](#setup--installation)
2. [Project Structure](#project-structure)
3. [Testing](#testing)
4. [Architecture](#architecture)
5. [Adding Features](#adding-features)
6. [Performance Optimization](#performance-optimization)
7. [Debugging](#debugging)
8. [Code Standards](#code-standards)
9. [Common Tasks](#common-tasks)
10. [Troubleshooting](#troubleshooting)

---

## Setup & Installation

### Prerequisites

- Go 1.21 or later
- SQLite3
- Git
- A text editor or IDE (VS Code recommended)

### Initial Setup

```bash
# Clone repository
git clone https://github.com/your-org/unified-go.git
cd unified-go

# Install Go dependencies
go mod download
go mod verify

# For Reading app
cd pkg/reading
go mod download

# For Piano app
cd pkg/piano
go mod download
```

### Development Environment

```bash
# Create local configuration
export GAIA_DB_PATH="./data"
export GAIA_LOG_LEVEL="debug"
export GAIA_ENV="development"

# Optional: Run database migrations
sqlite3 data/reading.db < migrations/reading_schema.sql
sqlite3 data/piano.db < migrations/piano_schema.sql
```

---

## Project Structure

### Reading App (`pkg/reading/`)

```
reading/
├── models.go                  # Data structures, 150-200 lines
├── service.go                 # Business logic, 200-300 lines
├── repository.go              # Database operations, 250-350 lines
├── router.go                  # HTTP handlers, 200-250 lines
├── handler.go                 # Response formatting, 50-100 lines
├── models_test.go             # Unit tests for models
├── service_test.go            # Unit tests for service
├── repository_test.go         # Unit tests for repository
├── integration_test.go        # Full integration tests + benchmarks
├── templates/                 # HTML templates
│   ├── base.html              # Base layout
│   ├── books.html             # Book listing
│   ├── read.html              # Reading interface
│   ├── dashboard.html         # Statistics
│   └── leaderboard.html       # Rankings
└── README.md                  # Documentation
```

### Piano App (`pkg/piano/`)

```
piano/
├── models.go                  # Data structures, 200-250 lines
├── service.go                 # Business logic, 250-350 lines
├── repository.go              # Database operations, 300-400 lines
├── router.go                  # HTTP handlers, 250-300 lines
├── handler.go                 # Response formatting, 50-100 lines
├── models_test.go             # Unit tests for models
├── service_test.go            # Unit tests for service
├── repository_test.go         # Unit tests for repository
├── integration_test.go        # Full integration tests + benchmarks
├── templates/                 # HTML templates
│   ├── base.html              # Base layout
│   ├── songs.html             # Song catalog
│   ├── practice.html          # Practice interface
│   ├── dashboard.html         # Statistics
│   └── leaderboard.html       # Rankings
└── README.md                  # Documentation
```

---

## Testing

### Test Strategy

The applications use a three-layer testing approach:

1. **Unit Tests** - Test individual functions and methods
2. **Repository Tests** - Test database operations
3. **Integration Tests** - Test complete API workflows

### Running Tests

```bash
# Run all tests for an app
go test -v ./pkg/reading
go test -v ./pkg/piano

# Run specific test file
go test -v -run ^TestCreateAndRetrieveBook ./pkg/reading

# Run with coverage
go test -cover ./pkg/reading
go test -coverprofile=coverage.out ./pkg/reading
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./pkg/reading
go test -bench=. -benchmem ./pkg/piano

# Run specific benchmark
go test -bench=BenchmarkReadingSession ./pkg/reading

# Run tests with race detector (detect concurrency issues)
go test -race ./pkg/reading ./pkg/piano

# Run tests with verbose output and specific time
go test -v -timeout 30s ./pkg/reading
```

### Test Organization

**Unit Tests** (models_test.go, service_test.go, repository_test.go):
```go
func TestModelValidation(t *testing.T) {
    // Arrange
    data := setupTestData()

    // Act
    result := performOperation(data)

    // Assert
    if !validate(result) {
        t.Errorf("Expected valid, got invalid")
    }
}
```

**Integration Tests** (integration_test.go):
```go
func TestCreateAndRetrieveBook(t *testing.T) {
    // Setup test infrastructure
    ti := setupIntegration(t)
    defer ti.db.Close()

    // Test complete workflow
    bookData := createTestBook()
    // ... make API call, verify response
}
```

**Benchmarks** (integration_test.go):
```go
func BenchmarkReadingSession(b *testing.B) {
    ti := setupBenchmark(b)
    defer ti.db.Close()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Run operation being benchmarked
    }
}
```

### Key Test Files

**Reading App:**
- `integration_test.go` - 9 integration tests + 3 benchmarks (~528 lines)
- `models_test.go` - Model validation and calculation tests
- `repository_test.go` - Database operation tests
- `service_test.go` - Business logic tests

**Piano App:**
- `integration_test.go` - 13 integration tests + 3 benchmarks (~512 lines)
- `models_test.go` - Model validation, MIDI handling, calculations
- `repository_test.go` - Database and MIDI blob operations
- `service_test.go` - Lesson processing, recommendations, theory

### Coverage Goals

- **Models**: 100% coverage (all validation paths)
- **Repository**: 95%+ coverage (all SQL paths)
- **Service**: 90%+ coverage (business logic)
- **Router**: 80%+ coverage (API endpoints)

### Writing New Tests

1. **Unit Test Template:**
```go
func TestNewFeature(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()

    repo := NewRepository(db)
    ctx := context.Background()

    // Execute
    result, err := repo.DoSomething(ctx, input)

    // Verify
    if err != nil {
        t.Fatalf("DoSomething() error = %v", err)
    }
    if result == nil {
        t.Error("Expected result, got nil")
    }
}
```

2. **Integration Test Template:**
```go
func TestNewAPIEndpoint(t *testing.T) {
    ti := setupIntegration(t)
    defer ti.db.Close()

    // Prepare test data
    data := map[string]interface{}{
        "param1": "value1",
        "param2": 123,
    }

    // Make request
    body, _ := json.Marshal(data)
    req := httptest.NewRequest("POST", "/api/endpoint", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Execute
    ti.router.ServeHTTP(w, req)

    // Verify
    if w.Code != http.StatusCreated {
        t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
    }
}
```

3. **Benchmark Template:**
```go
func BenchmarkNewOperation(b *testing.B) {
    ti := setupBenchmark(b)
    defer ti.db.Close()

    // Setup test data
    data := prepareData()

    // Run benchmark
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = performOperation(data)
    }
}
```

---

## Architecture

### Layered Architecture

```
┌─────────────────────────────────┐
│   HTTP Router / Handler Layer    │  router.go, handler.go
│   (Chi framework, JSON encoding) │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│    Service / Business Logic      │  service.go
│  (Validations, aggregations)     │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│   Repository / Data Access       │  repository.go
│      (SQL queries, CRUD)         │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│    SQLite Database               │  .db file
└─────────────────────────────────┘
```

### Data Flow Example: Create Reading Session

```
HTTP Request
    │
    ▼
POST /api/sessions (router.go)
    │
    ├─ Parse JSON body
    ├─ Validate required fields
    │
    ▼
service.ProcessSession() (service.go)
    │
    ├─ Calculate metrics (WPM, accuracy)
    ├─ Validate business rules
    │
    ▼
repo.SaveSession() (repository.go)
    │
    ├─ Build SQL INSERT
    ├─ Execute query
    │
    ▼
SQLite Database
    │
    ▼
Return session ID
    │
    ▼
respondJSON() (handler.go)
    │
    ▼
HTTP Response {status: "success", data: {...}}
```

### Key Design Patterns

1. **Repository Pattern**
   - Abstracts database operations
   - Easy to mock for testing
   - Enables database switching

2. **Service Layer Pattern**
   - Contains business logic
   - Validates inputs
   - Coordinates multiple repositories

3. **Handler Pattern**
   - Converts HTTP to domain objects
   - Validates HTTP requests
   - Formats responses

---

## Adding Features

### Step-by-Step: Adding a New API Endpoint

#### Example: Add "Save Reading Goal" endpoint

**1. Add Model** (`models.go`)
```go
type ReadingGoal struct {
    ID          uint
    UserID      uint
    WPMTarget   float64
    WeeksTarget int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Validate checks if ReadingGoal is valid
func (r *ReadingGoal) Validate() error {
    if r.UserID == 0 {
        return errors.New("user_id is required")
    }
    if r.WPMTarget < 0 {
        return errors.New("wpm_target must be positive")
    }
    return nil
}
```

**2. Add Repository Method** (`repository.go`)
```go
func (r *Repository) SaveReadingGoal(ctx context.Context, goal *ReadingGoal) (uint, error) {
    if goal == nil {
        return 0, errors.New("goal cannot be nil")
    }

    if err := goal.Validate(); err != nil {
        return 0, fmt.Errorf("invalid goal: %w", err)
    }

    stmt := `INSERT INTO reading_goals (user_id, wpm_target, weeks_target, created_at, updated_at)
             VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

    result, err := r.db.ExecContext(ctx, stmt, goal.UserID, goal.WPMTarget, goal.WeeksTarget)
    if err != nil {
        return 0, fmt.Errorf("failed to save goal: %w", err)
    }

    id, _ := result.LastInsertId()
    return uint(id), nil
}
```

**3. Add Service Method** (`service.go`)
```go
func (s *Service) SetReadingGoal(ctx context.Context, userID uint, wpmTarget float64, weeks int) (*ReadingGoal, error) {
    if userID == 0 {
        return nil, errors.New("user_id is required")
    }

    goal := &ReadingGoal{
        UserID:      userID,
        WPMTarget:   wpmTarget,
        WeeksTarget: weeks,
    }

    id, err := s.repo.SaveReadingGoal(ctx, goal)
    if err != nil {
        return nil, err
    }

    goal.ID = id
    return goal, nil
}
```

**4. Add Router Handler** (`router.go`)
```go
func (r *Router) SetReadingGoal(w http.ResponseWriter, req *http.Request) {
    var goalData struct {
        UserID    uint    `json:"user_id"`
        WPMTarget float64 `json:"wpm_target"`
        Weeks     int     `json:"weeks"`
    }

    if err := json.NewDecoder(req.Body).Decode(&goalData); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    goal, err := r.service.SetReadingGoal(req.Context(), goalData.UserID, goalData.WPMTarget, goalData.Weeks)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusCreated, goal)
}

// In the Routes() method, add:
router.Post("/api/goals", r.SetReadingGoal)
```

**5. Add Integration Test** (`integration_test.go`)
```go
func TestSetReadingGoal(t *testing.T) {
    ti := setupIntegration(t)
    defer ti.db.Close()

    goalData := map[string]interface{}{
        "user_id": 1,
        "wpm_target": 100.0,
        "weeks": 12,
    }

    body, _ := json.Marshal(goalData)
    req := httptest.NewRequest("POST", "/api/goals", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    ti.router.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Errorf("Expected 201, got %d", w.Code)
    }

    var response ReadingGoal
    json.Unmarshal(w.Body.Bytes(), &response)

    if response.WPMTarget != 100.0 {
        t.Errorf("Expected WPMTarget 100.0, got %f", response.WPMTarget)
    }
}
```

**6. Update Database Schema** (migration file)
```sql
CREATE TABLE IF NOT EXISTS reading_goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    wpm_target REAL NOT NULL,
    weeks_target INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_goals_user ON reading_goals(user_id);
```

---

## Performance Optimization

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./pkg/reading
go tool pprof -http=:8080 cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./pkg/reading
go tool pprof -http=:8080 mem.prof

# Benchmark with profiling
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./pkg/reading
```

### Common Optimizations

1. **Database Queries**
   - Add indexes on frequently queried columns
   - Use LIMIT for large result sets
   - Avoid N+1 queries with proper joins

2. **Memory**
   - Reuse buffers where possible
   - Avoid unnecessary allocations in hot loops
   - Use sync.Pool for temporary objects

3. **Caching**
   - Cache leaderboard rankings
   - Cache user statistics (update periodically)
   - Cache reading level calculations

### Benchmark Targets

**Reading App:**
- ReadingSession: < 50µs
- UserStatistics: < 100µs
- Leaderboard: < 50µs

**Piano App:**
- PracticeLesson: < 50µs
- UserProgress: < 50µs
- UserMetrics: < 100µs

---

## Debugging

### Using Print Debugging

```go
// Add debug output
fmt.Printf("DEBUG: userId=%d, bookId=%d\n", userID, bookID)

// Use log package
import "log"
log.Printf("Creating session for user %d", userID)
```

### Using Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run tests with debugger
dlv test ./pkg/reading

# Set breakpoint
(dlv) break TestCreateAndRetrieveBook
(dlv) continue
(dlv) next
(dlv) print variable_name
```

### Debugging Tests

```go
// Add debugging to tests
func TestComplexLogic(t *testing.T) {
    t.Logf("Starting test with userId=%d", userID)

    result := performOperation()

    t.Logf("Operation result: %+v", result)

    if !validate(result) {
        t.Logf("Validation failed for: %#v", result)
        t.Fail()
    }
}

// Run with verbose output
go test -v -run TestComplexLogic ./pkg/reading
```

### Examining Database

```bash
# Connect to SQLite database
sqlite3 data/reading.db

# View table structure
.schema reading_sessions

# Query data
SELECT * FROM reading_sessions WHERE user_id = 1;

# Check indexes
.indices

# Explain query plan
EXPLAIN QUERY PLAN SELECT * FROM reading_sessions WHERE user_id = 1;
```

---

## Code Standards

### Naming Conventions

```go
// Functions/Methods - CamelCase, exported starts with capital
func CreateBook()          // Exported
func processSession()      // Unexported

// Constants - UPPER_SNAKE_CASE or CamelCase
const MaxWPM = 200
const DefaultLanguage = "English"

// Variables - camelCase
var userID uint
var isActive bool

// Interfaces - er suffix
type Reader interface {}
type Validator interface {}
```

### Code Organization

1. **Package structure**
   ```go
   package reading

   import (
       "context"
       "database/sql"
       "fmt"
       "errors"
   )

   // Constants
   const (
       DefaultLanguage = "English"
   )

   // Types
   type Book struct {}

   // Public functions
   func NewRepository() {}

   // Private functions
   func calculateMetrics() {}
   ```

2. **Function ordering**
   - Public functions first
   - Grouped by functionality
   - Receiver methods after standalone functions

3. **Comment style**
   ```go
   // CreateBook creates a new book entry
   func (r *Repository) CreateBook(ctx context.Context, book *Book) (uint, error) {
       // Implementation...
   }
   ```

### Error Handling

```go
// Good
result, err := doSomething()
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}

// Bad
result, err := doSomething()
if err != nil {
    panic(err)
}
```

### Testing Style

```go
// Good - clear, focused
func TestSessionCreation(t *testing.T) {
    // Setup
    repo := setupTestRepo(t)

    // Execute
    id, err := repo.SaveSession(ctx, session)

    // Verify
    if err != nil {
        t.Fatal("SaveSession failed")
    }
    if id == 0 {
        t.Error("Expected non-zero ID")
    }
}

// Bad - unclear, testing too much
func TestRepo(t *testing.T) {
    // ...complex setup...
    // ...multiple assertions...
    // ...unclear what's being tested...
}
```

---

## Common Tasks

### Running a Specific Test

```bash
# Single test
go test -v -run TestCreateAndRetrieveBook ./pkg/reading

# Multiple tests matching pattern
go test -v -run "^TestCreate" ./pkg/reading

# All integration tests
go test -v -run "^Test" ./pkg/reading
```

### Adding a Database Migration

1. Create migration file: `migrations/005_add_reading_goals.sql`
2. Add schema changes
3. Update `repository_test.go` setupTestDB function
4. Run: `sqlite3 data/reading.db < migrations/005_add_reading_goals.sql`

### Checking Code Quality

```bash
# Format code
go fmt ./pkg/reading
go fmt ./pkg/piano

# Lint code (requires golangci-lint)
golangci-lint run ./pkg/reading ./pkg/piano

# Check for race conditions
go test -race ./pkg/reading ./pkg/piano

# Check test coverage
go test -coverprofile=coverage.out ./pkg/reading
go tool cover -html=coverage.out
```

### Updating Dependencies

```bash
# Check for updates
go list -u -m all

# Update specific dependency
go get -u github.com/user/package

# Update all dependencies
go get -u ./...

# Verify dependencies
go mod verify

# Clean up unused dependencies
go mod tidy
```

---

## Troubleshooting

### Common Issues & Solutions

**Issue: Tests fail with "database is locked"**
```
Solution:
- Ensure no other processes accessing test database
- Use in-memory databases (:memory:) for tests
- Check for unclosed database connections
```

**Issue: "nil pointer dereference" panic**
```
Solution:
- Check nil before using pointers
- Use guard clauses at function start
- Run with -race flag to detect issues early
```

**Issue: MIDI file handling errors (Piano)**
```
Solution:
- Verify MIDI header: 4D 54 68 64 (MThd)
- Check for proper blob encoding
- Ensure binary data not truncated
```

**Issue: Slow query performance**
```
Solution:
- Check for missing indexes
- Use EXPLAIN QUERY PLAN
- Profile with pprof
- Consider query optimization
```

**Issue: Tests pass locally but fail in CI**
```
Solution:
- Check for race conditions (use -race)
- Ensure no time dependencies
- Check for environment-specific issues
- Verify all mocks/stubs working
```

### Getting Help

1. **Check existing tests** - Look for similar test patterns
2. **Read error messages carefully** - They usually indicate the problem
3. **Use debugging tools** - dlv, pprof, sqlite3 CLI
4. **Check git history** - How was this implemented before?
5. **Ask on team channels** - Get help from other developers

---

## Checklist for New Features

- [ ] Feature designed and discussed
- [ ] Models defined with validation
- [ ] Repository methods implemented
- [ ] Service layer business logic added
- [ ] Router handler added with error handling
- [ ] Database schema updated (if needed)
- [ ] Unit tests written and passing
- [ ] Integration tests written and passing
- [ ] Benchmarks added (if performance critical)
- [ ] API documentation updated
- [ ] Code reviewed and approved
- [ ] Tests pass with -race flag
- [ ] Coverage meets 80%+ threshold
- [ ] Commit message is clear and descriptive

---

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Chi Router](https://github.com/go-chi/chi)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [Effective Go](https://golang.org/doc/effective_go)
- [Testing in Go](https://golang.org/doc/tutorial/add-a-test)

