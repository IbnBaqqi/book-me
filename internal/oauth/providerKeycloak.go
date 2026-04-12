package oauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/service"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// ProviderKeycloak handles Keycloak OIDC authentication.
type ProviderKeycloak struct {
	db               *database.DB
	config           *oauth2.Config
	session          *sessions.CookieStore
	redirectTokenURL string
	userInfoURL      string
}

// keycloakClaims holds the relevant fields from the Keycloak userinfo response.
type keycloakClaims struct {
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
}

// NewProviderKeycloak creates a new Keycloak OIDC provider.
func NewProviderKeycloak(
	db *database.DB,
	config *oauth2.Config,
	sessionSecret string,
	redirectTokenURL string,
	userInfoURL string,
) *ProviderKeycloak {
	return &ProviderKeycloak{
		db:               db,
		config:           config,
		session:          sessions.NewCookieStore([]byte(sessionSecret)),
		redirectTokenURL: redirectTokenURL,
		userInfoURL:      userInfoURL,
	}
}

// ExchangeCode exchanges the OAuth2 authorization code for a token.
func (p *ProviderKeycloak) ExchangeCode(r *http.Request) (*oauth2.Token, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, ErrInvalidOAuthCode
	}
	return p.config.Exchange(r.Context(), code)
}

// FetchUserInfo fetches user claims from the Keycloak userinfo endpoint.
func (p *ProviderKeycloak) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*keycloakClaims, error) {
	client := p.config.Client(ctx, token)

	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call userinfo endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status: %s", resp.Status)
	}

	var claims keycloakClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("userinfo response missing email")
	}
	// fmt.Println(claims)
	return &claims, nil
}

// FindOrCreateUser looks up a user by email or creates one from Keycloak claims.
func (p *ProviderKeycloak) FindOrCreateUser(ctx context.Context, claims *keycloakClaims) (database.User, error) {
	user, err := p.db.GetUserByEmail(ctx, claims.Email)
	if err == nil {
		return user, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		name := claims.PreferredUsername
		if name == "" {
			name = claims.Email
		}
		return p.db.CreateUser(ctx, database.CreateUserParams{
			Email: claims.Email,
			Name:  name,
			Role:  service.RoleStudent,
		})
	}

	return database.User{}, fmt.Errorf("database error: %w", err)
}
