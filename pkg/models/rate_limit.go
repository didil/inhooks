package models

import (
	"fmt"
	"time"
)

type RateLimitDecision struct {
	Allowed    bool
	Remaining  float64
	RetryAfter time.Duration
}

type RateLimitConfig struct {
	Capacity       int           `yaml:"capacity"`
	RefillRate     int           `yaml:"refillRate"`
	RefillInterval time.Duration `yaml:"refillInterval"`
}

func (r *RateLimitConfig) Validate() error {
	if r.Capacity <= 0 {
		return fmt.Errorf("rateLimit.capacity must be > 0")
	}
	if r.RefillRate <= 0 {
		return fmt.Errorf("rateLimit.refill_rate must be > 0")
	}
	if r.RefillInterval <= 0 {
		return fmt.Errorf("rateLimit.refill_interval must be > 0")
	}
	if r.RefillRate > r.Capacity {
		return fmt.Errorf("rateLimit.refill_rate (%d) cannot exceed capacity (%d)", r.RefillRate, r.Capacity)
	}
	return nil
}
