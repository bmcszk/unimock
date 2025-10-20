package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestUniConfig_LoadFromYAML_FixtureFileSupport(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "operations")

	// Create directories
	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures dir: %v", err)
	}

	// Create fixture files
	robotsFile := filepath.Join(fixturesDir, "robots.json")
	robotsContent := `{"robots": [{"id": "C10190", "status": "active", "name": "Robot C10190"}]}`
	err = os.WriteFile(robotsFile, []byte(robotsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create robots fixture file: %v", err)
	}

	statusFile := filepath.Join(fixturesDir, "status_C10190.json")
	statusContent := `{"status": "active", "battery": 85, "last_check": "2023-10-20T10:30:00Z"}`
	err = os.WriteFile(statusFile, []byte(statusContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create status fixture file: %v", err)
	}

	// Create configuration file with fixture references in temp root
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `sections:
  operations:
    path_pattern: "/internal/robots*"
    return_body: true

scenarios:
  - uuid: "list-robots"
    method: "GET"
    path: "/internal/robots"
    status_code: 200
    data: "@fixtures/operations/robots.json"

  - uuid: "robot-status-C10190"
    method: "GET"
    path: "/robots/C10190/status"
    status_code: 200
    data: "@fixtures/operations/status_C10190.json"
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load configuration
	uniConfig, err := config.LoadFromYAML(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify sections and scenarios
	verifyFixtureFileSupportSections(t, uniConfig)
	verifyFixtureFileSupportScenarios(t, uniConfig, robotsContent, statusContent)
}

func verifyFixtureFileSupportSections(t *testing.T, uniConfig *config.UniConfig) {
	t.Helper()
	if len(uniConfig.Sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(uniConfig.Sections))
	}

	operationsSection, exists := uniConfig.Sections["operations"]
	if !exists {
		t.Error("Expected operations section not found")
	}
	if operationsSection.PathPattern != "/internal/robots*" {
		t.Errorf("Expected path pattern '/internal/robots*', got '%s'", operationsSection.PathPattern)
	}
	if !operationsSection.ReturnBody {
		t.Error("Expected ReturnBody to be true")
	}
}

func verifyFixtureFileSupportScenarios(t *testing.T, uniConfig *config.UniConfig, robotsContent, statusContent string) {
	t.Helper()
	if len(uniConfig.Scenarios) != 2 {
		t.Errorf("Expected 2 scenarios, got %d", len(uniConfig.Scenarios))
	}

	// Test scenario conversion to model with fixture resolution
	scenarios := make([]model.Scenario, len(uniConfig.Scenarios))
	resolver := uniConfig.GetFixtureResolver()
	for i, scenarioConfig := range uniConfig.Scenarios {
		scenarios[i] = scenarioConfig.ToModelScenario(resolver)
	}

	// Verify scenarios
	verifyScenario(t, scenarios[0], "list-robots", "GET /internal/robots", 200, robotsContent)
	verifyScenario(t, scenarios[1], "robot-status-C10190", "GET /robots/C10190/status", 200, statusContent)
}

func verifyScenario(
	t *testing.T,
	scenario model.Scenario,
	expectedUUID string,
	expectedPath string,
	expectedStatusCode int,
	expectedData string,
) {
	t.Helper()
	if scenario.UUID != expectedUUID {
		t.Errorf("Expected UUID '%s', got '%s'", expectedUUID, scenario.UUID)
	}
	if scenario.RequestPath != expectedPath {
		t.Errorf("Expected request path '%s', got '%s'", expectedPath, scenario.RequestPath)
	}
	if scenario.StatusCode != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d", expectedStatusCode, scenario.StatusCode)
	}
	if scenario.Data != expectedData {
		t.Errorf("Expected data content, got '%s'", scenario.Data)
	}
}

func TestUniConfig_LoadFromYAML_MixedInlineAndFixtureData(t *testing.T) {
	tempDir := t.TempDir()
	fixtureContent := createTestFixture(t, tempDir, "fixtures", "data.json", `{"fixture": "data"}`)

	configFile := createTestConfig(t, tempDir, `scenarios:
  - uuid: "inline-data"
    method: "GET"
    path: "/inline"
    status_code: 200
    data: '{"inline": "data"}'

  - uuid: "fixture-data"
    method: "GET"
    path: "/fixture"
    status_code: 200
    data: "@fixtures/data.json"
`)

	uniConfig := loadConfigOrFatal(t, configFile)
	verifyMixedDataScenarios(t, uniConfig, fixtureContent)
}

func createTestFixture(t *testing.T, tempDir, fixtureDir, filename, content string) string {
	t.Helper()
	fixturesDir := filepath.Join(tempDir, fixtureDir)

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures dir: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, filename)
	err = os.WriteFile(fixtureFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}
	return content
}

func createTestConfig(t *testing.T, tempDir, configContent string) string {
	t.Helper()
	configFile := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	return configFile
}

func loadConfigOrFatal(t *testing.T, configFile string) *config.UniConfig {
	t.Helper()
	uniConfig, err := config.LoadFromYAML(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	return uniConfig
}

func verifyMixedDataScenarios(t *testing.T, uniConfig *config.UniConfig, fixtureContent string) {
	t.Helper()
	if len(uniConfig.Scenarios) != 2 {
		t.Errorf("Expected 2 scenarios, got %d", len(uniConfig.Scenarios))
	}

	scenarios := make([]model.Scenario, len(uniConfig.Scenarios))
	resolver := uniConfig.GetFixtureResolver()
	for i, scenarioConfig := range uniConfig.Scenarios {
		scenarios[i] = scenarioConfig.ToModelScenario(resolver)
	}

	// Verify inline data scenario
	inlineScenario := scenarios[0]
	if inlineScenario.Data != `{"inline": "data"}` {
		t.Errorf("Expected inline data, got '%s'", inlineScenario.Data)
	}

	// Verify fixture data scenario
	fixtureScenario := scenarios[1]
	if fixtureScenario.Data != fixtureContent {
		t.Errorf("Expected fixture content, got '%s'", fixtureScenario.Data)
	}
}

func TestUniConfig_LoadFromYAML_FixtureFileNotFound(t *testing.T) {
	// Setup
	tempDir := t.TempDir()

	// Create configuration file with non-existent fixture reference
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `scenarios:
  - uuid: "missing-fixture"
    method: "GET"
    path: "/missing"
    status_code: 200
    data: "@fixtures/nonexistent.json"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load configuration - should still work but ToModelScenario() will fail when called
	uniConfig, err := config.LoadFromYAML(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Config should load successfully
	if len(uniConfig.Scenarios) != 1 {
		t.Errorf("Expected 1 scenario, got %d", len(uniConfig.Scenarios))
	}

	// But ToModelScenario() should fail due to missing file
	scenario := uniConfig.Scenarios[0]
	resolver := uniConfig.GetFixtureResolver()
	modelScenario := scenario.ToModelScenario(resolver)
	// The current implementation doesn't return errors from ToModelScenario(), so this test
	// just verifies the behavior is consistent with current design
	if modelScenario.Data != "" {
		t.Errorf("Expected empty data for missing fixture, got '%s'", modelScenario.Data)
	}
}
