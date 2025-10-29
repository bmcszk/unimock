package e2e_test

import (
	"net/http"
	"testing"

	"github.com/bmcszk/unimock/pkg/model"
)

// TestSCEN_RH_001_GetExistingResource verifies SCEN-RH-001:
// Successful processing of a GET request for an existing individual resource.
func TestSCEN_RH_001_GetExistingResource(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "GET /test/resource/item123",
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
		Data:        `{"id": "item123", "data": "sample data"}`,
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_001.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_001.hresp")
}

// TestSCEN_RH_002_PostCreateResource verifies SCEN-RH-002:
// Successful creation of a new resource via a POST request.
func TestSCEN_RH_002_PostCreateResource(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.
		a_scenario(model.Scenario{
			RequestPath: "POST /test/collection",
			StatusCode:  http.StatusCreated,
			ContentType: "application/json",
			Location:    "/test/collection/newItem",
			Data:        `{"id": "newItem", "name": "New Item", "value": 42}`,
		}).and().
		a_scenario(model.Scenario{
			RequestPath: "GET /test/collection/newItem",
			StatusCode:  http.StatusOK,
			ContentType: "application/json",
			Data:        `{"id": "newItem", "name": "New Item", "value": 42}`,
		})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_002.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_002.hresp")
}

// TestSCEN_RH_003_PutUpdateResource verifies SCEN-RH-003:
// Successful update of an existing resource via a PUT request.
func TestSCEN_RH_003_PutUpdateResource(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.
		a_scenario(model.Scenario{
			RequestPath: "PUT /test/resource/itemToUpdate",
			StatusCode:  http.StatusOK,
			ContentType: "application/json",
			Data:        `{"id": "itemToUpdate", "status": "updated"}`,
		}).and().
		a_scenario(model.Scenario{
			RequestPath: "GET /test/resource/itemToUpdate",
			StatusCode:  http.StatusOK,
			ContentType: "application/json",
			Data:        `{"id": "itemToUpdate", "status": "updated"}`,
		})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_003.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_003.hresp")
}

// TestSCEN_RH_004_DeleteResource verifies SCEN-RH-004:
// Successful deletion of an existing resource via a DELETE request.
func TestSCEN_RH_004_DeleteResource(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.
		a_scenario(model.Scenario{
			RequestPath: "DELETE /test/resource/itemToDelete",
			StatusCode:  http.StatusNoContent,
		}).and().
		a_scenario(model.Scenario{
			RequestPath: "GET /test/resource/itemToDelete",
			StatusCode:  http.StatusNotFound,
			ContentType: "text/plain",
			Data:        "Resource not found",
		})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_004.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_004.hresp")
}

// TestSCEN_RH_005_GetIndividualResourceEndpoint verifies SCEN-RH-005:
// The application correctly uses Unimock for an individual resource endpoint (e.g., GET /mocks/{id}).
func TestSCEN_RH_005_GetIndividualResourceEndpoint(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "GET /mocks/specific-mock-id",
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
		Data:        `{"mockId": "specific-mock-id", "value": "This is a specific mock."}`,
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_005.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_005.hresp")
}

// TestSCEN_RH_006_GetCollectionEndpoint verifies SCEN-RH-006:
// The application correctly uses Unimock for a collection endpoint (e.g., GET /mocks).
func TestSCEN_RH_006_GetCollectionEndpoint(t *testing.T) {
	given, when, then := newParts(t)

	given.a_scenario(model.Scenario{
		RequestPath: "GET /mocks",
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
		Data: `[{"mockId":"mock1","value":"First mock in collection"},` +
			`{"mockId":"mock2","value":"Second mock in collection"}]`,
	})

	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_006.http")

	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_006.hresp")
}

// TestSCEN_RH_007_PostInvalidContentType verifies SCEN-RH-007:
// Unimock rejects a POST request with an unsupported Content-Type header with a 415 status.
func TestSCEN_RH_007_PostInvalidContentType(t *testing.T) {
	_, when, then := newParts(t)

	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_007.http")

	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_007.hresp")
}

// TestSCEN_RH_008_GetNonExistentResource verifies SCEN-RH-008:
// A GET request for a resource not configured in Unimock returns a 404 Not Found.
func TestSCEN_RH_008_GetNonExistentResource(t *testing.T) {
	_, when, then := newParts(t)

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_008.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_008.hresp")
}

// TestSCEN_RH_009_PathBasedRouting verifies SCEN-RH-009:
// Unimock correctly routes requests to different mock configurations based on the request path.
func TestSCEN_RH_009_PathBasedRouting(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.
		a_scenario(model.Scenario{
			RequestPath: "GET /test/routing/pathA",
			StatusCode:  http.StatusOK,
			ContentType: "text/plain",
			Data:        "Response A",
		}).and().
		a_scenario(model.Scenario{
			RequestPath: "GET /test/routing/pathB",
			StatusCode:  http.StatusOK,
			ContentType: "text/plain",
			Data:        "Response B",
		})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_009.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_009.hresp")
}

// TestSCEN_RH_010_WildcardPathMatching verifies SCEN-RH-010:
// Unimock supports wildcard matching in request paths (e.g., /users/*).
func TestSCEN_RH_010_WildcardPathMatching(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "GET /users/*",
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
		Data:        `{"status": "matched by wildcard"}`,
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_rh_010.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_rh_010.hresp")
}

// TestSCEN_SH_001_ExactPathScenarioMatch verifies SCEN-SH-001:
// A configured scenario is matched by its exact RequestPath.
func TestSCEN_SH_001_ExactPathScenarioMatch(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "GET /custom/scenario/exact",
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
		Data:        `{"message": "exact scenario matched"}`,
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_sh_001.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_sh_001.hresp")
}

// TestSCEN_SH_002_WildcardPathScenarioMatch verifies SCEN-SH-002:
// A configured scenario with a wildcard in RequestPath is matched.
func TestSCEN_SH_002_WildcardPathScenarioMatch(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "POST /custom/scenario/*",
		StatusCode:  http.StatusCreated,
		ContentType: "text/plain",
		Data:        "wildcard scenario matched",
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_sh_002.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_sh_002.hresp")
}

// TestSCEN_SH_003_ScenarioSkipsMockHandling verifies SCEN-SH-003:
// If a scenario matches, normal mock resource handling for the same path is skipped.
func TestSCEN_SH_003_ScenarioSkipsMockHandling(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "GET /api/users/override-put-id-e2e-003",
		StatusCode:  http.StatusTeapot,
		ContentType: "application/vnd.custom.teapot",
		Data:        `{"message": "I am a teapot because of the scenario!"}`,
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_sh_003.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_sh_003.hresp")
}

// TestSCEN_SH_004_ScenarioMethodMismatch verifies SCEN-SH-004:
// A scenario for a specific HTTP method does not match requests with other methods on the same path.
func TestSCEN_SH_004_ScenarioMethodMismatch(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "GET /api/test/method-specific",
		StatusCode:  http.StatusOK,
		ContentType: "text/plain",
		Data:        "GET scenario response for SCEN-SH-004",
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_sh_004.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_sh_004.hresp")
}

// TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation verifies SCEN-SH-005:
// A scenario can return an empty body, a specific status code, and a custom Location header.
func TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation(t *testing.T) {
	given, when, then := newParts(t)

	// given
	given.a_scenario(model.Scenario{
		RequestPath: "POST /api/actions/submit-task",
		StatusCode:  http.StatusCreated,
		Location:    "/tasks/status/new-task-123",
	})

	// when
	when.
		an_http_request_is_made_from_file("testdata/http/scen_sh_005.http")

	// then
	then.
		the_response_is_validated_against_file("testdata/http/scen_sh_005.hresp")
}
