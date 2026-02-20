package unified

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Service provides business logic for cross-app data aggregation
type Service struct {
	repo *Repository
}

// NewService creates a new unified service
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// NormalizeSkillLevel converts app-specific metrics to a 0-100 scale
func (s *Service) NormalizeSkillLevel(app string, rawMetrics map[string]float64) float64 {
	switch app {
	case "typing":
		return s.normalizeTypingMetrics(rawMetrics)
	case "math":
		return s.normalizeMathMetrics(rawMetrics)
	case "reading":
		return s.normalizeReadingMetrics(rawMetrics)
	case "piano":
		return s.normalizePianoMetrics(rawMetrics)
	default:
		return 0
	}
}

// normalizeTypingMetrics converts typing WPM and accuracy to 0-100 scale
// WPM scale: 0-50 WPM = 0-50 points, 50-100 WPM = 50-75 points, 100+ WPM = 75-100 points
// Accuracy scale: 0-95% = 0-75 points, 95-99% = 75-95 points, 99-100% = 95-100 points
func (s *Service) normalizeTypingMetrics(metrics map[string]float64) float64 {
	wpm := metrics["wpm"]
	accuracy := metrics["accuracy"]

	// Calculate WPM score (0-100)
	var wpmScore float64
	if wpm <= 50 {
		wpmScore = (wpm / 50) * 50
	} else if wpm <= 100 {
		wpmScore = 50 + ((wpm - 50) / 50) * 25
	} else {
		wpmScore = 75 + math.Min((wpm-100)/50, 1) * 25
	}

	// Calculate accuracy score (0-100)
	var accuracyScore float64
	if accuracy <= 95 {
		accuracyScore = (accuracy / 95) * 75
	} else if accuracy <= 99 {
		accuracyScore = 75 + ((accuracy - 95) / 4) * 20
	} else {
		accuracyScore = 95 + ((accuracy - 99) / 1) * 5
	}

	// Weighted average: 60% speed, 40% accuracy
	normalizedScore := (wpmScore * 0.6) + (accuracyScore * 0.4)
	return math.Min(normalizedScore, 100)
}

// normalizeMathMetrics converts math accuracy and speed to 0-100 scale
// Math is primarily accuracy-based with time as secondary factor
func (s *Service) normalizeMathMetrics(metrics map[string]float64) float64 {
	accuracy := metrics["accuracy"]        // 0-100
	avgTime := metrics["average_time"]     // seconds per problem
	difficulty := metrics["difficulty_score"] // 0-100

	// Accuracy is primary (70%)
	accuracyScore := math.Min(accuracy, 100)

	// Speed score: faster is better, scale by difficulty
	// Average time of 5 seconds = 100 score, 30 seconds = 0 score
	var speedScore float64
	if avgTime <= 5 {
		speedScore = 100
	} else if avgTime >= 30 {
		speedScore = 0
	} else {
		speedScore = 100 - ((avgTime - 5) / 25) * 100
	}

	// Difficulty adjustment (secondary, 10%)
	difficultyBoost := (difficulty / 100) * 10

	// Weighted calculation
	score := (accuracyScore * 0.7) + (speedScore * 0.2) + difficultyBoost
	return math.Min(score, 100)
}

// normalizeReadingMetrics converts reading WPM, comprehension, and accuracy to 0-100 scale
func (s *Service) normalizeReadingMetrics(metrics map[string]float64) float64 {
	wpm := metrics["wpm"]                          // words per minute
	comprehension := metrics["comprehension"]       // 0-100
	accuracy := metrics["accuracy"]                 // 0-100
	level := metrics["difficulty_level"]            // 1-5 scale

	// WPM scale: 100 WPM = 50 pts, 300 WPM = 100 pts
	var wpmScore float64
	if wpm <= 100 {
		wpmScore = (wpm / 100) * 50
	} else {
		wpmScore = 50 + math.Min((wpm-100)/200, 1) * 50
	}

	// Comprehension (40% of score)
	comprehensionScore := math.Min(comprehension, 100)

	// Accuracy (20% of score)
	accuracyScore := math.Min(accuracy, 100)

	// Difficulty bonus (10%)
	difficultyBonus := (level / 5) * 10

	// Weighted calculation: 30% WPM, 40% comprehension, 20% accuracy, 10% difficulty
	score := (wpmScore * 0.3) + (comprehensionScore * 0.4) + (accuracyScore * 0.2) + difficultyBonus
	return math.Min(score, 100)
}

// normalizePianoMetrics converts piano accuracy, tempo, and score to 0-100 scale
func (s *Service) normalizePianoMetrics(metrics map[string]float64) float64 {
	accuracy := metrics["accuracy"]              // 0-100
	tempoAccuracy := metrics["tempo_accuracy"]   // 0-100
	compositeScore := metrics["score"]           // 0-100
	difficulty := metrics["difficulty_score"]    // 0-100

	// All are already 0-100 scales
	// Weighted: 40% accuracy, 30% tempo, 20% score, 10% difficulty
	score := (accuracy * 0.4) + (tempoAccuracy * 0.3) + (compositeScore * 0.2) + (difficulty * 0.1)
	return math.Min(score, 100)
}

// CalculateOverallLevel calculates weighted average of all app levels
func (s *Service) CalculateOverallLevel(profile *UnifiedUserProfile) float64 {
	appCount := 0
	totalScore := 0.0

	if profile.TypingLevel > 0 {
		totalScore += profile.TypingLevel
		appCount++
	}
	if profile.MathLevel > 0 {
		totalScore += profile.MathLevel
		appCount++
	}
	if profile.ReadingLevel > 0 {
		totalScore += profile.ReadingLevel
		appCount++
	}
	if profile.PianoLevel > 0 {
		totalScore += profile.PianoLevel
		appCount++
	}

	if appCount == 0 {
		return 0
	}

	return totalScore / float64(appCount)
}

// MapSkillLevelToString converts normalized score to skill level text
func (s *Service) MapSkillLevelToString(score float64) string {
	switch {
	case score >= 90:
		return "expert"
	case score >= 75:
		return "advanced"
	case score >= 50:
		return "intermediate"
	case score >= 25:
		return "beginner"
	default:
		return "novice"
	}
}

// DetectTrends analyzes performance trends from recent sessions
func (s *Service) DetectTrends(sessions []UnifiedSession) map[string]float64 {
	trends := make(map[string]float64)

	if len(sessions) < 2 {
		return trends
	}

	// Group sessions by app
	appSessions := make(map[string][]UnifiedSession)
	for _, session := range sessions {
		appSessions[session.App] = append(appSessions[session.App], session)
	}

	// Calculate trend for each app
	for app, appSessionList := range appSessions {
		if len(appSessionList) < 2 {
			trends[app] = 0
			continue
		}

		// Calculate average score for first half and second half
		midpoint := len(appSessionList) / 2
		firstHalfScore := 0.0
		secondHalfScore := 0.0

		for i := 0; i < midpoint; i++ {
			firstHalfScore += appSessionList[i].PerformanceScore
		}
		firstHalfScore /= float64(midpoint)

		for i := midpoint; i < len(appSessionList); i++ {
			secondHalfScore += appSessionList[i].PerformanceScore
		}
		secondHalfScore /= float64(len(appSessionList) - midpoint)

		// Calculate percentage change
		if firstHalfScore == 0 {
			trends[app] = 0
		} else {
			trends[app] = ((secondHalfScore - firstHalfScore) / firstHalfScore) * 100
		}
	}

	return trends
}

// CalculateConsistencyScore measures practice consistency (0-100)
// Based on session frequency and regularity
func (s *Service) CalculateConsistencyScore(sessions []UnifiedSession) float64 {
	if len(sessions) < 3 {
		return float64(len(sessions)) * 20 // Partial credit for any sessions
	}

	// Sort sessions by time
	sortedSessions := make([]UnifiedSession, len(sessions))
	copy(sortedSessions, sessions)
	sort.Slice(sortedSessions, func(i, j int) bool {
		return sortedSessions[i].StartTime.Before(sortedSessions[j].StartTime)
	})

	// Calculate average time between sessions
	totalGap := time.Duration(0)
	for i := 1; i < len(sortedSessions); i++ {
		gap := sortedSessions[i].StartTime.Sub(sortedSessions[i-1].StartTime)
		totalGap += gap
	}

	avgGap := totalGap / time.Duration(len(sortedSessions)-1)

	// Ideal gap is 1 day (24 hours)
	// Score based on how close to ideal gap
	idealGap := 24 * time.Hour
	gapDiff := math.Abs(float64(avgGap - idealGap))
	maxDiff := float64(7 * 24 * time.Hour) // 7 days difference = 0 score

	if gapDiff >= maxDiff {
		return 0
	}

	return 100 * (1 - (gapDiff / maxDiff))
}

// GenerateRecommendations creates actionable recommendations for a user
func (s *Service) GenerateRecommendations(analytics *CrossAppAnalytics, profile *UnifiedUserProfile) *RecommendationData {
	recommendations := &RecommendationData{
		UserID:              profile.UserID,
		GeneratedAt:         time.Now(),
		AppRecommendations:  make([]AppRecommendation, 0),
		DifficultyAdvice:    "",
		PracticeTimeAdvice:  "",
		FocusAreas:          make([]FocusArea, 0),
		SuggestedGoals:      make([]Goal, 0),
		MilestoneDistance:   make(map[string]int),
	}

	// Analyze weakest app
	if analytics.WeakestApp != "" && profile.OverallLevel < 50 {
		recommendations.AppRecommendations = append(recommendations.AppRecommendations, AppRecommendation{
			App:             analytics.WeakestApp,
			Reason:          "This is your weakest area - practice will improve faster here",
			Priority:        "high",
			SuggestedAction: fmt.Sprintf("Spend 15-20 minutes daily on %s", analytics.WeakestApp),
			ExpectedBenefit: "Expected 10-15% improvement in 2 weeks",
		})
	}

	// Recommend app with lowest usage
	if analytics.MostPracticedApp != "" {
		recommendations.AppRecommendations = append(recommendations.AppRecommendations, AppRecommendation{
			App:             analytics.MostPracticedApp,
			Reason:          "You've made great progress here - maintain momentum",
			Priority:        "medium",
			SuggestedAction: fmt.Sprintf("Continue with %s while exploring others", analytics.MostPracticedApp),
			ExpectedBenefit: "Consolidate skills and reach mastery",
		})
	}

	// Generate difficulty advice
	if profile.OverallLevel < 40 {
		recommendations.DifficultyAdvice = "Focus on beginner level content to build foundations"
	} else if profile.OverallLevel < 65 {
		recommendations.DifficultyAdvice = "You're ready for intermediate challenges - try harder problems"
	} else if profile.OverallLevel < 85 {
		recommendations.DifficultyAdvice = "Advanced content is appropriate - push yourself further"
	} else {
		recommendations.DifficultyAdvice = "Expert level - challenge yourself with the hardest content"
	}

	// Practice time advice
	if analytics.TotalHoursPracticed < 5 {
		recommendations.PracticeTimeAdvice = "You've just started - consistency is key. Aim for 30 min daily"
	} else if analytics.TotalHoursPracticed < 20 {
		recommendations.PracticeTimeAdvice = "Good start! Increase to 45-60 minutes daily for faster progress"
	} else if analytics.TotalHoursPracticed < 50 {
		recommendations.PracticeTimeAdvice = "You're building solid skills - maintain current pace or increase"
	} else {
		recommendations.PracticeTimeAdvice = "Excellent commitment! Consider focusing on depth in your strongest area"
	}

	// Add focus areas (apps below overall level)
	if profile.TypingLevel > 0 && profile.TypingLevel < profile.OverallLevel-10 {
		recommendations.FocusAreas = append(recommendations.FocusAreas, FocusArea{
			App:             "typing",
			Area:            "Speed and Accuracy",
			CurrentLevel:    profile.TypingLevel,
			TargetLevel:     profile.OverallLevel,
			EstimatedSessions: int((profile.OverallLevel - profile.TypingLevel) / 5),
			Priority:        "medium",
		})
	}

	if profile.MathLevel > 0 && profile.MathLevel < profile.OverallLevel-10 {
		recommendations.FocusAreas = append(recommendations.FocusAreas, FocusArea{
			App:             "math",
			Area:            "Problem Solving",
			CurrentLevel:    profile.MathLevel,
			TargetLevel:     profile.OverallLevel,
			EstimatedSessions: int((profile.OverallLevel - profile.MathLevel) / 5),
			Priority:        "high",
		})
	}

	return recommendations
}

// CalculateDailyStreak determines how many consecutive days a user has practiced
func (s *Service) CalculateDailyStreak(sessions []UnifiedSession) int {
	if len(sessions) == 0 {
		return 0
	}

	// Sort sessions by date (most recent first)
	sortedSessions := make([]UnifiedSession, len(sessions))
	copy(sortedSessions, sessions)
	sort.Slice(sortedSessions, func(i, j int) bool {
		return sortedSessions[i].StartTime.After(sortedSessions[j].StartTime)
	})

	streak := 1
	previousDate := sortedSessions[0].StartTime.Truncate(24 * time.Hour)

	for i := 1; i < len(sortedSessions); i++ {
		currentDate := sortedSessions[i].StartTime.Truncate(24 * time.Hour)
		expectedDate := previousDate.AddDate(0, 0, -1)

		if currentDate.Equal(expectedDate) {
			streak++
			previousDate = currentDate
		} else if currentDate.Before(expectedDate) {
			// Gap detected, streak is broken
			break
		}
	}

	return streak
}

// CalculateSessionMetrics computes various metrics for a session
func (s *Service) CalculateSessionMetrics(session *UnifiedSession) {
	// This is already populated from the repository
	// This method can be used for additional calculations if needed
}

// GetAppInsights returns detailed insights for a specific app
func (s *Service) GetAppInsights(metrics *AppMetricsSummary) string {
	if metrics.SessionCount == 0 {
		return fmt.Sprintf("No sessions yet in %s. Start practicing to get insights!", metrics.App)
	}

	if metrics.ImprovementTrend > 5 {
		return fmt.Sprintf("%s: Great progress! You're improving at %.1f%% per week",
			metrics.App, metrics.ImprovementTrend)
	} else if metrics.ImprovementTrend < -5 {
		return fmt.Sprintf("%s: Performance has declined slightly (%.1f%%). Practice more to improve!",
			metrics.App, metrics.ImprovementTrend)
	}

	return fmt.Sprintf("%s: Steady progress. Keep practicing consistently!", metrics.App)
}

// EstimateMasteryTime calculates estimated time to reach a target level
func (s *Service) EstimateMasteryTime(currentLevel float64, targetLevel float64, improvementRate float64) float64 {
	if improvementRate <= 0 || currentLevel >= targetLevel {
		return 0
	}

	levelGap := targetLevel - currentLevel
	hoursPerDay := 0.75 // Average 45 minutes per day

	// Assuming linear improvement initially
	daysNeeded := levelGap / improvementRate * 7 / (hoursPerDay * 24)
	return daysNeeded
}

// CompareWithPeers provides percentile ranking compared to other users
func (s *Service) CompareWithPeers(userLevel float64, peerLevels []float64) float64 {
	if len(peerLevels) == 0 {
		return 50.0 // Default percentile if no peers
	}

	higherCount := 0
	for _, level := range peerLevels {
		if userLevel > level {
			higherCount++
		}
	}

	percentile := (float64(higherCount) / float64(len(peerLevels))) * 100
	return percentile
}

// IdentifyMilestones suggests upcoming milestones based on progress
func (s *Service) IdentifyMilestones(profile *UnifiedUserProfile) []string {
	milestones := make([]string, 0)

	// Level milestones
	if profile.OverallLevel >= 25 && profile.OverallLevel < 30 {
		milestones = append(milestones, "Beginner level completed - Ready for intermediate!")
	}
	if profile.OverallLevel >= 50 && profile.OverallLevel < 55 {
		milestones = append(milestones, "Intermediate level completed - Advanced challenges await!")
	}
	if profile.OverallLevel >= 75 && profile.OverallLevel < 80 {
		milestones = append(milestones, "Advanced level completed - You're almost an expert!")
	}
	if profile.OverallLevel >= 90 {
		milestones = append(milestones, "Expert level achieved - You've mastered this!")
	}

	// Activity milestones
	if profile.TotalSessionsAll == 10 {
		milestones = append(milestones, "10 sessions completed - Great consistency!")
	}
	if profile.TotalSessionsAll == 50 {
		milestones = append(milestones, "50 sessions completed - You're dedicated!")
	}
	if profile.TotalSessionsAll == 100 {
		milestones = append(milestones, "100 sessions completed - Incredible commitment!")
	}

	// Time milestones
	if profile.TotalPracticeMinutes >= 60 && profile.TotalPracticeMinutes < 70 {
		milestones = append(milestones, "1 hour of practice - You're building momentum!")
	}
	if profile.TotalPracticeMinutes >= 300 && profile.TotalPracticeMinutes < 310 {
		milestones = append(milestones, "5 hours of practice - Significant progress!")
	}
	if profile.TotalPracticeMinutes >= 600 && profile.TotalPracticeMinutes < 610 {
		milestones = append(milestones, "10 hours of practice - You're an achiever!")
	}

	return milestones
}
