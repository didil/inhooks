package services

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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

	timeOut := 1 * time.Second // minimum value accepted by redis is 1 second

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

func (s *RedisStoreSuite) TestLRangeAll() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`message-1`)
	value2 := []byte(`message-2`)

	queueKey := "q:processing"

	err := s.redisStore.Enqueue(ctx, queueKey, value1)
	s.NoError(err)
	err = s.redisStore.Enqueue(ctx, queueKey, value2)
	s.NoError(err)

	results, err := s.redisStore.LRangeAll(ctx, queueKey)
	s.NoError(err)

	s.Equal([]string{`message-1`, `message-2`}, results)
}

func (s *RedisStoreSuite) TestLRemRPush() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`message-1`)
	value2 := []byte(`message-2`)
	value3 := []byte(`message-3`)
	value4 := []byte(`message-4`)

	sourceQueueKey := "q:processing"
	destQueueKey := "q:ready"

	err := s.redisStore.Enqueue(ctx, sourceQueueKey, value1)
	s.NoError(err)
	err = s.redisStore.Enqueue(ctx, sourceQueueKey, value2)
	s.NoError(err)
	err = s.redisStore.Enqueue(ctx, sourceQueueKey, value3)
	s.NoError(err)
	err = s.redisStore.Enqueue(ctx, destQueueKey, value4)
	s.NoError(err)

	results, err := s.redisStore.LRangeAll(ctx, sourceQueueKey)
	s.NoError(err)
	s.Equal([]string{`message-1`, `message-2`, `message-3`}, results)

	results, err = s.redisStore.LRangeAll(ctx, destQueueKey)
	s.NoError(err)
	s.Equal([]string{`message-4`}, results)

	err = s.redisStore.LRemRPush(ctx, sourceQueueKey, destQueueKey, []string{"message-1", "message-3"})
	s.NoError(err)

	results, err = s.redisStore.LRangeAll(ctx, sourceQueueKey)
	s.NoError(err)
	s.Equal([]string{`message-2`}, results)

	results, err = s.redisStore.LRangeAll(ctx, destQueueKey)
	s.NoError(err)
	s.Equal([]string{`message-4`, "message-1", "message-3"}, results)
}

func (s *RedisStoreSuite) TestZRemRangeBelowScore() {
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

	count, err := s.redisStore.ZRemRangeBelowScore(ctx, queueKey, int(now.Unix()))
	s.NoError(err)

	s.Equal(2, count)

	queueResults, err := s.client.ZRange(ctx, queueKeyWithPrefix, 0, -1).Result()
	s.NoError(err)

	s.Equal([]string{"message-2", "message-4"}, queueResults)
}

func (s *RedisStoreSuite) TestZRemDel() {
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

	messageIDs := []string{m1ID, m3ID}
	messageKeys := []string{fmt.Sprintf("m:%s", m1ID), fmt.Sprintf("m:%s", m3ID)}

	_, err := s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Unix()), Member: m1ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Add(5 * time.Minute).Unix()), Member: m2ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Add(-5 * time.Minute).Unix()), Member: m3ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: float64(now.Add(20 * time.Minute).Unix()), Member: m4ID}).Result()
	s.NoError(err)

	err = s.redisStore.ZRemDel(ctx, queueKey, messageIDs, messageKeys)
	s.NoError(err)

	queueResults, err := s.client.ZRange(ctx, queueKeyWithPrefix, 0, -1).Result()
	s.NoError(err)

	s.Equal([]string{"message-2", "message-4"}, queueResults)

}

func (s *RedisStoreSuite) TestLLen() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	queueKey := "q:processing"

	length, err := s.redisStore.LLen(ctx, queueKey)
	s.NoError(err)
	s.Equal(int64(0), length)

	value1 := []byte(`{"id": 123}`)
	value2 := []byte(`{"id": 456}`)

	err = s.redisStore.Enqueue(ctx, queueKey, value1)
	s.NoError(err)

	length, err = s.redisStore.LLen(ctx, queueKey)
	s.NoError(err)
	s.Equal(int64(1), length)

	err = s.redisStore.Enqueue(ctx, queueKey, value2)
	s.NoError(err)

	length, err = s.redisStore.LLen(ctx, queueKey)
	s.NoError(err)
	s.Equal(int64(2), length)
}

func (s *RedisStoreSuite) TestZCard() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	queueKey := "q:scheduled"
	queueKeyWithPrefix := fmt.Sprintf("%s:%s", prefix, queueKey)

	count, err := s.redisStore.ZCard(ctx, queueKey)
	s.NoError(err)
	s.Equal(int64(0), count)

	m1ID := "message-1"
	m2ID := "message-2"
	m3ID := "message-3"

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: 100, Member: m1ID}).Result()
	s.NoError(err)

	count, err = s.redisStore.ZCard(ctx, queueKey)
	s.NoError(err)
	s.Equal(int64(1), count)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: 200, Member: m2ID}).Result()
	s.NoError(err)

	_, err = s.client.ZAdd(ctx, queueKeyWithPrefix, redis.Z{Score: 300, Member: m3ID}).Result()
	s.NoError(err)

	count, err = s.redisStore.ZCard(ctx, queueKey)
	s.NoError(err)
	s.Equal(int64(3), count)
}

func (s *RedisStoreSuite) TestMultiGet() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)
	value2 := []byte(`{"id": 456}`)

	messageKey1 := "m:abc123"
	messageKey2 := "m:def456"
	messageKey3 := "m:xyz789"

	err := s.redisStore.SetAndEnqueue(ctx, messageKey1, value1, "q:test", "abc123")
	s.NoError(err)
	err = s.redisStore.SetAndEnqueue(ctx, messageKey2, value2, "q:test", "def456")
	s.NoError(err)

	// fetch two existing keys and one non-existing key
	keys := []string{messageKey1, messageKey2, messageKey3}
	vals, err := s.redisStore.MultiGet(ctx, keys)
	s.NoError(err)

	s.Equal(2, len(vals))
	s.Equal(value1, vals[messageKey1])
	s.Equal(value2, vals[messageKey2])
	_, exists := vals[messageKey3]
	s.False(exists)
}

func (s *RedisStoreSuite) TestLRemDel() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	value1 := []byte(`{"id": 123}`)
	value2 := []byte(`{"id": 456}`)

	queueKey := "q:dead"
	messageKey1 := "m:abc123"
	messageKey2 := "m:def456"

	err := s.redisStore.SetAndEnqueue(ctx, messageKey1, value1, queueKey, "abc123")
	s.NoError(err)
	err = s.redisStore.SetAndEnqueue(ctx, messageKey2, value2, queueKey, "def456")
	s.NoError(err)

	queueResults, err := s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"abc123", "def456"}, queueResults)

	// remove abc123 from queue and delete its message key
	err = s.redisStore.LRemDel(ctx, queueKey, []string{"abc123"}, []string{messageKey1})
	s.NoError(err)

	queueResults, err = s.client.LRange(ctx, fmt.Sprintf("%s:%s", prefix, queueKey), 0, -1).Result()
	s.NoError(err)
	s.Equal([]string{"def456"}, queueResults)

	// message key for abc123 should be deleted
	val, err := s.client.Get(ctx, fmt.Sprintf("%s:%s", prefix, messageKey1)).Result()
	s.Error(err) // key should not exist
	s.Equal("", val)

	// message key for def456 should still exist
	val, err = s.client.Get(ctx, fmt.Sprintf("%s:%s", prefix, messageKey2)).Result()
	s.NoError(err)
	s.Equal(string(value2), val)
}

func (s *RedisStoreSuite) TestEval_ReturnsArray() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	script := `return {1, 2, 3}`
	res, err := s.redisStore.Eval(ctx, script, []string{"no-key"})
	s.NoError(err)
	s.Len(res, 3)
	s.Equal(int64(1), res[0])
	s.Equal(int64(2), res[1])
	s.Equal(int64(3), res[2])
}

func (s *RedisStoreSuite) TestEval_KeyPrefix() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	// Write to a key using the raw Redis client with prefix
	rawKey := fmt.Sprintf("%s:%s", prefix, "eval-test-key")
	err := s.client.Set(ctx, rawKey, "hello", 0).Err()
	s.NoError(err)

	// Eval using store — keys should be auto-prefixed
	// Wrap in table to return an array (Eval expects []interface{})
	script := `return {redis.call("GET", KEYS[1])}`
	res, err := s.redisStore.Eval(ctx, script, []string{"eval-test-key"})
	s.NoError(err)
	s.Len(res, 1)
	s.Equal("hello", res[0])
}

func (s *RedisStoreSuite) TestEval_Error() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	// Script that triggers a Redis error (wrong number of arguments to redis.call)
	script := `return redis.call("GET")`
	_, err := s.redisStore.Eval(ctx, script, []string{"some-key"})
	s.Error(err)
}

func (s *RedisStoreSuite) TestEval_TokenBucket_DrainAndRefill() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	rawKey := fmt.Sprintf("%s:%s", prefix, "rl:sink:drain-refill")
	s.client.Del(ctx, rawKey)

	capacity := 3
	refillRate := 2
	refillIntervalMs := 1000
	nowMs := time.Now().UnixMilli()

	// Drain all 3 tokens
	for i := 0; i < 3; i++ {
		res, err := s.redisStore.Eval(ctx, tokenBucketLua, []string{"rl:sink:drain-refill"}, capacity, refillRate, refillIntervalMs, nowMs)
		s.NoError(err)
		s.Len(res, 3)
		s.Equal(int64(1), res[0], "call %d should be allowed", i+1)
	}

	// 4th call should be denied
	res, err := s.redisStore.Eval(ctx, tokenBucketLua, []string{"rl:sink:drain-refill"}, capacity, refillRate, refillIntervalMs, nowMs)
	s.NoError(err)
	s.Equal(int64(0), res[0])
	retryAfterMs := res[2].(int64)
	s.Greater(retryAfterMs, int64(0), "retry_after should be positive")

	// Advance time past the refill interval and refill
	laterMs := nowMs + int64(refillIntervalMs)
	res, err = s.redisStore.Eval(ctx, tokenBucketLua, []string{"rl:sink:drain-refill"}, capacity, refillRate, refillIntervalMs, laterMs)
	s.NoError(err)
	s.Equal(int64(1), res[0], "should be allowed after refill")
}

func (s *RedisStoreSuite) TestEval_TokenBucket_CapacityCap() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	rawKey := fmt.Sprintf("%s:%s", prefix, "rl:sink:cap")
	s.client.Del(ctx, rawKey)

	capacity := 5
	refillRate := 100 // large refill
	refillIntervalMs := 1000
	nowMs := time.Now().UnixMilli()

	// Consume all tokens
	for i := 0; i < capacity; i++ {
		res, err := s.redisStore.Eval(ctx, tokenBucketLua, []string{"rl:sink:cap"}, capacity, refillRate, refillIntervalMs, nowMs)
		s.NoError(err)
		s.Equal(int64(1), res[0])
	}

	// Advance time significantly — refill should be capped at capacity
	laterMs := nowMs + int64(refillIntervalMs)*100
	res, err := s.redisStore.Eval(ctx, tokenBucketLua, []string{"rl:sink:cap"}, capacity, refillRate, refillIntervalMs, laterMs)
	s.NoError(err)
	s.Equal(int64(1), res[0])
	s.EqualValues(capacity-1, res[1], "should cap at capacity")
}

func (s *RedisStoreSuite) TestEval_TokenBucket_Atomicity() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	key := "rl:sink:atomic"
	rawKey := fmt.Sprintf("%s:%s", prefix, key)
	s.client.Del(ctx, rawKey)

	capacity := 10
	refillRate := 1
	refillIntervalMs := 1000
	nowMs := time.Now().UnixMilli()

	goroutines := 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	allowedCount := int32(0)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			res, err := s.redisStore.Eval(ctx, tokenBucketLua, []string{key}, capacity, refillRate, refillIntervalMs, nowMs)
			if err == nil && res[0].(int64) == 1 {
				atomic.AddInt32(&allowedCount, 1)
			}
		}()
	}

	wg.Wait()

	s.Equal(int32(capacity), allowedCount, "exactly capacity calls should be allowed")
}

func (s *RedisStoreSuite) TestEval_Peek() {
	ctx := context.Background()
	prefix := fmt.Sprintf("inhooks:%s", s.appConf.Redis.InhooksDBName)
	defer func() {
		err := testsupport.DeleteAllRedisKeys(ctx, s.client, prefix)
		s.NoError(err)
	}()

	rawKey := fmt.Sprintf("%s:%s", prefix, "rl:sink:peek")
	s.client.Del(ctx, rawKey)

	capacity := 5
	refillRate := 1
	refillIntervalMs := 1000
	nowMs := time.Now().UnixMilli()

	// Consume 2 tokens
	for i := 0; i < 2; i++ {
		_, err := s.redisStore.Eval(ctx, tokenBucketLua, []string{"rl:sink:peek"}, capacity, refillRate, refillIntervalMs, nowMs)
		s.NoError(err)
	}

	// Peek should show 3 remaining
	res, err := s.redisStore.Eval(ctx, peekLua, []string{"rl:sink:peek"}, capacity, refillRate, refillIntervalMs, nowMs)
	s.NoError(err)
	s.Len(res, 1)
	s.EqualValues(3, res[0])
}
