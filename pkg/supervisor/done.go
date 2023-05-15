package supervisor

import (
	"time"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleDoneQueue(f *models.Flow, sink *models.Sink) {
	logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))
	for {
		if s.appConf.Supervisor.DoneQueueCleanupEnabled {
			count, err := s.cleanupSvc.CleanupDoneQueue(s.ctx, f, sink, s.appConf.Supervisor.DoneQueueCleanupDelay)
			if err != nil {
				logger.Error("failed to cleanup done queue", zap.Error(err))
			}
			if count > 0 {
				logger.Info("done queue cleanup ok. messages removed", zap.Int("messagesCount", count))
			}
		}

		// wait before next check
		timer := time.NewTimer(s.appConf.Supervisor.DoneQueueCleanupInterval)

		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			continue
		}
	}
}
