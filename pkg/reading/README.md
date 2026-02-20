# Reading App - Package Documentation

A comprehensive reading practice application helping users improve reading speed, accuracy, and comprehension.

## Architecture Overview

The reading app follows a clean layered architecture:

```
HTTP Handlers (router.go) → Service Layer (service.go) → Repository (repository.go) → Database
```

### Key Components

- **Models** (models.go): Book, ReadingSession, ReadingStats, ComprehensionTest
- **Repository** (repository.go): CRUD operations for books, sessions, comprehension tests
- **Service** (service.go): Business logic - WPM calculation, accuracy, metrics
- **Router** (router.go): HTTP endpoints and request handling
- **Handlers** (handler.go): Individual endpoint handlers with response formatting

## Core Features

### 1. Reading Metrics

**WPM Calculation**: `(characters / 5) / minutes`
- 250 characters ÷ 5 ÷ 2 minutes = 25 WPM
- Assumes average word = 5 characters

**Accuracy**: `((total_chars - errors) / total_chars) × 100`
- Tracks character-level accuracy
- Range: 0-100%

**Comprehension**: `(correct_answers / total_questions) × 100`
- Based on assessment responses
- Used for adaptive difficulty

### 2. User Progress Tracking

- Session history with timestamps
- Average and best WPM trends
- Accuracy trends over time
- Reading level estimation
- Personalized recommendations

### 3. Leaderboard System

- Global rankings by WPM
- User statistics aggregation
- Performance comparisons
- Motivation through competition

## Database Schema

**books**: Reading passages (id, title, author, content, reading_level, language, word_count)

**reading_sessions**: Practice attempts (id, user_id, book_id, wpm, accuracy, duration)

**comprehension_tests**: Assessment data (id, session_id, question, user_answer, correct_answer, is_correct)

## API Endpoints

- POST /api/sessions - Create reading session
- GET /api/sessions/{id} - Get session details
- GET /api/users/{userId}/stats - User statistics
- GET /api/users/{userId}/progress - Progress overview
- GET /api/books - List passages
- POST /api/books - Create passage
- GET /api/leaderboard - Rankings

## Testing

Comprehensive test coverage includes:
- 82+ unit and integration tests
- Service calculation validation
- Repository CRUD operations
- Handler endpoint testing
- End-to-end workflows

Run: `go test ./pkg/reading -v`

## Performance Targets

- Save session: < 20ms
- Get stats: < 10ms
- List books: < 15ms
- Leaderboard: < 50ms

## Migration from Python

Migrated from Python/Flask to Go for:
- Better performance (5-10x faster)
- Improved concurrency handling
- Type safety and compile-time checks
- Easier deployment and scaling
