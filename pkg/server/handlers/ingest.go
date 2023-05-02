package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func (app *App) HandleIngest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	flowID := chi.URLParam(r, "flowID")
	logger := app.logger.With(zap.String("reqID", reqID), zap.String("flowID", flowID))

	logger.Info("new ingest request")

	// find the flow
	flow := app.inhooksConfigSvc.GetFlow(flowID)
	if flow == nil {
		logger.Error("ingest request failed: unknown flow", zap.String("flowID", flowID))
		app.WriteJSONErr(w, http.StatusNotFound, reqID, fmt.Errorf("unknown flow %s", flowID))
		return
	}

	// decode message
	_, err := app.messageDecoder.FromHttp(flowID, r)
	if err != nil {
		logger.Error("ingest request failed: unable to decode message", zap.Error(err))
		app.WriteJSONErr(w, http.StatusBadRequest, reqID, fmt.Errorf("unable to read data"))
		return
	}

	// TODO: enqueue message

	app.WriteJSONResponse(w, http.StatusOK, JSONOK{})
	logger.Info("ingest request succeeded")
}
