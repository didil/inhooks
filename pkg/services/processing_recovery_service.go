package services

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type ProcessingRecoveryService interface {
	MoveProcessingToReady(ctx context.Context, flow *models.Flow, sink *models.Sink, processingRecoveryInterval time.Duration) ([]string, error)
	AddToCache(mID string, ttl time.Duration)
}

type processingRecoveryService struct {
	redisStore RedisStore
	cache      *ristretto.Cache
}

func NewProcessingRecoveryService(redisStore RedisStore) (ProcessingRecoveryService, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 2e5,           // number of keys to track frequency of (200K).
		MaxCost:     2 * (1 << 20), // maximum cost of cache (1MB).
		BufferItems: 64,            // number of keys per Get buffer.
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to init cache")
	}

	svc := &processingRecoveryService{
		redisStore: redisStore,
		cache:      cache,
	}

	return svc, nil
}

const recoveryCacheDummyValue = "1"

func (s *processingRecoveryService) MoveProcessingToReady(ctx context.Context, flow *models.Flow, sink *models.Sink, ttl time.Duration) ([]string, error) {
	sourceQueueKey := queueKey(flow.ID, sink.ID, models.QueueStatusProcessing)
	destQueueKey := queueKey(flow.ID, sink.ID, models.QueueStatusReady)
	messageIDs, err := s.redisStore.LRangeAll(ctx, sourceQueueKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lrangeall")
	}
	if len(messageIDs) == 0 {
		// nothing in processing
		return nil, nil
	}

	messageIDsToMove := []string{}
	for _, mID := range messageIDs {
		if _, found := s.cache.Get(mID); found {
			// stuck message found
			messageIDsToMove = append(messageIDsToMove, mID)
		} else {
			s.cache.SetWithTTL(mID, recoveryCacheDummyValue, 1, ttl)
		}
	}

	if len(messageIDsToMove) == 0 {
		// no messages to move
		return messageIDsToMove, nil
	}

	err = s.redisStore.LRemRPush(ctx, sourceQueueKey, destQueueKey, messageIDsToMove)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lremrpush")
	}

	return messageIDsToMove, nil
}

// only used for testing
func (s *processingRecoveryService) AddToCache(mID string, ttl time.Duration) {
	s.cache.SetWithTTL(mID, recoveryCacheDummyValue, 1, ttl)
	s.cache.Wait()
}
