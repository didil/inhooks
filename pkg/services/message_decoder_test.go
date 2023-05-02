package services

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessageDecoderFromHttp_OK(t *testing.T) {
	flowID := "flow-1"
	jsonPayload := []byte(`{"id":"1234","status":"complete"}`)
	b := bytes.NewBuffer(jsonPayload)
	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/ingest/%s", flowID), b)

	r.Header = http.Header{
		"header-1": []string{"abc"},
		"header-2": []string{"def"},
	}

	d := NewMessageDecoder()
	message, err := d.FromHttp(flowID, r)
	assert.NoError(t, err)

	_, err = uuid.Parse(message.ID)
	assert.NoError(t, err)

	assert.Equal(t, flowID, message.FlowID)

	assert.Equal(t, r.Header, message.HttpHeaders)

	assert.Equal(t, jsonPayload, message.Payload)
}
