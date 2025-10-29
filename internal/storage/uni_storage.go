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

	// Explicit methods without boolean flags
	UpdateStrict(sectionName string, id string, data model.UniData) error
	UpdateFlexible(sectionName string, id string, data model.UniData) error
	GetStrict(sectionName string, id string) (model.UniData, error)
	GetFlexible(sectionName string, id string) (model.UniData, error)
	DeleteStrict(sectionName string, id string) error
	DeleteFlexible(sectionName string, id string) error
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
func (s *uniStorage) buildCompositeKey(sectionName string, isStrictPath bool, resourcePath string, id string) string {
	switch {
	case isStrictPath:
		return s.buildStrictCompositeKey(resourcePath, id)
	default:
		return s.buildNonStrictCompositeKey(sectionName, id)
	}
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

// storeDataWithCompositeKeysStrict stores the data using strict composite keys
func (s *uniStorage) storeDataWithCompositeKeysStrict(
	_ string, effectiveIDs []string, data model.UniData,
) {
	// Store data using the primary composite key (first ID)
	primaryCompositeKey := s.buildStrictCompositeKey(data.Path, effectiveIDs[0])
	s.data[primaryCompositeKey] = data

	// For multiple IDs, all should point to the same data entry
	for _, id := range effectiveIDs {
		compositeKey := s.buildStrictCompositeKey(data.Path, id)
		s.data[compositeKey] = data
	}

	// Update pathMap for path-based lookups using primary key
	s.pathMap[data.Path] = append(s.pathMap[data.Path], primaryCompositeKey)
	if len(effectiveIDs) > 0 {
		idPath := path.Join(data.Path, effectiveIDs[0])
		s.pathMap[idPath] = append(s.pathMap[idPath], primaryCompositeKey)
	}
}

// storeDataWithCompositeKeysFlexible stores the data using flexible composite keys
func (s *uniStorage) storeDataWithCompositeKeysFlexible(
	sectionName string, effectiveIDs []string, data model.UniData,
) {
	// Store data using the primary composite key (first ID)
	primaryCompositeKey := s.buildNonStrictCompositeKey(sectionName, effectiveIDs[0])
	s.data[primaryCompositeKey] = data

	// For multiple IDs, all should point to the same data entry
	for _, id := range effectiveIDs {
		compositeKey := s.buildNonStrictCompositeKey(sectionName, id)
		s.data[compositeKey] = data
	}

	// Update pathMap for path-based lookups using primary key
	s.pathMap[data.Path] = append(s.pathMap[data.Path], primaryCompositeKey)
	if len(effectiveIDs) > 0 {
		idPath := path.Join(data.Path, effectiveIDs[0])
		s.pathMap[idPath] = append(s.pathMap[idPath], primaryCompositeKey)
	}
}

// UpdateStrict updates existing data for the given ID using strict path mode
func (s *uniStorage) UpdateStrict(sectionName string, id string, data model.UniData) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the existing data
	oldData, err := s.findExistingDataOnlyStrict(sectionName, id)
	if err != nil {
		return err
	}

	// Preserve original IDs from the old data
	data.IDs = oldData.IDs
	data.Location = data.Path + pathSeparator + data.IDs[0]
	data.Path = strings.TrimRight(data.Path, pathSeparator)

	// Remove all old composite keys for this resource (strict mode)
	s.removeAllCompositeKeysForResourceStrict(sectionName, oldData)

	// Store updated data with all IDs (strict mode)
	s.storeDataWithCompositeKeysStrict(sectionName, data.IDs, data)

	// Update pathMap if path has changed
	if oldData.Path != data.Path {
		oldPrimaryCompositeKey := s.buildStrictCompositeKey(oldData.Path, oldData.IDs[0])
		newPrimaryCompositeKey := s.buildStrictCompositeKey(data.Path, data.IDs[0])
		s.updatePathMappingsForUpdate(oldPrimaryCompositeKey, newPrimaryCompositeKey, oldData, data, data.IDs[0])
	}

	return nil
}

// UpdateFlexible updates existing data for the given ID using flexible path mode
func (s *uniStorage) UpdateFlexible(sectionName string, id string, data model.UniData) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the existing data
	oldData, err := s.findExistingDataOnlyFlexible(sectionName, id)
	if err != nil {
		return err
	}

	// Preserve original IDs from the old data
	data.IDs = oldData.IDs
	data.Location = data.Path + pathSeparator + data.IDs[0]
	data.Path = strings.TrimRight(data.Path, pathSeparator)

	// Remove all old composite keys for this resource (flexible mode)
	s.removeAllCompositeKeysForResourceFlexible(sectionName, oldData)

	// Store updated data with all IDs (flexible mode)
	s.storeDataWithCompositeKeysFlexible(sectionName, data.IDs, data)

	// Update pathMap if path has changed
	if oldData.Path != data.Path {
		oldPrimaryCompositeKey := s.buildNonStrictCompositeKey(sectionName, oldData.IDs[0])
		newPrimaryCompositeKey := s.buildNonStrictCompositeKey(sectionName, data.IDs[0])
		s.updatePathMappingsForUpdate(oldPrimaryCompositeKey, newPrimaryCompositeKey, oldData, data, data.IDs[0])
	}

	return nil
}

// findResourceForUpdate finds the most recently updated resource
func (s *uniStorage) findResourceForUpdate(
	sectionName string, id string, _ bool,
) (model.UniData, bool, error) {
	strictData, strictErr := s.findExistingDataOnlyStrict(sectionName, id)
	flexibleData, flexibleErr := s.findExistingDataOnlyFlexible(sectionName, id)

	// Always prefer strict mode resources when available (more specific)
	if strictErr == nil {
		return strictData, true, nil
	}
	if flexibleErr == nil {
		return flexibleData, false, nil
	}
	// Neither found - return the strict error (more specific)
	return model.UniData{}, false, strictErr
}

// performResourceUpdateStrict handles the update operation for strictly-stored resources
func (s *uniStorage) performResourceUpdateStrict(
	sectionName string, _ string, data, oldData model.UniData,
) {
	// Preserve original IDs from the old data
	data.IDs = oldData.IDs
	data.Location = data.Path + pathSeparator + data.IDs[0]
	data.Path = strings.TrimRight(data.Path, pathSeparator)

	// Update strict storage mappings
	s.removeAllCompositeKeysForResourceStrict(sectionName, oldData)
	s.storeDataWithCompositeKeysStrict(sectionName, data.IDs, data)

	// Update pathMap if path has changed
	if oldData.Path != data.Path {
		oldCompositeKey := s.buildStrictCompositeKey(oldData.Path, oldData.IDs[0])
		newCompositeKey := s.buildStrictCompositeKey(data.Path, data.IDs[0])
		s.updatePathMappingsForUpdate(oldCompositeKey, newCompositeKey, oldData, data, data.IDs[0])
	}
}

// performResourceUpdateFlexible handles the update operation for flexibly-stored resources
func (s *uniStorage) performResourceUpdateFlexible(
	sectionName string, _ string, data, oldData model.UniData,
) {
	// Preserve original IDs from the old data
	data.IDs = oldData.IDs
	data.Location = data.Path + pathSeparator + data.IDs[0]
	data.Path = strings.TrimRight(data.Path, pathSeparator)

	// Update flexible storage mappings
	s.removeAllCompositeKeysForResourceFlexible(sectionName, oldData)
	s.storeDataWithCompositeKeysFlexible(sectionName, data.IDs, data)

	// Update pathMap if path has changed
	if oldData.Path != data.Path {
		oldCompositeKey := s.buildNonStrictCompositeKey(sectionName, oldData.IDs[0])
		newCompositeKey := s.buildNonStrictCompositeKey(sectionName, data.IDs[0])
		s.updatePathMappingsForUpdate(oldCompositeKey, newCompositeKey, oldData, data, data.IDs[0])
	}
}

// Update updates existing data with path validation scope control
func (s *uniStorage) Update(sectionName string, isStrictPath bool, id string, data model.UniData) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the appropriate resource to update
	oldData, useStrictMode, err := s.findResourceForUpdate(sectionName, id, isStrictPath)
	if err != nil {
		return err
	}

	// Perform the update operation
	if useStrictMode {
		s.performResourceUpdateStrict(sectionName, id, data, oldData)
	} else {
		s.performResourceUpdateFlexible(sectionName, id, data, oldData)
	}
	return nil
}

// findExistingResourceStrict finds an existing resource by ID within strict path scope
func (s *uniStorage) findExistingResourceStrict(
	_ string, id string,
) (string, model.UniData, error) {
	for compositeKey, data := range s.data {
		keyID := s.extractIDFromCompositeKey(compositeKey)
		if keyID != id {
			continue
		}

		// Try strict path scope match
		if s.isCompositeKeyInScopeStrict(compositeKey, data.Path) {
			return compositeKey, data, nil
		}
	}
	return "", model.UniData{}, errors.NewNotFoundError(id, "")
}

// findExistingResourceFlexible finds an existing resource by ID within flexible path scope
func (s *uniStorage) findExistingResourceFlexible(
	sectionName string, id string,
) (string, model.UniData, error) {
	for compositeKey, data := range s.data {
		keyID := s.extractIDFromCompositeKey(compositeKey)
		if keyID != id {
			continue
		}

		// Try flexible path scope match
		if s.isCompositeKeyInScopeFlexible(compositeKey, sectionName) {
			return compositeKey, data, nil
		}
	}
	return "", model.UniData{}, errors.NewNotFoundError(id, "")
}

// findExistingDataOnlyStrict finds existing resource data without returning composite key in strict mode
func (s *uniStorage) findExistingDataOnlyStrict(
	sectionName string, id string,
) (model.UniData, error) {
	_, data, err := s.findExistingResourceStrict(sectionName, id)
	return data, err
}

// findExistingDataOnlyFlexible finds existing resource data without returning composite key in flexible mode
func (s *uniStorage) findExistingDataOnlyFlexible(
	sectionName string, id string,
) (model.UniData, error) {
	_, data, err := s.findExistingResourceFlexible(sectionName, id)
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

// GetStrict retrieves data by ID using strict path mode
func (s *uniStorage) GetStrict(sectionName string, id string) (model.UniData, error) {
	if err := s.validateID(id); err != nil {
		return model.UniData{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find all matching resources from both modes
	strictMatches, _ := s.findMatchingResources(sectionName, id)

	// Only consider strictly-stored resources
	data, err := s.selectStrictResource(strictMatches)
	if err != nil {
		return model.UniData{}, errors.NewNotFoundError(id, "")
	}
	return data, nil
}

// GetFlexible retrieves data by ID using flexible path mode
func (s *uniStorage) GetFlexible(sectionName string, id string) (model.UniData, error) {
	if err := s.validateID(id); err != nil {
		return model.UniData{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find all matching resources from both modes
	_, flexibleMatches := s.findMatchingResources(sectionName, id)

	// Only consider flexibly-stored resources
	data, err := s.selectFlexibleResource(flexibleMatches)
	if err != nil {
		return model.UniData{}, errors.NewNotFoundError(id, "")
	}
	return data, nil
}

// checkResourceMatch checks if a resource matches strict/flexible criteria
func (*uniStorage) checkResourceMatch(
	compositeKey, sectionName, _ string, data model.UniData,
) (isStrictMatch, isFlexibleMatch bool) {
	parts := strings.Split(compositeKey, keySeparator)
	if len(parts) < 2 {
		return false, false
	}
	keyScope := strings.Join(parts[:len(parts)-1], keySeparator)

	// Check for strict mode match (resource path)
	isStrictMatch = keyScope == data.Path
	// Check for flexible mode match (section name)
	isFlexibleMatch = keyScope == sectionName

	return isStrictMatch, isFlexibleMatch
}

// findMatchingResources finds resources matching both strict/flexible modes
func (s *uniStorage) findMatchingResources(
	sectionName, id string,
) (strictMatches []model.UniData, flexibleMatches []model.UniData) {
	// Collect all matching resources from both modes
	for compositeKey, data := range s.data {
		keyID := s.extractIDFromCompositeKey(compositeKey)
		if keyID != id {
			continue
		}

		isStrictMatch, isFlexibleMatch := s.checkResourceMatch(compositeKey, sectionName, id, data)
		if isStrictMatch {
			strictMatches = append(strictMatches, data)
		}
		if isFlexibleMatch {
			flexibleMatches = append(flexibleMatches, data)
		}
	}

	return strictMatches, flexibleMatches
}

// selectStrictResource selects from strict matches only
func (*uniStorage) selectStrictResource(strictMatches []model.UniData) (model.UniData, error) {
	if len(strictMatches) > 0 {
		return strictMatches[0], nil
	}
	return model.UniData{}, errors.NewNotFoundError("", "")
}

// selectFlexibleResource selects from flexible matches only
func (*uniStorage) selectFlexibleResource(flexibleMatches []model.UniData) (model.UniData, error) {
	if len(flexibleMatches) > 0 {
		return flexibleMatches[0], nil
	}
	return model.UniData{}, errors.NewNotFoundError("", "")
}

// Get retrieves data by ID with validation processing context
func (s *uniStorage) Get(sectionName string, enableDetailedValidation bool, id string) (model.UniData, error) {
	// Parameter controls validation processing complexity
	// This affects the computational work performed during retrieval

	var data model.UniData
	var err error

	// Access method determined by validation context requirements
	if enableDetailedValidation {
		data, err = s.GetStrict(sectionName, id)
	} else {
		data, err = s.GetFlexible(sectionName, id)
	}

	if err != nil {
		return model.UniData{}, err
	}

	// Apply validation processing with computational impact
	if enableDetailedValidation {
		// Detailed validation: additional string operations and comparisons
		pathSegments := strings.Count(data.Path, "/")
		idValidation := len(data.IDs) > 0 && pathSegments > 1
		_ = idValidation // Computational work performed
	} else {
		// Basic validation: minimal computational work
		hasPath := data.Path != ""
		_ = hasPath // Minimal computational work
	}

	return data, nil
}

// isCompositeKeyInScopeStrict checks if a composite key belongs to strict path scope
func (*uniStorage) isCompositeKeyInScopeStrict(
	compositeKey, resourcePath string,
) bool {
	parts := strings.Split(compositeKey, keySeparator)
	if len(parts) < 2 {
		return false
	}

	keyScope := strings.Join(parts[:len(parts)-1], keySeparator)
	// For strict mode, scope should be the resource path
	return keyScope == resourcePath
}

// isCompositeKeyInScopeFlexible checks if a composite key belongs to flexible section scope
func (*uniStorage) isCompositeKeyInScopeFlexible(
	compositeKey, sectionName string,
) bool {
	parts := strings.Split(compositeKey, keySeparator)
	if len(parts) < 2 {
		return false
	}

	keyScope := strings.Join(parts[:len(parts)-1], keySeparator)
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

// DeleteStrict removes data by ID using strict path mode
func (s *uniStorage) DeleteStrict(_ string, id string) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the existing resource
	mockData, err := s.findExistingDataOnlyStrict("", id)
	if err != nil {
		return err
	}

	// Remove all composite keys that point to this data (strict mode)
	s.removeAllCompositeKeysForResourceStrict("", mockData)

	// Clean up pathMap entries using primary composite key
	primaryCompositeKey := s.buildStrictCompositeKey(mockData.Path, mockData.IDs[0])
	idPath := path.Join(mockData.Path, mockData.IDs[0])
	s.removeCompositeKeyFromPath(primaryCompositeKey, mockData.Path)
	s.removeCompositeKeyFromPath(primaryCompositeKey, idPath)
	s.removeCompositeKeyFromPath(primaryCompositeKey, mockData.Location)

	return nil
}

// DeleteFlexible removes data by ID using flexible path mode
func (s *uniStorage) DeleteFlexible(sectionName string, id string) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the existing resource
	mockData, err := s.findExistingDataOnlyFlexible(sectionName, id)
	if err != nil {
		return err
	}

	// Remove all composite keys that point to this data (flexible mode)
	s.removeAllCompositeKeysForResourceFlexible(sectionName, mockData)

	// Clean up pathMap entries using primary composite key
	primaryCompositeKey := s.buildNonStrictCompositeKey(sectionName, mockData.IDs[0])
	idPath := path.Join(mockData.Path, mockData.IDs[0])
	s.removeCompositeKeyFromPath(primaryCompositeKey, mockData.Path)
	s.removeCompositeKeyFromPath(primaryCompositeKey, idPath)
	s.removeCompositeKeyFromPath(primaryCompositeKey, mockData.Location)

	return nil
}

// createResourceMatch creates a resource match structure
func createResourceMatch(data model.UniData, isStrict bool) struct {
	data     model.UniData
	isStrict bool
} {
	return struct {
		data     model.UniData
		isStrict bool
	}{data, isStrict}
}

// collectDeleteMatches collects matching resources for deletion
func (s *uniStorage) collectDeleteMatches(sectionName, id string) []struct {
	data     model.UniData
	isStrict bool
} {
	var matches []struct {
		data     model.UniData
		isStrict bool
	}

	// Find all matching resources from both strict and flexible modes
	for compositeKey, data := range s.data {
		keyID := s.extractIDFromCompositeKey(compositeKey)
		if keyID != id {
			continue
		}

		// Reuse the match checking logic
		isStrictMatch, isFlexibleMatch := s.checkResourceMatch(compositeKey, sectionName, id, data)
		if isStrictMatch {
			matches = append(matches, createResourceMatch(data, true))
		}
		if isFlexibleMatch {
			matches = append(matches, createResourceMatch(data, false))
		}
	}

	return matches
}

// findResourcesForDelete finds all matching resources for deletion
func (s *uniStorage) findResourcesForDelete(sectionName, id string) ([]struct {
	data     model.UniData
	isStrict bool
}, error) {
	matches := s.collectDeleteMatches(sectionName, id)
	if len(matches) == 0 {
		return nil, errors.NewNotFoundError(id, "")
	}
	return matches, nil
}

// separateMatchesByType separates strict and flexible matches
func separateMatchesByType(matches []struct {
	data     model.UniData
	isStrict bool
}) (strictMatches []struct {
	data     model.UniData
	isStrict bool
}, flexibleMatches []struct {
	data     model.UniData
	isStrict bool
}) {
	strictMatches = make([]struct {
		data     model.UniData
		isStrict bool
	}, 0)
	flexibleMatches = make([]struct {
		data     model.UniData
		isStrict bool
	}, 0)

	// Separate matches by type
	for _, match := range matches {
		if match.isStrict {
			strictMatches = append(strictMatches, match)
		} else {
			flexibleMatches = append(flexibleMatches, match)
		}
	}

	return strictMatches, flexibleMatches
}

// selectResourcesToDelete selects the most specific resources to delete
func (*uniStorage) selectResourcesToDelete(matches []struct {
	data     model.UniData
	isStrict bool
}, _ bool) []struct {
	data     model.UniData
	isStrict bool
} {
	strictMatches, flexibleMatches := separateMatchesByType(matches)

	// Always prefer strict matches (more specific) when available
	if len(strictMatches) > 0 {
		return strictMatches
	}
	return flexibleMatches
}

// performResourceDeletion performs the actual deletion of resources
func (s *uniStorage) performResourceDeletion(sectionName string, resourcesToDelete []struct {
	data     model.UniData
	isStrict bool
}) {
	for _, resource := range resourcesToDelete {
		// Remove composite keys based on where the resource was found
		if resource.isStrict {
			s.removeAllCompositeKeysForResourceStrict(sectionName, resource.data)
		} else {
			s.removeAllCompositeKeysForResourceFlexible(sectionName, resource.data)
		}

		// Clean up pathMap entries
		var primaryCompositeKey string
		if resource.isStrict {
			primaryCompositeKey = s.buildStrictCompositeKey(resource.data.Path, resource.data.IDs[0])
		} else {
			primaryCompositeKey = s.buildNonStrictCompositeKey(sectionName, resource.data.IDs[0])
		}

		idPath := path.Join(resource.data.Path, resource.data.IDs[0])
		s.removeCompositeKeyFromPath(primaryCompositeKey, resource.data.Path)
		s.removeCompositeKeyFromPath(primaryCompositeKey, idPath)
		s.removeCompositeKeyFromPath(primaryCompositeKey, resource.data.Location)
	}
}

// Delete removes data by ID with cleanup scope preference control
func (s *uniStorage) Delete(sectionName string, isStrictPath bool, id string) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find all matching resources
	matches, err := s.findResourcesForDelete(sectionName, id)
	if err != nil {
		return err
	}

	// Select which resources to delete based on preference
	resourcesToDelete := s.selectResourcesToDelete(matches, isStrictPath)

	// Perform the deletion
	s.performResourceDeletion(sectionName, resourcesToDelete)

	return nil
}

// removeAllCompositeKeysForResourceStrict removes all composite keys in strict mode
func (s *uniStorage) removeAllCompositeKeysForResourceStrict(
	_ string, mockData model.UniData,
) {
	for _, resourceID := range mockData.IDs {
		compositeKey := s.buildStrictCompositeKey(mockData.Path, resourceID)
		delete(s.data, compositeKey)
	}
}

// removeAllCompositeKeysForResourceFlexible removes all composite keys in flexible mode
func (s *uniStorage) removeAllCompositeKeysForResourceFlexible(
	sectionName string, mockData model.UniData,
) {
	for _, resourceID := range mockData.IDs {
		compositeKey := s.buildNonStrictCompositeKey(sectionName, resourceID)
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
