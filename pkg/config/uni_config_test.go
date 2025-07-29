package config_test

import (
	"os"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

func TestUniConfig_ReturnBodyDefault(t *testing.T) {
	// Test that ReturnBody defaults to false
	mockConfig := config.NewUniConfig()

	// Add a test section
	mockConfig.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
	}

	section := mockConfig.Sections["test"]
	if section.ReturnBody {
		t.Errorf("Expected ReturnBody to default to false, got %v", section.ReturnBody)
	}
}

func TestUniConfig_LoadFromYAML_ReturnBodyFalse(t *testing.T) {
	// Create a temporary YAML file with return_body: false
	yamlContent := `test_section:
  path_pattern: "/api/test/*"
  body_id_paths:
    - "/id"
  return_body: false
`

	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Load configuration
	mockConfig, err := config.LoadFromYAML(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	section, exists := mockConfig.Sections["test_section"]
	if !exists {
		t.Fatal("Expected test_section to exist")
	}

	if section.ReturnBody {
		t.Errorf("Expected ReturnBody to be false, got %v", section.ReturnBody)
	}
}

func TestUniConfig_LoadFromYAML_ReturnBodyTrue(t *testing.T) {
	// Create a temporary YAML file with return_body: true
	yamlContent := `test_section:
  path_pattern: "/api/test/*"
  body_id_paths:
    - "/id"
  return_body: true
`

	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Load configuration
	mockConfig, err := config.LoadFromYAML(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	section, exists := mockConfig.Sections["test_section"]
	if !exists {
		t.Fatal("Expected test_section to exist")
	}

	if !section.ReturnBody {
		t.Errorf("Expected ReturnBody to be true, got %v", section.ReturnBody)
	}
}

func TestUniConfig_LoadFromYAML_ReturnBodyOmitted(t *testing.T) {
	// Create a temporary YAML file without return_body field (should default to false)
	yamlContent := `test_section:
  path_pattern: "/api/test/*"
  body_id_paths:
    - "/id"
`

	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Load configuration
	mockConfig, err := config.LoadFromYAML(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	section, exists := mockConfig.Sections["test_section"]
	if !exists {
		t.Fatal("Expected test_section to exist")
	}

	// Should default to false when omitted
	if section.ReturnBody {
		t.Errorf("Expected ReturnBody to default to false when omitted, got %v", section.ReturnBody)
	}
}

// TestSection_OnlyPathPatternField tests that Path field is removed and only PathPattern exists
func TestSection_OnlyPathPatternField(t *testing.T) {
	section := config.Section{
		PathPattern: "/api/users/*",
		BodyIDPaths: []string{"/id"},
	}

	// Test that PathPattern is properly set
	if section.PathPattern != "/api/users/*" {
		t.Errorf("Expected PathPattern to be '/api/users/*', got '%s'", section.PathPattern)
	}

	// Test that GetPathPattern returns PathPattern value
	if section.GetPathPattern() != "/api/users/*" {
		t.Errorf("Expected GetPathPattern() to return '/api/users/*', got '%s'", section.GetPathPattern())
	}
}

// TestSection_HeaderIDNames_SingleHeader tests single header ID extraction
func TestSection_HeaderIDNames_SingleHeader(t *testing.T) {
	section := config.Section{
		PathPattern:   "/api/users/*",
		HeaderIDNames: []string{"X-User-ID"},
		BodyIDPaths:   []string{"/id"},
	}

	if len(section.HeaderIDNames) != 1 {
		t.Errorf("Expected 1 header ID name, got %d", len(section.HeaderIDNames))
	}

	if section.HeaderIDNames[0] != "X-User-ID" {
		t.Errorf("Expected header ID name to be 'X-User-ID', got '%s'", section.HeaderIDNames[0])
	}
}

// TestSection_HeaderIDNames_MultipleHeaders tests multiple header ID extraction
func TestSection_HeaderIDNames_MultipleHeaders(t *testing.T) {
	section := config.Section{
		PathPattern:   "/api/users/*",
		HeaderIDNames: []string{"X-User-ID", "X-Resource-ID", "Authorization"},
		BodyIDPaths:   []string{"/id"},
	}

	expectedHeaders := []string{"X-User-ID", "X-Resource-ID", "Authorization"}

	if len(section.HeaderIDNames) != len(expectedHeaders) {
		t.Errorf("Expected %d header ID names, got %d", len(expectedHeaders), len(section.HeaderIDNames))
	}

	for i, expected := range expectedHeaders {
		if section.HeaderIDNames[i] != expected {
			t.Errorf("Expected header ID name %d to be '%s', got '%s'", i, expected, section.HeaderIDNames[i])
		}
	}
}

// TestUniConfig_LoadFromYAML_HeaderIDNames tests loading HeaderIDNames from YAML
func TestUniConfig_LoadFromYAML_HeaderIDNames(t *testing.T) {
	yamlContent := `test_section:
  path_pattern: "/api/test/*"
  body_id_paths:
    - "/id"
  header_id_names:
    - "X-Test-ID"
    - "X-Resource-ID"
`

	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	mockConfig, err := config.LoadFromYAML(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	section, exists := mockConfig.Sections["test_section"]
	if !exists {
		t.Fatal("Expected test_section to exist")
	}

	expectedHeaders := []string{"X-Test-ID", "X-Resource-ID"}
	if len(section.HeaderIDNames) != len(expectedHeaders) {
		t.Errorf("Expected %d header ID names, got %d", len(expectedHeaders), len(section.HeaderIDNames))
	}

	for i, expected := range expectedHeaders {
		if section.HeaderIDNames[i] != expected {
			t.Errorf("Expected header ID name %d to be '%s', got '%s'", i, expected, section.HeaderIDNames[i])
		}
	}
}

// TestUniConfig_LoadFromYAML_OnlyPathPattern tests that only path_pattern is supported, not path
func TestUniConfig_LoadFromYAML_OnlyPathPattern(t *testing.T) {
	yamlContent := `test_section:
  path_pattern: "/api/test/*"
  body_id_paths:
    - "/id"
`

	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	mockConfig, err := config.LoadFromYAML(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	section, exists := mockConfig.Sections["test_section"]
	if !exists {
		t.Fatal("Expected test_section to exist")
	}

	if section.PathPattern != "/api/test/*" {
		t.Errorf("Expected PathPattern to be '/api/test/*', got '%s'", section.PathPattern)
	}
}

// TestUniConfig_LoadFromYAML_UnifiedFormat tests loading the new unified format
func TestUniConfig_LoadFromYAML_UnifiedFormat(t *testing.T) {
	yamlContent := `sections:
  users:
    path_pattern: "/api/users/*"
    id_extraction:
      body_paths: ["/id"]
      header_names: ["X-User-ID", "Authorization"]
    return_body: true
  
scenarios:
  - uuid: "test-scenario"
    method: "GET"
    path: "/api/users/999"
    response:
      status_code: 404
      body: '{"error": "Not found"}'
`

	tempFile, err := os.CreateTemp("", "test_unified_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	mockConfig, err := config.LoadFromYAML(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load unified config: %v", err)
	}

	section, exists := mockConfig.Sections["users"]
	if !exists {
		t.Fatal("Expected users section to exist")
	}

	if section.PathPattern != "/api/users/*" {
		t.Errorf("Expected PathPattern to be '/api/users/*', got '%s'", section.PathPattern)
	}

	expectedHeaders := []string{"X-User-ID", "Authorization"}
	if len(section.HeaderIDNames) != len(expectedHeaders) {
		t.Errorf("Expected %d header names, got %d", len(expectedHeaders), len(section.HeaderIDNames))
	}

	for i, expected := range expectedHeaders {
		if section.HeaderIDNames[i] != expected {
			t.Errorf("Expected header %d to be '%s', got '%s'", i, expected, section.HeaderIDNames[i])
		}
	}

	if !section.ReturnBody {
		t.Errorf("Expected ReturnBody to be true, got %v", section.ReturnBody)
	}
}
