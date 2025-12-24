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
