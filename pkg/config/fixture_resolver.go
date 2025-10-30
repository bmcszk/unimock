package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// FixtureResolver handles loading fixture files referenced in configuration
type FixtureResolver struct {
	baseDir string
	cache   map[string]string
	mutex   sync.RWMutex
	// Precompiled regex for inline fixture references
	inlineFixtureRegex *regexp.Regexp
}

// NewFixtureResolver creates a new fixture resolver with the given base directory
func NewFixtureResolver(baseDir string) *FixtureResolver {
	// Compile regex to match fixture references within content: < ./path/to/file or <@ ./path/to/file
	inlineFixtureRegex := regexp.MustCompile(`<\s*@?\s*([^\s,}]+)`)

	return &FixtureResolver{
		baseDir:            baseDir,
		cache:              make(map[string]string),
		inlineFixtureRegex: inlineFixtureRegex,
	}
}

// ResolveFixture resolves a data string, supporting multiple fixture reference formats:
// - @fixtures/file.json syntax (backward compatibility)
// - < ./fixtures/file.ext syntax (go-restclient compatible)
// - <@ ./fixtures/file.ext syntax (future variable substitution)
// - Inline fixtures: {"key": < ./fixtures/file.json} syntax within body content
// - Inline data: Returns data as-is when no fixture references are found
func (fr *FixtureResolver) ResolveFixture(data string) (string, error) {
	trimmedData := strings.TrimSpace(data)

	// Handle @ syntax (backward compatibility)
	if strings.HasPrefix(trimmedData, "@") {
		return fr.resolveAtSyntax(trimmedData)
	}

	// Handle < and <@ syntax (go-restclient compatible) - direct replacement
	if strings.HasPrefix(trimmedData, "<") && !strings.Contains(trimmedData, "}") {
		return fr.resolveLessThanSyntax(trimmedData)
	}

	// Handle inline fixture references within body content
	return fr.resolveInlineFixtures(data), nil
}

// resolveAtSyntax handles @fixtures/file.json syntax
func (fr *FixtureResolver) resolveAtSyntax(data string) (string, error) {
	// Remove @ prefix to get file path
	filePath := strings.TrimSpace(data[1:])

	// Validate file path for security
	if err := fr.validateFilePath(filePath); err != nil {
		return "", err
	}

	return fr.loadFixtureFile(filePath)
}

// resolveLessThanSyntax handles < ./fixtures/file.ext and <@ ./fixtures/file.ext syntax
func (fr *FixtureResolver) resolveLessThanSyntax(data string) (string, error) {
	// Remove < prefix
	content := strings.TrimSpace(data[1:])

	// Remove optional @ prefix
	if strings.HasPrefix(content, "@") {
		content = strings.TrimSpace(content[1:])
	}

	// The remaining content is the file path
	filePath := strings.TrimSpace(content)

	// Validate file path for security
	if err := fr.validateFilePath(filePath); err != nil {
		return "", err
	}

	return fr.loadFixtureFile(filePath)
}

// loadFixtureFile loads a fixture file with caching
func (fr *FixtureResolver) loadFixtureFile(filePath string) (string, error) {
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

// resolveInlineFixtures handles fixture references within content (e.g., { "user": < ./fixtures/user.json })
func (fr *FixtureResolver) resolveInlineFixtures(data string) string {
	// Use regex to find all fixture references within the content
	return fr.inlineFixtureRegex.ReplaceAllStringFunc(data, fr.processInlineMatch)
}

// processInlineMatch processes a single regex match for inline fixture references
func (fr *FixtureResolver) processInlineMatch(match string) string {
	// Extract the file path from the match (remove "<" and optional "@")
	capture := fr.inlineFixtureRegex.FindStringSubmatch(match)
	if len(capture) < 2 {
		return match // Return original if regex didn't capture properly
	}

	filePath := strings.TrimSpace(capture[1])

	// Validate file path for security
	if err := fr.validateFilePath(filePath); err != nil {
		return match // Return original reference if path is invalid
	}

	// Try to load the fixture content
	if content, err := fr.loadFixtureFile(filePath); err == nil {
		return content
	}

	return match // Return original reference if loading fails
}

// validateFilePath ensures the file path is safe and doesn't allow path traversal
func (fr *FixtureResolver) validateFilePath(filePath string) error {
	if filePath == "" {
		return &InvalidFixturePathError{Path: filePath, Reason: "empty path"}
	}

	if err := fr.checkAbsolutePath(filePath); err != nil {
		return err
	}

	if err := fr.checkPathTraversal(filePath); err != nil {
		return err
	}

	return fr.checkCleanPath(filePath)
}

// checkAbsolutePath validates against absolute paths
func (*FixtureResolver) checkAbsolutePath(filePath string) error {
	// Check for absolute paths (filepath.IsAbs handles both Unix and Windows)
	if filepath.IsAbs(filePath) {
		return &InvalidFixturePathError{Path: filePath, Reason: "absolute path not allowed"}
	}

	// Additional checks for patterns that filepath.IsAbs might miss in cross-platform scenarios
	if strings.HasPrefix(filePath, "/") ||
		(len(filePath) >= 2 && filePath[1] == ':') || // Windows drive letter (e.g., C:)
		strings.HasPrefix(filePath, "\\") {
		return &InvalidFixturePathError{Path: filePath, Reason: "absolute path not allowed"}
	}

	return nil
}

// checkPathTraversal validates against path traversal sequences
func (*FixtureResolver) checkPathTraversal(filePath string) error {
	// Check for path traversal sequences (both Unix and Windows style)
	if !strings.Contains(filePath, "..") {
		return nil
	}

	// Check both forward and backward slash patterns
	if strings.Contains(filePath, "../") || strings.Contains(filePath, "..\\") ||
		strings.HasPrefix(filePath, "../") || strings.HasPrefix(filePath, "..\\") {
		return &InvalidFixturePathError{Path: filePath, Reason: "path traversal not allowed"}
	}
	return nil
}

// checkCleanPath validates cleaned path for traversal
func (*FixtureResolver) checkCleanPath(filePath string) error {
	// Clean the path and do a final check
	cleanPath := filepath.Clean(filePath)
	if strings.HasPrefix(cleanPath, "..") {
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
