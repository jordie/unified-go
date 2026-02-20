package typing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/jgirmay/unified-go/internal/middleware"
)

// createTestRouter creates a router with test database
func createTestRouter(t *testing.T) *Router {
	db := setupTestDB(t)
	authMW := middleware.NewAuthMiddleware("test-secret", "test_session")
	return NewRouter(newPoolFromDB(db), authMW)
}

// createSessionCookie creates a session cookie with user_id for testing
// Note: For these tests, we use mock contexts instead of actual cookies
func createSessionCookie(t *testing.T, userID int) *http.Cookie {
	store := sessions.NewCookieStore([]byte("test-secret"))

	session, err := store.New(&http.Request{}, "unified_session")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	middleware.SetAuthenticated(session, userID, "testuser")

	// For testing, we'll use a mock session in context instead
	return nil
}

// TestIndexHandler tests GET /typing/
func TestIndexHandler(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/typing/", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("IndexHandler() status code = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("IndexHandler() content type = %s, want text/html; charset=utf-8", ct)
	}

	if w.Body.Len() == 0 {
		t.Error("IndexHandler() returned empty body")
	}
}

// TestSaveResultHandler tests POST /typing/api/save_result
func TestSaveResultHandler(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	reqBody := SaveResultRequest{
		Content:    "the quick brown fox jumps over the lazy dog",
		TimeSpent:  60,
		ErrorCount: 2,
		TestMode:   "paragraphs",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/typing/api/save_result", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should return 401 Unauthorized (no session middleware)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("SaveResultHandler() status code = %d, want %d (no session)", w.Code, http.StatusUnauthorized)
	}
}

// TestStatsHandlerUnauthorized tests GET /typing/api/stats without auth
func TestStatsHandlerUnauthorized(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/typing/api/stats", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("StatsHandler() status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	if response.Error != "Unauthorized" {
		t.Errorf("StatsHandler() error = %s, want 'Unauthorized'", response.Error)
	}
}

// TestLeaderboardHandler tests GET /typing/api/leaderboard
func TestLeaderboardHandler(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/typing/api/leaderboard?limit=10", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("LeaderboardHandler() status code = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("LeaderboardHandler() content type = %s, want application/json", ct)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("LeaderboardHandler() failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("LeaderboardHandler() success = %v, want true", response.Success)
	}
}

// TestHistoryHandlerUnauthorized tests GET /typing/api/history without auth
func TestHistoryHandlerUnauthorized(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/typing/api/history?limit=20&offset=0", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HistoryHandler() status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestSettingsHandlerUnauthorized tests POST /typing/api/settings without auth
func TestSettingsHandlerUnauthorized(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	settings := map[string]interface{}{
		"difficulty": "hard",
		"test_mode":  "paragraphs",
	}

	bodyBytes, _ := json.Marshal(settings)
	req := httptest.NewRequest("POST", "/typing/api/settings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("SettingsHandler() status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestSaveResultHandlerMethodNotAllowed tests wrong HTTP method
func TestSaveResultHandlerMethodNotAllowed(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/typing/api/save_result", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("SaveResultHandler() status code = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// TestStatsHandlerMethodNotAllowed tests wrong HTTP method
func TestStatsHandlerMethodNotAllowed(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("POST", "/typing/api/stats", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("StatsHandler() status code = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// TestLeaderboardHandlerWithLimitParam tests limit query parameter
func TestLeaderboardHandlerWithLimitParam(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	tests := []struct {
		name     string
		limitStr string
		expectOK bool
	}{
		{"valid limit", "5", true},
		{"max limit", "1000", true},
		{"invalid limit", "abc", true}, // Should use default
		{"zero limit", "0", true},      // Should use default
		{"negative limit", "-5", true}, // Should use default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/typing/api/leaderboard?limit=" + tt.limitStr
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if tt.expectOK && w.Code != http.StatusOK {
				t.Errorf("LeaderboardHandler() status code = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

// TestHistoryHandlerPagination tests pagination parameters
func TestHistoryHandlerPaginationParams(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	tests := []struct {
		name   string
		limit  string
		offset string
		ok     bool
	}{
		{"default pagination", "", "", true},
		{"with limit", "10", "", true},
		{"with offset", "", "5", true},
		{"both params", "10", "5", true},
		{"invalid limit", "abc", "", true},  // Should use default
		{"invalid offset", "", "xyz", true}, // Should use default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/typing/api/history"
			if tt.limit != "" || tt.offset != "" {
				url += "?"
				if tt.limit != "" {
					url += "limit=" + tt.limit
				}
				if tt.offset != "" {
					if tt.limit != "" {
						url += "&"
					}
					url += "offset=" + tt.offset
				}
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// All should fail due to missing auth
			if w.Code != http.StatusUnauthorized {
				t.Errorf("HistoryHandler() status code = %d, want %d (no auth)", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

// TestSaveResultHandlerInvalidJSON tests invalid JSON in request body
// Note: Returns 401 because auth check happens before JSON parsing
func TestSaveResultHandlerInvalidJSON(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("POST", "/typing/api/save_result", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should return 401 Unauthorized because no session (auth check happens first)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("SaveResultHandler() status code = %d, want %d (no auth)", w.Code, http.StatusUnauthorized)
	}
}

// TestAPIResponseFormat tests that API responses are properly formatted
func TestAPIResponseFormat(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	req := httptest.NewRequest("GET", "/typing/api/leaderboard", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check required fields
	if response.Success != true {
		t.Error("APIResponse missing or incorrect 'success' field")
	}

	if response.Data == nil {
		t.Error("APIResponse missing 'data' field")
	}
}

// TestNewRouter tests router creation
func TestNewRouter(t *testing.T) {
	db := setupTestDB(t)
	authMW := middleware.NewAuthMiddleware("secret", "session")

	router := NewRouter(newPoolFromDB(db), authMW)

	if router == nil {
		t.Error("NewRouter() returned nil")
	}

	if router.service == nil {
		t.Error("NewRouter() service is nil")
	}

	if router.mux == nil {
		t.Error("NewRouter() mux is nil")
	}
}

// TestRegisterRoutes tests route registration
func TestRegisterRoutes(t *testing.T) {
	router := createTestRouter(t)
	mux := router.RegisterRoutes()

	if mux == nil {
		t.Error("RegisterRoutes() returned nil")
	}

	// Test that routes are registered
	routes := []string{
		"/typing/",
		"/typing/api/save_result",
		"/typing/api/stats",
		"/typing/api/leaderboard",
		"/typing/api/history",
		"/typing/api/settings",
	}

	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		w := httptest.NewRecorder()

		// Just check that the route is handled (may return 401/405 depending on route)
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s not registered (404)", route)
		}
	}
}
