package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMessageFetcher(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)
	timeSvc := mocks.NewMockTimeService(ctrl)

	messageFetcher := NewMessageFetcher(redisStore, timeSvc)

	flowID := "flow-1"
	sinkID := "sink-1"
	message := &models.Message{
		ID: "8d291081-a0ea-4511-9445-35f231d1c676",
	}
	timeout := 1 * time.Second

	messageBytes, err := json.Marshal(message)
	assert.NoError(t, err)

	redisStore.EXPECT().BLMove(ctx, timeout, "flow:flow-1:sink:sink-1:ready", "flow:flow-1:sink:sink-1:processing").Return(messageBytes, nil)

	m, err := messageFetcher.GetMessageForProcessing(ctx, timeout, flowID, sinkID)
	assert.NoError(t, err)
	assert.Equal(t, message, m)
}
