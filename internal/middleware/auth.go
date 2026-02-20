package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionKey is the context key for session data
type contextKey string

const SessionContextKey contextKey = "session"

// AuthMiddleware handles session validation
type AuthMiddleware struct {
	store *sessions.CookieStore
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(sessionSecret, sessionName string) *AuthMiddleware {
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	return &AuthMiddleware{
		store: store,
	}
}

// Handler returns the middleware handler function
func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := am.store.Get(r, "unified_session")
		if err != nil {
			// Session error, create new session
			session, _ = am.store.New(r, "unified_session")
		}

		// Add session to request context
		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth middleware ensures user is authenticated
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(SessionContextKey).(*sessions.Session)

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetSession retrieves session from request context
func GetSession(r *http.Request) *sessions.Session {
	if session := r.Context().Value(SessionContextKey); session != nil {
		return session.(*sessions.Session)
	}
	return nil
}

// GetUserID retrieves user ID from session
func GetUserID(r *http.Request) (int, bool) {
	session := GetSession(r)
	if session == nil {
		return 0, false
	}

	if userID, ok := session.Values["user_id"].(int); ok {
		return userID, true
	}

	return 0, false
}

// SetAuthenticated sets authentication status in session
func SetAuthenticated(session *sessions.Session, userID int, username string) {
	session.Values["authenticated"] = true
	session.Values["user_id"] = userID
	session.Values["username"] = username
}

// ClearSession clears authentication from session
func ClearSession(session *sessions.Session) {
	session.Values["authenticated"] = false
	delete(session.Values, "user_id")
	delete(session.Values, "username")
}
