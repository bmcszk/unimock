package handler_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestUniHandler_StrictPathBehavior(t *testing.T) {
	tests := getStrictPathTestCases()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executeStrictPathTest(t, tt)
		})
	}
}

// executeStrictPathTest runs a single strict path test case
func executeStrictPathTest(t *testing.T, tt strictPathTestCase) {
	t.Helper()
	
	// Setup handler with strict path configuration
	deps := setupStrictPathHandler(t, tt.strictPath)
	
	// Setup test data
	if tt.setupData != nil {
		tt.setupData(deps.store)
	}
	
	// Create and execute request
	req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
	if tt.contentType != "" {
		req.Header.Set("Content-Type", tt.contentType)
	}

	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	// Validate response
	validateStrictPathResponse(t, w, tt)
}

// validateStrictPathResponse validates the response for strict path tests
func validateStrictPathResponse(t *testing.T, w *httptest.ResponseRecorder, tt strictPathTestCase) {
	t.Helper()
	
	if w.Code != tt.expectedStatus {
		t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
	}

	if tt.expectedBodyContains != "" {
		respBody := strings.TrimSpace(w.Body.String())
		if !strings.Contains(respBody, tt.expectedBodyContains) {
			t.Errorf("expected body to contain '%s', got '%s'", tt.expectedBodyContains, respBody)
		}
	}
}

// strictPathTestCase represents a test case for strict path behavior
type strictPathTestCase struct {
	name               string
	strictPath         bool
	method             string
	path               string
	body               string
	contentType        string
	setupData          func(storage.UniStorage)
	expectedStatus     int
	expectedBodyContains string
	description        string
}

// getStrictPathTestCases returns all test cases for strict path behavior
func getStrictPathTestCases() []strictPathTestCase {
	return []strictPathTestCase{
		// GET behavior tests
		{
			name:       "GET with strict_path=false returns 404 for nonexistent individual resource",
			strictPath: false,
			method:     http.MethodGet,
			path:       "/users/nonexistent",
			setupData:  setupUsersCollection,
			expectedStatus: http.StatusNotFound,
			expectedBodyContains: "resource not found",
			description: "Individual resource requests should return 404 when resource doesn't exist",
		},
		{
			name:       "GET with strict_path=true should return 404",
			strictPath: true,
			method:     http.MethodGet,
			path:       "/users/nonexistent",
			setupData:  setupUsersCollection,
			expectedStatus: http.StatusNotFound,
			expectedBodyContains: "resource not found",
			description: "When strict_path=true, GET must match BOTH path pattern AND exact resource ID",
		},
		{
			name:       "GET existing resource works with strict_path=true",
			strictPath: true,
			method:     http.MethodGet,
			path:       "/users/123",
			setupData:  setupUsersCollection,
			expectedStatus: http.StatusOK,
			expectedBodyContains: `{"id": "123", "name": "test"}`,
			description: "When both path pattern and resource ID match, strict_path=true works normally",
		},
		// PUT behavior tests
		{
			name:        "PUT with strict_path=false should create new resource (upsert)",
			strictPath:  false,
			method:      http.MethodPut,
			path:        "/users/newuser",
			body:        `{"id": "newuser", "name": "New User"}`,
			contentType: "application/json",
			setupData:   setupUsersCollection,
			expectedStatus: http.StatusOK,
			expectedBodyContains: `{"id": "newuser", "name": "New User"}`,
			description: "When strict_path=false, PUT performs upsert (creates if doesn't exist)",
		},
		{
			name:        "PUT with strict_path=true should return 404 for new resource",
			strictPath:  true,
			method:      http.MethodPut,
			path:        "/users/newuser",
			body:        `{"id": "newuser", "name": "New User"}`,
			contentType: "application/json",
			setupData:   setupUsersCollection,
			expectedStatus: http.StatusNotFound,
			expectedBodyContains: "resource not found",
			description: "When strict_path=true, PUT must match BOTH path pattern AND exact resource ID",
		},
		{
			name:        "PUT existing resource works with strict_path=true",
			strictPath:  true,
			method:      http.MethodPut,
			path:        "/users/123",
			body:        `{"id": "123", "name": "Updated User"}`,
			contentType: "application/json",
			setupData:   setupUsersCollection,
			expectedStatus: http.StatusOK,
			expectedBodyContains: `{"id": "123", "name": "Updated User"}`,
			description: "When both path pattern and resource ID match, PUT with strict_path=true works",
		},
		// DELETE behavior tests
		{
			name:       "DELETE with strict_path=true should return 404 for nonexistent resource",
			strictPath: true,
			method:     http.MethodDelete,
			path:       "/users/nonexistent",
			setupData:  setupUsersCollection,
			expectedStatus: http.StatusNotFound,
			expectedBodyContains: "resource not found",
			description: "When strict_path=true, DELETE must match BOTH path pattern AND exact resource ID",
		},
		{
			name:       "DELETE existing resource works with strict_path=true",
			strictPath: true,
			method:     http.MethodDelete,
			path:       "/users/123",
			setupData:  setupUsersCollection,
			expectedStatus: http.StatusNoContent,
			description: "When both path pattern and resource ID match, DELETE with strict_path=true works",
		},
	}
}

func TestUniHandler_WildcardPatternMatching(t *testing.T) {
	tests := []struct {
		name           string
		pathPattern    string
		requestPath    string
		shouldMatch    bool
		method         string
		expectedStatus int
		description    string
	}{
		// Single wildcard tests
		{
			name:           "single wildcard matches exact segment",
			pathPattern:    "/users/*",
			requestPath:    "/users/123",
			shouldMatch:    true,
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound, // No data setup, should get 404
			description:    "Single * should match exactly one segment",
		},
		{
			name:           "single wildcard rejects multiple segments",
			pathPattern:    "/users/*",
			requestPath:    "/users/123/posts",
			shouldMatch:    false,
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound, // No matching section
			description:    "Single * should not match multiple segments",
		},

		// Recursive wildcard tests
		{
			name:           "recursive wildcard matches multiple segments",
			pathPattern:    "/api/**",
			requestPath:    "/api/v1/users/123",
			shouldMatch:    true,
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound, // No data setup, should get 404
			description:    "** should match multiple segments",
		},
		{
			name:           "recursive wildcard matches zero segments",
			pathPattern:    "/api/**",
			requestPath:    "/api",
			shouldMatch:    true,
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound, // No data setup, should get 404
			description:    "** should match zero segments",
		},

		// Mixed wildcard tests
		{
			name:           "mixed wildcards work together",
			pathPattern:    "/users/*/posts/**",
			requestPath:    "/users/123/posts/456/comments",
			shouldMatch:    true,
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound, // No data setup, should get 404
			description:    "* and ** should work together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWildcardHandlerMatch(t, tt)
		})
	}
}

// testWildcardHandlerMatch helper function to test wildcard handler matching
func testWildcardHandlerMatch(t *testing.T, tt struct {
	name           string
	pathPattern    string
	requestPath    string
	shouldMatch    bool
	method         string
	expectedStatus int
	description    string
}) {
	t.Helper()
	
	// Setup handler with the specific pattern
	deps := setupWildcardHandler(t, tt.pathPattern)
	
	// Create and execute request
	req := httptest.NewRequest(tt.method, tt.requestPath, nil)
	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	if tt.shouldMatch {
		validateExpectedMatch(t, w, tt.expectedStatus)
	} else {
		validateNoMatch(t, w)
	}
}

// validateExpectedMatch validates that the pattern matched as expected
func validateExpectedMatch(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if w.Code != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, w.Code)
	}
}

// validateNoMatch validates that no pattern matched
func validateNoMatch(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	respBody := w.Body.String()
	if !strings.Contains(respBody, "no matching section") {
		t.Errorf("expected 'no matching section' error, got '%s'", respBody)
	}
}

// Test helper functions

type strictPathHandlerDeps struct {
	handler *handler.UniHandler
	store   storage.UniStorage
	config  *config.UniConfig
	logger  *slog.Logger
}

func setupStrictPathHandler(t *testing.T, strictPath bool) strictPathHandlerDeps {
	t.Helper()
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/*",
				StrictPath:    strictPath,
				BodyIDPaths:   []string{"/id"},
				CaseSensitive: true,
				ReturnBody:    true, // For backward compatibility with existing tests
			},
		},
	}

	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, cfg)
	
	return strictPathHandlerDeps{
		handler: uniHandler,
		store:   store,
		config:  cfg,
		logger:  logger,
	}
}

func setupWildcardHandler(t *testing.T, pathPattern string) strictPathHandlerDeps {
	t.Helper()
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"test_pattern": {
				PathPattern:   pathPattern,
				StrictPath:    false,
				BodyIDPaths:   []string{"/id"},
				CaseSensitive: true,
			},
		},
	}

	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, cfg)
	
	return strictPathHandlerDeps{
		handler: uniHandler,
		store:   store,
		config:  cfg,
		logger:  logger,
	}
}

func setupUsersCollection(store storage.UniStorage) {
	testData := model.UniData{
		Path:        "/users/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test"}`),
	}
	// Create in both modes to support both strict and non-strict tests
	store.Create("users", false, testData)
	store.Create("users", true, testData)
}