package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bmcszk/unimock/pkg/model"
	"github.com/google/uuid"
)

// default backoff parameters.
const (
	defaultMaxAttempts = 3
	defaultBaseMS      = 500
	defaultMaxMS       = 30000
	httpClientTimeout  = 10 * time.Second
	uuidPlaceholder    = "{{uuid}}"
)

// Dispatcher delivers scenario-triggered webhooks asynchronously with retry and
// optional Standard Webhooks signing. Each Dispatch call spawns its own goroutine
// so the caller's request path is never blocked by network IO.
type Dispatcher struct {
	logger  *slog.Logger
	ring    *Ring
	client  *http.Client
	now     func() time.Time
	wait    func(context.Context, time.Duration) // ctx-aware backoff wait seam
	randInt func(n int) int                      // random source for backoff jitter, returns int in [0, n)
}

// NewDispatcher creates a Dispatcher using the provided logger and delivery ring.
// The HTTP client uses a fixed 10s Timeout (no zero-value client).
func NewDispatcher(logger *slog.Logger, ring *Ring) *Dispatcher {
	return &Dispatcher{
		logger: logger,
		ring:   ring,
		client: &http.Client{Timeout: httpClientTimeout},
		now:    time.Now,
		wait:   waitCtxOrTimeout,
		randInt: func(n int) int {
			if n <= 0 {
				return 0
			}
			return rand.Intn(n)
		},
	}
}

// waitCtxOrTimeout blocks for d or until ctx is done, whichever comes first.
// Context-aware waits keep delivery goroutines from lingering a full backoff
// interval after the dispatch deadline has already fired.
func waitCtxOrTimeout(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// Dispatch delivers a webhook asynchronously. It returns immediately; the optional
// onComplete callback fires when delivery has terminated (success or final failure).
// requestBody is the body of the matched HTTP request and is used as the payload
// when the webhook config does not specify a Body.
// requestPath is the matched scenario request path (e.g. "POST /orders") included
// in logs for traceability.
//
// Dispatch runs on its own goroutine; the goroutine outlives the caller's stack.
// When the supplied config has a blank URL or is nil, onComplete is still
// invoked exactly once before Dispatch returns, matching the contract for
// well-formed inputs (MINOR5).
func (d *Dispatcher) Dispatch(
	parentCtx context.Context,
	wh *model.WebhookConfig,
	requestPath string,
	requestBody []byte,
	onComplete func(),
) {
	if !dispatchable(wh) {
		if onComplete != nil {
			onComplete()
		}
		return
	}
	cfg := normalizeConfig(wh)
	go func() {
		defer func() {
			if onComplete != nil {
				onComplete()
			}
		}()
		d.deliver(parentCtx, cfg, requestPath, requestBody)
	}()
}

// dispatchable reports whether a webhook config can be dispatched at all
// (non-nil with a non-blank URL). Used by Dispatch to short-circuit configs
// that can never produce a delivery.
func dispatchable(wh *model.WebhookConfig) bool {
	return wh != nil && strings.TrimSpace(wh.URL) != ""
}

// DispatchTimeout returns the upper bound on a single webhook delivery's
// lifetime: worst-case retry envelope (MaxAttempts HTTP attempts at
// httpClientTimeout each, plus MaxAttempts-1 backoff waits capped at MaxMS)
// plus margin. The router uses it as the dispatch context deadline so the
// deadline provably covers the retry envelope instead of a fixed guess.
// Validation bounds (model.WebhookConfig.Validate: MaxAttempts<=12,
// MaxMS<=60000) keep the derived timeout at roughly 4.5 minutes worst case.
func DispatchTimeout(wh *model.WebhookConfig) time.Duration {
	cfg := normalizeConfig(wh)
	envelope := time.Duration(cfg.MaxAttempts)*httpClientTimeout +
		time.Duration(cfg.MaxAttempts-1)*time.Duration(cfg.MaxMS)*time.Millisecond
	return envelope + dispatchTimeoutMargin
}

// dispatchTimeoutMargin adds slack to the derived retry envelope for
// scheduling, body rendering, and signing overhead.
const dispatchTimeoutMargin = 5 * time.Second

// normalizeConfig applies defaults for fields the contract promises to handle
// (MaxAttempts, BaseMS, MaxMS, Method).
func normalizeConfig(wh *model.WebhookConfig) *model.WebhookConfig {
	out := *wh
	if out.MaxAttempts <= 0 {
		out.MaxAttempts = defaultMaxAttempts
	}
	if out.BaseMS <= 0 {
		out.BaseMS = defaultBaseMS
	}
	if out.MaxMS <= 0 {
		out.MaxMS = defaultMaxMS
	}
	if strings.TrimSpace(out.Method) == "" {
		out.Method = http.MethodPost
	}
	out.Method = strings.ToUpper(out.Method)
	return &out
}

// deliver executes the retry loop until success, context cancel, or attempts exhausted.
func (d *Dispatcher) deliver(
	parentCtx context.Context,
	wh *model.WebhookConfig,
	requestPath string,
	requestBody []byte,
) {
	deliveryID := uuid.NewString()
	lastDelivery := d.runAttempts(parentCtx, wh, requestBody, deliveryID)
	d.recordWithContext(lastDelivery, requestPath)
}

// runAttempts performs the retry loop body without the final record call.
func (d *Dispatcher) runAttempts(
	parentCtx context.Context,
	wh *model.WebhookConfig,
	requestBody []byte,
	deliveryID string,
) Delivery {
	var last Delivery
	for attempt := 1; attempt <= wh.MaxAttempts; attempt++ {
		if d.shouldAbort(parentCtx, deliveryID, wh.URL, attempt) {
			return d.canceledDelivery(deliveryID, wh.URL, attempt)
		}
		delivery := d.runSingleAttempt(parentCtx, wh, requestBody, deliveryID, attempt)
		if isSuccess(delivery) {
			return delivery
		}
		last = delivery
		if attempt == wh.MaxAttempts {
			break
		}
		d.wait(parentCtx, backoffDelay(d.randInt, wh, attempt))
	}
	return last
}

// backoffDelay computes the jittered backoff before the next retry attempt.
func backoffDelay(randInt func(int) int, wh *model.WebhookConfig, attempt int) time.Duration {
	jittered := jitter(randInt, ComputeBackoff(wh.BaseMS, wh.MaxMS, attempt-1))
	return time.Duration(jittered) * time.Millisecond
}

// shouldAbort returns true if the parent context is canceled. It exists as its own
// helper to keep runAttempts at low cognitive complexity.
func (*Dispatcher) shouldAbort(parentCtx context.Context, _, _ string, _ int) bool {
	return parentCtx != nil && parentCtx.Err() != nil
}

// isSuccess reports whether a delivery represents a successful (2xx) HTTP response.
func isSuccess(d Delivery) bool {
	return d.StatusCode >= 200 && d.StatusCode < 300 && d.Error == ""
}

// runSingleAttempt executes one HTTP attempt and fills attempt/url/ts metadata.
func (d *Dispatcher) runSingleAttempt(
	parentCtx context.Context,
	wh *model.WebhookConfig,
	requestBody []byte,
	deliveryID string,
	attempt int,
) Delivery {
	body := d.composeBody(wh, attempt, requestBody)
	delivery, err := d.attempt(parentCtx, wh, deliveryID, body)
	delivery.Attempt = attempt
	delivery.URL = wh.URL
	if delivery.TS.IsZero() {
		delivery.TS = d.now()
	}
	if err != nil && delivery.Error == "" {
		delivery.Error = err.Error()
	}
	return delivery
}

// canceledDelivery builds a Delivery record for an aborted context.
func (*Dispatcher) canceledDelivery(deliveryID, url string, attempt int) Delivery {
	return Delivery{
		ID:      deliveryID,
		URL:     url,
		Attempt: attempt,
		Error:   "context canceled",
		TS:      time.Now(),
	}
}

// ctxCanceled is intentionally left out; shouldAbort replaces it inline.

// attempt performs a single HTTP delivery and returns a Delivery with attempt metadata.
func (d *Dispatcher) attempt(
	parentCtx context.Context,
	wh *model.WebhookConfig,
	deliveryID string,
	body []byte,
) (Delivery, error) {
	attemptCtx := parentCtx
	if attemptCtx == nil {
		attemptCtx = context.Background()
	}
	req, err := http.NewRequestWithContext(attemptCtx, wh.Method, wh.URL, bytes.NewReader(body))
	if err != nil {
		return Delivery{ID: deliveryID, Error: fmt.Sprintf("build request: %s", err)}, err
	}
	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	d.applySigning(req, wh, body)

	resp, err := d.client.Do(req)
	if err != nil {
		return Delivery{ID: deliveryID, Error: err.Error()}, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Delivery{ID: deliveryID, StatusCode: resp.StatusCode,
			Error: fmt.Sprintf("non-2xx response: %d", resp.StatusCode)}, nil
	}
	return Delivery{ID: deliveryID, StatusCode: resp.StatusCode}, nil
}

// applySigning adds Standard Webhooks headers when SecretEnv is set and the env var
// resolves to a non-empty value. When the env var is empty/missing, signing is skipped.
//
// The signing secret is treated per the Standard Webhooks spec
// (https://github.com/standard-webhooks/standard-webhooks): if the env value
// starts with "whsec_", the remainder is base64-decoded (StdEncoding) and the
// raw bytes are used as the HMAC key. Otherwise, the env value is used as raw
// key bytes — preserved for back-compat with tests/users that supplied the
// plain-text secret directly.
//
// attemptWebhookID is a fresh UUID per attempt, so each attempt carries its
// own id header (the Standard Webhooks spec expects unique ids per delivery);
// the deliveryID itself (which is the per-delivery UUID) is recorded on the
// Delivery record and on the ring.
func (d *Dispatcher) applySigning(req *http.Request, wh *model.WebhookConfig, body []byte) {
	if strings.TrimSpace(wh.SecretEnv) == "" {
		return
	}
	secretRaw := os.Getenv(wh.SecretEnv)
	if secretRaw == "" {
		if d.logger != nil {
			d.logger.Warn("webhook signing skipped: env var empty",
				"secret_env", wh.SecretEnv, "url", wh.URL)
		}
		return
	}
	key, err := decodeWebhookSecret(secretRaw)
	if err != nil {
		if d.logger != nil {
			d.logger.Warn("webhook signing skipped: cannot decode secret",
				"secret_env", wh.SecretEnv, "url", wh.URL, "error", err)
		}
		return
	}
	// Per-attempt ID keeps each attempt individually signed.
	id := uuid.NewString()
	ts := strconvFormatUnix(d.now())
	mac := hmac.New(sha256.New, key)
	// hmac.Hash.Write never returns an error.
	_, _ = mac.Write([]byte(id))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(ts))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(body)
	sig := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("webhook-id", id)
	req.Header.Set("webhook-timestamp", ts)
	req.Header.Set("webhook-signature", sig)
}

// decodeWebhookSecret decodes a Standard Webhooks secret per spec:
//
//  1. If secretRaw starts with "whsec_", TrimPrefix and base64.StdEncoding-decode
//     the remainder and use those bytes as the HMAC key.
//  2. Otherwise fall back to the raw secret bytes (back-compat with plain-text
//     secrets used by earlier dispatcher tests/users).
//
// The function returns the resolved key bytes or a non-nil error when the
// whsec_-prefixed value is not valid base64. The plain-text fallback is
// intentional: secrets without the prefix are assumed to be raw bytes.
func decodeWebhookSecret(secretRaw string) ([]byte, error) {
	const prefix = "whsec_"
	if strings.HasPrefix(secretRaw, prefix) {
		encoded := strings.TrimPrefix(secretRaw, prefix)
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("decode whsec_ payload: %w", err)
		}
		return key, nil
	}
	return []byte(secretRaw), nil
}

// composeBody picks the configured Body (with {{uuid}} substitution) when present,
// otherwise falls back to the original request body. The attempt index is reserved
// for future per-attempt templating; today every attempt gets a fresh UUID.
func (*Dispatcher) composeBody(wh *model.WebhookConfig, _ int, requestBody []byte) []byte {
	template := wh.Body
	if template == "" {
		return requestBodyOrEmpty(requestBody)
	}
	return []byte(strings.ReplaceAll(template, uuidPlaceholder, uuid.NewString()))
}

// recordWithContext is like record but logs the matched scenario request path
// alongside the delivery, when present.
func (d *Dispatcher) recordWithContext(dlv Delivery, requestPath string) {
	if d.ring != nil {
		d.ring.Add(dlv)
	}
	d.logDelivery(dlv, requestPath)
}

// logDelivery emits the structured log line for a single delivery.
func (d *Dispatcher) logDelivery(dlv Delivery, requestPath string) {
	if d.logger == nil {
		return
	}
	attrs := []any{
		"id", dlv.ID,
		"url", dlv.URL,
		"attempt", dlv.Attempt,
		"status", dlv.StatusCode,
		"ts", dlv.TS,
	}
	if requestPath != "" {
		attrs = append(attrs, "scenario_request", requestPath)
	}
	if dlv.Error != "" {
		attrs = append(attrs, "error", dlv.Error)
	}
	if dlv.StatusCode >= 200 && dlv.StatusCode < 300 {
		d.logger.Info("webhook delivered", attrs...)
		return
	}
	d.logger.Warn("webhook delivery failed", attrs...)
}

// Snapshot returns the current delivery ring contents.
func (d *Dispatcher) Snapshot() []Delivery {
	if d.ring == nil {
		return []Delivery{}
	}
	return d.ring.Snapshot()
}

// ComputeBackoff returns the upper bound (in milliseconds) of the full-jitter delay
// for a given attempt: rand[0, min(maxMS, baseMS*2^attempt)].
// Exposed so tests can verify the cap behavior deterministically.
func ComputeBackoff(baseMS, maxMS, attempt int) int {
	if baseMS <= 0 {
		baseMS = defaultBaseMS
	}
	if maxMS <= 0 {
		maxMS = defaultMaxMS
	}
	if attempt < 0 {
		attempt = 0
	}
	delay := baseMS
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > maxMS {
			return maxMS
		}
	}
	if delay > maxMS {
		return maxMS
	}
	return delay
}

// jitter picks a random duration in [0, maxMs] using the dispatcher's randInt source.
// This implements the "full jitter" portion of the exponential backoff strategy.
func jitter(randInt func(int) int, maxMs int) int {
	if maxMs <= 0 {
		return 0
	}
	return randInt(maxMs + 1)
}

// strconvFormatUnix wraps strconv.FormatInt to keep imports clean in applySigning.
func strconvFormatUnix(t time.Time) string {
	return fmt.Sprintf("%d", t.Unix())
}

func requestBodyOrEmpty(b []byte) []byte {
	if b == nil {
		return []byte{}
	}
	return b
}
