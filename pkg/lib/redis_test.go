package lib

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestInitRedisClient(t *testing.T) {
	// Test case 1: Valid Redis URL
	t.Run("Valid Redis URL", func(t *testing.T) {
		conf := &AppConfig{
			Redis: RedisConfig{
				URL: "redis://localhostexample:6377",
			},
		}

		client, err := InitRedisClient(conf)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &redis.Client{}, client)

		// Clean up
		err = client.Close()
		assert.NoError(t, err)
	})

	// Test case 2: Invalid Redis URL
	t.Run("Invalid Redis URL", func(t *testing.T) {
		conf := &AppConfig{
			Redis: RedisConfig{
				URL: "invalid://url",
			},
		}

		client, err := InitRedisClient(conf)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to parse redis url")
	})
}
