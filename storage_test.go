package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestStorage_StoreAndGet(t *testing.T) {
	storage := NewStorage()

	data := &MockData{
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

	data1 := &MockData{
		Path:        "/test/123",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	data2 := &MockData{
		Path:        "/test/456",
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
	items, err := storage.GetByPath("/test/123")
	if err != nil {
		t.Fatalf("Failed to get by path: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Path != data1.Path {
		t.Errorf("Expected path %s, got %s", data1.Path, items[0].Path)
	}

	// Test collection path
	items, err = storage.GetByPath("/test")
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
	data := &MockData{
		Path:        "/test/123",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	err := storage.Create([]string{"123"}, data)
	if err != nil {
		t.Fatalf("Failed to create data: %v", err)
	}

	// Update data
	updatedData := &MockData{
		Path:        "/test/123",
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

	data1 := &MockData{
		Path:        "/test/123",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	data2 := &MockData{
		Path:        "/test/456",
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
	err = storage.ForEach(func(id string, data *MockData) error {
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
	data1 := &MockData{
		Path:        "/test/123",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
	}

	data2 := &MockData{
		Path:        "/test/456",
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

	// Test specific path
	items, err = storage.GetByPath("/test/123")
	if err != nil {
		t.Fatalf("Failed to get by path: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Path != data1.Path {
		t.Errorf("Expected path %s, got %s", data1.Path, items[0].Path)
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
			data := &MockData{
				Path:        fmt.Sprintf("/test/%d", i),
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
