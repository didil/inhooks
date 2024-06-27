package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestMessageProcessor(t *testing.T) {
	version.SetVersion("test")

	ctx := context.Background()
	cl := &http.Client{}
	p := NewMessageProcessor(cl)

	payload := []byte(`{"id": "the-payload"}`)
	transformedPayload := []byte(`{"id": "the-transformed-payload"}`)

	headers := http.Header{
		"X-Key":           []string{"123"},
		"User-Agent":      []string{"Sender-User-Agent"},
		"Content-Length":  []string{"21"},
		"Accept-Encoding": []string{"*"},
	}
	rawQuery := "k1=v1&k2=v2"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, transformedPayload, body)

		assert.Equal(t, rawQuery, req.URL.RawQuery)

		assert.Equal(t, http.Header{
			"X-Key":           []string{"123"},
			"User-Agent":      []string{"Inhooks/test (https://github.com/didil/inhooks)"},
			"Content-Length":  []string{"33"},
			"Accept-Encoding": []string{"*"},
		}, req.Header)
	}))
	defer s.Close()

	sink := &models.Sink{
		Type: "http",
		URL:  s.URL,
	}

	m := &models.Message{
		HttpHeaders: headers,
		RawQuery:    rawQuery,
		Payload:     payload,
	}

	err := p.Process(ctx, sink, m, transformedPayload)
	assert.NoError(t, err)
}

func TestMessageProcessor_userAgent(t *testing.T) {
	version.SetVersion("1.2.3")

	p := &messageProcessor{}

	expectedUserAgent := "Inhooks/1.2.3 (https://github.com/didil/inhooks)"
	actualUserAgent := p.userAgent()

	assert.Equal(t, expectedUserAgent, actualUserAgent)
}
