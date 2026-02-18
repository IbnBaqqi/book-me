package oauth

import (
	"fmt"
	"net/http"
)

// OauthError represents an oauth error with a status code
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

// Predefined errors - Auth errors
var (
	ErrInvalidToken = &OauthError{
		Message:    "invalid token",
		StatusCode: http.StatusUnauthorized,
	}
	ErrExpiredToken = &OauthError{
		Message:    "token has expired",
		StatusCode: http.StatusUnauthorized,
	}
	ErrUnauthorized = &OauthError{
		Message:    "unauthorized",
		StatusCode: http.StatusUnauthorized,
	}
	ErrInvalidCredentials = &OauthError{
		Message:    "invalid credentials",
		StatusCode: http.StatusUnauthorized,
	}
	ErrTokenRevoked = &OauthError{
		Message:    "token has been revoked",
		StatusCode: http.StatusUnauthorized,
	}
	ErrInvalidTokenType = &OauthError{
		Message:    "invalid token type",
		StatusCode: http.StatusUnauthorized,
	}
	ErrMissingAuthHeader = &OauthError{
		Message:    "missing authorization header",
		StatusCode: http.StatusUnauthorized,
	}
	ErrInvalidAuthHeader = &OauthError{
		Message:    "invalid authorization header format",
		StatusCode: http.StatusUnauthorized,
	}
	ErrEmptyBearerToken = &OauthError{
		Message:    "bearer token is empty",
		StatusCode: http.StatusUnauthorized,
	}
	ErrAuthenticationRequired = &OauthError{
		Message:    "authentication required",
		StatusCode: http.StatusUnauthorized,
	}
	ErrInsufficientPermissions = &OauthError{
		Message:    "insufficient permissions",
		StatusCode: http.StatusForbidden,
	}
)

// Predefined errors - User errors
var (
	ErrUserNotFound = &OauthError{
		Message:    "user not found",
		StatusCode: http.StatusNotFound,
	}
	ErrUserAlreadyExists = &OauthError{
		Message:    "user already exists",
		StatusCode: http.StatusBadRequest,
	}
	ErrInvalidRole = &OauthError{
		Message:    "invalid role",
		StatusCode: http.StatusBadRequest,
	}
	ErrGetUserFailed = &OauthError{
		Message:    "failed to get user",
		StatusCode: http.StatusInternalServerError,
	}
)

// Predefined errors - OAuth errors
var (
	ErrOAuthStateMismatch = &OauthError{
		Message:    "oauth state mismatch",
		StatusCode: http.StatusForbidden,
	}
	ErrOAuthExchangeFailed = &OauthError{
		Message:    "oauth token exchange failed",
		StatusCode: http.StatusUnauthorized,
	}
	ErrOAuthUserInfoFailed = &OauthError{
		Message:    "failed to fetch user info from OAuth provider",
		StatusCode: http.StatusInternalServerError,
	}
	ErrInvalidOAuthCode = &OauthError{
		Message:    "invalid or missing OAuth code",
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
	ErrOAuthStateFailed = &OauthError{
		Message:    "failed to generate or store OAuth state",
		StatusCode: http.StatusInternalServerError,
	}
	ErrOAuthSessionFailed = &OauthError{
		Message:    "failed to get session",
		StatusCode: http.StatusInternalServerError,
	}
)
