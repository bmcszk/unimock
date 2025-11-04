package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

// TDD-01: Create failing test for @fixtures/file.json syntax
func TestFixtureResolver_ResolveFixture_AtSyntax_SupportsBasicFixtureResolution(t *testing.T) {
	// Arrange: Create temporary directory and fixture file
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "operations")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "robots.json")
	expectedContent := `{"robots": [{"id": "R001", "name": "Alpha Robot", "status": "active"}]}`
	err = os.WriteFile(fixtureFile, []byte(expectedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	// Act: Create resolver and resolve fixture
	resolver := config.NewFixtureResolver(tempDir)
	result, err := resolver.ResolveFixture("@fixtures/operations/robots.json")

	// Assert: Should successfully resolve fixture content
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, result)
	}
}

// TDD-03: Create failing test for < ./fixtures/file.json syntax
func TestFixtureResolver_ResolveFixture_LessThanSyntax_SupportsGoRestClientCompatibleSyntax(t *testing.T) {
	// Arrange: Create temporary directory and fixture file
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "users")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "user_123.json")
	expectedContent := `{"id": "123", "name": "Test User", "email": "test@example.com"}`
	err = os.WriteFile(fixtureFile, []byte(expectedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	// Act: Create resolver and resolve fixture using < syntax
	resolver := config.NewFixtureResolver(tempDir)
	result, err := resolver.ResolveFixture("< ./fixtures/users/user_123.json")

	// Assert: Should successfully resolve fixture content
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, result)
	}
}

// TDD-05: Create failing test for <@ ./fixtures/file.json syntax
func TestFixtureResolver_ResolveFixture_LessAtSyntax_SupportsVariableSubstitutionSyntax(t *testing.T) {
	// Arrange: Create temporary directory and fixture file
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "products")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "product_456.json")
	expectedContent := `{"id": "456", "name": "Test Product", "price": 29.99}`
	err = os.WriteFile(fixtureFile, []byte(expectedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	// Act: Create resolver and resolve fixture using < @ syntax (space required after <)
	resolver := config.NewFixtureResolver(tempDir)
	result, err := resolver.ResolveFixture("< @ ./fixtures/products/product_456.json")

	// Assert: Should successfully resolve fixture content
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, result)
	}
}

// TDD-11: Create failing test for inline {"key": < ./file.json} syntax
func TestFixtureResolver_ResolveFixture_InlineFixture_SingleReference_SupportsBasicInlineResolution(t *testing.T) {
	// Arrange: Create temporary directory and fixture file
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "users")

	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	fixtureFile := filepath.Join(fixturesDir, "user.json")
	userContent := `{"id": "123", "name": "Test User"}`
	err = os.WriteFile(fixtureFile, []byte(userContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture file: %v", err)
	}

	// Act: Create resolver and resolve inline fixture
	resolver := config.NewFixtureResolver(tempDir)
	result, err := resolver.ResolveFixture(`{"user": < ./fixtures/users/user.json}`)

	// Assert: Should successfully resolve inline fixture
	expected := `{"user": {"id": "123", "name": "Test User"}}`
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expected {
		t.Errorf("Expected content %q, got %q", expected, result)
	}
}

// TDD-13: Create failing test for multiple inline fixtures
func TestFixtureResolver_ResolveFixture_InlineFixture_MultipleReferences_SupportsMultipleInlineFixtures(t *testing.T) {
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

	permissionsFile := filepath.Join(fixturesDir, "permissions.json")
	permissionsContent := `{"admin": false, "read": true, "write": false}`
	err = os.WriteFile(permissionsFile, []byte(permissionsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create permissions fixture file: %v", err)
	}

	// Act: Create resolver and resolve multiple inline fixtures
	resolver := config.NewFixtureResolver(tempDir)
	input := `{"user": < ./fixtures/user.json, "permissions": < ./fixtures/permissions.json}`
	result, err := resolver.ResolveFixture(input)

	// Assert: Should successfully resolve multiple inline fixtures
	expected := `{"user": {"id": "123", "name": "Test User"}, "permissions": {` +
		`"admin": false, "read": true, "write": false}}`
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != expected {
		t.Errorf("Expected content %q, got %q", expected, result)
	}
}

// TDD-17: Create failing test for path traversal attacks
func TestFixtureResolver_ResolveFixture_PathTraversal_AtSyntax_PreventsSecurityAttacks(t *testing.T) {
	resolver := config.NewFixtureResolver(t.TempDir())

	traversalPaths := []string{
		"@fixtures/../../../etc/passwd",
		"@fixtures/../../../../root/.ssh/id_rsa",
		"@fixtures/..\\..\\windows\\system32\\config\\sam",
	}

	for _, path := range traversalPaths {
		t.Run("path traversal test for "+path, func(t *testing.T) {
			assertPathTraversalBlocked(t, resolver, path)
		})
	}
}

// assertPathTraversalBlocked helper for path traversal testing
func assertPathTraversalBlocked(t *testing.T, resolver *config.FixtureResolver, path string) {
	t.Helper()
	_, err := resolver.ResolveFixture(path)
	if err == nil {
		t.Errorf("Expected security error for path %q, got nil", path)
	}
	var pathErr *config.InvalidFixturePathError
	if !errors.As(err, &pathErr) || !strings.Contains(pathErr.Error(), "path traversal") {
		t.Errorf("Expected path traversal error for %q, got: %v", path, err)
	}
}

// TDD-19: Create failing test for absolute paths
func TestFixtureResolver_ResolveFixture_AbsolutePath_AtSyntax_PreventsAbsolutePaths(t *testing.T) {
	resolver := config.NewFixtureResolver(t.TempDir())

	absolutePaths := []string{
		"@/etc/passwd",
		"@/tmp/secrets.txt",
		"@C:\\Windows\\System32\\config\\sam",
	}

	for _, path := range absolutePaths {
		t.Run("absolute path test for "+path, func(t *testing.T) {
			assertAbsolutePathBlocked(t, resolver, path)
		})
	}
}

// assertAbsolutePathBlocked helper for absolute path testing
func assertAbsolutePathBlocked(t *testing.T, resolver *config.FixtureResolver, path string) {
	t.Helper()
	_, err := resolver.ResolveFixture(path)
	if err == nil {
		t.Errorf("Expected security error for absolute path %q, got nil", path)
	}
	var pathErr *config.InvalidFixturePathError
	if !errors.As(err, &pathErr) || !strings.Contains(pathErr.Error(), "absolute path") {
		t.Errorf("Expected absolute path error for %q, got: %v", path, err)
	}
}

// TDD-21: Create failing test for empty references
func TestFixtureResolver_ResolveFixture_EmptyReference_AtSyntax_PreventsEmptyReferences(t *testing.T) {
	// Arrange: Create resolver
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)

	// Act: Try to resolve empty reference
	_, err := resolver.ResolveFixture("@")

	// Assert: Should return error for empty reference
	if err == nil {
		t.Error("Expected error for empty reference, got nil")
	}
	var pathErr *config.InvalidFixturePathError
	if !errors.As(err, &pathErr) || !strings.Contains(pathErr.Error(), "empty path") {
		t.Errorf("Expected empty path error, got: %v", err)
	}
}

// TDD-09: Create failing test for missing files
func TestFixtureResolver_ResolveFixture_MissingFile_ReturnsError(t *testing.T) {
	// Arrange: Create resolver without any fixture files
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)

	// Act: Try to resolve non-existent fixture
	result, err := resolver.ResolveFixture("@fixtures/nonexistent.json")

	// Assert: Should gracefully fallback to original data (no error)
	if err != nil {
		t.Errorf("Expected no error for missing file (graceful fallback), got: %v", err)
	}
	if result != "@fixtures/nonexistent.json" {
		t.Errorf("Expected original fixture reference, got: %q", result)
	}
}

// TDD-22: Create test for graceful fallback when file not found - email addresses
func TestFixtureResolver_ResolveFixture_EmailAddress_ReturnsOriginalData(t *testing.T) {
	// Arrange: Create resolver
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)

	// Act: Try to resolve email address (contains @ but is not a fixture reference)
	emailData := "user@example.com"
	result, err := resolver.ResolveFixture(emailData)

	// Assert: Should return original data and no error since it's not a valid fixture path
	if err != nil {
		t.Errorf("Expected no error for email address, got: %v", err)
	}
	if result != emailData {
		t.Errorf("Expected original email address %q, got %q", emailData, result)
	}
}

// TDD-23: Create test for graceful fallback when fixture file not found
func TestFixtureResolver_ResolveFixture_InvalidFixturePath_ReturnsOriginalData(t *testing.T) {
	// Arrange: Create resolver
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)

	// Act: Try to resolve non-existent fixture (should return original data)
	fixtureData := "@fixtures/nonexistent.json"
	result, err := resolver.ResolveFixture(fixtureData)

	// Assert: Should return original data when file not found (graceful fallback)
	if err != nil {
		t.Errorf("Expected no error for missing fixture (graceful fallback), got: %v", err)
	}
	if result != fixtureData {
		t.Errorf("Expected original fixture reference %q, got %q", fixtureData, result)
	}
}

// TDD-24: Create test for graceful fallback with < syntax when file not found
func TestFixtureResolver_ResolveFixture_LessThanSyntaxMissingFile_ReturnsOriginalData(t *testing.T) {
	// Arrange: Create resolver
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)

	// Act: Try to resolve non-existent fixture with < syntax
	fixtureData := "< ./fixtures/nonexistent.json"
	result, err := resolver.ResolveFixture(fixtureData)

	// Assert: Should return original data when file not found
	if err != nil {
		t.Errorf("Expected no error for missing fixture (graceful fallback), got: %v", err)
	}
	if result != fixtureData {
		t.Errorf("Expected original fixture reference %q, got %q", fixtureData, result)
	}
}

// TDD-25: Create test for inline fixtures with missing files - should return original reference
func TestFixtureResolver_ResolveFixture_InlineFixtureMissingFile_ReturnsOriginalData(t *testing.T) {
	// Arrange: Create resolver
	tempDir := t.TempDir()
	resolver := config.NewFixtureResolver(tempDir)

	// Act: Try to resolve inline fixture with missing file
	inlineData := `{"user": < ./fixtures/nonexistent.json, "status": "active"}`
	result, err := resolver.ResolveFixture(inlineData)

	// Assert: Should return original inline data when file not found
	if err != nil {
		t.Errorf("Expected no error for missing inline fixture (graceful fallback), got: %v", err)
	}
	// For inline fixtures, the missing reference should be preserved
	if !strings.Contains(result, "< ./fixtures/nonexistent.json") {
		t.Errorf("Expected original inline fixture reference to be preserved, got: %q", result)
	}
}
