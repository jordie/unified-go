package piano

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupServiceTestDB creates a test database with a service
func setupServiceTestDB(t *testing.T) (*sql.DB, *Service) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo)
	return db, service
}

// createTestMIDI returns a valid MIDI file header for testing
func createTestMIDI() []byte {
	// Valid MIDI file header: MThd (0x4D546864) followed by header data
	midi := []byte{0x4D, 0x54, 0x68, 0x64} // "MThd"
	midi = append(midi, 0x00, 0x00, 0x00, 0x06) // Header length (6 bytes)
	midi = append(midi, 0x00, 0x00) // Format type 0
	midi = append(midi, 0x00, 0x01) // Number of tracks (1)
	midi = append(midi, 0x00, 0x60) // Division (96 ticks per quarter note)
	return midi
}

// TestCalculateTempo tests tempo accuracy calculation
func TestCalculateTempo(t *testing.T) {
	_, service := setupServiceTestDB(t)

	tests := []struct {
		name                  string
		recordedBPM           float64
		targetBPM             float64
		expectedMinTempo      float64
		expectedMaxTempo      float64
	}{
		{
			name:                  "perfect tempo",
			recordedBPM:           120.0,
			targetBPM:             120.0,
			expectedMinTempo:      99.0,
			expectedMaxTempo:      100.0,
		},
		{
			name:                  "5% slower",
			recordedBPM:           114.0,
			targetBPM:             120.0,
			expectedMinTempo:      94.0,
			expectedMaxTempo:      96.0,
		},
		{
			name:                  "10% faster",
			recordedBPM:           132.0,
			targetBPM:             120.0,
			expectedMinTempo:      89.0,
			expectedMaxTempo:      91.0,
		},
		{
			name:                  "zero target BPM",
			recordedBPM:           120.0,
			targetBPM:             0.0,
			expectedMinTempo:      0.0,
			expectedMaxTempo:      0.0,
		},
		{
			name:                  "zero recorded BPM",
			recordedBPM:           0.0,
			targetBPM:             120.0,
			expectedMinTempo:      0.0,
			expectedMaxTempo:      0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempo := service.CalculateTempo(tt.recordedBPM, tt.targetBPM)
			if tempo < tt.expectedMinTempo || tempo > tt.expectedMaxTempo {
				t.Errorf("CalculateTempo(%.1f, %.1f) = %v, expected between %v and %v",
					tt.recordedBPM, tt.targetBPM, tempo, tt.expectedMinTempo, tt.expectedMaxTempo)
			}
		})
	}
}

// TestProcessLesson tests complete lesson processing
func TestProcessLesson(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// First create a song to use in the lesson
	repo := NewRepository(db)
	song := &Song{
		Title:         "Test Song",
		Composer:      "Test Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}
	songID, err := repo.SaveSong(ctx, song)
	if err != nil {
		t.Fatalf("SaveSong() error = %v", err)
	}

	session, err := service.ProcessLesson(ctx, 1, songID, 120.0, 30.0, 85, 100)
	if err != nil {
		t.Fatalf("ProcessLesson() error = %v", err)
	}

	if session.ID == 0 {
		t.Error("ProcessLesson() returned zero ID")
	}

	if session.UserID != 1 {
		t.Errorf("ProcessLesson() UserID = %v, want 1", session.UserID)
	}

	if session.NotesHit < 0 || session.NotesHit > session.NotesTotal {
		t.Errorf("ProcessLesson() NotesHit = %v, want between 0 and %d", session.NotesHit, session.NotesTotal)
	}

	if session.Duration <= 0 {
		t.Errorf("ProcessLesson() Duration = %v, want > 0", session.Duration)
	}
}

// TestProcessLessonInvalid tests invalid lesson processing
func TestProcessLessonInvalid(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	tests := []struct {
		name        string
		userID      uint
		songID      uint
		recordedBPM float64
		duration    float64
		notesCorrect int
		notesTotal  int
		wantErr     bool
	}{
		{"missing user_id", 0, 1, 120, 30, 85, 100, true},
		{"missing song_id", 1, 0, 120, 30, 85, 100, true},
		{"zero duration", 1, 1, 120, 0, 85, 100, true},
		{"negative notes_correct", 1, 1, 120, 30, -1, 100, true},
		{"zero notes_total", 1, 1, 120, 30, 50, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ProcessLesson(ctx, tt.userID, tt.songID, tt.recordedBPM, tt.duration, tt.notesCorrect, tt.notesTotal)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessLesson() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateLesson tests lesson recommendation generation
func TestGenerateLesson(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test song first
	repo := NewRepository(db)
	song := &Song{
		Title:         "Test Song",
		Composer:      "Test Composer",
		Difficulty:    "beginner",
		BPM:           100,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}
	_, err := repo.SaveSong(ctx, song)
	if err != nil {
		t.Fatalf("SaveSong() error = %v", err)
	}

	// Generate a lesson
	lesson, err := service.GenerateLesson(ctx, 1, "beginner")
	if err != nil {
		t.Fatalf("GenerateLesson() error = %v", err)
	}

	if lesson == nil {
		t.Error("GenerateLesson() returned nil")
	} else if lesson.Title == "" {
		t.Error("GenerateLesson() returned empty song title")
	}
}

// TestGetProgressionPath tests learning path determination
func TestGetProgressionPath(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	path, err := service.GetProgressionPath(ctx, 1)
	if err != nil {
		t.Fatalf("GetProgressionPath() error = %v", err)
	}

	if len(path) == 0 {
		t.Error("GetProgressionPath() returned empty path")
	}
}

// TestEvaluatePerformance tests performance evaluation
func TestEvaluatePerformance(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a song and lesson to evaluate
	repo := NewRepository(db)
	song := &Song{
		Title:         "Test Song",
		Composer:      "Test Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}
	songID, err := repo.SaveSong(ctx, song)
	if err != nil {
		t.Fatalf("SaveSong() error = %v", err)
	}

	// Process a lesson
	_, err = service.ProcessLesson(ctx, 1, songID, 120.0, 30.0, 85, 100)
	if err != nil {
		t.Fatalf("ProcessLesson() error = %v", err)
	}

	eval, err := service.EvaluatePerformance(ctx, 1)
	if err != nil {
		t.Fatalf("EvaluatePerformance() error = %v", err)
	}

	if eval["user_id"] != uint(1) {
		t.Errorf("EvaluatePerformance() user_id = %v, want 1", eval["user_id"])
	}

	if eval["total_lessons"] == nil {
		t.Error("EvaluatePerformance() missing total_lessons")
	}

	if eval["current_level"] == nil {
		t.Error("EvaluatePerformance() missing current_level")
	}
}

// TestGenerateMusicTheoryQuiz tests theory quiz generation
func TestGenerateMusicTheoryQuiz(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// GenerateMusicTheoryQuiz retrieves questions from repository
	// No setup needed - repository provides default questions

	quizzes, err := service.GenerateMusicTheoryQuiz(ctx, "beginner", 3)
	if err != nil {
		t.Fatalf("GenerateMusicTheoryQuiz() error = %v", err)
	}

	if len(quizzes) == 0 {
		t.Error("GenerateMusicTheoryQuiz() returned empty quiz list")
	}

	if len(quizzes) > 3 {
		t.Errorf("GenerateMusicTheoryQuiz() returned %d questions, want at most 3", len(quizzes))
	}
}

// TestAnalyzeMusicTheory tests music theory analysis
func TestAnalyzeMusicTheory(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	analysis, err := service.AnalyzeMusicTheory(ctx, 1)
	if err != nil {
		t.Fatalf("AnalyzeMusicTheory() error = %v", err)
	}

	if analysis["total_questions"] == nil {
		t.Error("AnalyzeMusicTheory() missing total_questions")
	}

	if analysis["correct_answers"] == nil {
		t.Error("AnalyzeMusicTheory() missing correct_answers")
	}
}

// TestGetUserMetrics tests comprehensive metrics retrieval
func TestGetUserMetrics(t *testing.T) {
	db, service := setupServiceTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a song and process a lesson
	repo := NewRepository(db)
	song := &Song{
		Title:         "Test Song",
		Composer:      "Test Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}
	songID, err := repo.SaveSong(ctx, song)
	if err != nil {
		t.Fatalf("SaveSong() error = %v", err)
	}

	_, err = service.ProcessLesson(ctx, 1, songID, 120.0, 30.0, 85, 100)
	if err != nil {
		t.Fatalf("ProcessLesson() error = %v", err)
	}

	metrics, err := service.GetUserMetrics(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserMetrics() error = %v", err)
	}

	if metrics["user_id"] != uint(1) {
		t.Errorf("GetUserMetrics() user_id = %v, want 1", metrics["user_id"])
	}

	if metrics["total_lessons"] == nil {
		t.Error("GetUserMetrics() missing total_lessons")
	}

	if metrics["current_level"] == nil {
		t.Error("GetUserMetrics() missing current_level")
	}
}

// BenchmarkCalculateAccuracy benchmarks accuracy calculation
func BenchmarkCalculateAccuracy(b *testing.B) {
	_, service := setupServiceTestDB(&testing.T{})

	notesCorrect := 85
	notesTotal := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateAccuracy(notesCorrect, notesTotal)
	}
}

// BenchmarkCalculateTempo benchmarks tempo calculation
func BenchmarkCalculateTempo(b *testing.B) {
	_, service := setupServiceTestDB(&testing.T{})

	recordedBPM := 118.5
	targetBPM := 120.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateTempo(recordedBPM, targetBPM)
	}
}

// BenchmarkCalculateCompositeScore benchmarks composite score calculation
func BenchmarkCalculateCompositeScore(b *testing.B) {
	_, service := setupServiceTestDB(&testing.T{})

	accuracy := 85.0
	tempo := 92.0
	theory := 78.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateCompositeScore(accuracy, tempo, theory)
	}
}

// BenchmarkProcessLesson benchmarks lesson processing
func BenchmarkProcessLesson(b *testing.B) {
	db, service := setupServiceTestDB(&testing.T{})
	defer db.Close()

	ctx := context.Background()

	// Create a song
	repo := NewRepository(db)
	song := &Song{
		Title:         "Test Song",
		Composer:      "Test Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}
	songID, _ := repo.SaveSong(ctx, song)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ProcessLesson(ctx, 1, songID, 120.0, 30.0, 85, 100)
	}
}

// BenchmarkEvaluatePerformance benchmarks performance evaluation
func BenchmarkEvaluatePerformance(b *testing.B) {
	db, service := setupServiceTestDB(&testing.T{})
	defer db.Close()

	ctx := context.Background()

	// Create a song and lesson
	repo := NewRepository(db)
	song := &Song{
		Title:         "Test Song",
		Composer:      "Test Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}
	songID, _ := repo.SaveSong(ctx, song)
	service.ProcessLesson(ctx, 1, songID, 120.0, 30.0, 85, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.EvaluatePerformance(ctx, 1)
	}
}
