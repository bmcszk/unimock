package config

import (
	"bytes"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// WildcardChar represents the single segment wildcard character used in path patterns
	WildcardChar = "*"
	// RecursiveWildcard represents the recursive wildcard for multiple segments
	RecursiveWildcard = "**"
	// PathSeparator represents the separator used in URL paths
	PathSeparator = "/"
	// noMatch represents an invalid match score
	noMatch = -1
)

// MockConfig represents the configuration for mock behavior
// It defines how Unimock handles different API endpoints and extracts IDs
// from various parts of HTTP requests.
type MockConfig struct {
	// Sections contains configuration for different API endpoint patterns
	// The map keys are section names (usually API resource names like "users" or "orders")
	// and the values are Section structs defining how to handle requests to those endpoints.
	Sections map[string]Section `yaml:",inline" json:"sections"`
}

// Section represents a configuration section for a specific API endpoint pattern
type Section struct {
	// PathPattern defines the URL pattern to match against.
	// Use * as a wildcard for single path segments, e.g. "/users/*" or "/users/*/orders/*"
	// Use ** as a wildcard for multiple path segments recursively, e.g. "/api/**"
	PathPattern string `yaml:"path_pattern" json:"path_pattern"`

	// StrictPath determines whether GET/PUT/DELETE operations require exact path pattern matching.
	// When true:
	//   - GET/PUT/DELETE must match the exact path pattern (stricter path validation)
	//   - PUT returns 404 if resource doesn't exist (no upsert behavior)
	//   - More restrictive path matching rules
	// When false (default):
	//   - More flexible path pattern matching (e.g., /users/* matches both /users/123 and /users)
	//   - PUT performs upsert operations (creates if doesn't exist)
	//   - Less restrictive path matching rules
	// Note: Individual resource requests (e.g., GET /users/123) always return 404 if resource doesn't exist
	StrictPath bool `yaml:"strict_path" json:"strict_path"`

	// BodyIDPaths defines the XPath-like paths to extract IDs from request bodies.
	// For JSON:
	//   - Use "/" to start from root
	//   - Use element names to navigate
	//   - Use "//" to search anywhere
	//   - Use "*" as wildcard
	//   - Use "text()" to get text content
	// Examples:
	//   - "/id" - extracts ID from root object
	//   - "/data/id" - extracts ID from nested object
	//   - "//id" - extracts any ID element anywhere
	//   - "/items/*/id" - extracts IDs from array of objects
	//   - "/user/id" - extracts ID from specific object
	//   - "//id[text()='123']" - extracts ID with specific value
	//
	// For XML:
	//   - Use "/" to start from root
	//   - Use element names to navigate
	//   - Use "//" to search anywhere
	//   - Use "*" as wildcard
	//   - Use "text()" to get text content
	// Examples:
	//   - "/root/id" - extracts ID from root element
	//   - "//id" - extracts any ID element
	//   - "/root/items/item/id" - extracts IDs from nested elements
	//   - "/root/*/id" - extracts IDs from any direct child
	//   - "//id[text()='123']" - extracts ID with specific value
	BodyIDPaths []string `yaml:"body_id_paths" json:"body_id_paths"`

	// HeaderIDName specifies the HTTP header name to extract IDs from.
	// If empty, no header-based ID extraction will be performed.
	HeaderIDName string `yaml:"header_id_name" json:"header_id_name"`

	// CaseSensitive determines whether path matching is case-sensitive.
	// If true, paths must match exactly including case.
	// If false, paths are matched case-insensitively.
	CaseSensitive bool `yaml:"case_sensitive" json:"case_sensitive"`

	// Transformations contains request/response transformation functions.
	// This field is only available when using Unimock as a library and is excluded from YAML serialization.
	// It allows programmatic modification of requests and responses for advanced testing scenarios.
	Transformations *TransformationConfig `yaml:"-" json:"-"`
}

// NewMockConfig creates an empty MockConfig with an initialized Sections map
func NewMockConfig() *MockConfig {
	return &MockConfig{
		Sections: make(map[string]Section),
	}
}

// LoadFromYAML loads a MockConfig from a YAML file at the given path
func LoadFromYAML(path string) (*MockConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := NewMockConfig()

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true) // Enable strict mode

	if err := decoder.Decode(config.Sections); err != nil {
		return nil, err
	}

	return config, nil
}

// isPatternMatch checks if a path matches a pattern with wildcards
func isPatternMatch(pattern, path string, caseSensitive bool) bool {
	matcher := pathMatcher{caseSensitive: caseSensitive}
	patternParts := strings.Split(strings.Trim(pattern, PathSeparator), PathSeparator)
	pathParts := strings.Split(strings.Trim(path, PathSeparator), PathSeparator)

	if !strings.Contains(pattern, WildcardChar) {
		return matcher.matchExactPath(pattern, path)
	}

	// Check for recursive wildcard patterns
	if strings.Contains(pattern, RecursiveWildcard) {
		return matcher.matchRecursivePattern(patternParts, pathParts)
	}

	return matcher.matchWildcardPattern(patternParts, pathParts)
}

// pathMatcher handles path matching with configurable case sensitivity
type pathMatcher struct {
	caseSensitive bool
}

// matchExactPath performs exact path matching
func (pm pathMatcher) matchExactPath(pattern, path string) bool {
	if pm.caseSensitive {
		return pattern == path
	}
	return strings.EqualFold(pattern, path)
}

// matchWildcardPattern performs wildcard pattern matching
func (pm pathMatcher) matchWildcardPattern(patternParts, pathParts []string) bool {
	if !isValidSegmentCount(patternParts, pathParts) {
		return false
	}

	return pm.matchSegments(patternParts, pathParts)
}

// isValidSegmentCount checks if segment counts are compatible for single wildcards
func isValidSegmentCount(patternParts, pathParts []string) bool {
	// For single wildcard patterns, the segment count must match exactly
	// OR for collection access, allow one less segment (e.g., /users/* matches /users)
	return len(patternParts) == len(pathParts) ||
		(len(patternParts) > 0 && len(pathParts) == len(patternParts)-1 &&
			patternParts[len(patternParts)-1] == WildcardChar)
}

// matchSegments compares pattern segments with path segments
func (pm pathMatcher) matchSegments(patternParts, pathParts []string) bool {
	// Handle collection access case: /users/* matches /users
	if pm.isCollectionAccess(patternParts, pathParts) {
		return pm.matchCollectionSegments(patternParts, pathParts)
	}

	return pm.matchNormalSegments(patternParts, pathParts)
}

// isCollectionAccess checks if this is a collection access pattern
func (*pathMatcher) isCollectionAccess(patternParts, pathParts []string) bool {
	return len(pathParts) == len(patternParts)-1 && len(patternParts) > 0 &&
		patternParts[len(patternParts)-1] == WildcardChar
}

// matchCollectionSegments matches collection access patterns
func (pm pathMatcher) matchCollectionSegments(patternParts, pathParts []string) bool {
	for i := 0; i < len(pathParts); i++ {
		if !pm.segmentMatches(patternParts[i], pathParts[i]) {
			return false
		}
	}
	return true
}

// matchNormalSegments matches normal patterns with exact segment counts
func (pm pathMatcher) matchNormalSegments(patternParts, pathParts []string) bool {
	maxLen := len(patternParts)
	if len(pathParts) < maxLen {
		maxLen = len(pathParts)
	}

	for i := 0; i < maxLen; i++ {
		if patternParts[i] == WildcardChar {
			continue
		}
		if !pm.segmentMatches(patternParts[i], pathParts[i]) {
			return false
		}
	}
	return true
}

// segmentMatches checks if a single segment matches
func (pm pathMatcher) segmentMatches(pattern, path string) bool {
	if pm.caseSensitive {
		return pattern == path
	}
	return strings.EqualFold(pattern, path)
}

// matchRecursivePattern handles patterns with ** recursive wildcards
func (pm pathMatcher) matchRecursivePattern(patternParts, pathParts []string) bool {
	return pm.matchRecursiveSegments(patternParts, pathParts, 0, 0)
}

// matchRecursiveSegments recursively matches pattern segments with path segments
func (pm pathMatcher) matchRecursiveSegments(patternParts, pathParts []string, patternIdx, pathIdx int) bool {
	// Check if all patterns consumed
	if patternIdx >= len(patternParts) {
		return pathIdx >= len(pathParts)
	}
	
	// Check if all paths consumed but patterns remain
	if pathIdx >= len(pathParts) {
		return allRemainingAreRecursiveWildcards(patternParts, patternIdx)
	}
	
	currentPattern := patternParts[patternIdx]
	
	switch currentPattern {
	case RecursiveWildcard:
		return pm.handleRecursiveWildcard(patternParts, pathParts, patternIdx, pathIdx)
	case WildcardChar:
		return pm.handleSingleWildcard(patternParts, pathParts, patternIdx, pathIdx)
	default:
		return pm.handleExactMatch(patternParts, pathParts, patternIdx, pathIdx, currentPattern)
	}
}

// allRemainingAreRecursiveWildcards checks if remaining pattern parts are all ** wildcards
func allRemainingAreRecursiveWildcards(patternParts []string, patternIdx int) bool {
	for i := patternIdx; i < len(patternParts); i++ {
		if patternParts[i] != RecursiveWildcard {
			return false
		}
	}
	return true
}

// handleRecursiveWildcard processes ** wildcards
func (pm pathMatcher) handleRecursiveWildcard(patternParts, pathParts []string, patternIdx, pathIdx int) bool {
	// ** can match zero or more segments
	for i := pathIdx; i <= len(pathParts); i++ {
		if pm.matchRecursiveSegments(patternParts, pathParts, patternIdx+1, i) {
			return true
		}
	}
	return false
}

// handleSingleWildcard processes * wildcards
func (pm pathMatcher) handleSingleWildcard(patternParts, pathParts []string, patternIdx, pathIdx int) bool {
	// * matches exactly one segment
	return pm.matchRecursiveSegments(patternParts, pathParts, patternIdx+1, pathIdx+1)
}

// handleExactMatch processes exact segment matches
func (pm pathMatcher) handleExactMatch(
	patternParts, pathParts []string, patternIdx, pathIdx int, currentPattern string,
) bool {
	if pm.segmentMatches(currentPattern, pathParts[pathIdx]) {
		return pm.matchRecursiveSegments(patternParts, pathParts, patternIdx+1, pathIdx+1)
	}
	return false
}

// MatchPath finds the section that matches the given path
func (c *MockConfig) MatchPath(path string) (string, *Section, error) {
	normalizedPath := strings.Trim(path, PathSeparator)

	// First try exact matches (no wildcards)
	if name, section := c.findExactMatch(normalizedPath); section != nil {
		return name, section, nil
	}

	// Then try wildcard matches, prioritizing longer patterns
	if name, section := c.findBestWildcardMatch(normalizedPath); section != nil {
		return name, section, nil
	}

	return "", nil, nil // No match found
}

// findExactMatch looks for exact pattern matches (no wildcards)
func (c *MockConfig) findExactMatch(normalizedPath string) (string, *Section) {
	for name, section := range c.Sections {
		pattern := strings.Trim(section.PathPattern, PathSeparator)
		if !strings.Contains(pattern, WildcardChar) {
			if isPatternMatch(pattern, normalizedPath, section.CaseSensitive) {
				s := section // Create a local copy
				return name, &s
			}
		}
	}
	return "", nil
}

// findBestWildcardMatch finds the best wildcard match by prioritizing longer patterns
func (c *MockConfig) findBestWildcardMatch(normalizedPath string) (string, *Section) {
	bestMatch := wildcardMatch{name: "", numSegments: noMatch}

	for name, section := range c.Sections {
		if match := c.evaluateWildcardSection(name, section, normalizedPath); match.isValid() {
			if match.isBetterThan(bestMatch) {
				bestMatch = match
			}
		}
	}

	return bestMatch.getResult(c)
}

// wildcardMatch represents a potential wildcard match
type wildcardMatch struct {
	name        string
	numSegments int
}

// isValid checks if the match is valid
func (m wildcardMatch) isValid() bool {
	return m.name != ""
}

// isBetterThan checks if this match is better than another
func (m wildcardMatch) isBetterThan(other wildcardMatch) bool {
	return m.numSegments > other.numSegments
}

// getResult returns the section for this match
func (m wildcardMatch) getResult(c *MockConfig) (string, *Section) {
	if !m.isValid() {
		return "", nil
	}
	matchedSection := c.Sections[m.name]
	return m.name, &matchedSection
}

// evaluateWildcardSection checks if a section matches and returns match info
func (*MockConfig) evaluateWildcardSection(name string, section Section, normalizedPath string) wildcardMatch {
	pattern := strings.Trim(section.PathPattern, PathSeparator)
	
	if !strings.Contains(pattern, WildcardChar) {
		return wildcardMatch{}
	}

	if !isPatternMatch(pattern, normalizedPath, section.CaseSensitive) {
		return wildcardMatch{}
	}

	// Calculate match score: prefer patterns with more specific segments
	// ** wildcards get lower priority than specific segments or *
	numSegments := len(strings.Split(pattern, PathSeparator))
	
	// Adjust score based on wildcard types
	score := numSegments * 100 // Base score
	patternParts := strings.Split(pattern, PathSeparator)
	for _, part := range patternParts {
		switch part {
		case RecursiveWildcard:
			score -= 50 // ** wildcards are less specific
		case WildcardChar:
			score -= 10 // * wildcards are somewhat less specific
		default:
			// Exact segments don't modify the score (most specific)
		}
	}
	
	return wildcardMatch{name: name, numSegments: score}
}
