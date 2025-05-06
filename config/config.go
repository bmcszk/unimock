package config

import (
	"bytes"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Sections map[string]Section `yaml:",inline"`
}

// Section represents a configuration section for a specific API endpoint pattern
type Section struct {
	// PathPattern defines the URL pattern to match against.
	// Use * as a wildcard for ID segments, e.g. "/users/*" or "/users/*/orders/*"
	PathPattern string `yaml:"path_pattern"`

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
	BodyIDPaths []string `yaml:"body_id_paths"`

	// HeaderIDName specifies the HTTP header name to extract IDs from.
	// If empty, no header-based ID extraction will be performed.
	HeaderIDName string `yaml:"header_id_name"`
}

// ConfigLoader defines the interface for loading configuration
type ConfigLoader interface {
	Load(path string) (*Config, error)
}

// YAMLConfigLoader implements ConfigLoader for YAML files
type YAMLConfigLoader struct{}

// Load implements ConfigLoader interface for YAML files
func (l *YAMLConfigLoader) Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Sections: make(map[string]Section),
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true) // Enable strict mode

	if err := decoder.Decode(config.Sections); err != nil {
		return nil, err
	}

	return config, nil
}

// MatchPath finds the first section that matches the given path
func (c *Config) MatchPath(path string) (string, *Section, error) {
	// Remove leading and trailing slashes for consistent matching
	path = strings.Trim(path, "/")

	// Try to find the most specific match first
	var bestMatch string
	var bestSection *Section
	var bestPatternLen int

	for name, section := range c.Sections {
		// Convert path pattern to match format
		pattern := strings.Trim(section.PathPattern, "/")

		// Split pattern and path into segments
		patternParts := strings.Split(pattern, "/")
		pathParts := strings.Split(path, "/")

		// Handle collection paths (no ID)
		if len(patternParts) > 0 && strings.HasSuffix(pattern, "*") {
			// Remove the * from pattern
			basePattern := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(path+"/", basePattern) {
				// Check if this is a more specific match
				if len(basePattern) > bestPatternLen {
					bestMatch = name
					bestSection = &section
					bestPatternLen = len(basePattern)
				}
				continue
			}
		}

		// Handle exact matches for collection paths
		if pattern == path {
			if len(pattern) > bestPatternLen {
				bestMatch = name
				bestSection = &section
				bestPatternLen = len(pattern)
			}
			continue
		}

		// Handle paths with ID
		if len(patternParts) == len(pathParts) {
			matched := true
			for i := 0; i < len(patternParts); i++ {
				if patternParts[i] == "*" {
					continue
				}
				if patternParts[i] != pathParts[i] {
					matched = false
					break
				}
			}
			if matched {
				if len(pattern) > bestPatternLen {
					bestMatch = name
					bestSection = &section
					bestPatternLen = len(pattern)
				}
			}
		}
	}

	if bestSection != nil {
		return bestMatch, bestSection, nil
	}

	return "", nil, nil
}
