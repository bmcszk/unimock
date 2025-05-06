package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/config"
	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/model"
	"github.com/bmcszk/unimock/internal/storage"
)

func TestScenarioPathMatching(t *testing.T) {
	// Create a mock logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create storages
	store := storage.NewStorage()
	scenarioStore := storage.NewScenarioStorage()

	// Create test config
	cfg := &config.Config{
		Sections: map[string]config.Section{
			"api": {
				PathPattern: "/api",
			},
		},
	}

	// Create handlers
	mainHandler := handler.NewHandler(store, cfg, logger)
	techHandler := handler.NewTechHandler(logger, time.Now())
	scenarioHandler := handler.NewScenarioHandler(scenarioStore, logger)

	// Create router
	router := &routeHandler{
		mainHandler:     mainHandler,
		techHandler:     techHandler,
		scenarioHandler: scenarioHandler,
		logger:          logger,
	}

	// Create test scenarios
	testScenario := &model.Scenario{
		UUID:        "test-uuid",
		Path:        "/api/test",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"message": "This is a test scenario"}`,
	}

	// Store the test scenario
	err := scenarioStore.Create(testScenario.UUID, testScenario)
	if err != nil {
		t.Fatalf("Failed to create test scenario: %v", err)
	}

	// Test cases
	tests := []struct {
		name           string
		path           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "matching scenario path",
			path:           "/api/test",
			wantStatusCode: 201,
			wantBody:       `{"message": "This is a test scenario"}`,
		},
		{
			name:           "non-matching path",
			path:           "/api/other",
			wantStatusCode: 400,
		},
		{
			name:           "scenario management endpoint",
			path:           "/_uni/scenarios",
			wantStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("Status code = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			if tt.wantBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if strings.TrimSpace(string(body)) != strings.TrimSpace(tt.wantBody) {
					t.Errorf("Body = %q, want %q", string(body), tt.wantBody)
				}
			}
		})
	}
}
