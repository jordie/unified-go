package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantPort int
		wantEnv  string
	}{
		{
			name:     "default values",
			envVars:  map[string]string{},
			wantPort: 5000,
			wantEnv:  "development",
		},
		{
			name: "custom port",
			envVars: map[string]string{
				"PORT": "8080",
			},
			wantPort: 8080,
			wantEnv:  "development",
		},
		{
			name: "production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
			},
			wantPort: 5000,
			wantEnv:  "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Set required SESSION_SECRET
			os.Setenv("SESSION_SECRET", "test-secret")

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %v, want %v", cfg.Port, tt.wantPort)
			}

			if cfg.Environment != tt.wantEnv {
				t.Errorf("Environment = %v, want %v", cfg.Environment, tt.wantEnv)
			}
		})
	}
}

func TestConfigEnvironmentChecks(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		isDev   bool
		isProd  bool
		isStage bool
	}{
		{
			name:    "development",
			env:     "development",
			isDev:   true,
			isProd:  false,
			isStage: false,
		},
		{
			name:    "production",
			env:     "production",
			isDev:   false,
			isProd:  true,
			isStage: false,
		},
		{
			name:    "staging",
			env:     "staging",
			isDev:   false,
			isProd:  false,
			isStage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}

			if got := cfg.IsDevelopment(); got != tt.isDev {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.isDev)
			}

			if got := cfg.IsProduction(); got != tt.isProd {
				t.Errorf("IsProduction() = %v, want %v", got, tt.isProd)
			}

			if got := cfg.IsStaging(); got != tt.isStage {
				t.Errorf("IsStaging() = %v, want %v", got, tt.isStage)
			}
		})
	}
}
