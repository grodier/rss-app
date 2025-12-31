package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/grodier/rss-app/internal/models"
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
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		SiteURL     string `json:"site_url"`
	}

	err := s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (s *Server) handleShowFeed(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil {
		s.notFoundResponse(w, r)
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
		s.serverErrorResponse(w, r, err)
	}
}
