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

// Delete removes data by ID
func (s *mockStorage) Delete(id string) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	storageID, exists := s.idMap[id]
	if !exists {
		// If the direct ID is not found, it might be a path trying to be deleted.
		// However, this Delete function is ID-specific.
		// Path-based deletion should be handled by a different mechanism or by iterating through GetByPath results.
		return errors.NewNotFoundError(id, " (as an individual resource ID)")
	}

	// Get the data for path cleanup and to ensure it exists
	mockData, dataExists := s.data[storageID]
	if !dataExists {
		// Data integrity issue: idMap has it, but data doesn't. Clean up idMap entry.
		delete(s.idMap, id)
		return errors.NewNotFoundError(id, " (data integrity issue: found in idMap but not in data store)")
	}

	// 1. Remove all external ID mappings for this storageID
	for extID, mappedStorageID := range s.idMap {
		if mappedStorageID == storageID {
			delete(s.idMap, extID)
		}
	}

	// 2. Remove the resource data
	delete(s.data, storageID)

	// 3. Clean up pathMap
	// Remove from the resource's base path
	if pathIDs, ok := s.pathMap[mockData.Path]; ok {
		newPathIDs := []string{}
		for _, sid := range pathIDs {
			if sid != storageID {
				newPathIDs = append(newPathIDs, sid)
			}
		}
		if len(newPathIDs) == 0 {
			delete(s.pathMap, mockData.Path)
		} else {
			s.pathMap[mockData.Path] = newPathIDs
		}
	}

	// Remove from all potential ID-specific paths.
	// This requires knowing all external IDs previously associated with mockData.Path.
	// Since we just deleted them from idMap, we don't have them directly.
	// However, the original pathMap construction added entries for path.Join(data.Path, ids[0]).
	// This part is tricky if multiple external IDs could form different id-specific paths.
	// For now, assume the main id-specific path was formed with the 'id' passed to this Delete function.
	// A more robust cleanup might involve iterating all paths in pathMap, but that's inefficient.
	// The current pathMap structure might need rethinking if arbitrary external IDs can form parts of distinct stored paths.
	// Let's stick to cleaning the path derived from the original id.
	// The original code also cleaned up path.Join(oldData.Path, id) in Update.

	// Re-evaluate path cleanup based on how paths are stored.
	// The current `pathMap` stores storageIDs under their `data.Path` (collection)
	// and potentially under `data.Path + "/" + externalId` (specific resource via one ID).
	// We've cleaned `mockData.Path`. We also need to clean any `mockData.Path + "/" + someExternalId` entries.

	// Iterate over a copy of pathMap keys to avoid issues with deleting from map during iteration
	pathsToClean := []string{}
	for p := range s.pathMap {
		// Check if path p starts with mockData.Path + "/" and corresponds to one of the (now deleted) external IDs.
		// This is still indirect. A better way would be if model.MockData stored all its external IDs.
		// For now, we only explicitly know the 'id' used for deletion.
		// Let's assume pathMap stores entries for mockData.Path (collection) and potentially mockData.Location (which is data.Path + "/" + ids[0] from Create).
		if p == mockData.Location { // mockData.Location should be the id-specific path created with the primary external ID
			pathsToClean = append(pathsToClean, p)
		}
	}
	// Also, if the 'id' used for deletion was not the one used to form mockData.Location,
	// its specific path (mockData.Path + "/" + id) might also be in pathMap.
	idSpecificPath := path.Join(mockData.Path, id)
	if _, exists := s.pathMap[idSpecificPath]; exists {
		pathsToClean = append(pathsToClean, idSpecificPath)
	}

	for _, p := range pathsToClean {
		if pathIDs, ok := s.pathMap[p]; ok {
			newPathIDs := []string{}
			for _, sid := range pathIDs {
				if sid != storageID {
					newPathIDs = append(newPathIDs, sid)
				}
			}
			if len(newPathIDs) == 0 {
				delete(s.pathMap, p)
			} else {
				s.pathMap[p] = newPathIDs
			}
		}
	}

	// If the id used for deletion itself formed part of a path key in pathMap
	// e.g. pathMap["/users/id123"] where "id123" is the 'id'
	// This specific path should also be cleaned if it only contained this storageID
	specificPathKey := path.Join(mockData.Path, id) // Assuming id is the last segment
	if currentPathIDs, ok := s.pathMap[specificPathKey]; ok {
		newSpecificPathIDs := []string{}
		for _, sid := range currentPathIDs {
			if sid != storageID {
				newSpecificPathIDs = append(newSpecificPathIDs, sid)
			}
		}
		if len(newSpecificPathIDs) == 0 {
			delete(s.pathMap, specificPathKey)
		} else {
			s.pathMap[specificPathKey] = newSpecificPathIDs
		}
	}

	return nil
}

// ForEach iterates over each stored item
// The 'id' passed to the callback function 'fn' will be one of the external IDs associated with the data.
// To be precise, it's the first external ID encountered when iterating idMap that maps to a given storageID.
// This might not be deterministic if multiple external IDs map to the same storageID.
// If a deterministic "primary" external ID is needed for ForEach, idMap iteration order dependency is an issue.
// For now, it provides *an* external ID.
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
