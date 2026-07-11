package services

import (
	"context"
	"testing"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)

	ctx := context.Background()

	flowID := "flow-1"
	sinkID := "sink-1"

	flow := &models.Flow{
		ID: flowID,
	}
	sink := &models.Sink{
		ID: sinkID,
	}

	// scheduled queue (sorted set)
	redisStore.EXPECT().ZCard(ctx, "f:flow-1:s:sink-1:q:scheduled").Return(int64(3), nil)
	// ready queue (list)
	redisStore.EXPECT().LLen(ctx, "f:flow-1:s:sink-1:q:ready").Return(int64(10), nil)
	// processing queue (list)
	redisStore.EXPECT().LLen(ctx, "f:flow-1:s:sink-1:q:processing").Return(int64(5), nil)
	// done queue (sorted set)
	redisStore.EXPECT().ZCard(ctx, "f:flow-1:s:sink-1:q:done").Return(int64(100), nil)
	// dead queue (list)
	redisStore.EXPECT().LLen(ctx, "f:flow-1:s:sink-1:q:dead").Return(int64(2), nil)

	s := NewQueueMetricsService(redisStore)
	err := s.UpdateMetrics(ctx, flow, sink)
	assert.NoError(t, err)
}

func TestUpdateMetrics_ErrorOnLLen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)

	ctx := context.Background()

	flow := &models.Flow{
		ID: "flow-1",
	}
	sink := &models.Sink{
		ID: "sink-1",
	}

	redisStore.EXPECT().ZCard(ctx, "f:flow-1:s:sink-1:q:scheduled").Return(int64(0), nil)
	redisStore.EXPECT().LLen(ctx, "f:flow-1:s:sink-1:q:ready").Return(int64(0), assert.AnError)

	s := NewQueueMetricsService(redisStore)
	err := s.UpdateMetrics(ctx, flow, sink)
	assert.Error(t, err)
}

func TestUpdateMetrics_ErrorOnZCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)

	ctx := context.Background()

	flow := &models.Flow{
		ID: "flow-1",
	}
	sink := &models.Sink{
		ID: "sink-1",
	}

	redisStore.EXPECT().ZCard(ctx, "f:flow-1:s:sink-1:q:scheduled").Return(int64(0), assert.AnError)

	s := NewQueueMetricsService(redisStore)
	err := s.UpdateMetrics(ctx, flow, sink)
	assert.Error(t, err)
}
