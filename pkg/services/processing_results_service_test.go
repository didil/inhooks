package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestProcessingResultsServiceHandleOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)
	retryCalculator := mocks.NewMockRetryCalculator(ctrl)

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().AnyTimes().Return(now)

	flowID := "flow-1"
	sinkID := "sink-1"
	mID := "message-1"

	messageKey := "f:flow-1:s:sink-1:m:message-1"
	sourceQueueKey := "f:flow-1:s:sink-1:q:processing"
	destQueueKey := "f:flow-1:s:sink-1:q:done"

	m := &models.Message{
		ID:     mID,
		FlowID: flowID,
		SinkID: sinkID,
		DeliveryAttempts: []*models.DeliveryAttempt{
			{
				At:     now.Add(-5 * time.Minute),
				Status: models.DeliveryAttemptStatusFailed,
				Error:  "some error",
			},
		},
	}

	mUpdated := *m
	mUpdated.DeliveryAttempts = append(m.DeliveryAttempts, &models.DeliveryAttempt{
		At:     now,
		Status: models.DeliveryAttemptStatusOK,
	})

	b, err := json.Marshal(&mUpdated)
	assert.NoError(t, err)

	redisStore.EXPECT().SetLRemZAdd(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID, float64(now.Unix())).Return(nil)

	s := NewProcessingResultsService(timeSvc, redisStore, retryCalculator)
	err = s.HandleOK(ctx, m)
	assert.NoError(t, err)
}

func TestProcessingResultsServiceHandleFailed_Dead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)
	retryCalculator := mocks.NewMockRetryCalculator(ctrl)

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().AnyTimes().Return(now)

	flowID := "flow-1"
	sinkID := "sink-1"
	mID := "message-1"

	messageKey := "f:flow-1:s:sink-1:m:message-1"
	sourceQueueKey := "f:flow-1:s:sink-1:q:processing"
	destQueueKey := "f:flow-1:s:sink-1:q:dead"

	retryInterval := 15 * time.Minute
	retryExpMultiplier := float64(1)
	maxAttempts := 3
	sink := &models.Sink{
		ID:                 sinkID,
		RetryInterval:      &retryInterval,
		RetryExpMultiplier: &retryExpMultiplier,
		MaxAttempts:        &maxAttempts,
	}

	m := &models.Message{
		ID:     mID,
		FlowID: flowID,
		SinkID: sinkID,
		DeliveryAttempts: []*models.DeliveryAttempt{
			{
				At:     now.Add(-10 * time.Minute),
				Status: models.DeliveryAttemptStatusFailed,
				Error:  "some error",
			},
			{
				At:     now.Add(-5 * time.Minute),
				Status: models.DeliveryAttemptStatusFailed,
				Error:  "other error",
			},
		},
	}

	processingErr := fmt.Errorf("new error")

	mUpdated := *m
	mUpdated.DeliverAfter = now.Add(retryInterval)
	mUpdated.DeliveryAttempts = append(m.DeliveryAttempts, &models.DeliveryAttempt{
		At:     now,
		Status: models.DeliveryAttemptStatusFailed,
		Error:  processingErr.Error(),
	})

	b, err := json.Marshal(&mUpdated)
	assert.NoError(t, err)

	redisStore.EXPECT().SetAndMove(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID).Return(nil)
	retryCalculator.EXPECT().NextAttemptInterval(len(m.DeliveryAttempts)+1, &retryInterval, &retryExpMultiplier).Return(retryInterval)

	s := NewProcessingResultsService(timeSvc, redisStore, retryCalculator)
	queuedInfo, err := s.HandleFailed(ctx, sink, m, processingErr)
	assert.NoError(t, err)
	assert.Equal(t, mID, queuedInfo.MessageID)
	assert.Equal(t, models.QueueStatusDead, queuedInfo.QueueStatus)
}

func TestProcessingResultsServiceHandleFailed_Scheduled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)
	retryCalculator := mocks.NewMockRetryCalculator(ctrl)

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().AnyTimes().Return(now)

	flowID := "flow-1"
	sinkID := "sink-1"
	mID := "message-1"

	messageKey := "f:flow-1:s:sink-1:m:message-1"
	sourceQueueKey := "f:flow-1:s:sink-1:q:processing"
	destQueueKey := "f:flow-1:s:sink-1:q:scheduled"

	retryInterval := 4 * time.Second
	retryExpMultiplier := float64(2)
	maxAttempts := 3

	sink := &models.Sink{
		ID:                 sinkID,
		RetryInterval:      &retryInterval,
		RetryExpMultiplier: &retryExpMultiplier,
		MaxAttempts:        &maxAttempts,
	}

	m := &models.Message{
		ID:     mID,
		FlowID: flowID,
		SinkID: sinkID,
		DeliveryAttempts: []*models.DeliveryAttempt{
			{
				At:     now.Add(-5 * time.Minute),
				Status: models.DeliveryAttemptStatusFailed,
				Error:  "some error",
			},
		},
	}

	nextAttemptInterval := 8 * time.Second

	processingErr := fmt.Errorf("new error")

	mUpdated := *m
	mUpdated.DeliverAfter = now.Add(nextAttemptInterval)
	mUpdated.DeliveryAttempts = append(m.DeliveryAttempts, &models.DeliveryAttempt{
		At:     now,
		Status: models.DeliveryAttemptStatusFailed,
		Error:  processingErr.Error(),
	})

	b, err := json.Marshal(&mUpdated)
	assert.NoError(t, err)

	redisStore.EXPECT().SetLRemZAdd(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID, float64(mUpdated.DeliverAfter.Unix())).Return(nil)
	retryCalculator.EXPECT().NextAttemptInterval(len(m.DeliveryAttempts)+1, &retryInterval, &retryExpMultiplier).Return(nextAttemptInterval)

	s := NewProcessingResultsService(timeSvc, redisStore, retryCalculator)
	queuedInfo, err := s.HandleFailed(ctx, sink, m, processingErr)
	assert.NoError(t, err)
	assert.Equal(t, mID, queuedInfo.MessageID)
	assert.Equal(t, models.QueueStatusScheduled, queuedInfo.QueueStatus)
	assert.Equal(t, mUpdated.DeliverAfter, queuedInfo.DeliverAfter)
}

func TestProcessingResultsServiceHandleFailed_Ready(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)
	retryCalculator := mocks.NewMockRetryCalculator(ctrl)

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().AnyTimes().Return(now)

	flowID := "flow-1"
	sinkID := "sink-1"
	mID := "message-1"

	messageKey := "f:flow-1:s:sink-1:m:message-1"
	sourceQueueKey := "f:flow-1:s:sink-1:q:processing"
	destQueueKey := "f:flow-1:s:sink-1:q:ready"

	retryInterval := 0 * time.Second
	retryExpMultiplier := float64(1)
	maxAttempts := 3
	sink := &models.Sink{
		ID:                 sinkID,
		RetryInterval:      &retryInterval,
		RetryExpMultiplier: &retryExpMultiplier,
		MaxAttempts:        &maxAttempts,
	}
	nextAttemptInterval := 0 * time.Second

	m := &models.Message{
		ID:     mID,
		FlowID: flowID,
		SinkID: sinkID,
		DeliveryAttempts: []*models.DeliveryAttempt{
			{
				At:     now.Add(-5 * time.Minute),
				Status: models.DeliveryAttemptStatusFailed,
				Error:  "some error",
			},
		},
	}

	processingErr := fmt.Errorf("new error")

	mUpdated := *m
	mUpdated.DeliverAfter = now.Add(retryInterval)
	mUpdated.DeliveryAttempts = append(m.DeliveryAttempts, &models.DeliveryAttempt{
		At:     now,
		Status: models.DeliveryAttemptStatusFailed,
		Error:  processingErr.Error(),
	})

	b, err := json.Marshal(&mUpdated)
	assert.NoError(t, err)

	redisStore.EXPECT().SetAndMove(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID).Return(nil)
	retryCalculator.EXPECT().NextAttemptInterval(len(m.DeliveryAttempts)+1, &retryInterval, &retryExpMultiplier).Return(nextAttemptInterval)

	s := NewProcessingResultsService(timeSvc, redisStore, retryCalculator)
	queuedInfo, err := s.HandleFailed(ctx, sink, m, processingErr)
	assert.NoError(t, err)
	assert.Equal(t, mID, queuedInfo.MessageID)
	assert.Equal(t, models.QueueStatusReady, queuedInfo.QueueStatus)
	assert.Equal(t, mUpdated.DeliverAfter, queuedInfo.DeliverAfter)
}
