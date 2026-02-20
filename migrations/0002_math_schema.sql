-- TARGET: math
-- Migration: Create initial Math app schema with 8 core tables

-- Users table (shared with other apps)
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Practice session results
CREATE TABLE IF NOT EXISTS results (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  mode TEXT NOT NULL CHECK(mode IN ('addition', 'subtraction', 'multiplication', 'division', 'mixed')),
  difficulty TEXT NOT NULL CHECK(difficulty IN ('easy', 'medium', 'hard', 'expert')),
  total_questions INTEGER NOT NULL,
  correct_answers INTEGER NOT NULL,
  total_time REAL NOT NULL,
  average_time REAL NOT NULL,
  accuracy REAL NOT NULL,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  CHECK(correct_answers >= 0 AND correct_answers <= total_questions),
  CHECK(accuracy >= 0 AND accuracy <= 100),
  CHECK(total_time > 0 AND average_time > 0)
);

-- Individual question history
CREATE TABLE IF NOT EXISTS question_history (
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
  FOREIGN KEY (user_id) REFERENCES users(id),
  CHECK(time_taken >= 0)
);

-- Recurring mistakes/errors
CREATE TABLE IF NOT EXISTS mistakes (
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
  UNIQUE(user_id, question),
  CHECK(error_count > 0)
);

-- Individual fact mastery tracking
CREATE TABLE IF NOT EXISTS mastery (
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
  UNIQUE(user_id, fact, mode),
  CHECK(correct_streak >= 0),
  CHECK(total_attempts >= 0),
  CHECK(mastery_level >= 0 AND mastery_level <= 100),
  CHECK(average_response_time >= 0)
);

-- User learning profile and characteristics
CREATE TABLE IF NOT EXISTS learning_profile (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER UNIQUE NOT NULL,
  learning_style TEXT CHECK(learning_style IN ('visual', 'sequential', 'global')),
  preferred_time_of_day TEXT CHECK(preferred_time_of_day IN ('morning', 'afternoon', 'evening')),
  attention_span_seconds INTEGER DEFAULT 300,
  best_streak_time TEXT,
  weak_time_of_day TEXT,
  avg_session_length INTEGER NOT NULL DEFAULT 0,
  total_practice_time INTEGER NOT NULL DEFAULT 0,
  profile_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  CHECK(attention_span_seconds > 0),
  CHECK(avg_session_length >= 0),
  CHECK(total_practice_time >= 0)
);

-- Time-of-day performance patterns
CREATE TABLE IF NOT EXISTS performance_patterns (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  hour_of_day INTEGER NOT NULL,
  day_of_week INTEGER NOT NULL,
  average_accuracy REAL NOT NULL DEFAULT 0,
  average_speed REAL NOT NULL DEFAULT 0,
  session_count INTEGER NOT NULL DEFAULT 0,
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  UNIQUE(user_id, hour_of_day, day_of_week),
  CHECK(hour_of_day >= 0 AND hour_of_day <= 23),
  CHECK(day_of_week >= 0 AND day_of_week <= 6),
  CHECK(average_accuracy >= 0 AND average_accuracy <= 100),
  CHECK(average_speed >= 0),
  CHECK(session_count >= 0)
);

-- Spaced Repetition Schedule (SM-2 Algorithm)
CREATE TABLE IF NOT EXISTS repetition_schedule (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  fact TEXT NOT NULL,
  mode TEXT,
  next_review TIMESTAMP NOT NULL,
  interval_days INTEGER NOT NULL DEFAULT 1,
  ease_factor REAL NOT NULL DEFAULT 2.5,
  review_count INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY (user_id) REFERENCES users(id),
  UNIQUE(user_id, fact, mode),
  CHECK(interval_days >= 1),
  CHECK(ease_factor >= 1.3 AND ease_factor <= 3.5),
  CHECK(review_count >= 0)
);

-- Create indexes for frequently queried columns
CREATE INDEX IF NOT EXISTS idx_results_user_id ON results(user_id);
CREATE INDEX IF NOT EXISTS idx_results_timestamp ON results(timestamp);
CREATE INDEX IF NOT EXISTS idx_question_history_user_id ON question_history(user_id);
CREATE INDEX IF NOT EXISTS idx_question_history_timestamp ON question_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_mistakes_user_id ON mistakes(user_id);
CREATE INDEX IF NOT EXISTS idx_mastery_user_id ON mastery(user_id);
CREATE INDEX IF NOT EXISTS idx_performance_patterns_user_id ON performance_patterns(user_id);
CREATE INDEX IF NOT EXISTS idx_repetition_schedule_user_id ON repetition_schedule(user_id);
CREATE INDEX IF NOT EXISTS idx_repetition_schedule_next_review ON repetition_schedule(next_review);
CREATE INDEX IF NOT EXISTS idx_learning_profile_user_id ON learning_profile(user_id);
