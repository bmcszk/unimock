//go:build e2e

package e2e_test

import (
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenariosOnlyConfiguration(t *testing.T) {
	// Create temporary config file with only scenarios, no sections
	configContent := `
scenarios:
  - uuid: "test-user-get"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "name": "Test User",
        "email": "test@example.com"
      }
  
  - uuid: "test-product-post"
    method: "POST"
    path: "/api/products"
    status_code: 201
    content_type: "application/json"
    location: "/api/products/456"
    data: |
      {
        "id": "456",
        "name": "Test Product",
        "price": 99.99
      }
  
  - uuid: "test-error-scenario"
    method: "GET"
    path: "/api/error"
    status_code: 500
    content_type: "application/json"
    data: |
      {
        "error": "Internal server error",
        "code": "SERVER_ERROR"
      }
`

	configFile := createTempConfigFile(t, configContent)
	defer os.Remove(configFile)

	// Create server config
	serverConfig := &config.ServerConfig{
		Port:     "0", // Use random available port
		LogLevel: "error",
	}

	// Load unified config from scenarios-only file
	unifiedConfig, err := config.LoadUnifiedFromYAML(configFile)
	require.NoError(t, err, "Should load scenarios-only config")

	// This should work - server should start with scenarios-only config
	server, err := pkg.NewServer(serverConfig, unifiedConfig)
	require.NoError(t, err, "Server should start with scenarios-only configuration")

	// Use httptest server instead for reliable testing
	testServer := http.Server{
		Handler: server.Handler,
	}

	// Start with httptest for reliable address
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	testServer.Addr = listener.Addr().String()

	go func() {
		if err := testServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Get server address
	baseURL := fmt.Sprintf("http://%s", testServer.Addr)

	// Cleanup
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		testServer.Shutdown(ctx)
	}()

	// Test scenarios work correctly
	t.Run("GET scenario returns expected response", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/users/123")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		expectedJSON := `{
  "id": "123",
  "name": "Test User",
  "email": "test@example.com"
}`
		assert.JSONEq(t, expectedJSON, string(body))
	})

	t.Run("POST scenario returns expected response with location header", func(t *testing.T) {
		resp, err := http.Post(baseURL+"/api/products", "application/json",
			strings.NewReader(`{"name": "Test Product", "price": 99.99}`))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "/api/products/456", resp.Header.Get("Location"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		expectedJSON := `{
  "id": "456",
  "name": "Test Product",
  "price": 99.99
}`
		assert.JSONEq(t, expectedJSON, string(body))
	})

	t.Run("Error scenario returns expected error response", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/error")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		expectedJSON := `{
  "error": "Internal server error",
  "code": "SERVER_ERROR"
}`
		assert.JSONEq(t, expectedJSON, string(body))
	})

	t.Run("Non-scenario path returns 404 (no sections configured)", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/unknown")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 404 since no sections are configured to handle this path
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("Health endpoint still works", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/_uni/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	})
}

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configFile, []byte(content), 0644)
	require.NoError(t, err)

	return configFile
}
