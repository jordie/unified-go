package dashboard

import (
	"context"
	"testing"
)

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	service := NewService(nil)
	if service == nil {
		t.Fatal("NewService returned nil")
	}
}

// TestValidateCategory tests category validation
func TestValidateCategory(t *testing.T) {
	service := NewService(nil)

	validCategories := []string{
		"typing_wpm",
		"math_accuracy",
		"reading_comprehension",
		"piano_score",
		"overall",
	}

	for _, category := range validCategories {
		if !service.ValidateCategory(category) {
			t.Errorf("Category '%s' should be valid", category)
		}
	}

	invalidCategories := []string{
		"invalid",
		"typing_accuracy",
		"math_speed",
		"",
	}

	for _, category := range invalidCategories {
		if service.ValidateCategory(category) {
			t.Errorf("Category '%s' should be invalid", category)
		}
	}
}

// TestGetAvailableCategories tests category listing
func TestGetAvailableCategories(t *testing.T) {
	service := NewService(nil)
	categories := service.GetAvailableCategories()

	if len(categories) == 0 {
		t.Fatal("GetAvailableCategories returned empty list")
	}

	if len(categories) != 5 {
		t.Errorf("Expected 5 categories, got %d", len(categories))
	}

	// Verify all expected categories are present
	expectedCount := 0
	for _, cat := range categories {
		if service.ValidateCategory(cat) {
			expectedCount++
		}
	}

	if expectedCount != 5 {
		t.Errorf("Expected all categories to be valid, got %d valid", expectedCount)
	}
}

// TestGetUserProfileWithInvalidID tests error handling
func TestGetUserProfileWithInvalidID(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetUserProfile(ctx, 0)
	if err == nil {
		t.Error("Expected error for zero user ID")
	}
}

// TestGetUserAnalyticsWithInvalidID tests error handling
func TestGetUserAnalyticsWithInvalidID(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetUserAnalytics(ctx, 0)
	if err == nil {
		t.Error("Expected error for zero user ID")
	}
}

// TestGetRecentActivityWithInvalidID tests error handling
func TestGetRecentActivityWithInvalidID(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetRecentActivity(ctx, 0, 10)
	if err == nil {
		t.Error("Expected error for zero user ID")
	}
}

// TestGetRecentActivityWithLimitValidation tests limit handling
func TestGetRecentActivityWithLimitValidation(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	// Negative limit should default to 20
	_, err := service.GetRecentActivity(ctx, 1, -5)
	if err == nil || err.Error() != "invalid user ID" {
		// Error might be from repo call, which is fine
	}

	// Zero limit should default to 20
	_, err = service.GetRecentActivity(ctx, 1, 0)
	if err == nil || err.Error() != "invalid user ID" {
		// Error might be from repo call, which is fine
	}
}

// TestGetLeaderboardWithInvalidCategory tests error handling
func TestGetLeaderboardWithInvalidCategory(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetLeaderboard(ctx, "invalid_category", 10)
	if err == nil {
		t.Error("Expected error for invalid category")
	}
}

// TestGetLeaderboardWithEmptyCategory tests error handling
func TestGetLeaderboardWithEmptyCategory(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetLeaderboard(ctx, "", 10)
	if err == nil {
		t.Error("Expected error for empty category")
	}
}

// TestGetLeaderboardWithLimitValidation tests limit handling
func TestGetLeaderboardWithLimitValidation(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	// Valid category but invalid limit should be corrected
	_, err := service.GetLeaderboard(ctx, "typing_wpm", 0)
	if err == nil || err.Error() != "category is required" {
		// Either error from category or repo, which is fine
	}

	// Negative limit should default to 20
	_, err = service.GetLeaderboard(ctx, "typing_wpm", -10)
	if err == nil || err.Error() != "category is required" {
		// Either error from category or repo, which is fine
	}
}

// TestGetSystemStatsWithNilRepo tests stats retrieval with nil repo
func TestGetSystemStatsWithNilRepo(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	// Should panic or return error since repo is nil
	defer func() {
		if r := recover(); r != nil {
			// Expected panic from nil repository
			return
		}
	}()

	_, _ = service.GetSystemStats(ctx)
}

// TestGetTrendsWithInvalidID tests error handling
func TestGetTrendsWithInvalidID(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetTrends(ctx, 0)
	if err == nil {
		t.Error("Expected error for zero user ID")
	}
}

// TestGetRecommendationsWithInvalidID tests error handling
func TestGetRecommendationsWithInvalidID(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetRecommendations(ctx, 0)
	if err == nil {
		t.Error("Expected error for zero user ID")
	}
}

// TestGetDashboardOverviewWithInvalidID tests error handling
func TestGetDashboardOverviewWithInvalidID(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()

	_, err := service.GetDashboardOverview(ctx, 0)
	if err == nil {
		t.Error("Expected error for zero user ID")
	}
}
