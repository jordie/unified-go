package piano

import (
	"fmt"
	"strings"
)

// Validator provides input validation for Piano app operations
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateSongInput validates song creation/update input
func (v *Validator) ValidateSongInput(title, composer string, bpm int, difficulty string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}

	if len(title) > 255 {
		return fmt.Errorf("title must be 255 characters or less")
	}

	if strings.TrimSpace(composer) == "" {
		return fmt.Errorf("composer is required")
	}

	if len(composer) > 255 {
		return fmt.Errorf("composer must be 255 characters or less")
	}

	if bpm < 30 || bpm > 300 {
		return fmt.Errorf("BPM must be between 30 and 300")
	}

	if !v.isValidDifficulty(difficulty) {
		return fmt.Errorf("difficulty must be one of: beginner, intermediate, advanced, expert")
	}

	return nil
}

// ValidateLessonInput validates lesson creation input
func (v *Validator) ValidateLessonInput(userID, songID uint, duration float64, notesCorrect, notesTotal int) error {
	if userID == 0 {
		return fmt.Errorf("user_id is required")
	}

	if songID == 0 {
		return fmt.Errorf("song_id is required")
	}

	if duration < 0 {
		return fmt.Errorf("duration must be non-negative")
	}

	if notesCorrect < 0 {
		return fmt.Errorf("notes_correct must be non-negative")
	}

	if notesTotal < 0 {
		return fmt.Errorf("notes_total must be non-negative")
	}

	if notesCorrect > notesTotal {
		return fmt.Errorf("notes_correct cannot exceed notes_total")
	}

	return nil
}

// ValidatePracticeSessionInput validates practice session input
func (v *Validator) ValidatePracticeSessionInput(userID, songID, lessonID uint, duration float64, notesHit, notesTotal int, tempoAvg float64) error {
	if userID == 0 {
		return fmt.Errorf("user_id is required")
	}

	if songID == 0 {
		return fmt.Errorf("song_id is required")
	}

	if duration < 0 {
		return fmt.Errorf("duration must be non-negative")
	}

	if notesHit < 0 {
		return fmt.Errorf("notes_hit must be non-negative")
	}

	if notesTotal < 0 {
		return fmt.Errorf("notes_total must be non-negative")
	}

	if notesHit > notesTotal {
		return fmt.Errorf("notes_hit cannot exceed notes_total")
	}

	if tempoAvg < 0 || tempoAvg > 500 {
		return fmt.Errorf("tempo_average must be between 0 and 500 BPM")
	}

	return nil
}

// ValidateScoreInput validates score input (0-100)
func (v *Validator) ValidateScoreInput(score float64) error {
	if score < 0 || score > 100 {
		return fmt.Errorf("score must be between 0 and 100")
	}
	return nil
}

// ValidatePagination validates pagination parameters
func (v *Validator) ValidatePagination(limit, offset int) error {
	if limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}

	if limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}

	if offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}

// ValidateMIDIFile validates MIDI file data
func (v *Validator) ValidateMIDIFile(data []byte, maxSizeBytes int64) error {
	if len(data) == 0 {
		return fmt.Errorf("MIDI file is empty")
	}

	if int64(len(data)) > maxSizeBytes {
		return fmt.Errorf("MIDI file is too large (max %d bytes)", maxSizeBytes)
	}

	// Check MIDI header signature
	if len(data) < 4 || data[0] != 0x4D || data[1] != 0x54 || data[2] != 0x68 || data[3] != 0x64 {
		return fmt.Errorf("invalid MIDI file format (must start with 'MThd')")
	}

	return nil
}

// ValidateUserID validates user ID
func (v *Validator) ValidateUserID(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("user_id must be positive")
	}
	return nil
}

// ValidateSongID validates song ID
func (v *Validator) ValidateSongID(songID uint) error {
	if songID == 0 {
		return fmt.Errorf("song_id is required")
	}
	return nil
}

// isValidDifficulty checks if difficulty is a valid value
func (v *Validator) isValidDifficulty(difficulty string) bool {
	validDifficulties := map[string]bool{
		"beginner":      true,
		"intermediate":  true,
		"advanced":      true,
		"expert":        true,
	}
	return validDifficulties[strings.ToLower(difficulty)]
}

// ValidateCreateSongRequest validates the complete create song request
type CreateSongRequest struct {
	Title         string `json:"title"`
	Composer      string `json:"composer"`
	Description   string `json:"description"`
	Difficulty    string `json:"difficulty"`
	BPM           int    `json:"bpm"`
	TimeSignature string `json:"time_signature"`
	KeySignature  string `json:"key_signature"`
	TotalNotes    int    `json:"total_notes"`
}

// Validate validates the entire request
func (csr *CreateSongRequest) Validate(v *Validator) error {
	if err := v.ValidateSongInput(csr.Title, csr.Composer, csr.BPM, csr.Difficulty); err != nil {
		return err
	}

	if csr.TotalNotes < 1 {
		return fmt.Errorf("total_notes must be at least 1")
	}

	if csr.TotalNotes > 10000 {
		return fmt.Errorf("total_notes cannot exceed 10000")
	}

	if strings.TrimSpace(csr.TimeSignature) == "" {
		return fmt.Errorf("time_signature is required")
	}

	if len(csr.TimeSignature) > 10 {
		return fmt.Errorf("time_signature must be 10 characters or less")
	}

	if strings.TrimSpace(csr.KeySignature) == "" {
		return fmt.Errorf("key_signature is required")
	}

	if len(csr.KeySignature) > 20 {
		return fmt.Errorf("key_signature must be 20 characters or less")
	}

	return nil
}

// CreatePracticeRequest represents a practice session creation request
type CreatePracticeRequest struct {
	UserID      int     `json:"user_id"`
	SongID      uint    `json:"song_id"`
	Duration    float64 `json:"duration"`
	NotesCorrect int    `json:"notes_correct"`
	NotesTotal  int     `json:"notes_total"`
	RecordedBPM float64 `json:"recorded_bpm"`
}

// Validate validates the entire request
func (cpr *CreatePracticeRequest) Validate(v *Validator) error {
	if err := v.ValidateUserID(cpr.UserID); err != nil {
		return err
	}

	if err := v.ValidateSongID(cpr.SongID); err != nil {
		return err
	}

	if cpr.Duration < 0 {
		return fmt.Errorf("duration must be non-negative")
	}

	if cpr.RecordedBPM < 0 || cpr.RecordedBPM > 500 {
		return fmt.Errorf("recorded_bpm must be between 0 and 500")
	}

	if err := v.ValidatePracticeSessionInput(uint(cpr.UserID), cpr.SongID, 0, cpr.Duration, cpr.NotesCorrect, cpr.NotesTotal, cpr.RecordedBPM); err != nil {
		return err
	}

	return nil
}
