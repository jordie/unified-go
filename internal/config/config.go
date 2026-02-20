package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port          int
	Host          string
	Environment   string
	DatabaseURL   string
	SessionSecret string
	SessionName   string
	CORSOrigins   []string
	StaticDir     string
	TemplateDir   string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnvAsInt("PORT", 5000),
		Host:        getEnv("HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "./data/unified.db"),
		SessionSecret: getEnv("SESSION_SECRET", generateDefaultSecret()),
		SessionName:   getEnv("SESSION_NAME", "unified_session"),
		CORSOrigins:   []string{
			getEnv("CORS_ORIGIN", "*"),
		},
		StaticDir:   getEnv("STATIC_DIR", "./static"),
		TemplateDir: getEnv("TEMPLATE_DIR", "./templates"),
	}

	// Validate required fields
	if cfg.SessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required")
	}

	return cfg, nil
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves environment variable as integer or returns default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// generateDefaultSecret creates a default session secret (should be overridden in production)
func generateDefaultSecret() string {
	return "unified-go-default-secret-change-in-production"
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsStaging returns true if running in staging mode
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}
