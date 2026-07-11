package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type CleanupService interface {
	CleanupDoneQueue(ctx context.Context, f *models.Flow, sink *models.Sink, doneQueueCleanupDelay time.Duration) (int, error)
	CleanupDeadQueue(ctx context.Context, f *models.Flow, sink *models.Sink, deadQueueCleanupDelay time.Duration) (int, error)
}

func NewCleanupService(redisStore RedisStore, timeSvc TimeService) CleanupService {
	return &cleanupService{
		redisStore: redisStore,
		timeSvc:    timeSvc,
	}
}

type cleanupService struct {
	redisStore RedisStore
	timeSvc    TimeService
}

func (s *cleanupService) CleanupDoneQueue(ctx context.Context, f *models.Flow, sink *models.Sink, doneQueueCleanupDelay time.Duration) (int, error) {
	doneQueueKey := queueKey(f.ID, sink.ID, models.QueueStatusDone)

	cutOffTimeEpoch := s.timeSvc.Now().Add(-doneQueueCleanupDelay).Unix()
	mIDs, err := s.redisStore.ZRangeBelowScore(ctx, doneQueueKey, float64(cutOffTimeEpoch))
	if err != nil {
		return 0, errors.Wrapf(err, "failed to zrange below score")
	}
	if len(mIDs) == 0 {
		// no messages do cleanup
		return 0, nil
	}

	// move message ids in chunks
	chunkSize := 50
	mIDChunks := lib.ChunkSliceBy(mIDs, chunkSize)

	for i := 0; i < len(mIDChunks); i++ {
		messageKeys := make([]string, 0, len(mIDChunks[i]))
		for _, mId := range mIDChunks[i] {
			mKey := messageKey(f.ID, sink.ID, mId)
			messageKeys = append(messageKeys, mKey)
		}

		err := s.redisStore.ZRemDel(ctx, doneQueueKey, mIDChunks[i], messageKeys)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to zremdel")
		}
	}

	return len(mIDs), nil
}

func (s *cleanupService) CleanupDeadQueue(ctx context.Context, f *models.Flow, sink *models.Sink, deadQueueCleanupDelay time.Duration) (int, error) {
	deadQueueKey := queueKey(f.ID, sink.ID, models.QueueStatusDead)

	mIDs, err := s.redisStore.LRangeAll(ctx, deadQueueKey)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to lrange all dead queue")
	}
	if len(mIDs) == 0 {
		return 0, nil
	}

	cutOffTime := s.timeSvc.Now().Add(-deadQueueCleanupDelay)
	chunkSize := 50
	mIDChunks := lib.ChunkSliceBy(mIDs, chunkSize)

	totalDeleted := 0

	for i := 0; i < len(mIDChunks); i++ {
		messageKeys := make([]string, 0, len(mIDChunks[i]))
		for _, mID := range mIDChunks[i] {
			messageKeys = append(messageKeys, messageKey(f.ID, sink.ID, mID))
		}

		// fetch all message payloads in one pipeline
		vals, err := s.redisStore.MultiGet(ctx, messageKeys)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to multi get messages")
		}

		// filter by last delivery attempt timestamp
		toDeleteIDs := make([]string, 0, len(mIDChunks[i]))
		toDeleteKeys := make([]string, 0, len(mIDChunks[i]))

		for idx, mID := range mIDChunks[i] {
			val, ok := vals[messageKeys[idx]]
			if !ok {
				// message key missing, clean it up
				toDeleteIDs = append(toDeleteIDs, mID)
				toDeleteKeys = append(toDeleteKeys, messageKeys[idx])
				continue
			}

			var msg models.Message
			if err := json.Unmarshal(val, &msg); err != nil {
				return 0, errors.Wrapf(err, "failed to unmarshal message")
			}

			if len(msg.DeliveryAttempts) == 0 {
				// no delivery attempts, shouldn't happen in dead queue, clean it up
				toDeleteIDs = append(toDeleteIDs, mID)
				toDeleteKeys = append(toDeleteKeys, messageKeys[idx])
				continue
			}

			lastAttempt := msg.DeliveryAttempts[len(msg.DeliveryAttempts)-1]
			if lastAttempt.At.Before(cutOffTime) {
				toDeleteIDs = append(toDeleteIDs, mID)
				toDeleteKeys = append(toDeleteKeys, messageKeys[idx])
			}
		}

		if len(toDeleteIDs) == 0 {
			continue
		}

		err = s.redisStore.LRemDel(ctx, deadQueueKey, toDeleteIDs, toDeleteKeys)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to lremdel dead queue")
		}

		totalDeleted += len(toDeleteIDs)
	}

	return totalDeleted, nil
}
