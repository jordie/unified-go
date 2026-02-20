# Piano App - Complete API Reference

## Base URL

```
http://localhost:8080/piano
```

## Authentication

All endpoints support optional user identification via:
- `X-User-ID` header
- `user_id` query parameter
- Session cookie (from login)

---

## Song Management

### GET /piano/api/songs

List all available piano songs with filtering and pagination.

**Parameters:**
```
limit    (query, optional, default: 20)     - Results per page (max 100)
offset   (query, optional, default: 0)      - Pagination offset
difficulty (query, optional) - Filter: beginner, intermediate, advanced, master
sort     (query, optional) - Sort field: title, composer, bpm, difficulty
```

**Example Requests:**
```bash
# Get all songs
curl http://localhost:8080/piano/api/songs

# Get first 10 songs
curl "http://localhost:8080/piano/api/songs?limit=10"

# Get beginner songs
curl "http://localhost:8080/piano/api/songs?difficulty=beginner"

# Pagination
curl "http://localhost:8080/piano/api/songs?limit=20&offset=20"
```

**Response (200 OK):**
```json
{
  "limit": 20,
  "offset": 0,
  "total": 20,
  "songs": [
    {
      "id": 1,
      "title": "Twinkle Twinkle Little Star",
      "composer": "Traditional",
      "description": "Classic beginner piece with simple melody",
      "difficulty": "beginner",
      "duration": 45.0,
      "bpm": 80,
      "time_signature": "4/4",
      "key_signature": "C Major",
      "total_notes": 26,
      "created_at": "2026-02-20T12:00:00Z"
    },
    ...
  ]
}
```

---

### GET /piano/api/songs/{id}

Get detailed information about a specific song.

**Path Parameters:**
```
id (required) - Song ID (integer)
```

**Example Requests:**
```bash
# Get song 1
curl http://localhost:8080/piano/api/songs/1

# With user context
curl -H "X-User-ID: 1" http://localhost:8080/piano/api/songs/1
```

**Response (200 OK):**
```json
{
  "id": 1,
  "title": "Twinkle Twinkle Little Star",
  "composer": "Traditional",
  "description": "Classic beginner piece with simple melody",
  "difficulty": "beginner",
  "duration": 45.0,
  "bpm": 80,
  "time_signature": "4/4",
  "key_signature": "C Major",
  "total_notes": 26,
  "midi_file": "base64_encoded_binary_data...",
  "created_at": "2026-02-20T12:00:00Z",
  "updated_at": "2026-02-20T12:00:00Z"
}
```

**Error Responses:**
```json
// 404 Not Found
{
  "error": "Song not found"
}

// 500 Internal Server Error
{
  "error": "Failed to get song"
}
```

---

### POST /piano/api/songs

Create a new song in the catalog.

**Request Body:**
```json
{
  "title": "String (required, 1-200 chars)",
  "composer": "String (required, 1-100 chars)",
  "description": "String (optional)",
  "difficulty": "String (required: beginner, intermediate, advanced, master)",
  "duration": "Number (required, seconds)",
  "bpm": "Number (required, 40-300)",
  "time_signature": "String (required, e.g., '4/4')",
  "key_signature": "String (required, e.g., 'C Major')",
  "total_notes": "Number (required, integer)",
  "midi_file": "String (optional, base64 encoded binary)"
}
```

**Example Request:**
```bash
curl -X POST http://localhost:8080/piano/api/songs \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Piece",
    "composer": "John Composer",
    "description": "A new classical piece",
    "difficulty": "intermediate",
    "duration": 300,
    "bpm": 100,
    "time_signature": "4/4",
    "key_signature": "G Major",
    "total_notes": 250
  }'
```

**Response (201 Created):**
```json
{
  "id": 21,
  "title": "New Piece",
  "composer": "John Composer",
  "description": "A new classical piece",
  "difficulty": "intermediate",
  "duration": 300,
  "bpm": 100,
  "time_signature": "4/4",
  "key_signature": "G Major",
  "total_notes": 250,
  "created_at": "2026-02-20T12:30:00Z"
}
```

---

## Lesson Management

### POST /piano/api/lessons

Start a new practice lesson on a song.

**Request Body:**
```json
{
  "song_id": "Number (required)",
  "user_id": "Number (required)"
}
```

**Example Request:**
```bash
curl -X POST http://localhost:8080/piano/api/lessons \
  -H "Content-Type: application/json" \
  -d '{
    "song_id": 1,
    "user_id": 1
  }'
```

**Response (201 Created):**
```json
{
  "id": 100,
  "user_id": 1,
  "song_id": 1,
  "start_time": "2026-02-20T12:40:00Z",
  "duration": 0,
  "notes_correct": 0,
  "notes_total": 26,
  "accuracy": 0,
  "tempo_accuracy": 0,
  "score": 0,
  "completed": false,
  "created_at": "2026-02-20T12:40:00Z"
}
```

---

### GET /piano/api/lessons/{id}

Get details about a specific lesson.

**Path Parameters:**
```
id (required) - Lesson ID (integer)
```

**Example Request:**
```bash
curl http://localhost:8080/piano/api/lessons/1
```

**Response (200 OK):**
```json
{
  "id": 1,
  "user_id": 1,
  "song_id": 1,
  "start_time": "2026-02-20T12:00:00Z",
  "end_time": "2026-02-20T12:00:45Z",
  "duration": 45.0,
  "notes_correct": 24,
  "notes_total": 26,
  "accuracy": 92.3,
  "tempo_accuracy": 95.0,
  "score": 93.65,
  "completed": true,
  "created_at": "2026-02-20T12:00:00Z"
}
```

---

### GET /piano/api/users/{userId}/lessons

Get all lessons for a specific user.

**Path Parameters:**
```
userId (required) - User ID (integer)
```

**Query Parameters:**
```
limit    (optional, default: 20)   - Results per page
offset   (optional, default: 0)    - Pagination offset
sort     (optional) - Sort field: created_at, score, accuracy
```

**Example Request:**
```bash
curl "http://localhost:8080/piano/api/users/1/lessons?limit=10"
```

**Response (200 OK):**
```json
{
  "lessons": [
    {
      "id": 3,
      "user_id": 1,
      "song_id": 3,
      "duration": 90.0,
      "notes_correct": 58,
      "notes_total": 64,
      "accuracy": 90.6,
      "score": 92.05,
      "completed": true,
      "created_at": "2026-02-19T14:00:00Z"
    },
    ...
  ],
  "total": 12,
  "limit": 10,
  "offset": 0
}
```

---

## User Progress & Metrics

### GET /piano/api/users/{userId}/progress

Get user's piano learning progress.

**Path Parameters:**
```
userId (required) - User ID (integer)
```

**Example Request:**
```bash
curl http://localhost:8080/piano/api/users/1/progress
```

**Response (200 OK):**
```json
{
  "user_id": 1,
  "lessons_completed": 12,
  "songs_mastered": 3,
  "total_practice_minutes": 180,
  "average_accuracy": 87.5,
  "accuracy_by_difficulty": {
    "beginner": 94.2,
    "intermediate": 89.1,
    "advanced": 78.5,
    "master": null
  },
  "recent_lessons": [
    {
      "song_id": 3,
      "title": "Ode to Joy",
      "accuracy": 90.6,
      "completed": "2026-02-20T14:00:00Z"
    }
  ]
}
```

---

### GET /piano/api/users/{userId}/metrics

Get detailed performance metrics for a user.

**Path Parameters:**
```
userId (required) - User ID (integer)
```

**Example Request:**
```bash
curl http://localhost:8080/piano/api/users/1/metrics
```

**Response (200 OK):**
```json
{
  "user_id": 1,
  "total_lessons": 12,
  "average_accuracy": 87.5,
  "best_score": 93.65,
  "total_practice_time_minutes": 180,
  "skill_level": "advanced",
  "practice_by_difficulty": {
    "beginner": {
      "lessons": 3,
      "average_accuracy": 94.2,
      "best_score": 97.5
    },
    "intermediate": {
      "lessons": 5,
      "average_accuracy": 89.1,
      "best_score": 92.3
    },
    "advanced": {
      "lessons": 4,
      "average_accuracy": 78.5,
      "best_score": 84.0
    },
    "master": {
      "lessons": 0,
      "average_accuracy": 0,
      "best_score": 0
    }
  },
  "improvement_percentage": 12.5,
  "estimated_next_milestone": "Master Level 50% accuracy"
}
```

---

### GET /piano/api/users/{userId}/evaluation

Get comprehensive performance evaluation for a user.

**Path Parameters:**
```
userId (required) - User ID (integer)
```

**Example Request:**
```bash
curl http://localhost:8080/piano/api/users/1/evaluation
```

**Response (200 OK):**
```json
{
  "user_id": 1,
  "current_level": "advanced",
  "assessment": {
    "rhythm_accuracy": 87.5,
    "tempo_consistency": 85.2,
    "note_accuracy": 89.3,
    "overall_performance": 87.3
  },
  "strengths": [
    "Excellent note accuracy",
    "Good finger technique",
    "Consistent practice"
  ],
  "areas_for_improvement": [
    "Tempo consistency on advanced pieces",
    "Dynamic expression control",
    "Hand position on larger jumps"
  ],
  "recommendations": [
    "Practice metronome exercises",
    "Work on expressive dynamics",
    "Strengthen finger independence"
  ]
}
```

---

## Leaderboard

### GET /piano/api/leaderboard

Get the global leaderboard of top performers.

**Query Parameters:**
```
limit      (optional, default: 100, max: 1000) - Number of results
difficulty (optional) - Filter: beginner, intermediate, advanced, master
sort_by    (optional) - Field: best_score, average_accuracy, lessons (default: best_score)
```

**Example Requests:**
```bash
# Top 100 performers
curl http://localhost:8080/piano/api/leaderboard

# Top 10
curl "http://localhost:8080/piano/api/leaderboard?limit=10"

# Beginner leaderboard
curl "http://localhost:8080/piano/api/leaderboard?difficulty=beginner"

# Sort by accuracy
curl "http://localhost:8080/piano/api/leaderboard?sort_by=average_accuracy"
```

**Response (200 OK):**
```json
{
  "leaderboard": [
    {
      "rank": 1,
      "user_id": 5,
      "best_score": 96.2,
      "average_accuracy": 92.1,
      "lessons_completed": 24
    },
    {
      "rank": 2,
      "user_id": 1,
      "best_score": 93.65,
      "average_accuracy": 87.5,
      "lessons_completed": 12
    },
    {
      "rank": 3,
      "user_id": 3,
      "best_score": 91.5,
      "average_accuracy": 85.2,
      "lessons_completed": 10
    }
  ],
  "limit": 100,
  "total_users": 9
}
```

---

## Music Theory

### POST /piano/api/theory-quiz

Generate a music theory quiz.

**Request Body:**
```json
{
  "difficulty": "String (optional: beginner, intermediate, advanced, master)",
  "topic": "String (optional: chord_identification, scales, intervals, notation, keys)",
  "questions_count": "Number (optional, default: 10, max: 50)"
}
```

**Example Request:**
```bash
curl -X POST http://localhost:8080/piano/api/theory-quiz \
  -H "Content-Type: application/json" \
  -d '{
    "difficulty": "intermediate",
    "topic": "chord_identification",
    "questions_count": 10
  }'
```

**Response (201 Created):**
```json
{
  "quiz_id": 1,
  "questions": [
    {
      "id": 1,
      "question": "What is this chord?",
      "options": ["C Major", "C Minor", "C Dominant", "C Diminished"],
      "difficulty": "intermediate"
    },
    ...
  ],
  "created_at": "2026-02-20T12:50:00Z"
}
```

---

### GET /piano/api/sessions/{sessionId}/analysis

Analyze quiz answers and get results.

**Path Parameters:**
```
sessionId (required) - Quiz session ID
```

**Query Parameters:**
```
user_id (optional) - User ID for scoring context
```

**Example Request:**
```bash
curl http://localhost:8080/piano/api/sessions/1/analysis?user_id=1
```

**Response (200 OK):**
```json
{
  "session_id": 1,
  "user_id": 1,
  "questions_answered": 10,
  "questions_correct": 8,
  "score": 80.0,
  "results": [
    {
      "question_id": 1,
      "correct": true,
      "user_answer": "C Major",
      "correct_answer": "C Major",
      "explanation": "This is a C Major triad..."
    },
    ...
  ],
  "feedback": "Good job! You scored 80%. Focus on diminished chords."
}
```

---

## Recommendations

### GET /piano/api/recommend/{userId}

Get personalized lesson recommendations for a user.

**Path Parameters:**
```
userId (required) - User ID (integer)
```

**Example Request:**
```bash
curl http://localhost:8080/piano/api/recommend/1
```

**Response (200 OK):**
```json
{
  "user_id": 1,
  "current_level": "advanced",
  "recommendations": [
    {
      "song_id": 14,
      "title": "Rondo Alla Turca",
      "composer": "Mozart",
      "difficulty": "advanced",
      "reason": "Perfect for your current level",
      "estimated_practice_time": "30 minutes"
    },
    {
      "song_id": 16,
      "title": "Goldberg Variations",
      "composer": "Bach",
      "difficulty": "master",
      "reason": "Challenge yourself with a master piece",
      "estimated_practice_time": "60 minutes"
    }
  ]
}
```

---

### GET /piano/api/progression-path/{userId}

Get the recommended progression path for a user.

**Path Parameters:**
```
userId (required) - User ID (integer)
```

**Example Request:**
```bash
curl http://localhost:8080/piano/api/progression-path/1
```

**Response (200 OK):**
```json
{
  "user_id": 1,
  "current_stage": "Intermediate Master",
  "progression_path": [
    {
      "stage": 1,
      "level": "Beginner",
      "status": "completed",
      "songs_completed": 3
    },
    {
      "stage": 2,
      "level": "Intermediate",
      "status": "in_progress",
      "songs_completed": 5,
      "songs_remaining": 2
    },
    {
      "stage": 3,
      "level": "Advanced",
      "status": "next",
      "songs_to_unlock": 5
    },
    {
      "stage": 4,
      "level": "Master",
      "status": "locked",
      "requirements": "Complete all advanced pieces"
    }
  ],
  "estimated_completion": "2026-04-20"
}
```

---

## UI Pages

### GET /piano/

Piano app homepage.

**Response:** HTML page with:
- Song catalog browser
- Quick practice buttons
- User dashboard link
- Statistics overview

---

### GET /piano/dashboard

User dashboard with progress and stats.

**Response:** HTML page with:
- User statistics
- Recent lessons
- Progress charts
- Leaderboard rankings
- Recommendation cards

---

## Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Successfully retrieved data |
| 201 | Created | Successfully created resource |
| 400 | Bad Request | Invalid parameters |
| 401 | Unauthorized | User not authenticated |
| 404 | Not Found | Resource doesn't exist |
| 500 | Server Error | Internal error |

---

## Error Handling

All error responses follow this format:

```json
{
  "error": "Error type",
  "message": "Detailed error message",
  "code": "ERROR_CODE",
  "timestamp": "2026-02-20T12:50:00Z"
}
```

---

## Rate Limiting

Not currently implemented. Contact development for production deployment needs.

---

## Webhooks

Not currently implemented. Contact development for integration needs.

---

**API Version:** 1.0
**Last Updated:** Phase 5 Subtask 4
**Status:** Core endpoints operational, handlers in optimization phase
