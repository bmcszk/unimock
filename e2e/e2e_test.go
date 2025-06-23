//go:build e2e

package e2e_test

import (
	"context"
	"testing"

	restclient "github.com/bmcszk/go-restclient"
)

const (
	defaultBaseURL = "http://localhost:8080"
	clientErrorMsg = "Failed to create client: %v"
)

// TestE2E_SCEN_RM_MULTI_ID_001 tests REQ-RM-MULTI-ID - Scenario 001
// Create a resource with multiple IDs (one from header, one from body JSON path) 
// and verify it can be retrieved by either ID.
func TestE2E_SCEN_RM_MULTI_ID_001(t *testing.T) {
	client, err := restclient.NewClient(restclient.WithBaseURL(defaultBaseURL))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := client.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_001.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_001.http: %v", err)
	}
	err = client.ValidateResponses("testdata/http/scen_rm_multi_id_001.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_001.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_002(t *testing.T) {
	client, err := restclient.NewClient(restclient.WithBaseURL(defaultBaseURL))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := client.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_002.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_002.http: %v", err)
	}
	err = client.ValidateResponses("testdata/http/scen_rm_multi_id_002.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_002.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_003(t *testing.T) {
	client, err := restclient.NewClient(restclient.WithBaseURL(defaultBaseURL))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := client.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_003.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_003.http: %v", err)
	}
	err = client.ValidateResponses("testdata/http/scen_rm_multi_id_003.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_003.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_004(t *testing.T) {
	client, err := restclient.NewClient(restclient.WithBaseURL(defaultBaseURL))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := client.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_004.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_004.http: %v", err)
	}
	err = client.ValidateResponses("testdata/http/scen_rm_multi_id_004.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_004.hresp: %v", err)
	}
}

func TestE2E_SCEN_RM_MULTI_ID_005(t *testing.T) {
	client, err := restclient.NewClient(restclient.WithBaseURL(defaultBaseURL))
	if err != nil {
		t.Fatalf(clientErrorMsg, err)
	}
	resps, err := client.ExecuteFile(context.Background(), "testdata/http/scen_rm_multi_id_005.http")
	if err != nil {
		t.Fatalf("Failed to execute file testdata/http/scen_rm_multi_id_005.http: %v", err)
	}
	err = client.ValidateResponses("testdata/http/scen_rm_multi_id_005.hresp", resps...)
	if err != nil {
		t.Fatalf("Failed to validate responses for testdata/http/scen_rm_multi_id_005.hresp: %v", err)
	}
}
