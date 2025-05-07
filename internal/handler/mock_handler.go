package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/bmcszk/unimock/internal/service"
)

// MockHandler represents our HTTP request handler
type MockHandler struct {
	service service.MockService
	logger  *slog.Logger
}

// NewMockHandler creates a new instance of Handler
func NewMockHandler(service service.MockService, logger *slog.Logger) *MockHandler {
	return &MockHandler{
		service: service,
		logger:  logger,
	}
}

// writeJSONResponse writes a JSON array response for collection endpoints
func writeJSONResponse(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(data)
	return err
}

// writeErrorResponse writes an error response with appropriate status code
func writeErrorResponse(w http.ResponseWriter, msg string, statusCode int) {
	http.Error(w, msg, statusCode)
}

// writeResourceResponse writes a JSON response for a single resource
func writeResourceResponse(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(data)
	return err
}

// HandleRequest processes the HTTP request and returns appropriate response
func (h *MockHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.HandleRequest(r.Context(), r)
	if err != nil {
		h.logger.Error("failed to handle request", "error", err)
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if resp.Body != nil {
		defer resp.Body.Close()
		if _, err := io.Copy(w, resp.Body); err != nil {
			h.logger.Error("failed to write response body", "error", err)
		}
	}
}

// ServeHTTP implements the http.Handler interface
func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandleRequest(w, r)
}
