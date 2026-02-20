package reading

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Book represents a book available for reading practice
type Book struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	Title         string        `json:"title" gorm:"index"`
	Author        string        `json:"author"`
	Content       string        `json:"content"`
	ReadingLevel  string        `json:"reading_level"` // beginner, intermediate, advanced
	Language      string        `json:"language" gorm:"default:english"`
	WordCount     int           `json:"word_count"`
	EstimatedTime float64       `json:"estimated_time_minutes"`
	CreatedAt     time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// ReadingSession represents a single reading practice session
type ReadingSession struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	UserID               uint      `json:"user_id" gorm:"index"`
	BookID               uint      `json:"book_id" gorm:"index"`
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	WPM                  float64   `json:"wpm"`           // Words per minute (0-500)
	Accuracy             float64   `json:"accuracy"`      // 0-100%
	ComprehensionScore   float64   `json:"comprehension"` // 0-100
	Duration             float64   `json:"duration"`      // Seconds
	WordsRead            int       `json:"words_read"`
	ErrorCount           int       `json:"error_count"`
	Completed            bool      `json:"completed"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ReadingStats represents aggregated statistics for a user
type ReadingStats struct {
	UserID              uint    `json:"user_id"`
	TotalBooksRead      int     `json:"total_books_read"`
	TotalSessionsCount  int     `json:"total_sessions"`
	TotalReadingTime    float64 `json:"total_reading_time"` // Seconds
	AverageWPM          float64 `json:"average_wpm"`
	BestWPM             float64 `json:"best_wpm"`
	AverageAccuracy     float64 `json:"average_accuracy"`
	AverageComprehension float64 `json:"average_comprehension"`
	FavoriteReadingLevel string  `json:"favorite_reading_level"`
	LastSessionTime      *time.Time `json:"last_session_time"`
}

// ComprehensionTest represents a comprehension quiz for a reading session
type ComprehensionTest struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SessionID    uint      `json:"session_id" gorm:"index"`
	Question     string    `json:"question"`
	UserAnswer   string    `json:"user_answer"`
	CorrectAnswer string   `json:"correct_answer"`
	IsCorrect    bool      `json:"is_correct"`
	Score        float64   `json:"score"` // 0-100
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ReadingTest represents test metadata
type ReadingTest struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SessionID    uint      `json:"session_id"`
	Questions    int       `json:"questions_count"`
	CorrectCount int       `json:"correct_count"`
	Score        float64   `json:"score"` // 0-100
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Validate performs validation on ReadingSession
func (rs *ReadingSession) Validate() error {
	if rs.UserID == 0 {
		return errors.New("user_id is required")
	}
	if rs.BookID == 0 {
		return errors.New("book_id is required")
	}
	if rs.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	if rs.WPM < 0 || rs.WPM > 500 {
		return fmt.Errorf("wpm must be between 0 and 500, got %f", rs.WPM)
	}
	if rs.Accuracy < 0 || rs.Accuracy > 100 {
		return fmt.Errorf("accuracy must be between 0 and 100, got %f", rs.Accuracy)
	}
	if rs.ComprehensionScore < 0 || rs.ComprehensionScore > 100 {
		return fmt.Errorf("comprehension score must be between 0 and 100, got %f", rs.ComprehensionScore)
	}
	return nil
}

// Validate performs validation on Book
func (b *Book) Validate() error {
	if b.Title == "" {
		return errors.New("title is required")
	}
	if b.Content == "" {
		return errors.New("content is required")
	}
	if b.ReadingLevel == "" {
		b.ReadingLevel = "intermediate"
	}
	validLevels := map[string]bool{"beginner": true, "intermediate": true, "advanced": true}
	if !validLevels[b.ReadingLevel] {
		return fmt.Errorf("invalid reading_level: %s", b.ReadingLevel)
	}
	if len(b.Content) < 50 {
		return errors.New("content is too short (minimum 50 characters)")
	}
	return nil
}

// Validate performs validation on ComprehensionTest
func (ct *ComprehensionTest) Validate() error {
	if ct.SessionID == 0 {
		return errors.New("session_id is required")
	}
	if ct.Question == "" {
		return errors.New("question is required")
	}
	if ct.CorrectAnswer == "" {
		return errors.New("correct_answer is required")
	}
	if ct.Score < 0 || ct.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", ct.Score)
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for ReadingSession
func (rs *ReadingSession) MarshalJSON() ([]byte, error) {
	type Alias ReadingSession
	return json.Marshal(&struct {
		*Alias
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(rs),
		StartTime: rs.StartTime.Format(time.RFC3339),
		EndTime:   rs.EndTime.Format(time.RFC3339),
		CreatedAt: rs.CreatedAt.Format(time.RFC3339),
		UpdatedAt: rs.UpdatedAt.Format(time.RFC3339),
	})
}

// CalculateReadingLevel estimates reading difficulty from word count and complexity
func CalculateReadingLevel(wordCount int) string {
	switch {
	case wordCount < 500:
		return "beginner"
	case wordCount < 2000:
		return "intermediate"
	default:
		return "advanced"
	}
}
