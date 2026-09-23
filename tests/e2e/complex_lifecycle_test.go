package e2e_test

import (
	"testing"
)

func TestComplexLifecycle(t *testing.T) {
	_, when, then := newParts(t)

	when.
		an_http_request_is_made_from_file("testdata/http/complex_lifecycle.http")

	then.
		the_response_is_validated_against_file("testdata/http/complex_lifecycle.hresp")
}
