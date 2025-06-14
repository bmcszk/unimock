package router

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

const (
	pathLogKey = "path"
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
func NewRouter(
	mockHandler, techHandler, scenarioHandler http.Handler, 
	scenarioService service.ScenarioService, 
	logger *slog.Logger, 
	mockConfig *config.MockConfig,
) *Router {
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
	requestPath := r.normalizePath(req.URL.Path)
	
	if r.handleScenario(w, req, requestPath) {
		return
	}
	
	if r.routeToSpecialHandlers(w, req, requestPath) {
		return
	}
	
	r.routeToMockHandler(w, req, requestPath)
}

// normalizePath normalizes the request path
func (*Router) normalizePath(path string) string {
	requestPath := strings.TrimSuffix(path, "/")
	if requestPath == "" {
		requestPath = "/"
	}
	return requestPath
}

// handleScenario checks and handles scenario matching
func (r *Router) handleScenario(w http.ResponseWriter, req *http.Request, requestPath string) bool {
	scenario := r.scenarioService.GetScenarioByPath(req.Context(), requestPath, req.Method)
	if scenario == nil {
		return false
	}

	r.logger.Info("found matching scenario",
		"method", req.Method,
		pathLogKey, requestPath,
		"uuid", scenario.UUID)

	r.writeScenarioResponse(w, scenario)
	return true
}

// writeScenarioResponse writes the scenario response
func (r *Router) writeScenarioResponse(w http.ResponseWriter, scenario *model.Scenario) {
	w.Header().Set("Content-Type", scenario.ContentType)
	if scenario.Location != "" {
		w.Header().Set("Location", scenario.Location)
	}
	
	if scenario.Headers != nil {
		for k, v := range scenario.Headers {
			w.Header().Set(k, v)
		}
	}
	
	w.WriteHeader(scenario.StatusCode)
	
	if _, err := w.Write([]byte(scenario.Data)); err != nil {
		r.logger.Error("failed to write scenario response in router", "error", err)
	}
}

// routeToSpecialHandlers routes to scenario or tech handlers
func (r *Router) routeToSpecialHandlers(w http.ResponseWriter, req *http.Request, requestPath string) bool {
	if strings.HasPrefix(requestPath, "/_uni/scenarios") {
		r.logger.Debug("routing to scenario handler", pathLogKey, requestPath)
		r.scenarioHandler.ServeHTTP(w, req)
		return true
	}

	if strings.HasPrefix(requestPath, "/_uni/") {
		r.logger.Debug("routing to technical handler", pathLogKey, requestPath)
		r.techHandler.ServeHTTP(w, req)
		return true
	}

	return false
}

// routeToMockHandler routes to the mock handler after validation
func (r *Router) routeToMockHandler(w http.ResponseWriter, req *http.Request, requestPath string) {
	if r.mockConfig == nil {
		r.logger.Error("router's mockConfig is nil", pathLogKey, requestPath)
		http.Error(w, "server configuration error", http.StatusInternalServerError)
		return
	}

	_, section, err := r.mockConfig.MatchPath(requestPath)
	if err != nil {
		r.logger.Error("error matching path in router", pathLogKey, requestPath, "error", err)
		http.Error(w, "error processing request path configuration", http.StatusInternalServerError)
		return
	}
	
	if section == nil {
		r.logger.Warn("no matching section found for path in router", pathLogKey, requestPath)
		http.Error(w, "Not Found: No matching mock configuration or active scenario for path", http.StatusNotFound)
		return
	}

	r.logger.Debug("routing to mock handler", pathLogKey, requestPath)
	r.mockHandler.ServeHTTP(w, req)
}
