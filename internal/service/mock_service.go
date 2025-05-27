package service

import (
	"context"
	"fmt"

	"github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

// mockService implements the MockService interface
type mockService struct {
	storage storage.MockStorage
	mockCfg *config.MockConfig
	// logger  *slog.Logger // Logger is currently unused in the service
}

// NewMockService creates a new instance of MockService
func NewMockService(storage storage.MockStorage, cfg *config.MockConfig) MockService {
	return &mockService{
		storage: storage,
		mockCfg: cfg,
		// logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)), // Example initialization if used
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

// UpdateResource updates an existing resource or creates it if it doesn't exist (upsert).
func (s *mockService) UpdateResource(ctx context.Context, path string, id string, data *model.MockData) error {
	err := s.storage.Update(id, data)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			// Resource not found, so attempt to create it.
			// The 'path' parameter isn't directly used by storage.Create,
			// as the mock data itself should contain necessary path info if storage needs it.
			// storage.Create expects a slice of IDs.
			// For a PUT to a specific resource /path/to/resourceId, 'id' is singular.
			createErr := s.storage.Create([]string{id}, data)
			if createErr != nil {
				// Handle potential conflict if, by some race, it was created between Update and Create.
				if _, conflictOk := createErr.(*errors.ConflictError); conflictOk {
					// If it's a conflict, it means it now exists, try Update again just in case.
					// This is a bit defensive but handles a potential race.
					retryErr := s.storage.Update(id, data)
					if retryErr != nil {
						return fmt.Errorf("failed to retry update after create conflict: %w", retryErr)
					}
					return nil // Successfully updated on retry
				}
				return fmt.Errorf("failed to create resource after not found on update: %w", createErr)
			}
			return nil // Successfully created
		}
		if _, ok := err.(*errors.InvalidRequestError); ok {
			return err
		}
		return fmt.Errorf("failed to update resource: %w", err)
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

// Helper functions (These were moved to handler or are no longer needed)
/*
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
*/

// Response creation functions are no longer used in the service layer.
// They were specific to an older handler implementation or example.
/*
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
*/
