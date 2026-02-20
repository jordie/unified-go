# Reading App API Reference

Complete documentation of all API endpoints for the reading application.

## Base URL

```
http://localhost:5051/api
```

## Authentication

Currently, the API uses simple user_id parameter for multi-user support. Future versions will implement token-based authentication.

## Response Format

All endpoints return JSON responses with consistent structure:

### Success Response
```json
{
    "success": true,
    "data": { /* response data */ },
    "error": null
}
```

### Error Response
```json
{
    "success": false,
    "data": null,
    "error": "Error message describing what went wrong"
}
```

## HTTP Status Codes

- **200 OK**: Successful GET request
- **201 Created**: Successful POST request (resource created)
- **400 Bad Request**: Invalid parameters or request body
- **404 Not Found**: Resource not found
- **500 Internal Server Error**: Server error

## Endpoints

### Sessions

#### POST /api/sessions
Create a new reading session.

**Request Body**
```json
{
    "user_id": 1,
    "book_id": 1,
    "content": "The full text that was read",
    "time_spent": 120.0,
    "errors": 2
}
```

**Parameters**
- `user_id` (integer, required): User ID
- `book_id` (integer, required): Book ID
- `content` (string, required): The content that was read
- `time_spent` (float, required): Time spent reading in seconds
- `errors` (integer, required): Number of errors made

**Response**
```json
{
    "id": 1,
    "user_id": 1,
    "book_id": 1,
    "wpm": 25.5,
    "accuracy": 95.2,
    "comprehension_score": 0,
    "duration": 120.0,
    "created_at": "2024-01-31T10:30:00Z"
}
```

**Status**: 201 Created

---

#### GET /api/sessions/{id}
Retrieve a specific reading session.

**Path Parameters**
- `id` (integer): Session ID

**Response**
```json
{
    "id": 1,
    "user_id": 1,
    "book_id": 1,
    "wpm": 25.5,
    "accuracy": 95.2,
    "comprehension_score": 0,
    "duration": 120.0,
    "created_at": "2024-01-31T10:30:00Z"
}
```

**Status**: 200 OK

---

#### GET /api/users/{userId}/sessions
List all sessions for a specific user.

**Path Parameters**
- `userId` (integer): User ID

**Query Parameters**
- `limit` (integer): Max sessions to return (default: 20)
- `offset` (integer): Pagination offset (default: 0)

**Response**
```json
{
    "sessions": [
        {
            "id": 1,
            "user_id": 1,
            "book_id": 1,
            "wpm": 25.5,
            "accuracy": 95.2,
            "duration": 120.0,
            "created_at": "2024-01-31T10:30:00Z"
        }
    ],
    "total": 10,
    "limit": 20,
    "offset": 0
}
```

**Status**: 200 OK

---

### User Statistics

#### GET /api/users/{userId}/stats
Get aggregated statistics for a user.

**Path Parameters**
- `userId` (integer): User ID

**Response**
```json
{
    "user_id": 1,
    "total_sessions_count": 15,
    "average_wpm": 45.2,
    "best_wpm": 62.5,
    "average_accuracy": 92.3,
    "total_reading_time_minutes": 180
}
```

**Status**: 200 OK

---

#### GET /api/users/{userId}/progress
Get user's progress overview.

**Path Parameters**
- `userId` (integer): User ID

**Response**
```json
{
    "estimated_level": "intermediate",
    "total_sessions": 15,
    "total_reading_time": 10800,
    "trend": {
        "direction": "improving",
        "change": 5.2
    },
    "recent_wpm": 48.5
}
```

**Status**: 200 OK

---

### Books

#### GET /api/books
List available reading passages.

**Query Parameters**
- `difficulty` (string): Filter by level (beginner|intermediate|advanced)
- `limit` (integer): Max books to return (default: 20)
- `offset` (integer): Pagination offset (default: 0)

**Response**
```json
{
    "books": [
        {
            "id": 1,
            "title": "Sample Book",
            "author": "Author Name",
            "content": "Book content here...",
            "reading_level": "beginner",
            "language": "English",
            "word_count": 500,
            "created_at": "2024-01-31T10:30:00Z"
        }
    ],
    "limit": 20,
    "offset": 0
}
```

**Status**: 200 OK

**Examples**
- `GET /api/books` - List all books
- `GET /api/books?difficulty=beginner` - Beginner books only
- `GET /api/books?limit=5&offset=0` - First 5 books

---

#### POST /api/books
Create a new reading passage.

**Request Body**
```json
{
    "title": "New Book",
    "author": "Author Name",
    "content": "Full book content...",
    "reading_level": "intermediate",
    "language": "English"
}
```

**Parameters**
- `title` (string, required): Book title
- `author` (string, required): Author name
- `content` (string, required): Book content (min 50 characters)
- `reading_level` (string): Level (beginner|intermediate|advanced)
- `language` (string): Language (default: English)

**Response**
```json
{
    "id": 2,
    "title": "New Book",
    "author": "Author Name",
    "content": "Full book content...",
    "reading_level": "intermediate",
    "language": "English",
    "word_count": 800,
    "created_at": "2024-01-31T10:30:00Z"
}
```

**Status**: 201 Created

---

#### GET /api/books/{id}
Retrieve a specific book.

**Path Parameters**
- `id` (integer): Book ID

**Response**
```json
{
    "id": 1,
    "title": "Sample Book",
    "author": "Author Name",
    "content": "Full book content...",
    "reading_level": "beginner",
    "language": "English",
    "word_count": 500,
    "created_at": "2024-01-31T10:30:00Z"
}
```

**Status**: 200 OK

---

### Comprehension

#### POST /api/comprehension
Submit a comprehension test answer.

**Request Body**
```json
{
    "session_id": 1,
    "question": "What was the main theme?",
    "user_answer": "The main theme was...",
    "correct_answer": "The main theme was..."
}
```

**Response**
```json
{
    "session_id": 1,
    "is_correct": true,
    "score": 100.0
}
```

**Status**: 201 Created

---

#### GET /api/sessions/{sessionId}/comprehension
Get comprehension questions for a session.

**Path Parameters**
- `sessionId` (integer): Session ID

**Response**
```json
{
    "session_id": 1,
    "questions": [
        {
            "id": 1,
            "question": "What was the main theme?",
            "correct_answer": "The answer..."
        }
    ]
}
```

**Status**: 200 OK

---

#### GET /api/sessions/{sessionId}/analysis
Analyze comprehension results.

**Path Parameters**
- `sessionId` (integer): Session ID

**Response**
```json
{
    "session_id": 1,
    "total_questions": 5,
    "correct_answers": 4,
    "score": 80.0,
    "feedback": "Good comprehension. Consider challenging yourself with harder texts."
}
```

**Status**: 200 OK

---

### Rankings

#### GET /api/leaderboard
Get top readers leaderboard.

**Query Parameters**
- `limit` (integer): Number of top readers (default: 10, max: 100)
- `metric` (string): Sort metric (wpm|accuracy|sessions, default: wpm)

**Response**
```json
{
    "leaderboard": [
        {
            "user_id": 1,
            "rank": 1,
            "average_wpm": 75.2,
            "best_wpm": 95.5,
            "sessions": 25,
            "accuracy": 94.2
        }
    ],
    "limit": 10
}
```

**Status**: 200 OK

**Examples**
- `GET /api/leaderboard` - Top 10 by WPM
- `GET /api/leaderboard?limit=5` - Top 5 readers
- `GET /api/leaderboard?metric=accuracy` - Top by accuracy

---

## Error Responses

### Common Error Codes

**400 Bad Request** - Invalid parameters
```json
{
    "success": false,
    "error": "book_id is required"
}
```

**404 Not Found** - Resource not found
```json
{
    "success": false,
    "error": "Session not found"
}
```

**500 Internal Server Error** - Server error
```json
{
    "success": false,
    "error": "Failed to save session"
}
```

---

## Rate Limiting

Currently no rate limiting is enforced. Future versions may implement:
- 100 requests per minute per user
- 1000 requests per minute per IP

---

## Pagination

Endpoints that return lists support pagination:

**Query Parameters**
- `limit`: Number of items per page (default: 20, max: 100)
- `offset`: Starting position (default: 0)

**Example**
```
GET /api/books?limit=10&offset=20
```

---

## Filtering

Some endpoints support filtering:

**Books Filtering**
- `difficulty`: beginner|intermediate|advanced
- `language`: English, Spanish, etc.

**Examples**
```
GET /api/books?difficulty=intermediate
GET /api/books?difficulty=advanced&limit=5
```

---

## Data Types

- `integer`: Whole numbers (1, 100, etc.)
- `float`: Decimal numbers (25.5, 95.2, etc.)
- `string`: Text values
- `boolean`: true|false
- `timestamp`: ISO 8601 format (2024-01-31T10:30:00Z)

---

## Best Practices

1. **Always validate input**: Check required fields before sending
2. **Handle errors gracefully**: Check success flag in response
3. **Use pagination**: Don't request all records at once
4. **Cache results**: Cache leaderboard and book list responses
5. **Rate limiting ready**: Prepare for future rate limiting

---

## Examples

### Complete Reading Session Flow

```bash
# 1. Get available beginner books
curl -X GET "http://localhost:5051/api/books?difficulty=beginner"

# 2. Select a book (e.g., book_id=1)

# 3. Create a reading session
curl -X POST "http://localhost:5051/api/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "book_id": 1,
    "content": "The text that was read",
    "time_spent": 120,
    "errors": 2
  }'

# 4. Get user statistics
curl -X GET "http://localhost:5051/api/users/1/stats"

# 5. View leaderboard
curl -X GET "http://localhost:5051/api/leaderboard"
```
