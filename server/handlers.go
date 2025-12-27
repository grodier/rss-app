package server

import (
	"fmt"
	"net/http"
)

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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
