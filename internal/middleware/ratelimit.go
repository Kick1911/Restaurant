package middleware

import (
	"net/http"
	"sync"

	"github.com/kick/sigma-connected/internal/auth"
	"github.com/kick/sigma-connected/pkg/response"
	"golang.org/x/time/rate"
)

var (
	clients = make(map[string]*rate.Limiter)
	mu      sync.Mutex
)

func getRateLimiter(userID string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := clients[userID]
	if !exists {
		limiter = rate.NewLimiter(10, 20)
		clients[userID] = limiter
	}

	return limiter
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			next.ServeHTTP(w, r)
			return
		}

		limiter := getRateLimiter(claims.UserID)
		if !limiter.Allow() {
			response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}
