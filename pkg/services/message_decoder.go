package services

import (
	"io"
	"net/http"

	"github.com/didil/inhooks/pkg/models"
	"github.com/google/uuid"
)

type MessageDecoder interface {
	FromHttp(flow *models.Flow, r *http.Request) (*models.Message, error)
}

type messageDecoder struct {
}

func NewMessageDecoder() MessageDecoder {
	return &messageDecoder{}
}

func (d *messageDecoder) FromHttp(flow *models.Flow, r *http.Request) (*models.Message, error) {
	m := &models.Message{}

	m.FlowID = flow.ID
	m.ID = uuid.New().String()
	m.HttpHeaders = r.Header

	var err error
	m.Payload, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return m, nil
}
