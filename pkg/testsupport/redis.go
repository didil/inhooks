package testsupport

import (
	"context"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func DeleteAllRedisKeys(ctx context.Context, redisClient *redis.Client, prefix string) error {
	keys, err := redisClient.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return errors.Wrapf(err, "failed to load keys")
	}

	pipe := redisClient.Pipeline()

	for _, key := range keys {
		pipe.Del(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	return err
}
