# Phase 4: Reading Application

A comprehensive reading comprehension platform that helps users improve their reading speed, accuracy, and comprehension through structured practice sessions.

## Overview

The Reading app provides an interactive environment for users to:
- **Read Books**: Access a curated library of books at various difficulty levels
- **Practice Reading**: Time-based reading sessions with real-time metrics
- **Track Progress**: Monitor WPM (words per minute), accuracy, and comprehension scores
- **Compete**: View leaderboards ranked by different metrics
- **Test Comprehension**: Answer questions after reading to verify understanding

## Architecture

```
pkg/reading/
├── models.go              # Data models (Book, Session, etc.)
├── service.go             # Business logic layer
├── repository.go          # Data persistence layer
├── router.go              # HTTP route handlers
├── handler.go             # Response formatting helpers
├── models_test.go         # Unit tests for models
├── service_test.go        # Unit tests for service
├── repository_test.go     # Unit tests for repository
├── integration_test.go    # Integration tests + benchmarks
├── templates/
│   ├── base.html          # Shared layout
│   ├── books.html         # Book listing
│   ├── read.html          # Reading interface
│   ├── dashboard.html     # User statistics
│   ├── leaderboard.html   # Competitive rankings
│   └── comprehension.html # Quiz interface
└── README.md              # This file
```

## Data Models

### Book
Represents a book in the library.

```go
type Book struct {
    ID               uint      // Unique identifier
    Title            string    // Book title
    Author           string    // Book author
    Content          string    // Full book text
    ReadingLevel     string    // "beginner", "intermediate", "advanced"
    Language         string    // Language code (e.g., "English")
    WordCount        int       // Total words in content
    EstimatedTimeMin float64   // Estimated reading time (minutes)
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### ReadingSession
Captures a single reading practice session.

```go
type ReadingSession struct {
    ID              uint      // Unique identifier
    UserID          uint      // User performing the reading
    BookID          uint      // Book being read
    StartTime       time.Time // When reading started
    EndTime         time.Time // When reading finished
    Duration        float64   // Time spent (seconds)
    WPM             float64   // Words per minute calculated
    Accuracy        float64   // Accuracy percentage (0-100)
    ComprehScore    float64   // Comprehension score (0-100)
    ErrorsDetected  int       // Number of reading errors
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### ComprehensionTest
Validates reading understanding.

```go
type ComprehensionTest struct {
    ID            uint      // Unique identifier
    SessionID     uint      // Associated reading session
    Question      string    // Test question
    CorrectAnswer string    // Correct answer
    UserAnswer    string    // User's answer
    Score         float64   // Score for this question (0-100)
    Difficulty    string    // Question difficulty level
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### UserProgress
Aggregated metrics for a user.

```go
type UserProgress struct {
    UserID              uint       // User identifier
    TotalSessions       int        // Total reading sessions completed
    TotalTimeSpent      float64    // Total time spent reading (minutes)
    AverageWPM          float64    // Average words per minute
    BestWPM             float64    // Highest WPM achieved
    AverageAccuracy     float64    // Average accuracy (0-100)
    AverageComprehScore float64    // Average comprehension score (0-100)
    ReadingLevel        string     // Estimated reading level
    LastSessionDate     *time.Time // When user last read
}
```

## API Endpoints

### Books
- `GET /api/books` - List all books (queryable by difficulty, language)
- `POST /api/books` - Create a new book
- `GET /api/books/{bookId}` - Get a specific book
- `PUT /api/books/{bookId}` - Update book metadata
- `DELETE /api/books/{bookId}` - Archive a book

### Reading Sessions
- `POST /api/sessions` - Create a new reading session
  ```json
  {
    "user_id": 1,
    "book_id": 5,
    "time_spent": 300.0,
    "errors": 2
  }
  ```
- `GET /api/sessions/{sessionId}` - Get session details
- `GET /api/users/{userId}/sessions` - Get user's reading history
- `PUT /api/sessions/{sessionId}` - Update session (e.g., mark complete)

### Comprehension Tests
- `POST /api/comprehension` - Create comprehension test
  ```json
  {
    "session_id": 123,
    "question": "What was the main theme?",
    "correct_answer": "Human resilience",
    "user_answer": "Overcoming adversity"
  }
  ```
- `GET /api/comprehension/{testId}` - Get test details
- `GET /api/sessions/{sessionId}/comprehension` - Get all tests for a session

### User Statistics
- `GET /api/users/{userId}/stats` - Get user progress statistics
- `GET /api/users/{userId}/progress` - Get detailed progress tracking
- `GET /api/leaderboard` - Get leaderboard rankings (queryable by metric)
  - Query params: `metric=wpm|accuracy|sessions|comprehension`
  - `?limit=10&offset=0` - Pagination

### Content Validation
- `POST /api/validate-content` - Validate book content
  ```json
  {
    "content": "Lorem ipsum...",
    "language": "English"
  }
  ```
  Response includes word count, validation status, and recommendations

## Database Schema

### books
```sql
CREATE TABLE books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT,
    content TEXT,
    reading_level TEXT,
    language TEXT DEFAULT 'english',
    word_count INTEGER,
    estimated_time_minutes REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### reading_sessions
```sql
CREATE TABLE reading_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    book_id INTEGER NOT NULL,
    start_time DATETIME,
    end_time DATETIME,
    wpm REAL,
    accuracy REAL,
    comprehension_score REAL,
    errors_detected INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### comprehension_tests
```sql
CREATE TABLE comprehension_tests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    question TEXT NOT NULL,
    correct_answer TEXT,
    user_answer TEXT,
    score REAL,
    difficulty TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user ON reading_sessions(user_id);
CREATE INDEX idx_comprehension_session ON comprehension_tests(session_id);
```

## Running the Application

### Prerequisites
- Go 1.21+
- SQLite3

### Setup
```bash
cd pkg/reading

# Install dependencies
go mod download

# Initialize database (automatic on first run)
# Tests create in-memory databases automatically
```

### Running Tests
```bash
# All tests
go test -v ./...

# Unit tests only
go test -v -run ^Test[A-Z] ./...

# Integration tests only
go test -v -run ^Test[A-Z] -integration ./...

# Benchmarks
go test -bench=Benchmark -run=^$ ./...

# Specific test
go test -v -run TestCreateAndRetrieveBook ./...
```

### Example Usage
```go
package main

import (
    "context"
    "database/sql"
    "github.com/jgirmay/unified-go/pkg/reading"
)

func main() {
    // Open database
    db, _ := sql.Open("sqlite3", "reading.db")
    defer db.Close()

    // Create repository and service
    repo := reading.NewRepository(db)
    service := reading.NewService(repo)

    // Create a book
    ctx := context.Background()
    book := &reading.Book{
        Title:        "The Great Gatsby",
        Author:       "F. Scott Fitzgerald",
        Content:      "...",
        ReadingLevel: "intermediate",
        Language:     "English",
        WordCount:    47000,
    }

    bookID, _ := service.repo.SaveBook(ctx, book)

    // Record a reading session
    session := &reading.ReadingSession{
        UserID:      1,
        BookID:      bookID,
        Duration:    600.0,  // 10 minutes
        ErrorsDetected: 3,
    }

    sessionID, _ := repo.SaveSession(ctx, session)

    // Calculate metrics
    wpm := reading.CalculateWPM(book.WordCount, session.Duration)
    accuracy := reading.CalculateAccuracy(book.Content, userInput)

    // Get user progress
    progress, _ := repo.GetUserProgress(ctx, 1)
    println("Average WPM:", progress.AverageWPM)
}
```

## Performance Metrics

### Benchmark Results (Apple M2)
- **ReadingSession**: ~18µs per operation
- **UserStatistics**: ~48µs per operation
- **Leaderboard Query**: ~9µs per operation
- **Content Validation**: ~8.7µs per operation

### Test Coverage
- 9 integration tests covering complete workflows
- 3 performance benchmarks
- 100+ total tests across unit, repository, and integration suites

## Features

### Reading Practice
- ✅ Timed reading sessions with auto-calculation
- ✅ WPM (words per minute) tracking
- ✅ Accuracy percentage calculation
- ✅ Difficulty levels (beginner/intermediate/advanced)
- ✅ Multiple language support

### Progress Tracking
- ✅ Personal statistics dashboard
- ✅ Historical session data
- ✅ Progress trends and metrics
- ✅ Level estimation based on performance

### Gamification
- ✅ Leaderboards (sortable by WPM, accuracy, sessions, comprehension)
- ✅ Personal best tracking
- ✅ Session completion streaks
- ✅ Difficulty-based challenges

### Comprehension
- ✅ Post-reading quizzes
- ✅ Multiple-choice questions
- ✅ Score calculation
- ✅ Comprehension level tracking

## Development

### Adding a New Feature

1. **Add Model** (models.go)
   ```go
   type MyNewType struct {
       ID   uint
       Name string
       // ... other fields
   }
   ```

2. **Add Repository Method** (repository.go)
   ```go
   func (r *Repository) SaveMyType(ctx context.Context, obj *MyNewType) (uint, error) {
       // Implementation
   }
   ```

3. **Add Service Method** (service.go)
   ```go
   func (s *Service) ProcessMyType(ctx context.Context, data *MyNewType) error {
       // Business logic
   }
   ```

4. **Add API Handler** (router.go)
   ```go
   func (r *Router) HandleMyType(w http.ResponseWriter, req *http.Request) {
       // HTTP handling
   }
   ```

5. **Add Tests** (integration_test.go)
   ```go
   func TestMyFeature(t *testing.T) {
       // Test implementation
   }
   ```

### Code Structure

- **Models Layer**: Data structures, validation, calculations
- **Repository Layer**: Database operations, SQL queries
- **Service Layer**: Business logic, aggregations, complex operations
- **Router Layer**: HTTP routing, request parsing, response formatting

## Troubleshooting

### Tests Failing

**Issue**: TestGetUserStatistics fails with "Expected 3 sessions, got 0"
- **Solution**: Ensure sessions are created before querying statistics

**Issue**: Database locked
- **Solution**: Check for concurrent access; tests use in-memory databases by default

### Performance Issues

**Issue**: Leaderboard queries are slow
- **Solution**: Add database indexes on frequently queried columns (user_id, created_at)

**Issue**: Content validation taking too long
- **Solution**: Consider caching word count calculations for large documents

## Contributing

When contributing to the Reading app:

1. Maintain test-first approach (write tests before implementation)
2. Keep integration tests synchronized with API changes
3. Run benchmarks before/after optimization
4. Update this README for new features
5. Ensure all tests pass: `go test -v ./...`

## Future Enhancements

- [ ] Text-to-speech integration
- [ ] Reading difficulty auto-adjustment
- [ ] ML-based comprehension question generation
- [ ] Real-time collaborative reading
- [ ] Mobile app integration
- [ ] Audio book support
- [ ] Reading speed records and achievements

## License

Part of the GAIA distributed development system for basic educational applications.
