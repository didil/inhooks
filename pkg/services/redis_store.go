package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisStore interface {
	Enqueue(ctx context.Context, key string, value []byte) error
	Dequeue(ctx context.Context, timeout time.Duration, key string) ([]byte, error)
	BLMove(ctx context.Context, timeout time.Duration, sourceQueueKey string, destQueueKey string) ([]byte, error)
}

type redisStore struct {
	client        *redis.Client
	inhooksDBName string
}

func NewRedisStore(client *redis.Client, inhooksDBName string) (RedisStore, error) {
	if inhooksDBName == "" {
		return nil, fmt.Errorf("env var REDIS_INHOOKS_DB_NAME not set")
	}

	st := &redisStore{
		client:        client,
		inhooksDBName: inhooksDBName,
	}

	return st, nil
}

func (s *redisStore) Enqueue(ctx context.Context, key string, value []byte) error {
	keyWithPrefix := s.keyWithPrefix(key)
	err := s.client.RPush(ctx, keyWithPrefix, value).Err()
	if err != nil {
		return errors.Wrapf(err, "failed to rpush to %s", keyWithPrefix)
	}

	return nil
}

func (s *redisStore) Dequeue(ctx context.Context, timeout time.Duration, key string) ([]byte, error) {
	keyWithPrefix := s.keyWithPrefix(key)
	res, err := s.client.BLPop(ctx, timeout, keyWithPrefix).Result()
	if err != nil {
		if err == redis.Nil {
			// no values
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to blpop. key: %s", keyWithPrefix)
	}
	if len(res) != 2 {
		return nil, errors.Wrapf(err, "blpop results should containe 2 elements. key: %s", keyWithPrefix)
	}

	return []byte(res[1]), nil
}

func (s *redisStore) BLMove(ctx context.Context, timeout time.Duration, sourceQueueKey string, destQueueKey string) ([]byte, error) {
	sourceKeyWithPrefix := s.keyWithPrefix(sourceQueueKey)
	destKeyWithPrefix := s.keyWithPrefix(destQueueKey)
	res, err := s.client.BLMove(ctx, sourceKeyWithPrefix, destKeyWithPrefix, "LEFT", "RIGHT", timeout).Result()
	if err != nil {
		if err == redis.Nil {
			// no values
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to blmove. source: %s dest: %s", sourceKeyWithPrefix, destKeyWithPrefix)
	}

	return []byte(res), nil
}

func (s *redisStore) keyWithPrefix(key string) string {
	return fmt.Sprintf("inhooks:%s:%s", s.inhooksDBName, key)
}
