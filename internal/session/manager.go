package session

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session with device fingerprint tracking
type Session struct {
	ID              string    `json:"id"`
	UserID          int64     `json:"user_id"`
	Username        string    `json:"username"`
	DeviceFingerprint string  `json:"device_fingerprint"`
	CreatedAt       time.Time `json:"created_at"`
	LastActivity    time.Time `json:"last_activity"`
	ExpiresAt       time.Time `json:"expires_at"`
	Active          bool      `json:"active"`
}

// User represents a unified user across all education apps
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`

	// App-specific data
	XP              int64  `json:"xp"`
	Level           int    `json:"level"`
	TotalSessions   int    `json:"total_sessions"`
	PreferredApp    string `json:"preferred_app"`
}

// Manager handles session lifecycle and user management
type Manager struct {
	db              *sql.DB
	sessionCache    map[string]*Session
	cacheMutex      sync.RWMutex
	sessionTimeout  time.Duration
	cleanupInterval time.Duration
}

// NewManager creates a new session manager
func NewManager(db *sql.DB) *Manager {
	m := &Manager{
		db:              db,
		sessionCache:    make(map[string]*Session),
		sessionTimeout:  24 * time.Hour,
		cleanupInterval: 1 * time.Hour,
	}

	// Start cleanup goroutine
	go m.cleanupExpiredSessions()

	return m
}

// CreateSession creates a new user session
func (m *Manager) CreateSession(userID int64, username, deviceFingerprint string) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(m.sessionTimeout)

	session := &Session{
		ID:                sessionID,
		UserID:            userID,
		Username:          username,
		DeviceFingerprint: deviceFingerprint,
		CreatedAt:         now,
		LastActivity:      now,
		ExpiresAt:         expiresAt,
		Active:            true,
	}

	// Save to cache
	m.cacheMutex.Lock()
	m.sessionCache[sessionID] = session
	m.cacheMutex.Unlock()

	// Save to database
	_, err := m.db.Exec(
		`INSERT INTO sessions (id, user_id, username, device_fingerprint, created_at, last_activity, expires_at, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, userID, username, deviceFingerprint, now, now, expiresAt, 1,
	)

	return session, err
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	// Check cache first
	m.cacheMutex.RLock()
	if session, exists := m.sessionCache[sessionID]; exists {
		m.cacheMutex.RUnlock()

		// Check if expired
		if time.Now().After(session.ExpiresAt) {
			_ = m.InvalidateSession(sessionID)
			return nil, errors.New("session expired")
		}

		return session, nil
	}
	m.cacheMutex.RUnlock()

	// Fetch from database
	session := &Session{}
	err := m.db.QueryRow(
		`SELECT id, user_id, username, device_fingerprint, created_at, last_activity, expires_at, active
		 FROM sessions WHERE id = ?`,
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.Username, &session.DeviceFingerprint,
		&session.CreatedAt, &session.LastActivity, &session.ExpiresAt, &session.Active)

	if err != nil {
		return nil, err
	}

	// Check if expired
	if !session.Active || time.Now().After(session.ExpiresAt) {
		_ = m.InvalidateSession(sessionID)
		return nil, errors.New("session expired")
	}

	// Update cache
	m.cacheMutex.Lock()
	m.sessionCache[sessionID] = session
	m.cacheMutex.Unlock()

	return session, nil
}

// UpdateLastActivity updates the last activity timestamp for a session
func (m *Manager) UpdateLastActivity(sessionID string) error {
	now := time.Now()

	// Update database
	_, err := m.db.Exec(
		`UPDATE sessions SET last_activity = ? WHERE id = ?`,
		now, sessionID,
	)

	if err != nil {
		return err
	}

	// Update cache
	m.cacheMutex.Lock()
	if session, exists := m.sessionCache[sessionID]; exists {
		session.LastActivity = now
	}
	m.cacheMutex.Unlock()

	return nil
}

// InvalidateSession marks a session as invalid
func (m *Manager) InvalidateSession(sessionID string) error {
	// Update database
	_, err := m.db.Exec(
		`UPDATE sessions SET active = 0 WHERE id = ?`,
		sessionID,
	)

	if err != nil {
		return err
	}

	// Remove from cache
	m.cacheMutex.Lock()
	delete(m.sessionCache, sessionID)
	m.cacheMutex.Unlock()

	return nil
}

// CreateUser creates a new user across all apps
func (m *Manager) CreateUser(username string) (*User, error) {
	now := time.Now()
	user := &User{
		Username:      username,
		CreatedAt:     now,
		LastActive:    now,
		XP:            0,
		Level:         1,
		TotalSessions: 1,
	}

	result, err := m.db.Exec(
		`INSERT INTO users (username, created_at, last_active, xp, level, total_sessions)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		username, now, now, 0, 1, 1,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

// GetUser retrieves a user by ID
func (m *Manager) GetUser(userID int64) (*User, error) {
	user := &User{}
	err := m.db.QueryRow(
		`SELECT id, username, created_at, last_active, xp, level, total_sessions
		 FROM users WHERE id = ?`,
		userID,
	).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastActive,
		&user.XP, &user.Level, &user.TotalSessions)

	return user, err
}

// GetUserByUsername retrieves a user by username
func (m *Manager) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := m.db.QueryRow(
		`SELECT id, username, created_at, last_active, xp, level, total_sessions
		 FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastActive,
		&user.XP, &user.Level, &user.TotalSessions)

	return user, err
}

// ListUsers retrieves all users
func (m *Manager) ListUsers() ([]*User, error) {
	rows, err := m.db.Query(
		`SELECT id, username, created_at, last_active, xp, level, total_sessions FROM users ORDER BY username`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LastActive,
			&user.XP, &user.Level, &user.TotalSessions)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// UpdateUserLastActive updates the last active timestamp for a user
func (m *Manager) UpdateUserLastActive(userID int64) error {
	now := time.Now()
	_, err := m.db.Exec(
		`UPDATE users SET last_active = ? WHERE id = ?`,
		now, userID,
	)
	return err
}

// AddUserXP adds XP to a user and handles level progression
func (m *Manager) AddUserXP(userID int64, xp int64) error {
	user, err := m.GetUser(userID)
	if err != nil {
		return err
	}

	newXP := user.XP + xp
	newLevel := 1 + (int(newXP) / 1000) // Level up every 1000 XP

	_, err = m.db.Exec(
		`UPDATE users SET xp = ?, level = ? WHERE id = ?`,
		newXP, newLevel, userID,
	)

	return err
}

// cleanupExpiredSessions periodically removes expired sessions from cache
func (m *Manager) cleanupExpiredSessions() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.cacheMutex.Lock()
		now := time.Now()

		for id, session := range m.sessionCache {
			if !session.Active || now.After(session.ExpiresAt) {
				delete(m.sessionCache, id)
			}
		}

		m.cacheMutex.Unlock()

		// Also clean database
		_ = m.db.QueryRow(`DELETE FROM sessions WHERE expires_at < ? OR active = 0`, now)
	}
}

// GetSessionByDeviceFingerprint attempts auto-login via device fingerprint
func (m *Manager) GetSessionByDeviceFingerprint(userID int64, deviceFingerprint string) (*Session, error) {
	session := &Session{}
	err := m.db.QueryRow(
		`SELECT id, user_id, username, device_fingerprint, created_at, last_activity, expires_at, active
		 FROM sessions WHERE user_id = ? AND device_fingerprint = ? AND active = 1 AND expires_at > ?`,
		userID, deviceFingerprint, time.Now(),
	).Scan(&session.ID, &session.UserID, &session.Username, &session.DeviceFingerprint,
		&session.CreatedAt, &session.LastActivity, &session.ExpiresAt, &session.Active)

	if err != nil {
		return nil, err
	}

	// Update last activity
	_ = m.UpdateLastActivity(session.ID)

	return session, nil
}
