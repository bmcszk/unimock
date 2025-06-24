package handler

import (
	"bytes"
	"context"
	"errors"
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
const (
	// Common strings
	errorLogKey = "error"
	pathLogKey  = "path"
	contentTypeHeader = "Content-Type"
)
// MockHandler provides clear, step-by-step HTTP method handlers
type MockHandler struct {
	service         service.MockService
	scenarioService service.ScenarioService
	logger          *slog.Logger
	mockCfg         *config.MockConfig
}
// NewMockHandler creates a new handler
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
// HandlePOST processes POST requests step by step
func (h *MockHandler) HandlePOST(ctx context.Context, req *http.Request) (*http.Response, error) {
	h.logger.Debug("starting POST request processing", "path", req.URL.Path)

	// Step 1: Find matching configuration section
	section, sectionName, err := h.findSection(req.URL.Path)
	if err != nil {
		h.logger.Warn("no matching section for POST", "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusNotFound, err.Error()), nil
	}

	// Step 2: Prepare POST data with ID extraction
	_, mockData, errResp := h.preparePostData(ctx, req, section, sectionName)
	if errResp != nil {
		return errResp, nil
	}
	// Step 3: Apply transformations and store resource
	transformedData, errResp := h.processPostRequest(ctx, req, mockData, section, sectionName)
	if errResp != nil {
		return errResp, nil
	}
	// Step 4: Build and return response
	return h.buildPOSTResponse(transformedData, section, sectionName)
}
// preparePostData extracts IDs and builds initial MockData for POST
func (h *MockHandler) preparePostData(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) ([]string, *model.MockData, *http.Response) {
	// Extract IDs from request
	ids, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil {
		h.logger.Error("failed to extract IDs for POST", "path", req.URL.Path, "error", err)
		if strings.Contains(err.Error(), "failed to parse JSON body") {
			return nil, nil, h.errorResponse(http.StatusBadRequest, "invalid request: failed to parse JSON body")
		}
		return nil, nil, h.errorResponse(http.StatusBadRequest, "failed to extract IDs")
	}

	// Generate UUID if no IDs found
	if len(ids) == 0 {
		generatedID := uuid.New().String()
		ids = []string{generatedID}
		h.logger.Debug("generated UUID for POST", "uuid", generatedID)
	}

	// Build MockData from request
	mockData, err := h.buildMockDataFromRequest(req, ids)
	if err != nil {
		h.logger.Error("failed to build MockData for POST", "error", err)
		return nil, nil, h.errorResponse(http.StatusBadRequest, "failed to process request data")
	}

	return ids, mockData, nil
}
// processPostRequest applies transformations and stores the resource
func (h *MockHandler) processPostRequest(
	ctx context.Context,
	req *http.Request,
	mockData *model.MockData,
	section *config.Section,
	sectionName string,
) (*model.MockData, *http.Response) {
	// Apply request transformations
	transformedData, err := h.applyRequestTransformations(mockData, section, sectionName)
	if err != nil {
		h.logger.Error("request transformation failed for POST", "error", err)
		return nil, h.errorResponse(http.StatusInternalServerError, "request transformation failed")
	}

	// Store the resource
	err = h.service.CreateResource(ctx, req.URL.Path, transformedData.IDs, transformedData)
	if err != nil {
		h.logger.Error("failed to create resource", "error", err)
		if strings.Contains(err.Error(), "already exists") {
			return nil, h.errorResponse(http.StatusConflict, "resource already exists")
		}
		return nil, h.errorResponse(http.StatusInternalServerError, "failed to create resource")
	}

	return transformedData, nil
}

// HandleGET processes GET requests step by step
func (h *MockHandler) HandleGET(ctx context.Context, req *http.Request) (*http.Response, error) {
	h.logger.Debug("starting GET request processing", "path", req.URL.Path)

	// Step 1: Find matching configuration section
	section, sectionName, err := h.findSection(req.URL.Path)
	if err != nil {
		h.logger.Warn("no matching section for GET", "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusNotFound, err.Error()), nil
	}

	// Step 2: Try to get individual resource first
	individualResp := h.tryGetIndividualResource(ctx, req, section, sectionName)
	if individualResp != nil {
		return individualResp, nil
	}

	// Step 3: Get collection of resources
	return h.getResourceCollection(ctx, req, section, sectionName)
}

// tryGetIndividualResource attempts to get an individual resource
func (h *MockHandler) tryGetIndividualResource(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) *http.Response {
	lastSegment := h.extractLastPathSegment(req.URL.Path)
	if lastSegment == "" || lastSegment == sectionName {
		return nil
	}

	resource, err := h.service.GetResource(ctx, req.URL.Path, lastSegment)
	if err != nil || resource == nil {
		return h.errorResponse(http.StatusNotFound, "resource not found")
	}

	// Apply strict path validation if enabled
	if section.StrictPath {
		if err := h.validateStrictPathAccess(req.URL.Path, resource.Path, section.PathPattern); err != nil {
			h.logger.Debug("strict path validation failed for GET", 
				"requestPath", req.URL.Path, "resourcePath", resource.Path, "error", err)
			return h.errorResponse(http.StatusNotFound, "resource not found")
		}
	}

	return h.buildTransformedResponse(resource, section, sectionName)
}
// getResourceCollection gets a collection of resources
func (h *MockHandler) getResourceCollection(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// For collection requests, use the base path from the pattern
	// e.g., for pattern "/users/*" and request "/users/nonexistent", look for resources at "/users"
	basePath := h.getCollectionBasePath(section.PathPattern, req.URL.Path)
	
	resources, err := h.service.GetResourcesByPath(ctx, basePath)
	if err != nil || len(resources) == 0 {
		return h.errorResponse(http.StatusNotFound, "resource not found"), nil
	}

	transformedResources, err := h.transformResourceCollection(resources, section, sectionName)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, "response transformation failed"), nil
	}

	return h.buildCollectionResponse(transformedResources), nil
}

// getCollectionBasePath determines the base path for collection queries
func (h *MockHandler) getCollectionBasePath(pattern, requestPath string) string {
	if !strings.Contains(pattern, "*") {
		return requestPath
	}
	
	return h.extractBasePathFromWildcard(pattern, requestPath)
}

// extractBasePathFromWildcard extracts base path from wildcard patterns
func (*MockHandler) extractBasePathFromWildcard(pattern, requestPath string) string {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	
	baseParts := make([]string, 0, len(patternParts))
	for i, part := range patternParts {
		if part == "*" || part == "**" {
			break
		}
		if i < len(requestParts) {
			baseParts = append(baseParts, requestParts[i])
		}
	}
	
	if len(baseParts) == 0 {
		return "/"
	}
	return "/" + strings.Join(baseParts, "/")
}

// transformResourceCollection applies transformations to a collection of resources
func (h *MockHandler) transformResourceCollection(
	resources []*model.MockData,
	section *config.Section,
	sectionName string,
) ([]*model.MockData, error) {
	transformedResources := make([]*model.MockData, len(resources))
	for i, resource := range resources {
		transformed, err := h.applyResponseTransformations(resource, section, sectionName)
		if err != nil {
			h.logger.Error("response transformation failed for collection item", "error", err)
			return nil, err
		}
		transformedResources[i] = transformed
	}
	return transformedResources, nil
}

// HandlePUT processes PUT requests step by step
func (h *MockHandler) HandlePUT(ctx context.Context, req *http.Request) (*http.Response, error) {
	h.logger.Debug("starting PUT request processing", "path", req.URL.Path)

	// Step 1: Find matching configuration section
	section, sectionName, err := h.findSection(req.URL.Path)
	if err != nil {
		h.logger.Warn("no matching section for PUT", "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusNotFound, err.Error()), nil
	}

	return h.processPUTRequest(ctx, req, section, sectionName)
}

// processPUTRequest handles the main PUT logic after section validation
func (h *MockHandler) processPUTRequest(
	ctx context.Context, req *http.Request, section *config.Section, sectionName string,
) (*http.Response, error) {
	// Extract ID from path
	ids, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil || len(ids) == 0 {
		h.logger.Error("failed to extract ID for PUT", "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusBadRequest, "ID required for PUT"), nil
	}

	// Build and transform data
	mockData, err := h.buildMockDataFromRequest(req, ids)
	if err != nil {
		h.logger.Error("failed to build MockData for PUT", "error", err)
		return h.errorResponse(http.StatusBadRequest, "failed to process request data"), nil
	}

	transformedData, err := h.applyRequestTransformations(mockData, section, sectionName)
	if err != nil {
		h.logger.Error("request transformation failed for PUT", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "request transformation failed"), nil
	}

	// Validate strict path if enabled
	if section.StrictPath {
		if resp := h.validateStrictPathForOperation(ctx, req.URL.Path, ids[0], 
			section.PathPattern, "PUT"); resp != nil {
			return resp, nil
		}
	}
	
	return h.executeResourceUpdate(ctx, req.URL.Path, ids[0], transformedData, section, sectionName)
}

// validateStrictPathForPUT checks resource existence for strict path validation
func (h *MockHandler) validateStrictPathForPUT(ctx context.Context, path, id string) error {
	existingResource, err := h.service.GetResource(ctx, path, id)
	if err != nil || existingResource == nil {
		h.logger.Debug("resource not found for strict PUT", "path", path, "id", id)
		return errors.New("resource not found")
	}
	return nil
}





// executeResourceUpdate performs the actual resource update and response building
func (h *MockHandler) executeResourceUpdate(
	ctx context.Context, path, id string, data *model.MockData, section *config.Section, sectionName string,
) (*http.Response, error) {
	err := h.service.UpdateResource(ctx, path, id, data)
	if err != nil {
		h.logger.Error("failed to update resource", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "failed to update resource"), nil
	}

	responseData, err := h.applyResponseTransformations(data, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed for PUT", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "response transformation failed"), nil
	}

	return h.buildSingleResourceResponse(responseData), nil
}

// HandleDELETE processes DELETE requests step by step
func (h *MockHandler) HandleDELETE(ctx context.Context, req *http.Request) (*http.Response, error) {
	h.logger.Debug("starting DELETE request processing", "path", req.URL.Path)

	// Step 1: Find matching configuration section
	section, sectionName, err := h.findSection(req.URL.Path)
	if err != nil {
		h.logger.Warn("no matching section for DELETE", "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusNotFound, err.Error()), nil
	}

	return h.processDELETERequest(ctx, req, section, sectionName)
}

// processDELETERequest handles the main DELETE logic after section validation
func (h *MockHandler) processDELETERequest(
	ctx context.Context, req *http.Request, section *config.Section, sectionName string,
) (*http.Response, error) {
	// Extract ID from path
	ids, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil || len(ids) == 0 {
		h.logger.Error("failed to extract ID for DELETE", "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusNotFound, "resource not found"), nil
	}

	// Validate strict path if enabled
	if section.StrictPath {
		if resp := h.validateStrictPathForOperation(ctx, req.URL.Path, ids[0], 
			section.PathPattern, "DELETE"); resp != nil {
			return resp, nil
		}
	}
	
	return h.executeResourceDeletion(ctx, req.URL.Path, ids[0])
}

// validateStrictPathForDELETE checks resource existence for strict path validation
func (h *MockHandler) validateStrictPathForDELETE(ctx context.Context, path, id string) error {
	existingResource, err := h.service.GetResource(ctx, path, id)
	if err != nil || existingResource == nil {
		h.logger.Debug("resource not found for strict DELETE", "path", path, "id", id)
		return errors.New("resource not found")
	}
	return nil
}

// executeResourceDeletion performs the actual resource deletion
func (h *MockHandler) executeResourceDeletion(ctx context.Context, path, id string) (*http.Response, error) {
	err := h.service.DeleteResource(ctx, path, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "resource not found"), nil
		}
		h.logger.Error("failed to delete resource", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "failed to delete resource"), nil
	}

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

// Helper methods for the simplified handlers

// findSection finds the matching configuration section for a request path
func (h *MockHandler) findSection(reqPath string) (*config.Section, string, error) {
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

// buildMockDataFromRequest creates MockData from HTTP request
func (*MockHandler) buildMockDataFromRequest(req *http.Request, ids []string) (*model.MockData, error) {
	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	
	// Restore body for potential re-reading
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	
	// Build MockData
	mockData := &model.MockData{
		Path:        strings.TrimSuffix(req.URL.Path, "/"),
		IDs:         ids,
		ContentType: req.Header.Get("Content-Type"),
		Body:        body,
	}
	
	// Set location for the resource
	if len(ids) > 0 {
		mockData.Location = mockData.Path + "/" + ids[0]
	}
	
	return mockData, nil
}

// applyRequestTransformations applies request transformations if configured
func (h *MockHandler) applyRequestTransformations(
	data *model.MockData,
	section *config.Section,
	sectionName string,
) (*model.MockData, error) {
	if section.Transformations == nil || !section.Transformations.HasRequestTransforms() {
		return data, nil
	}
	
	currentData := data
	for i, transform := range section.Transformations.RequestTransforms {
		h.logger.Debug("applying request transformation", "section", sectionName, "index", i)
		
		transformedData, err := h.safeExecuteTransform(transform, currentData, "request", sectionName)
		if err != nil {
			return nil, fmt.Errorf("request transformation %d failed: %w", i, err)
		}
		
		currentData = transformedData
	}
	
	return currentData, nil
}

// applyResponseTransformations applies response transformations if configured
func (h *MockHandler) applyResponseTransformations(
	data *model.MockData,
	section *config.Section,
	sectionName string,
) (*model.MockData, error) {
	if section.Transformations == nil || !section.Transformations.HasResponseTransforms() {
		return data, nil
	}
	
	currentData := data
	for i, transform := range section.Transformations.ResponseTransforms {
		h.logger.Debug("applying response transformation", "section", sectionName, "index", i)
		
		transformedData, err := h.safeExecuteTransform(transform, currentData, "response", sectionName)
		if err != nil {
			return nil, fmt.Errorf("response transformation %d failed: %w", i, err)
		}
		
		currentData = transformedData
	}
	
	return currentData, nil
}

// safeExecuteTransform executes a transformation with panic recovery
func (h *MockHandler) safeExecuteTransform(
	transform func(*model.MockData) (*model.MockData, error),
	data *model.MockData,
	transformType string,
	sectionName string,
) (transformedData *model.MockData, err error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("transformation panicked", 
				"type", transformType, 
				"section", sectionName, 
				"panic", r)
			err = fmt.Errorf("%s transformation panicked: %v", transformType, r)
			transformedData = nil
		}
	}()
	
	return transform(data)
}

// buildPOSTResponse builds response for POST requests
func (h *MockHandler) buildPOSTResponse(
	data *model.MockData,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// Apply response transformations
	responseData, err := h.applyResponseTransformations(data, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed for POST", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "response transformation failed"), nil
	}
	
	resp := &http.Response{
		StatusCode: http.StatusCreated,
		Header:     make(http.Header),
	}
	
	// Set Location header
	if responseData.Location != "" {
		resp.Header.Set("Location", responseData.Location)
	}
	
	// Only set response body if response transformations are configured
	if section.Transformations != nil && section.Transformations.HasResponseTransforms() {
		resp.Body = io.NopCloser(bytes.NewReader(responseData.Body))
		if responseData.ContentType != "" {
			resp.Header.Set("Content-Type", responseData.ContentType)
		}
	} else {
		resp.Body = io.NopCloser(strings.NewReader(""))
	}
	
	return resp, nil
}

// buildSingleResourceResponse builds response for individual resource
func (*MockHandler) buildSingleResourceResponse(data *model.MockData) *http.Response {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data.Body)),
	}
	
	if data.ContentType != "" {
		resp.Header.Set("Content-Type", data.ContentType)
	}
	
	if data.Location != "" {
		resp.Header.Set("Location", data.Location)
	}
	
	return resp
}

// buildCollectionResponse builds response for collection of resources
func (h *MockHandler) buildCollectionResponse(resources []*model.MockData) *http.Response {
	jsonItems := h.extractJSONItems(resources)
	responseBody := h.buildJSONArrayBody(jsonItems)
	
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}
}

// extractJSONItems filters JSON resources and returns their bodies
func (*MockHandler) extractJSONItems(resources []*model.MockData) [][]byte {
	var jsonItems [][]byte
	for _, resource := range resources {
		if strings.Contains(strings.ToLower(resource.ContentType), "json") {
			jsonItems = append(jsonItems, resource.Body)
		}
	}
	return jsonItems
}

// buildJSONArrayBody constructs a JSON array from items
func (h *MockHandler) buildJSONArrayBody(jsonItems [][]byte) []byte {
	if len(jsonItems) == 0 {
		return []byte("[]")
	}
	if len(jsonItems) == 1 {
		return h.buildSingleItemArray(jsonItems[0])
	}
	return h.buildMultiItemArray(jsonItems)
}

// buildSingleItemArray creates JSON array with single item
func (*MockHandler) buildSingleItemArray(item []byte) []byte {
	responseBody := append([]byte("["), item...)
	return append(responseBody, ']')
}

// buildMultiItemArray creates JSON array with multiple items
func (*MockHandler) buildMultiItemArray(jsonItems [][]byte) []byte {
	responseBody := []byte("[")
	for i, item := range jsonItems {
		responseBody = append(responseBody, item...)
		if i < len(jsonItems)-1 {
			responseBody = append(responseBody, ',')
		}
	}
	return append(responseBody, ']')
}

// errorResponse creates an error HTTP response
func (*MockHandler) errorResponse(statusCode int, message string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(message)),
	}
}

// extractIDs extracts IDs from the request using configured paths
func (h *MockHandler) extractIDs(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) ([]string, error) {
	if h.isPathBasedMethod(req.Method) {
		return h.extractPathBasedIDs(req, section, sectionName)
	}
	return h.extractBodyBasedIDs(ctx, req, section, sectionName)
}

// isPathBasedMethod checks if method uses path-based ID extraction
func (*MockHandler) isPathBasedMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodPut || method == http.MethodDelete
}

// extractPathBasedIDs extracts IDs from path for GET/PUT/DELETE methods
func (h *MockHandler) extractPathBasedIDs(
	req *http.Request,
	section *config.Section,
	sectionName string,
) ([]string, error) {
	pathSegments := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(section.PathPattern, "/"), "/")

	if h.shouldExtractFromPath(pathSegments, patternSegments) {
		lastSegment := pathSegments[len(pathSegments)-1]
		if lastSegment != "" && lastSegment != sectionName {
			return []string{lastSegment}, nil
		}
	}
	return nil, nil
}

// shouldExtractFromPath determines if ID should be extracted from path
func (*MockHandler) shouldExtractFromPath(pathSegments, patternSegments []string) bool {
	if len(patternSegments) == 0 || len(pathSegments) == 0 {
		return false
	}
	
	lastPattern := patternSegments[len(patternSegments)-1]
	
	// Handle recursive wildcard **
	if lastPattern == "**" {
		// For pattern /users/**, any path like /users/123 should extract 123
		return len(pathSegments) > len(patternSegments)-1
	}
	
	// Handle single wildcard *
	if lastPattern == "*" {
		return len(pathSegments) == len(patternSegments)
	}
	
	// For exact patterns, only extract if path is longer
	return len(pathSegments) > len(patternSegments)
}

// extractBodyBasedIDs extracts IDs from headers and body for POST requests
func (h *MockHandler) extractBodyBasedIDs(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) ([]string, error) {
	var collectedIDs []string
	seenIDs := make(map[string]bool)

	addID := func(id string) {
		if id != "" && id != sectionName && !seenIDs[id] {
			collectedIDs = append(collectedIDs, id)
			seenIDs[id] = true
		}
	}

	return h.extractPostIDs(ctx, req, section, addID)
}

// extractPostIDs extracts IDs from headers and body for POST requests
func (h *MockHandler) extractPostIDs(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	addID func(string),
) ([]string, error) {
	var collectedIDs []string

	// Try header ID extraction
	collectedIDs = h.tryExtractHeaderID(section, req, addID, collectedIDs)

	// Try body ID extraction
	bodyIDs, err := h.extractBodyIDs(ctx, req, section)
	if err != nil {
		return nil, err
	}
	for _, id := range bodyIDs {
		addID(id)
	}
	collectedIDs = append(collectedIDs, bodyIDs...)

	// Try path ID extraction as fallback
	if len(collectedIDs) == 0 {
		collectedIDs = h.tryExtractPathIDFallback(req, section, addID, collectedIDs)
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

// tryExtractPathIDFallback attempts to extract ID from path as fallback
func (*MockHandler) tryExtractPathIDFallback(
	req *http.Request,
	section *config.Section,
	addID func(string),
	collectedIDs []string,
) []string {
	pathSegments := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(section.PathPattern, "/"), "/")

	if len(pathSegments) > len(patternSegments) ||
		(len(pathSegments) == len(patternSegments) && strings.HasSuffix(section.PathPattern, "*")) {
		lastSegment := pathSegments[len(pathSegments)-1]
		addID(lastSegment)
		collectedIDs = append(collectedIDs, lastSegment)
	}
	return collectedIDs
}

// extractBodyIDs extracts IDs from request body
func (h *MockHandler) extractBodyIDs(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
) ([]string, error) {
	contentType := strings.ToLower(req.Header.Get("Content-Type"))

	if !h.isSupportedContentType(contentType) {
		return nil, nil
	}

	body, err := h.readAndRestoreRequestBody(req)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, nil
	}

	return h.parseIDsFromBody(ctx, body, contentType, section)
}

// isSupportedContentType checks if content type supports ID extraction
func (*MockHandler) isSupportedContentType(contentType string) bool {
	return strings.Contains(contentType, "json") || strings.Contains(contentType, "xml")
}

// readAndRestoreRequestBody reads and restores request body
func (*MockHandler) readAndRestoreRequestBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// parseIDsFromBody parses IDs from body content based on content type
func (h *MockHandler) parseIDsFromBody(
	ctx context.Context,
	body []byte,
	contentType string,
	section *config.Section,
) ([]string, error) {
	seenIDs := make(map[string]bool)
	var extractedIDs []string
	var err error

	if strings.Contains(contentType, "json") {
		extractedIDs, err = h.extractJSONIDs(body, section.BodyIDPaths, seenIDs)
	} else {
		extractedIDs, err = h.extractXMLIDs(body, section.BodyIDPaths, seenIDs)
	}

	if err != nil {
		h.logger.WarnContext(ctx, "error extracting IDs from body", "error", err)
		return nil, err
	}

	return extractedIDs, nil
}

// extractJSONIDs extracts IDs from JSON body
func (h *MockHandler) extractJSONIDs(body []byte, idPaths []string, seenIDs map[string]bool) ([]string, error) {
	doc, err := jsonquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %w", err)
	}

	var ids []string
	for _, path := range idPaths {
		pathIDs := h.extractJSONIDsFromPath(doc, path, seenIDs)
		ids = append(ids, pathIDs...)
	}
	return ids, nil
}

// extractJSONIDsFromPath extracts IDs from a specific JSON path
func (*MockHandler) extractJSONIDsFromPath(doc *jsonquery.Node, path string, seenIDs map[string]bool) []string {
	nodes, err := jsonquery.QueryAll(doc, path)
	if err != nil {
		return nil
	}

	var ids []string
	for _, node := range nodes {
		if idStr := fmt.Sprintf("%v", node.Value()); idStr != "" && !seenIDs[idStr] {
			ids = append(ids, idStr)
			seenIDs[idStr] = true
		}
	}
	return ids
}

// extractXMLIDs extracts IDs from XML body
func (h *MockHandler) extractXMLIDs(body []byte, idPaths []string, seenIDs map[string]bool) ([]string, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML body: %w", err)
	}

	var ids []string
	for _, path := range idPaths {
		pathIDs := h.extractXMLIDsFromPath(doc, path, seenIDs)
		ids = append(ids, pathIDs...)
	}
	return ids, nil
}

// extractXMLIDsFromPath extracts IDs from a specific XML path
func (*MockHandler) extractXMLIDsFromPath(doc *xmlquery.Node, path string, seenIDs map[string]bool) []string {
	nodes, err := xmlquery.QueryAll(doc, path)
	if err != nil {
		return nil
	}

	var ids []string
	for _, node := range nodes {
		if idStr := node.InnerText(); idStr != "" && !seenIDs[idStr] {
			ids = append(ids, idStr)
			seenIDs[idStr] = true
		}
	}
	return ids
}

// HandleRequest processes the HTTP request and returns appropriate response
func (h *MockHandler) HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")

	// Process the request using the appropriate handler
	switch req.Method {
	case http.MethodGet:
		return h.HandleGET(ctx, req)
	case http.MethodPost:
		return h.HandlePOST(ctx, req)
	case http.MethodPut:
		return h.HandlePUT(ctx, req)
	case http.MethodDelete:
		return h.HandleDELETE(ctx, req)
	default:
		return &http.Response{
			StatusCode: http.StatusMethodNotAllowed,
			Header:     make(http.Header),
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
	if resp.Body == nil {
		w.WriteHeader(resp.StatusCode)
		return
	}
	h.writeResponseBody(w, resp)
}

// writeResponseBody handles writing response with body
func (h *MockHandler) writeResponseBody(w http.ResponseWriter, resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("failed to read response body", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	if len(body) > 0 {
		h.writeBodyContent(w, body)
	}
}

// writeBodyContent writes the actual body content
func (h *MockHandler) writeBodyContent(w http.ResponseWriter, body []byte) {
	_, err := w.Write(body)
	if err != nil {
		h.logger.Error("failed to write response body", "error", err)
	}
}

// GetConfig returns the mock configuration
func (h *MockHandler) GetConfig() *config.MockConfig {
	return h.mockCfg
}