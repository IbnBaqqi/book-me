package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	// clientID     = "YOUR_CLIENT_ID"
	// redirectURI  = "http://localhost:8080/oauth/callback"
	// oauthAuthURL = "https://api.intra.42.fr/oauth/authorize"
	sessionName  = "bookme-session"
)


func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {

	state := generateRandomState()

	// Store state in session to prevent CSRF
	session, err := cfg.sessionStore.Get(r, sessionName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get session", err)
		return
	}
	session.Values["oauth_state"] = state
	session.Save(r, w)

	// Redirect to 42 Auth
	url := cfg.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}


func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b) // no error check as Read always succeeds
	return hex.EncodeToString(b)
}
