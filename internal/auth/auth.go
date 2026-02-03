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

type CustomClaims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	JwtSecret         string
	AccessTokenTTL time.Duration
}

type User struct {
	ID		int16
	Role	string
	Name	string
}

type contextKey struct{}

var userKey = contextKey{}

type TokenType string

const TokenTypeAccess TokenType = "book-me"

var (
	ErrInvalidToken			= errors.New("invalid token")
	ErrExpiredToken			= errors.New("expired token")
	ErrEmptyBearerToken     = errors.New("bearer token is empty")
	ErrInvalidBearerToken   = errors.New("bearer token is incorrect")
	ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")
)


func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey).(User)
	return user, ok
}

// to create a new auth service
func NewService(secret string) *Service {
	return &Service{
		JwtSecret:         secret,
		AccessTokenTTL: time.Hour, // Access Token Time-To-Live
	}
}

// Authentication middleware
func (s *Service) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenStr, err := GetBearerToken(r.Header)
		if err != nil {
			// No token or malformed token - unauthenticated request
			next.ServeHTTP(w, r)
			return
		}

		claims, err := s.VerifyAccessToken(tokenStr)
		if err != nil {
			// public endpoints
			// Invalid or expired token - unauthenticated request
			next.ServeHTTP(w, r)
			return
		}

		// Convert claims.Subject (string) to int16 for User.ID
		idInt64, err := strconv.ParseInt(claims.Subject, 10, 16)
		if err != nil {
			// If conversion fails, treat as unauthenticated
			next.ServeHTTP(w, r)
			return
		}
		user := User{
			ID:		int16(idInt64),
			Role:	claims.Role,
			Name:	claims.Name,
		}

		ctx := WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authorization middleware
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFromContext(r.Context()); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}


// This create a jwt token
func (s *Service) IssueAccessToken(user database.User) (string, error) {
	
	claims := CustomClaims{
		Name: user.Name,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(int64(user.ID), 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.AccessTokenTTL)),
			Issuer:    string(TokenTypeAccess),
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

// This validate the signature of the JWT and extract the claims
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
	rand.Read(tokenBytes) // no error check as Read always succeeds
	return hex.EncodeToString(tokenBytes)
}

// (Depreciated) This validate the signature of the JWT and extract the claims(userId)
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userIdString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return userId, nil
}
