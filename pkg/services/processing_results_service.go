package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type ProcessingResultsService interface {
	HandleFailed(ctx context.Context, sink *models.Sink, m *models.Message, processingErr error) error
	HandleOK(ctx context.Context, m *models.Message) error
}

type processingResultsService struct {
	timeSvc    TimeService
	redisStore RedisStore
}

func NewProcessingResultsService(timeSvc TimeService, redisStore RedisStore) ProcessingResultsService {
	return &processingResultsService{
		timeSvc:    timeSvc,
		redisStore: redisStore,
	}
}

func (s *processingResultsService) HandleFailed(ctx context.Context, sink *models.Sink, m *models.Message, processingErr error) error {
	now := s.timeSvc.Now()
	m.DeliveryAttempts = append(m.DeliveryAttempts,
		&models.DeliveryAttempt{
			At:     now,
			Status: models.DeliveryAttemptStatusFailed,
			Error:  processingErr,
		},
	)

	var retryAfter time.Duration
	if sink.RetryAfter == nil {
		retryAfter = 0
	} else {
		retryAfter = *sink.RetryAfter
	}
	m.DeliverAfter = now.Add(retryAfter)

	var maxAttempts int
	if sink.MaxAttempts == nil {
		maxAttempts = 0
	} else {
		maxAttempts = *sink.MaxAttempts
	}

	mKey := messageKey(m.FlowID, m.SinkID, m.ID)
	sourceQueueKey := queueKey(m.FlowID, m.SinkID, QueueStatusProcessing)

	b, err := json.Marshal(&m)
	if err != nil {
		return errors.Wrapf(err, "failed to encode message")
	}

	if len(m.DeliveryAttempts) >= maxAttempts {
		// update message and move to dead
		destQueueKey := queueKey(m.FlowID, m.SinkID, QueueStatusDead)
		err = s.redisStore.SetAndMove(ctx, mKey, b, sourceQueueKey, destQueueKey, m.ID)
		if err != nil {
			return errors.Wrapf(err, "failed to set and move to dead")
		}

		return nil
	}

	queueStatus := getQueueStatus(m, s.timeSvc)
	destQueueKey := queueKey(m.FlowID, m.SinkID, queueStatus)

	switch queueStatus {
	case QueueStatusReady:
		err = s.redisStore.SetAndMove(ctx, mKey, b, sourceQueueKey, destQueueKey, m.ID)
		if err != nil {
			return errors.Wrapf(err, "failed to set and enqueue ready message")
		}
	case QueueStatusScheduled:
		err = s.redisStore.SetLRemZAdd(ctx, mKey, b, sourceQueueKey, destQueueKey, m.ID, float64(m.DeliverAfter.Unix()))
		if err != nil {
			return errors.Wrapf(err, "failed to set and enqueue scheduled message")
		}
	default:
		return fmt.Errorf("unexpected queue status %s", queueStatus)
	}

	return nil
}
func (s *processingResultsService) HandleOK(ctx context.Context, m *models.Message) error {
	m.DeliveryAttempts = append(m.DeliveryAttempts,
		&models.DeliveryAttempt{
			At:     s.timeSvc.Now(),
			Status: models.DeliveryAttemptStatusOK,
		},
	)

	mKey := messageKey(m.FlowID, m.SinkID, m.ID)
	sourceQueueKey := queueKey(m.FlowID, m.SinkID, QueueStatusProcessing)
	destQueueKey := queueKey(m.FlowID, m.SinkID, QueueStatusDone)

	b, err := json.Marshal(&m)
	if err != nil {
		return errors.Wrapf(err, "failed to encode message")
	}

	// update message and move to done
	err = s.redisStore.SetAndMove(ctx, mKey, b, sourceQueueKey, destQueueKey, m.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to set and move to done")
	}

	return nil
}
