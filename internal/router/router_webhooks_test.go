package router_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/router"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/internal/webhooks"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestRouter_ScenarioWithWebhook_FiresDelivery(t *testing.T) {
	var calls atomic.Int64
	recv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer recv.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"api": {PathPattern: "/api"},
		},
	}
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())
	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, cfg)
	techHandler := handler.NewTechHandler(techService, logger, nil)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

	ring := webhooks.NewRing(10)
	dispatcher := webhooks.NewDispatcher(logger, ring)

	appRouter := router.NewRouter(
		uniHandler, techHandler, scenarioHandler,
		scenarioService, techService, logger, cfg, dispatcher,
	)

	scenario := model.Scenario{
		UUID:        "hooky",
		RequestPath: "POST /api/notify",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"ok":true}`,
		Webhook: &model.WebhookConfig{
			URL:         recv.URL,
			Method:      "POST",
			MaxAttempts: 1,
			BaseMS:      1,
			MaxMS:       5,
		},
	}
	if _, err := scenarioService.CreateScenario(context.TODO(), scenario); err != nil {
		t.Fatalf("create scenario: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/notify", strings.NewReader(`{"trigger":true}`))
	w := httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("scenario response status: got %d want 201", w.Code)
	}

	// Wait for async dispatch to land on the receiver.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && calls.Load() == 0 {
		time.Sleep(5 * time.Millisecond)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 webhook delivery, got %d", calls.Load())
	}

	// And the deliveries ring should reflect the same.
	records := dispatcher.Snapshot()
	if len(records) != 1 {
		t.Fatalf("expected 1 ring record, got %d", len(records))
	}
	if records[0].StatusCode != 200 {
		t.Errorf("ring status: got %d want 200", records[0].StatusCode)
	}
}

func TestRouter_ScenarioWithoutWebhook_DoesNotDispatch(t *testing.T) {
	var calls atomic.Int64
	recv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
	}))
	defer recv.Close()

	appRouter, scenarioService := setupTestRouter(t)

	scenario := model.Scenario{
		UUID:        "no-hook",
		RequestPath: "GET /api/plain",
		StatusCode:  200,
		Data:        `{"ok":true}`,
	}
	if _, err := scenarioService.CreateScenario(context.TODO(), scenario); err != nil {
		t.Fatalf("create scenario: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/plain", nil)
	w := httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", w.Code)
	}
	time.Sleep(50 * time.Millisecond)
	if calls.Load() != 0 {
		t.Errorf("expected no deliveries, got %d", calls.Load())
	}
}

func TestRouter_DeliveriesEndpoint_ExposedViaUnderscoreUnderscore(t *testing.T) {
	// Use the same setup as the scenario with webhook test to confirm the
	// /_uni/webhooks/deliveries endpoint is reachable from the router.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"api": {PathPattern: "/api"},
		},
	}
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())
	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, cfg)

	ring := webhooks.NewRing(10)
	ring.Add(webhooks.Delivery{ID: "abc", URL: "u", Attempt: 1, StatusCode: 200, TS: time.Now()})

	techHandler := handler.NewTechHandler(techService, logger, ring)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)
	dispatcher := webhooks.NewDispatcher(logger, ring)
	appRouter := router.NewRouter(
		uniHandler, techHandler, scenarioHandler,
		scenarioService, techService, logger, cfg, dispatcher,
	)

	req := httptest.NewRequest("GET", "/_uni/webhooks/deliveries", nil)
	w := httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"abc"`) {
		t.Errorf("body missing record id: %s", w.Body.String())
	}
}

func TestRouter_NilDispatcherIsTolerated(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"api": {PathPattern: "/api"},
		},
	}
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())
	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, cfg)
	techHandler := handler.NewTechHandler(techService, logger, nil)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)

	// dispatcher nil → scenarios with webhooks must not crash, just log a warning.
	appRouter := router.NewRouter(
		uniHandler, techHandler, scenarioHandler,
		scenarioService, techService, logger, cfg, nil,
	)

	scenario := model.Scenario{
		UUID:        "nil-disp",
		RequestPath: "POST /api/null-disp",
		StatusCode:  201,
		Data:        `{}`,
		Webhook:     &model.WebhookConfig{URL: "http://localhost:1/never"},
	}
	if _, err := scenarioService.CreateScenario(context.TODO(), scenario); err != nil {
		t.Fatalf("create scenario: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/null-disp", nil)
	w := httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status: got %d want 201", w.Code)
	}
}
