package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

// MockHandler represents our HTTP request handler
type MockHandler struct {
	service service.MockService
	logger  *slog.Logger
	mockCfg *config.MockConfig
}

// NewMockHandler creates a new instance of Handler
func NewMockHandler(service service.MockService, logger *slog.Logger, cfg *config.MockConfig) *MockHandler {
	return &MockHandler{
		service: service,
		logger:  logger,
		mockCfg: cfg,
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

// ExtractIDs extracts IDs from the request using configured paths
func (h *MockHandler) ExtractIDs(ctx context.Context, req *http.Request) ([]string, error) {
	// Find matching section
	var sectionName string
	var pathPattern string
	var bodyIDPaths []string
	var headerIDName string

	if h.mockCfg != nil {
		// Use the config
		var section *config.Section
		var err error
		sectionName, section, err = h.mockCfg.MatchPath(req.URL.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to match path pattern: %w", err)
		}
		if section == nil {
			return nil, fmt.Errorf("no matching section found for path: %s", req.URL.Path)
		}
		pathPattern = section.PathPattern
		bodyIDPaths = section.BodyIDPaths
		headerIDName = section.HeaderIDName
	} else {
		return nil, fmt.Errorf("service configuration is missing")
	}

	// For GET/PUT/DELETE requests, try to extract ID from path first
	if req.Method == http.MethodGet || req.Method == http.MethodPut || req.Method == http.MethodDelete {
		pathSegments := getPathInfo(req.URL.Path)
		patternSegments := getPathInfo(pathPattern)

		// Check if this is a resource path (contains an ID)
		if len(pathSegments) > len(patternSegments) ||
			(len(patternSegments) > 0 && len(pathSegments) > 0 &&
				patternSegments[len(patternSegments)-1] == "*" &&
				len(pathSegments) == len(patternSegments)) {

			// Use the last path segment as the ID
			lastSegment := pathSegments[len(pathSegments)-1]
			if isValidID(lastSegment, sectionName) {
				return []string{lastSegment}, nil
			}
		}

		// If we got here, it's a collection path without an ID
		return nil, nil
	}

	// For POST requests, try to extract ID from header first
	if headerIDName != "" {
		if id := req.Header.Get(headerIDName); id != "" {
			return []string{id}, nil
		}
	}

	// Try to extract ID from body
	contentType := req.Header.Get("Content-Type")
	contentTypeLower := strings.ToLower(contentType)
	if strings.Contains(contentTypeLower, "json") || strings.Contains(contentTypeLower, "xml") {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore body for later use

		if len(body) == 0 {
			return nil, nil
		}

		var ids []string
		seenIDs := make(map[string]bool) // Track unique IDs

		if strings.Contains(contentTypeLower, "json") {
			ids, err = extractJSONIDs(body, bodyIDPaths, seenIDs)
			if err != nil {
				return nil, err
			}
		} else if strings.Contains(contentTypeLower, "xml") {
			ids, err = extractXMLIDs(body, bodyIDPaths, seenIDs)
			if err != nil {
				return nil, err
			}
		}

		if len(ids) > 0 {
			return ids, nil
		}
	}

	// For non-JSON requests or if no IDs found in body, try to extract from path
	pathSegments := getPathInfo(req.URL.Path)
	if len(pathSegments) > 0 {
		lastSegment := pathSegments[len(pathSegments)-1]
		if isValidID(lastSegment, sectionName) {
			return []string{lastSegment}, nil
		}
	}

	return nil, nil
}

// Helper functions moved from service
func getPathInfo(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

func isValidID(segment string, sectionName string) bool {
	// Check if it's a valid JSON string and not the section name
	_, err := json.Marshal(segment)
	return err == nil && segment != sectionName
}

func extractJSONIDs(body []byte, paths []string, seenIDs map[string]bool) ([]string, error) {
	doc, err := jsonquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %w", err)
	}

	var ids []string
	for _, path := range paths {
		nodes, err := jsonquery.QueryAll(doc, path)
		if err != nil {
			continue
		}
		for _, node := range nodes {
			if id := fmt.Sprintf("%v", node.Value()); id != "" && !seenIDs[id] {
				ids = append(ids, id)
				seenIDs[id] = true
			}
		}
	}

	return ids, nil
}

func extractXMLIDs(body []byte, paths []string, seenIDs map[string]bool) ([]string, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML body: %w", err)
	}

	var ids []string
	for _, path := range paths {
		nodes, err := xmlquery.QueryAll(doc, path)
		if err != nil {
			continue
		}
		for _, node := range nodes {
			if id := node.InnerText(); id != "" && !seenIDs[id] {
				ids = append(ids, id)
				seenIDs[id] = true
			}
		}
	}

	return ids, nil
}

// HandleRequest processes the HTTP request and returns appropriate response
func (h *MockHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Extract IDs from the request
	ids, err := h.ExtractIDs(r.Context(), r)
	if err != nil {
		h.logger.Error("failed to extract IDs", "error", err)
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read the request body if present
	var body []byte
	if r.Body != nil {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error("failed to read request body", "error", err)
			writeErrorResponse(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		// Restore body for potential future use
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// Process different HTTP methods
	var statusCode int
	var responseBody []byte
	var responseHeaders = make(http.Header)

	switch r.Method {
	case http.MethodGet:
		if len(ids) > 0 {
			// Get single resource
			data, err := h.service.GetResource(r.Context(), r.URL.Path, ids[0])
			if err != nil {
				h.logger.Error("failed to get resource", "error", err)
				statusCode = http.StatusNotFound
				responseBody = []byte(err.Error())
			} else {
				statusCode = http.StatusOK
				responseBody = data.Body
				responseHeaders.Set("Content-Type", data.ContentType)
			}
		} else {
			// Get collection
			data, err := h.service.GetResourcesByPath(r.Context(), r.URL.Path)
			if err != nil {
				h.logger.Error("failed to get resources", "error", err)
				statusCode = http.StatusInternalServerError
				responseBody = []byte(err.Error())
			} else {
				// Create a JSON array with the bodies
				var rawBodies []interface{}

				// Sort the data by path for consistent ordering
				sort.Slice(data, func(i, j int) bool {
					return data[i].Path < data[j].Path
				})

				for _, item := range data {
					if strings.Contains(strings.ToLower(item.ContentType), "json") {
						var jsonData interface{}
						if err := json.Unmarshal(item.Body, &jsonData); err == nil {
							rawBodies = append(rawBodies, jsonData)
						} else {
							rawBodies = append(rawBodies, string(item.Body))
						}
					}
				}

				collectionBody, err := json.Marshal(rawBodies)
				if err != nil {
					h.logger.Error("failed to marshal collection", "error", err)
					statusCode = http.StatusInternalServerError
					responseBody = []byte("Failed to process collection")
				} else {
					statusCode = http.StatusOK
					responseBody = collectionBody
					responseHeaders.Set("Content-Type", "application/json")
				}
			}
		}

	case http.MethodPost:
		// Create resource
		data := &model.MockData{
			Path:        r.URL.Path,
			ContentType: r.Header.Get("Content-Type"),
			Body:        body,
		}

		err := h.service.CreateResource(r.Context(), r.URL.Path, ids, data)
		if err != nil {
			h.logger.Error("failed to create resource", "error", err)
			statusCode = http.StatusInternalServerError
			responseBody = []byte(err.Error())
		} else {
			statusCode = http.StatusCreated
			responseBody = data.Body
			responseHeaders.Set("Content-Type", data.ContentType)
		}

	case http.MethodPut:
		// Update resource
		if len(ids) == 0 {
			h.logger.Error("no ID found in PUT request")
			statusCode = http.StatusBadRequest
			responseBody = []byte("No ID found in request")
		} else {
			data := &model.MockData{
				Path:        r.URL.Path,
				ContentType: r.Header.Get("Content-Type"),
				Body:        body,
			}

			err := h.service.UpdateResource(r.Context(), r.URL.Path, ids[0], data)
			if err != nil {
				h.logger.Error("failed to update resource", "error", err)
				statusCode = http.StatusInternalServerError
				responseBody = []byte(err.Error())
			} else {
				statusCode = http.StatusOK
				responseBody = data.Body
				responseHeaders.Set("Content-Type", data.ContentType)
			}
		}

	case http.MethodDelete:
		// Delete resource
		if len(ids) == 0 {
			h.logger.Error("no ID found in DELETE request")
			statusCode = http.StatusBadRequest
			responseBody = []byte("No ID found in request")
		} else {
			err := h.service.DeleteResource(r.Context(), r.URL.Path, ids[0])
			if err != nil {
				h.logger.Error("failed to delete resource", "error", err)
				statusCode = http.StatusInternalServerError
				responseBody = []byte(err.Error())
			} else {
				statusCode = http.StatusNoContent
			}
		}

	default:
		statusCode = http.StatusMethodNotAllowed
		responseBody = []byte("Method not allowed")
	}

	// Copy response headers
	for key, values := range responseHeaders {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(statusCode)

	// Write response body
	if responseBody != nil {
		if _, err := w.Write(responseBody); err != nil {
			h.logger.Error("failed to write response body", "error", err)
		}
	}
}

// ServeHTTP implements the http.Handler interface
func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandleRequest(w, r)
}
