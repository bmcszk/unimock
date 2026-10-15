package e2e_test

import (
	"testing"
)

// TestStreaming_SSE_EventCountDeliversFramesThenEOF verifies an SSE streaming
// scenario: exactly event_count `data: ` frames, then EOF, with the SSE
// Content-Type and {{index}} substitution in frame bodies.
func TestStreaming_SSE_EventCountDeliversFramesThenEOF(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "stream-sse-001"
    method: "GET"
    path: "/events"
    status_code: 200
    stream:
      format: "sse"
      interval_ms: 10
      event_count: 3
      template: '{"n": {{index}}}'
`)

	when.
		a_streaming_request_is_made_to("/events")

	then.
		the_stream_content_type_is("text/event-stream").and().
		the_stream_contains_n_frames(3).and().
		the_stream_body_contains(`{"n": 0}`).and().
		the_stream_body_contains(`{"n": 2}`)
}

// TestStreaming_FramesArriveSpacedByInterval proves inter-frame pacing is
// honored on the wire: consecutive SSE frames arrive at least interval_ms
// apart, measured via per-read-event arrival timestamps (lower bound only —
// an upper bound would flake under CI scheduler jitter).
func TestStreaming_FramesArriveSpacedByInterval(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "stream-pacing-001"
    method: "GET"
    path: "/paced-events"
    status_code: 200
    stream:
      format: "sse"
      interval_ms: 300
      event_count: 5
      template: '{"n": {{index}}}'
`)

	when.
		a_streaming_request_is_made_to("/paced-events")

	then.
		the_stream_content_type_is("text/event-stream").and().
		the_stream_contains_n_frames(5).and().
		the_stream_frames_arrive_at_least_interval_ms_apart(300)
}

// TestStreaming_NDJSON_DeliversNewlineDelimitedFrames verifies an NDJSON
// streaming scenario: newline-terminated frames with the NDJSON Content-Type.
func TestStreaming_NDJSON_DeliversNewlineDelimitedFrames(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "stream-ndjson-001"
    method: "GET"
    path: "/ticks"
    status_code: 200
    stream:
      format: "ndjson"
      interval_ms: 10
      event_count: 2
      template: '{"tick": {{index}}}'
`)

	when.
		a_streaming_request_is_made_to("/ticks")

	then.
		the_stream_content_type_is("application/x-ndjson").and().
		the_stream_contains_n_frames(2).and().
		the_stream_body_contains(`{"tick": 1}`)
}

// TestStreaming_HoldOpen_ClientCancelEndsCleanly verifies hold_open streaming
// ends when the client disconnects (context cancel) and the test completes
// without hanging.
func TestStreaming_HoldOpen_ClientCancelEndsCleanly(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "stream-hold-001"
    method: "GET"
    path: "/live"
    status_code: 200
    stream:
      format: "sse"
      interval_ms: 20
      hold_open: true
      template: '{"i": {{index}}}'
`)

	when.
		a_streaming_request_is_made_to("/live").and().
		the_stream_receives_at_least_one_frame().and().
		the_client_cancels_the_stream()

	then.
		the_stream_content_type_is("text/event-stream")
}

// TestStreaming_ScenarioHeadersReachStreamedResponse proves scenario-supplied
// headers are applied to streamed responses too, not only static ones.
func TestStreaming_ScenarioHeadersReachStreamedResponse(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "stream-headers-001"
    method: "GET"
    path: "/header-events"
    status_code: 200
    headers:
      X-Scenario-Tag: "streamed"
    stream:
      format: "sse"
      interval_ms: 10
      event_count: 2
      template: '{"n": {{index}}}'
`)

	when.
		a_streaming_request_is_made_to("/header-events")

	then.
		the_stream_content_type_is("text/event-stream").and().
		the_stream_contains_n_frames(2).and().
		the_response_header_equals("X-Scenario-Tag", "streamed")
}
