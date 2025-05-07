package storage

import (
	"path"
	"strings"
	"sync"

	"github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/pkg/model"

	"github.com/google/uuid"
)

// MockStorage interface defines the operations for storing and retrieving data
type MockStorage interface {
	Create(ids []string, data *model.MockData) error
	Update(id string, data *model.MockData) error
	Get(id string) (*model.MockData, error)
	GetByPath(path string) ([]*model.MockData, error)
	Delete(id string) error
	ForEach(fn func(id string, data *model.MockData) error) error
}

// mockStorage implements the Storage interface
type mockStorage struct {
	mu      *sync.RWMutex
	data    map[string]*model.MockData // storageID -> data
	idMap   map[string]string          // externalID -> storageID
	pathMap map[string][]string        // path -> []storageID
}

// NewMockStorage creates a new instance of storage
func NewMockStorage() MockStorage {
	return &mockStorage{
		mu:      &sync.RWMutex{},
		data:    make(map[string]*model.MockData),
		idMap:   make(map[string]string),
		pathMap: make(map[string][]string),
	}
}

// validateData checks if the data is valid
func (s *mockStorage) validateData(data *model.MockData) error {
	if data == nil {
		return errors.NewInvalidRequestError("data cannot be nil")
	}
	return nil
}

// validateID checks if the ID is valid
func (s *mockStorage) validateID(id string) error {
	if id == "" {
		return errors.NewInvalidRequestError("ID cannot be empty")
	}
	return nil
}

// Create stores new data with the given IDs
func (s *mockStorage) Create(ids []string, data *model.MockData) error {
	if err := s.validateData(data); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for conflicts
	for _, id := range ids {
		if _, exists := s.idMap[id]; exists {
			return errors.NewConflictError(id)
		}
	}

	// Generate a new storage ID
	storageID := uuid.New().String()

	// Ensure path doesn't have trailing slash
	data.Path = strings.TrimRight(data.Path, "/")

	// Set location based on path and first ID
	if len(ids) > 0 {
		data.Location = data.Path + "/" + ids[0]
	} else {
		// Generate UUID for path-based storage
		generatedID := uuid.New().String()
		data.Location = data.Path + "/" + generatedID
		ids = []string{generatedID}
	}

	// Store the data
	s.data[storageID] = data

	// Map external IDs to storage ID
	for _, id := range ids {
		s.idMap[id] = storageID
	}

	// Add to pathMap for both original path and path with ID
	s.pathMap[data.Path] = append(s.pathMap[data.Path], storageID)
	if len(ids) > 0 {
		idPath := path.Join(data.Path, ids[0])
		s.pathMap[idPath] = append(s.pathMap[idPath], storageID)
	}

	return nil
}

// Update updates existing data for the given ID
func (s *mockStorage) Update(id string, data *model.MockData) error {
	if err := s.validateData(data); err != nil {
		return err
	}
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get existing storage ID
	storageID, exists := s.idMap[id]
	if !exists {
		return errors.NewNotFoundError(id, "")
	}

	// Get old data for path cleanup
	oldData := s.data[storageID]

	// Update the data
	s.data[storageID] = data

	// Update pathMap
	if oldData != nil && oldData.Path != data.Path {
		// Remove from old path
		if pathIDs, ok := s.pathMap[oldData.Path]; ok {
			for i, sid := range pathIDs {
				if sid == storageID {
					s.pathMap[oldData.Path] = append(pathIDs[:i], pathIDs[i+1:]...)
					break
				}
			}
		}
	}

	// Add to new path
	s.pathMap[data.Path] = append(s.pathMap[data.Path], storageID)

	return nil
}

// Get retrieves data by ID
func (s *mockStorage) Get(id string) (*model.MockData, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	storageID, exists := s.idMap[id]
	if !exists {
		return nil, errors.NewNotFoundError(id, "")
	}

	data, exists := s.data[storageID]
	if !exists {
		return nil, errors.NewNotFoundError(id, "")
	}

	return data, nil
}

// GetByPath retrieves all data stored at the given path
func (s *mockStorage) GetByPath(path string) ([]*model.MockData, error) {
	if path == "" {
		return nil, errors.NewInvalidRequestError("path cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.MockData
	seen := make(map[string]bool) // Track seen storage IDs to prevent duplicates

	// First try exact path match
	if storageIDs, ok := s.pathMap[path]; ok {
		for _, sid := range storageIDs {
			if data, exists := s.data[sid]; exists && !seen[sid] {
				seen[sid] = true
				result = append(result, data)
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}

	// If no exact match, try prefix match for collection paths
	for storedPath, storageIDs := range s.pathMap {
		if strings.HasPrefix(storedPath, path+"/") {
			for _, sid := range storageIDs {
				if data, exists := s.data[sid]; exists && !seen[sid] {
					seen[sid] = true
					result = append(result, data)
				}
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return nil, errors.NewNotFoundError("", path)
}

// Delete removes data by ID or path
func (s *mockStorage) Delete(id string) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Try to delete by ID first
	storageID, exists := s.idMap[id]
	if exists {
		data := s.data[storageID]
		if data != nil {
			// Remove from original path
			if pathIDs, ok := s.pathMap[data.Path]; ok {
				for i, sid := range pathIDs {
					if sid == storageID {
						s.pathMap[data.Path] = append(pathIDs[:i], pathIDs[i+1:]...)
						break
					}
				}
			}
			// Remove from ID path
			idPath := path.Join(data.Path, id)
			if pathIDs, ok := s.pathMap[idPath]; ok {
				for i, sid := range pathIDs {
					if sid == storageID {
						s.pathMap[idPath] = append(pathIDs[:i], pathIDs[i+1:]...)
						break
					}
				}
			}
		}
		delete(s.data, storageID)
		delete(s.idMap, id)
		return nil
	}

	// If ID-based deletion failed, try path-based deletion
	var found bool
	for storedPath, storageIDs := range s.pathMap {
		if strings.HasPrefix(storedPath, id+"/") || storedPath == id {
			for _, sid := range storageIDs {
				if _, exists := s.data[sid]; exists {
					// Find and remove all external IDs that map to this storage ID
					for extID, storedSID := range s.idMap {
						if storedSID == sid {
							delete(s.idMap, extID)
						}
					}
					delete(s.data, sid)
					found = true
				}
			}
			delete(s.pathMap, storedPath)
		}
	}

	if !found {
		return errors.NewNotFoundError(id, id)
	}

	return nil
}

// ForEach iterates over all stored data
func (s *mockStorage) ForEach(fn func(id string, data *model.MockData) error) error {
	if fn == nil {
		return errors.NewInvalidRequestError("callback function cannot be nil")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, data := range s.data {
		if err := fn(id, data); err != nil {
			return errors.NewStorageError("forEach", err)
		}
	}

	return nil
}
