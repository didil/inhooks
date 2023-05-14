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

	m1ID := "a5b6e039-f368-46fd-b0ed-ec9c68932179"
	m1 := &models.Message{
		ID:           m1ID,
		FlowID:       "flow-1",
		SourceID:     "source-1",
		SinkID:       "sink-1",
		RawQuery:     "x=123",
		Payload:      []byte(`{"id":"abc"}`),
		DeliverAfter: now.Add(-1 * time.Second),
	}

	messageKey1 := "f:flow-1:s:sink-1:m:a5b6e039-f368-46fd-b0ed-ec9c68932179"
	queueKey1 := "f:flow-1:s:sink-1:q:ready"
	m1Bytes, err := json.Marshal(&m1)
	assert.NoError(t, err)

	redisStore.EXPECT().
		SetAndEnqueue(ctx, messageKey1, m1Bytes, queueKey1, m1ID).
		Times(1).
		Return(nil)

	m2ID := "6e41b51c-1b90-4b0e-8504-3d0e633f8043"
	m2 := &models.Message{
		ID:           m2ID,
		FlowID:       "flow-1",
		SourceID:     "source-1",
		SinkID:       "sink-2",
		RawQuery:     "x=123",
		Payload:      []byte(`{"id":"abc"}`),
		DeliverAfter: now.Add(30 * time.Second),
	}

	messageKey2 := "f:flow-1:s:sink-2:m:6e41b51c-1b90-4b0e-8504-3d0e633f8043"
	queueKey2 := "f:flow-1:s:sink-2:q:scheduled"
	m2Bytes, err := json.Marshal(&m2)
	assert.NoError(t, err)

	redisStore.EXPECT().
		SetAndZAdd(ctx, messageKey2, m2Bytes, queueKey2, m2ID, float64(m2.DeliverAfter.Unix())).
		Times(1).
		Return(nil)

	queuedInfos, err := messageEnqueuer.Enqueue(ctx, []*models.Message{m1, m2})
	assert.NoError(t, err)

	expectedInfos := []*models.QueuedInfo{
		{MessageID: m1ID, QueueStatus: models.QueueStatusReady, DeliverAfter: m1.DeliverAfter},
		{MessageID: m2ID, QueueStatus: models.QueueStatusScheduled, DeliverAfter: m2.DeliverAfter},
	}

	assert.Equal(t, expectedInfos, queuedInfos)
}
