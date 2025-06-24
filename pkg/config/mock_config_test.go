package config_test

import (
	"os"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

func TestMockConfig_ReturnBodyDefault(t *testing.T) {
	// Test that ReturnBody defaults to false
	mockConfig := config.NewMockConfig()
	
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

func TestMockConfig_LoadFromYAML_ReturnBodyFalse(t *testing.T) {
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

func TestMockConfig_LoadFromYAML_ReturnBodyTrue(t *testing.T) {
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

func TestMockConfig_LoadFromYAML_ReturnBodyOmitted(t *testing.T) {
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