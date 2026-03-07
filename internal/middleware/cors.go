package middleware

import (
	"net/http"
	"strings"
)

var allowedOrigins = []string{
    "http://localhost:5173",
    "https://booking-calendar-chi.vercel.app",
    "https://*.hive.fi",
    "https://*.jgengo.dev",
}

func matchesOrigin(origin, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return origin == pattern
	}
	parts := strings.SplitN(pattern, "*", 2)
	return strings.HasPrefix(origin, parts[0]) && strings.HasSuffix(origin, parts[1])
}

// Cors ia a middleware to handle cors policy
func Cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        for _, allowed := range allowedOrigins {
            if matchesOrigin(origin, allowed) {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                break
            }
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}
