package config_test

import (
	"os"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

func TestUniConfig_WildcardPatternMatching(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		path        string
		shouldMatch bool
		description string
	}{
		// Single wildcard tests
		{
			name:        "single wildcard matches one segment",
			pattern:     "/users/*",
			path:        "/users/123",
			shouldMatch: true,
			description: "* should match single path segment",
		},
		{
			name:        "single wildcard matches collection access",
			pattern:     "/users/*",
			path:        "/users",
			shouldMatch: true,
			description: "* should match collection path (backward compatibility)",
		},
		{
			name:        "single wildcard does not match multiple segments",
			pattern:     "/users/*",
			path:        "/users/123/posts",
			shouldMatch: false,
			description: "* should not match multiple path segments",
		},
		{
			name:        "multiple single wildcards",
			pattern:     "/users/*/posts/*",
			path:        "/users/123/posts/456",
			shouldMatch: true,
			description: "multiple * should match their respective segments",
		},
		{
			name:        "multiple single wildcards missing segment",
			pattern:     "/users/*/posts/*",
			path:        "/users/123/posts",
			shouldMatch: false,
			description: "multiple * should require all segments",
		},

		// Recursive wildcard tests
		{
			name:        "recursive wildcard matches multiple segments",
			pattern:     "/api/**",
			path:        "/api/users/123/posts/456",
			shouldMatch: true,
			description: "** should match multiple path segments recursively",
		},
		{
			name:        "recursive wildcard matches zero segments",
			pattern:     "/api/**",
			path:        "/api",
			shouldMatch: true,
			description: "** should match zero segments",
		},
		{
			name:        "recursive wildcard matches one segment",
			pattern:     "/api/**",
			path:        "/api/users",
			shouldMatch: true,
			description: "** should match one segment",
		},
		{
			name:        "recursive wildcard with suffix",
			pattern:     "/api/**/health",
			path:        "/api/v1/admin/health",
			shouldMatch: true,
			description: "** should match multiple segments before suffix",
		},
		{
			name:        "recursive wildcard with suffix no match",
			pattern:     "/api/**/health",
			path:        "/api/v1/admin/status",
			shouldMatch: false,
			description: "** with suffix should require exact suffix match",
		},

		// Mixed wildcard tests
		{
			name:        "mixed single and recursive wildcards",
			pattern:     "/users/*/posts/**",
			path:        "/users/123/posts/456/comments/789",
			shouldMatch: true,
			description: "* and ** should work together",
		},
		{
			name:        "mixed wildcards with exact segments",
			pattern:     "/api/*/v1/**/users",
			path:        "/api/admin/v1/internal/users",
			shouldMatch: true,
			description: "mix of exact, *, and ** should work",
		},

		// Edge cases
		{
			name:        "pattern longer than path",
			pattern:     "/users/*/posts/*",
			path:        "/users/123",
			shouldMatch: false,
			description: "pattern longer than path should not match",
		},
		{
			name:        "empty path segments",
			pattern:     "/api/**",
			path:        "/api/",
			shouldMatch: true,
			description: "trailing slash should be handled",
		},
	}

	cfg := &config.UniConfig{
		Sections: make(map[string]config.Section),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWildcardPattern(t, cfg, tt)
		})
	}
}

// testWildcardPattern helper function to test pattern matching
func testWildcardPattern(t *testing.T, cfg *config.UniConfig, tt struct {
	name        string
	pattern     string
	path        string
	shouldMatch bool
	description string
}) {
	t.Helper()

	// Create a test section with the pattern
	section := config.Section{
		PathPattern:   tt.pattern,
		CaseSensitive: true,
	}
	cfg.Sections["test"] = section

	// Test the pattern matching
	sectionName, matchedSection, err := cfg.MatchPath(tt.path)

	if tt.shouldMatch {
		validateSuccessfulMatch(t, err, matchedSection, sectionName, tt.pattern, tt.path)
	} else {
		validateFailedMatch(t, matchedSection, tt.pattern, tt.path, sectionName)
	}
}

// validateSuccessfulMatch validates that a pattern match succeeded
func validateSuccessfulMatch(
	t *testing.T, err error, matchedSection *config.Section, sectionName, pattern, path string,
) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if matchedSection == nil {
		t.Errorf("expected to find matching section for pattern %s and path %s", pattern, path)
	}
	if sectionName != "test" {
		t.Errorf("expected section name 'test', got %s", sectionName)
	}
}

// validateFailedMatch validates that a pattern match failed as expected
func validateFailedMatch(t *testing.T, matchedSection *config.Section, pattern, path, sectionName string) {
	t.Helper()
	if matchedSection != nil {
		t.Errorf("expected no match for pattern %s and path %s, but got section %s",
			pattern, path, sectionName)
	}
}

func TestUniConfig_PatternPriority(t *testing.T) {
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"exact": {
				PathPattern:   "/api/users/123",
				CaseSensitive: true,
			},
			"single_wildcard": {
				PathPattern:   "/api/users/*",
				CaseSensitive: true,
			},
			"recursive_wildcard": {
				PathPattern:   "/api/**",
				CaseSensitive: true,
			},
			"mixed_wildcard": {
				PathPattern:   "/api/*/123",
				CaseSensitive: true,
			},
		},
	}

	tests := []struct {
		name            string
		path            string
		expectedSection string
		description     string
	}{
		{
			name:            "exact match should have highest priority",
			path:            "/api/users/123",
			expectedSection: "exact",
			description:     "exact patterns should be preferred over wildcards",
		},
		{
			name:            "specific wildcard over recursive",
			path:            "/api/users/456",
			expectedSection: "single_wildcard",
			description:     "* should be preferred over **",
		},
		{
			name:            "mixed wildcard should win over recursive",
			path:            "/api/posts/123",
			expectedSection: "mixed_wildcard",
			description:     "specific * patterns should be preferred over **",
		},
		{
			name:            "recursive wildcard as fallback",
			path:            "/api/v1/admin/health",
			expectedSection: "recursive_wildcard",
			description:     "** should match when no more specific pattern exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPatternPriority(t, cfg, tt)
		})
	}
}

// testPatternPriority helper function to test pattern priority
func testPatternPriority(t *testing.T, cfg *config.UniConfig, tt struct {
	name            string
	path            string
	expectedSection string
	description     string
}) {
	t.Helper()
	sectionName, matchedSection, err := cfg.MatchPath(tt.path)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if matchedSection == nil {
		t.Errorf("expected to find matching section for path %s", tt.path)
	}
	if sectionName != tt.expectedSection {
		t.Errorf("expected section %s, got %s for path %s", tt.expectedSection, sectionName, tt.path)
	}
}

func TestSection_StrictPathConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		yamlContent        string
		expectedStrictPath bool
		description        string
	}{
		{
			name: "strict_path true",
			yamlContent: `
test_section:
  path_pattern: "/users/*"
  strict_path: true
  body_id_paths: ["/id"]`,
			expectedStrictPath: true,
			description:        "strict_path should be parsed correctly when true",
		},
		{
			name: "strict_path false",
			yamlContent: `
test_section:
  path_pattern: "/users/*"
  strict_path: false
  body_id_paths: ["/id"]`,
			expectedStrictPath: false,
			description:        "strict_path should be parsed correctly when false",
		},
		{
			name: "strict_path omitted defaults to false",
			yamlContent: `
test_section:
  path_pattern: "/users/*"
  body_id_paths: ["/id"]`,
			expectedStrictPath: false,
			description:        "strict_path should default to false when omitted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStrictPathConfig(t, tt)
		})
	}
}

// testStrictPathConfig helper function to test strict path configuration
func testStrictPathConfig(t *testing.T, tt struct {
	name               string
	yamlContent        string
	expectedStrictPath bool
	description        string
}) {
	t.Helper()

	// Create a temporary file with YAML content
	tmpfile, err := createTempYAMLFile(tt.yamlContent)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer cleanupTempFile(tmpfile)

	// Load configuration from YAML
	cfg, err := config.LoadFromYAML(tmpfile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Check the strict_path value
	section, exists := cfg.Sections["test_section"]
	if !exists {
		t.Fatal("test_section not found in config")
	}

	if section.StrictPath != tt.expectedStrictPath {
		t.Errorf("expected StrictPath to be %v, got %v", tt.expectedStrictPath, section.StrictPath)
	}
}

// Helper functions for YAML testing
func createTempYAMLFile(content string) (string, error) {
	tmpfile, err := os.CreateTemp("", "config-test-*.yaml")
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		return "", err
	}

	return tmpfile.Name(), nil
}

func cleanupTempFile(filename string) {
	os.Remove(filename)
}
