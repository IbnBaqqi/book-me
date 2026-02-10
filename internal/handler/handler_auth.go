package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const sessionName = "bookme-session"

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	state := generateRandomState()

	// Store state in session to prevent CSRF
	session, err := h.session.Get(r, sessionName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get session", err)
		return
	}
	session.Values["oauth_state"] = state
	session.Save(r, w)

	// Redirect to 42 Auth
	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {

	// Validate CSRF state
	if err := h.validateState(w, r); err != nil {
		respondWithError(w, http.StatusForbidden, "Invalid or missing state — possible CSRF attack", err)
		return
	}

	// Exchange authorization code for token
	token, err := h.exchangeCode(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token exchange failed", err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15 * time.Second)
	defer cancel()
	// Get loggedIn User Info from 42
	user42, err := h.userService.Fetch42UserData(ctx, h.oauthConfig, token)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			respondWithError(w, http.StatusGatewayTimeout, "Request to 42 API timed out", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get user data from 42", err)
		return
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
		respondWithError(w, http.StatusForbidden, "Access Denied: Only Helsinki Campus Student Allowed", nil)
		return
	}

	// Find or create user - might use redis later for this TODO
	user, err := h.userService.FindOrCreateUser(r.Context(), user42)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get or create user", err)
		return
	}

	// Issue jwt
	jwtToken, err := h.auth.IssueAccessToken(user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	params := url.Values{}
	params.Add("token", jwtToken)
	params.Add("intra", user.Name)
	params.Add("role", strings.ToLower(user.Role))

	// final redirect
	finalRedirectURL := fmt.Sprintf("%s?%s", h.userService.RedirectTokenURI, params.Encode())
	http.Redirect(w, r, finalRedirectURL, http.StatusFound)
}

// validateState checks CSRF protection state
func (h *Handler) validateState(w http.ResponseWriter, r *http.Request) error {
	session, _ := h.session.Get(r, sessionName)
	
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

// exchangeCode exchanges OAuth code for access token
func (h *Handler) exchangeCode(r *http.Request) (*oauth2.Token, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, errors.New("missing authorization code")
	}

	return h.oauthConfig.Exchange(r.Context(), code)
}

// generateRandomState generates random state value
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b) // no error check as Read always succeeds
	return hex.EncodeToString(b)
}
