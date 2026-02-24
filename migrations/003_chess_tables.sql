-- ============================================================================
-- CHESS APP TABLES
-- ============================================================================

-- Chess games table
CREATE TABLE IF NOT EXISTS chess_games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    white_player_id INTEGER NOT NULL,
    black_player_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    time_control TEXT,
    time_per_side INTEGER,
    white_time_left INTEGER,
    black_time_left INTEGER,
    board_state TEXT NOT NULL,
    current_turn TEXT NOT NULL DEFAULT 'white',
    winner INTEGER,
    win_reason TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (white_player_id) REFERENCES users(id),
    FOREIGN KEY (black_player_id) REFERENCES users(id),
    FOREIGN KEY (winner) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_chess_games_status ON chess_games(status);
CREATE INDEX IF NOT EXISTS idx_chess_games_white_player ON chess_games(white_player_id);
CREATE INDEX IF NOT EXISTS idx_chess_games_black_player ON chess_games(black_player_id);
CREATE INDEX IF NOT EXISTS idx_chess_games_created_at ON chess_games(created_at);
CREATE INDEX IF NOT EXISTS idx_chess_games_winner ON chess_games(winner);

-- Chess moves table
CREATE TABLE IF NOT EXISTS chess_moves (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    move_number INTEGER NOT NULL,
    from_square TEXT NOT NULL,
    to_square TEXT NOT NULL,
    piece TEXT NOT NULL,
    is_capture INTEGER DEFAULT 0,
    is_check INTEGER DEFAULT 0,
    is_checkmate INTEGER DEFAULT 0,
    algebraic_notation TEXT,
    promotion TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES chess_games(id)
);

CREATE INDEX IF NOT EXISTS idx_chess_moves_game_id ON chess_moves(game_id);
CREATE INDEX IF NOT EXISTS idx_chess_moves_move_number ON chess_moves(move_number);

-- Chess players extended table
CREATE TABLE IF NOT EXISTS chess_players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE NOT NULL,
    profile_picture TEXT,
    bio TEXT,
    rating INTEGER DEFAULT 1200,
    rating_tier TEXT DEFAULT 'Bronze',
    games_played INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0,
    last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_chess_players_rating ON chess_players(rating);
CREATE INDEX IF NOT EXISTS idx_chess_players_tier ON chess_players(rating_tier);

-- Chess game results table
CREATE TABLE IF NOT EXISTS chess_game_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER UNIQUE NOT NULL,
    winner_id INTEGER NOT NULL,
    loser_id INTEGER NOT NULL,
    result_type TEXT NOT NULL,
    duration INTEGER,
    move_count INTEGER,
    rating_delta INTEGER,
    xp_earned INTEGER,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES chess_games(id),
    FOREIGN KEY (winner_id) REFERENCES users(id),
    FOREIGN KEY (loser_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_chess_results_winner ON chess_game_results(winner_id);
CREATE INDEX IF NOT EXISTS idx_chess_results_loser ON chess_game_results(loser_id);
CREATE INDEX IF NOT EXISTS idx_chess_results_game_id ON chess_game_results(game_id);

-- Chess player stats table
CREATE TABLE IF NOT EXISTS chess_player_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER UNIQUE NOT NULL,
    games_played INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0,
    win_rate REAL DEFAULT 0.0,
    average_game_duration INTEGER DEFAULT 0,
    favorite_opening TEXT,
    favorite_color TEXT,
    best_rating INTEGER DEFAULT 1200,
    lowest_rating INTEGER DEFAULT 1200,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES chess_players(id)
);

CREATE INDEX IF NOT EXISTS idx_chess_stats_player_id ON chess_player_stats(player_id);

-- Chess rating history table
CREATE TABLE IF NOT EXISTS chess_rating_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER NOT NULL,
    game_id INTEGER NOT NULL,
    old_rating INTEGER,
    new_rating INTEGER,
    delta INTEGER,
    opponent_rating INTEGER,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES chess_players(id),
    FOREIGN KEY (game_id) REFERENCES chess_games(id)
);

CREATE INDEX IF NOT EXISTS idx_rating_history_player ON chess_rating_history(player_id);
CREATE INDEX IF NOT EXISTS idx_rating_history_game ON chess_rating_history(game_id);

-- Chess achievements table
CREATE TABLE IF NOT EXISTS chess_achievements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    achievement_name TEXT NOT NULL,
    description TEXT,
    icon_url TEXT,
    criteria TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Chess player achievements table
CREATE TABLE IF NOT EXISTS chess_player_achievements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER NOT NULL,
    achievement_id INTEGER NOT NULL,
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES chess_players(id),
    FOREIGN KEY (achievement_id) REFERENCES chess_achievements(id)
);

CREATE INDEX IF NOT EXISTS idx_player_achievements ON chess_player_achievements(player_id);
CREATE INDEX IF NOT EXISTS idx_achievement_id ON chess_player_achievements(achievement_id);

-- Chess invitations table
CREATE TABLE IF NOT EXISTS chess_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_player_id INTEGER NOT NULL,
    to_player_id INTEGER NOT NULL,
    time_control TEXT,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP,
    FOREIGN KEY (from_player_id) REFERENCES users(id),
    FOREIGN KEY (to_player_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_invitations_to_player ON chess_invitations(to_player_id);
CREATE INDEX IF NOT EXISTS idx_invitations_from_player ON chess_invitations(from_player_id);

-- Leaderboard view for faster queries
CREATE VIEW IF NOT EXISTS chess_leaderboard AS
SELECT
    cp.id,
    cp.user_id as player_id,
    u.username,
    cp.rating,
    cp.games_played,
    CASE WHEN cp.games_played > 0 THEN ROUND(CAST(cp.wins AS FLOAT) * 100.0 / cp.games_played, 1) ELSE 0 END as win_rate,
    cp.rating_tier,
    cp.joined_at as created_at
FROM chess_players cp
JOIN users u ON cp.user_id = u.id
ORDER BY cp.rating DESC;
