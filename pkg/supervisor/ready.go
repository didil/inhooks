package supervisor

import (
	"context"

	"github.com/didil/inhooks/pkg/models"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleReadyQueue(ctx context.Context, f *models.Flow, sink *models.Sink) {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			m, err := s.messageFetcher.GetMessageForProcessing(ctx, s.appConf.Supervisor.ReadyWaitTime, f.ID, sink.ID)
			if err != nil {
				s.logger.Error("failed to get message for processing", zap.Error(err))
				continue
			}
			if m == nil {
				// no messages
				continue
			}

			s.logger.Info("processing message", zap.String("flowID", f.ID), zap.String("sinkID", sink.ID), zap.String("sinkType", string(sink.Type)), zap.String("messageID", m.ID))

			processingErr := s.messageProcessor.Process(ctx, sink, m)
			if processingErr != nil {
				err := s.processingResultsSvc.HandleFailed(ctx, sink, m, processingErr)
				if err != nil {
					s.logger.Error("could not handle failed processing", zap.Error(err))
					continue
				}
			} else {
				err := s.processingResultsSvc.HandleOK(ctx, sink, m)
				if err != nil {
					s.logger.Error("failed to handle ok processing", zap.Error(err))
					continue
				}
			}
		}
	}
}
