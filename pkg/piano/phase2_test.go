package piano

import (
	"testing"
	"time"
)

// TestValidationModule tests the comprehensive validation module
func TestValidationModule(t *testing.T) {
	v := NewValidator()

	t.Run("ValidateSongInput", func(t *testing.T) {
		tests := []struct {
			name       string
			title      string
			composer   string
			bpm        int
			difficulty string
			wantErr    bool
		}{
			{"valid song", "Sonata", "Composer", 120, "advanced", false},
			{"empty title", "", "Composer", 120, "beginner", true},
			{"low BPM", "Song", "Composer", 20, "beginner", true},
			{"bad difficulty", "Song", "Composer", 120, "extreme", true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := v.ValidateSongInput(tt.title, tt.composer, tt.bpm, tt.difficulty)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateSongInput() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidateMIDIFile", func(t *testing.T) {
		validMIDI := []byte{0x4D, 0x54, 0x68, 0x64}
		tests := []struct {
			name    string
			data    []byte
			maxSize int64
			wantErr bool
		}{
			{"valid MIDI", validMIDI, 1024 * 1024, false},
			{"empty data", []byte{}, 1024 * 1024, true},
			{"invalid header", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 1024 * 1024, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := v.ValidateMIDIFile(tt.data, tt.maxSize)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateMIDIFile() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidatePracticeSessionInput", func(t *testing.T) {
		tests := []struct {
			name       string
			userID     uint
			songID     uint
			notesHit   int
			notesTotal int
			wantErr    bool
		}{
			{"valid session", 1, 1, 80, 100, false},
			{"notes exceed total", 1, 1, 150, 100, true},
			{"zero song ID", 1, 0, 80, 100, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := v.ValidatePracticeSessionInput(tt.userID, tt.songID, 0, 30.0, tt.notesHit, tt.notesTotal, 120.0)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidatePracticeSessionInput() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidatePagination", func(t *testing.T) {
		tests := []struct {
			name    string
			limit   int
			offset  int
			wantErr bool
		}{
			{"valid", 20, 0, false},
			{"limit zero", 0, 0, true},
			{"limit too high", 2000, 0, true},
			{"negative offset", 20, -1, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := v.ValidatePagination(tt.limit, tt.offset)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidatePagination() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// TestMIDIServicePhase2 tests MIDI service features from Phase 2
func TestMIDIServicePhase2(t *testing.T) {
	ms := NewMIDIService()

	t.Run("ValidateMIDI", func(t *testing.T) {
		validMIDI := []byte{0x4D, 0x54, 0x68, 0x64, 0x00, 0x00, 0x00, 0x06}
		invalidMIDI := []byte{0xFF, 0xFF, 0xFF, 0xFF}

		if err := ms.ValidateMIDI(validMIDI); err != nil {
			t.Errorf("ValidateMIDI(valid) failed: %v", err)
		}

		if err := ms.ValidateMIDI(invalidMIDI); err == nil {
			t.Error("ValidateMIDI(invalid) should have failed")
		}
	})

	t.Run("CountNotes", func(t *testing.T) {
		midiData := []byte{0x4D, 0x54, 0x68, 0x64} // Valid header
		count, err := ms.CountNotes(midiData)
		if err != nil {
			t.Errorf("CountNotes() failed: %v", err)
		}
		if count < 0 {
			t.Errorf("CountNotes() returned negative count: %d", count)
		}
	})

	t.Run("HexConversion", func(t *testing.T) {
		originalData := []byte{0x4D, 0x54, 0x68, 0x64}
		hexData := ms.ConvertToHex(originalData)
		if hexData == "" {
			t.Error("ConvertToHex() returned empty string")
		}

		decodedData, err := ms.ConvertFromHex(hexData)
		if err != nil {
			t.Errorf("ConvertFromHex() failed: %v", err)
		}

		if string(originalData) != string(decodedData) {
			t.Error("Hex round-trip failed")
		}
	})

	t.Run("RecordingSession", func(t *testing.T) {
		session := ms.StartRecording()
		if session == nil {
			t.Fatal("StartRecording() returned nil")
		}

		// Add notes
		for i := 0; i < 5; i++ {
			if err := ms.AddNoteToRecording(session, 60+i, 100, 0.5); err != nil {
				t.Errorf("AddNoteToRecording() failed: %v", err)
			}
		}

		if session.NotesRecorded != 5 {
			t.Errorf("Expected 5 notes, got %d", session.NotesRecorded)
		}

		// Simulate some time passing
		time.Sleep(10 * time.Millisecond)

		_, err := ms.FinishRecording(session)
		if err != nil {
			t.Errorf("FinishRecording() failed: %v", err)
		}

		if session.Duration < 0.01 {
			t.Error("Duration not set after finishing")
		}
	})
}

// TestAuthModule tests authentication features from Phase 2
func TestAuthModule(t *testing.T) {
	t.Run("AuthErrors", func(t *testing.T) {
		err := ErrNotAuthenticated
		if err == nil {
			t.Fatal("ErrNotAuthenticated is nil")
		}
		if err.Code != 401 {
			t.Errorf("Expected code 401, got %d", err.Code)
		}
	})

	t.Run("PianoError", func(t *testing.T) {
		pe := NewPianoError(400, "test error")
		if pe.Code != 400 {
			t.Errorf("Expected code 400, got %d", pe.Code)
		}
		if pe.Message != "test error" {
			t.Errorf("Expected 'test error', got '%s'", pe.Message)
		}
		if pe.Error() != "test error" {
			t.Errorf("Error() should return message")
		}
	})
}

// TestRequestValidation tests request struct validation
func TestRequestValidation(t *testing.T) {
	v := NewValidator()

	t.Run("CreateSongRequest", func(t *testing.T) {
		tests := []struct {
			name    string
			req     *CreateSongRequest
			wantErr bool
		}{
			{
				"valid request",
				&CreateSongRequest{
					Title:         "Test",
					Composer:      "Test",
					Difficulty:    "beginner",
					BPM:           120,
					TimeSignature: "4/4",
					KeySignature:  "C Major",
					TotalNotes:    100,
				},
				false,
			},
			{
				"missing title",
				&CreateSongRequest{
					Title:         "",
					Composer:      "Test",
					Difficulty:    "beginner",
					BPM:           120,
					TimeSignature: "4/4",
					KeySignature:  "C Major",
					TotalNotes:    100,
				},
				true,
			},
			{
				"invalid notes",
				&CreateSongRequest{
					Title:         "Test",
					Composer:      "Test",
					Difficulty:    "beginner",
					BPM:           120,
					TimeSignature: "4/4",
					KeySignature:  "C Major",
					TotalNotes:    20000,
				},
				true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.req.Validate(v)
				if (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("CreatePracticeRequest", func(t *testing.T) {
		tests := []struct {
			name    string
			req     *CreatePracticeRequest
			wantErr bool
		}{
			{
				"valid request",
				&CreatePracticeRequest{
					UserID:       1,
					SongID:       1,
					Duration:     30.0,
					NotesCorrect: 80,
					NotesTotal:   100,
					RecordedBPM:  120.0,
				},
				false,
			},
			{
				"notes exceed total",
				&CreatePracticeRequest{
					UserID:       1,
					SongID:       1,
					Duration:     30.0,
					NotesCorrect: 150,
					NotesTotal:   100,
					RecordedBPM:  120.0,
				},
				true,
			},
			{
				"zero user ID",
				&CreatePracticeRequest{
					UserID:       0,
					SongID:       1,
					Duration:     30.0,
					NotesCorrect: 80,
					NotesTotal:   100,
					RecordedBPM:  120.0,
				},
				true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.req.Validate(v)
				if (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// TestScoringCalculations tests the scoring algorithms
func TestScoringCalculations(t *testing.T) {
	s := NewService(nil)

	t.Run("CalculateAccuracy", func(t *testing.T) {
		tests := []struct {
			correct  int
			total    int
			expected float64
		}{
			{100, 100, 100.0},
			{50, 100, 50.0},
			{0, 100, 0.0},
			{0, 0, 0.0},
		}
		for _, tt := range tests {
			score := s.CalculateAccuracy(tt.correct, tt.total)
			if score != tt.expected {
				t.Errorf("CalculateAccuracy(%d, %d) = %f, want %f", tt.correct, tt.total, score, tt.expected)
			}
		}
	})

	t.Run("CalculateTempo", func(t *testing.T) {
		tests := []struct {
			recorded float64
			target   float64
			minScore float64 // Range due to rounding
			maxScore float64
		}{
			{120.0, 120.0, 100.0, 100.0},
			{110.0, 120.0, 90.0, 92.0},  // ~91.67
			{60.0, 120.0, 40.0, 60.0},   // 50% error = 50 score
		}
		for _, tt := range tests {
			score := s.CalculateTempo(tt.recorded, tt.target)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("CalculateTempo(%f, %f) = %f, want between %f and %f", tt.recorded, tt.target, score, tt.minScore, tt.maxScore)
			}
		}
	})

	t.Run("CalculateCompositeScore", func(t *testing.T) {
		// Weights: Accuracy 50%, Tempo 30%, Theory 20%
		score := s.CalculateCompositeScore(100, 100, 100)
		if score != 100.0 {
			t.Errorf("Expected 100.0, got %f", score)
		}

		score = s.CalculateCompositeScore(90, 80, 70)
		expected := (90 * 0.5) + (80 * 0.3) + (70 * 0.2)
		if score != expected {
			t.Errorf("Expected %f, got %f", expected, score)
		}
	})
}
