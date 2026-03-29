package router

import (
	"docqa/internal/handlers"

	localMiddlweare "docqa/internal/middlweare"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(localMiddlweare.RequestLogger)

	r.Get("/health", h.Health)
	r.Post("/upload", h.Upload)
	r.Post("/ask", h.Ask)

	return r
}
