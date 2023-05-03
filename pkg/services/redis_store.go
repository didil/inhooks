package services

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisStore interface {
	Enqueue(ctx context.Context, key string, value []byte) error
	Dequeue(ctx context.Context, key string) ([]byte, error)
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
		return errors.Wrapf(err, "failed to lpush to %s", keyWithPrefix)
	}

	return nil
}

func (s *redisStore) Dequeue(ctx context.Context, key string) ([]byte, error) {
	keyWithPrefix := s.keyWithPrefix(key)
	res, err := s.client.RPop(ctx, keyWithPrefix).Bytes()
	if err != nil {
		if err == redis.Nil {
			// empty list
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to rpop from %s", keyWithPrefix)
	}

	return res, nil
}

func (s *redisStore) keyWithPrefix(key string) string {
	return fmt.Sprintf("inhooks:%s:%s", s.inhooksDBName, key)
}
