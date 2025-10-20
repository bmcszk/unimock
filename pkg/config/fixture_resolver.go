package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FixtureResolver handles loading fixture files referenced in configuration
type FixtureResolver struct {
	baseDir string
	cache   map[string]string
	mutex   sync.RWMutex
}

// NewFixtureResolver creates a new fixture resolver with the given base directory
func NewFixtureResolver(baseDir string) *FixtureResolver {
	return &FixtureResolver{
		baseDir: baseDir,
		cache:   make(map[string]string),
	}
}

// ResolveFixture resolves a data string, either returning inline data or loading from file
// If the data starts with "@", it treats it as a file reference relative to baseDir
// Otherwise, it returns the data as-is (inline data)
func (fr *FixtureResolver) ResolveFixture(data string) (string, error) {
	// Check if this is a file reference
	if !strings.HasPrefix(data, "@") {
		// Return inline data as-is
		return data, nil
	}

	// Remove @ prefix to get file path
	filePath := data[1:]

	// Validate file path for security
	if err := fr.validateFilePath(filePath); err != nil {
		return "", err
	}

	// Check cache first
	fr.mutex.RLock()
	if cached, exists := fr.cache[filePath]; exists {
		fr.mutex.RUnlock()
		return cached, nil
	}
	fr.mutex.RUnlock()

	// Resolve full path relative to base directory
	fullPath := filepath.Join(fr.baseDir, filePath)

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	// Cache the result
	contentStr := string(content)
	fr.mutex.Lock()
	fr.cache[filePath] = contentStr
	fr.mutex.Unlock()

	return contentStr, nil
}

// validateFilePath ensures the file path is safe and doesn't allow path traversal
func (*FixtureResolver) validateFilePath(filePath string) error {
	if filePath == "" {
		return &InvalidFixturePathError{Path: filePath, Reason: "empty path"}
	}

	// Check for absolute paths
	if filepath.IsAbs(filePath) {
		return &InvalidFixturePathError{Path: filePath, Reason: "absolute path not allowed"}
	}

	// Clean the path and check for path traversal attempts
	cleanPath := filepath.Clean(filePath)
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "../") {
		return &InvalidFixturePathError{Path: filePath, Reason: "path traversal not allowed"}
	}

	return nil
}

// ClearCache clears the internal file cache
func (fr *FixtureResolver) ClearCache() {
	fr.mutex.Lock()
	defer fr.mutex.Unlock()
	fr.cache = make(map[string]string)
}

// InvalidFixturePathError represents an error when a fixture path is invalid
type InvalidFixturePathError struct {
	Path   string
	Reason string
}

// Error returns the error message
func (e *InvalidFixturePathError) Error() string {
	return "invalid fixture path: " + e.Path + " (" + e.Reason + ")"
}
