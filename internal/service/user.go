package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/logger"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

// User42 holds user data returned from 42
type User42 struct {
	Email  string        `json:"email"`
	Name   string        `json:"login"`
	Staff  bool          `json:"staff?"`
	Campus []CampusUsers `json:"campus_users"`
}

type CampusUsers struct {
	ID      int  `json:"campus_id"`
	Primary bool `json:"is_primary"`
}

type UserService struct {
	db               *database.Queries
	RedirectTokenURI string
	IntraUserInfoURL string
}

// NewUserService create a new user service
func NewUserService(db *database.Queries, redirectTokenURI, intraUserInfoURL string) *UserService {
	return &UserService{
		db:               db,
		RedirectTokenURI: redirectTokenURI,
		IntraUserInfoURL: intraUserInfoURL,
	}
}

func (u *UserService) Fetch42UserData(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (*User42, error) {

	// A specialized HTTP client that handles the Authorization header and token refreshing automatically.
	oauthClient := oauthConfig.Client(ctx, token)

	// Create retry client for this request to avoid race condition
	retryClient := u.newRetryClient(oauthClient.Transport)

	// Create retryable request
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", u.IntraUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// RetryClient handles retries on 429 & 5xx reponse exponentially
	res, err := retryClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from intra: %w", err)
	}
	defer res.Body.Close()

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

// newRetryClient creates a retry client with the OAuth2 transport
func (u *UserService) newRetryClient(oauthTransport http.RoundTripper) *retryablehttp.Client {
	client := retryablehttp.NewClient()

	client.RetryMax =      3
	client.RetryWaitMin =  1 * time.Second
	client.RetryWaitMax =  5 * time.Second
	client.Logger =        &logger.RetryLogger{}
	client.Backoff =       retryablehttp.DefaultBackoff
	client.HTTPClient.Transport = oauthTransport
	
	return client
}

// findOrCreateUser gets existing user or creates new one
func (s *UserService) FindOrCreateUser(ctx context.Context, user42 *User42) (database.User, error) {
	// Try to find existing user
	user, err := s.db.GetUserByEmail(ctx, user42.Email)
	if err == nil {
		return user, nil 
	}

	// If not found, create new user
	if errors.Is(err, sql.ErrNoRows) {
		return s.createUser(ctx, user42)
	}

	// Database error
	return database.User{}, fmt.Errorf("database error: %w", err)
}

// createUser creates a new user from 42 data
func (s *UserService) createUser(ctx context.Context, user42 *User42) (database.User, error) {
	role := "STUDENT"
	if user42.Staff {
		role = "STAFF"
	}

	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		Email: user42.Email,
		Name:  user42.Name,
		Role:  role,
	})
	if err != nil {
		return database.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}