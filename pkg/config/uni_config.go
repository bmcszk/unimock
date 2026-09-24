// Package config provides configuration structures for the Unimock server.
// It includes ServerConfig for server settings and UniConfig for mock behavior.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmcszk/unimock/pkg/model"
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

// UniConfig represents the configuration for mock behavior
// It defines how Unimock handles different API endpoints and extracts IDs
// from various parts of HTTP requests.
type UniConfig struct {
	// Sections contains configuration for different API endpoint patterns
	// The map keys are section names (usually API resource names like "users" or "orders")
	// and the values are Section structs defining how to handle requests to those endpoints.
	// Supports both inline format (legacy) and explicit sections field (unified)
	Sections map[string]Section `yaml:"sections,omitempty" json:"sections"`

	// Scenarios contains predefined responses that override normal mock behavior
	Scenarios []ScenarioConfig `yaml:"scenarios,omitempty" json:"scenarios,omitempty"`

	// baseDir is the directory containing the configuration file (for fixture resolution)
	baseDir string

	// fixtureResolver handles loading fixture files referenced in configuration
	fixtureResolver *FixtureResolver
}

// ScenarioConfig represents a scenario definition in configuration
// It mirrors the model.Scenario structure but allows for flexible YAML parsing
type ScenarioConfig struct {
	// UUID is the unique identifier for the scenario (optional, auto-generated if empty)
	UUID string `yaml:"uuid,omitempty" json:"uuid,omitempty"`

	// Method is the HTTP method for this scenario (e.g., "GET", "POST", etc.)
	Method string `yaml:"method" json:"method"`

	// Path is the URL path pattern this scenario should match
	Path string `yaml:"path" json:"path"`

	// StatusCode is the HTTP status code to return (default: 200)
	StatusCode int `yaml:"status_code,omitempty" json:"status_code,omitempty"`

	// ContentType is the MIME type for the response (default: "application/json")
	ContentType string `yaml:"content_type,omitempty" json:"content_type,omitempty"`

	// Location header for redirects or resource creation responses
	Location string `yaml:"location,omitempty" json:"location,omitempty"`

	// Data is the response body content
	Data string `yaml:"data,omitempty" json:"data,omitempty"`

	// Headers contains additional HTTP headers to include in the response
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// Webhook optionally fires an outbound HTTP request after this scenario matches.
	Webhook *WebhookConfig `yaml:"webhook,omitempty" json:"webhook,omitempty"`

	// Stream optionally emits a server-generated stream (SSE or NDJSON) instead of
	// writing Data. Mutually exclusive with Data at validation time.
	Stream *StreamConfig `yaml:"stream,omitempty" json:"stream,omitempty"`
}

// StreamConfig mirrors model.StreamConfig for YAML deserialization and validation.
// The definition lives in stream_config.go; it is referenced from ScenarioConfig.

// WebhookConfig mirrors model.WebhookConfig for YAML deserialization and validation.
// Secret values are NEVER accepted inline; only the env var name is permitted.
type WebhookConfig struct {
	// URL is the absolute target URL the webhook is delivered to.
	URL string `yaml:"url" json:"url"`

	// Method is the HTTP method used for delivery. Defaults to POST. POST/PUT/PATCH only.
	Method string `yaml:"method,omitempty" json:"method,omitempty"`

	// Headers are extra HTTP headers sent with each delivery attempt.
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// Body is the request body sent to the target URL. "{{uuid}}" is replaced per attempt.
	Body string `yaml:"body,omitempty" json:"body,omitempty"`

	// SecretEnv is the name of the environment variable holding the HMAC secret.
	// The secret value itself is never accepted inline.
	SecretEnv string `yaml:"secret_env,omitempty" json:"secret_env,omitempty"`

	// MaxAttempts is the maximum total delivery attempts including the first. Default 3.
	MaxAttempts int `yaml:"max_attempts,omitempty" json:"max_attempts,omitempty"`

	// BaseMS is the base backoff in milliseconds. Default 500.
	BaseMS int `yaml:"base_ms,omitempty" json:"base_ms,omitempty"`

	// MaxMS caps the backoff delay in milliseconds. Default 30000.
	MaxMS int `yaml:"max_ms,omitempty" json:"max_ms,omitempty"`
}

// allowedWebhookMethods is the set of HTTP methods the dispatcher is allowed to use.
var allowedWebhookMethods = map[string]struct{}{
	httpMethodPOST:  {},
	httpMethodPUT:   {},
	httpMethodPATCH: {},
}

// HTTP method constants used in webhook validation (avoid pulling net/http into config's API surface).
const (
	httpMethodPOST  = "POST"
	httpMethodPUT   = "PUT"
	httpMethodPATCH = "PATCH"
)

// validate verifies the webhook configuration is acceptable. It returns the first error found.
// Used both when parsing YAML with strict known-fields and when checking config-derived values.
func (w *WebhookConfig) validate() error {
	if w == nil {
		return nil
	}
	if err := w.validateURL(); err != nil {
		return err
	}
	if err := w.validateMethod(); err != nil {
		return err
	}
	return w.validateRetries()
}

// validateURL checks the URL field parses and has a scheme + host.
func (w *WebhookConfig) validateURL() error {
	if strings.TrimSpace(w.URL) == "" {
		return errors.New("webhook: url is required")
	}
	parsed, err := url.Parse(w.URL)
	if err != nil {
		return fmt.Errorf("webhook: invalid url %q: %w", w.URL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("webhook: invalid url %q: must include scheme and host", w.URL)
	}
	return nil
}

// validateMethod enforces the POST/PUT/PATCH allowlist and normalizes empty method to POST.
func (w *WebhookConfig) validateMethod() error {
	method := strings.ToUpper(strings.TrimSpace(w.Method))
	if method == "" {
		method = httpMethodPOST
	}
	if _, ok := allowedWebhookMethods[method]; !ok {
		return fmt.Errorf("webhook: method %q not allowed (must be POST, PUT, or PATCH)", w.Method)
	}
	return nil
}

// validateRetries checks the retry-related fields are non-negative.
func (w *WebhookConfig) validateRetries() error {
	if w.MaxAttempts < 0 {
		return fmt.Errorf("webhook: maxAttempts must be >= 0, got %d", w.MaxAttempts)
	}
	if w.BaseMS < 0 {
		return fmt.Errorf("webhook: baseMs must be >= 0, got %d", w.BaseMS)
	}
	if w.MaxMS < 0 {
		return fmt.Errorf("webhook: maxMs must be >= 0, got %d", w.MaxMS)
	}
	return nil
}

// webhookYAMLKeys is the closed allowlist of YAML keys under `webhook:`.
// Any key not listed is rejected by the strict YAML decoder to prevent
// inline `secret` and other unintended fields from sneaking in.
var webhookYAMLKeys = map[string]struct{}{
	"url":          {},
	"method":       {},
	"headers":      {},
	"body":         {},
	"secret_env":   {},
	"max_attempts": {},
	"base_ms":      {},
	"max_ms":       {},
}

// ToModelScenario converts a ScenarioConfig to a model.Scenario
// Optionally accepts a FixtureResolver to resolve fixture references in Data field
func (sf *ScenarioConfig) ToModelScenario(fixtureResolver *FixtureResolver) model.Scenario {
	// Set defaults
	statusCode := sf.StatusCode
	if statusCode == 0 {
		statusCode = 200
	}

	contentType := sf.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	// Resolve fixture references in data if resolver is provided
	data := sf.Data
	if fixtureResolver != nil {
		resolvedData, err := fixtureResolver.ResolveFixture(data)
		if err == nil {
			data = resolvedData
		}
		// If resolution fails, use original data (backward compatibility)
	}

	// Combine method and path into RequestPath format
	requestPath := fmt.Sprintf("%s %s", strings.ToUpper(sf.Method), sf.Path)

	scenario := model.Scenario{
		UUID:        sf.UUID, // Will be auto-generated by scenario service if empty
		RequestPath: requestPath,
		StatusCode:  statusCode,
		ContentType: contentType,
		Location:    sf.Location,
		Data:        data,
		Headers:     sf.Headers,
	}
	if sf.Webhook != nil {
		scenario.Webhook = sf.Webhook.toModel()
	}
	if sf.Stream != nil {
		scenario.Stream = sf.Stream.toModel()
	}
	return scenario
}

// toModel converts the config webhook into the runtime model representation.
// The Method is normalized to upper case; default POST is applied here so the
// dispatcher can rely on it.
func (w *WebhookConfig) toModel() *model.WebhookConfig {
	method := strings.ToUpper(strings.TrimSpace(w.Method))
	if method == "" {
		method = httpMethodPOST
	}
	return &model.WebhookConfig{
		URL:         w.URL,
		Method:      method,
		Headers:     w.Headers,
		Body:        w.Body,
		SecretEnv:   w.SecretEnv,
		MaxAttempts: w.MaxAttempts,
		BaseMS:      w.BaseMS,
		MaxMS:       w.MaxMS,
	}
}

// IDExtractionConfig provides a simplified way to configure ID extraction
type IDExtractionConfig struct {
	// BodyPaths defines XPath-like paths to extract IDs from request bodies
	BodyPaths []string `yaml:"body_paths,omitempty" json:"body_paths,omitempty"`

	// HeaderNames specifies the HTTP header names to extract IDs from
	HeaderNames []string `yaml:"header_names,omitempty" json:"header_names,omitempty"`
}

// Section represents a configuration section for a specific API endpoint pattern
type Section struct {
	// PathPattern defines the URL pattern to match against.
	// Use * as a wildcard for single path segments, e.g. "/users/*" or "/users/*/orders/*"
	// Use ** as a wildcard for multiple path segments recursively, e.g. "/api/**"
	PathPattern string `yaml:"path_pattern" json:"path_pattern"`

	// StrictPath determines whether GET/PUT/DELETE operations require path structure compatibility.
	// When true:
	//   - Resources are only accessible via paths that extend their creation path
	//   - Example: Resource created at "/users/subpath" accessible via "/users/subpath/123" but not "/users/123"
	//   - PUT returns 404 if resource doesn't exist (no upsert behavior)
	//   - Enforces strict path structure validation for cross-path access prevention
	// When false (default):
	//   - Resources accessible via any path matching the section pattern (flexible cross-path access)
	//   - Example: Resource created at "/users/subpath" accessible via both "/users/subpath/123" and "/users/123"
	//   - PUT performs upsert operations (creates if doesn't exist)
	//   - Backward compatible behavior with flexible path matching
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

	// HeaderIDNames specifies the HTTP header names to extract IDs from.
	// Multiple headers can be specified to support different ID extraction methods.
	// If empty, no header-based ID extraction will be performed.
	HeaderIDNames []string `yaml:"header_id_names,omitempty" json:"header_id_names,omitempty"`

	// IDExtraction provides a simplified way to configure ID extraction in unified config
	IDExtraction *IDExtractionConfig `yaml:"id_extraction,omitempty" json:"id_extraction,omitempty"`

	// CaseSensitive determines whether path matching is case-sensitive.
	// If true, paths must match exactly including case.
	// If false, paths are matched case-insensitively.
	CaseSensitive bool `yaml:"case_sensitive" json:"case_sensitive"`

	// ReturnBody determines whether POST/PUT/DELETE operations should return the resource body.
	// When true, successful POST/PUT/DELETE operations return the created/updated resource in the response body.
	// When false (default), successful operations return an empty body with appropriate status codes.
	// This flag provides simple control over response body behavior without requiring transformations.
	ReturnBody bool `yaml:"return_body" json:"return_body"`

	// Transformations contains request/response transformation functions.
	// This field is only available when using Unimock as a library and is excluded from YAML serialization.
	// It allows programmatic modification of requests and responses for advanced testing scenarios.
	Transformations *TransformationConfig `yaml:"-" json:"-"`
}

// NewUniConfig creates an empty UniConfig with an initialized Sections map
func NewUniConfig() *UniConfig {
	return &UniConfig{
		Sections:  make(map[string]Section),
		Scenarios: []ScenarioConfig{},
	}
}

// LoadFromYAML loads a UniConfig from a YAML file at the given path
// Supports both legacy format (sections at root) and unified format (sections nested)
func LoadFromYAML(path string) (*UniConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg, unifiedErr := tryUnifiedFormat(data)
	if unifiedErr == nil {
		return finalizeUnifiedConfig(cfg, data, path)
	}
	if legacyCfg, legacyErr := tryLegacyFormat(data); legacyErr == nil {
		legacyCfg.initializeFixtureResolver(filepath.Dir(path))
		return legacyCfg, nil
	}
	// Surface the unified error for better debugging when both formats fail.
	return nil, unifiedErr
}

// finalizeUnifiedConfig runs the webhook-aware validation, normalization and
// fixture resolver initialization on a successfully unified-parsed config.
func finalizeUnifiedConfig(cfg *UniConfig, rawYAML []byte, path string) (*UniConfig, error) {
	if err := checkScenarioKeysInYAML(rawYAML); err != nil {
		return nil, err
	}
	if err := cfg.validateScenarios(); err != nil {
		return nil, err
	}
	cfg.Normalize()
	cfg.initializeFixtureResolver(filepath.Dir(path))
	return cfg, nil
}

// tryUnifiedFormat attempts to decode the YAML as the unified format with a loose
// decoder (preserves backward compatibility with unknown top-level fields) and
// returns the parsed config along with any decode error.
func tryUnifiedFormat(data []byte) (*UniConfig, error) {
	cfg := NewUniConfig()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(false)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}
	if len(cfg.Sections) == 0 && len(cfg.Scenarios) == 0 {
		return nil, errors.New("not unified format")
	}
	return cfg, nil
}

// tryLegacyFormat attempts to decode the YAML as the legacy format where sections
// appear at the root level. Uses a strict decoder to catch typos.
func tryLegacyFormat(data []byte) (*UniConfig, error) {
	var legacy struct {
		Sections map[string]Section `yaml:",inline"`
	}
	legacy.Sections = make(map[string]Section)
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&legacy); err != nil {
		return nil, err
	}
	cfg := NewUniConfig()
	cfg.Sections = legacy.Sections
	return cfg, nil
}

// checkScenarioKeysInYAML parses the YAML document and rejects any unknown keys
// under a `webhook:` or `stream:` mapping on a scenario. This is what makes the
// inline `secret:` and any other unknown webhook/stream field produce a clear
// error at config load time, while keeping the rest of the top-level decoder
// loose for backward compatibility with existing scenario schemas that may use
// legacy fields.
func checkScenarioKeysInYAML(data []byte) error {
	rootNode, err := decodeRootNode(data)
	if err != nil || rootNode == nil {
		return err
	}
	scenariosNode := findChildMappingValue(rootNode, "scenarios")
	if scenariosNode == nil || scenariosNode.Kind != yaml.SequenceNode {
		return nil
	}
	return checkEachScenarioKeys(scenariosNode)
}

// checkEachScenarioKeys dispatches per-scenario sub-mapping key checks for
// every entry in scenariosNode.
func checkEachScenarioKeys(scenariosNode *yaml.Node) error {
	for sIdx, scn := range scenariosNode.Content {
		if err := checkScenarioWebhookKeys(scn, sIdx); err != nil {
			return err
		}
		if err := checkScenarioStreamKeys(scn, sIdx); err != nil {
			return err
		}
	}
	return nil
}

// checkScenarioStreamKeys scans a scenario mapping node for a `stream:` sub-mapping
// and rejects any key outside the streamYAMLKeys allowlist with a clear error.
func checkScenarioStreamKeys(scn *yaml.Node, sIdx int) error {
	if scn == nil || scn.Kind != yaml.MappingNode {
		return nil
	}
	streamNode := findChildMappingValue(scn, "stream")
	if streamNode == nil || streamNode.Kind != yaml.MappingNode {
		return nil
	}
	return rejectUnknownStreamKeys(streamNode, sIdx)
}

// rejectUnknownStreamKeys returns an error for the first key in streamNode that
// is not in the streamYAMLKeys allowlist.
func rejectUnknownStreamKeys(streamNode *yaml.Node, sIdx int) error {
	for j := 0; j+1 < len(streamNode.Content); j += 2 {
		whKeyNode := streamNode.Content[j]
		whKey := whKeyNode.Value
		if _, ok := streamYAMLKeys[whKey]; ok {
			continue
		}
		return fmt.Errorf("scenarios[%d]: stream: unknown field %q at line %d "+
			"(allowed: format, interval_ms, event_count, hold_open, template)",
			sIdx, whKey, whKeyNode.Line)
	}
	return nil
}

// decodeRootNode unmarshals the YAML into a single root node for inspection.
func decodeRootNode(data []byte) (*yaml.Node, error) {
	var rootNode yaml.Node
	if err := yaml.Unmarshal(data, &rootNode); err != nil {
		return nil, err
	}
	if rootNode.Kind != yaml.DocumentNode || len(rootNode.Content) == 0 {
		return nil, nil
	}
	if rootNode.Content[0].Kind != yaml.MappingNode {
		return nil, nil
	}
	return rootNode.Content[0], nil
}

// findChildMappingValue returns the value node for the given key in a mapping node,
// or nil if not found.
func findChildMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// checkScenarioWebhookKeys scans a scenario mapping node for a `webhook:` sub-mapping
// and rejects any key outside the webhookYAMLKeys allowlist with a clear error.
func checkScenarioWebhookKeys(scn *yaml.Node, sIdx int) error {
	if scn == nil || scn.Kind != yaml.MappingNode {
		return nil
	}
	webhookNode := findChildMappingValue(scn, "webhook")
	if webhookNode == nil || webhookNode.Kind != yaml.MappingNode {
		return nil
	}
	return rejectUnknownWebhookKeys(webhookNode, sIdx)
}

// rejectUnknownWebhookKeys returns an error for the first key in webhookNode that
// is not in the webhookYAMLKeys allowlist.
func rejectUnknownWebhookKeys(webhookNode *yaml.Node, sIdx int) error {
	for j := 0; j+1 < len(webhookNode.Content); j += 2 {
		whKeyNode := webhookNode.Content[j]
		whKey := whKeyNode.Value
		if _, ok := webhookYAMLKeys[whKey]; ok {
			continue
		}
		return fmt.Errorf("scenarios[%d]: webhook: unknown field %q at line %d "+
			"(allowed: url, method, headers, body, secret_env, max_attempts, base_ms, max_ms)",
			sIdx, whKey, whKeyNode.Line)
	}
	return nil
}

// validateScenarios runs webhook-level and stream-level validation on every
// scenario in the config. Stream and Data are mutually exclusive: if both are
// set on a scenario, validation fails.
func (uc *UniConfig) validateScenarios() error {
	for i, sc := range uc.Scenarios {
		if err := validateOneScenario(i, sc); err != nil {
			return err
		}
	}
	return nil
}

// validateOneScenario validates webhook, stream and the stream/data exclusivity
// constraint for a single scenario at index i.
func validateOneScenario(i int, sc ScenarioConfig) error {
	if sc.Webhook != nil {
		if err := sc.Webhook.validate(); err != nil {
			return fmt.Errorf("scenarios[%d] (path=%q): %w", i, sc.Path, err)
		}
	}
	if sc.Stream != nil {
		if err := sc.Stream.validate(); err != nil {
			return fmt.Errorf("scenarios[%d] (path=%q): %w", i, sc.Path, err)
		}
	}
	if hasStreamAndData(sc) {
		return fmt.Errorf(
			"scenarios[%d] (path=%q): stream and data are mutually exclusive",
			i, sc.Path,
		)
	}
	return nil
}

// hasStreamAndData reports whether the scenario has both stream and data set.
func hasStreamAndData(sc ScenarioConfig) bool {
	return sc.Stream != nil && strings.TrimSpace(sc.Data) != ""
}

// initializeFixtureResolver sets up the fixture resolver with the configuration file's directory
func (uc *UniConfig) initializeFixtureResolver(baseDir string) {
	uc.baseDir = baseDir
	uc.fixtureResolver = NewFixtureResolver(baseDir)
}

// GetFixtureResolver returns the fixture resolver for this configuration
func (uc *UniConfig) GetFixtureResolver() *FixtureResolver {
	return uc.fixtureResolver
}

// Normalize ensures consistent field values across different configuration formats
func (uc *UniConfig) Normalize() {
	for name, section := range uc.Sections {
		section.Normalize()
		uc.Sections[name] = section
	}
}

// Normalize ensures consistent field values for different configuration formats
func (s *Section) Normalize() {
	s.normalizePathFields()
	s.normalizeIDExtractionFields()
}

// normalizePathFields handles path field variants (no longer needed - only PathPattern supported)
func (*Section) normalizePathFields() {
	// No normalization needed - only PathPattern is supported
}

// normalizeIDExtractionFields handles ID extraction field variants
func (s *Section) normalizeIDExtractionFields() {
	if s.IDExtraction == nil {
		return
	}

	if len(s.BodyIDPaths) == 0 && len(s.IDExtraction.BodyPaths) > 0 {
		s.BodyIDPaths = s.IDExtraction.BodyPaths
	}
	if len(s.HeaderIDNames) == 0 && len(s.IDExtraction.HeaderNames) > 0 {
		s.HeaderIDNames = s.IDExtraction.HeaderNames
	}
}

// GetPathPattern returns the path pattern
func (s *Section) GetPathPattern() string {
	return s.PathPattern
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
func (uc *UniConfig) MatchPath(path string) (string, *Section, error) {
	normalizedPath := strings.Trim(path, PathSeparator)

	// First try exact matches (no wildcards)
	if name, section := uc.findExactMatch(normalizedPath); section != nil {
		return name, section, nil
	}

	// Then try wildcard matches, prioritizing longer patterns
	if name, section := uc.findBestWildcardMatch(normalizedPath); section != nil {
		return name, section, nil
	}

	return "", nil, nil // No match found
}

// findExactMatch looks for exact pattern matches (no wildcards)
func (uc *UniConfig) findExactMatch(normalizedPath string) (string, *Section) {
	for name, section := range uc.Sections {
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
func (uc *UniConfig) findBestWildcardMatch(normalizedPath string) (string, *Section) {
	bestMatch := wildcardMatch{name: "", numSegments: noMatch}

	for name, section := range uc.Sections {
		if match := uc.evaluateWildcardSection(name, section, normalizedPath); match.isValid() {
			if match.isBetterThan(bestMatch) {
				bestMatch = match
			}
		}
	}

	return bestMatch.getResult(uc)
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
func (m wildcardMatch) getResult(uc *UniConfig) (string, *Section) {
	if !m.isValid() {
		return "", nil
	}
	matchedSection := uc.Sections[m.name]
	return m.name, &matchedSection
}

// evaluateWildcardSection checks if a section matches and returns match info
func (*UniConfig) evaluateWildcardSection(name string, section Section, normalizedPath string) wildcardMatch {
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
