package e2e_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	restclient "github.com/bmcszk/go-restclient"
	"github.com/bmcszk/unimock/pkg/client"
)

const (
	clientErrorMsg = "Failed to create client: %v"
)

// TestMain runs setup and cleanup for all E2E tests
func TestMain(m *testing.M) {
	// Clean up any existing scenarios before running tests
	cleanupAllScenarios()

	// Run the tests
	m.Run()

	// Clean up again after tests
	cleanupAllScenarios()
}

// cleanupAllScenarios removes all scenarios from the server to ensure clean test state
func cleanupAllScenarios() {
	unimockClient, err := client.NewClient(getBaseURL())
	if err != nil {
		fmt.Printf("Warning: Failed to create client for cleanup: %v\n", err)
		return
	}

	scenarios, err := unimockClient.ListScenarios(context.Background())
	if err != nil {
		fmt.Printf("Warning: Failed to list scenarios for cleanup: %v\n", err)
		return
	}

	for _, scenario := range scenarios {
		err := unimockClient.DeleteScenario(context.Background(), scenario.UUID)
		if err != nil {
			fmt.Printf("Warning: Failed to delete scenario %s: %v\n", scenario.UUID, err)
		}
	}
}

// getBaseURL returns the base URL for E2E tests from environment variable or default
func getBaseURL() string {
	if url := os.Getenv("UNIMOCK_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080" // Default to original port for backward compatibility
}

// TestE2E_SCEN_RM_MULTI_ID_001 tests REQ-RM-MULTI-ID - Scenario 001
// Create a resource with multiple IDs (one from header, one from body JSON path)
// and verify it can be retrieved by either ID.
func TestE2E_SCEN_RM_MULTI_ID_001(t *testing.T) {
	httpClient, err := restclient.NewClient(restclient.WithBaseURL(getBaseURL()))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := httpClient.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_001.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_001.http: %v", err)
	}
	err = httpClient.ValidateResponses("testdata/http/scen_rm_multi_id_001.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_001.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_002(t *testing.T) {
	httpClient, err := restclient.NewClient(restclient.WithBaseURL(getBaseURL()))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := httpClient.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_002.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_002.http: %v", err)
	}
	err = httpClient.ValidateResponses("testdata/http/scen_rm_multi_id_002.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_002.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_003(t *testing.T) {
	httpClient, err := restclient.NewClient(restclient.WithBaseURL(getBaseURL()))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := httpClient.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_003.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_003.http: %v", err)
	}
	err = httpClient.ValidateResponses("testdata/http/scen_rm_multi_id_003.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_003.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_004(t *testing.T) {
	httpClient, err := restclient.NewClient(restclient.WithBaseURL(getBaseURL()))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := httpClient.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_004.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_004.http: %v", err)
	}
	err = httpClient.ValidateResponses("testdata/http/scen_rm_multi_id_004.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_004.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_005(t *testing.T) {
	httpClient, err := restclient.NewClient(restclient.WithBaseURL(getBaseURL()))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := httpClient.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_005.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_005.http: %v", err)
	}
	err = httpClient.ValidateResponses("testdata/http/scen_rm_multi_id_005.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_005.hresp: %v", err)
	}
}
