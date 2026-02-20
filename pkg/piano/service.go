package piano

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// Service provides business logic for piano operations
type Service struct {
	repo *Repository
}

// NewService creates a new piano service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CalculateAccuracy calculates note accuracy as a percentage
// Compares correct notes against total notes played
func (s *Service) CalculateAccuracy(notesCorrect, notesTotal int) float64 {
	if notesTotal <= 0 {
		return 0.0
	}

	if notesCorrect < 0 {
		notesCorrect = 0
	}

	if notesCorrect > notesTotal {
		notesCorrect = notesTotal
	}

	// Calculate accuracy percentage
	accuracy := (float64(notesCorrect) / float64(notesTotal)) * 100.0

	// Clamp to 0-100%
	if accuracy < 0 {
		accuracy = 0
	} else if accuracy > 100 {
		accuracy = 100
	}

	// Round to 2 decimal places
	return math.Round(accuracy*100) / 100
}

// CalculateTempo calculates tempo accuracy based on recorded vs target BPM
// Returns a score from 0-100 where 100 is perfect match
func (s *Service) CalculateTempo(recordedBPM, targetBPM float64) float64 {
	if targetBPM <= 0 {
		return 0.0
	}

	if recordedBPM <= 0 {
		return 0.0
	}

	// Calculate BPM difference percentage
	difference := math.Abs(recordedBPM - targetBPM)
	percentDifference := (difference / targetBPM) * 100.0

	// Convert to accuracy score (100% difference = 0 score)
	tempoAccuracy := 100.0 - percentDifference

	// Clamp to 0-100%
	if tempoAccuracy < 0 {
		tempoAccuracy = 0
	} else if tempoAccuracy > 100 {
		tempoAccuracy = 100
	}

	// Round to 2 decimal places
	return math.Round(tempoAccuracy*100) / 100
}

// CalculateCompositeScore combines accuracy, tempo, and theory scores
// Weights: Accuracy 50%, Tempo 30%, Theory 20%
func (s *Service) CalculateCompositeScore(accuracy, tempo, theory float64) float64 {
	composite := (accuracy * 0.5) + (tempo * 0.3) + (theory * 0.2)

	// Clamp to 0-100%
	if composite < 0 {
		composite = 0
	} else if composite > 100 {
		composite = 100
	}

	// Round to 2 decimal places
	return math.Round(composite*100) / 100
}

// ProcessLesson processes a completed piano lesson
func (s *Service) ProcessLesson(ctx context.Context, userID uint, songID uint, recordedBPM float64, duration float64, notesCorrect, notesTotal int) (*PracticeSession, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if songID == 0 {
		return nil, errors.New("song_id is required")
	}

	if duration <= 0 {
		return nil, errors.New("duration must be positive")
	}

	if notesCorrect < 0 || notesTotal <= 0 {
		return nil, errors.New("invalid notes: total must be positive, correct must be non-negative")
	}

	// Retrieve song to get target BPM
	song, err := s.repo.GetSongByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	if song == nil {
		return nil, errors.New("song not found")
	}

	session := &PracticeSession{
		UserID:        userID,
		SongID:        songID,
		Duration:      duration,
		NotesHit:      notesCorrect,
		NotesTotal:    notesTotal,
		TempoAverage:  recordedBPM,
		RecordingMIDI: song.MIDIFile, // Use the song's MIDI as the recording
	}

	// Validate the session
	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Save to repository
	id, err := s.repo.SavePracticeSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to save practice session: %w", err)
	}

	session.ID = id
	return session, nil
}

// GenerateLesson creates a practice lesson recommendation based on user skill level
func (s *Service) GenerateLesson(ctx context.Context, userID uint, difficulty string) (*Song, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if difficulty == "" {
		difficulty = "intermediate"
	}

	// Validate difficulty
	validDifficulties := map[string]bool{"beginner": true, "intermediate": true, "advanced": true, "expert": true}
	if !validDifficulties[difficulty] {
		return nil, errors.New("invalid difficulty level")
	}

	// Get a random song of the requested difficulty from the repository
	songs, err := s.repo.GetSongs(ctx, difficulty, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get songs: %w", err)
	}

	if len(songs) == 0 {
		return nil, errors.New("no songs available for difficulty level")
	}

	return &songs[0], nil
}

// GetProgressionPath determines the optimal learning path for a user
func (s *Service) GetProgressionPath(ctx context.Context, userID uint) ([]string, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	// Get user progress
	progress, err := s.repo.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user progress: %w", err)
	}

	// Determine path based on estimated level
	path := []string{}
	switch progress.EstimatedLevel {
	case "beginner":
		path = []string{"beginner", "intermediate"}
	case "intermediate":
		path = []string{"intermediate", "advanced"}
	case "advanced":
		path = []string{"advanced", "expert"}
	case "expert":
		path = []string{"expert"}
	default:
		path = []string{"beginner", "intermediate", "advanced", "expert"}
	}

	return path, nil
}

// EvaluatePerformance provides a comprehensive performance evaluation
func (s *Service) EvaluatePerformance(ctx context.Context, userID uint) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	// Get user progress
	progress, err := s.repo.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user progress: %w", err)
	}

	// Get user lessons to calculate improvement trend
	lessons, err := s.repo.GetUserLessons(ctx, userID, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user lessons: %w", err)
	}

	// Calculate trend
	trend := "neutral"
	change := 0.0

	if len(lessons) >= 2 {
		firstScore := lessons[len(lessons)-1].Score
		lastScore := lessons[0].Score

		if firstScore > 0 {
			change = ((lastScore - firstScore) / firstScore) * 100
			if change > 5 {
				trend = "improving"
			} else if change < -5 {
				trend = "declining"
			}
		}
	}

	evaluation := map[string]interface{}{
		"user_id":             userID,
		"total_lessons":       progress.TotalLessonsCompleted,
		"average_score":       progress.AverageScore,
		"best_score":          progress.BestScore,
		"total_practice_time": progress.TotalPracticedMinutes,
		"current_level":       progress.EstimatedLevel,
		"fastest_tempo":       progress.FastestTempo,
		"trend":               trend,
		"change_percentage":   math.Round(change*100) / 100,
		"last_practiced":      progress.LastPracticedDate,
	}

	return evaluation, nil
}

// GenerateMusicTheoryQuiz creates a theory quiz based on difficulty
func (s *Service) GenerateMusicTheoryQuiz(ctx context.Context, difficulty string, questionCount int) ([]MusicQuestion, error) {
	if difficulty == "" {
		difficulty = "intermediate"
	}

	if questionCount <= 0 {
		questionCount = 5
	}

	if questionCount > 20 {
		questionCount = 20
	}

	// Validate difficulty
	validDifficulties := map[string]bool{"beginner": true, "intermediate": true, "advanced": true, "expert": true}
	if !validDifficulties[difficulty] {
		return nil, errors.New("invalid difficulty level")
	}

	// Get theory questions from repository
	questions, err := s.repo.GetTheoryQuestionsByDifficulty(ctx, difficulty, questionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get theory questions: %w", err)
	}

	if len(questions) == 0 {
		return nil, errors.New("no theory questions available for difficulty")
	}

	// Limit to requested count
	if len(questions) > questionCount {
		questions = questions[:questionCount]
	}

	return questions, nil
}

// AnalyzeMusicTheory evaluates music theory quiz answers for a session
func (s *Service) AnalyzeMusicTheory(ctx context.Context, sessionID uint) (map[string]interface{}, error) {
	if sessionID == 0 {
		return nil, errors.New("session_id is required")
	}

	// For now, return a basic analysis structure
	// In a full implementation, this would analyze actual quiz responses for the session
	return map[string]interface{}{
		"session_id":      sessionID,
		"total_questions": 0,
		"correct_answers": 0,
		"score":           0.0,
	}, nil
}

// GetUserMetrics returns comprehensive user metrics
func (s *Service) GetUserMetrics(ctx context.Context, userID uint) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	// Get user progress
	progress, err := s.repo.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user progress: %w", err)
	}

	// Get recent lessons for trend analysis
	lessons, err := s.repo.GetUserLessons(ctx, userID, 5, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user lessons: %w", err)
	}

	// Calculate 7-day metrics
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	sessionsLastWeek := 0
	practiceMinutesLastWeek := 0.0

	for _, lesson := range lessons {
		if lesson.StartTime.After(sevenDaysAgo) {
			sessionsLastWeek++
			practiceMinutesLastWeek += lesson.Duration
		}
	}

	metrics := map[string]interface{}{
		"user_id":                userID,
		"total_lessons":          progress.TotalLessonsCompleted,
		"average_score":          progress.AverageScore,
		"best_score":             progress.BestScore,
		"current_level":          progress.EstimatedLevel,
		"total_practice_minutes": progress.TotalPracticedMinutes,
		"sessions_last_week":     sessionsLastWeek,
		"practice_minutes_week":  math.Round(practiceMinutesLastWeek*100) / 100,
		"fastest_tempo":          progress.FastestTempo,
		"last_practiced":         progress.LastPracticedDate,
	}

	return metrics, nil
}
