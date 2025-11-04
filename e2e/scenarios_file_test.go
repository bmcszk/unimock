package e2e_test

import (
	"testing"
)

// TestScenarioFileLoading tests that scenarios can be loaded from a unified config file
func TestScenarioFileLoading_GET_ScenarioFromFile(t *testing.T) {
	given, when, then := newParts(t)

	given.
		unifiedConfigFile()

	when.
		a_get_request_is_made_to("/test-scenarios/user/123")

	then.
		the_response_is_successful()
}

func TestScenarioFileLoading_POST_ScenarioFromFile(t *testing.T) {
	given, when, then := newParts(t)

	given.
		unifiedConfigFile()

	when.
		a_post_request_is_made_to("/test-scenarios/users")

	then.
		the_post_response_is_successful()
}

func TestScenarioFileLoading_HEAD_ScenarioFromFile(t *testing.T) {
	given, when, then := newParts(t)

	given.
		unifiedConfigFile()

	when.
		a_head_request_is_made_to("/test-scenarios/user/789")

	then.
		the_head_response_is_successful()
}

// TestScenarioFileAndRuntimeAPIIntegration tests that file-based and runtime scenarios work together
func TestScenarioFileAndRuntimeAPIIntegration_FileScenarioWorks(t *testing.T) {
	given, when, then := newParts(t)

	given.
		integrationUnifiedConfigFile()

	when.
		a_get_request_is_made_to("/integration-test/file-resource")

	then.
		the_response_is_successful()
}

func TestScenarioFileAndRuntimeAPIIntegration_RuntimeScenarioWorksAlongsideFileScenario(t *testing.T) {
	given, when, then := newParts(t)

	given.
		integrationUnifiedConfigFile().and().
		a_runtime_scenario_is_created()

	when.
		a_get_request_is_made_to("/integration-test/runtime-resource")

	then.
		the_response_is_successful().and().
		the_file_scenario_still_works()
}
