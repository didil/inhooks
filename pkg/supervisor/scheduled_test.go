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

func TestSupervisorMoveDueScheduled(t *testing.T) {
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

	schedulerSvc := mocks.NewMockSchedulerService(ctrl)

	schedulerSvc.EXPECT().MoveDueScheduled(ctx, flow1, sink1)

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	s := NewSupervisor(
		WithSchedulerService(schedulerSvc),
		WithAppConfig(appConf),
		WithLogger(logger),
	)

	err = s.MoveDueScheduled(ctx, flow1, sink1)
	assert.NoError(t, err)
}
