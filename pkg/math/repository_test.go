package math

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t testing.TB) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Create tables
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			mode TEXT NOT NULL,
			difficulty TEXT NOT NULL,
			total_questions INTEGER NOT NULL,
			correct_answers INTEGER NOT NULL,
			total_time REAL NOT NULL,
			average_time REAL NOT NULL,
			accuracy REAL NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE question_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			question TEXT NOT NULL,
			user_answer TEXT,
			correct_answer TEXT NOT NULL,
			is_correct BOOLEAN NOT NULL,
			time_taken REAL NOT NULL,
			fact_family TEXT,
			mode TEXT,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE mistakes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			question TEXT NOT NULL,
			correct_answer TEXT NOT NULL,
			user_answer TEXT,
			mode TEXT,
			fact_family TEXT,
			error_count INTEGER NOT NULL DEFAULT 1,
			last_error TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, question)
		);

		CREATE TABLE mastery (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			fact TEXT NOT NULL,
			mode TEXT,
			correct_streak INTEGER NOT NULL DEFAULT 0,
			total_attempts INTEGER NOT NULL DEFAULT 0,
			mastery_level REAL NOT NULL DEFAULT 0,
			last_practiced TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			average_response_time REAL NOT NULL DEFAULT 0,
			fastest_time REAL,
			slowest_time REAL,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, fact, mode)
		);

		CREATE TABLE learning_profile (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			learning_style TEXT,
			preferred_time_of_day TEXT,
			attention_span_seconds INTEGER DEFAULT 300,
			best_streak_time TEXT,
			weak_time_of_day TEXT,
			avg_session_length INTEGER NOT NULL DEFAULT 0,
			total_practice_time INTEGER NOT NULL DEFAULT 0,
			profile_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE performance_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			hour_of_day INTEGER NOT NULL,
			day_of_week INTEGER NOT NULL,
			average_accuracy REAL NOT NULL DEFAULT 0,
			average_speed REAL NOT NULL DEFAULT 0,
			session_count INTEGER NOT NULL DEFAULT 0,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, hour_of_day, day_of_week)
		);

		CREATE TABLE repetition_schedule (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			fact TEXT NOT NULL,
			mode TEXT,
			next_review TIMESTAMP NOT NULL,
			interval_days INTEGER NOT NULL DEFAULT 1,
			ease_factor REAL NOT NULL DEFAULT 2.5,
			review_count INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, fact, mode)
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestSaveAndGetUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	user := &User{Username: "student1"}
	err := repo.SaveUser(ctx, user)
	if err != nil {
		t.Errorf("SaveUser() error = %v", err)
	}

	retrieved, err := repo.GetUser(ctx, user.ID)
	if err != nil {
		t.Errorf("GetUser() error = %v", err)
	}

	if retrieved.Username != user.Username {
		t.Errorf("Username = %v, want %v", retrieved.Username, user.Username)
	}
}

func TestGetUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	user := &User{Username: "math_student"}
	repo.SaveUser(ctx, user)

	retrieved, err := repo.GetUserByUsername(ctx, "math_student")
	if err != nil {
		t.Errorf("GetUserByUsername() error = %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("User ID = %d, want %d", retrieved.ID, user.ID)
	}
}

func TestSaveAndGetResult(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user first
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Create result
	result := &MathResult{
		UserID:         user.ID,
		Mode:           "addition",
		Difficulty:     "easy",
		TotalQuestions: 10,
		CorrectAnswers: 9,
		TotalTime:      30,
		AverageTime:    3,
		Accuracy:       90,
	}

	err := repo.SaveResult(ctx, result)
	if err != nil {
		t.Errorf("SaveResult() error = %v", err)
	}

	retrieved, err := repo.GetResult(ctx, result.ID)
	if err != nil {
		t.Errorf("GetResult() error = %v", err)
	}

	if retrieved.Accuracy != result.Accuracy {
		t.Errorf("Accuracy = %v, want %v", retrieved.Accuracy, result.Accuracy)
	}
}

func TestSaveAndGetMastery(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user first
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Create mastery
	mastery := &Mastery{
		UserID:          user.ID,
		Fact:            "3+5",
		Mode:            "addition",
		CorrectStreak:   5,
		TotalAttempts:   10,
		MasteryLevel:    80,
		FastestTime:     1.5,
		SlowestTime:     3.0,
	}

	err := repo.SaveMastery(ctx, mastery)
	if err != nil {
		t.Errorf("SaveMastery() error = %v", err)
	}

	retrieved, err := repo.GetMastery(ctx, user.ID, "3+5", "addition")
	if err != nil {
		t.Errorf("GetMastery() error = %v", err)
	}

	if retrieved.CorrectStreak != mastery.CorrectStreak {
		t.Errorf("CorrectStreak = %d, want %d", retrieved.CorrectStreak, mastery.CorrectStreak)
	}
}

func TestSaveAndGetMistake(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user first
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save mistake
	mistake := &Mistake{
		UserID:        user.ID,
		Question:      "3+5",
		CorrectAnswer: "8",
		UserAnswer:    "7",
		Mode:          "addition",
		FactFamily:    "plus_five",
		ErrorCount:    1,
	}

	err := repo.SaveMistake(ctx, mistake)
	if err != nil {
		t.Errorf("SaveMistake() error = %v", err)
	}

	retrieved, err := repo.GetMistake(ctx, user.ID, "3+5")
	if err != nil {
		t.Errorf("GetMistake() error = %v", err)
	}

	if retrieved.UserAnswer != "7" {
		t.Errorf("UserAnswer = %v, want 7", retrieved.UserAnswer)
	}
}

func TestSaveAndGetQuestionHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user first
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save question history
	history := &QuestionHistory{
		UserID:        user.ID,
		Question:      "5+5",
		UserAnswer:    "10",
		CorrectAnswer: "10",
		IsCorrect:     true,
		TimeTaken:     2.5,
		FactFamily:    "doubles",
		Mode:          "addition",
	}

	err := repo.SaveQuestionHistory(ctx, history)
	if err != nil {
		t.Errorf("SaveQuestionHistory() error = %v", err)
	}

	retrieved, err := repo.GetQuestionHistory(ctx, history.ID)
	if err != nil {
		t.Errorf("GetQuestionHistory() error = %v", err)
	}

	if !retrieved.IsCorrect {
		t.Error("IsCorrect should be true")
	}
}

func TestSaveAndGetLearningProfile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user first
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save learning profile
	profile := &LearningProfile{
		UserID:              user.ID,
		LearningStyle:       "visual",
		PreferredTimeOfDay:  "morning",
		AttentionSpanSeconds: 600,
		BestStreakTime:      "09:00",
		WeakTimeOfDay:       "evening",
	}

	err := repo.SaveLearningProfile(ctx, profile)
	if err != nil {
		t.Errorf("SaveLearningProfile() error = %v", err)
	}

	retrieved, err := repo.GetLearningProfile(ctx, user.ID)
	if err != nil {
		t.Errorf("GetLearningProfile() error = %v", err)
	}

	if retrieved.LearningStyle != "visual" {
		t.Errorf("LearningStyle = %v, want visual", retrieved.LearningStyle)
	}
}

func TestGetDueRepetitions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save repetition schedules
	now := time.Now()

	schedule1 := &RepetitionSchedule{
		UserID:       user.ID,
		Fact:         "3+5",
		Mode:         "addition",
		NextReview:   now.Add(-24 * time.Hour), // Due yesterday
		IntervalDays: 1,
		EaseFactor:   2.5,
		ReviewCount:  0,
	}

	schedule2 := &RepetitionSchedule{
		UserID:       user.ID,
		Fact:         "4+6",
		Mode:         "addition",
		NextReview:   now.Add(24 * time.Hour), // Due tomorrow
		IntervalDays: 1,
		EaseFactor:   2.5,
		ReviewCount:  0,
	}

	repo.SaveRepetitionSchedule(ctx, schedule1)
	repo.SaveRepetitionSchedule(ctx, schedule2)

	dueSchedules, err := repo.GetDueRepetitions(ctx, user.ID, 10)
	if err != nil {
		t.Errorf("GetDueRepetitions() error = %v", err)
	}

	if len(dueSchedules) != 1 {
		t.Errorf("Expected 1 due schedule, got %d", len(dueSchedules))
	}

	if len(dueSchedules) > 0 && dueSchedules[0].Fact != "3+5" {
		t.Errorf("Fact = %v, want 3+5", dueSchedules[0].Fact)
	}
}

func TestGetUserStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save some results
	for i := 0; i < 3; i++ {
		result := &MathResult{
			UserID:         user.ID,
			Mode:           "addition",
			Difficulty:     "easy",
			TotalQuestions: 10,
			CorrectAnswers: 9,
			TotalTime:      30,
			AverageTime:    3,
			Accuracy:       90,
		}
		repo.SaveResult(ctx, result)
	}

	// Save some question history so GetUserStats has data
	for i := 0; i < 5; i++ {
		history := &QuestionHistory{
			UserID:        user.ID,
			Question:      "3+5",
			UserAnswer:    "8",
			CorrectAnswer: "8",
			IsCorrect:     true,
			TimeTaken:     2.5,
			FactFamily:    "plus_five",
			Mode:          "addition",
		}
		repo.SaveQuestionHistory(ctx, history)
	}

	stats, err := repo.GetUserStats(ctx, user.ID)
	if err != nil {
		t.Errorf("GetUserStats() error = %v", err)
	}

	if stats.TotalSessions != 3 {
		t.Errorf("TotalSessions = %d, want 3", stats.TotalSessions)
	}

	if stats.AverageAccuracy <= 0 {
		t.Error("AverageAccuracy should be calculated")
	}

	if stats.TotalQuestions != 5 {
		t.Errorf("TotalQuestions = %d, want 5", stats.TotalQuestions)
	}
}

func TestGetMistakeAnalysis(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save mistakes
	for i := 0; i < 3; i++ {
		mistake := &Mistake{
			UserID:        user.ID,
			Question:      "3+5",
			CorrectAnswer: "8",
			UserAnswer:    "7",
			Mode:          "addition",
			FactFamily:    "plus_five",
			ErrorCount:    1,
		}
		repo.SaveMistake(ctx, mistake)
	}

	analyses, err := repo.GetMistakeAnalysis(ctx, user.ID)
	if err != nil {
		t.Errorf("GetMistakeAnalysis() error = %v", err)
	}

	if len(analyses) == 0 {
		t.Error("Expected mistake analyses")
	}

	if analyses[0].ErrorCount != 3 {
		t.Errorf("ErrorCount = %d, want 3", analyses[0].ErrorCount)
	}
}

func TestGetLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create users and results
	for userNum := 1; userNum <= 3; userNum++ {
		username := string(rune(userNum) + '0')
		user := &User{Username: username}
		repo.SaveUser(ctx, user)

		// Each user gets different accuracy
		accuracy := float64(80 + userNum*5)

		result := &MathResult{
			UserID:         user.ID,
			Mode:           "addition",
			Difficulty:     "easy",
			TotalQuestions: 10,
			CorrectAnswers: int(accuracy) / 10,
			TotalTime:      30,
			AverageTime:    3,
			Accuracy:       accuracy,
		}
		repo.SaveResult(ctx, result)
	}

	leaderboard, err := repo.GetLeaderboard(ctx, "sessions", 10)
	if err != nil {
		t.Errorf("GetLeaderboard() error = %v", err)
	}

	if len(leaderboard) == 0 {
		t.Error("Expected leaderboard entries")
	}
}

func TestGetBestPerformanceTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save performance patterns
	pattern := &PerformancePattern{
		UserID:          user.ID,
		HourOfDay:       10, // Morning
		DayOfWeek:       1,
		AverageAccuracy: 95,
		AverageSpeed:    2.0,
		SessionCount:    5,
	}

	err := repo.SavePerformancePattern(ctx, pattern)
	if err != nil {
		t.Errorf("SavePerformancePattern() error = %v", err)
	}

	bestTime, accuracy, err := repo.GetBestPerformanceTime(ctx, user.ID)
	if err != nil {
		t.Errorf("GetBestPerformanceTime() error = %v", err)
	}

	if bestTime != "morning" {
		t.Errorf("BestTime = %v, want morning", bestTime)
	}

	if accuracy != 95 {
		t.Errorf("Accuracy = %v, want 95", accuracy)
	}
}

func TestDeleteResult(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user and result
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	result := &MathResult{
		UserID:         user.ID,
		Mode:           "addition",
		Difficulty:     "easy",
		TotalQuestions: 10,
		CorrectAnswers: 9,
		TotalTime:      30,
		AverageTime:    3,
		Accuracy:       90,
	}

	repo.SaveResult(ctx, result)

	// Delete it
	err := repo.DeleteResult(ctx, result.ID)
	if err != nil {
		t.Errorf("DeleteResult() error = %v", err)
	}

	// Verify it's deleted
	_, err = repo.GetResult(ctx, result.ID)
	if err == nil {
		t.Error("Result should be deleted")
	}
}

func TestGetWeakFactFamilies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)
	ctx := context.Background()

	// Create user
	user := &User{Username: "student1"}
	repo.SaveUser(ctx, user)

	// Save mistakes in different families
	families := []string{"doubles", "plus_ten", "make_ten"}
	for _, family := range families {
		for i := 0; i < 3; i++ {
			mistake := &Mistake{
				UserID:        user.ID,
				Question:      family + "_q" + string(rune(i)+'0'),
				CorrectAnswer: "8",
				UserAnswer:    "7",
				Mode:          "addition",
				FactFamily:    family,
				ErrorCount:    1,
			}
			repo.SaveMistake(ctx, mistake)
		}
	}

	weakFamilies, err := repo.GetWeakFactFamilies(ctx, user.ID, 2)
	if err != nil {
		t.Errorf("GetWeakFactFamilies() error = %v", err)
	}

	if len(weakFamilies) == 0 {
		t.Error("Expected weak fact families")
	}
}
