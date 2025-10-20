package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

func TestFixtureResolver_ResolveFixture_ValidFileReference(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "operations")

	// Create directories
	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures dir: %v", err)
	}

	// Create fixture file
	fixtureFile := filepath.Join(fixturesDir, "robots.json")
	fixtureContent := `{"robots": [{"id": "C10190", "status": "active"}]}`
	err = os.WriteFile(fixtureFile, []byte(fixtureContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	resolver := config.NewFixtureResolver(tempDir)

	// Test
	result, err := resolver.ResolveFixture("@fixtures/operations/robots.json")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != fixtureContent {
		t.Errorf("Expected fixture content %q, got %q", fixtureContent, result)
	}
}

func TestFixtureResolver_ResolveFixture_InlineData(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)
	inlineData := `{"status": "ok"}`

	// Test
	result, err := resolver.ResolveFixture(inlineData)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for inline data, got: %v", err)
	}
	if result != inlineData {
		t.Errorf("Expected inline data %q, got %q", inlineData, result)
	}
}

func TestFixtureResolver_ResolveFixture_FileNotFound(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)

	// Test
	_, err := resolver.ResolveFixture("@fixtures/nonexistent.json")

	// Assert
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

func TestFixtureResolver_ResolveFixture_InvalidPath(t *testing.T) {
	resolver := config.NewFixtureResolver(t.TempDir())

	// Test paths that should error
	testErrorPaths(t, resolver, []string{
		"@fixtures/../../../etc/passwd",
		"@/etc/passwd",
		"@",
	})

	// Test inline data (no @ prefix)
	testInlineDataPath(t, resolver, "fixtures/data.json")
}

func testErrorPaths(t *testing.T, resolver *config.FixtureResolver, paths []string) {
	t.Helper()
	for _, path := range paths {
		t.Run("error for "+path, func(t *testing.T) {
			_, err := resolver.ResolveFixture(path)
			if err == nil {
				t.Errorf("Expected error for path %q, got nil", path)
			}
		})
	}
}

func testInlineDataPath(t *testing.T, resolver *config.FixtureResolver, path string) {
	t.Helper()
	_, err := resolver.ResolveFixture(path)
	if err != nil {
		t.Errorf("Expected no error for inline path %q, got: %v", path, err)
	}
}

func TestFixtureResolver_ResolveFixture_Caching(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures dir: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "test.json")
	fixtureContent := `{"test": "data"}`
	err = os.WriteFile(fixtureFile, []byte(fixtureContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	resolver := config.NewFixtureResolver(tempDir)

	// Test multiple calls to same file
	result1, err1 := resolver.ResolveFixture("@fixtures/test.json")
	result2, err2 := resolver.ResolveFixture("@fixtures/test.json")

	// Assert
	if err1 != nil || err2 != nil {
		t.Errorf("Expected no errors, got: %v, %v", err1, err2)
	}
	if result1 != fixtureContent || result2 != fixtureContent {
		t.Errorf("Expected fixture content %q, got %q and %q", fixtureContent, result1, result2)
	}
	if result1 != result2 {
		t.Error("Expected cached results to be identical")
	}
}

func TestFixtureResolver_ResolveFixture_XmlFile(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures dir: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "data.xml")
	fixtureContent := `<?xml version="1.0"?><data><item>test</item></data>`
	err = os.WriteFile(fixtureFile, []byte(fixtureContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	resolver := config.NewFixtureResolver(tempDir)

	// Test
	result, err := resolver.ResolveFixture("@fixtures/data.xml")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != fixtureContent {
		t.Errorf("Expected XML content %q, got %q", fixtureContent, result)
	}
}
