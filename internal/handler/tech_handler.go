package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/internal/service"
)

// TechHandler handles technical endpoints like health checks and metrics
type TechHandler struct {
	prefix  string
	service *service.TechService
	logger  *slog.Logger
}

// NewTechHandler creates a new instance of TechHandler
func NewTechHandler(techSvc *service.TechService, logger *slog.Logger) *TechHandler {
	return &TechHandler{
		prefix:  "/_uni/",
		service: techSvc,
		logger:  logger,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *TechHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log the request
	h.logger.Info("technical endpoint request",
		"method", r.Method,
		"path", r.URL.Path)


	// Only allow GET method for technical endpoints
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handle based on path
	path := strings.TrimPrefix(r.URL.Path, h.prefix)
	switch path {
	case "health":
		h.handleHealthCheck(w, r)
	case "metrics":
		h.handleMetrics(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleHealthCheck returns the health status of the service
func (h *TechHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Get health status from service
	response := h.service.GetHealthStatus(r.Context())

	// Write response
	h.writeJSONResponse(w, response)
}

// handleMetrics returns metrics about the service
func (h *TechHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Get metrics from service
	response := h.service.GetMetrics(r.Context())

	// Write response
	h.writeJSONResponse(w, response)
}

// writeJSONResponse writes a JSON response
func (h *TechHandler) writeJSONResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	// Marshal response to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("failed to marshal JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	_, err = w.Write(jsonData)
	if err != nil {
		h.logger.Error("failed to write response", "error", err)
	}
}
