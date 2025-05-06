package main

import (
	"fmt"
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

	// Add to pathMap
	s.pathMap[data.Path] = append(s.pathMap[data.Path], storageID)

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

	// For collection paths (e.g. /users), return all items under that path
	if strings.Count(strings.Trim(path, "/"), "/") == 0 {
		var result []*MockData
		for _, data := range s.data {
			if strings.HasPrefix(data.Path, path+"/") {
				result = append(result, data)
			}
		}
		if len(result) > 0 {
			return result, nil
		}
		return nil, fmt.Errorf("not found")
	}

	// For specific paths, return exact matches
	var result []*MockData
	for _, data := range s.data {
		if data.Path == path {
			result = append(result, data)
		}
	}
	if len(result) > 0 {
		return result, nil
	}

	// If no exact matches, try to find items with the path as a prefix
	for _, data := range s.data {
		if strings.HasPrefix(data.Path, path+"/") {
			result = append(result, data)
		}
	}
	if len(result) > 0 {
		return result, nil
	}

	return nil, fmt.Errorf("not found")
}

// Delete removes data by ID
func (s *storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If no ID provided, return error
	if id == "" {
		return fmt.Errorf("not found")
	}

	// Try to delete by ID
	storageID, exists := s.idMap[id]
	if !exists {
		return fmt.Errorf("not found")
	}

	// Remove from data
	data := s.data[storageID]
	delete(s.data, storageID)

	// Remove from idMap
	delete(s.idMap, id)

	// Remove from pathMap
	if data != nil {
		if pathIDs, ok := s.pathMap[data.Path]; ok {
			for i, sid := range pathIDs {
				if sid == storageID {
					s.pathMap[data.Path] = append(pathIDs[:i], pathIDs[i+1:]...)
					break
				}
			}
		}
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
