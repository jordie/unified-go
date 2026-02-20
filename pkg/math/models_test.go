package math

import (
	"testing"
	"time"
)

func TestUserValidation(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid user",
			user:    User{Username: "student1"},
			wantErr: false,
		},
		{
			name:    "empty username",
			user:    User{Username: ""},
			wantErr: true,
			errMsg:  "username cannot be empty",
		},
		{
			name:    "whitespace only username",
			user:    User{Username: "   "},
			wantErr: true,
			errMsg:  "username cannot be empty",
		},
		{
			name:    "username too long",
			user:    User{Username: string(make([]byte, 101))},
			wantErr: true,
			errMsg:  "username exceeds 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMathResultValidation(t *testing.T) {
	tests := []struct {
		name    string
		result  MathResult
		wantErr bool
	}{
		{
			name: "valid result",
			result: MathResult{
				Mode:           "addition",
				Difficulty:     "easy",
				TotalQuestions: 10,
				CorrectAnswers: 9,
				TotalTime:      60,
				AverageTime:    6,
				Accuracy:       90,
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			result: MathResult{
				Mode:           "invalid",
				Difficulty:     "easy",
				TotalQuestions: 10,
				CorrectAnswers: 9,
				TotalTime:      60,
				AverageTime:    6,
				Accuracy:       90,
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty",
			result: MathResult{
				Mode:           "addition",
				Difficulty:     "insane",
				TotalQuestions: 10,
				CorrectAnswers: 9,
				TotalTime:      60,
				AverageTime:    6,
				Accuracy:       90,
			},
			wantErr: true,
		},
		{
			name: "accuracy out of range",
			result: MathResult{
				Mode:           "addition",
				Difficulty:     "easy",
				TotalQuestions: 10,
				CorrectAnswers: 9,
				TotalTime:      60,
				AverageTime:    6,
				Accuracy:       150,
			},
			wantErr: true,
		},
		{
			name: "correct answers exceed total",
			result: MathResult{
				Mode:           "addition",
				Difficulty:     "easy",
				TotalQuestions: 10,
				CorrectAnswers: 15,
				TotalTime:      60,
				AverageTime:    6,
				Accuracy:       90,
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

func TestMasteryCalculation(t *testing.T) {
	tests := []struct {
		name               string
		mastery            Mastery
		baseAccuracy       float64
		speedBonusApplied  bool
		expectedMastery    float64
	}{
		{
			name: "high accuracy, high streak, speed bonus",
			mastery: Mastery{
				CorrectStreak: 5,
			},
			baseAccuracy:      0.95,
			speedBonusApplied: true,
			expectedMastery:   100, // 0.95*80 + 5*4 + 100 = 76+20+100 = 196, capped at 100
		},
		{
			name: "medium accuracy, medium streak",
			mastery: Mastery{
				CorrectStreak: 3,
			},
			baseAccuracy:      0.70,
			speedBonusApplied: false,
			expectedMastery:   68, // 0.70*80 + 3*4 = 56 + 12 = 68
		},
		{
			name: "low accuracy",
			mastery: Mastery{
				CorrectStreak: 0,
			},
			baseAccuracy:      0.40,
			speedBonusApplied: false,
			expectedMastery:   32, // 0.40*80 = 32
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mastery.CalculateMasteryLevel(tt.baseAccuracy, tt.speedBonusApplied)
			if tt.mastery.MasteryLevel != tt.expectedMastery {
				t.Errorf("CalculateMasteryLevel() got = %v, want %v", tt.mastery.MasteryLevel, tt.expectedMastery)
			}
		})
	}
}

func TestRepetitionScheduleSM2(t *testing.T) {
	tests := []struct {
		name            string
		quality         int
		initialEase     float64
		expectedEaseLow float64
		expectedEaseHigh float64
	}{
		{
			name:             "perfect response (quality 5)",
			quality:          5,
			initialEase:      2.5,
			expectedEaseLow:  2.6, // ease + 0.1
			expectedEaseHigh: 2.7,
		},
		{
			name:             "difficult response (quality 3)",
			quality:          3,
			initialEase:      2.5,
			expectedEaseLow:  2.3,
			expectedEaseHigh: 2.5,
		},
		{
			name:             "complete failure (quality 0)",
			quality:          0,
			initialEase:      2.5,
			expectedEaseLow:  1.3, // Clamped to minimum
			expectedEaseHigh: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := RepetitionSchedule{EaseFactor: tt.initialEase}
			rs.UpdateEaseFactor(tt.quality)

			// Allow small floating point differences
			if rs.EaseFactor < tt.expectedEaseLow || rs.EaseFactor > tt.expectedEaseHigh {
				if !(rs.EaseFactor >= 1.3 && rs.EaseFactor <= 3.5) {
					t.Errorf("UpdateEaseFactor() got = %v, want between %v and %v", rs.EaseFactor, 1.3, 3.5)
				}
			}
		})
	}
}

func TestRepetitionScheduleInterval(t *testing.T) {
	tests := []struct {
		name               string
		reviewCount        int
		currentInterval    int
		currentEaseFactor  float64
		quality            int
		expectedInterval   int
	}{
		{
			name:              "first review (quality 4)",
			reviewCount:       0,
			currentInterval:   1,
			currentEaseFactor: 2.5,
			quality:           4,
			expectedInterval:  1,
		},
		{
			name:              "second review (quality 4)",
			reviewCount:       1,
			currentInterval:   1,
			currentEaseFactor: 2.5,
			quality:           4,
			expectedInterval:  6,
		},
		{
			name:              "third review with ease factor",
			reviewCount:       2,
			currentInterval:   6,
			currentEaseFactor: 2.5,
			quality:           4,
			expectedInterval:  15, // 6 * 2.5 = 15
		},
		{
			name:              "failed review resets to 1",
			reviewCount:       5,
			currentInterval:   30,
			currentEaseFactor: 2.5,
			quality:           2,
			expectedInterval:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := RepetitionSchedule{
				ReviewCount:  tt.reviewCount,
				IntervalDays: tt.currentInterval,
				EaseFactor:   tt.currentEaseFactor,
			}
			nextInterval := rs.CalculateNextInterval(tt.quality)

			if nextInterval != tt.expectedInterval {
				t.Errorf("CalculateNextInterval() got = %v, want %v", nextInterval, tt.expectedInterval)
			}
		})
	}
}

func TestRepetitionScheduleIsDueForReview(t *testing.T) {
	tests := []struct {
		name         string
		nextReview   time.Time
		shouldBeDue  bool
	}{
		{
			name:        "due now",
			nextReview:  time.Now(),
			shouldBeDue: true,
		},
		{
			name:        "due in the past",
			nextReview:  time.Now().AddDate(0, 0, -1),
			shouldBeDue: true,
		},
		{
			name:        "not due (future)",
			nextReview:  time.Now().AddDate(0, 0, 1),
			shouldBeDue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := RepetitionSchedule{NextReview: tt.nextReview}
			isDue := rs.IsDueForReview()

			if isDue != tt.shouldBeDue {
				t.Errorf("IsDueForReview() got = %v, want %v", isDue, tt.shouldBeDue)
			}
		})
	}
}

func TestMasteryResponseTimeUpdate(t *testing.T) {
	m := Mastery{
		AverageResponseTime: 0,
		FastestTime:         0,
		SlowestTime:         0,
		TotalAttempts:       0,
	}

	// First update
	m.UpdateResponseTime(5.0)
	if m.AverageResponseTime != 5.0 {
		t.Errorf("First update: average = %v, want 5.0", m.AverageResponseTime)
	}
	if m.FastestTime != 5.0 {
		t.Errorf("First update: fastest = %v, want 5.0", m.FastestTime)
	}
	if m.SlowestTime != 5.0 {
		t.Errorf("First update: slowest = %v, want 5.0", m.SlowestTime)
	}

	// Second update (faster)
	m.UpdateResponseTime(3.0)
	if m.FastestTime != 3.0 {
		t.Errorf("Second update: fastest = %v, want 3.0", m.FastestTime)
	}
	if m.SlowestTime != 5.0 {
		t.Errorf("Second update: slowest = %v, want 5.0", m.SlowestTime)
	}

	// Third update (slower)
	m.UpdateResponseTime(7.0)
	if m.SlowestTime != 7.0 {
		t.Errorf("Third update: slowest = %v, want 7.0", m.SlowestTime)
	}

	// Average should be (5 + 3 + 7) / 3 = 5.0
	expectedAvg := (5.0 + 3.0 + 7.0) / 3.0
	if m.AverageResponseTime != expectedAvg {
		t.Errorf("Average = %v, want %v", m.AverageResponseTime, expectedAvg)
	}
}

func TestGetTimeOfDayFromHour(t *testing.T) {
	tests := []struct {
		hour     int
		expected string
	}{
		{5, "morning"},
		{8, "morning"},
		{11, "morning"},
		{12, "afternoon"},
		{14, "afternoon"},
		{16, "afternoon"},
		{17, "evening"},
		{20, "evening"},
		{23, "evening"},
		{0, "evening"},
		{4, "evening"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.hour)), func(t *testing.T) {
			result := GetTimeOfDayFromHour(tt.hour)
			if result != tt.expected {
				t.Errorf("GetTimeOfDayFromHour(%d) got = %v, want %v", tt.hour, result, tt.expected)
			}
		})
	}
}

func TestMistakeIncrement(t *testing.T) {
	m := Mistake{
		ErrorCount: 1,
		LastError:  time.Now().AddDate(0, 0, -1),
	}

	oldTime := m.LastError
	m.IncrementError()

	if m.ErrorCount != 2 {
		t.Errorf("ErrorCount = %d, want 2", m.ErrorCount)
	}

	if m.LastError.Before(oldTime) {
		t.Error("LastError should be updated to current time")
	}
}

func TestMathResultAccuracyCalculation(t *testing.T) {
	tests := []struct {
		name               string
		totalQuestions     int
		correctAnswers     int
		expectedAccuracy   float64
	}{
		{
			name:               "perfect accuracy",
			totalQuestions:     10,
			correctAnswers:     10,
			expectedAccuracy:   100.0,
		},
		{
			name:               "50% accuracy",
			totalQuestions:     10,
			correctAnswers:     5,
			expectedAccuracy:   50.0,
		},
		{
			name:               "zero questions",
			totalQuestions:     0,
			correctAnswers:     0,
			expectedAccuracy:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := MathResult{
				TotalQuestions: tt.totalQuestions,
				CorrectAnswers: tt.correctAnswers,
			}
			r.CalculateAccuracy()

			if r.Accuracy != tt.expectedAccuracy {
				t.Errorf("CalculateAccuracy() got = %v, want %v", r.Accuracy, tt.expectedAccuracy)
			}
		})
	}
}

func TestPerformancePatternValidation(t *testing.T) {
	tests := []struct {
		name    string
		pattern PerformancePattern
		wantErr bool
	}{
		{
			name: "valid pattern",
			pattern: PerformancePattern{
				HourOfDay:       10,
				DayOfWeek:       3,
				AverageAccuracy: 85.5,
				AverageSpeed:    2.5,
				SessionCount:    5,
			},
			wantErr: false,
		},
		{
			name: "invalid hour",
			pattern: PerformancePattern{
				HourOfDay:       25,
				DayOfWeek:       3,
				AverageAccuracy: 85.5,
			},
			wantErr: true,
		},
		{
			name: "invalid day of week",
			pattern: PerformancePattern{
				HourOfDay:       10,
				DayOfWeek:       8,
				AverageAccuracy: 85.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pattern.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
