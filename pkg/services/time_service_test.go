package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeService(t *testing.T) {
	s := NewTimeService()

	now := s.Now()

	// check now is relatively accurate
	assert.GreaterOrEqual(t, now, time.Now())
	assert.Less(t, now, time.Now().Add(500*time.Millisecond))
}
