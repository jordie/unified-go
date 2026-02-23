package reading

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ReadingApp manages the reading application
type ReadingApp struct {
	db *sql.DB
}

// NewReadingApp creates a new reading app instance
func NewReadingApp(db *sql.DB) *ReadingApp {
	return &ReadingApp{db: db}
}

// ============================================================================
// WORD MASTERY SYSTEM
// ============================================================================

// Word represents a vocabulary word at a specific level
type Word struct {
	ID            int64     `json:"id"`
	Word          string    `json:"word"`
	Level         int       `json:"level"`
	Phonetic      string    `json:"phonetic"`
	Definition    string    `json:"definition"`
	ExampleSentence string  `json:"example_sentence"`
	Category      string    `json:"category"`
	Difficulty    int       `json:"difficulty"`
	CreatedAt     time.Time `json:"created_at"`
}

// WordMastery tracks user's mastery of each word
type WordMastery struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Word           string    `json:"word"`
	CorrectCount   int       `json:"correct_count"`
	ErrorCount     int       `json:"error_count"`
	TotalAttempts  int       `json:"total_attempts"`
	Accuracy       float64   `json:"accuracy"`
	LastReviewedAt time.Time `json:"last_reviewed_at"`
	MasteredAt     *time.Time `json:"mastered_at,omitempty"`
}

// PracticeSession represents a single reading practice session
type PracticeSession struct {
	ID                 string
	UserID             int64
	Level              int
	TotalWords         int
	WordsCompleted     int
	CorrectAnswers     int
	Accuracy           float64
	TotalTime          int // seconds
	WordsAttempted     []string // Track words in THIS session to prevent repeats
	UsedWordIDs        map[int64]bool // Track word IDs to prevent duplicates
	StartedAt          time.Time
	CompletedAt        *time.Time
}

// ============================================================================
// WORD SELECTION (NO REPETITION)
// ============================================================================

// GetWordsForPractice returns unique words for practice, NEVER repeating within a session
func (app *ReadingApp) GetWordsForPractice(userID int64, count int, sessionUsedWords []string) ([]Word, error) {
	// Create a map of already-used words in this session (NO REPEATS!)
	usedWordsMap := make(map[string]bool)
	for _, word := range sessionUsedWords {
		usedWordsMap[strings.ToLower(word)] = true
	}

	// Get user's current level
	var currentLevel int
	err := app.db.QueryRow(
		`SELECT COALESCE(current_level, 1) FROM user_progress WHERE user_id = ?`,
		userID,
	).Scan(&currentLevel)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if currentLevel == 0 {
		currentLevel = 1
	}

	// Get words from current level, EXCLUDING:
	// 1. Already mastered words (3+ correct)
	// 2. Words already used in this session (CRITICAL!)
	rows, err := app.db.Query(`
		SELECT id, word, level, phonetic, definition, example_sentence, category, difficulty, created_at
		FROM words
		WHERE level = ?
		AND word NOT IN (
			SELECT word FROM word_mastery
			WHERE user_id = ? AND correct_count >= 3
		)
		ORDER BY RANDOM()
		LIMIT ?
	`, currentLevel, userID, count)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var w Word
		var exampleSentence *string
		err := rows.Scan(&w.ID, &w.Word, &w.Level, &w.Phonetic, &w.Definition,
			&exampleSentence, &w.Category, &w.Difficulty, &w.CreatedAt)
		if err != nil {
			continue
		}

		// CRITICAL: Skip if already used in this session
		if usedWordsMap[strings.ToLower(w.Word)] {
			continue
		}

		if exampleSentence != nil {
			w.ExampleSentence = *exampleSentence
		}

		words = append(words, w)

		// Stop once we have enough unique words
		if len(words) >= count {
			break
		}
	}

	return words, nil
}

// ============================================================================
// WORD MASTERY TRACKING
// ============================================================================

// RecordWordAttempt records a user's attempt on a word
func (app *ReadingApp) RecordWordAttempt(userID int64, word string, isCorrect bool) error {
	now := time.Now()

	// Check if mastery entry exists
	var existingID int64
	err := app.db.QueryRow(
		`SELECT id FROM word_mastery WHERE user_id = ? AND word = ?`,
		userID, word,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new mastery entry
		_, err = app.db.Exec(`
			INSERT INTO word_mastery (user_id, word, correct_count, error_count, total_attempts, last_reviewed_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`,
			userID,
			word,
			boolToInt(isCorrect),
			boolToInt(!isCorrect),
			1,
			now,
		)
		return err
	} else if err != nil {
		return err
	}

	// Update existing mastery entry
	_, err = app.db.Exec(`
		UPDATE word_mastery
		SET correct_count = correct_count + ?,
		    error_count = error_count + ?,
		    total_attempts = total_attempts + 1,
		    last_reviewed_at = ?,
		    mastered_at = CASE WHEN correct_count + ? >= 3 THEN ? ELSE mastered_at END
		WHERE user_id = ? AND word = ?
	`,
		boolToInt(isCorrect),
		boolToInt(!isCorrect),
		now,
		boolToInt(isCorrect),
		func() interface{} {
			if isCorrect {
				return now
			}
			return nil
		}(),
		userID,
		word,
	)

	return err
}

// GetWordMastery retrieves mastery info for a word
func (app *ReadingApp) GetWordMastery(userID int64, word string) (*WordMastery, error) {
	mastery := &WordMastery{}
	err := app.db.QueryRow(`
		SELECT id, user_id, word, correct_count, error_count, total_attempts, last_reviewed_at, mastered_at
		FROM word_mastery
		WHERE user_id = ? AND word = ?
	`, userID, word).Scan(
		&mastery.ID, &mastery.UserID, &mastery.Word, &mastery.CorrectCount,
		&mastery.ErrorCount, &mastery.TotalAttempts, &mastery.LastReviewedAt,
		&mastery.MasteredAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No mastery record yet
	}

	if err != nil {
		return nil, err
	}

	// Calculate accuracy
	if mastery.TotalAttempts > 0 {
		mastery.Accuracy = float64(mastery.CorrectCount*100) / float64(mastery.TotalAttempts)
	}

	return mastery, nil
}

// ============================================================================
// LEVEL PROGRESSION
// ============================================================================

// CalculateUserLevel determines the reading level based on mastery
func (app *ReadingApp) CalculateUserLevel(userID int64) (int, error) {
	// Get all words at each level and their mastery status
	rows, err := app.db.Query(`
		SELECT w.level, COUNT(DISTINCT w.id) as total_words,
		       COUNT(DISTINCT CASE WHEN wm.correct_count >= 3 THEN wm.word END) as mastered_words
		FROM words w
		LEFT JOIN word_mastery wm ON w.word = wm.word AND wm.user_id = ?
		GROUP BY w.level
		ORDER BY w.level
	`, userID)

	if err != nil {
		return 1, err
	}
	defer rows.Close()

	var currentLevel int = 1
	for rows.Next() {
		var level, totalWords, masteredWords int
		if err := rows.Scan(&level, &totalWords, &masteredWords); err != nil {
			continue
		}

		// If mastered 80% of words at this level, can advance
		if totalWords > 0 && (masteredWords*100)/totalWords >= 80 {
			currentLevel = level + 1
		} else {
			break // Stop at first level that's not 80% complete
		}
	}

	return currentLevel, nil
}

// ============================================================================
// SESSION MANAGEMENT
// ============================================================================

// CreatePracticeSession creates a new reading practice session
func (app *ReadingApp) CreatePracticeSession(userID int64, level int) *PracticeSession {
	return &PracticeSession{
		ID:           fmt.Sprintf("read_%d_%d", userID, time.Now().UnixNano()),
		UserID:       userID,
		Level:        level,
		WordsAttempted: []string{},
		UsedWordIDs:  make(map[int64]bool),
		StartedAt:    time.Now(),
	}
}

// ============================================================================
// STATISTICS
// ============================================================================

// GetUserStats retrieves reading statistics for a user
func (app *ReadingApp) GetUserStats(userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total words mastered
	var masteredCount int
	err := app.db.QueryRow(
		`SELECT COUNT(*) FROM word_mastery WHERE user_id = ? AND correct_count >= 3`,
		userID,
	).Scan(&masteredCount)
	if err != nil {
		masteredCount = 0
	}
	stats["words_mastered"] = masteredCount

	// Total attempts
	var totalAttempts int
	err = app.db.QueryRow(
		`SELECT SUM(total_attempts) FROM word_mastery WHERE user_id = ?`,
		userID,
	).Scan(&totalAttempts)
	if err != nil || totalAttempts == 0 {
		totalAttempts = 0
	}
	stats["total_attempts"] = totalAttempts

	// Average accuracy
	var avgAccuracy float64
	err = app.db.QueryRow(`
		SELECT
			CASE
				WHEN SUM(total_attempts) = 0 THEN 0
				ELSE (SUM(correct_count) * 100.0) / SUM(total_attempts)
			END
		FROM word_mastery WHERE user_id = ?
	`, userID).Scan(&avgAccuracy)
	if err != nil {
		avgAccuracy = 0
	}
	stats["average_accuracy"] = avgAccuracy

	// Current level
	level, _ := app.CalculateUserLevel(userID)
	stats["current_level"] = level

	return stats, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// InitDB initializes the reading app database tables
func (app *ReadingApp) InitDB() error {
	_, err := app.db.Exec(`
		CREATE TABLE IF NOT EXISTS words (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			word TEXT NOT NULL UNIQUE,
			level INTEGER NOT NULL,
			phonetic TEXT,
			definition TEXT,
			example_sentence TEXT,
			category TEXT,
			difficulty INTEGER DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS word_mastery (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			word TEXT NOT NULL,
			correct_count INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			total_attempts INTEGER DEFAULT 0,
			last_reviewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			mastered_at TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, word)
		);

		CREATE TABLE IF NOT EXISTS user_progress (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			current_level INTEGER DEFAULT 1,
			total_words_mastered INTEGER DEFAULT 0,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS reading_sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			total_words INTEGER NOT NULL,
			words_completed INTEGER DEFAULT 0,
			accuracy REAL DEFAULT 0,
			total_time INTEGER DEFAULT 0,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE INDEX IF NOT EXISTS idx_words_level ON words(level);
		CREATE INDEX IF NOT EXISTS idx_word_mastery_user ON word_mastery(user_id);
		CREATE INDEX IF NOT EXISTS idx_word_mastery_word ON word_mastery(word);
		CREATE INDEX IF NOT EXISTS idx_user_progress_user ON user_progress(user_id);
	`)
	return err
}
