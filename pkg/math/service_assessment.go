package math

import (
	"context"
	"fmt"
	"time"
)

// AssessmentEngine handles adaptive math assessments with binary search placement
type AssessmentEngine struct {
	repo *Repository
}

// NewAssessmentEngine creates a new assessment engine
func NewAssessmentEngine(repo *Repository) *AssessmentEngine {
	return &AssessmentEngine{repo: repo}
}

// AssessmentSession represents an active assessment session
type AssessmentSession struct {
	ID               uint
	UserID           uint
	StartTime        time.Time
	CurrentLevel     int
	MinLevel         int
	MaxLevel         int
	IsIncreasing     bool
	ResponseCount    int
	CorrectCount     int
	LastResponse     bool
	PlacementResult  *PlacementResult
	Questions        []*AssessmentQuestion
}

// AssessmentQuestion represents a question in an assessment
type AssessmentQuestion struct {
	ID       uint
	Level    int
	Question string
	Mode     string
}

// PlacementResult represents the result of an assessment
type PlacementResult struct {
	UserID            uint
	PlacedLevel       int
	Confidence        float64
	EstimatedAccuracy float64
	RecommendedMode   string
	TotalResponses    int
	CorrectResponses  int
	StartTime         time.Time
	EndTime           time.Time
}

// StartAssessment starts a new assessment session using binary search (15 levels)
func (e *AssessmentEngine) StartAssessment(ctx context.Context, userID uint, mode string) (*AssessmentSession, error) {
	session := &AssessmentSession{
		UserID:       userID,
		StartTime:    time.Now(),
		CurrentLevel: 7, // Start at middle of 1-15 range
		MinLevel:     1,
		MaxLevel:     15,
		IsIncreasing: true,
		ResponseCount: 0,
		CorrectCount: 0,
	}

	// Get starting question at middle level
	question := e.getQuestionForLevel(ctx, mode, session.CurrentLevel)
	if question != nil {
		session.Questions = append(session.Questions, question)
	}

	return session, nil
}

// ProcessResponse processes an assessment response and updates placement
func (e *AssessmentEngine) ProcessResponse(ctx context.Context, session *AssessmentSession, isCorrect bool, mode string) (*PlacementResult, error) {
	session.ResponseCount++
	session.LastResponse = isCorrect

	if isCorrect {
		session.CorrectCount++
		session.IsIncreasing = true
		// Move to higher level
		session.MinLevel = session.CurrentLevel
		session.CurrentLevel = (session.CurrentLevel + session.MaxLevel) / 2
	} else {
		session.IsIncreasing = false
		// Move to lower level
		session.MaxLevel = session.CurrentLevel
		session.CurrentLevel = (session.MinLevel + session.CurrentLevel) / 2
	}

	// Stop after 20 responses or when converged (range of 1)
	if session.ResponseCount >= 20 || (session.MaxLevel - session.MinLevel) <= 1 {
		result := e.DeterminePlacement(ctx, session)
		session.PlacementResult = result
		return result, nil
	}

	// Get next question
	question := e.getQuestionForLevel(ctx, mode, session.CurrentLevel)
	if question != nil {
		session.Questions = append(session.Questions, question)
	}

	return nil, nil
}

// DeterminePlacement determines final placement level based on assessment responses
func (e *AssessmentEngine) DeterminePlacement(ctx context.Context, session *AssessmentSession) *PlacementResult {
	// Final level is where user had consistent success
	finalLevel := session.MinLevel
	if session.MaxLevel > session.MinLevel {
		finalLevel = (session.MinLevel + session.MaxLevel) / 2
	}

	// Clamp to 1-15 range
	if finalLevel < 1 {
		finalLevel = 1
	}
	if finalLevel > 15 {
		finalLevel = 15
	}

	// Calculate confidence and estimated accuracy
	accuracy := float64(session.CorrectCount) / float64(session.ResponseCount)
	confidence := 0.0

	// Confidence increases with consistent responses near placement level
	if session.ResponseCount >= 15 {
		confidence = 0.95
	} else if session.ResponseCount >= 10 {
		confidence = 0.85
	} else if session.ResponseCount >= 5 {
		confidence = 0.70
	} else {
		confidence = 0.50
	}

	// Adjust confidence based on accuracy at final level
	if accuracy >= 0.85 {
		confidence += 0.10
	} else if accuracy < 0.50 {
		confidence -= 0.10
	}

	// Clamp confidence to 0-1
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	result := &PlacementResult{
		UserID:            session.UserID,
		PlacedLevel:       finalLevel,
		Confidence:        confidence,
		EstimatedAccuracy: accuracy,
		TotalResponses:    session.ResponseCount,
		CorrectResponses:  session.CorrectCount,
		StartTime:         session.StartTime,
		EndTime:           time.Now(),
	}

	// Recommend practice mode based on placement
	if finalLevel <= 5 {
		result.RecommendedMode = "basic"
	} else if finalLevel <= 10 {
		result.RecommendedMode = "intermediate"
	} else {
		result.RecommendedMode = "advanced"
	}

	return result
}

// getQuestionForLevel gets a question at a specific difficulty level (1-15)
func (e *AssessmentEngine) getQuestionForLevel(ctx context.Context, mode string, level int) *AssessmentQuestion {
	// Map 1-15 levels to difficulty progression
	// Levels 1-3: Easy (single digit addition)
	// Levels 4-7: Medium (two digit, simple operations)
	// Levels 8-11: Hard (two digit, complex operations)
	// Levels 12-15: Expert (three digit, multiple operations)

	difficulty := "easy"
	if level > 3 && level <= 7 {
		difficulty = "medium"
	} else if level > 7 && level <= 11 {
		difficulty = "hard"
	} else if level > 11 {
		difficulty = "expert"
	}

	question := &AssessmentQuestion{
		Level:    level,
		Question: fmt.Sprintf("Assessment question at level %d (%s)", level, difficulty),
		Mode:     mode,
	}

	return question
}

// GetAssessmentHistory retrieves assessment history for a user
func (e *AssessmentEngine) GetAssessmentHistory(ctx context.Context, userID uint, limit int) ([]*PlacementResult, error) {
	// Query database for assessment history
	// For now, return empty slice as placeholder
	return []*PlacementResult{}, nil
}

// GetCurrentLevel determines user's current level from assessment history
func (e *AssessmentEngine) GetCurrentLevel(ctx context.Context, userID uint) (int, error) {
	// Get user stats to estimate current level
	stats, err := e.repo.GetUserStats(ctx, userID)
	if err != nil {
		return 7, nil // Default to middle level
	}

	if stats == nil {
		return 7, nil
	}

	// Estimate level based on accuracy
	if stats.AverageAccuracy < 50 {
		return 3, nil
	} else if stats.AverageAccuracy < 70 {
		return 7, nil
	} else if stats.AverageAccuracy < 85 {
		return 10, nil
	} else {
		return 13, nil
	}
}

// AdaptiveDifficulty adjusts difficulty based on performance in real-time
func (e *AssessmentEngine) AdaptiveDifficulty(ctx context.Context, userID uint, currentLevel int, isCorrect bool) int {
	// If user gets 2 in a row correct, increase difficulty
	// If user gets 2 in a row wrong, decrease difficulty
	// This is used during practice sessions, not assessments

	if isCorrect {
		// Gradually increase difficulty
		newLevel := currentLevel + 1
		if newLevel > 15 {
			newLevel = 15
		}
		return newLevel
	} else {
		// Gradually decrease difficulty
		newLevel := currentLevel - 1
		if newLevel < 1 {
			newLevel = 1
		}
		return newLevel
	}
}

// AssessmentConfig represents assessment configuration
type AssessmentConfig struct {
	MinLevel         int     // Minimum difficulty level (1-15)
	MaxLevel         int     // Maximum difficulty level (1-15)
	TargetAccuracy   float64 // Target accuracy percentage (0-1)
	MinResponses     int     // Minimum responses before placement
	MaxResponses     int     // Maximum responses before placement
	ConvergenceRange int     // Level range for convergence
}

// GetAssessmentConfig returns default assessment configuration
func GetAssessmentConfig() *AssessmentConfig {
	return &AssessmentConfig{
		MinLevel:         1,
		MaxLevel:         15,
		TargetAccuracy:   0.75,
		MinResponses:     5,
		MaxResponses:     20,
		ConvergenceRange: 1,
	}
}

// ValidatePlacement checks if a placement is valid and within expected ranges
func (p *PlacementResult) Validate() error {
	if p.PlacedLevel < 1 || p.PlacedLevel > 15 {
		return fmt.Errorf("invalid placement level: %d (must be 1-15)", p.PlacedLevel)
	}

	if p.Confidence < 0 || p.Confidence > 1.0 {
		return fmt.Errorf("invalid confidence: %.2f (must be 0-1)", p.Confidence)
	}

	if p.EstimatedAccuracy < 0 || p.EstimatedAccuracy > 1.0 {
		return fmt.Errorf("invalid estimated accuracy: %.2f (must be 0-1)", p.EstimatedAccuracy)
	}

	if p.TotalResponses < 1 {
		return fmt.Errorf("must have at least 1 response")
	}

	if p.CorrectResponses < 0 || p.CorrectResponses > p.TotalResponses {
		return fmt.Errorf("invalid response counts")
	}

	return nil
}
