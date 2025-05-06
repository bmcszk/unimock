package main

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// MockData represents the data to be stored
type MockData struct {
	Path        string
	ContentType string
	Body        []byte
}

// Storage interface defines the operations for storing and retrieving data
type Storage interface {
	Create(ids []string, data *MockData) error
	Update(id string, data *MockData) error
	Get(id string) (*MockData, error)
	GetByPath(path string) ([]*MockData, error)
	Delete(id string) error
	ForEach(fn func(id string, data *MockData) error) error
}

// storage implements the Storage interface
type storage struct {
	mu      sync.RWMutex
	data    map[string]*MockData // storageID -> data
	idMap   map[string]string    // externalID -> storageID
	pathMap map[string][]string  // path -> []storageID
}

// NewStorage creates a new instance of storage
func NewStorage() Storage {
	return &storage{
		data:    make(map[string]*MockData),
		idMap:   make(map[string]string),
		pathMap: make(map[string][]string),
	}
}

// Create stores new data with the given IDs
func (s *storage) Create(ids []string, data *MockData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a new storage ID
	storageID := uuid.New().String()

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
func (s *storage) Update(id string, data *MockData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If no ID provided, return error
	if id == "" {
		return fmt.Errorf("not found")
	}

	// Get existing storage ID
	storageID, exists := s.idMap[id]
	if !exists {
		return fmt.Errorf("not found")
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
func (s *storage) Get(id string) (*MockData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storageID, exists := s.idMap[id]
	if !exists {
		return nil, fmt.Errorf("not found")
	}

	data, exists := s.data[storageID]
	if !exists {
		return nil, fmt.Errorf("not found")
	}

	return data, nil
}

// GetByPath retrieves all data stored at the given path
func (s *storage) GetByPath(path string) ([]*MockData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*MockData
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

	return nil, fmt.Errorf("not found")
}

// Delete removes data by ID or path
func (s *storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try to delete by ID first
	if storageID, exists := s.idMap[id]; exists {
		data := s.data[storageID]
		delete(s.data, storageID)
		delete(s.idMap, id)

		// Remove from pathMap
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
		return fmt.Errorf("not found")
	}

	return nil
}

// ForEach iterates over all stored data
func (s *storage) ForEach(fn func(id string, data *MockData) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, data := range s.data {
		if err := fn(id, data); err != nil {
			return err
		}
	}

	return nil
}
