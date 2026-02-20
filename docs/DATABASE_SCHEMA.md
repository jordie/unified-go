# Database Schema Reference

Complete documentation of all database tables, columns, and relationships.

## Overview

The unified-go platform uses SQLite3 with 16+ tables organized by application.

**Key Statistics:**
- 16 core tables
- 12+ indexes for query optimization
- Foreign key constraints enabled
- Total schema: ~50KB

## Core Tables

### users

User account information.

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE,
    password_hash TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    is_active INTEGER DEFAULT 1
);
```

## Reading App Tables

### reading_passages
Reading comprehension passages.

### reading_questions
Comprehension questions for passages.

### reading_results
User answers to reading questions.

### reading_user_stats
Aggregated reading statistics per user.

## Typing App Tables

### typing_tests
Typing speed tests.

### typing_results
Typing test results and metrics.

### typing_user_stats
Aggregated typing statistics per user.

## Math App Tables

### math_problems
Math practice problems (addition, subtraction, multiplication, division, fractions, algebra).

### math_solutions
User attempts on math problems.

### math_sessions
Math practice sessions.

### math_user_stats
Aggregated math statistics per user.

## Piano App Tables

### piano_lessons
Piano lesson content (placeholder for Phase 5).

### piano_progress
User piano progress tracking.

## Key Statistics

- All tables use INTEGER PRIMARY KEY AUTOINCREMENT
- Foreign key constraints enforced (PRAGMA foreign_keys = ON)
- Indexes created on user_id and frequently queried columns
- Unique constraints on username/email and per-user stats tables

## Example Queries

### Get user stats
```sql
SELECT * FROM reading_user_stats WHERE user_id = 1;
```

### Leaderboard
```sql
SELECT u.username, r.overall_accuracy, r.passages_completed
FROM users u
JOIN reading_user_stats r ON u.id = r.user_id
ORDER BY r.overall_accuracy DESC LIMIT 10;
```

## Maintenance

```bash
# Backup
sqlite3 data/unified.db ".dump" > backup.sql

# Vacuum (optimize)
sqlite3 data/unified.db "VACUUM;"

# Check integrity
sqlite3 data/unified.db "PRAGMA integrity_check;"
```

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for database issues.
