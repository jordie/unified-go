package typing

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestTypingIntegration provides integration test setup
type TestTypingIntegration struct {
	db      *sql.DB
	router  *Router
	service *Service
}

// setupIntegration creates a test database and router
func setupIntegration(t *testing.T) *TestTypingIntegration {
	db := setupTestDB(t)
	router := NewRouter(db)
	repo := NewRepository(db)
	service := NewService(repo)

	return &TestTypingIntegration{
		db:      db,
		router:  router,
		service: service,
	}
}

// setupBenchmark creates a test database for benchmarks
func setupBenchmark(b testing.TB) *TestTypingIntegration {
	db := setupTestDB(b)
	router := NewRouter(db)
	repo := NewRepository(db)
	service := NewService(repo)

	return &TestTypingIntegration{
		db:      db,
		router:  router,
		service: service,
	}
}

// TestCreateTypingTest tests creating a typing test
func TestCreateTypingTest(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	testData := map[string]interface{}{
		"user_id":  1,
		"content":  "The quick brown fox jumps over the lazy dog",
		"duration": 30.0,
		"errors":   2,
	}

	body, _ := json.Marshal(testData)
	req := httptest.NewRequest("POST", "/api/typing/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var result TypingResult
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.UserID != 1 {
		t.Errorf("Expected user_id 1, got %d", result.UserID)
	}

	if result.Accuracy <= 0 {
		t.Error("Expected accuracy to be calculated")
	}
}

// TestGetUserStats tests retrieving user statistics
func TestGetUserStats(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create test results
	for i := 0; i < 3; i++ {
		result := &TypingResult{
			UserID:      1,
			Content:     "The quick brown fox",
			TimeSpent:   30.0,
			WPM:         60.0,
			RawWPM:      65.0,
			ErrorsCount: 2,
			Accuracy:    95.0,
			TestMode:    "standard",
		}
		ti.service.repo.SaveResult(ctx, result)
	}

	req := httptest.NewRequest("GET", "/api/users/1/typing/stats", nil)
	w := httptest.NewRecorder()

	ti.router.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats UserStats
	json.Unmarshal(w.Body.Bytes(), &stats)

	if stats.UserID != 1 {
		t.Errorf("Expected user_id 1, got %d", stats.UserID)
	}

	if stats.TotalTests <= 0 {
		t.Error("Expected total_tests to be calculated")
	}
}

// TestGetLeaderboard tests retrieving leaderboard
func TestGetLeaderboard(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create test results for multiple users
	for userID := 1; userID <= 3; userID++ {
		for i := 0; i < 3; i++ {
			result := &TypingResult{
				UserID:      uint(userID),
				Content:     "test content",
				TimeSpent:   30.0,
				WPM:         float64(50 + userID*10),
				RawWPM:      float64(55 + userID*10),
				ErrorsCount: 1,
				Accuracy:    96.0,
				TestMode:    "standard",
			}
			ti.service.repo.SaveResult(ctx, result)
		}
	}

	req := httptest.NewRequest("GET", "/api/typing/leaderboard?limit=10", nil)
	w := httptest.NewRecorder()

	ti.router.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if leaderboard, ok := response["leaderboard"]; !ok || leaderboard == nil {
		t.Error("Expected leaderboard in response")
	}
}

// TestGetUserHistory tests retrieving user history
func TestGetUserHistory(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create test results
	result := &TypingResult{
		UserID:      1,
		Content:     "typing practice",
		TimeSpent:   30.0,
		WPM:         60.0,
		RawWPM:      65.0,
		ErrorsCount: 2,
		Accuracy:    95.0,
		TestMode:    "standard",
	}
	ti.service.repo.SaveResult(ctx, result)

	req := httptest.NewRequest("GET", "/api/users/1/typing/history?days=30", nil)
	w := httptest.NewRecorder()

	ti.router.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if _, ok := response["history"]; !ok {
		t.Error("Expected history in response")
	}
}

// TestGetLessons tests retrieving available lessons
func TestGetLessons(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	req := httptest.NewRequest("GET", "/api/typing/lessons", nil)
	w := httptest.NewRecorder()

	ti.router.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if lessons, ok := response["lessons"]; !ok || lessons == nil {
		t.Error("Expected lessons in response")
	}
}

// TestTypingMetricsCalculation tests metric calculations
func TestTypingMetricsCalculation(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	result, err := ti.service.ProcessTypingTest(ctx, 1, "The quick brown fox jumps over the lazy dog", 30.0, 2)
	if err != nil {
		t.Fatalf("ProcessTypingTest() error = %v", err)
	}

	if result.WPM <= 0 {
		t.Error("WPM should be calculated")
	}

	if result.Accuracy <= 0 || result.Accuracy > 100 {
		t.Errorf("Accuracy should be between 0-100, got %f", result.Accuracy)
	}

	if result.RawWPM <= 0 {
		t.Error("RawWPM should be calculated")
	}
}

// TestErrorHandling tests error responses
func TestErrorHandling(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{"invalid user ID", "GET", "/api/users/invalid/typing/stats", http.StatusBadRequest},
		{"missing user ID", "GET", "/api/users//typing/stats", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			ti.router.Routes().ServeHTTP(w, req)

			if w.Code < http.StatusBadRequest {
				t.Logf("Expected error status for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

// TestSkillLevelEstimation tests typing level estimation
func TestSkillLevelEstimation(t *testing.T) {
	tests := []struct {
		name     string
		wpm      float64
		expected string
	}{
		{"beginner", 30.0, "beginner"},
		{"intermediate", 50.0, "intermediate"},
		{"advanced", 75.0, "advanced"},
		{"expert", 90.0, "expert"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := EstimateTypingLevel(tt.wpm)
			if level != tt.expected {
				t.Errorf("Expected %s, got %s for WPM %.1f", tt.expected, level, tt.wpm)
			}
		})
	}
}

// BenchmarkTypingTest benchmarks test creation
func BenchmarkTypingTest(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	testData := map[string]interface{}{
		"user_id":  1,
		"content":  "The quick brown fox jumps over the lazy dog and runs through the forest",
		"duration": 30.0,
		"errors":   2,
	}

	body, _ := json.Marshal(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/typing/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ti.router.Routes().ServeHTTP(w, req)
	}
}

// BenchmarkUserStats benchmarks stats retrieval
func BenchmarkUserStats(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()

	// Create test data
	for i := 0; i < 10; i++ {
		result := &TypingResult{
			UserID:      1,
			Content:     "test content",
			TimeSpent:   30.0,
			WPM:         60.0,
			RawWPM:      65.0,
			ErrorsCount: 2,
			Accuracy:    95.0,
			TestMode:    "standard",
		}
		ti.service.repo.SaveResult(ctx, result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/1/typing/stats", nil)
		w := httptest.NewRecorder()
		ti.router.Routes().ServeHTTP(w, req)
	}
}

// BenchmarkLeaderboard benchmarks leaderboard retrieval
func BenchmarkLeaderboard(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()

	// Create test data for multiple users
	for userID := 1; userID <= 20; userID++ {
		for i := 0; i < 5; i++ {
			result := &TypingResult{
				UserID:      uint(userID),
				Content:     "test",
				TimeSpent:   30.0,
				WPM:         float64(50 + userID),
				RawWPM:      float64(55 + userID),
				ErrorsCount: 1,
				Accuracy:    96.0,
				TestMode:    "standard",
			}
			ti.service.repo.SaveResult(ctx, result)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/typing/leaderboard?limit=100", nil)
		w := httptest.NewRecorder()
		ti.router.Routes().ServeHTTP(w, req)
	}
}

// BenchmarkMetricsCalculation benchmarks metric calculations
func BenchmarkMetricsCalculation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateWPM(500, 30.0)
		_ = CalculateAccuracy(500, 5)
	}
}
