package handler_test

import (
	"io"
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
	"github.com/google/uuid"
)

func TestUniHandler_HandleRequest(t *testing.T) {
	deps := setupUniHandlerFull(t)
	testData := getTestData()
	populateTestData(deps.store, testData)

	tests := getHandlerTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uniHandler := prepareHandlerForTest(t, tt, deps, testData)
			executeTestRequest(t, tt, uniHandler)
		})
	}
}

// validatePostUUIDResponse validates POST response with UUID generation
func validatePostUUIDResponse(t *testing.T, w *httptest.ResponseRecorder, path string) {
	t.Helper()
	location := w.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header to be set")
		return
	}

	if !strings.HasPrefix(location, path+"/") {
		t.Errorf("expected Location header to start with %s/, got %s", path, location)
		return
	}

	extractedID := location[len(path)+1:]
	if _, err := uuid.Parse(extractedID); err != nil {
		t.Errorf("expected Location header to contain a valid UUID, got %s, error: %v", extractedID, err)
	}
}

// validateResponseBody validates response body content
func validateResponseBody(t *testing.T, w *httptest.ResponseRecorder, expectedBody string) {
	t.Helper()
	respBody := strings.TrimSpace(w.Body.String())

	if isErrorMessage(expectedBody) {
		if !strings.Contains(respBody, expectedBody) {
			t.Errorf("expected body to contain '%s', got '%s'", expectedBody, respBody)
		}
	} else {
		if respBody != expectedBody {
			t.Errorf("expected body '%s', got '%s'", expectedBody, respBody)
		}
	}
}

// uniHandlerDeps holds the dependencies for uni handler testing
type uniHandlerDeps struct {
	handler *handler.UniHandler
	store   storage.UniStorage
	config  *config.UniConfig
	logger  *slog.Logger
}

// setupUniHandlerFull creates and configures a mock handler with all dependencies for testing
func setupUniHandlerFull(t *testing.T) uniHandlerDeps {
	t.Helper()
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/*",
				BodyIDPaths:   []string{"/id"},
				CaseSensitive: true,
				ReturnBody:    true, // For backward compatibility with existing tests
			},
		},
	}

	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, cfg)

	return uniHandlerDeps{
		handler: uniHandler,
		store:   store,
		config:  cfg,
		logger:  logger,
	}
}

// getTestData returns the standard test data
func getTestData() []model.UniData {
	return []model.UniData{
		{
			Path:        "/users/123",
			ContentType: "application/json",
			Body:        []byte(`{"id": "123", "name": "test"}`),
		},
		{
			Path:        "/users/456",
			ContentType: "application/xml",
			Body:        []byte(`<user><id>456</id><name>test</name></user>`),
		},
		{
			Path:        "/users/789",
			ContentType: "application/octet-stream",
			Body:        []byte("binary data"),
		},
	}
}

// populateTestData adds test data to the storage
func populateTestData(store storage.UniStorage, testData []model.UniData) {
	for _, data := range testData {
		ids := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
		data.IDs = ids
		store.Create("users", false, data)
	}
}

// getHandlerTestCases returns the test cases for handler testing
func getHandlerTestCases() []struct {
	name           string
	method         string
	path           string
	contentType    string
	body           string
	expectedStatus int
	expectedBody   string
} {
	return []struct {
		name           string
		method         string
		path           string
		contentType    string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET existing resource",
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id": "123", "name": "test"}`,
		},
		{
			name:           "GET non-existent resource but collection exists",
			method:         http.MethodGet,
			path:           "/users/999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "resource not found",
		},
		{
			name:           "GET non-existent resource and collection",
			method:         http.MethodGet,
			path:           "/nonexistent/999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "no matching section found for path: /nonexistent/999",
		},
		{
			name:           "GET collection with mixed content types",
			method:         http.MethodGet,
			path:           "/users",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id": "123", "name": "test"}]`,
		},
		{
			name:           "GET empty collection",
			method:         http.MethodGet,
			path:           "/empty",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "no matching section found for path: /empty",
		},
		{
			name:           "POST new resource",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "application/json",
			body:           `{"id": "999", "name": "new"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "POST new resource with no ID (auto-generate UUID)",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "application/json",
			body:           `{"name": "new user, no id provided"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "POST with malformed JSON",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "application/json",
			body:           `{"id": "789", "name": "new"`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request: failed to parse JSON body",
		},
		{
			name:           "PUT existing resource",
			method:         http.MethodPut,
			path:           "/users/123",
			contentType:    "application/json",
			body:           `{"id": "123", "name": "updated"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PUT non-existent resource",
			method:         http.MethodPut,
			path:           "/users/unique-non-existent-id-9999",
			contentType:    "application/json",
			body:           `{"id": "unique-non-existent-id-9999", "name": "new"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id": "unique-non-existent-id-9999", "name": "new"}`,
		},
		{
			name:           "DELETE existing resource",
			method:         http.MethodDelete,
			path:           "/users/123",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DELETE non-existent resource but collection exists",
			method:         http.MethodDelete,
			path:           "/users/unique-non-existent-id-9999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "resource not found",
		},
		{
			name:           "DELETE non-existent resource and collection",
			method:         http.MethodDelete,
			path:           "/nonexistent/999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "no matching section found for path: /nonexistent/999",
		},
	}
}

// prepareHandlerForTest sets up the handler for test execution
func prepareHandlerForTest(t *testing.T, tt struct {
	name           string
	method         string
	path           string
	contentType    string
	body           string
	expectedStatus int
	expectedBody   string
}, deps uniHandlerDeps, testData []model.UniData) *handler.UniHandler {
	t.Helper()

	if needsCleanHandler(tt.name) {
		cleanStore := storage.NewUniStorage()
		populateTestData(cleanStore, testData)
		currentService := service.NewUniService(cleanStore, deps.config)
		currentScenarioStore := storage.NewScenarioStorage()
		currentScenarioService := service.NewScenarioService(currentScenarioStore)
		return handler.NewUniHandler(currentService, currentScenarioService, deps.logger, deps.config)
	}

	return deps.handler
}

// needsCleanHandler determines if a test needs a clean handler instance
func needsCleanHandler(testName string) bool {
	cleanHandlerTests := []string{
		"POST new resource with no ID (auto-generate UUID)",
		"POST new resource",
		"PUT non-existent resource",
		"DELETE non-existent resource but collection exists",
	}

	for _, name := range cleanHandlerTests {
		if testName == name {
			return true
		}
	}
	return false
}

// executeTestRequest executes the HTTP request and validates the response
func executeTestRequest(t *testing.T, tt struct {
	name           string
	method         string
	path           string
	contentType    string
	body           string
	expectedStatus int
	expectedBody   string
}, uniHandler *handler.UniHandler) {
	t.Helper()

	req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
	if tt.contentType != "" {
		req.Header.Set("Content-Type", tt.contentType)
	}

	w := httptest.NewRecorder()
	uniHandler.ServeHTTP(w, req)

	if w.Code != tt.expectedStatus {
		t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
	}

	if tt.name == "POST new resource with no ID (auto-generate UUID)" {
		validatePostUUIDResponse(t, w, tt.path)
	} else if tt.expectedBody != "" {
		validateResponseBody(t, w, tt.expectedBody)
	}
}

// isErrorMessage checks if the expected body is an error message
func isErrorMessage(expectedBody string) bool {
	errorPrefixes := []string{"invalid request:", "failed to", "no matching section", "resource not found"}
	for _, prefix := range errorPrefixes {
		if strings.Contains(expectedBody, prefix) {
			return true
		}
	}
	return false
}

func TestUniHandler_ReturnBodyFlag_POST(t *testing.T) {
	t.Run("POST with return_body false", testPOSTReturnBodyFalse)
	t.Run("POST with return_body true", testPOSTReturnBodyTrue)
}

func TestUniHandler_ReturnBodyFlag_PUT(t *testing.T) {
	t.Run("PUT with return_body false", testPUTReturnBodyFalse)
	t.Run("PUT with return_body true", testPUTReturnBodyTrue)
}

func TestUniHandler_ReturnBodyFlag_DELETE(t *testing.T) {
	t.Run("DELETE with return_body false", testDELETEReturnBodyFalse)
	t.Run("DELETE with return_body true", testDELETEReturnBodyTrue)
}

func testPOSTReturnBodyFalse(t *testing.T) {
	deps := setupUniHandlerFull(t)
	deps.config.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
		ReturnBody:  false,
	}

	requestBody := `{"id": "123", "name": "test"}`
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	if w.Body.String() != "" {
		t.Errorf("expected empty body when return_body is false, got: %s", w.Body.String())
	}
}

func testPOSTReturnBodyTrue(t *testing.T) {
	deps := setupUniHandlerFull(t)
	deps.config.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
		ReturnBody:  true,
	}

	requestBody := `{"id": "456", "name": "test with body"}`
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	if w.Body.String() == "" {
		t.Error("expected non-empty body when return_body is true")
	}
	if !strings.Contains(w.Body.String(), "test with body") {
		t.Errorf("expected body to contain request data, got: %s", w.Body.String())
	}
}

func testPUTReturnBodyFalse(t *testing.T) {
	deps := setupUniHandlerFull(t)
	deps.config.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
		ReturnBody:  false,
	}

	createResource(t, deps, "789", "initial")

	updateBody := `{"id": "789", "name": "updated"}`
	req := httptest.NewRequest("PUT", "/api/test/789", strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "" {
		t.Errorf("expected empty body when return_body is false, got: %s", w.Body.String())
	}
}

func testPUTReturnBodyTrue(t *testing.T) {
	deps := setupUniHandlerFull(t)
	deps.config.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
		ReturnBody:  true,
	}

	createResource(t, deps, "790", "initial")

	updateBody := `{"id": "790", "name": "updated with body"}`
	req := httptest.NewRequest("PUT", "/api/test/790", strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() == "" {
		t.Error("expected non-empty body when return_body is true")
	}
	if !strings.Contains(w.Body.String(), "updated with body") {
		t.Errorf("expected body to contain updated data, got: %s", w.Body.String())
	}
}

func testDELETEReturnBodyFalse(t *testing.T) {
	deps := setupUniHandlerFull(t)
	deps.config.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
		ReturnBody:  false,
	}

	createResource(t, deps, "delete1", "to delete")

	req := httptest.NewRequest("DELETE", "/api/test/delete1", nil)
	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if w.Body.String() != "" {
		t.Errorf("expected empty body when return_body is false, got: %s", w.Body.String())
	}
}

func testDELETEReturnBodyTrue(t *testing.T) {
	deps := setupUniHandlerFull(t)
	deps.config.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
		ReturnBody:  true,
	}

	createResource(t, deps, "delete2", "to delete")

	req := httptest.NewRequest("DELETE", "/api/test/delete2", nil)
	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if w.Body.String() != "{}" {
		t.Errorf("expected body '{}' when return_body is true, got: '%s'", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got: '%s'", w.Header().Get("Content-Type"))
	}
}

func createResource(t *testing.T, deps uniHandlerDeps, id, name string) {
	t.Helper()
	createBody := `{"id": "` + id + `", "name": "` + name + `"}`
	createReq := httptest.NewRequest("POST", "/api/test", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	deps.handler.ServeHTTP(createW, createReq)
}

func TestUniHandler_HEAD_Operation(t *testing.T) {
	deps := setupUniHandlerFull(t)

	// Create a test resource first
	createTestResourceForHead(t, deps, "test123", "test resource")

	tests := []headTestCase{
		{
			name:           "HEAD existing resource returns same headers as GET but no body",
			method:         "HEAD",
			path:           "/users/test123",
			expectedStatus: http.StatusOK,
			expectBody:     false,
			expectHeaders:  map[string]string{"Content-Type": "application/json"},
		},
		{
			name:           "GET existing resource returns headers and body",
			method:         "GET",
			path:           "/users/test123",
			expectedStatus: http.StatusOK,
			expectBody:     true,
			expectHeaders:  map[string]string{"Content-Type": "application/json"},
		},
		{
			name:           "HEAD non-existent resource returns 404 with no body",
			method:         "HEAD",
			path:           "/users/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectBody:     false,
		},
		{
			name:           "HEAD collection returns same headers as GET but no body",
			method:         "HEAD",
			path:           "/users",
			expectedStatus: http.StatusOK,
			expectBody:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executeHeadTest(t, deps, tt)
		})
	}
}

type headTestCase struct {
	name           string
	method         string
	path           string
	expectedStatus int
	expectBody     bool
	expectHeaders  map[string]string
}

func createTestResourceForHead(t *testing.T, deps uniHandlerDeps, id, name string) {
	t.Helper()
	createBody := `{"id": "` + id + `", "name": "` + name + `"}`
	createReq := httptest.NewRequest("POST", "/users", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	deps.handler.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("failed to create test resource, status: %d", createW.Code)
	}
}

func executeHeadTest(t *testing.T, deps uniHandlerDeps, tt headTestCase) {
	t.Helper()
	req := httptest.NewRequest(tt.method, tt.path, nil)
	w := httptest.NewRecorder()

	deps.handler.ServeHTTP(w, req)

	if w.Code != tt.expectedStatus {
		t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
	}

	checkExpectedHeaders(t, w, tt.expectHeaders)
	if tt.expectBody {
		checkBodyIsPresent(t, w)
	} else {
		checkBodyIsAbsent(t, w)
	}
}

func checkExpectedHeaders(t *testing.T, w *httptest.ResponseRecorder, expectHeaders map[string]string) {
	t.Helper()
	for key, expectedValue := range expectHeaders {
		if actualValue := w.Header().Get(key); actualValue != expectedValue {
			t.Errorf("expected header %s: %s, got: %s", key, expectedValue, actualValue)
		}
	}
}

func checkBodyIsPresent(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if len(body) == 0 {
		t.Error("expected body but got empty response")
	}
}

func checkBodyIsAbsent(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if len(body) != 0 {
		t.Errorf("expected empty body but got: %s", string(body))
	}
}

func TestUniHandler_HEAD_vs_GET_Consistency(t *testing.T) {
	deps := setupUniHandlerFull(t)

	// Create a test resource
	createTestResourceForHead(t, deps, "consistency123", "consistency test")

	// Make GET and HEAD requests
	getW := makeRequest(t, deps, "GET", "/users/consistency123")
	headW := makeRequest(t, deps, "HEAD", "/users/consistency123")

	// Verify consistency
	verifyStatusCodeConsistency(t, getW, headW)
	verifyHeaderConsistency(t, getW, headW)
	verifyBodyConsistency(t, getW, headW)
}

func makeRequest(t *testing.T, deps uniHandlerDeps, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)
	return w
}

func verifyStatusCodeConsistency(t *testing.T, getW, headW *httptest.ResponseRecorder) {
	t.Helper()
	if getW.Code != headW.Code {
		t.Errorf("status codes don't match: GET=%d, HEAD=%d", getW.Code, headW.Code)
	}
}

func verifyHeaderConsistency(t *testing.T, getW, headW *httptest.ResponseRecorder) {
	t.Helper()
	getHeaders := getW.Header()
	headHeaders := headW.Header()

	for key, getValues := range getHeaders {
		if key == "Content-Length" {
			continue // Skip Content-Length as it may differ
		}

		headValues := headHeaders[key]
		compareHeaderValues(t, key, getValues, headValues)
	}
}

func compareHeaderValues(t *testing.T, key string, getValues, headValues []string) {
	t.Helper()
	if len(headValues) != len(getValues) {
		t.Errorf("header %s value count mismatch: GET=%d, HEAD=%d", key, len(getValues), len(headValues))
		return
	}

	for i, getVal := range getValues {
		if headValues[i] != getVal {
			t.Errorf("header %s[%d] mismatch: GET=%s, HEAD=%s", key, i, getVal, headValues[i])
		}
	}
}

func verifyBodyConsistency(t *testing.T, getW, headW *httptest.ResponseRecorder) {
	t.Helper()
	getBody, _ := io.ReadAll(getW.Body)
	headBody, _ := io.ReadAll(headW.Body)

	if len(getBody) == 0 {
		t.Error("GET request should have body")
	}

	if len(headBody) != 0 {
		t.Errorf("HEAD request should have empty body, got: %s", string(headBody))
	}
}
