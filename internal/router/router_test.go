package router_test

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
	"github.com/bmcszk/unimock/internal/router"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestRouter_ServeHTTP(t *testing.T) {
	appRouter, scenarioService := setupTestRouter(t)
	setupTestScenarios(t, scenarioService)
	tests := getRouterTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executeRouterTest(t, appRouter, tt)
		})
	}
}

type routerTestCase struct {
	name             string
	method           string
	path             string
	wantStatusCode   int
	wantBodyContains string
}

func setupTestRouter(_ *testing.T) (*router.Router, *service.ScenarioService) {
	// Create a mock logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create storages
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()

	// Create test config
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"api": {
				PathPattern: "/api",
			},
		},
	}

	// Create services
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())

	// Create handlers
	uniHandler := handler.NewUniHandler(uniService, scenarioService, techService, logger, cfg)
	techHandler := handler.NewTechHandler(techService, logger)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

	// Create router
	appRouter := router.NewRouter(uniHandler, techHandler, scenarioHandler, scenarioService, logger, cfg)

	return appRouter, scenarioService
}

func setupTestScenarios(t *testing.T, scenarioService *service.ScenarioService) {
	t.Helper()
	scenarios := []model.Scenario{
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

	for _, scenario := range scenarios {
		_, err := scenarioService.CreateScenario(context.TODO(), scenario)
		if err != nil {
			t.Fatalf("Failed to create test scenario: %v", err)
		}
	}
}

func getRouterTestCases() []routerTestCase {
	return []routerTestCase{
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
}

func executeRouterTest(t *testing.T, appRouter *router.Router, tt routerTestCase) {
	t.Helper()
	req := httptest.NewRequest(tt.method, tt.path, nil)
	w := httptest.NewRecorder()

	appRouter.ServeHTTP(w, req)

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
}
