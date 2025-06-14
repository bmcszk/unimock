package config

import (
	"bytes"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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
	// Use * as a wildcard for ID segments, e.g. "/users/*" or "/users/*/orders/*"
	PathPattern string `yaml:"path_pattern" json:"path_pattern"`

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
	// Split pattern and path into segments
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Check for exact static path match
	if !strings.Contains(pattern, "*") {
		if caseSensitive {
			return pattern == path
		}
		return strings.EqualFold(pattern, path)
	}

	// For wildcard patterns, check each segment
	if strings.Contains(pattern, "*") {
		// Different number of segments means no match, unless pattern ends with *
		if len(patternParts) != len(pathParts) &&
			(len(patternParts) == 0 || patternParts[len(patternParts)-1] != "*" ||
				len(pathParts) < len(patternParts)-1) {
			return false
		}

		// Check each segment
		for i := 0; i < len(patternParts) && i < len(pathParts); i++ {
			// Wildcard matches anything
			if patternParts[i] == "*" {
				continue
			}
			// Exact match required for non-wildcard segments
			if caseSensitive {
				if patternParts[i] != pathParts[i] {
					return false
				}
			} else {
				if !strings.EqualFold(patternParts[i], pathParts[i]) {
					return false
				}
			}
		}
	}

	return true
}

// MatchPath finds the section that matches the given path
func (c *MockConfig) MatchPath(path string) (string, *Section, error) {
	// Remove leading and trailing slashes for consistent matching
	normalizedPath := strings.Trim(path, "/")

	// First prioritize exact matches (patterns without wildcards)
	for name, section := range c.Sections {
		pattern := strings.Trim(section.PathPattern, "/")
		if !strings.Contains(pattern, "*") { // Only consider non-wildcard patterns here
			if isPatternMatch(pattern, normalizedPath, section.CaseSensitive) {
				// Must return a pointer to a copy or a re-fetched section, not the loop variable's address
				s := section // Create a local copy
				return name, &s, nil
			}
		}
	}

	// Then try to find wildcard matches, prioritizing longer patterns first
	var bestMatchName string
	var bestMatchNumSegments = -1 // Use -1 to indicate no match found yet

	for name, section := range c.Sections {
		pattern := strings.Trim(section.PathPattern, "/")
		if strings.Contains(pattern, "*") { // Only consider wildcard patterns here
			if isPatternMatch(pattern, normalizedPath, section.CaseSensitive) {
				numSegments := len(strings.Split(pattern, "/"))
				if numSegments > bestMatchNumSegments {
					bestMatchNumSegments = numSegments
					bestMatchName = name
				}
			}
		}
	}

	if bestMatchName != "" {
		// Retrieve the actual section using the determined bestMatchName to avoid pointer issues with loop variable
		matchedSection := c.Sections[bestMatchName]
		return bestMatchName, &matchedSection, nil
	}

	return "", nil, nil // No match found
}
