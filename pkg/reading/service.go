package reading

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

// Service provides business logic for reading operations
type Service struct {
	repo *Repository
}

// NewService creates a new reading service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CalculateWPM calculates words per minute from content and time spent
// Formula: (characters / 5) / (minutes elapsed)
// The standard is 5 characters = 1 word
func (s *Service) CalculateWPM(content string, timeSpentSeconds float64) float64 {
	if timeSpentSeconds <= 0 {
		return 0
	}

	// Convert seconds to minutes
	minutes := timeSpentSeconds / 60.0

	// Count non-whitespace characters
	charCount := float64(len(strings.TrimSpace(content)))

	// Calculate WPM: (characters / 5) / minutes
	wpm := (charCount / 5.0) / minutes

	// Round to 2 decimal places
	return math.Round(wpm*100) / 100
}

// CalculateAccuracy calculates reading accuracy as a percentage
// Compares typed content against expected content character by character
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

// ProcessTestResult processes a completed reading session
func (s *Service) ProcessTestResult(ctx context.Context, userID uint, bookID uint, content string, timeSpentSeconds float64, errorCount int) (*ReadingSession, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if bookID == 0 {
		return nil, errors.New("book_id is required")
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

	// Calculate WPM
	wpm := s.CalculateWPM(content, timeSpentSeconds)

	// Calculate accuracy based on error count
	totalCharacters := float64(len(content))
	accuracy := 100.0
	if totalCharacters > 0 {
		accuracy = (1.0 - float64(errorCount)/totalCharacters) * 100.0
		if accuracy < 0 {
			accuracy = 0
		}
		if accuracy > 100 {
			accuracy = 100
		}
	}

	// Estimate comprehension score based on accuracy and WPM
	// Better accuracy and reasonable WPM indicate better comprehension
	comprehensionScore := (accuracy * 0.7) + (math.Min(wpm/2.0, 100) * 0.3)

	session := &ReadingSession{
		UserID:               userID,
		BookID:               bookID,
		Duration:             timeSpentSeconds,
		WPM:                  math.Round(wpm*100) / 100,
		Accuracy:             accuracy,
		ComprehensionScore:   math.Round(comprehensionScore*100) / 100,
		ErrorCount:           errorCount,
		Completed:            true,
		StartTime:            time.Now(),
		EndTime:              time.Now(),
	}

	// Validate the session
	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Save to repository
	id, err := s.repo.SaveLesson(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to save result: %w", err)
	}

	session.ID = id
	return session, nil
}

// GetUserStatistics retrieves aggregated statistics for a user
func (s *Service) GetUserStatistics(ctx context.Context, userID uint) (*ReadingStats, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	// If stats is empty (new user), return empty stats object
	if stats.TotalSessionsCount == 0 {
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

// GetLeaderboard retrieves the leaderboard of top readers
func (s *Service) GetLeaderboard(ctx context.Context, limit int) ([]ReadingStats, error) {
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

// GetUserTestHistory returns paginated test history for a user
func (s *Service) GetUserTestHistory(ctx context.Context, userID uint, limit, offset int) ([]ReadingSession, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if limit <= 0 || limit > 1000 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	sessions, err := s.repo.GetUserSessions(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get test history: %w", err)
	}

	return sessions, nil
}

// ValidateTestContent validates reading test content
func (s *Service) ValidateTestContent(content string) error {
	if len(strings.TrimSpace(content)) == 0 {
		return errors.New("content is required")
	}

	if len(content) < 50 {
		return errors.New("content is too short (minimum 50 characters)")
	}

	if len(content) > 100000 {
		return errors.New("content is too long (maximum 100000 characters)")
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

// EstimateUserLevel estimates the user's reading level based on WPM
func EstimateUserLevel(avgWPM float64) string {
	switch {
	case avgWPM < 100:
		return "beginner"
	case avgWPM < 200:
		return "intermediate"
	case avgWPM < 300:
		return "advanced"
	default:
		return "expert"
	}
}

// CalculateUserProgress calculates progress metrics for a user
func (s *Service) CalculateUserProgress(ctx context.Context, userID uint) (map[string]interface{}, error) {
	stats, err := s.GetUserStatistics(ctx, userID)
	if err != nil {
		return nil, err
	}

	progress := map[string]interface{}{
		"user_id":             userID,
		"total_tests":         stats.TotalSessionsCount,
		"average_wpm":         stats.AverageWPM,
		"best_wpm":            stats.BestWPM,
		"average_accuracy":    stats.AverageAccuracy,
		"average_comprehension": stats.AverageComprehension,
		"total_reading_time":  stats.TotalReadingTime,
		"estimated_level":     EstimateUserLevel(stats.AverageWPM),
	}

	// Calculate improvement trend if we have enough data
	if stats.TotalSessionsCount >= 2 {
		trend := s.calculateTrend(ctx, userID)
		progress["trend"] = trend
	}

	return progress, nil
}

// calculateTrend calculates improvement trend for a user
func (s *Service) calculateTrend(ctx context.Context, userID uint) map[string]interface{} {
	// Get last 10 sessions
	sessions, err := s.repo.GetUserSessions(ctx, userID, 10, 0)
	if err != nil || len(sessions) < 2 {
		return map[string]interface{}{
			"direction": "neutral",
			"change":    0.0,
		}
	}

	// Get first and last WPM from recent sessions
	// Note: sessions are returned in DESC order (most recent first)
	firstWPM := sessions[len(sessions)-1].WPM // Oldest in the list
	lastWPM := sessions[0].WPM                 // Most recent

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

// GetComprehensionAnalysis analyzes comprehension test results
func (s *Service) GetComprehensionAnalysis(ctx context.Context, sessionID uint) (map[string]interface{}, error) {
	tests, err := s.repo.GetComprehensionTests(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comprehension tests: %w", err)
	}

	if len(tests) == 0 {
		return map[string]interface{}{
			"total_questions": 0,
			"correct_answers": 0,
			"score":           0.0,
		}, nil
	}

	correct := 0
	for _, test := range tests {
		if test.IsCorrect {
			correct++
		}
	}

	score := (float64(correct) / float64(len(tests))) * 100

	return map[string]interface{}{
		"total_questions": len(tests),
		"correct_answers": correct,
		"score":           math.Round(score*100) / 100,
	}, nil
}

// RecommendBooks suggests books based on user's reading level
func (s *Service) RecommendBooks(ctx context.Context, userID uint, limit int) ([]Book, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 5
	}

	// Get user's statistics to determine reading level
	stats, err := s.GetUserStatistics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	// Determine recommended difficulty based on user level
	userLevel := EstimateUserLevel(stats.AverageWPM)

	// Users generally benefit from reading at their current level or slightly above
	var difficulties []string
	switch userLevel {
	case "beginner":
		difficulties = []string{"beginner", "intermediate"}
	case "intermediate":
		difficulties = []string{"intermediate", "advanced"}
	case "advanced", "expert":
		difficulties = []string{"advanced"}
	default:
		difficulties = []string{"intermediate"}
	}

	// Get books for the recommended difficulty levels
	books, err := s.repo.GetBooks(ctx, difficulties[0], limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommended books: %w", err)
	}

	// If not enough books at primary level, add from secondary level
	if len(books) < limit && len(difficulties) > 1 {
		moreBooks, err := s.repo.GetBooks(ctx, difficulties[1], limit-len(books), 0)
		if err == nil {
			books = append(books, moreBooks...)
		}
	}

	return books, nil
}

// SaveBook creates a new book in the system
func (s *Service) SaveBook(ctx context.Context, book *Book) (uint, error) {
	if book == nil {
		return 0, errors.New("book cannot be nil")
	}

	if err := book.Validate(); err != nil {
		return 0, fmt.Errorf("invalid book: %w", err)
	}

	// Auto-calculate reading level based on word count if not provided
	if book.ReadingLevel == "" {
		book.ReadingLevel = CalculateReadingLevel(book.WordCount)
	}

	return s.repo.SaveBook(ctx, book)
}

// GetBookRecommendations provides personalized book suggestions with reasoning
func (s *Service) GetBookRecommendations(ctx context.Context, userID uint) (map[string]interface{}, error) {
	books, err := s.RecommendBooks(ctx, userID, 5)
	if err != nil {
		return nil, err
	}

	stats, err := s.GetUserStatistics(ctx, userID)
	if err != nil {
		return nil, err
	}

	userLevel := EstimateUserLevel(stats.AverageWPM)

	recommendations := map[string]interface{}{
		"user_level":      userLevel,
		"average_wpm":     stats.AverageWPM,
		"recommended_books": books,
		"reason":          fmt.Sprintf("These books match your %s reading level", userLevel),
	}

	return recommendations, nil
}

// AnalyzeReadingPerformance provides detailed performance insights
func (s *Service) AnalyzeReadingPerformance(ctx context.Context, userID uint) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	stats, err := s.GetUserStatistics(ctx, userID)
	if err != nil {
		return nil, err
	}

	progress, err := s.CalculateUserProgress(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build comprehensive analysis
	analysis := map[string]interface{}{
		"user_id":         userID,
		"statistics":      stats,
		"progress":        progress,
		"strength":        s.identifyStrengths(stats),
		"improvement_area": s.identifyImprovementAreas(stats),
	}

	return analysis, nil
}

// identifyStrengths identifies what the user is doing well at
func (s *Service) identifyStrengths(stats *ReadingStats) []string {
	var strengths []string

	if stats.AverageWPM > 200 {
		strengths = append(strengths, "Excellent reading speed")
	}

	if stats.AverageAccuracy > 95 {
		strengths = append(strengths, "Outstanding accuracy")
	}

	if stats.AverageComprehension > 85 {
		strengths = append(strengths, "Strong comprehension")
	}

	if stats.TotalSessionsCount > 10 {
		strengths = append(strengths, "Consistent practice")
	}

	if len(strengths) == 0 {
		strengths = append(strengths, "Getting started on your reading journey")
	}

	return strengths
}

// identifyImprovementAreas identifies areas where the user can improve
func (s *Service) identifyImprovementAreas(stats *ReadingStats) []string {
	var areas []string

	if stats.AverageWPM < 100 {
		areas = append(areas, "Reading speed - try practicing with simpler texts")
	}

	if stats.AverageAccuracy < 80 {
		areas = append(areas, "Accuracy - slow down and focus on each word")
	}

	if stats.AverageComprehension < 75 {
		areas = append(areas, "Comprehension - review passages and take notes")
	}

	if stats.TotalSessionsCount < 5 {
		areas = append(areas, "Consistency - practice more regularly for better results")
	}

	if len(areas) == 0 {
		areas = append(areas, "You're doing great! Challenge yourself with harder texts")
	}

	return areas
}

// GetSessionDetails returns comprehensive information about a single reading session
func (s *Service) GetSessionDetails(ctx context.Context, sessionID uint) (map[string]interface{}, error) {
	if sessionID == 0 {
		return nil, errors.New("session_id is required")
	}

	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Get the book for this session
	book, err := s.repo.GetBookByID(ctx, session.BookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	// Get comprehension analysis
	comprehension, err := s.GetComprehensionAnalysis(ctx, sessionID)
	if err != nil {
		comprehension = map[string]interface{}{
			"total_questions": 0,
			"correct_answers": 0,
			"score":           0.0,
		}
	}

	details := map[string]interface{}{
		"session":       session,
		"book":          book,
		"comprehension": comprehension,
		"performance_rating": s.getRatingFromMetrics(session.WPM, session.Accuracy),
	}

	return details, nil
}

// getRatingFromMetrics converts metrics into a 1-5 star rating
func (s *Service) getRatingFromMetrics(wpm, accuracy float64) int {
	score := (wpm / 100) * 0.5 + (accuracy / 100) * 0.5

	switch {
	case score >= 2.0:
		return 5
	case score >= 1.6:
		return 4
	case score >= 1.2:
		return 3
	case score >= 0.8:
		return 2
	default:
		return 1
	}
}
