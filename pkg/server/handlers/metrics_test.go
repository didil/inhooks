package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/didil/inhooks/pkg/server"
	"github.com/didil/inhooks/pkg/server/handlers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleMetrics(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	app := handlers.NewApp(
		handlers.WithLogger(logger),
	)
	r := server.NewRouter(app)
	s := httptest.NewServer(r)
	defer s.Close()

	resp, err := http.Get(s.URL + "/api/v1/metrics")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8; escaping=values", resp.Header.Get("Content-Type"))
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// test body content
	assert.Contains(t, string(body), "go_goroutines")
}
