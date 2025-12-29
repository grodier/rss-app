package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/grodier/rss-app/models"
)

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": s.Env,
			"version":     s.Version,
		},
	}

	err := s.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
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

	feed := models.Feed{
		ID:          id,
		Title:       "Test Site",
		Description: "Description for a test feed",
		URL:         "https://test.com/rss.xml",
		SiteURL:     "https://test.com/",
		CreatedAt:   time.Now(),
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"feed": feed}, nil)
	if err != nil {
		s.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
