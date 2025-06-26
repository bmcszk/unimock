package handler_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUniHandler_MetricsIntegration(t *testing.T) {
	mockHandler, techService := setupMetricsTestHandler(t)
	ctx := context.Background()

	tests := getMetricsTestCases()

	// Execute all requests
	for _, tt := range tests {
		t.Run(tt.method+"_"+tt.path, func(t *testing.T) {
			executeMetricsTestRequest(ctx, t, mockHandler, tt)
		})
	}

	// Verify metrics
	verifyMetricsResults(ctx, t, techService, tests)
}

func TestUniHandler_MetricsTracking_AllHTTPMethods(t *testing.T) {
	mockHandler, techService := setupHTTPMethodsTestHandler(t)
	ctx := context.Background()

	// Create initial resource
	createInitialResource(ctx, t, mockHandler)

	// Test all HTTP methods
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD"}
	testAllHTTPMethods(ctx, t, mockHandler, methods)

	// Verify tracking
	verifyHTTPMethodsTracking(ctx, t, techService)
}

// Helper functions to reduce cognitive complexity

type metricsTestCase struct {
	method       string
	path         string
	body         string
	expectedCode int
}

func setupMetricsTestHandler(t *testing.T) (*handler.UniHandler, service.TechService) {
	t.Helper()
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	techService := service.NewTechService(time.Now())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/api/users/*",
				BodyIDPaths: []string{"/id"},
				ReturnBody:  true,
			},
			"products": {
				PathPattern: "/api/products/*",
				BodyIDPaths: []string{"/id"},
				ReturnBody:  false,
			},
		},
	}

	mockService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	mockHandler := handler.NewUniHandler(mockService, scenarioService, techService, logger, cfg)
	
	return mockHandler, techService
}

func setupHTTPMethodsTestHandler(t *testing.T) (*handler.UniHandler, service.TechService) {
	t.Helper()
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	techService := service.NewTechService(time.Now())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"test": {
				PathPattern: "/test/*",
				BodyIDPaths: []string{"/id"},
				ReturnBody:  true,
			},
		},
	}

	mockService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	mockHandler := handler.NewUniHandler(mockService, scenarioService, techService, logger, cfg)
	
	return mockHandler, techService
}

func getMetricsTestCases() []metricsTestCase {
	return []metricsTestCase{
		{"POST", "/api/users", `{"id": "1", "name": "John"}`, 201},
		{"GET", "/api/users/1", "", 200},
		{"PUT", "/api/users/1", `{"id": "1", "name": "John Updated"}`, 200},
		{"DELETE", "/api/users/1", "", 204},
		{"GET", "/api/users/999", "", 404}, // Non-existent resource
		{"POST", "/api/products", `{"id": "p1", "name": "Product 1"}`, 201},
		{"GET", "/api/products/p1", "", 200},
		{"GET", "/api/invalid", "", 404}, // No matching section
		{"INVALID", "/api/users", "", 405}, // Invalid method
	}
}

func executeMetricsTestRequest(
	ctx context.Context, t *testing.T, mockHandler *handler.UniHandler, tt metricsTestCase,
) {
	t.Helper()
	var req *http.Request
	if tt.body != "" {
		req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(tt.method, tt.path, nil)
	}

	resp, err := mockHandler.HandleRequest(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, tt.expectedCode, resp.StatusCode, 
		"Expected status code %d for %s %s", tt.expectedCode, tt.method, tt.path)
	if resp.Body != nil {
		resp.Body.Close()
	}
}

func verifyMetricsResults(ctx context.Context, t *testing.T, techService service.TechService, tests []metricsTestCase) {
	t.Helper()
	metrics := techService.GetMetrics(ctx)

	// Check total request count
	requestCount, ok := metrics["request_count"].(int64)
	require.True(t, ok, "request_count should be int64")
	assert.Equal(t, int64(len(tests)), requestCount, "Expected %d total requests", len(tests))

	// Check endpoint tracking
	apiEndpoints, ok := metrics["api_endpoints"].(map[string]int64)
	require.True(t, ok, "api_endpoints should be map[string]int64")

	// Verify specific endpoints were tracked
	expectedEndpoints := map[string]int64{
		"/api/users":     2, // POST, INVALID method
		"/api/users/1":   3, // GET, PUT, DELETE  
		"/api/users/999": 1, // GET (404)
		"/api/products":  1, // POST
		"/api/products/p1": 1, // GET
		"/api/invalid":   1, // GET (404)
	}

	for path, expectedCount := range expectedEndpoints {
		actualCount, exists := apiEndpoints[path]
		assert.True(t, exists, "Path %s should exist in metrics", path)
		assert.Equal(t, expectedCount, actualCount, "Expected %d requests for path %s", expectedCount, path)
	}

	// Check status code tracking
	statusCodeStats, ok := metrics["status_code_stats"].(map[string]map[string]int64)
	require.True(t, ok, "status_code_stats should be map[string]map[string]int64")

	// Verify status codes for specific paths
	usersPaths := statusCodeStats["/api/users"]
	assert.Equal(t, int64(1), usersPaths["201"], "Expected 1 POST request with 201 status")
	assert.Equal(t, int64(1), usersPaths["405"], "Expected 1 INVALID method request with 405 status")

	users1Path := statusCodeStats["/api/users/1"]
	assert.Equal(t, int64(2), users1Path["200"], "Expected 2 requests (GET + PUT) with 200 status")
	assert.Equal(t, int64(1), users1Path["204"], "Expected 1 DELETE request with 204 status") 

	users999Path := statusCodeStats["/api/users/999"]
	assert.Equal(t, int64(1), users999Path["404"], "Expected 1 GET request with 404 status")

	invalidPath := statusCodeStats["/api/invalid"]
	assert.Equal(t, int64(1), invalidPath["404"], "Expected 1 GET request with 404 status")
}

func createInitialResource(ctx context.Context, t *testing.T, mockHandler *handler.UniHandler) {
	t.Helper()
	postReq := httptest.NewRequest("POST", "/test", strings.NewReader(`{"id": "123", "data": "test"}`))
	postReq.Header.Set("Content-Type", "application/json")
	postResp, err := mockHandler.HandleRequest(ctx, postReq)
	require.NoError(t, err)
	if postResp.Body != nil {
		postResp.Body.Close()
	}
}

func testAllHTTPMethods(ctx context.Context, t *testing.T, mockHandler *handler.UniHandler, methods []string) {
	t.Helper()
	path := "/test/123"
	
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			testSingleHTTPMethod(ctx, t, mockHandler, method, path)
		})
	}
}

func testSingleHTTPMethod(ctx context.Context, t *testing.T, mockHandler *handler.UniHandler, method, path string) {
	t.Helper()
	var req *http.Request
	if method == "PUT" {
		req = httptest.NewRequest(method, path, strings.NewReader(`{"id": "123", "data": "updated"}`))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	resp, err := mockHandler.HandleRequest(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	if resp.Body != nil {
		resp.Body.Close()
	}
}

func verifyHTTPMethodsTracking(ctx context.Context, t *testing.T, techService service.TechService) {
	t.Helper()
	metrics := techService.GetMetrics(ctx)
	apiEndpoints, ok := metrics["api_endpoints"].(map[string]int64)
	require.True(t, ok, "api_endpoints should be map[string]int64")
	
	// Should have tracked: POST /test, GET /test/123, POST /test/123, PUT /test/123, DELETE /test/123, HEAD /test/123
	assert.True(t, apiEndpoints["/test"] > 0, "POST to /test should be tracked")
	assert.True(t, apiEndpoints["/test/123"] > 0, "Requests to /test/123 should be tracked")

	// Total should be 6 (1 POST + 5 methods to /test/123)
	requestCount, ok := metrics["request_count"].(int64)
	require.True(t, ok, "request_count should be int64")
	assert.Equal(t, int64(6), requestCount, "Expected 6 total requests")
}