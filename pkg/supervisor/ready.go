package supervisor

import (
	"context"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Supervisor) HandleReadyQueue(f *models.Flow, sink *models.Sink) {
	logger := s.logger.With(zap.String("flowID", f.ID), zap.String("sinkID", sink.ID))

	mChan := make(chan *models.Message, s.appConf.Supervisor.ReadyQueueConcurrency)

	for i := 0; i < s.appConf.Supervisor.ReadyQueueConcurrency; i++ {
		go s.startReadyProcessor(s.ctx, f, sink, mChan)
	}

	for {
		m, err := s.messageFetcher.GetMessageForProcessing(s.ctx, s.appConf.Supervisor.ReadyWaitTime, f.ID, sink.ID)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("failed to fetch message for processing", zap.Error(err))

			// wait before retrying
			timer := time.NewTimer(s.appConf.Supervisor.ErrSleepTime)

			select {
			case <-s.ctx.Done():
				return
			case <-timer.C:
				continue
			}
		}

		if m != nil {
			mChan <- m
		}

		// check if channel closed
		select {
		case <-s.ctx.Done():
			return
		default:
			continue
		}
	}
}

func (s *Supervisor) startReadyProcessor(ctx context.Context, f *models.Flow, sink *models.Sink, mChan chan *models.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-mChan:
			logger := s.logger.With(
				zap.String("flowID", f.ID),
				zap.String("sinkID", sink.ID),
				zap.String("sinkType", string(sink.Type)),
				zap.String("messageID", m.ID),
				zap.String("ingestedReqID", m.IngestedReqID),
			)

			logger.Info("processing message", zap.Int("attempt#", len(m.DeliveryAttempts)+1))

			processingErr := s.messageProcessor.Process(ctx, sink, m)
			if processingErr != nil {
				logger.Info("message processing failed")
				queuedInfo, err := s.processingResultsSvc.HandleFailed(ctx, sink, m, processingErr)
				if err != nil {
					logger.Error("could not handle failed processing", zap.Error(err))
					continue
				}
				logger.Info("message queued after failure", zap.String("queue", string(queuedInfo.QueueStatus)), zap.Time("nextAttemptAfter", queuedInfo.DeliverAfter))
			} else {
				logger.Info("message processed ok")
				err := s.processingResultsSvc.HandleOK(ctx, m)
				if err != nil {
					logger.Error("failed to handle ok processing", zap.Error(err))
					continue
				}
				logger.Info("message processed ok - finalized")
			}
		}
	}
}
