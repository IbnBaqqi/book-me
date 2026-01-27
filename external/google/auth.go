// auth_service.go
package google

import (
    // "bytes"
    "crypto/rsa"
    "crypto/x509"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
    client              *http.Client
    calendarScope       string
    privateKey          string
    serviceAccountEmail string
    tokenExpiration     int64
    tokenURI            string
}

func NewAuthService(calendarScope, privateKey, serviceAccountEmail, tokenURI string, tokenExpiration int64) *AuthService {
    return &AuthService{
        client:              &http.Client{Timeout: 10 * time.Second},
        calendarScope:       calendarScope,
        privateKey:          privateKey,
        serviceAccountEmail: serviceAccountEmail,
        tokenExpiration:     tokenExpiration,
        tokenURI:            tokenURI,
    }
}

func (s *AuthService) ProcessGoogleToken() (*AccessToken, error) {
    jwtToken, err := s.generateGoogleJwtToken()
    if err != nil {
        return nil, fmt.Errorf("failed to generate JWT: %w", err)
    }

    data := url.Values{}
    data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
    data.Set("assertion", jwtToken)

    req, err := http.NewRequest("POST", s.tokenURI, strings.NewReader(data.Encode()))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
    }

    var token AccessToken
    if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    token.CreatedAt = time.Now()
    return &token, nil
}

func (s *AuthService) loadPrivateKeyFromPem() (*rsa.PrivateKey, error) {
    // Clean up PEM format
    privateKeyPem := strings.ReplaceAll(s.privateKey, "-----BEGIN PRIVATE KEY-----", "")
    privateKeyPem = strings.ReplaceAll(privateKeyPem, "-----END PRIVATE KEY-----", "")
    privateKeyPem = strings.ReplaceAll(privateKeyPem, "\\\\n", "")
    privateKeyPem = strings.ReplaceAll(privateKeyPem, "\\n", "")
    privateKeyPem = strings.ReplaceAll(privateKeyPem, "\n", "")
    privateKeyPem = strings.TrimSpace(privateKeyPem)

    // Decode base64
    keyBytes, err := base64.StdEncoding.DecodeString(privateKeyPem)
    if err != nil {
        return nil, fmt.Errorf("failed to decode private key: %w", err)
    }

    // Parse PKCS8 private key
    key, err := x509.ParsePKCS8PrivateKey(keyBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse private key: %w", err)
    }

    rsaKey, ok := key.(*rsa.PrivateKey)
    if !ok {
        return nil, errors.New("key is not RSA private key")
    }

    return rsaKey, nil
}

func (s *AuthService) generateGoogleJwtToken() (string, error) {
    privateKey, err := s.loadPrivateKeyFromPem()
    if err != nil {
        return "", err
    }

    now := time.Now()
    claims := jwt.MapClaims{
        "iss":   s.serviceAccountEmail,
        "scope": s.calendarScope,
        "aud":   s.tokenURI,
        "iat":   now.Unix(),
        "exp":   now.Add(time.Duration(s.tokenExpiration) * time.Second).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(privateKey)
}