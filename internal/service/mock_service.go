package service

import (
	"context"
	"errors"
	"fmt"

	unimockerrors "github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

// mockService implements the MockService interface
type mockService struct {
	storage storage.MockStorage
	mockCfg *config.MockConfig
}

// NewMockService creates a new instance of MockService
func NewMockService(mockStorage storage.MockStorage, cfg *config.MockConfig) MockService {
	return &mockService{
		storage: mockStorage,
		mockCfg: cfg,
	}
}

// GetResource retrieves a resource by section and ID
func (s *mockService) GetResource(
	_ context.Context, sectionName string, isStrictPath bool, id string,
) (*model.MockData, error) {
	data, err := s.storage.Get(sectionName, isStrictPath, id)
	if err != nil {
		if _, ok := err.(*unimockerrors.NotFoundError); ok {
			return nil, errors.New("resource not found")
		}
		if _, ok := err.(*unimockerrors.InvalidRequestError); ok {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}
	return data, nil
}

// GetResourcesByPath retrieves all resources at a given path
func (s *mockService) GetResourcesByPath(_ context.Context, path string) ([]*model.MockData, error) {
	data, err := s.storage.GetByPath(path)
	if err != nil {
		if _, ok := err.(*unimockerrors.NotFoundError); ok {
			// Return empty array instead of error for collection endpoints
			return []*model.MockData{}, nil
		}
		if _, ok := err.(*unimockerrors.InvalidRequestError); ok {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get resources: %v", err)
	}
	return data, nil
}

// CreateResource creates a new resource
func (s *mockService) CreateResource(
	_ context.Context, sectionName string, isStrictPath bool, ids []string, data *model.MockData,
) error {
	if len(ids) == 0 {
		return unimockerrors.NewInvalidRequestError("no IDs found in request")
	}
	// Ensure MockData has the IDs set
	data.IDs = ids
	err := s.storage.Create(sectionName, isStrictPath, data)
	if err != nil {
		if _, ok := err.(*unimockerrors.ConflictError); ok {
			return errors.New("resource already exists")
		}
		if _, ok := err.(*unimockerrors.InvalidRequestError); ok {
			return err
		}
		return fmt.Errorf("failed to create resource: %v", err)
	}
	return nil
}

// UpdateResource updates an existing resource or creates it if it doesn't exist (upsert).
func (s *mockService) UpdateResource(
	_ context.Context, sectionName string, isStrictPath bool, id string, data *model.MockData,
) error {
	err := s.storage.Update(sectionName, isStrictPath, id, data)
	if err != nil {
		return s.handleUpdateError(err, sectionName, isStrictPath, id, data)
	}
	return nil
}

// handleUpdateError handles various update errors including upsert logic
func (s *mockService) handleUpdateError(
	err error, sectionName string, isStrictPath bool, id string, data *model.MockData,
) error {
	if _, ok := err.(*unimockerrors.NotFoundError); ok {
		return s.handleNotFoundUpdate(sectionName, isStrictPath, id, data)
	}
	if _, ok := err.(*unimockerrors.InvalidRequestError); ok {
		return err
	}
	return fmt.Errorf("failed to update resource: %w", err)
}

// handleNotFoundUpdate handles update when resource is not found (upsert create)
func (s *mockService) handleNotFoundUpdate(
	sectionName string, isStrictPath bool, id string, data *model.MockData,
) error {
	// Ensure MockData has the ID set for upsert create
	data.IDs = []string{id}
	createErr := s.storage.Create(sectionName, isStrictPath, data)
	if createErr != nil {
		return s.handleCreateConflict(createErr, sectionName, isStrictPath, id, data)
	}
	return nil
}

// handleCreateConflict handles potential conflicts during upsert create
func (s *mockService) handleCreateConflict(
	createErr error, sectionName string, isStrictPath bool, id string, data *model.MockData,
) error {
	if _, conflictOk := createErr.(*unimockerrors.ConflictError); !conflictOk {
		return fmt.Errorf("failed to create resource after not found on update: %w", createErr)
	}

	retryErr := s.storage.Update(sectionName, isStrictPath, id, data)
	if retryErr != nil {
		return fmt.Errorf("failed to retry update after create conflict: %w", retryErr)
	}
	return nil
}

// DeleteResource removes a resource
func (s *mockService) DeleteResource(_ context.Context, sectionName string, isStrictPath bool, id string) error {
	err := s.storage.Delete(sectionName, isStrictPath, id)
	if err != nil {
		if _, ok := err.(*unimockerrors.NotFoundError); ok {
			return errors.New("resource not found")
		}
		if _, ok := err.(*unimockerrors.InvalidRequestError); ok {
			return err
		}
		return fmt.Errorf("failed to delete resource: %v", err)
	}
	return nil
}

