package storage

import (
	"path"
	"strings"
	"sync"

	"github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/pkg/model"

	"github.com/google/uuid"
)

const (
	// pathSeparator represents the separator used in URL paths
	pathSeparator = "/"
	// keySeparator separates section/path from ID in composite keys
	keySeparator = ":"
)

// MockStorage interface defines the operations for storing and retrieving data
type MockStorage interface {
	Create(sectionName string, isStrictPath bool, data *model.MockData) error
	Update(sectionName string, isStrictPath bool, id string, data *model.MockData) error
	Get(sectionName string, isStrictPath bool, id string) (*model.MockData, error)
	GetByPath(requestPath string) ([]*model.MockData, error)
	Delete(sectionName string, isStrictPath bool, id string) error
	ForEach(fn func(id string, data *model.MockData) error) error
}

// mockStorage implements the Storage interface
type mockStorage struct {
	mu      *sync.RWMutex
	data    map[string]*model.MockData // compositeKey -> data
	pathMap map[string][]string        // path -> []compositeKey
}

// NewMockStorage creates a new instance of storage
func NewMockStorage() MockStorage {
	return &mockStorage{
		mu:      &sync.RWMutex{},
		data:    make(map[string]*model.MockData),
		pathMap: make(map[string][]string),
	}
}

// validateData checks if the data is valid
func (*mockStorage) validateData(data *model.MockData) error {
	if data == nil {
		return errors.NewInvalidRequestError("data cannot be nil")
	}
	return nil
}

// validateID checks if the ID is valid
func (*mockStorage) validateID(id string) error {
	if id == "" {
		return errors.NewInvalidRequestError("ID cannot be empty")
	}
	return nil
}

// buildCompositeKey creates a composite key for storage based on strict_path mode
//nolint:revive // isStrictPath flag is core to the design requirements
func (*mockStorage) buildCompositeKey(sectionName string, isStrictPath bool, resourcePath string, id string) string {
	if isStrictPath {
		// Strict mode: path:id (e.g., "/users/subpath:123")
		return resourcePath + keySeparator + id
	}
	// Non-strict mode: section:id (e.g., "users:123")
	return sectionName + keySeparator + id
}

// extractIDFromCompositeKey extracts the ID part from a composite key
func (*mockStorage) extractIDFromCompositeKey(compositeKey string) string {
	parts := strings.Split(compositeKey, keySeparator)
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return compositeKey
}

// Create stores new data using IDs from MockData.IDs field with section-aware conflict detection
func (s *mockStorage) Create(sectionName string, isStrictPath bool, data *model.MockData) error {
	if err := s.validateData(data); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate IDs and check for conflicts based on strict_path mode
	effectiveIDs, err := s.prepareIDsWithConflictCheck(sectionName, isStrictPath, data)
	if err != nil {
		return err
	}

	// Prepare data for storage
	finalIDs := s.prepareDataForStorage(effectiveIDs, data)

	// Store the data with composite keys
	s.storeDataWithCompositeKeys(sectionName, isStrictPath, finalIDs, data)

	return nil
}

// prepareIDsWithConflictCheck gets IDs from MockData and validates for conflicts based on strict_path mode
func (s *mockStorage) prepareIDsWithConflictCheck(
	sectionName string, isStrictPath bool, data *model.MockData,
) ([]string, error) {
	// Use IDs from MockData
	effectiveIDs := data.IDs

	// Check for conflicts using composite keys
	for _, id := range effectiveIDs {
		compositeKey := s.buildCompositeKey(sectionName, isStrictPath, data.Path, id)
		if _, exists := s.data[compositeKey]; exists {
			return nil, errors.NewConflictError(id)
		}
	}

	return effectiveIDs, nil
}

// prepareDataForStorage sets up data location and handles ID generation
func (*mockStorage) prepareDataForStorage(effectiveIDs []string, data *model.MockData) []string {
	// Ensure path doesn't have trailing slash
	data.Path = strings.TrimRight(data.Path, pathSeparator)

	// Set location based on path and first ID
	if len(effectiveIDs) > 0 {
		data.Location = data.Path + pathSeparator + effectiveIDs[0]
	} else {
		// Generate UUID for path-based storage
		generatedID := uuid.New().String()
		data.Location = data.Path + pathSeparator + generatedID
		effectiveIDs = []string{generatedID}
	}

	// Update the MockData with the effective IDs
	data.IDs = effectiveIDs

	return effectiveIDs
}

// storeDataWithCompositeKeys stores the data using composite keys and updates path mappings
func (s *mockStorage) storeDataWithCompositeKeys(
	sectionName string, isStrictPath bool, effectiveIDs []string, data *model.MockData,
) {
	// Store data using the primary composite key (first ID)
	primaryCompositeKey := s.buildCompositeKey(sectionName, isStrictPath, data.Path, effectiveIDs[0])
	s.data[primaryCompositeKey] = data
	
	// For multiple IDs, all should point to the same data entry
	// We achieve this by having all composite keys reference the same data object
	for _, id := range effectiveIDs {
		compositeKey := s.buildCompositeKey(sectionName, isStrictPath, data.Path, id)
		s.data[compositeKey] = data
	}

	// Update pathMap for path-based lookups using primary key
	s.pathMap[data.Path] = append(s.pathMap[data.Path], primaryCompositeKey)
	if len(effectiveIDs) > 0 {
		idPath := path.Join(data.Path, effectiveIDs[0])
		s.pathMap[idPath] = append(s.pathMap[idPath], primaryCompositeKey)
	}
}

// Update updates existing data for the given ID using section-aware composite key lookup
func (s *mockStorage) Update(sectionName string, isStrictPath bool, id string, data *model.MockData) error {
	if err := s.validateData(data); err != nil {
		return err
	}
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the existing data
	oldData, err := s.findExistingDataOnly(sectionName, isStrictPath, id)
	if err != nil {
		return err
	}

	// Preserve original IDs from the old data
	data.IDs = oldData.IDs
	data.Location = data.Path + pathSeparator + data.IDs[0]
	data.Path = strings.TrimRight(data.Path, pathSeparator)

	// Remove all old composite keys for this resource
	s.removeAllCompositeKeysForResource(sectionName, isStrictPath, oldData)
	
	// Store updated data with all IDs
	s.storeDataWithCompositeKeys(sectionName, isStrictPath, data.IDs, data)

	// Update pathMap if path has changed
	if oldData.Path != data.Path {
		oldPrimaryCompositeKey := s.buildCompositeKey(sectionName, isStrictPath, oldData.Path, oldData.IDs[0])
		newPrimaryCompositeKey := s.buildCompositeKey(sectionName, isStrictPath, data.Path, data.IDs[0])
		s.updatePathMappingsForUpdate(oldPrimaryCompositeKey, newPrimaryCompositeKey, oldData, data, data.IDs[0])
	}

	return nil
}

// findExistingResource finds an existing resource by ID within the section scope
//nolint:revive // isStrictPath flag is core to the design requirements
func (s *mockStorage) findExistingResource(
	sectionName string, isStrictPath bool, id string,
) (string, *model.MockData, error) {
	for compositeKey, data := range s.data {
		keyID := s.extractIDFromCompositeKey(compositeKey)
		if keyID != id {
			continue
		}
		
		// First try exact scope match
		if s.isCompositeKeyInScope(compositeKey, sectionName, isStrictPath, data.Path) {
			return compositeKey, data, nil
		}
		
		// Then try cross-strict_path lookup within the same section
		if s.isCompositeKeyInScope(compositeKey, sectionName, !isStrictPath, data.Path) {
			return compositeKey, data, nil
		}
	}
	return "", nil, errors.NewNotFoundError(id, "")
}

// findExistingDataOnly finds existing resource data without returning composite key
func (s *mockStorage) findExistingDataOnly(
	sectionName string, isStrictPath bool, id string,
) (*model.MockData, error) {
	_, data, err := s.findExistingResource(sectionName, isStrictPath, id)
	return data, err
}

// updatePathMappingsForUpdate handles path map updates when resource path changes
func (s *mockStorage) updatePathMappingsForUpdate(
	oldCompositeKey, newCompositeKey string, oldData, newData *model.MockData, id string,
) {
	// Remove from old paths
	oldIDPath := path.Join(oldData.Path, id)
	s.removeCompositeKeyFromPath(oldCompositeKey, oldData.Path)
	s.removeCompositeKeyFromPath(oldCompositeKey, oldIDPath)
	
	// Add to new paths
	newIDPath := path.Join(newData.Path, id)
	s.pathMap[newData.Path] = append(s.pathMap[newData.Path], newCompositeKey)
	s.pathMap[newIDPath] = append(s.pathMap[newIDPath], newCompositeKey)
}


// removeCompositeKeyFromPath removes composite key from a specific path mapping
func (s *mockStorage) removeCompositeKeyFromPath(compositeKey, resourcePath string) {
	pathKeys, ok := s.pathMap[resourcePath]
	if !ok {
		return
	}

	for i, key := range pathKeys {
		if key == compositeKey {
			s.pathMap[resourcePath] = append(pathKeys[:i], pathKeys[i+1:]...)
			break
		}
	}

	if len(s.pathMap[resourcePath]) == 0 {
		delete(s.pathMap, resourcePath)
	}
}

// Get retrieves data by ID using section-aware composite key lookup
func (s *mockStorage) Get(sectionName string, isStrictPath bool, id string) (*model.MockData, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// For Get operations, we need to search across all possible paths in the section
	// since we don't know the exact path when looking up by ID
	return s.findResourceByID(sectionName, isStrictPath, id)
}

// findResourceByID searches for a resource by ID within the given section and strict mode
//nolint:revive // isStrictPath flag is core to the design requirements
func (s *mockStorage) findResourceByID(sectionName string, isStrictPath bool, id string) (*model.MockData, error) {
	// Search through all stored data to find matching composite key
	for compositeKey, data := range s.data {
		// Extract the ID from the composite key
		keyID := s.extractIDFromCompositeKey(compositeKey)
		if keyID != id {
			continue
		}
		
		// Only return resources that match both section and strict mode
		if s.isCompositeKeyInScope(compositeKey, sectionName, isStrictPath, data.Path) {
			return data, nil
		}
	}
	
	return nil, errors.NewNotFoundError(id, "")
}

// isCompositeKeyInScope checks if a composite key belongs to the specified scope
//nolint:revive // isStrictPath flag is core to the design requirements
func (*mockStorage) isCompositeKeyInScope(
	compositeKey, sectionName string, isStrictPath bool, resourcePath string,
) bool {
	parts := strings.Split(compositeKey, keySeparator)
	if len(parts) < 2 {
		return false
	}
	
	keyScope := strings.Join(parts[:len(parts)-1], keySeparator)
	
	if isStrictPath {
		// For strict mode, scope should be the resource path
		return keyScope == resourcePath
	}
	// For non-strict mode, scope should be the section name
	return keyScope == sectionName
}

// GetByPath retrieves all data stored at the given path
func (s *mockStorage) GetByPath(requestPath string) ([]*model.MockData, error) {
	if requestPath == "" {
		return nil, errors.NewInvalidRequestError("path cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Normalize path by removing trailing slash
	requestPath = strings.TrimSuffix(requestPath, pathSeparator)
	
	result := s.getExactPathMatches(requestPath)
	if len(result) > 0 {
		return result, nil
	}
	
	result = s.getPrefixPathMatches(requestPath)
	if len(result) > 0 {
		return result, nil
	}

	return nil, errors.NewNotFoundError("resource not found", requestPath)
}

// getExactPathMatches finds resources with exact path matches
func (s *mockStorage) getExactPathMatches(requestPath string) []*model.MockData {
	var result []*model.MockData
	seen := make(map[*model.MockData]bool)
	
	if compositeKeys, ok := s.pathMap[requestPath]; ok {
		for _, key := range compositeKeys {
			if data, exists := s.data[key]; exists && !seen[data] {
				seen[data] = true
				result = append(result, data)
			}
		}
	}
	return result
}

// getPrefixPathMatches finds resources with prefix path matches
func (s *mockStorage) getPrefixPathMatches(requestPath string) []*model.MockData {
	var result []*model.MockData
	seen := make(map[*model.MockData]bool)
	
	for storedPath, compositeKeys := range s.pathMap {
		storedPath = strings.TrimSuffix(storedPath, pathSeparator)
		if !strings.HasPrefix(storedPath, requestPath+pathSeparator) {
			continue
		}
		result = s.addKeysToResult(result, seen, compositeKeys)
	}
	return result
}

// addKeysToResult adds composite keys to result if they exist and haven't been seen
func (s *mockStorage) addKeysToResult(
	result []*model.MockData, seen map[*model.MockData]bool, compositeKeys []string,
) []*model.MockData {
	for _, key := range compositeKeys {
		if data, exists := s.data[key]; exists && !seen[data] {
			seen[data] = true
			result = append(result, data)
		}
	}
	return result
}

// Delete removes data by ID using section-aware composite key lookup
func (s *mockStorage) Delete(sectionName string, isStrictPath bool, id string) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the existing resource
	mockData, err := s.findExistingDataOnly(sectionName, isStrictPath, id)
	if err != nil {
		return err
	}

	// Remove all composite keys that point to this data
	s.removeAllCompositeKeysForResource(sectionName, isStrictPath, mockData)

	// Clean up pathMap entries using primary composite key
	primaryCompositeKey := s.buildCompositeKey(sectionName, isStrictPath, mockData.Path, mockData.IDs[0])
	idPath := path.Join(mockData.Path, mockData.IDs[0])
	s.removeCompositeKeyFromPath(primaryCompositeKey, mockData.Path)
	s.removeCompositeKeyFromPath(primaryCompositeKey, idPath)
	s.removeCompositeKeyFromPath(primaryCompositeKey, mockData.Location)

	return nil
}

// removeAllCompositeKeysForResource removes all composite keys that reference the same data
func (s *mockStorage) removeAllCompositeKeysForResource(
	sectionName string, isStrictPath bool, mockData *model.MockData,
) {
	for _, resourceID := range mockData.IDs {
		compositeKey := s.buildCompositeKey(sectionName, isStrictPath, mockData.Path, resourceID)
		delete(s.data, compositeKey)
	}
}

// ForEach iterates over each stored item
// The 'id' passed to the callback function is the composite key (section:id or path:id)
func (s *mockStorage) ForEach(fn func(id string, data *model.MockData) error) error {
	if fn == nil {
		return errors.NewInvalidRequestError("callback function cannot be nil")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for compositeKey, data := range s.data {
		if err := fn(compositeKey, data); err != nil {
			return errors.NewStorageError("forEach", err)
		}
	}

	return nil
}
