package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// TODO look into timeout and context
func get42UserData(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (*User42, error) {

	// A specialized HTTP client that handles the Authorization header
	// and token refreshing automatically.
	client := oauthConfig.Client(ctx, token)

	res, err := client.Get("https://api.intra.42.fr/v2/me") //externalize later
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