package pkg_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/pkg"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer_ScenariosOnly(t *testing.T) {
	t.Run("should create server with empty sections and scenarios from file", func(t *testing.T) {
		// Create temporary config file with only scenarios
		configContent := `
scenarios:
  - uuid: "test-scenario"
    method: "GET"
    path: "/test"
    status_code: 200
    content_type: "application/json"
    data: '{"message": "test response"}'
`
		configFile := createTempConfigFile(t, configContent)
		defer os.Remove(configFile)

		serverConfig := &config.ServerConfig{
			Port:     "0",
			LogLevel: "error",
		}

		// Load unified config from file - should work with scenarios only
		unifiedConfig, err := config.LoadUnifiedFromYAML(configFile)
		require.NoError(t, err)

		server, err := pkg.NewServer(serverConfig, unifiedConfig)
		require.NoError(t, err, "Should create server with empty sections when scenarios exist")
		require.NotNil(t, server)
	})

	t.Run("should reject completely empty config (no sections, no scenarios)", func(t *testing.T) {
		// Create empty unified config
		unifiedConfig := &config.UnifiedConfig{
			Sections:  map[string]config.Section{},
			Scenarios: []config.ScenarioConfig{},
		}

		serverConfig := &config.ServerConfig{
			Port:     "0",
			LogLevel: "error",
		}

		server, err := pkg.NewServer(serverConfig, unifiedConfig)
		require.Error(t, err, "Should reject config with no sections and no scenarios")
		require.Nil(t, server)

		var configErr *pkg.ConfigError
		require.ErrorAs(t, err, &configErr)
		assert.Contains(t, configErr.Message, "no sections or scenarios defined")
	})

	t.Run("should accept config with sections but no scenarios", func(t *testing.T) {
		serverConfig := &config.ServerConfig{
			Port:     "0",
			LogLevel: "error",
		}

		unifiedConfig := &config.UnifiedConfig{
			Sections: map[string]config.Section{
				"users": {
					PathPattern: "/users/*",
					BodyIDPaths: []string{"/id"},
				},
			},
			Scenarios: []config.ScenarioConfig{},
		}

		server, err := pkg.NewServer(serverConfig, unifiedConfig)
		require.NoError(t, err, "Should accept config with sections")
		require.NotNil(t, server)
	})

	t.Run("should handle empty unified config gracefully", func(t *testing.T) {
		serverConfig := &config.ServerConfig{
			Port:     "0",
			LogLevel: "error",
		}

		// Nil unified config should be rejected
		server, err := pkg.NewServer(serverConfig, nil)
		require.Error(t, err, "Should reject nil unified config")
		require.Nil(t, server)

		var configErr *pkg.ConfigError
		require.ErrorAs(t, err, &configErr)
		assert.Contains(t, configErr.Message, "unified configuration is nil")
	})
}

func TestScenariosOnlyIntegration(t *testing.T) {
	t.Run("scenarios work without sections", func(t *testing.T) {
		// Create config with only scenarios
		configContent := `
scenarios:
  - uuid: "integration-test"
    method: "GET"
    path: "/api/test/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "message": "Integration test response"
      }
`
		configFile := createTempConfigFile(t, configContent)
		defer os.Remove(configFile)

		serverConfig := &config.ServerConfig{
			Port:     "0",
			LogLevel: "error",
		}

		// Load unified config from file
		unifiedConfig, err := config.LoadUnifiedFromYAML(configFile)
		require.NoError(t, err)

		server, err := pkg.NewServer(serverConfig, unifiedConfig)
		require.NoError(t, err)

		// Start test server
		testServer := httptest.NewServer(server.Handler)
		defer testServer.Close()

		// Test the scenario endpoint
		resp, err := http.Get(testServer.URL + "/api/test/123")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		expectedJSON := `{
  "id": "123",
  "message": "Integration test response"
}`
		assert.JSONEq(t, expectedJSON, string(body))

		// Test non-scenario endpoint returns 404
		resp2, err := http.Get(testServer.URL + "/api/unknown")
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, 404, resp2.StatusCode)
	})

	t.Run("POST scenario works without sections", func(t *testing.T) {
		configContent := `
scenarios:
  - uuid: "post-test"
    method: "POST"
    path: "/api/create"
    status_code: 201
    content_type: "application/json"
    location: "/api/created/456"
    data: '{"id": "456", "status": "created"}'
`
		configFile := createTempConfigFile(t, configContent)
		defer os.Remove(configFile)

		serverConfig := &config.ServerConfig{
			Port:     "0",
			LogLevel: "error",
		}

		// Load unified config from file
		unifiedConfig, err := config.LoadUnifiedFromYAML(configFile)
		require.NoError(t, err)

		server, err := pkg.NewServer(serverConfig, unifiedConfig)
		require.NoError(t, err)

		testServer := httptest.NewServer(server.Handler)
		defer testServer.Close()

		// Test POST scenario
		resp, err := http.Post(testServer.URL+"/api/create", "application/json",
			strings.NewReader(`{"data": "test"}`))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "/api/created/456", resp.Header.Get("Location"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"id": "456", "status": "created"}`, string(body))
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
