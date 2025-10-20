
package e2e_test

import (
	"context"
	"testing"

	restclient "github.com/bmcszk/go-restclient"
	"github.com/stretchr/testify/require"
)

func TestComplexLifecycle(t *testing.T) {
	// Given: A go-restclient instance
	// Base URL and other client configurations are expected to be handled by go-restclient
	// (e.g., via WithBaseURL if not using absolute URLs in .http file, or if {{baseUrl}} is set up).
	// We assume {{baseUrl}} in complex_lifecycle.http is correctly pointing to the Unimock instance.
	client, err := restclient.NewClient()
	require.NoError(t, err, "Failed to create go-restclient client")

	requestFilePath := "testdata/http/complex_lifecycle.http"
	expectedResponseFilePath := "testdata/http/complex_lifecycle.hresp"

	// Programmatic API variables can be passed if needed, but we defined static IDs in the .http file.
	// Example:
	// apiVars := map[string]string{
	// 	"primaryId": "e2e-static-prod-001", // Not needed if @primaryId is in .http
	// 	"secondaryId": "SKU-E2E-STATIC-001", // Not needed if @secondaryId is in .http
	// }

	// When: Executing all requests defined in the .http file
	// Passing nil for apiVars as they are in the file
	responses, execErr := client.ExecuteFile(context.Background(), requestFilePath)
	// It's important to check execErr AFTER validation for some types of errors,
	// but critical execution errors (like file not found for requests) should be checked here.
	// However, go-restclient.ValidateResponses will also fail if responses array is empty or nil.

	// Then: Validate all responses against the .hresp file
	validationErr := client.ValidateResponses(expectedResponseFilePath, responses...)

	// Consolidate error reporting
	if execErr != nil {
		t.Logf("Error during HTTP file execution: %v", execErr)
		// Optionally, print details of responses if any were partially processed
		for _, resp := range responses {
			if resp.Error != nil {
				t.Logf("Request (%s %s) had an execution error: %v\n", 
					resp.Request.Method, resp.Request.URL, resp.Error)
			}
		}
		require.FailNow(t, "Execution of .http file failed", execErr.Error())
	}

	require.NoErrorf(t, validationErr, "Validation of responses against '%s' failed", expectedResponseFilePath)

	t.Logf("Successfully executed and validated %d requests from %s against %s", 
		len(responses), requestFilePath, expectedResponseFilePath)
}
