package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	service         service.MockService
	scenarioService service.ScenarioService
	logger          *slog.Logger
	mockCfg         *config.MockConfig
}

// NewMockHandler creates a new instance of Handler
func NewMockHandler(service service.MockService, scenarioService service.ScenarioService, logger *slog.Logger, cfg *config.MockConfig) *MockHandler {
	return &MockHandler{
		service:         service,
		scenarioService: scenarioService,
		logger:          logger,
		mockCfg:         cfg,
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
	var collectedIDs []string
	seenIDs := make(map[string]bool)

	// Helper to add an ID if it's valid and not already seen
	addID := func(id string) {
		if id != "" && id != sectionName && !seenIDs[id] {
			collectedIDs = append(collectedIDs, id)
			seenIDs[id] = true
		}
	}

	// 1. Try to extract ID from path (primary for GET/PUT/DELETE, potential for POST)
	pathSegments := getPathInfo(req.URL.Path)
	if req.Method == http.MethodGet || req.Method == http.MethodPut || req.Method == http.MethodDelete {
		patternSegments := getPathInfo(section.PathPattern)
		if len(pathSegments) > len(patternSegments) || (len(patternSegments) > 0 && len(pathSegments) > 0 && patternSegments[len(patternSegments)-1] == "*" && len(pathSegments) == len(patternSegments)) {
			lastSegment := pathSegments[len(pathSegments)-1]
			addID(lastSegment)
			// For GET/PUT/DELETE, if path ID is found, it's usually the only one we care about for addressing the resource.
			// However, REQ-RM-MULTI-ID implies any ID can be used. Here, we assume the path ID is the target.
			// If `collectedIDs` has this path ID, we can return it directly for these methods.
			if len(collectedIDs) > 0 {
				return collectedIDs, nil
			}
		}
		// If no path ID for GET/PUT/DELETE, it's likely a collection operation or error, return empty or let POST logic run if applicable.
		// For now, if it's strictly GET/PUT/DELETE and no path ID, return no IDs.
		return nil, nil
	}

	// For POST requests (or if other methods fall through, though less likely with above return)
	// 2. Try primary Header ID Name
	if section.HeaderIDName != "" {
		addID(req.Header.Get(section.HeaderIDName))
	}

	// 4. Try Body IDs (Primary and Alternatives)
	contentType := req.Header.Get("Content-Type")
	contentTypeLower := strings.ToLower(contentType)
	if strings.Contains(contentTypeLower, "json") || strings.Contains(contentTypeLower, "xml") {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore body

		if len(body) > 0 {
			var bodyExtractionError error
			var extractedFromBody []string

			// Use only primary BodyIDPaths as alternative paths were removed
			primaryBodyIdPaths := section.BodyIDPaths

			if strings.Contains(contentTypeLower, "json") {
				extractedFromBody, bodyExtractionError = extractJSONIDs(body, primaryBodyIdPaths, seenIDs)
			} else if strings.Contains(contentTypeLower, "xml") {
				extractedFromBody, bodyExtractionError = extractXMLIDs(body, primaryBodyIdPaths, seenIDs)
			}

			if bodyExtractionError != nil {
				h.logger.WarnContext(ctx, "error extracting some IDs from body", "error", bodyExtractionError)
				// If body extraction itself fails (e.g. malformed JSON/XML), this should be a hard error.
				return nil, bodyExtractionError // Propagate body parsing errors
			}
			for _, id := range extractedFromBody { // Add successfully extracted and newly seen IDs
				addID(id) // addID handles uniqueness via seenIDs
			}
		}
	}

	// 5. Fallback for POST: if still no IDs, try last path segment (e.g., POST /collection/newIdToCreate)
	if req.Method == http.MethodPost && len(collectedIDs) == 0 {
		if len(pathSegments) > 0 {
			lastSegment := pathSegments[len(pathSegments)-1]
			// Check if lastSegment looks like a collection name vs an ID
			// This logic might need refinement based on how PathPattern is defined (e.g. /collection/* vs /collection)
			patternSegments := getPathInfo(section.PathPattern)
			if len(pathSegments) > len(patternSegments) || (len(pathSegments) == len(patternSegments) && strings.HasSuffix(section.PathPattern, "*")) {
				addID(lastSegment)
			}
		}
	}

	if len(collectedIDs) == 0 {
		// For POST, this signals to the caller to generate a UUID.
		// For other methods, if it reaches here, it implies no specific resource ID was found by path.
		return nil, nil // No IDs found, but not an error in extraction itself.
	}

	return collectedIDs, nil
}

// Helper functions moved from service
func getPathInfo(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

func extractJSONIDs(body []byte, idPaths []string, seenIDs map[string]bool) ([]string, error) {
	doc, err := jsonquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %w", err)
	}

	var ids []string
	for _, path := range idPaths {
		nodes, err := jsonquery.QueryAll(doc, path)
		if err != nil {
			continue
		}
		for _, node := range nodes {
			if idStr := fmt.Sprintf("%v", node.Value()); idStr != "" {
				if !seenIDs[idStr] {
					ids = append(ids, idStr)
				}
			}
		}
	}
	return ids, nil
}

func extractXMLIDs(body []byte, idPaths []string, seenIDs map[string]bool) ([]string, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML body: %w", err)
	}

	var ids []string
	for _, path := range idPaths {
		nodes, err := xmlquery.QueryAll(doc, path)
		if err != nil {
			continue
		}
		for _, node := range nodes {
			if idStr := node.InnerText(); idStr != "" {
				if !seenIDs[idStr] {
					ids = append(ids, idStr)
				}
			}
		}
	}
	return ids, nil
}

// HandleRequest processes the HTTP request and returns appropriate response
func (h *MockHandler) HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Normalize path by removing trailing slash
	req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")

	// Check for matching scenario first
	// Construct requestPath string (e.g., "GET /api/users")
	requestPath := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	if scenario, err := h.scenarioService.FindScenarioByRequestPath(ctx, requestPath); err == nil && scenario != nil {
		h.logger.Info("scenario matched", "requestPath", requestPath, "scenarioUUID", scenario.UUID)
		responseHeaders := http.Header{}
		if scenario.ContentType != "" {
			responseHeaders.Set("Content-Type", scenario.ContentType)
		}
		if scenario.Location != "" {
			responseHeaders.Set("Location", scenario.Location)
		}
		return &http.Response{
			StatusCode: scenario.StatusCode,
			Header:     responseHeaders,
			Body:       io.NopCloser(strings.NewReader(scenario.Data)),
		}, nil
	} else if err != nil {
		// Log the error but continue with normal mock handling as scenario finding is optional
		h.logger.Error("error finding scenario by request path", "requestPath", requestPath, "error", err)
	}

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
		var id string                   // Definitive ID for Location header, logging
		var createdResourceIDs []string // All IDs to be passed to storage.Create

		idsFromExtractor, err := h.extractIDs(ctx, req, section, sectionName)
		if err != nil { // An actual error occurred during extraction (e.g., bad JSON, IO error)
			h.logger.Error("failed to extract IDs for POST", "path", req.URL.Path, "error", err)
			return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("invalid request: " + err.Error()))}, nil
		} else if len(idsFromExtractor) == 0 { // No error from extractIDs, and no IDs were found from any source
			generatedUUID := uuid.New().String()
			id = generatedUUID
			createdResourceIDs = []string{id} // Use generated UUID for storage
			h.logger.Debug("POST: no IDs found by extractIDs, generated new UUID", "uuid", id)
		} else { // One or more IDs were successfully extracted by extractIDs
			id = idsFromExtractor[0]              // Use the first ID for Location header and logging
			createdResourceIDs = idsFromExtractor // Use all extracted IDs for storage
			if len(idsFromExtractor) > 1 {
				h.logger.Info("multiple IDs found in POST request, using first for Location, all for creation",
					"path", req.URL.Path, "all_ids", createdResourceIDs, "location_id", id)
			} else {
				h.logger.Debug("POST: using extracted ID(s)", "path", req.URL.Path, "ids", createdResourceIDs)
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
		bodyBytes, errIOReadAll := io.ReadAll(req.Body)
		if errIOReadAll != nil {
			h.logger.Error("failed to read request body for PUT", "path", req.URL.Path, "error", errIOReadAll.Error())
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("failed to read request body"))}, fmt.Errorf("reading body: %w", errIOReadAll)
		}

		contentType := req.Header.Get("Content-Type")
		if contentType == "" {
			h.logger.Error("Content-Type header is missing for PUT request", "path", req.URL.Path)
			return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("Content-Type header is missing"))}, nil
		}

		// Pass section and sectionName to extractIDs
		ids, err := h.extractIDs(ctx, req, section, sectionName)
		if err != nil || len(ids) == 0 {
			h.logger.Error("ID not found in PUT request", "path", req.URL.Path, "error", err)
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}, nil
		}

		putData := &model.MockData{
			Path:        req.URL.Path, // Or construct specific path using ids[0]
			ContentType: contentType,
			Body:        bodyBytes,
		}

		// Try to update resource (upsert logic is in the service)
		err = h.service.UpdateResource(ctx, req.URL.Path, ids[0], putData)

		if err != nil {
			if strings.Contains(err.Error(), "resource not found") { // Check if the service layer specifically returns this for a failed upsert's "find" part
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("resource not found after upsert attempt")),
				}, nil
			}
			// For other errors during update/create
			h.logger.Error("failed to update/create resource for PUT", "path", req.URL.Path, "id", ids[0], "error", err)
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("failed to process PUT request"))}, fmt.Errorf("processing PUT: %w", err)
		}

		// After successful upsert, fetch the resource to return it
		updatedResource, getErr := h.service.GetResource(ctx, req.URL.Path, ids[0])
		if getErr != nil {
			h.logger.Error("failed to get resource after PUT", "path", req.URL.Path, "id", ids[0], "error", getErr)
			// If we successfully upserted but can't get it, this is an internal error.
			// The resource should exist.
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("failed to retrieve resource after update"))}, fmt.Errorf("getting resource post-PUT: %w", getErr)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{updatedResource.ContentType}},
			Body:       io.NopCloser(bytes.NewReader(updatedResource.Body)),
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
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				h.logger.Warn("error closing response body in ServeHTTP defer", "error", closeErr)
			}
		}()
	}

	// Copy headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	if resp.Body != nil {
		bodyBytesToRespond, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			h.logger.Error("failed to read response body from HandleRequest's response", "error", readErr)
			// Set a server error status if possible, though headers might be sent
			if rr, ok := w.(*httptest.ResponseRecorder); ok && !rr.Flushed { // Check if WriteHeader has been called
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		// Explicitly set Content-Length if not already set by resp.Header
		if w.Header().Get("Content-Length") == "" {
			w.Header().Set("Content-Length", strconv.Itoa(len(bodyBytesToRespond)))
		}
		w.WriteHeader(resp.StatusCode) // Set status code AFTER Content-Length potentially

		if len(bodyBytesToRespond) > 0 {
			_, writeErr := w.Write(bodyBytesToRespond)
			if writeErr != nil {
				h.logger.Error("failed to write response body", "error", writeErr)
			}
		}
	} else {
		// If resp.Body is nil (e.g., for 204 No Content or if HandleRequest returns no body for other reasons)
		// still need to write the status code.
		w.WriteHeader(resp.StatusCode)
	}
}

// GetConfig returns the mock configuration
func (h *MockHandler) GetConfig() *config.MockConfig {
	return h.mockCfg
}
