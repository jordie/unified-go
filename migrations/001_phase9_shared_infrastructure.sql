-- TARGET: all
-- Phase 9 Shared Infrastructure Schema
-- Creates unified user, session, and shared data tables

-- ============================================================================
-- USERS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    xp INTEGER DEFAULT 0,
    level INTEGER DEFAULT 1,
    total_sessions INTEGER DEFAULT 0,
    preferred_app TEXT DEFAULT 'typing'
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- ============================================================================
-- SESSIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    username TEXT NOT NULL,
    device_fingerprint TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    active INTEGER DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(active);

-- ============================================================================
-- DEVICE FINGERPRINTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS device_fingerprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    fingerprint TEXT NOT NULL UNIQUE,
    device_name TEXT,
    last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    trusted INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_device_fingerprints_user_id ON device_fingerprints(user_id);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_fingerprint ON device_fingerprints(fingerprint);

-- ============================================================================
-- XP LOG TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS xp_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    app_name TEXT NOT NULL,
    amount INTEGER NOT NULL,
    source TEXT,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_xp_log_user_id ON xp_log(user_id);
CREATE INDEX IF NOT EXISTS idx_xp_log_app_name ON xp_log(app_name);
CREATE INDEX IF NOT EXISTS idx_xp_log_created_at ON xp_log(created_at);

-- ============================================================================
-- GOALS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    app_name TEXT NOT NULL,
    goal_type TEXT,
    target_value INTEGER,
    current_value INTEGER DEFAULT 0,
    due_date TIMESTAMP,
    completed INTEGER DEFAULT 0,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_goals_user_id ON goals(user_id);
CREATE INDEX IF NOT EXISTS idx_goals_app_name ON goals(app_name);

-- ============================================================================
-- TYPING APP TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS typing_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    wpm INTEGER NOT NULL,
    raw_wpm INTEGER,
    accuracy REAL NOT NULL,
    test_type TEXT,
    test_mode TEXT,
    test_duration INTEGER,
    total_characters INTEGER,
    correct_characters INTEGER,
    incorrect_characters INTEGER,
    errors INTEGER,
    time_taken REAL,
    text_snippet TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_typing_results_user_id ON typing_results(user_id);
CREATE INDEX IF NOT EXISTS idx_typing_results_created_at ON typing_results(created_at);

CREATE TABLE IF NOT EXISTS typing_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE,
    total_tests INTEGER DEFAULT 0,
    average_wpm REAL DEFAULT 0,
    average_accuracy REAL DEFAULT 0,
    best_wpm INTEGER DEFAULT 0,
    total_time_typed INTEGER DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS races (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    difficulty TEXT DEFAULT 'medium',
    placement INTEGER,
    wpm INTEGER,
    accuracy REAL,
    race_time REAL,
    xp_earned INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS racing_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE,
    total_races INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    podiums INTEGER DEFAULT 0,
    total_xp INTEGER DEFAULT 0,
    current_car TEXT DEFAULT '🚗',
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- ============================================================================
-- MATH APP TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS math_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    problem_id TEXT,
    operation TEXT,
    operand1 INTEGER,
    operand2 INTEGER,
    difficulty TEXT,
    user_answer REAL,
    correct_answer REAL,
    is_correct INTEGER,
    time_taken REAL,
    session_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_math_results_user_id ON math_results(user_id);
CREATE INDEX IF NOT EXISTS idx_math_results_created_at ON math_results(created_at);

CREATE TABLE IF NOT EXISTS math_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE,
    total_problems INTEGER DEFAULT 0,
    correct_count INTEGER DEFAULT 0,
    incorrect_count INTEGER DEFAULT 0,
    accuracy REAL DEFAULT 0,
    average_time REAL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS math_weaknesses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    operation TEXT,
    difficulty TEXT,
    error_rate REAL,
    priority INTEGER,
    recommended_action TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- ============================================================================
-- READING APP TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS reading_passages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    content TEXT,
    difficulty_level INTEGER,
    word_count INTEGER,
    grade_level INTEGER,
    category TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reading_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    passage_id INTEGER,
    words_read INTEGER,
    correct_words INTEGER,
    incorrect_words INTEGER,
    accuracy REAL,
    time_spent INTEGER,
    reading_speed REAL,
    comprehension_score REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (passage_id) REFERENCES reading_passages(id)
);

CREATE INDEX IF NOT EXISTS idx_reading_results_user_id ON reading_results(user_id);

CREATE TABLE IF NOT EXISTS reading_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE,
    total_words_read INTEGER DEFAULT 0,
    average_accuracy REAL DEFAULT 0,
    average_speed REAL DEFAULT 0,
    average_comprehension REAL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS word_mastery (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    word TEXT,
    mastery_level INTEGER DEFAULT 0,
    attempt_count INTEGER DEFAULT 0,
    correct_count INTEGER DEFAULT 0,
    last_attempt TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_word_mastery_user_id ON word_mastery(user_id);

-- ============================================================================
-- PIANO APP TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS piano_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    duration INTEGER,
    notes_played INTEGER,
    notes_correct INTEGER,
    accuracy REAL,
    level INTEGER,
    difficulty TEXT,
    piece_id TEXT,
    exercise_type TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_piano_sessions_user_id ON piano_sessions(user_id);

CREATE TABLE IF NOT EXISTS piano_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE,
    current_level INTEGER DEFAULT 1,
    average_accuracy REAL DEFAULT 0,
    total_notes INTEGER DEFAULT 0,
    correct_notes INTEGER DEFAULT 0,
    streak_days INTEGER DEFAULT 0,
    total_practice_time INTEGER DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS piano_badges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    badge_name TEXT,
    description TEXT,
    icon_url TEXT,
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- ============================================================================
-- PROGRESS TRACKING TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_journals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    app_name TEXT,
    title TEXT,
    content TEXT,
    reflection TEXT,
    mood_rating INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS progress_milestones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    app_name TEXT,
    title TEXT,
    description TEXT,
    achieved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_users_xp ON users(xp);
CREATE INDEX IF NOT EXISTS idx_users_level ON users(level);
CREATE INDEX IF NOT EXISTS idx_xp_log_created_at ON xp_log(created_at DESC);

-- ============================================================================
-- TRIGGERS FOR AUTOMATIC UPDATES
-- ============================================================================

-- Update users.last_active when a session is created
CREATE TRIGGER IF NOT EXISTS update_user_last_active_on_session
AFTER INSERT ON sessions
BEGIN
    UPDATE users SET last_active = CURRENT_TIMESTAMP WHERE id = NEW.user_id;
END;

-- Clean up expired sessions periodically (in application code)
-- Sessions with expires_at < NOW() and active = 0 should be cleaned up
