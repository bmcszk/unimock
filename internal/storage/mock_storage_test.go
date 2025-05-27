package storage

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/bmcszk/unimock/pkg/model"
)

func TestMockStorage_CRUD(t *testing.T) {
	storage := NewMockStorage()

	// Test data with mixed content types
	testData := []*model.MockData{
		{
			Path:        "/test/123",
			ContentType: "application/json",
			Body:        []byte(`{"id": "123"}`),
		},
		{
			Path:        "/test/456",
			ContentType: "application/xml",
			Body:        []byte(`<data><id>456</id></data>`),
		},
		{
			Path:        "/test/789",
			ContentType: "application/octet-stream",
			Body:        []byte("binary data"),
		},
	}

	// Test Create
	for _, data := range testData {
		id := data.Path[strings.LastIndex(data.Path, "/")+1:]
		err := storage.Create([]string{id}, data)
		if err != nil {
			t.Errorf("Failed to store data: %v", err)
		}
	}

	// Test Get
	for _, data := range testData {
		id := data.Path[strings.LastIndex(data.Path, "/")+1:]
		retrieved, err := storage.Get(id)
		if err != nil {
			t.Errorf("Failed to get data: %v", err)
		}
		if string(retrieved.Body) != string(data.Body) {
			t.Errorf("Expected body %s, got %s", string(data.Body), string(retrieved.Body))
		}
	}

	// Test GetByPath
	items, err := storage.GetByPath("/test")
	if err != nil {
		t.Fatalf("Failed to get by path: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	// Test GetByPath with trailing slash
	items, err = storage.GetByPath("/test/")
	if err != nil {
		t.Fatalf("Failed to get by path with trailing slash: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	// Test GetByPath with case sensitivity
	_, err = storage.GetByPath("/Test")
	if err == nil {
		t.Error("Expected error for case-sensitive path mismatch, got nil")
	} else if !strings.Contains(err.Error(), "resource not found") {
		t.Errorf("Expected 'resource not found' error, got %v", err)
	}

	// Test GetByPath with empty collection
	_, err = storage.GetByPath("/empty")
	if err == nil {
		t.Error("Expected error for empty collection, got nil")
	} else if !strings.Contains(err.Error(), "resource not found") {
		t.Errorf("Expected 'resource not found' error, got %v", err)
	}

	// Test Update
	updatedData := &model.MockData{
		Path:        "/test/123",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "updated": true}`),
	}
	err = storage.Update("123", updatedData)
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	// Test Delete
	err = storage.Delete("123")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// Verify deletion
	_, err = storage.Get("123")
	if err == nil {
		t.Error("Expected error when getting deleted data")
	}
}

func TestMockStorage_ConcurrentAccess(t *testing.T) {
	storage := NewMockStorage()
	var wg sync.WaitGroup

	// Test concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("test%d", i)
			data := &model.MockData{
				Path:        "/test",
				ContentType: "application/json",
				Body:        []byte(fmt.Sprintf(`{"id": "%d"}`, i)),
			}
			err := storage.Create([]string{id}, data)
			if err != nil {
				t.Errorf("Failed to store data: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all data was stored
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("test%d", i)
		_, err := storage.Get(id)
		if err != nil {
			t.Errorf("Failed to get data for %s: %v", id, err)
		}
	}
}

func TestMockStorage_ErrorCases(t *testing.T) {
	storage := NewMockStorage()

	// Test updating non-existent ID
	err := storage.Update("nonexistent", &model.MockData{})
	if err == nil {
		t.Error("Expected error when updating non-existent ID")
	}

	// Test deleting non-existent ID
	err = storage.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent ID")
	}

	// Test getting non-existent ID
	_, err = storage.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent ID")
	}

	// Test creating duplicate ID
	data := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}
	err = storage.Create([]string{"123"}, data)
	if err != nil {
		t.Fatalf("Failed to create initial data: %v", err)
	}
	err = storage.Create([]string{"123"}, data)
	if err == nil {
		t.Error("Expected error when creating duplicate ID")
	}

	// Test malformed paths
	_, err = storage.GetByPath("")
	if err == nil {
		t.Error("Expected error when getting with empty path")
	}

	_, err = storage.GetByPath("invalid/path/")
	if err == nil {
		t.Error("Expected error when getting with invalid path")
	}
}
