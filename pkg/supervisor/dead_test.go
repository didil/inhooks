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

func TestSupervisor_HandleDeadQueue(t *testing.T) {
	appConf, err := testsupport.InitAppConfig(context.Background())
	assert.NoError(t, err)

	appConf.Supervisor.DeadQueueCleanupInterval = 45 * time.Second
	appConf.Supervisor.DeadQueueCleanupDelay = 24 * time.Hour
	appConf.Supervisor.DeadQueueCleanupEnabled = true

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

	count := 3
	cleanupSvc.EXPECT().
		CleanupDeadQueue(gomock.Any(), flow1, sink1, appConf.Supervisor.DeadQueueCleanupDelay).
		DoAndReturn(func(ctx context.Context, f *models.Flow, sink *models.Sink, deadQueueCleanupDelay time.Duration) (int, error) {
			s.Shutdown()

			return count, nil
		})

	s.HandleDeadQueue(flow1, sink1)
}
