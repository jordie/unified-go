package typing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/jgirmay/unified-go/internal/database"
)

// Service provides business logic for typing operations
type Service struct {
	repo *Repository
}

// NewService creates a new typing service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CalculateWPM calculates words per minute from content and time spent
// Formula: (characters / 5) / (minutes elapsed)
// The standard in typing is 5 characters = 1 word
func (s *Service) CalculateWPM(content string, timeSpentSeconds float64) float64 {
	if timeSpentSeconds <= 0 {
		return 0
	}

	// Convert seconds to minutes
	minutes := timeSpentSeconds / 60.0

	// Count non-whitespace characters
	charCount := float64(len(strings.TrimSpace(content)))

	// Calculate WPM: (characters / 5) / minutes
	// Using 5 as the standard character count per word
	wpm := (charCount / 5.0) / minutes

	// Round to 2 decimal places
	return math.Round(wpm*100) / 100
}

// CalculateAccuracy calculates typing accuracy as a percentage
// Compares typed content against expected content character by character
// Ignores differences in whitespace at the beginning/end
func (s *Service) CalculateAccuracy(typed, expected string) float64 {
	typed = strings.TrimSpace(typed)
	expected = strings.TrimSpace(expected)

	if len(expected) == 0 {
		if len(typed) == 0 {
			return 100.0
		}
		return 0.0
	}

	// Use the length of the expected content as the baseline
	correctChars := 0
	minLen := len(typed)
	if len(expected) < minLen {
		minLen = len(expected)
	}

	// Count matching characters
	for i := 0; i < minLen; i++ {
		if typed[i] == expected[i] {
			correctChars++
		}
	}

	// Add penalty for length mismatch
	lengthDifference := len(expected) - len(typed)
	if lengthDifference > 0 {
		// User typed less than expected
		correctChars -= lengthDifference
	}

	// Ensure we don't have negative correct characters
	if correctChars < 0 {
		correctChars = 0
	}

	// Calculate accuracy percentage
	accuracy := (float64(correctChars) / float64(len(expected))) * 100.0

	// Clamp to 0-100%
	if accuracy < 0 {
		accuracy = 0
	} else if accuracy > 100 {
		accuracy = 100
	}

	// Round to 2 decimal places
	return math.Round(accuracy*100) / 100
}

// ProcessTestResult processes a completed typing test
// Validates input, calculates WPM and accuracy, saves to repository
func (s *Service) ProcessTestResult(ctx context.Context, userID uint, content string, timeSpentSeconds float64, errorCount int) (*TypingResult, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if len(strings.TrimSpace(content)) == 0 {
		return nil, errors.New("content is required")
	}

	if timeSpentSeconds <= 0 {
		return nil, errors.New("time_spent must be positive")
	}

	if errorCount < 0 {
		return nil, errors.New("errors cannot be negative")
	}

	// Calculate WPM and accuracy
	wpm := s.CalculateWPM(content, timeSpentSeconds)
	rawWPM := wpm // Raw WPM before error adjustment

	// Adjust WPM for errors (optional: subtract errors from WPM)
	// This is a common practice in typing tests
	errorPenalty := float64(errorCount) * 0.5 // Each error reduces WPM by 0.5
	adjustedWPM := wpm - errorPenalty
	if adjustedWPM < 0 {
		adjustedWPM = 0
	}

	// For accuracy, we use a reference text approach
	// Since we don't have the original reference text, we calculate based on errors
	// Assuming average word length of 5 characters
	estimatedLength := float64(len(content))
	accuracy := 100.0
	if estimatedLength > 0 {
		accuracy = (1.0 - float64(errorCount)/estimatedLength) * 100.0
		if accuracy < 0 {
			accuracy = 0
		}
		if accuracy > 100 {
			accuracy = 100
		}
	}

	result := &TypingResult{
		UserID:      userID,
		Content:     content,
		TimeSpent:   timeSpentSeconds,
		WPM:         adjustedWPM,
		RawWPM:      rawWPM,
		Accuracy:    math.Round(accuracy*100) / 100,
		ErrorsCount: errorCount,
	}

	// Validate the result
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("invalid result: %w", err)
	}

	// Save to repository
	id, err := s.repo.SaveResult(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save result: %w", err)
	}

	result.ID = id
	return result, nil
}

// GetUserStatistics retrieves aggregated statistics for a user
func (s *Service) GetUserStatistics(ctx context.Context, userID uint) (*UserStats, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	// If stats is empty (new user), return empty stats object
	if stats.TotalTests == 0 {
		stats.UserID = userID
		return stats, nil
	}

	// Ensure values are reasonable
	if stats.AverageWPM < 0 {
		stats.AverageWPM = 0
	}
	if stats.BestWPM < 0 {
		stats.BestWPM = 0
	}
	if stats.AverageAccuracy < 0 {
		stats.AverageAccuracy = 0
	} else if stats.AverageAccuracy > 100 {
		stats.AverageAccuracy = 100
	}

	return stats, nil
}

// GetLeaderboard retrieves the leaderboard of top users
// Returns users sorted by best WPM in descending order
func (s *Service) GetLeaderboard(ctx context.Context, limit int) ([]UserStats, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 1000 {
		limit = 1000
	}

	stats, err := s.repo.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return stats, nil
}

// GetUserTestCount returns the total number of tests for a user
func (s *Service) GetUserTestCount(ctx context.Context, userID uint) (int, error) {
	if userID == 0 {
		return 0, errors.New("user_id is required")
	}

	count, err := s.repo.GetTestCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get test count: %w", err)
	}

	return count, nil
}

// GetUserTestHistory returns paginated test history for a user
func (s *Service) GetUserTestHistory(ctx context.Context, userID uint, limit, offset int) ([]TypingTest, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if limit <= 0 || limit > 1000 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	tests, err := s.repo.GetUserTests(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get test history: %w", err)
	}

	return tests, nil
}

// CalculateUserProgress calculates progress metrics for a user
func (s *Service) CalculateUserProgress(ctx context.Context, userID uint) (map[string]interface{}, error) {
	stats, err := s.GetUserStatistics(ctx, userID)
	if err != nil {
		return nil, err
	}

	progress := map[string]interface{}{
		"user_id":           userID,
		"total_tests":       stats.TotalTests,
		"average_wpm":       stats.AverageWPM,
		"best_wpm":          stats.BestWPM,
		"average_accuracy":  stats.AverageAccuracy,
		"total_time_typed":  stats.TotalTimeTyped,
		"estimated_level":   estimateUserLevel(stats.AverageWPM),
		"improvement_trend": calculateTrend(ctx, s.repo, userID),
	}

	return progress, nil
}

// estimateUserLevel estimates the user's typing level based on WPM
// Beginner: < 40 WPM
// Intermediate: 40-60 WPM
// Advanced: 60-80 WPM
// Expert: 80+ WPM
func estimateUserLevel(avgWPM float64) string {
	switch {
	case avgWPM < 40:
		return "beginner"
	case avgWPM < 60:
		return "intermediate"
	case avgWPM < 80:
		return "advanced"
	default:
		return "expert"
	}
}

// calculateTrend calculates the user's improvement trend
// Returns a map with trend direction and percentage improvement
func calculateTrend(ctx context.Context, repo *Repository, userID uint) map[string]interface{} {
	// Get last 10 tests
	tests, err := repo.GetUserTests(ctx, userID, 10, 0)
	if err != nil || len(tests) < 2 {
		return map[string]interface{}{
			"direction": "neutral",
			"change":    0.0,
		}
	}

	// Get first and last WPM from recent tests
	firstWPM := tests[len(tests)-1].WPM // Oldest test
	lastWPM := tests[0].WPM              // Most recent test

	if firstWPM == 0 {
		return map[string]interface{}{
			"direction": "neutral",
			"change":    0.0,
		}
	}

	// Calculate percentage improvement
	change := ((lastWPM - firstWPM) / firstWPM) * 100

	direction := "neutral"
	if change > 5 {
		direction = "improving"
	} else if change < -5 {
		direction = "declining"
	}

	return map[string]interface{}{
		"direction": direction,
		"change":    math.Round(change*100) / 100,
	}
}

// ValidateTestContent validates typing test content
// Checks for reasonable content length and character variety
func (s *Service) ValidateTestContent(content string) error {
	if len(strings.TrimSpace(content)) == 0 {
		return errors.New("content is required")
	}

	if len(content) < 10 {
		return errors.New("content is too short (minimum 10 characters)")
	}

	if len(content) > 10000 {
		return errors.New("content is too long (maximum 10000 characters)")
	}

	// Check if content has at least some variety
	hasLetters := false

	for _, r := range content {
		if unicode.IsLetter(r) {
			hasLetters = true
			break
		}
	}

	if !hasLetters {
		return errors.New("content must contain letters")
	}

	return nil
}

// NewServiceWithPool is a helper to create a Service with database pool
func NewServiceWithPool(db *database.Pool) *Service {
	repo := NewRepository(db)
	return NewService(repo)
}
