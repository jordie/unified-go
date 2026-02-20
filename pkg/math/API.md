# Math App REST API Reference

## Overview

The Math App exposes a comprehensive REST API with 20 endpoints organized into 6 functional categories. All endpoints return JSON responses with a consistent structure.

## Base URL

```
http://localhost:5000/math/api
https://192.168.1.231:5051/math/api (DEV)
https://192.168.1.231:5052/math/api (QA)
```

## Response Format

### Success Response
```json
{
  "success": true,
  "data": { /* endpoint-specific data */ },
  "error": null
}
```

### Error Response
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid difficulty level",
    "details": {}
  }
}
```

## Error Codes

| Code | HTTP | Description |
|------|------|-------------|
| VALIDATION_ERROR | 400 | Invalid input parameters |
| NOT_FOUND | 404 | Resource not found |
| UNAUTHORIZED | 401 | Authentication required |
| INTERNAL_ERROR | 500 | Server error |
| DB_ERROR | 500 | Database error |

---

## 1. Core Practice Endpoints

### GET /api/practice/next-question

Get next question for practice session.

**Parameters:**
```
user_id: int (required)
mode: string (required) - addition|subtraction|multiplication|division|mixed|remediation
difficulty: string (optional) - easy|medium|hard|expert (default: based on level)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "question_id": "q_123",
    "question": "7 + 8 = ?",
    "mode": "addition",
    "difficulty": "medium",
    "fact_family": "addition_double",
    "hint": "Think of 7 + 7 = 14, then add 1 more",
    "time_limit": 30
  }
}
```

### POST /api/practice/check-answer

Submit answer and get result.

**Body:**
```json
{
  "user_id": 1,
  "question_id": "q_123",
  "user_answer": "15",
  "time_taken": 2.5,
  "mode": "addition"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "is_correct": true,
    "correct_answer": "15",
    "explanation": "7 + 8 = 15 ✓",
    "mastery_level": 75,
    "next_interval": "6 days",
    "streak": 5,
    "xp_earned": 10,
    "feedback": "Great! Keep going!"
  }
}
```

### POST /api/practice/save-session

Save completed practice session.

**Body:**
```json
{
  "user_id": 1,
  "mode": "addition",
  "difficulty": "medium",
  "total_questions": 10,
  "correct_answers": 8,
  "total_time": 120.5,
  "questions": [
    {
      "question": "5 + 3",
      "user_answer": "8",
      "is_correct": true,
      "time_taken": 2.5,
      "fact_family": "addition_basic"
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "session_id": "sess_456",
    "accuracy": 80.0,
    "average_time": 12.05,
    "xp_earned": 80,
    "level_up": false,
    "badge_earned": null,
    "saved_at": "2026-02-20T10:30:00Z"
  }
}
```

### GET /api/practice/leaderboard

Get top performers by accuracy.

**Parameters:**
```
limit: int (optional, default: 10, max: 100)
mode: string (optional) - Filter by mode
period: string (optional) - all|week|month
```

**Response:**
```json
{
  "success": true,
  "data": {
    "leaderboard": [
      {
        "rank": 1,
        "user_id": 5,
        "username": "mathstar",
        "accuracy": 95.5,
        "xp": 5000,
        "level": 20
      },
      {
        "rank": 2,
        "user_id": 3,
        "username": "numbermaster",
        "accuracy": 92.3,
        "xp": 4500,
        "level": 19
      }
    ],
    "period": "all"
  }
}
```

---

## 2. Practice Management Endpoints

### GET /api/users/{user_id}

Get user profile and stats.

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": 1,
    "username": "john",
    "level": 12,
    "xp": 3500,
    "total_sessions": 45,
    "total_questions": 450,
    "overall_accuracy": 82.3,
    "created_at": "2025-01-15T08:00:00Z",
    "last_active": "2026-02-20T10:30:00Z"
  }
}
```

### POST /api/users

Create new user or get existing user.

**Body:**
```json
{
  "username": "newstudent",
  "auto_create": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": 42,
    "username": "newstudent",
    "created": true,
    "created_at": "2026-02-20T10:30:00Z"
  }
}
```

### GET /api/practice/history

Get user's practice session history.

**Parameters:**
```
user_id: int (required)
limit: int (optional, default: 20)
offset: int (optional, default: 0)
mode: string (optional) - Filter by mode
```

**Response:**
```json
{
  "success": true,
  "data": {
    "sessions": [
      {
        "session_id": "sess_456",
        "mode": "addition",
        "difficulty": "medium",
        "total_questions": 10,
        "correct_answers": 8,
        "accuracy": 80.0,
        "total_time": 120.5,
        "created_at": "2026-02-20T10:00:00Z"
      }
    ],
    "total_count": 45,
    "limit": 20,
    "offset": 0
  }
}
```

---

## 3. Learning Analytics Endpoints

### GET /api/analytics/dashboard

Get comprehensive user analytics.

**Parameters:**
```
user_id: int (required)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "overall": {
      "total_sessions": 45,
      "total_questions": 450,
      "overall_accuracy": 82.3,
      "average_session_length": 10,
      "total_practice_time": 450
    },
    "by_mode": {
      "addition": {
        "accuracy": 85.0,
        "sessions": 15,
        "level": 10
      },
      "subtraction": {
        "accuracy": 80.0,
        "sessions": 12,
        "level": 8
      }
    },
    "time_analysis": {
      "best_time": "morning (9am-12pm)",
      "best_accuracy": 85.5,
      "by_hour": [
        { "hour": 9, "accuracy": 85.0, "sessions": 5 },
        { "hour": 10, "accuracy": 86.0, "sessions": 6 }
      ]
    },
    "weak_families": [
      {
        "family": "subtraction_regrouping",
        "accuracy": 65.0,
        "attempts": 20,
        "errors": 7
      }
    ],
    "progress_trend": {
      "last_7_days": "+3.5%",
      "last_30_days": "+8.2%",
      "trajectory": "improving"
    }
  }
}
```

### GET /api/analytics/mistakes

Get most common mistakes.

**Parameters:**
```
user_id: int (required)
limit: int (optional, default: 10)
mode: string (optional)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "mistakes": [
      {
        "question": "12 - 5",
        "correct_answer": "7",
        "user_answers": {
          "6": 3,
          "8": 2,
          "7": 5
        },
        "total_errors": 5,
        "last_error": "2026-02-19T15:30:00Z",
        "pattern": "Subtraction with regrouping"
      }
    ]
  }
}
```

### GET /api/analytics/weakness-report

Get detailed weakness analysis with recommendations.

**Parameters:**
```
user_id: int (required)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "weak_families": [
      {
        "family": "subtraction_regrouping",
        "accuracy": 65.0,
        "sessions_attempted": 20,
        "recommendation": "Practice borrowing from tens place",
        "target_accuracy": 85,
        "priority": "high"
      }
    ],
    "overall_recommendation": "Focus on subtraction with regrouping. You're close!",
    "estimated_time_to_mastery": "5-7 days"
  }
}
```

### GET /api/analytics/progress-chart

Get progress data for charting.

**Parameters:**
```
user_id: int (required)
period: string (optional) - week|month|year (default: month)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "period": "month",
    "data_points": [
      {
        "date": "2026-01-20",
        "accuracy": 75.0,
        "sessions": 3,
        "questions": 30
      },
      {
        "date": "2026-01-21",
        "accuracy": 78.0,
        "sessions": 4,
        "questions": 40
      }
    ]
  }
}
```

---

## 4. Spaced Repetition Endpoints

### GET /api/sr/due-facts

Get facts due for spaced repetition review.

**Parameters:**
```
user_id: int (required)
limit: int (optional, default: 10)
mode: string (optional)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "due_facts": [
      {
        "fact": "7 + 8",
        "mode": "addition",
        "next_review": "2026-02-20T12:00:00Z",
        "interval_days": 6,
        "ease_factor": 2.6,
        "review_count": 2,
        "days_overdue": 0
      }
    ],
    "total_due": 12,
    "total_scheduled": 156
  }
}
```

### POST /api/sr/record-review

Record a spaced repetition review.

**Body:**
```json
{
  "user_id": 1,
  "fact": "7 + 8",
  "mode": "addition",
  "quality": 5,
  "response_time": 2.5
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "fact": "7 + 8",
    "new_ease_factor": 2.7,
    "new_interval": 16,
    "next_review": "2026-03-07T12:00:00Z",
    "review_count": 3,
    "quality_rating": "Perfect"
  }
}
```

---

## 5. Assessment Endpoints

### POST /api/assessment/start

Start adaptive placement assessment.

**Body:**
```json
{
  "user_id": 1,
  "mode": "mixed",
  "question_count": 15
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "assessment_id": "assess_789",
    "starting_level": 7,
    "total_levels": 15,
    "question_count": 15,
    "started_at": "2026-02-20T10:30:00Z"
  }
}
```

### POST /api/assessment/submit-response

Submit assessment response.

**Body:**
```json
{
  "assessment_id": "assess_789",
  "user_id": 1,
  "question_number": 1,
  "is_correct": true,
  "response_time": 3.5
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "assessment_complete": false,
    "next_level": 9,
    "progress": "1/15",
    "current_accuracy": 100.0
  }
}
```

### GET /api/assessment/results/{assessment_id}

Get assessment results.

**Response:**
```json
{
  "success": true,
  "data": {
    "assessment_id": "assess_789",
    "placed_level": 10,
    "accuracy": 80.0,
    "confidence": 0.85,
    "total_questions": 12,
    "correct_answers": 10,
    "recommendation": "Your optimal level is 10 - Medium difficulty",
    "completed_at": "2026-02-20T10:45:00Z"
  }
}
```

---

## 6. Fact Family Endpoints

### GET /api/fact-families

Get all available fact families.

**Response:**
```json
{
  "success": true,
  "data": {
    "families": [
      {
        "id": "addition_doubles",
        "name": "Doubles (n + n)",
        "examples": ["2+2", "3+3", "5+5"],
        "difficulty": "easy",
        "mode": "addition"
      },
      {
        "id": "addition_make_ten",
        "name": "Make Ten (n + (10-n))",
        "examples": ["3+7", "4+6", "5+5"],
        "difficulty": "medium",
        "mode": "addition"
      }
    ],
    "total": 22
  }
}
```

### GET /api/fact-families/{family_id}/progress

Get mastery for specific fact family.

**Parameters:**
```
user_id: int (required)
family_id: string (required)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "family_id": "addition_doubles",
    "family_name": "Doubles",
    "mastery_level": 85,
    "accuracy": 85.0,
    "sessions": 15,
    "total_attempts": 30,
    "correct_attempts": 26,
    "status": "mastered"
  }
}
```

### POST /api/fact-families/{family_id}/practice

Start practice session for fact family.

**Body:**
```json
{
  "user_id": 1,
  "family_id": "addition_doubles",
  "question_count": 10
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "session_id": "fam_sess_101",
    "family_id": "addition_doubles",
    "question_count": 10,
    "started_at": "2026-02-20T10:30:00Z"
  }
}
```

---

## 7. Advanced Endpoints

### GET /api/practice/recommendation

Get personalized practice recommendation.

**Parameters:**
```
user_id: int (required)
```

**Response:**
```json
{
  "success": true,
  "data": {
    "recommendation": "Focus on subtraction with regrouping",
    "reason": "Your accuracy is 65% - below your average of 82%",
    "suggested_mode": "remediation",
    "suggested_difficulty": "medium",
    "expected_benefit": "Increase overall accuracy by 3-5%",
    "estimated_time": "15-20 minutes",
    "priority": "high"
  }
}
```

### POST /api/practice/claim-badge

Award achievement badge.

**Body:**
```json
{
  "user_id": 1,
  "badge": "100_perfect_streak",
  "earned_at": "2026-02-20T10:30:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "badge": "100_perfect_streak",
    "description": "Achieved 100 perfect answers in a row",
    "image_url": "/badges/perfect_streak.png",
    "earned_at": "2026-02-20T10:30:00Z",
    "is_new": true
  }
}
```

---

## Example Client Usage

### JavaScript Fetch
```javascript
// Get next question
const response = await fetch('/math/api/practice/next-question', {
  method: 'GET',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ user_id: 1, mode: 'addition', difficulty: 'medium' })
});
const data = await response.json();
if (data.success) {
  displayQuestion(data.data.question);
}

// Submit answer
const checkResponse = await fetch('/math/api/practice/check-answer', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    user_id: 1,
    question_id: data.data.question_id,
    user_answer: '15',
    time_taken: 2.5,
    mode: 'addition'
  })
});
const result = await checkResponse.json();
if (result.data.is_correct) {
  showSuccess('Correct! ' + result.data.feedback);
}
```

### cURL
```bash
# Get user analytics
curl -X GET 'http://localhost:5000/math/api/analytics/dashboard?user_id=1' \
  -H 'Content-Type: application/json'

# Submit practice answer
curl -X POST 'http://localhost:5000/math/api/practice/check-answer' \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": 1,
    "question_id": "q_123",
    "user_answer": "15",
    "time_taken": 2.5,
    "mode": "addition"
  }'
```

---

## Rate Limiting

- Default: 100 requests per minute per user
- Heavy operations (/api/analytics/): 10 per minute
- Response headers include: `X-RateLimit-Remaining`, `X-RateLimit-Reset`

## Timeouts

- Standard endpoints: 5 second timeout
- Analytics endpoints: 10 second timeout
- Assessment endpoints: 2 second timeout per question

## Status Codes

- `200 OK` - Successful request
- `201 Created` - Resource created
- `400 Bad Request` - Invalid parameters
- `401 Unauthorized` - Authentication required
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

**API Version:** 1.0.0
**Last Updated:** Phase 6 Migration
**Status:** Production Ready
