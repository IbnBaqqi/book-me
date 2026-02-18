// Package oauth provides OAuth authentication flow.
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/database"
)

// Service orchestrates the OAuth authentication flow.
type Service struct {
	provider *Provider42
}

// NewService creates a new OAuth service for a provider.
func NewService(provider *Provider42) *Service {

	return &Service{
		provider: provider,
	}
}

const sessionName = "bookme-session"

// InitiateLogin generates a state token and returns the OAuth authorization URL.
func (s *Service) InitiateLogin(w http.ResponseWriter, r *http.Request) (string, error) {

	state := generateRandomState()

	// Store state in session to prevent CSRF
	session, err := s.provider.session.Get(r, sessionName)
	if err != nil {
		return "", ErrOAuthSessionFailed
	}
	session.Values["oauth_state"] = state
	if err := session.Save(r, w); err != nil {
		slog.Error("failed to save session", "error", err)
		return "", &OauthError{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	url := s.provider.config.AuthCodeURL(state)

	return url, nil
}

// HandleCallback is a service layer function that handles Oauth callback.
func (s *Service) HandleCallback(r *http.Request) (database.User, error) {

	// Exchange authorization code for token
	token, err := s.provider.ExchangeCode(r)
	if err != nil {
		return database.User{}, ErrOAuthExchangeFailed
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	user42, err := s.provider.Fetch42UserData(ctx, s.provider.config, token)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return database.User{}, ErrOAuthTimeout
		}
		slog.Error("failed to fetch user data from 42 intra", "error", err)
		return database.User{}, ErrOAuthUserInfoFailed
	}

	// Validate Campus
	isHive := false
	for _, camp := range user42.Campus {
		if camp.ID == 13 && camp.Primary {
			isHive = true
			break
		}
	}
	if !isHive {
		return database.User{}, ErrInvalidCampus
	}

	// Find or create user - might use redis later for this TODO
	user, err := s.provider.FindOrCreateUser(r.Context(), user42)
	if err != nil {
		slog.Error("Unable to Find or create user", "error", err)
		return database.User{}, &OauthError{
			Message: "failed to find or create user",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return user, nil
}

// ValidateState checks CSRF protection state.
func (s *Service) ValidateState(w http.ResponseWriter, r *http.Request) error {
	session, _ := s.provider.session.Get(r, sessionName)

	// get saved state and compare with incoming state
	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || expectedState != r.URL.Query().Get("state") {
		slog.Error("invalid or missing state")
		return &OauthError{
			Err:        errors.New("invalid or missing state"),
			StatusCode: http.StatusForbidden,
		}
	}

	delete(session.Values, "oauth_state")
	if err := session.Save(r, w); err != nil {
		slog.Error("failed to save session", "error", err)
		return &OauthError{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	return nil
}

// GetRedirectTokenURL returns the OAuth provider redirect URL.
func (s *Service) GetRedirectTokenURL() string {
	return s.provider.redirectTokenURL
}

// generateRandomState generates a cryptograph secure random state token.
func generateRandomState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
