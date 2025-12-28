package server

import (
	"fmt"
	"net/http"
)

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": s.Env,
		"version":     s.Version,
	}

	err := s.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		s.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

func (s *Server) handleCreateFeed(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new feed")
}

func (s *Server) handleShowFeed(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "show the details of feed %d\n", id)
}
