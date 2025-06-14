package router

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestRouter_ServeHTTP(t *testing.T) {
	// Create a mock logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create storages
	store := storage.NewMockStorage()
	scenarioStore := storage.NewScenarioStorage()

	// Create test config
	cfg := &config.MockConfig{
		Sections: map[string]config.Section{
			"api": {
				PathPattern: "/api",
			},
		},
	}

	// Create services
	mockService := service.NewMockService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())

	// Create handlers
	mockHandler := handler.NewMockHandler(mockService, scenarioService, logger, cfg)
	techHandler := handler.NewTechHandler(techService, logger)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

	// Create router
	router := NewRouter(mockHandler, techHandler, scenarioHandler, scenarioService, logger, cfg)

	// Setup test scenarios
	scenarios := []*model.Scenario{
		{
			UUID:        "test-scenario-1",
			RequestPath: "GET /api/test",
			StatusCode:  201,
			ContentType: "application/json",
			Data:        `{"message": "This is a test scenario"}`,
		},
		{
			UUID:        "test-scenario-2",
			RequestPath: "POST /api/test",
			StatusCode:  202,
			ContentType: "application/json",
			Data:        `{"message": "This is a POST scenario"}`,
		},
	}

	// Store the test scenarios
	for _, scenario := range scenarios {
		err := scenarioService.CreateScenario(context.TODO(), scenario)
		if err != nil {
			t.Fatalf("Failed to create test scenario: %v", err)
		}
	}

	// Test cases
	tests := []struct {
		name             string
		method           string
		path             string
		wantStatusCode   int
		wantBodyContains string
	}{
		{
			name:             "scenario GET request",
			method:           http.MethodGet,
			path:             "/api/test",
			wantStatusCode:   201,
			wantBodyContains: "This is a test scenario",
		},
		{
			name:             "scenario POST request",
			method:           http.MethodPost,
			path:             "/api/test",
			wantStatusCode:   202,
			wantBodyContains: "This is a POST scenario",
		},
		{
			name:           "scenario management endpoint",
			method:         http.MethodGet,
			path:           "/_uni/scenarios",
			wantStatusCode: 200,
		},
		{
			name:             "technical endpoint",
			method:           http.MethodGet,
			path:             "/_uni/health",
			wantStatusCode:   200,
			wantBodyContains: "status",
		},
		{
			name:             "regular API endpoint",
			method:           http.MethodGet,
			path:             "/api/other",
			wantStatusCode:   http.StatusNotFound,
			wantBodyContains: "Not Found: No matching mock configuration or active scenario for path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("Status code = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			if tt.wantBodyContains != "" {
				body, _ := io.ReadAll(resp.Body)
				if !strings.Contains(string(body), tt.wantBodyContains) {
					t.Errorf("Body = %q, want to contain %q", string(body), tt.wantBodyContains)
				}
			}
		})
	}
}
