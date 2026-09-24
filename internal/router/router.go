package router

import (
	"bytes"
	"context"
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

// StreamResponseWriter is the contract the router needs from a stream writer
// implementation. Defined here so tests can supply a fake without depending on
// the handler package.
type StreamResponseWriter interface {
	WriteStream(ctx context.Context, w http.ResponseWriter, sc *model.StreamConfig)
}

// Deps bundles optional side-effecting collaborators (webhook dispatcher,
// stream writer) so the NewRouter constructor stays at 8 parameters.
type Deps struct {
	// WebhookDispatcher fires outbound webhooks for scenarios with a Webhook
	// config. May be nil; nil is tolerated but logged at warn time.
	WebhookDispatcher *webhooks.Dispatcher
	// StreamWriter writes server-generated streams for scenarios with a Stream
	// config. May be nil; nil surfaces as 500 on the response when matched.
	StreamWriter StreamResponseWriter
}

// Router wraps a Chi router with scenario handling capabilities
type Router struct {
	router            chi.Router
	uniHandler        http.Handler
	techHandler       http.Handler
	scenarioHandler   http.Handler
	scenarioService   *service.ScenarioService
	techService       *service.TechService
	logger            *slog.Logger
	uniConfig         *config.UniConfig
	deps              Deps
}

// NewRouter creates a new Router instance with Chi. deps may have nil fields;
// a nil WebhookDispatcher means webhook scenarios are served but no delivery
// is attempted; a nil StreamWriter means streaming scenarios return 500.
func NewRouter(
	uniHandler, techHandler, scenarioHandler http.Handler,
	scenarioService *service.ScenarioService,
	techService *service.TechService,
	logger *slog.Logger,
	uniConfig *config.UniConfig,
	deps Deps,
) *Router {
	r := &Router{
		uniHandler:      uniHandler,
		techHandler:     techHandler,
		scenarioHandler: scenarioHandler,
		scenarioService: scenarioService,
		techService:     techService,
		logger:          logger,
		uniConfig:       uniConfig,
		deps:            deps,
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

// responseWriter wraps http.ResponseWriter to capture status codes. It also
// forwards Flush() (and the optional CloseNotifier / Hijacker / Pusher) so
// downstream handlers that need to flush frames (e.g. stream writers) keep
// working under the metrics middleware.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush forwards to the underlying ResponseWriter when it supports flushing.
// Returns silently when the wrapped writer is not an http.Flusher (e.g. an
// httptest.ResponseRecorder used in some tests) so callers stay non-fatal.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
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
	if r.deps.WebhookDispatcher == nil {
		if r.logger != nil {
			r.logger.Warn("scenario has webhook but no dispatcher is wired",
				"scenario_uuid", scenario.UUID, "webhook_url", scenario.Webhook.URL)
		}
		return
	}
	body := captureRequestBody(req)
	requestPath := r.normalizePath(req.URL.Path)
	r.deps.WebhookDispatcher.Dispatch(req.Context(), scenario.Webhook, req.Method+" "+requestPath, body, nil)
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
	if scenario.Stream != nil {
		r.writeStreamScenarioResponse(w, req, scenario)
		return
	}
	r.writeStaticScenarioResponse(w, req, scenario)
}

// writeStaticScenarioResponse writes the static (non-streaming) response body
// path: headers + status + (optional) Data. HEAD requests get headers only.
func (r *Router) writeStaticScenarioResponse(
	w http.ResponseWriter, req *http.Request, scenario model.Scenario,
) {
	applyScenarioHeaders(w, scenario)
	w.WriteHeader(scenario.StatusCode)
	// For HEAD requests, don't write response body
	if req.Method == http.MethodHead {
		return
	}
	if _, err := w.Write([]byte(scenario.Data)); err != nil {
		r.logger.Error("failed to write scenario response in router", "error", err)
	}
}

// writeStreamScenarioResponse delegates to the stream writer when the scenario
// has a Stream config. HEAD requests fall back to a static empty 200 with the
// scenario headers only (no frames are sent).
func (r *Router) writeStreamScenarioResponse(
	w http.ResponseWriter, req *http.Request, scenario model.Scenario,
) {
	if req.Method == http.MethodHead {
		applyScenarioHeaders(w, scenario)
		w.WriteHeader(scenario.StatusCode)
		return
	}
	if r.deps.StreamWriter == nil {
		r.failMissingStreamWriter(w, scenario)
		return
	}
	r.deps.StreamWriter.WriteStream(req.Context(), w, scenario.Stream)
}

// failMissingStreamWriter writes a 500 when a stream scenario fires but no
// stream writer is wired into the router. Kept as its own helper to keep
// writeStreamScenarioResponse at low complexity.
func (r *Router) failMissingStreamWriter(w http.ResponseWriter, scenario model.Scenario) {
	if r.logger != nil {
		r.logger.Error("scenario has stream but no stream writer is wired",
			"scenario_uuid", scenario.UUID)
	}
	http.Error(w, "streaming responses are not enabled", http.StatusInternalServerError)
}

// applyScenarioHeaders writes Content-Type, Location, and the scenario-supplied
// extra headers onto w. Centralised so static and stream branches share the
// same header semantics.
func applyScenarioHeaders(w http.ResponseWriter, scenario model.Scenario) {
	w.Header().Set("Content-Type", scenario.ContentType)
	if scenario.Location != "" {
		w.Header().Set("Location", scenario.Location)
	}
	for k, v := range scenario.Headers {
		w.Header().Set(k, v)
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

