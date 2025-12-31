package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleHealthcheck(t *testing.T) {
	// Create a test server instance with minimal dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	version := "test-version"
	env := "test-env"
	s := &Server{
		logger:  logger,
		Version: version,
		Env:     env,
	}

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

func TestHandleCreateFeed(t *testing.T) {
	tests := []struct {
		name             string
		body             string
		expectedStatus   int
		expectedResponse string
		expectedErrors   map[string]string
	}{
		{
			name: "valid feed creation",
			body: `{
				"title": "Test Site",
				"description": "Description for a test feed",
				"url": "https://test.com/rss.xml",
				"site_url": "https://test.com/"
			}`,
			expectedStatus:   http.StatusOK,
			expectedResponse: "{Title:Test Site Description:Description for a test feed URL:https://test.com/rss.xml SiteURL:https://test.com/}\n",
		},
		{
			name:             "empty body feed creation",
			body:             "",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"body must not be empty"}` + "\n",
		},
		{
			name:             "incorrect content type",
			body:             `{"title": 123}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"body contains incorrect JSON type for field \"title\""}` + "\n",
		},
		{
			name:             "incorrect json type",
			body:             `["foo", "bar"]`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"body contains incorrect JSON type (at character 1)"}` + "\n",
		},
		{
			name:             "malformed json",
			body:             `{"title": "Moana", }`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"body contains badly-formed JSON (at character 20)"}` + "\n",
		},
		{
			name:             "unknown field in json",
			body:             `{"title": "Test Site", "unknown_field": "value"}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"body contains unknown key \"unknown_field\""}` + "\n",
		},
		{
			name:             "multiple json values",
			body:             `{"title": "Test Site"} {"description": "Another description"}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"body must only contain a single JSON value"}` + "\n",
		},
		{
			name: "missing title",
			body: `{
				"description": "Description for a test feed",
				"url": "https://test.com/rss.xml",
				"site_url": "https://test.com/"
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedErrors: map[string]string{
				"title": "must be provided",
			},
		},
		{
			name: "title too long",
			body: `{
				"title": "` + strings.Repeat("a", 501) + `",
				"description": "Description for a test feed",
				"url": "https://test.com/rss.xml",
				"site_url": "https://test.com/"
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedErrors: map[string]string{
				"title": "must not be more than 500 bytes long",
			},
		},
		{
			name: "missing description",
			body: `{
				"title": "Test Site",
				"url": "https://test.com/rss.xml",
				"site_url": "https://test.com/"
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedErrors: map[string]string{
				"description": "must be provided",
			},
		},
		{
			name: "missing url",
			body: `{
				"title": "Test Site",
				"description": "Description for a test feed",
				"site_url": "https://test.com/"
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedErrors: map[string]string{
				"url": "must be provided",
			},
		},
		{
			name: "missing site_url",
			body: `{
				"title": "Test Site",
				"description": "Description for a test feed",
				"url": "https://test.com/rss.xml"
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedErrors: map[string]string{
				"site_url": "must be provided",
			},
		},
		{
			name: "multiple validation failures",
			body: `{
				"title": "",
				"description": "",
				"url": "",
				"site_url": ""
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedErrors: map[string]string{
				"title":       "must be provided",
				"description": "must be provided",
				"url":         "must be provided",
				"site_url":    "must be provided",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			s := &Server{
				logger: logger,
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/admin/feeds", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			s.handleCreateFeed(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedErrors != nil {
				// Parse the response to check for validation errors
				var envelope map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &envelope)
				if err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				errors, ok := envelope["error"].(map[string]any)
				if !ok {
					t.Fatal("expected errors to be a map in the response")
				}

				// Check that all expected errors are present
				for key, expectedMsg := range tt.expectedErrors {
					actualMsg, exists := errors[key]
					if !exists {
						t.Errorf("expected error for field %s, but it was not present", key)
						continue
					}
					if actualMsg != expectedMsg {
						t.Errorf("expected error for %s to be '%s', got '%s'", key, expectedMsg, actualMsg)
					}
				}

				// Check that no unexpected errors are present
				if len(errors) != len(tt.expectedErrors) {
					t.Errorf("expected %d errors, got %d", len(tt.expectedErrors), len(errors))
				}
			} else if tt.expectedResponse != "" {
				if rr.Body.String() != tt.expectedResponse {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedResponse)
				}
			}
		})
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
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			s := &Server{
				logger: logger,
			}

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
