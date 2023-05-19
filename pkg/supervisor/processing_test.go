package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSupervisorHandleProcessingQueue(t *testing.T) {
	appConf, err := testsupport.InitAppConfig(context.Background())
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

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithProcessingRecoveryService(processingRecoverySvc),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	processingRecoverySvc.EXPECT().MoveProcessingToReady(gomock.Any(), flow1, sink1, 2*appConf.Supervisor.ProcessingRecoveryInterval).
		DoAndReturn(func(ctx context.Context, flow *models.Flow, sink *models.Sink, processingRecoveryInterval time.Duration) ([]string, error) {
			s.Shutdown()
			return movedMessageIds, nil
		})

	s.HandleProcessingQueue(flow1, sink1)
}
