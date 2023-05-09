package services

import (
	"context"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInhooksConfigService_Load_OK(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	logger := zap.NewNop()
	s := NewInhooksConfigService(logger, appConf)
	err = s.Load("../testsupport/testdata/inhooksconfig/simple.yml")
	assert.NoError(t, err)

	flow1 := s.FindFlowForSource("source-1-slug")
	assert.NotNil(t, flow1)

	flow1ById := s.GetFlow("flow-1")
	assert.Equal(t, flow1, flow1ById)

	delay1 := 0 * time.Second
	maxAttempts1 := 3

	assert.Equal(t, &models.Flow{
		ID: "flow-1",
		Source: &models.Source{
			ID:   "source-1",
			Slug: "source-1-slug",
			Type: "http",
		},
		Sinks: []*models.Sink{
			{
				ID:          "sink-1",
				Type:        "http",
				URL:         "https://example.com/sink",
				Delay:       &delay1,
				MaxAttempts: &maxAttempts1,
			},
		},
	}, flow1)

	flow2 := s.FindFlowForSource("source-2-slug")
	assert.NotNil(t, flow2)

	flow2ById := s.GetFlow("flow-2")
	assert.Equal(t, flow2, flow2ById)

	delay2 := 15 * time.Minute
	retryAfter2 := 2 * time.Minute
	maxAttempts2 := 5
	assert.Equal(t, &models.Flow{
		ID: "flow-2",
		Source: &models.Source{
			ID:   "source-2",
			Slug: "source-2-slug",
			Type: "http",
		},
		Sinks: []*models.Sink{
			{
				ID:          "sink-2",
				Type:        "http",
				URL:         "https://example.com/sink",
				Delay:       &delay2,
				RetryAfter:  &retryAfter2,
				MaxAttempts: &maxAttempts2,
			},
		},
	}, flow2)

	inexistentFlow := s.FindFlowForSource("source-3-slug")
	assert.Nil(t, inexistentFlow)

	inexistentFlow = s.GetFlow("flow-3")
	assert.Nil(t, inexistentFlow)

	flows := s.GetFlows()
	assert.Equal(t, map[string]*models.Flow{
		"flow-1": flow1,
		"flow-2": flow2,
	}, flows)

}

func TestInhooksConfigService_Load_DupFlow(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	logger := zap.NewNop()
	s := NewInhooksConfigService(logger, appConf)
	err = s.Load("../testsupport/testdata/inhooksconfig/dup-flow.yml")
	assert.ErrorContains(t, err, "validation err: flow ids must be unique. duplicate flow id: flow-1")
}
