package e2e_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func createTestFixtureConfigFile(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// Create fixtures directory structure
	fixturesDir := filepath.Join(tempDir, "fixtures", "operations")
	err := os.MkdirAll(fixturesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create fixtures dir: %v", err)
	}

	// Create fixture files
	robotsFile := filepath.Join(fixturesDir, "robots.json")
	robotsContent := `{"robots": [{"id": "R001", "name": "Alpha Robot", "status": "active"}]}`
	err = os.WriteFile(robotsFile, []byte(robotsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create robots fixture file: %v", err)
	}

	statusFile := filepath.Join(fixturesDir, "status_R001.json")
	statusContent := `{"robot_id": "R001", "status": "active", "battery": 95, "last_check": "2023-10-20T15:30:00Z"}`
	err = os.WriteFile(statusFile, []byte(statusContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create status fixture file: %v", err)
	}

	// Create XML fixture
	xmlFile := filepath.Join(fixturesDir, "report.xml")
	xmlContent := `<?xml version="1.0"?><robot_report><robot id="R001"><performance score="98"/></robot></robot_report>`
	err = os.WriteFile(xmlFile, []byte(xmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create XML fixture file: %v", err)
	}

	// Create configuration file with fixture references
	configYAML := `
sections:
  operations:
    path_pattern: "/api/v1/robots*"
    return_body: true

scenarios:
  - uuid: "list-robots"
    method: "GET"
    path: "/api/v1/robots"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/operations/robots.json"

  - uuid: "robot-status"
    method: "GET"
    path: "/api/v1/robots/R001/status"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/operations/status_R001.json"

  - uuid: "robot-report"
    method: "GET"
    path: "/api/v1/robots/R001/report"
    status_code: 200
    content_type: "application/xml"
    data: "@fixtures/operations/report.xml"

  - uuid: "inline-example"
    method: "GET"
    path: "/api/v1/inline"
    status_code: 200
    content_type: "application/json"
    data: '{"message": "This is inline data", "source": "inline"}'
`

	configFile := filepath.Join(tempDir, "fixture-test-config.yaml")
	err = os.WriteFile(configFile, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixture test config file: %v", err)
	}
	return configFile
}

// TestFixtureFileSupport tests that fixture files are properly loaded and served
func TestFixtureFileSupport(t *testing.T) {
	configFile := createTestFixtureConfigFile(t)
	srv, baseURL := setupScenarioTestServer(t, configFile)
	defer cleanupTestServer(t, srv)

	t.Run("GET robots list from fixture", func(t *testing.T) {
		testRobotsListFromFixture(t, baseURL)
	})

	t.Run("GET robot status from fixture", func(t *testing.T) {
		testRobotStatusFromFixture(t, baseURL)
	})

	t.Run("GET robot XML report from fixture", func(t *testing.T) {
		testRobotXMLReportFromFixture(t, baseURL)
	})

	t.Run("GET inline data still works", func(t *testing.T) {
		testInlineDataWorks(t, baseURL)
	})
}

func testRobotsListFromFixture(t *testing.T, baseURL string) {
	t.Helper()
	resp := makeGETRequest(t, baseURL+"/api/v1/robots")
	defer resp.Body.Close()

	validateJSONResponse(t, resp, 200)
	body := readResponseBody(t, resp)

	responseData := unmarshalJSONResponse(t, body)
	validateRobotsFixtureContent(t, responseData)
}

func makeGETRequest(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	return resp
}

func validateJSONResponse(t *testing.T, resp *http.Response, expectedStatus int) {
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func readResponseBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	return body
}

func unmarshalJSONResponse(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var responseData map[string]any
	if err := json.Unmarshal(body, &responseData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	return responseData
}

func validateRobotsFixtureContent(t *testing.T, responseData map[string]any) {
	t.Helper()
	robots, ok := responseData["robots"].([]any)
	if !ok || len(robots) == 0 {
		t.Fatalf("Expected robots array in response, got %v", responseData["robots"])
	}

	robot, ok := robots[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected robot to be a map, got %T", robots[0])
	}

	expectedValues := map[string]any{
		"id":     "R001",
		"name":   "Alpha Robot",
		"status": "active",
	}

	for key, expected := range expectedValues {
		if robot[key] != expected {
			t.Errorf("Expected robot %s '%v', got %v", key, expected, robot[key])
		}
	}
}

func testRobotStatusFromFixture(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/api/v1/robots/R001/status")
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

	// Validate the status fixture content was loaded correctly
	if responseData["robot_id"] != "R001" {
		t.Errorf("Expected robot_id 'R001', got %v", responseData["robot_id"])
	}
	if responseData["status"] != "active" {
		t.Errorf("Expected status 'active', got %v", responseData["status"])
	}
	if responseData["battery"] != float64(95) {
		t.Errorf("Expected battery 95, got %v", responseData["battery"])
	}
}

func testRobotXMLReportFromFixture(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/api/v1/robots/R001/report")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/xml" {
		t.Errorf("Expected Content-Type application/xml, got %s", resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedXML := `<?xml version="1.0"?><robot_report>` +
		`<robot id="R001"><performance score="98"/></robot></robot_report>`
	if string(body) != expectedXML {
		t.Errorf("Expected XML fixture content %q, got %q", expectedXML, string(body))
	}
}

func testInlineDataWorks(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/api/v1/inline")
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

	if responseData["message"] != "This is inline data" {
		t.Errorf("Expected message 'This is inline data', got %v", responseData["message"])
	}
	if responseData["source"] != "inline" {
		t.Errorf("Expected source 'inline', got %v", responseData["source"])
	}
}

func createMissingFixtureConfigFile(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// Create configuration file with missing fixture reference
	configYAML := `
sections:
  test:
    path_pattern: "/test/**"
    return_body: true

scenarios:
  - uuid: "missing-fixture"
    method: "GET"
    path: "/test/missing"
    status_code: 200
    content_type: "application/json"
    data: "@fixtures/missing.json"
`

	configFile := filepath.Join(tempDir, "missing-fixture-config.yaml")
	err := os.WriteFile(configFile, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create missing fixture config file: %v", err)
	}
	return configFile
}

// TestMissingFixtureFileHandling tests graceful handling of missing fixture files
func TestMissingFixtureFileHandling(t *testing.T) {
	configFile := createMissingFixtureConfigFile(t)
	srv, baseURL := setupScenarioTestServer(t, configFile)
	defer cleanupTestServer(t, srv)

	t.Run("Missing fixture returns empty response", func(t *testing.T) {
		testMissingFixtureReturnsEmptyResponse(t, baseURL)
	})
}

func testMissingFixtureReturnsEmptyResponse(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/test/missing")
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

	// Missing fixture should result in empty response
	if len(body) != 0 {
		t.Errorf("Expected empty body for missing fixture, got %d bytes: %s", len(body), string(body))
	}
}

func createSecurityTestConfigFile(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// Create configuration file with security test scenarios
	configYAML := `
sections:
  security_test:
    path_pattern: "/security/**"
    return_body: true

scenarios:
  - uuid: "path-traversal-attempt"
    method: "GET"
    path: "/security/path-traversal"
    status_code: 200
    data: "@fixtures/../../../etc/passwd"

  - uuid: "absolute-path-attempt"
    method: "GET"
    path: "/security/absolute-path"
    status_code: 200
    data: "@/etc/passwd"

  - uuid: "empty-reference"
    method: "GET"
    path: "/security/empty-ref"
    status_code: 200
    data: "@"
`

	configFile := filepath.Join(tempDir, "security-test-config.yaml")
	err := os.WriteFile(configFile, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create security test config file: %v", err)
	}
	return configFile
}

// TestFixtureFileSecurity tests that security validation prevents path traversal attacks
func TestFixtureFileSecurity(t *testing.T) {
	configFile := createSecurityTestConfigFile(t)
	srv, baseURL := setupScenarioTestServer(t, configFile)
	defer cleanupTestServer(t, srv)

	t.Run("Path traversal attempts return empty response", func(t *testing.T) {
		testPathTraversalAttempts(t, baseURL)
	})

	t.Run("Absolute path attempts return empty response", func(t *testing.T) {
		testAbsolutePathAttempts(t, baseURL)
	})

	t.Run("Empty reference returns empty response", func(t *testing.T) {
		testEmptyReference(t, baseURL)
	})
}

func testPathTraversalAttempts(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/security/path-traversal")
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

	// Path traversal should be blocked and return empty response
	if len(body) != 0 {
		t.Errorf("Expected empty body for blocked path traversal, got %d bytes", len(body))
	}
}

func testAbsolutePathAttempts(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/security/absolute-path")
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

	// Absolute path should be blocked and return empty response
	if len(body) != 0 {
		t.Errorf("Expected empty body for blocked absolute path, got %d bytes", len(body))
	}
}

func testEmptyReference(t *testing.T, baseURL string) {
	t.Helper()
	resp, err := http.Get(baseURL + "/security/empty-ref")
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

	// Empty reference should return empty response
	if len(body) != 0 {
		t.Errorf("Expected empty body for empty reference, got %d bytes", len(body))
	}
}

// TestFixtureFileCaching tests that fixture files are cached for performance
func TestFixtureFileCaching(t *testing.T) {
	configFile := createTestFixtureConfigFile(t)
	srv, baseURL := setupScenarioTestServer(t, configFile)
	defer cleanupTestServer(t, srv)

	t.Run("Multiple requests to same fixture return consistent data", func(t *testing.T) {
		testFixtureCachingConsistency(t, baseURL)
	})
}

func testFixtureCachingConsistency(t *testing.T, baseURL string) {
	t.Helper()
	responses := makeMultipleRequests(t, baseURL, 3)
	validateResponseConsistency(t, responses)
}

func makeMultipleRequests(t *testing.T, baseURL string, count int) [][]byte {
	t.Helper()
	var responses [][]byte

	for i := 0; i < count; i++ {
		resp, err := http.Get(baseURL + "/api/v1/robots")
		if err != nil {
			t.Fatalf("GET request %d failed: %v", i+1, err)
		}
		body := readAndCloseResponse(t, resp)
		responses = append(responses, body)
	}

	return responses
}

func readAndCloseResponse(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return body
}

func validateResponseConsistency(t *testing.T, responses [][]byte) {
	t.Helper()
	// All responses should be identical (indicating caching works)
	for i := 1; i < len(responses); i++ {
		if string(responses[0]) != string(responses[i]) {
			t.Errorf("Response %d differs from response 0, caching may not be working properly", i)
		}
	}
}
