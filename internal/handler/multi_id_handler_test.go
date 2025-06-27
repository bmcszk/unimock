package handler_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func setupMultiIDTestHandler() *handler.UniHandler {
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"products": {
				PathPattern:   "/products/*",
				BodyIDPaths:   []string{"/id", "/details/upc", "/internalCode"},
				HeaderIDName:  "X-Primary-ID",
				CaseSensitive: true,
			},
		},
	}

	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	return handler.NewUniHandler(uniService, scenarioService, logger, cfg)
}

func TestMultiIDPostExtraction(t *testing.T) {
	uniHandler := setupMultiIDTestHandler()

	// Create request with multiple IDs in body and header
	jsonBody := `{"id": "prod123", "details": {"upc": "123456789"}, "internalCode": "INT456"}`
	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Primary-ID", "primary123")

	// Execute request
	w := httptest.NewRecorder()
	uniHandler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify Location header
	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/products/") {
		t.Errorf("expected Location header to start with /products/, got %s", location)
	}
}

func createTestResource(uniHandler *handler.UniHandler) {
	jsonBody := `{"id": "prod456", "details": {"upc": "987654321"}, ` +
		`"internalCode": "INT789", "name": "Test Product"}`
	postReq := httptest.NewRequest("POST", "/products", bytes.NewBufferString(jsonBody))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Primary-ID", "primary456")

	w := httptest.NewRecorder()
	uniHandler.ServeHTTP(w, postReq)
}

func TestMultiIDGetRetrieval(t *testing.T) {
	uniHandler := setupMultiIDTestHandler()
	createTestResource(uniHandler)

	// Test GET with different IDs extracted from body and header
	testIDs := []string{"prod456", "primary456", "987654321", "INT789"}
	
	for _, testID := range testIDs {
		getReq := httptest.NewRequest("GET", "/products/"+testID, nil)
		w := httptest.NewRecorder()
		uniHandler.ServeHTTP(w, getReq)

		if w.Code != http.StatusOK {
			t.Errorf("GET with ID %s failed: status %d", testID, w.Code)
		}
	}
}

func TestMultiIDConflictPrevention(t *testing.T) {
	uniHandler := setupMultiIDTestHandler()

	// Create first resource
	jsonBody1 := `{"id": "conflict_test_1", "details": {"upc": "CONFLICT_UPC"}, "internalCode": "INT001"}`
	postReq1 := httptest.NewRequest("POST", "/products", bytes.NewBufferString(jsonBody1))
	postReq1.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	uniHandler.ServeHTTP(w, postReq1)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create first resource: status %d", w.Code)
	}

	// Try to create second resource with conflicting UPC
	jsonBody2 := `{"id": "conflict_test_2", "details": {"upc": "CONFLICT_UPC"}, "internalCode": "INT002"}`
	postReq2 := httptest.NewRequest("POST", "/products", bytes.NewBufferString(jsonBody2))
	postReq2.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	uniHandler.ServeHTTP(w, postReq2)

	// Should get conflict error
	if w.Code != http.StatusConflict {
		t.Errorf("expected conflict status %d, got %d, body: %s", 
			http.StatusConflict, w.Code, w.Body.String())
	}
}

func TestMultiIDUpdateAndDelete(t *testing.T) {
	uniHandler := setupMultiIDTestHandler()

	// Create a resource
	jsonBody := `{"id": "update_test", "details": {"upc": "UPDATE_UPC"}, "internalCode": "UPDATE_INT"}`
	postReq := httptest.NewRequest("POST", "/products", bytes.NewBufferString(jsonBody))
	postReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	uniHandler.ServeHTTP(w, postReq)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create resource: status %d", w.Code)
	}

	// Update using different ID (UPC)
	updateBody := `{"id": "update_test", "details": {"upc": "UPDATE_UPC"}, ` +
		`"internalCode": "UPDATE_INT", "updated": true}`
	putReq := httptest.NewRequest("PUT", "/products/UPDATE_UPC", bytes.NewBufferString(updateBody))
	putReq.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	uniHandler.ServeHTTP(w, putReq)

	if w.Code != http.StatusOK {
		t.Errorf("PUT failed: status %d, body: %s", w.Code, w.Body.String())
	}

	// Verify update is visible via original ID
	getReq := httptest.NewRequest("GET", "/products/update_test", nil)
	w = httptest.NewRecorder()
	uniHandler.ServeHTTP(w, getReq)

	if w.Code != http.StatusOK {
		t.Errorf("GET after PUT failed: status %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"updated": true`) {
		t.Errorf("Update not visible, body: %s", w.Body.String())
	}

	// Delete using internal code
	delReq := httptest.NewRequest("DELETE", "/products/UPDATE_INT", nil)
	w = httptest.NewRecorder()
	uniHandler.ServeHTTP(w, delReq)

	if w.Code != http.StatusNoContent {
		t.Errorf("DELETE failed: status %d", w.Code)
	}

	// Verify resource is gone via original ID
	getReq = httptest.NewRequest("GET", "/products/update_test", nil)
	w = httptest.NewRecorder()
	uniHandler.ServeHTTP(w, getReq)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET after DELETE should return 404, got %d", w.Code)
	}
}

func TestStorageMultiIDSupport(t *testing.T) {
	store := storage.NewUniStorage()

	// Create resource with multiple IDs
	ids := []string{"id1", "id2", "id3"}
	data := model.UniData{
		Path:        "/test",
		ContentType: "application/json",
		Body:        []byte(`{"test": "data"}`),
		IDs:         ids,
	}

	err := store.Create("test", false, data)
	if err != nil {
		t.Fatalf("failed to create resource with multiple IDs: %v", err)
	}

	// Verify all IDs can retrieve the same resource
	for _, id := range ids {
		retrieved, err := store.Get("test", false, id)
		if err != nil {
			t.Errorf("failed to get resource by ID %s: %v", id, err)
			continue
		}

		if string(retrieved.Body) != string(data.Body) {
			t.Errorf("retrieved data for ID %s doesn't match original", id)
		}
	}
}

func TestStorageIDConflictPrevention(t *testing.T) {
	store := storage.NewUniStorage()

	// Create first resource
	ids1 := []string{"conflict1", "shared"}
	data1 := model.UniData{
		Path:        "/test1",
		ContentType: "application/json",
		Body:        []byte(`{"test": "data1"}`),
	}

	data1.IDs = ids1
	err := store.Create("test", false, data1)
	if err != nil {
		t.Fatalf("failed to create first resource: %v", err)
	}

	// Try to create second resource with conflicting ID
	ids2 := []string{"conflict2", "shared"} // "shared" conflicts
	data2 := model.UniData{
		Path:        "/test2",
		ContentType: "application/json",
		Body:        []byte(`{"test": "data2"}`),
	}

	data2.IDs = ids2
	err = store.Create("test", false, data2)
	if err == nil {
		t.Error("expected conflict error, but creation succeeded")
	}
}