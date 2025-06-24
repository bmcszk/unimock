//go:build !e2e

package storage

import (
	"testing"

	"github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompositeKeyConflictDetection tests the new composite key-based conflict detection
func TestCompositeKeyConflictDetection(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(storage MockStorage)
		sectionName    string
		isStrictPath   bool
		resourcePath   string
		id             string
		expectConflict bool
		description    string
	}{
		{
			name: "strict_path=true: different paths, same ID - no conflict",
			setup: func(storage MockStorage) {
				data := &model.MockData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				storage.Create("users", true, data)
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
			setup: func(storage MockStorage) {
				data := &model.MockData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				storage.Create("users", true, data)
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
			setup: func(storage MockStorage) {
				data := &model.MockData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				storage.Create("users", false, data)
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
			setup: func(storage MockStorage) {
				data := &model.MockData{
					Path: "/users/subpath",
					IDs:  []string{"123"},
					Body: []byte(`{"id": "123", "name": "test"}`),
				}
				storage.Create("users", false, data)
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
			setup: func(storage MockStorage) {
				data := &model.MockData{
					Path: "/users/admin",
					IDs:  []string{"456"},
					Body: []byte(`{"id": "456", "name": "admin"}`),
				}
				storage.Create("users", true, data)
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
			storage := NewMockStorage()
			
			// Setup existing data
			tt.setup(storage)
			
			// Try to create a new resource that might conflict
			newData := &model.MockData{
				Path: tt.resourcePath,
				IDs:  []string{tt.id},
				Body: []byte(`{"id": "` + tt.id + `", "name": "new resource"}`),
			}
			
			err := storage.Create(tt.sectionName, tt.isStrictPath, newData)
			
			if tt.expectConflict {
				assert.Error(t, err, tt.description)
				assert.IsType(t, &errors.ConflictError{}, err, "Should return ConflictError")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestCompositeKeyGeneration tests the composite key generation logic
func TestCompositeKeyGeneration(t *testing.T) {
	storage := &mockStorage{}
	
	tests := []struct {
		name         string
		sectionName  string
		isStrictPath bool
		resourcePath string
		id           string
		expectedKey  string
	}{
		{
			name:         "strict mode uses path:id format",
			sectionName:  "users",
			isStrictPath: true,
			resourcePath: "/users/subpath",
			id:           "123",
			expectedKey:  "/users/subpath:123",
		},
		{
			name:         "non-strict mode uses section:id format",
			sectionName:  "users",
			isStrictPath: false,
			resourcePath: "/users/subpath",
			id:           "123",
			expectedKey:  "users:123",
		},
		{
			name:         "strict mode with root path",
			sectionName:  "products",
			isStrictPath: true,
			resourcePath: "/products",
			id:           "abc",
			expectedKey:  "/products:abc",
		},
		{
			name:         "non-strict mode with complex path",
			sectionName:  "orders",
			isStrictPath: false,
			resourcePath: "/api/v1/orders/pending",
			id:           "order-456",
			expectedKey:  "orders:order-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := storage.buildCompositeKey(tt.sectionName, tt.isStrictPath, tt.resourcePath, tt.id)
			assert.Equal(t, tt.expectedKey, key)
		})
	}
}

// TestSectionAwareResourceAccess tests that resources are properly scoped by section/strict_path mode
func TestSectionAwareResourceAccess(t *testing.T) {
	storage := NewMockStorage()
	
	// Create resources in different scopes with different IDs
	strictData := &model.MockData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "strict admin"}`),
	}
	err := storage.Create("users", true, strictData)
	require.NoError(t, err)
	
	nonStrictData := &model.MockData{
		Path: "/users/regular",
		IDs:  []string{"456"}, // Different ID to avoid conflict
		Body: []byte(`{"id": "456", "name": "non-strict user"}`),
	}
	err = storage.Create("users", false, nonStrictData)
	require.NoError(t, err)
	
	// Try to access resources
	tests := []struct {
		name             string
		sectionName      string
		isStrictPath     bool
		id               string
		expectedData     *model.MockData
		shouldFind       bool
		description      string
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := storage.Get(tt.sectionName, tt.isStrictPath, tt.id)
			
			if tt.shouldFind {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, result, tt.description)
				if tt.expectedData != nil {
					assert.Equal(t, tt.expectedData.Path, result.Path)
					assert.Equal(t, tt.expectedData.IDs, result.IDs)
				}
			} else {
				assert.Error(t, err, tt.description)
				assert.Nil(t, result, tt.description)
			}
		})
	}
}

// TestResourceUpdateConflictDetection tests conflict detection during updates
func TestResourceUpdateConflictDetection(t *testing.T) {
	storage := NewMockStorage()
	
	// Create initial resource
	initialData := &model.MockData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "original"}`),
	}
	err := storage.Create("users", true, initialData)
	require.NoError(t, err)
	
	// Test updating existing resource
	updatedData := &model.MockData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "updated"}`),
	}
	
	err = storage.Update("users", true, "123", updatedData)
	assert.NoError(t, err, "Should be able to update existing resource")
	
	// Verify update
	result, err := storage.Get("users", true, "123")
	require.NoError(t, err)
	assert.Contains(t, string(result.Body), "updated")
	
	// Test updating non-existent resource
	nonExistentData := &model.MockData{
		Path: "/users/admin",
		IDs:  []string{"999"},
		Body: []byte(`{"id": "999", "name": "should not work"}`),
	}
	
	err = storage.Update("users", true, "999", nonExistentData)
	assert.Error(t, err, "Should not be able to update non-existent resource")
	assert.IsType(t, &errors.NotFoundError{}, err)
}

// TestResourceDeletionWithScoping tests deletion works correctly with section scoping
func TestResourceDeletionWithScoping(t *testing.T) {
	storage := NewMockStorage()
	
	// Create resources in different scopes with different IDs
	strictData := &model.MockData{
		Path: "/users/admin",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "strict admin"}`),
	}
	err := storage.Create("users", true, strictData)
	require.NoError(t, err)
	
	nonStrictData := &model.MockData{
		Path: "/users/regular",
		IDs:  []string{"456"}, // Different ID to avoid conflict
		Body: []byte(`{"id": "456", "name": "non-strict user"}`),
	}
	err = storage.Create("users", false, nonStrictData)
	require.NoError(t, err)
	
	// Delete strict resource
	err = storage.Delete("users", true, "123")
	assert.NoError(t, err, "Should be able to delete strict resource")
	
	// Verify strict resource is gone
	_, err = storage.Get("users", true, "123")
	assert.Error(t, err, "Strict resource should be deleted")
	
	// Verify non-strict resource still exists
	result, err := storage.Get("users", false, "456")
	assert.NoError(t, err, "Non-strict resource should still exist")
	assert.Contains(t, string(result.Body), "non-strict user")
	
	// Delete non-strict resource
	err = storage.Delete("users", false, "456")
	assert.NoError(t, err, "Should be able to delete non-strict resource")
	
	// Verify both resources are gone
	_, err = storage.Get("users", false, "456")
	assert.Error(t, err, "Non-strict resource should be deleted")
}

// TestNonStrictModeIDUniquenessPerSection tests that IDs must be unique per section in non-strict mode
func TestNonStrictModeIDUniquenessPerSection(t *testing.T) {
	storage := NewMockStorage()
	
	// Create first resource in non-strict mode
	firstData := &model.MockData{
		Path: "/users/regular",
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "first user"}`),
	}
	err := storage.Create("users", false, firstData)
	require.NoError(t, err)
	
	// Try to create second resource with same ID in same section
	secondData := &model.MockData{
		Path: "/users/admin", // Different path but same section and ID
		IDs:  []string{"123"},
		Body: []byte(`{"id": "123", "name": "second user"}`),
	}
	err = storage.Create("users", false, secondData)
	assert.Error(t, err, "Should not allow duplicate ID in same section for non-strict mode")
	assert.IsType(t, &errors.ConflictError{}, err)
}