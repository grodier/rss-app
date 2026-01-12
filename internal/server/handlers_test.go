package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/grodier/rss-app/internal/models"
	"github.com/grodier/rss-app/internal/pgsql"
)

// validFeedBody is a shared test fixture for valid feed creation requests
var validFeedBody = `{
	"title": "Test Site",
	"description": "Description for a test feed",
	"url": "https://test.com/rss.xml",
	"site_url": "https://test.com/"
}`

// testServerOptions configures optional dependencies for test server
type testServerOptions struct {
	feedService models.FeedService
	version     string
	env         string
}

// newTestServer creates a Server instance configured for testing.
// Options can be nil for default test configuration.
func newTestServer(opts *testServerOptions) *Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	s := &Server{
		logger: logger,
	}

	if opts != nil {
		if opts.feedService != nil {
			s.FeedService = opts.feedService
		}
		if opts.version != "" {
			s.Version = opts.version
		}
		if opts.env != "" {
			s.Env = opts.env
		}
	}

	return s
}

func TestHandleHealthcheck(t *testing.T) {
	version := "test-version"
	env := "test-env"
	s := newTestServer(&testServerOptions{
		version: version,
		env:     env,
	})

	// Create a new HTTP request for the healthcheck endpoint
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler directly
	s.handleHealthcheck(rr, req)

	// Assert the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Assert the Content-Type header
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	// Assert the response body
	var envelope map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &envelope)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check status field
	expectedStatus := "available"
	if envelope["status"] != expectedStatus {
		t.Errorf("expected status to be %v, got %v", expectedStatus, envelope["status"])
	}

	// Check system_info field
	systemInfo, ok := envelope["system_info"].(map[string]any)
	if !ok {
		t.Fatal("expected system_info to be a map")
	}

	if systemInfo["environment"] != env {
		t.Errorf("expected environment to be %v, got %v", env, systemInfo["environment"])
	}

	if systemInfo["version"] != version {
		t.Errorf("expected version to be %v, got %v", version, systemInfo["version"])
	}
}

func TestHandleCreateFeed_Success(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{},
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/feeds", strings.NewReader(validFeedBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleCreateFeed(rr, req)

	// Assert status
	if rr.Code != http.StatusCreated {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusCreated)
	}

	// Assert headers
	if got := rr.Header().Get("Location"); got != "/v1/feeds/1" {
		t.Errorf("got Location %q, want %q", got, "/v1/feeds/1")
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("got Content-Type %q, want %q", got, "application/json")
	}

	// Assert body structure
	var envelope struct {
		Feed struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
			SiteURL     string `json:"site_url"`
		} `json:"feed"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Check populated fields
	if envelope.Feed.ID != 1 {
		t.Errorf("got id %d, want 1", envelope.Feed.ID)
	}
	if envelope.Feed.Title != "Test Site" {
		t.Errorf("got title %q, want %q", envelope.Feed.Title, "Test Site")
	}
	if envelope.Feed.Description != "Description for a test feed" {
		t.Errorf("got description %q, want %q", envelope.Feed.Description, "Description for a test feed")
	}
	if envelope.Feed.URL != "https://test.com/rss.xml" {
		t.Errorf("got url %q, want %q", envelope.Feed.URL, "https://test.com/rss.xml")
	}
	if envelope.Feed.SiteURL != "https://test.com/" {
		t.Errorf("got site_url %q, want %q", envelope.Feed.SiteURL, "https://test.com/")
	}
}

func TestHandleCreateFeed_JSONParsingErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{"empty body", "", "body must not be empty"},
		{"wrong type for field", `{"title": 123}`, `body contains incorrect JSON type for field "title"`},
		{"array instead of object", `["foo", "bar"]`, "body contains incorrect JSON type (at character 1)"},
		{"malformed json", `{"title": "Moana", }`, "body contains badly-formed JSON (at character 20)"},
		{"unknown field", `{"title": "Test Site", "unknown_field": "value"}`, `body contains unknown key "unknown_field"`},
		{"multiple json values", `{"title": "Test Site"} {"description": "Another description"}`, "body must only contain a single JSON value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(nil)

			req := httptest.NewRequest(http.MethodPost, "/v1/admin/feeds", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.handleCreateFeed(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusBadRequest)
			}

			var resp struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.Error != tt.wantError {
				t.Errorf("got error %q, want %q", resp.Error, tt.wantError)
			}
		})
	}
}

func TestHandleCreateFeed_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantErrors map[string]string
	}{
		{
			name:       "missing title",
			body:       `{"description": "Description for a test feed", "url": "https://test.com/rss.xml", "site_url": "https://test.com/"}`,
			wantErrors: map[string]string{"title": "must be provided"},
		},
		{
			name:       "title too long",
			body:       `{"title": "` + strings.Repeat("a", 501) + `", "description": "Description for a test feed", "url": "https://test.com/rss.xml", "site_url": "https://test.com/"}`,
			wantErrors: map[string]string{"title": "must not be more than 500 bytes long"},
		},
		{
			name:       "missing description",
			body:       `{"title": "Test Site", "url": "https://test.com/rss.xml", "site_url": "https://test.com/"}`,
			wantErrors: map[string]string{"description": "must be provided"},
		},
		{
			name:       "missing url",
			body:       `{"title": "Test Site", "description": "Description for a test feed", "site_url": "https://test.com/"}`,
			wantErrors: map[string]string{"url": "must be provided"},
		},
		{
			name:       "missing site_url",
			body:       `{"title": "Test Site", "description": "Description for a test feed", "url": "https://test.com/rss.xml"}`,
			wantErrors: map[string]string{"site_url": "must be provided"},
		},
		{
			name:       "multiple validation failures",
			body:       `{"title": "", "description": "", "url": "", "site_url": ""}`,
			wantErrors: map[string]string{"title": "must be provided", "description": "must be provided", "url": "must be provided", "site_url": "must be provided"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(nil)

			req := httptest.NewRequest(http.MethodPost, "/v1/admin/feeds", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.handleCreateFeed(rr, req)

			if rr.Code != http.StatusUnprocessableEntity {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusUnprocessableEntity)
			}

			var resp struct {
				Error map[string]string `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			for field, wantMsg := range tt.wantErrors {
				if resp.Error[field] != wantMsg {
					t.Errorf("field %q: got %q, want %q", field, resp.Error[field], wantMsg)
				}
			}

			if len(resp.Error) != len(tt.wantErrors) {
				t.Errorf("got %d errors, want %d", len(resp.Error), len(tt.wantErrors))
			}
		})
	}
}

func TestHandleCreateFeed_ServiceError(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			createFn: func(feed *models.Feed) error {
				return errors.New("database connection failed")
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/feeds", strings.NewReader(validFeedBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleCreateFeed(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	wantError := "the server encountered a problem and could not process your request"
	if resp.Error != wantError {
		t.Errorf("got error %q, want %q", resp.Error, wantError)
	}
}

func TestHandleShowFeed_Success(t *testing.T) {
	expectedFeed := &models.Feed{
		ID:          1,
		Title:       "Test Feed",
		Description: "A test description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
		Language:    "en",
	}

	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				if id != 1 {
					t.Errorf("unexpected id: got %d, want 1", id)
				}
				return expectedFeed, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/feeds/1", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("got Content-Type %q, want %q", got, "application/json")
	}

	var envelope struct {
		Feed struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
			SiteURL     string `json:"site_url"`
			Language    string `json:"language"`
		} `json:"feed"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if envelope.Feed.ID != 1 {
		t.Errorf("got id %d, want 1", envelope.Feed.ID)
	}
	if envelope.Feed.Title != "Test Feed" {
		t.Errorf("got title %q, want %q", envelope.Feed.Title, "Test Feed")
	}
	if envelope.Feed.Description != "A test description" {
		t.Errorf("got description %q, want %q", envelope.Feed.Description, "A test description")
	}
	if envelope.Feed.URL != "https://example.com/feed.xml" {
		t.Errorf("got url %q, want %q", envelope.Feed.URL, "https://example.com/feed.xml")
	}
	if envelope.Feed.SiteURL != "https://example.com" {
		t.Errorf("got site_url %q, want %q", envelope.Feed.SiteURL, "https://example.com")
	}
	if envelope.Feed.Language != "en" {
		t.Errorf("got language %q, want %q", envelope.Feed.Language, "en")
	}
}

func TestHandleShowFeed_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"non-integer", "abc"},
		{"zero", "0"},
		{"negative", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(nil)

			req := httptest.NewRequest(http.MethodGet, "/v1/feeds/"+tt.id, nil)
			rr := httptest.NewRecorder()

			s.router().ServeHTTP(rr, req)

			if rr.Code != http.StatusNotFound {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}

func TestHandleShowFeed_NotFound(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				return nil, pgsql.ErrRecordNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/feeds/999", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandleShowFeed_ServiceError(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				return nil, errors.New("database connection failed")
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/feeds/1", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// validUpdateFeedBody is a shared test fixture for valid partial feed update requests
var validUpdateFeedBody = `{
	"title": "Updated Title",
	"description": "Updated description"
}`

func TestHandleUpdateFeed_Success(t *testing.T) {
	existingFeed := &models.Feed{
		ID:          1,
		Title:       "Original Title",
		Description: "Original description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
		Version:     1,
	}

	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				if id != 1 {
					t.Errorf("unexpected id: got %d, want 1", id)
				}
				// Return a copy to simulate database behavior
				feed := *existingFeed
				return &feed, nil
			},
			updateFn: func(feed *models.Feed) error {
				// Verify the feed has been updated with new values
				if feed.Title != "Updated Title" {
					t.Errorf("expected title to be updated, got %q", feed.Title)
				}
				if feed.Description != "Updated description" {
					t.Errorf("expected description to be updated, got %q", feed.Description)
				}
				// Original values should be preserved (partial update)
				if feed.URL != "https://example.com/feed.xml" {
					t.Errorf("expected URL to be preserved, got %q", feed.URL)
				}
				if feed.SiteURL != "https://example.com" {
					t.Errorf("expected SiteURL to be preserved, got %q", feed.SiteURL)
				}
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/1", strings.NewReader(validUpdateFeedBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("got Content-Type %q, want %q", got, "application/json")
	}

	var envelope struct {
		Feed struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
			SiteURL     string `json:"site_url"`
		} `json:"feed"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if envelope.Feed.ID != 1 {
		t.Errorf("got id %d, want 1", envelope.Feed.ID)
	}
	if envelope.Feed.Title != "Updated Title" {
		t.Errorf("got title %q, want %q", envelope.Feed.Title, "Updated Title")
	}
	if envelope.Feed.Description != "Updated description" {
		t.Errorf("got description %q, want %q", envelope.Feed.Description, "Updated description")
	}
}

func TestHandleUpdateFeed_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"non-integer", "abc"},
		{"zero", "0"},
		{"negative", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(nil)

			req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/"+tt.id, strings.NewReader(validUpdateFeedBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.router().ServeHTTP(rr, req)

			if rr.Code != http.StatusNotFound {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}

func TestHandleUpdateFeed_NotFound(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				return nil, pgsql.ErrRecordNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/999", strings.NewReader(validUpdateFeedBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateFeed_JSONParsingErrors(t *testing.T) {
	existingFeed := &models.Feed{
		ID:          1,
		Title:       "Original Title",
		Description: "Original description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
	}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{"empty body", "", "body must not be empty"},
		{"wrong type for field", `{"title": 123}`, `body contains incorrect JSON type for field "title"`},
		{"array instead of object", `["foo", "bar"]`, "body contains incorrect JSON type (at character 1)"},
		{"malformed json", `{"title": "Updated", }`, "body contains badly-formed JSON (at character 22)"},
		{"unknown field", `{"title": "Updated", "unknown_field": "value"}`, `body contains unknown key "unknown_field"`},
		{"multiple json values", `{"title": "Updated"} {"description": "Another"}`, "body must only contain a single JSON value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(&testServerOptions{
				feedService: &mockFeedService{
					getFn: func(id int64) (*models.Feed, error) {
						feed := *existingFeed
						return &feed, nil
					},
				},
			})

			req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/1", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.router().ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusBadRequest)
			}

			var resp struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.Error != tt.wantError {
				t.Errorf("got error %q, want %q", resp.Error, tt.wantError)
			}
		})
	}
}

func TestHandleUpdateFeed_ValidationErrors(t *testing.T) {
	existingFeed := &models.Feed{
		ID:          1,
		Title:       "Original Title",
		Description: "Original description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
	}

	tests := []struct {
		name       string
		body       string
		wantErrors map[string]string
	}{
		{
			name:       "empty title",
			body:       `{"title": ""}`,
			wantErrors: map[string]string{"title": "must be provided"},
		},
		{
			name:       "title too long",
			body:       `{"title": "` + strings.Repeat("a", 501) + `"}`,
			wantErrors: map[string]string{"title": "must not be more than 500 bytes long"},
		},
		{
			name:       "empty description",
			body:       `{"description": ""}`,
			wantErrors: map[string]string{"description": "must be provided"},
		},
		{
			name:       "empty url",
			body:       `{"url": ""}`,
			wantErrors: map[string]string{"url": "must be provided"},
		},
		{
			name:       "empty site_url",
			body:       `{"site_url": ""}`,
			wantErrors: map[string]string{"site_url": "must be provided"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(&testServerOptions{
				feedService: &mockFeedService{
					getFn: func(id int64) (*models.Feed, error) {
						feed := *existingFeed
						return &feed, nil
					},
				},
			})

			req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/1", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.router().ServeHTTP(rr, req)

			if rr.Code != http.StatusUnprocessableEntity {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusUnprocessableEntity)
			}

			var resp struct {
				Error map[string]string `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			for field, wantMsg := range tt.wantErrors {
				if resp.Error[field] != wantMsg {
					t.Errorf("field %q: got %q, want %q", field, resp.Error[field], wantMsg)
				}
			}

			if len(resp.Error) != len(tt.wantErrors) {
				t.Errorf("got %d errors, want %d", len(resp.Error), len(tt.wantErrors))
			}
		})
	}
}

func TestHandleUpdateFeed_GetServiceError(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				return nil, errors.New("database connection failed")
			},
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/1", strings.NewReader(validUpdateFeedBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	wantError := "the server encountered a problem and could not process your request"
	if resp.Error != wantError {
		t.Errorf("got error %q, want %q", resp.Error, wantError)
	}
}

func TestHandleUpdateFeed_UpdateServiceError(t *testing.T) {
	existingFeed := &models.Feed{
		ID:          1,
		Title:       "Original Title",
		Description: "Original description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
		Version:     1,
	}

	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				feed := *existingFeed
				return &feed, nil
			},
			updateFn: func(feed *models.Feed) error {
				return errors.New("database connection failed")
			},
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/1", strings.NewReader(validUpdateFeedBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	wantError := "the server encountered a problem and could not process your request"
	if resp.Error != wantError {
		t.Errorf("got error %q, want %q", resp.Error, wantError)
	}
}

func TestHandleUpdateFeed_EditConflict(t *testing.T) {
	existingFeed := &models.Feed{
		ID:          1,
		Title:       "Original Title",
		Description: "Original description",
		URL:         "https://example.com/feed.xml",
		SiteURL:     "https://example.com",
		Version:     1,
	}

	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getFn: func(id int64) (*models.Feed, error) {
				feed := *existingFeed
				return &feed, nil
			},
			updateFn: func(feed *models.Feed) error {
				return pgsql.ErrEditConflict
			},
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/feeds/1", strings.NewReader(validUpdateFeedBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusConflict)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	wantError := "unable to update the record due to an edit conflict, please try again"
	if resp.Error != wantError {
		t.Errorf("got error %q, want %q", resp.Error, wantError)
	}
}

func TestHandleDeleteFeed_Success(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			deleteFn: func(id int64) error {
				if id != 1 {
					t.Errorf("unexpected id: got %d, want 1", id)
				}
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/v1/feeds/1", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("got Content-Type %q, want %q", got, "application/json")
	}

	var envelope struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if envelope.Message != "feed successfully deleted" {
		t.Errorf("got message %q, want %q", envelope.Message, "feed successfully deleted")
	}
}

func TestHandleDeleteFeed_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"non-integer", "abc"},
		{"zero", "0"},
		{"negative", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(nil)

			req := httptest.NewRequest(http.MethodDelete, "/v1/feeds/"+tt.id, nil)
			rr := httptest.NewRecorder()

			s.router().ServeHTTP(rr, req)

			if rr.Code != http.StatusNotFound {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}

func TestHandleDeleteFeed_NotFound(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			deleteFn: func(id int64) error {
				return pgsql.ErrRecordNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/v1/feeds/999", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteFeed_ServiceError(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			deleteFn: func(id int64) error {
				return errors.New("database connection failed")
			},
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/v1/feeds/1", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	wantError := "the server encountered a problem and could not process your request"
	if resp.Error != wantError {
		t.Errorf("got error %q, want %q", resp.Error, wantError)
	}
}

func TestHandleListFeeds_Success(t *testing.T) {
	expectedFeeds := []*models.Feed{
		{
			ID:          1,
			Title:       "Test Feed 1",
			Description: "Description 1",
			URL:         "https://example1.com/feed.xml",
			SiteURL:     "https://example1.com",
			Language:    "en",
			Version:     1,
		},
		{
			ID:          2,
			Title:       "Test Feed 2",
			Description: "Description 2",
			URL:         "https://example2.com/feed.xml",
			SiteURL:     "https://example2.com",
			Language:    "es",
			Version:     1,
		},
	}
	expectedMetadata := models.Metadata{
		CurrentPage:  1,
		PageSize:     20,
		FirstPage:    1,
		LastPage:     1,
		TotalRecords: 2,
	}

	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getAllFn: func(title, url string, filters models.Filters) ([]*models.Feed, models.Metadata, error) {
				// Verify default parameters
				if title != "" {
					t.Errorf("expected empty title filter, got %q", title)
				}
				if url != "" {
					t.Errorf("expected empty url filter, got %q", url)
				}
				if filters.Page != 1 {
					t.Errorf("expected page 1, got %d", filters.Page)
				}
				if filters.PageSize != 20 {
					t.Errorf("expected page_size 20, got %d", filters.PageSize)
				}
				if filters.Sort != "id" {
					t.Errorf("expected sort 'id', got %q", filters.Sort)
				}
				return expectedFeeds, expectedMetadata, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/feeds", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("got Content-Type %q, want %q", got, "application/json")
	}

	var envelope struct {
		Feeds []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
			SiteURL     string `json:"site_url"`
			Language    string `json:"language"`
		} `json:"feeds"`
		Metadata struct {
			CurrentPage  int `json:"current_page"`
			PageSize     int `json:"page_size"`
			FirstPage    int `json:"first_page"`
			LastPage     int `json:"last_page"`
			TotalRecords int `json:"total_records"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(envelope.Feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(envelope.Feeds))
	}

	if envelope.Feeds[0].ID != 1 {
		t.Errorf("got id %d, want 1", envelope.Feeds[0].ID)
	}
	if envelope.Feeds[0].Title != "Test Feed 1" {
		t.Errorf("got title %q, want %q", envelope.Feeds[0].Title, "Test Feed 1")
	}

	if envelope.Metadata.CurrentPage != 1 {
		t.Errorf("got current_page %d, want 1", envelope.Metadata.CurrentPage)
	}
	if envelope.Metadata.TotalRecords != 2 {
		t.Errorf("got total_records %d, want 2", envelope.Metadata.TotalRecords)
	}
}

func TestHandleListFeeds_WithFilters(t *testing.T) {
	tests := []struct {
		name          string
		queryString   string
		wantTitle     string
		wantURL       string
		wantPage      int
		wantPageSize  int
		wantSort      string
	}{
		{
			name:          "filter by title",
			queryString:   "?title=technology",
			wantTitle:     "technology",
			wantURL:       "",
			wantPage:      1,
			wantPageSize:  20,
			wantSort:      "id",
		},
		{
			name:          "filter by url",
			queryString:   "?url=https://example.com",
			wantTitle:     "",
			wantURL:       "https://example.com",
			wantPage:      1,
			wantPageSize:  20,
			wantSort:      "id",
		},
		{
			name:          "custom pagination",
			queryString:   "?page=2&page_size=10",
			wantTitle:     "",
			wantURL:       "",
			wantPage:      2,
			wantPageSize:  10,
			wantSort:      "id",
		},
		{
			name:          "sort by title ascending",
			queryString:   "?sort=title",
			wantTitle:     "",
			wantURL:       "",
			wantPage:      1,
			wantPageSize:  20,
			wantSort:      "title",
		},
		{
			name:          "sort by title descending",
			queryString:   "?sort=-title",
			wantTitle:     "",
			wantURL:       "",
			wantPage:      1,
			wantPageSize:  20,
			wantSort:      "-title",
		},
		{
			name:          "sort by url ascending",
			queryString:   "?sort=url",
			wantTitle:     "",
			wantURL:       "",
			wantPage:      1,
			wantPageSize:  20,
			wantSort:      "url",
		},
		{
			name:          "sort by url descending",
			queryString:   "?sort=-url",
			wantTitle:     "",
			wantURL:       "",
			wantPage:      1,
			wantPageSize:  20,
			wantSort:      "-url",
		},
		{
			name:          "sort by id descending",
			queryString:   "?sort=-id",
			wantTitle:     "",
			wantURL:       "",
			wantPage:      1,
			wantPageSize:  20,
			wantSort:      "-id",
		},
		{
			name:          "combined filters and pagination",
			queryString:   "?title=tech&page=3&page_size=5&sort=-title",
			wantTitle:     "tech",
			wantURL:       "",
			wantPage:      3,
			wantPageSize:  5,
			wantSort:      "-title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(&testServerOptions{
				feedService: &mockFeedService{
					getAllFn: func(title, url string, filters models.Filters) ([]*models.Feed, models.Metadata, error) {
						if title != tt.wantTitle {
							t.Errorf("got title %q, want %q", title, tt.wantTitle)
						}
						if url != tt.wantURL {
							t.Errorf("got url %q, want %q", url, tt.wantURL)
						}
						if filters.Page != tt.wantPage {
							t.Errorf("got page %d, want %d", filters.Page, tt.wantPage)
						}
						if filters.PageSize != tt.wantPageSize {
							t.Errorf("got page_size %d, want %d", filters.PageSize, tt.wantPageSize)
						}
						if filters.Sort != tt.wantSort {
							t.Errorf("got sort %q, want %q", filters.Sort, tt.wantSort)
						}
						return []*models.Feed{}, models.Metadata{}, nil
					},
				},
			})

			req := httptest.NewRequest(http.MethodGet, "/v1/feeds"+tt.queryString, nil)
			rr := httptest.NewRecorder()

			s.router().ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleListFeeds_EmptyResult(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getAllFn: func(title, url string, filters models.Filters) ([]*models.Feed, models.Metadata, error) {
				return []*models.Feed{}, models.Metadata{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/feeds", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	var envelope struct {
		Feeds    []any          `json:"feeds"`
		Metadata models.Metadata `json:"metadata"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(envelope.Feeds) != 0 {
		t.Errorf("expected empty feeds array, got %d feeds", len(envelope.Feeds))
	}
}

func TestHandleListFeeds_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantErrors map[string]string
	}{
		{
			name:       "page zero",
			query:      "?page=0",
			wantErrors: map[string]string{"page": "must be greater than zero"},
		},
		{
			name:       "page negative",
			query:      "?page=-1",
			wantErrors: map[string]string{"page": "must be greater than zero"},
		},
		{
			name:       "page too large",
			query:      "?page=10000001",
			wantErrors: map[string]string{"page": "must be a maximum of 10 million"},
		},
		{
			name:       "page_size zero",
			query:      "?page_size=0",
			wantErrors: map[string]string{"page_size": "must be greater than zero"},
		},
		{
			name:       "page_size negative",
			query:      "?page_size=-5",
			wantErrors: map[string]string{"page_size": "must be greater than zero"},
		},
		{
			name:       "page_size too large",
			query:      "?page_size=101",
			wantErrors: map[string]string{"page_size": "must not be more than 100"},
		},
		{
			name:       "invalid sort value",
			query:      "?sort=invalid",
			wantErrors: map[string]string{"sort": "invalid sort value"},
		},
		{
			name:       "invalid sort with dash prefix",
			query:      "?sort=-invalid",
			wantErrors: map[string]string{"sort": "invalid sort value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(&testServerOptions{
				feedService: &mockFeedService{},
			})

			req := httptest.NewRequest(http.MethodGet, "/v1/feeds"+tt.query, nil)
			rr := httptest.NewRecorder()

			s.router().ServeHTTP(rr, req)

			if rr.Code != http.StatusUnprocessableEntity {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusUnprocessableEntity)
			}

			var resp struct {
				Error map[string]string `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			for field, wantMsg := range tt.wantErrors {
				if resp.Error[field] != wantMsg {
					t.Errorf("field %q: got %q, want %q", field, resp.Error[field], wantMsg)
				}
			}
		})
	}
}

func TestHandleListFeeds_ServiceError(t *testing.T) {
	s := newTestServer(&testServerOptions{
		feedService: &mockFeedService{
			getAllFn: func(title, url string, filters models.Filters) ([]*models.Feed, models.Metadata, error) {
				return nil, models.Metadata{}, errors.New("database connection failed")
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/feeds", nil)
	rr := httptest.NewRecorder()

	s.router().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	wantError := "the server encountered a problem and could not process your request"
	if resp.Error != wantError {
		t.Errorf("got error %q, want %q", resp.Error, wantError)
	}
}
