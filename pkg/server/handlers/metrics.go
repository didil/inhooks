package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func (app *App) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	logger := app.logger.With(zap.String("reqID", reqID))

	logger.Info("new metrics request")
	promhttp.Handler().ServeHTTP(w, r)
	logger.Info("metrics request succeeded")
}
