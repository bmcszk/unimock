package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenarioHandler_Create(t *testing.T) {
	// Create a new storage and service for testing
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

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
	scenarioHandler.ServeHTTP(rr, req)

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
	// Create a new storage and service for testing
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

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
	scenarioHandler.ServeHTTP(rr, req)

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
	scenarioHandler.ServeHTTP(rr, req)

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
	// Create a new storage and service for testing
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

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
		scenarioHandler.ServeHTTP(rr, req)
	}

	// Now test the list endpoint
	req, err := http.NewRequest("GET", "/_uni/scenarios", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	scenarioHandler.ServeHTTP(rr, req)

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
	// Create a new storage and service for testing
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

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
	scenarioHandler.ServeHTTP(rr, req)

	// Get the created scenario's UUID
	var createdScenario model.Scenario
	json.Unmarshal(rr.Body.Bytes(), &createdScenario)
	uuid := createdScenario.UUID

	// Update the scenario
	updatedScenario := createdScenario
	updatedScenario.StatusCode = 201
	updatedScenario.Data = `{"message":"Updated World!"}`

	body, _ = json.Marshal(updatedScenario)
	req, _ = http.NewRequest("PUT", "/_uni/scenarios/"+uuid, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	scenarioHandler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify the scenario was updated
	req, _ = http.NewRequest("GET", "/_uni/scenarios/"+uuid, nil)
	rr = httptest.NewRecorder()
	scenarioHandler.ServeHTTP(rr, req)

	var updated model.Scenario
	json.Unmarshal(rr.Body.Bytes(), &updated)

	if updated.StatusCode != 201 {
		t.Errorf("Expected StatusCode 201, got %d", updated.StatusCode)
	}
	if updated.Data != `{"message":"Updated World!"}` {
		t.Errorf(`Expected Data {"message":"Updated World!"}, got %s`, updated.Data)
	}
}

func TestScenarioHandler_Delete(t *testing.T) {
	// Create a new storage and service for testing
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

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
	scenarioHandler.ServeHTTP(rr, req)

	// Get the created scenario's UUID
	var createdScenario model.Scenario
	json.Unmarshal(rr.Body.Bytes(), &createdScenario)
	uuid := createdScenario.UUID

	// Now delete the scenario
	req, _ = http.NewRequest("DELETE", "/_uni/scenarios/"+uuid, nil)
	rr = httptest.NewRecorder()
	scenarioHandler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	// Try to get the deleted scenario - should return 404
	req, _ = http.NewRequest("GET", "/_uni/scenarios/"+uuid, nil)
	rr = httptest.NewRecorder()
	scenarioHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestScenarioHandler_NotFound(t *testing.T) {
	// Create a new storage and service for testing
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

	// Request a non-existent scenario
	req, _ := http.NewRequest("GET", "/_uni/scenarios/non-existent", nil)
	rr := httptest.NewRecorder()
	scenarioHandler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestScenarioHandler_InvalidJSON(t *testing.T) {
	// Create a new storage and service for testing
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

	// Create a request with invalid JSON
	req, _ := http.NewRequest("POST", "/_uni/scenarios", bytes.NewBufferString(`{"invalid JSON`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	scenarioHandler.ServeHTTP(rr, req)

	// Check the status code - should be BadRequest for invalid JSON
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetScenarioByPath(t *testing.T) {
	store := storage.NewScenarioStorage()
	scenarioService := service.NewScenarioService(store)

	// Create some scenarios
	testScenarios := []model.Scenario{
		{UUID: "uuid1", RequestPath: "GET /api/test", Data: "test data 1", StatusCode: http.StatusOK},
		{UUID: "uuid2", RequestPath: "GET /api/users/*", Data: "user data", StatusCode: http.StatusOK},
		{UUID: "uuid3", RequestPath: "POST /api/test", Data: "post data", StatusCode: http.StatusCreated},
	}

	for i := range testScenarios {
		_, err := scenarioService.CreateScenario(context.TODO(), testScenarios[i])
		require.NoError(t, err, "Failed to create scenario %s in TestGetScenarioByPath", testScenarios[i].UUID)
	}

	// Test cases
	tests := []struct {
		name       string
		path       string
		method     string
		expectUUID string
		expectNil  bool
	}{
		{"Exact match", "/api/test", "GET", "uuid1", false},
		{"Path with parameter", "/api/users/123", "GET", "uuid2", false},
		{"POST request", "/api/test", "POST", "uuid3", false},
		{"Non-existent path", "/api/nonexistent", "GET", "", true},
		{"Wrong method", "/api/test", "DELETE", "", true},
	}

	for _, tt := range tests {
		tc := tt // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			scenario, found := scenarioService.GetScenarioByPath(context.TODO(), tc.path, tc.method)
			if tc.expectNil {
				assert.False(t, found)
			} else {
				assert.True(t, found)
				assert.Equal(t, tc.expectUUID, scenario.UUID)
			}
		})
	}
}
