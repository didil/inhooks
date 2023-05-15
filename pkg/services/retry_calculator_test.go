package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryCalculator_ConstantBackoff(t *testing.T) {
	c := NewRetryCalculator()

	retryInterval := 5 * time.Second
	retryExpMultiplier := float64(1)

	assert.Equal(t, 5*time.Second, c.NextAttemptInterval(1, &retryInterval, &retryExpMultiplier))
	assert.Equal(t, 5*time.Second, c.NextAttemptInterval(2, &retryInterval, &retryExpMultiplier))
	assert.Equal(t, 5*time.Second, c.NextAttemptInterval(3, &retryInterval, &retryExpMultiplier))
}

func TestRetryCalculator_ExponentialBackoff(t *testing.T) {
	c := NewRetryCalculator()

	retryInterval := 3 * time.Second
	retryExpMultiplier := float64(2)

	assert.Equal(t, 3*time.Second, c.NextAttemptInterval(1, &retryInterval, &retryExpMultiplier))
	assert.Equal(t, 6*time.Second, c.NextAttemptInterval(2, &retryInterval, &retryExpMultiplier))
	assert.Equal(t, 12*time.Second, c.NextAttemptInterval(3, &retryInterval, &retryExpMultiplier))
	assert.Equal(t, 24*time.Second, c.NextAttemptInterval(4, &retryInterval, &retryExpMultiplier))
}
