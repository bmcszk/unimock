package storage

import (
	"testing"

	"github.com/bmcszk/unimock/internal/model"
)

func TestScenarioStorage_CRUD(t *testing.T) {
	// Create new storage
	storage := NewScenarioStorage()

	// Test creating a scenario
	scenario := &model.Scenario{
		UUID:        "test-id",
		RequestPath: "GET /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Test scenario"}`,
	}

	// Create
	err := storage.Create("test-id", scenario)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	// Get
	retrieved, err := storage.Get("test-id")
	if err != nil {
		t.Fatalf("Failed to get scenario: %v", err)
	}

	// Verify fields match
	if retrieved.UUID != scenario.UUID {
		t.Errorf("Expected UUID %s, got %s", scenario.UUID, retrieved.UUID)
	}
	if retrieved.RequestPath != scenario.RequestPath {
		t.Errorf("Expected RequestPath %s, got %s", scenario.RequestPath, retrieved.RequestPath)
	}
	if retrieved.StatusCode != scenario.StatusCode {
		t.Errorf("Expected StatusCode %d, got %d", scenario.StatusCode, retrieved.StatusCode)
	}
	if retrieved.Data != scenario.Data {
		t.Errorf("Expected Data %s, got %s", scenario.Data, retrieved.Data)
	}

	// Update
	updatedScenario := &model.Scenario{
		UUID:        "test-id",
		RequestPath: "PUT /api/updated",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"message":"Updated scenario"}`,
	}

	err = storage.Update("test-id", updatedScenario)
	if err != nil {
		t.Fatalf("Failed to update scenario: %v", err)
	}

	// Get updated
	retrieved, err = storage.Get("test-id")
	if err != nil {
		t.Fatalf("Failed to get updated scenario: %v", err)
	}

	// Verify updated fields
	if retrieved.RequestPath != updatedScenario.RequestPath {
		t.Errorf("Expected updated RequestPath %s, got %s", updatedScenario.RequestPath, retrieved.RequestPath)
	}
	if retrieved.StatusCode != updatedScenario.StatusCode {
		t.Errorf("Expected updated StatusCode %d, got %d", updatedScenario.StatusCode, retrieved.StatusCode)
	}
	if retrieved.Data != updatedScenario.Data {
		t.Errorf("Expected updated Data %s, got %s", updatedScenario.Data, retrieved.Data)
	}

	// List
	scenarios := storage.List()
	if len(scenarios) != 1 {
		t.Errorf("Expected 1 scenario in list, got %d", len(scenarios))
	}

	// Delete
	err = storage.Delete("test-id")
	if err != nil {
		t.Fatalf("Failed to delete scenario: %v", err)
	}

	// Verify deleted
	_, err = storage.Get("test-id")
	if err == nil {
		t.Error("Expected error when getting deleted scenario, got nil")
	}

	// List should be empty
	scenarios = storage.List()
	if len(scenarios) != 0 {
		t.Errorf("Expected 0 scenarios after deletion, got %d", len(scenarios))
	}
}

func TestScenarioStorage_ErrorCases(t *testing.T) {
	storage := NewScenarioStorage()

	// Get non-existent scenario
	_, err := storage.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent scenario, got nil")
	}

	// Update non-existent scenario
	err = storage.Update("non-existent", &model.Scenario{UUID: "non-existent"})
	if err == nil {
		t.Error("Expected error when updating non-existent scenario, got nil")
	}

	// Delete non-existent scenario
	err = storage.Delete("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent scenario, got nil")
	}

	// Create with nil scenario
	err = storage.Create("test-id", nil)
	if err == nil {
		t.Error("Expected error when creating nil scenario, got nil")
	}
}
