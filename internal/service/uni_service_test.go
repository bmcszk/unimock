package service_test

import (
	"context"
	"testing"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUniService_CreateResource(t *testing.T) {
	// Setup
	store := storage.NewUniStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
				BodyIDPaths: []string{"/id"},
			},
		},
	}
	svc := service.NewUniService(store, cfg)

	// Test data
	data := model.UniData{
		Path:        "/users/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test"}`),
	}

	// Test creating a resource
	err := svc.CreateResource(context.Background(), "users", false, []string{"123"}, data)
	assert.NoError(t, err)

	// Verify resource was created
	result, err := svc.GetResource(context.Background(), "users", false, "123")
	assert.NoError(t, err)
	assert.Equal(t, data.IDs, result.IDs)
	assert.Equal(t, data.Body, result.Body)
}

func TestUniService_GetResource(t *testing.T) {
	// Setup
	store := storage.NewUniStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
			},
		},
	}
	svc := service.NewUniService(store, cfg)

	// Test data
	data := model.UniData{
		Path:        "/users/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test"}`),
	}

	// Create resource first
	err := svc.CreateResource(context.Background(), "users", false, []string{"123"}, data)
	require.NoError(t, err)

	// Test getting the resource
	result, err := svc.GetResource(context.Background(), "users", false, "123")
	assert.NoError(t, err)
	assert.Equal(t, data.IDs, result.IDs)
	assert.Equal(t, data.ContentType, result.ContentType)
	assert.Equal(t, data.Body, result.Body)
}

func TestUniService_GetResource_NotFound(t *testing.T) {
	// Setup
	store := storage.NewUniStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
			},
		},
	}
	svc := service.NewUniService(store, cfg)

	// Test getting non-existent resource
	_, err := svc.GetResource(context.Background(), "users", false, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUniService_UpdateResource(t *testing.T) {
	// Setup
	store := storage.NewUniStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
			},
		},
	}
	svc := service.NewUniService(store, cfg)

	// Test data
	originalData := model.UniData{
		Path:        "/users/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test"}`),
	}

	updatedData := model.UniData{
		Path:        "/users/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "updated"}`),
	}

	// Create resource first
	err := svc.CreateResource(context.Background(), "users", false, []string{"123"}, originalData)
	require.NoError(t, err)

	// Update the resource
	err = svc.UpdateResource(context.Background(), "users", false, "123", updatedData)
	assert.NoError(t, err)

	// Verify resource was updated
	result, err := svc.GetResource(context.Background(), "users", false, "123")
	assert.NoError(t, err)
	assert.Equal(t, updatedData.ContentType, result.ContentType)
	assert.Equal(t, updatedData.Body, result.Body)
}

func TestUniService_DeleteResource(t *testing.T) {
	// Setup
	store := storage.NewUniStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
			},
		},
	}
	svc := service.NewUniService(store, cfg)

	// Test data
	data := model.UniData{
		Path:        "/users/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test"}`),
	}

	// Create resource first
	err := svc.CreateResource(context.Background(), "users", false, []string{"123"}, data)
	require.NoError(t, err)

	// Delete the resource
	err = svc.DeleteResource(context.Background(), "users", false, "123")
	assert.NoError(t, err)

	// Verify resource was deleted
	_, err = svc.GetResource(context.Background(), "users", false, "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUniService_GetResourcesByPath(t *testing.T) {
	// Setup
	store := storage.NewUniStorage()
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/users/*",
			},
		},
	}
	svc := service.NewUniService(store, cfg)

	// Create multiple resources
	data1 := model.UniData{
		Path:        "/users/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "user1"}`),
	}
	data2 := model.UniData{
		Path:        "/users/456",
		IDs:         []string{"456"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "456", "name": "user2"}`),
	}

	err := svc.CreateResource(context.Background(), "users", false, []string{"123"}, data1)
	require.NoError(t, err)
	err = svc.CreateResource(context.Background(), "users", false, []string{"456"}, data2)
	require.NoError(t, err)

	// Get resources by path
	resources, err := svc.GetResourcesByPath(context.Background(), "/users")
	assert.NoError(t, err)
	assert.Len(t, resources, 2)

	// Verify both resources are returned
	paths := make([]string, len(resources))
	for i, resource := range resources {
		paths[i] = resource.Path
	}
	assert.Contains(t, paths, "/users/123")
	assert.Contains(t, paths, "/users/456")
}