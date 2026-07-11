package supervisor

import (
	"time"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleQueueMetrics(f *models.Flow, sink *models.Sink) {
	logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))
	for {
		err := s.queueMetricsSvc.UpdateMetrics(s.ctx, f, sink)
		if err != nil {
			logger.Error("failed to update queue metrics", zap.Error(err))
		}

		// wait before next collection
		timer := time.NewTimer(s.appConf.Supervisor.QueueMetricsInterval)

		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			continue
		}
	}
}
