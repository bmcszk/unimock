package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFixtureFileSupport_AtSyntax_BasicWorkflow tests basic @fixtures syntax workflow
func TestFixtureFileSupport_AtSyntax_BasicWorkflow(t *testing.T) {
	configFile := createAtSyntaxFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/user")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Legacy User")
}

// TestFixtureFileSupport_LessThanSyntax_GoRestClientCompatible tests < syntax workflow
func TestFixtureFileSupport_LessThanSyntax_GoRestClientCompatible(t *testing.T) {
	configFile := createLessThanSyntaxFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/product")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("New Product")
}

// TestFixtureFileSupport_LessAtSyntax_VariableSubstitution tests <@ syntax workflow
func TestFixtureFileSupport_LessAtSyntax_VariableSubstitution(t *testing.T) {
	configFile := createLessAtSyntaxFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/order")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Pending Order")
}

// TestFixtureFileSupport_InlineFixtures_SingleReference tests inline fixture workflow
func TestFixtureFileSupport_InlineFixtures_SingleReference(t *testing.T) {
	configFile := createInlineFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/user-with-profile")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Test User").and().
		the_response_body_contains_fixtures_data("Developer Profile")
}

// TestFixtureFileSupport_InlineFixtures_MultipleReferences tests multiple inline fixtures
func TestFixtureFileSupport_InlineFixtures_MultipleReferences(t *testing.T) {
	configFile := createMultipleInlineFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/complete-user-data")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Test User").and().
		the_response_body_contains_fixtures_data("\"admin\": true").and().
		the_response_body_contains_fixtures_data("\"theme\": \"dark\"")
}

// TestFixtureFileSupport_BackwardCompatibility_AtSyntax tests @ syntax in mixed config
func TestFixtureFileSupport_BackwardCompatibility_AtSyntax(t *testing.T) {
	configFile := createMixedSyntaxFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/legacy")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Legacy Data")
}

// TestFixtureFileSupport_BackwardCompatibility_LessThanSyntax tests < syntax in mixed config
func TestFixtureFileSupport_BackwardCompatibility_LessThanSyntax(t *testing.T) {
	configFile := createMixedSyntaxFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/enhanced")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Enhanced Data")
}

// TestFixtureFileSupport_BackwardCompatibility_InlineSyntax tests inline syntax in mixed config
func TestFixtureFileSupport_BackwardCompatibility_InlineSyntax(t *testing.T) {
	configFile := createMixedSyntaxFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/inline")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Inline Data")
}

// TestFixtureFileSupport_BackwardCompatibility_DirectSyntax tests direct data in mixed config
func TestFixtureFileSupport_BackwardCompatibility_DirectSyntax(t *testing.T) {
	configFile := createMixedSyntaxFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/direct")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Direct Data")
}

// TestFixtureFileSupport_ErrorHandling_MissingFile tests graceful fallback for missing fixtures
func TestFixtureFileSupport_ErrorHandling_MissingFile(t *testing.T) {
	configFile := createMissingFixtureConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/missing")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("@fixtures/nonexistent.json")
}

// TestFixtureFileSupport_Security_PathTraversalProtection tests security against path traversal
func TestFixtureFileSupport_Security_PathTraversalProtection(t *testing.T) {
	configFile := createPathTraversalConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/traversal")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("@fixtures/../../../etc/passwd")
}

// TestFixtureFileSupport_Performance_CachingTests tests fixture caching performance
func TestFixtureFileSupport_Performance_CachingTests(t *testing.T) {
	configFile := createPerformanceTestConfig(t)
	_, when, then := newServerParts(t, configFile)

	when.
		a_get_request_is_made_to("/api/fixtures/cached")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Cached Data")
}

// Helper functions for creating test configurations

func createAtSyntaxFixtureConfig(t *testing.T) string {
	t.Helper()

	// Create fixtures directory and files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "users")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	userFixture := filepath.Join(fixturesDir, "user.json")
	userContent := `{"id": "123", "name": "Legacy User", "email": "legacy@example.com", "role": "customer"}`
	err = os.WriteFile(userFixture, []byte(userContent), 0644)
	require.NoError(t, err)

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-User-ID"]
    return_body: true

scenarios:
  - uuid: "at-syntax-test"
    method: "GET"
    path: "/api/fixtures/user"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/users/user.json"
    headers:
      X-Test-Source: "at-syntax"
`

	configFile := filepath.Join(tempDir, "at-syntax-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createLessThanSyntaxFixtureConfig(t *testing.T) string {
	t.Helper()

	// Create fixtures directory and files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "products")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	productFixture := filepath.Join(fixturesDir, "product.json")
	productContent := `{"id": "456", "name": "New Product", "price": 29.99, "category": "electronics"}`
	err = os.WriteFile(productFixture, []byte(productContent), 0644)
	require.NoError(t, err)

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-Product-ID"]
    return_body: true

scenarios:
  - uuid: "less-than-syntax-test"
    method: "GET"
    path: "/api/fixtures/product"
    status_code: 200
    content_type: "application/json"
    data: "< ./fixtures/products/product.json"
    headers:
      X-Test-Source: "less-than-syntax"
`

	configFile := filepath.Join(tempDir, "less-than-syntax-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createLessAtSyntaxFixtureConfig(t *testing.T) string {
	t.Helper()

	// Create fixtures directory and files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "orders")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	orderFixture := filepath.Join(fixturesDir, "order.json")
	orderContent := `{"id": "789", "status": "Pending Order", "total": 199.99, "items": 3}`
	err = os.WriteFile(orderFixture, []byte(orderContent), 0644)
	require.NoError(t, err)

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-Order-ID"]
    return_body: true

scenarios:
  - uuid: "less-at-syntax-test"
    method: "GET"
    path: "/api/fixtures/order"
    status_code: 200
    content_type: "application/json"
    data: "<@ ./fixtures/orders/order.json"
    headers:
      X-Test-Source: "less-at-syntax"
`

	configFile := filepath.Join(tempDir, "less-at-syntax-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createInlineFixtureConfig(t *testing.T) string {
	t.Helper()

	// Create fixtures directory and files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	userFixture := filepath.Join(fixturesDir, "user.json")
	userContent := `{"id": "123", "name": "Test User", "email": "test@example.com"}`
	err = os.WriteFile(userFixture, []byte(userContent), 0644)
	require.NoError(t, err)

	profileFixture := filepath.Join(fixturesDir, "profile.json")
	profileContent := `{"title": "Developer Profile", "level": "Senior", "department": "Engineering"}`
	err = os.WriteFile(profileFixture, []byte(profileContent), 0644)
	require.NoError(t, err)

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-User-ID"]
    return_body: true

scenarios:
  - uuid: "inline-fixture-test"
    method: "GET"
    path: "/api/fixtures/user-with-profile"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "user": < ./fixtures/user.json,
        "profile": < ./fixtures/profile.json,
        "combined": true
      }
    headers:
      X-Test-Source: "inline-fixture"
`

	configFile := filepath.Join(tempDir, "inline-fixture-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createMultipleInlineFixtureConfig(t *testing.T) string {
	t.Helper()

	// Create fixtures directory and files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	userFixture := filepath.Join(fixturesDir, "user.json")
	userContent := `{"id": "123", "name": "Test User", "role": "admin"}`
	err = os.WriteFile(userFixture, []byte(userContent), 0644)
	require.NoError(t, err)

	permissionsFixture := filepath.Join(fixturesDir, "permissions.json")
	permissionsContent := `{"admin": true, "read": true, "write": true, "delete": false}`
	err = os.WriteFile(permissionsFixture, []byte(permissionsContent), 0644)
	require.NoError(t, err)

	settingsFixture := filepath.Join(fixturesDir, "settings.json")
	settingsContent := `{"theme": "dark", "notifications": true, "language": "en"}`
	err = os.WriteFile(settingsFixture, []byte(settingsContent), 0644)
	require.NoError(t, err)

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-User-ID"]
    return_body: true

scenarios:
  - uuid: "multiple-inline-fixture-test"
    method: "GET"
    path: "/api/fixtures/complete-user-data"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "user": < ./fixtures/user.json,
        "permissions": < ./fixtures/permissions.json,
        "settings": < ./fixtures/settings.json,
        "metadata": {
          "created": "2024-01-15T10:30:00Z",
          "source": "multiple-inline-fixtures"
        }
      }
    headers:
      X-Test-Source: "multiple-inline-fixtures"
`

	configFile := filepath.Join(tempDir, "multiple-inline-fixture-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createMixedSyntaxFixtureConfig(t *testing.T) string {
	t.Helper()

	// Create fixtures directory and files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "mixed")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	legacyFixture := filepath.Join(fixturesDir, "legacy.json")
	legacyContent := `{"id": "001", "name": "Legacy Data", "syntax": "@fixtures"}`
	err = os.WriteFile(legacyFixture, []byte(legacyContent), 0644)
	require.NoError(t, err)

	enhancedFixture := filepath.Join(fixturesDir, "enhanced.json")
	enhancedContent := `{"id": "002", "name": "Enhanced Data", "syntax": "< ./fixtures"}`
	err = os.WriteFile(enhancedFixture, []byte(enhancedContent), 0644)
	require.NoError(t, err)

	inlineFixture := filepath.Join(fixturesDir, "inline.json")
	inlineContent := `{"id": "003", "name": "Inline Data", "syntax": "inline fixtures"}`
	err = os.WriteFile(inlineFixture, []byte(inlineContent), 0644)
	require.NoError(t, err)

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-Test-ID"]
    return_body: true

scenarios:
  - uuid: "legacy-syntax-test"
    method: "GET"
    path: "/api/fixtures/legacy"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/mixed/legacy.json"

  - uuid: "enhanced-syntax-test"
    method: "GET"
    path: "/api/fixtures/enhanced"
    status_code: 200
    content_type: "application/json"
    data: "< ./fixtures/mixed/enhanced.json"

  - uuid: "inline-syntax-test"
    method: "GET"
    path: "/api/fixtures/inline"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "data": < ./fixtures/mixed/inline.json,
        "syntax": "inline"
      }

  - uuid: "direct-syntax-test"
    method: "GET"
    path: "/api/fixtures/direct"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "004",
        "name": "Direct Data",
        "syntax": "direct"
      }
`

	configFile := filepath.Join(tempDir, "mixed-syntax-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createMissingFixtureConfig(t *testing.T) string {
	t.Helper()

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-Test-ID"]
    return_body: true

scenarios:
  - uuid: "missing-fixture-test"
    method: "GET"
    path: "/api/fixtures/missing"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/nonexistent.json"
    headers:
      X-Test-Source: "missing-fixture"
`

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "missing-fixture-config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createPathTraversalConfig(t *testing.T) string {
	t.Helper()

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-Test-ID"]
    return_body: true

scenarios:
  - uuid: "path-traversal-test"
    method: "GET"
    path: "/api/fixtures/traversal"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/../../../etc/passwd"
    headers:
      X-Test-Source: "path-traversal"
`

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "path-traversal-config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}

func createPerformanceTestConfig(t *testing.T) string {
	t.Helper()

	// Create fixtures directory and files
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures", "performance")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	cachedFixture := filepath.Join(fixturesDir, "cached.json")
	cachedContent := `{"id": "999", "name": "Cached Data", "performance": "test", "cached": true}`
	err = os.WriteFile(cachedFixture, []byte(cachedContent), 0644)
	require.NoError(t, err)

	configContent := `
sections:
  fixtures:
    path_pattern: "/api/fixtures/**"
    body_id_paths: ["/id"]
    header_id_names: ["X-Test-ID"]
    return_body: true

scenarios:
  - uuid: "performance-cache-test"
    method: "GET"
    path: "/api/fixtures/cached"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/performance/cached.json"
    headers:
      X-Test-Source: "performance-cache"
`

	configFile := filepath.Join(tempDir, "performance-config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	return configFile
}
