package services

import (
	"testing"

	"github.com/didil/inhooks/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestInhooksConfigService_Load_OK(t *testing.T) {
	s := NewInhooksConfigService()
	err := s.Load("../testdata/inhooksconfig/simple.yml")
	assert.NoError(t, err)

	flow := s.FindFlowForSource("source-1-slug")
	assert.NotNil(t, flow)

	assert.Equal(t, "flow-1", flow.ID)
	expectedSource := &models.Source{
		ID:   "source-1",
		Slug: "source-1-slug",
		Type: "http",
	}
	assert.Equal(t, expectedSource, flow.Source)

	assert.Len(t, flow.Sinks, 1)

	sink := flow.Sinks[0]
	expectedSink := &models.Sink{
		ID:   "sink-1",
		Type: "http",
		URL:  "https://example.com/sink",
	}
	assert.Equal(t, expectedSink, sink)

	inexistentFlow := s.FindFlowForSource("flow-2")
	assert.Nil(t, inexistentFlow)
}

func TestInhooksConfigService_Load_DupFlow(t *testing.T) {
	s := NewInhooksConfigService()
	err := s.Load("../testdata/inhooksconfig/dup-flow.yml")
	assert.ErrorContains(t, err, "validation err: flow ids must be unique. duplicate flow id: flow-1")
}
