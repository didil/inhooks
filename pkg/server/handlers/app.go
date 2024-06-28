package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/didil/inhooks/pkg/services"
	"go.uber.org/zap"
)

type App struct {
	logger             *zap.Logger
	inhooksConfigSvc   services.InhooksConfigService
	messageBuilder     services.MessageBuilder
	messageEnqueuer    services.MessageEnqueuer
	messageVerifier    services.MessageVerifier
	messageTransformer services.MessageTransformer
}

type AppOpt func(app *App)

func NewApp(opts ...AppOpt) *App {
	app := &App{}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func WithLogger(logger *zap.Logger) AppOpt {
	return func(app *App) {
		app.logger = logger
	}
}

func WithInhooksConfigService(inhooksConfigSvc services.InhooksConfigService) AppOpt {
	return func(app *App) {
		app.inhooksConfigSvc = inhooksConfigSvc
	}
}

func WithMessageBuilder(messageBuilder services.MessageBuilder) AppOpt {
	return func(app *App) {
		app.messageBuilder = messageBuilder
	}
}

func WithMessageEnqueuer(messageEnqueuer services.MessageEnqueuer) AppOpt {
	return func(app *App) {
		app.messageEnqueuer = messageEnqueuer
	}
}

func WithMessageVerifier(messageVerifier services.MessageVerifier) AppOpt {
	return func(app *App) {
		app.messageVerifier = messageVerifier
	}
}

func WithMessageTransformer(messageTransformer services.MessageTransformer) AppOpt {
	return func(app *App) {
		app.messageTransformer = messageTransformer
	}
}

type JSONErr struct {
	Error string `json:"error"`
	ReqID string `json:"reqID,omitempty"`
}

type JSONOK struct {
}

func (app *App) WriteJSONErr(w http.ResponseWriter, statusCode int, reqID string, err error) {
	jsonErr := &JSONErr{
		Error: err.Error(),
		ReqID: reqID,
	}
	app.WriteJSONResponse(w, statusCode, jsonErr)
}

func (app *App) WriteJSONResponse(w http.ResponseWriter, statusCode int, resp any) {
	w.WriteHeader(statusCode)
	writeErr := json.NewEncoder(w).Encode(resp)
	if writeErr != nil {
		app.logger.Error("json write err", zap.Error(writeErr))
	}
}
