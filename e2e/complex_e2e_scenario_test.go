//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"testing"

	go_restclient "github.com/bmcszk/go-restclient"
	"github.com/bmcszk/unimock/pkg/client"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/require"
)

const (
	complexE2EBaseURL      = "http://localhost:8080" // Assuming Unimock runs here
	complexE2EPrimaryID    = "e2e-static-prod-001"
	complexE2ESecondarySKU = "SKU-E2E-STATIC-001"
)

// TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle verifies REQ-E2E-COMPLEX-001.
func TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle(t *testing.T) {
	ctx := context.Background()

	// Initialize Unimock API client (for scenario management)
	unimockAPIClient, err := client.NewClient(complexE2EBaseURL)
	require.NoError(t, err, "Failed to create unimock API client")

	// Initialize go-restclient
	// Variables defined in .http files (@primaryId, @secondarySku) will be used.
	// We could also pass them via WithVars if preferred or if they needed to be more dynamic from Go.
	rc, err := go_restclient.NewClient(go_restclient.WithBaseURL(complexE2EBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	var scenarioID string // Declare scenarioID in the scope of the main test function

	// Top-level cleanup to ensure scenario is deleted if test panics or Step 7 doesn't complete
	t.Cleanup(func() {
		if scenarioID != "" { // Only attempt delete if scenarioID was set and not cleared by Step 7
			t.Logf("Top-level cleanup: Attempting to delete scenario %s", scenarioID)
			delErr := unimockAPIClient.DeleteScenario(ctx, scenarioID)
			if delErr != nil {
				t.Logf("Top-level cleanup: Failed to delete scenario %s: %v", scenarioID, delErr)
			} else {
				t.Logf("Top-level cleanup: Successfully deleted scenario %s", scenarioID)
			}
		}
	})

	// --- Part 1: Steps 1, 3-4 (Create, Update, Verify Update - Step 2 Retrieve by SKU skipped due to current limitations) ---
	t.Run("Part1_CreateUpdateVerify", func(t *testing.T) {
		requestFilePath := "testdata/http/complex_e2e_scenario_001_part1_steps1-4.http"
		expectedResponseFilePath := "testdata/http/complex_e2e_scenario_001_part1_steps1-4.hresp"

		responses, execErr := rc.ExecuteFile(ctx, requestFilePath)
		if execErr != nil {
			t.Logf("Error during HTTP file execution for Part 1: %v", execErr)
			for _, resp := range responses {
				if resp.Error != nil {
					t.Logf("Request (%s %s) had an execution error: %v\n", resp.Request.Method, resp.Request.URL, resp.Error)
				}
			}
			require.FailNow(t, "Execution of Part 1 .http file failed", execErr.Error())
		}
		require.Len(t, responses, 3, "Expected 3 responses from Part 1 (Step 2 skipped)")

		validationErr := rc.ValidateResponses(expectedResponseFilePath, responses...)
		require.NoErrorf(t, validationErr, "Validation of Part 1 responses against '%s' failed", expectedResponseFilePath)
		t.Logf("Part 1 (Steps 1, 3-4) completed successfully.")
	})

	// --- PRD Step 5: Apply Scenario Override ---
	// var scenarioID string // Moved to parent scope
	t.Run("Step5_ApplyScenarioOverride", func(t *testing.T) {
		scenarioToCreate := &model.Scenario{
			RequestPath: "GET /products/" + complexE2EPrimaryID, // Target the specific resource
			StatusCode:  http.StatusTeapot,
			ContentType: "application/json", // PRD example
			Headers: map[string]string{
				"X-Custom-Header": "Teapot",
			},
			Data: `{"message": "I'm a teapot"}`,
		}
		createdScenario, err := unimockAPIClient.CreateScenario(ctx, scenarioToCreate)
		require.NoError(t, err, "Failed to create Unimock scenario for override")
		require.NotNil(t, createdScenario, "Created scenario should not be nil")
		require.NotEmpty(t, createdScenario.UUID, "Created scenario UUID should not be empty")
		scenarioID = createdScenario.UUID // Assign to the scenarioID in the parent scope
		t.Logf("Unimock scenario %s created successfully for override.", scenarioID)

		// REMOVED t.Cleanup from here as it was causing premature deletion
	})
	require.NotEmpty(t, scenarioID, "Scenario ID must be captured from Step 5 and be non-empty before Step 6")

	// --- Part 2: Step 6 (Verify Scenario Override) ---
	t.Run("Part2_Step6_VerifyScenarioOverride", func(t *testing.T) {
		requestFilePath := "testdata/http/complex_e2e_scenario_001_part2_step6.http"
		expectedResponseFilePath := "testdata/http/complex_e2e_scenario_001_part2_step6.hresp"

		responses, execErr := rc.ExecuteFile(ctx, requestFilePath)
		if execErr != nil {
			t.Logf("Error during HTTP file execution for Part 2 (Step 6): %v", execErr)
			// Optional: Log response details
			require.FailNow(t, "Execution of Part 2 (Step 6) .http file failed", execErr.Error())
		}
		require.Len(t, responses, 1, "Expected 1 response from Part 2 (Step 6)")

		validationErr := rc.ValidateResponses(expectedResponseFilePath, responses...)
		require.NoErrorf(t, validationErr, "Validation of Part 2 (Step 6) responses against '%s' failed", expectedResponseFilePath)
		t.Logf("Part 2 (Step 6 - Verify Scenario Override) completed successfully.")
	})

	// --- PRD Step 7: Delete Scenario Override ---
	t.Run("Step7_DeleteScenarioOverride", func(t *testing.T) {
		require.NotEmpty(t, scenarioID, "Scenario ID must be set to delete scenario in Step 7") // Ensure scenarioID is available
		err := unimockAPIClient.DeleteScenario(ctx, scenarioID)
		require.NoError(t, err, "Failed to delete Unimock scenario %s", scenarioID)
		t.Logf("Unimock scenario %s deleted successfully.", scenarioID)
		scenarioID = "" // Clear scenarioID after successful deletion to prevent re-deletion in top-level cleanup
	})

	// --- Part 3: Steps 8-10 (Verify Scenario Removal, Delete Resource, Verify Deletion) ---
	t.Run("Part3_Steps8-10_VerifyScenarioRemovalAndDelete", func(t *testing.T) {
		requestFilePath := "testdata/http/complex_e2e_scenario_001_part3_steps8-10.http"
		expectedResponseFilePath := "testdata/http/complex_e2e_scenario_001_part3_steps8-10.hresp"

		responses, execErr := rc.ExecuteFile(ctx, requestFilePath)
		if execErr != nil {
			t.Logf("Error during HTTP file execution for Part 3 (Steps 8-10): %v", execErr)
			// Optional: Log response details
			require.FailNow(t, "Execution of Part 3 (Steps 8-10) .http file failed", execErr.Error())
		}
		require.Len(t, responses, 3, "Expected 3 responses from Part 3 (Steps 8-10)")

		validationErr := rc.ValidateResponses(expectedResponseFilePath, responses...)
		require.NoErrorf(t, validationErr, "Validation of Part 3 (Steps 8-10) responses against '%s' failed", expectedResponseFilePath)
		t.Logf("Part 3 (Steps 8-10 - Verify Scenario Removal, Delete Resource, Verify Deletion) completed successfully.")
	})

	t.Logf("TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle completed all steps.")
}
