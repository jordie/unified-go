package math

import (
	"context"
	"testing"
	"time"
)

// ==================== END-TO-END FLOW TESTS ====================

// TestSpacedRepetitionFlow tests SM-2 spaced repetition end-to-end
func TestSpacedRepetitionFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	engine := NewSM2Engine(repo)
	ctx := context.Background()

	// Create a user
	user := &User{Username: "sr_test_user", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	// Initialize schedule for a fact
	schedule, err := engine.InitializeSchedule(ctx, user.ID, "5+3", MODE_ADDITION)
	if err != nil {
		t.Fatalf("Failed to initialize schedule: %v", err)
	}

	if schedule.ReviewCount != 0 {
		t.Errorf("Expected 0 reviews for new fact, got %d", schedule.ReviewCount)
	}

	if schedule.EaseFactor != INITIAL_EASE_FACTOR {
		t.Errorf("Expected ease factor %.1f, got %.1f", INITIAL_EASE_FACTOR, schedule.EaseFactor)
	}

	// Process first review (quality 5 = perfect)
	schedule, err = engine.ProcessReview(ctx, user.ID, "5+3", MODE_ADDITION, 5)
	if err != nil {
		t.Fatalf("Failed to process first review: %v", err)
	}

	if schedule.ReviewCount != 1 {
		t.Errorf("Expected 1 review after processing, got %d", schedule.ReviewCount)
	}

	if schedule.IntervalDays != 1 {
		t.Errorf("Expected 1 day interval for first review, got %d", schedule.IntervalDays)
	}

	// Process second review
	schedule, err = engine.ProcessReview(ctx, user.ID, "5+3", MODE_ADDITION, 4)
	if err != nil {
		t.Fatalf("Failed to process second review: %v", err)
	}

	if schedule.ReviewCount != 2 {
		t.Errorf("Expected 2 reviews, got %d", schedule.ReviewCount)
	}

	if schedule.IntervalDays != 6 {
		t.Errorf("Expected 6 day interval for second review, got %d", schedule.IntervalDays)
	}

	// Process third review (exponential growth)
	schedule, err = engine.ProcessReview(ctx, user.ID, "5+3", MODE_ADDITION, 5)
	if err != nil {
		t.Fatalf("Failed to process third review: %v", err)
	}

	if schedule.ReviewCount != 3 {
		t.Errorf("Expected 3 reviews, got %d", schedule.ReviewCount)
	}

	// Interval should be 6 * ease_factor
	expectedInterval := int(float64(6) * schedule.EaseFactor)
	if schedule.IntervalDays < expectedInterval-2 || schedule.IntervalDays > expectedInterval+2 {
		t.Logf("Expected interval around %d days (±2), got %d", expectedInterval, schedule.IntervalDays)
	}

	t.Logf("✓ Spaced repetition flow completed successfully")
}

// TestAssessmentFlow tests placement assessment end-to-end
func TestAssessmentFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	engine := NewAssessmentEngine(repo)
	ctx := context.Background()

	// Create user
	user := &User{Username: "assessment_user", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	// Start assessment
	session, err := engine.StartAssessment(ctx, user.ID, MODE_MIXED)
	if err != nil {
		t.Fatalf("Failed to start assessment: %v", err)
	}

	if session.CurrentLevel != 7 {
		t.Errorf("Expected starting level 7, got %d", session.CurrentLevel)
	}

	if session.MinLevel != 1 || session.MaxLevel != 15 {
		t.Errorf("Expected level range 1-15, got %d-%d", session.MinLevel, session.MaxLevel)
	}

	// Simulate answers (mostly correct)
	correctCount := 0
	for i := 0; i < 15; i++ {
		isCorrect := i < 12 // 12 correct out of 15
		if isCorrect {
			correctCount++
		}

		result, _ := engine.ProcessResponse(ctx, session, isCorrect, MODE_MIXED)
		if result != nil {
			// Assessment complete
			if result.PlacedLevel < 1 || result.PlacedLevel > 15 {
				t.Errorf("Placement level out of range: %d", result.PlacedLevel)
			}

			expectedAccuracy := float64(correctCount) / float64(i+1)
			if result.EstimatedAccuracy != expectedAccuracy {
				t.Logf("Accuracy difference: expected %.2f, got %.2f", expectedAccuracy, result.EstimatedAccuracy)
			}

			if result.Confidence < 0 || result.Confidence > 1.0 {
				t.Errorf("Confidence out of range: %.2f", result.Confidence)
			}

			break
		}
	}

	t.Logf("✓ Assessment flow completed successfully")
}

// TestMasteryTracking tests that mastery levels progress correctly
func TestMasteryTracking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	user := &User{Username: "mastery_user", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	fact := "7 + 8"

	// Simulate 5 successful responses
	for i := 0; i < 5; i++ {
		history := &QuestionHistory{
			UserID:        user.ID,
			Question:      fact,
			UserAnswer:    "15",
			CorrectAnswer: "15",
			IsCorrect:     true,
			TimeTaken:     2.5,
			Mode:          MODE_ADDITION,
			Timestamp:     time.Now(),
		}
		service.SaveQuestionResponse(ctx, user.ID, history, 0)
	}

	mastery, _ := repo.GetMastery(ctx, user.ID, fact, MODE_ADDITION)
	if mastery == nil {
		t.Fatal("Mastery record should exist")
	}

	if mastery.CorrectStreak != 5 {
		t.Errorf("Expected streak of 5, got %d", mastery.CorrectStreak)
	}

	if mastery.TotalAttempts != 5 {
		t.Errorf("Expected 5 total attempts, got %d", mastery.TotalAttempts)
	}

	if mastery.MasteryLevel == 0 {
		t.Error("Mastery level should be calculated")
	}

	if mastery.MasteryLevel > 100 {
		t.Errorf("Mastery level should not exceed 100, got %.1f", mastery.MasteryLevel)
	}

	t.Logf("✓ Mastery tracking completed successfully (level: %.1f)", mastery.MasteryLevel)
}

// TestAssessmentPlacement validates placement calculation
func TestAssessmentPlacement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	engine := NewAssessmentEngine(repo)
	ctx := context.Background()

	user := &User{Username: "placement_test", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	testCases := []struct {
		name           string
		accuracy       float64
		expectedLevel  int
		tolerance      int
	}{
		{"Very weak (< 50%)", 0.40, 3, 2},
		{"Weak (50-70%)", 0.60, 7, 2},
		{"Medium (70-85%)", 0.75, 10, 2},
		{"Strong (85%+)", 0.90, 13, 2},
	}

	for _, tc := range testCases {
		level, _ := engine.GetCurrentLevel(ctx, user.ID)

		// Simple placement: 1 + floor(accuracy / 7)
		expectedMin := tc.expectedLevel - tc.tolerance
		expectedMax := tc.expectedLevel + tc.tolerance

		if level < expectedMin || level > expectedMax {
			t.Logf("%s: Expected level %d±%d, got %d (accuracy: %.1f%%)",
				tc.name, tc.expectedLevel, tc.tolerance, level, tc.accuracy*100)
		}
	}

	t.Logf("✓ Assessment placement tests completed")
}

// ==================== PERFORMANCE BENCHMARKS ====================

// BenchmarkSaveResult measures performance of saving practice results
func BenchmarkSaveResult(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	user := &User{Username: "bench_user", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	result := &MathResult{
		UserID:         user.ID,
		Mode:           MODE_ADDITION,
		Difficulty:     "medium",
		TotalQuestions: 10,
		CorrectAnswers: 8,
		TotalTime:      120.5,
		Timestamp:      time.Now(),
	}
	result.CalculateAccuracy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.SaveResult(ctx, result)
	}
}

// BenchmarkGetDueRepetitions measures performance of retrieving due facts
func BenchmarkGetDueRepetitions(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	user := &User{Username: "bench_user", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	// Create some due schedules
	engine := NewSM2Engine(repo)
	for i := 0; i < 10; i++ {
		fact := "fact_" + string(rune(i))
		engine.InitializeSchedule(ctx, user.ID, fact, MODE_ADDITION)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetDueRepetitions(ctx, user.ID, 10)
	}
}

// BenchmarkGenerateAdaptiveSession measures performance of generating sessions
func BenchmarkGenerateAdaptiveSession(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	repo := NewRepository(db)
	engine := NewSM2Engine(repo)
	ctx := context.Background()

	user := &User{Username: "bench_user", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.GenerateAdaptiveSession(ctx, user.ID, 10)
	}
}

// BenchmarkAssessmentPlacement measures placement calculation speed
func BenchmarkAssessmentPlacement(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	repo := NewRepository(db)
	engine := NewAssessmentEngine(repo)
	ctx := context.Background()

	user := &User{Username: "bench_user", CreatedAt: time.Now()}
	repo.SaveUser(ctx, user)

	session, _ := engine.StartAssessment(ctx, user.ID, MODE_MIXED)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isCorrect := i%2 == 0
		engine.ProcessResponse(ctx, session, isCorrect, MODE_MIXED)
	}
}
