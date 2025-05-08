package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
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
		{
			name:        "POST with ID in custom JSON content type",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "text/json",
			body:        `{"id": "123", "name": "test"}`,
			expectedIDs: []string{"123"},
		},
		{
			name:        "POST with ID in custom XML content type",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "text/xml",
			body:        `<user><id>123</id><name>test</name></user>`,
			expectedIDs: []string{"123"},
		},
		{
			name:        "POST with ID in vendor JSON content type",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/vnd.company.json",
			body:        `{"id": "123", "name": "test"}`,
			expectedIDs: []string{"123"},
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
			"products": {
				PathPattern: "/products/*",
				BodyIDPaths: []string{"/id"},
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
		// Mixed content type test data
		{
			Path:        "/products/101",
			ContentType: "application/json",
			Body:        []byte(`{"id": "101", "name": "Product A", "price": 29.99}`),
		},
		{
			Path:        "/products/102",
			ContentType: "application/xml",
			Body:        []byte(`<product><id>102</id><name>Product B</name><price>49.99</price></product>`),
		},
		{
			Path:        "/products/103",
			ContentType: "application/json",
			Body:        []byte(`{"id": "103", "name": "Product C", "price": 19.99}`),
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
			expectedBody:   `{"id": "123", "name": "test"}`,
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
			expectedBody:   `[{"id": "123", "name": "test"}]`,
		},
		{
			name:           "POST new resource",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "application/json",
			body:           `{"id": "789", "name": "new"}`, // Use ID that doesn't exist yet
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id": "789", "name": "new"}`,
		},
		{
			name:           "PUT existing resource",
			method:         http.MethodPut,
			path:           "/users/123",
			contentType:    "application/json",
			body:           `{"id": "123", "name": "updated"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id": "123", "name": "updated"}`,
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
		{
			name:           "POST with custom JSON content type",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "text/json",
			body:           `{"id": "text-json-456", "name": "custom json"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "POST with custom XML content type",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "text/xml",
			body:           `<user><id>text-xml-457</id><name>custom xml</name></user>`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "POST with vendor JSON content type",
			method:         http.MethodPost,
			path:           "/users",
			contentType:    "application/vnd.company.json+v1",
			body:           `{"id": "vnd-json-458", "name": "vendor json"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "GET collection with mixed content types",
			method:         http.MethodGet,
			path:           "/products",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id": "101", "name": "Product A", "price": 29.99},{"id": "103", "name": "Product C", "price": 19.99}]`,
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

				// For JSON responses, compare with more flexibility
				if strings.Contains(resp.Header.Get("Content-Type"), "json") {
					// Parse expected and actual JSON
					var expectedJSON, actualJSON interface{}
					if err := json.Unmarshal([]byte(tt.expectedBody), &expectedJSON); err != nil {
						t.Errorf("invalid expected JSON: %v", err)
						return
					}
					if err := json.Unmarshal(body, &actualJSON); err != nil {
						t.Errorf("invalid actual JSON: %v", err)
						return
					}

					// Compare the parsed JSON
					if !reflect.DeepEqual(expectedJSON, actualJSON) {
						t.Errorf("body JSON = %v, want %v", actualJSON, expectedJSON)
					}
				} else {
					// For non-JSON responses, do a direct string comparison
					if string(body) != tt.expectedBody {
						t.Errorf("body = %q, want %q", string(body), tt.expectedBody)
					}
				}
			}
		})
	}
}

// TestMockService_GetCollection tests the collection retrieval functionality with various scenarios
func TestMockService_GetCollection(t *testing.T) {
	// Create test config
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
				BodyIDPaths: []string{"/id"},
			},
			"products": {
				PathPattern: "/products/*",
				BodyIDPaths: []string{"/id"},
			},
			"orders": {
				PathPattern: "/orders/*",
				BodyIDPaths: []string{"/id"},
			},
			"articles": {
				PathPattern: "/articles/*",
				BodyIDPaths: []string{"/id"},
			},
		},
	}

	// Add a new section to the config for empty testing
	cfg.Sections["categories"] = config.Section{
		PathPattern: "/categories/*",
		BodyIDPaths: []string{"/id"},
	}

	// Create service
	store := storage.NewMockStorage()
	service := NewMockService(store, cfg)

	// Test scenario 1: Empty collection
	// No data for /articles path - should return empty array

	// Test scenario 2: Collection with only XML content
	xmlData := []*model.MockData{
		{
			Path:        "/orders/x1",
			ContentType: "application/xml",
			Body:        []byte(`<order><id>x1</id><customer>Alice</customer><total>50.00</total></order>`),
		},
		{
			Path:        "/orders/x2",
			ContentType: "application/xml",
			Body:        []byte(`<order><id>x2</id><customer>Bob</customer><total>75.00</total></order>`),
		},
	}
	for _, data := range xmlData {
		ids := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
		if err := store.Create(ids, data); err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	// Test scenario 3: Collection with multiple JSON items
	jsonData := []*model.MockData{
		{
			Path:        "/users/101",
			ContentType: "application/json",
			Body:        []byte(`{"id": "101", "name": "John Doe", "email": "john@example.com"}`),
		},
		{
			Path:        "/users/102",
			ContentType: "application/json",
			Body:        []byte(`{"id": "102", "name": "Jane Smith", "email": "jane@example.com"}`),
		},
		{
			Path:        "/users/103",
			ContentType: "application/json",
			Body:        []byte(`{"id": "103", "name": "Bob Johnson", "email": "bob@example.com"}`),
		},
	}
	for _, data := range jsonData {
		ids := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
		if err := store.Create(ids, data); err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	// Test scenario 4: Collection with various JSON structures
	mixedJsonData := []*model.MockData{
		{
			Path:        "/products/p1",
			ContentType: "application/json",
			Body:        []byte(`{"id": "p1", "name": "Laptop", "price": 999.99, "inStock": true}`),
		},
		{
			Path:        "/products/p2",
			ContentType: "application/json",
			Body:        []byte(`{"id": "p2", "name": "Mouse", "price": 19.99, "specs": {"wireless": true, "dpi": 1200}}`),
		},
		{
			Path:        "/products/p3",
			ContentType: "application/json",
			Body:        []byte(`{"id": "p3", "name": "Keyboard", "price": 49.99, "variants": ["black", "white", "rgb"]}`),
		},
	}
	for _, data := range mixedJsonData {
		ids := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
		if err := store.Create(ids, data); err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	// Test scenario 5: Collection with larger JSON data
	largeJsonData := make([]*model.MockData, 0)
	for i := 1; i <= 50; i++ {
		data := &model.MockData{
			Path:        fmt.Sprintf("/articles/a%d", i),
			ContentType: "application/json",
			Body:        []byte(fmt.Sprintf(`{"id": "a%d", "title": "Article %d", "content": "%s", "tags": ["tag1", "tag2"], "published": true, "views": %d}`, i, i, strings.Repeat("Lorem ipsum dolor sit amet. ", 5), i*100)),
		}
		largeJsonData = append(largeJsonData, data)
		ids := []string{data.Path[strings.LastIndex(data.Path, "/")+1:]}
		if err := store.Create(ids, data); err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedItems  int
		validate       func(t *testing.T, body []byte)
	}{
		{
			name:           "Empty collection",
			path:           "/categories", // Using the new empty section
			expectedStatus: http.StatusOK,
			expectedItems:  0,
			validate: func(t *testing.T, body []byte) {
				var data []interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(data) != 0 {
					t.Errorf("expected empty array, got %d items", len(data))
				}
			},
		},
		{
			name:           "XML-only collection (should be empty)",
			path:           "/orders",
			expectedStatus: http.StatusOK,
			expectedItems:  0,
			validate: func(t *testing.T, body []byte) {
				var data []interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(data) != 0 {
					t.Errorf("expected empty array for XML-only content, got %d items", len(data))
				}
			},
		},
		{
			name:           "Multiple JSON items with same structure",
			path:           "/users",
			expectedStatus: http.StatusOK,
			expectedItems:  3,
			validate: func(t *testing.T, body []byte) {
				var data []map[string]interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(data) != 3 {
					t.Errorf("expected 3 items, got %d", len(data))
				}

				// Check for specific fields
				for _, item := range data {
					id, ok := item["id"].(string)
					if !ok {
						t.Errorf("missing or invalid 'id' field")
					}
					if _, ok := item["name"].(string); !ok {
						t.Errorf("missing or invalid 'name' field for id %s", id)
					}
					if _, ok := item["email"].(string); !ok {
						t.Errorf("missing or invalid 'email' field for id %s", id)
					}
				}
			},
		},
		{
			name:           "Various JSON structures",
			path:           "/products",
			expectedStatus: http.StatusOK,
			expectedItems:  3,
			validate: func(t *testing.T, body []byte) {
				var data []map[string]interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(data) != 3 {
					t.Errorf("expected 3 items, got %d", len(data))
				}

				// Verify each specific item has the expected structure
				foundSpecs := false
				foundVariants := false
				foundInStock := false

				for _, item := range data {
					if specs, ok := item["specs"].(map[string]interface{}); ok {
						foundSpecs = true
						if _, ok := specs["wireless"].(bool); !ok {
							t.Errorf("missing or invalid 'wireless' field in specs")
						}
					}
					if variants, ok := item["variants"].([]interface{}); ok {
						foundVariants = true
						if len(variants) != 3 {
							t.Errorf("expected 3 variants, got %d", len(variants))
						}
					}
					if _, ok := item["inStock"].(bool); ok {
						foundInStock = true
					}
				}

				if !foundSpecs {
					t.Errorf("missing 'specs' object in response")
				}
				if !foundVariants {
					t.Errorf("missing 'variants' array in response")
				}
				if !foundInStock {
					t.Errorf("missing 'inStock' boolean in response")
				}
			},
		},
		{
			name:           "Large collection of JSON items",
			path:           "/articles",
			expectedStatus: http.StatusOK,
			expectedItems:  50,
			validate: func(t *testing.T, body []byte) {
				var data []map[string]interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(data) != 50 {
					t.Errorf("expected 50 items, got %d", len(data))
				}

				// Check that all items have the expected fields
				for _, item := range data {
					id, ok := item["id"].(string)
					if !ok {
						t.Errorf("missing or invalid 'id' field")
					}
					if _, ok := item["title"].(string); !ok {
						t.Errorf("missing or invalid 'title' field for id %s", id)
					}
					if _, ok := item["content"].(string); !ok {
						t.Errorf("missing or invalid 'content' field for id %s", id)
					}
					if tags, ok := item["tags"].([]interface{}); !ok || len(tags) != 2 {
						t.Errorf("missing or invalid 'tags' field for id %s", id)
					}
				}

				// Check that the total response size is large (> 10KB)
				if len(body) < 10*1024 {
					t.Errorf("expected large response (>10KB), got %d bytes", len(body))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			// Handle request
			resp, err := service.HandleRequest(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status code = %d, want %d", resp.StatusCode, tt.expectedStatus)
			}

			// Read body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			// Validate response
			if tt.validate != nil {
				tt.validate(t, body)
			}
		})
	}
}
