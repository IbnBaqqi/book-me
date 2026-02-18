// Package auth provides JWT authentication and authorization.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomClaims represents custom JWT claims.
type CustomClaims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Service handles JWT authentication.
type Service struct {
	JwtSecret         string
	AccessTokenTTL time.Duration
}

// User represents an authenticated User.
type User struct {
	ID		int64
	Role	string
	Name	string
}

type contextKey struct{}

var userKey = contextKey{}

type tokenType string

const tokenTypeAccess tokenType = "access"

// Predefined errors - Auth errors
var (
	ErrInvalidToken			= errors.New("invalid token")
	ErrExpiredToken			= errors.New("expired token")
	ErrEmptyBearerToken     = errors.New("bearer token is empty")
	ErrInvalidBearerToken   = errors.New("bearer token is incorrect")
	ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")
)

// WithUser save user into context
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext get the saved user during Auth from context
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey).(User)
	return user, ok
}

// NewService creates a new auth service(JWT).
func NewService(secret string) *Service {
	return &Service{
		JwtSecret:         secret,
		AccessTokenTTL: time.Hour,
	}
}

// IssueAccessToken create a jwt token
func (s *Service) IssueAccessToken(user database.User) (string, error) {
	
	claims := CustomClaims{
		Name: user.Name,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.AccessTokenTTL)),
			Issuer:    string(tokenTypeAccess),
		},
	}
	signingKey := []byte(s.JwtSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return jwtToken, nil
}

// VerifyAccessToken validate the signature of the JWT and extract the claims
func (s *Service) VerifyAccessToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&CustomClaims{},
		func(token *jwt.Token) (any, error) {
			// enforce signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(s.JwtSecret), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetBearerToken return the Bearer Token from request header
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		return "", ErrInvalidBearerToken
	}
	if token == "" {
		return "", ErrEmptyBearerToken
	}
	return token, nil
}

// MakeRefreshToken makes a random 256 bit token encoded in hex
func MakeRefreshToken() string {
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes) // no error check as Read always succeeds
	return hex.EncodeToString(tokenBytes)
}

// ValidateJWT validates the signature of the JWT and extracts the claims(userID).
// 
// Deprecated: Use VerifyAccessToken instead.
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(tokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return userID, nil
}
