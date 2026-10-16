package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"ems-backend/internal/response"
	"ems-backend/internal/service"
)

type contextKey string

const (
	// AdminContextKey is used to store and retrieve AdminClaims in the request context.
	AdminContextKey contextKey = "admin_claims"
)

// RequireAuth validates the Bearer JWT token on protected endpoints.
func RequireAuth(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip OPTIONS requests handled by CORS
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Missing Authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Unauthorized(w, "Invalid Authorization header format. Expected 'Bearer <token>'")
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
			if tokenStr == "" {
				response.Unauthorized(w, "Token is empty")
				return
			}

			claims, err := authService.ValidateToken(tokenStr)
			if err != nil {
				response.Unauthorized(w, "Invalid or expired authentication session")
				return
			}

			// Attach claims to context
			ctx := context.WithValue(r.Context(), AdminContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAdminClaims extracts the authenticated AdminClaims from context if available.
func GetAdminClaims(ctx context.Context) (*service.AdminClaims, bool) {
	claims, ok := ctx.Value(AdminContextKey).(*service.AdminClaims)
	return claims, ok
}

// clientRecord tracks request timestamps for rate limiting.
type clientRecord struct {
	attempts []time.Time
}

// LoginRateLimiter provides rate limiting for authentication requests.
type LoginRateLimiter struct {
	mu          sync.Mutex
	records     map[string]*clientRecord
	maxAttempts int
	window      time.Duration
}

// NewLoginRateLimiter creates a rate limiter with specified max attempts and time window.
func NewLoginRateLimiter(maxAttempts int, window time.Duration) *LoginRateLimiter {
	limiter := &LoginRateLimiter{
		records:     make(map[string]*clientRecord),
		maxAttempts: maxAttempts,
		window:      window,
	}

	// Background routine to evict stale client records periodically
	go limiter.cleanupLoop(window * 2)

	return limiter
}

func (l *LoginRateLimiter) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		l.mu.Lock()
		cutoff := time.Now().Add(-l.window)
		for ip, rec := range l.records {
			var active []time.Time
			for _, t := range rec.attempts {
				if t.After(cutoff) {
					active = append(active, t)
				}
			}
			if len(active) == 0 {
				delete(l.records, ip)
			} else {
				rec.attempts = active
			}
		}
		l.mu.Unlock()
	}
}

// Limit returns an HTTP middleware enforcing login rate limits by client IP.
func (l *LoginRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractClientIP(r)

		l.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-l.window)

		rec, exists := l.records[ip]
		if !exists {
			rec = &clientRecord{}
			l.records[ip] = rec
		}

		// Prune older attempts
		var active []time.Time
		for _, t := range rec.attempts {
			if t.After(cutoff) {
				active = append(active, t)
			}
		}
		rec.attempts = active

		if len(rec.attempts) >= l.maxAttempts {
			l.mu.Unlock()
			response.TooManyRequests(w, "Too many login attempts. Please wait a minute before trying again.")
			return
		}

		rec.attempts = append(rec.attempts, now)
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
