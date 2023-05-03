package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/testsupport"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type RedisStoreSuite struct {
	suite.Suite
	client     *redis.Client
	redisStore RedisStore
	appConf    *lib.AppConfig
}

func TestRedisStoreSuite(t *testing.T) {
	suite.Run(t, new(RedisStoreSuite))
}

func (s *RedisStoreSuite) SetupTest() {
	err := lib.LoadEnvFromFile("../../.env.test")
	s.NoError(err)

	ctx := context.Background()
	appConf, err := lib.ProcessAppConfig(ctx)
	s.NoError(err)

	s.appConf = appConf

	client, err := lib.InitRedisClient(appConf)
	s.NoError(err)

	s.client = client

	redisStore, err := NewRedisStore(client, appConf.Redis.PrefixInhooksDBName)
	s.NoError(err)

	s.redisStore = redisStore
}

func (s *RedisStoreSuite) TestEnqueue_Dequeue() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.PrefixInhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)
	value2 := []byte(`{"id": 456}`)

	queueKey := "q:abc"

	err := s.redisStore.Enqueue(ctx, queueKey, value1)
	s.NoError(err)
	err = s.redisStore.Enqueue(ctx, queueKey, value2)
	s.NoError(err)

	results, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKey), 0, -1).Result()
	s.NoError(err)

	s.Equal([]string{`{"id": 123}`, `{"id": 456}`}, results)

	val2, err := s.redisStore.Dequeue(ctx, queueKey)
	s.NoError(err)
	s.Equal(val2, value2)

	val1, err := s.redisStore.Dequeue(ctx, queueKey)
	s.NoError(err)
	s.Equal(val1, value1)

	extraVal, err := s.redisStore.Dequeue(ctx, queueKey)
	s.NoError(err)

	s.Nil(extraVal)
}
