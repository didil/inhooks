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

func TestMoveDueScheduled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	now := time.Now()
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

	sourceQueueKey := "f:flow-1:s:sink-1:q:scheduled"
	destQueueKey := "f:flow-1:s:sink-1:q:ready"

	mIDs := []string{"message-1", "message-2", "message-3"}
	redisStore.EXPECT().ZRangeBelowScore(ctx, sourceQueueKey, float64(now.Unix())).Return(mIDs, nil)
	redisStore.EXPECT().ZRemRpush(ctx, mIDs, sourceQueueKey, destQueueKey).Return(nil)

	s := NewSchedulerService(redisStore, timeSvc)
	err := s.MoveDueScheduled(ctx, flow, sink)
	assert.NoError(t, err)

}
