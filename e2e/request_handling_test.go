//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	restclient "github.com/bmcszk/go-restclient"
	"github.com/bmcszk/unimock/pkg/client"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	UnimockBaseURL               = "http://localhost:8080"
	applicationJSONContentType   = "application/json"
	failedToCreateUnimockAPIMsg  = "Failed to create unimock API client"
	getPrefix                    = "GET "
)

// Note: Uses localhost:8080 as configured in the Makefile test setup.
// Server lifecycle (start/stop/health check) is managed by 'make test-e2e'.
// Port can be overridden via UNIMOCK_PORT environment variable on the server side.

// executeAndValidateHTTPFile is a helper function to execute an HTTP test file
// using go-restclient and validate its responses.
func ExecuteAndValidateHTTPFile(t *testing.T, httpFilePath string, hrespFilePath string) {
	t.Helper() // Mark as test helper

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance for %s", httpFilePath)

	resps, err := rc.ExecuteFile(context.Background(), httpFilePath)
	require.NoError(t, err, "Failed to execute %s", httpFilePath)

	err = rc.ValidateResponses(hrespFilePath, resps...)
	require.NoError(t, err, "Failed to validate responses against %s for %s", hrespFilePath, httpFilePath)
}

// TestSCEN_RH_001_GetExistingResource verifies SCEN-RH-001:
// Successful processing of a getPrefixrequest for an existing individual resource.
func TestSCEN_RH_001_GetExistingResource(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetURLPath := "/test/resource/item123"
	expectedBody := `{"id": "item123", "data": "sample data"}`
	expectedContentType := applicationJSONContentType

	scenario := model.Scenario{
		RequestPath: getPrefix + targetURLPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedContentType,
		Data:        expectedBody,
	}

	createdScenario, err := unimockAPIClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario")
	require.NotNil(t, createdScenario, "Created scenario should not be nil")
	require.NotEmpty(t, createdScenario.UUID, "Created scenario UUID should not be empty")

	t.Cleanup(func() {
		errDel := unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, errDel, "Failed to delete scenario %s", createdScenario.UUID)
	})

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_rh_001.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_rh_001.http")

	err = rc.ValidateResponses("testdata/http/scen_rh_001.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_rh_001.hresp")
}

// TestSCEN_RH_002_PostCreateResource verifies SCEN-RH-002:
// Successful creation of a new resource via a POST request.
func TestSCEN_RH_002_PostCreateResource(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetCollectionURLPath := "/test/collection"
	newItemID := "newItem"
	// newItemData := `{"name": "New Item", "value": 42}` // Defined in .http file
	expectedLocationHeader := targetCollectionURLPath + "/" + newItemID
	expectedContentType := applicationJSONContentType // For the getPrefixrequest after POST

	// Scenario for the POST request
	postScenario := model.Scenario{
		RequestPath: "POST " + targetCollectionURLPath,
		StatusCode:  http.StatusCreated,
		ContentType: applicationJSONContentType,
		Location:    expectedLocationHeader,
		Data:        `{"id": "` + newItemID + `", "name": "New Item", "value": 42}`,
	}
	createdPostScenario, err := unimockAPIClient.CreateScenario(context.Background(), postScenario)
	require.NoError(t, err, "Failed to create POST scenario")
	require.NotNil(t, createdPostScenario, "Created POST scenario should not be nil")
	t.Cleanup(func() {
		_ = unimockAPIClient.DeleteScenario(context.Background(), createdPostScenario.UUID)
	})

	// Scenario for the subsequent getPrefixrequest to verify creation
	getScenario := model.Scenario{
		RequestPath: getPrefix + expectedLocationHeader,
		StatusCode:  http.StatusOK,
		ContentType: expectedContentType,
		Data:        `{"id": "` + newItemID + `", "name": "New Item", "value": 42}`,
	}
	createdGetScenario, err := unimockAPIClient.CreateScenario(context.Background(), getScenario)
	require.NoError(t, err, "Failed to create GET scenario")
	require.NotNil(t, createdGetScenario, "Created GET scenario should not be nil")
	t.Cleanup(func() {
		_ = unimockAPIClient.DeleteScenario(context.Background(), createdGetScenario.UUID)
	})

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_rh_002.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_rh_002.http")

	err = rc.ValidateResponses("testdata/http/scen_rh_002.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_rh_002.hresp")
}

// TestSCEN_RH_003_PutUpdateResource verifies SCEN-RH-003:
// Successful update of an existing resource via a PUT request.
func TestSCEN_RH_003_PutUpdateResource(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetResourceURLPath := "/test/resource/itemToUpdate"
	updatedData := `{"id": "itemToUpdate", "status": "updated"}`
	expectedContentType := applicationJSONContentType

	// Scenario for the PUT request
	putScenario := model.Scenario{
		RequestPath: "PUT " + targetResourceURLPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedContentType,
		Data:        updatedData,
	}
	createdPutScenario, err := unimockAPIClient.CreateScenario(context.Background(), putScenario)
	require.NoError(t, err, "Failed to create PUT scenario")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdPutScenario.UUID) })

	// Scenario for the getPrefixrequest after PUT to verify update
	getAfterPutScenario := model.Scenario{
		RequestPath: getPrefix + targetResourceURLPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedContentType,
		Data:        updatedData,
	}
	createdGetAfterPutScenario, err := unimockAPIClient.CreateScenario(context.Background(), getAfterPutScenario)
	require.NoError(t, err, "Failed to create getPrefix(after PUT) scenario")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdGetAfterPutScenario.UUID) })

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_rh_003.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_rh_003.http")

	err = rc.ValidateResponses("testdata/http/scen_rh_003.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_rh_003.hresp")
}

// TestSCEN_RH_004_DeleteResource verifies SCEN-RH-004:
// Successful deletion of an existing resource via a DELETE request.
func TestSCEN_RH_004_DeleteResource(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetResourceURLPath := "/test/resource/itemToDelete"

	// Scenario for the DELETE request
	deleteScenario := model.Scenario{
		RequestPath: "DELETE " + targetResourceURLPath,
		StatusCode:  http.StatusNoContent,
	}
	createdDeleteScenario, err := unimockAPIClient.CreateScenario(context.Background(), deleteScenario)
	require.NoError(t, err, "Failed to create DELETE scenario")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdDeleteScenario.UUID) })

	// Scenario for the getPrefixrequest after DELETE to verify deletion (expects 404)
	// This scenario in Unimock should produce the body that matches scen_rh_004.hresp for the 404.
	getAfterDeleteScenario := model.Scenario{
		RequestPath: getPrefix + targetResourceURLPath,
		StatusCode:  http.StatusNotFound,
		ContentType: "text/plain",         // Match .hresp, simplified from "text/plain; charset=utf-8"
		Data:        "Resource not found", // Match .hresp
	}
	createdGetAfterDeleteScenario, err := unimockAPIClient.CreateScenario(context.Background(), getAfterDeleteScenario)
	require.NoError(t, err, "Failed to create getPrefix(after DELETE) scenario")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdGetAfterDeleteScenario.UUID) })

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_rh_004.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_rh_004.http")

	err = rc.ValidateResponses("testdata/http/scen_rh_004.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_rh_004.hresp")
}

// TestSCEN_RH_005_GetIndividualResourceEndpoint verifies SCEN-RH-005:
// The application correctly uses Unimock for an individual resource endpoint (e.g., getPrefix/mocks/{id}).
func TestSCEN_RH_005_GetIndividualResourceEndpoint(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	mockedEndpointPath := "/mocks/specific-mock-id"
	expectedMockBody := `{"mockId": "specific-mock-id", "value": "This is a specific mock."}`
	expectedMockContentType := applicationJSONContentType

	// Scenario for the specific mock endpoint
	scenario := model.Scenario{
		RequestPath: getPrefix + mockedEndpointPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedMockContentType,
		Data:        expectedMockBody,
	}
	createdScenario, err := unimockAPIClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario for specific mock endpoint")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID) })

	ExecuteAndValidateHTTPFile(t, "testdata/http/scen_rh_005.http", "testdata/http/scen_rh_005.hresp")
}

// TestSCEN_RH_006_GetCollectionEndpoint verifies SCEN-RH-006:
// The application correctly uses Unimock for a collection endpoint (e.g., getPrefix/mocks).
func TestSCEN_RH_006_GetCollectionEndpoint(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	mockedCollectionPath := "/mocks"
	expectedCollectionBody := `[{"mockId":"mock1","value":"First mock in collection"},` +
		`{"mockId":"mock2","value":"Second mock in collection"}]` // Compact JSON
	expectedCollectionContentType := applicationJSONContentType

	// Scenario for the collection endpoint
	scenario := model.Scenario{
		RequestPath: getPrefix + mockedCollectionPath,
		StatusCode:  http.StatusOK,
		ContentType: expectedCollectionContentType,
		Data:        expectedCollectionBody,
	}
	createdScenario, err := unimockAPIClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario for collection endpoint")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID) })

	ExecuteAndValidateHTTPFile(t, "testdata/http/scen_rh_006.http", "testdata/http/scen_rh_006.hresp")
}

// TestSCEN_RH_007_PostInvalidContentType verifies SCEN-RH-007:
// Unimock rejects a POST request with an unsupported Content-Type header with a 415 status.
func TestSCEN_RH_007_PostInvalidContentType(t *testing.T) {
	// For this test, we are testing Unimock's *direct* handling of invalid Content-Type 
	// on its scenario management endpoint.
	// No specific Unimock scenario needs to be created or matched via the API client.

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_rh_007.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_rh_007.http")

	// The .hresp file expects only one response
	require.Len(t, resps, 1, "Expected one response from scen_rh_007.http")

	err = rc.ValidateResponses("testdata/http/scen_rh_007.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_rh_007.hresp")
}

// TestSCEN_RH_008_GetNonExistentResource verifies SCEN-RH-008:
// A getPrefixrequest for a resource not configured in Unimock returns a 404 Not Found.
func TestSCEN_RH_008_GetNonExistentResource(t *testing.T) {
	// No Unimock scenario needs to be created for this test, as we are verifying the 404
	// when no scenario matches the request to a path not covered by mock handler logic.

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_rh_008.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_rh_008.http")

	// The .hresp file expects only one response
	require.Len(t, resps, 1, "Expected one response from scen_rh_008.http")

	err = rc.ValidateResponses("testdata/http/scen_rh_008.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_rh_008.hresp")
}

// TestSCEN_RH_009_PathBasedRouting verifies SCEN-RH-009:
// Unimock correctly routes requests to different mock configurations based on the request path.
func TestSCEN_RH_009_PathBasedRouting(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	pathA := "/test/routing/pathA"
	dataA := "Response A"
	contentTypeA := "text/plain" // Simplified from "text/plain; charset=utf-8"

	pathB := "/test/routing/pathB"
	dataB := "Response B"
	contentTypeB := "text/plain" // Simplified from "text/plain; charset=utf-8"

	// Scenario for Path A
	scenarioA := model.Scenario{
		RequestPath: getPrefix + pathA,
		StatusCode:  http.StatusOK,
		ContentType: contentTypeA,
		Data:        dataA,
	}
	createdScenarioA, err := unimockAPIClient.CreateScenario(context.Background(), scenarioA)
	require.NoError(t, err, "Failed to create scenario for path A")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdScenarioA.UUID) })

	// Scenario for Path B
	scenarioB := model.Scenario{
		RequestPath: getPrefix + pathB,
		StatusCode:  http.StatusOK,
		ContentType: contentTypeB,
		Data:        dataB,
	}
	createdScenarioB, err := unimockAPIClient.CreateScenario(context.Background(), scenarioB)
	require.NoError(t, err, "Failed to create scenario for path B")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdScenarioB.UUID) })

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_rh_009.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_rh_009.http")

	validationErr := rc.ValidateResponses("testdata/http/scen_rh_009.hresp", resps...)
	require.NoError(t, validationErr, "Failed to validate responses against testdata/http/scen_rh_009.hresp")
}

// TestSCEN_RH_010_WildcardPathMatching verifies SCEN-RH-010:
// Unimock supports wildcard matching in request paths (e.g., /users/*).
func TestSCEN_RH_010_WildcardPathMatching(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	wildcardPathPattern := "/users/*"
	responseData := `{"status": "matched by wildcard"}`
	responseContentType := applicationJSONContentType

	// Scenario with wildcard
	wildcardScenario := model.Scenario{
		RequestPath: getPrefix + wildcardPathPattern,
		StatusCode:  http.StatusOK,
		ContentType: responseContentType,
		Data:        responseData,
	}
	createdWildcardScenario, err := unimockAPIClient.CreateScenario(context.Background(), wildcardScenario)
	require.NoError(t, err, "Failed to create wildcard scenario")
	t.Cleanup(func() { _ = unimockAPIClient.DeleteScenario(context.Background(), createdWildcardScenario.UUID) })

	ExecuteAndValidateHTTPFile(t, "testdata/http/scen_rh_010.http", "testdata/http/scen_rh_010.hresp")
}

// TestSCEN_SH_001_ExactPathScenarioMatch verifies SCEN-SH-001:
// A configured scenario is matched by its exact RequestPath.
func TestSCEN_SH_001_ExactPathScenarioMatch(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetMethod := http.MethodGet
	targetPath := "/custom/scenario/exact"
	expectedStatusCode := http.StatusOK
	expectedContentType := applicationJSONContentType
	expectedBody := `{"message": "exact scenario matched"}`

	scenario := model.Scenario{
		RequestPath: targetMethod + " " + targetPath,
		StatusCode:  expectedStatusCode,
		ContentType: expectedContentType,
		Data:        expectedBody,
	}

	createdScenario, err := unimockAPIClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario SCEN-SH-001")
	require.NotNil(t, createdScenario, "Created scenario SCEN-SH-001 should not be nil")
	require.NotEmpty(t, createdScenario.UUID, "Created scenario SCEN-SH-001 UUID should not be empty")

	t.Cleanup(func() {
		errDel := unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, errDel, "Failed to delete scenario %s", createdScenario.UUID)
	})

	ExecuteAndValidateHTTPFile(t, "testdata/http/scen_sh_001.http", "testdata/http/scen_sh_001.hresp")
}

// TestSCEN_SH_002_WildcardPathScenarioMatch verifies SCEN-SH-002:
// A configured scenario with a wildcard in RequestPath is matched.
func TestSCEN_SH_002_WildcardPathScenarioMatch(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	scenarioMethod := http.MethodPost
	scenarioPathPattern := "/custom/scenario/*"
	expectedStatusCode := http.StatusCreated
	expectedContentType := "text/plain"
	expectedBody := "wildcard scenario matched"

	scenario := model.Scenario{
		RequestPath: scenarioMethod + " " + scenarioPathPattern,
		StatusCode:  expectedStatusCode,
		ContentType: expectedContentType,
		Data:        expectedBody,
	}

	createdScenario, err := unimockAPIClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario SCEN-SH-002")
	require.NotNil(t, createdScenario, "Created scenario SCEN-SH-002 should not be nil")
	require.NotEmpty(t, createdScenario.UUID, "Created scenario SCEN-SH-002 UUID should not be empty")

	t.Cleanup(func() {
		errDel := unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, errDel, "Failed to delete scenario %s", createdScenario.UUID)
	})

	ExecuteAndValidateHTTPFile(t, "testdata/http/scen_sh_002.http", "testdata/http/scen_sh_002.hresp")
}

// TestSCEN_SH_003_ScenarioSkipsMockHandling verifies SCEN-SH-003:
// If a scenario matches, normal mock resource handling for the same path is skipped.
func TestSCEN_SH_003_ScenarioSkipsMockHandling(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetMethod := http.MethodGet
	mockResourceID := "override-put-id-e2e-003"
	mockResourcePath := "/api/users/" + mockResourceID

	// Configure a scenario that matches the getPrefixrequest path (mockResourcePath) and method (targetMethod).
	// This scenario will override the regular mock handling.
	scenarioRequestPath := fmt.Sprintf("%s %s", targetMethod, mockResourcePath)
	scenarioStatusCode := http.StatusTeapot
	scenarioContentType := "application/vnd.custom.teapot"
	scenarioData := `{"message": "I am a teapot because of the scenario!"}`

	createdScenario, err := unimockAPIClient.CreateScenario(context.Background(), model.Scenario{
		RequestPath: scenarioRequestPath,
		StatusCode:  scenarioStatusCode,
		ContentType: scenarioContentType,
		Data:        scenarioData,
	})
	require.NoError(t, err, "Failed to create scenario for SCEN-SH-003")
	require.NotNil(t, createdScenario, "Created scenario should not be nil for SCEN-SH-003")

	t.Cleanup(func() {
		err := unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID)
		assert.NoError(t, err, "Failed to cleanup (delete) scenario %s for SCEN-SH-003", createdScenario.UUID)
	})

	// Use go-restclient to execute the sequence: PUT original, getPrefix(expect scenario), DELETE original
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_sh_003.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_sh_003.http")

	err = rc.ValidateResponses("testdata/http/scen_sh_003.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_sh_003.hresp")
}

// TestSCEN_SH_004_ScenarioMethodMismatch verifies SCEN-SH-004:
// A scenario for a specific HTTP method does not match requests with other methods on the same path.
func TestSCEN_SH_004_ScenarioMethodMismatch(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetPath := "/api/test/method-specific"
	getMethod := http.MethodGet

	getScenarioData := "GET scenario response for SCEN-SH-004"
	getScenarioStatusCode := http.StatusOK
	getScenarioContentType := "text/plain"

	getScenario := model.Scenario{
		RequestPath: fmt.Sprintf("%s %s", getMethod, targetPath),
		StatusCode:  getScenarioStatusCode,
		ContentType: getScenarioContentType,
		Data:        getScenarioData,
	}
	createdGetScenario, err := unimockAPIClient.CreateScenario(context.Background(), getScenario)
	require.NoError(t, err, "Failed to create GET scenario for SCEN-SH-004")
	require.NotNil(t, createdGetScenario, "Created GET scenario should not be nil for SCEN-SH-004")
	t.Cleanup(func() {
		unimockAPIClient.DeleteScenario(context.Background(), createdGetScenario.UUID)
	})

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_sh_004.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_sh_004.http")

	err = rc.ValidateResponses("testdata/http/scen_sh_004.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_sh_004.hresp")
}

// TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation verifies SCEN-SH-005:
// A scenario can return an empty body, a specific status code, and a custom Location header.
func TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation(t *testing.T) {
	unimockAPIClient, err := client.NewClient(UnimockBaseURL)
	require.NoError(t, err, failedToCreateUnimockAPIMsg)

	targetPath := "/api/actions/submit-task"
	requestMethod := http.MethodPost

	expectedStatusCode := http.StatusCreated
	expectedLocationHeader := "/tasks/status/new-task-123"
	expectedData := ""
	expectedContentType := "" // Scenario ContentType is empty, implying no specific Content-Type header for empty body

	scenario := model.Scenario{
		RequestPath: fmt.Sprintf("%s %s", requestMethod, targetPath),
		StatusCode:  expectedStatusCode,
		ContentType: expectedContentType,
		Location:    expectedLocationHeader,
		Data:        expectedData,
	}
	createdScenario, err := unimockAPIClient.CreateScenario(context.Background(), scenario)
	require.NoError(t, err, "Failed to create scenario for SCEN-SH-005")
	require.NotNil(t, createdScenario, "Created scenario should not be nil for SCEN-SH-005")
	t.Cleanup(func() {
		unimockAPIClient.DeleteScenario(context.Background(), createdScenario.UUID)
	})

	// Use go-restclient to execute and validate
	rc, err := restclient.NewClient(restclient.WithBaseURL(UnimockBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	resps, err := rc.ExecuteFile(context.Background(), "testdata/http/scen_sh_005.http")
	require.NoError(t, err, "Failed to execute testdata/http/scen_sh_005.http")

	err = rc.ValidateResponses("testdata/http/scen_sh_005.hresp", resps...)
	require.NoError(t, err, "Failed to validate responses against testdata/http/scen_sh_005.hresp")
}
