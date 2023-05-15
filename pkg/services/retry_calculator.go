package services

import (
	"math"
	"time"
)

type RetryCalculator interface {
	NextAttemptInterval(attemptsCount int, retryInterval *time.Duration, retryExpMultiplier *float64) time.Duration
}

func NewRetryCalculator() RetryCalculator {
	return &retryCalculator{}
}

type retryCalculator struct {
}

func (c *retryCalculator) NextAttemptInterval(attemptsCount int, retryInterval *time.Duration, retryExpMultiplier *float64) time.Duration {
	var expMultiplier float64
	if retryExpMultiplier == nil {
		expMultiplier = 1 // constant backoff
	} else {
		expMultiplier = *retryExpMultiplier
	}

	var baseInterval time.Duration
	if retryInterval == nil {
		baseInterval = 0
	} else {
		baseInterval = *retryInterval
	}

	// interval formula =  "base interval" * "exponential backoff multiplier" ** "number of previous attempts - 1"
	interval := time.Duration(float64(baseInterval) * math.Pow(expMultiplier, float64(attemptsCount-1)))

	return interval
}
