package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestScenarioService_GetScenarioByPath(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	service := NewScenarioService(store)

	// Setup test data
	scenarios := []*model.Scenario{
		{
			UUID:        "test-scenario-1",
			RequestPath: "GET /api/users",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"users": []}`,
		},
		{
			UUID:        "test-scenario-2",
			RequestPath: "POST /api/users",
			StatusCode:  201,
			ContentType: "application/json",
			Data:        `{"id": "123"}`,
		},
		{
			UUID:        "test-scenario-3",
			RequestPath: "GET /api/users/*",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"id": "123", "name": "test"}`,
		},
		{
			UUID:        "test-scenario-4",
			RequestPath: "PUT /api/users/*",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"id": "123", "name": "updated"}`,
		},
	}

	// Store test scenarios
	for _, scenario := range scenarios {
		err := service.CreateScenario(context.Background(), scenario)
		if err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	tests := []struct {
		name           string
		path           string
		method         string
		expectedUUID   string
		expectedStatus int
		expectedData   string
	}{
		{
			name:           "GET request",
			path:           "/api/users",
			method:         "GET",
			expectedUUID:   "test-scenario-1",
			expectedStatus: 200,
			expectedData:   `{"users": []}`,
		},
		{
			name:           "POST request",
			path:           "/api/users",
			method:         "POST",
			expectedUUID:   "test-scenario-2",
			expectedStatus: 201,
			expectedData:   `{"id": "123"}`,
		},
		{
			name:           "GET request with wildcard",
			path:           "/api/users/123",
			method:         "GET",
			expectedUUID:   "test-scenario-3",
			expectedStatus: 200,
			expectedData:   `{"id": "123", "name": "test"}`,
		},
		{
			name:           "PUT request with wildcard",
			path:           "/api/users/123",
			method:         "PUT",
			expectedUUID:   "test-scenario-4",
			expectedStatus: 200,
			expectedData:   `{"id": "123", "name": "updated"}`,
		},
		{
			name:   "non-existent path",
			path:   "/api/other",
			method: "GET",
		},
		{
			name:   "non-existent method",
			path:   "/api/users",
			method: "DELETE",
		},
		{
			name:   "case-sensitive method",
			path:   "/api/users",
			method: "get", // Should not match "GET"
		},
		{
			name:   "case-sensitive path",
			path:   "/API/users",
			method: "GET", // Should not match "/api/users"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario := service.GetScenarioByPath(context.Background(), tt.path, tt.method)

			if tt.expectedUUID == "" {
				if scenario != nil {
					t.Errorf("expected nil scenario, got %+v", scenario)
				}
				return
			}

			if scenario == nil {
				t.Error("expected scenario, got nil")
				return
			}

			if scenario.UUID != tt.expectedUUID {
				t.Errorf("UUID = %q, want %q", scenario.UUID, tt.expectedUUID)
			}

			if scenario.StatusCode != tt.expectedStatus {
				t.Errorf("StatusCode = %d, want %d", scenario.StatusCode, tt.expectedStatus)
			}

			if scenario.Data != tt.expectedData {
				t.Errorf("Data = %q, want %q", scenario.Data, tt.expectedData)
			}
		})
	}
}

func TestScenarioService_ListScenarios(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	service := NewScenarioService(store)

	// Setup test data
	scenarios := []*model.Scenario{
		{
			UUID:        "test-scenario-1",
			RequestPath: "GET /api/users",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"users": []}`,
		},
		{
			UUID:        "test-scenario-2",
			RequestPath: "POST /api/users",
			StatusCode:  201,
			ContentType: "application/json",
			Data:        `{"id": "123"}`,
		},
		{
			UUID:        "test-scenario-3",
			RequestPath: "GET /api/orders",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"orders": []}`,
		},
	}

	// Store test scenarios
	for _, scenario := range scenarios {
		err := service.CreateScenario(context.Background(), scenario)
		if err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	// Get all scenarios
	allScenarios := service.ListScenarios(context.Background())

	// Check count
	if len(allScenarios) != len(scenarios) {
		t.Errorf("got %d scenarios, want %d", len(allScenarios), len(scenarios))
	}

	// Check each scenario
	for i, scenario := range scenarios {
		found := false
		for _, s := range allScenarios {
			if s.UUID == scenario.UUID {
				found = true
				if s.RequestPath != scenario.RequestPath {
					t.Errorf("scenario[%d].RequestPath = %q, want %q", i, s.RequestPath, scenario.RequestPath)
				}
				if s.StatusCode != scenario.StatusCode {
					t.Errorf("scenario[%d].StatusCode = %d, want %d", i, s.StatusCode, scenario.StatusCode)
				}
				if s.Data != scenario.Data {
					t.Errorf("scenario[%d].Data = %q, want %q", i, s.Data, scenario.Data)
				}
				break
			}
		}
		if !found {
			t.Errorf("scenario %q not found in list", scenario.UUID)
		}
	}
}

func TestScenarioService_GetScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	service := NewScenarioService(store)

	// Setup test data
	scenario := &model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	}

	err := service.CreateScenario(context.Background(), scenario)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	tests := []struct {
		name           string
		uuid           string
		expectedStatus int
		expectedData   string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "existing scenario",
			uuid:           "test-scenario",
			expectedStatus: 200,
			expectedData:   `{"users": []}`,
		},
		{
			name:          "non-existent scenario",
			uuid:          "non-existent",
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name:          "empty UUID",
			uuid:          "",
			expectError:   true,
			errorContains: "invalid request: scenario ID cannot be empty",
		},
		{
			name:          "invalid UUID format",
			uuid:          "invalid-uuid",
			expectError:   true,
			errorContains: "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := service.GetScenario(context.Background(), tt.uuid)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errorContains)) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if scenario.StatusCode != tt.expectedStatus {
				t.Errorf("StatusCode = %d, want %d", scenario.StatusCode, tt.expectedStatus)
			}

			if scenario.Data != tt.expectedData {
				t.Errorf("Data = %q, want %q", scenario.Data, tt.expectedData)
			}
		})
	}
}

func TestScenarioService_CreateScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	service := NewScenarioService(store)

	tests := []struct {
		name           string
		scenario       *model.Scenario
		expectedStatus int
		expectedData   string
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid scenario",
			scenario: &model.Scenario{
				RequestPath: "GET /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectedStatus: 200,
			expectedData:   `{"users": []}`,
		},
		{
			name: "scenario with UUID",
			scenario: &model.Scenario{
				UUID:        "custom-uuid",
				RequestPath: "GET /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectedStatus: 200,
			expectedData:   `{"users": []}`,
		},
		{
			name: "scenario with location",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				Location:    "/_uni/scenarios/test-scenario",
				RequestPath: "GET /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "resource already exists",
		},
		{
			name: "duplicate UUID",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				RequestPath: "GET /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "resource already exists",
		},
		{
			name: "invalid request path",
			scenario: &model.Scenario{
				RequestPath: "INVALID /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "invalid HTTP method in request path: INVALID",
		},
		{
			name: "invalid status code",
			scenario: &model.Scenario{
				RequestPath: "GET /api/users",
				StatusCode:  999,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "invalid status code: 999",
		},
		{
			name: "invalid content type (now valid text/plain)",
			scenario: &model.Scenario{
				RequestPath: "GET /api/users",
				StatusCode:  200,
				ContentType: "text/plain",
				Data:        `{"users": []}`,
			},
			expectError:    false,
			errorContains:  "",
			expectedStatus: 200,
			expectedData:   `{"users": []}`,
		},
		{
			name: "content type with spaces",
			scenario: &model.Scenario{
				UUID:        "create-test-content-type-with-spaces",
				RequestPath: "GET /path",
				StatusCode:  200,
				ContentType: "application/ json_with_spaces",
				Data:        "test data",
			},
			expectError:   true,
			errorContains: "invalid content type: contains whitespace characters",
		},
		{
			name: "create with valid empty content type",
			scenario: &model.Scenario{
				UUID:        "create-test-empty-content-type",
				RequestPath: "GET /path",
				StatusCode:  200,
				ContentType: "",
				Data:        "test data",
			},
			expectError:    false,
			errorContains:  "",
			expectedStatus: 200,
			expectedData:   "test data",
		},
	}

	// Create first scenario for duplicate UUID test
	err := service.CreateScenario(context.Background(), &model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	})
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateScenario(context.Background(), tt.scenario)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errorContains)) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify scenario was created
			scenario, err := service.GetScenario(context.Background(), tt.scenario.UUID)
			if err != nil {
				t.Errorf("failed to get created scenario: %v", err)
				return
			}

			if scenario.StatusCode != tt.expectedStatus {
				t.Errorf("StatusCode = %d, want %d", scenario.StatusCode, tt.expectedStatus)
			}

			if scenario.Data != tt.expectedData {
				t.Errorf("Data = %q, want %q", scenario.Data, tt.expectedData)
			}

			// Verify location was set correctly
			if tt.scenario.Location == "" {
				expectedLocation := "/_uni/scenarios/" + tt.scenario.UUID
				if scenario.Location != expectedLocation {
					t.Errorf("Location = %q, want %q", scenario.Location, expectedLocation)
				}
			} else {
				if scenario.Location != tt.scenario.Location {
					t.Errorf("Location = %q, want %q", scenario.Location, tt.scenario.Location)
				}
			}
		})
	}
}

func TestScenarioService_UpdateScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	service := NewScenarioService(store)

	// Setup test data
	scenario := &model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	}

	err := service.CreateScenario(context.Background(), scenario)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	tests := []struct {
		name           string
		scenario       *model.Scenario
		expectedStatus int
		expectedData   string
		expectError    bool
		errorContains  string
	}{
		{
			name: "update existing scenario",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				RequestPath: "GET /api/users",
				StatusCode:  201,
				ContentType: "application/json",
				Data:        `{"users": [{"id": "123"}]}`,
			},
			expectedStatus: 201,
			expectedData:   `{"users": [{"id": "123"}]}`,
		},
		{
			name: "update with different UUID",
			scenario: &model.Scenario{
				UUID:        "different-uuid",
				RequestPath: "GET /api/users",
				StatusCode:  201,
				ContentType: "application/json",
				Data:        `{"users": [{"id": "123"}]}`,
			},
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name: "non-existent scenario",
			scenario: &model.Scenario{
				UUID:        "non-existent",
				RequestPath: "GET /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name: "invalid request path",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				RequestPath: "INVALID /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "invalid HTTP method in request path: INVALID",
		},
		{
			name: "invalid status code",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				RequestPath: "GET /api/users",
				StatusCode:  999,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "invalid status code: 999",
		},
		{
			name: "invalid content type (now valid text/plain)",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				RequestPath: "GET /api/users",
				StatusCode:  200,
				ContentType: "text/plain",
				Data:        `{"users": []}`,
			},
			expectError:    false,
			errorContains:  "",
			expectedStatus: 200,
			expectedData:   `{"users": []}`,
		},
		{
			name: "content type with spaces",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				RequestPath: "GET /path",
				StatusCode:  200,
				ContentType: "application/ json_with_spaces",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "invalid content type: contains whitespace characters",
		},
		{
			name: "update with valid empty content type",
			scenario: &model.Scenario{
				UUID:        "test-scenario",
				RequestPath: "GET /path",
				StatusCode:  200,
				ContentType: "",
				Data:        `{"users": []}`,
			},
			expectError:    false,
			errorContains:  "",
			expectedStatus: 200,
			expectedData:   `{"users": []}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateScenario(context.Background(), tt.scenario.UUID, tt.scenario)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errorContains)) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify scenario was updated
			scenario, err := service.GetScenario(context.Background(), tt.scenario.UUID)
			if err != nil {
				t.Errorf("failed to get updated scenario: %v", err)
				return
			}

			if scenario.StatusCode != tt.expectedStatus {
				t.Errorf("StatusCode = %d, want %d", scenario.StatusCode, tt.expectedStatus)
			}

			if scenario.Data != tt.expectedData {
				t.Errorf("Data = %q, want %q", scenario.Data, tt.expectedData)
			}
		})
	}
}

func TestScenarioService_DeleteScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	service := NewScenarioService(store)

	// Setup test data
	scenario := &model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	}

	err := service.CreateScenario(context.Background(), scenario)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	tests := []struct {
		name          string
		uuid          string
		expectError   bool
		errorContains string
	}{
		{
			name: "delete existing scenario",
			uuid: "test-scenario",
		},
		{
			name:          "delete non-existent scenario",
			uuid:          "non-existent",
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name:          "empty UUID",
			uuid:          "",
			expectError:   true,
			errorContains: "invalid request: scenario ID cannot be empty",
		},
		{
			name:          "invalid UUID format",
			uuid:          "invalid-uuid",
			expectError:   true,
			errorContains: "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteScenario(context.Background(), tt.uuid)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errorContains)) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify scenario was deleted
			if _, err = service.GetScenario(context.Background(), tt.uuid); err == nil {
				t.Error("scenario still exists after deletion")
			} else if !bytes.Contains([]byte(err.Error()), []byte("resource not found")) {
				t.Errorf("unexpected error when getting deleted scenario: %v", err)
			}
		})
	}
}
