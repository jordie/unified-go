package math

import (
	"context"
	"fmt"
	"time"
)

// Service handles business logic for the math app
type Service struct {
	repo *Repository
}

// NewService creates a new service instance
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ProcessPracticeResult processes a practice session and updates user data
func (s *Service) ProcessPracticeResult(ctx context.Context, userID uint, result *MathResult) error {
	if err := result.Validate(); err != nil {
		return fmt.Errorf("invalid result: %w", err)
	}

	result.UserID = userID
	result.CalculateAccuracy()
	result.CalculateAverageTime()

	// Save the result
	if err := s.repo.SaveResult(ctx, result); err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	// Update learning profile with practice time
	profile, _ := s.repo.GetLearningProfile(ctx, userID)
	if profile == nil {
		profile = &LearningProfile{UserID: userID}
	}
	profile.TotalPracticeTime += int(result.TotalTime)
	profile.AvgSessionLength = int(float64(profile.TotalPracticeTime) / float64(result.TotalQuestions))

	if err := s.repo.SaveLearningProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to update learning profile: %w", err)
	}

	return nil
}

// SaveQuestionResponse saves a question attempt and updates mastery/mistakes
func (s *Service) SaveQuestionResponse(ctx context.Context, userID uint, history *QuestionHistory, quality int) error {
	if err := history.Validate(); err != nil {
		return fmt.Errorf("invalid question history: %w", err)
	}

	history.UserID = userID

	// Save question history
	if err := s.repo.SaveQuestionHistory(ctx, history); err != nil {
		return fmt.Errorf("failed to save question history: %w", err)
	}

	// Determine fact family if not provided
	if history.FactFamily == "" {
		history.FactFamily = ClassifyFactFamily(history.Question, history.Mode)
	}

	// Update mastery
	mastery, _ := s.repo.GetMastery(ctx, userID, history.Question, history.Mode)
	if mastery == nil {
		mastery = &Mastery{
			UserID: userID,
			Fact:   history.Question,
			Mode:   history.Mode,
		}
	}

	if history.IsCorrect {
		mastery.CorrectStreak++
	} else {
		mastery.CorrectStreak = 0

		// Record mistake
		mistake := &Mistake{
			UserID:        userID,
			Question:      history.Question,
			CorrectAnswer: history.CorrectAnswer,
			UserAnswer:    history.UserAnswer,
			Mode:          history.Mode,
			FactFamily:    history.FactFamily,
			ErrorCount:    1,
		}

		if err := s.repo.SaveMistake(ctx, mistake); err != nil {
			return fmt.Errorf("failed to save mistake: %w", err)
		}
	}

	// Update response time statistics
	mastery.UpdateResponseTime(history.TimeTaken)

	// Calculate mastery level
	baseAccuracy := float64(mastery.CorrectStreak) / float64(mastery.TotalAttempts)
	speedBonus := history.TimeTaken < mastery.AverageResponseTime && mastery.AverageResponseTime > 0
	mastery.CalculateMasteryLevel(baseAccuracy, speedBonus)

	if err := s.repo.SaveMastery(ctx, mastery); err != nil {
		return fmt.Errorf("failed to save mastery: %w", err)
	}

	return nil
}

// UpdatePerformancePattern updates time-of-day performance data
func (s *Service) UpdatePerformancePattern(ctx context.Context, userID uint, accuracy float64) error {
	now := time.Now()
	hour := now.Hour()
	dayOfWeek := int(now.Weekday())

	pattern, _ := s.repo.GetPerformancePattern(ctx, userID, hour, dayOfWeek)
	if pattern == nil {
		pattern = &PerformancePattern{
			UserID:    userID,
			HourOfDay: hour,
			DayOfWeek: dayOfWeek,
		}
	}

	// Update moving average for accuracy
	if pattern.SessionCount > 0 {
		pattern.AverageAccuracy = (pattern.AverageAccuracy*float64(pattern.SessionCount) + accuracy) / float64(pattern.SessionCount+1)
	} else {
		pattern.AverageAccuracy = accuracy
	}

	pattern.SessionCount++

	if err := pattern.Validate(); err != nil {
		return fmt.Errorf("invalid performance pattern: %w", err)
	}

	if err := s.repo.SavePerformancePattern(ctx, pattern); err != nil {
		return fmt.Errorf("failed to save performance pattern: %w", err)
	}

	return nil
}

// AnalyzeUserLearning generates learning profile analysis
func (s *Service) AnalyzeUserLearning(ctx context.Context, userID uint) (*LearningAnalysis, error) {
	analysis := &LearningAnalysis{
		UserID: userID,
	}

	// Get user stats
	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	analysis.Stats = stats

	// Get performance patterns
	patterns, err := s.repo.GetPatternsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get patterns: %w", err)
	}

	if len(patterns) > 0 {
		// Find best and worst times
		bestAccuracy := 0.0
		worstAccuracy := 100.0
		bestTime := ""
		worstTime := ""

		for _, p := range patterns {
			if p.AverageAccuracy > bestAccuracy {
				bestAccuracy = p.AverageAccuracy
				bestTime = GetTimeOfDayFromHour(p.HourOfDay)
			}
			if p.AverageAccuracy < worstAccuracy {
				worstAccuracy = p.AverageAccuracy
				worstTime = GetTimeOfDayFromHour(p.HourOfDay)
			}
		}

		analysis.BestTimeOfDay = bestTime
		analysis.BestTimeAccuracy = bestAccuracy
		analysis.WeakTimeOfDay = worstTime
		analysis.WeakTimeAccuracy = worstAccuracy
	}

	// Get weak fact families
	weakFamilies, err := s.repo.GetWeakFactFamilies(ctx, userID, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to get weak families: %w", err)
	}
	analysis.WeakFactFamilies = weakFamilies

	// Get learning profile
	profile, _ := s.repo.GetLearningProfile(ctx, userID)
	if profile != nil {
		analysis.LearningProfile = profile
	}

	return analysis, nil
}

// GeneratePracticeRecommendations generates personalized practice recommendations
func (s *Service) GeneratePracticeRecommendations(ctx context.Context, userID uint, mode string) (*PracticeRecommendation, error) {
	rec := &PracticeRecommendation{
		UserID: userID,
		Mode:   mode,
	}

	// Get weak fact families
	weakFamilies, err := s.repo.GetWeakFactFamilies(ctx, userID, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to get weak families: %w", err)
	}

	rec.WeakAreas = weakFamilies

	// Get recent activity
	activity, err := s.repo.GetRecentActivity(ctx, userID, 24)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	rec.RecentQuestionsAnswered = activity["questions_answered"]
	rec.RecentAccuracyPercent = activity["recent_accuracy_percent"]

	// Get best performance time
	bestTime, bestAccuracy, err := s.repo.GetBestPerformanceTime(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get best performance time: %w", err)
	}

	rec.BestPracticeTime = bestTime
	rec.BestTimeAccuracy = bestAccuracy

	// Generate recommendation text
	if len(weakFamilies) > 0 {
		rec.Recommendation = "Focus on weak fact families: "
		for family := range weakFamilies {
			rec.Recommendation += family + ", "
		}
	} else {
		rec.Recommendation = "Keep up the great work! Try harder difficulty levels."
	}

	if bestTime != "" {
		rec.Recommendation += " Practice during " + bestTime + " for best results."
	}

	return rec, nil
}

// GetDueForReview returns facts due for SM-2 review
func (s *Service) GetDueForReview(ctx context.Context, userID uint, limit int) ([]*RepetitionSchedule, error) {
	schedules, err := s.repo.GetDueRepetitions(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get due repetitions: %w", err)
	}

	return schedules, nil
}

// UpdateRepetitionAfterReview updates SM-2 schedule after review
func (s *Service) UpdateRepetitionAfterReview(ctx context.Context, userID uint, fact string, mode string, quality int) error {
	schedule, err := s.repo.GetRepetitionSchedule(ctx, userID, fact, mode)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	if schedule == nil {
		// Create new schedule if it doesn't exist
		schedule = &RepetitionSchedule{
			UserID:       userID,
			Fact:         fact,
			Mode:         mode,
			NextReview:   time.Now(),
			IntervalDays: 1,
			EaseFactor:   INITIAL_EASE_FACTOR,
			ReviewCount:  0,
		}
	}

	// Update schedule using SM-2
	schedule.ScheduleNextReview(quality)

	if err := s.repo.SaveRepetitionSchedule(ctx, schedule); err != nil {
		return fmt.Errorf("failed to save schedule: %w", err)
	}

	return nil
}

// LearningAnalysis represents comprehensive learning analysis
type LearningAnalysis struct {
	UserID              uint
	Stats               *UserStats
	LearningProfile     *LearningProfile
	BestTimeOfDay       string
	BestTimeAccuracy    float64
	WeakTimeOfDay       string
	WeakTimeAccuracy    float64
	WeakFactFamilies    map[string]int
}

// PracticeRecommendation represents personalized practice recommendations
type PracticeRecommendation struct {
	UserID                  uint
	Mode                    string
	WeakAreas               map[string]int
	RecentQuestionsAnswered int
	RecentAccuracyPercent   int
	BestPracticeTime        string
	BestTimeAccuracy        float64
	Recommendation          string
}

// ClassifyFactFamily determines the fact family for a question
func ClassifyFactFamily(question string, mode string) string {
	// This is a placeholder - will be expanded in service_phonics.go
	// For now, return a generic classification

	if mode == MODE_ADDITION {
		return "addition_basic"
	} else if mode == MODE_SUBTRACTION {
		return "subtraction_basic"
	} else if mode == MODE_MULTIPLICATION {
		return "multiplication_basic"
	} else if mode == MODE_DIVISION {
		return "division_basic"
	}

	return "mixed_basic"
}

// DeterminePracticeMode determines the best practice mode based on performance
func (s *Service) DeterminePracticeMode(ctx context.Context, userID uint) (string, error) {
	// Get mistake analysis
	analyses, err := s.repo.GetMistakeAnalysis(ctx, userID)
	if err != nil {
		return MODE_MIXED, nil
	}

	if len(analyses) == 0 {
		return MODE_MIXED, nil
	}

	// Find mode with most errors
	modeErrors := make(map[string]int)
	for range analyses {
		// Extract mode from analysis if available
		// For now, just count errors
		modeErrors["mixed"]++
	}

	return MODE_MIXED, nil
}

// CalculatePracticeIntensity calculates recommended practice intensity
func (s *Service) CalculatePracticeIntensity(ctx context.Context, userID uint) int {
	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return 10 // Default: 10 questions
	}

	if stats == nil {
		return 10
	}

	// Base intensity on accuracy
	if stats.AverageAccuracy < 60 {
		return 20 // More practice for low accuracy
	} else if stats.AverageAccuracy < 80 {
		return 15
	} else {
		return 10
	}
}

// EstimateCompletionTime estimates time to master a fact family
func (s *Service) EstimateCompletionTime(ctx context.Context, userID uint, factFamily string) int {
	// Get mistakes in this family
	mistakes, err := s.repo.GetMistakesByFactFamily(ctx, userID, factFamily)
	if err != nil || len(mistakes) == 0 {
		return 7 // Default: 7 days
	}

	// Calculate based on error count
	errorCount := 0
	for _, m := range mistakes {
		errorCount += m.ErrorCount
	}

	// Estimate: 2 days per error + base 5 days
	days := 5 + (errorCount * 2)
	if days > 30 {
		days = 30 // Cap at 30 days
	}

	return days
}

// GetNextLevelUpMilestone returns the next mastery milestone
func (s *Service) GetNextLevelUpMilestone(ctx context.Context, userID uint) (int, error) {
	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return 0, err
	}

	if stats == nil {
		return 1, nil
	}

	// Define milestone levels
	milestones := []int{5, 10, 25, 50, 100, 250, 500}

	for _, milestone := range milestones {
		if stats.TotalMastered < milestone {
			return milestone, nil
		}
	}

	return 1000, nil // Ultimate milestone
}
