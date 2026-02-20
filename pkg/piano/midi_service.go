package piano

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"
)

// MIDIService handles MIDI file operations for playback and recording
type MIDIService struct{}

// NewMIDIService creates a new MIDI service
func NewMIDIService() *MIDIService {
	return &MIDIService{}
}

// MIDINote represents a single MIDI note event
type MIDINote struct {
	Note     int       // MIDI note number (0-127)
	Velocity int       // Note velocity (0-127)
	Duration float64   // Duration in milliseconds
	StartTime float64  // Start time in milliseconds from beginning
}

// ValidateMIDI checks if the MIDI data is valid
func (ms *MIDIService) ValidateMIDI(data []byte) error {
	if len(data) < 4 {
		return errors.New("MIDI data too short")
	}

	// Check MIDI header signature "MThd" (0x4D546864)
	if data[0] != 0x4D || data[1] != 0x54 || data[2] != 0x68 || data[3] != 0x64 {
		return errors.New("invalid MIDI header - must start with 'MThd'")
	}

	return nil
}

// GetMIDIDuration extracts the duration of a MIDI file in milliseconds
// This is a simplified version that reads the time division and looks for the end of track
func (ms *MIDIService) GetMIDIDuration(data []byte) (float64, error) {
	if err := ms.ValidateMIDI(data); err != nil {
		return 0, err
	}

	reader := bytes.NewReader(data)

	// Skip MIDI header (14 bytes: 4 for "MThd", 4 for header length, 2 for format, 2 for tracks, 2 for division)
	_, err := reader.Seek(8, io.SeekStart)
	if err != nil {
		return 0, err
	}

	var format, tracks, division uint16
	if err := binary.Read(reader, binary.BigEndian, &format); err != nil {
		return 0, err
	}
	if err := binary.Read(reader, binary.BigEndian, &tracks); err != nil {
		return 0, err
	}
	if err := binary.Read(reader, binary.BigEndian, &division); err != nil {
		return 0, err
	}

	// Default tempo: 120 BPM
	tempoBPM := 120.0

	// Rough duration calculation (simplified)
	// In a real implementation, you'd parse all track events and sum their deltas
	// For now, return an estimate based on file size and common MIDI patterns
	fileSize := float64(len(data))
	estimatedDuration := (fileSize / 128.0) * (60000.0 / tempoBPM)

	return estimatedDuration, nil
}

// ExtractNotes extracts individual note events from MIDI data
// Returns a simplified list of notes for display/analysis
func (ms *MIDIService) ExtractNotes(data []byte) ([]MIDINote, error) {
	if err := ms.ValidateMIDI(data); err != nil {
		return nil, err
	}

	// This is a simplified note extraction
	// A full implementation would properly parse MIDI events
	var notes []MIDINote

	// In a real implementation, you would:
	// 1. Parse the MIDI header to get time division
	// 2. Iterate through track events
	// 3. Extract Note On/Note Off events
	// 4. Calculate note durations from delta times
	// 5. Build the MIDINote list

	// For now, return an empty list (proper parsing can be added later)
	return notes, nil
}

// CountNotes returns the approximate number of notes in a MIDI file
func (ms *MIDIService) CountNotes(data []byte) (int, error) {
	if err := ms.ValidateMIDI(data); err != nil {
		return 0, err
	}

	// Count Note On events in the MIDI data
	// MIDI Note On event: status byte 0x90-0x9F (format depends on channel)
	count := 0
	for i := 0; i < len(data)-2; i++ {
		// Look for Note On events (0x90 = channel 1, 0x91 = channel 2, etc.)
		if (data[i] >= 0x90 && data[i] <= 0x9F) && data[i+2] > 0 {
			// Second byte is note number, third byte is velocity
			// Only count if velocity > 0 (0 velocity = note off)
			count++
		}
	}

	return count, nil
}

// ConvertToBase64 encodes MIDI binary data to base64 for transmission
func (ms *MIDIService) ConvertToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// ConvertFromHex decodes hex-encoded MIDI data back to binary
func (ms *MIDIService) ConvertFromHex(hexData string) ([]byte, error) {
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, fmt.Errorf("invalid hex encoding: %w", err)
	}
	return data, nil
}

// RecordingSession represents a MIDI recording session
type RecordingSession struct {
	MIDIData      []byte        // Recorded MIDI data
	Duration      float64       // Duration in seconds
	NotesRecorded int           // Number of notes recorded
	StartTime     int64         // Unix timestamp when recording started
	EndTime       int64         // Unix timestamp when recording ended
	AverageVelocity float64     // Average note velocity (0-127)
}

// StartRecording initializes a new recording session
func (ms *MIDIService) StartRecording() *RecordingSession {
	return &RecordingSession{
		MIDIData: make([]byte, 0),
		StartTime: time.Now().UnixMilli(),
	}
}

// AddNoteToRecording simulates adding a note event to a recording
// In a real implementation, this would accumulate MIDI events
func (ms *MIDIService) AddNoteToRecording(session *RecordingSession, note int, velocity int, duration float64) error {
	if session == nil {
		return errors.New("recording session is nil")
	}

	if note < 0 || note > 127 {
		return errors.New("MIDI note must be between 0 and 127")
	}

	if velocity < 0 || velocity > 127 {
		return errors.New("MIDI velocity must be between 0 and 127")
	}

	session.NotesRecorded++
	session.AverageVelocity = (session.AverageVelocity*float64(session.NotesRecorded-1) + float64(velocity)) / float64(session.NotesRecorded)

	return nil
}

// FinishRecording completes a recording session and returns the recorded MIDI data
func (ms *MIDIService) FinishRecording(session *RecordingSession) ([]byte, error) {
	if session == nil {
		return nil, errors.New("recording session is nil")
	}

	session.EndTime = time.Now().UnixMilli()
	session.Duration = float64(session.EndTime-session.StartTime) / 1000.0

	// In a real implementation, this would generate proper MIDI binary format
	// For now, return the accumulated MIDI data
	return session.MIDIData, nil
}

// CalculateBPMFromRecording estimates the BPM from recorded note timings
func (ms *MIDIService) CalculateBPMFromRecording(session *RecordingSession, targetBPM float64) float64 {
	if session.NotesRecorded < 2 || session.Duration <= 0 {
		return targetBPM // Return target if not enough data
	}

	// Estimate BPM based on note density
	// Assume average song has 4 notes per beat
	notesPerSecond := float64(session.NotesRecorded) / session.Duration
	estimatedBPM := (notesPerSecond * 60.0) / 4.0

	return estimatedBPM
}
