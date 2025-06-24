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
	"github.com/google/uuid"
)

func TestMockHandler_HandleRequest(t *testing.T) {
	deps := setupMockHandlerFull(t)
	testData := getTestData()
	populateTestData(deps.store, testData)

	tests := getHandlerTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := prepareHandlerForTest(t, tt, deps, testData)
			executeTestRequest(t, tt, mockHandler)
		})
	}
}

// validatePostUUIDResponse validates POST response with UUID generation
func validatePostUUIDResponse(t *testing.T, w *httptest.ResponseRecorder, path string) {
	t.Helper()
	location := w.Header().Get("Location")
	if location == "" {
		t.Errorf("expected Location header to be set")
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

// mockHandlerDeps holds the dependencies for mock handler testing
type mockHandlerDeps struct {
	handler *handler.MockHandler
	store   storage.MockStorage
	config  *config.MockConfig
	logger  *slog.Logger
}

// setupMockHandlerFull creates and configures a mock handler with all dependencies for testing
func setupMockHandlerFull(t *testing.T) mockHandlerDeps {
	t.Helper()
	store := storage.NewMockStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/*",
				BodyIDPaths:   []string{"/id"},
				CaseSensitive: true,
				ReturnBody:    true, // Maintain backward compatibility for existing tests
			},
		},
	}

	mockService := service.NewMockService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	mockHandler := handler.NewMockHandler(mockService, scenarioService, logger, cfg)
	
	return mockHandlerDeps{
		handler: mockHandler,
		store:   store,
		config:  cfg,
		logger:  logger,
	}
}

// getTestData returns the standard test data
func getTestData() []*model.MockData {
	return []*model.MockData{
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
func populateTestData(store storage.MockStorage, testData []*model.MockData) {
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
}, deps mockHandlerDeps, testData []*model.MockData) *handler.MockHandler {
	t.Helper()

	if needsCleanHandler(tt.name) {
		cleanStore := storage.NewMockStorage()
		populateTestData(cleanStore, testData)
		currentService := service.NewMockService(cleanStore, deps.config)
		currentScenarioStore := storage.NewScenarioStorage()
		currentScenarioService := service.NewScenarioService(currentScenarioStore)
		return handler.NewMockHandler(currentService, currentScenarioService, deps.logger, deps.config)
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
}, mockHandler *handler.MockHandler) {
	t.Helper()

	req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
	if tt.contentType != "" {
		req.Header.Set("Content-Type", tt.contentType)
	}

	w := httptest.NewRecorder()
	mockHandler.ServeHTTP(w, req)

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

func TestMockHandler_ReturnBodyFlag_POST(t *testing.T) {
	t.Run("POST with return_body false", testPOSTReturnBodyFalse)
	t.Run("POST with return_body true", testPOSTReturnBodyTrue)
}

func TestMockHandler_ReturnBodyFlag_PUT(t *testing.T) {
	t.Run("PUT with return_body false", testPUTReturnBodyFalse)
	t.Run("PUT with return_body true", testPUTReturnBodyTrue)
}

func TestMockHandler_ReturnBodyFlag_DELETE(t *testing.T) {
	t.Run("DELETE with return_body false", testDELETEReturnBodyFalse)
	t.Run("DELETE with return_body true", testDELETEReturnBodyTrue)
}

func testPOSTReturnBodyFalse(t *testing.T) {
	deps := setupMockHandlerFull(t)
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
	deps := setupMockHandlerFull(t)
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
	deps := setupMockHandlerFull(t)
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
	deps := setupMockHandlerFull(t)
	deps.config.Sections["test"] = config.Section{
		PathPattern: "/api/test/*",
		BodyIDPaths: []string{"/id"},
		ReturnBody:  true,
	}
	
	createResource(t, deps, "999", "initial")
	
	updateBody := `{"id": "999", "name": "updated value"}`
	req := httptest.NewRequest("PUT", "/api/test/999", strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	deps.handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() == "" {
		t.Error("expected non-empty body when return_body is true")
	}
	if !strings.Contains(w.Body.String(), "updated value") {
		t.Errorf("expected body to contain updated data, got: %s", w.Body.String())
	}
}

func testDELETEReturnBodyFalse(t *testing.T) {
	deps := setupMockHandlerFull(t)
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
	deps := setupMockHandlerFull(t)
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

func createResource(t *testing.T, deps mockHandlerDeps, id, name string) {
	t.Helper()
	createBody := `{"id": "` + id + `", "name": "` + name + `"}`
	createReq := httptest.NewRequest("POST", "/api/test", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	deps.handler.ServeHTTP(createW, createReq)
}