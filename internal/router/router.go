package router

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/webhooks"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

const (
	pathLogKey = "path"
)

// maxDispatchBodyCapture is the upper bound on request body bytes retained for
// webhook payloads. Bodies larger than this are truncated to keep memory bounded.
const maxDispatchBodyCapture = 1 << 20 // 1 MiB

// Router wraps a Chi router with scenario handling capabilities
type Router struct {
	router           chi.Router
	uniHandler       http.Handler
	techHandler      http.Handler
	scenarioHandler  http.Handler
	scenarioService  *service.ScenarioService
	techService      *service.TechService
	logger           *slog.Logger
	uniConfig        *config.UniConfig
	webhookDispatcher *webhooks.Dispatcher
}

// NewRouter creates a new Router instance with Chi. The webhookDispatcher may be
// nil; in that case scenarios with a Webhook config are still served but no
// outbound delivery is attempted.
func NewRouter(
	uniHandler, techHandler, scenarioHandler http.Handler,
	scenarioService *service.ScenarioService,
	techService *service.TechService,
	logger *slog.Logger,
	uniConfig *config.UniConfig,
	webhookDispatcher *webhooks.Dispatcher,
) *Router {
	r := &Router{
		uniHandler:        uniHandler,
		techHandler:       techHandler,
		scenarioHandler:   scenarioHandler,
		scenarioService:   scenarioService,
		techService:       techService,
		logger:            logger,
		uniConfig:         uniConfig,
		webhookDispatcher: webhookDispatcher,
	}

	r.setupRoutes()
	return r
}

// setupRoutes configures the Chi router with all routes and middleware
func (r *Router) setupRoutes() {
	r.router = chi.NewRouter()
	
	// Add middleware
	r.router.Use(middleware.RequestID)
	r.router.Use(r.loggingMiddleware)
	r.router.Use(r.metricsMiddleware)
	r.router.Use(middleware.Recoverer)
	
	// Add scenario handling middleware (runs before route matching)
	r.router.Use(r.scenarioMiddleware)
	
	// Technical endpoints (/_uni/*)
	r.router.Mount("/_uni/scenarios", r.scenarioHandler)
	r.router.Mount("/_uni", r.techHandler)
	
	// Catch-all route for uni handler (must be last)
	r.router.HandleFunc("/*", r.uniHandlerFunc)
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// loggingMiddleware adds request logging
func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.logger.Debug("incoming request",
			"method", req.Method,
			"path", req.URL.Path,
			"remote_addr", req.RemoteAddr)
		next.ServeHTTP(w, req)
	})
}

// metricsMiddleware tracks request metrics using TechService
func (r *Router) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath := r.normalizePath(req.URL.Path)
		
		// Create a response writer wrapper to capture status code
		ww := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // default status code
		}
		
		// Increment request count before processing
		r.techService.IncrementRequestCount(req.Context(), requestPath)
		
		// Process request
		next.ServeHTTP(ww, req)
		
		// Track response after processing
		r.techService.TrackResponse(req.Context(), requestPath, ww.statusCode)
	})
}

// responseWriter wraps http.ResponseWriter to capture status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizePath normalizes the request path
func (*Router) normalizePath(path string) string {
	requestPath := strings.TrimSuffix(path, "/")
	if requestPath == "" {
		requestPath = "/"
	}
	return requestPath
}

// scenarioMiddleware checks for scenario matches before route handling
func (r *Router) scenarioMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath := r.normalizePath(req.URL.Path)

		// Skip scenario matching for technical endpoints
		if strings.HasPrefix(requestPath, "/_uni/") {
			next.ServeHTTP(w, req)
			return
		}

		scenario, found := r.scenarioService.GetScenarioByPath(req.Context(), requestPath, req.Method)
		if found {
			r.logger.Info("found matching scenario",
				"method", req.Method,
				pathLogKey, requestPath,
				"uuid", scenario.UUID)

			r.maybeDispatchScenarioWebhook(req, scenario)
			r.writeScenarioResponse(w, req, scenario)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// maybeDispatchScenarioWebhook fires the scenario's webhook asynchronously when one
// is configured and a dispatcher is wired in. The request body is read once (and
// restored on the request) so the downstream handler still sees it intact.
func (r *Router) maybeDispatchScenarioWebhook(req *http.Request, scenario model.Scenario) {
	if scenario.Webhook == nil {
		return
	}
	if r.webhookDispatcher == nil {
		if r.logger != nil {
			r.logger.Warn("scenario has webhook but no dispatcher is wired",
				"scenario_uuid", scenario.UUID, "webhook_url", scenario.Webhook.URL)
		}
		return
	}
	body := captureRequestBody(req)
	requestPath := r.normalizePath(req.URL.Path)
	r.webhookDispatcher.Dispatch(req.Context(), scenario.Webhook, req.Method+" "+requestPath, body, nil)
}

// captureRequestBody returns the request body bytes (capped) and restores the
// body so subsequent handlers can still read it. Errors are tolerated: the body
// is simply treated as empty.
func captureRequestBody(req *http.Request) []byte {
	if req.Body == nil {
		return nil
	}
	defer func() {
		_ = req.Body.Close()
	}()
	limited := io.LimitReader(req.Body, maxDispatchBodyCapture)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil
	}
	// Restore body for any subsequent handler.
	req.Body = io.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))
	return data
}

// writeScenarioResponse writes the scenario response
func (r *Router) writeScenarioResponse(w http.ResponseWriter, req *http.Request, scenario model.Scenario) {
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
	
	// For HEAD requests, don't write response body
	if req.Method != http.MethodHead {
		if _, err := w.Write([]byte(scenario.Data)); err != nil {
			r.logger.Error("failed to write scenario response in router", "error", err)
		}
	}
}

// uniHandlerFunc wraps the uni handler with path validation
func (r *Router) uniHandlerFunc(w http.ResponseWriter, req *http.Request) {
	requestPath := r.normalizePath(req.URL.Path)
	
	if r.uniConfig == nil {
		r.logger.Error("router's uniConfig is nil", pathLogKey, requestPath)
		http.Error(w, "server configuration error", http.StatusInternalServerError)
		return
	}

	_, section, err := r.uniConfig.MatchPath(requestPath)
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

	r.logger.Debug("routing to uni handler", pathLogKey, requestPath)
	r.uniHandler.ServeHTTP(w, req)
}

