package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// Login handles user login / sign-in
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	// Initiate oauth2 flow (login)
	url, err := h.oauth.InitiateLogin(w, r)
	if err != nil {
		handleError(w, err)
	}

	// redirect to oauth2 authorization server
	http.Redirect(w, r, url, http.StatusFound)
}

// Callback handles callback from Oauth flow
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {

	// Validate CSRF state
	if err := h.oauth.ValidateState(w, r); err != nil {
		handleError(w, err)
		return
	}

	// Oauth 42 service handles callback
	user, err := h.oauth.HandleCallback(r)
	if err != nil {
		handleError(w, err)
		return
	}

	// Issue jwt
	jwtToken, err := h.auth.IssueAccessToken(user)
	if err != nil {
		slog.Error("Failed to generate token", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	params := url.Values{}
	params.Add("token", jwtToken)
	params.Add("intra", user.Name)
	params.Add("role", strings.ToLower(user.Role))

	// final redirect
	finalRedirectURL := fmt.Sprintf("%s?%s", h.oauth.GetRedirectTokenURL(), params.Encode())
	http.Redirect(w, r, finalRedirectURL, http.StatusFound)
}
