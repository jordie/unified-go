package piano

import (
	"encoding/json"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Song represents a piano song/piece available for practice
type Song struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title" gorm:"index"`
	Composer      string    `json:"composer"`
	Description   string    `json:"description"`
	MIDIFile      []byte    `json:"midi_file" gorm:"type:BLOB"` // Binary MIDI data
	Difficulty    string    `json:"difficulty"`  // beginner, intermediate, advanced, expert
	Duration      float64   `json:"duration"`    // Seconds
	BPM           int       `json:"bpm"`         // Beats per minute (40-300)
	TimeSignature string    `json:"time_signature"` // e.g., "4/4", "3/4"
	KeySignature  string    `json:"key_signature"`  // e.g., "C major", "G minor"
	TotalNotes    int       `json:"total_notes"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PianoLesson represents a lesson session where a user practices a song
type PianoLesson struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"index"`
	SongID         uint      `json:"song_id" gorm:"index"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Duration       float64   `json:"duration"` // Seconds
	NotesCorrect   int       `json:"notes_correct"`
	NotesTotal     int       `json:"notes_total"`
	Accuracy       float64   `json:"accuracy"` // 0-100 (notes_correct / notes_total * 100)
	TempoAccuracy  float64   `json:"tempo_accuracy"` // 0-100
	Score          float64   `json:"score"` // 0-100 (composite score)
	Completed      bool      `json:"completed"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PracticeSession represents a recorded practice session with MIDI data
type PracticeSession struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"index"`
	SongID        uint      `json:"song_id" gorm:"index"`
	LessonID      uint      `json:"lesson_id" gorm:"index"`
	RecordingMIDI []byte    `json:"recording_midi" gorm:"type:BLOB"` // Binary MIDI recording
	Duration      float64   `json:"duration"` // Seconds
	NotesHit      int       `json:"notes_hit"`
	NotesTotal    int       `json:"notes_total"`
	TempoAverage  float64   `json:"tempo_average"` // BPM
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserProgress represents aggregated progress for a user
type UserProgress struct {
	UserID                  uint    `json:"user_id"`
	TotalLessonsCompleted   int     `json:"total_lessons_completed"`
	TotalPracticedMinutes   float64 `json:"total_practiced_minutes"`
	AverageScore            float64 `json:"average_score"`
	BestScore               float64 `json:"best_score"`
	FastestTempo            float64 `json:"fastest_tempo"` // BPM
	BestDifficulty          string  `json:"best_difficulty"`
	TotalSongsMastered      int     `json:"total_songs_mastered"`
	CurrentLevel            string  `json:"current_level"` // beginner, intermediate, advanced, expert
	EstimatedLevel          string  `json:"estimated_level"`
	LastPracticedDate       *time.Time `json:"last_practiced_date"`
}

// MusicTheoryQuiz represents a music theory quiz
type MusicTheoryQuiz struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index"`
	LessonID   uint      `json:"lesson_id" gorm:"index"`
	Topic      string    `json:"topic"` // scales, intervals, chords, rhythm, etc.
	Questions  string    `json:"questions" gorm:"type:TEXT"` // JSON encoded
	Answers    string    `json:"answers" gorm:"type:TEXT"` // JSON encoded user answers
	Score      float64   `json:"score"` // 0-100
	Difficulty string    `json:"difficulty"` // beginner, intermediate, advanced
	Completed  bool      `json:"completed"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// MusicQuestion represents a single music theory question
type MusicQuestion struct {
	ID      uint     `json:"id"`
	Question string   `json:"question"`
	Options []string `json:"options"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation string `json:"explanation"`
}

// Validate performs validation on Song
func (s *Song) Validate() error {
	if s.Title == "" {
		return errors.New("title is required")
	}
	if s.Composer == "" {
		return errors.New("composer is required")
	}
	if len(s.MIDIFile) == 0 {
		return errors.New("midi_file is required")
	}
	// Validate MIDI header (should start with MThd: 0x4D546864)
	if len(s.MIDIFile) < 4 || s.MIDIFile[0] != 0x4D || s.MIDIFile[1] != 0x54 {
		return errors.New("invalid MIDI file format (must start with 'MThd')")
	}
	if s.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	if s.BPM < 40 || s.BPM > 300 {
		return fmt.Errorf("bpm must be between 40 and 300, got %d", s.BPM)
	}
	if s.Difficulty == "" {
		s.Difficulty = "intermediate"
	}
	validDifficulties := map[string]bool{"beginner": true, "intermediate": true, "advanced": true, "expert": true}
	if !validDifficulties[s.Difficulty] {
		return fmt.Errorf("invalid difficulty: %s", s.Difficulty)
	}
	return nil
}

// Validate performs validation on PianoLesson
func (pl *PianoLesson) Validate() error {
	if pl.UserID == 0 {
		return errors.New("user_id is required")
	}
	if pl.SongID == 0 {
		return errors.New("song_id is required")
	}
	if pl.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	if pl.NotesTotal <= 0 {
		return errors.New("notes_total must be positive")
	}
	if pl.NotesCorrect < 0 || pl.NotesCorrect > pl.NotesTotal {
		return fmt.Errorf("notes_correct must be between 0 and %d, got %d", pl.NotesTotal, pl.NotesCorrect)
	}
	if pl.Accuracy < 0 || pl.Accuracy > 100 {
		return fmt.Errorf("accuracy must be between 0 and 100, got %f", pl.Accuracy)
	}
	if pl.TempoAccuracy < 0 || pl.TempoAccuracy > 100 {
		return fmt.Errorf("tempo_accuracy must be between 0 and 100, got %f", pl.TempoAccuracy)
	}
	if pl.Score < 0 || pl.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", pl.Score)
	}
	return nil
}

// Validate performs validation on PracticeSession
func (ps *PracticeSession) Validate() error {
	if ps.UserID == 0 {
		return errors.New("user_id is required")
	}
	if ps.SongID == 0 {
		return errors.New("song_id is required")
	}
	if len(ps.RecordingMIDI) == 0 {
		return errors.New("recording_midi is required")
	}
	// Validate MIDI header
	if len(ps.RecordingMIDI) < 4 || ps.RecordingMIDI[0] != 0x4D || ps.RecordingMIDI[1] != 0x54 {
		return errors.New("invalid MIDI recording format")
	}
	if ps.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	if ps.NotesTotal <= 0 {
		return errors.New("notes_total must be positive")
	}
	return nil
}

// CalculateAccuracy calculates accuracy percentage from correct/total notes
func CalculateAccuracy(notesCorrect, notesTotal int) float64 {
	if notesTotal <= 0 {
		return 0
	}
	accuracy := (float64(notesCorrect) / float64(notesTotal)) * 100
	if accuracy > 100 {
		accuracy = 100
	}
	if accuracy < 0 {
		accuracy = 0
	}
	return accuracy
}

// CalculateTempoAccuracy calculates tempo accuracy percentage
func CalculateTempoAccuracy(actualTempo, targetTempo float64) float64 {
	if targetTempo <= 0 {
		return 0
	}
	difference := abs(actualTempo - targetTempo)
	accuracy := (1.0 - (difference / targetTempo)) * 100
	if accuracy > 100 {
		accuracy = 100
	}
	if accuracy < 0 {
		accuracy = 0
	}
	return accuracy
}

// CalculateCompositeScore calculates overall lesson score
// Formula: (accuracy * 0.7) + (tempo_accuracy * 0.2) + (theory_score * 0.1)
func CalculateCompositeScore(accuracy, tempoAccuracy, theoryScore float64) float64 {
	score := (accuracy * 0.7) + (tempoAccuracy * 0.2) + (theoryScore * 0.1)
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

// EstimatePianoLevel estimates piano skill level from average score
func EstimatePianoLevel(averageScore float64) string {
	switch {
	case averageScore < 50:
		return "beginner"
	case averageScore < 70:
		return "intermediate"
	case averageScore < 85:
		return "advanced"
	default:
		return "expert"
	}
}

// MarshalMIDI returns MIDI data as hex string for JSON
func (s *Song) MarshalMIDI() string {
	if len(s.MIDIFile) == 0 {
		return ""
	}
	return hex.EncodeToString(s.MIDIFile)
}

// UnmarshalMIDI decodes hex-encoded MIDI data
func (s *Song) UnmarshalMIDI(hexData string) error {
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return err
	}
	s.MIDIFile = data
	return nil
}

// MarshalJSON implements custom JSON marshaling for Song
func (s *Song) MarshalJSON() ([]byte, error) {
	type Alias Song
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		MIDIFile  string `json:"midi_file"` // Hex encoded
	}{
		Alias:     (*Alias)(s),
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
		MIDIFile:  hex.EncodeToString(s.MIDIFile),
	})
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
