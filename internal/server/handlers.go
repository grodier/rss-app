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
		Title       *string `json:"title"`
		Description *string `json:"description"`
		URL         *string `json:"url"`
		SiteURL     *string `json:"site_url"`
		Language    *string `json:"language"`
	}

	err = s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		feed.Title = *input.Title
	}

	if input.Description != nil {
		feed.Description = *input.Description
	}

	if input.URL != nil {
		feed.URL = *input.URL
	}

	if input.SiteURL != nil {
		feed.SiteURL = *input.SiteURL
	}

	if input.Language != nil {
		feed.Language = *input.Language
	}

	v := validator.NewValidator()

	if models.ValidateFeed(v, feed); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = s.FeedService.Update(feed)
	if err != nil {
		switch {
		case err == pgsql.ErrEditConflict:
			s.editConflictResponse(w, r)
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

func (s *Server) handleDeleteFeed(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil {
		s.notFoundResponse(w, r)
		return
	}

	err = s.FeedService.Delete(id)
	if err != nil {
		switch {
		case err == pgsql.ErrRecordNotFound:
			s.notFoundResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"message": "feed successfully deleted"}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *Server) handleListFeeds(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		URL   string
		models.Filters
	}

	v := validator.NewValidator()

	qs := r.URL.Query()

	input.Title = s.readString(qs, "title", "")
	input.URL = s.readString(qs, "url", "")

	input.Filters.Page = s.readInt(qs, "page", 1, v)
	input.Filters.PageSize = s.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = s.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "url", "-id", "-title", "-url"}

	if models.ValidateFilters(v, input.Filters); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}
