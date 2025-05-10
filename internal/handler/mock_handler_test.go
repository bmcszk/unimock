package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestMockHandler_HandleRequest(t *testing.T) {
	// Setup
	store := storage.NewMockStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
				BodyIDPaths: []string{"/id"},
			},
		},
	}

	// Create service and handler
	mockService := service.NewMockService(store, cfg)
	handler := NewMockHandler(mockService, logger, cfg)

	// Test data
	testData := []*model.MockData{
		{
			Path:        "/users/123",
			ContentType: "application/json",
			Body:        []byte(`{"id": "123", "name": "test"}`),
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
			name:           "GET non-existent resource",
			method:         http.MethodGet,
			path:           "/users/999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "resource not found",
		},
		{
			name:           "POST new resource",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "application/json",
			body:           `{"id": "456", "name": "new"}`,
			expectedStatus: http.StatusCreated,
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
			name:           "DELETE existing resource",
			method:         http.MethodDelete,
			path:           "/users/123",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check body if specified
			if tt.expectedBody != "" {
				if !strings.Contains(w.Body.String(), tt.expectedBody) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

// Added from mock_service_test.go
func TestMockHandler_ExtractIDs(t *testing.T) {
	// Create test config
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:  "/users/*",
				HeaderIDName: "X-Resource-ID",
				BodyIDPaths: []string{
					"/id",
					"/data/id",
					"//id",
				},
			},
			"orders": {
				PathPattern:  "/orders/*",
				HeaderIDName: "X-Order-ID",
				BodyIDPaths: []string{
					"/orderId",
					"/order/id",
				},
			},
		},
	}

	// Create service and handler
	store := storage.NewMockStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := service.NewMockService(store, cfg)
	handler := NewMockHandler(mockService, logger, cfg)

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
		// More test cases from the original mock_service_test.go
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Extract IDs
			ids, err := handler.ExtractIDs(context.Background(), req)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check IDs
			if len(ids) != len(tt.expectedIDs) {
				t.Errorf("got %d IDs, want %d", len(ids), len(tt.expectedIDs))
				return
			}

			for i, id := range ids {
				if id != tt.expectedIDs[i] {
					t.Errorf("ID[%d] = %q, want %q", i, id, tt.expectedIDs[i])
				}
			}
		})
	}
}

// Helper functions
func jsonEqual(a, b interface{}) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aJSON, bJSON)
}

func createMockData(path string, body []byte) *model.MockData {
	return &model.MockData{
		Path:        path,
		ContentType: "application/json",
		Body:        body,
	}
}
