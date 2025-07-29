//go:build e2e

package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/unimock/pkg"
	"github.com/bmcszk/unimock/pkg/config"
)

func createTestUnifiedConfigFile(t *testing.T) string {
	t.Helper()
	unifiedConfigYAML := `
sections:
  test_scenarios:
    path_pattern: "/test-scenarios/**"
    body_id_paths:
      - "/id"
    header_id_names: ["X-User-ID"]
    return_body: true

scenarios:
  - uuid: "e2e-test-scenario-001"
    method: "GET"
    path: "/test-scenarios/user/123"
    status_code: 200
    content_type: "application/json"
    headers:
      X-Test-Header: "from-file"
    data: |
      {
        "id": "123",
        "name": "File Scenario User",
        "source": "yaml-file"
      }

  - uuid: "e2e-test-scenario-002"
    method: "POST"
    path: "/test-scenarios/users"
    status_code: 201
    content_type: "application/json"
    location: "/test-scenarios/users/456"
    data: |
      {
        "id": "456",
        "name": "Created from File",
        "source": "yaml-file"
      }

  - uuid: "e2e-test-scenario-003"
    method: "HEAD"
    path: "/test-scenarios/user/789"
    status_code: 200
    content_type: "application/json"
    headers:
      X-User-ID: "789"
      Last-Modified: "Wed, 21 Oct 2015 07:28:00 GMT"
`

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-unified-config.yaml")
	err := os.WriteFile(configFile, []byte(unifiedConfigYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create unified config file: %v", err)
	}
	return configFile
}

func setupScenarioTestServer(t *testing.T, configFile string) (*http.Server, string) {
	t.Helper()
	serverConfig := &config.ServerConfig{
		Port:       "0",
		LogLevel:   "info",
		ConfigPath: configFile,
	}

	uniConfig, err := config.LoadFromYAML(configFile)
	if err != nil {
		t.Fatalf("Failed to load unified config: %v", err)
	}

	srv, err := pkg.NewServer(serverConfig, uniConfig)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	baseURL := fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)
	return srv, baseURL
}

func cleanupTestServer(t *testing.T, srv *http.Server) {
	t.Helper()
	if err := srv.Close(); err != nil {
		t.Logf("Error closing server: %v", err)
	}
}

func createIntegrationUnifiedConfigFile(t *testing.T) string {
	t.Helper()
	unifiedConfigYAML := `
sections:
  integration_test:
    path_pattern: "/integration-test/**"
    body_id_paths:
      - "/id"
    header_id_names: ["X-Resource-ID"]
    return_body: true

scenarios:
  - uuid: "file-scenario-001"
    method: "GET"
    path: "/integration-test/file-resource"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "source": "file",
        "type": "file-scenario"
      }
`

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "integration-unified-config.yaml")
	err := os.WriteFile(configFile, []byte(unifiedConfigYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create unified config file: %v", err)
	}
	return configFile
}

// TestScenarioFileLoading tests that scenarios can be loaded from a unified config file
func TestScenarioFileLoading(t *testing.T) {
	configFile := createTestUnifiedConfigFile(t)
	srv, baseURL := setupScenarioTestServer(t, configFile)
	defer cleanupTestServer(t, srv)

	t.Run("GET scenario from file", func(t *testing.T) {
		testGETScenarioFromFile(t, baseURL)
	})

	t.Run("POST scenario from file", func(t *testing.T) {
		testPOSTScenarioFromFile(t, baseURL)
	})

	t.Run("HEAD scenario from file", func(t *testing.T) {
		testHEADScenarioFromFile(t, baseURL)
	})
}

func testGETScenarioFromFile(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/test-scenarios/user/123")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	if resp.Header.Get("X-Test-Header") != "from-file" {
		t.Errorf("Expected X-Test-Header 'from-file', got %s", resp.Header.Get("X-Test-Header"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var responseData map[string]any
	if err := json.Unmarshal(body, &responseData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	validateScenarioResponse(t, responseData, "123", "File Scenario User", "yaml-file")
}

func testPOSTScenarioFromFile(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Post(baseURL+"/test-scenarios/users", "application/json", nil)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Location") != "/test-scenarios/users/456" {
		t.Errorf("Expected Location '/test-scenarios/users/456', got %s", resp.Header.Get("Location"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var responseData map[string]any
	if err := json.Unmarshal(body, &responseData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	validateScenarioResponse(t, responseData, "456", "Created from File", "yaml-file")
}

func testHEADScenarioFromFile(t *testing.T, baseURL string) {
	t.Helper()
	resp := makeHEADRequest(t, baseURL+"/test-scenarios/user/789")
	defer resp.Body.Close()

	validateHEADResponse(t, resp)
	validateHEADHeaders(t, resp)
	validateEmptyBody(t, resp)
}

func makeHEADRequest(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		t.Fatalf("Failed to create HEAD request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HEAD request failed: %v", err)
	}
	return resp
}

func validateHEADResponse(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func validateHEADHeaders(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.Header.Get("X-User-ID") != "789" {
		t.Errorf("Expected X-User-ID '789', got %s", resp.Header.Get("X-User-ID"))
	}
	expected := "Wed, 21 Oct 2015 07:28:00 GMT"
	if resp.Header.Get("Last-Modified") != expected {
		t.Errorf("Expected Last-Modified '%s', got %s", expected, resp.Header.Get("Last-Modified"))
	}
}

func validateEmptyBody(t *testing.T, resp *http.Response) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("Expected empty body for HEAD request, got %d bytes", len(body))
	}
}

func validateScenarioResponse(
	t *testing.T, responseData map[string]any, expectedID, expectedName, expectedSource string,
) {
	t.Helper()
	if responseData["id"] != expectedID {
		t.Errorf("Expected id '%s', got %v", expectedID, responseData["id"])
	}
	if responseData["name"] != expectedName {
		t.Errorf("Expected name '%s', got %v", expectedName, responseData["name"])
	}
	if responseData["source"] != expectedSource {
		t.Errorf("Expected source '%s', got %v", expectedSource, responseData["source"])
	}
}

// TestScenarioFileAndRuntimeAPIIntegration tests that file-based and runtime scenarios work together
func TestScenarioFileAndRuntimeAPIIntegration(t *testing.T) {
	configFile := createIntegrationUnifiedConfigFile(t)
	srv, baseURL := setupScenarioTestServer(t, configFile)
	defer cleanupTestServer(t, srv)

	t.Run("File scenario works", func(t *testing.T) {
		testFileScenarioWorks(t, baseURL)
	})

	t.Run("Runtime API scenario works alongside file scenario", func(t *testing.T) {
		testRuntimeAPIScenarioIntegration(t, baseURL)
	})
}

func testFileScenarioWorks(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/integration-test/file-resource")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var responseData map[string]any
	if err := json.Unmarshal(body, &responseData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if responseData["source"] != "file" {
		t.Errorf("Expected source 'file', got %v", responseData["source"])
	}
}

func testRuntimeAPIScenarioIntegration(t *testing.T, baseURL string) {
	t.Helper()
	// Create and add runtime scenario
	createRuntimeScenario(t, baseURL)

	// Test the runtime scenario works
	testRuntimeScenarioResponse(t, baseURL)

	// Verify file scenario still works
	verifyFileScenarioStillWorks(t, baseURL)
}

func createRuntimeScenario(t *testing.T, baseURL string) {
	t.Helper()
	runtimeScenarioJSON := `{
		"uuid": "runtime-scenario-001",
		"requestPath": "GET /integration-test/runtime-resource",
		"statusCode": 200,
		"contentType": "application/json",
		"data": "{\"source\": \"runtime\", \"type\": \"runtime-scenario\"}"
	}`

	resp, err := http.Post(baseURL+"/_uni/scenarios", "application/json",
		strings.NewReader(runtimeScenarioJSON))
	if err != nil {
		t.Fatalf("Failed to create runtime scenario: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201 for scenario creation, got %d", resp.StatusCode)
	}
}

func testRuntimeScenarioResponse(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/integration-test/runtime-resource")
	if err != nil {
		t.Fatalf("GET request for runtime scenario failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var responseData map[string]any
	if err := json.Unmarshal(body, &responseData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if responseData["source"] != "runtime" {
		t.Errorf("Expected source 'runtime', got %v", responseData["source"])
	}
}

func verifyFileScenarioStillWorks(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/integration-test/file-resource")
	if err != nil {
		t.Fatalf("GET request for file scenario failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for file scenario, got %d", resp.StatusCode)
	}
}
