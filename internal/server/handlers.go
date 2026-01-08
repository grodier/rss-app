package server

import (
	"fmt"
	"net/http"

	"github.com/grodier/rss-app/internal/models"
	"github.com/grodier/rss-app/internal/pgsql"
	"github.com/grodier/rss-app/internal/validator"
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

	feed := &models.Feed{
		Title:       input.Title,
		Description: input.Description,
		URL:         input.URL,
		SiteURL:     input.SiteURL,
	}

	v := validator.NewValidator()

	if models.ValidateFeed(v, feed); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = s.FeedService.Create(feed)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/feeds/%d", feed.ID))

	err = s.writeJSON(w, http.StatusCreated, envelope{"feed": feed}, headers)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *Server) handleShowFeed(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil {
		s.notFoundResponse(w, r)
		return
	}

	feed, err := s.FeedService.Get(id)
	if err != nil {
		switch {
		case err == pgsql.ErrRecordNotFound:
			s.notFoundResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"feed": feed}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *Server) handleUpdateFeed(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil {
		s.notFoundResponse(w, r)
		return
	}

	feed, err := s.FeedService.Get(id)
	if err != nil {
		switch {
		case err == pgsql.ErrRecordNotFound:
			s.notFoundResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		SiteURL     string `json:"site_url"`
		Language    string `json:"language"`
	}

	err = s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	feed.Title = input.Title
	feed.Description = input.Description
	feed.URL = input.URL
	feed.SiteURL = input.SiteURL
	feed.Language = input.Language

	v := validator.NewValidator()

	if models.ValidateFeed(v, feed); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = s.FeedService.Update(feed)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"feed": feed}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}
