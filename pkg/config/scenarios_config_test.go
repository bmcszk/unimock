package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

func TestLoadScenariosFromYAML_ValidFile(t *testing.T) {
	yamlContent := `
scenarios:
  - uuid: "test-001"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {"id": "123", "name": "Test User"}
`
	tempFile := createTempScenariosFile(t, yamlContent)
	defer os.Remove(tempFile)

	cfg, err := config.LoadScenariosFromYAML(tempFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(cfg.Scenarios) != 1 {
		t.Fatalf("Expected 1 scenario, got %d", len(cfg.Scenarios))
	}

	scenario := cfg.Scenarios[0]
	if scenario.UUID != "test-001" {
		t.Errorf("Expected UUID 'test-001', got '%s'", scenario.UUID)
	}
	if scenario.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", scenario.Method)
	}
}

func TestLoadScenariosFromYAML_EmptyPath(t *testing.T) {
	cfg, err := config.LoadScenariosFromYAML("")
	if err != nil {
		t.Fatalf("Expected no error for empty path, got: %v", err)
	}

	if len(cfg.Scenarios) != 0 {
		t.Fatalf("Expected 0 scenarios, got %d", len(cfg.Scenarios))
	}
}

func TestLoadScenariosFromYAML_NonexistentFile(t *testing.T) {
	_, err := config.LoadScenariosFromYAML("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}

	if !strings.Contains(err.Error(), "scenarios file not found") {
		t.Errorf("Expected 'scenarios file not found' error, got: %v", err)
	}
}

func TestLoadScenariosFromYAML_InvalidYAML(t *testing.T) {
	invalidYAML := `scenarios: [invalid yaml`
	tempFile := createTempScenariosFile(t, invalidYAML)
	defer os.Remove(tempFile)

	_, err := config.LoadScenariosFromYAML(tempFile)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestLoadScenariosFromYAML_ValidationErrors(t *testing.T) {
	invalidScenarioYAML := `
scenarios:
  - method: "INVALID"
    path: "/api/users"
`
	tempFile := createTempScenariosFile(t, invalidScenarioYAML)
	defer os.Remove(tempFile)

	_, err := config.LoadScenariosFromYAML(tempFile)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid method") {
		t.Errorf("Expected 'invalid method' error, got: %v", err)
	}
}

func TestScenariosConfig_ToModelScenarios(t *testing.T) {
	yamlContent := `
scenarios:
  - uuid: "test-001"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
`
	tempFile := createTempScenariosFile(t, yamlContent)
	defer os.Remove(tempFile)

	cfg, err := config.LoadScenariosFromYAML(tempFile)
	if err != nil {
		t.Fatalf("Failed to load scenarios: %v", err)
	}

	models := cfg.ToModelScenarios()
	if len(models) != 1 {
		t.Fatalf("Expected 1 model scenario, got %d", len(models))
	}

	model := models[0]
	if model.UUID != "test-001" {
		t.Errorf("Expected UUID 'test-001', got '%s'", model.UUID)
	}
	if model.RequestPath != "GET /api/users/123" {
		t.Errorf("Expected RequestPath 'GET /api/users/123', got '%s'", model.RequestPath)
	}
	if model.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d", model.StatusCode)
	}
}

// createTempScenariosFile creates a temporary YAML file for testing
func createTempScenariosFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-scenarios.yaml")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tmpFile
}