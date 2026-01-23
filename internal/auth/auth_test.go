package auth

import (
	"net/http"
	"testing"
	// "time"

	// "github.com/google/uuid"
)

// func TestValidateJWT(t *testing.T) {
// 	userID := uuid.New()
// 	validToken, _ := IssueAccessToken(userID, "secret", time.Hour)

// 	tests := []struct {
// 		name        string
// 		tokenString string
// 		tokenSecret string
// 		wantUserID  uuid.UUID
// 		wantErr     bool
// 	}{
// 		{
// 			name:        "Valid token",
// 			tokenString: validToken,
// 			tokenSecret: "secret",
// 			wantUserID:  userID,
// 			wantErr:     false,
// 		},
// 		{
// 			name:        "Invalid token",
// 			tokenString: "invalid.token.string",
// 			tokenSecret: "secret",
// 			wantUserID:  uuid.Nil,
// 			wantErr:     true,
// 		},
// 		{
// 			name:        "Wrong secret",
// 			tokenString: validToken,
// 			tokenSecret: "wrong_secret",
// 			wantUserID:  uuid.Nil,
// 			wantErr:     true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if gotUserID != tt.wantUserID {
// 				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
// 			}
// 		})
// 	}
// }

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name:    "no authorization header",
			header:  http.Header{},
			wantErr: true,
		},
		{
			name: "valid bearer token",
			header: http.Header{
				"Authorization": []string{"Bearer my-token"},
			},
			wantToken: "my-token",
			wantErr:   false,
		},
		{
			name: "wrong scheme",
			header: http.Header{
				"Authorization": []string{"Basic abc123"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.header)

			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Fatalf("expected token %q, got %q", tt.wantToken, token)
			}
		})
	}
}
