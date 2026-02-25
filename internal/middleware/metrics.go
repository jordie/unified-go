package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/metrics"
)

// bodyLogWriter wraps gin.ResponseWriter to capture response body size
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write implements io.Writer for capturing response body
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// MetricsMiddleware creates a middleware that records HTTP request metrics
func MetricsMiddleware(registry *metrics.HTTPMetricsRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time and increment active requests
		start := time.Now()
		registry.IncrementActiveRequests()
		defer registry.DecrementActiveRequests()

		// Get request info
		method := c.Request.Method
		path := c.FullPath() // Use template path like /api/typing/users/:id
		reqSize := c.Request.ContentLength

		// Wrap response writer to capture response size
		blw := &bodyLogWriter{body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = blw

		// Call next handler
		c.Next()

		// Calculate metrics
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		respSize := int64(blw.body.Len())
		appName := extractAppName(path)

		// Record metrics
		registry.RecordRequest(method, path, status, appName, duration, reqSize, respSize)
	}
}

// extractAppName extracts the app name from the request path
// Examples:
//   - /api/math/problems -> "math"
//   - /api/typing/users/:id -> "typing"
//   - /health -> "core"
//   - /api/users -> "core"
func extractAppName(path string) string {
	if path == "" || path == "/" {
		return "core"
	}

	// Handle /metrics endpoint
	if path == "/metrics" {
		return "core"
	}

	// Handle /api/{app}/... paths
	if strings.HasPrefix(path, "/api/") {
		parts := strings.Split(path[5:], "/") // Remove "/api/" prefix
		if len(parts) > 0 && parts[0] != "" {
			// Check if it's a known app
			app := parts[0]
			switch app {
			case "math", "typing", "reading", "piano":
				return app
			case "users", "auth", "health":
				return "core"
			default:
				return "core"
			}
		}
	}

	return "core"
}

// readBodyAndRestore reads request body and restores it for handlers
// This is a helper for capturing request body in middleware if needed
func readBodyAndRestore(c *gin.Context) []byte {
	body := c.Request.Body
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil
	}
	// Restore body for subsequent handlers
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes
}
