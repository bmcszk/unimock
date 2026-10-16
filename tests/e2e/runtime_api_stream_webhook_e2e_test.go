package e2e_test

import (
	"fmt"
	"testing"
)

// TestRuntimeAPI_StreamScenarioCreateAndStreams verifies a valid SSE stream
// scenario created through the runtime API streams immediately, without a
// server restart.
func TestRuntimeAPI_StreamScenarioCreateAndStreams(t *testing.T) {
	given, when, then := newParts(t)

	given.
		runtimeAPIUnifiedConfigFile().and().
		a_runtime_scenario_create_request_is_made_with(`{
			"uuid": "runtime-stream-001",
			"requestPath": "GET /runtime-api/live",
			"statusCode": 200,
			"contentType": "text/event-stream",
			"stream": {
				"format": "sse",
				"intervalMs": 10,
				"eventCount": 3,
				"template": "{\"n\": {{index}}}"
			}
		}`)

	when.
		a_streaming_request_is_made_to("/runtime-api/live")

	then.
		the_runtime_scenario_was_accepted().and().
		the_stream_content_type_is("text/event-stream").and().
		the_stream_contains_n_frames(3).and().
		the_stream_body_contains(`{"n": 2}`)
}

// TestRuntimeAPI_StreamScenarioRejectsHoldOpenWithoutInterval verifies the
// runtime create path enforces the shared model validation: hold_open without
// a positive interval_ms is rejected with the model error text.
func TestRuntimeAPI_StreamScenarioRejectsHoldOpenWithoutInterval(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "runtime-seed-001"
    method: "GET"
    path: "/runtime-api/seed"
    status_code: 200
`)

	when.
		a_runtime_scenario_create_request_is_made_with(`{
			"uuid": "runtime-stream-bad-001",
			"requestPath": "GET /runtime-api/hold",
			"statusCode": 200,
			"contentType": "text/event-stream",
			"stream": {
				"format": "sse",
				"intervalMs": 0,
				"holdOpen": true,
				"template": "{\"i\": {{index}}}"
			}
		}`)

	then.
		the_runtime_scenario_was_rejected_with("hold_open requires interval_ms > 0")
}

// TestRuntimeAPI_WebhookScenarioCreateAndFires verifies a webhook scenario
// created through the runtime API fires a delivery to the receiver and records
// it at /_uni/webhooks/deliveries.
func TestRuntimeAPI_WebhookScenarioCreateAndFires(t *testing.T) {
	given, when, then := newParts(t)

	given.
		runtimeAPIUnifiedConfigFile().and().
		a_webhook_receiver_is_started().and().
		a_runtime_scenario_create_request_is_made_with(fmt.Sprintf(`{
			"uuid": "runtime-webhook-001",
			"requestPath": "POST /runtime-api/orders",
			"statusCode": 201,
			"contentType": "application/json",
			"data": "{\"id\":\"order-1\",\"status\":\"created\"}",
			"webhook": {
				"url": %q,
				"method": "POST",
				"body": "{\"event\":\"order.created\",\"id\":\"{{uuid}}\"}",
				"maxAttempts": 1,
				"baseMs": 10,
				"maxMs": 50
			}
		}`, when.webhookReceiver.URL))

	when.
		a_post_request_with_open_body_is_made_to("/runtime-api/orders")

	then.
		the_runtime_scenario_was_accepted().and().
		wait_for_webhook_delivery().and().
		the_webhook_receiver_saw_n_requests(1).and().
		the_webhook_receiver_saw_method("POST").and().
		the_webhook_receiver_saw_body_containing(`"event":"order.created"`).and().
		the_webhook_delivery_is_recorded_at_uni_deliveries()
}

// TestRuntimeAPI_WebhookScenarioRejectsInvalidConfig verifies the runtime
// create path enforces the webhook retry ceiling shared with YAML load.
func TestRuntimeAPI_WebhookScenarioRejectsInvalidConfig(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "runtime-webhook-seed-001"
    method: "GET"
    path: "/runtime-api/webhook-seed"
    status_code: 200
`)

	when.
		a_runtime_scenario_create_request_is_made_with(`{
			"uuid": "runtime-webhook-bad-001",
			"requestPath": "POST /runtime-api/orders",
			"statusCode": 201,
			"contentType": "application/json",
			"data": "{}",
			"webhook": {
				"url": "http://localhost:1/hook",
				"maxAttempts": 9999
			}
		}`)

	then.
		the_runtime_scenario_was_rejected_with("maxAttempts must be <= 12, got 9999")
}

// TestRuntimeAPI_StreamWebhookMutuallyExclusive verifies the shared model
// validation rejects a runtime scenario carrying both stream and webhook —
// the same rule YAML load enforces at startup.
func TestRuntimeAPI_StreamWebhookMutuallyExclusive(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "runtime-xor-seed-001"
    method: "GET"
    path: "/runtime-api/xor-seed"
    status_code: 200
`)

	when.
		a_runtime_scenario_create_request_is_made_with(`{
			"uuid": "runtime-xor-bad-001",
			"requestPath": "GET /runtime-api/xor",
			"statusCode": 200,
			"stream": {
				"format": "sse",
				"eventCount": 1,
				"intervalMs": 10,
				"template": "{\"i\": {{index}}}"
			},
			"webhook": {
				"url": "http://127.0.0.1:1/hook"
			}
		}`)

	then.
		the_runtime_scenario_was_rejected_with("stream and webhook are mutually exclusive")
}
