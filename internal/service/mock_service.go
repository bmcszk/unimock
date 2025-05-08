package service

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
	"github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/google/uuid"
)

// mockService implements the MockService interface
type mockService struct {
	storage storage.MockStorage
	mockCfg *config.MockConfig
	logger  *slog.Logger
}

// NewMockService creates a new instance of MockService
func NewMockService(storage storage.MockStorage, cfg *config.MockConfig) MockService {
	return &mockService{
		storage: storage,
		mockCfg: cfg,
	}
}

// ExtractIDs extracts IDs from the request using configured paths
func (s *mockService) ExtractIDs(ctx context.Context, req *http.Request) ([]string, error) {
	// Find matching section
	var sectionName string
	var pathPattern string
	var bodyIDPaths []string
	var headerIDName string

	if s.mockCfg != nil {
		// Use the config
		var section *config.Section
		var err error
		sectionName, section, err = s.mockCfg.MatchPath(req.URL.Path)
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
			ids, err = s.extractJSONIDs(body, bodyIDPaths, seenIDs)
			if err != nil {
				return nil, err
			}
		} else if strings.Contains(contentTypeLower, "xml") {
			ids, err = s.extractXMLIDs(body, bodyIDPaths, seenIDs)
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

// HandleRequest processes the HTTP request and returns appropriate response
func (s *mockService) HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Validate content type for POST and PUT requests
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		contentType := req.Header.Get("Content-Type")
		contentTypeLower := strings.ToLower(contentType)
		if !strings.Contains(contentTypeLower, "json") && !strings.Contains(contentTypeLower, "xml") {
			return nil, errors.NewInvalidRequestError(fmt.Sprintf("unsupported content type: %s", contentType))
		}
	}

	// Get section configuration for this path
	var sectionExists bool
	if s.mockCfg != nil {
		_, section, err := s.mockCfg.MatchPath(req.URL.Path)
		if err != nil {
			return nil, errors.NewInvalidRequestError(err.Error())
		}
		if section == nil {
			return nil, errors.NewInvalidRequestError("no matching section found for path: " + req.URL.Path)
		}
		sectionExists = true
	} else {
		return nil, errors.NewInvalidRequestError("service configuration is missing")
	}

	if !sectionExists {
		return nil, errors.NewInvalidRequestError("service configuration is missing")
	}

	// Extract IDs from the request
	ids, err := s.ExtractIDs(ctx, req)
	if err != nil {
		return nil, errors.NewInvalidRequestError(err.Error())
	}

	switch req.Method {
	case http.MethodGet:
		// First try to get by ID if available
		if len(ids) > 0 {
			data, err := s.GetResource(ctx, req.URL.Path, ids[0])
			if err == nil {
				return createResourceResponse(data), nil
			}
			// If specific ID lookup fails, continue to path-based lookup
		}

		// Fallback to path-based retrieval
		data, err := s.GetResourcesByPath(ctx, req.URL.Path)
		if err != nil {
			return nil, err
		}
		return createCollectionResponse(data), nil

	case http.MethodPost:
		// For POST requests, need to extract ID from body or generate one
		if len(ids) == 0 {
			// If no ID in the request, try to extract from the body
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, errors.NewInvalidRequestError(fmt.Sprintf("failed to read request body: %v", err))
			}
			req.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore body for later use

			// Try to extract ID from JSON body
			contentType := req.Header.Get("Content-Type")
			contentTypeLower := strings.ToLower(contentType)
			if strings.Contains(contentTypeLower, "json") {
				var jsonData map[string]interface{}
				if err := json.Unmarshal(body, &jsonData); err == nil {
					if id, ok := jsonData["id"].(string); ok && id != "" {
						ids = []string{id}
					}
				}
			}

			// If still no ID, generate one
			if len(ids) == 0 {
				ids = []string{uuid.New().String()}
			}
		}

		// Check if an ID already exists before creating
		for _, id := range ids {
			if _, err := s.storage.Get(id); err == nil {
				// Resource with this ID already exists
				return nil, fmt.Errorf("resource already exists")
			}
		}

		// Now create the resource
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, errors.NewInvalidRequestError(fmt.Sprintf("failed to read request body: %v", err))
		}

		data := &model.MockData{
			Path:        req.URL.Path,
			ContentType: req.Header.Get("Content-Type"),
			Body:        body,
		}

		if err := s.CreateResource(ctx, req.URL.Path, ids, data); err != nil {
			return nil, err
		}

		return createCreatedResponse(data), nil

	case http.MethodPut:
		if len(ids) == 0 {
			return nil, errors.NewInvalidRequestError("no ID provided for PUT request")
		}

		// Check if resource exists directly with storage
		_, err = s.storage.Get(ids[0])
		if err != nil {
			return nil, fmt.Errorf("resource not found")
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, errors.NewInvalidRequestError(fmt.Sprintf("failed to read request body: %v", err))
		}

		data := &model.MockData{
			Path:        req.URL.Path,
			ContentType: req.Header.Get("Content-Type"),
			Body:        body,
		}

		if err := s.UpdateResource(ctx, req.URL.Path, ids[0], data); err != nil {
			return nil, err
		}

		return createResourceResponse(data), nil

	case http.MethodDelete:
		if len(ids) == 0 {
			return nil, errors.NewInvalidRequestError("no ID provided for DELETE request")
		}

		// First try to delete by ID
		err = s.DeleteResource(ctx, req.URL.Path, ids[0])
		if err == nil {
			return createNoContentResponse(), nil
		}

		// If ID-based deletion fails, try path-based deletion
		if strings.Contains(err.Error(), "resource not found") {
			// Try deleting by path
			err = s.storage.Delete(req.URL.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to delete resource: %v", err)
			}
			return createNoContentResponse(), nil
		}

		return nil, err

	default:
		return nil, fmt.Errorf("method %s not allowed", req.Method)
	}
}

// GetResource retrieves a resource by path and ID
func (s *mockService) GetResource(ctx context.Context, path string, id string) (*model.MockData, error) {
	data, err := s.storage.Get(id)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			return nil, fmt.Errorf("resource not found")
		}
		if _, ok := err.(*errors.InvalidRequestError); ok {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}
	return data, nil
}

// GetResourcesByPath retrieves all resources at a given path
func (s *mockService) GetResourcesByPath(ctx context.Context, path string) ([]*model.MockData, error) {
	data, err := s.storage.GetByPath(path)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			// Return empty array instead of error for collection endpoints
			return []*model.MockData{}, nil
		}
		if _, ok := err.(*errors.InvalidRequestError); ok {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get resources: %v", err)
	}
	return data, nil
}

// CreateResource creates a new resource
func (s *mockService) CreateResource(ctx context.Context, path string, ids []string, data *model.MockData) error {
	if len(ids) == 0 {
		return errors.NewInvalidRequestError("no IDs found in request")
	}
	err := s.storage.Create(ids, data)
	if err != nil {
		if _, ok := err.(*errors.ConflictError); ok {
			return fmt.Errorf("resource already exists")
		}
		if _, ok := err.(*errors.InvalidRequestError); ok {
			return err
		}
		return fmt.Errorf("failed to create resource: %v", err)
	}
	return nil
}

// UpdateResource updates an existing resource
func (s *mockService) UpdateResource(ctx context.Context, path string, id string, data *model.MockData) error {
	err := s.storage.Update(id, data)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			return fmt.Errorf("resource not found")
		}
		if _, ok := err.(*errors.InvalidRequestError); ok {
			return err
		}
		return fmt.Errorf("failed to update resource: %v", err)
	}
	return nil
}

// DeleteResource removes a resource
func (s *mockService) DeleteResource(ctx context.Context, path string, id string) error {
	err := s.storage.Delete(id)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			return fmt.Errorf("resource not found")
		}
		if _, ok := err.(*errors.InvalidRequestError); ok {
			return err
		}
		return fmt.Errorf("failed to delete resource: %v", err)
	}
	return nil
}

// Helper functions

func getPathInfo(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

func isValidID(segment string, sectionName string) bool {
	// Check if it's a valid JSON string and not the section name
	_, err := json.Marshal(segment)
	return err == nil && segment != sectionName
}

func (s *mockService) extractJSONIDs(body []byte, paths []string, seenIDs map[string]bool) ([]string, error) {
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

func (s *mockService) extractXMLIDs(body []byte, paths []string, seenIDs map[string]bool) ([]string, error) {
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

func createResourceResponse(data *model.MockData) *http.Response {
	headers := map[string][]string{
		"Content-Type": {data.ContentType},
	}

	// Add location header if it exists
	if data.Location != "" {
		headers["Location"] = []string{data.Location}
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     headers,
		Body:       io.NopCloser(bytes.NewReader(data.Body)),
	}
}

func createCollectionResponse(data []*model.MockData) *http.Response {
	// Sort the data by path to ensure consistent ordering
	sort.Slice(data, func(i, j int) bool {
		return data[i].Path < data[j].Path
	})

	// Create a JSON array with the bodies from each MockData
	var rawBodies []interface{}

	for _, item := range data {
		// More flexible content type handling - treat any content type containing "json" as JSON
		if strings.Contains(strings.ToLower(item.ContentType), "json") {
			// For JSON content, parse it directly
			var jsonData interface{}
			if err := json.Unmarshal(item.Body, &jsonData); err == nil {
				rawBodies = append(rawBodies, jsonData)
			} else {
				// If JSON parsing fails, use as string
				rawBodies = append(rawBodies, string(item.Body))
			}
		}
		// XML content types will be supported in future updates
	}

	// Marshal the array of bodies
	body, err := json.Marshal(rawBodies)
	if err != nil {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
		}
	}

	// For collection responses, we use a default content type of application/json
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     headers,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func createCreatedResponse(data *model.MockData) *http.Response {
	headers := map[string][]string{
		"Content-Type": {data.ContentType},
	}

	// Add location header if it exists
	if data.Location != "" {
		headers["Location"] = []string{data.Location}
	}

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     headers,
		Body:       io.NopCloser(bytes.NewReader(data.Body)),
	}
}

func createNoContentResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusNoContent,
	}
}
