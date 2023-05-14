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

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().Return(now)

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

	redisStore.EXPECT().SetAndMove(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID).Return(nil)

	s := NewProcessingResultsService(timeSvc, redisStore)
	err = s.HandleOK(ctx, m)
	assert.NoError(t, err)
}

func TestProcessingResultsServiceHandleFailed_Dead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().Return(now)

	flowID := "flow-1"
	sinkID := "sink-1"
	mID := "message-1"

	messageKey := "f:flow-1:s:sink-1:m:message-1"
	sourceQueueKey := "f:flow-1:s:sink-1:q:processing"
	destQueueKey := "f:flow-1:s:sink-1:q:dead"

	retryAfter := 3 * time.Minute
	maxAttempts := 2
	sink := &models.Sink{
		ID:          sinkID,
		RetryAfter:  &retryAfter,
		MaxAttempts: &maxAttempts,
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

	processingErr := fmt.Errorf("new error")

	mUpdated := *m
	mUpdated.DeliverAfter = now.Add(retryAfter)
	mUpdated.DeliveryAttempts = append(m.DeliveryAttempts, &models.DeliveryAttempt{
		At:     now,
		Status: models.DeliveryAttemptStatusFailed,
		Error:  processingErr.Error(),
	})

	b, err := json.Marshal(&mUpdated)
	assert.NoError(t, err)

	redisStore.EXPECT().SetAndMove(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID).Return(nil)

	s := NewProcessingResultsService(timeSvc, redisStore)
	requeuedInfo, err := s.HandleFailed(ctx, sink, m, processingErr)
	assert.NoError(t, err)
	assert.Equal(t, models.QueueStatusDead, requeuedInfo.QueueStatus)
}

func TestProcessingResultsServiceHandleFailed_Scheduled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().Return(now)

	flowID := "flow-1"
	sinkID := "sink-1"
	mID := "message-1"

	messageKey := "f:flow-1:s:sink-1:m:message-1"
	sourceQueueKey := "f:flow-1:s:sink-1:q:processing"
	destQueueKey := "f:flow-1:s:sink-1:q:scheduled"

	retryAfter := 3 * time.Minute
	maxAttempts := 3
	sink := &models.Sink{
		ID:          sinkID,
		RetryAfter:  &retryAfter,
		MaxAttempts: &maxAttempts,
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

	processingErr := fmt.Errorf("new error")

	mUpdated := *m
	mUpdated.DeliverAfter = now.Add(retryAfter)
	mUpdated.DeliveryAttempts = append(m.DeliveryAttempts, &models.DeliveryAttempt{
		At:     now,
		Status: models.DeliveryAttemptStatusFailed,
		Error:  processingErr.Error(),
	})

	b, err := json.Marshal(&mUpdated)
	assert.NoError(t, err)

	redisStore.EXPECT().SetLRemZAdd(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID, float64(mUpdated.DeliverAfter.Unix())).Return(nil)

	s := NewProcessingResultsService(timeSvc, redisStore)
	requeuedInfo, err := s.HandleFailed(ctx, sink, m, processingErr)
	assert.NoError(t, err)
	assert.Equal(t, models.QueueStatusScheduled, requeuedInfo.QueueStatus)
	assert.Equal(t, mUpdated.DeliverAfter, requeuedInfo.DeliverAfter)
}

func TestProcessingResultsServiceHandleFailed_Ready(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().Return(now)

	flowID := "flow-1"
	sinkID := "sink-1"
	mID := "message-1"

	messageKey := "f:flow-1:s:sink-1:m:message-1"
	sourceQueueKey := "f:flow-1:s:sink-1:q:processing"
	destQueueKey := "f:flow-1:s:sink-1:q:ready"

	retryAfter := 0 * time.Second
	maxAttempts := 3
	sink := &models.Sink{
		ID:          sinkID,
		RetryAfter:  &retryAfter,
		MaxAttempts: &maxAttempts,
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

	processingErr := fmt.Errorf("new error")

	mUpdated := *m
	mUpdated.DeliverAfter = now.Add(retryAfter)
	mUpdated.DeliveryAttempts = append(m.DeliveryAttempts, &models.DeliveryAttempt{
		At:     now,
		Status: models.DeliveryAttemptStatusFailed,
		Error:  processingErr.Error(),
	})

	b, err := json.Marshal(&mUpdated)
	assert.NoError(t, err)

	redisStore.EXPECT().SetAndMove(ctx, messageKey, b, sourceQueueKey, destQueueKey, mID).Return(nil)

	s := NewProcessingResultsService(timeSvc, redisStore)
	requeuedInfo, err := s.HandleFailed(ctx, sink, m, processingErr)
	assert.NoError(t, err)
	assert.Equal(t, models.QueueStatusReady, requeuedInfo.QueueStatus)
	assert.Equal(t, mUpdated.DeliverAfter, requeuedInfo.DeliverAfter)
}
