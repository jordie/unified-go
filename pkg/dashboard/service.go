package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jgirmay/unified-go/pkg/unified"
)

// Service provides business logic for the dashboard
type Service struct {
	unifiedRepo *unified.Repository
	unifiedSvc  *unified.Service
}

// NewService creates a new dashboard service
func NewService(unifiedRepo *unified.Repository) *Service {
	unifiedSvc := unified.NewService(unifiedRepo)
	return &Service{
		unifiedRepo: unifiedRepo,
		unifiedSvc:  unifiedSvc,
	}
}

// GetUserProfile aggregates and returns user profile data
func (s *Service) GetUserProfile(ctx context.Context, userID uint) (*unified.UnifiedUserProfile, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	profile, err := s.unifiedRepo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Calculate overall level if we have individual app levels
	if profile.TypingLevel > 0 || profile.MathLevel > 0 || profile.ReadingLevel > 0 || profile.PianoLevel > 0 {
		profile.OverallLevel = s.unifiedSvc.CalculateOverallLevel(profile)
	}

	return profile, nil
}

// GetUserAnalytics returns cross-app analytics for a user
func (s *Service) GetUserAnalytics(ctx context.Context, userID uint) (*unified.CrossAppAnalytics, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	analytics, err := s.unifiedRepo.GetCrossAppAnalytics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

// GetRecentActivity returns recent sessions for a user
func (s *Service) GetRecentActivity(ctx context.Context, userID uint, limit int) ([]unified.UnifiedSession, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sessions, err := s.unifiedRepo.GetRecentSessions(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent sessions: %w", err)
	}

	return sessions, nil
}

// GetLeaderboard returns a leaderboard for a specific category
func (s *Service) GetLeaderboard(ctx context.Context, category string, limit int) (*unified.UnifiedLeaderboard, error) {
	if category == "" {
		return nil, fmt.Errorf("category is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	lb, err := s.unifiedRepo.GetUnifiedLeaderboard(ctx, category, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return lb, nil
}

// GetSystemStats returns platform-wide statistics
func (s *Service) GetSystemStats(ctx context.Context) (*unified.SystemStats, error) {
	stats, err := s.unifiedRepo.GetSystemStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system stats: %w", err)
	}

	return stats, nil
}

// GetTrends calculates and returns platform trends
func (s *Service) GetTrends(ctx context.Context, userID uint) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Get recent sessions
	sessions, err := s.unifiedRepo.GetRecentSessions(ctx, userID, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Calculate trends
	trends := s.unifiedSvc.DetectTrends(sessions)
	consistency := s.unifiedSvc.CalculateConsistencyScore(sessions)
	streak := s.unifiedSvc.CalculateDailyStreak(sessions)

	result := map[string]interface{}{
		"weekly_trends": trends,
		"consistency":   consistency,
		"daily_streak":  streak,
		"session_count": len(sessions),
		"timestamp":     time.Now(),
	}

	return result, nil
}

// GetRecommendations returns personalized recommendations for a user
func (s *Service) GetRecommendations(ctx context.Context, userID uint) (*unified.RecommendationData, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Get profile
	profile, err := s.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Get analytics
	analytics, err := s.GetUserAnalytics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Generate recommendations
	recommendations := s.unifiedSvc.GenerateRecommendations(analytics, profile)

	return recommendations, nil
}

// GetDashboardOverview returns a comprehensive dashboard overview
func (s *Service) GetDashboardOverview(ctx context.Context, userID uint) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Get profile
	profile, err := s.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Get analytics
	analytics, err := s.GetUserAnalytics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Get recent activity
	sessions, err := s.GetRecentActivity(ctx, userID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Get recommendations
	recommendations, err := s.GetRecommendations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Get milestones
	milestones := s.unifiedSvc.IdentifyMilestones(profile)

	// Get trends
	trends, err := s.GetTrends(ctx, userID)
	if err != nil {
		trends = make(map[string]interface{})
	}

	overview := map[string]interface{}{
		"user_profile":    profile,
		"analytics":       analytics,
		"recent_activity": sessions,
		"recommendations": recommendations,
		"milestones":      milestones,
		"trends":          trends,
		"updated_at":      time.Now(),
	}

	return overview, nil
}

// ValidateCategory checks if a leaderboard category is valid
func (s *Service) ValidateCategory(category string) bool {
	validCategories := map[string]bool{
		"typing_wpm":           true,
		"math_accuracy":        true,
		"reading_comprehension": true,
		"piano_score":          true,
		"overall":              true,
	}
	return validCategories[category]
}

// GetAvailableCategories returns all available leaderboard categories
func (s *Service) GetAvailableCategories() []string {
	return []string{
		"typing_wpm",
		"math_accuracy",
		"reading_comprehension",
		"piano_score",
		"overall",
	}
}
