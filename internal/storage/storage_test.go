package storage

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bmcszk/unimock/internal/model"
)

func TestStorage_StoreAndGet(t *testing.T) {
	storage := NewStorage()

	data := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	// Store data
	err := storage.Create([]string{"123"}, data)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Get data
	retrieved, err := storage.Get("123")
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if retrieved.Path != data.Path {
		t.Errorf("Expected path %s, got %s", data.Path, retrieved.Path)
	}
}

func TestStorage_GetByPath(t *testing.T) {
	storage := NewStorage()

	data1 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	data2 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "456"}`),
	}

	// Store data
	err := storage.Create([]string{"123"}, data1)
	if err != nil {
		t.Fatalf("Failed to store data1: %v", err)
	}

	err = storage.Create([]string{"456"}, data2)
	if err != nil {
		t.Fatalf("Failed to store data2: %v", err)
	}

	// Test exact path
	items, err := storage.GetByPath("/test")
	if err != nil {
		t.Fatalf("Failed to get by path: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}
}

func TestStorage_Update(t *testing.T) {
	storage := NewStorage()

	// Create initial data
	data := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	err := storage.Create([]string{"123"}, data)
	if err != nil {
		t.Fatalf("Failed to create data: %v", err)
	}

	// Update data
	updatedData := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "updated": true}`),
	}

	err = storage.Update("123", updatedData)
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	// Verify update
	retrieved, err := storage.Get("123")
	if err != nil {
		t.Fatalf("Failed to get updated data: %v", err)
	}

	if string(retrieved.Body) != string(updatedData.Body) {
		t.Errorf("Expected body %s, got %s", string(updatedData.Body), string(retrieved.Body))
	}

	// Test updating non-existent ID
	err = storage.Update("nonexistent", updatedData)
	if err == nil {
		t.Error("Expected error when updating non-existent ID")
	}
}

func TestStorage_ForEach(t *testing.T) {
	storage := NewStorage()

	data1 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	data2 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "456"}`),
	}

	// Store data
	err := storage.Create([]string{"123"}, data1)
	if err != nil {
		t.Fatalf("Failed to store data1: %v", err)
	}

	err = storage.Create([]string{"456"}, data2)
	if err != nil {
		t.Fatalf("Failed to store data2: %v", err)
	}

	// Test ForEach
	count := 0
	err = storage.ForEach(func(id string, data *model.MockData) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ForEach failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 items, got %d", count)
	}
}

func TestStorage_GetByPathMultipleElements(t *testing.T) {
	storage := NewStorage()

	// Create test data
	data1 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	data2 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "456"}`),
	}

	// Store data
	err := storage.Create([]string{"123"}, data1)
	if err != nil {
		t.Fatalf("Failed to store data1: %v", err)
	}

	err = storage.Create([]string{"456"}, data2)
	if err != nil {
		t.Fatalf("Failed to store data2: %v", err)
	}

	// Test collection path
	items, err := storage.GetByPath("/test")
	if err != nil {
		t.Fatalf("Failed to get by path: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}
}

func TestStorage_ConcurrentAccess(t *testing.T) {
	storage := NewStorage()
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

func TestStorage_DeleteOneOfMultipleElements(t *testing.T) {
	storage := NewStorage()

	// Create two elements with the same path
	data1 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test1"}`),
	}

	data2 := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "456", "name": "test2"}`),
	}

	// Store both elements
	err := storage.Create([]string{"123"}, data1)
	if err != nil {
		t.Fatalf("Failed to store data1: %v", err)
	}

	err = storage.Create([]string{"456"}, data2)
	if err != nil {
		t.Fatalf("Failed to store data2: %v", err)
	}

	// Verify both elements are stored
	items, err := storage.GetByPath("/test")
	if err != nil {
		t.Fatalf("Failed to get by path: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	// Delete one element
	err = storage.Delete("123")
	if err != nil {
		t.Fatalf("Failed to delete data1: %v", err)
	}

	// Verify the other element is still accessible by path
	items, err = storage.GetByPath("/test")
	if err != nil {
		t.Fatalf("Failed to get by path after deletion: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item after deletion, got %d", len(items))
	}
	if string(items[0].Body) != string(data2.Body) {
		t.Errorf("Expected remaining item to be data2, got %s", string(items[0].Body))
	}

	// Verify the deleted element is not accessible by ID
	_, err = storage.Get("123")
	if err == nil {
		t.Error("Expected error when getting deleted item by ID")
	}

	// Verify the remaining element is still accessible by ID
	remaining, err := storage.Get("456")
	if err != nil {
		t.Fatalf("Failed to get remaining item by ID: %v", err)
	}
	if string(remaining.Body) != string(data2.Body) {
		t.Errorf("Expected remaining item to be data2, got %s", string(remaining.Body))
	}
}

func TestStorage_JsonWithIdInBody(t *testing.T) {
	storage := NewStorage()

	// Create JSON data with ID in body
	data := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test", "value": 42}`),
	}

	// Store data
	err := storage.Create([]string{"123"}, data)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Test retrieval by ID
	retrieved, err := storage.Get("123")
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if string(retrieved.Body) != string(data.Body) {
		t.Errorf("Expected body %s, got %s", string(data.Body), string(retrieved.Body))
	}
	if retrieved.Path != data.Path {
		t.Errorf("Expected path %s, got %s", data.Path, retrieved.Path)
	}
	if retrieved.ContentType != data.ContentType {
		t.Errorf("Expected content type %s, got %s", data.ContentType, retrieved.ContentType)
	}

	// Test retrieval by collection path
	items, err := storage.GetByPath("/test")
	if err != nil {
		t.Fatalf("Failed to get by collection path: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item in collection, got %d", len(items))
	}
	if string(items[0].Body) != string(data.Body) {
		t.Errorf("Expected body %s, got %s", string(data.Body), string(items[0].Body))
	}
	if items[0].Path != data.Path {
		t.Errorf("Expected path %s, got %s", data.Path, items[0].Path)
	}
	if items[0].ContentType != data.ContentType {
		t.Errorf("Expected content type %s, got %s", data.ContentType, items[0].ContentType)
	}
}
