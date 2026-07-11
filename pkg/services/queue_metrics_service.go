package services

import (
	"context"

	"github.com/didil/inhooks/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var queueSizeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "queue_size",
	Help: "Number of messages in a queue",
}, []string{"flow_id", "sink_id", "queue_status"})

type QueueMetricsService interface {
	UpdateMetrics(ctx context.Context, f *models.Flow, sink *models.Sink) error
}

type queueMetricsService struct {
	redisStore RedisStore
}

func NewQueueMetricsService(redisStore RedisStore) QueueMetricsService {
	return &queueMetricsService{
		redisStore: redisStore,
	}
}

func (s *queueMetricsService) UpdateMetrics(ctx context.Context, f *models.Flow, sink *models.Sink) error {
	queueStatuses := []models.QueueStatus{
		models.QueueStatusScheduled,
		models.QueueStatusReady,
		models.QueueStatusProcessing,
		models.QueueStatusDone,
		models.QueueStatusDead,
	}

	for _, qs := range queueStatuses {
		size, err := s.getQueueSize(ctx, f.ID, sink.ID, qs)
		if err != nil {
			return err
		}
		queueSizeGauge.WithLabelValues(f.ID, sink.ID, string(qs)).Set(float64(size))
	}

	return nil
}

func (s *queueMetricsService) getQueueSize(ctx context.Context, flowID string, sinkID string, queueStatus models.QueueStatus) (int64, error) {
	qKey := queueKey(flowID, sinkID, queueStatus)

	switch queueStatus {
	case models.QueueStatusScheduled, models.QueueStatusDone:
		return s.redisStore.ZCard(ctx, qKey)
	default:
		return s.redisStore.LLen(ctx, qKey)
	}
}
