package services

import (
	"context"
	"time"

	"github.com/didil/inhooks/pkg/models"
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
		// enqueue to dead
		return nil
	}

	_ = getQueueStatus(m, s.timeSvc)

	// move queues

	return nil
}
func (s *processingResultsService) HandleOK(ctx context.Context, sink *models.Sink, m *models.Message) error {
	m.DeliveryAttempts = append(m.DeliveryAttempts,
		&models.DeliveryAttempt{
			At:     s.timeSvc.Now(),
			Status: models.DeliveryAttemptStatusOK,
		},
	)

	// need to move specific item from processing queue to done queue

	// move queues

	return nil
}
