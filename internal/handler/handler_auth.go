package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/IbnBaqqi/book-me/internal/database"
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

	// Check State (CSRF Protection)
	session, _ := h.session.Get(r, sessionName)

	expectedState, ok := session.Values["oauth_state"].(string) // get saved state and compare with incoming state
	if !ok || expectedState != r.URL.Query().Get("state") {
		respondWithError(w, http.StatusForbidden, "Invalid or missing state — possible CSRF attack.", nil)
		return
	}

	// remove session
	delete(session.Values, "oauth_state")
	session.Save(r, w)

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		respondWithError(w, http.StatusBadRequest, "Missing code in callback", nil)
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token exchange failed", err)
		return
	}

	// Get loggedIn User Info from 42
	user42, err := h.userService.Fetch42UserData(r.Context(), h.oauthConfig, token)
	if err != nil {
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

	// Find or create user
	user, err := h.db.GetUserByEmail(r.Context(), user42.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // if user doesn't exist

			role := "STUDENT"
			if user42.Staff {
				role = "STAFF"
			}
			// create user
			newUser, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
				Email: user42.Email,
				Name:  user42.Name,
				Role:  role,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
				return
			}
			user = newUser
		} else {
			// Actual database error
			respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
			return
		}
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

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b) // no error check as Read always succeeds
	return hex.EncodeToString(b)
}
