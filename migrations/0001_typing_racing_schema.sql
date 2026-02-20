-- TARGET: typing
-- Migration: Create typing racing tables for Phase 5 migration
-- Description: Creates races and user_racing_stats tables for racing mode support

-- Create races table for recording race sessions
CREATE TABLE IF NOT EXISTS races (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    mode TEXT DEFAULT 'standard',
    placement INTEGER NOT NULL,
    wpm REAL NOT NULL,
    accuracy REAL NOT NULL,
    race_time REAL NOT NULL,
    xp_earned INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create user_racing_stats table for aggregated racing statistics
CREATE TABLE IF NOT EXISTS user_racing_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE NOT NULL,
    total_races INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    podiums INTEGER DEFAULT 0,
    total_xp INTEGER DEFAULT 0,
    current_car TEXT DEFAULT '🚗',
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_races_user_id ON races(user_id);
CREATE INDEX IF NOT EXISTS idx_races_created_at ON races(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_races_placement ON races(placement);
CREATE INDEX IF NOT EXISTS idx_user_racing_stats_total_xp ON user_racing_stats(total_xp DESC);
CREATE INDEX IF NOT EXISTS idx_user_racing_stats_wins ON user_racing_stats(wins DESC);
