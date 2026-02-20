package unified

import (
	"time"
)

// UnifiedUserProfile represents aggregated user data across all apps
type UnifiedUserProfile struct {
	UserID              uint
	Username            string

	// Aggregate metrics
	TotalSessionsAll    int
	TotalPracticeMinutes float64
	LastActivityDate    time.Time
	AccountCreated      time.Time

	// Per-app stats (pointers to app-specific types)
	TypingStats         interface{} // typing.UserStats
	MathStats           interface{} // math stats structure
	ReadingStats        interface{} // reading.ReadingStats
	PianoStats          interface{} // piano.UserProgress

	// Normalized skill levels (0-100 scale)
	TypingLevel         float64
	MathLevel           float64
	ReadingLevel        float64
	PianoLevel          float64
	OverallLevel        float64  // Weighted average

	// Engagement
	DailyStreakDays     int
	WeeklyActiveApps    []string
}

// UnifiedSession represents a normalized session across apps
type UnifiedSession struct {
	ID          uint
	UserID      uint
	App         string  // "typing", "math", "reading", "piano"
	StartTime   time.Time
	EndTime     time.Time
	Duration    float64

	// Normalized metrics (0-100 scale)
	PerformanceScore float64
	AccuracyScore    float64
	SpeedScore       float64
	ComprehensionScore float64

	// App-specific labels
	ActivityLabel   string  // e.g., "WPM: 75", "Accuracy: 92%", "Score: 85/100"

	// Original metrics (app-specific)
	OriginalData    map[string]interface{}
}

// CrossAppAnalytics represents advanced insights across apps
type CrossAppAnalytics struct {
	UserID              uint

	// Time patterns
	BestPracticeTimeOfDay string   // morning, afternoon, evening
	TotalHoursPracticed   float64
	AvgSessionLength      float64

	// Performance distribution
	StrongestApp        string
	WeakestApp          string
	MostPracticedApp    string
	TotalAppsPracticed  int

	// Trends (7-day and 30-day)
	WeeklyProgress      map[string]float64  // App -> % improvement
	MonthlyProgress     map[string]float64

	// App-specific metrics
	AppMetrics          map[string]*AppMetricsSummary

	// Recommendations
	RecommendedApp      string
	RecommendedFocus    string
	NextMilestone       string
	SuggestedActions    []string
}

// AppMetricsSummary contains key metrics for a single app
type AppMetricsSummary struct {
	App                 string
	SessionCount        int
	AveragePerformance  float64
	BestPerformance     float64
	LastSessionTime     time.Time
	TotalTimeSpent      float64 // minutes
	ConsistencyScore    float64 // 0-100 based on regularity
	ImprovementTrend    float64 // % change over 7 days
}

// UnifiedLeaderboard represents cross-app rankings
type UnifiedLeaderboard struct {
	Category string  // "typing_wpm", "math_accuracy", "reading_comprehension", etc.
	Entries  []LeaderboardEntry
	UpdatedAt time.Time
}

// LeaderboardEntry represents a single entry in a leaderboard
type LeaderboardEntry struct {
	Rank        int
	UserID      uint
	Username    string
	App         string
	MetricValue float64
	MetricLabel string  // "125 WPM", "95% accuracy", etc.
	Timestamp   time.Time
}

// SystemStats represents platform-wide statistics
type SystemStats struct {
	TotalUsers          int
	ActiveUsersToday    int
	ActiveUsersThisWeek int
	ActiveUsersThisMonth int

	// App usage
	AppUsageCount       map[string]int    // App -> session count
	AppAverageScore     map[string]float64 // App -> avg performance

	// Global trends
	MostPopularApp      string
	MostImprovedUsers   []UserImprovement
	TopPerformers       []TopPerformer

	// Engagement
	AverageDailyStreak  float64
	TotalSessionsAll    int
	AverageSessionLength float64
	PlatformUptime      float64 // percentage
}

// UserImprovement represents a user with notable improvement
type UserImprovement struct {
	UserID          uint
	Username        string
	App             string
	ImprovementRate float64 // % improvement
	Period          string  // "week", "month"
}

// TopPerformer represents a top-performing user
type TopPerformer struct {
	UserID      uint
	Username    string
	App         string
	MetricValue float64
	MetricName  string
	Rank        int
}

// SkillLevel represents a user's skill level in an app
type SkillLevel struct {
	App              string
	NormalizedScore  float64 // 0-100
	Level            string  // "beginner", "intermediate", "advanced", "expert"
	SessionCount     int
	MasteredItems    int
	CurrentFocus     string
}

// UserDailyActivity tracks activity for a specific day
type UserDailyActivity struct {
	UserID          uint
	Date            time.Time
	AppsUsed        []string
	SessionCount    int
	TotalMinutes    float64
	AverageScore    float64
	SessionDetails  []DailySessionDetail
}

// DailySessionDetail represents a single session on a given day
type DailySessionDetail struct {
	App             string
	SessionCount    int
	Duration        float64
	PerformanceScore float64
	TimeOfDay       string // morning, afternoon, evening, night
}

// RecommendationData represents actionable recommendations for a user
type RecommendationData struct {
	UserID              uint
	GeneratedAt         time.Time

	// Recommendations by type
	AppRecommendations  []AppRecommendation
	DifficultyAdvice    string
	PracticeTimeAdvice  string
	FocusAreas          []FocusArea

	// Goals
	SuggestedGoals      []Goal
	MilestoneDistance   map[string]int // App -> sessions until next milestone
}

// AppRecommendation represents a recommendation for a specific app
type AppRecommendation struct {
	App             string
	Reason          string
	Priority        string // "low", "medium", "high"
	SuggestedAction string
	ExpectedBenefit string
}

// FocusArea represents an area where the user should focus
type FocusArea struct {
	App             string
	Area            string
	CurrentLevel    float64
	TargetLevel     float64
	EstimatedSessions int
	Priority        string
}

// Goal represents a specific goal for a user
type Goal struct {
	ID              uint
	UserID          uint
	App             string
	Description     string
	TargetValue     float64
	CurrentValue    float64
	Deadline        *time.Time
	Progress        float64 // 0-100
	Status          string  // "active", "completed", "abandoned"
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AggregationMetrics tracks metrics about the aggregation process
type AggregationMetrics struct {
	UserID           uint
	LastAggregation  time.Time
	DataFreshness    map[string]time.Time // App -> last update time
	AggregationScore float64 // 0-100 based on data completeness
	MissingApps      []string
	Notes            string
}
