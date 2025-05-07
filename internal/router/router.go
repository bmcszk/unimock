package router

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/internal/service"
)

// Router is a http.Handler that routes requests to the appropriate handler based on path prefix
type Router struct {
	mainHandler     http.Handler
	techHandler     http.Handler
	scenarioHandler http.Handler
	scenarioService service.ScenarioService
	logger          *slog.Logger
}

// NewRouter creates a new Router instance
func NewRouter(mainHandler, techHandler, scenarioHandler http.Handler, scenarioService service.ScenarioService, logger *slog.Logger) *Router {
	return &Router{
		mainHandler:     mainHandler,
		techHandler:     techHandler,
		scenarioHandler: scenarioHandler,
		scenarioService: scenarioService,
		logger:          logger,
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// First, check if there's a scenario matching this path and method
	scenario := r.scenarioService.GetScenarioByPath(req.Context(), req.URL.Path, req.Method)
	if scenario != nil {
		r.logger.Info("found matching scenario",
			"method", req.Method,
			"path", req.URL.Path,
			"uuid", scenario.UUID)

		// Set the status code and content type from the scenario
		w.Header().Set("Content-Type", scenario.ContentType)
		w.WriteHeader(scenario.StatusCode)

		// Write the response body from the scenario
		w.Write([]byte(scenario.Data))
		return
	}

	if strings.HasPrefix(req.URL.Path, "/_uni/scenarios") {
		r.logger.Debug("routing to scenario handler", "path", req.URL.Path)
		r.scenarioHandler.ServeHTTP(w, req)
		return
	}

	if strings.HasPrefix(req.URL.Path, "/_uni/") {
		r.logger.Debug("routing to technical handler", "path", req.URL.Path)
		r.techHandler.ServeHTTP(w, req)
		return
	}

	r.logger.Debug("routing to main handler", "path", req.URL.Path)
	r.mainHandler.ServeHTTP(w, req)
}
