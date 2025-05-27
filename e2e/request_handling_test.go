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

// TestSCEN_RH_003_PutUpdateResource verifies SCEN-RH-003:
// Successful processing of a PUT request to update an existing resource.
func TestSCEN_RH_003_PutUpdateResource(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	// - A mock resource `{"id": "existingItem", "value": "old data"}` exists at `/test/collection/existingItem`.
	//   (This might require a setup step to create/ensure this resource exists before the PUT)
	// - Unimock is configured to allow PUT requests to `/test/collection/{id}`.

	resourceID := "existingItem"
	putURL := unimockBaseURL + "/test/collection/" + resourceID
	updatedRequestBody := `{"id": "existingItem", "value": "updated data"}`

	// Step 1: Send a PUT request to `/test/collection/existingItem`
	client := &http.Client{}
	putReq, err := http.NewRequest(http.MethodPut, putURL, strings.NewReader(updatedRequestBody))
	require.NoError(t, err, "Failed to create PUT request")
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := client.Do(putReq)
	require.NoError(t, err, "Failed to send PUT request")
	defer putResp.Body.Close()

	// Expected Result (PUT):
	// - HTTP status code 200 OK (or 204 No Content).
	// For this test, we'll assert for 200 OK. If 204 is also acceptable, the assertion can be adjusted.
	assert.Equal(t, http.StatusOK, putResp.StatusCode, "HTTP status code should be 200 OK for PUT")

	// Verify resource update by sending a GET request
	getResp, err := http.Get(putURL)
	require.NoError(t, err, "Failed to send GET request to verify resource update")
	defer getResp.Body.Close()

	// Expected Result (GET):
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "GET after PUT: HTTP status code should be 200 OK")

	bodyBytes, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "GET after PUT: Failed to read response body")
	assert.JSONEq(t, updatedRequestBody, string(bodyBytes), "GET after PUT: Response body does not match updated content")

	assert.Equal(t, "application/json", getResp.Header.Get("Content-Type"), "GET after PUT: Content-Type header does not match")
}

// TestSCEN_RH_004_DeleteResource verifies SCEN-RH-004:
// Successful processing of a DELETE request for an existing individual resource.
func TestSCEN_RH_004_DeleteResource(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	// - A mock resource exists at `/test/resource/itemToDelete`.
	//   (This might require a setup step to create/ensure this resource exists before the DELETE)

	resourceID := "itemToDelete"
	resourceURL := unimockBaseURL + "/test/resource/" + resourceID

	// Step 1: Send a DELETE request to `/test/resource/itemToDelete`.
	client := &http.Client{}
	delReq, err := http.NewRequest(http.MethodDelete, resourceURL, nil)
	require.NoError(t, err, "Failed to create DELETE request")

	delResp, err := client.Do(delReq)
	require.NoError(t, err, "Failed to send DELETE request")
	defer delResp.Body.Close()

	// Expected Result (DELETE):
	// - HTTP status code 200 OK or 204 No Content.
	// We will accept either. Other responses are a failure.
	assert.True(t, delResp.StatusCode == http.StatusOK || delResp.StatusCode == http.StatusNoContent,
		"HTTP status code should be 200 OK or 204 No Content for DELETE, got %d", delResp.StatusCode)

	// Verify resource deletion by sending a GET request
	getResp, err := http.Get(resourceURL)
	require.NoError(t, err, "Failed to send GET request to verify resource deletion")
	defer getResp.Body.Close()

	// Expected Result (GET after DELETE):
	// - Subsequent GET request to `/test/resource/itemToDelete` returns 404 Not Found.
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode, "GET after DELETE: HTTP status code should be 404 Not Found")
}

// TestSCEN_RH_005_GetIndividualResourceEndpoint verifies SCEN-RH-005:
// GET request for an individual resource endpoint.
func TestSCEN_RH_005_GetIndividualResourceEndpoint(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	// - A mock resource is configured at `/individual/item001`.
	//   (Response body and Content-Type will be assumed for this test)

	targetURL := unimockBaseURL + "/individual/item001"
	// Assuming a simple JSON response for the configured mock as it's not specified in the scenario
	expectedBody := `{"itemId": "item001", "description": "Individual item endpoint test"}`
	expectedContentType := "application/json"

	// Steps:
	// 1. Send a GET request to `/individual/item001`.
	resp, err := http.Get(targetURL)
	require.NoError(t, err, "Failed to send GET request")
	defer resp.Body.Close()

	// Expected Result:
	// - Service returns 200 OK with the configured mock response for `/individual/item001`.
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK")

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	assert.JSONEq(t, expectedBody, string(bodyBytes), "Response body does not match expected configured mock")

	assert.Equal(t, expectedContentType, resp.Header.Get("Content-Type"), "Content-Type header does not match expected")
}

// TestSCEN_RH_006_GetCollectionEndpoint verifies SCEN-RH-006:
// GET request for a collection endpoint.
func TestSCEN_RH_006_GetCollectionEndpoint(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	// - Multiple mock resources are configured under `/collection/items/`.
	//   For this test, assume `/collection/items/itemA` and `/collection/items/itemB` are mocked.
	//   Mock for /collection/items/itemA: `{"id": "itemA", "type": "gadget"}`
	//   Mock for /collection/items/itemB: `{"id": "itemB", "type": "widget"}`

	targetURL := unimockBaseURL + "/collection/items"

	// Expected: A JSON array of the items in the collection.
	// The exact format (e.g. full objects vs links) depends on Unimock's implementation detail for collection GETs.
	// For this test, we assume it returns an array of the full JSON bodies of the items.
	expectedBody := `[
		{"id": "itemA", "type": "gadget"},
		{"id": "itemB", "type": "widget"}
	]`
	// Note: Order in the array might not be guaranteed by the system,
	// so a more robust check might be needed if order is not deterministic (e.g. unmarshal and compare sets).
	// For now, assert.JSONEq handles potential reordering of keys within objects, but not array element order.

	// Steps:
	// 1. Send a GET request to `/collection/items`.
	resp, err := http.Get(targetURL)
	require.NoError(t, err, "Failed to send GET request to collection endpoint")
	defer resp.Body.Close()

	// Expected Result:
	// - Service returns 200 OK.
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK for collection GET")

	// - Response body is a JSON array containing representations of resources.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body for collection GET")
	assert.JSONEq(t, expectedBody, string(bodyBytes), "Response body for collection GET does not match expected array")

	// - Content-Type header is `application/json` (assuming, as it's a JSON array).
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type for collection GET should be application/json")
}

// TestSCEN_RH_007_PostInvalidContentType verifies SCEN-RH-007:
// Service rejects a POST request with an unsupported or invalid Content-Type header.
func TestSCEN_RH_007_PostInvalidContentType(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	// - A specific endpoint `/restricted_post` is configured to only accept `application/json` for POST requests.
	//   (This configuration needs to be part of the Unimock setup for this test to be meaningful)

	targetURL := unimockBaseURL + "/restricted_post"
	xmlBody := `<payload><data>test</data></payload>`

	// Steps:
	// 1. Send a POST request to `/restricted_post` with Content-Type `application/xml`.
	resp, err := http.Post(targetURL, "application/xml", strings.NewReader(xmlBody))
	require.NoError(t, err, "Failed to send POST request with invalid Content-Type")
	defer resp.Body.Close()

	// Expected Result:
	// - Service returns an HTTP 415 Unsupported Media Type status code.
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode, "HTTP status code should be 415 for unsupported Content-Type")
}
