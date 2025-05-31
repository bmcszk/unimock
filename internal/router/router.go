package router

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/pkg/config"
)

// Router is a http.Handler that routes requests to the appropriate handler based on path prefix
type Router struct {
	mockHandler     http.Handler
	techHandler     http.Handler
	scenarioHandler http.Handler
	scenarioService service.ScenarioService
	logger          *slog.Logger
	mockConfig      *config.MockConfig
}

// NewRouter creates a new Router instance
func NewRouter(mockHandler, techHandler, scenarioHandler http.Handler, scenarioService service.ScenarioService, logger *slog.Logger, mockConfig *config.MockConfig) *Router {
	return &Router{
		mockHandler:     mockHandler,
		techHandler:     techHandler,
		scenarioHandler: scenarioHandler,
		scenarioService: scenarioService,
		logger:          logger,
		mockConfig:      mockConfig,
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Normalize path once
	requestPath := strings.TrimSuffix(req.URL.Path, "/")
	if requestPath == "" { // Handle case of request to "/"
		requestPath = "/" // Or some other defined root behavior
	}

	// First, check if there's a scenario matching this path and method
	scenario := r.scenarioService.GetScenarioByPath(req.Context(), requestPath, req.Method)
	if scenario != nil {
		r.logger.Info("found matching scenario",
			"method", req.Method,
			"path", requestPath,
			"uuid", scenario.UUID)

		// Set the status code and content type from the scenario
		w.Header().Set("Content-Type", scenario.ContentType)
		if scenario.Location != "" {
			w.Header().Set("Location", scenario.Location)
		}
		// Apply custom headers from scenario
		if scenario.Headers != nil {
			for k, v := range scenario.Headers {
				w.Header().Set(k, v)
			}
		}
		w.WriteHeader(scenario.StatusCode)

		// Write the response body from the scenario
		if _, err := w.Write([]byte(scenario.Data)); err != nil {
			r.logger.Error("failed to write scenario response in router", "error", err)
		}
		return
	}

	if strings.HasPrefix(requestPath, "/_uni/scenarios") {
		r.logger.Debug("routing to scenario handler", "path", requestPath)
		r.scenarioHandler.ServeHTTP(w, req)
		return
	}

	if strings.HasPrefix(requestPath, "/_uni/") {
		r.logger.Debug("routing to technical handler", "path", requestPath)
		r.techHandler.ServeHTTP(w, req)
		return
	}

	// Check if the path matches any section pattern
	if cfg := r.mockConfig; cfg != nil {
		_, section, err := cfg.MatchPath(requestPath)
		if err != nil {
			r.logger.Error("error matching path in router", "path", requestPath, "error", err)
			http.Error(w, "error processing request path configuration", http.StatusInternalServerError)
			return
		}
		if section == nil {
			r.logger.Warn("no matching section found for path in router", "path", requestPath)
			http.Error(w, "Not Found: No matching mock configuration or active scenario for path", http.StatusNotFound)
			return
		}
	} else {
		r.logger.Error("router's mockConfig is nil", "path", requestPath)
		http.Error(w, "server configuration error", http.StatusInternalServerError)
		return
	}

	r.logger.Debug("routing to mock handler", "path", requestPath)
	r.mockHandler.ServeHTTP(w, req)
}
