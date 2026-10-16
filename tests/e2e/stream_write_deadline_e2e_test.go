package e2e_test

import (
	"io"
	"testing"
	"time"
)

// the_stream_is_read_to_eof blocks until the server closes the stream, records
// the wall time spent reading, and keeps the body for frame-count assertions.
func (p *parts) the_stream_is_read_to_eof() *parts {
	p.Helper()
	p.require.NotNil(p.streamingResponse, "streaming response not opened")
	start := time.Now()
	body, err := io.ReadAll(p.streamingResponse.Body)
	p.streamingElapsed = time.Since(start)
	p.streamingErr = err
	p.streamingBody = string(body)
	return p
}

// the_stream_ended_cleanly asserts the stream reached a clean EOF with no
// transport error (a server-side truncation surfaces as an unexpected EOF).
func (p *parts) the_stream_ended_cleanly() *parts {
	p.Helper()
	p.require.NoError(p.streamingErr, "stream must end with clean EOF, not a transport error")
	return p
}

// the_stream_elapsed_at_least asserts the wall time from response open to EOF.
func (p *parts) the_stream_elapsed_at_least(want time.Duration) *parts {
	p.Helper()
	p.require.GreaterOrEqual(p.streamingElapsed, want,
		"stream lifetime must exceed the server WriteTimeout window")
	return p
}

// TestStreaming_StreamLongerThanServerWriteTimeoutCompletes proves a stream
// whose total duration (~12s) exceeds the server's 10s WriteTimeout still
// delivers every frame and closes cleanly — the per-frame write-deadline
// refresh keeps the connection alive past the static server deadline.
func TestStreaming_StreamLongerThanServerWriteTimeoutCompletes(t *testing.T) {
	given, when, then := newParts(t)

	given.
		configFileFromContent(`
scenarios:
  - uuid: "stream-write-deadline-001"
    method: "GET"
    path: "/long-events"
    status_code: 200
    stream:
      format: "sse"
      interval_ms: 600
      event_count: 20
      template: '{"n": {{index}}}'
`)

	when.
		a_streaming_request_is_made_to("/long-events").and().
		the_stream_is_read_to_eof()

	then.
		the_stream_contains_n_frames(20).and().
		the_stream_ended_cleanly().and().
		the_stream_elapsed_at_least(11 * time.Second)
}
