package handler

import (
	// "context"
	// "errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	// "time"

	// "golang.org/x/oauth2"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	url, err := h.oauth.InitiateLogin(w, r)
	if err != nil {
		handleError(w, err)
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {

	// Validate CSRF state
	if err := h.oauth.ValidateState(w, r); err != nil {
		respondWithError(w, http.StatusForbidden, "Invalid or missing state", err)
		return
	}

	// // Exchange authorization code for token
	// token, err := h.exchangeCode(r)
	// if err != nil {
	// 	respondWithError(w, http.StatusUnauthorized, "token exchange failed", err)
	// 	return
	// }

	// ctx, cancel := context.WithTimeout(r.Context(), 15 * time.Second)
	// defer cancel()
	// // Get loggedIn User Info from 42
	// user42, err := h.userService.Fetch42UserData(ctx, h.oauthConfig, token)
	// if err != nil {
	// 	if errors.Is(err, context.DeadlineExceeded) {
	// 		respondWithError(w, http.StatusGatewayTimeout, "Request to 42 API timed out", err)
	// 		return
	// 	}
	// 	respondWithError(w, http.StatusInternalServerError, "Failed to get user data from 42", err)
	// 	return
	// }

	// // Validate Campus
	// isHive := false
	// for _, camp := range user42.Campus {
	// 	if camp.ID == 13 && camp.Primary {
	// 		isHive = true
	// 		break
	// 	}
	// }

	// if !isHive {
	// 	respondWithError(w, http.StatusForbidden, "Access Denied: Only Helsinki Campus Student Allowed", nil)
	// 	return
	// }

	// // Find or create user - might use redis later for this TODO
	// user, err := h.userService.FindOrCreateUser(r.Context(), user42)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "failed to get or create user", err)
	// 	return
	// }

	user, err := h.oauth.Handlecallback(r)
	if err != nil {
		handleError(w, err)
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
	finalRedirectURL := fmt.Sprintf("%s?%s", h.oauth.GetRedirectTokenURL(), params.Encode())
	http.Redirect(w, r, finalRedirectURL, http.StatusFound)
}
