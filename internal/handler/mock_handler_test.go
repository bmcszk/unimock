package handler

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/google/uuid"
)

func TestMockHandler_HandleRequest(t *testing.T) {
	// Setup
	store := storage.NewMockStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/*",
				BodyIDPaths:   []string{"/id"},
				CaseSensitive: true,
			},
		},
	}

	// Create service and handler
	mockService := service.NewMockService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	handler := NewMockHandler(mockService, scenarioService, logger, cfg)

	// Test data with mixed content types
	testData := []*model.MockData{
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

	// Store test data
	for _, data := range testData {
		ids := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
		store.Create(ids, data)
	}

	tests := []struct {
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
			expectedBody:   "invalid request: invalid JSON",
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special handling for tests that modify state or depend on a clean slate
			if tt.name == "POST new resource with no ID (auto-generate UUID)" ||
				tt.name == "POST new resource" ||
				tt.name == "PUT non-existent resource" ||
				tt.name == "DELETE non-existent resource but collection exists" {

				cleanStore := storage.NewMockStorage()
				// Repopulate with baseline testData for other tests that might expect it
				// This is a simple way to reset state for this test suite structure.
				// Ideally, each t.Run would set up its own specific required data.
				for _, data := range testData { // testData is from the outer scope
					initialIds := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
					cleanStore.Create(initialIds, data)
				}
				currentService := service.NewMockService(cleanStore, cfg) // cfg is from outer scope
				// Create a new scenarioService for this specific test scope too, as handler is reassigned
				currentScenarioStore := storage.NewScenarioStorage()
				currentScenarioService := service.NewScenarioService(currentScenarioStore)
				handler = NewMockHandler(currentService, currentScenarioService, logger, cfg) // re-assign handler to the one in the outer scope
			}

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.name == "POST new resource with no ID (auto-generate UUID)" {
				location := w.Header().Get("Location")
				if location == "" {
					t.Errorf("expected Location header to be set")
				}
				if !strings.HasPrefix(location, tt.path+"/") {
					t.Errorf("expected Location header to start with %s/, got %s", tt.path, location)
				}
				extractedID := location[len(tt.path)+1:]
				_, err := uuid.Parse(extractedID)
				if err != nil {
					t.Errorf("expected Location header to contain a valid UUID, got %s, error: %v", extractedID, err)
				}
			} else if tt.expectedBody != "" {
				if !strings.EqualFold(w.Body.String(), tt.expectedBody) {
					t.Errorf("expected body to match %q, got %q", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

// Added from mock_service_test.go
func TestMockHandler_ExtractIDs(t *testing.T) {
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/*",
				HeaderIDName:  "X-Resource-ID",
				BodyIDPaths:   []string{"/id", "/data/id", "//id"},
				CaseSensitive: true,
			},
			"orders": {
				PathPattern:   "/orders/*",
				HeaderIDName:  "X-Order-ID",
				BodyIDPaths:   []string{"/orderId", "/order/id"},
				CaseSensitive: false,
			},
		},
	}

	store := storage.NewMockStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := service.NewMockService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	handler := NewMockHandler(mockService, scenarioService, logger, cfg)
	ctx := context.Background()

	tests := []struct {
		name          string
		method        string
		path          string
		headerKey     string
		headerValue   string
		body          string
		contentType   string
		expectedIDs   []string
		expectedError string
	}{
		// Path extraction tests
		{
			name:        "Extract ID from path - GET",
			method:      http.MethodGet,
			path:        "/users/123",
			expectedIDs: []string{"123"},
		},
		{
			name:        "Extract ID from path - PUT",
			method:      http.MethodPut,
			path:        "/users/abc",
			expectedIDs: []string{"abc"},
		},
		{
			name:        "Extract ID from path - DELETE",
			method:      http.MethodDelete,
			path:        "/orders/xyz",
			expectedIDs: []string{"xyz"},
		},
		{
			name:   "No ID in path for GET (collection)",
			method: http.MethodGet,
			path:   "/users",
			// No error, but nil IDs expected for collection GET
		},
		// Header extraction tests (for POST)
		{
			name:        "Extract ID from header - POST",
			method:      http.MethodPost,
			path:        "/users",
			headerKey:   "X-Resource-ID",
			headerValue: "header-id-1",
			expectedIDs: []string{"header-id-1"},
		},
		// Body extraction tests (for POST)
		{
			name:        "Extract ID from JSON body - POST",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/json",
			body:        `{"id": "json-id-1"}`,
			expectedIDs: []string{"json-id-1"},
		},
		{
			name:        "Extract multiple IDs from JSON body - POST",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/json",
			body:        `[{"id": "json-id-1"}, {"id": "json-id-2"}]`, // Assuming BodyIDPaths is like "//id"
			expectedIDs: []string{"json-id-1", "json-id-2"},
		},
		{
			name:        "Extract ID from XML body - POST",
			method:      http.MethodPost,
			path:        "/users", // uses "users" section config
			contentType: "application/xml",
			body:        "<data><id>xml-id-1</id></data>",
			expectedIDs: []string{"xml-id-1"},
		},
		// Fallback and error tests
		{
			name:          "No ID found - POST",
			method:        http.MethodPost,
			path:          "/users",
			expectedError: "no IDs found in request",
		},
		{
			name:          "Invalid JSON - POST",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "application/json",
			body:          `{"id": "bad-json`,
			expectedError: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerValue)
			}
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Determine sectionName and section for the current test case path
			var currentSection *config.Section
			var currentSectionName string
			var err error

			// Simplified section matching for test setup
			// In real code, handler.getSectionForRequest is used.
			pathForSectionLookup := strings.Trim(tt.path, "/")
			if strings.Contains(pathForSectionLookup, "/") { // e.g. /users/123 -> users
				currentSectionName = pathForSectionLookup[:strings.Index(pathForSectionLookup, "/")]
			} else { // e.g. /users -> users
				currentSectionName = pathForSectionLookup
			}

			sec, ok := cfg.Sections[currentSectionName]
			if !ok {
				// Try matching with path pattern logic if direct name lookup fails (e.g. for /orders/xyz)
				for name, s := range cfg.Sections {
					trimmedPattern := strings.Trim(s.PathPattern, "/")
					basePattern := trimmedPattern
					if strings.Contains(trimmedPattern, "/*") {
						basePattern = trimmedPattern[:strings.Index(trimmedPattern, "/*")]
					}
					if currentSectionName == basePattern || strings.HasPrefix(currentSectionName, basePattern+"/") {
						currentSection = &s
						currentSectionName = name // Use the defined section name
						break
					}
				}
				if currentSection == nil {
					t.Fatalf("Test setup error: could not find section for path %s (derived section name: %s)", tt.path, currentSectionName)
				}
			} else {
				currentSection = &sec
			}

			ids, err := handler.extractIDs(ctx, req, currentSection, currentSectionName)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error to contain %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(ids, tt.expectedIDs) {
					t.Errorf("expected IDs %v, got %v", tt.expectedIDs, ids)
				}
			}
		})
	}
}
