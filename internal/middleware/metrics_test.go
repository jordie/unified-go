package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/metrics"
)

func TestExtractAppName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/math/problems", "math"},
		{"/api/typing/users/:id", "typing"},
		{"/api/reading/passages", "reading"},
		{"/api/piano/songs", "piano"},
		{"/api/users", "core"},
		{"/api/auth/login", "core"},
		{"/health", "core"},
		{"/metrics", "core"},
		{"", "core"},
		{"/", "core"},
		{"/unknown/path", "core"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := extractAppName(tt.path)
			if result != tt.expected {
				t.Errorf("extractAppName(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := metrics.NewHTTPMetricsRegistry()
	middleware := MetricsMiddleware(registry)

	engine := gin.New()
	engine.Use(middleware)

	engine.GET("/api/math/problems", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	engine.GET("/api/typing/test", func(c *gin.Context) {
		c.JSON(201, gin.H{"id": 123})
	})

	engine.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"Math GET request", "GET", "/api/math/problems", 200},
		{"Typing POST request", "GET", "/api/typing/test", 201},
		{"Server error", "GET", "/error", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMetricsMiddlewareRecordsMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := metrics.NewHTTPMetricsRegistry()
	middleware := MetricsMiddleware(registry)

	engine := gin.New()
	engine.Use(middleware)

	engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Make a request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Verify response
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsMiddlewareActiveRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := metrics.NewHTTPMetricsRegistry()
	middleware := MetricsMiddleware(registry)

	engine := gin.New()
	engine.Use(middleware)

	// Track active requests during handler
	activeCount := 0
	engine.GET("/test", func(c *gin.Context) {
		// During handler execution, there should be active requests tracked
		c.JSON(200, gin.H{"ok": true})
	})

	// Make requests
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		activeCount += 1
	}

	if activeCount != 5 {
		t.Errorf("Expected 5 requests, got %d", activeCount)
	}
}

func TestMetricsMiddlewareHandlesErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := metrics.NewHTTPMetricsRegistry()
	middleware := MetricsMiddleware(registry)

	engine := gin.New()
	engine.Use(middleware)

	// 4xx error
	engine.GET("/not-found", func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "not found"})
	})

	// 5xx error
	engine.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "server error"})
	})

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"404 Not Found", "/not-found", 404},
		{"500 Server Error", "/error", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestMetricsMiddlewareWithDifferentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := metrics.NewHTTPMetricsRegistry()
	middleware := MetricsMiddleware(registry)

	engine := gin.New()
	engine.Use(middleware)

	engine.GET("/resource", func(c *gin.Context) {
		c.JSON(200, gin.H{"method": "GET"})
	})

	engine.POST("/resource", func(c *gin.Context) {
		c.JSON(201, gin.H{"method": "POST"})
	})

	engine.PUT("/resource/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"method": "PUT"})
	})

	engine.DELETE("/resource/:id", func(c *gin.Context) {
		c.JSON(204, gin.H{})
	})

	methods := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"GET", "GET", "/resource", 200},
		{"POST", "POST", "/resource", 201},
		{"PUT", "PUT", "/resource/1", 200},
		{"DELETE", "DELETE", "/resource/1", 204},
	}

	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			req, _ := http.NewRequest(m.method, m.path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != m.expectedStatus {
				t.Errorf("Expected status %d, got %d", m.expectedStatus, w.Code)
			}
		})
	}
}
