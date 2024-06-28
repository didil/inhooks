package server

import (
	"github.com/didil/inhooks/pkg/server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(app *handlers.App) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/ingest/{sourceSlug}", app.HandleIngest)

		r.Post("/transform", app.HandleTransform)
		r.Get("/metrics", app.HandleMetrics)
	})

	return r
}
