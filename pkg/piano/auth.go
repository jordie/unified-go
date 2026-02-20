package piano

import (
	"net/http"

	"github.com/jgirmay/unified-go/internal/middleware"
)

// AuthMiddleware wraps the global auth middleware for piano-specific use
type PianoAuthMiddleware struct {
	globalAuth *middleware.AuthMiddleware
}

// NewPianoAuthMiddleware creates a new piano auth middleware wrapper
func NewPianoAuthMiddleware(globalAuth *middleware.AuthMiddleware) *PianoAuthMiddleware {
	return &PianoAuthMiddleware{
		globalAuth: globalAuth,
	}
}

// RequireAuth is a middleware that requires user authentication for protected routes
func (pam *PianoAuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from session
		userID, ok := middleware.GetUserID(r)
		if !ok || userID == 0 {
			// Redirect unauthenticated users to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuthJSON is middleware that requires auth and returns JSON error
func (pam *PianoAuthMiddleware) RequireAuthJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from session
		userID, ok := middleware.GetUserID(r)
		if !ok || userID == 0 {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"error": "authentication required",
				"status": 401,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromRequest extracts user ID from the session
func GetUserIDFromRequest(r *http.Request) int {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		return 0
	}
	return userID
}

// ErrNotAuthenticated indicates the user is not authenticated
var ErrNotAuthenticated = NewPianoError(401, "user must be authenticated")

// PianoError represents a piano app error
type PianoError struct {
	Code    int
	Message string
}

// NewPianoError creates a new piano error
func NewPianoError(code int, message string) *PianoError {
	return &PianoError{
		Code:    code,
		Message: message,
	}
}

// Error implements the error interface
func (pe *PianoError) Error() string {
	return pe.Message
}
