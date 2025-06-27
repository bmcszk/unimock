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

// UniHandler provides clear, step-by-step HTTP method handlers
type UniHandler struct {
	service         *service.UniService
	scenarioService *service.ScenarioService
	logger          *slog.Logger
	uniCfg         *config.UniConfig
}
// NewUniHandler creates a new handler
func NewUniHandler(
	uniService *service.UniService,
	scenarioService *service.ScenarioService,
	logger *slog.Logger,
	cfg *config.UniConfig,
) *UniHandler {
	return &UniHandler{
		service:         uniService,
		scenarioService: scenarioService,
		logger:          logger,
		uniCfg:         cfg,
	}
}



// HandlePOST processes POST requests step by step
func (h *UniHandler) HandlePOST(ctx context.Context, req *http.Request) (*http.Response, error) {
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
// preparePostData extracts IDs and builds initial UniData for POST
func (h *UniHandler) preparePostData(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) ([]string, model.UniData, *http.Response) {
	// Extract IDs from request
	ids, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil {
		h.logger.Error("failed to extract IDs for POST", "path", req.URL.Path, "error", err)
		if strings.Contains(err.Error(), "failed to parse JSON body") {
			return nil, model.UniData{}, h.errorResponse(http.StatusBadRequest, 
				"invalid request: failed to parse JSON body")
		}
		return nil, model.UniData{}, h.errorResponse(http.StatusBadRequest, "failed to extract IDs")
	}

	// Generate UUID if no IDs found
	if len(ids) == 0 {
		generatedID := uuid.New().String()
		ids = []string{generatedID}
		h.logger.Debug("generated UUID for POST", "uuid", generatedID)
	}

	// Build UniData from request
	mockData, err := h.buildUniDataFromRequest(req, ids)
	if err != nil {
		h.logger.Error("failed to build UniData for POST", "error", err)
		return nil, model.UniData{}, h.errorResponse(http.StatusBadRequest, "failed to process request data")
	}

	return ids, mockData, nil
}
// processPostRequest applies transformations and stores the resource
func (h *UniHandler) processPostRequest(
	ctx context.Context,
	_ *http.Request,
	mockData model.UniData,
	section *config.Section,
	sectionName string,
) (model.UniData, *http.Response) {
	// Apply request transformations
	transformedData, err := h.applyRequestTransformations(mockData, section, sectionName)
	if err != nil {
		h.logger.Error("request transformation failed for POST", "error", err)
		return model.UniData{}, h.errorResponse(http.StatusInternalServerError, "request transformation failed")
	}

	// Store the resource
	err = h.service.CreateResource(ctx, sectionName, section.StrictPath, transformedData.IDs, transformedData)
	if err != nil {
		h.logger.Error("failed to create resource", "error", err)
		if strings.Contains(err.Error(), "already exists") {
			return model.UniData{}, h.errorResponse(http.StatusConflict, "resource already exists")
		}
		return model.UniData{}, h.errorResponse(http.StatusInternalServerError, "failed to create resource")
	}

	return transformedData, nil
}

// HandleGET processes GET requests step by step
func (h *UniHandler) HandleGET(ctx context.Context, req *http.Request) (*http.Response, error) {
	return h.handleGetOrHead(ctx, req, false)
}

// HandleHEAD processes HEAD requests (same as GET but no body)
func (h *UniHandler) HandleHEAD(ctx context.Context, req *http.Request) (*http.Response, error) {
	return h.handleGetOrHead(ctx, req, true)
}

// handleGetOrHead processes GET and HEAD requests with optional body suppression
//nolint:revive // suppressBody flag is appropriate for differentiating GET vs HEAD behavior
func (h *UniHandler) handleGetOrHead(
	ctx context.Context, req *http.Request, suppressBody bool,
) (*http.Response, error) {
	methodName := "GET"
	if suppressBody {
		methodName = "HEAD"
	}
	h.logger.Debug("starting "+methodName+" request processing", "path", req.URL.Path)

	// Step 1: Find matching configuration section
	section, sectionName, err := h.findSection(req.URL.Path)
	if err != nil {
		h.logger.Warn("no matching section for "+methodName, "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusNotFound, err.Error()), nil
	}

	// Step 2: Try to get individual resource first
	individualResp := h.tryGetIndividualResource(ctx, req, section, sectionName)
	if individualResp != nil {
		if suppressBody {
			return h.suppressResponseBody(individualResp), nil
		}
		return individualResp, nil
	}

	// Step 3: Get collection of resources
	resp := h.getResourceCollection(ctx, req, section, sectionName)
	
	if suppressBody {
		return h.suppressResponseBody(resp), nil
	}
	return resp, nil
}

// tryGetIndividualResource attempts to get an individual resource
func (h *UniHandler) tryGetIndividualResource(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) *http.Response {
	lastSegment := h.extractLastPathSegment(req.URL.Path)
	if lastSegment == "" || lastSegment == sectionName {
		return nil
	}

	resource, err := h.service.GetResource(ctx, sectionName, section.StrictPath, lastSegment)
	if err != nil {
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
func (h *UniHandler) getResourceCollection(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) *http.Response {
	// For collection requests, use the base path from the pattern
	// e.g., for pattern "/users/*" and request "/users/nonexistent", look for resources at "/users"
	basePath := h.getCollectionBasePath(section.PathPattern, req.URL.Path)
	
	resources, err := h.service.GetResourcesByPath(ctx, basePath)
	if err != nil || len(resources) == 0 {
		return h.errorResponse(http.StatusNotFound, "resource not found")
	}

	transformedResources, err := h.transformResourceCollection(resources, section, sectionName)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, "response transformation failed")
	}

	return h.buildCollectionResponse(transformedResources)
}

// getCollectionBasePath determines the base path for collection queries
func (h *UniHandler) getCollectionBasePath(pattern, requestPath string) string {
	if !strings.Contains(pattern, "*") {
		return requestPath
	}
	
	return h.extractBasePathFromWildcard(pattern, requestPath)
}

// extractBasePathFromWildcard extracts base path from wildcard patterns
func (*UniHandler) extractBasePathFromWildcard(pattern, requestPath string) string {
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
func (h *UniHandler) transformResourceCollection(
	resources []model.UniData,
	section *config.Section,
	sectionName string,
) ([]model.UniData, error) {
	transformedResources := make([]model.UniData, len(resources))
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
func (h *UniHandler) HandlePUT(ctx context.Context, req *http.Request) (*http.Response, error) {
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
func (h *UniHandler) processPUTRequest(
	ctx context.Context, req *http.Request, section *config.Section, sectionName string,
) (*http.Response, error) {
	// Extract ID from path
	ids, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil || len(ids) == 0 {
		h.logger.Error("failed to extract ID for PUT", "path", req.URL.Path, "error", err)
		return h.errorResponse(http.StatusBadRequest, "ID required for PUT"), nil
	}

	// Build and transform data
	mockData, err := h.buildUniDataFromRequest(req, ids)
	if err != nil {
		h.logger.Error("failed to build UniData for PUT", "error", err)
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
			section.PathPattern, "PUT", sectionName, section.StrictPath); resp != nil {
			return resp, nil
		}
	}
	
	return h.executeResourceUpdate(ctx, ids[0], transformedData, section, sectionName)
}

// validateResourceExists checks if a resource exists for operations requiring existing resources
func (h *UniHandler) validateResourceExists(
	ctx context.Context, sectionName string, isStrictPath bool, id string, operation string,
) error {
	_, err := h.service.GetResource(ctx, sectionName, isStrictPath, id)
	if err != nil {
		h.logger.Debug("resource not found for strict operation", 
			"sectionName", sectionName, "id", id, "operation", operation)
		return errors.New("resource not found")
	}
	return nil
}





// executeResourceUpdate performs the actual resource update and response building
func (h *UniHandler) executeResourceUpdate(
	ctx context.Context, id string, data model.UniData, section *config.Section, sectionName string,
) (*http.Response, error) {
	err := h.service.UpdateResource(ctx, sectionName, section.StrictPath, id, data)
	if err != nil {
		h.logger.Error("failed to update resource", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "failed to update resource"), nil
	}

	responseData, err := h.applyResponseTransformations(data, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed for PUT", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "response transformation failed"), nil
	}

	return h.buildPUTResponse(responseData, section), nil
}

// HandleDELETE processes DELETE requests step by step
func (h *UniHandler) HandleDELETE(ctx context.Context, req *http.Request) (*http.Response, error) {
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
func (h *UniHandler) processDELETERequest(
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
			section.PathPattern, "DELETE", sectionName, section.StrictPath); resp != nil {
			return resp, nil
		}
	}
	
	return h.executeResourceDeletion(ctx, ids[0], section, sectionName)
}


// executeResourceDeletion performs the actual resource deletion
func (h *UniHandler) executeResourceDeletion(
	ctx context.Context, id string, section *config.Section, sectionName string,
) (*http.Response, error) {
	err := h.service.DeleteResource(ctx, sectionName, section.StrictPath, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "resource not found"), nil
		}
		h.logger.Error("failed to delete resource", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "failed to delete resource"), nil
	}

	return h.buildDELETEResponse(section), nil
}

// Helper methods for the simplified handlers

// findSection finds the matching configuration section for a request path
func (h *UniHandler) findSection(reqPath string) (*config.Section, string, error) {
	if h.uniCfg == nil {
		return nil, "", errors.New("service configuration is missing")
	}
	
	sectionName, section, err := h.uniCfg.MatchPath(reqPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to match path pattern: %w", err)
	}
	
	if section == nil {
		return nil, "", fmt.Errorf("no matching section found for path: %s", reqPath)
	}
	
	return section, sectionName, nil
}

// buildUniDataFromRequest creates UniData from HTTP request
func (*UniHandler) buildUniDataFromRequest(req *http.Request, ids []string) (model.UniData, error) {
	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return model.UniData{}, fmt.Errorf("failed to read request body: %w", err)
	}
	
	// Restore body for potential re-reading
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	
	// Build UniData
	mockData := model.UniData{
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
func (h *UniHandler) applyRequestTransformations(
	data model.UniData,
	section *config.Section,
	sectionName string,
) (model.UniData, error) {
	if section.Transformations == nil || !section.Transformations.HasRequestTransforms() {
		return data, nil
	}
	
	currentData := data
	for i, transform := range section.Transformations.RequestTransforms {
		h.logger.Debug("applying request transformation", "section", sectionName, "index", i)
		
		transformedData, err := transform(currentData)
		if err != nil {
			return model.UniData{}, fmt.Errorf("request transformation %d failed: %w", i, err)
		}
		
		currentData = transformedData
	}
	
	return currentData, nil
}

// applyResponseTransformations applies response transformations if configured
func (h *UniHandler) applyResponseTransformations(
	data model.UniData,
	section *config.Section,
	sectionName string,
) (model.UniData, error) {
	if section.Transformations == nil || !section.Transformations.HasResponseTransforms() {
		return data, nil
	}
	
	currentData := data
	for i, transform := range section.Transformations.ResponseTransforms {
		h.logger.Debug("applying response transformation", "section", sectionName, "index", i)
		
		transformedData, err := transform(currentData)
		if err != nil {
			return model.UniData{}, fmt.Errorf("response transformation %d failed: %w", i, err)
		}
		
		currentData = transformedData
	}
	
	return currentData, nil
}


// extractIDs extracts IDs from the request using configured paths
func (h *UniHandler) extractIDs(
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
func (*UniHandler) isPathBasedMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead ||
		method == http.MethodPut || method == http.MethodDelete
}

// extractPathBasedIDs extracts IDs from path for GET/HEAD/PUT/DELETE methods
func (h *UniHandler) extractPathBasedIDs(
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
func (*UniHandler) shouldExtractFromPath(pathSegments, patternSegments []string) bool {
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
func (h *UniHandler) extractBodyBasedIDs(
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
func (h *UniHandler) extractPostIDs(
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
func (*UniHandler) tryExtractHeaderID(
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
func (*UniHandler) tryExtractPathIDFallback(
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
func (h *UniHandler) extractBodyIDs(
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
func (*UniHandler) isSupportedContentType(contentType string) bool {
	return strings.Contains(contentType, "json") || strings.Contains(contentType, "xml")
}

// readAndRestoreRequestBody reads and restores request body
func (*UniHandler) readAndRestoreRequestBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// parseIDsFromBody parses IDs from body content based on content type
func (h *UniHandler) parseIDsFromBody(
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
func (h *UniHandler) extractJSONIDs(body []byte, idPaths []string, seenIDs map[string]bool) ([]string, error) {
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
func (*UniHandler) extractJSONIDsFromPath(doc *jsonquery.Node, path string, seenIDs map[string]bool) []string {
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
func (h *UniHandler) extractXMLIDs(body []byte, idPaths []string, seenIDs map[string]bool) ([]string, error) {
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
func (*UniHandler) extractXMLIDsFromPath(doc *xmlquery.Node, path string, seenIDs map[string]bool) []string {
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
func (h *UniHandler) HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")

	// Process the request using the appropriate handler
	var resp *http.Response
	var err error
	
	switch req.Method {
	case http.MethodGet:
		resp, err = h.HandleGET(ctx, req)
	case http.MethodHead:
		resp, err = h.HandleHEAD(ctx, req)
	case http.MethodPost:
		resp, err = h.HandlePOST(ctx, req)
	case http.MethodPut:
		resp, err = h.HandlePUT(ctx, req)
	case http.MethodDelete:
		resp, err = h.HandleDELETE(ctx, req)
	default:
		resp = &http.Response{
			StatusCode: http.StatusMethodNotAllowed,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("method not allowed")),
		}
		err = nil
	}
	
	
	return resp, err
}

// ServeHTTP implements the http.Handler interface
func (h *UniHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
func (*UniHandler) copyHeaders(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
}

// writeResponse writes the response body and status code
func (h *UniHandler) writeResponse(w http.ResponseWriter, resp *http.Response) {
	if resp.Body == nil {
		w.WriteHeader(resp.StatusCode)
		return
	}
	h.writeResponseBody(w, resp)
}

// writeResponseBody handles writing response with body
func (h *UniHandler) writeResponseBody(w http.ResponseWriter, resp *http.Response) {
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
func (h *UniHandler) writeBodyContent(w http.ResponseWriter, body []byte) {
	_, err := w.Write(body)
	if err != nil {
		h.logger.Error("failed to write response body", "error", err)
	}
}

// suppressResponseBody removes body from response while preserving headers and status
func (*UniHandler) suppressResponseBody(resp *http.Response) *http.Response {
	if resp == nil {
		return resp
	}
	
	// Create new response with same headers and status but empty body
	newResp := &http.Response{
		StatusCode: resp.StatusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
		Proto:      resp.Proto,
		ProtoMajor: resp.ProtoMajor,
		ProtoMinor: resp.ProtoMinor,
	}
	
	// Copy all headers
	for k, v := range resp.Header {
		newResp.Header[k] = v
	}
	
	// Close original body if it exists
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
	
	return newResp
}

// GetConfig returns the mock configuration
func (h *UniHandler) GetConfig() *config.UniConfig {
	return h.uniCfg
}