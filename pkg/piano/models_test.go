package piano

import (
	"testing"
	"time"
)

// TestSongValidation tests Song struct validation
func TestSongValidation(t *testing.T) {
	validMIDI := []byte{0x4D, 0x54, 0x68, 0x64} // "MThd" - valid MIDI header

	tests := []struct {
		name    string
		song    *Song
		wantErr bool
	}{
		{
			name: "valid song",
			song: &Song{
				Title:         "Moonlight Sonata",
				Composer:      "Beethoven",
				MIDIFile:      validMIDI,
				Difficulty:    "advanced",
				Duration:      600.0,
				BPM:           120,
				TimeSignature: "4/4",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			song: &Song{
				Title:    "",
				Composer: "Composer",
				MIDIFile: validMIDI,
				Duration: 600.0,
				BPM:      120,
			},
			wantErr: true,
		},
		{
			name: "missing composer",
			song: &Song{
				Title:    "Title",
				Composer: "",
				MIDIFile: validMIDI,
				Duration: 600.0,
				BPM:      120,
			},
			wantErr: true,
		},
		{
			name: "missing MIDI file",
			song: &Song{
				Title:     "Title",
				Composer:  "Composer",
				MIDIFile:  nil,
				Duration:  600.0,
				BPM:       120,
			},
			wantErr: true,
		},
		{
			name: "invalid MIDI header",
			song: &Song{
				Title:    "Title",
				Composer: "Composer",
				MIDIFile: []byte{0x00, 0x00, 0x00, 0x00},
				Duration: 600.0,
				BPM:      120,
			},
			wantErr: true,
		},
		{
			name: "BPM too low",
			song: &Song{
				Title:    "Title",
				Composer: "Composer",
				MIDIFile: validMIDI,
				Duration: 600.0,
				BPM:      20,
			},
			wantErr: true,
		},
		{
			name: "BPM too high",
			song: &Song{
				Title:    "Title",
				Composer: "Composer",
				MIDIFile: validMIDI,
				Duration: 600.0,
				BPM:      350,
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty",
			song: &Song{
				Title:      "Title",
				Composer:   "Composer",
				MIDIFile:   validMIDI,
				Duration:   600.0,
				BPM:        120,
				Difficulty: "impossible",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.song.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPianoLessonValidation tests PianoLesson struct validation
func TestPianoLessonValidation(t *testing.T) {
	tests := []struct {
		name    string
		lesson  *PianoLesson
		wantErr bool
	}{
		{
			name: "valid lesson",
			lesson: &PianoLesson{
				UserID:        1,
				SongID:        1,
				Duration:      300.0,
				NotesCorrect:  95,
				NotesTotal:    100,
				Accuracy:      95.0,
				TempoAccuracy: 88.5,
				Score:         90.0,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			lesson: &PianoLesson{
				UserID:      0,
				SongID:      1,
				Duration:    300.0,
				NotesTotal:  100,
			},
			wantErr: true,
		},
		{
			name: "missing song_id",
			lesson: &PianoLesson{
				UserID:     1,
				SongID:     0,
				Duration:   300.0,
				NotesTotal: 100,
			},
			wantErr: true,
		},
		{
			name: "zero duration",
			lesson: &PianoLesson{
				UserID:     1,
				SongID:     1,
				Duration:   0,
				NotesTotal: 100,
			},
			wantErr: true,
		},
		{
			name: "notes_correct exceeds total",
			lesson: &PianoLesson{
				UserID:       1,
				SongID:       1,
				Duration:     300.0,
				NotesCorrect: 150,
				NotesTotal:   100,
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy",
			lesson: &PianoLesson{
				UserID:        1,
				SongID:        1,
				Duration:      300.0,
				NotesCorrect:  95,
				NotesTotal:    100,
				Accuracy:      150.0,
				TempoAccuracy: 85.0,
			},
			wantErr: true,
		},
		{
			name: "invalid tempo_accuracy",
			lesson: &PianoLesson{
				UserID:        1,
				SongID:        1,
				Duration:      300.0,
				NotesCorrect:  95,
				NotesTotal:    100,
				Accuracy:      95.0,
				TempoAccuracy: -10.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lesson.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPracticeSessionValidation tests PracticeSession struct validation
func TestPracticeSessionValidation(t *testing.T) {
	validMIDI := []byte{0x4D, 0x54, 0x68, 0x64} // "MThd"

	tests := []struct {
		name    string
		session *PracticeSession
		wantErr bool
	}{
		{
			name: "valid practice session",
			session: &PracticeSession{
				UserID:        1,
				SongID:        1,
				RecordingMIDI: validMIDI,
				Duration:      300.0,
				NotesHit:      95,
				NotesTotal:    100,
				TempoAverage:  120.0,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			session: &PracticeSession{
				UserID:        0,
				SongID:        1,
				RecordingMIDI: validMIDI,
				Duration:      300.0,
				NotesTotal:    100,
			},
			wantErr: true,
		},
		{
			name: "missing song_id",
			session: &PracticeSession{
				UserID:        1,
				SongID:        0,
				RecordingMIDI: validMIDI,
				Duration:      300.0,
				NotesTotal:    100,
			},
			wantErr: true,
		},
		{
			name: "missing MIDI recording",
			session: &PracticeSession{
				UserID:        1,
				SongID:        1,
				RecordingMIDI: nil,
				Duration:      300.0,
				NotesTotal:    100,
			},
			wantErr: true,
		},
		{
			name: "invalid MIDI header",
			session: &PracticeSession{
				UserID:        1,
				SongID:        1,
				RecordingMIDI: []byte{0x00, 0x00},
				Duration:      300.0,
				NotesTotal:    100,
			},
			wantErr: true,
		},
		{
			name: "zero duration",
			session: &PracticeSession{
				UserID:        1,
				SongID:        1,
				RecordingMIDI: validMIDI,
				Duration:      0,
				NotesTotal:    100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCalculateAccuracy tests accuracy calculation
func TestCalculateAccuracy(t *testing.T) {
	tests := []struct {
		notesCorrect int
		notesTotal   int
		expected     float64
	}{
		{100, 100, 100.0},
		{95, 100, 95.0},
		{50, 100, 50.0},
		{0, 100, 0.0},
		{0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run("accuracy calculation", func(t *testing.T) {
			result := CalculateAccuracy(tt.notesCorrect, tt.notesTotal)
			if result != tt.expected {
				t.Errorf("CalculateAccuracy(%d, %d) = %f, want %f", tt.notesCorrect, tt.notesTotal, result, tt.expected)
			}
		})
	}
}

// TestCalculateTempoAccuracy tests tempo accuracy calculation
func TestCalculateTempoAccuracy(t *testing.T) {
	tests := []struct {
		actual   float64
		target   float64
		name     string
	}{
		{120.0, 120.0, "perfect tempo"},
		{100.0, 120.0, "slow tempo"},
		{140.0, 120.0, "fast tempo"},
		{0, 120.0, "zero actual tempo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTempoAccuracy(tt.actual, tt.target)
			if result < 0 || result > 100 {
				t.Errorf("CalculateTempoAccuracy(%f, %f) = %f, expected 0-100", tt.actual, tt.target, result)
			}
		})
	}
}

// TestCalculateCompositeScore tests composite score calculation
func TestCalculateCompositeScore(t *testing.T) {
	tests := []struct {
		accuracy      float64
		tempoAccuracy float64
		theoryScore   float64
		name          string
	}{
		{90.0, 85.0, 80.0, "good performance"},
		{100.0, 100.0, 100.0, "perfect performance"},
		{50.0, 50.0, 50.0, "average performance"},
		{0, 0, 0, "poor performance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCompositeScore(tt.accuracy, tt.tempoAccuracy, tt.theoryScore)
			if result < 0 || result > 100 {
				t.Errorf("CalculateCompositeScore(%f, %f, %f) = %f, expected 0-100",
					tt.accuracy, tt.tempoAccuracy, tt.theoryScore, result)
			}
		})
	}
}

// TestEstimatePianoLevel tests piano level estimation
func TestEstimatePianoLevel(t *testing.T) {
	tests := []struct {
		score         float64
		expectedLevel string
	}{
		{30.0, "beginner"},
		{60.0, "intermediate"},
		{75.0, "advanced"},
		{95.0, "expert"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedLevel, func(t *testing.T) {
			result := EstimatePianoLevel(tt.score)
			if result != tt.expectedLevel {
				t.Errorf("EstimatePianoLevel(%f) = %s, want %s", tt.score, result, tt.expectedLevel)
			}
		})
	}
}

// TestMIDIHexEncoding tests MIDI hex encoding/decoding
func TestMIDIHexEncoding(t *testing.T) {
	originalMIDI := []byte{0x4D, 0x54, 0x68, 0x64, 0x00, 0x00, 0x00, 0x06}

	song := &Song{
		Title:    "Test",
		Composer: "Test",
		MIDIFile: originalMIDI,
		Duration: 300.0,
		BPM:      120,
	}

	hexString := song.MarshalMIDI()
	if hexString == "" {
		t.Error("MarshalMIDI() returned empty string")
	}

	song2 := &Song{}
	err := song2.UnmarshalMIDI(hexString)
	if err != nil {
		t.Errorf("UnmarshalMIDI() error = %v", err)
	}

	if len(song2.MIDIFile) != len(originalMIDI) {
		t.Errorf("UnmarshalMIDI() length mismatch: %d vs %d", len(song2.MIDIFile), len(originalMIDI))
	}
}

// TestMIDIValidation tests MIDI file validation
func TestMIDIValidation(t *testing.T) {
	validMIDI := []byte{0x4D, 0x54, 0x68, 0x64}

	tests := []struct {
		name    string
		midi    []byte
		wantErr bool
	}{
		{"valid MIDI", validMIDI, false},
		{"invalid header", []byte{0x00, 0x00, 0x00, 0x00}, true},
		{"empty MIDI", []byte{}, true},
		{"partial MIDI", []byte{0x4D, 0x54}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			song := &Song{
				Title:    "Test",
				Composer: "Test",
				MIDIFile: tt.midi,
				Duration: 300.0,
				BPM:      120,
			}
			err := song.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMusicTheoryQuiz tests MusicTheoryQuiz struct
func TestMusicTheoryQuiz(t *testing.T) {
	quiz := &MusicTheoryQuiz{
		UserID:     1,
		Topic:      "scales",
		Score:      85.0,
		Difficulty: "intermediate",
		Completed:  true,
	}

	if quiz.UserID == 0 {
		t.Error("UserID should be set")
	}

	if quiz.Score < 0 || quiz.Score > 100 {
		t.Error("Score should be between 0-100")
	}
}

// TestUserProgress tests UserProgress struct
func TestUserProgress(t *testing.T) {
	now := time.Now()
	progress := &UserProgress{
		UserID:                  1,
		TotalLessonsCompleted:   10,
		TotalPracticedMinutes:   300.0,
		AverageScore:            82.5,
		BestScore:               95.0,
		FastestTempo:            140.0,
		BestDifficulty:          "advanced",
		CurrentLevel:            "intermediate",
		LastPracticedDate:       &now,
	}

	if progress.UserID == 0 {
		t.Error("UserID should be set")
	}

	if progress.TotalLessonsCompleted < 0 {
		t.Error("TotalLessonsCompleted should be non-negative")
	}

	if progress.AverageScore > progress.BestScore {
		t.Error("AverageScore should not exceed BestScore")
	}
}

// TestSongMarshalJSON tests JSON marshaling
func TestSongMarshalJSON(t *testing.T) {
	midi := []byte{0x4D, 0x54, 0x68, 0x64}
	song := &Song{
		ID:       1,
		Title:    "Test Song",
		Composer: "Test Composer",
		MIDIFile: midi,
		Duration: 300.0,
		BPM:      120,
		CreatedAt: time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC),
	}

	data, err := song.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalJSON() returned empty data")
	}

	jsonStr := string(data)
	if jsonStr == "" {
		t.Error("JSON string should not be empty")
	}
}
