package supervisor

import (
	"context"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleScheduledQueue(ctx context.Context, f *models.Flow, sink *models.Sink) {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))
			err := s.MoveDueScheduled(ctx, f, sink)
			if err != nil {
				logger.Error("failed to move due scheduled", zap.Error(err))
			}
			// wait before next check
			time.Sleep(s.appConf.Supervisor.SchedulerInterval)
		}
	}
}

func (s *Supervisor) MoveDueScheduled(ctx context.Context, f *models.Flow, sink *models.Sink) error {
	err := s.schedulerSvc.MoveDueScheduled(ctx, f, sink)
	if err != nil {
		return err
	}

	return nil
}
