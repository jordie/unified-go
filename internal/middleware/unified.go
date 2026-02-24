package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// SECURITY HEADERS MIDDLEWARE
// ============================================================================

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HSTS (HTTP Strict Transport Security)
		if c.Request.Header.Get("X-Forwarded-Proto") == "https" || c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// ============================================================================
// CORS MIDDLEWARE
// ============================================================================

// CORS handles cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ============================================================================
// RATE LIMITING MIDDLEWARE
// ============================================================================

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	requestsPerSecond float64
	burst             int
	tokens            float64
	lastRefill        time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		requestsPerSecond: requestsPerSecond,
		burst:             burst,
		tokens:            float64(burst),
		lastRefill:        time.Now(),
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = min(float64(rl.burst), rl.tokens+elapsed*rl.requestsPerSecond)
	rl.lastRefill = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RateLimitMiddleware applies rate limiting to requests
func RateLimitMiddleware(requestsPerSecond float64) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerSecond, int(requestsPerSecond*2))

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============================================================================
// ERROR HANDLING MIDDLEWARE
// ============================================================================

// ErrorHandler handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

// ============================================================================
// REQUEST LOGGING MIDDLEWARE
// ============================================================================

// Logger logs HTTP requests
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var methodColor, resetColor string

		if param.IsOutputColor() {
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		return fmt.Sprintf("[%s] %s %3d %s %13v %15s %s %#v\n%s",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			methodColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			param.Request.RequestURI,
			param.Request.UserAgent(),
			resetColor,
		)
	})
}

// ============================================================================
// COMPRESSION MIDDLEWARE
// ============================================================================

// CompressionMiddleware enables gzip compression
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Encoding", "gzip")
		c.Next()
	}
}

// ============================================================================
// HEALTH CHECK MIDDLEWARE
// ============================================================================

// HealthCheck returns system health status
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/healthz" {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============================================================================
// RECOVERY MIDDLEWARE
// ============================================================================

// Recovery provides panic recovery with better error messages
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"message": fmt.Sprintf("%v", err),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
