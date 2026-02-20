# GAIA Educational Apps - API Documentation

Complete API reference for Phase 4 (Reading) and Phase 5 (Piano) educational applications.

## Base URLs

- **Reading App**: `http://localhost:8081/api`
- **Piano App**: `http://localhost:8082/api`

## Response Format

All endpoints return JSON responses with consistent formatting.

### Success Response
```json
{
  "status": "success",
  "data": { /* response data */ },
  "timestamp": "2026-02-20T10:00:00Z"
}
```

### Error Response
```json
{
  "status": "error",
  "error": "Error message",
  "code": "ERROR_CODE",
  "timestamp": "2026-02-20T10:00:00Z"
}
```

### HTTP Status Codes
- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

# Phase 4: Reading App API

## Books Endpoints

### List Books
```http
GET /books
```

**Query Parameters:**
- `difficulty` - Filter by difficulty (beginner, intermediate, advanced)
- `language` - Filter by language (English, Spanish, French, etc.)
- `limit` - Results per page (default: 20)
- `offset` - Pagination offset (default: 0)

**Example Request:**
```bash
curl "http://localhost:8081/api/books?difficulty=intermediate&language=English&limit=10"
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "books": [
      {
        "id": 1,
        "title": "The Great Gatsby",
        "author": "F. Scott Fitzgerald",
        "reading_level": "intermediate",
        "language": "English",
        "word_count": 47094,
        "estimated_time_minutes": 180,
        "created_at": "2026-02-20T10:00:00Z",
        "updated_at": "2026-02-20T10:00:00Z"
      }
    ],
    "total": 145,
    "page": 1,
    "limit": 10
  }
}
```

### Get Book Details
```http
GET /books/{bookId}
```

**Path Parameters:**
- `bookId` - Book ID (required)

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "title": "The Great Gatsby",
    "author": "F. Scott Fitzgerald",
    "content": "In my younger and more vulnerable years...",
    "reading_level": "intermediate",
    "language": "English",
    "word_count": 47094,
    "estimated_time_minutes": 180,
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

### Create Book
```http
POST /books
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "New Book",
  "author": "Author Name",
  "content": "Full book content...",
  "reading_level": "beginner",
  "language": "English",
  "word_count": 5000
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "id": 146,
    "title": "New Book",
    "author": "Author Name",
    "reading_level": "beginner",
    "language": "English",
    "word_count": 5000,
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

## Reading Sessions Endpoints

### Create Reading Session
```http
POST /sessions
Content-Type: application/json
```

**Request Body:**
```json
{
  "user_id": 1,
  "book_id": 5,
  "duration": 600.0,
  "errors": 3
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "id": 123,
    "user_id": 1,
    "book_id": 5,
    "duration": 600.0,
    "wpm": 78.5,
    "accuracy": 98.5,
    "comprehension_score": 85.0,
    "errors_detected": 3,
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

### Get Session Details
```http
GET /sessions/{sessionId}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 123,
    "user_id": 1,
    "book_id": 5,
    "start_time": "2026-02-20T09:50:00Z",
    "end_time": "2026-02-20T10:00:00Z",
    "duration": 600.0,
    "wpm": 78.5,
    "accuracy": 98.5,
    "comprehension_score": 85.0,
    "errors_detected": 3
  }
}
```

### Get User Sessions
```http
GET /users/{userId}/sessions
```

**Query Parameters:**
- `limit` - Results per page (default: 20)
- `offset` - Pagination offset (default: 0)

**Response:**
```json
{
  "status": "success",
  "data": {
    "sessions": [
      { /* session objects */ }
    ],
    "total": 42,
    "page": 1,
    "limit": 20
  }
}
```

## User Statistics Endpoints

### Get User Progress
```http
GET /users/{userId}/progress
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "total_sessions": 42,
    "total_time_spent": 420.5,
    "average_wpm": 78.2,
    "best_wpm": 95.3,
    "average_accuracy": 96.8,
    "average_comprehension_score": 82.5,
    "reading_level": "advanced",
    "last_session_date": "2026-02-20T10:00:00Z"
  }
}
```

### Get User Statistics
```http
GET /users/{userId}/stats
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "total_sessions": 42,
    "total_time_spent": 420.5,
    "average_wpm": 78.2,
    "best_wpm": 95.3,
    "average_accuracy": 96.8,
    "percentage_improvement": 15.5,
    "most_read_difficulty": "intermediate"
  }
}
```

### Get Leaderboard
```http
GET /leaderboard
```

**Query Parameters:**
- `metric` - Sort by: `wpm`, `accuracy`, `sessions`, `comprehension` (default: wpm)
- `limit` - Results per page (default: 20)
- `offset` - Pagination offset (default: 0)

**Example Request:**
```bash
curl "http://localhost:8081/api/leaderboard?metric=accuracy&limit=10"
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "metric": "accuracy",
    "entries": [
      {
        "rank": 1,
        "user_id": 5,
        "username": "reader_pro",
        "total_sessions": 120,
        "avg_wpm": 92.3,
        "best_wpm": 110.5,
        "accuracy": 99.2,
        "comprehension_score": 90.5
      },
      {
        "rank": 2,
        "user_id": 3,
        "username": "speed_reader",
        "total_sessions": 85,
        "avg_wpm": 88.5,
        "best_wpm": 105.2,
        "accuracy": 98.8,
        "comprehension_score": 87.2
      }
    ],
    "total": 1250
  }
}
```

## Comprehension Endpoints

### Create Comprehension Test
```http
POST /comprehension
Content-Type: application/json
```

**Request Body:**
```json
{
  "session_id": 123,
  "question": "What was the main theme of the book?",
  "correct_answer": "The American Dream",
  "user_answer": "Dreams and aspirations"
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "id": 456,
    "session_id": 123,
    "question": "What was the main theme of the book?",
    "correct_answer": "The American Dream",
    "user_answer": "Dreams and aspirations",
    "score": 85.0,
    "difficulty": "intermediate",
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

### Get Session Comprehension Tests
```http
GET /sessions/{sessionId}/comprehension
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "session_id": 123,
    "tests": [
      { /* comprehension test objects */ }
    ],
    "average_score": 82.5,
    "total_questions": 10
  }
}
```

## Content Validation

### Validate Content
```http
POST /validate-content
Content-Type: application/json
```

**Request Body:**
```json
{
  "content": "Lorem ipsum dolor sit amet...",
  "language": "English"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "valid": true,
    "word_count": 1250,
    "character_count": 7500,
    "reading_level": "intermediate",
    "estimated_time_minutes": 8,
    "warnings": []
  }
}
```

---

# Phase 5: Piano App API

## Songs Endpoints

### List Songs
```http
GET /songs
```

**Query Parameters:**
- `difficulty` - Filter by difficulty (beginner, intermediate, advanced, master)
- `composer` - Filter by composer name
- `key_signature` - Filter by key (C Major, D minor, etc.)
- `limit` - Results per page (default: 20)
- `offset` - Pagination offset (default: 0)

**Example Request:**
```bash
curl "http://localhost:8082/api/songs?difficulty=advanced&composer=Beethoven"
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "songs": [
      {
        "id": 1,
        "title": "Moonlight Sonata",
        "composer": "Beethoven",
        "difficulty": "advanced",
        "bpm": 60,
        "time_signature": "4/4",
        "key_signature": "C# minor",
        "total_notes": 1000,
        "duration": 600.0,
        "created_at": "2026-02-20T10:00:00Z"
      }
    ],
    "total": 87,
    "page": 1,
    "limit": 20
  }
}
```

### Get Song Details
```http
GET /songs/{songId}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "title": "Moonlight Sonata",
    "composer": "Beethoven",
    "description": "A beautiful nocturne composed in 1801...",
    "difficulty": "advanced",
    "bpm": 60,
    "time_signature": "4/4",
    "key_signature": "C# minor",
    "total_notes": 1000,
    "duration": 600.0,
    "midi_file": "base64_encoded_midi_data",
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

### Create Song
```http
POST /songs
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Fur Elise",
  "composer": "Beethoven",
  "difficulty": "intermediate",
  "bpm": 120,
  "time_signature": "2/4",
  "key_signature": "A minor",
  "total_notes": 500,
  "duration": 180.0,
  "description": "Famous piano piece",
  "midi_file": "base64_encoded_midi_data"
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "id": 88,
    "title": "Fur Elise",
    "composer": "Beethoven",
    "difficulty": "intermediate",
    "bpm": 120,
    "total_notes": 500,
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

## Practice Sessions Endpoints

### Record Practice Session
```http
POST /practice
Content-Type: application/json
```

**Request Body:**
```json
{
  "user_id": 1,
  "song_id": 5,
  "recorded_bpm": 118.5,
  "duration": 300.0,
  "notes_correct": 450,
  "notes_total": 500
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "id": 456,
    "user_id": 1,
    "song_id": 5,
    "duration": 300.0,
    "notes_hit": 450,
    "notes_total": 500,
    "tempo_average": 118.5,
    "accuracy": 90.0,
    "tempo_accuracy": 98.75,
    "score": 93.2,
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

### Get Session Details
```http
GET /practice/{sessionId}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 456,
    "user_id": 1,
    "song_id": 5,
    "duration": 300.0,
    "notes_hit": 450,
    "notes_total": 500,
    "tempo_average": 118.5,
    "accuracy": 90.0,
    "tempo_accuracy": 98.75,
    "score": 93.2
  }
}
```

### Get User Sessions
```http
GET /users/{userId}/sessions
```

**Query Parameters:**
- `limit` - Results per page (default: 20)
- `offset` - Pagination offset (default: 0)

**Response:**
```json
{
  "status": "success",
  "data": {
    "sessions": [
      { /* practice session objects */ }
    ],
    "total": 127,
    "page": 1,
    "limit": 20
  }
}
```

## MIDI Endpoints

### Download Song MIDI
```http
GET /midi/{songId}
```

**Response:** Binary MIDI file (application/octet-stream)

**Example Request:**
```bash
curl "http://localhost:8082/api/midi/1" --output "song.mid"
```

### Upload MIDI File
```http
POST /midi/upload
Content-Type: audio/midi
```

**Request Body:** Binary MIDI file data

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "uploaded": true,
    "filename": "recording.mid",
    "size_bytes": 2048,
    "valid_midi": true
  }
}
```

### Analyze MIDI
```http
POST /midi/analyze
Content-Type: application/json
```

**Request Body:**
```json
{
  "midi_data": "4d546864000000060000000100600000000000"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "valid": true,
    "format": 0,
    "tracks": 1,
    "ticks_per_quarter_note": 96,
    "duration_seconds": 240.0,
    "notes_detected": 450
  }
}
```

## User Metrics Endpoints

### Get User Progress
```http
GET /users/{userId}/progress
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "total_lessons_completed": 42,
    "total_practiced_minutes": 1250.5,
    "average_score": 82.3,
    "best_score": 98.5,
    "average_accuracy": 89.2,
    "average_tempo": 118.5,
    "fastest_tempo": 140.0,
    "best_difficulty": "advanced",
    "current_level": "advanced",
    "last_practiced_date": "2026-02-20T10:00:00Z"
  }
}
```

### Get User Metrics
```http
GET /users/{userId}/metrics
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "total_lessons": 42,
    "total_time": 1250.5,
    "average_score": 82.3,
    "best_score": 98.5,
    "accuracy_metrics": {
      "average": 89.2,
      "best": 100.0,
      "trend": "improving"
    },
    "tempo_metrics": {
      "average": 118.5,
      "fastest": 140.0,
      "trend": "improving"
    },
    "recent_scores": [85.2, 88.5, 91.3, 89.2, 92.5],
    "difficulty_progression": ["beginner", "intermediate", "advanced"]
  }
}
```

### Get Performance Analysis
```http
GET /users/{userId}/performance
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "session_count": 42,
    "score_trend": "positive",
    "accuracy_trend": "improving",
    "tempo_trend": "improving",
    "weak_areas": ["sixteenth_notes", "arpeggios"],
    "strong_areas": ["scales", "chords"],
    "recommendation": "Focus on sixteenth note exercises"
  }
}
```

### Get Leaderboard
```http
GET /leaderboard
```

**Query Parameters:**
- `metric` - Sort by: `score`, `accuracy`, `tempo`, `lessons` (default: score)
- `limit` - Results per page (default: 20)
- `offset` - Pagination offset (default: 0)

**Response:**
```json
{
  "status": "success",
  "data": {
    "metric": "score",
    "entries": [
      {
        "rank": 1,
        "user_id": 12,
        "username": "piano_master",
        "total_lessons": 200,
        "average_score": 96.5,
        "best_score": 100.0,
        "accuracy": 98.2,
        "tempo": 135.0
      }
    ],
    "total": 3500
  }
}
```

## Music Theory Endpoints

### Generate Theory Quiz
```http
POST /theory-quiz
Content-Type: application/json
```

**Request Body:**
```json
{
  "difficulty": "intermediate",
  "count": 5,
  "user_id": 1
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "id": 789,
    "user_id": 1,
    "difficulty": "intermediate",
    "topic": "scales",
    "questions": [
      {
        "id": 1,
        "question": "What notes are in the C major scale?",
        "options": ["C D E F G A B", "C D Eb F G Ab Bb", "C D E F# G A B"],
        "correct_answer": "C D E F G A B"
      }
    ],
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

### Get Quiz Details
```http
GET /theory-quiz/{quizId}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 789,
    "user_id": 1,
    "difficulty": "intermediate",
    "topic": "scales",
    "questions": [
      { /* question objects */ }
    ],
    "completed": false,
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

### Submit Quiz Answers
```http
POST /theory-quiz/{quizId}/submit
Content-Type: application/json
```

**Request Body:**
```json
{
  "answers": [
    {"question_id": 1, "answer": "C D E F G A B"},
    {"question_id": 2, "answer": "7 semitones"}
  ]
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 789,
    "score": 90.0,
    "completed": true,
    "results": [
      {
        "question_id": 1,
        "correct": true,
        "user_answer": "C D E F G A B"
      }
    ]
  }
}
```

## Recommendations Endpoints

### Get Lesson Recommendations
```http
GET /recommend/{userId}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "current_level": "intermediate",
    "recommended_songs": [
      {
        "id": 42,
        "title": "Prelude in C Major",
        "composer": "Bach",
        "difficulty": "intermediate",
        "reason": "Matches your skill level and improves fundamentals"
      }
    ],
    "next_steps": ["Master scales", "Practice arpeggios", "Learn chord progressions"]
  }
}
```

### Get Progression Path
```http
GET /progression-path/{userId}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "current_level": "intermediate",
    "path": [
      {
        "level": "beginner",
        "status": "completed",
        "songs": 15,
        "completion_date": "2026-01-15T10:00:00Z"
      },
      {
        "level": "intermediate",
        "status": "in_progress",
        "songs": 8,
        "completion_percent": 53
      },
      {
        "level": "advanced",
        "status": "locked",
        "songs": 0,
        "unlock_requirement": "Complete 15 intermediate songs"
      }
    ]
  }
}
```

### Get Next Lesson
```http
GET /next-lesson/{userId}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user_id": 1,
    "next_song": {
      "id": 50,
      "title": "Invention No. 1",
      "composer": "Bach",
      "difficulty": "intermediate",
      "bpm": 120,
      "total_notes": 400,
      "reason": "Recommended for improving hand coordination"
    }
  }
}
```

---

## Error Handling

All endpoints follow consistent error handling. Common error codes:

- `INVALID_REQUEST` - Malformed request or invalid parameters
- `NOT_FOUND` - Resource not found
- `VALIDATION_ERROR` - Request validation failed
- `DATABASE_ERROR` - Database operation failed
- `UNAUTHORIZED` - Authentication required
- `INTERNAL_ERROR` - Server error

**Example Error Response:**
```json
{
  "status": "error",
  "error": "Book with ID 999 not found",
  "code": "NOT_FOUND",
  "details": {
    "resource": "book",
    "id": 999
  },
  "timestamp": "2026-02-20T10:00:00Z"
}
```

---

## Rate Limiting

API rate limits (subject to deployment configuration):
- **Standard Users**: 100 requests per minute
- **Premium Users**: 1000 requests per minute
- **Burst Limit**: 20 requests per second

Rate limit headers included in responses:
- `X-RateLimit-Limit` - Total requests allowed
- `X-RateLimit-Remaining` - Requests remaining
- `X-RateLimit-Reset` - Time when limit resets (Unix timestamp)

---

## Testing

### Using cURL

**List books:**
```bash
curl -X GET "http://localhost:8081/api/books?limit=5"
```

**Create reading session:**
```bash
curl -X POST "http://localhost:8081/api/sessions" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "book_id": 5, "duration": 600.0, "errors": 3}'
```

**Get piano leaderboard:**
```bash
curl -X GET "http://localhost:8082/api/leaderboard?metric=score&limit=10"
```

### Using Postman

1. Import endpoints from this documentation
2. Set base URL variables for each app
3. Use environment variables for user IDs and resource IDs
4. Test both success and error scenarios

---

## Changelog

### Version 1.0.0 (2026-02-20)
- Initial API release for Reading and Piano apps
- Full CRUD operations for books, songs, sessions
- User statistics and leaderboards
- Music theory quiz support
- MIDI file handling
- Performance analysis endpoints
