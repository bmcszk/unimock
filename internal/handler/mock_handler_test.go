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
	handler := NewMockHandler(mockService, logger, cfg)

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
			expectedBody:   "resource not found",
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
			expectedBody:   "resource not found",
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
			expectedBody:   "invalid JSON",
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
			expectedStatus: http.StatusNotFound,
			expectedBody:   "resource not found",
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
			expectedBody:   "resource not found",
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
				handler = NewMockHandler(currentService, logger, cfg)     // re-assign handler to the one in the outer scope
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := service.NewMockService(store, cfg)
	handler := NewMockHandler(mockService, logger, cfg)
	ctx := context.Background()

	tests := []struct {
		name          string
		method        string
		path          string
		headers       map[string]string
		body          string
		contentType   string
		expectedIDs   []string
		expectError   bool
		errorContains string
	}{
		{
			name:        "GET with ID in path",
			method:      http.MethodGet,
			path:        "/users/123",
			expectedIDs: []string{"123"},
		},
		{
			name:          "GET with case-sensitive path",
			method:        http.MethodGet,
			path:          "/Users/123",
			expectedIDs:   nil,
			expectError:   true,
			errorContains: "invalid path",
		},
		{
			name:        "GET with case-insensitive path",
			method:      http.MethodGet,
			path:        "/Orders/123",
			expectedIDs: []string{"123"},
		},
		{
			name:        "GET collection path",
			method:      http.MethodGet,
			path:        "/users",
			expectedIDs: nil,
		},
		{
			name:        "POST with ID in header",
			method:      http.MethodPost,
			path:        "/users",
			headers:     map[string]string{"X-Resource-ID": "123"},
			expectedIDs: []string{"123"},
		},
		{
			name:        "POST with ID in JSON body",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/json",
			body:        `{"id": "123", "name": "test"}`,
			expectedIDs: []string{"123"},
		},
		{
			name:        "POST with ID in nested JSON body",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/json",
			body:        `{"data": {"id": "123"}, "name": "test"}`,
			expectedIDs: []string{"123"},
		},
		{
			name:        "POST with ID in XML body",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/xml",
			body:        `<user><id>123</id><name>test</name></user>`,
			expectedIDs: []string{"123"},
		},
		{
			name:        "POST with ID in nested XML body",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/xml",
			body:        `<user><data><id>123</id></data><name>test</name></user>`,
			expectedIDs: []string{"123"},
		},
		{
			name:          "POST with missing ID",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "application/json",
			body:          `{"name": "test"}`,
			expectError:   true,
			errorContains: "no IDs found in request",
		},
		{
			name:          "POST with malformed JSON",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "application/json",
			body:          `{"id": "123", "name": "test"`,
			expectError:   true,
			errorContains: "invalid JSON",
		},
		{
			name:          "POST with malformed XML",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "application/xml",
			body:          `<user><id>123</id><name>test</user>`,
			expectError:   true,
			errorContains: "invalid XML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ids, err := handler.extractIDs(ctx, req)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				if tt.errorContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errorContains)) {
					t.Errorf("expected error message to contain %q, got %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if !reflect.DeepEqual(ids, tt.expectedIDs) {
					t.Errorf("expected IDs to be %v, got %v", tt.expectedIDs, ids)
				}
			}
		})
	}
}
