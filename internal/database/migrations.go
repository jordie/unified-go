package database

import (
	"database/sql"
	"fmt"
	"log"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations is the list of all database migrations
var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_users_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				email TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		`,
	},
	{
		Version: 2,
		Name:    "create_sessions_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS sessions (
				id TEXT PRIMARY KEY,
				user_id INTEGER,
				data TEXT,
				expires_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
			CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
		`,
	},
	{
		Version: 3,
		Name:    "create_migrations_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				name TEXT NOT NULL,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`,
	},
	{
		Version: 4,
		Name:    "create_app_data_tables",
		SQL: `
			-- Typing app data
			CREATE TABLE IF NOT EXISTS typing_progress (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				lesson_id TEXT NOT NULL,
				wpm INTEGER,
				accuracy REAL,
				completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_typing_progress_user_id ON typing_progress(user_id);

			-- Math app data
			CREATE TABLE IF NOT EXISTS math_progress (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				problem_type TEXT NOT NULL,
				correct_answers INTEGER DEFAULT 0,
				total_attempts INTEGER DEFAULT 0,
				completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_math_progress_user_id ON math_progress(user_id);

			-- Reading app data
			CREATE TABLE IF NOT EXISTS reading_progress (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				book_id TEXT NOT NULL,
				page_number INTEGER DEFAULT 0,
				comprehension_score REAL,
				completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_reading_progress_user_id ON reading_progress(user_id);

			-- Piano app data
			CREATE TABLE IF NOT EXISTS piano_progress (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				song_id TEXT NOT NULL,
				accuracy REAL,
				completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_piano_progress_user_id ON piano_progress(user_id);
		`,
	},
	{
		Version: 5,
		Name:    "create_piano_app_tables",
		SQL: `
			-- Piano songs catalog
			CREATE TABLE IF NOT EXISTS songs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				title TEXT NOT NULL,
				composer TEXT NOT NULL,
				description TEXT,
				midi_file BLOB,
				difficulty TEXT,
				duration REAL,
				bpm INTEGER,
				time_signature TEXT,
				key_signature TEXT,
				total_notes INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_songs_difficulty ON songs(difficulty);
			CREATE INDEX IF NOT EXISTS idx_songs_composer ON songs(composer);

			-- Piano lessons (practice sessions)
			CREATE TABLE IF NOT EXISTS piano_lessons (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				song_id INTEGER NOT NULL,
				start_time DATETIME,
				end_time DATETIME,
				duration REAL,
				notes_correct INTEGER,
				notes_total INTEGER,
				accuracy REAL,
				tempo_accuracy REAL,
				score REAL,
				completed INTEGER DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_piano_lessons_user_id ON piano_lessons(user_id);
			CREATE INDEX IF NOT EXISTS idx_piano_lessons_song_id ON piano_lessons(song_id);
			CREATE INDEX IF NOT EXISTS idx_piano_lessons_created_at ON piano_lessons(created_at);

			-- Practice recordings
			CREATE TABLE IF NOT EXISTS practice_sessions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				song_id INTEGER NOT NULL,
				lesson_id INTEGER,
				recording_midi BLOB,
				duration REAL,
				notes_hit INTEGER,
				notes_total INTEGER,
				tempo_average REAL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE,
				FOREIGN KEY (lesson_id) REFERENCES piano_lessons(id) ON DELETE SET NULL
			);
			CREATE INDEX IF NOT EXISTS idx_practice_sessions_user_id ON practice_sessions(user_id);
			CREATE INDEX IF NOT EXISTS idx_practice_sessions_song_id ON practice_sessions(song_id);

			-- Music theory quizzes
			CREATE TABLE IF NOT EXISTS music_theory_quizzes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				questions TEXT,
				answers TEXT,
				score REAL,
				completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_music_theory_quizzes_user_id ON music_theory_quizzes(user_id);

			-- User music metrics
			CREATE TABLE IF NOT EXISTS user_music_metrics (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER UNIQUE NOT NULL,
				total_lessons INTEGER DEFAULT 0,
				average_accuracy REAL DEFAULT 0,
				best_score REAL DEFAULT 0,
				total_practice_time_minutes INTEGER DEFAULT 0,
				skill_level TEXT DEFAULT 'beginner',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_user_music_metrics_user_id ON user_music_metrics(user_id);
		`,
	},
}

// RunMigrations executes all pending database migrations
func RunMigrations(db *Pool) error {
	// Ensure migrations table exists
	if err := ensureMigrationsTable(db.DB); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := getCurrentVersion(db.DB)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	log.Printf("Current database version: %d", currentVersion)

	// Run pending migrations
	applied := 0
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Name)

		tx, err := db.DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute migration SQL
		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}

		// Record migration
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version,
			migration.Name,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		applied++
		log.Printf("Migration %d applied successfully", migration.Version)
	}

	if applied == 0 {
		log.Println("No pending migrations to apply")
	} else {
		log.Printf("Applied %d migrations successfully", applied)
	}

	return nil
}

// ensureMigrationsTable creates the migrations table if it doesn't exist
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// getCurrentVersion returns the latest applied migration version
func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
