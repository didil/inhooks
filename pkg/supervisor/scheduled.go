package supervisor

import (
	"time"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleScheduledQueue(f *models.Flow, sink *models.Sink) {
	logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))
	for {
		err := s.schedulerSvc.MoveDueScheduled(s.ctx, f, sink)
		if err != nil {
			logger.Error("failed to move due scheduled", zap.Error(err))
		}

		// wait before next check
		timer := time.NewTimer(s.appConf.Supervisor.SchedulerInterval)

		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			continue
		}
	}
}
