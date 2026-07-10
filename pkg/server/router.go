package server

import (
	"github.com/didil/inhooks/api"
	"github.com/didil/inhooks/pkg/server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(app *handlers.App) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	api.HandlerWithOptions(app, api.ChiServerOptions{
		BaseURL:    "/api/v1",
		BaseRouter: r,
	})
	return r
}
