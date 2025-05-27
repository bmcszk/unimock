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

// getSectionForRequest finds the matching configuration section for a given request path.
func (h *MockHandler) getSectionForRequest(reqPath string) (*config.Section, string, error) {
	if h.mockCfg == nil {
		return nil, "", fmt.Errorf("service configuration is missing")
	}
	sectionName, section, err := h.mockCfg.MatchPath(reqPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to match path pattern: %w", err)
	}
	if section == nil {
		return nil, "", fmt.Errorf("no matching section found for path: %s", reqPath)
	}
	return section, sectionName, nil
}

// extractIDs extracts IDs from the request using configured paths
func (h *MockHandler) extractIDs(ctx context.Context, req *http.Request, section *config.Section, sectionName string) ([]string, error) {
	// For GET/PUT/DELETE requests, try to extract ID from path first
	if req.Method == http.MethodGet || req.Method == http.MethodPut || req.Method == http.MethodDelete {
		pathSegments := getPathInfo(req.URL.Path)
		patternSegments := getPathInfo(section.PathPattern)

		// Check if this is a resource path (contains an ID)
		if len(pathSegments) > len(patternSegments) || (len(patternSegments) > 0 && len(pathSegments) > 0 && patternSegments[len(patternSegments)-1] == "*" && len(pathSegments) == len(patternSegments)) {
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
	section, sectionName, err := h.getSectionForRequest(req.URL.Path)
	if err != nil {
		// If no matching section, typically a 404 or specific error from getSectionForRequest
		// The router should ideally prevent this, but if it happens:
		h.logger.Warn("no matching section in HandleRequest", "path", req.URL.Path, "error", err)
		return &http.Response{
			StatusCode: http.StatusNotFound, // Or http.StatusBadRequest depending on error
			Body:       io.NopCloser(strings.NewReader(err.Error())),
		}, nil
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
		var id string                   // This will be the definitive ID for the resource
		var createdResourceIDs []string // IDs to be used for storage

		// Pass section and sectionName to extractIDs
		idsFromExtractor, err := h.extractIDs(ctx, req, section, sectionName)
		if err != nil {
			if err.Error() == "no IDs found in request" {
				generatedUUID := uuid.New().String()
				id = generatedUUID                // Assign to id for Location header and logging
				createdResourceIDs = []string{id} // Use generated ID for storage
				err = nil                         // Clear the error, as we've handled it by generating an ID
				h.logger.Debug("POST: no ID found in request, generated new UUID", "uuid", id)
			} else {
				h.logger.Error("failed to extract IDs for POST", "path", req.URL.Path, "error", err)
				return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("invalid request: " + err.Error()))}, nil
			}
		} else if len(idsFromExtractor) == 0 { // No error, but no IDs found (e.g. empty body for relevant content types)
			generatedUUID := uuid.New().String()
			id = generatedUUID                // Assign to id for Location header and logging
			createdResourceIDs = []string{id} // Use generated ID for storage
			h.logger.Debug("POST: extractIDs returned no error but empty IDs, generated new UUID", "uuid", id)
		} else {
			id = idsFromExtractor[0]              // Use the first ID for Location header and logging if multiple are found
			createdResourceIDs = idsFromExtractor // Use all extracted IDs for storage if that's the intent for batch
			if len(idsFromExtractor) > 1 {
				h.logger.Warn("multiple IDs found in POST request, using the first for Location header, all for creation", "path", req.URL.Path, "all_ids", idsFromExtractor, "location_id", id)
			}
		}

		// Read the body for storage
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

		if err := h.service.CreateResource(ctx, req.URL.Path, createdResourceIDs, data); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return &http.Response{
					StatusCode: http.StatusConflict,
					Body:       io.NopCloser(strings.NewReader(err.Error())),
				}, nil
			}
			h.logger.Error("failed to create resource for POST", "path", req.URL.Path, "id", id, "error", err)
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("failed to create resource"))}, nil
		}
		h.logger.Info("resource created via POST", "path", req.URL.Path, "id", id)

		location := req.URL.Path
		// Ensure a single ID is used for the location header
		location = fmt.Sprintf("%s/%s", strings.TrimSuffix(location, "/"), id)

		// For Unimock, we'll return 201 with Location and empty body, consistent with typical REST APIs.
		// The logic for handling CollectionJSON here was primarily for logging or specific response shaping
		// if it were different from the standard POST response. Since POST creates a single resource,
		// and the Location header points to that single resource, complex CollectionJSON logic isn't needed here.

		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     http.Header{"Location": []string{location}},
			Body:       io.NopCloser(strings.NewReader("")), // Empty body for 201
		}, nil

	case http.MethodPut:
		// Pass section and sectionName to extractIDs
		ids, err := h.extractIDs(ctx, req, section, sectionName)
		if err != nil || len(ids) == 0 {
			h.logger.Error("ID not found in PUT request", "path", req.URL.Path, "error", err)
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
		err = h.service.UpdateResource(ctx, req.URL.Path, ids[0], &model.MockData{
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
		// Pass section and sectionName to extractIDs
		ids, err := h.extractIDs(ctx, req, section, sectionName)
		if err != nil || len(ids) == 0 {
			h.logger.Error("ID not found in DELETE request", "path", req.URL.Path, "error", err)
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}, nil
		}

		// Try to delete resource
		err = h.service.DeleteResource(ctx, req.URL.Path, ids[0])
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

	if resp != nil && resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
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
