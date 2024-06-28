package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/didil/inhooks/pkg/models"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type TransformRequest struct {
	Body                string                      `json:"body"`
	Headers             map[string][]string         `json:"headers"`
	TransformDefinition *models.TransformDefinition `json:"transformDefinition"`
}

type TransformResponse struct {
	Body    string              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

func (app *App) HandleTransform(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	logger := app.logger.With(zap.String("reqID", reqID))

	logger.Info("new transform request")

	transformRequest := &TransformRequest{}
	err := json.NewDecoder(r.Body).Decode(transformRequest)
	if err != nil {
		logger.Error("transform request failed: unable to decode request body", zap.Error(err))
		app.WriteJSONErr(w, http.StatusBadRequest, reqID, fmt.Errorf("unable to decode request body"))
		return
	}

	if transformRequest.TransformDefinition == nil {
		logger.Error("transform request failed: transform definition is required")
		app.WriteJSONErr(w, http.StatusBadRequest, reqID, fmt.Errorf("transform definition is required"))
		return
	}

	m := &models.Message{
		Payload:     []byte(transformRequest.Body),
		HttpHeaders: transformRequest.Headers,
	}

	err = app.messageTransformer.Transform(ctx, transformRequest.TransformDefinition, m)
	if err != nil {
		logger.Error("transform request failed: unable to transform message", zap.Error(err))
		app.WriteJSONErr(w, http.StatusInternalServerError, reqID, fmt.Errorf("unable to transform message"))
		return
	}

	transformResponse := &TransformResponse{
		Body:    string(m.Payload),
		Headers: m.HttpHeaders,
	}

	app.WriteJSONResponse(w, http.StatusOK, transformResponse)

	logger.Info("transform request succeeded")

}
