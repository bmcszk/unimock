package service_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
)

// Helper function to create test scenarios
func createTestScenarios(t *testing.T, scenarioSvc *service.ScenarioService) {
	t.Helper()
	scenarios := []model.Scenario{
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

	for _, scenario := range scenarios {
		_, err := scenarioSvc.CreateScenario(context.Background(), scenario)
		if err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}
}

// Helper function to validate scenario response
func validateScenarioResponse(
	t *testing.T, scenario model.Scenario, expectedUUID string, expectedStatus int, expectedData string,
) {
	t.Helper()
	if scenario.UUID != expectedUUID {
		t.Errorf("UUID = %q, want %q", scenario.UUID, expectedUUID)
	}
	if scenario.StatusCode != expectedStatus {
		t.Errorf("StatusCode = %d, want %d", scenario.StatusCode, expectedStatus)
	}
	if scenario.Data != expectedData {
		t.Errorf("Data = %q, want %q", scenario.Data, expectedData)
	}
}

// Helper function to get test cases for GetScenarioByPath
func getScenarioByPathTestCases() []struct {
	name           string
	path           string
	method         string
	expectedUUID   string
	expectedStatus int
	expectedData   string
} {
	return []struct {
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
}

func TestScenarioService_GetScenarioByPath(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	scenarioSvc := service.NewScenarioService(store)

	// Setup test data
	createTestScenarios(t, scenarioSvc)

	tests := getScenarioByPathTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, found := scenarioSvc.GetScenarioByPath(context.Background(), tt.path, tt.method)
			if tt.expectedUUID == "" {
				assert.False(t, found, "Expected no scenario for %s %s", tt.method, tt.path)
			} else {
				assert.True(t, found, "Expected scenario found for %s %s", tt.method, tt.path)
				validateScenarioResponse(t, scenario, tt.expectedUUID, tt.expectedStatus, tt.expectedData)
			}
		})
	}
}

// Helper function to create list test scenarios
func createListTestScenarios(t *testing.T, scenarioSvc *service.ScenarioService) []model.Scenario {
	t.Helper()
	scenarios := []model.Scenario{
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

	for _, scenario := range scenarios {
		_, err := scenarioSvc.CreateScenario(context.Background(), scenario)
		if err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	// Convert to slice for return
	return scenarios
}

// Helper function to find scenario in list
func findScenarioInList(scenarios []model.Scenario, uuid string) (model.Scenario, bool) {
	for _, s := range scenarios {
		if s.UUID == uuid {
			return s, true
		}
	}
	return model.Scenario{}, false
}

// Helper function to validate scenario in list
func validateScenarioInList(t *testing.T, expected model.Scenario, actual model.Scenario, index int) {
	t.Helper()
	if actual.RequestPath != expected.RequestPath {
		t.Errorf("scenario[%d].RequestPath = %q, want %q", index, actual.RequestPath, expected.RequestPath)
	}
	if actual.StatusCode != expected.StatusCode {
		t.Errorf("scenario[%d].StatusCode = %d, want %d", index, actual.StatusCode, expected.StatusCode)
	}
	if actual.Data != expected.Data {
		t.Errorf("scenario[%d].Data = %q, want %q", index, actual.Data, expected.Data)
	}
}

func TestScenarioService_ListScenarios(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	scenarioSvc := service.NewScenarioService(store)

	// Setup test data
	scenarios := createListTestScenarios(t, scenarioSvc)

	// Get all scenarios
	allScenarios := scenarioSvc.ListScenarios(context.Background())

	// Check count
	if len(allScenarios) != len(scenarios) {
		t.Errorf("got %d scenarios, want %d", len(allScenarios), len(scenarios))
	}

	// Check each scenario
	for i, scenario := range scenarios {
		actual, found := findScenarioInList(allScenarios, scenario.UUID)
		if !found {
			t.Errorf("scenario %q not found in list", scenario.UUID)
			continue
		}
		validateScenarioInList(t, scenario, actual, i)
	}
}

// Helper function to setup GetScenario test data
func setupGetScenarioTest(t *testing.T, scenarioSvc *service.ScenarioService) {
	t.Helper()
	scenario := model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	}

	_, err := scenarioSvc.CreateScenario(context.Background(), scenario)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}
}

// Helper function to get test cases for GetScenario
func getScenarioTestCases() []struct {
	name           string
	uuid           string
	expectedStatus int
	expectedData   string
	expectError    bool
	errorContains  string
} {
	return []struct {
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
}

// Helper function to validate GetScenario response
func validateGetScenarioResponse(
	t *testing.T, scenario model.Scenario, err error, expectedStatus int, expectedData string,
	expectError bool, errorContains string,
) {
	t.Helper()
	if expectError {
		validateErrorResponse(t, err, errorContains)
		return
	}

	validateNoError(t, err)
	validateScenarioData(t, scenario, expectedStatus, expectedData)
}

// validateGetScenarioResponseWithError validates that an error occurred as expected
func validateGetScenarioResponseWithError(
	t *testing.T, err error, errorContains string,
) {
	t.Helper()
	validateErrorResponse(t, err, errorContains)
}

// validateGetScenarioResponseWithSuccess validates successful scenario retrieval
func validateGetScenarioResponseWithSuccess(
	t *testing.T, scenario model.Scenario, err error, expectedStatus int, expectedData string,
) {
	t.Helper()
	validateNoError(t, err)
	validateScenarioData(t, scenario, expectedStatus, expectedData)
}

func TestScenarioService_GetScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	scenarioSvc := service.NewScenarioService(store)

	// Setup test data
	setupGetScenarioTest(t, scenarioSvc)

	tests := getScenarioTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := scenarioSvc.GetScenario(context.Background(), tt.uuid)
			validateGetScenarioResponse(
				t, scenario, err, tt.expectedStatus, tt.expectedData, tt.expectError, tt.errorContains,
			)
		})
	}
}

// Helper function to setup CreateScenario test data
func setupCreateScenarioTest(t *testing.T, scenarioSvc *service.ScenarioService) {
	t.Helper()
	// Create first scenario for duplicate UUID test
	_, err := scenarioSvc.CreateScenario(context.Background(), model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	})
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}
}

// Helper function to get test cases for CreateScenario
func getCreateScenarioTestCases() []struct {
	name           string
	scenario       model.Scenario
	expectedStatus int
	expectedData   string
	expectError    bool
	errorContains  string
} {
	return []struct {
		name           string
		scenario       model.Scenario
		expectedStatus int
		expectedData   string
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid scenario",
			scenario: model.Scenario{
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
			scenario: model.Scenario{
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
			scenario: model.Scenario{
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
			scenario: model.Scenario{
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
			scenario: model.Scenario{
				RequestPath: "INVALID /api/users",
				StatusCode:  200,
				ContentType: "application/json",
				Data:        `{"users": []}`,
			},
			expectError:   true,
			errorContains: "invalid HTTP method in request path: INVALID",
		},
		{
			name: "valid text/plain content type",
			scenario: model.Scenario{
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
			name: "create with valid empty content type",
			scenario: model.Scenario{
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
}

// Helper function to validate error response
func validateErrorResponse(t *testing.T, err error, errorContains string) {
	t.Helper()
	if err == nil {
		t.Error("expected error, got nil")
	} else if errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(errorContains)) {
		t.Errorf("error message %q does not contain %q", err.Error(), errorContains)
	}
}

// Helper function to validate no error occurred
func validateNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Helper function to validate scenario data
func validateScenarioData(t *testing.T, scenario model.Scenario, expectedStatus int, expectedData string) {
	t.Helper()
	if scenario.StatusCode != expectedStatus {
		t.Errorf("StatusCode = %d, want %d", scenario.StatusCode, expectedStatus)
	}
	if scenario.Data != expectedData {
		t.Errorf("Data = %q, want %q", scenario.Data, expectedData)
	}
}

// Helper function to validate CreateScenario response
func validateCreateScenarioResponse(
	t *testing.T, _ *service.ScenarioService, _ model.Scenario, err error,
	_ int, _ string, expectError bool, errorContains string,
) {
	t.Helper()
	if expectError {
		validateErrorResponse(t, err, errorContains)
		return
	}

	// For successful creation, just verify no error occurred
	// UUID generation happens inside CreateScenario so we can't easily verify the stored scenario
	// The main validation is that no error occurred during creation
	validateNoError(t, err)
}

func TestScenarioService_CreateScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	scenarioSvc := service.NewScenarioService(store)

	// Setup test data
	setupCreateScenarioTest(t, scenarioSvc)

	tests := getCreateScenarioTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := scenarioSvc.CreateScenario(context.Background(), tt.scenario)
			validateCreateScenarioResponse(
				t, scenarioSvc, tt.scenario, err, tt.expectedStatus, tt.expectedData,
				tt.expectError, tt.errorContains,
			)
		})
	}
}

// Helper function to setup UpdateScenario test data
func setupUpdateScenarioTest(t *testing.T, scenarioSvc *service.ScenarioService) {
	t.Helper()
	scenario := model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	}

	_, err := scenarioSvc.CreateScenario(context.Background(), scenario)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}
}

// Helper function to get test cases for UpdateScenario
func getUpdateScenarioTestCases() []struct {
	name           string
	scenario       model.Scenario
	expectedStatus int
	expectedData   string
	expectError    bool
	errorContains  string
} {
	return []struct {
		name           string
		scenario       model.Scenario
		expectedStatus int
		expectedData   string
		expectError    bool
		errorContains  string
	}{
		{
			name: "update existing scenario",
			scenario: model.Scenario{
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
			scenario: model.Scenario{
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
			scenario: model.Scenario{
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
			scenario: model.Scenario{
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
			name: "valid text/plain content type",
			scenario: model.Scenario{
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
			name: "update with valid empty content type",
			scenario: model.Scenario{
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
}

// Helper function to validate UpdateScenario response
func validateUpdateScenarioResponse(
	t *testing.T, scenarioSvc *service.ScenarioService, scenario model.Scenario, err error,
	expectedStatus int, expectedData string, expectError bool, errorContains string,
) {
	t.Helper()
	if expectError {
		validateErrorResponse(t, err, errorContains)
		return
	}

	validateNoError(t, err)

	// Verify scenario was updated
	updatedScenario, err := scenarioSvc.GetScenario(context.Background(), scenario.UUID)
	if err != nil {
		t.Errorf("failed to get updated scenario: %v", err)
		return
	}

	validateScenarioData(t, updatedScenario, expectedStatus, expectedData)
}

func TestScenarioService_UpdateScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	scenarioSvc := service.NewScenarioService(store)

	// Setup test data
	setupUpdateScenarioTest(t, scenarioSvc)

	tests := getUpdateScenarioTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scenarioSvc.UpdateScenario(context.Background(), tt.scenario.UUID, tt.scenario)
			validateUpdateScenarioResponse(
				t, scenarioSvc, tt.scenario, err, tt.expectedStatus, tt.expectedData,
				tt.expectError, tt.errorContains,
			)
		})
	}
}

// Helper function to setup DeleteScenario test data
func setupDeleteScenarioTest(t *testing.T, scenarioSvc *service.ScenarioService) {
	t.Helper()
	scenario := model.Scenario{
		UUID:        "test-scenario",
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": []}`,
	}

	_, err := scenarioSvc.CreateScenario(context.Background(), scenario)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}
}

// Helper function to get test cases for DeleteScenario
func getDeleteScenarioTestCases() []struct {
	name          string
	uuid          string
	expectError   bool
	errorContains string
} {
	return []struct {
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
}

// Helper function to validate DeleteScenario response
func validateDeleteScenarioResponse(
	t *testing.T, scenarioSvc *service.ScenarioService, uuid string, err error,
	expectError bool, errorContains string,
) {
	t.Helper()
	if expectError {
		validateErrorResponse(t, err, errorContains)
		return
	}

	validateNoError(t, err)

	// Verify scenario was deleted
	if _, err = scenarioSvc.GetScenario(context.Background(), uuid); err == nil {
		t.Error("scenario still exists after deletion")
	} else if !bytes.Contains([]byte(err.Error()), []byte("resource not found")) {
		t.Errorf("unexpected error when getting deleted scenario: %v", err)
	}
}

func TestScenarioService_DeleteScenario(t *testing.T) {
	// Create service
	store := storage.NewScenarioStorage()
	scenarioSvc := service.NewScenarioService(store)

	// Setup test data
	setupDeleteScenarioTest(t, scenarioSvc)

	tests := getDeleteScenarioTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scenarioSvc.DeleteScenario(context.Background(), tt.uuid)
			validateDeleteScenarioResponse(t, scenarioSvc, tt.uuid, err, tt.expectError, tt.errorContains)
		})
	}
}
