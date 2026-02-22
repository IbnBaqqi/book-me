package oauth

import (
	"fmt"
	"net/http"
)

// OauthError represents an oauth error with a status code
//
//nolint:revive // intentional naming for clarity
type OauthError struct {
	Err        error
	Message    string
	StatusCode int
}

// Error implements the error interface
func (e *OauthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *OauthError) Unwrap() error {
	return e.Err
}

// Wrap wraps an existing error with additional context
func (e *OauthError) Wrap(err error) *OauthError {
	return &OauthError{
		Err:        err,
		Message:    e.Message,
		StatusCode: e.StatusCode,
	}
}

// Predefined errors - OAuth errors
var (
	ErrOAuthStateMismatch = &OauthError{
		Message:    "oauth state mismatch",
		StatusCode: http.StatusForbidden,
	}
	ErrOAuthExchangeFailed = &OauthError{
		Message:    "oauth token exchange failed",
		StatusCode: http.StatusInternalServerError,
	}
	ErrOAuthUserInfoFailed = &OauthError{
		Message:    "failed to fetch user info from oauth provider",
		StatusCode: http.StatusInternalServerError,
	}
	ErrInvalidOAuthCode = &OauthError{
		Message:    "invalid or missing oauth code",
		StatusCode: http.StatusBadRequest,
	}
	ErrOAuthTimeout = &OauthError{
		Message:    "oauth request timeout",
		StatusCode: http.StatusGatewayTimeout,
	}
	ErrInvalidCampus = &OauthError{
		Message:    "access denied: only helsinki campus student allowed",
		StatusCode: http.StatusForbidden,
	}
	ErrOAuthSessionFailed = &OauthError{
		Message:    "failed to get session",
		StatusCode: http.StatusInternalServerError,
	}
	ErrFailedToSaveSession = &OauthError{
		Message:    "failed to save session",
		StatusCode: http.StatusInternalServerError,
	}
	ErrFailedToFindorCreateUser = &OauthError{
		Message:    "failed to find or create user",
		StatusCode: http.StatusInternalServerError,
	}
	ErrInvalidOrMissingState = &OauthError{
		Message:    "invalid or missing state",
		StatusCode: http.StatusForbidden,
	}
)
