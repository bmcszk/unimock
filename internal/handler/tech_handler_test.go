package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/service"
)

func TestTechHandler_HealthCheck(t *testing.T) {
	// Create a new tech service and handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	techService := service.NewTechService(time.Now())
	handler := NewTechHandler(techService, logger)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/_uni/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Our handler satisfies http.Handler, so we can call its ServeHTTP method
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	// Verify the response contains expected fields
	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", status)
	}

	if _, ok := response["uptime"]; !ok {
		t.Errorf("Expected uptime field to be present")
	}
}

func TestTechHandler_Metrics(t *testing.T) {
	// Create a new tech service and handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	techService := service.NewTechService(time.Now())
	handler := NewTechHandler(techService, logger)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/_uni/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Our handler satisfies http.Handler, so we can call its ServeHTTP method
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	// Verify the response contains expected fields
	if _, ok := response["request_count"]; !ok {
		t.Errorf("Expected request_count field to be present")
	}

	if _, ok := response["api_endpoints"]; !ok {
		t.Errorf("Expected api_endpoints field to be present")
	}
}

func TestTechHandler_NotFound(t *testing.T) {
	// Create a new tech service and handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	techService := service.NewTechService(time.Now())
	handler := NewTechHandler(techService, logger)

	// Create a request to pass to our handler with an invalid path
	req, err := http.NewRequest("GET", "/_uni/invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Our handler satisfies http.Handler, so we can call its ServeHTTP method
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestTechHandler_MethodNotAllowed(t *testing.T) {
	// Create a new tech service and handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	techService := service.NewTechService(time.Now())
	handler := NewTechHandler(techService, logger)

	// Create a request to pass to our handler with an invalid method
	req, err := http.NewRequest("POST", "/_uni/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Our handler satisfies http.Handler, so we can call its ServeHTTP method
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
