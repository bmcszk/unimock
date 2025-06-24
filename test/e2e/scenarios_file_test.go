//go:build e2e

package e2e

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

// TestScenarioFileLoading tests that scenarios can be loaded from a YAML file
func TestScenarioFileLoading(t *testing.T) {
	// Create a temporary scenarios file
	scenariosYAML := `
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
	scenariosFile := filepath.Join(tempDir, "test-scenarios.yaml")
	err := os.WriteFile(scenariosFile, []byte(scenariosYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create scenarios file: %v", err)
	}

	// Create server configuration with scenarios file
	serverConfig := &config.ServerConfig{
		Port:          "0", // Use random port
		LogLevel:      "info",
		ConfigPath:    "../../config.yaml",
		ScenariosFile: scenariosFile,
	}

	// Load mock configuration
	mockConfig, err := config.LoadFromYAML("../../config.yaml")
	if err != nil {
		t.Fatalf("Failed to load mock config: %v", err)
	}

	// Create and start server
	srv, err := pkg.NewServer(serverConfig, mockConfig)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background on a listener
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start and get the actual port
	time.Sleep(100 * time.Millisecond)
	baseURL := fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)

	// Ensure server is stopped after test
	defer func() {
		if err := srv.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	}()

	// Test scenarios loaded from file
	t.Run("GET scenario from file", func(t *testing.T) {
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

		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if responseData["id"] != "123" {
			t.Errorf("Expected id '123', got %v", responseData["id"])
		}

		if responseData["name"] != "File Scenario User" {
			t.Errorf("Expected name 'File Scenario User', got %v", responseData["name"])
		}

		if responseData["source"] != "yaml-file" {
			t.Errorf("Expected source 'yaml-file', got %v", responseData["source"])
		}
	})

	t.Run("POST scenario from file", func(t *testing.T) {
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

		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if responseData["id"] != "456" {
			t.Errorf("Expected id '456', got %v", responseData["id"])
		}

		if responseData["name"] != "Created from File" {
			t.Errorf("Expected name 'Created from File', got %v", responseData["name"])
		}
	})

	t.Run("HEAD scenario from file", func(t *testing.T) {
		req, err := http.NewRequest("HEAD", baseURL+"/test-scenarios/user/789", nil)
		if err != nil {
			t.Fatalf("Failed to create HEAD request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("HEAD request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
		}

		if resp.Header.Get("X-User-ID") != "789" {
			t.Errorf("Expected X-User-ID '789', got %s", resp.Header.Get("X-User-ID"))
		}

		if resp.Header.Get("Last-Modified") != "Wed, 21 Oct 2015 07:28:00 GMT" {
			t.Errorf("Expected Last-Modified 'Wed, 21 Oct 2015 07:28:00 GMT', got %s", resp.Header.Get("Last-Modified"))
		}

		// HEAD should have empty body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if len(body) != 0 {
			t.Errorf("Expected empty body for HEAD request, got %d bytes", len(body))
		}
	})
}

// TestScenarioFileAndRuntimeAPIIntegration tests that file-based and runtime scenarios work together
func TestScenarioFileAndRuntimeAPIIntegration(t *testing.T) {
	// Create a temporary scenarios file with one scenario
	scenariosYAML := `
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
	scenariosFile := filepath.Join(tempDir, "integration-scenarios.yaml")
	err := os.WriteFile(scenariosFile, []byte(scenariosYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create scenarios file: %v", err)
	}

	// Create server configuration with scenarios file
	serverConfig := &config.ServerConfig{
		Port:          "0", // Use random port
		LogLevel:      "info",
		ConfigPath:    "../../config.yaml",
		ScenariosFile: scenariosFile,
	}

	// Load mock configuration
	mockConfig, err := config.LoadFromYAML("../../config.yaml")
	if err != nil {
		t.Fatalf("Failed to load mock config: %v", err)
	}

	// Create and start server
	srv, err := pkg.NewServer(serverConfig, mockConfig)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background on a listener
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	baseURL := fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)

	// Ensure server is stopped after test
	defer func() {
		if err := srv.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	}()

	// Test that file scenario works
	t.Run("File scenario works", func(t *testing.T) {
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

		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if responseData["source"] != "file" {
			t.Errorf("Expected source 'file', got %v", responseData["source"])
		}
	})

	// Add a runtime scenario via API and test it works alongside file scenario
	t.Run("Runtime API scenario works alongside file scenario", func(t *testing.T) {
		// Create a runtime scenario via API
		runtimeScenarioJSON := `{
			"uuid": "runtime-scenario-001",
			"requestPath": "GET /integration-test/runtime-resource",
			"statusCode": 200,
			"contentType": "application/json",
			"data": "{\"source\": \"runtime\", \"type\": \"runtime-scenario\"}"
		}`

		// Add scenario via API
		resp, err := http.Post(baseURL+"/_uni/scenarios", "application/json", 
			strings.NewReader(runtimeScenarioJSON))
		if err != nil {
			t.Fatalf("Failed to create runtime scenario: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("Expected status 201 for scenario creation, got %d", resp.StatusCode)
		}

		// Test the runtime scenario works
		resp, err = http.Get(baseURL + "/integration-test/runtime-resource")
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

		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if responseData["source"] != "runtime" {
			t.Errorf("Expected source 'runtime', got %v", responseData["source"])
		}

		// Verify file scenario still works
		resp, err = http.Get(baseURL + "/integration-test/file-resource")
		if err != nil {
			t.Fatalf("GET request for file scenario failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200 for file scenario, got %d", resp.StatusCode)
		}
	})
}