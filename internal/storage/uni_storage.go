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

// UniStorage interface defines the operations for storing and retrieving data
type UniStorage interface {
	Create(sectionName string, isStrictPath bool, data model.UniData) error
	Update(sectionName string, isStrictPath bool, id string, data model.UniData) error
	Get(sectionName string, isStrictPath bool, id string) (model.UniData, error)
	GetByPath(requestPath string) ([]model.UniData, error)
	Delete(sectionName string, isStrictPath bool, id string) error
	ForEach(fn func(id string, data model.UniData) error) error
}

// uniStorage implements the Storage interface
type uniStorage struct {
	mu      *sync.RWMutex
	data    map[string]model.UniData // compositeKey -> data
	pathMap map[string][]string      // path -> []compositeKey
}

// NewUniStorage creates a new instance of storage
func NewUniStorage() UniStorage {
	return &uniStorage{
		mu:      &sync.RWMutex{},
		data:    make(map[string]model.UniData),
		pathMap: make(map[string][]string),
	}
}

// validateID checks if the ID is valid
func (*uniStorage) validateID(id string) error {
	if id == "" {
		return errors.NewInvalidRequestError("ID cannot be empty")
	}
	return nil
}

// buildCompositeKey creates a composite key for storage based on strict_path mode
func (u *uniStorage) buildCompositeKey(sectionName string, isStrictPath bool, resourcePath string, id string) string {
	if isStrictPath {
		return u.buildStrictCompositeKey(resourcePath, id)
	}
	return u.buildNonStrictCompositeKey(sectionName, id)
}

// buildStrictCompositeKey creates a composite key in strict mode: path:id
func (*uniStorage) buildStrictCompositeKey(resourcePath string, id string) string {
	return resourcePath + keySeparator + id
}

// buildNonStrictCompositeKey creates a composite key in non-strict mode: section:id
func (*uniStorage) buildNonStrictCompositeKey(sectionName string, id string) string {
	return sectionName + keySeparator + id
}

// extractIDFromCompositeKey extracts the ID part from a composite key
func (*uniStorage) extractIDFromCompositeKey(compositeKey string) string {
	parts := strings.Split(compositeKey, keySeparator)
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return compositeKey
}

// Create stores new data using IDs from UniData.IDs field with section-aware conflict detection
func (s *uniStorage) Create(sectionName string, isStrictPath bool, data model.UniData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate IDs and check for conflicts based on strict_path mode
	effectiveIDs, err := s.prepareIDsWithConflictCheck(sectionName, isStrictPath, data)
	if err != nil {
		return err
	}

	// Prepare data for storage
	finalIDs := s.prepareDataForStorage(effectiveIDs, &data)

	// Store the data with composite keys
	s.storeDataWithCompositeKeys(sectionName, isStrictPath, finalIDs, data)

	return nil
}

// prepareIDsWithConflictCheck gets IDs from UniData and validates for conflicts based on strict_path mode
func (s *uniStorage) prepareIDsWithConflictCheck(
	sectionName string, isStrictPath bool, data model.UniData,
) ([]string, error) {
	// Use IDs from UniData
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
func (*uniStorage) prepareDataForStorage(effectiveIDs []string, data *model.UniData) []string {
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

	// Update the UniData with the effective IDs
	data.IDs = effectiveIDs

	return effectiveIDs
}

// storeDataWithCompositeKeys stores the data using composite keys and updates path mappings
func (s *uniStorage) storeDataWithCompositeKeys(
	sectionName string, isStrictPath bool, effectiveIDs []string, data model.UniData,
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
func (s *uniStorage) Update(sectionName string, isStrictPath bool, id string, data model.UniData) error {
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
func (s *uniStorage) findExistingResource(
	sectionName string, isStrictPath bool, id string,
) (string, model.UniData, error) {
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
	return "", model.UniData{}, errors.NewNotFoundError(id, "")
}

// findExistingDataOnly finds existing resource data without returning composite key
func (s *uniStorage) findExistingDataOnly(
	sectionName string, isStrictPath bool, id string,
) (model.UniData, error) {
	_, data, err := s.findExistingResource(sectionName, isStrictPath, id)
	return data, err
}

// updatePathMappingsForUpdate handles path map updates when resource path changes
func (s *uniStorage) updatePathMappingsForUpdate(
	oldCompositeKey, newCompositeKey string, oldData, newData model.UniData, id string,
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
func (s *uniStorage) removeCompositeKeyFromPath(compositeKey, resourcePath string) {
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
func (s *uniStorage) Get(sectionName string, isStrictPath bool, id string) (model.UniData, error) {
	if err := s.validateID(id); err != nil {
		return model.UniData{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// For Get operations, we need to search across all possible paths in the section
	// since we don't know the exact path when looking up by ID
	data, err := s.findResourceByID(sectionName, isStrictPath, id)
	if err != nil {
		return model.UniData{}, err
	}
	return data, nil
}

// findResourceByID searches for a resource by ID within the given section and strict mode
func (s *uniStorage) findResourceByID(sectionName string, isStrictPath bool, id string) (model.UniData, error) {
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

	return model.UniData{}, errors.NewNotFoundError(id, "")
}

// isCompositeKeyInScope checks if a composite key belongs to the specified scope
func (*uniStorage) isCompositeKeyInScope(
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
func (s *uniStorage) GetByPath(requestPath string) ([]model.UniData, error) {
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
func (s *uniStorage) getExactPathMatches(requestPath string) []model.UniData {
	var result []model.UniData
	seen := make(map[string]bool)

	if compositeKeys, ok := s.pathMap[requestPath]; ok {
		for _, key := range compositeKeys {
			if data, exists := s.data[key]; exists && !seen[key] {
				seen[key] = true
				result = append(result, data)
			}
		}
	}
	return result
}

// getPrefixPathMatches finds resources with prefix path matches
func (s *uniStorage) getPrefixPathMatches(requestPath string) []model.UniData {
	var result []model.UniData
	seen := make(map[string]bool)

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
func (s *uniStorage) addKeysToResult(
	result []model.UniData, seen map[string]bool, compositeKeys []string,
) []model.UniData {
	for _, key := range compositeKeys {
		if data, exists := s.data[key]; exists && !seen[key] {
			seen[key] = true
			result = append(result, data)
		}
	}
	return result
}

// Delete removes data by ID using section-aware composite key lookup
func (s *uniStorage) Delete(sectionName string, isStrictPath bool, id string) error {
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
func (s *uniStorage) removeAllCompositeKeysForResource(
	sectionName string, isStrictPath bool, mockData model.UniData,
) {
	for _, resourceID := range mockData.IDs {
		compositeKey := s.buildCompositeKey(sectionName, isStrictPath, mockData.Path, resourceID)
		delete(s.data, compositeKey)
	}
}

// ForEach iterates over each stored item
// The 'id' passed to the callback function is the composite key (section:id or path:id)
func (s *uniStorage) ForEach(fn func(id string, data model.UniData) error) error {
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
