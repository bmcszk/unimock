package storage_test

import (
	"testing"

	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
)

// Helper function to create test scenario
func createTestScenario() model.Scenario {
	return model.Scenario{
		UUID:        "test-id",
		RequestPath: "GET /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Test scenario"}`,
	}
}

// Helper function to create updated test scenario
func createUpdatedTestScenario() model.Scenario {
	return model.Scenario{
		UUID:        "test-id",
		RequestPath: "PUT /api/updated",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"message":"Updated scenario"}`,
	}
}

// Helper function to validate scenario fields
func validateScenarioFields(t *testing.T, retrieved, expected model.Scenario) {
	t.Helper()
	if retrieved.UUID != expected.UUID {
		t.Errorf("Expected UUID %s, got %s", expected.UUID, retrieved.UUID)
	}
	if retrieved.RequestPath != expected.RequestPath {
		t.Errorf("Expected RequestPath %s, got %s", expected.RequestPath, retrieved.RequestPath)
	}
	if retrieved.StatusCode != expected.StatusCode {
		t.Errorf("Expected StatusCode %d, got %d", expected.StatusCode, retrieved.StatusCode)
	}
	if retrieved.Data != expected.Data {
		t.Errorf("Expected Data %s, got %s", expected.Data, retrieved.Data)
	}
}

// Helper function to test create and get operations
func testCreateAndGet(t *testing.T, storageInstance storage.ScenarioStorage, scenario model.Scenario) {
	t.Helper()
	// Create
	err := storageInstance.Create("test-id", scenario)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	// Get
	retrieved, err := storageInstance.Get("test-id")
	if err != nil {
		t.Fatalf("Failed to get scenario: %v", err)
	}

	validateScenarioFields(t, retrieved, scenario)
}

// Helper function to test update operations
func testUpdate(t *testing.T, storageInstance storage.ScenarioStorage, updatedScenario model.Scenario) {
	t.Helper()
	// Update
	err := storageInstance.Update("test-id", updatedScenario)
	if err != nil {
		t.Fatalf("Failed to update scenario: %v", err)
	}

	// Get updated
	retrieved, err := storageInstance.Get("test-id")
	if err != nil {
		t.Fatalf("Failed to get updated scenario: %v", err)
	}

	validateScenarioFields(t, retrieved, updatedScenario)
}

// Helper function to test list and delete operations
func testListAndDelete(t *testing.T, storageInstance storage.ScenarioStorage) {
	t.Helper()
	// List
	scenarios := storageInstance.List()
	if len(scenarios) != 1 {
		t.Errorf("Expected 1 scenario in list, got %d", len(scenarios))
	}

	// Delete
	err := storageInstance.Delete("test-id")
	if err != nil {
		t.Fatalf("Failed to delete scenario: %v", err)
	}

	// Verify deleted
	_, err = storageInstance.Get("test-id")
	if err == nil {
		t.Error("Expected error when getting deleted scenario, got nil")
	}

	// List should be empty
	scenarios = storageInstance.List()
	if len(scenarios) != 0 {
		t.Errorf("Expected 0 scenarios after deletion, got %d", len(scenarios))
	}
}

func TestScenarioStorage_CRUD(t *testing.T) {
	// Create new storage
	storageInstance := storage.NewScenarioStorage()

	// Test creating a scenario
	scenario := createTestScenario()
	testCreateAndGet(t, storageInstance, scenario)

	// Test update
	updatedScenario := createUpdatedTestScenario()
	testUpdate(t, storageInstance, updatedScenario)

	// Test list and delete
	testListAndDelete(t, storageInstance)
}

func TestScenarioStorage_ErrorCases(t *testing.T) {
	storageInstance := storage.NewScenarioStorage()

	// Get non-existent scenario
	_, err := storageInstance.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent scenario, got nil")
	}

	// Update non-existent scenario
	err = storageInstance.Update("non-existent", model.Scenario{UUID: "non-existent"})
	if err == nil {
		t.Error("Expected error when updating non-existent scenario, got nil")
	}

	// Delete non-existent scenario
	err = storageInstance.Delete("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent scenario, got nil")
	}

	// Create with empty ID
	err = storageInstance.Create("", model.Scenario{})
	if err == nil {
		t.Error("Expected error when creating scenario with empty ID, got nil")
	}
}
