package piano

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
)

// Repository handles database operations for piano app
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new piano repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveSong saves a piano song to the database with MIDI blob
func (r *Repository) SaveSong(ctx context.Context, song *Song) (uint, error) {
	if song == nil {
		return 0, errors.New("song cannot be nil")
	}

	if err := song.Validate(); err != nil {
		return 0, fmt.Errorf("invalid song: %w", err)
	}

	stmt := `INSERT INTO songs (title, composer, description, midi_file, difficulty, duration, bpm, time_signature, key_signature, total_notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.ExecContext(ctx, stmt,
		song.Title, song.Composer, song.Description, song.MIDIFile, song.Difficulty,
		song.Duration, song.BPM, song.TimeSignature, song.KeySignature, song.TotalNotes)

	if err != nil {
		return 0, fmt.Errorf("failed to save song: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return uint(id), nil
}

// GetSongByID retrieves a song by ID with MIDI data
func (r *Repository) GetSongByID(ctx context.Context, songID uint) (*Song, error) {
	if songID == 0 {
		return nil, errors.New("song_id is required")
	}

	var song Song
	stmt := `SELECT id, title, composer, description, midi_file, difficulty, duration, bpm, time_signature, key_signature, total_notes, created_at, updated_at FROM songs WHERE id = ?`

	err := r.db.QueryRowContext(ctx, stmt, songID).Scan(
		&song.ID, &song.Title, &song.Composer, &song.Description, &song.MIDIFile, &song.Difficulty,
		&song.Duration, &song.BPM, &song.TimeSignature, &song.KeySignature, &song.TotalNotes, &song.CreatedAt, &song.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("song not found")
		}
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	return &song, nil
}

// GetSongs retrieves songs with optional filtering
func (r *Repository) GetSongs(ctx context.Context, difficulty string, limit, offset int) ([]Song, error) {
	if limit <= 0 || limit > 1000 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	stmt := `SELECT id, title, composer, description, midi_file, difficulty, duration, bpm, time_signature, key_signature, total_notes, created_at, updated_at FROM songs`

	if difficulty != "" {
		stmt += ` WHERE difficulty = ?`
	}

	stmt += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	var rows *sql.Rows
	var err error

	if difficulty != "" {
		rows, err = r.db.QueryContext(ctx, stmt, difficulty, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, stmt, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get songs: %w", err)
	}
	defer rows.Close()

	var songs []Song
	for rows.Next() {
		var song Song
		if err := rows.Scan(&song.ID, &song.Title, &song.Composer, &song.Description, &song.MIDIFile, &song.Difficulty,
			&song.Duration, &song.BPM, &song.TimeSignature, &song.KeySignature, &song.TotalNotes, &song.CreatedAt, &song.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan song: %w", err)
		}
		songs = append(songs, song)
	}

	return songs, rows.Err()
}

// SaveLesson saves a piano lesson to the database
func (r *Repository) SaveLesson(ctx context.Context, lesson *PianoLesson) (uint, error) {
	if lesson == nil {
		return 0, errors.New("lesson cannot be nil")
	}

	if err := lesson.Validate(); err != nil {
		return 0, fmt.Errorf("invalid lesson: %w", err)
	}

	stmt := `INSERT INTO piano_lessons (user_id, song_id, start_time, end_time, duration, notes_correct, notes_total, accuracy, tempo_accuracy, score, completed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.ExecContext(ctx, stmt,
		lesson.UserID, lesson.SongID, lesson.StartTime, lesson.EndTime, lesson.Duration,
		lesson.NotesCorrect, lesson.NotesTotal, lesson.Accuracy, lesson.TempoAccuracy, lesson.Score, lesson.Completed)

	if err != nil {
		return 0, fmt.Errorf("failed to save lesson: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return uint(id), nil
}

// SavePracticeSession saves a practice session with MIDI recording blob
func (r *Repository) SavePracticeSession(ctx context.Context, session *PracticeSession) (uint, error) {
	if session == nil {
		return 0, errors.New("session cannot be nil")
	}

	if err := session.Validate(); err != nil {
		return 0, fmt.Errorf("invalid session: %w", err)
	}

	stmt := `INSERT INTO practice_sessions (user_id, song_id, lesson_id, recording_midi, duration, notes_hit, notes_total, tempo_average, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.ExecContext(ctx, stmt,
		session.UserID, session.SongID, session.LessonID, session.RecordingMIDI, session.Duration,
		session.NotesHit, session.NotesTotal, session.TempoAverage)

	if err != nil {
		return 0, fmt.Errorf("failed to save practice session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return uint(id), nil
}

// GetMIDIRecording retrieves a MIDI recording blob from a practice session
func (r *Repository) GetMIDIRecording(ctx context.Context, sessionID uint) ([]byte, error) {
	if sessionID == 0 {
		return nil, errors.New("session_id is required")
	}

	var midiData []byte
	stmt := `SELECT recording_midi FROM practice_sessions WHERE id = ?`

	err := r.db.QueryRowContext(ctx, stmt, sessionID).Scan(&midiData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("failed to get MIDI recording: %w", err)
	}

	return midiData, nil
}

// GetUserLessons retrieves all piano lessons for a user
func (r *Repository) GetUserLessons(ctx context.Context, userID uint, limit, offset int) ([]PianoLesson, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	if limit <= 0 || limit > 1000 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	stmt := `SELECT id, user_id, song_id, start_time, end_time, duration, notes_correct, notes_total, accuracy, tempo_accuracy, score, completed, created_at, updated_at
		FROM piano_lessons WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, stmt, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user lessons: %w", err)
	}
	defer rows.Close()

	var lessons []PianoLesson
	for rows.Next() {
		var lesson PianoLesson
		if err := rows.Scan(&lesson.ID, &lesson.UserID, &lesson.SongID, &lesson.StartTime, &lesson.EndTime,
			&lesson.Duration, &lesson.NotesCorrect, &lesson.NotesTotal, &lesson.Accuracy, &lesson.TempoAccuracy,
			&lesson.Score, &lesson.Completed, &lesson.CreatedAt, &lesson.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lesson: %w", err)
		}
		lessons = append(lessons, lesson)
	}

	return lessons, rows.Err()
}

// GetUserProgress calculates aggregated progress for a user
func (r *Repository) GetUserProgress(ctx context.Context, userID uint) (*UserProgress, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	progress := &UserProgress{
		UserID: userID,
	}

	// Count completed practice sessions
	var sessionCount int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM practice_sessions WHERE user_id = ?`, userID).Scan(&sessionCount)
	progress.TotalLessonsCompleted = sessionCount

	// Get sessions for aggregation
	stmt := `SELECT duration, notes_hit, notes_total, tempo_average FROM practice_sessions WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user progress: %w", err)
	}
	defer rows.Close()

	var totalScore, maxScore, maxTempo, totalDuration float64
	sessionCount = 0

	for rows.Next() {
		var duration, tempoAverage float64
		var notesHit, notesTotal int
		if err := rows.Scan(&duration, &notesHit, &notesTotal, &tempoAverage); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		sessionCount++
		totalDuration += duration

		// Calculate accuracy and tempo accuracy for this session
		accuracy := CalculateAccuracy(notesHit, notesTotal)
		tempoAccuracy := CalculateTempoAccuracy(tempoAverage, 120.0) // Default target tempo

		// Calculate composite score
		score := CalculateCompositeScore(accuracy, tempoAccuracy, 0) // No theory score in practice session

		totalScore += score
		if score > maxScore {
			maxScore = score
		}
		if tempoAccuracy > maxTempo {
			maxTempo = tempoAccuracy
		}
	}

	if sessionCount > 0 {
		progress.AverageScore = totalScore / float64(sessionCount)
		progress.TotalPracticedMinutes = totalDuration / 60.0
	}

	progress.BestScore = maxScore
	progress.FastestTempo = maxTempo
	progress.CurrentLevel = EstimatePianoLevel(progress.AverageScore)

	// Get last practiced date
	var lastTime *sql.NullTime
	r.db.QueryRowContext(ctx, `SELECT MAX(created_at) FROM practice_sessions WHERE user_id = ?`, userID).Scan(&lastTime)
	if lastTime != nil && lastTime.Valid {
		progress.LastPracticedDate = &lastTime.Time
	}

	return progress, nil
}

// SaveMusicTheoryQuiz saves a music theory quiz
func (r *Repository) SaveMusicTheoryQuiz(ctx context.Context, quiz *MusicTheoryQuiz) (uint, error) {
	if quiz == nil {
		return 0, errors.New("quiz cannot be nil")
	}

	stmt := `INSERT INTO music_theory_quizzes (user_id, lesson_id, topic, questions, answers, score, difficulty, completed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.ExecContext(ctx, stmt,
		quiz.UserID, quiz.LessonID, quiz.Topic, quiz.Questions, quiz.Answers, quiz.Score, quiz.Difficulty, quiz.Completed)

	if err != nil {
		return 0, fmt.Errorf("failed to save music theory quiz: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return uint(id), nil
}

// GetTheoryQuestionsByDifficulty retrieves theory questions by difficulty
func (r *Repository) GetTheoryQuestionsByDifficulty(ctx context.Context, difficulty string, limit int) ([]MusicQuestion, error) {
	if limit <= 0 || limit > 1000 {
		limit = 10
	}

	// In a real implementation, this would query from a questions table
	// For now, return predefined questions based on difficulty
	questions := []MusicQuestion{
		{
			ID:            1,
			Question:      "What is the C major scale?",
			Options:       []string{"C D E F G A B", "C D E F# G A B", "C D Eb F G Ab Bb"},
			CorrectAnswer: "C D E F G A B",
			Explanation:   "The C major scale consists of all white keys on the piano",
		},
		{
			ID:            2,
			Question:      "What is a perfect fifth interval?",
			Options:       []string{"5 semitones", "7 semitones", "12 semitones"},
			CorrectAnswer: "7 semitones",
			Explanation:   "A perfect fifth is 7 semitones apart",
		},
	}

	if len(questions) > limit {
		questions = questions[:limit]
	}

	return questions, nil
}

// GetLessonByID retrieves a specific piano lesson
func (r *Repository) GetLessonByID(ctx context.Context, lessonID uint) (*PianoLesson, error) {
	if lessonID == 0 {
		return nil, errors.New("lesson_id is required")
	}

	var lesson PianoLesson
	stmt := `SELECT id, user_id, song_id, start_time, end_time, duration, notes_correct, notes_total, accuracy, tempo_accuracy, score, completed, created_at, updated_at FROM piano_lessons WHERE id = ?`

	err := r.db.QueryRowContext(ctx, stmt, lessonID).Scan(
		&lesson.ID, &lesson.UserID, &lesson.SongID, &lesson.StartTime, &lesson.EndTime,
		&lesson.Duration, &lesson.NotesCorrect, &lesson.NotesTotal, &lesson.Accuracy, &lesson.TempoAccuracy,
		&lesson.Score, &lesson.Completed, &lesson.CreatedAt, &lesson.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("lesson not found")
		}
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	return &lesson, nil
}

// GetLeaderboard retrieves top pianists by average score
func (r *Repository) GetLeaderboard(ctx context.Context, limit int) ([]UserProgress, error) {
	if limit <= 0 || limit > 1000 {
		limit = 10
	}

	// Get distinct users
	stmt := `SELECT DISTINCT user_id FROM piano_lessons WHERE completed = 1 ORDER BY user_id`

	rows, err := r.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var leaderboard []UserProgress
	for rows.Next() {
		var userID uint
		if err := rows.Scan(&userID); err != nil {
			continue
		}

		progress, err := r.GetUserProgress(ctx, userID)
		if err == nil && progress != nil {
			leaderboard = append(leaderboard, *progress)
		}
	}

	// Sort by best score descending
	for i := 0; i < len(leaderboard)-1; i++ {
		for j := i + 1; j < len(leaderboard); j++ {
			if leaderboard[j].BestScore > leaderboard[i].BestScore {
				leaderboard[i], leaderboard[j] = leaderboard[j], leaderboard[i]
			}
		}
	}

	// Limit results
	if len(leaderboard) > limit {
		leaderboard = leaderboard[:limit]
	}

	return leaderboard, nil
}

// StoreMIDIAsHex stores MIDI data as hex-encoded string (for backup/export)
func (r *Repository) StoreMIDIAsHex(ctx context.Context, midiData []byte) string {
	return hex.EncodeToString(midiData)
}

// RetrieveMIDIFromHex retrieves MIDI data from hex-encoded string
func (r *Repository) RetrieveMIDIFromHex(ctx context.Context, hexData string) ([]byte, error) {
	return hex.DecodeString(hexData)
}
