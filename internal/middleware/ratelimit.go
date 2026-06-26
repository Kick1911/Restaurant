package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kick/sigma-connected/internal/auth"
	"github.com/kick/sigma-connected/pkg/response"
)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

func (rl *RateLimiter) allow(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s", userID)
	now := time.Now().UnixMicro()
	window := int64(time.Second / time.Microsecond)

	pipe := rl.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-window))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, 2*time.Second)
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limit: %w", err)
	}

	count := cmds[2].(*redis.IntCmd).Val()
	return count <= 20, nil
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := rl.allow(r.Context(), claims.UserID)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "rate limit check failed")
			return
		}
		if !allowed {
			response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}
