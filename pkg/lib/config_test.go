package lib

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitAppConfig(t *testing.T) {
	ctx := context.Background()

	oldRedisInhooksDBName := os.Getenv("REDIS_INHOOKS_DB_NAME")
	defer func() {
		os.Setenv("REDIS_INHOOKS_DB_NAME", oldRedisInhooksDBName)
	}()

	os.Setenv("REDIS_INHOOKS_DB_NAME", "mydb")

	appConf, err := InitAppConfig(ctx)
	assert.NoError(t, err)

	assert.Equal(t, "mydb", appConf.Redis.InhooksDBName)
}
