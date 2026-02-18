// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages per-IP rate limiting.
type RateLimiter struct {
	mu              sync.Mutex
	visitors        map[string]*visitor
	rate            rate.Limit
	burst           int
	cleanupInterval time.Duration
	lastCleanup     time.Time
	trustProxy      bool
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter.
//   - r: requests per second
//   - b: maximum burst size
//   - trustProxy: whether to trust X-Forwarded-For / X-Real-IP headers
func NewRateLimiter(r rate.Limit, b int, trustProxy bool) *RateLimiter {
	return &RateLimiter{
		visitors:        make(map[string]*visitor),
		rate:            r,
		burst:           b,
		cleanupInterval: 3 * time.Minute,
		lastCleanup:     time.Now(),
		trustProxy:      trustProxy,
	}
}

// getVisitor returns the rate limiter for the given IP,
// creating one if it doesn't exist. Cleans up stale entries inline.
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Inline cleanup of stale visitors.
	if now.Sub(rl.lastCleanup) > rl.cleanupInterval {
		for k, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.cleanupInterval {
				delete(rl.visitors, k)
			}
		}
		rl.lastCleanup = now
	}

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: now}
		return limiter
	}

	v.lastSeen = now
	return v.limiter
}

// Limit is the middleware that enforces rate limiting.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.getIP(r)
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			slog.Warn("rate limit exceeded", "ip", ip, "method", r.Method, "path", r.URL.Path)
			w.Header().Set("Retry-After", "6")
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP extracts the client IP address from the request.
func (rl *RateLimiter) getIP(r *http.Request) string {
	if rl.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.SplitN(xff, ",", 2)
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			return strings.TrimSpace(ip)
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
