package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type ProcessingResultsService interface {
	HandleFailed(ctx context.Context, sink *models.Sink, m *models.Message, processingErr error) error
	HandleOK(ctx context.Context, sink *models.Sink, m *models.Message) error
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
	m.DeliveryAttempts = append(m.DeliveryAttempts,
		&models.DeliveryAttempt{
			At:     s.timeSvc.Now(),
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
	m.DeliverAfter = s.timeSvc.Now().Add(retryAfter)
	// enqueue to ready, scheduled, or dead

	var maxAttempts int
	if sink.MaxAttempts == nil {
		maxAttempts = 0
	} else {
		maxAttempts = *sink.MaxAttempts
	}

	if len(m.DeliveryAttempts) >= maxAttempts {
		//TODO: enqueue to dead
		return nil
	}

	_ = getQueueStatus(m, s.timeSvc)

	//TODO: move queues

	return nil
}
func (s *processingResultsService) HandleOK(ctx context.Context, sink *models.Sink, m *models.Message) error {
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
		return errors.Wrapf(err, "failed to encode message to set and move to done for flow: %s source: %s sink: %s", m.FlowID, m.SourceID, m.SinkID)
	}

	// update message and move to done
	err = s.redisStore.SetAndMove(ctx, mKey, b, sourceQueueKey, destQueueKey, m.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to set and move to done for flow: %s source: %s sink: %s", m.FlowID, m.SourceID, m.SinkID)
	}

	return nil
}
