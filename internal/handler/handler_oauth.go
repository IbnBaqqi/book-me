package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// Login42 handles user login / sign-in
func (h *Handler) Login42(w http.ResponseWriter, r *http.Request) {

	// Initiate oauth2 flow
	url, err := h.oauth.Initiate42Login(w, r)
	slog.Error("got here and url", "error", url)
	if err != nil {
		handleError(w, err)
		return
	}

	// redirect to oauth2 authorization server
	http.Redirect(w, r, url, http.StatusFound)
}

// Callback42 handles callback from Oauth flow
func (h *Handler) Callback42(w http.ResponseWriter, r *http.Request) {

	// Validate CSRF state
	if err := h.oauth.ValidateState(w, r); err != nil {
		handleError(w, err)
		return
	}

	user, err := h.oauth.Handle42Callback(r)
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

// LoginKeycloak handles user Hive's keycloak login / sign-in
func (h *Handler) LoginKeycloak(w http.ResponseWriter, r *http.Request) {

	// Initiate oauth2 flow
	url, err := h.oauth.InitiateKeycloakLogin(w, r)
	slog.Debug("got to initiate key login", "url", url) //remove later
	if err != nil {
		handleError(w, err)
		return
	}

	// redirect to oauth2 authorization server
	http.Redirect(w, r, url, http.StatusFound)
}

// CallbackKeyclok handles callback from Hive's keycloak Oauth flow
func (h *Handler) CallbackKeycloak(w http.ResponseWriter, r *http.Request) {

	slog.Debug("got to callback") // remove later
	// Validate CSRF state
	if err := h.oauth.ValidateState(w, r); err != nil {
		handleError(w, err)
		return
	}

	user, err := h.oauth.HandleKeycloakCallback(r)
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
