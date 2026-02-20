package math

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary in-memory SQLite database for testing
func setupTestDB(t testing.TB) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Create tables
	createTablesSQL := `
	CREATE TABLE math_problems (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT,
		difficulty TEXT,
		question TEXT,
		answer REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE math_solutions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		problem_id INTEGER,
		attempt REAL,
		correct INTEGER,
		time_spent REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE math_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		problem_type TEXT,
		difficulty TEXT,
		total_problems INTEGER,
		correct_answers INTEGER,
		score REAL,
		time_spent REAL,
		started_at TIMESTAMP,
		completed_at TIMESTAMP,
		average_time_per_problem REAL
	);

	CREATE TABLE math_user_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE,
		total_problems INTEGER DEFAULT 0,
		correct_answers INTEGER DEFAULT 0,
		accuracy REAL DEFAULT 0,
		average_time_per_problem REAL DEFAULT 0,
		best_score REAL DEFAULT 0,
		total_time_spent INTEGER DEFAULT 0,
		sessions_completed INTEGER DEFAULT 0,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_math_solutions_user_id ON math_solutions(user_id);
	CREATE INDEX idx_math_sessions_user_id ON math_sessions(user_id);
	`

	if _, err := db.Exec(createTablesSQL); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// Insert test user
	if _, err := db.Exec("INSERT INTO math_user_stats (user_id) VALUES (1)"); err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	return db
}

type testInstance struct {
	db      *sql.DB
	service *Service
	router  *Router
}

func setupIntegration(t testing.TB) *testInstance {
	db := setupTestDB(t)
	router := NewRouter(db)
	return &testInstance{
		db:      db,
		service: router.service,
		router:  router,
	}
}

func setupBenchmark(b *testing.B) *testInstance {
	db := setupTestDB(b)
	router := NewRouter(db)
	return &testInstance{
		db:      db,
		service: router.service,
		router:  router,
	}
}

// TestGenerateProblem tests problem generation
func TestGenerateProblem(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	testCases := []struct {
		problemType ProblemType
		difficulty  DifficultyLevel
	}{
		{Addition, Easy},
		{Subtraction, Medium},
		{Multiplication, Hard},
		{Division, VeryHard},
	}

	for _, tc := range testCases {
		question, answer, err := ti.service.GenerateProblem(tc.problemType, tc.difficulty)
		if err != nil {
			t.Fatalf("GenerateProblem() error = %v", err)
		}

		if question == "" {
			t.Errorf("GenerateProblem() returned empty question for %s", tc.problemType)
		}

		if answer <= 0 && tc.problemType != Division {
			t.Errorf("GenerateProblem() returned invalid answer for %s", tc.problemType)
		}
	}
}

// TestCompleteSession tests completing a quiz session
func TestCompleteSession(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()
	session := &QuizSession{
		UserID:                1,
		ProblemType:           Addition,
		Difficulty:            Easy,
		TotalProblems:         10,
		CorrectAnswers:        8,
		TimeSpent:             60.0,
		StartedAt:             time.Now().Add(-time.Minute),
		AverageTimePerProblem: 6.0,
	}

	err := ti.service.CompleteSession(ctx, session)
	if err != nil {
		t.Fatalf("CompleteSession() error = %v", err)
	}

	if session.Score != 80.0 {
		t.Errorf("CompleteSession() score = %f, want 80.0", session.Score)
	}
}

// TestGetUserStats tests retrieving user statistics
func TestGetUserStats(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Complete a session first
	session := &QuizSession{
		UserID:                1,
		ProblemType:           Addition,
		Difficulty:            Easy,
		TotalProblems:         10,
		CorrectAnswers:        8,
		TimeSpent:             60.0,
		StartedAt:             time.Now().Add(-time.Minute),
		AverageTimePerProblem: 6.0,
	}

	err := ti.service.CompleteSession(ctx, session)
	if err != nil {
		t.Fatalf("CompleteSession() error = %v", err)
	}

	// Retrieve stats
	stats, err := ti.service.GetUserStats(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if stats.UserID != 1 {
		t.Errorf("GetUserStats() returned incorrect user_id: %d", stats.UserID)
	}

	if stats.SessionsCompleted == 0 {
		t.Errorf("GetUserStats() returned no sessions")
	}
}

// TestGetLeaderboard tests retrieving leaderboard
func TestGetLeaderboard(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	ctx := context.Background()

	// Create multiple sessions
	for i := 1; i <= 5; i++ {
		correctAnswers := 6 + i
		if correctAnswers > 10 {
			correctAnswers = 10
		}
		session := &QuizSession{
			UserID:                uint(i),
			ProblemType:           Addition,
			Difficulty:            Easy,
			TotalProblems:         10,
			CorrectAnswers:        correctAnswers,
			TimeSpent:             60.0,
			StartedAt:             time.Now().Add(-time.Minute),
			AverageTimePerProblem: 6.0,
		}

		// Insert user stat entry if not exists
		ti.db.Exec("INSERT OR IGNORE INTO math_user_stats (user_id) VALUES (?)", i)

		err := ti.service.CompleteSession(ctx, session)
		if err != nil {
			t.Fatalf("CompleteSession() error = %v", err)
		}
	}

	leaderboard, err := ti.service.GetLeaderboard(ctx, 10)
	if err != nil {
		t.Fatalf("GetLeaderboard() error = %v", err)
	}

	if len(leaderboard) == 0 {
		t.Errorf("GetLeaderboard() returned empty leaderboard")
	}
}

// TestCalculateScore tests score calculation
func TestCalculateScore(t *testing.T) {
	testCases := []struct {
		correct int
		total   int
		want    float64
	}{
		{10, 10, 100.0},
		{5, 10, 50.0},
		{8, 10, 80.0},
		{0, 10, 0.0},
	}

	for _, tc := range testCases {
		got := CalculateScore(tc.correct, tc.total)
		if got != tc.want {
			t.Errorf("CalculateScore(%d, %d) = %f, want %f", tc.correct, tc.total, got, tc.want)
		}
	}
}

// TestEstimateMathLevel tests math skill level estimation
func TestEstimateMathLevel(t *testing.T) {
	testCases := []struct {
		accuracy float64
		want     string
	}{
		{40.0, "beginner"},
		{60.0, "intermediate"},
		{80.0, "advanced"},
		{90.0, "expert"},
	}

	for _, tc := range testCases {
		got := EstimateMathLevel(tc.accuracy)
		if got != tc.want {
			t.Errorf("EstimateMathLevel(%f) = %s, want %s", tc.accuracy, got, tc.want)
		}
	}
}

// TestSessionValidation tests QuizSession validation
func TestSessionValidation(t *testing.T) {
	testCases := []struct {
		name    string
		session *QuizSession
		wantErr bool
	}{
		{
			name: "valid session",
			session: &QuizSession{
				UserID:         1,
				TotalProblems:  10,
				CorrectAnswers: 8,
				Score:          80.0,
				TimeSpent:      60.0,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			session: &QuizSession{
				TotalProblems:  10,
				CorrectAnswers: 8,
				Score:          80.0,
			},
			wantErr: true,
		},
		{
			name: "invalid correct_answers",
			session: &QuizSession{
				UserID:         1,
				TotalProblems:  10,
				CorrectAnswers: 15,
				Score:          80.0,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.session.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestHTTPEndpoints tests HTTP endpoints
func TestHTTPEndpoints(t *testing.T) {
	ti := setupIntegration(t)
	defer ti.db.Close()

	t.Run("GenerateProblem", func(t *testing.T) {
		body := map[string]interface{}{
			"problem_type": "addition",
			"difficulty":   "easy",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/math/problem", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ti.router.Routes().ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("GenerateProblem() status = %d, want 200", w.Code)
		}
	})

	t.Run("GetProblemTypes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/math/problem/types", nil)
		w := httptest.NewRecorder()
		ti.router.Routes().ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("GetProblemTypes() status = %d, want 200", w.Code)
		}
	})

	t.Run("CompleteSession", func(t *testing.T) {
		body := map[string]interface{}{
			"user_id":          1,
			"problem_type":     "addition",
			"difficulty":       "easy",
			"total_problems":   10,
			"correct_answers":  8,
			"time_spent":       60.0,
			"started_at":       time.Now().Add(-time.Minute),
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/math/session/complete", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ti.router.Routes().ServeHTTP(w, req)

		if w.Code != 201 {
			t.Errorf("CompleteSession() status = %d, want 201", w.Code)
		}
	})

	t.Run("GetUserStats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/1/math/stats", nil)
		w := httptest.NewRecorder()
		ti.router.Routes().ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("GetUserStats() status = %d, want 200", w.Code)
		}
	})
}

// BenchmarkGenerateProblem benchmarks problem generation
func BenchmarkGenerateProblem(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti.service.GenerateProblem(Addition, Easy)
	}
}

// BenchmarkCompleteSession benchmarks session completion
func BenchmarkCompleteSession(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := &QuizSession{
			UserID:                1,
			ProblemType:           Addition,
			Difficulty:            Easy,
			TotalProblems:         10,
			CorrectAnswers:        8,
			TimeSpent:             60.0,
			StartedAt:             time.Now().Add(-time.Minute),
			AverageTimePerProblem: 6.0,
		}
		ti.service.CompleteSession(ctx, session)
	}
}

// BenchmarkGetUserStats benchmarks stats retrieval
func BenchmarkGetUserStats(b *testing.B) {
	ti := setupBenchmark(b)
	defer ti.db.Close()

	ctx := context.Background()

	// Create initial data
	session := &QuizSession{
		UserID:                1,
		ProblemType:           Addition,
		Difficulty:            Easy,
		TotalProblems:         10,
		CorrectAnswers:        8,
		TimeSpent:             60.0,
		StartedAt:             time.Now().Add(-time.Minute),
		AverageTimePerProblem: 6.0,
	}
	ti.service.CompleteSession(ctx, session)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti.service.GetUserStats(ctx, 1)
	}
}

// BenchmarkCalculateScore benchmarks score calculation
func BenchmarkCalculateScore(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateScore(8, 10)
	}
}
