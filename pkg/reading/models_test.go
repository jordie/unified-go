package reading

import (
	"testing"
	"time"
)

// TestBookValidation tests Book struct validation
func TestBookValidation(t *testing.T) {
	tests := []struct {
		name    string
		book    *Book
		wantErr bool
	}{
		{
			name: "valid book",
			book: &Book{
				Title:        "The Great Gatsby",
				Author:       "F. Scott Fitzgerald",
				Content:      "In my younger and more vulnerable years, my father gave me advice that I've been turning over in my mind ever since.",
				ReadingLevel: "intermediate",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			book: &Book{
				Title:   "",
				Author:  "Author",
				Content: "This is some valid content for a book.",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			book: &Book{
				Title:  "Title",
				Author: "Author",
			},
			wantErr: true,
		},
		{
			name: "content too short",
			book: &Book{
				Title:   "Title",
				Author:  "Author",
				Content: "short",
			},
			wantErr: true,
		},
		{
			name: "invalid reading level",
			book: &Book{
				Title:        "Title",
				Author:       "Author",
				Content:      "This is some valid content with enough characters.",
				ReadingLevel: "expert",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.book.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReadingSessionValidation tests ReadingSession struct validation
func TestReadingSessionValidation(t *testing.T) {
	tests := []struct {
		name    string
		session *ReadingSession
		wantErr bool
	}{
		{
			name: "valid session",
			session: &ReadingSession{
				UserID:               1,
				BookID:               1,
				WPM:                  150.0,
				Accuracy:             95.5,
				ComprehensionScore:   85.0,
				Duration:             600.0,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			session: &ReadingSession{
				UserID:    0,
				BookID:    1,
				Duration:  600.0,
				WPM:       150.0,
				Accuracy:  95.0,
			},
			wantErr: true,
		},
		{
			name: "missing book_id",
			session: &ReadingSession{
				UserID:   1,
				BookID:   0,
				Duration: 600.0,
			},
			wantErr: true,
		},
		{
			name: "zero duration",
			session: &ReadingSession{
				UserID:   1,
				BookID:   1,
				Duration: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid WPM (too high)",
			session: &ReadingSession{
				UserID:   1,
				BookID:   1,
				Duration: 600.0,
				WPM:      550.0,
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy (negative)",
			session: &ReadingSession{
				UserID:   1,
				BookID:   1,
				Duration: 600.0,
				WPM:      150.0,
				Accuracy: -10.0,
			},
			wantErr: true,
		},
		{
			name: "invalid comprehension (over 100)",
			session: &ReadingSession{
				UserID:               1,
				BookID:               1,
				Duration:             600.0,
				WPM:                  150.0,
				Accuracy:             95.0,
				ComprehensionScore:   150.0,
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

// TestComprehensionTestValidation tests ComprehensionTest struct validation
func TestComprehensionTestValidation(t *testing.T) {
	tests := []struct {
		name    string
		test    *ComprehensionTest
		wantErr bool
	}{
		{
			name: "valid comprehension test",
			test: &ComprehensionTest{
				SessionID:     1,
				Question:      "What is the main theme?",
				UserAnswer:    "Love and ambition",
				CorrectAnswer: "Love and ambition",
				IsCorrect:     true,
				Score:         100.0,
			},
			wantErr: false,
		},
		{
			name: "missing session_id",
			test: &ComprehensionTest{
				SessionID: 0,
				Question:  "What is the theme?",
			},
			wantErr: true,
		},
		{
			name: "missing question",
			test: &ComprehensionTest{
				SessionID: 1,
				Question:  "",
			},
			wantErr: true,
		},
		{
			name: "missing correct_answer",
			test: &ComprehensionTest{
				SessionID:     1,
				Question:      "What is it?",
				CorrectAnswer: "",
			},
			wantErr: true,
		},
		{
			name: "invalid score (over 100)",
			test: &ComprehensionTest{
				SessionID:     1,
				Question:      "Question",
				CorrectAnswer: "Answer",
				Score:         150.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.test.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReadingLevelCalculation tests reading level calculation from word count
func TestReadingLevelCalculation(t *testing.T) {
	tests := []struct {
		wordCount       int
		expectedLevel   string
	}{
		{100, "beginner"},
		{400, "beginner"},
		{500, "intermediate"},
		{1000, "intermediate"},
		{1999, "intermediate"},
		{2000, "advanced"},
		{5000, "advanced"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedLevel, func(t *testing.T) {
			level := CalculateReadingLevel(tt.wordCount)
			if level != tt.expectedLevel {
				t.Errorf("CalculateReadingLevel(%d) = %s, want %s", tt.wordCount, level, tt.expectedLevel)
			}
		})
	}
}

// TestReadingSessionTimestamps tests timestamp handling in ReadingSession
func TestReadingSessionTimestamps(t *testing.T) {
	session := &ReadingSession{
		UserID:    1,
		BookID:    1,
		Duration:  600.0,
		WPM:       150.0,
		Accuracy:  95.0,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(10 * time.Minute),
	}

	if session.StartTime.After(session.EndTime) {
		t.Error("StartTime should be before EndTime")
	}
}

// TestReadingSessionMarshalJSON tests JSON marshaling
func TestReadingSessionMarshalJSON(t *testing.T) {
	session := &ReadingSession{
		ID:       1,
		UserID:   1,
		BookID:   1,
		Duration: 600.0,
		WPM:      150.0,
		Accuracy: 95.0,
		StartTime: time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 2, 20, 10, 10, 0, 0, time.UTC),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := session.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalJSON() returned empty data")
	}

	// Verify it contains ISO format dates
	if string(data) == "" {
		t.Error("MarshalJSON() returned empty string")
	}
}

// TestReadingSessionDefaultValues tests default value assignment
func TestReadingSessionDefaultValues(t *testing.T) {
	session := &ReadingSession{
		UserID:  1,
		BookID:  1,
		Duration: 300.0,
	}

	if session.Completed {
		t.Error("New session should not be marked as completed")
	}

	if session.ErrorCount != 0 {
		t.Error("New session should have 0 error count")
	}
}

// TestReadingStatsAggregation tests ReadingStats struct creation
func TestReadingStatsAggregation(t *testing.T) {
	stats := &ReadingStats{
		UserID:               1,
		TotalBooksRead:       5,
		TotalSessionsCount:   10,
		TotalReadingTime:     36000.0, // 10 hours
		AverageWPM:           200.0,
		BestWPM:              250.0,
		AverageAccuracy:      92.5,
		AverageComprehension: 85.0,
		FavoriteReadingLevel: "intermediate",
	}

	if stats.UserID != 1 {
		t.Error("Stats UserID mismatch")
	}

	if stats.TotalBooksRead <= 0 {
		t.Error("TotalBooksRead should be positive")
	}

	if stats.AverageWPM > stats.BestWPM {
		t.Error("AverageWPM should not exceed BestWPM")
	}
}
