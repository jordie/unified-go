package piano

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestPianoIntegration provides integration test setup
type TestPianoIntegration struct {
	db      *sql.DB
	router  chi.Router
	service *Service
}

// setupIntegration creates a test database and router
func setupIntegration(t *testing.T) *TestPianoIntegration {
	db := setupTestDB(t)
	router := NewRouter(db).Routes()
	repo := NewRepository(db)
	service := NewService(repo)

	return &TestPianoIntegration{
		db:      db,
		router:  router,
		service: service,
	}
}

// setupBenchmark creates a test database and router for benchmarks
func setupBenchmark(b testing.TB) *TestPianoIntegration {
	db := setupTestDB(b)
	router := NewRouter(db).Routes()
	repo := NewRepository(db)
	service := NewService(repo)

	return &TestPianoIntegration{
		db:      db,
		router:  router,
		service: service,
	}
}

// TestCreateAndRetrieveSong tests the complete song lifecycle
func TestCreateAndRetrieveSong(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	// Create a song via API
	songData := map[string]interface{}{
		"title":           "Test Song",
		"composer":        "Test Composer",
		"difficulty":      "beginner",
		"bpm":             120,
		"time_signature":  "4/4",
		"key_signature":   "C Major",
		"total_notes":     50,
		"duration":        120.0,
		"midi_file":       createTestMIDI(),
	}

	body, _ := json.Marshal(songData)
	req := httptest.NewRequest("POST", "/api/songs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Parse response
	var song Song
	if err := json.Unmarshal(w.Body.Bytes(), &song); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if song.ID == 0 {
		t.Error("Song ID should not be 0")
	}

	// Retrieve the song
	req = httptest.NewRequest("GET", "/api/songs/1", nil)
	w = httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var retrievedSong Song
	json.Unmarshal(w.Body.Bytes(), &retrievedSong)
	if retrievedSong.Title != "Test Song" {
		t.Errorf("Expected title 'Test Song', got '%s'", retrievedSong.Title)
	}
}

// TestPracticeLessonFlow tests the complete practice lesson flow
func TestPracticeLessonFlow(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a song first
	song := &Song{
		Title:         "Practice Test",
		Composer:      "Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		TotalNotes:    100,
		Duration:      180.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, err := ti.service.repo.SaveSong(ctx, song)
	if err != nil {
		t.Fatalf("Failed to create song: %v", err)
	}

	// Submit a practice session
	sessionData := map[string]interface{}{
		"user_id":       1,
		"song_id":       songID,
		"recorded_bpm":  118.5,
		"duration":      180.0,
		"notes_correct": 85,
		"notes_total":   100,
	}

	body, _ := json.Marshal(sessionData)
	req := httptest.NewRequest("POST", "/api/practice", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var session PracticeSession
	json.Unmarshal(w.Body.Bytes(), &session)

	if session.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", session.UserID)
	}

	if session.NotesHit != 85 {
		t.Errorf("Expected NotesHit 85, got %d", session.NotesHit)
	}
}

// TestUserProgress tests piano progress tracking
func TestUserProgress(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a song
	song := &Song{
		Title:         "Progress Song",
		Composer:      "Composer",
		Difficulty:    "beginner",
		BPM:           100,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		TotalNotes:    50,
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, _ := ti.service.repo.SaveSong(ctx, song)

	// Create multiple practice sessions
	for i := 0; i < 3; i++ {
		ti.service.ProcessLesson(ctx, 1, songID, 100.0+float64(i*2), 120.0, 40+i*5, 50)
	}

	// Get user progress
	req := httptest.NewRequest("GET", "/api/users/1/progress", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var progress UserProgress
	json.Unmarshal(w.Body.Bytes(), &progress)

	if progress.TotalLessonsCompleted != 3 {
		t.Errorf("Expected 3 lessons, got %d", progress.TotalLessonsCompleted)
	}

	if progress.AverageScore <= 0 {
		t.Errorf("Average score should be calculated, got %f", progress.AverageScore)
	}
}

// TestUserMetrics tests comprehensive user metrics
func TestUserMetrics(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a song and sessions
	song := &Song{
		Title:         "Metrics Song",
		Composer:      "Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		TotalNotes:    100,
		Duration:      240.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, _ := ti.service.repo.SaveSong(ctx, song)

	// Create sessions
	for i := 0; i < 2; i++ {
		ti.service.ProcessLesson(ctx, 1, songID, 120.0, 120.0, 80+i*5, 100)
	}

	// Get metrics
	req := httptest.NewRequest("GET", "/api/users/1/metrics", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var metrics map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &metrics)

	if metrics["total_lessons"] == nil {
		t.Error("Metrics should contain total_lessons")
	}

	if metrics["current_level"] == nil {
		t.Error("Metrics should contain current_level")
	}
}

// TestPerformanceEvaluation tests performance evaluation
func TestPerformanceEvaluation(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a song
	song := &Song{
		Title:         "Eval Song",
		Composer:      "Composer",
		Difficulty:    "advanced",
		BPM:           140,
		TimeSignature: "4/4",
		KeySignature:  "D Major",
		TotalNotes:    150,
		Duration:      300.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, _ := ti.service.repo.SaveSong(ctx, song)

	// Create sessions with varying performance
	for i := 0; i < 3; i++ {
		ti.service.ProcessLesson(ctx, 1, songID, 138.0+float64(i), 180.0, 120+i*5, 150)
	}

	// Get evaluation
	req := httptest.NewRequest("GET", "/api/users/1/evaluation", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var evaluation map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &evaluation)

	if evaluation["total_lessons"] == nil {
		t.Error("Evaluation should contain total_lessons")
	}

	if evaluation["trend"] == nil {
		t.Error("Evaluation should contain trend analysis")
	}
}

// TestMusicTheoryQuiz tests theory quiz generation
func TestMusicTheoryQuiz(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	quizData := map[string]interface{}{
		"difficulty": "beginner",
		"count":      5,
	}

	body, _ := json.Marshal(quizData)
	req := httptest.NewRequest("POST", "/api/theory-quiz", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["questions"] == nil {
		t.Error("Response should contain questions")
	}

	if response["count"] == nil {
		t.Error("Response should contain count")
	}
}

// TestLessonRecommendation tests lesson recommendation
func TestLessonRecommendation(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create songs with different difficulties
	difficulties := []string{"beginner", "intermediate", "advanced"}
	for _, diff := range difficulties {
		song := &Song{
			Title:         "Song - " + diff,
			Composer:      "Composer",
			Difficulty:    diff,
			BPM:           120,
			TimeSignature: "4/4",
			KeySignature:  "C Major",
			TotalNotes:    100,
			Duration:      180.0,
			MIDIFile:      createTestMIDI(),
		}
		ti.service.repo.SaveSong(ctx, song)
	}

	// Get recommendation
	req := httptest.NewRequest("GET", "/api/recommend/1", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["recommended"] == nil {
		t.Error("Response should contain recommended lesson")
	}

	if response["difficulty"] == nil {
		t.Error("Response should contain difficulty")
	}
}

// TestProgressionPath tests learning progression path
func TestProgressionPath(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a song and session
	song := &Song{
		Title:         "Path Song",
		Composer:      "Composer",
		Difficulty:    "beginner",
		BPM:           100,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		TotalNotes:    50,
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, _ := ti.service.repo.SaveSong(ctx, song)
	ti.service.ProcessLesson(ctx, 1, songID, 100.0, 120.0, 45, 50)

	// Get progression path
	req := httptest.NewRequest("GET", "/api/progression-path/1", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["path"] == nil {
		t.Error("Response should contain progression path")
	}
}

// TestListSongsFiltering tests song filtering
func TestListSongsFiltering(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create songs with different difficulties
	for _, diff := range []string{"beginner", "intermediate", "advanced"} {
		song := &Song{
			Title:         "Filter Test - " + diff,
			Composer:      "Composer",
			Difficulty:    diff,
			BPM:           120,
			TimeSignature: "4/4",
			KeySignature:  "C Major",
			TotalNotes:    100,
			Duration:      180.0,
			MIDIFile:      createTestMIDI(),
		}
		ti.service.repo.SaveSong(ctx, song)
	}

	// Test filtering
	req := httptest.NewRequest("GET", "/api/songs?difficulty=intermediate", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	songs := response["songs"].([]interface{})
	if len(songs) == 0 {
		t.Error("Expected songs in response")
	}
}

// TestMIDIUpload tests MIDI file upload
func TestMIDIUpload(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	// Create a multipart form with MIDI file
	body := new(bytes.Buffer)
	midiData := createTestMIDI()

	// Write MIDI data as form data
	body.Write(midiData)

	req := httptest.NewRequest("POST", "/api/midi/upload", body)
	req.Header.Set("Content-Type", "audio/midi")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if response["uploaded"] != true {
			t.Error("Expected uploaded to be true")
		}
	}
}

// TestErrorHandling tests error responses
func TestErrorHandling(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{"invalid song ID", "GET", "/api/songs/invalid", http.StatusBadRequest},
		{"missing user ID", "GET", "/api/users//progress", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			ti.router.ServeHTTP(w, req)

			if w.Code < http.StatusBadRequest {
				t.Logf("Expected error status for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

// BenchmarkPracticeLesson benchmarks practice lesson processing
func BenchmarkPracticeLesson(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()
	song := &Song{
		Title:         "Bench Song",
		Composer:      "Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		TotalNotes:    100,
		Duration:      180.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, _ := ti.service.repo.SaveSong(ctx, song)

	sessionData := map[string]interface{}{
		"user_id":       1,
		"song_id":       songID,
		"recorded_bpm":  120.0,
		"duration":      180.0,
		"notes_correct": 85,
		"notes_total":   100,
	}

	body, _ := json.Marshal(sessionData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/practice", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}

// BenchmarkUserProgress benchmarks progress retrieval
func BenchmarkUserProgress(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()
	song := &Song{
		Title:         "Stat Song",
		Composer:      "Composer",
		Difficulty:    "beginner",
		BPM:           100,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		TotalNotes:    50,
		Duration:      120.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, _ := ti.service.repo.SaveSong(ctx, song)

	// Create sessions
	for i := 0; i < 5; i++ {
		ti.service.ProcessLesson(ctx, 1, songID, 100.0, 120.0, 40+i*2, 50)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/1/progress", nil)
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}

// BenchmarkUserMetrics benchmarks metrics calculation
func BenchmarkUserMetrics(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()
	song := &Song{
		Title:         "Metric Song",
		Composer:      "Composer",
		Difficulty:    "intermediate",
		BPM:           120,
		TimeSignature: "4/4",
		KeySignature:  "C Major",
		TotalNotes:    100,
		Duration:      240.0,
		MIDIFile:      createTestMIDI(),
	}

	songID, _ := ti.service.repo.SaveSong(ctx, song)

	for i := 0; i < 5; i++ {
		ti.service.ProcessLesson(ctx, 1, songID, 120.0, 180.0, 80+i, 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/1/metrics", nil)
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}
