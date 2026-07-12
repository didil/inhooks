package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimitConfig_Validate_OK(t *testing.T) {
	cases := []struct {
		name string
		cfg  RateLimitConfig
	}{
		{
			name: "simple 1s refill",
			cfg:  RateLimitConfig{Capacity: 10, RefillRate: 5, RefillInterval: time.Second},
		},
		{
			name: "refill equals capacity",
			cfg:  RateLimitConfig{Capacity: 50, RefillRate: 50, RefillInterval: time.Minute},
		},
		{
			name: "sub-second interval",
			cfg:  RateLimitConfig{Capacity: 10, RefillRate: 5, RefillInterval: 100 * time.Millisecond},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, tc.cfg.Validate())
		})
	}
}

func TestRateLimitConfig_Validate_Errors(t *testing.T) {
	cases := []struct {
		name    string
		cfg     RateLimitConfig
		wantErr string
	}{
		{
			name:    "zero capacity",
			cfg:     RateLimitConfig{Capacity: 0, RefillRate: 5, RefillInterval: time.Second},
			wantErr: "rateLimit.capacity must be > 0",
		},
		{
			name:    "negative capacity",
			cfg:     RateLimitConfig{Capacity: -1, RefillRate: 5, RefillInterval: time.Second},
			wantErr: "rateLimit.capacity must be > 0",
		},
		{
			name:    "zero refill rate",
			cfg:     RateLimitConfig{Capacity: 10, RefillRate: 0, RefillInterval: time.Second},
			wantErr: "rateLimit.refill_rate must be > 0",
		},
		{
			name:    "zero refill interval",
			cfg:     RateLimitConfig{Capacity: 10, RefillRate: 5, RefillInterval: 0},
			wantErr: "rateLimit.refill_interval must be > 0",
		},
		{
			name:    "refill rate exceeds capacity",
			cfg:     RateLimitConfig{Capacity: 10, RefillRate: 20, RefillInterval: time.Second},
			wantErr: "rateLimit.refill_rate (20) cannot exceed capacity (10)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
