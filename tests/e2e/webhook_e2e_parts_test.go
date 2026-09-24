package e2e_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// receivedWebhook captures a single inbound webhook call recorded by the
// in-test httptest receiver used by the webhook E2E helpers.
type receivedWebhook struct {
	Method string
	URL    string
	Header http.Header
	Body   []byte
}

// a_post_request_with_open_body_is_made_to issues a POST request but leaves
// the response body open so the underlying request context stays alive long
// enough for any async webhook delivery to fire. The body is closed via
// t.Cleanup at test end. Use for webhook scenarios where the dispatcher
// depends on req.Context().
func (p *parts) a_post_request_with_open_body_is_made_to(url string) {
	p.Helper()
	resp, err := http.Post(p.baseURL+url, "application/json", nil)
	p.require.NoError(err)
	body, err := io.ReadAll(resp.Body)
	p.require.NoError(err)
	p.responses = []any{&ResponseWithBody{Response: resp, BodyContent: body}}
	p.Cleanup(func() { _ = resp.Body.Close() })
}

// a_webhook_receiver_is_started spins up an in-process httptest server that
// records every inbound request on parts.webhookRequests under webhookMu.
func (p *parts) a_webhook_receiver_is_started() *parts {
	p.Helper()
	p.webhookMu.Lock()
	p.webhookRequests = nil
	p.webhookFailCount = 0
	p.webhookMu.Unlock()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		p.webhookMu.Lock()
		p.webhookRequests = append(p.webhookRequests, receivedWebhook{
			Method: r.Method,
			URL:    r.URL.String(),
			Header: r.Header.Clone(),
			Body:   body,
		})
		attempt := len(p.webhookRequests)
		fail := p.webhookFailCount
		p.webhookMu.Unlock()
		if attempt <= fail {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	p.webhookReceiver = srv
	p.Cleanup(srv.Close)
	return p
}

// the_webhook_receiver_fails_first_attempt makes the next inbound request fail
// with 500; subsequent requests return 200.
func (p *parts) the_webhook_receiver_fails_first_attempt() *parts {
	p.Helper()
	p.webhookMu.Lock()
	p.webhookFailCount = 1
	p.webhookMu.Unlock()
	return p
}

// webhookSnapshot returns a defensive copy of recorded requests under the lock.
func (p *parts) webhookSnapshot() []receivedWebhook {
	p.webhookMu.Lock()
	defer p.webhookMu.Unlock()
	out := make([]receivedWebhook, len(p.webhookRequests))
	copy(out, p.webhookRequests)
	return out
}

// the_webhook_receiver_saw_n_requests asserts the receiver got exactly n calls.
func (p *parts) the_webhook_receiver_saw_n_requests(n int) *parts {
	p.Helper()
	got := len(p.webhookSnapshot())
	p.require.Equal(n, got, "webhook receiver call count")
	return p
}

// the_webhook_receiver_saw_method asserts the latest recorded method matches.
func (p *parts) the_webhook_receiver_saw_method(method string) *parts {
	p.Helper()
	recs := p.webhookSnapshot()
	p.require.NotEmpty(recs, "no webhook deliveries recorded")
	p.require.Equal(method, recs[len(recs)-1].Method, "webhook HTTP method")
	return p
}

// the_webhook_receiver_saw_body_containing asserts the latest body contains substr.
func (p *parts) the_webhook_receiver_saw_body_containing(substr string) *parts {
	p.Helper()
	recs := p.webhookSnapshot()
	p.require.NotEmpty(recs, "no webhook deliveries recorded")
	p.require.Contains(string(recs[len(recs)-1].Body), substr, "webhook body")
	return p
}

// the_webhook_signature_header_has_v1_prefix asserts Standard Webhooks signing.
func (p *parts) the_webhook_signature_header_has_v1_prefix() *parts {
	p.Helper()
	recs := p.webhookSnapshot()
	p.require.NotEmpty(recs, "no webhook deliveries recorded")
	sig := recs[len(recs)-1].Header.Get("webhook-signature")
	p.require.True(strings.HasPrefix(sig, "v1,"), "expected webhook-signature to start with v1,, got %q", sig)
	return p
}

// the_webhook_delivery_is_recorded_at_uni_deliveries polls the deliveries endpoint
// and asserts at least one delivery record was captured after the scenario fired.
// Polling is required because the dispatcher records to its ring buffer AFTER
// the HTTP response is returned to the receiver; wait_for_webhook_delivery
// alone is not enough to guarantee the ring reflects the delivery.
func (p *parts) the_webhook_delivery_is_recorded_at_uni_deliveries() {
	p.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(p.baseURL + "/_uni/webhooks/deliveries")
		p.require.NoError(err, "GET /_uni/webhooks/deliveries")
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		p.require.NoError(err, "read deliveries body")
		p.require.Equal(http.StatusOK, resp.StatusCode, "deliveries endpoint status")
		if strings.TrimSpace(string(body)) != "" && strings.TrimSpace(string(body)) != "[]" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	p.Fatal("timed out waiting for delivery record at /_uni/webhooks/deliveries")
}

// a_streaming_request_is_made_to opens a GET request with a cancel context and
// stores the live response in parts.streamingResponse for incremental reads.
func (p *parts) a_streaming_request_is_made_to(url string) *parts {
	p.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	p.streamingCancel = cancel
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+url, nil)
	p.require.NoError(err, "build streaming request")
	resp, err := http.DefaultClient.Do(req)
	p.require.NoError(err, "execute streaming request")
	p.streamingResponse = resp
	p.streamingCT = resp.Header.Get("Content-Type")
	p.Cleanup(func() { _ = resp.Body.Close() })
	return p
}

// the_stream_content_type_is asserts the open streaming response Content-Type.
func (p *parts) the_stream_content_type_is(ct string) *parts {
	p.Helper()
	p.require.NotNil(p.streamingResponse, "streaming response not opened")
	p.require.Equal(ct, p.streamingCT, "streaming Content-Type")
	return p
}

// countStreamFrames counts frames in a fully-buffered stream body per format.
func countStreamFrames(body, contentType string) int {
	lines := strings.Split(body, "\n")
	if strings.HasPrefix(contentType, "text/event-stream") {
		return countPrefixed(lines, "data: ")
	}
	return countNonEmpty(lines)
}

// countPrefixed counts lines with the given prefix.
func countPrefixed(lines []string, prefix string) int {
	n := 0
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			n++
		}
	}
	return n
}

// countNonEmpty counts non-blank lines.
func countNonEmpty(lines []string) int {
	n := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

// streamingBodyRead caches the fully-read stream body so multiple
// the_stream_body_contains assertions don't re-read a consumed body.
func (p *parts) streamingBodyRead() string {
	p.Helper()
	if p.streamingBody != "" {
		return p.streamingBody
	}
	p.require.NotNil(p.streamingResponse, "streaming response not opened")
	body, err := io.ReadAll(p.streamingResponse.Body)
	p.require.NoError(err, "read streaming body")
	p.streamingBody = string(body)
	return p.streamingBody
}

// the_stream_contains_n_frames reads the full streaming body and asserts the frame count.
func (p *parts) the_stream_contains_n_frames(n int) *parts {
	p.Helper()
	got := countStreamFrames(p.streamingBodyRead(), p.streamingCT)
	p.require.Equal(n, got, "stream frame count for %q", p.streamingCT)
	return p
}

// the_stream_body_contains asserts the buffered stream body contains substring.
func (p *parts) the_stream_body_contains(substr string) *parts {
	p.Helper()
	p.require.Contains(p.streamingBodyRead(), substr, "streaming body")
	return p
}

// the_client_cancels_the_stream aborts the streaming request via context cancel.
func (p *parts) the_client_cancels_the_stream() *parts {
	p.Helper()
	if p.streamingCancel != nil {
		p.streamingCancel()
	}
	if p.streamingResponse != nil && p.streamingResponse.Body != nil {
		_ = p.streamingResponse.Body.Close()
	}
	return p
}

// the_stream_receives_at_least_one_frame blocks until one byte (one frame prefix) is read.
func (p *parts) the_stream_receives_at_least_one_frame() *parts {
	p.Helper()
	p.require.NotNil(p.streamingResponse, "streaming response not opened")
	type readResult struct {
		n   int
		err error
	}
	ch := make(chan readResult, 1)
	go func() {
		buf := make([]byte, 1)
		n, err := p.streamingResponse.Body.Read(buf)
		ch <- readResult{n: n, err: err}
	}()
	select {
	case res := <-ch:
		p.require.NoError(res.err, "expected at least one frame before cancel")
		p.require.Greater(res.n, 0, "stream read returned 0 bytes")
	case <-time.After(2 * time.Second):
		p.Fatal("timed out waiting for first stream frame")
	}
	return p
}

// wait_for_webhook_delivery is a small polling helper that returns once the
// receiver has observed at least one call, so subsequent assertions don't race
// the async dispatcher.
func (p *parts) wait_for_webhook_delivery() *parts {
	p.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(p.webhookSnapshot()) >= 1 {
			return p
		}
		time.Sleep(5 * time.Millisecond)
	}
	p.Fatal("timed out waiting for webhook delivery to land at receiver")
	return p
}

// wait_for_n_webhook_deliveries polls until the receiver has observed at least n
// calls. Use this for retry-style tests where the first attempt is the trigger
// to keep polling — without it, single-shot wait_for_webhook_delivery returns
// after the failing first attempt and races the retry backoff.
func (p *parts) wait_for_n_webhook_deliveries(n int) *parts {
	p.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(p.webhookSnapshot()) >= n {
			return p
		}
		time.Sleep(5 * time.Millisecond)
	}
	p.Fatalf("timed out waiting for %d webhook deliveries at receiver", n)
	return p
}
