package oauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/logger"
	"github.com/IbnBaqqi/book-me/internal/service"
	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

// User42 represent the user data response from 42 user info endpoint
type User42 struct {
	Email  string        `json:"email"`
	Name   string        `json:"login"`
	Staff  bool          `json:"staff?"`
	Campus []CampusUsers `json:"campus_users"`
}

// CampusUsers represents User42 campus information
type CampusUsers struct {
	ID      int  `json:"campus_id"`
	Primary bool `json:"is_primary"`
}

// Provider42 represents the info & dependencies for 42 OAuth
type Provider42 struct {
	db               *database.DB
	config           *oauth2.Config
	session          *sessions.CookieStore
	redirectTokenURL string
	userInfoURL      string
}

// NewProvider42 creates a new 42 OAuth provider
func NewProvider42(
	db *database.DB,
	config *oauth2.Config,
	sessionSecret string,
	redirectTokenURL string,
	userInfoURL string,
) *Provider42 {

	return &Provider42{
		db:               db,
		config:           config,
		session:          sessions.NewCookieStore([]byte(sessionSecret)),
		redirectTokenURL: redirectTokenURL,
		userInfoURL:      userInfoURL,
	}
}

// ExchangeCode exchanges OAuth code for access token
func (p *Provider42) ExchangeCode(r *http.Request) (*oauth2.Token, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, errors.New("missing authorization code")
	}

	return p.config.Exchange(r.Context(), code)
}

// Fetch42UserData fetch user data from 42 intranet
func (p *Provider42) Fetch42UserData(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (*User42, error) {

	// A specialized HTTP client that handles the Authorization header and token refreshing automatically.
	oauthClient := oauthConfig.Client(ctx, token)

	// Create retry client for request
	retryClient := p.NewRetryClient(oauthClient.Transport)

	// Create retryable request
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", p.userInfoURL, nil)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request") // change
	}

	// RetryClient handles retries on 429 & 5xx response exponentially
	res, err := retryClient.Do(req)
	if err != nil {
		slog.Error("failed to fetch user data from 42 intra", "err", err)
		return nil, fmt.Errorf("failed to fetch user data from 42 intra")
	}
	
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("intra api returned error status: %s", res.Status)
	}

	var user42 User42
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&user42)
	if err != nil {
		return nil, fmt.Errorf("failed to decode intra user data: %w", err)
	}

	return &user42, nil
}

// NewRetryClient creates a retry client with the OAuth2 transport
func (p *Provider42) NewRetryClient(oauthTransport http.RoundTripper) *retryablehttp.Client {
	client := retryablehttp.NewClient()

	client.RetryMax = 3
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 5 * time.Second
	client.Logger = &logger.RetryLogger{}
	client.Backoff = retryablehttp.DefaultBackoff
	client.HTTPClient.Transport = oauthTransport

	return client
}

// FindOrCreateUser gets existing user or creates new one
func (p *Provider42) FindOrCreateUser(ctx context.Context, user42 *User42) (database.User, error) {
	// Try to find existing user
	user, err := p.db.GetUserByEmail(ctx, user42.Email)
	if err == nil {
		return user, nil
	}

	// If not found, create new user
	if errors.Is(err, sql.ErrNoRows) {
		return p.createUser(ctx, user42)
	}

	// Database error
	return database.User{}, fmt.Errorf("database error: %w", err)
}

// createUser creates a new user from 42 data
func (p *Provider42) createUser(ctx context.Context, user42 *User42) (database.User, error) {
	role := service.RoleStudent
	if user42.Staff {
		role = service.RoleStaff
	}

	user, err := p.db.CreateUser(ctx, database.CreateUserParams{
		Email: user42.Email,
		Name:  user42.Name,
		Role:  role,
	})
	if err != nil {
		return database.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
