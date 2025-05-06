package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// TechHandler handles technical endpoints like health checks and metrics
type TechHandler struct {
	prefix         string
	logger         *slog.Logger
	startTime      time.Time
	requestCounter atomic.Int64
	endpointStats  map[string]atomic.Int64
}

// NewTechHandler creates a new instance of TechHandler
func NewTechHandler(logger *slog.Logger, startTime time.Time) *TechHandler {
	return &TechHandler{
		prefix:        "/_uni/",
		logger:        logger,
		startTime:     startTime,
		endpointStats: make(map[string]atomic.Int64),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *TechHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log the request
	h.logger.Info("technical endpoint request",
		"method", r.Method,
		"path", r.URL.Path)

	// Increment request counter
	h.requestCounter.Add(1)

	// Track endpoint stats
	if stats, exists := h.endpointStats[r.URL.Path]; exists {
		stats.Add(1)
	} else {
		var counter atomic.Int64
		counter.Add(1)
		h.endpointStats[r.URL.Path] = counter
	}

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
	// Calculate uptime
	uptime := time.Since(h.startTime).String()

	// Create response
	response := map[string]interface{}{
		"status": "ok",
		"uptime": uptime,
	}

	// Write response
	h.writeJSONResponse(w, response)
}

// handleMetrics returns metrics about the service
func (h *TechHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Create API endpoint stats
	apiEndpoints := make(map[string]int64)
	for path, counter := range h.endpointStats {
		apiEndpoints[path] = counter.Load()
	}

	// Create response
	response := map[string]interface{}{
		"request_count": h.requestCounter.Load(),
		"api_endpoints": apiEndpoints,
	}

	// Write response
	h.writeJSONResponse(w, response)
}

// writeJSONResponse writes a JSON response
func (h *TechHandler) writeJSONResponse(w http.ResponseWriter, data interface{}) {
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
