package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type MessageEnqueuer interface {
	Enqueue(ctx context.Context, messages []*models.Message) error
}

func NewMessageEnqueuer(redisStore RedisStore, timeSvc TimeService) MessageEnqueuer {
	return &messageEnqueuer{
		redisStore: redisStore,
		timeSvc:    timeSvc,
	}
}

type messageEnqueuer struct {
	redisStore RedisStore
	timeSvc    TimeService
}

func (e *messageEnqueuer) Enqueue(ctx context.Context, messages []*models.Message) error {
	for _, m := range messages {
		var queueStatus QueueStatus

		if m.DeliverAfter.After(e.timeSvc.Now()) {
			// schedule in the future
			queueStatus = QueueStatusScheduled
		} else {
			// ready to process
			queueStatus = QueueStatusReady
		}

		b, err := json.Marshal(&m)
		if err != nil {
			return errors.Wrapf(err, "failed to encode message for flow: %s source: %s sink: %s", m.FlowID, m.SourceID, m.SinkID)
		}

		err = e.redisStore.Enqueue(ctx, queueKey(m.FlowID, m.SinkID, queueStatus), b)
		if err != nil {
			return errors.Wrapf(err, "failed to enqueue message for flow: %s source: %s sink: %s", m.FlowID, m.SourceID, m.SinkID)
		}
	}
	return nil
}

func queueKey(flowID string, sinkID string, queueStatus QueueStatus) string {
	return fmt.Sprintf("flow:%s:sink:%s:%s", flowID, sinkID, queueStatus)
}

type QueueStatus string

const (
	QueueStatusScheduled QueueStatus = "scheduled"
	QueueStatusReady     QueueStatus = "ready"
)
