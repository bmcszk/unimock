package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

// validateStrictPathAccess checks if request path pattern matches resource's original path pattern
// when strict_path=true to prevent cross-path access
func (h *MockHandler) validateStrictPathAccess(requestPath, resourcePath, sectionPattern string) error {
	// For strict path validation, we need to check if the request is trying to access
	// a resource via a different path structure than where it was created
	
	// Key insight from PR comment:
	// - POST /users/subpath body: {"id": 1} creates resource at path="/users/subpath"
	// - GET /users/1 should fail (different structure: /users/* vs /users/subpath)  
	// - GET /users/subpath/1 should work (same structure: extends /users/subpath)
	
	// Check if request path structure is compatible with resource creation path
	isCompatible := h.arePathsCompatible(requestPath, resourcePath, sectionPattern)
	
	h.logger.Debug("strict path validation", 
		"requestPath", requestPath, 
		"resourcePath", resourcePath, 
		"sectionPattern", sectionPattern,
		"compatible", isCompatible)
	
	if !isCompatible {
		return fmt.Errorf("path structure mismatch: request path %s not compatible with resource path %s", 
			requestPath, resourcePath)
	}
	
	return nil
}

// arePathsCompatible checks if two paths are compatible for strict path validation
func (*MockHandler) arePathsCompatible(requestPath, resourcePath, _ string) bool {
	// Case 1: Exact match (e.g., GET /users/123 where resource was created at /users via POST /users)
	if requestPath == resourcePath {
		return true
	}
	
	// Case 2: Request extends resource path with ID 
	// (e.g., GET /users/subpath/123 where resource was created at /users/subpath)
	if strings.HasPrefix(requestPath, resourcePath+"/") {
		return true
	}
	
	// Case 3: Different structures should be rejected
	// e.g., GET /users/123 vs resource created at /users/subpath
	return false
}

// validateStrictPathForOperation validates strict path access for PUT/DELETE operations
func (h *MockHandler) validateStrictPathForOperation(ctx context.Context, reqPath, id, 
	sectionPattern, operation, sectionName string, isStrictPath bool) *http.Response {
	// Check resource existence first
	var err error
	if operation == "PUT" {
		err = h.validateStrictPathForPUT(ctx, sectionName, isStrictPath, id)
	} else {
		err = h.validateStrictPathForDELETE(ctx, sectionName, isStrictPath, id)
	}
	
	if err != nil {
		return h.errorResponse(http.StatusNotFound, "resource not found")
	}
	
	// Additional strict path access validation
	existingResource, err := h.service.GetResource(ctx, sectionName, isStrictPath, id)
	if err == nil && existingResource != nil {
		if err := h.validateStrictPathAccess(reqPath, existingResource.Path, sectionPattern); err != nil {
			h.logger.Debug("strict path access validation failed for "+operation, 
				"requestPath", reqPath, "resourcePath", existingResource.Path, "error", err)
			return h.errorResponse(http.StatusNotFound, "resource not found")
		}
	}
	
	return nil
}

// extractLastPathSegment extracts the last segment from a URL path
func (*MockHandler) extractLastPathSegment(urlPath string) string {
	pathSegments := strings.Split(strings.Trim(urlPath, "/"), "/")
	if len(pathSegments) == 0 {
		return ""
	}
	return pathSegments[len(pathSegments)-1]
}

// buildTransformedResponse applies transformations and builds the response
func (h *MockHandler) buildTransformedResponse(resource *model.MockData, 
	section *config.Section, sectionName string) *http.Response {
	transformedData, err := h.applyResponseTransformations(resource, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed for GET", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "response transformation failed")
	}
	return h.buildSingleResourceResponse(transformedData)
}