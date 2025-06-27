package router_test

import (
	"context"
	"io"
	"log/slog"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_ScenariosIgnoreReturnBodyConfig(t *testing.T) {
	// Setup router with return_body=false configuration
	appRouter, scenarioService := setupTestRouterWithReturnBodyFalse(t)

	// Create a scenario
	scenario := model.Scenario{
		UUID:        "test-body-scenario",
		RequestPath: "POST /api/resources",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"id": "123", "created": true}`,
	}

	_, err := scenarioService.CreateScenario(context.TODO(), scenario)
	require.NoError(t, err)

	// Make request that matches scenario
	req := httptest.NewRequest("POST", "/api/resources", strings.NewReader(`{"name": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	appRouter.ServeHTTP(w, req)

	// Verify scenario ALWAYS returns body despite return_body=false config
	assert.Equal(t, 201, w.Code)
	assert.Contains(t, w.Body.String(), `{"id": "123", "created": true}`)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestRouter_ScenarioVsMockReturnBodyBehavior(t *testing.T) {
	appRouter, scenarioService := setupTestRouterWithReturnBodyFalse(t)

	// Create scenario for /api/test
	scenario := model.Scenario{
		UUID:        "comparison-scenario",
		RequestPath: "POST /api/test",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"scenario": "response"}`,
	}
	_, err := scenarioService.CreateScenario(context.TODO(), scenario)
	require.NoError(t, err)

	// Test 1: Request matching scenario - should return body
	req1 := httptest.NewRequest("POST", "/api/test", strings.NewReader(`{"test": "data"}`))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	appRouter.ServeHTTP(w1, req1)

	assert.Equal(t, 201, w1.Code)
	assert.Contains(t, w1.Body.String(), `{"scenario": "response"}`)

	// Test 2: Request to different path (no scenario match) - should respect return_body=false
	req2 := httptest.NewRequest("POST", "/api/other", strings.NewReader(`{"test": "data"}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	appRouter.ServeHTTP(w2, req2)

	assert.Equal(t, 201, w2.Code)
	assert.Empty(t, w2.Body.String()) // Mock handler respects return_body=false
}

func TestRouter_ScenariosAlwaysReturnUUID(t *testing.T) {
	appRouter, _ := setupTestRouterWithReturnBodyFalse(t)

	// Test 1: POST scenario without UUID should get one generated
	reqBody := `{"requestPath": "GET /test/uuid", "statusCode": 200, ` +
		`"contentType": "application/json", "data": "{\"message\": \"test\"}"}`
	req := httptest.NewRequest("POST", "/_uni/scenarios", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	appRouter.ServeHTTP(w, req)

	// Verify response includes generated UUID
	assert.Equal(t, 201, w.Code)
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, `"uuid":`)
	assert.NotContains(t, responseBody, `"uuid":""`)
	assert.NotContains(t, responseBody, `"uuid":"",`)
}

func setupTestRouterWithReturnBodyFalse(t *testing.T) (*router.Router, *service.ScenarioService) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()

	// Configuration with return_body=false
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"api": {
				PathPattern: "/api/*",
				BodyIDPaths: []string{"/id"},
				ReturnBody:  false, // This should be ignored by scenarios
			},
		},
	}

	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())

	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, cfg)
	techHandler := handler.NewTechHandler(techService, logger)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

	return router.NewRouter(
		uniHandler, techHandler, scenarioHandler, 
		scenarioService, techService, logger, cfg,
	), scenarioService
}