package math

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// AudioRecordRequest represents an incoming audio file upload
type AudioRecordRequest struct {
	UserID      uint   `json:"user_id"`
	SessionID   string `json:"session_id,omitempty"`
	MimeType    string `json:"mime_type"`
	Description string `json:"description,omitempty"`
}

// AudioRecordResponse represents the response from audio recording
type AudioRecordResponse struct {
	Success bool   `json:"success"`
	AudioID string `json:"audio_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AudioTranscribeRequest represents a transcription request
type AudioTranscribeRequest struct {
	AudioID string `json:"audio_id"`
	UserID  uint   `json:"user_id"`
}

// AudioTranscribeResponse represents the transcription result
type AudioTranscribeResponse struct {
	Success       bool   `json:"success"`
	AudioID       string `json:"audio_id"`
	Transcript    string `json:"transcript,omitempty"`
	Confidence    float64 `json:"confidence,omitempty"`
	Error         string `json:"error,omitempty"`
}

// RecordAudio handles POST /api/math/audio/record
func (h *Handler) RecordAudio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse multipart form with max 10MB file size
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse form: %v", err),
		})
		return
	}

	// Get audio file from request
	file, fileHeader, err := r.FormFile("audio")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   "No audio file provided",
		})
		return
	}
	defer file.Close()

	// Get user ID from form
	userIDStr := r.FormValue("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   "Invalid user_id",
		})
		return
	}

	// Create audio storage directory if it doesn't exist
	audioDir := "data/audio/math"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create audio directory: %v", err),
		})
		return
	}

	// Generate unique audio ID with timestamp
	audioID := fmt.Sprintf("math_%d_%d", userID, time.Now().UnixNano())
	audioPath := filepath.Join(audioDir, audioID+filepath.Ext(fileHeader.Filename))

	// Save audio file
	dst, err := os.Create(audioPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save audio file: %v", err),
		})
		return
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AudioRecordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write audio file: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AudioRecordResponse{
		Success: true,
		AudioID: audioID,
		Message: fmt.Sprintf("Audio file recorded successfully (%d bytes)", fileHeader.Size),
	})
}

// TranscribeAudio handles POST /api/math/audio/transcribe
func (h *Handler) TranscribeAudio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AudioTranscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioTranscribeResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.AudioID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AudioTranscribeResponse{
			Success: false,
			Error:   "audio_id is required",
		})
		return
	}

	// For now, return a placeholder transcription
	// In production, this would call Google Speech-to-Text API, Whisper, or similar service
	transcript := "This is a placeholder transcription. Audio processing service not yet integrated."

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AudioTranscribeResponse{
		Success:    true,
		AudioID:    req.AudioID,
		Transcript: transcript,
		Confidence: 0.0, // Placeholder confidence
	})
}
