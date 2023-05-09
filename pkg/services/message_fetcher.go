package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type MessageFetcher interface {
	GetMessageForProcessing(ctx context.Context, timeout time.Duration, flowID string, sinkID string) (*models.Message, error)
}

func NewMessageFetcher(redisStore RedisStore, timeSvc TimeService) MessageFetcher {
	return &messageFetcher{
		redisStore: redisStore,
		timeSvc:    timeSvc,
	}
}

type messageFetcher struct {
	redisStore RedisStore
	timeSvc    TimeService
}

func (f *messageFetcher) GetMessageForProcessing(ctx context.Context, timeout time.Duration, flowID string, sinkID string) (*models.Message, error) {
	sourceQueueKey := queueKey(flowID, sinkID, QueueStatusReady)
	destQueueKey := queueKey(flowID, sinkID, QueueStatusProcessing)
	mIDBytes, err := f.redisStore.BLMove(ctx, timeout, sourceQueueKey, destQueueKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to redis blmove. flow: %s sink: %s", flowID, sinkID)
	}
	if mIDBytes == nil {
		return nil, nil
	}

	mID := string(mIDBytes)

	mKey := messageKey(flowID, sinkID, mID)
	b, err := f.redisStore.Get(ctx, mKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to redis get. flow: %s sink: %s", flowID, sinkID)
	}

	var m models.Message
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal message. m: %s", string(b))
	}

	return &m, nil
}
