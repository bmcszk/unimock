package handler

import (
	"bytes"
	"context"
	"errors"
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

const (
	// lastElementIndex represents the offset to get the last element of a slice
	lastElementIndex = 1
	
	// Common strings
	errorLogKey = "error"
	pathLogKey  = "path"
	idLogKey    = "id"
	contentTypeHeader = "Content-Type"
	pathSeparator = "/"
	resourceNotFoundMsg = "resource not found"
	
	// Array sizes
	singleItem = 1
)

// MockHandler represents our HTTP request handler
type MockHandler struct {
	service         service.MockService
	scenarioService service.ScenarioService
	logger          *slog.Logger
	mockCfg         *config.MockConfig
}

// NewMockHandler creates a new instance of Handler
func NewMockHandler(
	mockService service.MockService, 
	scenarioService service.ScenarioService, 
	logger *slog.Logger, 
	cfg *config.MockConfig,
) *MockHandler {
	return &MockHandler{
		service:         mockService,
		scenarioService: scenarioService,
		logger:          logger,
		mockCfg:         cfg,
	}
}

// getSectionForRequest finds the matching configuration section for a given request path.
func (h *MockHandler) getSectionForRequest(reqPath string) (*config.Section, string, error) {
	if h.mockCfg == nil {
		return nil, "", errors.New("service configuration is missing")
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

// transformRequest applies request transformations to MockData before processing
func (h *MockHandler) transformRequest(
	data *model.MockData,
	section *config.Section,
	sectionName string,
) (*model.MockData, error) {
	// Skip if no transformations configured
	if section.Transformations == nil || !section.Transformations.HasRequestTransforms() {
		return data, nil
	}

	currentData := data
	for i, transform := range section.Transformations.RequestTransforms {
		h.logger.Debug("applying request transformation", "section", sectionName, "transform_index", i)
		
		// Execute transformation with panic recovery
		transformedData, err := h.safeExecuteRequestTransform(transform, currentData, sectionName)
		if err != nil {
			h.logger.Error(
				"request transformation failed",
				"section", sectionName,
				"transform_index", i,
				errorLogKey, err,
			)
			return nil, fmt.Errorf("request transformation failed: %w", err)
		}
		
		currentData = transformedData
	}

	return currentData, nil
}

// transformResponse applies response transformations to MockData after processing
func (h *MockHandler) transformResponse(
	data *model.MockData,
	section *config.Section,
	sectionName string,
) (*model.MockData, error) {
	// Skip if no transformations configured
	if section.Transformations == nil || !section.Transformations.HasResponseTransforms() {
		return data, nil
	}

	currentData := data
	for i, transform := range section.Transformations.ResponseTransforms {
		h.logger.Debug("applying response transformation", "section", sectionName, "transform_index", i)
		
		// Execute transformation with panic recovery
		transformedData, err := h.safeExecuteResponseTransform(transform, currentData, sectionName)
		if err != nil {
			h.logger.Error(
				"response transformation failed",
				"section", sectionName,
				"transform_index", i,
				errorLogKey, err,
			)
			return nil, fmt.Errorf("response transformation failed: %w", err)
		}
		
		currentData = transformedData
	}

	return currentData, nil
}

// safeExecuteRequestTransform executes a request transformation function with panic recovery
func (h *MockHandler) safeExecuteRequestTransform(
	transform config.RequestTransformFunc,
	data *model.MockData,
	sectionName string,
) (transformedData *model.MockData, err error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("request transformation panicked", "section", sectionName, "panic", r)
			err = fmt.Errorf("request transformation panicked: %v", r)
			transformedData = nil
		}
	}()

	return transform(data)
}

// safeExecuteResponseTransform executes a response transformation function with panic recovery
func (h *MockHandler) safeExecuteResponseTransform(
	transform config.ResponseTransformFunc,
	data *model.MockData,
	sectionName string,
) (transformedData *model.MockData, err error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("response transformation panicked", "section", sectionName, "panic", r)
			err = fmt.Errorf("response transformation panicked: %v", r)
			transformedData = nil
		}
	}()

	return transform(data)
}


// extractIDs extracts IDs from the request using configured paths
func (h *MockHandler) extractIDs(
	ctx context.Context, 
	req *http.Request, 
	section *config.Section, 
	sectionName string,
) ([]string, error) {
	var collectedIDs []string
	seenIDs := make(map[string]bool)

	// Helper to add an ID if it's valid and not already seen
	addID := func(id string) {
		if id != "" && id != sectionName && !seenIDs[id] {
			collectedIDs = append(collectedIDs, id)
			seenIDs[id] = true
		}
	}

	// Handle path-based ID extraction for GET/PUT/DELETE methods
	if req.Method == http.MethodGet || req.Method == http.MethodPut || req.Method == http.MethodDelete {
		return h.extractPathIDs(req, section, addID)
	}

	// For POST requests, extract IDs from headers and body
	return h.extractPostIDs(ctx, req, section, addID, collectedIDs)
}

// extractPathIDs extracts IDs from the request path for GET/PUT/DELETE methods
func (*MockHandler) extractPathIDs(req *http.Request, section *config.Section, addID func(string)) ([]string, error) {
	var collectedIDs []string
	
	pathSegments := getPathInfo(req.URL.Path)
	patternSegments := getPathInfo(section.PathPattern)
	
	if len(pathSegments) > len(patternSegments) || 
		(len(patternSegments) > 0 && len(pathSegments) > 0 && 
		 patternSegments[len(patternSegments)-lastElementIndex] == "*" && 
		 len(pathSegments) == len(patternSegments)) {
		lastSegment := pathSegments[len(pathSegments)-lastElementIndex]
		addID(lastSegment)
		
		// For GET/PUT/DELETE, if path ID is found, return it directly
		if lastSegment != "" {
			collectedIDs = append(collectedIDs, lastSegment)
			return collectedIDs, nil
		}
	}
	
	// If no path ID for GET/PUT/DELETE, return empty
	return nil, nil
}

// extractPostIDs extracts IDs from headers and body for POST requests
func (h *MockHandler) extractPostIDs(
	ctx context.Context, 
	req *http.Request, 
	section *config.Section, 
	addID func(string), 
	collectedIDs []string,
) ([]string, error) {
	collectedIDs = h.tryExtractHeaderID(section, req, addID, collectedIDs)
	
	bodyIDs, err := h.extractBodyIDs(ctx, req, section, addID)
	if err != nil {
		return nil, err
	}
	collectedIDs = append(collectedIDs, bodyIDs...)

	if len(collectedIDs) == 0 {
		collectedIDs = h.tryExtractPathID(req, section, addID, collectedIDs)
	}

	if len(collectedIDs) == 0 {
		return nil, nil
	}

	return collectedIDs, nil
}

// tryExtractHeaderID attempts to extract ID from request headers
func (*MockHandler) tryExtractHeaderID(
	section *config.Section, 
	req *http.Request, 
	addID func(string), 
	collectedIDs []string,
) []string {
	if section.HeaderIDName != "" {
		headerID := req.Header.Get(section.HeaderIDName)
		if headerID != "" {
			addID(headerID)
			collectedIDs = append(collectedIDs, headerID)
		}
	}
	return collectedIDs
}

// tryExtractPathID attempts to extract ID from request path
func (h *MockHandler) tryExtractPathID(
	req *http.Request, 
	section *config.Section, 
	addID func(string), 
	collectedIDs []string,
) []string {
	pathSegments := getPathInfo(req.URL.Path)
	if len(pathSegments) == 0 {
		return collectedIDs
	}

	lastSegment := pathSegments[len(pathSegments)-lastElementIndex]
	patternSegments := getPathInfo(section.PathPattern)
	
	if h.shouldUsePathSegment(pathSegments, patternSegments, section.PathPattern) {
		addID(lastSegment)
		collectedIDs = append(collectedIDs, lastSegment)
	}
	
	return collectedIDs
}

// shouldUsePathSegment determines if path segment should be used as ID
func (*MockHandler) shouldUsePathSegment(pathSegments, patternSegments []string, pathPattern string) bool {
	return len(pathSegments) > len(patternSegments) || 
		(len(pathSegments) == len(patternSegments) && strings.HasSuffix(pathPattern, "*"))
}

// extractBodyIDs extracts IDs from request body
func (h *MockHandler) extractBodyIDs(
	ctx context.Context, 
	req *http.Request, 
	section *config.Section, 
	addID func(string),
) ([]string, error) {
	contentTypeLower := strings.ToLower(req.Header.Get("Content-Type"))
	
	if !h.isSupportedContentType(contentTypeLower) {
		return nil, nil
	}

	body, err := h.readAndRestoreBody(req)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, nil
	}

	return h.extractIDsFromBody(ctx, body, contentTypeLower, section, addID)
}

// isSupportedContentType checks if content type is supported for ID extraction
func (*MockHandler) isSupportedContentType(contentType string) bool {
	return strings.Contains(contentType, "json") || strings.Contains(contentType, "xml")
}

// readAndRestoreBody reads request body and restores it for later use
func (*MockHandler) readAndRestoreBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// extractIDsFromBody extracts IDs from parsed body content
func (h *MockHandler) extractIDsFromBody(
	ctx context.Context, 
	body []byte, 
	contentType string, 
	section *config.Section, 
	addID func(string),
) ([]string, error) {
	var extractedFromBody []string
	var err error
	seenIDs := make(map[string]bool)

	if strings.Contains(contentType, "json") {
		extractedFromBody, err = extractJSONIDs(body, section.BodyIDPaths, seenIDs)
	} else {
		extractedFromBody, err = extractXMLIDs(body, section.BodyIDPaths, seenIDs)
	}

	if err != nil {
		h.logger.WarnContext(ctx, "error extracting some IDs from body", "error", err)
		return nil, err
	}

	for _, id := range extractedFromBody {
		addID(id)
	}

	return extractedFromBody, nil
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
		pathIDs := extractJSONIDsFromPath(doc, path, seenIDs)
		ids = append(ids, pathIDs...)
	}
	return ids, nil
}

func extractJSONIDsFromPath(doc *jsonquery.Node, path string, seenIDs map[string]bool) []string {
	nodes, err := jsonquery.QueryAll(doc, path)
	if err != nil {
		return nil
	}

	var ids []string
	for _, node := range nodes {
		if idStr := fmt.Sprintf("%v", node.Value()); idStr != "" {
			if !seenIDs[idStr] {
				ids = append(ids, idStr)
			}
		}
	}
	return ids
}

func extractXMLIDs(body []byte, idPaths []string, seenIDs map[string]bool) ([]string, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML body: %w", err)
	}

	var ids []string
	for _, path := range idPaths {
		pathIDs := extractXMLIDsFromPath(doc, path, seenIDs)
		ids = append(ids, pathIDs...)
	}
	return ids, nil
}

func extractXMLIDsFromPath(doc *xmlquery.Node, path string, seenIDs map[string]bool) []string {
	nodes, err := xmlquery.QueryAll(doc, path)
	if err != nil {
		return nil
	}

	var ids []string
	for _, node := range nodes {
		if idStr := node.InnerText(); idStr != "" {
			if !seenIDs[idStr] {
				ids = append(ids, idStr)
			}
		}
	}
	return ids
}

// HandleRequest processes the HTTP request and returns appropriate response
func (h *MockHandler) HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")

	section, sectionName, err := h.getSectionForRequest(req.URL.Path)
	if err != nil {
		h.logger.Warn("no matching section in HandleRequest", pathLogKey, req.URL.Path, errorLogKey, err)
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(err.Error())),
		}, nil
	}

	// Process the request using the appropriate handler with transformation support
	switch req.Method {
	case http.MethodGet:
		return h.handleGetWithTransformations(ctx, req, section, sectionName)
	case http.MethodPost:
		return h.handlePostWithTransformations(ctx, req, section, sectionName)
	case http.MethodPut:
		return h.handlePutWithTransformations(ctx, req, section, sectionName)
	case http.MethodDelete:
		return h.handleDeleteWithTransformations(ctx, req, section, sectionName)
	default:
		return &http.Response{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       io.NopCloser(strings.NewReader("method not allowed")),
		}, nil
	}
}

// handleGetWithTransformations handles GET requests with transformation support
func (h *MockHandler) handleGetWithTransformations(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// For GET requests, we need to first retrieve the resource, then apply response transformations
	resp, err := h.handleGet(ctx, req, sectionName)
	if err != nil {
		return nil, err
	}

	// Convert response to MockData for transformation
	mockData, err := h.responsesToMockData(resp)
	if err != nil {
		h.logger.Error("failed to convert response to MockData", errorLogKey, err)
		return h.internalServerError("failed to process response"), nil
	}

	// Apply response transformations
	transformedData, err := h.transformResponse(mockData, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed", errorLogKey, err)
		return h.internalServerError("response transformation failed"), nil
	}

	// Convert back to HTTP response
	return h.mockDataToResponse(transformedData, resp.StatusCode), nil
}

// handlePostWithTransformations handles POST requests with transformation support
func (h *MockHandler) handlePostWithTransformations(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// Extract IDs and prepare initial data
	locationID, initialData, errResp := h.preparePostData(ctx, req, section, sectionName)
	if errResp != nil {
		return errResp, nil
	}

	// Apply transformations and create resource
	transformedData, errResp := h.processPostTransformations(ctx, req, initialData, section, sectionName)
	if errResp != nil {
		return errResp, nil
	}

	// Build and return final response
	return h.buildPostResponse(ctx, req, transformedData, section, sectionName, locationID)
}

// handlePutWithTransformations handles PUT requests with transformation support
func (h *MockHandler) handlePutWithTransformations(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// Prepare PUT data and validate
	initialData, errResp := h.preparePutData(ctx, req, section, sectionName)
	if errResp != nil {
		return errResp, nil
	}

	// Process transformations and update
	transformedData, errResp := h.processPutTransformations(ctx, req, initialData, section, sectionName)
	if errResp != nil {
		return errResp, nil
	}

	// Build and return response
	return h.buildPutResponse(ctx, req, transformedData, section, sectionName)
}

// handleDeleteWithTransformations handles DELETE requests with transformation support
func (h *MockHandler) handleDeleteWithTransformations(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// For DELETE, we just use the original handler since there's typically no response body to transform
	return h.handleDelete(ctx, req, section, sectionName)
}

// Helper methods for transformation integration

// responsesToMockData converts an HTTP response to MockData
func (*MockHandler) responsesToMockData(resp *http.Response) (*model.MockData, error) {
	if resp == nil {
		return nil, errors.New("response is nil")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Close the original body and replace it
	_ = resp.Body.Close() // Ignore error - this is just cleanup
	resp.Body = io.NopCloser(bytes.NewReader(body))

	return &model.MockData{
		ContentType: resp.Header.Get(contentTypeHeader),
		Body:        body,
		Location:    resp.Header.Get("Location"),
	}, nil
}

// mockDataToResponse converts MockData back to an HTTP response
func (*MockHandler) mockDataToResponse(data *model.MockData, statusCode int) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data.Body)),
	}

	if data.ContentType != "" {
		resp.Header.Set(contentTypeHeader, data.ContentType)
	}

	if data.Location != "" {
		resp.Header.Set("Location", data.Location)
	}

	return resp
}

// internalServerError creates a 500 Internal Server Error response
func (*MockHandler) internalServerError(message string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(message)),
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
			_ = resp.Body.Close()
		}()
	}

	h.copyHeaders(w, resp)
	h.writeResponse(w, resp)
}


// copyHeaders copies response headers to the writer
func (*MockHandler) copyHeaders(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
}

// writeResponse writes the response body and status code
func (h *MockHandler) writeResponse(w http.ResponseWriter, resp *http.Response) {
	if resp.Body != nil {
		h.writeBodyResponse(w, resp)
	} else {
		w.WriteHeader(resp.StatusCode)
	}
}

// writeBodyResponse handles writing response with body
func (h *MockHandler) writeBodyResponse(w http.ResponseWriter, resp *http.Response) {
	bodyBytesToRespond, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		h.handleBodyReadError(w, readErr)
		return
	}

	h.setContentLength(w, bodyBytesToRespond)
	w.WriteHeader(resp.StatusCode)

	if len(bodyBytesToRespond) > 0 {
		h.writeBodyData(w, bodyBytesToRespond)
	}
}

// handleBodyReadError handles errors when reading response body
func (h *MockHandler) handleBodyReadError(w http.ResponseWriter, readErr error) {
	h.logger.Error("failed to read response body from HandleRequest's response", errorLogKey, readErr)
	if rr, ok := w.(*httptest.ResponseRecorder); ok && !rr.Flushed {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// setContentLength sets the Content-Length header if not already set
func (*MockHandler) setContentLength(w http.ResponseWriter, body []byte) {
	if w.Header().Get("Content-Length") == "" {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	}
}







// writeBodyData writes the actual body data
func (h *MockHandler) writeBodyData(w http.ResponseWriter, body []byte) {
	_, writeErr := w.Write(body)
	if writeErr != nil {
		h.logger.Error("failed to write response body", errorLogKey, writeErr)
	}
}

// handleGet processes GET requests
func (h *MockHandler) handleGet(ctx context.Context, req *http.Request, sectionName string) (*http.Response, error) {
	pathSegments := getPathInfo(req.URL.Path)
	lastSegment := pathSegments[len(pathSegments)-lastElementIndex]

	if lastSegment != "" && lastSegment != sectionName {
		resource, err := h.service.GetResource(ctx, req.URL.Path, lastSegment)
		if err == nil && resource != nil {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{resource.ContentType}},
				Body:       io.NopCloser(bytes.NewReader(resource.Body)),
			}, nil
		}
	}

	resources, err := h.service.GetResourcesByPath(ctx, req.URL.Path)
	if err == nil && len(resources) > 0 {
		return h.buildCollectionResponse(resources)
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("resource not found")),
	}, nil
}

// buildCollectionResponse builds a JSON array response from resources
func (*MockHandler) buildCollectionResponse(resources []*model.MockData) (*http.Response, error) {
	items := extractJSONItems(resources)
	responseBody := buildJSONArrayBody(items)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}, nil
}

// extractJSONItems filters JSON resources and returns their bodies
func extractJSONItems(resources []*model.MockData) [][]byte {
	var items [][]byte
	for _, r := range resources {
		if strings.Contains(strings.ToLower(r.ContentType), "json") {
			items = append(items, r.Body)
		}
	}
	return items
}

// buildJSONArrayBody constructs a JSON array from items
func buildJSONArrayBody(items [][]byte) []byte {
	if len(items) == 0 {
		return []byte("[]")
	}
	
	if len(items) == singleItem {
		return buildSingleItemArray(items[0])
	}
	
	return buildMultiItemArray(items)
}

// buildSingleItemArray creates JSON array with single item
func buildSingleItemArray(item []byte) []byte {
	responseBody := append([]byte("["), item...)
	return append(responseBody, byte(']'))
}

// buildMultiItemArray creates JSON array with multiple items
func buildMultiItemArray(items [][]byte) []byte {
	responseBody := []byte("[")
	for i, item := range items {
		responseBody = append(responseBody, item...)
		if i < len(items)-singleItem {
			responseBody = append(responseBody, []byte(",")...)
		}
	}
	return append(responseBody, byte(']'))
}


// extractPostResourceIDs extracts or generates resource IDs for POST requests
func (h *MockHandler) extractPostResourceIDs(
	ctx context.Context, 
	req *http.Request, 
	section *config.Section, 
	sectionName string,
) (string, []string, error) {
	idsFromExtractor, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil {
		h.logger.Error("failed to extract IDs for POST", pathLogKey, req.URL.Path, errorLogKey, err)
		return "", nil, err
	}

	if len(idsFromExtractor) == 0 {
		generatedUUID := uuid.New().String()
		h.logger.Debug("POST: no IDs found by extractIDs, generated new UUID", "uuid", generatedUUID)
		return generatedUUID, []string{generatedUUID}, nil
	}

	id := idsFromExtractor[0]
	if len(idsFromExtractor) > singleItem {
		h.logger.Info("multiple IDs found in POST request, using first for Location, all for creation",
			pathLogKey, req.URL.Path, "all_ids", idsFromExtractor, "location_id", id)
	} else {
		h.logger.Debug("POST: using extracted ID(s)", pathLogKey, req.URL.Path, "ids", idsFromExtractor)
	}

	return id, idsFromExtractor, nil
}






// handleDelete processes DELETE requests
func (h *MockHandler) handleDelete(
	ctx context.Context, 
	req *http.Request, 
	section *config.Section, 
	sectionName string,
) (*http.Response, error) {
	ids, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil || len(ids) == 0 {
		h.logger.Error("ID not found in DELETE request", pathLogKey, req.URL.Path, errorLogKey, err)
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(resourceNotFoundMsg)),
		}, nil
	}

	err = h.service.DeleteResource(ctx, req.URL.Path, ids[0])
	if err != nil {
		if err.Error() == "resource not found" {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(resourceNotFoundMsg)),
			}, nil
		}
		return nil, err
	}

	return &http.Response{
		StatusCode: http.StatusNoContent,
	}, nil
}

// GetConfig returns the mock configuration
func (h *MockHandler) GetConfig() *config.MockConfig {
	return h.mockCfg
}
