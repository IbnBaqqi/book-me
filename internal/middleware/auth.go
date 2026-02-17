package middleware

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/IbnBaqqi/book-me/internal/auth"
)

// Authenticate extracts and validates JWT token, adding user to context if valid.
// Allows request to continue even without valid token (for public endpoints).
func Authenticate(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := auth.GetBearerToken(r.Header)
			if err != nil {
				// No token or malformed - continue without auth
				slog.Debug("no auth token", "path", r.URL.Path, "error", err.Error())
				next.ServeHTTP(w, r)
				return
			}

			claims, err := authService.VerifyAccessToken(tokenStr)
			if err != nil {
				slog.Warn("invalid auth token", "path", r.URL.Path, "error", err.Error())
				next.ServeHTTP(w, r)
				return
			}

			// Parse user ID from claims
			id, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				slog.Error("invalid user ID in token", "subject", claims.Subject, "error", err.Error())
				next.ServeHTTP(w, r)
				return
			}

			user := auth.User{
				ID:   id,
				Role: claims.Role,
				Name: claims.Name,
			}

			slog.Debug("authenticated request", "user_id", user.ID, "role", user.Role, "path", r.URL.Path)
			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth ensures user is authenticated, returning 401 if not.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			slog.Warn("unauthorized access attempt", "path", r.URL.Path, "method", r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		slog.Debug("authorized request", "user_id", user.ID, "role", user.Role, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
