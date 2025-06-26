//go:build !e2e

package storage_test

import (
	"testing"

	"github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompositeKeyConflictDetection tests the new composite key-based conflict detection
func TestCompositeKeyConflictDetection(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(store storage.UniStorage)
		sectionName    string
		isStrictPath   bool
		resourcePath   string
		id             string
		expectConflict bool
		description    string
	}{
		{
			name: "strict_path=true: different paths, same ID - no conflict",
			setup: func(store storage.UniStorage) {
				data := model.UniData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				store.Create("users", true, data)
			},
			sectionName:    "users",
			isStrictPath:   true,
			resourcePath:   "/users/different",
			id:             "123",
			expectConflict: false,
			description:    "With strict_path=true, same ID on different paths should not conflict",
		},
		{
			name: "strict_path=true: same path, same ID - conflict",
			setup: func(store storage.UniStorage) {
				data := model.UniData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				store.Create("users", true, data)
			},
			sectionName:    "users",
			isStrictPath:   true,
			resourcePath:   "/users/subpath",
			id:             "123",
			expectConflict: true,
			description:    "With strict_path=true, same path and ID should conflict",
		},
		{
			name: "strict_path=false: different paths, same ID - conflict",
			setup: func(store storage.UniStorage) {
				data := model.UniData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				store.Create("users", false, data)
			},
			sectionName:    "users",
			isStrictPath:   false,
			resourcePath:   "/users/different",
			id:             "123",
			expectConflict: true,
			description:    "With strict_path=false, same section and ID should conflict across paths",
		},
		{
			name: "strict_path=false: different sections, same ID - no conflict",
			setup: func(store storage.UniStorage) {
				data := model.UniData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				store.Create("users", false, data)
			},
			sectionName:    "products",
			isStrictPath:   false,
			resourcePath:   "/products",
			id:             "123",
			expectConflict: false,
			description:    "With strict_path=false, different sections should not conflict even with same ID",
		},
		{
			name: "mixed modes: strict creates resource, non-strict tries same section/ID",
			setup: func(store storage.UniStorage) {
				data := model.UniData{
					Path: "/users/admin",
					IDs:  []string{"456"},
					Body: []byte(`{"id": "456", "name": "admin"}`),
				}
				store.Create("users", true, data)
			},
			sectionName:    "users",
			isStrictPath:   false,
			resourcePath:   "/users/regular",
			id:             "456",
			expectConflict: false,
			description:    "Resources created in strict mode should be isolated from non-strict mode lookups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewUniStorage()
			
			// Setup existing data
			tt.setup(store)
			
			// Try to create a new resource that might conflict
			newData := model.UniData{
				Path: tt.resourcePath,
				IDs:  []string{tt.id},
				Body: []byte(`{"id": "` + tt.id + `", "name": "new resource"}`),
			}
			
			err := store.Create(tt.sectionName, tt.isStrictPath, newData)
			
			if tt.expectConflict {
				assert.Error(t, err, tt.description)
				assert.IsType(t, &errors.ConflictError{}, err, "Should return ConflictError")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}


// TestSectionAwareResourceAccess tests that resources are properly scoped by section/strict_path mode
func TestSectionAwareResourceAccess(t *testing.T) {
	store := storage.NewUniStorage()
	
	// Setup test data
	strictData, nonStrictData := setupSectionAwareTestData(t, store)
	
	// Run test cases
	testCases := getSectionAwareTestCases(strictData, nonStrictData)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			validateSectionAwareAccess(t, store, tt)
		})
	}
}

// setupSectionAwareTestData creates test resources in different scopes
func setupSectionAwareTestData(t *testing.T, store storage.UniStorage) (strictData, nonStrictData model.UniData) {
	t.Helper()
	
	// Create resources in different scopes with different IDs
	strictData = model.UniData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "strict admin"}`),
	}
	err := store.Create("users", true, strictData)
	require.NoError(t, err)
	
	nonStrictData = model.UniData{
		Path: "/users/regular",
		IDs:  []string{"456"}, // Different ID to avoid conflict
		Body: []byte(`{"id": "456", "name": "non-strict user"}`),
	}
	err = store.Create("users", false, nonStrictData)
	require.NoError(t, err)
	
	return strictData, nonStrictData
}

// sectionAwareTestCase defines a test case for section-aware access
type sectionAwareTestCase struct {
	name         string
	sectionName  string
	isStrictPath bool
	id           string
	expectedData model.UniData
	shouldFind   bool
	description  string
}

// getSectionAwareTestCases returns test cases for section-aware access
func getSectionAwareTestCases(strictData, nonStrictData model.UniData) []sectionAwareTestCase {
	return []sectionAwareTestCase{
		{
			name:         "strict mode finds strict resource",
			sectionName:  "users",
			isStrictPath: true,
			id:           "123",
			expectedData: strictData,
			shouldFind:   true,
			description:  "Strict mode should find resource created in strict mode",
		},
		{
			name:         "non-strict mode finds non-strict resource",
			sectionName:  "users",
			isStrictPath: false,
			id:           "456",
			expectedData: nonStrictData,
			shouldFind:   true,
			description:  "Non-strict mode should find resource created in non-strict mode",
		},
		{
			name:         "strict mode cannot access non-strict resource",
			sectionName:  "users",
			isStrictPath: true,
			id:           "456", // This ID exists in non-strict mode but not strict
			shouldFind:   false,
			description:  "Strict mode should not find resources from non-strict mode",
		},
		{
			name:         "non-strict mode cannot access strict resource",
			sectionName:  "users",
			isStrictPath: false,
			id:           "123", // This ID exists in strict mode but not non-strict
			shouldFind:   false,
			description:  "Non-strict mode should not find resources from strict mode",
		},
	}
}

// validateSectionAwareAccess validates a single section-aware access test case
func validateSectionAwareAccess(t *testing.T, store storage.UniStorage, tt sectionAwareTestCase) {
	t.Helper()
	
	result, err := store.Get(tt.sectionName, tt.isStrictPath, tt.id)
	
	if tt.shouldFind {
		assert.NoError(t, err, tt.description)
		assert.NotEmpty(t, result.Path, tt.description)
		if tt.expectedData.Path != "" {
			assert.Equal(t, tt.expectedData.Path, result.Path)
			assert.Equal(t, tt.expectedData.IDs, result.IDs)
		}
	} else {
		assert.Error(t, err, tt.description)
		// Since result is now a value type, check for zero value instead of nil
		assert.Empty(t, result.Path, tt.description)
	}
}

// TestResourceUpdateConflictDetection tests conflict detection during updates
func TestResourceUpdateConflictDetection(t *testing.T) {
	store := storage.NewUniStorage()
	
	// Create initial resource
	initialData := model.UniData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "original"}`),
	}
	err := store.Create("users", true, initialData)
	require.NoError(t, err)
	
	// Test updating existing resource
	updatedData := model.UniData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "updated"}`),
	}
	
	err = store.Update("users", true, "123", updatedData)
	assert.NoError(t, err, "Should be able to update existing resource")
	
	// Verify update
	result, err := store.Get("users", true, "123")
	require.NoError(t, err)
	assert.Contains(t, string(result.Body), "updated")
	
	// Test updating non-existent resource
	nonExistentData := model.UniData{
		Path: "/users/admin",
		IDs:  []string{"999"},
		Body: []byte(`{"id": "999", "name": "should not work"}`),
	}
	
	err = store.Update("users", true, "999", nonExistentData)
	assert.Error(t, err, "Should not be able to update non-existent resource")
	assert.IsType(t, &errors.NotFoundError{}, err)
}

// TestResourceDeletionWithScoping tests deletion works correctly with section scoping
func TestResourceDeletionWithScoping(t *testing.T) {
	store := storage.NewUniStorage()
	
	// Create resources in different scopes with different IDs
	strictData := model.UniData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "strict admin"}`),
	}
	err := store.Create("users", true, strictData)
	require.NoError(t, err)
	
	nonStrictData := model.UniData{
		Path: "/users/regular",
		IDs:  []string{"456"}, // Different ID to avoid conflict
		Body: []byte(`{"id": "456", "name": "non-strict user"}`),
	}
	err = store.Create("users", false, nonStrictData)
	require.NoError(t, err)
	
	// Delete strict resource
	err = store.Delete("users", true, "123")
	assert.NoError(t, err, "Should be able to delete strict resource")
	
	// Verify strict resource is gone
	_, err = store.Get("users", true, "123")
	assert.Error(t, err, "Strict resource should be deleted")
	
	// Verify non-strict resource still exists
	result, err := store.Get("users", false, "456")
	assert.NoError(t, err, "Non-strict resource should still exist")
	assert.Contains(t, string(result.Body), "non-strict user")
	
	// Delete non-strict resource
	err = store.Delete("users", false, "456")
	assert.NoError(t, err, "Should be able to delete non-strict resource")
	
	// Verify both resources are gone
	_, err = store.Get("users", false, "456")
	assert.Error(t, err, "Non-strict resource should be deleted")
}

// TestNonStrictModeIDUniquenessPerSection tests that IDs must be unique per section in non-strict mode
func TestNonStrictModeIDUniquenessPerSection(t *testing.T) {
	store := storage.NewUniStorage()
	
	// Create first resource in non-strict mode
	firstData := model.UniData{
		Path: "/users/regular",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "first user"}`),
	}
	err := store.Create("users", false, firstData)
	require.NoError(t, err)
	
	// Try to create second resource with same ID in same section
	secondData := model.UniData{
		Path: "/users/admin", // Different path but same section and ID
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "second user"}`),
	}
	err = store.Create("users", false, secondData)
	assert.Error(t, err, "Should not allow duplicate ID in same section for non-strict mode")
	assert.IsType(t, &errors.ConflictError{}, err)
}