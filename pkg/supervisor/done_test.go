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

func TestSupervisor_HandleDoneQueue(t *testing.T) {
	appConf, err := testsupport.InitAppConfig(context.Background())
	assert.NoError(t, err)

	appConf.Supervisor.DoneQueueCleanupInterval = 45 * time.Second
	appConf.Supervisor.DoneQueueCleanupDelay = 5 * time.Hour
	appConf.Supervisor.DoneQueueCleanupEnabled = true

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

	cleanupSvc := mocks.NewMockCleanupService(ctrl)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithCleanupService(cleanupSvc),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	count := 2
	cleanupSvc.EXPECT().
		CleanupDoneQueue(gomock.Any(), flow1, sink1, appConf.Supervisor.DoneQueueCleanupDelay).
		DoAndReturn(func(ctx context.Context, f *models.Flow, sink *models.Sink, doneQueueCleanupDelay time.Duration) (int, error) {
			s.Shutdown()

			return count, nil
		})

	s.HandleDoneQueue(flow1, sink1)

}
