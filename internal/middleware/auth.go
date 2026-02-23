package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jgirmay/GAIA_GO/internal/session"
)

// AuthContext keys for storing authentication data in Gin context
const (
	SessionKey     = "session"
	UserIDKey      = "user_id"
	UsernameKey    = "username"
	AuthHeaderName = "X-Session-ID"
)

// AuthMiddleware creates a middleware that enforces session validation
func AuthMiddleware(sessionManager *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from cookie or header
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			sessionID = c.GetHeader(AuthHeaderName)
		}

		if sessionID == "" {
			// Not authenticated - allow to continue but mark as unauthenticated
			c.Set(UserIDKey, int64(0))
			c.Set(UsernameKey, "Guest")
			c.Next()
			return
		}

		// Validate session
		sess, err := sessionManager.GetSession(sessionID)
		if err != nil {
			// Session invalid or expired
			c.SetCookie("session_id", "", -1, "/", "", false, true)
			c.Set(UserIDKey, int64(0))
			c.Set(UsernameKey, "Guest")
			c.Next()
			return
		}

		// Update last activity
		_ = sessionManager.UpdateLastActivity(sessionID)

		// Store in context
		c.Set(SessionKey, sess)
		c.Set(UserIDKey, sess.UserID)
		c.Set(UsernameKey, sess.Username)

		c.Next()
	}
}

// RequireAuth creates a middleware that requires authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(UserIDKey)
		if !exists || userID.(int64) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// OptionalAuth allows both authenticated and unauthenticated access but sets context
func OptionalAuth(sessionManager *session.Manager) gin.HandlerFunc {
	return AuthMiddleware(sessionManager)
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) (int64, error) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, errors.New("user id not in context")
	}
	id, ok := userID.(int64)
	if !ok {
		return 0, errors.New("invalid user id type")
	}
	if id == 0 {
		return 0, errors.New("not authenticated")
	}
	return id, nil
}

// GetUsername retrieves the authenticated username from context
func GetUsername(c *gin.Context) (string, error) {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return "", errors.New("username not in context")
	}
	name, ok := username.(string)
	if !ok {
		return "", errors.New("invalid username type")
	}
	return name, nil
}

// GetSession retrieves the session from context
func GetSession(c *gin.Context) (*session.Session, error) {
	sess, exists := c.Get(SessionKey)
	if !exists {
		return nil, errors.New("session not in context")
	}
	s, ok := sess.(*session.Session)
	if !ok {
		return nil, errors.New("invalid session type")
	}
	return s, nil
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return false
	}
	return userID.(int64) != 0
}

// SetAuthCookie sets the session cookie for authentication
func SetAuthCookie(c *gin.Context, sessionID string, maxAge int) {
	c.SetCookie(
		"session_id",
		sessionID,
		maxAge,
		"/",
		"",
		false, // secure - set to true in production
		true,  // httpOnly
	)
}

// ClearAuthCookie removes the authentication cookie
func ClearAuthCookie(c *gin.Context) {
	c.SetCookie(
		"session_id",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
}
