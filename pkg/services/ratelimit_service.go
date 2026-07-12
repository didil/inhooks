package services

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

//go:embed luascripts/token_bucket.lua
var tokenBucketLua string

//go:embed luascripts/peek.lua
var peekLua string

type RateLimiter interface {
	Allow(ctx context.Context, key string, cfg *models.RateLimitConfig) models.RateLimitDecision
	Peek(ctx context.Context, key string, cfg *models.RateLimitConfig) (float64, error)
}

type tokenBucketLimiter struct {
	store   RedisStore
	timeSvc TimeService
	logger  *zap.Logger
}

func NewTokenBucketLimiter(store RedisStore, timeSvc TimeService, logger *zap.Logger) RateLimiter {
	return &tokenBucketLimiter{
		store:   store,
		timeSvc: timeSvc,
		logger:  logger,
	}
}

func (l *tokenBucketLimiter) Allow(ctx context.Context, key string, cfg *models.RateLimitConfig) models.RateLimitDecision {
	args := buildRateLimitArgs(cfg, l.timeSvc.Now().UnixMilli())

	res, err := l.store.Eval(ctx, tokenBucketLua, []string{key}, args...)
	if err != nil {
		l.logger.Warn("rate limiter error, failing open", zap.Error(err))
		return models.RateLimitDecision{Allowed: true, Remaining: float64(cfg.Capacity)}
	}

	allowed, remaining, retryAfter, err := parseAllowResult(res)
	if err != nil {
		l.logger.Warn("rate limiter parse error, failing open", zap.Error(err))
		return models.RateLimitDecision{Allowed: true, Remaining: float64(cfg.Capacity)}
	}

	return models.RateLimitDecision{
		Allowed:    allowed,
		Remaining:  remaining,
		RetryAfter: retryAfter,
	}
}

func (l *tokenBucketLimiter) Peek(ctx context.Context, key string, cfg *models.RateLimitConfig) (float64, error) {
	args := buildRateLimitArgs(cfg, l.timeSvc.Now().UnixMilli())

	res, err := l.store.Eval(ctx, peekLua, []string{key}, args...)
	if err != nil {
		return 0, err
	}

	if len(res) < 1 {
		return 0, fmt.Errorf("peek: empty result")
	}

	tokens, err := toFloat64(res[0])
	if err != nil {
		return 0, fmt.Errorf("peek: %w", err)
	}
	return tokens, nil
}

func buildRateLimitArgs(cfg *models.RateLimitConfig, nowMs int64) []interface{} {
	return []interface{}{
		cfg.Capacity,
		cfg.RefillRate,
		cfg.RefillInterval.Milliseconds(),
		nowMs,
	}
}

func parseAllowResult(res []interface{}) (bool, float64, time.Duration, error) {
	if len(res) < 3 {
		return false, 0, 0, fmt.Errorf("rate_limit: expected 3 values, got %d", len(res))
	}

	allowedI, ok := res[0].(int64)
	if !ok {
		return false, 0, 0, fmt.Errorf("rate_limit: allowed has unexpected type %T", res[0])
	}

	remaining, err := toFloat64(res[1])
	if err != nil {
		return false, 0, 0, fmt.Errorf("rate_limit: remaining: %w", err)
	}

	retryAfterMs, ok := res[2].(int64)
	if !ok {
		return false, 0, 0, fmt.Errorf("rate_limit: retry_after has unexpected type %T", res[2])
	}

	return allowedI == 1, remaining, time.Duration(retryAfterMs) * time.Millisecond, nil
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int64:
		return float64(val), nil
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}
