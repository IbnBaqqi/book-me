package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	RedirectTokenURI string
	IntraUserInfoURL string
	RetryClient      *retryablehttp.Client
}

// NewUserService create a new user service
func NewUserService(redirectTokenURI, intraUserInfoURL string) *UserService {
	retryClient := retryablehttp.NewClient()

	// Configure retry behavior
	retryClient.RetryMax = 3                   // Maximum number of retries
	retryClient.RetryWaitMin = 1 * time.Second // Minimum wait between retries
	retryClient.RetryWaitMax = 5 * time.Second // Maximum wait between retries
	retryClient.Logger = nil

	// Custom retry policy, don't retry on 4xx errors
	retryClient.CheckRetry = customRetryPolicy

	// Use custom logger adapter
	retryClient.Logger = &logger.RetryLogger{}

	// Default backoff strategy is exponential
	retryClient.Backoff = retryablehttp.DefaultBackoff

	return &UserService{
		RedirectTokenURI: redirectTokenURI,
		IntraUserInfoURL: intraUserInfoURL,
		RetryClient:      retryClient,
	}
}

// TODO look into timeout and context
func (u *UserService) Fetch42UserData(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (*User42, error) {

	// A specialized HTTP client that handles the Authorization header and token refreshing automatically.
	standardClient := oauthConfig.Client(ctx, token)

	// Wrap the OAuth2 transport with retryable client
	u.RetryClient.HTTPClient = standardClient

	// Create retryable request
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", u.IntraUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := u.RetryClient.Do(req)
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

// customRetryPolicy determines which errors/status codes should be retried
func customRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// Always retry on connection errors
	if err != nil {
		return true, err
	}

	// Don't retry on client errors (4xx)
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return false, nil
	}

	// Retry on server errors (5xx) and other non-2xx responses
	if resp.StatusCode == 0 || resp.StatusCode >= 500 {
		return true, nil
	}

	// Use default policy for everything else
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}

// func getOrCreateUser(ctx context.Context, user42 User42) {

// }
