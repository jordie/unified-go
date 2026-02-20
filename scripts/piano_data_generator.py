#!/usr/bin/env python3
"""
Piano App Test Data Generator
Generates realistic test data for comprehensive Piano app testing.

Usage:
    python3 scripts/piano_data_generator.py --generate all
    python3 scripts/piano_data_generator.py --generate users
    python3 scripts/piano_data_generator.py --generate lessons
    python3 scripts/piano_data_generator.py --stats
    python3 scripts/piano_data_generator.py --clean
"""

import sqlite3
import argparse
import random
from datetime import datetime, timedelta
from pathlib import Path

DATABASE = Path("data/unified.db")

# Song catalog (20 pieces)
SONGS = {
    "beginner": [
        (1, "Twinkle Twinkle Little Star", "Traditional", 45, 80, 26),
        (2, "Mary Had a Little Lamb", "Sarah Josepha Hale", 50, 75, 25),
        (3, "Ode to Joy", "Ludwig van Beethoven", 55, 85, 28),
        (4, "Happy Birthday", "Mildred J. Hill", 40, 100, 24),
        (5, "Jingle Bells", "James Pierpont", 60, 120, 30),
    ],
    "intermediate": [
        (6, "Moonlight Sonata (1st Movement)", "Beethoven", 180, 70, 120),
        (7, "Für Elise", "Beethoven", 150, 90, 110),
        (8, "Nocturne Op. 9 No. 2", "Chopin", 200, 60, 130),
        (9, "Waltz of the Flowers", "Tchaikovsky", 240, 140, 140),
        (10, "Prelude in C Major", "Bach", 160, 80, 100),
    ],
    "advanced": [
        (11, "Sonata No. 8 (Pathétique)", "Beethoven", 900, 80, 200),
        (12, "Ballade No. 1", "Chopin", 1200, 70, 220),
        (13, "Hungarian Rhapsody No. 2", "Liszt", 1400, 160, 240),
        (14, "Rondo Alla Turca", "Mozart", 480, 140, 180),
        (15, "La Campanella", "Liszt", 600, 180, 200),
    ],
    "master": [
        (16, "Goldberg Variations BWV 988", "Bach", 2400, 100, 300),
        (17, "The Art of Fugue", "Bach", 2800, 80, 320),
        (18, "Transcendental Etude No. 4", "Liszt", 1800, 180, 280),
        (19, "Piano Concerto No. 1", "Tchaikovsky", 2000, 120, 300),
        (20, "Rachmaninoff Prelude Op. 3 No. 2", "Rachmaninoff", 900, 60, 250),
    ],
}

USERS = [
    (1, "alice_piano", "alice@piano.local"),
    (2, "bob_beginner", "bob@piano.local"),
    (3, "carol_advanced", "carol@piano.local"),
    (4, "dave_master", "dave@piano.local"),
    (5, "eve_intermediate", "eve@piano.local"),
    (6, "frank_beginner", "frank@piano.local"),
    (7, "grace_advanced", "grace@piano.local"),
    (8, "henry_master", "henry@piano.local"),
    (9, "iris_intermediate", "iris@piano.local"),
    (10, "jack_beginner", "jack@piano.local"),
    (11, "kate_advanced", "kate@piano.local"),
    (12, "liam_expert", "liam@piano.local"),
    (13, "mona_intermediate", "mona@piano.local"),
    (14, "noah_beginner", "noah@piano.local"),
    (15, "olivia_advanced", "olivia@piano.local"),
    (16, "paul_master", "paul@piano.local"),
    (17, "quinn_intermediate", "quinn@piano.local"),
    (18, "rachel_beginner", "rachel@piano.local"),
    (19, "sam_advanced", "sam@piano.local"),
    (20, "tara_expert", "tara@piano.local"),
]

def connect_db():
    """Connect to database."""
    return sqlite3.connect(DATABASE)

def generate_users(conn):
    """Insert test users."""
    cursor = conn.cursor()
    
    for uid, username, email in USERS:
        try:
            cursor.execute(
                "INSERT OR IGNORE INTO users (id, username, email, password_hash, created_at) "
                "VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)",
                (uid, username, email, f"hash_{username}")
            )
        except Exception as e:
            print(f"  Error inserting user {username}: {e}")
    
    conn.commit()
    count = cursor.execute("SELECT COUNT(*) FROM users WHERE email LIKE '%@piano.local'").fetchone()[0]
    print(f"✅ Generated {count} test users")

def generate_lessons(conn, num_per_user=5):
    """Generate realistic piano lessons for each user."""
    cursor = conn.cursor()
    
    # User-to-difficulty mapping
    user_levels = {
        # Beginner users (2, 6, 10, 14, 18)
        2: ("beginner", 1, 5), 6: ("beginner", 1, 5), 10: ("beginner", 1, 5),
        14: ("beginner", 1, 5), 18: ("beginner", 1, 5),
        # Intermediate users (5, 9, 13, 17)
        5: ("intermediate", 6, 10), 9: ("intermediate", 6, 10),
        13: ("intermediate", 6, 10), 17: ("intermediate", 6, 10),
        # Advanced users (3, 7, 11, 15, 19)
        3: ("advanced", 11, 15), 7: ("advanced", 11, 15),
        11: ("advanced", 11, 15), 15: ("advanced", 11, 15), 19: ("advanced", 11, 15),
        # Master users (4, 8, 12, 16, 20)
        4: ("master", 16, 20), 8: ("master", 16, 20),
        12: ("master", 16, 20), 16: ("master", 16, 20), 20: ("master", 16, 20),
        1: ("advanced", 11, 15),  # Alice is advanced
    }
    
    lesson_count = 0
    
    for uid, username, _ in USERS:
        if uid not in user_levels:
            continue
            
        level, song_start, song_end = user_levels[uid]
        
        # Generate 3-5 lessons per user over past 7 days
        for i in range(random.randint(3, 5)):
            song_id = random.randint(song_start, song_end)
            days_ago = random.randint(1, 7)
            
            # Get song info for realistic duration
            song_info = None
            for songs in SONGS.values():
                for song in songs:
                    if song[0] == song_id:
                        song_info = song
                        break
            
            if not song_info:
                continue
            
            song_id, title, composer, duration, bpm, total_notes = song_info
            
            # Generate realistic metrics
            accuracy = random.uniform(75, 97)
            tempo_accuracy = random.uniform(80, 99)
            notes_correct = int(total_notes * accuracy / 100)
            score = (accuracy * 0.7) + (tempo_accuracy * 0.3)
            
            start_time = (datetime.now() - timedelta(days=days_ago)).isoformat()
            end_time = (datetime.now() - timedelta(days=days_ago, seconds=-int(duration))).isoformat()
            
            try:
                cursor.execute(
                    "INSERT INTO piano_lessons "
                    "(user_id, song_id, start_time, end_time, duration, notes_correct, "
                    "notes_total, accuracy, tempo_accuracy, score, completed, created_at) "
                    "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, CURRENT_TIMESTAMP)",
                    (uid, song_id, start_time, end_time, duration, notes_correct,
                     total_notes, round(accuracy, 2), round(tempo_accuracy, 2), round(score, 2))
                )
                lesson_count += 1
            except Exception as e:
                print(f"  Error inserting lesson for user {uid}: {e}")
    
    conn.commit()
    print(f"✅ Generated {lesson_count} practice lessons")

def show_stats(conn):
    """Display database statistics."""
    cursor = conn.cursor()
    
    print("\n📊 Database Statistics:")
    print("=" * 50)
    
    # User stats
    user_count = cursor.execute("SELECT COUNT(*) FROM users WHERE email LIKE '%@piano.local'").fetchone()[0]
    print(f"  Users: {user_count}")
    
    # Song stats
    song_count = cursor.execute("SELECT COUNT(*) FROM songs").fetchone()[0]
    print(f"  Songs: {song_count}")
    
    # Lesson stats
    lesson_count = cursor.execute("SELECT COUNT(*) FROM piano_lessons").fetchone()[0]
    print(f"  Lessons: {lesson_count}")
    
    # Accuracy stats
    stats = cursor.execute(
        "SELECT AVG(accuracy), MIN(accuracy), MAX(accuracy) FROM piano_lessons"
    ).fetchone()
    if stats[0]:
        print(f"  Average Accuracy: {stats[0]:.2f}%")
        print(f"  Min/Max Accuracy: {stats[1]:.2f}% / {stats[2]:.2f}%")
    
    # Users per difficulty
    print("\n  Lessons by Difficulty:")
    for difficulty in ["beginner", "intermediate", "advanced", "master"]:
        count = cursor.execute(
            "SELECT COUNT(*) FROM piano_lessons pl "
            "JOIN songs s ON pl.song_id = s.id "
            "WHERE s.difficulty = ?",
            (difficulty,)
        ).fetchone()[0]
        print(f"    {difficulty.title()}: {count}")
    
    print("=" * 50 + "\n")

def clean_data(conn):
    """Clean all Piano app test data."""
    cursor = conn.cursor()
    
    cursor.execute("DELETE FROM piano_lessons WHERE user_id > 1 OR user_id IN (SELECT id FROM users WHERE email LIKE '%@piano.local')")
    cursor.execute("DELETE FROM users WHERE email LIKE '%@piano.local'")
    cursor.execute("DELETE FROM songs")
    
    conn.commit()
    print("✅ Cleaned all Piano app test data")

def main():
    parser = argparse.ArgumentParser(description="Piano app test data generator")
    parser.add_argument("--generate", choices=["all", "users", "lessons"], 
                       help="Generate test data")
    parser.add_argument("--stats", action="store_true", help="Show database statistics")
    parser.add_argument("--clean", action="store_true", help="Clean all test data")
    
    args = parser.parse_args()
    
    if not DATABASE.exists():
        print(f"❌ Database not found: {DATABASE}")
        return
    
    conn = connect_db()
    
    try:
        if args.generate == "all":
            print("Generating test data...")
            generate_users(conn)
            generate_lessons(conn)
            show_stats(conn)
        elif args.generate == "users":
            print("Generating users...")
            generate_users(conn)
        elif args.generate == "lessons":
            print("Generating lessons...")
            generate_lessons(conn)
        elif args.stats:
            show_stats(conn)
        elif args.clean:
            confirm = input("⚠️  This will delete all Piano app test data. Continue? (y/N): ")
            if confirm.lower() == "y":
                clean_data(conn)
        else:
            parser.print_help()
    finally:
        conn.close()

if __name__ == "__main__":
    main()
