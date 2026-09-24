package webhooks_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/webhooks"
	"github.com/bmcszk/unimock/pkg/model"
)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// waitFor polls the condition up to timeout.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

type echoCapture struct {
	mu       sync.Mutex
	lastReq  *http.Request
	lastBody []byte
}

// anEchoServer returns a test HTTP server that counts requests, captures the most
// recent one and always responds with 200 OK. Use httptest.NewServer directly when
// a custom status code or response body is needed.
func anEchoServer(t *testing.T) (*httptest.Server, *atomic.Int64, *echoCapture) {
	t.Helper()
	var (
		count   atomic.Int64
		capture echoCapture
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		body, _ := io.ReadAll(r.Body)
		capture.mu.Lock()
		capture.lastReq = r
		capture.lastBody = append([]byte(nil), body...)
		capture.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, &count, &capture
}

func TestDispatcher_SuccessFirstTry(t *testing.T) {
	srv, count, capture := anEchoServer(t)

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{URL: srv.URL, Method: "POST"}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", []byte(`{"hello":"world"}`), func() { close(done) })
	waitDone(t, done)

	if count.Load() != 1 {
		t.Errorf("expected 1 request, got %d", count.Load())
	}
	assertRequestMethod(t, capture, http.MethodPost)
	assertRequestBodyContains(t, capture, "hello")

	records := d.Snapshot()
	if len(records) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(records))
	}
	if records[0].StatusCode != 200 {
		t.Errorf("status: got %d", records[0].StatusCode)
	}
	if records[0].Attempt != 1 {
		t.Errorf("attempt: got %d", records[0].Attempt)
	}
	if records[0].ID == "" {
		t.Error("delivery ID must be set")
	}
}

// waitDone blocks on the dispatch completion channel, failing the test on timeout.
func waitDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("dispatch did not complete in time")
	}
}

func assertRequestMethod(t *testing.T, capture *echoCapture, want string) {
	t.Helper()
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.lastReq == nil {
		t.Fatal("no request captured")
	}
	if capture.lastReq.Method != want {
		t.Errorf("method: got %s want %s", capture.lastReq.Method, want)
	}
}

func assertRequestBodyContains(t *testing.T, capture *echoCapture, substr string) {
	t.Helper()
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if !strings.Contains(string(capture.lastBody), substr) {
		t.Errorf("body: got %q, want substring %q", string(capture.lastBody), substr)
	}
}

func TestDispatcher_RetryThenSuccess(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{
		URL:         srv.URL,
		Method:      "POST",
		MaxAttempts: 5,
		BaseMS:      1,
		MaxMS:       5,
	}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", []byte(`{}`), func() { close(done) })
	waitDone(t, done)

	if calls.Load() != 3 {
		t.Errorf("expected 3 calls (2 fail + 1 success), got %d", calls.Load())
	}

	records := d.Snapshot()
	if len(records) != 1 {
		t.Fatalf("expected 1 final delivery, got %d", len(records))
	}
	if records[0].StatusCode != 200 {
		t.Errorf("final status: got %d", records[0].StatusCode)
	}
	if records[0].Attempt != 3 {
		t.Errorf("final attempt: got %d, want 3", records[0].Attempt)
	}
}

func TestDispatcher_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{
		URL:         srv.URL,
		MaxAttempts: 3,
		BaseMS:      1,
		MaxMS:       5,
	}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	<-done

	if calls.Load() != 3 {
		t.Errorf("expected exactly 3 calls, got %d", calls.Load())
	}
	records := d.Snapshot()
	if len(records) != 1 {
		t.Fatalf("expected 1 final delivery record, got %d", len(records))
	}
	if records[0].StatusCode != 502 {
		t.Errorf("final status: got %d", records[0].StatusCode)
	}
	if records[0].Error == "" {
		t.Error("expected non-empty Error after all attempts failed")
	}
}

func TestDispatcher_UsesConfiguredBody(t *testing.T) {
	srv, _, capture := anEchoServer(t)

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{
		URL:  srv.URL,
		Body: `static payload uuid={{uuid}}`,
	}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", []byte(`original`), func() { close(done) })
	<-done

	capture.mu.Lock()
	defer capture.mu.Unlock()
	body := string(capture.lastBody)
	if !strings.HasPrefix(body, "static payload uuid=") {
		t.Errorf("body: got %q", body)
	}
	if body == "static payload uuid={{uuid}}" {
		t.Error("{{uuid}} was not substituted")
	}
}

func TestDispatcher_UsesRequestBodyWhenNoConfiguredBody(t *testing.T) {
	srv, _, capture := anEchoServer(t)

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{URL: srv.URL}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", []byte(`{"original":true}`), func() { close(done) })
	<-done

	capture.mu.Lock()
	defer capture.mu.Unlock()
	if string(capture.lastBody) != `{"original":true}` {
		t.Errorf("body: got %q, want %q", string(capture.lastBody), `{"original":true}`)
	}
}

func TestDispatcher_AppliesConfiguredMethodAndHeaders(t *testing.T) {
	srv, _, capture := anEchoServer(t)

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{
		URL:     srv.URL,
		Method:  "PUT",
		Headers: map[string]string{"X-Custom": "yes"},
	}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	<-done

	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.lastReq.Method != http.MethodPut {
		t.Errorf("method: got %s", capture.lastReq.Method)
	}
	if capture.lastReq.Header.Get("X-Custom") != "yes" {
		t.Error("custom header missing")
	}
}

func TestDispatcher_SigningHeaders(t *testing.T) {
	const secret = "super-secret"
	t.Setenv("WH_SECRET", secret)

	srv, count, capture := anEchoServer(t)

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{
		URL:       srv.URL,
		SecretEnv: "WH_SECRET",
		Body:      `{"id":"{{uuid}}"}`,
	}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	<-done

	if count.Load() != 1 {
		t.Fatalf("calls: %d", count.Load())
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()

	id := capture.lastReq.Header.Get("webhook-id")
	ts := capture.lastReq.Header.Get("webhook-timestamp")
	sig := capture.lastReq.Header.Get("webhook-signature")

	if id == "" || ts == "" || sig == "" {
		t.Fatalf("missing signing headers: id=%q ts=%q sig=%q", id, ts, sig)
	}
	// Validate HMAC: v1, base64(HMAC-SHA256(secret, "<id>.<ts>.<body>"))
	body := string(capture.lastBody)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(id + "." + ts + "." + body))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !strings.HasPrefix(sig, "v1,") {
		t.Errorf("sig prefix: got %q", sig)
	}
	gotSig := strings.TrimPrefix(sig, "v1,")
	if gotSig != want {
		t.Errorf("HMAC mismatch:\n got %q\nwant %q", gotSig, want)
	}
	// timestamp must be a valid unix int
	if _, err := strconv.ParseInt(ts, 10, 64); err != nil {
		t.Errorf("timestamp not unix int: %v", err)
	}
}

func TestDispatcher_SigningSkippedWhenEnvEmpty(t *testing.T) {
	// Ensure WH_NOTSET is unset regardless of test environment.
	_ = os.Unsetenv("WH_NOTSET")

	srv, _, capture := anEchoServer(t)

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{
		URL:       srv.URL,
		SecretEnv: "WH_NOTSET",
	}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	<-done

	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.lastReq.Header.Get("webhook-id") != "" {
		t.Error("expected signing skipped (no webhook-id header)")
	}
	if capture.lastReq.Header.Get("webhook-signature") != "" {
		t.Error("expected signing skipped (no webhook-signature header)")
	}
}

func TestDispatcher_StopsOn2xx(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusAccepted) // 2xx
	}))
	defer srv.Close()

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{URL: srv.URL, MaxAttempts: 5, BaseMS: 1, MaxMS: 5}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	<-done

	if calls.Load() != 1 {
		t.Errorf("expected exactly 1 call (2xx stops retries), got %d", calls.Load())
	}
}

func TestDispatcher_RetriesOn4xx(t *testing.T) {
	// Spec says "stop on 2xx"; 4xx is not 2xx so retries continue up to MaxAttempts.
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{URL: srv.URL, MaxAttempts: 3, BaseMS: 1, MaxMS: 5}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	<-done

	if calls.Load() != 3 {
		t.Errorf("expected 3 attempts on 4xx, got %d", calls.Load())
	}
}

func TestDispatcher_RecordsTransportError(t *testing.T) {
	// Use an unreachable address (closed listener).
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close() // immediately close so URL is unreachable

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{URL: url, MaxAttempts: 2, BaseMS: 1, MaxMS: 5}

	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	<-done

	records := d.Snapshot()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].StatusCode != 0 {
		t.Errorf("expected statusCode 0 on transport error, got %d", records[0].StatusCode)
	}
	if records[0].Error == "" {
		t.Error("expected non-empty Error on transport failure")
	}
}

func TestDispatcher_DispatchIsAsync(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// slow handler to prove dispatch does not block the caller
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{URL: srv.URL, MaxAttempts: 1}

	start := time.Now()
	done := make(chan struct{})
	d.Dispatch(context.Background(), wh, "POST /x", nil, func() { close(done) })
	// Caller should not block on the slow server: return path is essentially instant.
	elapsed := time.Since(start)
	if elapsed > 20*time.Millisecond {
		t.Errorf("Dispatch should return promptly, took %s", elapsed)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("async dispatch never completed")
	}

	// And the ring should eventually have the record.
	waitFor(t, time.Second, func() bool { return len(d.Snapshot()) == 1 })
}

func TestBackoff_Bounds(t *testing.T) {
	// base=100ms, max=1000ms. Attempt index 0..6 → expected cap at 1000.
	for n := 0; n < 10; n++ {
		d := time.Duration(webhooks.ComputeBackoff(100, 1000, n)) * time.Millisecond
		if d < 0 {
			t.Errorf("attempt %d: negative delay %s", n, d)
		}
		if d > 1000*time.Millisecond {
			t.Errorf("attempt %d: delay %s exceeds max", n, d)
		}
	}
}

func TestBackoff_Caps(t *testing.T) {
	d := webhooks.ComputeBackoff(100, 1000, 20)
	if d > 1000 {
		t.Errorf("expected cap at 1000, got %d", d)
	}
}

func TestComputeBackoffSequence_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		base    int
		max     int
		attempt int
		maxExp  int
	}{
		{name: "first attempt", base: 100, max: 1000, attempt: 0, maxExp: 100},
		{name: "second attempt", base: 100, max: 1000, attempt: 1, maxExp: 200},
		{name: "third attempt", base: 100, max: 1000, attempt: 2, maxExp: 400},
		{name: "cap kicks in", base: 100, max: 1000, attempt: 5, maxExp: 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBackoffTrials(t, tt.base, tt.max, tt.attempt, tt.maxExp)
		})
	}
}

func runBackoffTrials(t *testing.T, base, maxMS, attempt, maxExp int) {
	t.Helper()
	for i := 0; i < 50; i++ {
		got := webhooks.ComputeBackoff(base, maxMS, attempt)
		if got < 0 || got > maxExp {
			t.Fatalf("attempt %d trial %d: got %d out of [0, %d]", attempt, i, got, maxExp)
		}
	}
}

func TestDispatcher_NilOnCompleteIsTolerated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{URL: srv.URL, MaxAttempts: 1}

	d.Dispatch(context.Background(), wh, "POST /x", nil, nil)
	waitFor(t, 2*time.Second, func() bool { return len(d.Snapshot()) == 1 })
}

// TestDispatcher_ParentContextTimeoutBoundsDelivery verifies the WithTimeout
// path that the router relies on (bean unimock-ae8a AC #4). A short-timeout
// parent context is passed to Dispatch; the receiver hangs and accepts but
// never responds. The dispatch goroutine MUST terminate within the timeout
// window instead of leaking forever. This is the "test WithTimeout path
// directly" approach the bean recommends since webhookDispatchTimeout is a
// 60-second package const that cannot be overridden per-call.
func TestDispatcher_ParentContextTimeoutBoundsDelivery(t *testing.T) {
	// Receiver that accepts the connection, reads the request, then blocks
	// until its own context is canceled (which happens when the client closes
	// the request). This is a real hang — no 200 response is ever written.
	hanging := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer hanging.Close()

	d := webhooks.NewDispatcher(quietLogger(), webhooks.NewRing(10))
	wh := &model.WebhookConfig{
		URL:         hanging.URL,
		Method:      "POST",
		MaxAttempts: 1,
		BaseMS:      1,
		MaxMS:       5,
	}

	// Use a tight 300ms parent timeout so the test runs in well under a second.
	// This mirrors the contract the router relies on: dispatchCtx has a
	// deadline that bounds the async goroutine lifetime.
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	d.Dispatch(ctx, wh, "POST /hang", []byte(`{}`), func() { close(done) })

	select {
	case <-done:
		// Delivery terminated within the timeout window.
	case <-time.After(3 * time.Second):
		t.Fatal("dispatch goroutine did not terminate within 3s of parent ctx timeout")
	}

	records := d.Snapshot()
	if len(records) == 0 {
		t.Fatal("expected at least one delivery record after bounded dispatch")
	}
	if records[0].Error == "" {
		t.Errorf("expected non-empty error after ctx timeout, got %+v", records[0])
	}
	if !strings.Contains(strings.ToLower(records[0].Error), "context") &&
		!strings.Contains(strings.ToLower(records[0].Error), "deadline") &&
		!strings.Contains(strings.ToLower(records[0].Error), "canceled") {
		t.Errorf("expected ctx-derived error, got %q", records[0].Error)
	}
}
