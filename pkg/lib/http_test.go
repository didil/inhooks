package lib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpClient(t *testing.T) {
	timeout := 5 * time.Second
	appConf := &AppConfig{
		HTTPClient: HTTPClientConfig{
			Timeout: timeout,
		},
	}

	cl := NewHttpClient(appConf)

	assert.Equal(t, timeout, cl.Timeout)
}
