package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestWebhook_FiresOnScenarioMatch verifies that a POST scenario with a
// webhook config triggers an outbound POST to the configured URL, returns the
// scenario response normally, and exposes the delivery via /_uni/webhooks/deliveries.
func TestWebhook_FiresOnScenarioMatch(t *testing.T) {
	given, when, then := newParts(t)

	// given - start receiver first so we know its URL; then write a scenario
	// whose webhook target points at that receiver.
	given.
		a_webhook_receiver_is_started().and().
		configFileFromContent(fmt.Sprintf(`
scenarios:
  - uuid: "webhook-basic-001"
    method: "POST"
    path: "/orders"
    status_code: 201
    content_type: "application/json"
    data: |
      {"id":"order-1","status":"created"}
    webhook:
      url: %q
      method: POST
      body: '{"order":"created","id":"{{uuid}}"}'
      max_attempts: 1
      base_ms: 10
      max_ms: 50
`, when.webhookReceiver.URL))

	// when - POST to the scenario; the response is returned synchronously and
	// the open body keeps the request context alive for async delivery.
	when.
		a_post_request_with_open_body_is_made_to("/orders")

	// then - scenario response was delivered; webhook fired exactly once with the expected method and body
	then.
		the_response_body_contains_fixtures_data(`"order-1"`).and().
		wait_for_webhook_delivery().and().
		the_webhook_receiver_saw_n_requests(1).and().
		the_webhook_receiver_saw_method(http.MethodPost).and().
		the_webhook_receiver_saw_body_containing(`"order":"created"`).and().
		the_webhook_delivery_is_recorded_at_uni_deliveries()
}

// TestWebhook_RetriesUntilSuccess verifies the dispatcher retries failed
// attempts: receiver fails the first request with 500, accepts the second;
// the receiver observes exactly 2 attempts.
func TestWebhook_RetriesUntilSuccess(t *testing.T) {
	given, when, then := newParts(t)

	// given - receiver fails first attempt, then accepts; scenario configured with 2 max attempts
	given.
		a_webhook_receiver_is_started().and().
		the_webhook_receiver_fails_first_attempt().and().
		configFileFromContent(fmt.Sprintf(`
scenarios:
  - uuid: "webhook-retry-001"
    method: "POST"
    path: "/notify"
    status_code: 202
    content_type: "application/json"
    data: |
      {"queued":true}
    webhook:
      url: %q
      method: POST
      body: '{"retry":"yes","id":"{{uuid}}"}'
      max_attempts: 2
      base_ms: 10
      max_ms: 50
`, when.webhookReceiver.URL))

	// when - trigger the scenario
	when.
		a_post_request_with_open_body_is_made_to("/notify")

	// then - scenario response is delivered to caller; receiver saw exactly 2 attempts
	then.
		the_response_body_contains_fixtures_data(`"queued":true`).and().
		wait_for_n_webhook_deliveries(2).and().
		the_webhook_receiver_saw_n_requests(2)
}

// TestWebhook_SignatureHeaderHasV1Prefix verifies that when secret_env is set and
// resolves to a non-empty value, the dispatcher signs the webhook with the
// Standard Webhooks header prefixed with "v1,".
func TestWebhook_SignatureHeaderHasV1Prefix(t *testing.T) {
	// Set the secret env var BEFORE the receiver is started so the dispatcher
	// reads it when it signs the delivery.
	t.Setenv("UNIMOCK_WEBHOOK_SECRET", "shhh-this-is-a-test-secret")

	given, when, then := newParts(t)

	// given - scenario referencing the env var name for the HMAC secret
	given.
		a_webhook_receiver_is_started().and().
		configFileFromContent(fmt.Sprintf(`
scenarios:
  - uuid: "webhook-signing-001"
    method: "POST"
    path: "/signed"
    status_code: 200
    content_type: "application/json"
    data: |
      {"ok":true}
    webhook:
      url: %q
      method: POST
      body: '{"id":"{{uuid}}"}'
      secret_env: UNIMOCK_WEBHOOK_SECRET
      max_attempts: 1
      base_ms: 10
      max_ms: 50
`, when.webhookReceiver.URL))

	// when - trigger the scenario; signing happens inside the dispatcher
	when.
		a_post_request_with_open_body_is_made_to("/signed")

	// then - webhook-signature header is present with v1, prefix
	then.
		the_response_body_contains_fixtures_data(`"ok":true`).and().
		wait_for_webhook_delivery().and().
		the_webhook_receiver_saw_n_requests(1).and().
		the_webhook_signature_header_has_v1_prefix()
}

// TestWebhook_DeliveriesEndpointReflectsFiredWebhook verifies that after a
// webhook fires, the /_uni/webhooks/deliveries endpoint reports it.
func TestWebhook_DeliveriesEndpointReflectsFiredWebhook(t *testing.T) {
	given, when, then := newParts(t)

	// given - scenario that fires a single webhook
	given.
		a_webhook_receiver_is_started().and().
		configFileFromContent(fmt.Sprintf(`
scenarios:
  - uuid: "webhook-deliveries-001"
    method: "POST"
    path: "/audit"
    status_code: 200
    content_type: "application/json"
    data: |
      {"audit":"logged"}
    webhook:
      url: %q
      method: POST
      body: '{"event":"audit","id":"{{uuid}}"}'
      max_attempts: 1
      base_ms: 10
      max_ms: 50
`, when.webhookReceiver.URL))

	// when - trigger the scenario
	when.
		a_post_request_with_open_body_is_made_to("/audit")

	// then - give the dispatcher a moment, then verify the deliveries endpoint
	then.
		the_response_body_contains_fixtures_data(`"audit":"logged"`).and().
		wait_for_webhook_delivery().and().
		the_webhook_delivery_is_recorded_at_uni_deliveries()
}
