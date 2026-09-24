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

// streamRouterTestServer mounts the full unimock Router under httptest.NewServer
// and registers a streaming scenario. The returned server can be hit with a
// real http.Client so http.Flusher is exercised end-to-end.
func streamRouterTestServer(t *testing.T, scenario model.Scenario) *httptest.Server {
	t.Helper()
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
	appRouter := router.NewRouter(
		uniHandler, techHandler, scenarioHandler,
		scenarioService, techService, logger, cfg,
		router.Deps{StreamWriter: handler.NewStreamWriter(logger)},
	)
	if _, err := scenarioService.CreateScenario(context.TODO(), scenario); err != nil {
		t.Fatalf("create scenario: %v", err)
	}
	srv := httptest.NewServer(appRouter)
	t.Cleanup(srv.Close)
	return srv
}

// countSSEFrames reads r and counts the number of `data: ` lines.
func countSSEFrames(t *testing.T, r io.Reader) int {
	t.Helper()
	body, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	count := 0
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "data: ") {
			count++
		}
	}
	return count
}

func TestRouter_StreamScenario_SSE(t *testing.T) {
	srv := streamRouterTestServer(t, model.Scenario{
		UUID:        "stream-sse",
		RequestPath: "GET /api/events",
		StatusCode:  200,
		Stream: &model.StreamConfig{
			Format:     "sse",
			EventCount: 2,
			IntervalMS: 10,
			Template:   "tick-{{index}}",
		},
	})

	resp, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	assertStreamHeaders(t, resp, "text/event-stream")
	if got := countSSEFrames(t, resp.Body); got != 2 {
		t.Errorf("frames: got %d want 2", got)
	}
}

// assertStreamHeaders verifies the contract headers for a streaming response.
func assertStreamHeaders(t *testing.T, resp *http.Response, contentType string) {
	t.Helper()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != contentType {
		t.Errorf("Content-Type: got %q want %q", got, contentType)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control: got %q want no-cache", got)
	}
	if got := resp.Header.Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering: got %q want no", got)
	}
}

func TestRouter_StreamScenario_HEAD_ReturnsHeadersNoBody(t *testing.T) {
	srv := streamRouterTestServer(t, model.Scenario{
		UUID:        "stream-head",
		RequestPath: "HEAD /api/head-stream",
		StatusCode:  200,
		ContentType: "application/json",
		Stream: &model.StreamConfig{
			Format:     "sse",
			EventCount: 5,
			Template:   "should-not-appear",
		},
	})

	req, err := http.NewRequest(http.MethodHead, srv.URL+"/api/head-stream", nil)
	if err != nil {
		t.Fatalf("HEAD req: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HEAD do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("HEAD body read: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("HEAD body should be empty, got: %q", body)
	}
}

func TestRouter_StreamScenario_NilStreamWriter_Returns500(t *testing.T) {
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
	appRouter := router.NewRouter(
		uniHandler, techHandler, scenarioHandler,
		scenarioService, techService, logger, cfg,
		router.Deps{}, // no stream writer
	)
	scenario := model.Scenario{
		UUID:        "no-streamer",
		RequestPath: "GET /api/no-streamer",
		StatusCode:  200,
		Stream: &model.StreamConfig{
			Format:     "sse",
			EventCount: 1,
			Template:   "x",
		},
	}
	if _, err := scenarioService.CreateScenario(context.TODO(), scenario); err != nil {
		t.Fatalf("create scenario: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/no-streamer", nil)
	w := httptest.NewRecorder()
	appRouter.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d want 500", w.Code)
	}
}