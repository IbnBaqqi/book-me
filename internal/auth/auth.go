package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
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
	secret         string
	accessTokenTTL time.Duration
}

type TokenType string

const TokenTypeAccess TokenType = "book-me"

var (
	ErrInvalidToken			= errors.New("invalid token")
	ErrExpiredToken			= errors.New("expired token")
	ErrEmptyBearerToken     = errors.New("bearer token is empty")
	ErrInvalidBearerToken   = errors.New("bearer token is incorrect")
	ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")
)

func NewService(secret string) *Service {
	return &Service{
		secret:         secret,
		accessTokenTTL: time.Hour, // Access Token Time-To-Live
	}
}

// This create a jwt token
func (s *Service) IssueAccessToken(user database.User) (string, error) {

	claims := CustomClaims{
		Name: user.Name,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			Issuer:    string(TokenTypeAccess),
		},
	}
	signingKey := []byte(s.secret)
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
			return []byte(s.secret), nil
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
	bearerToken := headers.Get("Authorization")
	if bearerToken == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	token, ok := strings.CutPrefix(bearerToken, "Bearer ")
	if !ok {
		return "", ErrInvalidBearerToken
	}
	if token == "" {
		return "", ErrEmptyBearerToken
	}
	return token, nil
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

// MakeRefreshToken makes a random 256 bit token encoded in hex
func MakeRefreshToken() string {
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes) // no error check as Read always succeeds
	return hex.EncodeToString(tokenBytes)
}
