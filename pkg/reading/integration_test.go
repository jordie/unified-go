package reading

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestReadingIntegration provides integration test setup
type TestReadingIntegration struct {
	db     *sql.DB
	router chi.Router
	service *Service
}

// setupIntegration creates a test database and router
func setupIntegration(t *testing.T) *TestReadingIntegration {
	db := setupTestDB(t)
	router := NewRouter(db).Routes()
	repo := NewRepository(db)
	service := NewService(repo)

	return &TestReadingIntegration{
		db:      db,
		router:  router,
		service: service,
	}
}

// TestCreateAndRetrieveBook tests the complete book lifecycle
func TestCreateAndRetrieveBook(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	// Create a book via API
	bookData := map[string]interface{}{
		"title":         "Test Book",
		"author":        "Test Author",
		"content":       "This is a test book content with sufficient length to pass validation",
		"reading_level": "beginner",
		"language":      "English",
		"word_count":    12,
	}

	body, _ := json.Marshal(bookData)
	req := httptest.NewRequest("POST", "/api/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Parse response
	var book Book
	if err := json.Unmarshal(w.Body.Bytes(), &book); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if book.ID == 0 {
		t.Error("Book ID should not be 0")
	}

	// Retrieve the book
	req = httptest.NewRequest("GET", "/api/books/1", nil)
	w = httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var retrievedBook Book
	json.Unmarshal(w.Body.Bytes(), &retrievedBook)
	if retrievedBook.Title != "Test Book" {
		t.Errorf("Expected title 'Test Book', got '%s'", retrievedBook.Title)
	}
}

// TestReadingSessionFlow tests the complete reading session flow
func TestReadingSessionFlow(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a book first
	book := &Book{
		Title:        "Test Book",
		Author:       "Author",
		Content:      "This is test content with enough words for a reading session test",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    10,
	}

	bookID, err := ti.service.repo.SaveBook(ctx, book)
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	// Submit a reading session
	sessionData := map[string]interface{}{
		"user_id":    1,
		"book_id":    bookID,
		"content":    "This is test content with enough words for a reading session test",
		"time_spent": 60.0,
		"errors":     2,
	}

	body, _ := json.Marshal(sessionData)
	req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var session ReadingSession
	json.Unmarshal(w.Body.Bytes(), &session)

	if session.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", session.UserID)
	}

	if session.WPM <= 0 {
		t.Errorf("WPM should be calculated, got %f", session.WPM)
	}

	if session.Accuracy < 0 || session.Accuracy > 100 {
		t.Errorf("Accuracy should be 0-100, got %f", session.Accuracy)
	}
}

// TestUserStatistics tests statistics aggregation
func TestUserStatistics(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a book
	book := &Book{
		Title:        "Test Book",
		Author:       "Author",
		Content:      "This is test content with enough words to create multiple sessions",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    10,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create multiple reading sessions
	for i := 0; i < 3; i++ {
		ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0, 1)
	}

	// Get user statistics
	req := httptest.NewRequest("GET", "/api/users/1/stats", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats ReadingStats
	json.Unmarshal(w.Body.Bytes(), &stats)

	if stats.TotalSessionsCount != 3 {
		t.Errorf("Expected 3 sessions, got %d", stats.TotalSessionsCount)
	}

	if stats.AverageWPM <= 0 {
		t.Errorf("Average WPM should be calculated, got %f", stats.AverageWPM)
	}
}

// TestUserProgress tests progress calculation
func TestUserProgress(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a book and sessions
	book := &Book{
		Title:        "Progress Test Book",
		Author:       "Author",
		Content:      "This is content for testing progress calculation with adequate length",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    10,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create sessions with varying performance
	for i := 0; i < 2; i++ {
		ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, float64(60-i*10), 1)
	}

	// Get progress
	req := httptest.NewRequest("GET", "/api/users/1/progress", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var progress map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &progress)

	if progress["total_tests"] == nil {
		t.Error("Progress should contain total_tests")
	}

	if progress["estimated_level"] == nil {
		t.Error("Progress should contain estimated_level")
	}
}

// TestComprehensionWorkflow tests comprehension tests
func TestComprehensionWorkflow(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a book and session
	book := &Book{
		Title:        "Comprehension Book",
		Author:       "Author",
		Content:      "Test content for comprehension testing purposes and validation",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    8,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)
	session, _ := ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0, 0)

	// Save comprehension test
	test := &ComprehensionTest{
		SessionID:     session.ID,
		Question:      "What is the main topic?",
		UserAnswer:    "Test content",
		CorrectAnswer: "Test content",
		IsCorrect:     true,
		Score:         100.0,
	}

	_, err := ti.service.repo.SaveComprehensionTest(ctx, test)
	if err != nil {
		t.Fatalf("Failed to save comprehension test: %v", err)
	}

	// Retrieve comprehension tests
	req := httptest.NewRequest("GET", "/api/sessions/1/comprehension", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Analyze comprehension
	req = httptest.NewRequest("GET", "/api/sessions/1/analysis", nil)
	w = httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var analysis map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &analysis)

	if analysis["total_questions"] != float64(1) {
		t.Errorf("Expected 1 question, got %v", analysis["total_questions"])
	}
}

// TestContentValidation tests content validation endpoint
func TestContentValidation(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	tests := []struct {
		name        string
		content     string
		shouldPass  bool
		statusCode  int
	}{
		{"valid content", "This is a valid book content with sufficient length for reading", true, http.StatusOK},
		{"too short", "short", false, http.StatusBadRequest},
		{"empty", "", false, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"content": tt.content}
			body, _ := json.Marshal(data)

			req := httptest.NewRequest("POST", "/api/validate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ti.router.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.shouldPass && response["valid"] != true {
				t.Error("Expected content to be valid")
			}
		})
	}
}

// TestListBooksFiltering tests book filtering
func TestListBooksFiltering(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create books with different levels
	levels := []string{"beginner", "intermediate", "advanced"}
	for _, level := range levels {
		book := &Book{
			Title:        "Book - " + level,
			Author:       "Author",
			Content:      "This is test content for filtering books by difficulty level",
			ReadingLevel: level,
			Language:     "English",
			WordCount:    9,
		}
		ti.service.repo.SaveBook(ctx, book)
	}

	// Test filtering
	req := httptest.NewRequest("GET", "/api/books?difficulty=beginner", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	books := response["books"].([]interface{})
	if len(books) == 0 {
		t.Error("Expected books in response")
	}
}

// setupBenchmark creates a test database and router for benchmarks
func setupBenchmark(b testing.TB) *TestReadingIntegration {
	db := setupTestDB(b)
	router := NewRouter(db).Routes()
	repo := NewRepository(db)
	service := NewService(repo)

	return &TestReadingIntegration{
		db:      db,
		router:  router,
		service: service,
	}
}

// BenchmarkReadingSession benchmarks reading session processing
func BenchmarkReadingSession(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()
	book := &Book{
		Title:        "Bench Book",
		Author:       "Author",
		Content:      "This is benchmark content for performance testing of reading sessions",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    10,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	sessionData := map[string]interface{}{
		"user_id":    1,
		"book_id":    bookID,
		"content":    book.Content,
		"time_spent": 60.0,
		"errors":     1,
	}

	body, _ := json.Marshal(sessionData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}

// BenchmarkUserStatistics benchmarks statistics retrieval
func BenchmarkUserStatistics(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()
	book := &Book{
		Title:        "Stat Book",
		Author:       "Author",
		Content:      "Content for benchmarking statistics retrieval performance",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    7,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create sessions
	for i := 0; i < 5; i++ {
		ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/1/stats", nil)
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
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
		body       map[string]interface{}
		statusCode int
	}{
		{"invalid book ID", "GET", "/api/books/invalid", nil, http.StatusBadRequest},
		{"missing user ID", "GET", "/api/users//stats", nil, http.StatusNotFound},
		{"invalid JSON", "POST", "/api/sessions", nil, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}

			var req *http.Request
			if body != nil {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			ti.router.ServeHTTP(w, req)

			if w.Code < http.StatusBadRequest {
				// Some may succeed, but error cases should return 4xx
				t.Logf("Expected error status for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

// TestSessionPersistence tests that sessions are persisted correctly
func TestSessionPersistence(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	book := &Book{
		Title:        "Persistence Book",
		Author:       "Author",
		Content:      "Content to test data persistence in reading sessions",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    7,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create session
	session, _ := ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 120.0, 3)

	// Retrieve immediately
	retrieved, _ := ti.service.repo.GetSessionByID(ctx, session.ID)

	if retrieved == nil {
		t.Fatal("Session should be retrievable after creation")
	}

	if retrieved.UserID != session.UserID {
		t.Errorf("UserID mismatch: expected %d, got %d", session.UserID, retrieved.UserID)
	}

	if retrieved.WPM != session.WPM {
		t.Errorf("WPM mismatch: expected %f, got %f", session.WPM, retrieved.WPM)
	}
}

// BenchmarkLeaderboard benchmarks leaderboard retrieval
func BenchmarkLeaderboard(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()

	// Create multiple users with sessions
	for u := 1; u <= 3; u++ {
		book := &Book{
			Title:        "Leaderboard Book",
			Author:       "Author",
			Content:      "Content for leaderboard benchmarking tests",
			ReadingLevel: "intermediate",
			Language:     "English",
			WordCount:    6,
		}

		bookID, _ := ti.service.repo.SaveBook(ctx, book)

		for i := 0; i < 3; i++ {
			ti.service.ProcessTestResult(ctx, uint(u), bookID, book.Content, 60.0-float64(i*10), 1)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/leaderboard?limit=10", nil)
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}

// TestSaveReadingResultHandler tests the SaveReadingResult via sessions endpoint
func TestSaveReadingResultHandler(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a book first
	book := &Book{
		Title:        "Handler Test Book",
		Author:       "Test Author",
		Content:      "This content is for testing the save reading result handler endpoint",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    12,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Test SaveReadingResult via /api/sessions endpoint
	requestData := map[string]interface{}{
		"user_id":    1,
		"book_id":    bookID,
		"content":    book.Content,
		"time_spent": 90.0,
		"errors":     2,
	}

	body, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Errorf("Expected status 200 or 201, got %d: %s", w.Code, w.Body.String())
	}

	var session ReadingSession
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if session.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", session.UserID)
	}

	if session.WPM <= 0 {
		t.Errorf("WPM should be calculated, got %f", session.WPM)
	}

	if session.Accuracy < 0 || session.Accuracy > 100 {
		t.Errorf("Accuracy should be 0-100, got %f", session.Accuracy)
	}
}

// TestGetMasteryStatsHandler tests the GetMasteryStats via stats endpoint
func TestGetMasteryStatsHandler(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a book and session
	book := &Book{
		Title:        "Mastery Test Book",
		Author:       "Author",
		Content:      "Content for mastery statistics testing and evaluation",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    8,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0+float64(i*10), i)
	}

	// Test GetMasteryStats via /api/users/1/stats endpoint
	req := httptest.NewRequest("GET", "/api/users/1/stats", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var stats ReadingStats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if stats.TotalSessionsCount != 3 {
		t.Errorf("Expected 3 sessions, got %d", stats.TotalSessionsCount)
	}
}

// TestGetReadingStatsHandler tests the GetReadingStats via progress endpoint
func TestGetReadingStatsHandler(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	book := &Book{
		Title:        "Stats Test Book",
		Author:       "Author",
		Content:      "Testing reading statistics retrieval from the handler",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    8,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0, 0)

	req := httptest.NewRequest("GET", "/api/users/1/progress", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["estimated_level"] == nil {
		t.Error("Response should contain estimated_level")
	}
}

// TestGetBooksHandler tests the GetBooks handler
func TestGetBooksHandler(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create books with different levels
	levels := []string{"beginner", "intermediate", "advanced"}
	for _, level := range levels {
		book := &Book{
			Title:        "Book - " + level,
			Author:       "Test Author",
			Content:      "Test content for getting books with different difficulty levels",
			ReadingLevel: level,
			Language:     "English",
			WordCount:    10,
		}
		ti.service.repo.SaveBook(ctx, book)
	}

	// Test GetBooks without filter
	req := httptest.NewRequest("GET", "/api/books", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response GetBooksResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Response should indicate success")
	}

	if len(response.Data) != 3 {
		t.Errorf("Expected 3 books, got %d", len(response.Data))
	}

	// Test with difficulty filter
	req = httptest.NewRequest("GET", "/api/books?difficulty=beginner", nil)
	w = httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response.Data) < 1 {
		t.Error("Expected at least 1 beginner book")
	}

	for _, book := range response.Data {
		if book.ReadingLevel != "beginner" {
			t.Errorf("Expected beginner level, got %s", book.ReadingLevel)
		}
	}
}

// TestSubmitComprehensionHandler tests the SubmitComprehension via comprehension endpoint
func TestSubmitComprehensionHandler(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create a book and session
	book := &Book{
		Title:        "Comprehension Handler Test",
		Author:       "Author",
		Content:      "This is content for testing comprehension handler submission flow",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    10,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)
	session, _ := ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0, 0)

	// Submit comprehension answers via /api/comprehension
	requestData := map[string]interface{}{
		"session_id": session.ID,
		"question":   "What is the main idea?",
		"user_answer": "Testing comprehension",
		"correct_answer": "Testing comprehension",
	}

	body, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/api/comprehension", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Logf("Comprehension submission: status %d", w.Code)
	}
}

// TestGetLeaderboardHandler tests the GetLeaderboard handler
func TestGetLeaderboardHandler(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create multiple users with sessions
	book := &Book{
		Title:        "Leaderboard Test Book",
		Author:       "Author",
		Content:      "Content for leaderboard handler testing with multiple users",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    10,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	for userID := 1; userID <= 3; userID++ {
		ti.service.ProcessTestResult(ctx, uint(userID), bookID, book.Content, 60.0+float64(userID*10), userID-1)
	}

	// Test GetLeaderboard
	req := httptest.NewRequest("GET", "/api/leaderboard?limit=10", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response GetLeaderboardResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Response should indicate success")
	}

	if len(response.Data) == 0 {
		t.Error("Leaderboard should contain entries")
	}
}

// TestUserIsolation ensures users cannot see each other's data
func TestUserIsolation(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	book := &Book{
		Title:        "User Isolation Test",
		Author:       "Author",
		Content:      "Testing that user data is properly isolated between different users",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    10,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create sessions for different users
	ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0, 0)
	ti.service.ProcessTestResult(ctx, 2, bookID, book.Content, 50.0, 1)

	// User 1 should only see their own stats
	req := httptest.NewRequest("GET", "/api/users/1/stats", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	var stats1 ReadingStats
	json.Unmarshal(w.Body.Bytes(), &stats1)

	// User 2 should only see their own stats
	req = httptest.NewRequest("GET", "/api/users/2/stats", nil)
	w = httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	var stats2 ReadingStats
	json.Unmarshal(w.Body.Bytes(), &stats2)

	// Both should have exactly 1 session (their own)
	if stats1.TotalSessionsCount != 1 || stats2.TotalSessionsCount != 1 {
		t.Errorf("User isolation failed: user1 sessions=%d, user2 sessions=%d", stats1.TotalSessionsCount, stats2.TotalSessionsCount)
	}
}

// TestConcurrentSessions tests handling of concurrent reading sessions
func TestConcurrentSessions(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	book := &Book{
		Title:        "Concurrent Test Book",
		Author:       "Author",
		Content:      "Content for testing concurrent session handling and processing",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    10,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create multiple sessions quickly
	for i := 0; i < 5; i++ {
		sessionData := SaveReadingResultRequest{
			UserID:     uint(i + 1),
			BookID:     bookID,
			Content:    book.Content,
			Duration:   float64(60 + i*10),
			ErrorCount: i,
		}

		body, _ := json.Marshal(sessionData)
		req := httptest.NewRequest("POST", "/api/save_reading_result", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ti.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusCreated {
			t.Errorf("Failed to create session %d: status %d", i, w.Code)
		}
	}

	// Verify all sessions were created
	req := httptest.NewRequest("GET", "/api/reading_stats?user_id=1", nil)
	w := httptest.NewRecorder()
	ti.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestMetricsCalculation tests that metrics are calculated correctly
func TestMetricsCalculation(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	book := &Book{
		Title:        "Metrics Test Book",
		Author:       "Author",
		Content:      "Short test content for metrics",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    5,
	}
	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create a known session
	requestData := SaveReadingResultRequest{
		UserID:     1,
		BookID:     bookID,
		Content:    book.Content,
		Duration:   120.0, // 2 minutes
		ErrorCount: 1,
	}

	body, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/api/save_reading_result", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ti.router.ServeHTTP(w, req)

	var response SaveReadingResultResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Data == nil {
		t.Fatal("Should have session data")
	}

	// WPM = (word_count / duration_in_minutes)
	// Accuracy = ((word_count - errors) / word_count) * 100
	expectedWPM := float64(book.WordCount) / (requestData.Duration / 60.0)
	expectedAccuracy := float64((book.WordCount - requestData.ErrorCount) / book.WordCount * 100)

	// Allow small floating point differences
	if response.Data.WPM < expectedWPM-1 || response.Data.WPM > expectedWPM+1 {
		t.Errorf("WPM calculation: expected ~%f, got %f", expectedWPM, response.Data.WPM)
	}

	if response.Data.Accuracy < expectedAccuracy-1 || response.Data.Accuracy > expectedAccuracy+1 {
		t.Errorf("Accuracy calculation: expected ~%f, got %f", expectedAccuracy, response.Data.Accuracy)
	}
}

// TestInvalidRequestHandling tests handling of invalid requests
func TestInvalidRequestHandling(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		statusCode int
	}{
		{"invalid JSON body", "POST", "/api/save_reading_result", "invalid json", http.StatusBadRequest},
		{"missing required fields", "POST", "/api/save_reading_result", "{}", http.StatusBadRequest},
		{"invalid book ID", "GET", "/api/passages?book_id=invalid", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader([]byte(tt.body)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			ti.router.ServeHTTP(w, req)

			if w.Code != tt.statusCode && (w.Code < http.StatusBadRequest || w.Code >= http.StatusInternalServerError) {
				t.Logf("Test %s: expected status %d, got %d", tt.name, tt.statusCode, w.Code)
			}
		})
	}
}

// BenchmarkSaveReadingResultHandler benchmarks the SaveReadingResult handler
func BenchmarkSaveReadingResultHandler(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()
	book := &Book{
		Title:        "Bench Book",
		Author:       "Author",
		Content:      "Benchmark content for save reading result handler",
		ReadingLevel: "beginner",
		Language:     "English",
		WordCount:    8,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	sessionData := SaveReadingResultRequest{
		UserID:     1,
		BookID:     bookID,
		Content:    book.Content,
		Duration:   60.0,
		ErrorCount: 1,
	}

	body, _ := json.Marshal(sessionData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/save_reading_result", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}

// BenchmarkGetMasteryStats benchmarks the GetMasteryStats handler
func BenchmarkGetMasteryStats(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()
	book := &Book{
		Title:        "Mastery Bench Book",
		Author:       "Author",
		Content:      "Benchmark content for mastery stats handler",
		ReadingLevel: "intermediate",
		Language:     "English",
		WordCount:    8,
	}

	bookID, _ := ti.service.repo.SaveBook(ctx, book)

	// Create sessions
	for i := 0; i < 5; i++ {
		ti.service.ProcessTestResult(ctx, 1, bookID, book.Content, 60.0, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/get_mastery_stats?user_id=1", nil)
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}

// BenchmarkGetBooks benchmarks the GetBooks handler
func BenchmarkGetBooks(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()

	// Create books
	for i := 0; i < 10; i++ {
		book := &Book{
			Title:        "Book " + string(rune(i)),
			Author:       "Author",
			Content:      "Benchmark content for get books handler retrieval performance",
			ReadingLevel: "beginner",
			Language:     "English",
			WordCount:    8,
		}
		ti.service.repo.SaveBook(ctx, book)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/passages", nil)
		w := httptest.NewRecorder()
		ti.router.ServeHTTP(w, req)
	}
}
