package handlers

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func (app *App) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}
