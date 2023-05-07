package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeService(t *testing.T) {
	s := NewTimeService()

	n := time.Now()

	now := s.Now()

	// check now is relatively accurate
	assert.GreaterOrEqual(t, now, n)
	assert.Less(t, now, n.Add(100*time.Millisecond))
}
