package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) router() http.Handler {
	router := chi.NewRouter()

	router.Get("/v1/healthcheck", s.handleHealthcheck)
	router.Post("/v1/admin/feeds", s.handleCreateFeed)
	router.Get("/v1/feeds/{id}", s.handleShowFeed)
	return router
}
