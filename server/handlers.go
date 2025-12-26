package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleCreateFeed(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new feed")
}

func (s *Server) handleShowFeed(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
	}

	fmt.Fprintf(w, "show the details of movie %d\n", id)
}
