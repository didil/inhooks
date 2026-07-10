package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func (app *App) Metrics(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	logger := app.logger.With(zap.String("reqID", reqID))

	logger.Info("new metrics request")
	promhttp.Handler().ServeHTTP(w, r)
	logger.Info("metrics request succeeded")
}
