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

func TestHandleShowFeed(t *testing.T) {
	tests := []struct {
		name             string
		id               string
		expectedStatus   int
		expectedResponse map[string]any
	}{
		{
			name:           "valid id",
			id:             "1",
			expectedStatus: http.StatusOK,
			expectedResponse: map[string]any{
				"id":          float64(1),
				"title":       "Test Site",
				"description": "Description for a test feed",
				"url":         "https://test.com/rss.xml",
				"site_url":    "https://test.com/",
			},
		},
		{
			name:           "non-integer id",
			id:             "abc",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "id less than 1",
			id:             "0",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "negative id",
			id:             "-5",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(nil)

			req := httptest.NewRequest(http.MethodGet, "/v1/feeds/"+tt.id, nil)
			rr := httptest.NewRecorder()

			// Use the router to properly handle URL parameters
			handler := s.router()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// For valid responses, check JSON structure
			if tt.expectedStatus == http.StatusOK && tt.expectedResponse != nil {
				// Assert the Content-Type header
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("handler returned wrong content type: got %v want %v", contentType, "application/json")
				}

				// Parse the envelope
				var envelope map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &envelope)
				if err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				// Extract the feed from the envelope
				feed, ok := envelope["feed"].(map[string]any)
				if !ok {
					t.Fatal("expected feed to be a map in the envelope")
				}

				// Check all expected fields
				for key, expectedValue := range tt.expectedResponse {
					if feed[key] != expectedValue {
						t.Errorf("expected %s to be %v, got %v", key, expectedValue, feed[key])
					}
				}
			}
		})
	}
}
