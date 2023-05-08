package services

import (
	"io"
	"net/http"

	"github.com/didil/inhooks/pkg/models"
	"github.com/google/uuid"
)

type MessageBuilder interface {
	FromHttp(flow *models.Flow, r *http.Request) ([]*models.Message, error)
}

type messageBuilder struct {
	timeSvc TimeService
}

func NewMessageBuilder(timeSvc TimeService) MessageBuilder {
	return &messageBuilder{
		timeSvc: timeSvc,
	}
}

func (b *messageBuilder) FromHttp(flow *models.Flow, r *http.Request) ([]*models.Message, error) {
	httpHeaders := r.Header
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	query := r.URL.RawQuery

	messages := []*models.Message{}

	for _, s := range flow.Sinks {
		m := &models.Message{}

		m.FlowID = flow.ID
		m.SourceID = flow.Source.ID
		m.SinkID = s.ID
		m.ID = uuid.New().String()
		m.HttpHeaders = httpHeaders
		m.RawQuery = query
		m.Payload = payload

		// init processing info
		if s.Delay != nil {
			m.DeliverAfter = b.timeSvc.Now().Add(*s.Delay)
		}

		messages = append(messages, m)
	}

	return messages, nil
}
