# Phase 5: Piano Application

A comprehensive piano learning platform that helps users improve their musical skills through structured practice sessions, performance tracking, and personalized recommendations.

## Overview

The Piano app provides an interactive environment for users to:
- **Learn Songs**: Access a library of piano pieces at various difficulty levels
- **Practice Lessons**: Record practice sessions with MIDI support
- **Track Performance**: Monitor accuracy, tempo control, and skill progression
- **Study Theory**: Complete interactive music theory quizzes
- **Compete**: View leaderboards ranked by score, accuracy, and tempo
- **Get Recommendations**: Receive personalized lesson suggestions based on skill level

## Architecture

```
pkg/piano/
├── models.go              # Data models (Song, Session, etc.)
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
│   ├── songs.html         # Song catalog
│   ├── practice.html      # Practice interface with recording
│   ├── dashboard.html     # User statistics and progress
│   ├── leaderboard.html   # Competitive rankings
│   └── theory.html        # Music theory quiz interface
└── README.md              # This file
```

## Data Models

### Song
Represents a piano piece in the catalog.

```go
type Song struct {
    ID            uint      // Unique identifier
    Title         string    // Song title
    Composer      string    // Composer name
    Difficulty    string    // "beginner", "intermediate", "advanced", "master"
    BPM           int       // Beats per minute (tempo)
    TimeSignature string    // Time signature (e.g., "4/4")
    KeySignature  string    // Key (e.g., "C Major")
    TotalNotes    int       // Total number of notes in piece
    Duration      float64   // Duration in seconds
    Description   string    // Learning notes and tips
    MIDIFile      []byte    // Binary MIDI file data (blob)
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### PracticeSession
Captures a single practice recording.

```go
type PracticeSession struct {
    ID            uint      // Unique identifier
    UserID        uint      // User practicing
    SongID        uint      // Song being practiced
    LessonID      uint      // Associated lesson (optional)
    RecordingMIDI []byte    // Binary MIDI recording (blob)
    Duration      float64   // Practice duration (seconds)
    NotesHit      int       // Correct notes played
    NotesTotal    int       // Total notes in piece
    TempoAverage  float64   // Average tempo (BPM)
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### PianoLesson
Represents a structured lesson unit.

```go
type PianoLesson struct {
    ID            uint      // Unique identifier
    UserID        uint      // Student
    SongID        uint      // Song being learned
    StartTime     time.Time // Lesson start
    EndTime       time.Time // Lesson end
    Duration      float64   // Total duration (seconds)
    NotesCorrect  int       // Correct notes played
    NotesTotal    int       // Total notes required
    Accuracy      float64   // Accuracy percentage (0-100)
    TempoAccuracy float64   // Tempo accuracy percentage (0-100)
    Score         float64   // Composite score (0-100)
    Completed     bool      // Lesson completion status
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### MusicTheoryQuiz
Interactive music theory assessment.

```go
type MusicTheoryQuiz struct {
    ID         uint      // Unique identifier
    UserID     uint      // Student taking quiz
    LessonID   uint      // Associated lesson
    Topic      string    // Quiz topic ("scales", "intervals", "chords", etc.)
    Questions  string    // JSON array of questions
    Answers    string    // JSON array of user answers
    Score      float64   // Quiz score (0-100)
    Difficulty string    // "beginner", "intermediate", "advanced"
    Completed  bool      // Completion status
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### UserProgress
Aggregated metrics for a user.

```go
type UserProgress struct {
    UserID                uint       // User identifier
    TotalLessonsCompleted int        // Total lessons finished
    TotalPracticedMinutes float64    // Total practice time (minutes)
    AverageScore          float64    // Average composite score (0-100)
    BestScore             float64    // Highest score achieved
    AverageAccuracy       float64    // Average note accuracy (0-100)
    AverageTempo          float64    // Average tempo (BPM)
    FastestTempo          float64    // Fastest tempo achieved (BPM)
    BestDifficulty        string     // Hardest successfully completed difficulty
    CurrentLevel          string     // Estimated skill level
    LastPracticedDate     *time.Time // Most recent practice date
}
```

### MusicQuestion
Theory quiz question data.

```go
type MusicQuestion struct {
    ID            uint      // Question ID
    Question      string    // Question text
    Options       []string  // Multiple choice options
    CorrectAnswer string    // Correct answer
    Explanation   string    // Educational explanation
}
```

## API Endpoints

### Songs
- `GET /api/songs` - List all songs (filterable by difficulty, composer, key)
- `POST /api/songs` - Create a new song
  ```json
  {
    "title": "Moonlight Sonata",
    "composer": "Beethoven",
    "difficulty": "advanced",
    "bpm": 60,
    "time_signature": "4/4",
    "key_signature": "C# minor",
    "total_notes": 1000,
    "duration": 600.0,
    "midi_file": "binary_data"
  }
  ```
- `GET /api/songs/{songId}` - Get song details with MIDI
- `PUT /api/songs/{songId}` - Update song metadata

### Practice Sessions
- `POST /api/practice` - Record a practice session
  ```json
  {
    "user_id": 1,
    "song_id": 5,
    "recorded_bpm": 120.0,
    "duration": 300.0,
    "notes_correct": 850,
    "notes_total": 1000
  }
  ```
- `GET /api/practice/{sessionId}` - Get session details
- `GET /api/users/{userId}/sessions` - Get user's practice history
- `GET /api/midi/{songId}` - Download MIDI file for a song

### User Metrics
- `GET /api/users/{userId}/progress` - Get user progress tracking
- `GET /api/users/{userId}/metrics` - Get comprehensive performance metrics
- `GET /api/users/{userId}/performance` - Analyze performance trends
- `GET /api/leaderboard` - Get rankings (queryable by metric)
  - Query params: `metric=score|accuracy|tempo|lessons`
  - `?limit=10&offset=0` - Pagination

### Music Theory
- `POST /api/theory-quiz` - Generate a music theory quiz
  ```json
  {
    "difficulty": "intermediate",
    "count": 5,
    "user_id": 1
  }
  ```
- `GET /api/theory-quiz/{quizId}` - Get quiz details
- `POST /api/theory-quiz/{quizId}/submit` - Submit quiz answers
- `GET /api/theory-questions` - Get available theory questions

### Recommendations
- `GET /api/recommend/{userId}` - Get personalized lesson recommendations
- `GET /api/progression-path/{userId}` - Get suggested learning progression
- `GET /api/next-lesson/{userId}` - Get next recommended lesson

### MIDI Operations
- `POST /api/midi/upload` - Upload a MIDI file
- `GET /api/midi/{songId}` - Download song MIDI
- `POST /api/midi/analyze` - Analyze a MIDI recording
  ```json
  {
    "midi_data": "hex_encoded_midi"
  }
  ```

## Database Schema

### songs
```sql
CREATE TABLE songs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    composer TEXT,
    difficulty TEXT,
    bpm INTEGER,
    time_signature TEXT,
    key_signature TEXT,
    total_notes INTEGER,
    duration REAL,
    description TEXT,
    midi_file BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### piano_lessons
```sql
CREATE TABLE piano_lessons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    song_id INTEGER NOT NULL,
    start_time DATETIME,
    end_time DATETIME,
    duration REAL,
    notes_correct INTEGER,
    notes_total INTEGER,
    accuracy REAL,
    tempo_accuracy REAL,
    score REAL,
    completed BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### practice_sessions
```sql
CREATE TABLE practice_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    song_id INTEGER NOT NULL,
    lesson_id INTEGER,
    recording_midi BLOB,
    duration REAL,
    notes_hit INTEGER,
    notes_total INTEGER,
    tempo_average REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### music_theory_quizzes
```sql
CREATE TABLE music_theory_quizzes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    lesson_id INTEGER,
    topic TEXT,
    questions TEXT,
    answers TEXT,
    score REAL,
    difficulty TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_lessons_user ON piano_lessons(user_id);
CREATE INDEX idx_sessions_user ON practice_sessions(user_id);
CREATE INDEX idx_quizzes_user ON music_theory_quizzes(user_id);
```

## Running the Application

### Prerequisites
- Go 1.21+
- SQLite3
- MIDI support libraries (optional, for MIDI analysis)

### Setup
```bash
cd pkg/piano

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
go test -v -run TestCreateAndRetrieveSong ./...
```

### Example Usage
```go
package main

import (
    "context"
    "database/sql"
    "github.com/jgirmay/unified-go/pkg/piano"
)

func main() {
    // Open database
    db, _ := sql.Open("sqlite3", "piano.db")
    defer db.Close()

    // Create repository and service
    repo := piano.NewRepository(db)
    service := piano.NewService(repo)

    // Create a song
    ctx := context.Background()
    song := &piano.Song{
        Title:         "Moonlight Sonata",
        Composer:      "Beethoven",
        Difficulty:    "advanced",
        BPM:           60,
        TimeSignature: "4/4",
        KeySignature:  "C# minor",
        TotalNotes:    1000,
        Duration:      600.0,
        MIDIFile:      midiData,
    }

    songID, _ := repo.SaveSong(ctx, song)

    // Record a practice session
    session, _ := service.ProcessLesson(ctx, 1, songID, 120.0, 300.0, 850, 1000)

    // Calculate metrics
    accuracy := piano.CalculateAccuracy(session.NotesHit, session.NotesTotal)
    tempoAcc := piano.CalculateTempoAccuracy(session.TempoAverage, song.BPM)
    score := piano.CalculateCompositeScore(accuracy, tempoAcc, 0)

    // Get user progress
    progress, _ := repo.GetUserProgress(ctx, 1)
    println("Average Score:", progress.AverageScore)
    println("Skill Level:", progress.CurrentLevel)
}
```

## Performance Metrics

### Benchmark Results (Apple M2)
- **PracticeLesson**: ~25µs per operation
- **UserProgress**: ~22µs per operation
- **UserMetrics**: ~33µs per operation
- **CalculateAccuracy**: ~0.31ns per operation
- **CalculateTempoAccuracy**: ~0.32ns per operation
- **CalculateCompositeScore**: ~0.34ns per operation

### Test Coverage
- 13 integration tests covering complete workflows
- 3 performance benchmarks
- 100+ total tests across unit, repository, and integration suites

## Features

### Learning & Practice
- ✅ Structured lesson system with progression
- ✅ MIDI file support for song playback and recording
- ✅ Practice session recording and analysis
- ✅ Real-time accuracy tracking
- ✅ Tempo control and monitoring
- ✅ Difficulty levels (beginner/intermediate/advanced/master)

### Performance Tracking
- ✅ Personal statistics dashboard
- ✅ Historical session data and trends
- ✅ Accuracy and tempo metrics
- ✅ Skill level estimation
- ✅ Progress visualization

### Gamification
- ✅ Leaderboards (sortable by score, accuracy, tempo, lessons)
- ✅ Personal best tracking
- ✅ Achievement badges
- ✅ Difficulty-based challenges

### Music Theory
- ✅ Interactive theory quizzes
- ✅ Topic-based questions (scales, intervals, chords, etc.)
- ✅ Multiple difficulty levels
- ✅ Educational explanations
- ✅ Score tracking and analysis

### Recommendations
- ✅ Personalized lesson suggestions
- ✅ Difficulty-based progression paths
- ✅ Skill-level matching
- ✅ Learning optimization

## MIDI Support

### MIDI File Handling

The Piano app stores and processes MIDI files as binary blobs. MIDI files are used for:
- Song playback and learning
- Practice session recording (storing user's performance)
- Performance analysis

### MIDI Operations

```go
// Create song with MIDI
song := &piano.Song{
    Title:    "Song Name",
    MIDIFile: midiData, // []byte containing MIDI data
}

// Record practice session with MIDI
session := &piano.PracticeSession{
    UserID:        1,
    SongID:        songID,
    RecordingMIDI: recordedMIDI, // User's practice recording
    NotesHit:      850,
    NotesTotal:    1000,
}

// Analyze MIDI
midiHex := fmt.Sprintf("%x", midiData)
// Verify MIDI header: 4D 54 68 64 (MThd)
```

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

### Scoring Algorithm

The Piano app uses a composite scoring system:

```
Accuracy Score (0-100):
  = (NotesCorrect / NotesTotal) * 100

Tempo Accuracy (0-100):
  = 100 - (|ActualTempo - TargetTempo| / TargetTempo * 100)
  Capped at 0-100

Composite Score (0-100):
  = (Accuracy * 0.7) + (TempoAccuracy * 0.2) + (TheoryScore * 0.1)

Skill Level:
  - Beginner: Average Score < 50
  - Intermediate: 50-70
  - Advanced: 70-85
  - Master: >= 85
```

### Code Structure

- **Models Layer**: Data structures, validation, score calculations
- **Repository Layer**: Database operations, MIDI blob handling
- **Service Layer**: Business logic, lesson processing, recommendations
- **Router Layer**: HTTP routing, MIDI uploads/downloads, response formatting

## Troubleshooting

### Tests Failing

**Issue**: TestUserProgress fails with "Expected 3 lessons, got 0"
- **Solution**: Ensure practice sessions are created before querying progress

**Issue**: MIDI data corruption
- **Solution**: Verify MIDI header starts with `4D 54 68 64` (MThd)

### Performance Issues

**Issue**: Leaderboard queries are slow
- **Solution**: Add database indexes on frequently queried columns

**Issue**: MIDI file uploads timing out
- **Solution**: Increase upload timeout; compress MIDI files before transfer

### MIDI Problems

**Issue**: "Invalid MIDI file" error
- **Solution**: Verify MIDI file validity; check file starts with proper header

**Issue**: MIDI data truncated
- **Solution**: Ensure binary data is correctly encoded when storing/retrieving

## Contributing

When contributing to the Piano app:

1. Maintain test-first approach (write tests before implementation)
2. Keep integration tests synchronized with API changes
3. Run benchmarks before/after optimization
4. Update this README for new features
5. Ensure all tests pass: `go test -v ./...`
6. Validate MIDI file handling for binary operations

## Future Enhancements

- [ ] Real-time MIDI input via USB keyboard
- [ ] Audio playback with real MIDI synthesis
- [ ] Performance visualization graphs
- [ ] Automatic difficulty adjustment
- [ ] Ensemble/multiplayer features
- [ ] Sheet music display integration
- [ ] Mobile app support
- [ ] ML-based practice recommendations
- [ ] Video tutorials for each piece
- [ ] Metronome with tempo adjustment

## License

Part of the GAIA distributed development system for basic educational applications.
