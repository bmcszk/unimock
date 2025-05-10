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
	if oldData == nil { // Should not happen if storageID exists, but as a safeguard
		return errors.NewNotFoundError(id, "data integrity issue: storageID exists but no data")
	}

	// Determine old and new id-specific paths
	oldIDPath := path.Join(oldData.Path, id)
	newIDPath := path.Join(data.Path, id)

	// Update the data itself
	s.data[storageID] = data

	// Update pathMap if the base path or the full path (if id changed, though id is fixed here) has changed
	// This handles the case where the resource is moved to a different collection path.
	if oldData.Path != data.Path {
		// Remove from old base path in pathMap
		if pathIDs, ok := s.pathMap[oldData.Path]; ok {
			for i, sid := range pathIDs {
				if sid == storageID {
					s.pathMap[oldData.Path] = append(pathIDs[:i], pathIDs[i+1:]...)
					break
				}
			}
			if len(s.pathMap[oldData.Path]) == 0 {
				delete(s.pathMap, oldData.Path)
			}
		}
		// Remove from old id-specific path in pathMap
		if pathIDs, ok := s.pathMap[oldIDPath]; ok {
			for i, sid := range pathIDs {
				if sid == storageID {
					s.pathMap[oldIDPath] = append(pathIDs[:i], pathIDs[i+1:]...)
					break
				}
			}
			if len(s.pathMap[oldIDPath]) == 0 {
				delete(s.pathMap, oldIDPath)
			}
		}

		// Add to new base path in pathMap
		s.pathMap[data.Path] = append(s.pathMap[data.Path], storageID)
		// Add to new id-specific path in pathMap
		s.pathMap[newIDPath] = append(s.pathMap[newIDPath], storageID)
	} else if oldIDPath != newIDPath { // Path is same, but ID somehow changed - defensive
		// This case should not happen if 'id' parameter to Update is the key externalID
		// and it's not changing. If data.Location or similar implies ID change, this needs thought.
		// For now, assume external 'id' is constant for an update call.
	}

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

	// Normalize path by removing trailing slash
	path = strings.TrimSuffix(path, "/")

	// Only allow exact case-sensitive match
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

	// For collections, only allow case-sensitive prefix match
	for storedPath, storageIDs := range s.pathMap {
		storedPath = strings.TrimSuffix(storedPath, "/")
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

	return nil, errors.NewNotFoundError("resource not found", path)
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
