//go:build e2e

package e2e

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: This assumes Unimock is running at http://localhost:8080
// and has been pre-configured with the required mock.
// Future tasks should address proper Unimock setup/teardown and dynamic configuration for tests.
const unimockBaseURL = "http://localhost:8080"

// TestSCEN_RH_001_GetExistingResource verifies SCEN-RH-001:
// Successful processing of a GET request for an existing individual resource.
func TestSCEN_RH_001_GetExistingResource(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	// - A mock resource is configured at `/test/resource/item123`
	//   with content `{"id": "item123", "data": "sample data"}`
	//   and Content-Type `application/json`.

	targetURL := unimockBaseURL + "/test/resource/item123"
	expectedBody := `{"id": "item123", "data": "sample data"}`
	expectedContentType := "application/json"

	// Steps:
	// 1. Send a GET request to `/test/resource/item123`.
	resp, err := http.Get(targetURL)
	require.NoError(t, err, "Failed to send GET request")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 200 OK.
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK")

	// - Response body is `{"id": "item123", "data": "sample data"}`.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	assert.JSONEq(t, expectedBody, string(bodyBytes), "Response body does not match expected")

	// - Content-Type header is `application/json`.
	assert.Equal(t, expectedContentType, resp.Header.Get("Content-Type"), "Content-Type header does not match")
}

// TestSCEN_RH_002_PostCreateResource verifies SCEN-RH-002:
// Successful processing of a POST request to create a new resource.
func TestSCEN_RH_002_PostCreateResource(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	// - Unimock is configured to allow POST requests to `/test/collection`
	//   and extract ID from body path `/id`.

	postURL := unimockBaseURL + "/test/collection"
	requestBody := `{"id": "newItem", "value": "new data"}`
	expectedLocationHeader := "/test/collection/newItem"
	expectedGetBody := requestBody

	// Steps:
	// 1. Send a POST request to `/test/collection` with body
	//    `{"id": "newItem", "value": "new data"}` and Content-Type `application/json`.
	postResp, err := http.Post(postURL, "application/json", strings.NewReader(requestBody))
	require.NoError(t, err, "Failed to send POST request")
	defer postResp.Body.Close()

	// Expected Result (POST):
	// - HTTP status code 201 Created.
	assert.Equal(t, http.StatusCreated, postResp.StatusCode, "HTTP status code should be 201 Created")

	// - Location header is present and points to `/test/collection/newItem`.
	locationHeader := postResp.Header.Get("Location")
	assert.Equal(t, expectedLocationHeader, locationHeader, "Location header does not match")

	// Verify resource creation by sending a GET request to the Location URL
	getURL := unimockBaseURL + locationHeader // Assuming Location header is a relative path
	getResp, err := http.Get(getURL)
	require.NoError(t, err, "Failed to send GET request to verify resource creation")
	defer getResp.Body.Close()

	// Expected Result (GET):
	// - HTTP status code 200 OK.
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "GET request: HTTP status code should be 200 OK")

	// - Response body is `{"id": "newItem", "value": "new data"}`.
	bodyBytes, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "GET request: Failed to read response body")
	assert.JSONEq(t, expectedGetBody, string(bodyBytes), "GET request: Response body does not match expected")

	// - Content-Type header is `application/json`.
	assert.Equal(t, "application/json", getResp.Header.Get("Content-Type"), "GET request: Content-Type header does not match")
}
