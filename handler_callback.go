package main

import "net/http"


func (cfg *apiConfig) handlerCallback(w http.ResponseWriter, r *http.Request) {

	// Check State (CSRF Protection)
	session, _ := cfg.sessionStore.Get(r, sessionName)

	expectedState, ok := session.Values["oauth_state"].(string) // get savedState
	if !ok || expectedState != r.URL.Query().Get("state") {
		respondWithError(w, http.StatusForbidden, "Invalid or missing state — possible CSRF attack.", nil)
		return
	}

	// remove session
	delete(session.Values, "oauth_state")
	session.Save(r, w)

	// Exchange code for token
	code := r.URL.Query().Get("code")

	token, err := cfg.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token exchange failed", err)
		return
	}
}