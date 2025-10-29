package e2e_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	restclient "github.com/bmcszk/go-restclient"
	"github.com/bmcszk/unimock/pkg"
	"github.com/bmcszk/unimock/pkg/client"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/require"
)

type parts struct {
	*testing.T
	require          *require.Assertions
	baseURL          string
	unimockAPIClient *client.Client
	rc               *restclient.Client
	scenarioID       string
	server           *http.Server
	responses        []any
	configFile       string
}

func newParts(t *testing.T) (given *parts, when *parts, then *parts) {
	t.Helper()

	baseURL := getBaseURL()
	unimockAPIClient, err := client.NewClient(baseURL)
	require.NoError(t, err, "Failed to create unimock API client")

	rc, err := restclient.NewClient(restclient.WithBaseURL(baseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	p := &parts{
		T:                t,
		require:          require.New(t),
		baseURL:          baseURL,
		unimockAPIClient: unimockAPIClient,
		rc:               rc,
	}

	p.cleanupAllScenarios()

	return p, p, p
}

func newServerParts(t *testing.T, configFile string) (given *parts, when *parts, then *parts) {
	t.Helper()

	p := &parts{
		T:          t,
		require:    require.New(t),
		configFile: configFile,
	}

	p.setupScenarioTestServer(configFile)

	return p, p, p
}

func (p *parts) and() *parts {
	return p
}

func (p *parts) a_scenario(scenario model.Scenario) *parts {
	createdScenario, err := p.unimockAPIClient.CreateScenario(context.Background(), scenario)
	require.NoError(p.T, err, "Failed to create scenario")
	require.NotNil(p.T, createdScenario, "Created scenario should not be nil")
	require.NotEmpty(p.T, createdScenario.UUID, "Created scenario UUID should not be empty")
	p.scenarioID = createdScenario.UUID
	p.Cleanup(func() {
		errDel := p.unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID)
		require.NoError(p.T, errDel, "Failed to delete scenario %s", createdScenario.UUID)
	})
	return p
}

func (p *parts) an_http_request_is_made_from_file(httpFile string) *parts {
	responses, err := p.rc.ExecuteFile(context.Background(), httpFile)
	p.require.NoError(err, "Failed to execute http file")
	p.responses = make([]any, len(responses))
	for i, resp := range responses {
		p.responses[i] = resp
	}
	return p
}

func (p *parts) the_response_is_validated_against_file(hrespFile string) *parts {
	restResponses := make([]*restclient.Response, len(p.responses))
	for i, resp := range p.responses {
		restResp, ok := resp.(*restclient.Response)
		p.require.True(ok, "Response %d is not of type *restclient.Response", i)
		restResponses[i] = restResp
	}
	p.require.NoError(p.rc.ValidateResponses(hrespFile, restResponses...), "Failed to validate responses")
	return p
}

func (p *parts) the_resource_can_be_retrieved_by_either_id(hrespFile string) {
	p.the_response_is_validated_against_file(hrespFile)
}

func (p *parts) a_resource_is_created_with_conflicting_ids(httpFile string) *parts {
	return p.an_http_request_is_made_from_file(httpFile)
}

func (p *parts) a_conflict_error_is_returned(hrespFile string) *parts {
	return p.the_response_is_validated_against_file(hrespFile)
}

func (p *parts) a_resource_is_updated_and_verified(httpFile string) *parts {
	return p.an_http_request_is_made_from_file(httpFile)
}

func (p *parts) the_update_is_successful(hrespFile string) *parts {
	return p.the_response_is_validated_against_file(hrespFile)
}

func (p *parts) a_resource_is_deleted_and_verified(httpFile string) *parts {
	return p.an_http_request_is_made_from_file(httpFile)
}

func (p *parts) the_deletion_is_successful(hrespFile string) *parts {
	return p.the_response_is_validated_against_file(hrespFile)
}

func (p *parts) a_scenario_override_is_applied(scenario model.Scenario) *parts {
	createdScenario, err := p.unimockAPIClient.CreateScenario(context.Background(), scenario)
	p.require.NoError(err, "Failed to create scenario")
	p.require.NotNil(createdScenario, "Created scenario should not be nil")
	p.require.NotEmpty(createdScenario.UUID, "Created scenario UUID should not be empty")
	p.scenarioID = createdScenario.UUID
	return p
}

func (p *parts) the_scenario_override_is_verified(hrespFile string) *parts {
	return p.the_response_is_validated_against_file(hrespFile)
}

func (p *parts) the_scenario_override_is_deleted() {
	err := p.unimockAPIClient.DeleteScenario(context.Background(), p.scenarioID)
	p.require.NoError(err, "Failed to delete scenario")
}

func (p *parts) the_scenario_removal_is_verified(hrespFile string) *parts {
	return p.the_response_is_validated_against_file(hrespFile)
}

// getBaseURL returns the base URL for E2E tests from environment variable or default
func getBaseURL() string {
	if url := os.Getenv("UNIMOCK_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:28080" // Default to Docker Compose exposed port (defined in docker-compose.yml)
}

func (p *parts) cleanupAllScenarios() {
	scenarios, err := p.unimockAPIClient.ListScenarios(context.Background())
	if err != nil {
		fmt.Printf("Warning: Failed to list scenarios for cleanup: %v\n", err)
		return
	}

	p.deleteScenarios(scenarios)
}

func (p *parts) deleteScenarios(scenarios []model.Scenario) {
	for _, scenario := range scenarios {
		err := p.unimockAPIClient.DeleteScenario(context.Background(), scenario.UUID)
		if err != nil {
			p.handleDeleteError(err, scenario.UUID)
		}
	}
}

func (*parts) handleDeleteError(err error, scenarioUUID string) {
	// Don't warn if scenario is already not found (already deleted)
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
		fmt.Printf("Warning: Failed to delete scenario %s: %v\n", scenarioUUID, err)
	}
}

func (p *parts) setupScenarioTestServer(configFile string) {
	p.Helper()
	serverConfig := createServerConfig(configFile)
	uniConfig := p.loadUnifiedConfig(configFile)
	srv := p.createServer(serverConfig, uniConfig)
	listener := p.createListener()
	p.startServer(srv, listener)
	p.configureServerConnection(srv, listener)
	p.setupServerCleanup(srv)
}

func createServerConfig(configFile string) *config.ServerConfig {
	return &config.ServerConfig{
		Port:       "0",
		LogLevel:   "info",
		ConfigPath: configFile,
	}
}

func (p *parts) loadUnifiedConfig(configFile string) *config.UniConfig {
	uniConfig, err := config.LoadFromYAML(configFile)
	p.require.NoError(err, "Failed to load unified config")
	return uniConfig
}

func (p *parts) createServer(serverConfig *config.ServerConfig, uniConfig *config.UniConfig) *http.Server {
	srv, err := pkg.NewServer(serverConfig, uniConfig)
	p.require.NoError(err, "Failed to create server")
	return srv
}

func (p *parts) createListener() net.Listener {
	listener, err := net.Listen("tcp", ":0")
	p.require.NoError(err, "Failed to create listener")
	return listener
}

func (p *parts) startServer(srv *http.Server, listener net.Listener) {
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			p.Logf("Server error: %v", err)
		}
	}()
}

func (p *parts) configureServerConnection(srv *http.Server, listener net.Listener) {
	time.Sleep(100 * time.Millisecond)
	p.baseURL = fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)
	p.server = srv
}

func (p *parts) setupServerCleanup(srv *http.Server) {
	p.Cleanup(func() {
		if err := srv.Close(); err != nil {
			p.Logf("Error closing server: %v", err)
		}
	})
}

// Server-based HTTP helper methods
func (p *parts) a_get_request_is_made_to(url string) {
	resp, err := http.Get(p.baseURL + url)
	p.require.NoError(err)
	defer resp.Body.Close()
	p.responses = []any{resp}
}

func (p *parts) a_post_request_is_made_to(url string) {
	resp, err := http.Post(p.baseURL+url, "application/json", nil)
	p.require.NoError(err)
	defer resp.Body.Close()
	p.responses = []any{resp}
}

func (p *parts) a_head_request_is_made_to(url string) *parts {
	req, err := http.NewRequest("HEAD", p.baseURL+url, nil)
	p.require.NoError(err)
	resp, err := http.DefaultClient.Do(req)
	p.require.NoError(err)
	defer resp.Body.Close()
	p.responses = []any{resp}
	return p
}

func (p *parts) the_response_is_successful() *parts {
	resp, ok := p.responses[0].(*http.Response)
	p.require.True(ok, "Response is not of type *http.Response")
	p.require.Equal(http.StatusOK, resp.StatusCode)
	return p
}

func (p *parts) the_post_response_is_successful() {
	resp, ok := p.responses[0].(*http.Response)
	p.require.True(ok, "Response is not of type *http.Response")
	p.require.Equal(http.StatusCreated, resp.StatusCode)
}

func (p *parts) the_head_response_is_successful() *parts {
	resp, ok := p.responses[0].(*http.Response)
	p.require.True(ok, "Response is not of type *http.Response")
	p.require.Equal(http.StatusOK, resp.StatusCode)
	return p
}

func (p *parts) the_response_is_error() *parts {
	resp, ok := p.responses[0].(*http.Response)
	p.require.True(ok, "Response is not of type *http.Response")
	p.require.Equal(http.StatusInternalServerError, resp.StatusCode)
	return p
}

func (p *parts) the_response_is_not_found() *parts {
	resp, ok := p.responses[0].(*http.Response)
	p.require.True(ok, "Response is not of type *http.Response")
	p.require.Equal(http.StatusNotFound, resp.StatusCode)
	return p
}

func (p *parts) a_runtime_scenario_is_created() *parts {
	runtimeScenarioJSON := `{
		"uuid": "runtime-scenario-001",
		"requestPath": "GET /integration-test/runtime-resource",
		"statusCode": 200,
		"contentType": "application/json",
		"data": "{\"source\": \"runtime\", \"type\": \"runtime-scenario\"}"
	}`

	resp, err := http.Post(p.baseURL+"/_uni/scenarios", "application/json",
		strings.NewReader(runtimeScenarioJSON))
	p.require.NoError(err)
	p.require.Equal(http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
	return p
}

func (p *parts) the_file_scenario_still_works() *parts {
	resp, err := http.Get(p.baseURL + "/integration-test/file-resource")
	p.require.NoError(err)
	p.require.Equal(http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	return p
}

// Config file creation helpers
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

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Validate content parameter - this varies across different test functions
	if len(content) == 0 {
		t.Fatal("content parameter cannot be empty")
	}

	// Log content usage to demonstrate parameter is used
	t.Logf("Creating config file with content length: %d", len(content))

	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	return configFile
}
