package supervisor

import (
	"context"
	"fmt"
	"testing"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSupervisorFetchAndProcess_OK(t *testing.T) {
	ctx := context.Background()

	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	appConf.Supervisor.ErrSleepTime = 0

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

	messageFetcher.EXPECT().GetMessageForProcessing(ctx, appConf.Supervisor.ReadyWaitTime, flowId1, sinkID1).Return(m, nil)
	messageProcessor.EXPECT().Process(ctx, sink1, m).Return(nil)
	processingResultsService.EXPECT().HandleOK(ctx, m).Return(nil)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithMessageFetcher(messageFetcher),
		WithMessageProcessor(messageProcessor),
		WithProcessingResultsService(processingResultsService),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	err = s.FetchAndProcess(ctx, flow1, sink1)
	assert.NoError(t, err)
}

func TestSupervisorFetchAndProcess_Failed(t *testing.T) {
	ctx := context.Background()

	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	appConf.Supervisor.ErrSleepTime = 0

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

	processingErr := fmt.Errorf("processing error")

	messageFetcher := mocks.NewMockMessageFetcher(ctrl)
	messageProcessor := mocks.NewMockMessageProcessor(ctrl)
	processingResultsService := mocks.NewMockProcessingResultsService(ctrl)

	messageFetcher.EXPECT().GetMessageForProcessing(ctx, appConf.Supervisor.ReadyWaitTime, flowId1, sinkID1).Return(m, nil)
	messageProcessor.EXPECT().Process(ctx, sink1, m).Return(processingErr)
	processingResultsService.EXPECT().HandleFailed(ctx, sink1, m, processingErr).Return(nil)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithMessageFetcher(messageFetcher),
		WithMessageProcessor(messageProcessor),
		WithProcessingResultsService(processingResultsService),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	err = s.FetchAndProcess(ctx, flow1, sink1)
	assert.NoError(t, err)
}
