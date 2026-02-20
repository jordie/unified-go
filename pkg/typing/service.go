package typing

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// Service provides business logic for typing functionality
type Service struct {
	repo *Repository
}

// NewService creates a new typing service
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ProcessTypingTest processes a typing test and calculates metrics
func (s *Service) ProcessTypingTest(ctx context.Context, userID uint, content string, duration float64, errors int) (*TypingResult, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if duration <= 0 {
		return nil, errors.New("duration must be positive")
	}

	if errors < 0 {
		return nil, errors.New("errors cannot be negative")
	}

	// Calculate metrics
	wpm := CalculateWPM(len(content), duration)
	rawWPM := CalculateRawWPM(len(content), duration)
	accuracy := CalculateAccuracy(len(content), errors)

	result := &TypingResult{
		UserID:      userID,
		Content:     content,
		TimeSpent:   duration,
		WPM:         wpm,
		RawWPM:      rawWPM,
		ErrorsCount: errors,
		Accuracy:    accuracy,
		TestMode:    "standard",
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

// GetUserProgress retrieves user's typing progress and statistics
func (s *Service) GetUserProgress(ctx context.Context, userID uint) (*UserStats, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// GetLeaderboard retrieves top typers by WPM
func (s *Service) GetLeaderboard(ctx context.Context, limit int) ([]UserStats, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	leaderboard, err := s.repo.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return leaderboard, nil
}

// GetUserHistory retrieves user's test history
func (s *Service) GetUserHistory(ctx context.Context, userID uint, days int) ([]TypingResult, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if days <= 0 {
		days = 30 // Default to last 30 days
	}

	history, err := s.repo.GetTestHistory(ctx, userID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	return history, nil
}

// CalculateWPM calculates corrected words per minute
func CalculateWPM(charCount int, durationSeconds float64) float64 {
	if durationSeconds == 0 {
		return 0
	}

	minutes := durationSeconds / 60.0
	words := float64(charCount) / 5.0 // Standard: 5 characters = 1 word
	wpm := words / minutes

	if wpm < 0 {
		return 0
	}

	return math.Round(wpm*10) / 10 // Round to 1 decimal place
}

// CalculateRawWPM calculates raw words per minute (before accuracy adjustment)
func CalculateRawWPM(charCount int, durationSeconds float64) float64 {
	return CalculateWPM(charCount, durationSeconds)
}

// CalculateAccuracy calculates typing accuracy percentage
func CalculateAccuracy(charCount int, errors int) float64 {
	if charCount == 0 {
		return 100.0
	}

	if errors > charCount {
		errors = charCount
	}

	correctChars := charCount - errors
	accuracy := (float64(correctChars) / float64(charCount)) * 100.0

	if accuracy < 0 {
		accuracy = 0
	}
	if accuracy > 100 {
		accuracy = 100
	}

	return math.Round(accuracy*10) / 10 // Round to 1 decimal place
}

// EstimateTypingLevel estimates user typing skill level based on WPM
func EstimateTypingLevel(averageWPM float64) string {
	switch {
	case averageWPM < 40:
		return "beginner"
	case averageWPM < 60:
		return "intermediate"
	case averageWPM < 80:
		return "advanced"
	default:
		return "expert"
	}
}

// CalculateProgressTrend calculates progress trend
func CalculateProgressTrend(previousWPM, currentWPM float64) string {
	if currentWPM > previousWPM {
		return "improving"
	} else if currentWPM < previousWPM {
		return "declining"
	}
	return "stable"
}
