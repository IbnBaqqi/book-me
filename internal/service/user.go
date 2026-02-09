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
}

// NewUserService create a new user service
func NewUserService(redirectTokenURI, intraUserInfoURL string) *UserService {
	return &UserService{
		RedirectTokenURI: redirectTokenURI,
		IntraUserInfoURL: intraUserInfoURL,
	}
}

// TODO look into timeout and context
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
