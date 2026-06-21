package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	maxAttempts   = 5
	lockDuration  = 15 * time.Minute
	counterPrefix = "failed_login:"
	lockPrefix    = "lockout:"
)

type BruteForceProtector struct {
	rdb *redis.Client
}

func NewBruteForceProtector(rdb *redis.Client) *BruteForceProtector {
	return &BruteForceProtector{rdb: rdb}
}

func (b *BruteForceProtector) RecordFailedAttempt(ctx context.Context, email string) error {
	counterKey := counterPrefix + email
	pipe := b.rdb.Pipeline()
	pipe.Incr(ctx, counterKey)
	pipe.Expire(ctx, counterKey, lockDuration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("record failed attempt: %w", err)
	}

	count, err := b.rdb.Get(ctx, counterKey).Int()
	if err != nil {
		return fmt.Errorf("get attempt count: %w", err)
	}

	if count >= maxAttempts {
		lockKey := lockPrefix + email
		err := b.rdb.Set(ctx, lockKey, "1", lockDuration).Err()
		if err != nil {
			return fmt.Errorf("set lockout: %w", err)
		}
	}

	return nil
}

func (b *BruteForceProtector) IsLocked(ctx context.Context, email string) (bool, error) {
	lockKey := lockPrefix + email
	exists, err := b.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return false, fmt.Errorf("check lockout: %w", err)
	}
	return exists > 0, nil
}

func (b *BruteForceProtector) ResetAttempts(ctx context.Context, email string) error {
	counterKey := counterPrefix + email
	lockKey := lockPrefix + email
	pipe := b.rdb.Pipeline()
	pipe.Del(ctx, counterKey)
	pipe.Del(ctx, lockKey)
	_, err := pipe.Exec(ctx)
	return err
}
