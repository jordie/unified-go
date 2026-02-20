package unified

import (
	"math"
	"testing"
	"time"
)

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	service := NewService(nil)
	if service == nil {
		t.Fatal("NewService returned nil")
	}
}

// TestNormalizeTypingMetrics tests typing metrics normalization
func TestNormalizeTypingMetrics(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		name     string
		metrics  map[string]float64
		minScore float64
		maxScore float64
	}{
		{
			name: "beginner typing",
			metrics: map[string]float64{
				"wpm":      30,
				"accuracy": 90,
			},
			minScore: 30,
			maxScore: 50,
		},
		{
			name: "intermediate typing",
			metrics: map[string]float64{
				"wpm":      75,
				"accuracy": 96,
			},
			minScore: 60,
			maxScore: 80,
		},
		{
			name: "advanced typing",
			metrics: map[string]float64{
				"wpm":      120,
				"accuracy": 99.5,
			},
			minScore: 80,
			maxScore: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.normalizeTypingMetrics(tt.metrics)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Expected score between %f and %f, got %f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

// TestNormalizeMathMetrics tests math metrics normalization
func TestNormalizeMathMetrics(t *testing.T) {
	service := NewService(nil)

	metrics := map[string]float64{
		"accuracy":        85.0,
		"average_time":    10.0,
		"difficulty_score": 70.0,
	}

	score := service.normalizeMathMetrics(metrics)
	if score < 0 || score > 100 {
		t.Errorf("Expected score between 0 and 100, got %f", score)
	}
}

// TestNormalizeReadingMetrics tests reading metrics normalization
func TestNormalizeReadingMetrics(t *testing.T) {
	service := NewService(nil)

	metrics := map[string]float64{
		"wpm":              200,
		"comprehension":    88.0,
		"accuracy":         92.0,
		"difficulty_level": 3,
	}

	score := service.normalizeReadingMetrics(metrics)
	if score < 0 || score > 100 {
		t.Errorf("Expected score between 0 and 100, got %f", score)
	}
}

// TestNormalizePianoMetrics tests piano metrics normalization
func TestNormalizePianoMetrics(t *testing.T) {
	service := NewService(nil)

	metrics := map[string]float64{
		"accuracy":        90.0,
		"tempo_accuracy":  85.0,
		"score":           88.0,
		"difficulty_score": 75.0,
	}

	score := service.normalizePianoMetrics(metrics)
	if score < 0 || score > 100 {
		t.Errorf("Expected score between 0 and 100, got %f", score)
	}

	// Calculate expected: 90*0.4 + 85*0.3 + 88*0.2 + 75*0.1 = 36 + 25.5 + 17.6 + 7.5 = 86.6
	if math.Abs(score-86.6) > 0.1 {
		t.Errorf("Expected 86.6, got %f", score)
	}
}

// TestNormalizeSkillLevel tests unified normalization method
func TestNormalizeSkillLevel(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		app     string
		metrics map[string]float64
	}{
		{"typing", map[string]float64{"wpm": 75, "accuracy": 95}},
		{"math", map[string]float64{"accuracy": 85, "average_time": 10, "difficulty_score": 70}},
		{"reading", map[string]float64{"wpm": 200, "comprehension": 88, "accuracy": 92, "difficulty_level": 3}},
		{"piano", map[string]float64{"accuracy": 90, "tempo_accuracy": 85, "score": 88, "difficulty_score": 75}},
	}

	for _, tt := range tests {
		score := service.NormalizeSkillLevel(tt.app, tt.metrics)
		if score < 0 || score > 100 {
			t.Errorf("%s: Expected score between 0 and 100, got %f", tt.app, score)
		}
	}
}

// TestCalculateOverallLevel tests overall level calculation
func TestCalculateOverallLevel(t *testing.T) {
	service := NewService(nil)

	profile := &UnifiedUserProfile{
		TypingLevel:  80.0,
		MathLevel:    75.0,
		ReadingLevel: 85.0,
		PianoLevel:   70.0,
	}

	overall := service.CalculateOverallLevel(profile)
	expected := 77.5
	if math.Abs(overall-expected) > 0.1 {
		t.Errorf("Expected overall %f, got %f", expected, overall)
	}
}

// TestMapSkillLevelToString tests skill level mapping
func TestMapSkillLevelToString(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		score    float64
		expected string
	}{
		{95, "expert"},
		{80, "advanced"},
		{60, "intermediate"},
		{35, "beginner"},
		{10, "novice"},
	}

	for _, tt := range tests {
		level := service.MapSkillLevelToString(tt.score)
		if level != tt.expected {
			t.Errorf("Score %f: Expected %s, got %s", tt.score, tt.expected, level)
		}
	}
}

// TestDetectTrends tests trend detection
func TestDetectTrends(t *testing.T) {
	service := NewService(nil)

	now := time.Now()
	sessions := []UnifiedSession{
		{
			App:               "typing",
			PerformanceScore:  70.0,
			StartTime:         now.AddDate(0, 0, -4),
		},
		{
			App:               "typing",
			PerformanceScore:  75.0,
			StartTime:         now.AddDate(0, 0, -3),
		},
		{
			App:               "typing",
			PerformanceScore:  80.0,
			StartTime:         now.AddDate(0, 0, -2),
		},
		{
			App:               "math",
			PerformanceScore:  60.0,
			StartTime:         now.AddDate(0, 0, -3),
		},
		{
			App:               "math",
			PerformanceScore:  55.0,
			StartTime:         now.AddDate(0, 0, -1),
		},
	}

	trends := service.DetectTrends(sessions)

	if typingTrend, ok := trends["typing"]; !ok || typingTrend <= 0 {
		t.Errorf("Expected positive typing trend, got %v", typingTrend)
	}

	if mathTrend, ok := trends["math"]; !ok || mathTrend >= 0 {
		t.Errorf("Expected negative math trend, got %v", mathTrend)
	}
}

// TestCalculateConsistencyScore tests consistency calculation
func TestCalculateConsistencyScore(t *testing.T) {
	service := NewService(nil)

	now := time.Now()

	// Good consistency: sessions every day
	goodSessions := []UnifiedSession{
		{StartTime: now.AddDate(0, 0, -3)},
		{StartTime: now.AddDate(0, 0, -2)},
		{StartTime: now.AddDate(0, 0, -1)},
		{StartTime: now},
	}

	score := service.CalculateConsistencyScore(goodSessions)
	if score < 70 {
		t.Errorf("Good consistency should score high, got %f", score)
	}

	// Poor consistency: sessions days apart
	poorSessions := []UnifiedSession{
		{StartTime: now.AddDate(0, 0, -10)},
		{StartTime: now.AddDate(0, 0, -5)},
		{StartTime: now},
	}

	poorScore := service.CalculateConsistencyScore(poorSessions)
	if poorScore > 50 {
		t.Errorf("Poor consistency should score low, got %f", poorScore)
	}
}

// TestCalculateDailyStreak tests streak calculation
func TestCalculateDailyStreak(t *testing.T) {
	service := NewService(nil)

	now := time.Now()

	// 5-day streak
	sessions := []UnifiedSession{
		{StartTime: now.AddDate(0, 0, -4).Add(2 * time.Hour)},
		{StartTime: now.AddDate(0, 0, -3).Add(2 * time.Hour)},
		{StartTime: now.AddDate(0, 0, -2).Add(2 * time.Hour)},
		{StartTime: now.AddDate(0, 0, -1).Add(2 * time.Hour)},
		{StartTime: now.Add(2 * time.Hour)},
	}

	streak := service.CalculateDailyStreak(sessions)
	if streak != 5 {
		t.Errorf("Expected 5-day streak, got %d", streak)
	}

	// Broken streak
	brokenSessions := []UnifiedSession{
		{StartTime: now.AddDate(0, 0, -4).Add(2 * time.Hour)},
		{StartTime: now.AddDate(0, 0, -3).Add(2 * time.Hour)},
		{StartTime: now.AddDate(0, 0, -1).Add(2 * time.Hour)}, // Gap on day -2
		{StartTime: now.Add(2 * time.Hour)},
	}

	brokenStreak := service.CalculateDailyStreak(brokenSessions)
	if brokenStreak != 2 {
		t.Errorf("Expected 2-day streak (after break), got %d", brokenStreak)
	}
}

// TestGenerateRecommendations tests recommendation generation
func TestGenerateRecommendations(t *testing.T) {
	service := NewService(nil)

	analytics := &CrossAppAnalytics{
		UserID:            1,
		WeakestApp:        "piano",
		MostPracticedApp:  "typing",
		TotalHoursPracticed: 15.0,
	}

	profile := &UnifiedUserProfile{
		UserID:         1,
		TypingLevel:    85.0,
		MathLevel:      70.0,
		ReadingLevel:   80.0,
		PianoLevel:     50.0,
		OverallLevel:   71.25,
	}

	recommendations := service.GenerateRecommendations(analytics, profile)

	if recommendations == nil {
		t.Fatal("Recommendations is nil")
	}

	if recommendations.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", recommendations.UserID)
	}

	if recommendations.DifficultyAdvice == "" {
		t.Error("DifficultyAdvice is empty")
	}

	if recommendations.PracticeTimeAdvice == "" {
		t.Error("PracticeTimeAdvice is empty")
	}

	if len(recommendations.AppRecommendations) == 0 {
		t.Error("No app recommendations generated")
	}
}

// TestEstimateMasteryTime tests mastery time estimation
func TestEstimateMasteryTime(t *testing.T) {
	service := NewService(nil)

	// From level 50 to level 90 with 5% improvement per week
	// With 45 min/day, that's 5.25 hours/week
	// (90-50) / 5 * 7 / (0.75*24) ≈ 3.1 days - the formula is fast because 5% per week is aggressive
	days := service.EstimateMasteryTime(50.0, 90.0, 5.0)
	if days < 1 || days > 10 {
		t.Errorf("Expected mastery time between 1 and 10 days, got %f", days)
	}

	// Already at target level
	days = service.EstimateMasteryTime(90.0, 90.0, 5.0)
	if days != 0 {
		t.Errorf("Expected 0 days for equal level, got %f", days)
	}
}

// TestCompareWithPeers tests percentile calculation
func TestCompareWithPeers(t *testing.T) {
	service := NewService(nil)

	peerLevels := []float64{50, 60, 70, 80, 90}

	// User with score 75 should be around 60th percentile
	percentile := service.CompareWithPeers(75.0, peerLevels)
	if percentile < 40 || percentile > 80 {
		t.Errorf("Expected percentile around 60, got %f", percentile)
	}

	// User with highest score
	topPercentile := service.CompareWithPeers(95.0, peerLevels)
	if topPercentile < 80 {
		t.Errorf("Expected high percentile, got %f", topPercentile)
	}
}

// TestIdentifyMilestones tests milestone identification
func TestIdentifyMilestones(t *testing.T) {
	service := NewService(nil)

	profile := &UnifiedUserProfile{
		UserID:                1,
		OverallLevel:          26, // Just passed beginner
		TotalSessionsAll:      10,
		TotalPracticeMinutes:  65, // Just passed 1 hour
	}

	milestones := service.IdentifyMilestones(profile)

	if len(milestones) == 0 {
		t.Error("No milestones identified")
	}

	// Should have beginner completion milestone
	found := false
	for _, m := range milestones {
		if m == "Beginner level completed - Ready for intermediate!" {
			found = true
		}
	}
	if !found {
		t.Error("Expected beginner completion milestone")
	}
}

// TestGetAppInsights tests app-specific insights
func TestGetAppInsights(t *testing.T) {
	service := NewService(nil)

	// Improving metrics
	improver := &AppMetricsSummary{
		App:              "typing",
		SessionCount:     5,
		ImprovementTrend: 7.5,
	}

	insight := service.GetAppInsights(improver)
	if insight == "" {
		t.Error("Insight is empty for improving app")
	}

	// Declining metrics
	decliner := &AppMetricsSummary{
		App:              "math",
		SessionCount:     3,
		ImprovementTrend: -8.0,
	}

	declineInsight := service.GetAppInsights(decliner)
	if declineInsight == "" {
		t.Error("Insight is empty for declining app")
	}

	// No sessions
	newApp := &AppMetricsSummary{
		App:          "reading",
		SessionCount: 0,
	}

	newInsight := service.GetAppInsights(newApp)
	if newInsight == "" {
		t.Error("Insight is empty for new app")
	}
}

// TestNormalizeAllApps tests all normalization paths
func TestNormalizeAllApps(t *testing.T) {
	service := NewService(nil)

	apps := []string{"typing", "math", "reading", "piano", "unknown"}

	for _, app := range apps {
		metrics := make(map[string]float64)
		// Add some default metrics
		metrics["accuracy"] = 80.0
		metrics["wpm"] = 75.0

		score := service.NormalizeSkillLevel(app, metrics)
		if app == "unknown" {
			if score != 0 {
				t.Errorf("Unknown app should return 0, got %f", score)
			}
		} else if score < 0 || score > 100 {
			t.Errorf("%s: Expected 0-100, got %f", app, score)
		}
	}
}
