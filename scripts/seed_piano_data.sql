-- Piano App Database Seed Script
-- Creates sample songs, practice sessions, and user metrics

-- Clear existing data
DELETE FROM songs WHERE id > 0;
DELETE FROM piano_lessons WHERE id > 0;
DELETE FROM practice_sessions WHERE id > 0;

-- Insert beginner level songs
INSERT INTO songs (title, composer, description, difficulty, duration, bpm, time_signature, key_signature, total_notes, created_at)
VALUES
  ('Twinkle Twinkle Little Star', 'Traditional', 'Classic beginner piece with simple melody', 'beginner', 45.0, 80, '4/4', 'C Major', 26, datetime('now')),
  ('Mary Had a Little Lamb', 'Sarah Josepha Hale', 'Nursery rhyme, great for beginners', 'beginner', 40.0, 85, '4/4', 'C Major', 22, datetime('now')),
  ('Ode to Joy', 'Ludwig van Beethoven', 'Famous melody from Symphony No. 9', 'beginner', 90.0, 100, '4/4', 'D Major', 64, datetime('now')),
  ('Happy Birthday', 'Mildred J. Hill', 'Beloved celebration song', 'beginner', 30.0, 120, '4/4', 'G Major', 15, datetime('now')),
  ('Jingle Bells', 'James Pierpont', 'Holiday favorite', 'beginner', 60.0, 110, '4/4', 'C Major', 35, datetime('now')),

-- Insert intermediate level songs
  ('Moonlight Sonata (1st Movement)', 'Ludwig van Beethoven', 'Gentle and expressive', 'intermediate', 300.0, 60, '4/4', 'C# Minor', 256, datetime('now')),
  ('Für Elise', 'Ludwig van Beethoven', 'Romantic masterpiece', 'intermediate', 270.0, 76, '4/4', 'A Minor', 224, datetime('now')),
  ('Nocturne Op. 9 No. 2', 'Frédéric Chopin', 'Lyrical and flowing', 'intermediate', 330.0, 72, '4/4', 'E♭ Major', 298, datetime('now')),
  ('Waltz of the Flowers', 'Pyotr Ilyich Tchaikovsky', 'Grand and elegant', 'intermediate', 240.0, 132, '3/4', 'D Major', 280, datetime('now')),
  ('Prelude in C Major', 'Johann Sebastian Bach', 'Flowing arpeggio patterns', 'intermediate', 180.0, 84, '4/4', 'C Major', 156, datetime('now')),

-- Insert advanced level songs
  ('Sonata No. 8 (Pathétique)', 'Ludwig van Beethoven', 'Dramatic and passionate', 'advanced', 480.0, 52, '4/4', 'C Minor', 512, datetime('now')),
  ('Ballade No. 1', 'Frédéric Chopin', 'Complex and emotional', 'advanced', 540.0, 108, '4/4', 'G Minor', 580, datetime('now')),
  ('Hungarian Rhapsody No. 2', 'Franz Liszt', 'Virtuosic and brilliant', 'advanced', 600.0, 140, '2/4', 'C# Minor', 892, datetime('now')),
  ('Rondo Alla Turca', 'Wolfgang Amadeus Mozart', 'Spirited Turkish march', 'advanced', 420.0, 160, '2/4', 'A Major', 456, datetime('now')),
  ('La Campanella', 'Franz Liszt', 'Technical showcase piece', 'advanced', 480.0, 152, '2/4', 'G# Minor', 724, datetime('now')),

-- Insert master level songs
  ('Goldberg Variations BWV 988', 'Johann Sebastian Bach', 'Ultimate contrapuntal masterpiece', 'master', 1200.0, 120, '4/4', 'G Major', 2048, datetime('now')),
  ('The Art of Fugue', 'Johann Sebastian Bach', 'Complex polyphonic work', 'master', 1440.0, 80, '4/4', 'C Minor', 2400, datetime('now')),
  ('Transcendental Etude No. 4', 'Franz Liszt', 'Extreme technical difficulty', 'master', 600.0, 200, '4/4', 'D Minor', 1024, datetime('now')),
  ('Piano Concerto No. 1', 'Pyotr Ilyich Tchaikovsky', 'Grand orchestral arrangement', 'master', 1800.0, 100, '4/4', 'B♭ Minor', 3200, datetime('now')),
  ('Rachmaninoff Prelude Op. 3 No. 2', 'Sergei Rachmaninoff', 'Powerful dramatic piece', 'master', 420.0, 60, '4/4', 'C# Minor', 512, datetime('now'));

-- Insert practice sessions (20 sample lessons)
INSERT INTO piano_lessons (user_id, song_id, start_time, end_time, duration, notes_correct, notes_total, accuracy, tempo_accuracy, score, completed, created_at)
VALUES
  (1, 1, datetime('now', '-7 days'), datetime('now', '-7 days', '+45 seconds'), 45, 24, 26, 92.3, 95.0, 93.65, 1, datetime('now', '-7 days')),
  (1, 2, datetime('now', '-6 days'), datetime('now', '-6 days', '+40 seconds'), 40, 20, 22, 90.9, 92.0, 91.45, 1, datetime('now', '-6 days')),
  (1, 3, datetime('now', '-5 days'), datetime('now', '-5 days', '+90 seconds'), 90, 58, 64, 90.6, 93.5, 92.05, 1, datetime('now', '-5 days')),
  (2, 1, datetime('now', '-7 days'), datetime('now', '-7 days', '+50 seconds'), 50, 22, 26, 84.6, 88.0, 86.3, 1, datetime('now', '-7 days')),
  (2, 4, datetime('now', '-4 days'), datetime('now', '-4 days', '+35 seconds'), 35, 12, 15, 80.0, 85.0, 82.5, 1, datetime('now', '-4 days')),
  (2, 5, datetime('now', '-3 days'), datetime('now', '-3 days', '+65 seconds'), 65, 30, 35, 85.7, 87.5, 86.6, 1, datetime('now', '-3 days')),
  (3, 6, datetime('now', '-6 days'), datetime('now', '-6 days', '+300 seconds'), 300, 225, 256, 87.9, 91.0, 89.45, 1, datetime('now', '-6 days')),
  (3, 7, datetime('now', '-5 days'), datetime('now', '-5 days', '+270 seconds'), 270, 198, 224, 88.4, 89.5, 88.95, 1, datetime('now', '-5 days')),
  (3, 8, datetime('now', '-4 days'), datetime('now', '-4 days', '+330 seconds'), 330, 260, 298, 87.2, 90.0, 88.6, 1, datetime('now', '-4 days')),
  (4, 11, datetime('now', '-7 days'), datetime('now', '-7 days', '+480 seconds'), 480, 430, 512, 84.0, 85.5, 84.75, 1, datetime('now', '-7 days')),
  (4, 12, datetime('now', '-5 days'), datetime('now', '-5 days', '+540 seconds'), 540, 485, 580, 83.6, 86.0, 84.8, 1, datetime('now', '-5 days')),
  (4, 13, datetime('now', '-3 days'), datetime('now', '-3 days', '+600 seconds'), 600, 750, 892, 84.1, 87.5, 85.8, 1, datetime('now', '-3 days')),
  (5, 16, datetime('now', '-6 days'), datetime('now', '-6 days', '+1200 seconds'), 1200, 1850, 2048, 90.3, 94.0, 92.15, 1, datetime('now', '-6 days')),
  (5, 17, datetime('now', '-4 days'), datetime('now', '-4 days', '+1440 seconds'), 1440, 2150, 2400, 89.6, 93.5, 91.55, 1, datetime('now', '-4 days')),
  (5, 18, datetime('now', '-2 days'), datetime('now', '-2 days', '+600 seconds'), 600, 920, 1024, 89.8, 92.5, 91.15, 1, datetime('now', '-2 days')),
  (6, 2, datetime('now', '-5 days'), datetime('now', '-5 days', '+45 seconds'), 45, 18, 22, 81.8, 83.0, 82.4, 1, datetime('now', '-5 days')),
  (6, 3, datetime('now', '-3 days'), datetime('now', '-3 days', '+95 seconds'), 95, 54, 64, 84.4, 86.5, 85.45, 1, datetime('now', '-3 days')),
  (7, 9, datetime('now', '-4 days'), datetime('now', '-4 days', '+240 seconds'), 240, 240, 280, 85.7, 88.0, 86.85, 1, datetime('now', '-4 days')),
  (8, 10, datetime('now', '-3 days'), datetime('now', '-3 days', '+180 seconds'), 180, 140, 156, 89.7, 91.0, 90.35, 1, datetime('now', '-3 days')),
  (9, 14, datetime('now', '-2 days'), datetime('now', '-2 days', '+420 seconds'), 420, 365, 456, 80.0, 82.5, 81.25, 1, datetime('now', '-2 days'));

-- Create/update user music metrics
INSERT OR REPLACE INTO user_music_metrics (user_id, total_lessons, average_accuracy, best_score, total_practice_time_minutes, skill_level, created_at)
SELECT
  COALESCE(u.id, pl.user_id) as user_id,
  COALESCE(COUNT(pl.id), 0) as total_lessons,
  COALESCE(CAST(AVG(pl.accuracy) AS REAL), 0) as average_accuracy,
  COALESCE(MAX(pl.score), 0) as best_score,
  COALESCE(CAST(SUM(pl.duration) / 60.0 AS INTEGER), 0) as total_practice_time_minutes,
  CASE
    WHEN COALESCE(AVG(pl.accuracy), 0) < 60 THEN 'beginner'
    WHEN COALESCE(AVG(pl.accuracy), 0) < 75 THEN 'intermediate'
    WHEN COALESCE(AVG(pl.accuracy), 0) < 90 THEN 'advanced'
    ELSE 'master'
  END as skill_level,
  datetime('now')
FROM users u
LEFT JOIN piano_lessons pl ON u.id = pl.user_id
GROUP BY COALESCE(u.id, pl.user_id);
