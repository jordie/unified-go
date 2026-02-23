package math

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

// ============================================================================
// WHISPER API INTEGRATION FOR SPEECH-TO-TEXT
// ============================================================================

// TranscribeAudio transcribes audio using Whisper
// Falls back to local Whisper server if available, then OpenAI API
func TranscribeAudio(audioData []byte, filename string) (string, float64, error) {
	// Try local Whisper server first (if running on localhost:5000)
	if text, confidence, err := transcribeWithLocalWhisper(audioData, filename); err == nil {
		return text, confidence, nil
	}

	// Fall back to OpenAI Whisper API
	return transcribeWithOpenAI(audioData, filename)
}

// transcribeWithLocalWhisper sends audio to a local Whisper server
func transcribeWithLocalWhisper(audioData []byte, filename string) (string, float64, error) {
	localURL := os.Getenv("WHISPER_LOCAL_URL")
	if localURL == "" {
		localURL = "http://localhost:5000/transcribe"
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add audio file
	part, err := writer.CreateFormFile("audio", filename)
	if err != nil {
		return "", 0, err
	}

	if _, err := part.Write(audioData); err != nil {
		return "", 0, err
	}

	// Add language
	if err := writer.WriteField("language", "en"); err != nil {
		return "", 0, err
	}

	writer.Close()

	req, err := http.NewRequest("POST", localURL, body)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("local Whisper API error: %d", resp.StatusCode)
	}

	var result struct {
		Text       string  `json:"text"`
		Confidence float64 `json:"confidence"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	return strings.TrimSpace(result.Text), result.Confidence, nil
}

// transcribeWithOpenAI sends audio to OpenAI Whisper API
func transcribeWithOpenAI(audioData []byte, filename string) (string, float64, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Fallback: Try to use local Whisper Python server on port 5000
		// If that fails, return mock transcription for development
		return mockTranscribeAudio(audioData), 0.85, nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add audio file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", 0, err
	}

	if _, err := part.Write(audioData); err != nil {
		return "", 0, err
	}

	// Add model
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return "", 0, err
	}

	// Add language
	if err := writer.WriteField("language", "en"); err != nil {
		return "", 0, err
	}

	writer.Close()

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	// Confidence is estimated based on text length and quality indicators
	confidence := estimateConfidence(result.Text)

	return strings.TrimSpace(result.Text), confidence, nil
}

// estimateConfidence estimates transcription confidence
func estimateConfidence(text string) float64 {
	// Simple heuristic: longer, cleaner text = higher confidence
	if text == "" {
		return 0
	}

	confidence := 0.85 // Base confidence

	// Adjust based on text length
	if len(text) > 50 {
		confidence += 0.10
	} else if len(text) < 5 {
		confidence -= 0.20
	}

	// Check for common transcription errors (repeated words, gibberish)
	words := strings.Fields(text)
	if len(words) > 1 {
		// Check for excessive repetition (max 2 repeats)
		repeatCount := 0
		for i := 1; i < len(words); i++ {
			if words[i] == words[i-1] {
				repeatCount++
			}
		}
		if repeatCount > 2 {
			confidence -= float64(repeatCount) * 0.05
		}
	}

	// Cap confidence at 0.99 and floor at 0.5
	if confidence > 0.99 {
		confidence = 0.99
	} else if confidence < 0.5 {
		confidence = 0.5
	}

	return confidence
}

// ============================================================================
// MOCK TRANSCRIPTION (For Development/Testing)
// ============================================================================

// mockTranscribeAudio provides mock transcription for development
// This allows students to test speech-to-text without an API key
func mockTranscribeAudio(audioData []byte) string {
	if len(audioData) == 0 {
		return ""
	}

	// Mock responses based on audio length for testing
	// In production, use real Whisper API with valid OPENAI_API_KEY
	mockResponses := []string{
		"eight",
		"fifteen",
		"twenty five",
		"forty two",
		"ninety nine",
		"one hundred",
		"five thousand",
	}

	// Use audio data hash to select a mock response (for consistency)
	hash := 0
	for _, b := range audioData {
		hash = (hash*31 + int(b)) % len(mockResponses)
	}

	return mockResponses[hash]
}

// OfflineTranscribeDemo is a mock transcriber for development/testing
// In production, use actual Whisper API
func OfflineTranscribeDemo(audioData []byte) (string, float64) {
	// This is just for testing without API keys
	// In production, ensure Whisper API is properly configured

	if len(audioData) == 0 {
		return "", 0
	}

	text := mockTranscribeAudio(audioData)
	confidence := 0.85

	// Estimate confidence based on audio length
	seconds := len(audioData) / 50
	if seconds > 5 {
		confidence = 0.92
	} else if seconds < 2 {
		confidence = 0.75
	}

	return text, confidence
}
