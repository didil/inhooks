package services

import (
	"context"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type SchedulerService interface {
	MoveDueScheduled(ctx context.Context, f *models.Flow, sink *models.Sink) error
}

func NewSchedulerService(redisStore RedisStore, timeSvc TimeService) SchedulerService {
	return &schedulerService{
		redisStore: redisStore,
		timeSvc:    timeSvc,
	}
}

type schedulerService struct {
	redisStore RedisStore
	timeSvc    TimeService
}

func (s *schedulerService) MoveDueScheduled(ctx context.Context, f *models.Flow, sink *models.Sink) error {
	scheduledQueueKey := queueKey(f.ID, sink.ID, models.QueueStatusScheduled)
	mIDs, err := s.redisStore.ZRangeBelowScore(ctx, scheduledQueueKey, float64(s.timeSvc.Now().Unix()))
	if err != nil {
		return errors.Wrapf(err, "failed to zrange below score")
	}
	if len(mIDs) == 0 {
		// no messages due
		return nil
	}

	// move message ids in chunks
	chunkSize := 50
	mIDChunks := lib.ChunkSliceBy(mIDs, chunkSize)

	sourceQueueKey := queueKey(f.ID, sink.ID, models.QueueStatusScheduled)
	destQueueKey := queueKey(f.ID, sink.ID, models.QueueStatusReady)

	for i := 0; i < len(mIDChunks); i++ {
		err := s.redisStore.ZRemRpush(ctx, mIDChunks[i], sourceQueueKey, destQueueKey)
		if err != nil {
			return errors.Wrapf(err, "failed to zremrpush")
		}
	}

	return nil
}
