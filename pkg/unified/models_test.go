package unified

import (
	"testing"
	"time"
)

// TestUnifiedUserProfileCreation tests creating a user profile
func TestUnifiedUserProfileCreation(t *testing.T) {
	profile := &UnifiedUserProfile{
		UserID:               1,
		Username:             "testuser",
		TotalSessionsAll:     10,
		TotalPracticeMinutes: 120.5,
		TypingLevel:          85.0,
		MathLevel:            90.0,
		ReadingLevel:         75.0,
		PianoLevel:           80.0,
		DailyStreakDays:      5,
	}

	if profile.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", profile.UserID)
	}

	if profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", profile.Username)
	}

	if profile.TotalSessionsAll != 10 {
		t.Errorf("Expected 10 sessions, got %d", profile.TotalSessionsAll)
	}

	expectedOverall := (85.0 + 90.0 + 75.0 + 80.0) / 4
	// Note: OverallLevel would be calculated separately
	_ = expectedOverall
}

// TestUnifiedSession tests creating a unified session
func TestUnifiedSession(t *testing.T) {
	now := time.Now()
	session := &UnifiedSession{
		ID:                 1,
		UserID:             1,
		App:                "typing",
		StartTime:          now,
		EndTime:            now.Add(5 * time.Minute),
		Duration:           5.0,
		PerformanceScore:   85.0,
		AccuracyScore:      90.0,
		SpeedScore:         80.0,
		ActivityLabel:      "WPM: 75",
		OriginalData:       make(map[string]interface{}),
	}

	if session.App != "typing" {
		t.Errorf("Expected app 'typing', got '%s'", session.App)
	}

	if session.PerformanceScore != 85.0 {
		t.Errorf("Expected performance score 85.0, got %f", session.PerformanceScore)
	}
}

// TestCrossAppAnalytics tests creating analytics data
func TestCrossAppAnalytics(t *testing.T) {
	analytics := &CrossAppAnalytics{
		UserID:               1,
		TotalHoursPracticed:  10.5,
		AvgSessionLength:     30.0,
		StrongestApp:         "typing",
		WeakestApp:           "piano",
		MostPracticedApp:     "math",
		TotalAppsPracticed:   4,
		WeeklyProgress:       make(map[string]float64),
		MonthlyProgress:      make(map[string]float64),
		AppMetrics:           make(map[string]*AppMetricsSummary),
		RecommendedApp:       "piano",
		RecommendedFocus:     "tempo accuracy",
	}

	if analytics.TotalAppsPracticed != 4 {
		t.Errorf("Expected 4 apps practiced, got %d", analytics.TotalAppsPracticed)
	}

	if analytics.RecommendedApp != "piano" {
		t.Errorf("Expected recommended app 'piano', got '%s'", analytics.RecommendedApp)
	}
}

// TestAppMetricsSummary tests app metrics
func TestAppMetricsSummary(t *testing.T) {
	now := time.Now()
	metrics := &AppMetricsSummary{
		App:                "typing",
		SessionCount:       15,
		AveragePerformance: 82.5,
		BestPerformance:    95.0,
		LastSessionTime:    now,
		TotalTimeSpent:     120.0,
		ConsistencyScore:   85.0,
		ImprovementTrend:   5.5,
	}

	if metrics.App != "typing" {
		t.Errorf("Expected app 'typing', got '%s'", metrics.App)
	}

	if metrics.SessionCount != 15 {
		t.Errorf("Expected 15 sessions, got %d", metrics.SessionCount)
	}

	if metrics.ImprovementTrend != 5.5 {
		t.Errorf("Expected 5.5 improvement trend, got %f", metrics.ImprovementTrend)
	}
}

// TestUnifiedLeaderboard tests leaderboard creation
func TestUnifiedLeaderboard(t *testing.T) {
	entries := []LeaderboardEntry{
		{
			Rank:        1,
			UserID:      1,
			Username:    "alice",
			App:         "typing",
			MetricValue: 120.0,
			MetricLabel: "120 WPM",
		},
		{
			Rank:        2,
			UserID:      2,
			Username:    "bob",
			App:         "typing",
			MetricValue: 105.0,
			MetricLabel: "105 WPM",
		},
	}

	lb := &UnifiedLeaderboard{
		Category:  "typing_wpm",
		Entries:   entries,
		UpdatedAt: time.Now(),
	}

	if len(lb.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(lb.Entries))
	}

	if lb.Entries[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", lb.Entries[0].Rank)
	}

	if lb.Entries[0].MetricValue != 120.0 {
		t.Errorf("Expected metric 120.0, got %f", lb.Entries[0].MetricValue)
	}
}

// TestSystemStats tests system statistics
func TestSystemStats(t *testing.T) {
	stats := &SystemStats{
		TotalUsers:          100,
		ActiveUsersToday:    25,
		ActiveUsersThisWeek: 60,
		ActiveUsersThisMonth: 85,
		AppUsageCount:       make(map[string]int),
		AppAverageScore:     make(map[string]float64),
	}

	stats.AppUsageCount["typing"] = 500
	stats.AppUsageCount["math"] = 450
	stats.AppUsageCount["reading"] = 300
	stats.AppUsageCount["piano"] = 250

	stats.AppAverageScore["typing"] = 82.5
	stats.AppAverageScore["math"] = 78.0
	stats.AppAverageScore["reading"] = 85.0
	stats.AppAverageScore["piano"] = 80.0

	if stats.TotalUsers != 100 {
		t.Errorf("Expected 100 users, got %d", stats.TotalUsers)
	}

	if stats.AppUsageCount["typing"] != 500 {
		t.Errorf("Expected 500 typing sessions, got %d", stats.AppUsageCount["typing"])
	}
}

// TestSkillLevel tests skill level representation
func TestSkillLevel(t *testing.T) {
	skill := &SkillLevel{
		App:             "piano",
		NormalizedScore: 82.5,
		Level:           "intermediate",
		SessionCount:    25,
		MasteredItems:   5,
		CurrentFocus:    "tempo accuracy",
	}

	if skill.App != "piano" {
		t.Errorf("Expected app 'piano', got '%s'", skill.App)
	}

	if skill.Level != "intermediate" {
		t.Errorf("Expected level 'intermediate', got '%s'", skill.Level)
	}

	if skill.NormalizedScore != 82.5 {
		t.Errorf("Expected score 82.5, got %f", skill.NormalizedScore)
	}
}

// TestRecommendationData tests recommendation structures
func TestRecommendationData(t *testing.T) {
	recommendation := &RecommendationData{
		UserID:             1,
		GeneratedAt:        time.Now(),
		AppRecommendations: make([]AppRecommendation, 0),
		DifficultyAdvice:   "Try intermediate level problems",
		PracticeTimeAdvice: "Practice 30 minutes daily",
		FocusAreas:         make([]FocusArea, 0),
		SuggestedGoals:     make([]Goal, 0),
		MilestoneDistance:  make(map[string]int),
	}

	if recommendation.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", recommendation.UserID)
	}

	if recommendation.DifficultyAdvice != "Try intermediate level problems" {
		t.Errorf("Unexpected advice")
	}
}

// TestGoal tests goal structure
func TestGoal(t *testing.T) {
	now := time.Now()
	deadline := now.AddDate(0, 0, 30)
	goal := &Goal{
		ID:           1,
		UserID:       1,
		App:          "typing",
		Description:  "Reach 100 WPM",
		TargetValue:  100.0,
		CurrentValue: 75.0,
		Deadline:     &deadline,
		Progress:     75.0,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if goal.Description != "Reach 100 WPM" {
		t.Errorf("Unexpected goal description")
	}

	if goal.Progress != 75.0 {
		t.Errorf("Expected 75 percent progress, got %f", goal.Progress)
	}

	if goal.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", goal.Status)
	}
}

// TestUserDailyActivity tests daily activity tracking
func TestUserDailyActivity(t *testing.T) {
	today := time.Now()
	activity := &UserDailyActivity{
		UserID:       1,
		Date:         today,
		AppsUsed:     []string{"typing", "math"},
		SessionCount: 5,
		TotalMinutes: 45.0,
		AverageScore: 82.5,
		SessionDetails: []DailySessionDetail{
			{
				App:              "typing",
				SessionCount:     3,
				Duration:         30.0,
				PerformanceScore: 85.0,
				TimeOfDay:        "morning",
			},
		},
	}

	if len(activity.AppsUsed) != 2 {
		t.Errorf("Expected 2 apps used, got %d", len(activity.AppsUsed))
	}

	if activity.TotalMinutes != 45.0 {
		t.Errorf("Expected 45 minutes, got %f", activity.TotalMinutes)
	}
}
