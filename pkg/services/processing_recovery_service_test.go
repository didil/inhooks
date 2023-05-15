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

func TestProcessingRecoveryService_Empty_Cache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)

	s, err := NewProcessingRecoveryService(redisStore)
	assert.NoError(t, err)

	flowID := "flow-id"
	flow := &models.Flow{
		ID: flowID,
	}

	sinkID := "sink-id"
	sink := &models.Sink{
		ID: sinkID,
	}

	sourceQueueKey := "f:flow-id:s:sink-id:q:processing"
	//	destQueueKey := "f:flow-id:s:sink-id:q:ready"

	processingMessagesId := []string{"message-1", "message-2", "message-3"}
	redisStore.EXPECT().LRangeAll(ctx, sourceQueueKey).Return(processingMessagesId, nil)
	ttl := 50 * time.Millisecond
	movedMessageIDs, err := s.MoveProcessingToReady(ctx, flow, sink, ttl)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, movedMessageIDs)

	//redisStore.EXPECT().LRemRPush(ctx, sourceQueueKey, destQueueKey)
}

func TestProcessingRecoveryService_Entry_In_Cache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)

	s, err := NewProcessingRecoveryService(redisStore)
	assert.NoError(t, err)

	flowID := "flow-id"
	flow := &models.Flow{
		ID: flowID,
	}

	sinkID := "sink-id"
	sink := &models.Sink{
		ID: sinkID,
	}

	sourceQueueKey := "f:flow-id:s:sink-id:q:processing"
	destQueueKey := "f:flow-id:s:sink-id:q:ready"

	processingMessagesId := []string{"message-1", "message-2", "message-3"}
	movedMessageID := "message-2"
	redisStore.EXPECT().LRangeAll(ctx, sourceQueueKey).Return(processingMessagesId, nil)
	redisStore.EXPECT().LRemRPush(ctx, sourceQueueKey, destQueueKey, []string{"message-2"})

	ttl := 1 * time.Second

	s.AddToCache(movedMessageID, ttl)

	movedMessageIDs, err := s.MoveProcessingToReady(ctx, flow, sink, ttl)
	assert.NoError(t, err)
	assert.Equal(t, []string{movedMessageID}, movedMessageIDs)
}

func TestProcessingRecoveryService_Entry_In_Cache_Expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	redisStore := mocks.NewMockRedisStore(ctrl)

	s, err := NewProcessingRecoveryService(redisStore)
	assert.NoError(t, err)

	flowID := "flow-id"
	flow := &models.Flow{
		ID: flowID,
	}

	sinkID := "sink-id"
	sink := &models.Sink{
		ID: sinkID,
	}

	sourceQueueKey := "f:flow-id:s:sink-id:q:processing"

	processingMessagesId := []string{"message-1", "message-2", "message-3"}
	movedMessageID := "message-2"
	redisStore.EXPECT().LRangeAll(ctx, sourceQueueKey).Return(processingMessagesId, nil)

	ttl := 10 * time.Millisecond

	s.AddToCache(movedMessageID, ttl)

	time.Sleep(ttl)

	movedMessageIDs, err := s.MoveProcessingToReady(ctx, flow, sink, ttl)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, movedMessageIDs)
}
