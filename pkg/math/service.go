package math

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Service provides business logic for math functionality
type Service struct {
	repo *Repository
}

// NewService creates a new math service
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GenerateProblem generates a random math problem based on type and difficulty
func (s *Service) GenerateProblem(problemType ProblemType, difficulty DifficultyLevel) (string, float64, error) {
	var num1, num2 float64
	var question string
	var answer float64

	switch difficulty {
	case Easy:
		num1 = float64(rand.Intn(10) + 1)
		num2 = float64(rand.Intn(10) + 1)
	case Medium:
		num1 = float64(rand.Intn(50) + 1)
		num2 = float64(rand.Intn(50) + 1)
	case Hard:
		num1 = float64(rand.Intn(100) + 1)
		num2 = float64(rand.Intn(100) + 1)
	case VeryHard:
		num1 = float64(rand.Intn(1000) + 1)
		num2 = float64(rand.Intn(1000) + 1)
	}

	switch problemType {
	case Addition:
		question = fmt.Sprintf("What is %.0f + %.0f?", num1, num2)
		answer = num1 + num2
	case Subtraction:
		if num1 < num2 {
			num1, num2 = num2, num1
		}
		question = fmt.Sprintf("What is %.0f - %.0f?", num1, num2)
		answer = num1 - num2
	case Multiplication:
		question = fmt.Sprintf("What is %.0f × %.0f?", num1, num2)
		answer = num1 * num2
	case Division:
		if num2 == 0 {
			num2 = 1
		}
		question = fmt.Sprintf("What is %.0f ÷ %.0f?", num1, num2)
		answer = num1 / num2
	case Fractions:
		n1 := int(num1) % 10
		d1 := int(num2)%9 + 1
		n2 := int(num1) % 10
		d2 := int(num2)%9 + 1
		question = fmt.Sprintf("What is %d/%d + %d/%d?", n1, d1, n2, d2)
		answer = float64(n1)/float64(d1) + float64(n2)/float64(d2)
	case Algebra:
		question = fmt.Sprintf("Solve: x + %.0f = %.0f", num1, num1+num2)
		answer = num2
	default:
		return "", 0, fmt.Errorf("unsupported problem type: %s", problemType)
	}

	return question, answer, nil
}

// RecordSolution records a user's attempt at solving a problem
func (s *Service) RecordSolution(ctx context.Context, solution *ProblemSolution) error {
	if err := solution.Validate(); err != nil {
		return err
	}

	// Save to repository
	_, err := s.repo.SaveSolution(ctx, solution)
	if err != nil {
		return fmt.Errorf("failed to save solution: %w", err)
	}

	return nil
}

// CompleteSession records a completed quiz session
func (s *Service) CompleteSession(ctx context.Context, session *QuizSession) error {
	if err := session.Validate(); err != nil {
		return err
	}

	// Calculate score
	session.Score = CalculateScore(session.CorrectAnswers, session.TotalProblems)
	session.CompletedAt = time.Now()

	// Save to repository
	_, err := s.repo.SaveSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Update user stats
	if err := s.repo.updateUserStats(ctx, session.UserID); err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}

	return nil
}

// GetUserStats retrieves user's math statistics
func (s *Service) GetUserStats(ctx context.Context, userID uint) (*UserMathStats, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// GetProblemTypeStats retrieves stats for a specific problem type
func (s *Service) GetProblemTypeStats(ctx context.Context, userID uint, problemType ProblemType) (*MathResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	result, err := s.repo.GetProblemTypeStats(ctx, userID, problemType)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem type stats: %w", err)
	}

	return result, nil
}

// GetLeaderboard retrieves top math performers
func (s *Service) GetLeaderboard(ctx context.Context, limit int) ([]UserMathStats, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	leaderboard, err := s.repo.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return leaderboard, nil
}

// GetUserSessions retrieves user's quiz sessions
func (s *Service) GetUserSessions(ctx context.Context, userID uint, limit, offset int) ([]QuizSession, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	sessions, err := s.repo.GetUserSessions(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return sessions, nil
}

// CalculateScore calculates quiz score based on correct answers
func CalculateScore(correctAnswers, totalProblems int) float64 {
	if totalProblems == 0 {
		return 0
	}

	accuracy := float64(correctAnswers) / float64(totalProblems)
	score := accuracy * 100.0

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return math.Round(score*10) / 10
}

// CalculateAccuracy calculates accuracy percentage
func CalculateAccuracy(correctAnswers, totalProblems int) float64 {
	if totalProblems == 0 {
		return 100.0
	}

	if correctAnswers > totalProblems {
		correctAnswers = totalProblems
	}

	accuracy := (float64(correctAnswers) / float64(totalProblems)) * 100.0

	if accuracy < 0 {
		accuracy = 0
	}
	if accuracy > 100 {
		accuracy = 100
	}

	return math.Round(accuracy*10) / 10
}

// CalculateAverageTimePerProblem calculates average time per problem
func CalculateAverageTimePerProblem(totalTime float64, totalProblems int) float64 {
	if totalProblems == 0 {
		return 0
	}

	avgTime := totalTime / float64(totalProblems)
	return math.Round(avgTime*10) / 10
}

// EstimateMathLevel estimates user's math skill level
func EstimateMathLevel(accuracy float64) string {
	switch {
	case accuracy < 50:
		return "beginner"
	case accuracy < 70:
		return "intermediate"
	case accuracy < 85:
		return "advanced"
	default:
		return "expert"
	}
}
