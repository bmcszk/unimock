package e2e_test

import (
	"fmt"
	"testing"
)

func TestDualBox_WebhookBroadcastBetweenInstances(t *testing.T) {
	given, when, then := newParts(t)

	given.
		a_second_unimock_instance_is_started_with(`
scenarios:
  - uuid: "dual-box-receiver-001"
    method: "POST"
    path: "/hook"
    status_code: 200
    content_type: "application/json"
    data: |
      {"received":true}
`).and().
		configFileFromContent(fmt.Sprintf(`
scenarios:
  - uuid: "dual-box-broadcaster-001"
    method: "POST"
    path: "/orders"
    status_code: 201
    content_type: "application/json"
    data: |
      {"id":"order-1","status":"created"}
    webhook:
      url: %q
      method: POST
      body: '{"event":"order.created","source":"box-a"}'
      max_attempts: 1
      base_ms: 10
      max_ms: 50
`, given.receiver.baseURL+"/hook"))

	when.
		a_post_request_with_open_body_is_made_to("/orders")

	then.
		the_response_body_contains_fixtures_data(`"order-1"`).and().
		the_second_instance_received_the_webhook_request("/hook").and().
		the_second_instance_responded_200_to("/hook").and().
		the_delivery_to_is_recorded_as_successful(then.receiver.baseURL + "/hook").and().
		the_second_instance_records_no_outbound_deliveries()
}

// receiver box keeps serving own streams while absorbing cross-instance deliveries.
func TestDualBox_ReceiverConfirmsDeliveryAndStreams(t *testing.T) {
	given, when, then := newParts(t)

	given.
		a_second_unimock_instance_is_started_with(`
scenarios:
  - uuid: "dual-box-receiver-002"
    method: "POST"
    path: "/hook"
    status_code: 200
    content_type: "application/json"
    data: |
      {"received":true}

  - uuid: "dual-box-stream-002"
    method: "GET"
    path: "/events"
    status_code: 200
    stream:
      format: "sse"
      interval_ms: 10
      event_count: 3
      template: '{"n": {{index}}}'
`).and().
		configFileFromContent(fmt.Sprintf(`
scenarios:
  - uuid: "dual-box-broadcaster-002"
    method: "POST"
    path: "/orders"
    status_code: 201
    content_type: "application/json"
    data: |
      {"id":"order-2","status":"created"}
    webhook:
      url: %q
      method: POST
      body: '{"event":"order.created","source":"box-a"}'
      max_attempts: 1
      base_ms: 10
      max_ms: 50
`, given.receiver.baseURL+"/hook"))

	when.
		a_post_request_with_open_body_is_made_to("/orders").and().
		a_streaming_request_is_made_to_the_second_instance("/events")

	then.
		the_response_body_contains_fixtures_data(`"order-2"`).and().
		the_second_instance_received_the_webhook_request("/hook").and().
		the_second_instance_responded_200_to("/hook").and().
		the_delivery_to_is_recorded_as_successful(then.receiver.baseURL + "/hook").and().
		the_stream_content_type_is("text/event-stream").and().
		the_stream_contains_n_frames(3).and().
		the_stream_body_contains(`{"n": 2}`)
}
