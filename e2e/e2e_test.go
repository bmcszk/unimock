package e2e_test

import (
	"testing"
)

// TestE2E_SCEN_RM_MULTI_ID_001 tests REQ-RM-MULTI-ID - Scenario 001
// Create a resource with multiple IDs (one from header, one from body JSON path)
// and verify it can be retrieved by either ID.
func TestE2E_SCEN_RM_MULTI_ID_001(t *testing.T) {
	_, when, then := newParts(t)

	// given
	// No setup needed for this test

	// when
	when.
		the_resource_is_created_with_multiple_ids("testdata/http/scen_rm_multi_id_001.http")

	// then
	then.
		the_resource_can_be_retrieved_by_either_id("testdata/http/scen_rm_multi_id_001.hresp")
}

func TestE2E_SCEN_RM_MULTI_ID_002(t *testing.T) {
	_, when, then := newParts(t)

	// given
	// No setup needed for this test

	// when
	when.
		a_resource_is_updated_and_verified("testdata/http/scen_rm_multi_id_002.http")

	// then
	then.
		the_update_is_successful("testdata/http/scen_rm_multi_id_002.hresp")
}

func TestE2E_SCEN_RM_MULTI_ID_003(t *testing.T) {
	_, when, then := newParts(t)

	// given
	// No setup needed for this test

	// when
	when.
		a_resource_is_deleted_and_verified("testdata/http/scen_rm_multi_id_003.http")

	// then
	then.
		the_deletion_is_successful("testdata/http/scen_rm_multi_id_003.hresp")
}

func TestE2E_SCEN_RM_MULTI_ID_004(t *testing.T) {
	_, when, then := newParts(t)

	// given
	// No setup needed for this test

	// when
	when.
		a_resource_is_created_with_conflicting_ids("testdata/http/scen_rm_multi_id_004.http")

	// then
	then.
		a_conflict_error_is_returned("testdata/http/scen_rm_multi_id_004.hresp")
}

func TestE2E_SCEN_RM_MULTI_ID_005(t *testing.T) {
	_, when, then := newParts(t)

	// given
	// No setup needed for this test

	// when
	when.
		the_resource_is_created_with_multiple_ids("testdata/http/scen_rm_multi_id_005.http")

	// then
	then.
		the_resource_can_be_retrieved_by_either_id("testdata/http/scen_rm_multi_id_005.hresp")
}
