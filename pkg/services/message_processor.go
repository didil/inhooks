package services

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/version"
	"github.com/pkg/errors"
)

type MessageProcessor interface {
	Process(ctx context.Context, sink *models.Sink, m *models.Message) error
}

type messageProcessor struct {
	httpClient *http.Client
}

func NewMessageProcessor(httpClient *http.Client) MessageProcessor {
	return &messageProcessor{
		httpClient: httpClient,
	}
}

func (p *messageProcessor) Process(ctx context.Context, sink *models.Sink, m *models.Message) error {
	switch sink.Type {
	case models.SinkTypeHttp:
		return p.processHTTP(ctx, sink, m)
	default:
		return fmt.Errorf("unkown sink type %s", sink.Type)
	}
}

func (p *messageProcessor) processHTTP(ctx context.Context, sink *models.Sink, m *models.Message) error {
	buf := bytes.NewBuffer(m.Payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sink.URL, buf)
	if err != nil {
		return errors.Wrapf(err, "failed to build http request. sink: %s m:%s", sink.ID, m.ID)
	}

	req.Header = m.HttpHeaders
	if m.RawQuery != "" {
		req.URL.RawQuery = m.RawQuery
	}

	req.Header["User-Agent"] = []string{p.userAgent()}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to send http request. sink: %s m:%s", sink.ID, m.ID)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.Wrapf(err, "http reponse error %d. sink: %s m:%s", resp.StatusCode, sink.ID, m.ID)
	}

	return nil
}

func (p *messageProcessor) userAgent() string {
	return fmt.Sprintf("Inhooks/v%s (https://github.com/didil/inhooks)", version.GetVersion())
}
