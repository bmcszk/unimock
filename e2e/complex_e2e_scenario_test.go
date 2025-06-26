//go:build e2e

package e2e_test

import (
	"context"
	"net/http"
	"testing"

	restclient "github.com/bmcszk/go-restclient"
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

	// Initialize clients
	unimockAPIClient, rc := setupTestClients(t)
	
	var scenarioID string
	setupScenarioCleanup(ctx, t, unimockAPIClient, &scenarioID)

	// Execute test parts
	runPart1CreateUpdateVerify(ctx, t, rc)
	scenarioID = runStep5ApplyScenarioOverride(ctx, t, unimockAPIClient)
	runPart2VerifyScenarioOverride(ctx, t, rc)
	runStep7DeleteScenarioOverride(ctx, t, unimockAPIClient, &scenarioID)
	runPart3VerifyScenarioRemovalAndDelete(ctx, t, rc)

	t.Logf("TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle completed all steps.")
}

type httpClient interface {
	ExecuteFile(context.Context, string) ([]*restclient.Response, error)
	ValidateResponses(string, ...*restclient.Response) error
}

func setupTestClients(t *testing.T) (*client.Client, httpClient) {
	t.Helper()
	// Initialize Unimock API client (for scenario management)
	unimockAPIClient, err := client.NewClient(complexE2EBaseURL)
	require.NoError(t, err, "Failed to create unimock API client")

	// Initialize go-restclient
	rc, err := restclient.NewClient(restclient.WithBaseURL(complexE2EBaseURL))
	require.NoError(t, err, "Failed to create go-restclient instance")

	return unimockAPIClient, rc
}

func setupScenarioCleanup(ctx context.Context, t *testing.T, unimockAPIClient *client.Client, scenarioID *string) {
	t.Helper()
	t.Cleanup(func() {
		if *scenarioID != "" {
			t.Logf("Top-level cleanup: Attempting to delete scenario %s", *scenarioID)
			delErr := unimockAPIClient.DeleteScenario(ctx, *scenarioID)
			if delErr != nil {
				t.Logf("Top-level cleanup: Failed to delete scenario %s: %v", *scenarioID, delErr)
			} else {
				t.Logf("Top-level cleanup: Successfully deleted scenario %s", *scenarioID)
			}
		}
	})
}

func runPart1CreateUpdateVerify(ctx context.Context, t *testing.T, rc httpClient) {
	t.Helper()
	t.Run("Part1_CreateUpdateVerify", func(t *testing.T) {
		requestFilePath := "testdata/http/complex_e2e_scenario_001_part1_steps1-4.http"
		expectedResponseFilePath := "testdata/http/complex_e2e_scenario_001_part1_steps1-4.hresp"

		responses, execErr := rc.ExecuteFile(ctx, requestFilePath)
		if execErr != nil {
			t.Logf("Error during HTTP file execution for Part 1: %v", execErr)
			require.FailNow(t, "Execution of Part 1 .http file failed", execErr.Error())
		}

		validationErr := rc.ValidateResponses(expectedResponseFilePath, responses...)
		require.NoErrorf(t, validationErr, 
			"Validation of Part 1 responses against '%s' failed", expectedResponseFilePath)
		t.Logf("Part 1 (Steps 1, 3-4) completed successfully.")
	})
}

func runStep5ApplyScenarioOverride(ctx context.Context, t *testing.T, unimockAPIClient *client.Client) string {
	t.Helper()
	var scenarioID string
	t.Run("Step5_ApplyScenarioOverride", func(t *testing.T) {
		scenarioToCreate := &model.Scenario{
			RequestPath: "GET /products/" + complexE2EPrimaryID,
			StatusCode:  http.StatusTeapot,
			ContentType: "application/json",
			Headers: map[string]string{
				"X-Custom-Header": "Teapot",
			},
			Data: `{"message": "I'm a teapot"}`,
		}
		createdScenario, err := unimockAPIClient.CreateScenario(ctx, *scenarioToCreate)
		require.NoError(t, err, "Failed to create Unimock scenario for override")
		require.NotNil(t, createdScenario, "Created scenario should not be nil")
		require.NotEmpty(t, createdScenario.UUID, "Created scenario UUID should not be empty")
		scenarioID = createdScenario.UUID
		t.Logf("Unimock scenario %s created successfully for override.", scenarioID)
	})
	return scenarioID
}

func runPart2VerifyScenarioOverride(ctx context.Context, t *testing.T, rc httpClient) {
	t.Helper()
	t.Run("Part2_Step6_VerifyScenarioOverride", func(t *testing.T) {
		requestFilePath := "testdata/http/complex_e2e_scenario_001_part2_step6.http"
		expectedResponseFilePath := "testdata/http/complex_e2e_scenario_001_part2_step6.hresp"

		responses, execErr := rc.ExecuteFile(ctx, requestFilePath)
		if execErr != nil {
			t.Logf("Error during HTTP file execution for Part 2 (Step 6): %v", execErr)
			require.FailNow(t, "Execution of Part 2 (Step 6) .http file failed", execErr.Error())
		}

		validationErr := rc.ValidateResponses(expectedResponseFilePath, responses...)
		require.NoErrorf(t, validationErr, 
			"Validation of Part 2 (Step 6) responses against '%s' failed", expectedResponseFilePath)
		t.Logf("Part 2 (Step 6 - Verify Scenario Override) completed successfully.")
	})
}

func runStep7DeleteScenarioOverride(
	ctx context.Context, 
	t *testing.T, 
	unimockAPIClient *client.Client, 
	scenarioID *string,
) {
	t.Helper()
	t.Run("Step7_DeleteScenarioOverride", func(t *testing.T) {
		require.NotEmpty(t, *scenarioID, "Scenario ID must be set to delete scenario in Step 7")
		err := unimockAPIClient.DeleteScenario(ctx, *scenarioID)
		require.NoError(t, err, "Failed to delete Unimock scenario %s", *scenarioID)
		t.Logf("Unimock scenario %s deleted successfully.", *scenarioID)
		*scenarioID = ""
	})
}

func runPart3VerifyScenarioRemovalAndDelete(ctx context.Context, t *testing.T, rc httpClient) {
	t.Helper()
	t.Run("Part3_Steps8-10_VerifyScenarioRemovalAndDelete", func(t *testing.T) {
		requestFilePath := "testdata/http/complex_e2e_scenario_001_part3_steps8-10.http"
		expectedResponseFilePath := "testdata/http/complex_e2e_scenario_001_part3_steps8-10.hresp"

		responses, execErr := rc.ExecuteFile(ctx, requestFilePath)
		if execErr != nil {
			t.Logf("Error during HTTP file execution for Part 3 (Steps 8-10): %v", execErr)
			require.FailNow(t, "Execution of Part 3 (Steps 8-10) .http file failed", execErr.Error())
		}

		validationErr := rc.ValidateResponses(expectedResponseFilePath, responses...)
		require.NoErrorf(t, validationErr, 
			"Validation of Part 3 (Steps 8-10) responses against '%s' failed", expectedResponseFilePath)
		t.Logf("Part 3 (Steps 8-10 - Verify Scenario Removal, Delete Resource, Verify Deletion) " +
			"completed successfully.")
	})
}
