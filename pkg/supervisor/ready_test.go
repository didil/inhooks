package supervisor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSupervisorHandleReadyQueue_OK(t *testing.T) {
	ctx := context.Background()

	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	appConf.Supervisor.ErrSleepTime = 0
	appConf.Supervisor.ReadyQueueConcurrency = 1

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowId1 := "flow-1"
	sinkID1 := "sink-1"

	sink1 := &models.Sink{
		ID: sinkID1,
	}

	flow1 := &models.Flow{
		ID:    flowId1,
		Sinks: []*models.Sink{sink1},
	}

	mID1 := "message-1"

	m := &models.Message{
		ID: mID1,
	}

	messageFetcher := mocks.NewMockMessageFetcher(ctrl)
	messageProcessor := mocks.NewMockMessageProcessor(ctrl)
	processingResultsService := mocks.NewMockProcessingResultsService(ctrl)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithMessageFetcher(messageFetcher),
		WithMessageProcessor(messageProcessor),
		WithProcessingResultsService(processingResultsService),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	fetcherCallCount := 0

	messageFetcher.EXPECT().
		GetMessageForProcessing(gomock.Any(), appConf.Supervisor.ReadyWaitTime, flowId1, sinkID1).AnyTimes().
		DoAndReturn(func(ctx context.Context, timeout time.Duration, flowID string, sinkID string) (*models.Message, error) {
			fetcherCallCount++

			if fetcherCallCount == 1 {
				return m, nil
			}

			return nil, nil
		})

	messageProcessor.EXPECT().
		Process(gomock.Any(), sink1, m).
		DoAndReturn(func(ctx context.Context, sink *models.Sink, m *models.Message) error {
			return nil
		})

	processingResultsService.EXPECT().HandleOK(gomock.Any(), m).DoAndReturn(func(ctx context.Context, m *models.Message) error {
		s.Shutdown()
		return nil
	})

	s.HandleReadyQueue(flow1, sink1)
}

func TestSupervisorHandleReadyQueue_Failed(t *testing.T) {
	ctx := context.Background()

	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	appConf.Supervisor.ErrSleepTime = 0
	appConf.Supervisor.ReadyQueueConcurrency = 1

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flowId1 := "flow-1"
	sinkID1 := "sink-1"

	sink1 := &models.Sink{
		ID: sinkID1,
	}

	flow1 := &models.Flow{
		ID:    flowId1,
		Sinks: []*models.Sink{sink1},
	}

	mID1 := "message-1"

	m := &models.Message{
		ID: mID1,
	}

	messageFetcher := mocks.NewMockMessageFetcher(ctrl)
	messageProcessor := mocks.NewMockMessageProcessor(ctrl)
	processingResultsService := mocks.NewMockProcessingResultsService(ctrl)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithMessageFetcher(messageFetcher),
		WithMessageProcessor(messageProcessor),
		WithProcessingResultsService(processingResultsService),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	processingErr := fmt.Errorf("processing error")

	fetcherCallCount := 0
	messageFetcher.EXPECT().
		GetMessageForProcessing(gomock.Any(), appConf.Supervisor.ReadyWaitTime, flowId1, sinkID1).AnyTimes().
		DoAndReturn(func(ctx context.Context, timeout time.Duration, flowID string, sinkID string) (*models.Message, error) {
			fetcherCallCount++

			if fetcherCallCount == 1 {
				return m, nil
			}

			return nil, nil
		})

	messageProcessor.EXPECT().
		Process(gomock.Any(), sink1, m).
		DoAndReturn(func(ctx context.Context, sink *models.Sink, m *models.Message) error {
			return processingErr
		})

	processingResultsService.EXPECT().
		HandleFailed(gomock.Any(), sink1, m, processingErr).
		DoAndReturn(func(ctx context.Context, sink *models.Sink, m *models.Message, processingErr error) (*models.QueuedInfo, error) {
			s.Shutdown()

			return &models.QueuedInfo{QueueStatus: models.QueueStatusReady}, nil
		})

	s.HandleReadyQueue(flow1, sink1)
}
