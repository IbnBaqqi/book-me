package google

import "sync"

type TokenManager struct {
    authService *AuthService
    mu          sync.RWMutex
    cachedToken *AccessToken
}

func NewTokenManager(authService *AuthService) *TokenManager {
    return &TokenManager{
        authService: authService,
    }
}

func (tm *TokenManager) GetAccessToken() (string, error) {
    tm.mu.RLock()
    if tm.cachedToken != nil && !tm.cachedToken.IsExpired() {
        token := tm.cachedToken.Token
        tm.mu.RUnlock()
        return token, nil
    }
    tm.mu.RUnlock()

    // Acquire write lock to refresh token
    tm.mu.Lock()
    defer tm.mu.Unlock()

    // Double-check after acquiring write lock (another goroutine might have refreshed it)
    if tm.cachedToken != nil && !tm.cachedToken.IsExpired() {
        return tm.cachedToken.Token, nil
    }

    token, err := tm.authService.ProcessGoogleToken()
    if err != nil {
        return "", err
    }

    tm.cachedToken = token
    return token.Token, nil
}