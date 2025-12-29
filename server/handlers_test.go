package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	s := &Server{
		logger: logger,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/feeds", nil)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler directly
	s.handleCreateFeed(rr, req)

	// Assert the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Assert the response body
	expected := "create a new feed\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
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
