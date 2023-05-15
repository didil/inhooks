package supervisor

import (
	"time"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

// move stuck messages from processing to ready queue periodically
func (s *Supervisor) HandleProcessingQueue(f *models.Flow, sink *models.Sink) {
	logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))
	for {
		// cache keys for twice the processing recovery interval
		// this avoids the recovery process from interfering with legitimate retry attempts
		ttl := 2 * s.appConf.Supervisor.ProcessingRecoveryInterval
		movedMessageIds, err := s.processingRecoverySvc.MoveProcessingToReady(s.ctx, f, sink, ttl)
		if err != nil {
			logger.Error("failed to move processing to ready", zap.Error(err))
		}
		if len(movedMessageIds) > 0 {
			logger.Info("moved stuck messages from processing to ready", zap.Strings("messageIDs", movedMessageIds))
		}

		// wait before next check
		timer := time.NewTimer(s.appConf.Supervisor.ProcessingRecoveryInterval)

		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			continue
		}
	}
}
