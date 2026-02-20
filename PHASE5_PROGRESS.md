# Phase 5 Implementation Progress - Piano App

**Branch**: `feature/phase5-piano-app-0220`
**Started**: 2026-02-20
**Target**: 100% completion in 2.5-3 weeks (most complex phase)
**Manager**: manager_piano
**Task IDs**: 80 (parent), subtasks 1-7

## Implementation Checklist

### Subtask 1: Create Data Models
**File**: `pkg/piano/models.go`
**Status**: ⏳ NOT STARTED

Required:
- [ ] Song struct (ID, Title, Composer, MIDIFile []byte, Difficulty, Duration, BPM, TimeSignature, CreatedAt)
- [ ] PianoLesson struct (ID, UserID, SongID, StartTime, EndTime, DurationPracticed, NotesCorrect, NotesTotal, TempoAccuracy, Score, CreatedAt)
- [ ] MusicTheoryQuiz struct (ID, UserID, LessonID, Questions []MusicQuestion, Score, Difficulty)
- [ ] UserProgress struct (UserID, TotalLessonsCompleted, TotalPracticedMinutes, AverageScore, FastestTempo, BestDifficulty)
- [ ] PracticeSession struct (ID, UserID, SongID, RecordingMIDI []byte, Duration, NotesHit, Tempo, CreatedAt)
- [ ] JSON marshaling/unmarshaling (with MIDI blob support)
- [ ] Validation methods (tempo 40-300 BPM, score 0-100, difficulty levels)
- [ ] Unit tests (10+ tests)

**Commit**: `models: Add piano data models with MIDI support`

---

### Subtask 2: Create Repository Layer
**File**: `pkg/piano/repository.go`
**Status**: ⏳ NOT STARTED

Required Functions:
- [ ] GetSongs(ctx, difficulty, limit, offset) ([]Song, error)
- [ ] GetSongByID(ctx, songID) (*Song, error)
- [ ] SaveLesson(ctx, lesson) error
- [ ] GetUserLessons(ctx, userID, limit, offset) ([]PianoLesson, error)
- [ ] SavePracticeSession(ctx, session) error
- [ ] StoreMIDIRecording(ctx, recordingData []byte) (string, error)
- [ ] GetMIDIRecording(ctx, recordingID) ([]byte, error)
- [ ] SaveMusicTheoryQuiz(ctx, quiz) error
- [ ] GetUserProgress(ctx, userID) (*UserProgress, error)
- [ ] GetTheoryQuestionsByDifficulty(ctx, difficulty) ([]MusicQuestion, error)
- [ ] Error wrapping with context
- [ ] MIDI blob handling and indexing
- [ ] Unit tests with mocks (15+ tests)

**Uses**: `internal/database/pool.go`, BLOB storage for MIDI files

**Commit**: `repo: Add piano data repository with MIDI support`

---

### Subtask 3: Create Service Layer
**File**: `pkg/piano/service.go`
**Status**: ⏳ NOT STARTED

Required Functions:
- [ ] CalculateAccuracy(recordedMIDI, expectedMIDI) float64
- [ ] CalculateTempo(recordedMIDI) float64
- [ ] AnalyzeMusicTheory(answers) float64
- [ ] ProcessLesson(userID, songID, recordedMIDI, duration) (*PianoLesson, error)
- [ ] GenerateLesson(userID, difficulty) (*PianoLesson, error)
- [ ] GetProgressionPath(userID) ([]Song, error)
- [ ] EvaluatePerformance(userID) (*UserProgress, error)
- [ ] GenerateMusicTheoryQuiz(difficulty, count) ([]MusicQuestion, error)
- [ ] MIDIParsing and analysis utilities
- [ ] Business logic implementations
- [ ] Unit tests (22+ tests - this is the most complex phase)

**Accuracy Formula**: (notes_hit / total_notes) * 100
**Tempo Accuracy**: abs(actual_tempo - target_tempo) / target_tempo
**Overall Score**: (accuracy * 70%) + (tempo_accuracy * 20%) + (theory_score * 10%)

**Commit**: `service: Add piano service with MIDI analysis and music theory`

---

### Subtask 4: Implement Router & Handlers
**Files**: `pkg/piano/router.go` + `handler.go`
**Status**: ⏳ NOT STARTED

Routes to implement:
- [ ] GET /piano/api/songs - List available songs (with filtering)
- [ ] GET /piano/api/songs/:id - Get song details with MIDI preview
- [ ] POST /piano/api/lessons/start - Start piano lesson
- [ ] POST /piano/api/lessons/:id/submit - Submit recorded lesson (MIDI upload)
- [ ] GET /piano/api/lessons/:id/playback - Stream recorded MIDI playback
- [ ] GET /piano/api/theory/quiz - Generate music theory quiz
- [ ] POST /piano/api/theory/submit - Submit theory answers
- [ ] GET /piano/api/progress - User progress stats
- [ ] GET /piano/api/scores/history - Lesson history with scores
- [ ] GET /piano/api/sheet-music/:id - Generate SVG sheet music visualization

Handler responsibilities:
- [ ] Parse JSON requests with MIDI blob uploads
- [ ] MIME type validation (audio/midi)
- [ ] Call services for processing
- [ ] Return JSON responses + streaming MIDI playback
- [ ] Error handling with context
- [ ] Integration tests (14+ tests)

**Commit**: `router: Add piano API routes and handlers with MIDI streaming`

---

### Subtask 5: Template Conversion
**Path**: `templates/piano/`
**Status**: ⏳ NOT STARTED

Convert from Jinja2 to Go html/template:
- [ ] templates/piano/index.html
- [ ] templates/piano/song_list.html
- [ ] templates/piano/lesson_player.html (with audio playback controls)
- [ ] templates/piano/theory_quiz.html
- [ ] templates/piano/progress.html
- [ ] templates/piano/sheet_music_viewer.html
- [ ] Static file integration (CSS/JS for MIDI playback, SVG rendering)

Template syntax changes:
- [ ] {{ variable }} stays same
- [ ] {% if %} → {{if}}
- [ ] {% for %} → {{range}}
- [ ] Music theory content scaffolding
- [ ] Sheet music SVG generation

**Commit**: `templates: Convert piano Jinja2 templates to Go html/template with sheet music`

---

### Subtask 6: Integration Tests
**File**: `pkg/piano/integration_test.go`
**Status**: ⏳ NOT STARTED

Test scenarios:
- [ ] GetSongs endpoint - Filtering by difficulty
- [ ] Lesson submission endpoint - MIDI upload and processing
- [ ] MIDI playback endpoint - Stream validation
- [ ] Accuracy calculation - Note detection from MIDI
- [ ] Tempo calculation - BPM extraction and comparison
- [ ] Theory quiz - Question generation and scoring
- [ ] Progress tracking - Score aggregation across lessons
- [ ] Sheet music generation - SVG output validation
- [ ] MIDI blob storage and retrieval
- [ ] Session validation - Auth middleware works
- [ ] Database persistence - MIDI files survive restart
- [ ] Load test - 50 concurrent lesson submissions (<30ms avg)

**Minimum**: 14+ passing tests
**Performance Target**: <30ms average response time
**Note**: This phase is most complex due to MIDI processing and music theory

**Commit**: `test: Add piano integration tests and MIDI analysis validation`

---

### Subtask 7: Documentation & Cleanup
**File**: `pkg/piano/README.md`
**Status**: ⏳ NOT STARTED

Documentation:
- [ ] API endpoints with MIDI upload examples
- [ ] Data models and schema (including MIDI blob storage)
- [ ] MIDI file format requirements
- [ ] Music theory question library structure
- [ ] Service business logic (accuracy, tempo, theory scoring)
- [ ] Sheet music SVG generation algorithm
- [ ] Testing approach and performance characteristics

Cleanup:
- [ ] Remove debug logging
- [ ] Run `go fmt`
- [ ] Run `go vet`
- [ ] Remove unused code
- [ ] MIDI library cleanup

**Commit**: `docs: Add piano package documentation and MIDI processing guide`

---

## Completion Criteria

### All Required
- [ ] All 7 files implemented (models, repo, service, router, templates, tests, docs)
- [ ] 55+ tests passing (higher count due to MIDI complexity)
- [ ] `go test ./pkg/piano` passes completely
- [ ] Zero compilation warnings
- [ ] All templates render correctly
- [ ] MIDI streaming functional
- [ ] Sheet music generation working
- [ ] API endpoints tested manually
- [ ] Documentation complete
- [ ] Code review ready

### Test Results Target
```
TestModels:        10+ passing
TestRepository:    15+ passing
TestService:       22+ passing (most complex)
TestRouter:        14+ passing
TestIntegration:   10+ passing
────────────────────────────────
TOTAL:           55+ passing
```

---

## Progress Timeline

| Week | Milestone | Status |
|------|-----------|--------|
| Week 1 | Models + Repo + Service | ⏳ Pending |
| Week 1-2 | Router + Templates + MIDI | ⏳ Pending |
| Week 2-3 | Integration Tests + Docs | ⏳ Pending |
| Week 3 | Code Review Ready | ⏳ Pending |

---

## Piano Phase Complexity Notes

This phase is the most complex of all app conversions due to:
1. **MIDI File Processing** - Binary file parsing and audio playback
2. **Music Theory** - Complex validation logic for musical concepts
3. **Real-time Playback** - Streaming MIDI recordings to browser
4. **SVG Sheet Music** - Generating music notation from MIDI
5. **Tempo & Accuracy Analysis** - Sophisticated timing calculations

Expected to take slightly longer than other phases (2.5-3 weeks).

---

## Next Steps

1. **BEGIN SUBTASK 1**: Create `pkg/piano/models.go`
2. Follow commit strategy (7 commits total)
3. Update this file as progress is made
4. Report blockers or issues immediately (especially MIDI library integration)

**Current Status**: Ready for implementation to start

---

**Last Updated**: 2026-02-20 10:34
**Manager**: manager_piano
**Branch**: feature/phase5-piano-app-0220
