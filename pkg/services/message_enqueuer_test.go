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

func TestMessageEnqueuer(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisStore := mocks.NewMockRedisStore(ctrl)

	timeSvc := mocks.NewMockTimeService(ctrl)
	now := time.Date(2023, 05, 5, 8, 9, 12, 0, time.UTC)
	timeSvc.EXPECT().Now().Times(2).Return(now)

	messageEnqueuer := NewMessageEnqueuer(redisStore, timeSvc)

	m1 := &models.Message{
		ID:           "a5b6e039-f368-46fd-b0ed-ec9c68932179",
		FlowID:       "flow-1",
		SourceID:     "source-1",
		SinkID:       "sink-1",
		RawQuery:     "x=123",
		Payload:      []byte(`{"id":"abc"}`),
		DeliverAfter: now.Add(-1 * time.Second),
	}

	k1 := "flow:flow-1:sink:sink-1:ready"
	m1Bytes, err := json.Marshal(&m1)
	assert.NoError(t, err)

	redisStore.EXPECT().
		Enqueue(ctx, k1, m1Bytes).
		Times(1).
		Return(nil)

	m2 := &models.Message{
		ID:           "a5b6e039-f368-46fd-b0ed-ec9c68932179",
		FlowID:       "flow-1",
		SourceID:     "source-1",
		SinkID:       "sink-2",
		RawQuery:     "x=123",
		Payload:      []byte(`{"id":"abc"}`),
		DeliverAfter: now.Add(30 * time.Second),
	}

	k2 := "flow:flow-1:sink:sink-2:scheduled"
	m2Bytes, err := json.Marshal(&m2)
	assert.NoError(t, err)

	redisStore.EXPECT().
		Enqueue(ctx, k2, m2Bytes).
		Times(1).
		Return(nil)

	err = messageEnqueuer.Enqueue(ctx, []*models.Message{m1, m2})
	assert.NoError(t, err)
}
