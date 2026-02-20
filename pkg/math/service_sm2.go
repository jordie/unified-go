package math

import (
	"context"
	"fmt"
	"math"
	"time"
)

// SM2Engine handles the SuperMemo 2 spaced repetition algorithm
type SM2Engine struct {
	repo *Repository
}

// NewSM2Engine creates a new SM-2 engine
func NewSM2Engine(repo *Repository) *SM2Engine {
	return &SM2Engine{repo: repo}
}

// DetermineQualityFromPerformance determines quality rating based on performance metrics
func (e *SM2Engine) DetermineQualityFromPerformance(isCorrect bool, responseTimeMS int, averageTimeMS int) int {
	if !isCorrect {
		// Wrong answer = lowest quality
		if responseTimeMS > averageTimeMS*2 {
			return QUALITY_BLACKOUT // Very slow + wrong
		}
		return QUALITY_WRONG
	}

	// Correct answer - quality depends on speed
	if averageTimeMS == 0 {
		return QUALITY_CORRECT // First attempt
	}

	responseTimeSeconds := float64(responseTimeMS) / 1000.0
	averageTimeSeconds := float64(averageTimeMS) / 1000.0

	speedRatio := responseTimeSeconds / averageTimeSeconds

	if speedRatio < 0.8 {
		return QUALITY_PERFECT // Much faster
	} else if speedRatio < 1.2 {
		return QUALITY_CORRECT // Close to average
	} else if speedRatio < 1.5 {
		return QUALITY_DIFFICULT // Slower
	} else {
		return QUALITY_DIFFICULT // Much slower
	}
}

// InitializeSchedule creates a new spaced repetition schedule for a fact
func (e *SM2Engine) InitializeSchedule(ctx context.Context, userID uint, fact string, mode string) (*RepetitionSchedule, error) {
	// Check if schedule already exists
	existing, _ := e.repo.GetRepetitionSchedule(ctx, userID, fact, mode)
	if existing != nil {
		return existing, nil
	}

	// Create new schedule
	schedule := &RepetitionSchedule{
		UserID:       userID,
		Fact:         fact,
		Mode:         mode,
		NextReview:   time.Now(),
		IntervalDays: 1,
		EaseFactor:   INITIAL_EASE_FACTOR,
		ReviewCount:  0,
	}

	if err := e.repo.SaveRepetitionSchedule(ctx, schedule); err != nil {
		return nil, fmt.Errorf("failed to initialize schedule: %w", err)
	}

	return schedule, nil
}

// ProcessReview processes a review and updates the schedule
func (e *SM2Engine) ProcessReview(ctx context.Context, userID uint, fact string, mode string, quality int) (*RepetitionSchedule, error) {
	schedule, err := e.repo.GetRepetitionSchedule(ctx, userID, fact, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	if schedule == nil {
		// Initialize if doesn't exist
		schedule, err = e.InitializeSchedule(ctx, userID, fact, mode)
		if err != nil {
			return nil, err
		}
	}

	// Clamp quality to valid range
	if quality < 0 {
		quality = 0
	} else if quality > 5 {
		quality = 5
	}

	// Update ease factor using SM-2 formula
	schedule.UpdateEaseFactor(quality)

	// Calculate next interval
	var nextInterval int
	if quality < 3 {
		// Failed - reset to 1 day
		nextInterval = 1
	} else if schedule.ReviewCount == 0 {
		nextInterval = 1
	} else if schedule.ReviewCount == 1 {
		nextInterval = 6
	} else {
		// Subsequent reviews follow exponential growth
		nextInterval = int(math.Round(float64(schedule.IntervalDays) * schedule.EaseFactor))
	}

	schedule.IntervalDays = nextInterval
	schedule.NextReview = time.Now().AddDate(0, 0, nextInterval)
	schedule.ReviewCount++

	if err := schedule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid schedule after update: %w", err)
	}

	if err := e.repo.SaveRepetitionSchedule(ctx, schedule); err != nil {
		return nil, fmt.Errorf("failed to save schedule: %w", err)
	}

	return schedule, nil
}

// GenerateAdaptiveSession generates an adaptive practice session
func (e *SM2Engine) GenerateAdaptiveSession(ctx context.Context, userID uint, sessionSize int) (*AdaptiveSession, error) {
	session := &AdaptiveSession{
		UserID:    userID,
		StartTime: time.Now(),
		Limit:     sessionSize,
	}

	// Get due facts (40% of session)
	dueCount := (sessionSize * 4) / 10
	if dueCount > 0 {
		dueSchedules, err := e.repo.GetDueRepetitions(ctx, userID, dueCount)
		if err != nil {
			return nil, fmt.Errorf("failed to get due facts: %w", err)
		}

		for _, schedule := range dueSchedules {
			session.DueItems = append(session.DueItems, schedule.Fact)
		}
	}

	// Get new facts (60% of session)
	newCount := sessionSize - dueCount
	if newCount > 0 {
		// Get all mastery records to identify new facts
		allSchedules, err := e.repo.GetAllRepetitions(ctx, userID)
		if err == nil {
			// Count how many we have
			existingCount := len(allSchedules)
			if existingCount < sessionSize*3 {
				// We need more facts - add placeholders for new facts
				newFactsNeeded := newCount
				for i := 0; i < newFactsNeeded; i++ {
					// Placeholder - actual facts come from question generation
					session.NewItems = append(session.NewItems, fmt.Sprintf("new_fact_%d", i))
				}
			}
		}
	}

	return session, nil
}

// AnalyzeSM2Progress analyzes user's SM-2 progress and statistics
func (e *SM2Engine) AnalyzeSM2Progress(ctx context.Context, userID uint) (*SM2Progress, error) {
	progress := &SM2Progress{
		UserID: userID,
	}

	// Get all schedules
	schedules, err := e.repo.GetAllRepetitions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	progress.TotalFacts = len(schedules)

	if progress.TotalFacts == 0 {
		return progress, nil
	}

	// Analyze by review count and ease factor
	easeSum := 0.0
	reviewCounts := make(map[int]int)
	dueCount := 0
	var minEase, maxEase float64
	minEase = 3.5
	maxEase = 1.3

	for _, schedule := range schedules {
		easeSum += schedule.EaseFactor
		reviewCounts[schedule.ReviewCount]++

		if schedule.EaseFactor < minEase {
			minEase = schedule.EaseFactor
		}
		if schedule.EaseFactor > maxEase {
			maxEase = schedule.EaseFactor
		}

		if schedule.IsDueForReview() {
			dueCount++
		}
	}

	progress.AverageEaseFactor = easeSum / float64(progress.TotalFacts)
	progress.MinEaseFactor = minEase
	progress.MaxEaseFactor = maxEase
	progress.DueFacts = dueCount
	progress.ReviewCounts = reviewCounts

	// Calculate retention rate (facts with ease >= 2.5 are well-retained)
	wellRetained := 0
	for _, schedule := range schedules {
		if schedule.EaseFactor >= 2.5 {
			wellRetained++
		}
	}

	if progress.TotalFacts > 0 {
		progress.RetentionPercent = (float64(wellRetained) / float64(progress.TotalFacts)) * 100
	}

	return progress, nil
}

// GetOptimalReviewTime calculates optimal time to review a fact
func (e *SM2Engine) GetOptimalReviewTime(schedule *RepetitionSchedule) time.Duration {
	if schedule == nil {
		return 0
	}

	now := time.Now()
	if schedule.NextReview.Before(now) {
		return 0 // Already due
	}

	return schedule.NextReview.Sub(now)
}

// RecommendReviewCount recommends how many facts to review based on due count
func (e *SM2Engine) RecommendReviewCount(dueCount int) int {
	// Recommend reviewing 10-30 items based on backlog
	if dueCount == 0 {
		return 10 // Default daily review
	} else if dueCount < 20 {
		return dueCount + 5 // Review due + 5 new
	} else if dueCount < 50 {
		return 25 // Cap at 25
	} else {
		return 30 // Maximum session
	}
}

// AdaptiveSession represents an adaptive practice session
type AdaptiveSession struct {
	UserID    uint
	StartTime time.Time
	Limit     int
	DueItems  []string
	NewItems  []string
}

// SM2Progress represents user's spaced repetition progress
type SM2Progress struct {
	UserID              uint
	TotalFacts          int
	AverageEaseFactor   float64
	MinEaseFactor       float64
	MaxEaseFactor       float64
	DueFacts            int
	RetentionPercent    float64
	ReviewCounts        map[int]int
}

// SM2Constants for algorithm tuning
const (
	// Quality threshold for successful review
	QUALITY_THRESHOLD = 3

	// Interval constants
	MIN_INTERVAL = 1
	MAX_INTERVAL = 365
)

// CalculateEaseFactorChange calculates the change in ease factor
func CalculateEaseFactorChange(quality int) float64 {
	// Formula: 0.1 - (5 - q) * (0.08 + (5 - q) * 0.02)
	q := float64(quality)
	return 0.1 - (5.0-q)*(0.08+(5.0-q)*0.02)
}

// ClampEaseFactor clamps ease factor to valid range
func ClampEaseFactor(ease float64) float64 {
	if ease < MIN_EASE_FACTOR {
		return MIN_EASE_FACTOR
	}
	if ease > MAX_EASE_FACTOR {
		return MAX_EASE_FACTOR
	}
	return ease
}

// CalculateInterval calculates the next interval in days
func CalculateInterval(reviewCount int, currentInterval int, easeFactor float64, quality int) int {
	if quality < QUALITY_THRESHOLD {
		return 1 // Reset on failure
	}

	if reviewCount == 0 {
		return 1
	}
	if reviewCount == 1 {
		return 6
	}

	nextInterval := int(math.Round(float64(currentInterval) * easeFactor))
	if nextInterval < MIN_INTERVAL {
		nextInterval = MIN_INTERVAL
	}
	if nextInterval > MAX_INTERVAL {
		nextInterval = MAX_INTERVAL
	}

	return nextInterval
}
