package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisStore interface {
	Get(ctx context.Context, messageKey string) ([]byte, error)
	SetAndEnqueue(ctx context.Context, messageKey string, value []byte, queueKey string, messageID string) error
	SetAndZAdd(ctx context.Context, messageKey string, value []byte, queueKey string, messageID string, score float64) error
	SetAndMove(ctx context.Context, messageKey string, value []byte, sourceQueueKey, destQueueKey string, messageID string) error
	SetLRemZAdd(ctx context.Context, messageKey string, value []byte, sourceQueueKey, destQueueKey string, messageID string, score float64) error
	Enqueue(ctx context.Context, key string, value []byte) error
	Dequeue(ctx context.Context, timeout time.Duration, key string) ([]byte, error)
	BLMove(ctx context.Context, timeout time.Duration, sourceQueueKey string, destQueueKey string) ([]byte, error)
	ZRangeBelowScore(ctx context.Context, queueKey string, score float64) ([]string, error)
	ZRemRpush(ctx context.Context, messageIDs []string, sourceQueueKey string, destQueueKey string) error
	LRangeAll(ctx context.Context, queueKey string) ([]string, error)
	LRemRPush(ctx context.Context, sourceQueueKey, destQueueKey string, messageIDs []string) error
	ZRemRangeBelowScore(ctx context.Context, queueKey string, maxScore int) (int, error)
	ZRemDel(ctx context.Context, queueKey string, messageIDs []string, messageKeys []string) error
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

func (s *redisStore) Get(ctx context.Context, messageKey string) ([]byte, error) {
	messageKeyWithPrefix := s.keyWithPrefix(messageKey)
	res, err := s.client.Get(ctx, messageKeyWithPrefix).Result()
	if err != nil {
		if err == redis.Nil {
			// no values
			return nil, nil
		}

		return nil, err
	}

	return []byte(res), nil
}

func (s *redisStore) SetAndEnqueue(ctx context.Context, messageKey string, value []byte, queueKey string, messageID string) error {
	pipe := s.client.TxPipeline()

	messageKeyWithPrefix := s.keyWithPrefix(messageKey)
	pipe.Set(ctx, messageKeyWithPrefix, value, 0)

	queueKeyWithPrefix := s.keyWithPrefix(queueKey)
	pipe.RPush(ctx, queueKeyWithPrefix, messageID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *redisStore) SetAndZAdd(ctx context.Context, messageKey string, value []byte, queueKey string, messageID string, score float64) error {
	pipe := s.client.TxPipeline()

	messageKeyWithPrefix := s.keyWithPrefix(messageKey)
	pipe.Set(ctx, messageKeyWithPrefix, value, 0)

	queueKeyWithPrefix := s.keyWithPrefix(queueKey)
	z := redis.Z{
		Score:  score,
		Member: messageID,
	}
	pipe.ZAdd(ctx, queueKeyWithPrefix, z)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *redisStore) SetAndMove(ctx context.Context, messageKey string, value []byte, sourceQueueKey, destQueueKey string, messageID string) error {
	pipe := s.client.TxPipeline()

	messageKeyWithPrefix := s.keyWithPrefix(messageKey)
	pipe.Set(ctx, messageKeyWithPrefix, value, 0)

	sourceKeyWithPrefix := s.keyWithPrefix(sourceQueueKey)
	pipe.LRem(ctx, sourceKeyWithPrefix, 0, messageID)

	destKeyWithPrefix := s.keyWithPrefix(destQueueKey)
	pipe.RPush(ctx, destKeyWithPrefix, messageID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *redisStore) SetLRemZAdd(ctx context.Context, messageKey string, value []byte, sourceQueueKey, destQueueKey string, messageID string, score float64) error {
	pipe := s.client.TxPipeline()

	messageKeyWithPrefix := s.keyWithPrefix(messageKey)
	pipe.Set(ctx, messageKeyWithPrefix, value, 0)

	sourceKeyWithPrefix := s.keyWithPrefix(sourceQueueKey)
	pipe.LRem(ctx, sourceKeyWithPrefix, 0, messageID)

	destKeyWithPrefix := s.keyWithPrefix(destQueueKey)
	z := redis.Z{
		Score:  score,
		Member: messageID,
	}
	pipe.ZAdd(ctx, destKeyWithPrefix, z)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
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

func (s *redisStore) ZRangeBelowScore(ctx context.Context, queueKey string, maxScore float64) ([]string, error) {
	queueKeyWithPrefix := s.keyWithPrefix(queueKey)

	args := redis.ZRangeArgs{
		Key:     queueKeyWithPrefix,
		Start:   "-inf",
		Stop:    maxScore,
		ByScore: true,
	}

	vals, err := s.client.ZRangeArgs(ctx, args).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to zrange. queueKey: %s", queueKeyWithPrefix)
	}

	return vals, nil
}

func (s *redisStore) ZRemRpush(ctx context.Context, messageIDs []string, sourceQueueKey string, destQueueKey string) error {
	pipe := s.client.TxPipeline()

	sourceKeyWithPrefix := s.keyWithPrefix(sourceQueueKey)
	pipe.ZRem(ctx, sourceKeyWithPrefix, messageIDs)

	destKeyWithPrefix := s.keyWithPrefix(destQueueKey)
	pipe.RPush(ctx, destKeyWithPrefix, messageIDs)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *redisStore) LRangeAll(ctx context.Context, queueKey string) ([]string, error) {
	queueKeyWithPrefix := s.keyWithPrefix(queueKey)

	vals, err := s.client.LRange(ctx, queueKeyWithPrefix, 0, -1).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lrange. queueKey: %s", queueKeyWithPrefix)
	}

	return vals, nil
}

func (s *redisStore) LRemRPush(ctx context.Context, sourceQueueKey, destQueueKey string, messageIDs []string) error {
	pipe := s.client.TxPipeline()

	sourceKeyWithPrefix := s.keyWithPrefix(sourceQueueKey)
	destKeyWithPrefix := s.keyWithPrefix(destQueueKey)

	for _, messageID := range messageIDs {
		pipe.LRem(ctx, sourceKeyWithPrefix, 0, messageID)
		pipe.RPush(ctx, destKeyWithPrefix, messageID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *redisStore) ZRemRangeBelowScore(ctx context.Context, queueKey string, maxScore int) (int, error) {
	queueKeyWithPrefix := s.keyWithPrefix(queueKey)

	count, err := s.client.ZRemRangeByScore(ctx, queueKeyWithPrefix, "-inf", strconv.Itoa(maxScore)).Result()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to zremrangebyscore. queueKey: %s", queueKeyWithPrefix)
	}

	return int(count), nil
}

func (s *redisStore) ZRemDel(ctx context.Context, queueKey string, messageIDs []string, messageKeys []string) error {
	pipe := s.client.TxPipeline()

	queueKeyWithPrefix := s.keyWithPrefix(queueKey)
	pipe.ZRem(ctx, queueKeyWithPrefix, messageIDs)

	messageKeysWithPrefix := []string{}
	for _, messageKey := range messageKeys {
		messageKeyWithPrefix := s.keyWithPrefix(messageKey)
		messageKeysWithPrefix = append(messageKeysWithPrefix, messageKeyWithPrefix)
	}

	pipe.Del(ctx, messageKeysWithPrefix...)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
