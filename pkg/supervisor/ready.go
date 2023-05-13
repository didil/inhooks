package supervisor

import (
	"context"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleReadyQueue(ctx context.Context, f *models.Flow, sink *models.Sink) {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			err := s.FetchAndProcess(ctx, f, sink)
			if err != nil {
				s.logger.Error("failed to fetch and processed", zap.Error(err))
				// wait before retrying
				time.Sleep(s.appConf.Supervisor.ErrSleepTime)
			}
		}
	}
}

func (s *Supervisor) FetchAndProcess(ctx context.Context, f *models.Flow, sink *models.Sink) error {
	m, err := s.messageFetcher.GetMessageForProcessing(ctx, s.appConf.Supervisor.ReadyWaitTime, f.ID, sink.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to get message for processing")
	}
	if m == nil {
		// no messages
		return nil
	}

	s.logger.Info("processing message", zap.String("flowID", f.ID), zap.String("sinkID", sink.ID), zap.String("sinkType", string(sink.Type)), zap.String("messageID", m.ID))

	processingErr := s.messageProcessor.Process(ctx, sink, m)
	if processingErr != nil {
		err := s.processingResultsSvc.HandleFailed(ctx, sink, m, processingErr)
		if err != nil {
			return errors.Wrapf(err, "could not handle failed processing")
		}
	} else {
		err := s.processingResultsSvc.HandleOK(ctx, m)
		if err != nil {
			return errors.Wrapf(err, "failed to handle ok processing")
		}
	}

	return nil
}
