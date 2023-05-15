package services

import (
	"context"
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
