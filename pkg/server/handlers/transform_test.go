package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/server"
	"github.com/didil/inhooks/pkg/server/handlers"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleTransform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	messageTransformer := mocks.NewMockMessageTransformer(ctrl)

	app := handlers.NewApp(
		handlers.WithLogger(logger),
		handlers.WithMessageTransformer(messageTransformer),
	)
	r := server.NewRouter(app)
	s := httptest.NewServer(r)
	defer s.Close()

	transformDefinition := &models.TransformDefinition{
		Type:   models.TransformTypeJavascript,
		Script: "function transform(bodyStr, headers) { return [bodyStr.toUpperCase(), headers]; }",
	}

	requestBody := handlers.TransformRequest{
		Body:                "hello world",
		Headers:             map[string][]string{"Content-Type": {"text/plain"}},
		TransformDefinition: transformDefinition,
	}

	messageTransformer.EXPECT().Transform(
		gomock.Any(),
		transformDefinition,
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, td *models.TransformDefinition, m *models.Message) error {
		m.Payload = []byte("HELLO WORLD")
		return nil
	})

	buf, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, s.URL+"/api/v1/transform", bytes.NewBuffer(buf))
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var transformResponse handlers.TransformResponse
	err = json.NewDecoder(resp.Body).Decode(&transformResponse)
	assert.NoError(t, err)

	assert.Equal(t, "HELLO WORLD", transformResponse.Body)
	assert.Equal(t, map[string][]string{"Content-Type": {"text/plain"}}, transformResponse.Headers)
}

func TestHandleTransform_BadRequest(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	app := handlers.NewApp(
		handlers.WithLogger(logger),
	)
	r := server.NewRouter(app)
	s := httptest.NewServer(r)
	defer s.Close()

	req, err := http.NewRequest(http.MethodPost, s.URL+"/api/v1/transform", bytes.NewBufferString("invalid json"))
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var jsonErr handlers.JSONErr
	err = json.NewDecoder(resp.Body).Decode(&jsonErr)
	assert.NoError(t, err)
	assert.Contains(t, jsonErr.Error, "unable to decode request body")
}

func TestHandleTransform_MissingTransformDefinition(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	app := handlers.NewApp(
		handlers.WithLogger(logger),
	)
	r := server.NewRouter(app)
	s := httptest.NewServer(r)
	defer s.Close()

	requestBody := handlers.TransformRequest{
		Body:    "hello world",
		Headers: map[string][]string{"Content-Type": {"text/plain"}},
	}

	buf, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, s.URL+"/api/v1/transform", bytes.NewBuffer(buf))
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var jsonErr handlers.JSONErr
	err = json.NewDecoder(resp.Body).Decode(&jsonErr)
	assert.NoError(t, err)
	assert.Contains(t, jsonErr.Error, "transform definition is required")
}
