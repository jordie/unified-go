package math

import (
	"fmt"
	"time"
)

// ProblemType represents the type of math problem
type ProblemType string

const (
	Addition       ProblemType = "addition"
	Subtraction    ProblemType = "subtraction"
	Multiplication ProblemType = "multiplication"
	Division       ProblemType = "division"
	Fractions      ProblemType = "fractions"
	Algebra        ProblemType = "algebra"
)

// DifficultyLevel represents problem difficulty
type DifficultyLevel string

const (
	Easy       DifficultyLevel = "easy"
	Medium     DifficultyLevel = "medium"
	Hard       DifficultyLevel = "hard"
	VeryHard   DifficultyLevel = "very_hard"
)

// Problem represents a single math problem
type Problem struct {
	ID        uint          `json:"id"`
	Type      ProblemType   `json:"type"`
	Difficulty DifficultyLevel `json:"difficulty"`
	Question  string        `json:"question"`
	Options   []string      `json:"options,omitempty"`
	Answer    float64       `json:"answer,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// ProblemSolution represents a user's solution to a problem
type ProblemSolution struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	ProblemID uint      `json:"problem_id"`
	Attempt   float64   `json:"attempt"`
	Correct   bool      `json:"correct"`
	TimeSpent float64   `json:"time_spent"`
	CreatedAt time.Time `json:"created_at"`
}

// QuizSession represents a practice session
type QuizSession struct {
	ID              uint          `json:"id"`
	UserID          uint          `json:"user_id"`
	ProblemType     ProblemType   `json:"problem_type"`
	Difficulty      DifficultyLevel `json:"difficulty"`
	TotalProblems   int           `json:"total_problems"`
	CorrectAnswers  int           `json:"correct_answers"`
	Score           float64       `json:"score"`
	TimeSpent       float64       `json:"time_spent"`
	StartedAt       time.Time     `json:"started_at"`
	CompletedAt     time.Time     `json:"completed_at,omitempty"`
	AverageTimePerProblem float64 `json:"average_time_per_problem"`
}

// UserMathStats represents aggregated user statistics
type UserMathStats struct {
	UserID                uint          `json:"user_id"`
	TotalProblems         int           `json:"total_problems"`
	CorrectAnswers        int           `json:"correct_answers"`
	Accuracy              float64       `json:"accuracy"`
	AverageTimePerProblem float64       `json:"average_time_per_problem"`
	BestScore             float64       `json:"best_score"`
	TotalTimeSpent        int           `json:"total_time_spent"`
	SessionsCompleted     int           `json:"sessions_completed"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// MathResult aggregates statistics for a problem type
type MathResult struct {
	ProblemType     ProblemType   `json:"problem_type"`
	Difficulty      DifficultyLevel `json:"difficulty"`
	TotalAttempts   int           `json:"total_attempts"`
	CorrectAnswers  int           `json:"correct_answers"`
	Accuracy        float64       `json:"accuracy"`
	AverageTimePerProblem float64 `json:"average_time_per_problem"`
}

// Validate validates a ProblemSolution
func (ps *ProblemSolution) Validate() error {
	if ps.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if ps.ProblemID == 0 {
		return fmt.Errorf("problem_id is required")
	}
	if ps.TimeSpent < 0 {
		return fmt.Errorf("time_spent cannot be negative")
	}
	return nil
}

// Validate validates a QuizSession
func (qs *QuizSession) Validate() error {
	if qs.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if qs.TotalProblems <= 0 {
		return fmt.Errorf("total_problems must be positive")
	}
	if qs.CorrectAnswers < 0 || qs.CorrectAnswers > qs.TotalProblems {
		return fmt.Errorf("correct_answers must be between 0 and total_problems")
	}
	if qs.Score < 0 || qs.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100")
	}
	if qs.TimeSpent < 0 {
		return fmt.Errorf("time_spent cannot be negative")
	}
	return nil
}

// Validate validates UserMathStats
func (ums *UserMathStats) Validate() error {
	if ums.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if ums.TotalProblems < 0 {
		return fmt.Errorf("total_problems cannot be negative")
	}
	if ums.CorrectAnswers < 0 || ums.CorrectAnswers > ums.TotalProblems {
		return fmt.Errorf("correct_answers must be between 0 and total_problems")
	}
	if ums.Accuracy < 0 || ums.Accuracy > 100 {
		return fmt.Errorf("accuracy must be between 0 and 100")
	}
	if ums.BestScore < 0 || ums.BestScore > 100 {
		return fmt.Errorf("best_score must be between 0 and 100")
	}
	return nil
}
