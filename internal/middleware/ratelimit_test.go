package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	// Create a rate limiter: 2 requests per second, burst of 2
	limiter := NewRateLimiter(rate.Limit(2), 2)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with rate limiter
	limitedHandler := limiter.Limit(handler)

	t.Run("allows requests within limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"

		// First request should succeed
		w := httptest.NewRecorder()
		limitedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Second request should succeed (within burst)
		w = httptest.NewRecorder()
		limitedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:1234"

		// Use up the burst
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			limitedHandler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("request %d: expected status 200, got %d", i+1, w.Code)
			}
		}

		// Next request should be rate limited
		w := httptest.NewRecorder()
		limitedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("expected status 429, got %d", w.Code)
		}

		if w.Body.String() == "" {
			t.Error("expected error message in response body")
		}
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.3:1234"

		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.4:1234"

		// Both IPs should be able to make requests
		w1 := httptest.NewRecorder()
		limitedHandler.ServeHTTP(w1, req1)

		w2 := httptest.NewRecorder()
		limitedHandler.ServeHTTP(w2, req2)

		if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
			t.Errorf("expected both requests to succeed, got %d and %d", w1.Code, w2.Code)
		}
	})
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:       "from RemoteAddr",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "from X-Real-IP",
			remoteAddr:    "192.168.1.1:1234",
			xRealIP:       "10.0.0.1",
			expectedIP:    "10.0.0.1",
		},
		{
			name:          "from X-Forwarded-For",
			remoteAddr:    "192.168.1.1:1234",
			xForwardedFor: "10.0.0.2",
			expectedIP:    "10.0.0.2",
		},
		{
			name:          "X-Forwarded-For takes precedence",
			remoteAddr:    "192.168.1.1:1234",
			xForwardedFor: "10.0.0.2",
			xRealIP:       "10.0.0.1",
			expectedIP:    "10.0.0.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getIP(req)
			if ip != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}
