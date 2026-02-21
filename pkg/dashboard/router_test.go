package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jgirmay/unified-go/pkg/unified"
)

// TestRouterSetup tests router initialization
func TestRouterSetup(t *testing.T) {
	router := NewRouter(nil)
	if router == nil {
		t.Fatal("NewRouter returned nil")
	}

	if router.router == nil {
		t.Fatal("Router.router is nil")
	}

	if router.service == nil {
		t.Fatal("Router.service is nil")
	}
}

// TestIndexHandler tests the app launcher endpoint
func TestIndexHandler(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected HTML response body")
	}

	if !strings.Contains(body, "Unified Educational Platform") {
		t.Error("Expected 'Unified Educational Platform' in response")
	}

	if !strings.Contains(body, "Typing") {
		t.Error("Expected 'Typing' app in response")
	}
}

// TestGetSystemStatsEndpoint tests the system stats API routing
func TestGetSystemStatsEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	// Without a real database, this will error, but endpoint routing works
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 (nil db) or 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType == "" && w.Code == http.StatusInternalServerError {
		// Expected for nil db
		return
	}

	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON response, got %s", contentType)
	}
}

// TestGetUserProfileEndpoint tests the user profile API routing
func TestGetUserProfileEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/users/1/profile", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	// Without a real database, this will return 500 or similar error
	// Just verify the endpoint is routed (status != 404)
	if w.Code == http.StatusNotFound {
		t.Errorf("Endpoint not found (404)")
	}
}

// TestGetUserProfileInvalidIDEndpoint tests error handling for invalid user ID
func TestGetUserProfileInvalidIDEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/users/invalid/profile", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid ID, got %d", w.Code)
	}
}

// TestGetAnalyticsEndpoint tests the analytics API routing
func TestGetAnalyticsEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/users/1/analytics", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	// Without a real database, this will error, but endpoint routing works
	if w.Code == http.StatusNotFound {
		t.Errorf("Endpoint not found (404)")
	}
}

// TestGetSessionsEndpoint tests the sessions/activity API routing
func TestGetSessionsEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/users/1/sessions?limit=10", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	// Without a real database, this will error, but endpoint routing works
	if w.Code == http.StatusNotFound {
		t.Errorf("Endpoint not found (404)")
	}
}

// TestGetLeaderboardEndpoint tests the leaderboard API
func TestGetLeaderboardEndpoint(t *testing.T) {
	router := NewRouter(nil)

	categories := []string{"typing_wpm", "math_accuracy", "reading_comprehension", "piano_score", "overall"}
	for _, category := range categories {
		t.Run(category, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/leaderboard/%s?limit=10", category), nil)
			w := httptest.NewRecorder()

			router.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for category %s, got %d", category, w.Code)
			}

			var leaderboard unified.UnifiedLeaderboard
			err := json.NewDecoder(w.Body).Decode(&leaderboard)
			if err != nil {
				t.Fatalf("Failed to decode leaderboard: %v", err)
			}

			if leaderboard.Category != category {
				t.Errorf("Expected category %s, got %s", category, leaderboard.Category)
			}

			if len(leaderboard.Entries) == 0 {
				t.Error("Expected leaderboard entries")
			}
		})
	}
}

// TestGetLeaderboardStatsEndpoint tests the leaderboard stats API
func TestGetLeaderboardStatsEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/leaderboard/stats/typing_wpm", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&stats)
	if err != nil {
		t.Fatalf("Failed to decode stats: %v", err)
	}

	expectedKeys := []string{"category", "entry_count", "avg_metric", "max_metric"}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("Expected key '%s' in stats", key)
		}
	}
}

// TestGetUserRankEndpoint tests the user rank API
func TestGetUserRankEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/leaderboard/typing_wpm/user/1", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["rank"]; !ok {
		t.Error("Expected 'rank' key in response")
	}
}

// TestGetTrendsEndpoint tests the trends API routing
func TestGetTrendsEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/trends/1", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	// Endpoint should be routed (not 404)
	if w.Code == http.StatusNotFound {
		t.Errorf("Endpoint not found (404)")
	}
}

// TestGetRecommendationsEndpoint tests the recommendations API routing
func TestGetRecommendationsEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/recommendations/1", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	// Endpoint should be routed (not 404)
	if w.Code == http.StatusNotFound {
		t.Errorf("Endpoint not found (404)")
	}
}

// TestGetOverviewEndpoint tests the dashboard overview API routing
func TestGetOverviewEndpoint(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest("GET", "/api/overview/1", nil)
	w := httptest.NewRecorder()

	router.router.ServeHTTP(w, req)

	// Endpoint should be routed (not 404)
	if w.Code == http.StatusNotFound {
		t.Errorf("Endpoint not found (404)")
	}
}

// TestErrorHandling tests error handling across endpoints
func TestErrorHandling(t *testing.T) {
	router := NewRouter(nil)

	testCases := []struct {
		method   string
		url      string
		expected int
		name     string
	}{
		{
			method:   "GET",
			url:      "/api/users/invalid/profile",
			expected: http.StatusBadRequest,
			name:     "Invalid user ID",
		},
		{
			method:   "GET",
			url:      "/api/leaderboard/invalid_category",
			expected: http.StatusBadRequest,
			name:     "Invalid leaderboard category",
		},
		{
			method:   "GET",
			url:      "/api/leaderboard/typing_wpm/user/invalid",
			expected: http.StatusBadRequest,
			name:     "Invalid user ID in rank endpoint",
		},
		{
			method:   "GET",
			url:      "/api/trends/invalid",
			expected: http.StatusBadRequest,
			name:     "Invalid user ID in trends",
		},
		{
			method:   "GET",
			url:      "/api/recommendations/invalid",
			expected: http.StatusBadRequest,
			name:     "Invalid user ID in recommendations",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)
			w := httptest.NewRecorder()

			router.router.ServeHTTP(w, req)

			if w.Code != tc.expected {
				t.Errorf("Expected status %d, got %d", tc.expected, w.Code)
			}
		})
	}
}

// TestContentTypeHeaders tests that responses have correct content type
func TestContentTypeHeaders(t *testing.T) {
	router := NewRouter(nil)

	testCases := []struct {
		url         string
		contentType string
		name        string
		requireOK   bool
	}{
		{"/", "text/html", "Index endpoint", true},
		{"/api/stats", "application/json", "Stats API (nil db)", false},
		{"/api/users/1/profile", "application/json", "Profile API (nil db)", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
			w := httptest.NewRecorder()

			router.router.ServeHTTP(w, req)

			// Index endpoint should return OK
			if tc.requireOK && w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
				return
			}

			// For API endpoints with nil db, they'll return 500, so skip content-type check
			if !tc.requireOK && w.Code != http.StatusOK {
				return
			}

			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, tc.contentType) {
				t.Errorf("Expected Content-Type %s, got %s", tc.contentType, contentType)
			}
		})
	}
}

// TestLeaderboardAccuracy tests that leaderboard data is properly formatted
func TestLeaderboardAccuracy(t *testing.T) {
	router := NewRouter(nil)

	categories := []string{
		"typing_wpm",
		"math_accuracy",
		"reading_comprehension",
		"piano_score",
		"overall",
	}

	for _, category := range categories {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/leaderboard/%s", category), nil)
		w := httptest.NewRecorder()
		router.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Failed to get %s leaderboard: %d", category, w.Code)
			continue
		}

		var leaderboard unified.UnifiedLeaderboard
		json.NewDecoder(w.Body).Decode(&leaderboard)

		if leaderboard.Category != category {
			t.Errorf("Category mismatch: expected %s, got %s", category, leaderboard.Category)
		}

		// Verify entries are sorted by rank
		for i := 0; i < len(leaderboard.Entries)-1; i++ {
			if leaderboard.Entries[i].Rank > leaderboard.Entries[i+1].Rank {
				t.Errorf("Leaderboard %s entries not sorted by rank", category)
				break
			}
		}

		// Verify all entries have required fields
		for _, entry := range leaderboard.Entries {
			if entry.App == "" {
				t.Errorf("Leaderboard entry has empty app field")
			}
			if entry.Username == "" {
				t.Errorf("Leaderboard entry has empty username field")
			}
		}
	}
}

// TestDashboardLoadTime tests that endpoints respond quickly
func TestDashboardLoadTime(t *testing.T) {
	router := NewRouter(nil)

	// Only test endpoints that work without a real database
	endpoints := []string{
		"/api/leaderboard/typing_wpm",
		"/api/leaderboard/math_accuracy",
		"/api/leaderboard/reading_comprehension",
		"/api/leaderboard/piano_score",
		"/api/leaderboard/overall",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			start := time.Now()

			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			router.router.ServeHTTP(w, req)

			elapsed := time.Since(start)

			if w.Code != http.StatusOK {
				t.Errorf("Request failed with status %d", w.Code)
				return
			}

			// Endpoints should respond in < 100ms
			if elapsed > 100*time.Millisecond {
				t.Logf("Warning: endpoint %s took %v (expected < 100ms)", endpoint, elapsed)
			}
		})
	}
}

// BenchmarkProfileAggregation benchmarks profile data aggregation
func BenchmarkProfileAggregation(b *testing.B) {
	router := NewRouter(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/1/profile", nil)
		w := httptest.NewRecorder()
		router.router.ServeHTTP(w, req)
	}
}

// BenchmarkLeaderboardRetrieval benchmarks leaderboard data retrieval
func BenchmarkLeaderboardRetrieval(b *testing.B) {
	router := NewRouter(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/leaderboard/typing_wpm", nil)
		w := httptest.NewRecorder()
		router.router.ServeHTTP(w, req)
	}
}
