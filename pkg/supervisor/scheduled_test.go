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

func TestSupervisor_HandleScheduledQueue(t *testing.T) {
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

	schedulerSvc := mocks.NewMockSchedulerService(ctrl)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithSchedulerService(schedulerSvc),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	schedulerSvc.EXPECT().MoveDueScheduled(gomock.Any(), flow1, sink1).
		Do(func(ctx context.Context, f *models.Flow, sink *models.Sink) error {
			s.Shutdown()
			return nil
		})

	s.HandleScheduledQueue(flow1, sink1)
}
