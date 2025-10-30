package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

// TestBackwardCompatibility_VerifiesLegacyAndEnhancedSyntaxWork ensures backward compatibility
// This test verifies that both legacy @fixtures syntax and new enhanced syntax work together
func TestBackwardCompatibility_VerifiesLegacyAndEnhancedSyntaxWork(t *testing.T) {
	// Arrange: Create temporary directory with mixed fixture syntax configuration
	tempDir := t.TempDir()
	userContent, productContent := setupBackwardCompatibilityFixtures(t, tempDir)
	configFile := createMixedSyntaxConfig(t, tempDir)

	// Act: Load configuration
	uniConfig, err := config.LoadFromYAML(configFile)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Assert: Verify configuration loaded correctly
	assertBackwardCompatibilityConfiguration(t, uniConfig)
	assertLegacyFixtureSyntax(t, uniConfig, userContent)
	assertEnhancedFixtureSyntax(t, uniConfig, productContent)
	assertInlineFixtureSyntax(t, uniConfig, userContent)
	assertDirectDataSyntax(t, uniConfig)
	assertScenarioProperties(t, uniConfig)
}

// setupBackwardCompatibilityFixtures creates test fixture files for backward compatibility testing
func setupBackwardCompatibilityFixtures(t *testing.T, tempDir string) (userContent string, productContent string) {
	t.Helper()

	// Create fixtures directory
	fixturesDir := filepath.Join(tempDir, "fixtures")
	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	// Create test fixture files
	userFixture := filepath.Join(fixturesDir, "user.json")
	userContent = `{"id": "123", "name": "Legacy User", "email": "legacy@example.com"}`
	err = os.WriteFile(userFixture, []byte(userContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create user fixture: %v", err)
	}

	productFixture := filepath.Join(fixturesDir, "product.json")
	productContent = `{"id": "456", "name": "New Product", "price": 29.99}`
	err = os.WriteFile(productFixture, []byte(productContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create product fixture: %v", err)
	}

	return userContent, productContent
}

// createMixedSyntaxConfig creates a configuration file with mixed fixture syntax
func createMixedSyntaxConfig(t *testing.T, tempDir string) string {
	t.Helper()

	configContent := `
# Mixed configuration using both legacy and enhanced fixture syntax
scenarios:
  # Legacy @fixtures syntax (backward compatibility)
  - uuid: "legacy-user-scenario"
    method: "GET"
    path: "/api/legacy-user"
    status_code: 200
    data: "@fixtures/user.json"
    headers:
      Content-Type: "application/json"

  # Enhanced < syntax (go-restclient compatible)
  - uuid: "enhanced-product-scenario"
    method: "GET"
    path: "/api/enhanced-product"
    status_code: 200
    data: "< ./fixtures/product.json"
    headers:
      Content-Type: "application/json"

  # Enhanced inline fixture syntax
  - uuid: "inline-fixture-scenario"
    method: "POST"
    path: "/api/inline-user"
    status_code: 201
    data: '{"user": < ./fixtures/user.json, "created": true}'
    headers:
      Content-Type: "application/json"

  # Direct string data (no fixture resolution)
  - uuid: "direct-data-scenario"
    method: "GET"
    path: "/api/direct"
    status_code: 200
    data: '{"message": "direct data without fixture"}'
    headers:
      Content-Type: "application/json"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	return configFile
}

// assertBackwardCompatibilityConfiguration verifies basic configuration loading
func assertBackwardCompatibilityConfiguration(t *testing.T, uniConfig *config.UniConfig) {
	t.Helper()

	if len(uniConfig.Scenarios) != 4 {
		t.Errorf("Expected 4 scenarios, got %d", len(uniConfig.Scenarios))
	}

	// Verify fixture resolver is initialized
	if uniConfig.GetFixtureResolver() == nil {
		t.Error("Fixture resolver should be initialized")
	}
}

// assertLegacyFixtureSyntax verifies @fixtures syntax works
func assertLegacyFixtureSyntax(t *testing.T, uniConfig *config.UniConfig, expectedUserContent string) {
	t.Helper()

	legacyScenario := uniConfig.Scenarios[0]
	legacyModel := legacyScenario.ToModelScenario(uniConfig.GetFixtureResolver())
	if legacyModel.Data != expectedUserContent {
		t.Errorf("Legacy scenario data mismatch\nExpected: %s\nGot: %s", expectedUserContent, legacyModel.Data)
	}
}

// assertEnhancedFixtureSyntax verifies < syntax works
func assertEnhancedFixtureSyntax(t *testing.T, uniConfig *config.UniConfig, expectedProductContent string) {
	t.Helper()

	enhancedScenario := uniConfig.Scenarios[1]
	enhancedModel := enhancedScenario.ToModelScenario(uniConfig.GetFixtureResolver())
	if enhancedModel.Data != expectedProductContent {
		t.Errorf("Enhanced scenario data mismatch\nExpected: %s\nGot: %s", expectedProductContent, enhancedModel.Data)
	}
}

// assertInlineFixtureSyntax verifies inline fixture syntax works
func assertInlineFixtureSyntax(t *testing.T, uniConfig *config.UniConfig, _ string) {
	t.Helper()

	inlineScenario := uniConfig.Scenarios[2]
	inlineModel := inlineScenario.ToModelScenario(uniConfig.GetFixtureResolver())
	expectedInline := `{"user": {"id": "123", "name": "Legacy User", "email": "legacy@example.com"}, "created": true}`
	if inlineModel.Data != expectedInline {
		t.Errorf("Inline scenario data mismatch\nExpected: %s\nGot: %s", expectedInline, inlineModel.Data)
	}
}

// assertDirectDataSyntax verifies direct data (no fixture resolution) works
func assertDirectDataSyntax(t *testing.T, uniConfig *config.UniConfig) {
	t.Helper()

	directScenario := uniConfig.Scenarios[3]
	directModel := directScenario.ToModelScenario(uniConfig.GetFixtureResolver())
	expectedDirect := `{"message": "direct data without fixture"}`
	if directModel.Data != expectedDirect {
		t.Errorf("Direct scenario data mismatch\nExpected: %s\nGot: %s", expectedDirect, directModel.Data)
	}
}

// assertScenarioProperties verifies scenario properties are preserved
func assertScenarioProperties(t *testing.T, uniConfig *config.UniConfig) {
	t.Helper()

	legacyScenario := uniConfig.Scenarios[0]
	legacyModel := legacyScenario.ToModelScenario(uniConfig.GetFixtureResolver())

	if legacyModel.UUID != "legacy-user-scenario" {
		t.Errorf("Expected UUID 'legacy-user-scenario', got '%s'", legacyModel.UUID)
	}
	if legacyModel.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", legacyModel.StatusCode)
	}
	if legacyModel.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", legacyModel.ContentType)
	}
}

// TestBackwardCompatibility_VerifiesFixtureErrorHandling tests error handling in fixture resolution
func TestBackwardCompatibility_VerifiesFixtureErrorHandling(t *testing.T) {
	// Arrange: Create configuration with invalid fixture reference
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
scenarios:
  - uuid: "invalid-fixture-scenario"
    method: "GET"
    path: "/api/invalid"
    status_code: 200
    data: "@fixtures/nonexistent.json"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Act: Load configuration
	uniConfig, err := config.LoadFromYAML(configFile)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Assert: Verify graceful fallback to original data when fixture resolution fails
	scenario := uniConfig.Scenarios[0]
	model := scenario.ToModelScenario(uniConfig.GetFixtureResolver())

	// Should fallback to original fixture reference when file doesn't exist
	expectedData := "@fixtures/nonexistent.json"
	if model.Data != expectedData {
		t.Errorf("Expected fallback to original data '%s', got '%s'", expectedData, model.Data)
	}
}

// TestBackwardCompatibility_VerifiesConfigurationIntegration tests integration with server
func TestBackwardCompatibility_VerifiesConfigurationIntegration(t *testing.T) {
	// This test verifies that the enhanced fixture support integrates properly
	// with the existing configuration system and doesn't break existing functionality

	// Arrange: Create a complete configuration with sections and scenarios
	tempDir := t.TempDir()
	orderContent := setupConfigurationIntegrationFixtures(t, tempDir)
	configFile := createComprehensiveConfig(t, tempDir)

	// Act: Load configuration
	uniConfig, err := config.LoadFromYAML(configFile)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Assert: Verify both sections and scenarios work correctly
	assertConfigurationIntegration(t, uniConfig, orderContent)
}

// setupConfigurationIntegrationFixtures creates fixtures for integration testing
func setupConfigurationIntegrationFixtures(t *testing.T, tempDir string) string {
	t.Helper()

	// Create fixtures directory and files
	fixturesDir := filepath.Join(tempDir, "fixtures")
	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	orderFixture := filepath.Join(fixturesDir, "order.json")
	orderContent := `{"id": "789", "status": "pending", "total": 99.99}`
	err = os.WriteFile(orderFixture, []byte(orderContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create order fixture: %v", err)
	}

	return orderContent
}

// createComprehensiveConfig creates a complete configuration with sections and scenarios
func createComprehensiveConfig(t *testing.T, tempDir string) string {
	t.Helper()

	configContent := `
# Complete configuration with sections and enhanced scenarios
sections:
  orders:
    path_pattern: "/api/orders/*"
    body_id_paths: ["/id"]
    header_id_names: ["X-Order-ID"]
    return_body: true
    strict_path: false

scenarios:
  - uuid: "order-created"
    method: "POST"
    path: "/api/orders"
    status_code: 201
    data: "< ./fixtures/order.json"
    headers:
      Content-Type: "application/json"
      Location: "/api/orders/789"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	return configFile
}

// assertConfigurationIntegration verifies configuration integration works correctly
func assertConfigurationIntegration(t *testing.T, uniConfig *config.UniConfig, expectedOrderContent string) {
	t.Helper()

	// Verify section and scenario counts
	if len(uniConfig.Sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(uniConfig.Sections))
	}

	if len(uniConfig.Scenarios) != 1 {
		t.Errorf("Expected 1 scenario, got %d", len(uniConfig.Scenarios))
	}

	// Verify section configuration
	assertSectionConfiguration(t, uniConfig)

	// Verify scenario with fixture resolution
	assertScenarioWithFixtureResolution(t, uniConfig, expectedOrderContent)

	// Verify path matching still works
	assertPathMatchingIntegration(t, uniConfig)
}

// assertSectionConfiguration verifies section configuration is correct
func assertSectionConfiguration(t *testing.T, uniConfig *config.UniConfig) {
	t.Helper()

	orderSection := uniConfig.Sections["orders"]
	if orderSection.PathPattern != "/api/orders/*" {
		t.Errorf("Expected path pattern '/api/orders/*', got '%s'", orderSection.PathPattern)
	}
	if !orderSection.ReturnBody {
		t.Error("Expected ReturnBody to be true")
	}
}

// assertScenarioWithFixtureResolution verifies scenario fixture resolution works
func assertScenarioWithFixtureResolution(t *testing.T, uniConfig *config.UniConfig, expectedOrderContent string) {
	t.Helper()

	scenario := uniConfig.Scenarios[0]
	model := scenario.ToModelScenario(uniConfig.GetFixtureResolver())
	if model.Data != expectedOrderContent {
		t.Errorf("Expected resolved fixture data '%s', got '%s'", expectedOrderContent, model.Data)
	}
}

// assertPathMatchingIntegration verifies path matching integration works
func assertPathMatchingIntegration(t *testing.T, uniConfig *config.UniConfig) {
	t.Helper()

	sectionName, matchedSection, err := uniConfig.MatchPath("/api/orders/123")
	if err != nil {
		t.Errorf("Path matching failed: %v", err)
	}
	if sectionName != "orders" {
		t.Errorf("Expected section name 'orders', got '%s'", sectionName)
	}
	if matchedSection == nil {
		t.Error("Expected matched section, got nil")
	}
}
