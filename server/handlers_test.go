package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandleHealthcheck(t *testing.T) {
	// Create a test server instance with minimal dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	s := &Server{
		logger: logger,
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

	// Assert the response body
	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
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
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id",
			id:             "1",
			expectedStatus: http.StatusOK,
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
		})
	}
}
