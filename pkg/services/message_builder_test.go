package services

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessageBuilderFromHttp_OK(t *testing.T) {
	flowID := "flow-1"
	sourceID := "source-1"
	sink1ID := "sink-1"
	sink2ID := "sink-2"
	jsonPayload := []byte(`{"id":"1234","status":"complete"}`)
	b := bytes.NewBuffer(jsonPayload)
	rawQuery := "x=abc&yz=this%20is%20ok"
	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/ingest/%s?%s", flowID, rawQuery), b)

	r.Header = http.Header{
		"header-1": []string{"abc"},
		"header-2": []string{"def"},
	}

	flow := &models.Flow{
		ID: flowID,
		Source: &models.Source{
			ID: sourceID,
		},
		Sinks: []*models.Sink{
			{
				ID: sink1ID,
			},
			{
				ID:    sink2ID,
				Delay: 5 * time.Minute,
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeSvc := mocks.NewMockTimeService(ctrl)
	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().Times(2).Return(now)

	d := NewMessageBuilder(timeSvc)
	messages, err := d.FromHttp(flow, r)
	assert.NoError(t, err)

	m1 := messages[0]

	_, err = uuid.Parse(m1.ID)
	assert.NoError(t, err)

	assert.Equal(t, flowID, m1.FlowID)
	assert.Equal(t, sourceID, m1.SourceID)
	assert.Equal(t, sink1ID, m1.SinkID)
	assert.Equal(t, rawQuery, m1.RawQuery)
	assert.Equal(t, r.Header, m1.HttpHeaders)
	assert.Equal(t, jsonPayload, m1.Payload)
	assert.Equal(t, now, m1.DeliverAfter)

	m2 := messages[1]

	_, err = uuid.Parse(m2.ID)
	assert.NoError(t, err)

	assert.Equal(t, flowID, m2.FlowID)
	assert.Equal(t, sourceID, m2.SourceID)
	assert.Equal(t, sink2ID, m2.SinkID)
	assert.Equal(t, rawQuery, m2.RawQuery)
	assert.Equal(t, r.Header, m2.HttpHeaders)
	assert.Equal(t, jsonPayload, m2.Payload)
	assert.Equal(t, now.Add(5*time.Minute), m2.DeliverAfter)
}
