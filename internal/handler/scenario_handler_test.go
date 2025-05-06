package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bmcszk/unimock/internal/model"
	"github.com/bmcszk/unimock/internal/storage"
)

func TestScenarioHandler_Create(t *testing.T) {
	// Create a new storage for testing
	storage := storage.NewScenarioStorage()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := NewScenarioHandler(storage, logger)

	// Create a new scenario
	scenario := model.Scenario{
		RequestPath: "GET /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Hello, World!"}`,
	}

	// Convert scenario to JSON
	body, err := json.Marshal(scenario)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request
	req, err := http.NewRequest("POST", "/_uni/scenarios", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder
	rr := httptest.NewRecorder()

	// Serve HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check response
	var response model.Scenario
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	// Verify the created scenario has a UUID
	if response.UUID == "" {
		t.Error("Expected UUID to be present in response")
	}

	// Verify the path was preserved
	if response.RequestPath != scenario.RequestPath {
		t.Errorf("Expected requestPath %s, got %s", scenario.RequestPath, response.RequestPath)
	}
}

func TestScenarioHandler_Get(t *testing.T) {
	// Create a new storage for testing
	storage := storage.NewScenarioStorage()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := NewScenarioHandler(storage, logger)

	// Create a new scenario
	scenario := model.Scenario{
		RequestPath: "GET /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Hello, World!"}`,
	}

	// Create the scenario
	body, _ := json.Marshal(scenario)
	req, _ := http.NewRequest("POST", "/_uni/scenarios", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Get the created scenario's UUID
	var createdScenario model.Scenario
	json.Unmarshal(rr.Body.Bytes(), &createdScenario)
	uuid := createdScenario.UUID

	// Now test the GET endpoint
	req, err := http.NewRequest("GET", "/_uni/scenarios/"+uuid, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response
	var response model.Scenario
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	// Verify the scenario data
	if response.UUID != uuid {
		t.Errorf("Expected UUID %s, got %s", uuid, response.UUID)
	}
	if response.RequestPath != scenario.RequestPath {
		t.Errorf("Expected requestPath %s, got %s", scenario.RequestPath, response.RequestPath)
	}
}

func TestScenarioHandler_List(t *testing.T) {
	// Create a new storage for testing
	storage := storage.NewScenarioStorage()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := NewScenarioHandler(storage, logger)

	// Create multiple scenarios
	scenarios := []model.Scenario{
		{
			RequestPath: "GET /api/test1",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"message":"Hello, World 1!"}`,
		},
		{
			RequestPath: "GET /api/test2",
			StatusCode:  201,
			ContentType: "application/json",
			Data:        `{"message":"Hello, World 2!"}`,
		},
	}

	// Create each scenario
	for _, scenario := range scenarios {
		body, _ := json.Marshal(scenario)
		req, _ := http.NewRequest("POST", "/_uni/scenarios", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Now test the list endpoint
	req, err := http.NewRequest("GET", "/_uni/scenarios", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response
	var response []*model.Scenario
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	// Verify we have the expected number of scenarios
	if len(response) != len(scenarios) {
		t.Errorf("Expected %d scenarios, got %d", len(scenarios), len(response))
	}
}

func TestScenarioHandler_Update(t *testing.T) {
	// Create a new storage for testing
	storage := storage.NewScenarioStorage()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := NewScenarioHandler(storage, logger)

	// Create a new scenario
	scenario := model.Scenario{
		RequestPath: "GET /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Hello, World!"}`,
	}

	// Create the scenario
	body, _ := json.Marshal(scenario)
	req, _ := http.NewRequest("POST", "/_uni/scenarios", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Get the created scenario's UUID
	var createdScenario model.Scenario
	json.Unmarshal(rr.Body.Bytes(), &createdScenario)
	uuid := createdScenario.UUID

	// Update the scenario
	updatedScenario := model.Scenario{
		UUID:        uuid,
		RequestPath: "PUT /api/test-updated",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"message":"Updated message"}`,
	}

	// Convert updated scenario to JSON
	body, err := json.Marshal(updatedScenario)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to update
	req, err = http.NewRequest("PUT", "/_uni/scenarios/"+uuid, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response
	var response model.Scenario
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	// Verify the updated fields
	if response.RequestPath != updatedScenario.RequestPath {
		t.Errorf("Expected requestPath %s, got %s", updatedScenario.RequestPath, response.RequestPath)
	}
	if response.StatusCode != updatedScenario.StatusCode {
		t.Errorf("Expected status code %d, got %d", updatedScenario.StatusCode, response.StatusCode)
	}
	if response.Data != updatedScenario.Data {
		t.Errorf("Expected data %s, got %s", updatedScenario.Data, response.Data)
	}
}

func TestScenarioHandler_Delete(t *testing.T) {
	// Create a new storage for testing
	storage := storage.NewScenarioStorage()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := NewScenarioHandler(storage, logger)

	// Create a new scenario
	scenario := model.Scenario{
		RequestPath: "GET /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Hello, World!"}`,
	}

	// Create the scenario
	body, _ := json.Marshal(scenario)
	req, _ := http.NewRequest("POST", "/_uni/scenarios", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Get the created scenario's UUID
	var createdScenario model.Scenario
	json.Unmarshal(rr.Body.Bytes(), &createdScenario)
	uuid := createdScenario.UUID

	// Delete the scenario
	req, err := http.NewRequest("DELETE", "/_uni/scenarios/"+uuid, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	// Try to get the deleted scenario
	req, err = http.NewRequest("GET", "/_uni/scenarios/"+uuid, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the scenario is gone (should get 404)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestScenarioHandler_NotFound(t *testing.T) {
	// Create a new storage for testing
	storage := storage.NewScenarioStorage()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := NewScenarioHandler(storage, logger)

	// Create a request with a non-existent UUID
	req, err := http.NewRequest("GET", "/_uni/scenarios/non-existent-uuid", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestScenarioHandler_InvalidJSON(t *testing.T) {
	// Create a new storage for testing
	storage := storage.NewScenarioStorage()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := NewScenarioHandler(storage, logger)

	// Create a request with invalid JSON
	req, err := http.NewRequest("POST", "/_uni/scenarios", bytes.NewBufferString(`{"invalid json`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetScenarioByPath(t *testing.T) {
	// Create a mock logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create test scenarios
	scenarios := []*model.Scenario{
		{
			UUID:        "123",
			RequestPath: "GET /api/users",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `[{"id": 1, "name": "John"}]`,
		},
		{
			UUID:        "456",
			RequestPath: "POST /api/products",
			StatusCode:  201,
			ContentType: "application/json",
			Data:        `[{"id": 1, "name": "Product A"}]`,
		},
		{
			UUID:        "789",
			RequestPath: "GET /api/products",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `[{"id": 1, "name": "Product B"}]`,
		},
	}

	// Create storage with test scenarios
	store := storage.NewScenarioStorage()
	for _, s := range scenarios {
		_ = store.Create(s.UUID, s)
	}

	// Create scenario handler
	handler := NewScenarioHandler(store, logger)

	// Test cases
	tests := []struct {
		name     string
		path     string
		method   string
		wantUUID string
		wantNil  bool
	}{
		{
			name:     "existing path with GET method",
			path:     "/api/users",
			method:   "GET",
			wantUUID: "123",
			wantNil:  false,
		},
		{
			name:     "existing path with POST method",
			path:     "/api/products",
			method:   "POST",
			wantUUID: "456",
			wantNil:  false,
		},
		{
			name:     "existing path with GET method 2",
			path:     "/api/products",
			method:   "GET",
			wantUUID: "789",
			wantNil:  false,
		},
		{
			name:     "existing path with wrong method",
			path:     "/api/users",
			method:   "POST",
			wantUUID: "",
			wantNil:  true,
		},
		{
			name:     "non-existing path",
			path:     "/api/nonexistent",
			method:   "GET",
			wantUUID: "",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario := handler.GetScenarioByPath(tt.path, tt.method)

			if tt.wantNil {
				if scenario != nil {
					t.Errorf("GetScenarioByPath(%q, %q) = %v, want nil", tt.path, tt.method, scenario)
				}
			} else {
				if scenario == nil {
					t.Errorf("GetScenarioByPath(%q, %q) = nil, want scenario with UUID %q", tt.path, tt.method, tt.wantUUID)
				} else if scenario.UUID != tt.wantUUID {
					t.Errorf("GetScenarioByPath(%q, %q) = scenario with UUID %q, want scenario with UUID %q", tt.path, tt.method, scenario.UUID, tt.wantUUID)
				}
			}
		})
	}
}
