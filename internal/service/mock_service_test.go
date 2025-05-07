package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/internal/model"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
)

func TestMockService_ExtractIDs(t *testing.T) {
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

	// Create service
	store := storage.NewMockStorage()
	service := NewMockService(store, cfg)

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
		{
			name:        "POST with multiple IDs in JSON body",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/json",
			body:        `{"id": "123", "data": {"id": "456"}, "name": "test"}`,
			expectedIDs: []string{"123", "456"},
		},
		{
			name:        "PUT with ID in path and header",
			method:      http.MethodPut,
			path:        "/users/123",
			headers:     map[string]string{"X-Resource-ID": "456"},
			expectedIDs: []string{"123"}, // Path ID takes precedence
		},
		{
			name:        "DELETE with ID in path",
			method:      http.MethodDelete,
			path:        "/users/123",
			expectedIDs: []string{"123"},
		},
		{
			name:        "Different section with ID in path",
			method:      http.MethodGet,
			path:        "/orders/789",
			expectedIDs: []string{"789"},
		},
		{
			name:        "Different section with ID in header",
			method:      http.MethodPost,
			path:        "/orders",
			headers:     map[string]string{"X-Order-ID": "789"},
			expectedIDs: []string{"789"},
		},
		{
			name:          "invalid path pattern",
			method:        http.MethodGet,
			path:          "/invalid/path",
			expectError:   true,
			errorContains: "no matching section found",
		},
		{
			name:          "invalid JSON body",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "application/json",
			body:          `{"id": "123", "name": "test"`, // Missing closing brace
			expectError:   true,
			errorContains: "failed to parse JSON body",
		},
		{
			name:          "invalid XML body",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "application/xml",
			body:          `<user><id>123</id><name>test</user>`, // Missing closing tag
			expectError:   true,
			errorContains: "failed to parse XML body",
		},
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
			ids, err := service.ExtractIDs(context.Background(), req)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errorContains)) {
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

func TestMockService_HandleRequest(t *testing.T) {
	// Create test config
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
				BodyIDPaths: []string{"/id"},
			},
			"orders": {
				PathPattern: "/orders/*",
			},
		},
	}

	// Create service
	store := storage.NewMockStorage()
	service := NewMockService(store, cfg)

	// Setup test data
	testData := []*model.MockData{
		{
			Path:        "/users/123",
			ContentType: "application/json",
			Body:        []byte(`{"id": "123", "name": "test"}`),
		},
		{
			Path:        "/orders/456",
			ContentType: "application/json",
			Body:        []byte(`{"id": "456", "status": "pending"}`),
		},
	}

	for _, data := range testData {
		ids := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
		err := store.Create(ids, data)
		if err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		expectedStatus int
		expectedBody   string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "GET existing resource",
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Path":"/users/123","Location":"/users/123/123","ContentType":"application/json","Body":"eyJpZCI6ICIxMjMiLCAibmFtZSI6ICJ0ZXN0In0="}`,
		},
		{
			name:          "GET non-existent resource",
			method:        http.MethodGet,
			path:          "/users/999", // Using ID that doesn't exist
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name:           "GET collection",
			method:         http.MethodGet,
			path:           "/users",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"Path":"/users/123","Location":"/users/123/123","ContentType":"application/json","Body":"eyJpZCI6ICIxMjMiLCAibmFtZSI6ICJ0ZXN0In0="}]`,
		},
		{
			name:           "POST new resource",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "application/json",
			body:           `{"id": "789", "name": "new"}`, // Use ID that doesn't exist yet
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"Path":"/users","Location":"/users/789","ContentType":"application/json","Body":"eyJpZCI6ICI3ODkiLCAibmFtZSI6ICJuZXcifQ=="}`,
		},
		{
			name:           "PUT existing resource",
			method:         http.MethodPut,
			path:           "/users/123",
			contentType:    "application/json",
			body:           `{"id": "123", "name": "updated"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Path":"/users/123","Location":"","ContentType":"application/json","Body":"eyJpZCI6ICIxMjMiLCAibmFtZSI6ICJ1cGRhdGVkIn0="}`,
		},
		{
			name:          "PUT non-existent resource",
			method:        http.MethodPut,
			path:          "/users/999", // Using ID that doesn't exist
			contentType:   "application/json",
			body:          `{"id": "999", "name": "new"}`,
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name:           "DELETE existing resource",
			method:         http.MethodDelete,
			path:           "/users/123",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:          "DELETE non-existent resource",
			method:        http.MethodDelete,
			path:          "/users/999", // Using ID that doesn't exist
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name:          "invalid method",
			method:        "INVALID",
			path:          "/users/123",
			expectError:   true,
			errorContains: "method INVALID not allowed",
		},
		{
			name:          "invalid content type",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "invalid/type",
			body:          `{"id": "123"}`,
			expectError:   true,
			errorContains: "unsupported content type",
		},
		{
			name:          "invalid JSON body",
			method:        http.MethodPost,
			path:          "/users",
			contentType:   "application/json",
			body:          `{"id": "123", "name": "test"`, // Missing closing brace
			expectError:   true,
			errorContains: "failed to parse JSON body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Handle request
			resp, err := service.HandleRequest(context.Background(), req)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errorContains)) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check response
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status code = %d, want %d", resp.StatusCode, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read response body: %v", err)
					return
				}
				if string(body) != tt.expectedBody {
					t.Errorf("body = %q, want %q", string(body), tt.expectedBody)
				}
			}
		})
	}
}
