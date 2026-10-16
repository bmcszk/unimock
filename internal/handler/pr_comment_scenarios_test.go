//go:build !e2e

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPRCommentScenario1_FlexiblePathMatching tests the scenario from PR feedback:
// Given: pattern "/users/**", POST /users/subpath body: {"id": 1}, strict_path=false
// When: GET/PUT/DELETE /users/1
// Then: Operations succeed (resource accessible via extracted ID)
func TestPRCommentScenario1_FlexiblePathMatching(t *testing.T) {
	// Setup
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/**",
				StrictPath:    false, // Flexible matching
				BodyIDPaths:   []string{"/id"},
				HeaderIDNames: []string{},
				CaseSensitive: false,
			},
		},
	}

	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	h := handler.NewUniHandler(uniService, scenarioService, logger, cfg)

	// Step 1: POST /users/subpath with body: {"id": 1}
	postBody := `{"id": 1, "name": "test user"}`
	postReq := httptest.NewRequest(http.MethodPost, "/users/subpath", bytes.NewReader([]byte(postBody)))
	postReq.Header.Set("Content-Type", "application/json")

	postResp, err := h.HandlePOST(context.Background(), postReq)
	require.NoError(t, err)
	defer postResp.Body.Close()
	assert.Equal(t, http.StatusCreated, postResp.StatusCode)
	t.Logf("POST response status: %d", postResp.StatusCode)

	// Step 2: GET /users/1 - should succeed with flexible path matching
	getReq := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	getResp, err := h.HandleGET(context.Background(), getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "GET should succeed with flexible path matching")

	// Verify response contains the resource
	var getResponseData map[string]any
	err = json.NewDecoder(getResp.Body).Decode(&getResponseData)
	require.NoError(t, err)
	assert.Equal(t, float64(1), getResponseData["id"], "Resource should be accessible via extracted ID")
	assert.Equal(t, "test user", getResponseData["name"])
	t.Logf("GET response: %+v", getResponseData)

	// Step 3: PUT /users/1 - should succeed (update existing resource)
	putBody := `{"id": 1, "name": "updated user", "status": "active"}`
	putReq := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewReader([]byte(putBody)))
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := h.HandlePUT(context.Background(), putReq)
	require.NoError(t, err)
	defer putResp.Body.Close()
	assert.Equal(t, http.StatusOK, putResp.StatusCode, "PUT should succeed with flexible path matching")
	t.Logf("PUT response status: %d", putResp.StatusCode)

	// Step 4: DELETE /users/1 - should succeed
	deleteReq := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	deleteResp, err := h.HandleDELETE(context.Background(), deleteReq)
	require.NoError(t, err)
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode, "DELETE should succeed with flexible path matching")
	t.Logf("DELETE response status: %d", deleteResp.StatusCode)

	// Step 5: Verify resource is deleted - GET should return 404
	getAfterDeleteReq := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	getAfterDeleteResp, err := h.HandleGET(context.Background(), getAfterDeleteReq)
	require.NoError(t, err)
	defer getAfterDeleteResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getAfterDeleteResp.StatusCode, "Resource should be deleted")
}

// TestPRCommentScenario2_StrictPathMatching tests the scenario from PR feedback:
// Given: pattern "/users/**", POST /users/subpath body: {"id": 1}, strict_path=true
// When: GET/PUT/DELETE /users/1
// Then: 404 Not Found (strict path validation enforced)
func TestPRCommentScenario2_StrictPathMatching(t *testing.T) {
	// Setup
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/**",
				StrictPath:    true, // Strict path validation
				BodyIDPaths:   []string{"/id"},
				HeaderIDNames: []string{},
				CaseSensitive: false,
			},
		},
	}

	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	h := handler.NewUniHandler(uniService, scenarioService, logger, cfg)

	// Step 1: POST /users/subpath with body: {"id": 1}
	postBody := `{"id": 1, "name": "test user"}`
	postReq := httptest.NewRequest(http.MethodPost, "/users/subpath", bytes.NewReader([]byte(postBody)))
	postReq.Header.Set("Content-Type", "application/json")

	postResp, err := h.HandlePOST(context.Background(), postReq)
	require.NoError(t, err)
	defer postResp.Body.Close()
	assert.Equal(t, http.StatusCreated, postResp.StatusCode)
	t.Logf("POST response status: %d", postResp.StatusCode)

	// Step 2: GET /users/1 - should return 404 with strict path matching
	getReq := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	getResp, err := h.HandleGET(context.Background(), getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode,
		"GET should return 404 with strict path matching - resource at /users/subpath not accessible via /users/1")
	t.Logf("GET response status: %d (expected 404)", getResp.StatusCode)

	// Step 3: PUT /users/1 - should return 404 with strict path matching
	putBody := `{"id": 1, "name": "updated user", "status": "active"}`
	putReq := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewReader([]byte(putBody)))
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := h.HandlePUT(context.Background(), putReq)
	require.NoError(t, err)
	defer putResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, putResp.StatusCode,
		"PUT should return 404 with strict path matching - no access to resource created at different path")
	t.Logf("PUT response status: %d (expected 404)", putResp.StatusCode)

	// Step 4: DELETE /users/1 - should return 404 with strict path matching
	deleteReq := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	deleteResp, err := h.HandleDELETE(context.Background(), deleteReq)
	require.NoError(t, err)
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode,
		"DELETE should return 404 with strict path matching")
	t.Logf("DELETE response status: %d (expected 404)", deleteResp.StatusCode)

	// Step 5: Verify the resource IS accessible via the correct path structure /users/subpath/1
	getCorrectReq := httptest.NewRequest(http.MethodGet, "/users/subpath/1", nil)
	getCorrectResp, err := h.HandleGET(context.Background(), getCorrectReq)
	require.NoError(t, err)
	defer getCorrectResp.Body.Close()
	assert.Equal(t, http.StatusOK, getCorrectResp.StatusCode,
		"Resource should be accessible via correct path structure /users/subpath/1")
}

// TestUpsertBehaviorWithStrictPathFalse verifies upsert works when strict_path=false
func TestUpsertBehaviorWithStrictPathFalse(t *testing.T) {
	// Setup
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"products": {
				PathPattern:   "/products/*",
				StrictPath:    false, // Allow upsert
				BodyIDPaths:   []string{"/id"},
				HeaderIDNames: []string{},
				CaseSensitive: false,
			},
		},
	}

	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	h := handler.NewUniHandler(uniService, scenarioService, logger, cfg)

	// Step 1: PUT /products/999 (non-existent) - should create resource (upsert)
	putBody := `{"id": "999", "name": "new product", "price": 100}`
	putReq := httptest.NewRequest(http.MethodPut, "/products/999", bytes.NewReader([]byte(putBody)))
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := h.HandlePUT(context.Background(), putReq)
	require.NoError(t, err)
	defer putResp.Body.Close()
	assert.Equal(t, http.StatusOK, putResp.StatusCode, "PUT should succeed (upsert create)")
	t.Logf("PUT (upsert) response status: %d", putResp.StatusCode)

	// Step 2: GET /products/999 - should return the created resource
	getReq := httptest.NewRequest(http.MethodGet, "/products/999", nil)
	getResp, err := h.HandleGET(context.Background(), getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "GET should find upsert-created resource")

	var getResponseData map[string]any
	err = json.NewDecoder(getResp.Body).Decode(&getResponseData)
	require.NoError(t, err)
	assert.Equal(t, "999", getResponseData["id"])
	assert.Equal(t, "new product", getResponseData["name"])
	assert.Equal(t, float64(100), getResponseData["price"])
	t.Logf("GET response after upsert: %+v", getResponseData)

	// Step 3: PUT /products/999 again - should update existing resource
	updateBody := `{"id": "999", "name": "updated product", "price": 150}`
	updateReq := httptest.NewRequest(http.MethodPut, "/products/999", bytes.NewReader([]byte(updateBody)))
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := h.HandlePUT(context.Background(), updateReq)
	require.NoError(t, err)
	defer updateResp.Body.Close()
	assert.Equal(t, http.StatusOK, updateResp.StatusCode, "PUT should succeed (update existing)")
	t.Logf("PUT (update) response status: %d", updateResp.StatusCode)

	// Step 4: GET /products/999 - should return the updated resource
	getFinalReq := httptest.NewRequest(http.MethodGet, "/products/999", nil)
	getFinalResp, err := h.HandleGET(context.Background(), getFinalReq)
	require.NoError(t, err)
	defer getFinalResp.Body.Close()
	assert.Equal(t, http.StatusOK, getFinalResp.StatusCode)

	var getFinalResponseData map[string]any
	err = json.NewDecoder(getFinalResp.Body).Decode(&getFinalResponseData)
	require.NoError(t, err)
	assert.Equal(t, "updated product", getFinalResponseData["name"])
	assert.Equal(t, float64(150), getFinalResponseData["price"])
	t.Logf("GET response after update: %+v", getFinalResponseData)
}

// TestStrictPathPreventsUpsert verifies strict_path=true prevents upsert
func TestStrictPathPreventsUpsert(t *testing.T) {
	// Setup
	cfg := &config.UniConfig{
		Sections: map[string]config.Section{
			"admin": {
				PathPattern:   "/admin/users/*",
				StrictPath:    true, // Prevent upsert
				BodyIDPaths:   []string{"/id"},
				HeaderIDNames: []string{},
				CaseSensitive: false,
			},
		},
	}

	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	uniService := service.NewUniService(store, cfg)
	scenarioService := service.NewScenarioService(scenarioStore)
	h := handler.NewUniHandler(uniService, scenarioService, logger, cfg)

	// Step 1: PUT /admin/users/999 (non-existent) - should return 404 (no upsert)
	putBody := `{"id": "999", "name": "admin user", "role": "admin"}`
	putReq := httptest.NewRequest(http.MethodPut, "/admin/users/999", bytes.NewReader([]byte(putBody)))
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := h.HandlePUT(context.Background(), putReq)
	require.NoError(t, err)
	defer putResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, putResp.StatusCode,
		"PUT should return 404 when strict_path=true and resource doesn't exist (no upsert)")
	t.Logf("PUT response status: %d (expected 404, no upsert)", putResp.StatusCode)

	// Step 2: GET /admin/users/999 - should also return 404
	getReq := httptest.NewRequest(http.MethodGet, "/admin/users/999", nil)
	getResp, err := h.HandleGET(context.Background(), getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode, "Resource should not exist")

	// Step 3: Create resource via POST first
	postReq := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader([]byte(putBody)))
	postReq.Header.Set("Content-Type", "application/json")

	postResp, err := h.HandlePOST(context.Background(), postReq)
	require.NoError(t, err)
	defer postResp.Body.Close()
	assert.Equal(t, http.StatusCreated, postResp.StatusCode, "POST should create the resource")
	t.Logf("POST response status: %d", postResp.StatusCode)

	// Step 4: Now PUT /admin/users/999 should succeed (updating existing resource)
	updateBody := `{"id": "999", "name": "updated admin", "role": "super-admin"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/admin/users/999", bytes.NewReader([]byte(updateBody)))
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := h.HandlePUT(context.Background(), updateReq)
	require.NoError(t, err)
	defer updateResp.Body.Close()
	assert.Equal(t, http.StatusOK, updateResp.StatusCode,
		"PUT should succeed when resource exists, even with strict_path=true")
	t.Logf("PUT (update existing) response status: %d", updateResp.StatusCode)
}
