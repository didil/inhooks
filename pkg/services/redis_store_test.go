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

func (s *RedisStoreSuite) TestSetAndEnqueue_SetAndMove() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)
	value2 := []byte(`{"id": 456}`)
	value3 := []byte(`{"id": 789}`)

	queueKeyProcessing := "q:processing"
	messageID1 := "abc123"
	messageKey1 := "messages:abc123"
	messageID2 := "def456"
	messageKey2 := "messages:def456"
	messageID3 := "xyz789"
	messageKey3 := "messages:xyz789"

	err := s.redisStore.SetAndEnqueue(ctx, messageKey1, value1, queueKeyProcessing, messageID1)
	s.NoError(err)

	queueResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKeyProcessing), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"abc123"}, queueResults)

	val, err := s.redisStore.Get(ctx, messageKey1)
	s.NoError(err)
	s.Equal(value1, val)

	err = s.redisStore.SetAndEnqueue(ctx, messageKey2, value2, queueKeyProcessing, messageID2)
	s.NoError(err)

	err = s.redisStore.SetAndEnqueue(ctx, messageKey3, value3, queueKeyProcessing, messageID3)
	s.NoError(err)

	queueResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKeyProcessing), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"abc123", "def456", "xyz789"}, queueResults)

	value2Updated := []byte(`{"id": 456, "updated": true}`)
	queueKeyDone := "q:done"

	err = s.redisStore.SetAndMove(ctx, messageKey2, value2Updated, queueKeyProcessing, queueKeyDone, messageID2)
	s.NoError(err)

	val, err = s.redisStore.Get(ctx, messageKey2)
	s.NoError(err)
	s.Equal(value2Updated, val)

	queueResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKeyProcessing), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"abc123", "xyz789"}, queueResults)

	queueResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKeyDone), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"def456"}, queueResults)

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

func (s *RedisStoreSuite) TestSetLRemZAdd() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)
	value2 := []byte(`{"id": 456}`)
	deliverAfter := time.Date(2023, 05, 5, 8, 9, 24, 0, time.UTC).Unix()

	queueKeyScheduled := "q:scheduled"
	queueKeyProcessing := "q:processing"

	messageID1 := "abc123"
	messageKey1 := "messages:abc123"

	messageID2 := "def456"
	messageKey2 := "messages:def456"

	err := s.redisStore.SetAndEnqueue(ctx, messageKey1, value1, queueKeyProcessing, messageID1)
	s.NoError(err)
	err = s.redisStore.SetAndEnqueue(ctx, messageKey2, value2, queueKeyProcessing, messageID2)
	s.NoError(err)

	queueResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKeyProcessing), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"abc123", "def456"}, queueResults)

	err = s.redisStore.SetLRemZAdd(ctx, messageKey1, value1, queueKeyProcessing, queueKeyScheduled, messageID1, float64(deliverAfter))
	s.NoError(err)

	queueKeyScheduledWithPrefix := fmt.Sprintf("%s:%s", prefix, queueKeyScheduled)

	queueResults, err = s.client.ZRange(ctx, queueKeyScheduledWithPrefix, 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"abc123"}, queueResults)

	queueResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKeyProcessing), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"def456"}, queueResults)

	val, err := s.client.Get(ctx, fmt.Sprintf("%s:%s", prefix, messageKey1)).Result()
	s.NoError(err)

	s.Equal(string(value1), val)
}

func (s *RedisStoreSuite) TestZRangeBelowScore() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	now := time.Date(2023, 05, 5, 8, 9, 24, 0, time.UTC)

	queueKey := "q:scheduled"
	queueKeyWithPrefix := fmt.Sprintf("%s:%s", prefix, queueKey)

	m1ID := "message-1"
	m2ID := "message-2"
	m3ID := "message-3"
	m4ID := "message-4"

	_, err := s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Unix()), Member: m1ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Add(5 * time.Minute).Unix()), Member: m2ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Add(-5 * time.Minute).Unix()), Member: m3ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Add(20 * time.Minute).Unix()), Member: m4ID}).Result()
	s.NoError(err)

	vals, err := s.redisStore.ZRangeBelowScore(ctx, queueKey, float64(now.Unix()))
	s.NoError(err)

	s.Equal([]string{"message-3", "message-1"}, vals)
}

func (s *RedisStoreSuite) TestZRemRpush() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	now := time.Date(2023, 05, 5, 8, 9, 24, 0, time.UTC)

	sourceQueueKey := "q:scheduled"
	sourceQueueKeyWithPrefix := fmt.Sprintf("%s:%s", prefix, sourceQueueKey)

	destQueueKey := "q:ready"
	destQueueKeyWithPrefix := fmt.Sprintf("%s:%s", prefix, destQueueKey)

	m1ID := "message-1"
	m2ID := "message-2"
	m3ID := "message-3"
	m4ID := "message-4"

	_, err := s.client.ZAdd(ctx, sourceQueueKeyWithPrefix, redis.Z{Score: float64(now.Unix()), Member: m1ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, sourceQueueKeyWithPrefix, redis.Z{Score: float64(now.Add(5 * time.Minute).Unix()), Member: m2ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, sourceQueueKeyWithPrefix, redis.Z{Score: float64(now.Add(-5 * time.Minute).Unix()), Member: m3ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, sourceQueueKeyWithPrefix, redis.Z{Score: float64(now.Add(20 * time.Minute).Unix()), Member: m4ID}).Result()
	s.NoError(err)

	mIDs := []string{"message-3", "message-1"}

	err = s.redisStore.ZRemRpush(ctx, mIDs, sourceQueueKey, destQueueKey)
	s.NoError(err)

	queueResults, err := s.client.ZRange(ctx, sourceQueueKeyWithPrefix, 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"message-2", "message-4"}, queueResults)

	queueResults, err = s.client.LRange(ctx, destQueueKeyWithPrefix, 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"message-3", "message-1"}, queueResults)

}
