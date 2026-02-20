package typing

import (
	"encoding/json"
	"testing"
	"time"
)

// TestTypingTestValidation tests the Validate method for TypingTest
func TestTypingTestValidation(t *testing.T) {
	tests := []struct {
		name    string
		test    *TypingTest
		wantErr bool
	}{
		{
			name: "valid typing test",
			test: &TypingTest{
				ID:       1,
				UserID:   1,
				WPM:      60.5,
				Accuracy: 95.5,
				Duration: 120.0,
				Errors:   2,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			test: &TypingTest{
				ID:       1,
				UserID:   0,
				WPM:      60.5,
				Accuracy: 95.5,
				Duration: 120.0,
			},
			wantErr: true,
		},
		{
			name: "negative wpm",
			test: &TypingTest{
				ID:       1,
				UserID:   1,
				WPM:      -10.0,
				Accuracy: 95.5,
				Duration: 120.0,
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy above 100",
			test: &TypingTest{
				ID:       1,
				UserID:   1,
				WPM:      60.5,
				Accuracy: 101.0,
				Duration: 120.0,
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy below 0",
			test: &TypingTest{
				ID:       1,
				UserID:   1,
				WPM:      60.5,
				Accuracy: -1.0,
				Duration: 120.0,
			},
			wantErr: true,
		},
		{
			name: "invalid duration",
			test: &TypingTest{
				ID:       1,
				UserID:   1,
				WPM:      60.5,
				Accuracy: 95.5,
				Duration: -10.0,
			},
			wantErr: true,
		},
		{
			name: "negative errors",
			test: &TypingTest{
				ID:       1,
				UserID:   1,
				WPM:      60.5,
				Accuracy: 95.5,
				Duration: 120.0,
				Errors:   -1,
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

// TestTypingResultValidation tests the Validate method for TypingResult
func TestTypingResultValidation(t *testing.T) {
	tests := []struct {
		name    string
		result  *TypingResult
		wantErr bool
	}{
		{
			name: "valid typing result",
			result: &TypingResult{
				ID:          1,
				UserID:      1,
				Content:     "the quick brown fox",
				TimeSpent:   120.0,
				WPM:         60.5,
				Accuracy:    95.5,
				ErrorsCount: 2,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			result: &TypingResult{
				ID:        1,
				UserID:    0,
				Content:   "test",
				TimeSpent: 120.0,
				WPM:       60.5,
				Accuracy:  95.5,
			},
			wantErr: true,
		},
		{
			name: "missing content",
			result: &TypingResult{
				ID:        1,
				UserID:    1,
				Content:   "",
				TimeSpent: 120.0,
				WPM:       60.5,
				Accuracy:  95.5,
			},
			wantErr: true,
		},
		{
			name: "invalid time spent",
			result: &TypingResult{
				ID:        1,
				UserID:    1,
				Content:   "test",
				TimeSpent: -10.0,
				WPM:       60.5,
				Accuracy:  95.5,
			},
			wantErr: true,
		},
		{
			name: "negative wpm",
			result: &TypingResult{
				ID:        1,
				UserID:    1,
				Content:   "test",
				TimeSpent: 120.0,
				WPM:       -10.0,
				Accuracy:  95.5,
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy",
			result: &TypingResult{
				ID:        1,
				UserID:    1,
				Content:   "test",
				TimeSpent: 120.0,
				WPM:       60.5,
				Accuracy:  105.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUserStatsValidation tests the Validate method for UserStats
func TestUserStatsValidation(t *testing.T) {
	tests := []struct {
		name    string
		stats   *UserStats
		wantErr bool
	}{
		{
			name: "valid user stats",
			stats: &UserStats{
				UserID:          1,
				TotalTests:      10,
				AverageWPM:      60.5,
				BestWPM:         75.0,
				AverageAccuracy: 95.5,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			stats: &UserStats{
				UserID:          0,
				TotalTests:      10,
				AverageWPM:      60.5,
				BestWPM:         75.0,
				AverageAccuracy: 95.5,
			},
			wantErr: true,
		},
		{
			name: "negative total tests",
			stats: &UserStats{
				UserID:          1,
				TotalTests:      -1,
				AverageWPM:      60.5,
				BestWPM:         75.0,
				AverageAccuracy: 95.5,
			},
			wantErr: true,
		},
		{
			name: "negative average wpm",
			stats: &UserStats{
				UserID:          1,
				TotalTests:      10,
				AverageWPM:      -10.0,
				BestWPM:         75.0,
				AverageAccuracy: 95.5,
			},
			wantErr: true,
		},
		{
			name: "invalid average accuracy above 100",
			stats: &UserStats{
				UserID:          1,
				TotalTests:      10,
				AverageWPM:      60.5,
				BestWPM:         75.0,
				AverageAccuracy: 101.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stats.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTypingTestJSONMarshaling tests JSON marshaling for TypingTest
func TestTypingTestJSONMarshaling(t *testing.T) {
	now := time.Now()
	tt := &TypingTest{
		ID:        1,
		UserID:    1,
		TestTime:  now,
		WPM:       60.5,
		Accuracy:  95.5,
		Duration:  120.0,
		Errors:    2,
		CreatedAt: now,
	}

	// Test marshaling
	data, err := json.Marshal(tt)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Test unmarshaling
	var result TypingTest
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.ID != tt.ID || result.UserID != tt.UserID || result.WPM != tt.WPM {
		t.Errorf("Unmarshal() result mismatch: got %+v, want %+v", result, tt)
	}
}

// TestTypingResultJSONMarshaling tests JSON marshaling for TypingResult
func TestTypingResultJSONMarshaling(t *testing.T) {
	now := time.Now()
	tr := &TypingResult{
		ID:          1,
		UserID:      1,
		Content:     "the quick brown fox",
		TimeSpent:   120.0,
		WPM:         60.5,
		Accuracy:    95.5,
		ErrorsCount: 2,
		CreatedAt:   now,
	}

	// Test marshaling
	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Test unmarshaling
	var result TypingResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.ID != tr.ID || result.UserID != tr.UserID || result.Content != tr.Content {
		t.Errorf("Unmarshal() result mismatch: got %+v, want %+v", result, tr)
	}
}

// TestUserStatsJSONMarshaling tests JSON marshaling for UserStats
func TestUserStatsJSONMarshaling(t *testing.T) {
	now := time.Now()
	us := &UserStats{
		UserID:          1,
		TotalTests:      10,
		AverageWPM:      60.5,
		BestWPM:         75.0,
		AverageAccuracy: 95.5,
		LastUpdated:     now,
	}

	// Test marshaling
	data, err := json.Marshal(us)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Test unmarshaling
	var result UserStats
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.UserID != us.UserID || result.TotalTests != us.TotalTests || result.AverageWPM != us.AverageWPM {
		t.Errorf("Unmarshal() result mismatch: got %+v, want %+v", result, us)
	}
}
