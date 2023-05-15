package supervisor

import (
	"context"
	"testing"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSupervisorMoveProcessingToReady(t *testing.T) {
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

	processingRecoverySvc := mocks.NewMockProcessingRecoveryService(ctrl)
	movedMessageIds := []string{"message-1", "message-2"}
	processingRecoverySvc.EXPECT().MoveProcessingToReady(ctx, flow1, sink1, 2*appConf.Supervisor.ProcessingRecoveryInterval).Return(movedMessageIds, nil)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithProcessingRecoveryService(processingRecoverySvc),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	messageIds, err := s.MoveProcessingToReady(ctx, flow1, sink1)
	assert.NoError(t, err)
	assert.Equal(t, movedMessageIds, messageIds)
}
