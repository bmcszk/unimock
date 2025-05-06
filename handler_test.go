package main

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
	"time"
)

// Helper function to create a request with timeout
func createRequest(t *testing.T, method, path string, body string) *http.Request {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req = req.WithContext(ctx)
	t.Cleanup(func() {
		cancel()
	})
	return req
}

func TestHandler_ExtractIDs(t *testing.T) {
	storage := NewStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewHandler(storage, []string{"//id", "//@id"}, "X-Resource-ID", logger)

	tests := []struct {
		name        string
		method      string
		contentType string
		headerID    string
		body        string
		path        string
		expectedIDs []string
		expectError bool
	}{
		{
			name:        "JSON with ID in body",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id": "123", "name": "test"}`,
			path:        "/users",
			expectedIDs: []string{"123"},
		},
		{
			name:        "JSON with ID in header",
			method:      http.MethodPost,
			contentType: "application/json",
			headerID:    "123",
			body:        `{"name": "test"}`,
			path:        "/users",
			expectedIDs: []string{"123"},
		},
		{
			name:        "JSON without ID",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"name": "test"}`,
			path:        "/users",
			expectError: true,
		},
		{
			name:        "XML with ID in body",
			method:      http.MethodPost,
			contentType: "application/xml",
			body:        `<root><id>123</id></root>`,
			path:        "/users",
			expectedIDs: []string{"123"},
		},
		{
			name:        "XML without ID - use path",
			method:      http.MethodPost,
			contentType: "application/xml",
			body:        `<root><name>test</name></root>`,
			path:        "/users/123",
			expectedIDs: []string{"123"},
		},
		{
			name:        "Deep path without ID",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        `plain text`,
			path:        "/users/123/orders/456",
			expectedIDs: []string{"456"},
		},
		{
			name:        "Collection path without ID",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        `plain text`,
			path:        "/users",
			expectedIDs: []string{},
		},
		{
			name:        "GET with ID in path",
			method:      http.MethodGet,
			contentType: "application/json",
			path:        "/users/123",
			expectedIDs: []string{"123"},
		},
		{
			name:        "GET collection path",
			method:      http.MethodGet,
			contentType: "application/json",
			path:        "/users",
			expectedIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			if tt.headerID != "" {
				req.Header.Set("X-Resource-ID", tt.headerID)
			}

			ids, err := handler.extractIDs(req)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("extractIDs() error = %v", err)
				return
			}

			if len(ids) != len(tt.expectedIDs) {
				t.Errorf("Expected %d IDs, got %d", len(tt.expectedIDs), len(ids))
				return
			}

			// Create a map for easier comparison
			idMap := make(map[string]bool)
			for _, id := range ids {
				idMap[id] = true
			}

			for _, expectedID := range tt.expectedIDs {
				if !idMap[expectedID] {
					t.Errorf("Expected ID %s not found in %v", expectedID, ids)
				}
			}
		})
	}
}

func TestHandler_HandleRequest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		body           string
		expectedStatus int
		expectedBody   string
		checkHeaders   map[string]string
		isXML          bool
		needsSetup     bool
		setupData      []struct {
			path string
			body string
		}
	}{
		{
			name:           "GET non-existent resource",
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusNotFound,
			needsSetup:     false,
		},
		{
			name:           "POST new resource to collection",
			method:         http.MethodPost,
			path:           "/users",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"id": "123", "name": "test"}`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id": "123", "name": "test"}`,
			checkHeaders: map[string]string{
				"Location":     "/users/123",
				"Content-Type": "application/json",
			},
			needsSetup: false,
		},
		{
			name:           "POST new resource with deep path",
			method:         http.MethodPost,
			path:           "/users/123/orders",
			headers:        map[string]string{"Content-Type": "application/xml"},
			body:           `<order><n>test<n></order>`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `<order><n>test<n></order>`,
			checkHeaders: map[string]string{
				"Location":     "/users/123/orders/456",
				"Content-Type": "application/xml",
			},
			isXML:      true,
			needsSetup: false,
		},
		{
			name:           "GET existing resource",
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id": "123", "name": "test"}`,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "test"}`,
				},
			},
		},
		{
			name:           "GET collection",
			method:         http.MethodGet,
			path:           "/users",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id": "123", "name": "test"}]`,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "test"}`,
				},
			},
		},
		{
			name:           "POST duplicate resource",
			method:         http.MethodPost,
			path:           "/users",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"id": "123", "name": "test2"}`,
			expectedStatus: http.StatusConflict,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "test"}`,
				},
			},
		},
		{
			name:           "PUT non-existent resource",
			method:         http.MethodPut,
			path:           "/users/456",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"id": "456", "name": "test3"}`,
			expectedStatus: http.StatusNotFound,
			needsSetup:     false,
		},
		{
			name:           "PUT existing resource",
			method:         http.MethodPut,
			path:           "/users/123",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"id": "123", "name": "updated"}`,
			expectedStatus: http.StatusOK,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "test"}`,
				},
			},
		},
		{
			name:           "GET updated resource",
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id": "123", "name": "updated"}`,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "updated"}`,
				},
			},
		},
		{
			name:           "DELETE non-existent resource",
			method:         http.MethodDelete,
			path:           "/users/456",
			expectedStatus: http.StatusNotFound,
			needsSetup:     false,
		},
		{
			name:           "DELETE existing resource",
			method:         http.MethodDelete,
			path:           "/users/123",
			expectedStatus: http.StatusNoContent,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "test"}`,
				},
			},
		},
		{
			name:           "GET deleted resource",
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusNotFound,
			needsSetup:     false,
		},
		{
			name:           "PUT with ID in path",
			method:         http.MethodPut,
			path:           "/users/123",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"name": "updated"}`,
			expectedStatus: http.StatusOK,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "test"}`,
				},
			},
		},
		{
			name:           "PUT with deep path",
			method:         http.MethodPut,
			path:           "/users/123/orders/456",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"status": "updated"}`,
			expectedStatus: http.StatusOK,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123/orders/456",
					body: `{"id": "456", "status": "pending"}`,
				},
			},
		},
		{
			name:           "DELETE by ID",
			method:         http.MethodDelete,
			path:           "/users/123",
			expectedStatus: http.StatusNoContent,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123",
					body: `{"id": "123", "name": "test"}`,
				},
			},
		},
		{
			name:           "DELETE by path fallback",
			method:         http.MethodDelete,
			path:           "/users/123/orders",
			expectedStatus: http.StatusNoContent,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/users/123/orders/1",
					body: `{"id": "1", "status": "pending"}`,
				},
				{
					path: "/users/123/orders/2",
					body: `{"id": "2", "status": "completed"}`,
				},
			},
		},
		{
			name:           "DELETE non-existent by ID and path",
			method:         http.MethodDelete,
			path:           "/users/999",
			expectedStatus: http.StatusNotFound,
			needsSetup:     false,
		},
		{
			name:           "POST JSON with ID in body and verify retrieval",
			method:         http.MethodPost,
			path:           "/test",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"id": "123", "name": "test", "value": 42}`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id": "123", "name": "test", "value": 42}`,
			checkHeaders: map[string]string{
				"Location":     "/test/123",
				"Content-Type": "application/json",
			},
			needsSetup: false,
		},
		{
			name:           "GET JSON with ID in body after creation",
			method:         http.MethodGet,
			path:           "/test/123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id": "123", "name": "test", "value": 42}`,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/test/123",
					body: `{"id": "123", "name": "test", "value": 42}`,
				},
			},
		},
		{
			name:           "GET collection with JSON items",
			method:         http.MethodGet,
			path:           "/test",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id": "123", "name": "test", "value": 42}]`,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/test/123",
					body: `{"id": "123", "name": "test", "value": 42}`,
				},
			},
		},
		{
			name:           "Delete one of multiple elements",
			method:         http.MethodDelete,
			path:           "/test/123",
			expectedStatus: http.StatusNoContent,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/test/123",
					body: `{"id": "123", "name": "test1", "value": 42}`,
				},
				{
					path: "/test/456",
					body: `{"id": "456", "name": "test2", "value": 43}`,
				},
			},
		},
		{
			name:           "Verify remaining element after deletion",
			method:         http.MethodGet,
			path:           "/test",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id": "456", "name": "test2", "value": 43}]`,
			needsSetup:     true,
			setupData: []struct {
				path string
				body string
			}{
				{
					path: "/test/456",
					body: `{"id": "456", "name": "test2", "value": 43}`,
				},
			},
		},
		{
			name:           "Verify deleted element is not accessible",
			method:         http.MethodGet,
			path:           "/test/123",
			expectedStatus: http.StatusNotFound,
			needsSetup:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new storage for each test case
			storage := NewStorage()
			handler := NewHandler(storage, []string{"//id"}, "X-Resource-ID", logger)

			// For tests that require existing data, set it up
			if tt.needsSetup {
				for _, data := range tt.setupData {
					// Extract IDs from the path
					pathSegments := strings.Split(strings.Trim(data.path, "/"), "/")
					var ids []string
					if len(pathSegments) > 0 {
						lastSegment := pathSegments[len(pathSegments)-1]
						if _, err := json.Marshal(lastSegment); err == nil {
							ids = append(ids, lastSegment)
						}
					}

					// Store the data
					if err := storage.Create(ids, &MockData{
						Path:        data.path,
						ContentType: "application/json",
						Body:        []byte(data.body),
					}); err != nil {
						t.Fatalf("Failed to setup test data: %v", err)
					}
				}
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				if tt.isXML {
					if w.Body.String() != tt.expectedBody {
						t.Errorf("expected body %s, got %s", tt.expectedBody, w.Body.String())
					}
				} else {
					var expected, actual interface{}
					if err := json.Unmarshal([]byte(tt.expectedBody), &expected); err != nil {
						t.Fatalf("failed to unmarshal expected body: %v", err)
					}
					if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
						t.Fatalf("failed to unmarshal actual body: %v", err)
					}
					if !jsonEqual(expected, actual) {
						t.Errorf("expected body %s, got %s", tt.expectedBody, w.Body.String())
					}
				}
			}

			for k, v := range tt.checkHeaders {
				if w.Header().Get(k) != v {
					t.Errorf("expected header %s=%s, got %s", k, v, w.Header().Get(k))
				}
			}

			// For DELETE tests, verify the data was actually deleted
			if tt.method == http.MethodDelete && tt.expectedStatus == http.StatusNoContent {
				ids, err := handler.extractIDs(httptest.NewRequest(http.MethodGet, tt.path, nil))
				if err != nil {
					t.Fatalf("Failed to extract IDs for verification: %v", err)
				}
				if len(ids) > 0 {
					if _, err := storage.Get(ids[0]); err == nil {
						t.Error("Expected resource to be deleted but it still exists")
					}
				}
				// Check path-based deletion
				items, err := storage.GetByPath(tt.path)
				if err == nil && len(items) > 0 {
					t.Error("Expected path resources to be deleted but they still exist")
				}
			}
		})
	}
}

func jsonEqual(a, b interface{}) bool {
	aj, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bj, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aj, bj)
}
