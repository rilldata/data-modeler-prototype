package ratelimit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"math"
)

// Limiter returns an error if quota per key is exceeded.
type Limiter interface {
	Limit(ctx context.Context, limitKey string, limit redis_rate.Limit) error
}

// Redis offers rate limiting functionality using a Redis-based rate limiter from the `go-redis/redis_rate`.
// The Redis supports the concept of 'No-operation' (Noop) that performs no rate limiting.
// This can be useful in local/testing environments or when rate limiting is not required.
type Redis struct {
	*redis_rate.Limiter
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{Limiter: redis_rate.NewLimiter(client)}
}

func (l *Redis) Limit(ctx context.Context, limitKey string, limit redis_rate.Limit) error {
	if limit == Unlimited {
		return nil
	}

	if limit.IsZero() {
		return NewQuotaExceededError("Resource quota not provided")
	}

	rateResult, err := l.Allow(ctx, limitKey, limit)
	if err != nil {
		return err
	}

	if rateResult.Allowed == 0 {
		return NewQuotaExceededError(fmt.Sprintf("Rate limit exceeded. Try again in %v seconds", rateResult.RetryAfter))
	}

	return nil
}

type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

func (n Noop) Limit(ctx context.Context, limitKey string, limit redis_rate.Limit) error {
	return nil
}

var Default = redis_rate.PerMinute(60)

var Sensitive = redis_rate.PerMinute(10)

var Public = redis_rate.PerMinute(250)

var Unlimited = redis_rate.PerSecond(math.MaxInt)

var Zero = redis_rate.Limit{}

type QuotaExceededError struct {
	message string
}

func (e QuotaExceededError) Error() string {
	return e.message
}

func NewQuotaExceededError(message string) QuotaExceededError {
	return QuotaExceededError{message}
}

func AuthLimitKey(methodName, authID string) string {
	return fmt.Sprintf("auth:%s:%s", methodName, authID)
}

func AnonLimitKey(methodName, peer string) string {
	return fmt.Sprintf("anon:%s:%s", methodName, peer)
}
