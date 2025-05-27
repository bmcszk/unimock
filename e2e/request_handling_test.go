//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/pkg/client"
	"github.com/bmcszk/unimock/pkg/model"
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
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	targetURLPath := "/test/resource/item123"
	expectedBody := `{"id": "item123", "data": "sample data"}`
	expectedContentType := "application/json"

	scenario := &model.Scenario{
		RequestPath: "GET " + targetURLPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedContentType,
		Data:        expectedBody,
	}

	createdScenario, err := unimockClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario")
	require.NotNil(t, createdScenario, "Created scenario should not be nil")
	require.NotEmpty(t, createdScenario.UUID, "Created scenario UUID should not be empty")

	t.Cleanup(func() {
		errDel := unimockClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, errDel, "Failed to delete scenario %s", createdScenario.UUID)
	})

	targetURL := unimockBaseURL + targetURLPath

	// Steps:
	// 1. Send a GET request to `/test/resource/item123`.
	resp, err := http.Get(targetURL)
	require.NoError(t, err, "Failed to send GET request")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 200 OK.
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK")

	// - Response body is `{\"id\": \"item123\", \"data\": \"sample data\"}`.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	assert.JSONEq(t, expectedBody, string(bodyBytes), "Response body does not match expected")

	// - Content-Type header is `application/json`.
	assert.Equal(t, expectedContentType, resp.Header.Get("Content-Type"), "Content-Type header does not match")
}

// TestSCEN_RH_002_PostCreateResource verifies SCEN-RH-002:
// Successful creation of a new resource via a POST request.
func TestSCEN_RH_002_PostCreateResource(t *testing.T) {
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	targetCollectionURLPath := "/test/collection"
	newItemID := "newItem"
	newItemData := `{"name": "New Item", "value": 42}`
	expectedLocationHeader := targetCollectionURLPath + "/" + newItemID
	expectedContentType := "application/json" // For the GET request after POST

	// Scenario for the POST request
	postScenario := &model.Scenario{
		RequestPath: "POST " + targetCollectionURLPath,
		StatusCode:  http.StatusCreated,
		ContentType: "application/json", // ContentType of the POST response (if any body)
		Location:    expectedLocationHeader,
		Data:        `{"id": "` + newItemID + `", "name": "New Item", "value": 42}`, // Body of POST response
	}
	createdPostScenario, err := unimockClient.CreateScenario(context.Background(), postScenario)
	require.NoError(t, err, "Failed to create POST scenario")
	require.NotNil(t, createdPostScenario, "Created POST scenario should not be nil")
	t.Cleanup(func() {
		_ = unimockClient.DeleteScenario(context.Background(), createdPostScenario.UUID)
	})

	// Scenario for the subsequent GET request to verify creation
	getScenario := &model.Scenario{
		RequestPath: "GET " + expectedLocationHeader,
		StatusCode:  http.StatusOK,
		ContentType: expectedContentType,
		Data:        `{"id": "` + newItemID + `", "name": "New Item", "value": 42}`,
	}
	createdGetScenario, err := unimockClient.CreateScenario(context.Background(), getScenario)
	require.NoError(t, err, "Failed to create GET scenario")
	require.NotNil(t, createdGetScenario, "Created GET scenario should not be nil")
	t.Cleanup(func() {
		_ = unimockClient.DeleteScenario(context.Background(), createdGetScenario.UUID)
	})

	targetURL := unimockBaseURL + targetCollectionURLPath

	// Steps:
	// 1. Send a POST request to `/test/collection` with body `{"name": "New Item", "value": 42}`.
	resp, err := http.Post(targetURL, "application/json", strings.NewReader(newItemData))
	require.NoError(t, err, "Failed to send POST request")
	defer resp.Body.Close()

	// Expected Result for POST:
	// - HTTP status code 201 Created.
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "HTTP status code should be 201 Created")
	// - Location header points to the new resource (e.g., `/test/collection/newItem`).
	assert.Equal(t, expectedLocationHeader, resp.Header.Get("Location"), "Location header does not match")

	// 2. Send a GET request to the URL from the Location header.
	getResp, err := http.Get(unimockBaseURL + resp.Header.Get("Location"))
	require.NoError(t, err, "Failed to send GET request to new resource URL")
	defer getResp.Body.Close()

	// Expected Result for GET:
	// - HTTP status code 200 OK.
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "GET request: HTTP status code should be 200 OK")
	// - Response body matches the data of the newly created item.
	bodyBytes, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "GET request: Failed to read response body")
	assert.JSONEq(t, `{"id": "`+newItemID+`", "name": "New Item", "value": 42}`, string(bodyBytes), "GET request: Response body does not match expected")
	// - Content-Type header is `application/json`.
	assert.Equal(t, expectedContentType, getResp.Header.Get("Content-Type"), "GET request: Content-Type header does not match")
}

// TestSCEN_RH_003_PutUpdateResource verifies SCEN-RH-003:
// Successful update of an existing resource via a PUT request.
func TestSCEN_RH_003_PutUpdateResource(t *testing.T) {
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	targetResourceURLPath := "/test/resource/itemToUpdate"
	updatedData := `{"id": "itemToUpdate", "status": "updated"}`
	expectedContentType := "application/json"

	// Scenario for the PUT request
	putScenario := &model.Scenario{
		RequestPath: "PUT " + targetResourceURLPath,
		StatusCode:  http.StatusOK, // Or 204 No Content if no body is returned
		ContentType: expectedContentType,
		Data:        updatedData, // Body of PUT response
	}
	createdPutScenario, err := unimockClient.CreateScenario(context.Background(), putScenario)
	require.NoError(t, err, "Failed to create PUT scenario")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdPutScenario.UUID) })

	// Scenario for the GET request after PUT to verify update
	getAfterPutScenario := &model.Scenario{
		RequestPath: "GET " + targetResourceURLPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedContentType,
		Data:        updatedData,
	}
	createdGetAfterPutScenario, err := unimockClient.CreateScenario(context.Background(), getAfterPutScenario)
	require.NoError(t, err, "Failed to create GET (after PUT) scenario")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdGetAfterPutScenario.UUID) })

	putTargetURL := unimockBaseURL + targetResourceURLPath

	// Steps:
	// 1. Send a PUT request to `/test/resource/itemToUpdate` with body `{"status": "updated"}`.
	req, err := http.NewRequest(http.MethodPut, putTargetURL, strings.NewReader(updatedData))
	require.NoError(t, err, "Failed to create PUT request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to send PUT request")
	defer resp.Body.Close()

	// Expected Result for PUT:
	// - HTTP status code 200 OK (or 204 No Content).
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK for PUT")
	bodyBytesPut, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read PUT response body")
	assert.JSONEq(t, updatedData, string(bodyBytesPut), "PUT response body does not match expected")

	// 2. Send a GET request to `/test/resource/itemToUpdate`.
	getResp, err := http.Get(unimockBaseURL + targetResourceURLPath)
	require.NoError(t, err, "Failed to send GET request after PUT")
	defer getResp.Body.Close()

	// Expected Result for GET:
	// - HTTP status code 200 OK.
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "GET after PUT: HTTP status code should be 200 OK")
	// - Response body reflects the updated content.
	bodyBytesGet, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "GET after PUT: Failed to read response body")
	assert.JSONEq(t, updatedData, string(bodyBytesGet), "GET after PUT: Response body does not match updated content")
	// - Content-Type header is `application/json`.
	assert.Equal(t, expectedContentType, getResp.Header.Get("Content-Type"), "GET after PUT: Content-Type header does not match")
}

// TestSCEN_RH_004_DeleteResource verifies SCEN-RH-004:
// Successful deletion of an existing resource via a DELETE request.
func TestSCEN_RH_004_DeleteResource(t *testing.T) {
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	targetResourceURLPath := "/test/resource/itemToDelete"

	// Scenario for the DELETE request
	deleteScenario := &model.Scenario{
		RequestPath: "DELETE " + targetResourceURLPath,
		StatusCode:  http.StatusNoContent, // Or http.StatusOK if a body is returned
	}
	createdDeleteScenario, err := unimockClient.CreateScenario(context.Background(), deleteScenario)
	require.NoError(t, err, "Failed to create DELETE scenario")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdDeleteScenario.UUID) })

	// Scenario for the GET request after DELETE to verify deletion (expects 404)
	getAfterDeleteScenario := &model.Scenario{
		RequestPath: "GET " + targetResourceURLPath,
		StatusCode:  http.StatusNotFound,
		ContentType: "text/plain", // Or whatever unimock returns for 404 by default for scenarios
		Data:        "Resource not found",
	}
	createdGetAfterDeleteScenario, err := unimockClient.CreateScenario(context.Background(), getAfterDeleteScenario)
	require.NoError(t, err, "Failed to create GET (after DELETE) scenario")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdGetAfterDeleteScenario.UUID) })

	deleteTargetURL := unimockBaseURL + targetResourceURLPath

	// Steps:
	// 1. Send a DELETE request to `/test/resource/itemToDelete`.
	req, err := http.NewRequest(http.MethodDelete, deleteTargetURL, nil)
	require.NoError(t, err, "Failed to create DELETE request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to send DELETE request")
	defer resp.Body.Close()

	// Expected Result for DELETE:
	// - HTTP status code 200 OK or 204 No Content.
	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, resp.StatusCode, "HTTP status code should be 200 OK or 204 No Content for DELETE, got %d", resp.StatusCode)

	// 2. Send a GET request to `/test/resource/itemToDelete`.
	getResp, err := http.Get(unimockBaseURL + targetResourceURLPath)
	require.NoError(t, err, "Failed to send GET request after DELETE")
	defer getResp.Body.Close()

	// Expected Result for GET:
	// - HTTP status code 404 Not Found.
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode, "GET after DELETE: HTTP status code should be 404 Not Found")
}

// TestSCEN_RH_005_GetIndividualResourceEndpoint verifies SCEN-RH-005:
// The application correctly uses Unimock for an individual resource endpoint (e.g., GET /mocks/{id}).
func TestSCEN_RH_005_GetIndividualResourceEndpoint(t *testing.T) {
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	mockedEndpointPath := "/mocks/specific-mock-id"
	expectedMockBody := `{"mockId": "specific-mock-id", "value": "This is a specific mock."}`
	expectedMockContentType := "application/json"

	// Scenario for the specific mock endpoint
	scenario := &model.Scenario{
		RequestPath: "GET " + mockedEndpointPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedMockContentType,
		Data:        expectedMockBody,
	}
	createdScenario, err := unimockClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario for specific mock endpoint")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdScenario.UUID) })

	targetURL := unimockBaseURL + mockedEndpointPath

	// Steps:
	// 1. Send a GET request to `/mocks/specific-mock-id`.
	resp, err := http.Get(targetURL)
	require.NoError(t, err, "Failed to send GET request to specific mock endpoint")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 200 OK.
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK")
	// - Response body matches the configured mock for this specific ID.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	assert.JSONEq(t, expectedMockBody, string(bodyBytes), "Response body does not match expected configured mock")
	// - Content-Type header matches the configured mock (e.g., `application/json`).
	assert.Equal(t, expectedMockContentType, resp.Header.Get("Content-Type"), "Content-Type header does not match expected")
}

// TestSCEN_RH_006_GetCollectionEndpoint verifies SCEN-RH-006:
// The application correctly uses Unimock for a collection endpoint (e.g., GET /mocks).
func TestSCEN_RH_006_GetCollectionEndpoint(t *testing.T) {
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	mockedCollectionPath := "/mocks"
	expectedCollectionBody := `[
		{"mockId": "mock1", "value": "First mock in collection"},
		{"mockId": "mock2", "value": "Second mock in collection"}
	]`
	expectedCollectionContentType := "application/json"

	// Scenario for the collection endpoint
	scenario := &model.Scenario{
		RequestPath: "GET " + mockedCollectionPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedCollectionContentType,
		Data:        expectedCollectionBody,
	}
	createdScenario, err := unimockClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario for collection endpoint")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdScenario.UUID) })

	targetURL := unimockBaseURL + mockedCollectionPath

	// Steps:
	// 1. Send a GET request to `/mocks`.
	resp, err := http.Get(targetURL)
	require.NoError(t, err, "Failed to send GET request to collection endpoint")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 200 OK.
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status code should be 200 OK for collection GET")
	// - Response body is an array of mock objects as configured in Unimock for the collection.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body for collection GET")
	assert.JSONEq(t, expectedCollectionBody, string(bodyBytes), "Response body for collection GET does not match expected array")
	// - Content-Type header is `application/json`.
	assert.Equal(t, expectedCollectionContentType, resp.Header.Get("Content-Type"), "Content-Type for collection GET should be application/json")
}

// TestSCEN_RH_007_PostInvalidContentType verifies SCEN-RH-007:
// Unimock rejects a POST request with an unsupported Content-Type header with a 415 status.
func TestSCEN_RH_007_PostInvalidContentType(t *testing.T) {
	// For this test, we are testing Unimock's *direct* handling of invalid Content-Type on its scenario management endpoint,
	// not scenario-based responses after a scenario is created. So, no specific scenario needs to be *matched* via the client.
	// We just need to ensure Unimock is running. The Makefile already handles this.

	// Let's test POSTing a new scenario with an invalid content type.
	scenarioCreationURL := unimockBaseURL + "/_uni/scenarios"
	scenarioPayload := `{"requestPath": "GET /foo", "statusCode": 200, "contentType": "text/plain", "data": "hello"}`

	resp, err := http.Post(scenarioCreationURL, "application/xml", strings.NewReader(scenarioPayload))
	require.NoError(t, err, "Failed to send POST request with invalid Content-Type")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 415 Unsupported Media Type.
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode, "HTTP status code should be 415 for unsupported Content-Type")
}

// TestSCEN_RH_008_GetNonExistentResource verifies SCEN-RH-008:
// A GET request for a resource not configured in Unimock returns a 404 Not Found.
func TestSCEN_RH_008_GetNonExistentResource(t *testing.T) {
	// No scenario needs to be created for this test, as we are verifying the 404
	// when no scenario matches. Unimock should be running (handled by Makefile).
	// unimockClient, err := client.NewClient(unimockBaseURL) // Client might be needed if we want to ensure no scenarios exist for path
	// require.NoError(t, err, "Failed to create unimock client")

	targetURLPath := "/test/non_existent_resource/random123"
	targetURL := unimockBaseURL + targetURLPath

	// Steps:
	// 1. Send a GET request to `/test/non_existent_resource/random123`.
	resp, err := http.Get(targetURL)
	require.NoError(t, err, "Failed to send GET request to non-existent resource")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 404 Not Found.
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "HTTP status code should be 404 for non-existent resource")
}

// TestSCEN_RH_009_PathBasedRouting verifies SCEN-RH-009:
// Unimock correctly routes requests to different mock configurations based on the request path.
func TestSCEN_RH_009_PathBasedRouting(t *testing.T) {
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	pathA := "/test/routing/pathA"
	dataA := "Response A"
	contentTypeA := "text/plain"

	pathB := "/test/routing/pathB"
	dataB := "Response B"
	contentTypeB := "text/plain" // Can be different, e.g., application/json

	// Scenario for Path A
	scenarioA := &model.Scenario{
		RequestPath: "GET " + pathA,
		StatusCode:  http.StatusOK,
		ContentType: contentTypeA,
		Data:        dataA,
	}
	createdScenarioA, err := unimockClient.CreateScenario(context.Background(), scenarioA)
	require.NoError(t, err, "Failed to create scenario for path A")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdScenarioA.UUID) })

	// Scenario for Path B
	scenarioB := &model.Scenario{
		RequestPath: "GET " + pathB,
		StatusCode:  http.StatusOK,
		ContentType: contentTypeB,
		Data:        dataB,
	}
	createdScenarioB, err := unimockClient.CreateScenario(context.Background(), scenarioB)
	require.NoError(t, err, "Failed to create scenario for path B")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdScenarioB.UUID) })

	// Steps & Expected Results:
	// 1. Send a GET request to `/test/routing/pathA`.
	respA, errA := http.Get(unimockBaseURL + pathA)
	require.NoError(t, errA, "/path/A: Failed to send GET request")
	defer respA.Body.Close()
	assert.Equal(t, http.StatusOK, respA.StatusCode, "/path/A: HTTP status code should be 200 OK")
	bodyA, _ := io.ReadAll(respA.Body)
	assert.Equal(t, dataA, string(bodyA), "/path/A: Response body does not match")
	assert.True(t, strings.HasPrefix(respA.Header.Get("Content-Type"), contentTypeA), "/path/A: Content-Type does not match, got %s", respA.Header.Get("Content-Type"))

	// 2. Send a GET request to `/test/routing/pathB`.
	respB, errB := http.Get(unimockBaseURL + pathB)
	require.NoError(t, errB, "/path/B: Failed to send GET request")
	defer respB.Body.Close()
	assert.Equal(t, http.StatusOK, respB.StatusCode, "/path/B: HTTP status code should be 200 OK")
	bodyB, _ := io.ReadAll(respB.Body)
	assert.Equal(t, dataB, string(bodyB), "/path/B: Response body does not match")
	assert.True(t, strings.HasPrefix(respB.Header.Get("Content-Type"), contentTypeB), "/path/B: Content-Type does not match, got %s", respB.Header.Get("Content-Type"))
}

// TestSCEN_RH_010_WildcardPathMatching verifies SCEN-RH-010:
// Unimock supports wildcard matching in request paths (e.g., /users/*).
// This test was added based on the new scenario SCEN-RH-010.
func TestSCEN_RH_010_WildcardPathMatching(t *testing.T) {
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	wildcardPathPattern := "/users/*"
	specificPath1 := "/users/user123/profile"
	specificPath2 := "/users/anotherUser/settings"
	nonMatchingPath := "/customers/data"

	responseData := `{"status": "matched by wildcard"}`
	responseContentType := "application/json"

	// Scenario with wildcard
	wildcardScenario := &model.Scenario{
		RequestPath: "GET " + wildcardPathPattern,
		StatusCode:  http.StatusOK,
		ContentType: responseContentType,
		Data:        responseData,
	}
	createdWildcardScenario, err := unimockClient.CreateScenario(context.Background(), wildcardScenario)
	require.NoError(t, err, "Failed to create wildcard scenario")
	t.Cleanup(func() { _ = unimockClient.DeleteScenario(context.Background(), createdWildcardScenario.UUID) })

	// Scenario for a non-matching path to ensure wildcard isn't overly greedy (optional, but good practice)
	// For this test, we'll rely on Unimock's default 404 if this specific scenario isn't hit.

	// Test case 1: Path matching wildcard
	resp1, err1 := http.Get(unimockBaseURL + specificPath1)
	require.NoError(t, err1, "Failed to GET %s", specificPath1)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode, "Status code for %s", specificPath1)
	body1, _ := io.ReadAll(resp1.Body)
	assert.JSONEq(t, responseData, string(body1), "Body for %s", specificPath1)

	// Test case 2: Another path matching wildcard
	resp2, err2 := http.Get(unimockBaseURL + specificPath2)
	require.NoError(t, err2, "Failed to GET %s", specificPath2)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode, "Status code for %s", specificPath2)
	body2, _ := io.ReadAll(resp2.Body)
	assert.JSONEq(t, responseData, string(body2), "Body for %s", specificPath2)

	// Test case 3: Path NOT matching wildcard (should 404)
	resp3, err3 := http.Get(unimockBaseURL + nonMatchingPath)
	require.NoError(t, err3, "Failed to GET %s", nonMatchingPath)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp3.StatusCode, "Status code for non-matching path %s should be 404", nonMatchingPath)
}

// TestSCEN_SH_001_ExactPathScenarioMatch verifies SCEN-SH-001:
// A configured scenario is matched by its exact RequestPath.
func TestSCEN_SH_001_ExactPathScenarioMatch(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	// - A scenario is configured with
	//   RequestPath: "GET /custom/scenario/exact",
	//   StatusCode:  200,
	//   ContentType: "application/json",
	//   Data:        "{\"message\": \"exact scenario matched\"}".
	targetMethod := http.MethodGet
	targetPath := "/custom/scenario/exact"
	expectedStatusCode := http.StatusOK
	expectedContentType := "application/json"
	expectedBody := `{"message": "exact scenario matched"}`

	scenario := &model.Scenario{
		RequestPath: targetMethod + " " + targetPath,
		StatusCode:  expectedStatusCode,
		ContentType: expectedContentType,
		Data:        expectedBody,
	}

	createdScenario, err := unimockClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario SCEN-SH-001")
	require.NotNil(t, createdScenario, "Created scenario SCEN-SH-001 should not be nil")
	require.NotEmpty(t, createdScenario.UUID, "Created scenario SCEN-SH-001 UUID should not be empty")

	t.Cleanup(func() {
		errDel := unimockClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, errDel, "Failed to delete scenario %s", createdScenario.UUID)
	})

	// Steps:
	// 1. Send a GET request to `/custom/scenario/exact`.
	requestURL := unimockBaseURL + targetPath
	req, err := http.NewRequest(targetMethod, requestURL, nil)
	require.NoError(t, err, "Failed to create GET request for SCEN-SH-001")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to send GET request for SCEN-SH-001")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 200 OK.
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "SCEN-SH-001: HTTP status code mismatch")

	// - Response body is `{\"message\": \"exact scenario matched\"}`.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "SCEN-SH-001: Failed to read response body")
	assert.JSONEq(t, expectedBody, string(bodyBytes), "SCEN-SH-001: Response body mismatch")

	// - Content-Type header is `application/json`.
	assert.Equal(t, expectedContentType, resp.Header.Get("Content-Type"), "SCEN-SH-001: Content-Type header mismatch")
}

// TestSCEN_SH_002_WildcardPathScenarioMatch verifies SCEN-SH-002:
// A configured scenario with a wildcard in RequestPath is matched.
func TestSCEN_SH_002_WildcardPathScenarioMatch(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	// - A scenario is configured with
	//   RequestPath: "POST /custom/scenario/*",
	//   StatusCode:  201,
	//   ContentType: "text/plain",
	//   Data:        "wildcard scenario matched".
	scenarioMethod := http.MethodPost
	scenarioPathPattern := "/custom/scenario/*"
	requestPathToMatch := "/custom/scenario/anything/here" // This path should be matched by the wildcard
	expectedStatusCode := http.StatusCreated
	expectedContentType := "text/plain"
	expectedBody := "wildcard scenario matched"

	scenario := &model.Scenario{
		RequestPath: scenarioMethod + " " + scenarioPathPattern,
		StatusCode:  expectedStatusCode,
		ContentType: expectedContentType,
		Data:        expectedBody,
	}

	createdScenario, err := unimockClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario SCEN-SH-002")
	require.NotNil(t, createdScenario, "Created scenario SCEN-SH-002 should not be nil")
	require.NotEmpty(t, createdScenario.UUID, "Created scenario SCEN-SH-002 UUID should not be empty")

	t.Cleanup(func() {
		errDel := unimockClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, errDel, "Failed to delete scenario %s", createdScenario.UUID)
	})

	// Steps:
	// 1. Send a POST request to `/custom/scenario/anything/here` with any body.
	requestURL := unimockBaseURL + requestPathToMatch
	postBody := strings.NewReader(`{"some_key": "some_value"}`)
	req, err := http.NewRequest(scenarioMethod, requestURL, postBody)
	require.NoError(t, err, "Failed to create POST request for SCEN-SH-002")
	req.Header.Set("Content-Type", "application/json") // Body content type, not expected response content type

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to send POST request for SCEN-SH-002")
	defer resp.Body.Close()

	// Expected Result:
	// - HTTP status code 201 Created.
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "SCEN-SH-002: HTTP status code mismatch")

	// - Response body is `wildcard scenario matched`.
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "SCEN-SH-002: Failed to read response body")
	assert.Equal(t, expectedBody, string(bodyBytes), "SCEN-SH-002: Response body mismatch")

	// - Content-Type header is `text/plain`.
	assert.Equal(t, expectedContentType, resp.Header.Get("Content-Type"), "SCEN-SH-002: Content-Type header mismatch")
}

// TestSCEN_SH_003_ScenarioSkipsMockHandling verifies SCEN-SH-003:
// If a scenario matches, normal mock resource handling for the same path is skipped.
func TestSCEN_SH_003_ScenarioSkipsMockHandling(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	targetMethod := http.MethodGet
	// Use a path that falls under a configured section in config.yaml (e.g., "users" -> "/api/users/*")
	mockResourceID := "override-put-id-e2e-003"
	mockResourcePath := "/api/users/" + mockResourceID // e.g. /api/users/override-put-id-e2e-003

	// 1. Create a regular mock resource at mockResourcePath by using PUT.
	// This resource should be overridden by the scenario.
	originalMockData := fmt.Sprintf(`{"id": "%s", "value": "this should be skipped"}`, mockResourceID)
	originalMockContentType := "application/json"

	putReqURL := unimockBaseURL + mockResourcePath
	putReq, err := http.NewRequest(http.MethodPut, putReqURL, strings.NewReader(originalMockData))
	require.NoError(t, err, "Failed to create PUT request for original mock data SCEN-SH-003")
	putReq.Header.Set("Content-Type", originalMockContentType)

	putResp, err := http.DefaultClient.Do(putReq)
	require.NoError(t, err, "Failed to PUT original mock data SCEN-SH-003")
	// For PUT to /users/{id}, if it creates or updates, it should be 200 OK.
	// The previous error was 404 because /override/path had no section.
	// Now /users/override-put-id-003 should be handled by the "users" section.
	require.True(t, putResp.StatusCode == http.StatusOK || putResp.StatusCode == http.StatusCreated, "PUT original mock data should be successful (200 or 201), got %d", putResp.StatusCode)
	defer putResp.Body.Close()

	// 2. Configure a scenario that matches the same GET request path (mockResourcePath) and method (targetMethod).
	scenarioRequestPath := fmt.Sprintf("%s %s", targetMethod, mockResourcePath) // e.g. "GET /users/override-put-id-003"
	scenarioStatusCode := http.StatusTeapot
	scenarioContentType := "application/vnd.custom.teapot"
	scenarioData := `{"message": "I am a teapot because of the scenario!"}`

	createdScenario, err := unimockClient.CreateScenario(context.Background(), &model.Scenario{
		RequestPath: scenarioRequestPath,
		StatusCode:  scenarioStatusCode,
		ContentType: scenarioContentType,
		Data:        scenarioData,
	})
	require.NoError(t, err, "Failed to create scenario for SCEN-SH-003")
	require.NotNil(t, createdScenario, "Created scenario should not be nil for SCEN-SH-003")

	// Ensure scenario is deleted even if test fails midway
	defer func() {
		err := unimockClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, err, "Failed to cleanup (delete) scenario %s for SCEN-SH-003", createdScenario.UUID)
	}()

	// Steps:
	// 1. Send a GET request to mockResourcePath.
	getReqURL := unimockBaseURL + mockResourcePath
	getReq, err := http.NewRequest(targetMethod, getReqURL, nil)
	require.NoError(t, err, "Failed to create GET request for SCEN-SH-003")

	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err, "Failed to execute GET request for SCEN-SH-003")
	defer getResp.Body.Close()

	// Expected Result:
	// - HTTP Status Code should be from the SCENARIO (418 I'm a teapot).
	// - Content-Type header should be from the SCENARIO.
	// - Response body should be from the SCENARIO.
	bodyBytes, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "Failed to read response body for SCEN-SH-003")

	assert.Equal(t, scenarioStatusCode, getResp.StatusCode, "SCEN-SH-003: HTTP status code mismatch, scenario should override")
	assert.Equal(t, scenarioContentType, getResp.Header.Get("Content-Type"), "SCEN-SH-003: Content-Type header mismatch, scenario should override")
	assert.JSONEq(t, scenarioData, string(bodyBytes), "SCEN-SH-003: Response body mismatch, scenario should override")

	// Additional verification: ensure the original mock data was not returned.
	assert.NotEqual(t, originalMockContentType, getResp.Header.Get("Content-Type"), "SCEN-SH-003: Content-Type should not match original mock")
	assert.NotEqual(t, originalMockData, string(bodyBytes), "SCEN-SH-003: Body should not match original mock (raw string comparison)")

	// Cleanup: Delete the mock resource created at the start. The scenario is cleaned by defer.
	// This is important if the scenario didn't match and the original resource was served.
	delReqURL := unimockBaseURL + mockResourcePath
	delReq, err := http.NewRequest(http.MethodDelete, delReqURL, nil)
	require.NoError(t, err, "Failed to create DELETE request for cleanup SCEN-SH-003")
	delResp, err := http.DefaultClient.Do(delReq)
	require.NoError(t, err, "Failed to execute DELETE request for cleanup SCEN-SH-003")
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode, "Cleanup DELETE of mock resource failed for SCEN-SH-003")
	defer delResp.Body.Close()
}

// TestSCEN_SH_004_ScenarioMethodMismatch verifies SCEN-SH-004:
// A scenario for a specific HTTP method does not match requests with other methods on the same path.
func TestSCEN_SH_004_ScenarioMethodMismatch(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	targetPath := "/api/test/method-specific" // Using /api/... to match config.yaml sections if needed
	getMethod := http.MethodGet
	postMethod := http.MethodPost

	// - A scenario is configured for GET /api/test/method-specific
	getScenarioData := "GET scenario response for SCEN-SH-004"
	getScenarioStatusCode := http.StatusOK
	getScenarioContentType := "text/plain"

	getScenario := &model.Scenario{
		RequestPath: fmt.Sprintf("%s %s", getMethod, targetPath),
		StatusCode:  getScenarioStatusCode,
		ContentType: getScenarioContentType,
		Data:        getScenarioData,
	}
	createdGetScenario, err := unimockClient.CreateScenario(context.Background(), getScenario)
	require.NoError(t, err, "Failed to create GET scenario for SCEN-SH-004")
	require.NotNil(t, createdGetScenario, "Created GET scenario should not be nil for SCEN-SH-004")
	t.Cleanup(func() {
		unimockClient.DeleteScenario(context.Background(), createdGetScenario.UUID)
	})

	// Steps:
	// 1. Send a GET request to targetPath.
	getReqURL := unimockBaseURL + targetPath
	getReq, err := http.NewRequest(getMethod, getReqURL, nil)
	require.NoError(t, err, "Failed to create GET request for SCEN-SH-004")
	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err, "Failed to execute GET request for SCEN-SH-004")
	defer getResp.Body.Close()

	// Expected Result for GET:
	assert.Equal(t, getScenarioStatusCode, getResp.StatusCode, "SCEN-SH-004 GET: HTTP status code mismatch")
	bodyBytesGet, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "SCEN-SH-004 GET: Failed to read response body")
	assert.Equal(t, getScenarioData, string(bodyBytesGet), "SCEN-SH-004 GET: Response body mismatch")
	assert.Equal(t, getScenarioContentType, getResp.Header.Get("Content-Type"), "SCEN-SH-004 GET: Content-Type mismatch")

	// 2. Send a POST request to targetPath.
	postReqURL := unimockBaseURL + targetPath
	postBody := strings.NewReader(`{"data": "some post data"}`)
	postReq, err := http.NewRequest(postMethod, postReqURL, postBody)
	require.NoError(t, err, "Failed to create POST request for SCEN-SH-004")
	postReq.Header.Set("Content-Type", "application/json")

	postResp, err := http.DefaultClient.Do(postReq)
	require.NoError(t, err, "Failed to execute POST request for SCEN-SH-004")
	defer postResp.Body.Close()

	// Expected Result for POST:
	// Should be 404 Not Found or 405 Method Not Allowed as no scenario/mock for POST exists.
	// The mock handler typically returns 404 if no section matches or if the method isn't handled by a matched section's logic.
	// It returns 405 if a section matches but the specific method has no implementation (default case in switch).
	// Given that /api/test/method-specific likely won't match a section in config.yaml, a 404 from the router or early in HandleRequest is expected.
	assert.Equal(t, http.StatusNotFound, postResp.StatusCode, "SCEN-SH-004 POST: Expected 404 Not Found as no POST scenario/mock exists")
}

// TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation verifies SCEN-SH-005:
// A scenario can return an empty body, a specific status code, and a custom Location header.
func TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation(t *testing.T) {
	// Preconditions:
	// - Unimock service is running.
	unimockClient, err := client.NewClient(unimockBaseURL)
	require.NoError(t, err, "Failed to create unimock client")

	targetPath := "/api/actions/submit-task" // Using /api/... for consistency with config.yaml if sections matter
	requestMethod := http.MethodPost

	// - Scenario configuration
	expectedStatusCode := http.StatusCreated
	expectedLocationHeader := "/tasks/status/new-task-123"
	expectedData := ""
	// ContentType can be empty or omitted for no-body responses; the scenario handler should cope.
	// If a ContentType *is* set in the scenario (e.g. "text/plain"), it should be returned.
	// If not set in scenario, observe actual Content-Type (might be none, or a default by http library/handler).
	// For this test, we will explicitly set it to empty in the scenario and assert it's not set or is a default.
	expectedContentType := "" // Or a specific one like "text/plain" if that's the default unimock sends

	scenario := &model.Scenario{
		RequestPath: fmt.Sprintf("%s %s", requestMethod, targetPath),
		StatusCode:  expectedStatusCode,
		ContentType: expectedContentType, // Explicitly empty
		Location:    expectedLocationHeader,
		Data:        expectedData,
	}
	createdScenario, err := unimockClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario for SCEN-SH-005")
	require.NotNil(t, createdScenario, "Created scenario should not be nil for SCEN-SH-005")
	t.Cleanup(func() {
		unimockClient.DeleteScenario(context.Background(), createdScenario.UUID)
	})

	// Steps:
	// 1. Send a POST request to targetPath.
	reqURL := unimockBaseURL + targetPath
	reqBody := strings.NewReader(`{"action": "start"}`)
	req, err := http.NewRequest(requestMethod, reqURL, reqBody)
	require.NoError(t, err, "Failed to create POST request for SCEN-SH-005")
	req.Header.Set("Content-Type", "application/json") // Request body type

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to execute POST request for SCEN-SH-005")
	defer resp.Body.Close()

	// Expected Result:
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "SCEN-SH-005: HTTP status code mismatch")
	assert.Equal(t, expectedLocationHeader, resp.Header.Get("Location"), "SCEN-SH-005: Location header mismatch")

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "SCEN-SH-005: Failed to read response body")
	assert.Equal(t, expectedData, string(bodyBytes), "SCEN-SH-005: Response body should be empty")

	// Check Content-Type: if scenario specified empty, it might be omitted by server or a default might be set.
	// For this test, if scenario.ContentType was "", an absent Content-Type header is ideal for 201/204.
	// However, http.ResponseWriter might set a default if `Write()` is called, even with 0 bytes.
	// If `HandleRequest` returns a body (even empty string) and ContentType in scenario is "", then handler sets CT. If scenario CT is empty and Data is empty string, it's possible no CT is set by `HandleRequest`'s resp.Header.
	// `ServeHTTP` sets Content-Length if not set. It doesn't add CT if missing.
	actualContentType := resp.Header.Get("Content-Type")
	if expectedContentType == "" {
		// Allow empty or a common default like "text/plain; charset=utf-8" which Go's HTTP server might add
		assert.True(t, actualContentType == "" || strings.HasPrefix(actualContentType, "text/plain"),
			"SCEN-SH-005: Content-Type should be empty or a default text/plain, got '%s'", actualContentType)
	} else {
		assert.Equal(t, expectedContentType, actualContentType, "SCEN-SH-005: Content-Type mismatch")
	}
}
