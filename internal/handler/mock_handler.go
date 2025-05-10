package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/google/uuid"
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

// extractIDs extracts IDs from the request using configured paths
func (h *MockHandler) extractIDs(ctx context.Context, req *http.Request) ([]string, error) {
	// Find matching section
	var sectionName string
	var section *config.Section
	var err error

	if h.mockCfg != nil {
		// Use the config
		sectionName, section, err = h.mockCfg.MatchPath(req.URL.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to match path pattern: %w", err)
		}
		if section == nil {
			return nil, fmt.Errorf("invalid path")
		}
	} else {
		return nil, fmt.Errorf("service configuration is missing")
	}

	// For GET/PUT/DELETE requests, try to extract ID from path first
	if req.Method == http.MethodGet || req.Method == http.MethodPut || req.Method == http.MethodDelete {
		pathSegments := getPathInfo(req.URL.Path)
		patternSegments := getPathInfo(section.PathPattern)

		// Check if this is a resource path (contains an ID)
		if len(pathSegments) > len(patternSegments) ||
			(len(patternSegments) > 0 && len(pathSegments) > 0 &&
				patternSegments[len(patternSegments)-1] == "*" &&
				len(pathSegments) == len(patternSegments)) {

			// Use the last path segment as the ID
			lastSegment := pathSegments[len(pathSegments)-1]
			if lastSegment != "" && lastSegment != sectionName {
				return []string{lastSegment}, nil
			}
		}

		// If we got here, it's a collection path without an ID
		return nil, nil
	}

	// For POST requests, try to extract ID from header first
	if section.HeaderIDName != "" {
		if id := req.Header.Get(section.HeaderIDName); id != "" {
			if id != sectionName {
				return []string{id}, nil
			}
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
			return nil, fmt.Errorf("no IDs found in request")
		}

		var ids []string
		seenIDs := make(map[string]bool) // Track unique IDs

		if strings.Contains(contentTypeLower, "json") {
			ids, err = extractJSONIDs(body, section.BodyIDPaths, seenIDs)
			if err != nil {
				return nil, fmt.Errorf("invalid JSON")
			}
		} else if strings.Contains(contentTypeLower, "xml") {
			ids, err = extractXMLIDs(body, section.BodyIDPaths, seenIDs)
			if err != nil {
				return nil, fmt.Errorf("invalid XML")
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
		if lastSegment != "" && lastSegment != sectionName {
			return []string{lastSegment}, nil
		}
	}

	return nil, fmt.Errorf("no IDs found in request")
}

// Helper functions moved from service
func getPathInfo(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
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
func (h *MockHandler) HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Normalize path by removing trailing slash
	req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")

	// Find matching section
	var section *config.Section
	var sectionName string
	var err error

	if h.mockCfg != nil {
		// Use the config
		sectionName, section, err = h.mockCfg.MatchPath(req.URL.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to match path pattern: %w", err)
		}
		if section == nil {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}, nil
		}
	} else {
		return nil, fmt.Errorf("service configuration is missing")
	}

	// Handle different HTTP methods
	switch req.Method {
	case http.MethodGet:
		// First try to get resource by ID
		pathSegments := getPathInfo(req.URL.Path)
		lastSegment := pathSegments[len(pathSegments)-1]

		if lastSegment != "" && lastSegment != sectionName {
			// Try to get resource by ID
			resource, err := h.service.GetResource(ctx, req.URL.Path, lastSegment)
			if err == nil && resource != nil {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{resource.ContentType}},
					Body:       io.NopCloser(bytes.NewReader(resource.Body)),
				}, nil
			}
		}

		// If resource not found by ID, try to get collection at the exact path
		resources, err := h.service.GetResourcesByPath(ctx, req.URL.Path)
		if err == nil && len(resources) > 0 {
			var items [][]byte
			for _, r := range resources {
				if strings.Contains(strings.ToLower(r.ContentType), "json") {
					items = append(items, r.Body)
				}
			}
			var responseBody []byte
			if len(items) == 1 {
				responseBody = append([]byte("["), items[0]...)
				responseBody = append(responseBody, byte(']'))
			} else if len(items) > 1 {
				responseBody = append(responseBody, byte('['))
				for i, item := range items {
					responseBody = append(responseBody, item...)
					if i < len(items)-1 {
						responseBody = append(responseBody, []byte(",")...)
					}
				}
				responseBody = append(responseBody, byte(']'))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		}

		// If neither resource nor collection found, return 404
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("resource not found")),
		}, nil

	case http.MethodPost:
		// Accept any content type
		// Extract IDs from request
		ids, err := h.extractIDs(ctx, req)
		if err != nil {
			if err.Error() == "no IDs found in request" {
				// If no ID found, generate a new UUID
				generatedUUID := uuid.New().String()
				ids = []string{generatedUUID}
				err = nil // Clear the error, as we've handled it by generating an ID
				h.logger.Debug("POST: no ID found in request, generated new UUID", "uuid", generatedUUID)
			} else {
				// For other errors from extractIDs (e.g., invalid JSON), return BadRequest
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader(err.Error())),
				}, nil
			}
		} else if len(ids) == 0 { // Should ideally be caught by extractIDs error, but as a safeguard
			generatedUUID := uuid.New().String()
			ids = []string{generatedUUID}
			h.logger.Debug("POST: extractIDs returned no error but empty IDs, generated new UUID", "uuid", generatedUUID)
		}

		// Read request body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}

		// Create resource
		contentType := req.Header.Get("Content-Type")
		data := &model.MockData{
			Path:        req.URL.Path,
			ContentType: contentType,
			Body:        body,
		}

		if err := h.service.CreateResource(ctx, req.URL.Path, ids, data); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return &http.Response{
					StatusCode: http.StatusConflict,
					Body:       io.NopCloser(strings.NewReader(err.Error())),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(err.Error())),
			}, nil
		}

		// Set Location header to the resource path
		location := req.URL.Path
		if len(ids) > 0 {
			location = fmt.Sprintf("%s/%s", strings.TrimSuffix(location, "/"), ids[0])
		}

		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     http.Header{"Location": []string{location}},
		}, nil

	case http.MethodPut:
		// Try to update resource by ID
		pathSegments := getPathInfo(req.URL.Path)
		lastSegment := pathSegments[len(pathSegments)-1]

		if lastSegment == "" || lastSegment == sectionName {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}, nil
		}

		existingResource, err := h.service.GetResource(ctx, req.URL.Path, lastSegment)

		if err != nil || existingResource == nil {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}, nil
		}

		// Read request body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}

		// Try to update resource
		err = h.service.UpdateResource(ctx, req.URL.Path, lastSegment, &model.MockData{
			Path:        req.URL.Path,
			ContentType: req.Header.Get("Content-Type"),
			Body:        body,
		})

		if err != nil {
			if err.Error() == "resource not found" {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("resource not found")),
				}, nil
			}
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
		}, nil

	case http.MethodDelete:
		// Try to delete resource by ID
		pathSegments := getPathInfo(req.URL.Path)
		lastSegment := pathSegments[len(pathSegments)-1]

		if lastSegment == "" || lastSegment == sectionName {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}, nil
		}

		existingResource, err := h.service.GetResource(ctx, req.URL.Path, lastSegment)

		if err != nil || existingResource == nil {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}, nil
		}

		// Try to delete resource
		err = h.service.DeleteResource(ctx, req.URL.Path, lastSegment)
		if err != nil {
			if err.Error() == "resource not found" {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("resource not found")),
				}, nil
			}
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusNoContent,
		}, nil

	default:
		return &http.Response{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       io.NopCloser(strings.NewReader("method not allowed")),
		}, nil
	}
}

// ServeHTTP implements the http.Handler interface
func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	resp, err := h.HandleRequest(ctx, r)
	if err != nil {
		h.logger.Error("failed to handle request", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	// Write response body
	if resp.Body != nil {
		if _, err := io.Copy(w, resp.Body); err != nil {
			h.logger.Error("failed to write response body", "error", err)
		}
	}
}

// GetConfig returns the mock configuration
func (h *MockHandler) GetConfig() *config.MockConfig {
	return h.mockCfg
}
