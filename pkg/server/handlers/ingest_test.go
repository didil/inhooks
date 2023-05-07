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

func TestUpdateLB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inhooksConfigSvc := mocks.NewMockInhooksConfigService(ctrl)
	messageBuilder := mocks.NewMockMessageBuilder(ctrl)
	messageEnqueuer := mocks.NewMockMessageEnqueuer(ctrl)
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	app := handlers.NewApp(
		handlers.WithLogger(logger),
		handlers.WithInhooksConfigService(inhooksConfigSvc),
		handlers.WithMessageBuilder(messageBuilder),
		handlers.WithMessageEnqueuer(messageEnqueuer),
	)
	r := server.NewRouter(app)
	s := httptest.NewServer(r)
	defer s.Close()

	flow := &models.Flow{
		ID: "flow-1",
	}

	inhooksConfigSvc.EXPECT().FindFlowForSource("my-source").Return(flow)

	messages := []*models.Message{
		{
			ID: "107f942d-f693-45f4-83e6-9a67197bdfe9",
		},
		{
			ID: "cdd3b72a-97b2-447e-b88d-4ae9e43f80a2",
		},
	}

	messageBuilder.EXPECT().FromHttp(flow, gomock.AssignableToTypeOf(&http.Request{})).Return(messages, nil)

	messageEnqueuer.EXPECT().Enqueue(gomock.Any(), messages).Return(nil)

	buf := bytes.NewBufferString(`{"id": "abc"}`)

	req, err := http.NewRequest(http.MethodPost, s.URL+"/api/v1/ingest/my-source", buf)
	assert.NoError(t, err)

	cl := &http.Client{}
	resp, err := cl.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	jsonOK := &handlers.JSONOK{}
	err = json.NewDecoder(resp.Body).Decode(jsonOK)
	assert.NoError(t, err)
}
