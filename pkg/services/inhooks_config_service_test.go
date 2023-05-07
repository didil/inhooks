package services

import (
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInhooksConfigService_Load_OK(t *testing.T) {
	logger := zap.NewNop()
	s := NewInhooksConfigService(logger)
	err := s.Load("../testsupport/testdata/inhooksconfig/simple.yml")
	assert.NoError(t, err)

	flow1 := s.FindFlowForSource("source-1-slug")
	assert.NotNil(t, flow1)

	flow1ById := s.GetFlow("flow-1")
	assert.Equal(t, flow1, flow1ById)

	assert.Equal(t, &models.Flow{
		ID: "flow-1",
		Source: &models.Source{
			ID:   "source-1",
			Slug: "source-1-slug",
			Type: "http",
		},
		Sinks: []*models.Sink{
			{
				ID:   "sink-1",
				Type: "http",
				URL:  "https://example.com/sink",
			},
		},
	}, flow1)

	flow2 := s.FindFlowForSource("source-2-slug")
	assert.NotNil(t, flow2)

	flow2ById := s.GetFlow("flow-2")
	assert.Equal(t, flow2, flow2ById)

	assert.Equal(t, &models.Flow{
		ID: "flow-2",
		Source: &models.Source{
			ID:   "source-2",
			Slug: "source-2-slug",
			Type: "http",
		},
		Sinks: []*models.Sink{
			{
				ID:    "sink-2",
				Type:  "http",
				URL:   "https://example.com/sink",
				Delay: 15 * time.Minute,
			},
		},
	}, flow2)

	inexistentFlow := s.FindFlowForSource("source-3-slug")
	assert.Nil(t, inexistentFlow)

	inexistentFlow = s.GetFlow("flow-3")
	assert.Nil(t, inexistentFlow)
}

func TestInhooksConfigService_Load_DupFlow(t *testing.T) {
	logger := zap.NewNop()
	s := NewInhooksConfigService(logger)
	err := s.Load("../testsupport/testdata/inhooksconfig/dup-flow.yml")
	assert.ErrorContains(t, err, "validation err: flow ids must be unique. duplicate flow id: flow-1")
}
