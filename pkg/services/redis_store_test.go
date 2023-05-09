package services

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	ctx := context.Background()

	appConf, err := testsupport.InitAppConfig(ctx)
	s.NoError(err)

	s.appConf = appConf

	client, err := lib.InitRedisClient(appConf)
	s.NoError(err)

	s.client = client

	redisStore, err := NewRedisStore(client, appConf.Redis.InhooksDBName)
	s.NoError(err)

	s.redisStore = redisStore
}

func (s *RedisStoreSuite) TestEnqueue_Dequeue() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
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

	timeOut := 1 * time.Second

	val1, err := s.redisStore.Dequeue(ctx, timeOut, queueKey)
	s.NoError(err)
	s.Equal(val1, value1)

	val2, err := s.redisStore.Dequeue(ctx, timeOut, queueKey)
	s.NoError(err)
	s.Equal(val2, value2)

	extraVal, err := s.redisStore.Dequeue(ctx, timeOut, queueKey)
	s.NoError(err)

	s.Nil(extraVal)
}

func (s *RedisStoreSuite) TestBLMove() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)
	value2 := []byte(`{"id": 456}`)
	value3 := []byte(`{"id": 789}`)

	sourceQueueKey := "q:ready"
	destQueueKey := "q:processing"

	err := s.redisStore.Enqueue(ctx, sourceQueueKey, value1)
	s.NoError(err)
	err = s.redisStore.Enqueue(ctx, sourceQueueKey, value2)
	s.NoError(err)
	err = s.redisStore.Enqueue(ctx, destQueueKey, value3)
	s.NoError(err)

	sourceResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, sourceQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{`{"id": 123}`, `{"id": 456}`}, sourceResults)

	destResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, destQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{`{"id": 789}`}, destResults)

	timeOut := 1 * time.Second

	val1, err := s.redisStore.BLMove(ctx, timeOut, sourceQueueKey, destQueueKey)
	s.NoError(err)
	s.Equal(val1, value1)

	sourceResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, sourceQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{`{"id": 456}`}, sourceResults)

	destResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, destQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{`{"id": 789}`, `{"id": 123}`}, destResults)

	val2, err := s.redisStore.BLMove(ctx, timeOut, sourceQueueKey, destQueueKey)
	s.NoError(err)
	s.Equal(val2, value2)

	sourceResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, sourceQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{}, sourceResults)

	destResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, destQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{`{"id": 789}`, `{"id": 123}`, `{"id": 456}`}, destResults)

	noVal, err := s.redisStore.BLMove(ctx, timeOut, sourceQueueKey, destQueueKey)
	s.NoError(err)
	s.Nil(noVal)
}

func (s *RedisStoreSuite) TestSetAndEnqueue() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)

	queueKey := "q:ready"
	messageID := "abc123"
	messageKey := "messages:abc123"

	err := s.redisStore.SetAndEnqueue(ctx, messageKey, value1, queueKey, messageID)
	s.NoError(err)

	queueResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKey), 0, -1).Result()
	s.NoError(err)

	s.Equal([]string{"abc123"}, queueResults)

	val, err := s.client.Get(ctx, fmt.Sprintf("%s:%s", prefix, messageKey)).Result()
	s.NoError(err)

	s.Equal(string(value1), val)
}

func (s *RedisStoreSuite) TestSetAndZAdd() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)
	deliverAfter := time.Date(2023, 05, 5, 8, 9, 24, 0, time.UTC).Unix()

	setKey := "scheduled"
	messageID := "abc123"
	messageKey := "messages:abc123"

	err := s.redisStore.SetAndZAdd(ctx, messageKey, value1, setKey, messageID, float64(deliverAfter))
	s.NoError(err)

	setKeyWithPrefix := fmt.Sprintf("%s:%s", prefix, setKey)

	prevDate := time.Date(2023, 05, 4, 8, 9, 24, 0, time.UTC).Unix()
	otherID := "my-id"

	_, err = s.client.ZAdd(ctx, setKeyWithPrefix, redis.Z{Score: float64(prevDate), Member: otherID}).Result()
	s.NoError(err)

	queueResults, err := s.client.ZRange(ctx, setKeyWithPrefix, 0, -1).Result()
	s.NoError(err)

	s.Equal([]string{"my-id", "abc123"}, queueResults)

	val, err := s.client.Get(ctx, fmt.Sprintf("%s:%s", prefix, messageKey)).Result()
	s.NoError(err)

	s.Equal(string(value1), val)
}

func (s *RedisStoreSuite) TestGetAndBLMove() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": "message-123"}`)
	value2 := []byte(`{"id": "message-456"}`)
	value3 := []byte(`{"id": "message-789"}`)

	m1ID := "message-123"
	m2ID := "message-456"
	m3ID := "message-789"

	sourceQueueKey := "q:ready"
	destQueueKey := "q:processing"
	m1Key := "m:message-123"
	m2Key := "m:message-456"
	m3Key := "m:message-789"

	err := s.redisStore.SetAndEnqueue(ctx, m1Key, value1, sourceQueueKey, m1ID)
	s.NoError(err)
	err = s.redisStore.SetAndEnqueue(ctx, m2Key, value2, sourceQueueKey, m2ID)
	s.NoError(err)
	err = s.redisStore.SetAndEnqueue(ctx, m3Key, value3, destQueueKey, m3ID)
	s.NoError(err)

	sourceResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, sourceQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"message-123", "message-456"}, sourceResults)

	val1FromGet, err := s.client.Get(ctx, fmt.Sprintf("%s:%s", prefix, m1Key)).Result()
	s.NoError(err)
	s.Equal(string(value1), val1FromGet)

	destResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, destQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"message-789"}, destResults)

	timeOut := 1 * time.Second

	myM1ID, val1, err := s.redisStore.GetAndBLMove(ctx, m1Key, sourceQueueKey, destQueueKey, timeOut)
	s.NoError(err)
	s.Equal(value1, val1)
	s.Equal(m1ID, myM1ID)

	sourceResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, sourceQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"message-456"}, sourceResults)

	destResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, destQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"message-789", "message-123"}, destResults)

	m2ID, val2, err := s.redisStore.GetAndBLMove(ctx, timeOut, m2Key, sourceQueueKey, destQueueKey)
	s.NoError(err)
	s.Equal(val2, value2)

	sourceResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, sourceQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{}, sourceResults)

	destResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, destQueueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{`{"id": "message-789"}`, `{"id": "message-123"}`, `{"id": "message-456"}`}, destResults)

	noMsdID, noVal, err := s.redisStore.BLMove(ctx, timeOut, sourceQueueKey, destQueueKey)
	s.NoError(err)
	s.Nil(noVal)

}
