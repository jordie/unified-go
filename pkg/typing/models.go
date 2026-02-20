package typing

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// TypingTest represents a single typing test result
type TypingTest struct {
	ID          uint      `json:"id" db:"id"`
	UserID      uint      `json:"user_id" db:"user_id"`
	TestTime    time.Time `json:"test_time" db:"timestamp"`
	WPM         float64   `json:"wpm" db:"wpm"`
	RawWPM      float64   `json:"raw_wpm" db:"raw_wpm"`
	Accuracy    float64   `json:"accuracy" db:"accuracy"`
	Duration    float64   `json:"duration" db:"time_taken"`
	Errors      int       `json:"errors" db:"errors"`
	TestMode    string    `json:"test_mode" db:"test_mode"`
	TextSnippet string    `json:"text_snippet" db:"text_snippet"`
	CreatedAt   time.Time `json:"created_at" db:"timestamp"`
}

// TypingResult represents a typing test result with content
type TypingResult struct {
	ID          uint      `json:"id" db:"id"`
	UserID      uint      `json:"user_id" db:"user_id"`
	Content     string    `json:"content" db:"text_snippet"`
	TimeSpent   float64   `json:"time_spent" db:"time_taken"`
	WPM         float64   `json:"wpm" db:"wpm"`
	RawWPM      float64   `json:"raw_wpm" db:"raw_wpm"`
	ErrorsCount int       `json:"errors_count" db:"errors"`
	Accuracy    float64   `json:"accuracy" db:"accuracy"`
	TestMode    string    `json:"test_mode" db:"test_mode"`
	CreatedAt   time.Time `json:"created_at" db:"timestamp"`
}

// UserStats represents aggregated user statistics
type UserStats struct {
	UserID          uint      `json:"user_id" db:"user_id"`
	TotalTests      int       `json:"total_tests" db:"total_tests"`
	AverageWPM      float64   `json:"average_wpm" db:"average_wpm"`
	BestWPM         float64   `json:"best_wpm" db:"best_wpm"`
	AverageAccuracy float64   `json:"average_accuracy" db:"average_accuracy"`
	TotalTimeTyped  int       `json:"total_time_typed" db:"total_time_typed"`
	LastUpdated     time.Time `json:"last_updated" db:"last_updated"`
}

// Validate checks if a TypingTest is valid
func (t *TypingTest) Validate() error {
	if t.UserID == 0 {
		return errors.New("user_id is required")
	}
	if t.WPM < 0 {
		return errors.New("wpm cannot be negative")
	}
	if t.Accuracy < 0 || t.Accuracy > 100 {
		return errors.New("accuracy must be between 0 and 100")
	}
	if t.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	if t.Errors < 0 {
		return errors.New("errors cannot be negative")
	}
	return nil
}

// Validate checks if a TypingResult is valid
func (r *TypingResult) Validate() error {
	if r.UserID == 0 {
		return errors.New("user_id is required")
	}
	if len(r.Content) == 0 {
		return errors.New("content is required")
	}
	if r.TimeSpent <= 0 {
		return errors.New("time_spent must be positive")
	}
	if r.WPM < 0 {
		return errors.New("wpm cannot be negative")
	}
	if r.Accuracy < 0 || r.Accuracy > 100 {
		return errors.New("accuracy must be between 0 and 100")
	}
	if r.ErrorsCount < 0 {
		return errors.New("errors_count cannot be negative")
	}
	return nil
}

// Validate checks if UserStats is valid
func (s *UserStats) Validate() error {
	if s.UserID == 0 {
		return errors.New("user_id is required")
	}
	if s.TotalTests < 0 {
		return errors.New("total_tests cannot be negative")
	}
	if s.AverageWPM < 0 {
		return errors.New("average_wpm cannot be negative")
	}
	if s.BestWPM < 0 {
		return errors.New("best_wpm cannot be negative")
	}
	if s.AverageAccuracy < 0 || s.AverageAccuracy > 100 {
		return errors.New("average_accuracy must be between 0 and 100")
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for TypingTest
func (t *TypingTest) MarshalJSON() ([]byte, error) {
	type Alias TypingTest
	return json.Marshal(&struct {
		TestTime  string `json:"test_time"`
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		TestTime:  t.TestTime.Format(time.RFC3339),
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		Alias:     (*Alias)(t),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for TypingTest
func (t *TypingTest) UnmarshalJSON(data []byte) error {
	type Alias TypingTest
	aux := &struct {
		TestTime  string `json:"test_time"`
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	if aux.TestTime != "" {
		t.TestTime, err = time.Parse(time.RFC3339, aux.TestTime)
		if err != nil {
			return err
		}
	}
	if aux.CreatedAt != "" {
		t.CreatedAt, err = time.Parse(time.RFC3339, aux.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for TypingResult
func (r *TypingResult) MarshalJSON() ([]byte, error) {
	type Alias TypingResult
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		Alias:     (*Alias)(r),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for TypingResult
func (r *TypingResult) UnmarshalJSON(data []byte) error {
	type Alias TypingResult
	aux := &struct {
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.CreatedAt != "" {
		var err error
		r.CreatedAt, err = time.Parse(time.RFC3339, aux.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for UserStats
func (s *UserStats) MarshalJSON() ([]byte, error) {
	type Alias UserStats
	return json.Marshal(&struct {
		LastUpdated string `json:"last_updated"`
		*Alias
	}{
		LastUpdated: s.LastUpdated.Format(time.RFC3339),
		Alias:       (*Alias)(s),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for UserStats
func (s *UserStats) UnmarshalJSON(data []byte) error {
	type Alias UserStats
	aux := &struct {
		LastUpdated string `json:"last_updated"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.LastUpdated != "" {
		var err error
		s.LastUpdated, err = time.Parse(time.RFC3339, aux.LastUpdated)
		if err != nil {
			return err
		}
	}
	return nil
}

// ScanRow scans a database row into a TypingTest
func (t *TypingTest) ScanRow(rows *sql.Rows) error {
	return rows.Scan(
		&t.ID,
		&t.UserID,
		&t.WPM,
		&t.RawWPM,
		&t.Accuracy,
		&t.Errors,
		&t.Duration,
		&t.TestMode,
		&t.TextSnippet,
		&t.CreatedAt,
	)
}

// ScanRow scans a database row into a TypingResult
func (r *TypingResult) ScanRow(rows *sql.Rows) error {
	return rows.Scan(
		&r.ID,
		&r.UserID,
		&r.Content,
		&r.TimeSpent,
		&r.WPM,
		&r.RawWPM,
		&r.ErrorsCount,
		&r.Accuracy,
		&r.TestMode,
		&r.CreatedAt,
	)
}

// ScanRow scans a database row into UserStats
func (s *UserStats) ScanRow(rows *sql.Rows) error {
	return rows.Scan(
		&s.UserID,
		&s.TotalTests,
		&s.AverageWPM,
		&s.BestWPM,
		&s.AverageAccuracy,
		&s.TotalTimeTyped,
		&s.LastUpdated,
	)
}
