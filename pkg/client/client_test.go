package client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/unimock/pkg/client"
	"github.com/bmcszk/unimock/pkg/model"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectError bool
	}{
		{
			name:        "valid URL",
			baseURL:     "http://localhost:8080",
			expectError: false,
		},
		{
			name:        "empty URL defaults to localhost",
			baseURL:     "",
			expectError: false,
		},
		{
			name:        "invalid URL",
			baseURL:     "://invalid-url",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testClient, err := client.NewClient(tc.baseURL)

			if tc.expectError {
				assertErrorExpected(t, err)
				return
			}

			assertSuccessfulClientCreation(t, testClient, err, tc.baseURL)
		})
	}
}

// assertErrorExpected verifies that an error was returned when expected
func assertErrorExpected(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// assertSuccessfulClientCreation verifies that a client was created successfully
func assertSuccessfulClientCreation(t *testing.T, testClient *client.Client, err error, baseURL string) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if testClient == nil {
		t.Error("expected client, got nil")
		return
	}

	validateClientURL(t, testClient, baseURL)
	validateHTTPClient(t, testClient)
}

// validateClientURL checks that the client's base URL is set correctly
func validateClientURL(t *testing.T, testClient *client.Client, expectedURL string) {
	t.Helper()
	actualURL := testClient.BaseURL.String()
	
	if expectedURL == "" {
		if actualURL != "http://localhost:8080" {
			t.Errorf("expected default base URL, got %s", actualURL)
		}
	} else if actualURL != expectedURL {
		t.Errorf("expected base URL %s, got %s", expectedURL, actualURL)
	}
}

// validateHTTPClient checks that the HTTP client is properly initialized
func validateHTTPClient(t *testing.T, testClient *client.Client) {
	t.Helper()
	if testClient.HTTPClient == nil {
		t.Error("expected HTTP client, got nil")
	}
}

func TestClientOperations(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	apiClient, ctx := setupTestClient(t, server.URL)

	t.Run("GetScenario", func(t *testing.T) {
		testGetScenario(ctx, t, apiClient)
	})

	t.Run("ListScenarios", func(t *testing.T) {
		testListScenarios(ctx, t, apiClient)
	})

	t.Run("CreateScenario", func(t *testing.T) {
		testCreateScenario(ctx, t, apiClient)
	})

	t.Run("UpdateScenario", func(t *testing.T) {
		testUpdateScenario(ctx, t, apiClient)
	})

	t.Run("DeleteScenario", func(t *testing.T) {
		testDeleteScenario(ctx, t, apiClient)
	})
}

// createTestServer creates a mock HTTP server for testing
func createTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleTestServerRequest(w, r)
	}))
}

// handleTestServerRequest handles incoming requests to the test server
func handleTestServerRequest(w http.ResponseWriter, r *http.Request) {
	testScenario := getTestScenario()
	testScenarios := []*model.Scenario{testScenario}

	w.Header().Set("Content-Type", "application/json")

	switch {
	case isListScenariosRequest(r):
		json.NewEncoder(w).Encode(testScenarios)
	case isGetScenarioRequest(r):
		handleGetScenarioRequest(w, r, testScenario)
	case isCreateScenarioRequest(r):
		handleCreateScenarioRequest(w, r)
	case isUpdateScenarioRequest(r):
		handleUpdateScenarioRequest(w, r)
	case isDeleteScenarioRequest(r):
		handleDeleteScenarioRequest(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Endpoint not found"))
	}
}

// getTestScenario returns a test scenario for mocking
func getTestScenario() *model.Scenario {
	return &model.Scenario{
		UUID:        "test-uuid",
		RequestPath: "GET /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Test data"}`,
	}
}

// Request type checking functions
func isListScenariosRequest(r *http.Request) bool {
	return r.Method == http.MethodGet && r.URL.Path == "/_uni/scenarios"
}

func isGetScenarioRequest(r *http.Request) bool {
	return r.Method == http.MethodGet && 
		(r.URL.Path == "/_uni/scenarios/test-uuid" || r.URL.Path == "/_uni/scenarios/not-found")
}

func isCreateScenarioRequest(r *http.Request) bool {
	return r.Method == http.MethodPost && r.URL.Path == "/_uni/scenarios"
}

func isUpdateScenarioRequest(r *http.Request) bool {
	return r.Method == http.MethodPut && 
		(r.URL.Path == "/_uni/scenarios/test-uuid" || r.URL.Path == "/_uni/scenarios/not-found")
}

func isDeleteScenarioRequest(r *http.Request) bool {
	return r.Method == http.MethodDelete && 
		(r.URL.Path == "/_uni/scenarios/test-uuid" || r.URL.Path == "/_uni/scenarios/not-found")
}

// Request handlers
func handleGetScenarioRequest(w http.ResponseWriter, r *http.Request, testScenario *model.Scenario) {
	if r.URL.Path == "/_uni/scenarios/not-found" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
		return
	}
	json.NewEncoder(w).Encode(testScenario)
}

func handleCreateScenarioRequest(w http.ResponseWriter, r *http.Request) {
	var scenario model.Scenario
	if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}
	if scenario.UUID == "" {
		scenario.UUID = "new-uuid"
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scenario)
}

func handleUpdateScenarioRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/_uni/scenarios/not-found" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
		return
	}
	
	var scenario model.Scenario
	if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scenario)
}

func handleDeleteScenarioRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/_uni/scenarios/not-found" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// setupTestClient creates a client and context for testing
func setupTestClient(t *testing.T, serverURL string) (*client.Client, context.Context) {
	t.Helper()
	apiClient, err := client.NewClient(serverURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	return apiClient, ctx
}

// Individual test functions
func testGetScenario(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	// Test getting an existing scenario
	scenario, err := apiClient.GetScenario(ctx, "test-uuid")
	if err != nil {
		t.Errorf("Failed to get scenario: %v", err)
	}
	// Check if scenario is not empty (zero value check)
	if scenario.UUID == "" {
		t.Fatal("Expected scenario with UUID, got empty scenario")
	}
	if scenario.UUID != "test-uuid" {
		t.Errorf("Expected UUID test-uuid, got %s", scenario.UUID)
	}

	// Test getting a non-existent scenario
	_, err = apiClient.GetScenario(ctx, "not-found")
	if err == nil {
		t.Error("Expected error for non-existent scenario, got nil")
	}
}

func testListScenarios(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	scenarios, err := apiClient.ListScenarios(ctx)
	if err != nil {
		t.Errorf("Failed to list scenarios: %v", err)
	}
	if len(scenarios) != 1 {
		t.Errorf("Expected 1 scenario, got %d", len(scenarios))
	}
	if scenarios[0].UUID != "test-uuid" {
		t.Errorf("Expected UUID test-uuid, got %s", scenarios[0].UUID)
	}
}

func testCreateScenario(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	newScenario := &model.Scenario{
		RequestPath: "POST /api/test",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"message":"New test data"}`,
	}

	created, err := apiClient.CreateScenario(ctx, *newScenario)
	if err != nil {
		t.Errorf("Failed to create scenario: %v", err)
	}
	if created.UUID == "" {
		t.Fatal("Expected created scenario with UUID, got empty scenario")
	}
	if created.UUID != "new-uuid" {
		t.Errorf("Expected UUID new-uuid, got %s", created.UUID)
	}
	if created.RequestPath != newScenario.RequestPath {
		t.Errorf("Expected request path %s, got %s", newScenario.RequestPath, created.RequestPath)
	}
}

func testUpdateScenario(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	updateScenario := &model.Scenario{
		UUID:        "test-uuid",
		RequestPath: "PUT /api/test",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message":"Updated test data"}`,
	}

	// Test updating an existing scenario
	updated, err := apiClient.UpdateScenario(ctx, "test-uuid", *updateScenario)
	if err != nil {
		t.Errorf("Failed to update scenario: %v", err)
	}
	if updated.UUID == "" {
		t.Fatal("Expected updated scenario with UUID, got empty scenario")
	}
	if updated.RequestPath != updateScenario.RequestPath {
		t.Errorf("Expected request path %s, got %s", updateScenario.RequestPath, updated.RequestPath)
	}

	// Test updating a non-existent scenario
	_, err = apiClient.UpdateScenario(ctx, "not-found", *updateScenario)
	if err == nil {
		t.Error("Expected error for non-existent scenario, got nil")
	}
}

func testDeleteScenario(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	// Test deleting an existing scenario
	err := apiClient.DeleteScenario(ctx, "test-uuid")
	if err != nil {
		t.Errorf("Failed to delete scenario: %v", err)
	}

	// Test deleting a non-existent scenario
	err = apiClient.DeleteScenario(ctx, "not-found")
	if err == nil {
		t.Error("Expected error for non-existent scenario, got nil")
	}
}

func TestClientContextCancellation(t *testing.T) {
	// Create a test server that delays for a bit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Sleep to simulate a delay
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	// Create a client with the test server URL
	apiClient2, err := client.NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to make a request with canceled context
	_, err = apiClient2.GetScenario(ctx, "test-uuid")
	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}
}

// ========================================
// Universal HTTP Methods Tests
// ========================================

func TestUniversalHTTPMethods(t *testing.T) {
	server := createUniversalHTTPTestServer()
	defer server.Close()

	apiClient, ctx := setupTestClient(t, server.URL)

	t.Run("GET", func(t *testing.T) {
		testGetMethod(ctx, t, apiClient)
	})

	t.Run("HEAD", func(t *testing.T) {
		testHeadMethod(ctx, t, apiClient)
	})

	t.Run("POST", func(t *testing.T) {
		testPostMethod(ctx, t, apiClient)
	})

	t.Run("PUT", func(t *testing.T) {
		testPutMethod(ctx, t, apiClient)
	})

	t.Run("DELETE", func(t *testing.T) {
		testDeleteMethod(ctx, t, apiClient)
	})

	t.Run("PATCH", func(t *testing.T) {
		testPatchMethod(ctx, t, apiClient)
	})

	t.Run("OPTIONS", func(t *testing.T) {
		testOptionsMethod(ctx, t, apiClient)
	})
}

func TestJSONMethods(t *testing.T) {
	server := createUniversalHTTPTestServer()
	defer server.Close()

	apiClient, ctx := setupTestClient(t, server.URL)

	t.Run("PostJSON", func(t *testing.T) {
		testPostJSONMethod(ctx, t, apiClient)
	})

	t.Run("PutJSON", func(t *testing.T) {
		testPutJSONMethod(ctx, t, apiClient)
	})

	t.Run("PatchJSON", func(t *testing.T) {
		testPatchJSONMethod(ctx, t, apiClient)
	})
}

func TestURLHandling(t *testing.T) {
	server := createUniversalHTTPTestServer()
	defer server.Close()

	apiClient, ctx := setupTestClient(t, server.URL)

	testRelativePath(ctx, t, apiClient)
	testAbsoluteURL(ctx, t, apiClient, server.URL)
}

func testRelativePath(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	resp, err := apiClient.Get(ctx, "/api/users", nil)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func testAbsoluteURL(ctx context.Context, t *testing.T, apiClient *client.Client, serverURL string) {
	t.Helper()
	absoluteURL := serverURL + "/api/users"
	resp, err := apiClient.Get(ctx, absoluteURL, nil)
	if err != nil {
		t.Fatalf("Failed to make GET request with absolute URL: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestErrorHandling(t *testing.T) {
	server := createErrorTestServer()
	defer server.Close()

	apiClient, ctx := setupTestClient(t, server.URL)

	t.Run("HTTPError", func(t *testing.T) {
		resp, err := apiClient.Get(ctx, "/error", nil)
		if err != nil {
			t.Fatalf("Expected no error for HTTP error response, got: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("JSONSerializationError", func(t *testing.T) {
		// Create a data structure that can't be serialized to JSON
		invalidData := make(chan int)
		_, err := apiClient.PostJSON(ctx, "/api/users", nil, invalidData)
		if err == nil {
			t.Error("Expected JSON serialization error, got nil")
		}
	})
}

func TestContextCancellation(t *testing.T) {
	server := createSlowTestServer()
	defer server.Close()

	apiClient, err := client.NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to make a request with canceled context
	_, err = apiClient.Get(ctx, "/api/users", nil)
	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}
}

func TestHealthCheck(t *testing.T) {
	server := createHealthCheckTestServer()
	defer server.Close()

	apiClient, err := client.NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test successful health check
	resp, err := apiClient.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if string(resp.Body) != `{"status":"ok"}` {
		t.Errorf("Expected health check response, got %s", string(resp.Body))
	}
}

// Test server implementations

func createUniversalHTTPTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test-value")
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"method":"GET","path":"` + r.URL.Path + `"}`))
		case http.MethodHead:
			w.WriteHeader(http.StatusOK)
			// HEAD responses should not have body
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			body, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"method":"` + r.Method + `","body":"` + string(body) + `"}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodOptions:
			w.Header().Set("Allow", "GET,HEAD,POST,PUT,PATCH,DELETE,OPTIONS")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func createErrorTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
}

func createSlowTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Sleep to simulate a slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
}

func createHealthCheckTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_uni/health" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Individual test functions

func testGetMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	headers := map[string]string{"X-Custom-Header": "custom-value"}
	
	resp, err := apiClient.Get(ctx, "/api/users", headers)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}

	validateBasicResponse(t, resp, 200)
	validateResponseHeaders(t, resp)
	validateResponseBody(t, resp, "GET")
}

func testHeadMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	headers := map[string]string{"X-Custom-Header": "custom-value"}
	
	resp, err := apiClient.Head(ctx, "/api/users", headers)
	if err != nil {
		t.Fatalf("Failed to make HEAD request: %v", err)
	}

	validateBasicResponse(t, resp, 200)
	validateResponseHeaders(t, resp)
	// HEAD responses should have empty body
	if len(resp.Body) != 0 {
		t.Errorf("Expected empty body for HEAD request, got %d bytes", len(resp.Body))
	}
}

func testPostMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	headers := map[string]string{"Content-Type": "application/json"}
	body := []byte(`{"name":"test"}`)
	
	resp, err := apiClient.Post(ctx, "/api/users", headers, body)
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}

	validateBasicResponse(t, resp, 201)
	validateResponseHeaders(t, resp)
	validateResponseBody(t, resp, "POST")
}

func testPutMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	headers := map[string]string{"Content-Type": "application/json"}
	body := []byte(`{"name":"updated"}`)
	
	resp, err := apiClient.Put(ctx, "/api/users", headers, body)
	if err != nil {
		t.Fatalf("Failed to make PUT request: %v", err)
	}

	validateBasicResponse(t, resp, 201)
	validateResponseHeaders(t, resp)
	validateResponseBody(t, resp, "PUT")
}

func testDeleteMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	headers := map[string]string{"X-Custom-Header": "custom-value"}
	
	resp, err := apiClient.Delete(ctx, "/api/users", headers)
	if err != nil {
		t.Fatalf("Failed to make DELETE request: %v", err)
	}

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
	validateResponseHeaders(t, resp)
}

func testPatchMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	headers := map[string]string{"Content-Type": "application/json"}
	body := []byte(`{"name":"patched"}`)
	
	resp, err := apiClient.Patch(ctx, "/api/users", headers, body)
	if err != nil {
		t.Fatalf("Failed to make PATCH request: %v", err)
	}

	validateBasicResponse(t, resp, 201)
	validateResponseHeaders(t, resp)
	validateResponseBody(t, resp, "PATCH")
}

func testOptionsMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	headers := map[string]string{"X-Custom-Header": "custom-value"}
	
	resp, err := apiClient.Options(ctx, "/api/users", headers)
	if err != nil {
		t.Fatalf("Failed to make OPTIONS request: %v", err)
	}

	validateBasicResponse(t, resp, 200)
	validateResponseHeaders(t, resp)
	
	// Check for Allow header
	allowHeader := resp.Headers.Get("Allow")
	if allowHeader == "" {
		t.Error("Expected Allow header in OPTIONS response")
	}
}

func testPostJSONMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	data := map[string]any{
		"name": "test",
		"age":  25,
	}
	
	resp, err := apiClient.PostJSON(ctx, "/api/users", nil, data)
	if err != nil {
		t.Fatalf("Failed to make PostJSON request: %v", err)
	}

	validateBasicResponse(t, resp, 201)
	validateResponseHeaders(t, resp)
	validateResponseBody(t, resp, "POST")
}

func testPutJSONMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	data := map[string]any{
		"name": "updated",
		"age":  30,
	}
	
	resp, err := apiClient.PutJSON(ctx, "/api/users", nil, data)
	if err != nil {
		t.Fatalf("Failed to make PutJSON request: %v", err)
	}

	validateBasicResponse(t, resp, 201)
	validateResponseHeaders(t, resp)
	validateResponseBody(t, resp, "PUT")
}

func testPatchJSONMethod(ctx context.Context, t *testing.T, apiClient *client.Client) {
	t.Helper()
	data := map[string]any{
		"name": "patched",
	}
	
	resp, err := apiClient.PatchJSON(ctx, "/api/users", nil, data)
	if err != nil {
		t.Fatalf("Failed to make PatchJSON request: %v", err)
	}

	validateBasicResponse(t, resp, 201)
	validateResponseHeaders(t, resp)
	validateResponseBody(t, resp, "PATCH")
}

// Validation helper functions

func validateBasicResponse(t *testing.T, resp *client.Response, expectedStatus int) {
	t.Helper()
	if resp == nil {
		t.Fatal("Response is nil")
	}
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

func validateResponseHeaders(t *testing.T, resp *client.Response) {
	t.Helper()
	if resp.Headers == nil {
		t.Fatal("Response headers are nil")
	}
	
	testHeader := resp.Headers.Get("X-Test-Header")
	if testHeader != "test-value" {
		t.Errorf("Expected X-Test-Header 'test-value', got '%s'", testHeader)
	}
}

func validateResponseBody(t *testing.T, resp *client.Response, expectedMethod string) {
	t.Helper()
	if len(resp.Body) == 0 {
		t.Error("Expected response body, got empty")
		return
	}
	
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, expectedMethod) {
		t.Errorf("Expected response body to contain method '%s', got: %s", expectedMethod, bodyStr)
	}
}
