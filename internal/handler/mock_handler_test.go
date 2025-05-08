package handler

import (
	"bytes"
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
	handler := NewMockHandler(mockService, logger)

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
			expectedStatus: http.StatusInternalServerError,
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
