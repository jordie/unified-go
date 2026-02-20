.PHONY: help build run test clean install dev docker

# Default target
help:
	@echo "Unified Educational Platform - Go Edition"
	@echo ""
	@echo "Available targets:"
	@echo "  make build      - Build the binary"
	@echo "  make run        - Run the server"
	@echo "  make dev        - Run in development mode with auto-reload"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make install    - Install dependencies"
	@echo "  make db-reset   - Reset the database"
	@echo "  make docker     - Build Docker image (Phase 3)"

# Build the application
build:
	@echo "Building unified-go..."
	go build -o unified-go cmd/server/main.go
	@echo "Build complete: ./unified-go"

# Build with optimizations for production
build-prod:
	@echo "Building for production..."
	go build -ldflags="-s -w" -o unified-go cmd/server/main.go
	@echo "Production build complete: ./unified-go"

# Run the server
run: build
	@echo "Starting server..."
	./unified-go

# Development mode (requires air for auto-reload)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without auto-reload..."; \
		go run cmd/server/main.go; \
	fi

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f unified-go
	rm -f coverage.out coverage.html
	rm -rf tmp/
	@echo "Clean complete"

# Reset database (WARNING: deletes all data)
db-reset:
	@echo "WARNING: This will delete all database data!"
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		rm -rf data/*.db*; \
		echo "Database reset complete"; \
	else \
		echo "Cancelled"; \
	fi

# Check database tables
db-tables:
	@echo "Database tables:"
	@sqlite3 data/unified.db ".tables"

# Show database schema
db-schema:
	@echo "Database schema:"
	@sqlite3 data/unified.db ".schema"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed"; \
		echo "Install: brew install golangci-lint"; \
	fi

# Security check (requires gosec)
security:
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not installed"; \
		echo "Install: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Docker build (Phase 3)
docker:
	@echo "Docker support coming in Phase 3"

# Show current environment
env:
	@echo "Current environment:"
	@echo "  PORT: $${PORT:-5000}"
	@echo "  ENVIRONMENT: $${ENVIRONMENT:-development}"
	@echo "  DATABASE_URL: $${DATABASE_URL:-./data/unified.db}"

# Quick health check
health:
	@echo "Checking server health..."
	@curl -s http://localhost:5000/health | jq . || echo "Server not responding"
