package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Pool represents a database connection pool
type Pool struct {
	*sql.DB
}

// InitPool initializes a new SQLite database connection pool with WAL mode
func InitPool(databaseURL string) (*Pool, error) {
	// Ensure directory exists
	dir := filepath.Dir(databaseURL)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", databaseURL+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable WAL mode and other optimizations
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=1000000000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			log.Printf("Warning: failed to execute %s: %v", pragma, err)
		}
	}

	log.Println("Database connection pool initialized with WAL mode")

	return &Pool{DB: db}, nil
}

// Close closes the database connection pool
func (p *Pool) Close() error {
	log.Println("Closing database connection pool")
	return p.DB.Close()
}

// HealthCheck verifies the database connection is healthy
func (p *Pool) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Stats returns database pool statistics
func (p *Pool) Stats() sql.DBStats {
	return p.DB.Stats()
}
