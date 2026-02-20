# Phase 5 - Subtask 3: Test All Piano Endpoints

## Comprehensive Endpoint Testing Plan

### 1. Song Management Endpoints

#### GET /piano/api/songs
- List all songs with pagination
- Parameters: limit, offset
- Expected: 200 OK, all 20 songs

#### GET /piano/api/songs?difficulty={level}
- Filter by difficulty: beginner, intermediate, advanced, master
- Expected: 200 OK, filtered results

#### GET /piano/api/songs/{id}
- Get specific song details
- Expected: 200 OK with song data

#### POST /piano/api/songs
- Create new song
- Expected: 201 Created

### 2. Lesson Management Endpoints

#### POST /piano/api/lessons
- Start new practice lesson
- Expected: 201 Created

#### GET /piano/api/lessons/{id}
- Get lesson details
- Expected: 200 OK

#### GET /piano/api/users/{userId}/lessons
- Get user's lesson history
- Expected: 200 OK

### 3. User Stats Endpoints

#### GET /piano/api/users/{userId}/progress
- Get user progress
- Expected: 200 OK

#### GET /piano/api/users/{userId}/metrics
- Get performance metrics
- Expected: 200 OK

#### GET /piano/api/users/{userId}/evaluation
- Performance evaluation
- Expected: 200 OK

### 4. Music Theory Endpoints

#### POST /piano/api/theory-quiz
- Generate theory quiz
- Expected: 201 Created

#### GET /piano/api/sessions/{sessionId}/analysis
- Analyze theory answers
- Expected: 200 OK

### 5. Leaderboard & Rankings

#### GET /piano/api/leaderboard
- Get top performers
- Parameters: limit
- Expected: 200 OK

### 6. MIDI Operations

#### POST /piano/api/midi/upload
- Upload MIDI recording
- Expected: 201 Created

#### GET /piano/api/midi/{sessionId}
- Download MIDI
- Expected: 200 OK

### 7. Recommendations

#### GET /piano/api/recommend/{userId}
- Get lesson recommendations
- Expected: 200 OK

#### GET /piano/api/progression-path/{userId}
- Get progression path
- Expected: 200 OK

### 8. UI Pages

#### GET /piano/
- Piano app homepage
- Expected: 200 OK, HTML

#### GET /piano/dashboard
- User dashboard
- Expected: 200 OK, HTML

---

## Testing Results
(To be filled in)
