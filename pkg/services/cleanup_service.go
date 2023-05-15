package services

import (
	"context"
	"time"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type CleanupService interface {
	CleanupDoneQueue(ctx context.Context, f *models.Flow, sink *models.Sink, doneQueueCleanupDelay time.Duration) (int, error)
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
		return 0, err
	}
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
