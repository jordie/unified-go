-- Generate 20 test users for Piano app
-- Usage: sqlite3 data/unified.db < scripts/generate_piano_users.sql

-- Insert test users (extending existing users table)
INSERT OR IGNORE INTO users (id, username, email, password_hash, created_at)
VALUES
  (1, 'alice_piano', 'alice@piano.local', 'hash_alice', CURRENT_TIMESTAMP),
  (2, 'bob_beginner', 'bob@piano.local', 'hash_bob', CURRENT_TIMESTAMP),
  (3, 'carol_advanced', 'carol@piano.local', 'hash_carol', CURRENT_TIMESTAMP),
  (4, 'dave_master', 'dave@piano.local', 'hash_dave', CURRENT_TIMESTAMP),
  (5, 'eve_intermediate', 'eve@piano.local', 'hash_eve', CURRENT_TIMESTAMP),
  (6, 'frank_beginner', 'frank@piano.local', 'hash_frank', CURRENT_TIMESTAMP),
  (7, 'grace_advanced', 'grace@piano.local', 'hash_grace', CURRENT_TIMESTAMP),
  (8, 'henry_master', 'henry@piano.local', 'hash_henry', CURRENT_TIMESTAMP),
  (9, 'iris_intermediate', 'iris@piano.local', 'hash_iris', CURRENT_TIMESTAMP),
  (10, 'jack_beginner', 'jack@piano.local', 'hash_jack', CURRENT_TIMESTAMP),
  (11, 'kate_advanced', 'kate@piano.local', 'hash_kate', CURRENT_TIMESTAMP),
  (12, 'liam_expert', 'liam@piano.local', 'hash_liam', CURRENT_TIMESTAMP),
  (13, 'mona_intermediate', 'mona@piano.local', 'hash_mona', CURRENT_TIMESTAMP),
  (14, 'noah_beginner', 'noah@piano.local', 'hash_noah', CURRENT_TIMESTAMP),
  (15, 'olivia_advanced', 'olivia@piano.local', 'hash_olivia', CURRENT_TIMESTAMP),
  (16, 'paul_master', 'paul@piano.local', 'hash_paul', CURRENT_TIMESTAMP),
  (17, 'quinn_intermediate', 'quinn@piano.local', 'hash_quinn', CURRENT_TIMESTAMP),
  (18, 'rachel_beginner', 'rachel@piano.local', 'hash_rachel', CURRENT_TIMESTAMP),
  (19, 'sam_advanced', 'sam@piano.local', 'hash_sam', CURRENT_TIMESTAMP),
  (20, 'tara_expert', 'tara@piano.local', 'hash_tara', CURRENT_TIMESTAMP);

-- Verify insertion
SELECT COUNT(*) as total_users FROM users WHERE email LIKE '%@piano.local';
