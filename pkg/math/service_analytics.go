package math

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// AnalyticsEngine handles learning analytics and pattern detection
type AnalyticsEngine struct {
	repo *Repository
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(repo *Repository) *AnalyticsEngine {
	return &AnalyticsEngine{repo: repo}
}

// UserAnalytics represents comprehensive user analytics
type UserAnalytics struct {
	UserID              uint
	TotalSessionsCount  int
	TotalQuestionsCount int
	CorrectAnswersCount int
	OverallAccuracy     float64
	AverageSessionTime  int
	BestAccuracy        float64
	WorstAccuracy       float64
	StreakData          *StreakAnalysis
	TimeOfDayAnalysis   *TimeOfDayAnalysis
	FactFamilyAnalysis  *FactFamilyAnalysis
	ProgressTrend       *ProgressTrend
	RetentionMetrics    *RetentionMetrics
}

// StreakAnalysis represents streak statistics
type StreakAnalysis struct {
	CurrentStreak  int
	LongestStreak  int
	StreakStartDay time.Time
	StreakEndDay   time.Time
}

// TimeOfDayAnalysis represents performance by time of day
type TimeOfDayAnalysis struct {
	BestTimeOfDay       string
	BestTimeAccuracy    float64
	WorstTimeOfDay      string
	WorstTimeAccuracy   float64
	MorningAccuracy     float64
	AfternoonAccuracy   float64
	EveningAccuracy     float64
	MorningSessionCount int
	AfternoonSessionCount int
	EveningSessionCount   int
}

// FactFamilyAnalysis represents fact family performance
type FactFamilyAnalysis struct {
	TotalFamilies    int
	MasteredFamilies int
	StrengthAreas    map[string]float64
	WeakAreas        map[string]float64
	MostPracticed    string
	MostMastered     string
	MostMistaken     string
}

// ProgressTrend represents user progress over time
type ProgressTrend struct {
	DaysActive        int
	AccuracyTrend     float64 // -100 to +100 (percentage change)
	SessionTrend      float64 // -100 to +100 (percentage change)
	SpeedTrend        float64 // -100 to +100 (percentage improvement)
	LastWeekAccuracy  float64
	LastMonthAccuracy float64
	ProjectedLevel    int
}

// RetentionMetrics represents information retention analysis
type RetentionMetrics struct {
	FactsLearned           int
	FactsRetained          int
	RetentionPercentage    float64
	AverageEaseFactor      float64
	AverageIntervalDays    int
	DueFactsCount          int
	OverdueFactsCount      int
	EstimatedLearningDays  int
}

// GetUserAnalytics generates comprehensive analytics for a user
func (e *AnalyticsEngine) GetUserAnalytics(ctx context.Context, userID uint) (*UserAnalytics, error) {
	analytics := &UserAnalytics{
		UserID: userID,
	}

	// Get basic stats
	stats, err := e.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	if stats != nil {
		analytics.TotalSessionsCount = stats.TotalSessions
		analytics.TotalQuestionsCount = stats.TotalQuestions
		analytics.CorrectAnswersCount = stats.CorrectAnswers
		analytics.OverallAccuracy = stats.AverageAccuracy
		analytics.BestAccuracy = stats.BestAccuracy
	}

	// Get SM-2 progress
	sm2Progress, err := e.GetSM2Analytics(ctx, userID)
	if err == nil && sm2Progress != nil {
		analytics.RetentionMetrics = &RetentionMetrics{
			FactsLearned:          sm2Progress.TotalFacts,
			FactsRetained:         int(float64(sm2Progress.TotalFacts) * sm2Progress.RetentionPercent / 100),
			RetentionPercentage:   sm2Progress.RetentionPercent,
			AverageEaseFactor:     sm2Progress.AverageEaseFactor,
			AverageIntervalDays:   3, // Placeholder
			DueFactsCount:         sm2Progress.DueFacts,
		}
	}

	// Get time-of-day analysis
	patterns, err := e.repo.GetPatternsByUser(ctx, userID)
	if err == nil {
		analytics.TimeOfDayAnalysis = e.analyzeTimeOfDay(patterns)
	}

	// Get fact family analysis
	weakFamilies, err := e.repo.GetWeakFactFamilies(ctx, userID, 5)
	if err == nil {
		analytics.FactFamilyAnalysis = e.analyzeFactFamilies(ctx, userID, weakFamilies)
	}

	// Get progress trend
	analytics.ProgressTrend = e.calculateProgressTrend(ctx, userID)

	// Get streak analysis
	analytics.StreakData = e.analyzeStreak(ctx, userID)

	return analytics, nil
}

// GetSM2Analytics returns SM-2 spaced repetition analytics
func (e *AnalyticsEngine) GetSM2Analytics(ctx context.Context, userID uint) (*SM2Progress, error) {
	engine := NewSM2Engine(e.repo)
	return engine.AnalyzeSM2Progress(ctx, userID)
}

// analyzeTimeOfDay analyzes performance by time of day
func (e *AnalyticsEngine) analyzeTimeOfDay(patterns []*PerformancePattern) *TimeOfDayAnalysis {
	analysis := &TimeOfDayAnalysis{
		BestTimeAccuracy:  0,
		WorstTimeAccuracy: 100,
	}

	morningAccuracies := []float64{}
	afternoonAccuracies := []float64{}
	eveningAccuracies := []float64{}

	for _, p := range patterns {
		timeOfDay := GetTimeOfDayFromHour(p.HourOfDay)

		if p.AverageAccuracy > analysis.BestTimeAccuracy {
			analysis.BestTimeAccuracy = p.AverageAccuracy
			analysis.BestTimeOfDay = timeOfDay
		}

		if p.AverageAccuracy < analysis.WorstTimeAccuracy {
			analysis.WorstTimeAccuracy = p.AverageAccuracy
			analysis.WorstTimeOfDay = timeOfDay
		}

		// Categorize by time of day
		if timeOfDay == "morning" {
			morningAccuracies = append(morningAccuracies, p.AverageAccuracy)
			analysis.MorningSessionCount += p.SessionCount
		} else if timeOfDay == "afternoon" {
			afternoonAccuracies = append(afternoonAccuracies, p.AverageAccuracy)
			analysis.AfternoonSessionCount += p.SessionCount
		} else if timeOfDay == "evening" {
			eveningAccuracies = append(eveningAccuracies, p.AverageAccuracy)
			analysis.EveningSessionCount += p.SessionCount
		}
	}

	// Calculate averages by time period
	if len(morningAccuracies) > 0 {
		analysis.MorningAccuracy = average(morningAccuracies)
	}
	if len(afternoonAccuracies) > 0 {
		analysis.AfternoonAccuracy = average(afternoonAccuracies)
	}
	if len(eveningAccuracies) > 0 {
		analysis.EveningAccuracy = average(eveningAccuracies)
	}

	return analysis
}

// analyzeFactFamilies analyzes fact family performance
func (e *AnalyticsEngine) analyzeFactFamilies(ctx context.Context, userID uint, weakFamilies map[string]int) *FactFamilyAnalysis {
	analysis := &FactFamilyAnalysis{
		StrengthAreas: make(map[string]float64),
		WeakAreas:     make(map[string]float64),
	}

	// Get all masteries to count total and mastered families
	masteries, _ := e.repo.GetMasteryByUser(ctx, userID, MODE_MIXED)
	analysis.TotalFamilies = len(masteries)

	masteredCount := 0
	for _, m := range masteries {
		if m.MasteryLevel >= 80 {
			masteredCount++
			analysis.StrengthAreas[m.Fact] = float64(m.MasteryLevel)
		}
	}
	analysis.MasteredFamilies = masteredCount

	// Add weak areas
	for family, count := range weakFamilies {
		analysis.WeakAreas[family] = float64(count)
	}

	// Find most/least practiced
	if len(masteries) > 0 {
		sort.Slice(masteries, func(i, j int) bool {
			return masteries[i].TotalAttempts > masteries[j].TotalAttempts
		})
		analysis.MostPracticed = masteries[0].Fact

		sort.Slice(masteries, func(i, j int) bool {
			return masteries[i].MasteryLevel > masteries[j].MasteryLevel
		})
		analysis.MostMastered = masteries[0].Fact
	}

	return analysis
}

// calculateProgressTrend calculates learning progress trends
func (e *AnalyticsEngine) calculateProgressTrend(ctx context.Context, userID uint) *ProgressTrend {
	trend := &ProgressTrend{}

	// Get recent activity for trend calculation
	activity, _ := e.repo.GetRecentActivity(ctx, userID, 24*30) // Last 30 days

	if activity != nil {
		questionsAnswered := activity["questions_answered"]
		if questionsAnswered > 0 {
			trend.DaysActive = questionsAnswered / 10 // Rough estimate
		}

		recentAccuracy := activity["recent_accuracy_percent"]
		trend.LastMonthAccuracy = float64(recentAccuracy) / 100.0
	}

	// Calculate trend
	// If last week was better than last month, trend is positive
	trend.AccuracyTrend = (trend.LastWeekAccuracy - trend.LastMonthAccuracy) * 100

	// Estimate projected level (1-15)
	if trend.LastMonthAccuracy < 0.5 {
		trend.ProjectedLevel = 5
	} else if trend.LastMonthAccuracy < 0.7 {
		trend.ProjectedLevel = 8
	} else if trend.LastMonthAccuracy < 0.85 {
		trend.ProjectedLevel = 11
	} else {
		trend.ProjectedLevel = 14
	}

	return trend
}

// analyzeStreak analyzes user's practice streak
func (e *AnalyticsEngine) analyzeStreak(ctx context.Context, userID uint) *StreakAnalysis {
	analysis := &StreakAnalysis{
		StreakStartDay: time.Now(),
	}

	// Get recent sessions to calculate streak
	results, _ := e.repo.GetResultsByUser(ctx, userID, 100, 0)

	if len(results) == 0 {
		return analysis
	}

	// Sort by timestamp descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	// Calculate current streak from most recent
	currentStreak := 1
	analysis.StreakStartDay = results[0].Timestamp

	for i := 0; i < len(results)-1; i++ {
		date1 := results[i].Timestamp.Format("2006-01-02")
		date2 := results[i+1].Timestamp.Format("2006-01-02")

		if date1 == date2 {
			// Same day, continue
			continue
		}

		// Check if consecutive days
		day1 := results[i].Timestamp.Day()
		day2 := results[i+1].Timestamp.Day()

		if day1-day2 == 1 {
			currentStreak++
		} else {
			break
		}
	}

	analysis.CurrentStreak = currentStreak
	analysis.LongestStreak = currentStreak // Simplified - would need more data

	return analysis
}

// GetWeakAreas identifies weak fact families needing practice
func (e *AnalyticsEngine) GetWeakAreas(ctx context.Context, userID uint, limit int) (map[string]int, error) {
	return e.repo.GetWeakFactFamilies(ctx, userID, limit)
}

// GetStrengthAreas identifies mastered fact families
func (e *AnalyticsEngine) GetStrengthAreas(ctx context.Context, userID uint) (map[string]float64, error) {
	masteries, err := e.repo.GetMasteryByUser(ctx, userID, MODE_MIXED)
	if err != nil {
		return nil, fmt.Errorf("failed to get masteries: %w", err)
	}

	areas := make(map[string]float64)
	for _, m := range masteries {
		if m.MasteryLevel >= 80 {
			areas[m.Fact] = float64(m.MasteryLevel)
		}
	}

	return areas, nil
}

// GenerateInsight generates a personalized learning insight
func (e *AnalyticsEngine) GenerateInsight(ctx context.Context, userID uint) (string, error) {
	analytics, err := e.GetUserAnalytics(ctx, userID)
	if err != nil {
		return "", err
	}

	// Generate insight based on analytics
	if analytics.OverallAccuracy < 0.6 {
		return "Your accuracy is below 60%. Focus on mastering the fundamentals before moving to harder problems.", nil
	}

	if analytics.OverallAccuracy >= 0.9 {
		return "Excellent work! You're mastering the material. Try more challenging problems.", nil
	}

	if analytics.TimeOfDayAnalysis != nil && analytics.TimeOfDayAnalysis.BestTimeOfDay != "" {
		return fmt.Sprintf("You perform best in the %s. Try to schedule practice sessions during this time.", analytics.TimeOfDayAnalysis.BestTimeOfDay), nil
	}

	if len(analytics.FactFamilyAnalysis.WeakAreas) > 0 {
		// Get first weak area
		for family := range analytics.FactFamilyAnalysis.WeakAreas {
			return fmt.Sprintf("You're struggling with %s. Would you like to focus practice on this area?", family), nil
		}
	}

	return "Keep practicing! Consistency is key to improvement.", nil
}

// Helper function to calculate average
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// GetMasteryByUser returns all mastery records for a user
func (e *AnalyticsEngine) GetMasteryByUser(ctx context.Context, userID uint) ([]*Mastery, error) {
	return e.repo.GetMasteryByUser(ctx, userID, MODE_MIXED)
}

// CompareUsers compares statistics between two users
func (e *AnalyticsEngine) CompareUsers(ctx context.Context, userID1, userID2 uint) (map[string]interface{}, error) {
	stats1, err := e.repo.GetUserStats(ctx, userID1)
	if err != nil {
		return nil, err
	}

	stats2, err := e.repo.GetUserStats(ctx, userID2)
	if err != nil {
		return nil, err
	}

	comparison := make(map[string]interface{})
	comparison["user1_accuracy"] = stats1.AverageAccuracy
	comparison["user2_accuracy"] = stats2.AverageAccuracy
	comparison["user1_sessions"] = stats1.TotalSessions
	comparison["user2_sessions"] = stats2.TotalSessions
	comparison["accuracy_difference"] = stats1.AverageAccuracy - stats2.AverageAccuracy

	return comparison, nil
}

// PredictMastery predicts when a fact will be mastered (in days)
func (e *AnalyticsEngine) PredictMastery(ctx context.Context, userID uint, factFamily string) int {
	// Get mistakes in this family
	mistakes, _ := e.repo.GetMistakesByFactFamily(ctx, userID, factFamily)

	if len(mistakes) == 0 {
		return 1 // Already mastered or no errors
	}

	// Estimate based on error count
	errorCount := 0
	for _, m := range mistakes {
		errorCount += m.ErrorCount
	}

	// Rough estimate: 2 days per error + base 3 days
	days := 3 + (errorCount * 2)
	if days > 30 {
		days = 30
	}

	return days
}
