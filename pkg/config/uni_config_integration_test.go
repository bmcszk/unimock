package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

// TDD-23: Create failing test for scenario config integration
func TestUniConfig_ToModelScenario_WithFixtureResolver_IntegratesFixtureResolution(t *testing.T) {
	// Arrange: Create temporary directory and fixture file
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "users")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "user_123.json")
	expectedFixtureContent := `{"id": "123", "name": "Test User", "email": "test@example.com"}`
	err = os.WriteFile(fixtureFile, []byte(expectedFixtureContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	// Create scenario config with fixture reference
	scenarioConfig := &config.ScenarioConfig{
		UUID:        "test-user",
		Method:      "GET",
		Path:        "/api/users/123",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        "@fixtures/users/user_123.json", // This should be resolved
	}

	// Act: Create resolver and convert scenario
	resolver := config.NewFixtureResolver(tempDir)
	result := scenarioConfig.ToModelScenario(resolver)

	// Assert: Should resolve fixture content in Data field
	expectedRequestPath := "GET /api/users/123"
	if result.RequestPath != expectedRequestPath {
		t.Errorf("Expected RequestPath %q, got %q", expectedRequestPath, result.RequestPath)
	}
	if result.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d", result.StatusCode)
	}
	if result.ContentType != "application/json" {
		t.Errorf("Expected ContentType %q, got %q", "application/json", result.ContentType)
	}
	if result.Data != expectedFixtureContent {
		t.Errorf("Expected Data %q, got %q", expectedFixtureContent, result.Data)
	}
}

// Test integration with < syntax
func TestUniConfig_ToModelScenario_WithFixtureResolver_IntegratesLessThanSyntax(t *testing.T) {
	// Arrange: Create temporary directory and fixture file
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "products")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "product_456.json")
	expectedFixtureContent := `{"id": "456", "name": "Test Product", "price": 29.99}`
	err = os.WriteFile(fixtureFile, []byte(expectedFixtureContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	// Create scenario config with < syntax fixture reference
	scenarioConfig := &config.ScenarioConfig{
		UUID:        "test-product",
		Method:      "GET",
		Path:        "/api/products/456",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        "< ./fixtures/products/product_456.json", // This should be resolved
	}

	// Act: Create resolver and convert scenario
	resolver := config.NewFixtureResolver(tempDir)
	result := scenarioConfig.ToModelScenario(resolver)

	// Assert: Should resolve fixture content in Data field
	if result.Data != expectedFixtureContent {
		t.Errorf("Expected Data %q, got %q", expectedFixtureContent, result.Data)
	}
}

// Test integration with inline fixtures
func TestUniConfig_ToModelScenario_WithFixtureResolver_IntegratesInlineFixtures(t *testing.T) {
	// Arrange: Create temporary directory and fixture files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	userFile := filepath.Join(fixturesDir, "user.json")
	userContent := `{"id": "123", "name": "Test User"}`
	err = os.WriteFile(userFile, []byte(userContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create user fixture file: %v", err)
	}

	// Create scenario config with inline fixture reference
	scenarioConfig := &config.ScenarioConfig{
		UUID:        "test-inline",
		Method:      "GET",
		Path:        "/api/users/123/inline",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"user": < ./fixtures/user.json, "timestamp": "2024-01-15T10:30:00Z"}`,
	}

	// Act: Create resolver and convert scenario
	resolver := config.NewFixtureResolver(tempDir)
	result := scenarioConfig.ToModelScenario(resolver)

	// Assert: Should resolve inline fixture content
	expectedData := `{"user": {"id": "123", "name": "Test User"}, "timestamp": "2024-01-15T10:30:00Z"}`
	if result.Data != expectedData {
		t.Errorf("Expected Data %q, got %q", expectedData, result.Data)
	}
}

// Test that inline data passes through unchanged (backward compatibility)
func TestUniConfig_ToModelScenario_WithFixtureResolver_PassesThroughInlineData(t *testing.T) {
	// Arrange: Create scenario config with inline data
	inlineData := `{"message": "This is inline data", "status": "ok"}`
	scenarioConfig := &config.ScenarioConfig{
		UUID:        "test-inline-data",
		Method:      "GET",
		Path:        "/api/inline",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        inlineData,
	}

	// Act: Create resolver and convert scenario
	resolver := config.NewFixtureResolver(t.TempDir())
	result := scenarioConfig.ToModelScenario(resolver)

	// Assert: Should pass through inline data unchanged
	if result.Data != inlineData {
		t.Errorf("Expected Data %q, got %q", inlineData, result.Data)
	}
}
