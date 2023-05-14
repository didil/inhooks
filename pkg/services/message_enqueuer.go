package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
)

type MessageEnqueuer interface {
	Enqueue(ctx context.Context, messages []*models.Message) ([]*models.QueuedInfo, error)
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

func (e *messageEnqueuer) Enqueue(ctx context.Context, messages []*models.Message) ([]*models.QueuedInfo, error) {
	queuedInfos := []*models.QueuedInfo{}

	for _, m := range messages {
		queueStatus := getQueueStatus(m, e.timeSvc.Now())

		err := e.redisEnqueue(ctx, m, queueStatus)
		if err != nil {
			return nil, err
		}

		queuedInfos = append(queuedInfos, &models.QueuedInfo{MessageID: m.ID, QueueStatus: queueStatus, DeliverAfter: m.DeliverAfter})
	}

	return queuedInfos, nil
}

func (e *messageEnqueuer) redisEnqueue(ctx context.Context, m *models.Message, queueStatus models.QueueStatus) error {
	b, err := json.Marshal(&m)
	if err != nil {
		return errors.Wrapf(err, "failed to encode message for sink: %s", m.SinkID)
	}

	mKey := messageKey(m.FlowID, m.SinkID, m.ID)
	qKey := queueKey(m.FlowID, m.SinkID, queueStatus)

	switch queueStatus {
	case models.QueueStatusReady:
		err = e.redisStore.SetAndEnqueue(ctx, mKey, b, qKey, m.ID)
		if err != nil {
			return errors.Wrapf(err, "failed to set and enqueue message for sink: %s", m.SinkID)
		}
	case models.QueueStatusScheduled:
		err = e.redisStore.SetAndZAdd(ctx, mKey, b, qKey, m.ID, float64(m.DeliverAfter.Unix()))
		if err != nil {
			return errors.Wrapf(err, "failed to set and enqueue message for sink: %s", m.SinkID)
		}
	default:
		return fmt.Errorf("unexpected queue status %s", queueStatus)
	}

	return nil
}

func getQueueStatus(m *models.Message, now time.Time) models.QueueStatus {
	if m.DeliverAfter.After(now) {
		// schedule in the future
		return models.QueueStatusScheduled
	}
	// ready to process
	return models.QueueStatusReady
}

func flowSinkKeyPrefix(flowID string, sinkID string) string {
	return fmt.Sprintf("f:%s:s:%s", flowID, sinkID)
}

func messageKey(flowID string, sinkID string, messageID string) string {
	return fmt.Sprintf("%s:m:%s", flowSinkKeyPrefix(flowID, sinkID), messageID)
}

func queueKey(flowID string, sinkID string, queueStatus models.QueueStatus) string {
	return fmt.Sprintf("%s:q:%s", flowSinkKeyPrefix(flowID, sinkID), queueStatus)
}
