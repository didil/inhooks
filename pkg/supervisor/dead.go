package supervisor

import (
	"time"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleDeadQueue(f *models.Flow, sink *models.Sink) {
	logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))
	for {
		if s.appConf.Supervisor.DeadQueueCleanupEnabled {
			count, err := s.cleanupSvc.CleanupDeadQueue(s.ctx, f, sink, s.appConf.Supervisor.DeadQueueCleanupDelay)
			if err != nil {
				logger.Error("failed to cleanup dead queue", zap.Error(err))
			}
			if count > 0 {
				logger.Info("dead queue cleanup ok. messages removed", zap.Int("messagesCount", count))
			}
		}

		// wait before next check
		timer := time.NewTimer(s.appConf.Supervisor.DeadQueueCleanupInterval)

		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			continue
		}
	}
}
