package storage_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestMockStorage_CRUD(t *testing.T) { //nolint:revive
	testStorage := storage.NewMockStorage()

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
		// Set IDs in MockData
		data.IDs = []string{id}
		err := testStorage.Create(data)
		if err != nil {
			t.Errorf("Failed to store data: %v", err)
		}
	}

	// Test Get
	for _, data := range testData {
		id := data.Path[strings.LastIndex(data.Path, "/")+1:]
		retrieved, err := testStorage.Get(id)
		if err != nil {
			t.Errorf("Failed to get data: %v", err)
		}
		if string(retrieved.Body) != string(data.Body) {
			t.Errorf("Expected body %s, got %s", string(data.Body), string(retrieved.Body))
		}
	}

	// Test GetByPath
	items, err := testStorage.GetByPath("/test")
	if err != nil {
		t.Fatalf("Failed to get by path: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	// Test GetByPath with trailing slash
	items, err = testStorage.GetByPath("/test/")
	if err != nil {
		t.Fatalf("Failed to get by path with trailing slash: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	// Test GetByPath with case sensitivity
	_, err = testStorage.GetByPath("/Test")
	if err == nil {
		t.Error("Expected error for case-sensitive path mismatch, got nil")
	} else if !strings.Contains(err.Error(), "resource not found") {
		t.Errorf("Expected 'resource not found' error, got %v", err)
	}

	// Test GetByPath with empty collection
	_, err = testStorage.GetByPath("/empty")
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
	err = testStorage.Update("123", updatedData)
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	// Test Delete
	err = testStorage.Delete("123")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// Verify deletion
	_, err = testStorage.Get("123")
	if err == nil {
		t.Error("Expected error when getting deleted data")
	}
}

// Helper function to create concurrent test data
func createConcurrentTestData(t *testing.T, testStorage storage.MockStorage, i int) {
	t.Helper()
	id := fmt.Sprintf("test%d", i)
	data := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(fmt.Sprintf(`{"id": "%d"}`, i)),
		IDs:         []string{id},
	}
	err := testStorage.Create(data)
	if err != nil {
		t.Errorf("Failed to store data: %v", err)
	}
}

// Helper function to verify concurrent test data
func verifyConcurrentTestData(t *testing.T, testStorage storage.MockStorage) {
	t.Helper()
	// Verify all data was stored
	for i := 0; i < 10; i++ { //nolint:mnd
		id := fmt.Sprintf("test%d", i)
		_, err := testStorage.Get(id)
		if err != nil {
			t.Errorf("Failed to get data for %s: %v", id, err)
		}
	}
}

func TestMockStorage_ConcurrentAccess(t *testing.T) {
	testStorage := storage.NewMockStorage()
	var wg sync.WaitGroup

	// Test concurrent writes
	for i := 0; i < 10; i++ { //nolint:mnd
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			createConcurrentTestData(t, testStorage, i)
		}(i)
	}

	wg.Wait()

	verifyConcurrentTestData(t, testStorage)
}

func TestMockStorage_ErrorCases(t *testing.T) {
	testStorage := storage.NewMockStorage()

	// Test updating non-existent ID
	err := testStorage.Update("nonexistent", &model.MockData{})
	if err == nil {
		t.Error("Expected error when updating non-existent ID")
	}

	// Test deleting non-existent ID
	err = testStorage.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent ID")
	}

	// Test getting non-existent ID
	_, err = testStorage.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent ID")
	}

	// Test creating duplicate ID
	data := &model.MockData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"id": "123"}`),
		IDs:         []string{"123"},
	}
	err = testStorage.Create(data)
	if err != nil {
		t.Fatalf("Failed to create initial data: %v", err)
	}
	err = testStorage.Create(data)
	if err == nil {
		t.Error("Expected error when creating duplicate ID")
	}

	// Test malformed paths
	_, err = testStorage.GetByPath("")
	if err == nil {
		t.Error("Expected error when getting with empty path")
	}

	_, err = testStorage.GetByPath("invalid/path/")
	if err == nil {
		t.Error("Expected error when getting with invalid path")
	}
}

// Helper function to test create with multiple IDs
func testCreateWithMultipleIDs(t *testing.T, storageInstance storage.MockStorage) {
	t.Helper()
	multiIdData1 := &model.MockData{
		Path:        "/multi/data1",
		ContentType: "text/plain",
		Body:        []byte("data for multi-id 1"),
	}
	externalIDs1 := []string{"id1_main", "id1_alias1", "id1_alias2"}
	multiIdData1.IDs = externalIDs1
	err := storageInstance.Create(multiIdData1)
	if err != nil {
		t.Fatalf("CreateWithMultipleIDs: failed to create resource: %v", err)
	}

	for _, id := range externalIDs1 {
		retrieved, errGet := storageInstance.Get(id)
		if errGet != nil {
			t.Errorf("CreateWithMultipleIDs: Get(%s) failed: %v", id, errGet)
		}
		if retrieved == nil || string(retrieved.Body) != string(multiIdData1.Body) {
			t.Errorf(
				"CreateWithMultipleIDs: Get(%s) returned incorrect data. Got %v, Expected %v",
				id, retrieved, multiIdData1,
			)
		}
	}
}

// Helper function to test delete by one ID removes all mappings
func testDeleteByOneIDRemovesAllMappings(t *testing.T, storageInstance storage.MockStorage) {
	t.Helper()
	multiIdData2 := &model.MockData{
		Path:        "/multi/data2",
		ContentType: "text/plain",
		Body:        []byte("data for multi-id 2"),
	}
	externalIDs2 := []string{"id2_main", "id2_alias1"}
	multiIdData2.IDs = externalIDs2
	err := storageInstance.Create(multiIdData2)
	if err != nil {
		t.Fatalf("DeleteByOneID: failed to create initial resource: %v", err)
	}

	err = storageInstance.Delete(externalIDs2[0]) // Delete by "id2_main"
	if err != nil {
		t.Fatalf("DeleteByOneID: Delete(%s) failed: %v", externalIDs2[0], err)
	}

	for _, id := range externalIDs2 {
		_, errGet := storageInstance.Get(id)
		if errGet == nil {
			t.Errorf("DeleteByOneID: Get(%s) should have failed after delete, but succeeded.", id)
		}
	}
}

// Helper function to test update by one ID affects single resource
func testUpdateByOneIDAffectsSingleResource(t *testing.T, storageInstance storage.MockStorage) { //nolint:revive
	t.Helper()
	multiIdData3 := &model.MockData{
		Path:        "/multi/data3",
		ContentType: "text/plain",
		Body:        []byte("original data3"),
	}
	externalIDs3 := []string{"id3_main", "id3_alias1"}
	multiIdData3.IDs = externalIDs3
	err := storageInstance.Create(multiIdData3)
	if err != nil {
		t.Fatalf("UpdateByOneID: failed to create initial resource: %v", err)
	}

	updatedData3Body := "updated data3"
	updatePayload := &model.MockData{
		Path:        "/multi/data3",
		ContentType: "text/plain",
		Body:        []byte(updatedData3Body),
	}
	err = storageInstance.Update(externalIDs3[1], updatePayload) // Update using "id3_alias1"
	if err != nil {
		t.Fatalf("UpdateByOneID: Update(%s) failed: %v", externalIDs3[1], err)
	}

	for _, id := range externalIDs3 {
		retrieved, errGet := storageInstance.Get(id)
		if errGet != nil {
			t.Errorf("UpdateByOneID: Get(%s) failed after update: %v", id, errGet)
		}
		if retrieved == nil || string(retrieved.Body) != updatedData3Body {
			t.Errorf(
				"UpdateByOneID: Get(%s) returned incorrect data after update. Got %s, Expected %s",
				id, string(retrieved.Body), updatedData3Body,
			)
		}
	}
}

// Helper function to test conflict on create with existing external ID
func testConflictOnCreateWithExistingExternalID(t *testing.T, storageInstance storage.MockStorage) {
	t.Helper()
	conflictData1 := &model.MockData{
		Path:        "/conflict/data1",
		ContentType: "text/plain",
		Body:        []byte("conflict data 1"),
	}
	externalIDsConflict1 := []string{"common_id", "unique_c1"}
	conflictData1.IDs = externalIDsConflict1
	err := storageInstance.Create(conflictData1)
	if err != nil {
		t.Fatalf("ConflictOnCreate: failed to create initial resource for conflict test: %v", err)
	}

	conflictData2 := &model.MockData{
		Path:        "/conflict/data2",
		ContentType: "text/plain",
		Body:        []byte("conflict data 2"),
	}
	externalIDsConflict2 := []string{"common_id", "unique_c2"} // Attempt to reuse "common_id"
	conflictData2.IDs = externalIDsConflict2
	err = storageInstance.Create(conflictData2)
	if err == nil {
		t.Errorf("ConflictOnCreate: expected conflict error when creating with duplicate external ID, but got nil")
	} else {
		// Check for specific conflict error type if desired, e.g., errors.IsConflict(err)
		t.Logf("ConflictOnCreate: correctly received error for duplicate ID: %v", err)
	}

	// Verify that the original resource with "common_id" is still accessible and unchanged
	retrievedConflict, errGet := storageInstance.Get("common_id")
	if errGet != nil {
		t.Errorf("ConflictOnCreate: Get(\"common_id\") failed after conflict attempt: %v", errGet)
	}
	if retrievedConflict == nil || string(retrievedConflict.Body) != string(conflictData1.Body) {
		t.Errorf(
			"ConflictOnCreate: Get(\"common_id\") returned incorrect data. Got %v, Expected %v",
			retrievedConflict, conflictData1,
		)
	}
	// Verify the second resource (that caused conflict) was not created with its unique ID either
	_, errGet = storageInstance.Get("unique_c2")
	if errGet == nil {
		t.Errorf("ConflictOnCreate: Get(\"unique_c2\") should have failed as creation was aborted, but succeeded.")
	}
}

func TestMockStorage_MultiID(t *testing.T) {
	storageInstance := storage.NewMockStorage() // Use public API

	// 1. CreateWithMultipleIDs & Get by any ID
	testCreateWithMultipleIDs(t, storageInstance)

	// 2. DeleteByOneID_RemovesAllMappings
	testDeleteByOneIDRemovesAllMappings(t, storageInstance)

	// 3. UpdateByOneID_AffectsSingleResource
	testUpdateByOneIDAffectsSingleResource(t, storageInstance)

	// 4. ConflictOnCreateWithExistingExternalID
	testConflictOnCreateWithExistingExternalID(t, storageInstance)
}
