package e2e_test

import (
	"net/http"
	"testing"

	"github.com/bmcszk/unimock/pkg/model"
)

// TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle verifies REQ-E2E-COMPLEX-001.
func TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle(t *testing.T) {
	given, when, then := newParts(t)

	// Part 1: Create, Update, Verify
	when.
		an_http_request_is_made_from_file("testdata/http/complex_e2e_scenario_001_part1_steps1-4.http")

	then.
		the_response_is_validated_against_file("testdata/http/complex_e2e_scenario_001_part1_steps1-4.hresp")

	// Part 2: Apply and Verify Scenario Override
	httpFile2 := "testdata/http/complex_e2e_scenario_001_part2_step6.http"
	hrespFile2 := "testdata/http/complex_e2e_scenario_001_part2_step6.hresp"
	given.
		a_scenario_override_is_applied(model.Scenario{
			RequestPath: "GET /products/e2e-static-prod-001",
			StatusCode:  http.StatusTeapot,
			ContentType: "application/json",
			Headers: map[string]string{
				"X-Custom-Header": "Teapot",
			},
			Data: `{"message": "I'm a teapot"}`,
		})

	when.
		an_http_request_is_made_from_file(httpFile2)

	then.
		the_scenario_override_is_verified(hrespFile2)

	// Part 3: Delete Scenario and Verify
	httpFile3 := "testdata/http/complex_e2e_scenario_001_part3_steps8-10.http"
	hrespFile3 := "testdata/http/complex_e2e_scenario_001_part3_steps8-10.hresp"

	given.
		the_scenario_override_is_deleted()

	when.
		an_http_request_is_made_from_file(httpFile3)

	then.
		the_scenario_removal_is_verified(hrespFile3)
}
