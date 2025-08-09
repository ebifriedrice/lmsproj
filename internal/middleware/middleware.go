package middleware

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
)

// Middleware struct holds dependencies for middleware.
type Middleware struct {
	SessionManager *scs.SessionManager
}

// NewMiddleware creates a new Middleware struct.
func NewMiddleware(sessionManager *scs.SessionManager) *Middleware {
	return &Middleware{
		SessionManager: sessionManager,
	}
}

// RequireAuthentication is a middleware that requires a user to be authenticated.
func (m *Middleware) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated by looking for a user ID in the session.
		userID := m.SessionManager.GetInt64(r.Context(), "authenticatedUserID")
		if userID == 0 {
			// If the user is not authenticated, redirect them to the login page.
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If the user is authenticated, call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin is a middleware that requires a user to be an admin.
// It should be used after RequireAuthentication.
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user's role from the session.
		role := m.SessionManager.GetString(r.Context(), "userRole")
		if role != "admin" {
			// If the user is not an admin, return a forbidden error.
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// If the user is an admin, call the next handler.
		next.ServeHTTP(w, r)
	})
}
