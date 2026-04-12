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
	"github.com/gorilla/sessions"
)

// Service orchestrates the OAuth authentication flow.
type Service struct {
	provider42  *Provider42
	providerKey *ProviderKeycloak
}

// NewService creates a new OAuth service.
func NewService(provider42 *Provider42, providerKey *ProviderKeycloak) *Service {
	return &Service{
		provider42:  provider42,
		providerKey: providerKey,
	}
}

const sessionName = "bookme-session"

// Initiate42Login generates a state token and returns the 42 OAuth authorization URL.
func (s *Service) Initiate42Login(w http.ResponseWriter, r *http.Request) (string, error) {
	state := generateRandomState()

	session, err := s.provider42.session.Get(r, sessionName)
	if err != nil {
		return "", ErrOAuthSessionFailed
	}
	session.Values["oauth_state"] = state
	if err := session.Save(r, w); err != nil {
		slog.Error("failed to save session", "error", err)
		return "", ErrFailedToSaveSession
	}

	return s.provider42.config.AuthCodeURL(state), nil
}

// Handle42Callback handles the 42 OAuth callback.
func (s *Service) Handle42Callback(r *http.Request) (database.User, error) {
	token, err := s.provider42.ExchangeCode(r)
	if err != nil {
		slog.Error("oauth code exchange failed", "error", err)
		return database.User{}, ErrOAuthExchangeFailed
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	user42, err := s.provider42.Fetch42UserData(ctx, s.provider42.config, token)
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

	user, err := s.provider42.FindOrCreateUser(r.Context(), user42)
	if err != nil {
		slog.Error("unable to find or create user", "error", err)
		return database.User{}, ErrFailedToFindorCreateUser
	}

	return user, nil
}

// InitiateKeycloakLogin generates a state token and returns the Keycloak authorization URL.
func (s *Service) InitiateKeycloakLogin(w http.ResponseWriter, r *http.Request) (string, error) {
	state := generateRandomState()

	session, err := s.providerKey.session.Get(r, sessionName)
	if err != nil {
		return "", ErrOAuthSessionFailed
	}
	session.Values["oauth_state"] = state
	if err := session.Save(r, w); err != nil {
		slog.Error("failed to save session", "error", err)
		return "", ErrFailedToSaveSession
	}

	return s.providerKey.config.AuthCodeURL(state), nil
}

// HandleKeycloakCallback handles the Keycloak OIDC callback.
func (s *Service) HandleKeycloakCallback(r *http.Request) (database.User, error) {
	token, err := s.providerKey.ExchangeCode(r)
	if err != nil {
		slog.Error("keycloak code exchange failed", "error", err)
		return database.User{}, ErrOAuthExchangeFailed
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	claims, err := s.providerKey.FetchUserInfo(ctx, token)
	if err != nil {
		slog.Error("failed to fetch keycloak user info", "error", err)
		return database.User{}, ErrOAuthUserInfoFailed
	}

	user, err := s.providerKey.FindOrCreateUser(r.Context(), claims)
	if err != nil {
		slog.Error("unable to find or create keycloak user", "error", err)
		return database.User{}, ErrFailedToFindorCreateUser
	}

	return user, nil
}

// ValidateState checks CSRF protection state.
// It checks both provider sessions since only one will have the state set.
func (s *Service) ValidateState(w http.ResponseWriter, r *http.Request) error {
	// Try 42 session first, then Keycloak
	for _, store := range []*sessions.CookieStore{s.provider42.session, s.providerKey.session} {
		session, err := store.Get(r, sessionName)
		if err != nil {
			continue
		}
		expectedState, ok := session.Values["oauth_state"].(string)
		if !ok || expectedState == "" {
			continue
		}
		if expectedState != r.URL.Query().Get("state") {
			return ErrInvalidOrMissingState
		}
		delete(session.Values, "oauth_state")
		if err := session.Save(r, w); err != nil {
			slog.Error("failed to save session", "error", err)
			return ErrFailedToSaveSession
		}
		return nil
	}

	slog.Error("invalid or missing state")
	return ErrInvalidOrMissingState
}

// GetRedirectTokenURL returns the OAuth provider redirect URL.
func (s *Service) GetRedirectTokenURL() string {
	return s.provider42.redirectTokenURL
}

// generateRandomState generates a cryptographically secure random state token.
func generateRandomState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
