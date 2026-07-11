package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCleanUpServiceCleanupDoneQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	now := time.Date(2023, 05, 5, 8, 46, 20, 0, time.UTC)
	timeSvc.EXPECT().Now().Return(now)

	ctx := context.Background()

	flowId := "flow-1"
	sinkID := "sink-1"

	flow := &models.Flow{
		ID: flowId,
	}
	sink := &models.Sink{
		ID: sinkID,
	}

	queueKey := "f:flow-1:s:sink-1:q:done"

	doneQueueCleanupDelay := 30 * time.Minute
	cutoffTime := time.Date(2023, 05, 5, 8, 16, 20, 0, time.UTC)

	mIds := []string{"message-1", "message-2"}
	messageKeys := []string{"f:flow-1:s:sink-1:m:message-1", "f:flow-1:s:sink-1:m:message-2"}

	redisStore.EXPECT().ZRangeBelowScore(ctx, queueKey, float64(cutoffTime.Unix())).Return(mIds, nil)
	redisStore.EXPECT().ZRemDel(ctx, queueKey, mIds, messageKeys).Return(nil)

	s := NewCleanupService(redisStore, timeSvc)
	count, err := s.CleanupDoneQueue(ctx, flow, sink, doneQueueCleanupDelay)
	assert.NoError(t, err)

	assert.Equal(t, 2, count)
}

func TestCleanUpServiceCleanupDeadQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	now := time.Date(2023, 05, 5, 8, 46, 20, 0, time.UTC)
	timeSvc.EXPECT().Now().Return(now)

	ctx := context.Background()

	flowId := "flow-1"
	sinkID := "sink-1"

	flow := &models.Flow{
		ID: flowId,
	}
	sink := &models.Sink{
		ID: sinkID,
	}

	deadQueueKey := "f:flow-1:s:sink-1:q:dead"
	deadQueueCleanupDelay := 30 * time.Minute

	mIds := []string{"message-1", "message-2"}
	messageKey1 := "f:flow-1:s:sink-1:m:message-1"
	messageKey2 := "f:flow-1:s:sink-1:m:message-2"
	messageKeys := []string{messageKey1, messageKey2}

	// message-1 is old, message-2 is recent
	msg1 := &models.Message{
		ID: "message-1",
		DeliveryAttempts: []*models.DeliveryAttempt{
			{At: now.Add(-1 * time.Hour), Status: models.DeliveryAttemptStatusFailed, Error: "err"},
		},
	}
	msg1Bytes, _ := json.Marshal(msg1)

	msg2 := &models.Message{
		ID: "message-2",
		DeliveryAttempts: []*models.DeliveryAttempt{
			{At: now.Add(-1 * time.Minute), Status: models.DeliveryAttemptStatusFailed, Error: "err"},
		},
	}
	msg2Bytes, _ := json.Marshal(msg2)

	vals := map[string][]byte{
		messageKey1: msg1Bytes,
		messageKey2: msg2Bytes,
	}

	redisStore.EXPECT().LRangeAll(ctx, deadQueueKey).Return(mIds, nil)
	redisStore.EXPECT().MultiGet(ctx, messageKeys).Return(vals, nil)
	redisStore.EXPECT().LRemDel(ctx, deadQueueKey, []string{"message-1"}, []string{messageKey1}).Return(nil)

	s := NewCleanupService(redisStore, timeSvc)
	count, err := s.CleanupDeadQueue(ctx, flow, sink, deadQueueCleanupDelay)
	assert.NoError(t, err)

	assert.Equal(t, 1, count)
}

func TestCleanUpServiceCleanupDeadQueue_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	ctx := context.Background()

	flow := &models.Flow{ID: "flow-1"}
	sink := &models.Sink{ID: "sink-1"}

	deadQueueKey := "f:flow-1:s:sink-1:q:dead"

	redisStore.EXPECT().LRangeAll(ctx, deadQueueKey).Return([]string{}, nil)

	s := NewCleanupService(redisStore, timeSvc)
	count, err := s.CleanupDeadQueue(ctx, flow, sink, 30*time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
