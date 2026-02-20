# Complete API Endpoints Reference

All 33+ endpoints for the unified-go educational platform with request/response examples.

## Table of Contents

1. [Health & Status](#health--status)
2. [Reading App](#reading-app-9-endpoints)
3. [Typing App](#typing-app-9-endpoints)
4. [Math App](#math-app-7-endpoints)
5. [Piano App](#piano-app-pending)

---

## Health & Status

### GET /health

Server health check endpoint.

**Request:**
```bash
curl http://localhost:8080/health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "go_version": "go1.21.0",
  "uptime": "2h30m15s",
  "timestamp": "2025-02-20T15:30:45Z",
  "goroutines": 42,
  "environment": "development"
}
```

---

## Reading App (9 endpoints)

### GET /reading/

Reading app homepage with interactive interface.

**Request:**
```bash
curl http://localhost:8080/reading/
```

**Response:** HTML page with reading dashboard

---

### GET /api/reading/passages

List available reading passages with filters.

**Request:**
```bash
curl "http://localhost:8080/api/reading/passages?difficulty=medium&limit=10&offset=0"
```

**Query Parameters:**
- `difficulty` (optional): easy, medium, hard
- `limit` (default: 20): Number of results
- `offset` (default: 0): Pagination offset

**Response (200 OK):**
```json
{
  "passages": [
    {
      "id": 1,
      "title": "The Renaissance",
      "difficulty": "medium",
      "content": "The Renaissance was a period of European cultural...",
      "word_count": 450,
      "estimated_reading_time_seconds": 180,
      "created_at": "2025-02-01T10:00:00Z"
    }
  ],
  "total": 42,
  "limit": 10,
  "offset": 0
}
```

---

### GET /api/reading/passages/{id}

Get a specific passage with questions.

**Request:**
```bash
curl http://localhost:8080/api/reading/passages/1
```

**Response (200 OK):**
```json
{
  "id": 1,
  "title": "The Renaissance",
  "difficulty": "medium",
  "content": "The Renaissance was a period...",
  "word_count": 450,
  "estimated_reading_time_seconds": 180,
  "questions": [
    {
      "id": 101,
      "text": "What was the main focus of Renaissance art?",
      "options": [
        "Religious themes",
        "Political propaganda",
        "Humanist ideals",
        "Military strategy"
      ],
      "type": "multiple_choice"
    }
  ],
  "created_at": "2025-02-01T10:00:00Z"
}
```

---

### POST /api/reading/answer

Submit reading comprehension answers.

**Request:**
```bash
curl -X POST http://localhost:8080/api/reading/answer \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "passage_id": 1,
    "question_id": 101,
    "answer_selected": "Humanist ideals"
  }'
```

**Response (201 Created):**
```json
{
  "result_id": 501,
  "passage_id": 1,
  "question_id": 101,
  "user_id": 1,
  "correct": true,
  "answer_provided": "Humanist ideals",
  "correct_answer": "Humanist ideals",
  "timestamp": "2025-02-20T15:30:45Z"
}
```

---

### GET /api/users/{userId}/reading/stats

Get user reading statistics.

**Request:**
```bash
curl -H "X-User-ID: 1" http://localhost:8080/api/users/1/reading/stats
```

**Response (200 OK):**
```json
{
  "user_id": 1,
  "passages_completed": 12,
  "questions_answered": 48,
  "questions_correct": 44,
  "accuracy": 91.67,
  "average_reading_speed": 245.5,
  "best_accuracy": 100,
  "total_time_spent_minutes": 180,
  "difficulty_progress": {
    "easy": {"completed": 4, "accuracy": 98.5},
    "medium": {"completed": 5, "accuracy": 91.2},
    "hard": {"completed": 3, "accuracy": 85.0}
  },
  "last_session": "2025-02-20T14:00:00Z"
}
```

---

### GET /api/reading/leaderboard

Reading accuracy leaderboard.

**Request:**
```bash
curl "http://localhost:8080/api/reading/leaderboard?limit=10"
```

**Query Parameters:**
- `limit` (default: 100, max: 1000): Number of results

**Response (200 OK):**
```json
{
  "leaderboard": [
    {
      "rank": 1,
      "user_id": 5,
      "username": "alice",
      "accuracy": 94.5,
      "passages_completed": 18,
      "total_time_minutes": 240
    },
    {
      "rank": 2,
      "user_id": 3,
      "username": "bob",
      "accuracy": 92.3,
      "passages_completed": 15,
      "total_time_minutes": 210
    }
  ],
  "limit": 10
}
```

---

### GET /api/users/{userId}/reading/sessions

Get user reading session history.

**Request:**
```bash
curl -H "X-User-ID: 1" "http://localhost:8080/api/users/1/reading/sessions?limit=20&offset=0"
```

**Query Parameters:**
- `limit` (default: 20): Results per page
- `offset` (default: 0): Pagination offset

**Response (200 OK):**
```json
{
  "sessions": [
    {
      "id": 201,
      "user_id": 1,
      "passage_id": 1,
      "started_at": "2025-02-20T14:00:00Z",
      "completed_at": "2025-02-20T14:05:30Z",
      "duration_seconds": 330,
      "questions_answered": 5,
      "questions_correct": 5,
      "accuracy": 100,
      "reading_speed_wpm": 260
    }
  ],
  "total": 12,
  "limit": 20,
  "offset": 0
}
```

---

### GET /reading/dashboard

Reading dashboard UI page.

**Request:**
```bash
curl http://localhost:8080/reading/dashboard
```

**Response:** HTML dashboard with stats and progress charts

---

## Typing App (9 endpoints)

### GET /typing/

Typing app homepage.

**Request:**
```bash
curl http://localhost:8080/typing/
```

**Response:** HTML page with typing practice interface

---

### POST /api/typing/test

Create and submit a typing test.

**Request:**
```bash
curl -X POST http://localhost:8080/api/typing/test \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "text": "The quick brown fox jumps over the lazy dog",
    "duration_seconds": 60,
    "user_input": "The quick brown fox jumps over the lazy dog",
    "start_time": "2025-02-20T15:20:00Z",
    "end_time": "2025-02-20T15:21:00Z"
  }'
```

**Response (201 Created):**
```json
{
  "test_id": 101,
  "user_id": 1,
  "words_per_minute": 85.5,
  "accuracy": 98.5,
  "characters_typed": 44,
  "errors": 1,
  "duration_seconds": 60,
  "timestamp": "2025-02-20T15:21:00Z"
}
```

---

### GET /api/typing/test/{id}

Get specific typing test details.

**Request:**
```bash
curl http://localhost:8080/api/typing/test/101
```

**Response (200 OK):**
```json
{
  "id": 101,
  "user_id": 1,
  "text": "The quick brown fox jumps over the lazy dog",
  "duration_seconds": 60,
  "words_per_minute": 85.5,
  "accuracy": 98.5,
  "characters_typed": 44,
  "errors": 1,
  "error_details": [{"position": 23, "expected": "r", "typed": "t"}],
  "timestamp": "2025-02-20T15:21:00Z"
}
```

---

### GET /api/users/{userId}/typing/tests

Get user typing test history.

**Request:**
```bash
curl -H "X-User-ID: 1" "http://localhost:8080/api/users/1/typing/tests?limit=20&offset=0"
```

**Query Parameters:**
- `limit` (default: 20): Results per page
- `offset` (default: 0): Pagination offset

**Response (200 OK):**
```json
{
  "tests": [
    {
      "id": 101,
      "duration_seconds": 60,
      "words_per_minute": 85.5,
      "accuracy": 98.5,
      "timestamp": "2025-02-20T15:21:00Z"
    },
    {
      "id": 100,
      "duration_seconds": 60,
      "words_per_minute": 82.3,
      "accuracy": 96.2,
      "timestamp": "2025-02-20T14:15:00Z"
    }
  ],
  "total": 24,
  "limit": 20,
  "offset": 0
}
```

---

### GET /api/users/{userId}/typing/stats

Get user typing statistics.

**Request:**
```bash
curl -H "X-User-ID: 1" http://localhost:8080/api/users/1/typing/stats
```

**Response (200 OK):**
```json
{
  "user_id": 1,
  "total_tests": 24,
  "average_wpm": 78.5,
  "best_wpm": 92.3,
  "average_accuracy": 94.2,
  "best_accuracy": 99.5,
  "total_characters_typed": 15840,
  "total_time_minutes": 132,
  "recent_improvement": 8.5,
  "last_test": "2025-02-20T15:21:00Z"
}
```

---

### GET /api/typing/leaderboard

Typing speed leaderboard.

**Request:**
```bash
curl "http://localhost:8080/api/typing/leaderboard?limit=10"
```

**Query Parameters:**
- `limit` (default: 100, max: 1000): Number of results

**Response (200 OK):**
```json
{
  "leaderboard": [
    {
      "rank": 1,
      "user_id": 8,
      "username": "typer",
      "best_wpm": 125.5,
      "average_wpm": 98.3,
      "average_accuracy": 97.2,
      "tests_completed": 42
    },
    {
      "rank": 2,
      "user_id": 1,
      "username": "student",
      "best_wpm": 92.3,
      "average_wpm": 78.5,
      "average_accuracy": 94.2,
      "tests_completed": 24
    }
  ],
  "limit": 10
}
```

---

### GET /typing/dashboard

Typing dashboard UI.

**Request:**
```bash
curl http://localhost:8080/typing/dashboard
```

**Response:** HTML dashboard with progress charts

---

### GET /api/typing/recommended-text

Get recommended typing practice text.

**Request:**
```bash
curl "http://localhost:8080/api/typing/recommended-text?difficulty=medium"
```

**Query Parameters:**
- `difficulty` (default: medium): easy, medium, hard

**Response (200 OK):**
```json
{
  "text": "The Renaissance was a period of European cultural...",
  "word_count": 85,
  "estimated_duration_seconds": 45,
  "difficulty": "medium"
}
```

---

## Math App (7 endpoints)

### GET /math/

Math app homepage.

**Request:**
```bash
curl http://localhost:8080/math/
```

**Response:** HTML page with math problem selector

---

### POST /api/math/problem

Generate a new math problem.

**Request:**
```bash
curl -X POST http://localhost:8080/api/math/problem \
  -H "Content-Type: application/json" \
  -d '{
    "problem_type": "addition",
    "difficulty": "easy"
  }'
```

**Query Parameters:**
- `problem_type`: addition, subtraction, multiplication, division, fractions, algebra
- `difficulty`: easy, medium, hard, very_hard

**Response (200 OK):**
```json
{
  "question": "15 + 23 = ?",
  "answer": 38
}
```

---

### POST /api/math/session/complete

Complete a math practice session.

**Request:**
```bash
curl -X POST http://localhost:8080/api/math/session/complete \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "problem_type": "addition",
    "difficulty": "easy",
    "total_problems": 10,
    "correct_answers": 9,
    "time_spent": 120.5
  }'
```

**Response (201 Created):**
```json
{
  "session_id": 301,
  "score": 90.0,
  "accuracy": 90.0,
  "message": "Session completed successfully"
}
```

---

### GET /api/users/{userId}/math/stats

Get user math statistics.

**Request:**
```bash
curl -H "X-User-ID: 1" http://localhost:8080/api/users/1/math/stats
```

**Response (200 OK):**
```json
{
  "stats": {
    "user_id": 1,
    "total_problems": 487,
    "correct_answers": 412,
    "accuracy": 84.6,
    "average_time_per_problem": 12.3,
    "best_score": 100,
    "total_time_spent": 5987,
    "sessions_completed": 48
  },
  "math_level": "advanced",
  "next_recommendation": "Challenge yourself with hard problems"
}
```

---

### GET /api/math/leaderboard

Math accuracy leaderboard.

**Request:**
```bash
curl "http://localhost:8080/api/math/leaderboard?limit=10"
```

**Query Parameters:**
- `limit` (default: 100, max: 1000): Number of results

**Response (200 OK):**
```json
{
  "leaderboard": [
    {
      "rank": 1,
      "user_id": 12,
      "best_score": 100,
      "accuracy": 97.3,
      "sessions_completed": 56
    },
    {
      "rank": 2,
      "user_id": 1,
      "best_score": 100,
      "accuracy": 84.6,
      "sessions_completed": 48
    }
  ],
  "limit": 10
}
```

---

### GET /api/users/{userId}/math/sessions

Get user math session history.

**Request:**
```bash
curl -H "X-User-ID: 1" "http://localhost:8080/api/users/1/math/sessions?limit=20"
```

**Query Parameters:**
- `limit` (default: 20): Results per page
- `offset` (default: 0): Pagination offset

**Response (200 OK):**
```json
{
  "sessions": [
    {
      "id": 301,
      "user_id": 1,
      "problem_type": "addition",
      "difficulty": "easy",
      "total_problems": 10,
      "correct_answers": 9,
      "score": 90.0,
      "time_spent": 120.5,
      "timestamp": "2025-02-20T15:15:00Z"
    }
  ],
  "total": 48,
  "limit": 20,
  "offset": 0
}
```

---

### GET /math/dashboard

Math dashboard UI.

**Request:**
```bash
curl http://localhost:8080/math/dashboard
```

**Response:** HTML dashboard with stats and progress

---

## Piano App (pending)

Piano app endpoints coming in Phase 5.

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request",
  "message": "problem_type is required"
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "message": "User not found"
}
```

### 404 Not Found
```json
{
  "error": "Not found",
  "message": "Passage with id 999 not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "message": "Database connection failed"
}
```

---

## Authentication

For endpoints requiring user context, include:

```bash
curl -H "X-User-ID: 1" http://localhost:8080/api/users/1/typing/stats
```

Or via cookie (if session is configured):
```bash
curl -b "session=abc123" http://localhost:8080/api/users/1/typing/stats
```

---

## Rate Limiting

Not implemented in current version. Add with middleware if needed:

```go
r.Use(middleware.RateLimit(100 * time.Second, 1000))
```

---

## Pagination

All list endpoints support:
- `limit` (default: varies, typically 20-100)
- `offset` (default: 0)

Example:
```bash
curl "http://localhost:8080/api/reading/passages?limit=10&offset=20"
```

---

## Timestamps

All timestamps are ISO 8601 format:
```
2025-02-20T15:30:45Z
```

Parse with:
```go
t, _ := time.Parse(time.RFC3339, "2025-02-20T15:30:45Z")
```

---

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common API errors.
