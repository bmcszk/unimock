//go:build e2e

package e2e

import (
	"io"
	"net/http"
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
