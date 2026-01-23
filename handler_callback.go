package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
)

func (cfg *apiConfig) handlerCallback(w http.ResponseWriter, r *http.Request) {

	// Check State (CSRF Protection)
	session, _ := cfg.sessionStore.Get(r, sessionName)

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

	token, err := cfg.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token exchange failed", err)
		return
	}

	// Get loggedIn User Info from 42
	user42, err := get42UserData(r.Context(), cfg.oauthConfig, token)
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
        respondWithError(w, http.StatusForbidden, "Access Denied: Only Helsinki Campus Student Allowed", nil) //log that another campus tried to login
        return
    }

	// Find or create user
	user, err := cfg.db.GetUserByEmail(r.Context(), user42.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // if user doesn't exist

			role := "STUDENT"
			if user42.Staff {
				role = "STAFF"
			}
			// create user
			newUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
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
            respondWithError(w, http.StatusInternalServerError, "Database error", err)
            return
		}
	}

	// Issue jwt
	jwtToken, err := auth.GenerateJWT( user, cfg.jwtSecret, time.Hour)
	if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err)
        return
    }

	// TODO: decode user, store in DB, issue my own JWT

	// final redirect and TODO add jwt token once created
	finalRedirectURL := cfg.redirectTokenURI + jwtToken
	http.Redirect(w, r, finalRedirectURL, http.StatusFound)
}
