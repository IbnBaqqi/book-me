package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
)

func TestAuthenticate(t *testing.T) {
	// Create auth service
	authService := auth.NewService("test-secret-key")

	// Create a test user and token
	testUser := database.User{
		ID:   123,
		Name: "Test User",
		Role: "STUDENT",
	}
	token, err := authService.IssueAccessToken(testUser)
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(user.Name))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("anonymous"))
		}
	})

	// Wrap with auth middleware
	authMiddleware := Authenticate(authService)
	wrappedHandler := authMiddleware(handler)

	t.Run("allows request without token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "anonymous" {
			t.Errorf("expected 'anonymous', got %s", w.Body.String())
		}
	})

	t.Run("authenticates valid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "Test User" {
			t.Errorf("expected 'Test User', got %s", w.Body.String())
		}
	})

	t.Run("allows request with invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "anonymous" {
			t.Errorf("expected 'anonymous', got %s", w.Body.String())
		}
	})

	t.Run("allows request with malformed auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "NotBearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "anonymous" {
			t.Errorf("expected 'anonymous', got %s", w.Body.String())
		}
	})
}

func TestRequireAuth(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("protected"))
	})

	// Wrap with RequireAuth
	wrappedHandler := RequireAuth(handler)

	t.Run("blocks unauthenticated request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}
	})

	t.Run("allows authenticated request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		// Add user to context
		user := auth.User{
			ID:   123,
			Name: "Test User",
			Role: "STUDENT",
		}
		ctx := auth.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "protected" {
			t.Errorf("expected 'protected', got %s", w.Body.String())
		}
	})
}

func TestAuthenticateAndRequireAuth(t *testing.T) {
	// Create auth service
	authService := auth.NewService("test-secret-key")

	// Create a test user and token
	testUser := database.User{
		ID:   456,
		Name: "Jane Doe",
		Role: "STAFF",
	}
	token, err := authService.IssueAccessToken(testUser)
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := auth.UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(user.Role))
	})

	// Chain both middlewares
	authMiddleware := Authenticate(authService)
	wrappedHandler := authMiddleware(RequireAuth(handler))

	t.Run("full auth flow with valid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "STAFF" {
			t.Errorf("expected 'STAFF', got %s", w.Body.String())
		}
	})

	t.Run("full auth flow without token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})
}
