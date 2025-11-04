package e2e_test

import (
	"testing"
)

// TestFixtureFileSupport_AtSyntax_BasicWorkflow tests basic @fixtures syntax workflow
func TestFixtureFileSupport_AtSyntax_BasicWorkflow(t *testing.T) {
	given, when, then := newParts(t)

	given.
		atSyntaxFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/user")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Legacy User")
}

// TestFixtureFileSupport_LessThanSyntax_GoRestClientCompatible tests < syntax workflow
func TestFixtureFileSupport_LessThanSyntax_GoRestClientCompatible(t *testing.T) {
	given, when, then := newParts(t)

	given.
		lessThanSyntaxFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/product")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("New Product")
}

// TestFixtureFileSupport_LessAtSyntax_VariableSubstitution tests <@ syntax workflow
func TestFixtureFileSupport_LessAtSyntax_VariableSubstitution(t *testing.T) {
	given, when, then := newParts(t)

	given.
		lessAtSyntaxFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/order")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Pending Order")
}

// TestFixtureFileSupport_InlineFixtures_SingleReference tests inline fixture workflow
func TestFixtureFileSupport_InlineFixtures_SingleReference(t *testing.T) {
	given, when, then := newParts(t)

	given.
		inlineFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/user-with-profile")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Test User").and().
		the_response_body_contains_fixtures_data("Developer Profile")
}

// TestFixtureFileSupport_InlineFixtures_MultipleReferences tests multiple inline fixtures
func TestFixtureFileSupport_InlineFixtures_MultipleReferences(t *testing.T) {
	given, when, then := newParts(t)

	given.
		multipleInlineFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/complete-user-data")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Test User").and().
		the_response_body_contains_fixtures_data("\"admin\": true").and().
		the_response_body_contains_fixtures_data("\"theme\": \"dark\"")
}

// TestFixtureFileSupport_BackwardCompatibility_AtSyntax tests @ syntax in mixed config
func TestFixtureFileSupport_BackwardCompatibility_AtSyntax(t *testing.T) {
	given, when, then := newParts(t)

	given.
		mixedSyntaxFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/legacy")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Legacy Data")
}

// TestFixtureFileSupport_BackwardCompatibility_LessThanSyntax tests < syntax in mixed config
func TestFixtureFileSupport_BackwardCompatibility_LessThanSyntax(t *testing.T) {
	given, when, then := newParts(t)

	given.
		mixedSyntaxFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/enhanced")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Enhanced Data")
}

// TestFixtureFileSupport_BackwardCompatibility_InlineSyntax tests inline syntax in mixed config
func TestFixtureFileSupport_BackwardCompatibility_InlineSyntax(t *testing.T) {
	given, when, then := newParts(t)

	given.
		mixedSyntaxFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/inline")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Inline Data")
}

// TestFixtureFileSupport_BackwardCompatibility_DirectSyntax tests direct data in mixed config
func TestFixtureFileSupport_BackwardCompatibility_DirectSyntax(t *testing.T) {
	given, when, then := newParts(t)

	given.
		mixedSyntaxFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/direct")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Direct Data")
}

// TestFixtureFileSupport_ErrorHandling_MissingFile tests graceful fallback for missing fixtures
func TestFixtureFileSupport_ErrorHandling_MissingFile(t *testing.T) {
	given, when, then := newParts(t)

	given.
		missingFixtureConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/missing")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("@fixtures/nonexistent.json")
}

// TestFixtureFileSupport_Security_PathTraversalProtection tests security against path traversal
func TestFixtureFileSupport_Security_PathTraversalProtection(t *testing.T) {
	given, when, then := newParts(t)

	given.
		pathTraversalConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/traversal")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("@fixtures/../../../etc/passwd")
}

// TestFixtureFileSupport_Performance_CachingTests tests fixture caching performance
func TestFixtureFileSupport_Performance_CachingTests(t *testing.T) {
	given, when, then := newParts(t)

	given.
		performanceTestConfig()

	when.
		a_get_request_is_made_to("/api/fixtures/cached")

	then.
		the_response_is_successful().and().
		the_response_body_contains_fixtures_data("Cached Data")
}
