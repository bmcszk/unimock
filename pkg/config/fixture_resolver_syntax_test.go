package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

// Helper function to set up test resolver with fixture file
func setupTestResolver(t *testing.T) (*config.FixtureResolver, string) {
	t.Helper()
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures")

	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "test.json")
	if err := os.WriteFile(fixtureFile, []byte(`{"test": "data"}`), 0644); err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	return config.NewFixtureResolver(tempDir), `{"test": "data"}`
}

// Test that <./path (no space after <) is returned as-is
func TestFixtureResolver_ResolveFixture_NoSpaceAfterLessThan_ReturnsAsIs(t *testing.T) {
	resolver, _ := setupTestResolver(t)
	input := "<./fixtures/test.json"
	result, err := resolver.ResolveFixture(input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != input {
		t.Errorf("Expected %q, got %q", input, result)
	}
}

// Test that <@ ./path (with space after @) resolves file
func TestFixtureResolver_ResolveFixture_AtSyntax_ResolvesFile(t *testing.T) {
	resolver, expectedContent := setupTestResolver(t)
	input := "<@ ./fixtures/test.json"
	result, err := resolver.ResolveFixture(input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expectedContent {
		t.Errorf("Expected %q, got %q", expectedContent, result)
	}
}

// Test that <@./path (no space after @) is returned as-is
func TestFixtureResolver_ResolveFixture_NoSpaceAfterAt_ReturnsAsIs(t *testing.T) {
	resolver, _ := setupTestResolver(t)
	input := "<@./fixtures/test.json"
	result, err := resolver.ResolveFixture(input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != input {
		t.Errorf("Expected %q, got %q", input, result)
	}
}

// Test that < ./path (with space) resolves file
func TestFixtureResolver_ResolveFixture_SpaceAfterLessThan_ResolvesFile(t *testing.T) {
	resolver, expectedContent := setupTestResolver(t)
	input := "< ./fixtures/test.json"
	result, err := resolver.ResolveFixture(input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expectedContent {
		t.Errorf("Expected %q, got %q", expectedContent, result)
	}
}

// Test that < @./path (space after < but not after @) is returned as-is (invalid)
func TestFixtureResolver_ResolveFixture_SpaceBeforeAtNoSpaceAfter_ReturnsAsIs(t *testing.T) {
	resolver, _ := setupTestResolver(t)
	input := "< @./fixtures/test.json"
	result, err := resolver.ResolveFixture(input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != input {
		t.Errorf("Expected %q, got %q", input, result)
	}
}

// Test that <  ./path (multiple spaces) resolves file
func TestFixtureResolver_ResolveFixture_MultipleSpaces_ResolvesFile(t *testing.T) {
	resolver, expectedContent := setupTestResolver(t)
	input := "<  ./fixtures/test.json"
	result, err := resolver.ResolveFixture(input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expectedContent {
		t.Errorf("Expected %q, got %q", expectedContent, result)
	}
}
