package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
func NewUserService(redirectTokenURI, intraUserInfoURL string) *UserService{
	return &UserService{
		RedirectTokenURI: redirectTokenURI,
		IntraUserInfoURL: intraUserInfoURL,
	}
}

// TODO look into timeout and context
func (u *UserService) Get42UserData(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (*User42, error) {

	// A specialized HTTP client that handles the Authorization header
	// and token refreshing automatically.
	client := oauthConfig.Client(ctx, token)

	res, err := client.Get(u.IntraUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from intra: %s", err)
	}
	// close response body
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

// func getOrCreateUser(ctx context.Context, user42 User42) {

// }
