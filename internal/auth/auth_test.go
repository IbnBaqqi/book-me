package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/IbnBaqqi/book-me/internal/database"
)

func TestIssueAccessToken(t *testing.T) {
	service := NewService("test-secret-key")

	testUser := database.User{
		ID:   123,
		Name: "John Doe",
		Role: "STUDENT",
	}

	token, err := service.IssueAccessToken(testUser)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Verify the token can be parsed
	claims, err := service.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("failed to verify issued token: %v", err)
	}

	if claims.Name != testUser.Name {
		t.Errorf("expected name %s, got %s", testUser.Name, claims.Name)
	}
	if claims.Role != testUser.Role {
		t.Errorf("expected role %s, got %s", testUser.Role, claims.Role)
	}
}

func TestVerifyAccessToken(t *testing.T) {
	service := NewService("test-secret-key")

	testUser := database.User{
		ID:   456,
		Name: "Jane Smith",
		Role: "STAFF",
	}

	validToken, _ := service.IssueAccessToken(testUser)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkFunc func(*testing.T, *CustomClaims)
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
			checkFunc: func(t *testing.T, claims *CustomClaims) {
				if claims.Name != "Jane Smith" {
					t.Errorf("expected name Jane Smith, got %s", claims.Name)
				}
				if claims.Role != "STAFF" {
					t.Errorf("expected role STAFF, got %s", claims.Role)
				}
			},
		},
		{
			name:    "invalid token format",
			token:   "invalid.token.string",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not-a-jwt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.VerifyAccessToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, claims)
			}
		})
	}
}

func TestVerifyAccessToken_WrongSecret(t *testing.T) {
	service1 := NewService("secret-1")
	service2 := NewService("secret-2")

	testUser := database.User{
		ID:   789,
		Name: "Bob",
		Role: "STUDENT",
	}

	token, _ := service1.IssueAccessToken(testUser)

	// Try to verify with different secret
	_, err := service2.VerifyAccessToken(token)
	if err == nil {
		t.Error("expected error when verifying with wrong secret")
	}
}

func TestVerifyAccessToken_ExpiredToken(t *testing.T) {
	service := NewService("test-secret")
	service.AccessTokenTTL = -1 * time.Hour // Already expired

	testUser := database.User{
		ID:   999,
		Name: "Expired User",
		Role: "STUDENT",
	}

	token, _ := service.IssueAccessToken(testUser)

	_, err := service.VerifyAccessToken(token)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    http.Header
		wantToken string
		wantErr   error
	}{
		{
			name:    "no authorization header",
			header:  http.Header{},
			wantErr: ErrNoAuthHeaderIncluded,
		},
		{
			name: "valid bearer token",
			header: http.Header{
				"Authorization": []string{"Bearer my-token-123"},
			},
			wantToken: "my-token-123",
			wantErr:   nil,
		},
		{
			name: "wrong scheme",
			header: http.Header{
				"Authorization": []string{"Basic abc123"},
			},
			wantErr: ErrInvalidBearerToken,
		},
		{
			name: "empty bearer token",
			header: http.Header{
				"Authorization": []string{"Bearer "},
			},
			wantErr: ErrEmptyBearerToken,
		},
		{
			name: "bearer with spaces",
			header: http.Header{
				"Authorization": []string{"Bearer  token-with-leading-space"},
			},
			wantToken: " token-with-leading-space",
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.header)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != tt.wantToken {
				t.Errorf("expected token %q, got %q", tt.wantToken, token)
			}
		})
	}
}

func TestWithUserAndUserFromContext(t *testing.T) {
	user := User{
		ID:   123,
		Name: "Test User",
		Role: "STUDENT",
	}

	ctx := context.Background()

	// Test UserFromContext with no user
	_, ok := UserFromContext(ctx)
	if ok {
		t.Error("expected no user in empty context")
	}

	// Test WithUser
	ctx = WithUser(ctx, user)

	// Test UserFromContext with user
	retrievedUser, ok := UserFromContext(ctx)
	if !ok {
		t.Fatal("expected user in context")
	}

	if retrievedUser.ID != user.ID {
		t.Errorf("expected ID %d, got %d", user.ID, retrievedUser.ID)
	}
	if retrievedUser.Name != user.Name {
		t.Errorf("expected name %s, got %s", user.Name, retrievedUser.Name)
	}
	if retrievedUser.Role != user.Role {
		t.Errorf("expected role %s, got %s", user.Role, retrievedUser.Role)
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token1 := MakeRefreshToken()
	token2 := MakeRefreshToken()

	if token1 == "" {
		t.Error("expected non-empty token")
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected token length 64, got %d", len(token1))
	}

	if token1 == token2 {
		t.Error("expected different tokens, got same")
	}
}

func TestNewService(t *testing.T) {
	secret := "my-secret-key"
	service := NewService(secret)

	if service.JwtSecret != secret {
		t.Errorf("expected secret %s, got %s", secret, service.JwtSecret)
	}

	if service.AccessTokenTTL != time.Hour {
		t.Errorf("expected TTL 1 hour, got %v", service.AccessTokenTTL)
	}
}
