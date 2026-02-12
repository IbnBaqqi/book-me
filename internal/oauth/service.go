package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/database"
)

// Oauth Service orchestrates the OAuth authentication flow
type Service struct {
	provider   *Provider42
}

// NewService creates a new OAuth service
func NewService(provider *Provider42) *Service {
	
	return &Service{
		provider:   provider,
	}
}

const sessionName = "bookme-session"

// InitiateLogin generates a state token and returns the OAuth authorization URL
func (s *Service) InitiateLogin(w http.ResponseWriter, r *http.Request) (string, error) {
	
	state := generateRandomState()

	// Store state in session to prevent CSRF
	session, err := s.provider.session.Get(r, sessionName)
	if err != nil {
		return "", ErrOAuthSessionFailed
	}
	session.Values["oauth_state"] = state
	session.Save(r, w)

	url := s.provider.config.AuthCodeURL(state)

	return url, nil
}

// func (s *Service) HandleCallback(ctx context.Context, code, state string) (*UserInfo, error) {

// }

func (s *Service) Handlecallback(r *http.Request) (database.User, error) {

	// Exchange authorization code for token
	token, err := s.provider.ExchangeCode(r)
	if err != nil {
		return database.User{}, ErrOAuthExchangeFailed
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15 * time.Second)
	defer cancel()

	// Get loggedIn User Info from 42
	user42, err := s.provider.Fetch42UserData(ctx, s.provider.config, token)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return database.User{}, ErrOAuthTimeout
		}
		return database.User{}, &OauthError{
			StatusCode: http.StatusBadGateway,
        	Message:    err.Error(),
        	Err:        err,
		}
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
		// TODO need to find a way to return the error and also log like respondWithError
		// respondWithError(w, http.StatusInternalServerError, "failed to get or create user", err)
		return database.User{}, err
	}

	return user, nil
}

// validateState checks CSRF protection state
func (s *Service) ValidateState(w http.ResponseWriter, r *http.Request) error {
	session, _ := s.provider.session.Get(r, sessionName)
	
	// get saved state and compare with incoming state
	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || expectedState != r.URL.Query().Get("state") {
		return errors.New("state mismatch")
	}

	// Clear state from session
	delete(session.Values, "oauth_state")
	session.Save(r, w)
	
	return nil
}

// GetAuthURL returns the OAuth authorization URL with state
func (s *Service) GetRedirectTokenURL() string {
	return s.provider.redirectTokenURI
}

// generateRandomState generates a cryptograph secure random state token
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b) // no error check as Read always succeeds
	return hex.EncodeToString(b)
}