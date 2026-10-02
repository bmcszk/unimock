package model

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Validate runs the v1 contract validation on the scenario. It checks the
// Webhook (URL scheme + host, method allowlist with default, non-negative
// retry fields, and upper bounds) and the Stream (format allowlist,
// termination mode exactly-one, template non-empty and free of \n/\r, and
// the new hold_open-requires-interval constraint). It also enforces the
// stream/data mutual-exclusion rule. The error strings are intended to be
// reused by callers that prefix them with "stream: " / "webhook: " so the
// YAML-config-level error text stays stable.
//
// This validator is the single source of truth shared by YAML load and
// runtime scenario create/update (bean unimock-vhxx MAJOR4).
func (s *Scenario) Validate() error {
	if err := s.Webhook.Validate(); err != nil {
		return err
	}
	if err := s.Stream.Validate(); err != nil {
		return err
	}
	if s.Stream != nil && strings.TrimSpace(s.Data) != "" {
		return errors.New("stream and data are mutually exclusive")
	}
	return nil
}

// Validate enforces the webhook contract: URL parses, scheme+host present,
// method allowed and default POST on empty, retry fields non-negative with
// upper bounds (bounds cap the derived dispatch timeout ~4min; see
// internal/webhooks.DispatchTimeout). Returned errors are unprefixed — callers
// that need the "webhook: ..." prefix should wrap with fmt.Errorf("webhook: %w", err).
func (w *WebhookConfig) Validate() error {
	if w == nil {
		return nil
	}
	if err := w.ValidateURL(); err != nil {
		return err
	}
	if err := w.ValidateMethod(); err != nil {
		return err
	}
	return w.ValidateRetries()
}

// ValidateURL checks the URL field parses and has a scheme + host. Errors are
// unprefixed; callers wrap with "webhook: %w".
func (w *WebhookConfig) ValidateURL() error {
	if strings.TrimSpace(w.URL) == "" {
		return errors.New("url is required")
	}
	parsed, err := url.Parse(w.URL)
	if err != nil {
		return fmt.Errorf("invalid url %q: %w", w.URL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid url %q: must include scheme and host", w.URL)
	}
	return nil
}

// ValidateMethod enforces the POST/PUT/PATCH allowlist and normalizes empty
// method to POST. The reported error reflects the original (un-normalized)
// method so the user sees what they wrote.
func (w *WebhookConfig) ValidateMethod() error {
	method := strings.ToUpper(strings.TrimSpace(w.Method))
	if method == "" {
		return nil
	}
	if !isAllowedWebhookMethod(method) {
		return fmt.Errorf("method %q not allowed (must be POST, PUT, or PATCH)", w.Method)
	}
	return nil
}

// isAllowedWebhookMethod reports whether method is in the POST/PUT/PATCH set.
// Defined as a package-private helper so the model has no net/http import.
func isAllowedWebhookMethod(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH":
		return true
	default:
		return false
	}
}

// MaxAttemptsCeiling bounds MaxAttempts on the derived dispatch deadline.
// 12 attempts × 10s client timeout + 11 × 60s backoff cap ≈ 4.5min.
const MaxAttemptsCeiling = 12

// MaxBackoffMillees bound cap on the per-attempt backoff in ms.
const MaxBackoffMillees = 60000

// ValidateRetries checks MaxAttempts / BaseMS / MaxMS are non-negative and
// within the upper bounds that keep the derived dispatch timeout sensible.
// Errors are unprefixed.
func (w *WebhookConfig) ValidateRetries() error {
	if w.MaxAttempts < 0 {
		return fmt.Errorf("maxAttempts must be >= 0, got %d", w.MaxAttempts)
	}
	if w.MaxAttempts > MaxAttemptsCeiling {
		return fmt.Errorf("maxAttempts must be <= %d, got %d", MaxAttemptsCeiling, w.MaxAttempts)
	}
	if w.BaseMS < 0 {
		return fmt.Errorf("baseMs must be >= 0, got %d", w.BaseMS)
	}
	if w.MaxMS < 0 {
		return fmt.Errorf("maxMs must be >= 0, got %d", w.MaxMS)
	}
	if w.MaxMS > MaxBackoffMillees {
		return fmt.Errorf("maxMs must be <= %d, got %d", MaxBackoffMillees, w.MaxMS)
	}
	return nil
}

// allowedStreamFormats is the closed set of streaming formats supported by v1.
var allowedStreamFormats = map[string]struct{}{
	"sse":    {},
	"ndjson": {},
}

// Validate enforces the stream configuration contract: format allowlist,
// IntervalMS >= 0, Template non-empty and free of newline/carriage-return,
// and exactly-one of EventCount > 0 / HoldOpen, with the extra rule that
// HoldOpen requires IntervalMS > 0 (would otherwise spin a hot loop).
// Errors are unprefixed; callers wrap with "stream: %w".
func (s *StreamConfig) Validate() error {
	if s == nil {
		return nil
	}
	if err := s.validateFormat(); err != nil {
		return err
	}
	if s.IntervalMS < 0 {
		return fmt.Errorf("interval_ms must be >= 0, got %d", s.IntervalMS)
	}
	if err := s.validateTemplate(); err != nil {
		return err
	}
	return s.validateTermination()
}

// validateFormat checks the stream format is on the allowlist.
func (s *StreamConfig) validateFormat() error {
	if _, ok := allowedStreamFormats[s.Format]; !ok {
		return fmt.Errorf("format %q not allowed (must be sse or ndjson)", s.Format)
	}
	return nil
}

// validateTemplate checks the template is non-empty and single-line
// (newlines would break SSE data: and NDJSON line framing).
func (s *StreamConfig) validateTemplate() error {
	if strings.TrimSpace(s.Template) == "" {
		return errors.New("template is required")
	}
	if strings.ContainsAny(s.Template, "\n\r") {
		return errors.New("template must not contain newline or carriage-return characters")
	}
	return nil
}

// validateTermination checks exactly one of event_count>0 / hold_open is set
// and hold_open carries a positive interval (a zero interval would hot-loop).
func (s *StreamConfig) validateTermination() error {
	if (s.EventCount > 0) == s.HoldOpen {
		return fmt.Errorf(
			"exactly one of event_count>0 or hold_open=true must be set "+
				"(got event_count=%d, hold_open=%v)", s.EventCount, s.HoldOpen,
		)
	}
	if s.HoldOpen && s.IntervalMS <= 0 {
		return errors.New("hold_open requires interval_ms > 0")
	}
	return nil
}

// Scenario represents a predefined mock scenario for specific API requests
// Scenarios allow bypassing the normal mocking behavior for certain paths,
// enabling precise control over specific API responses.
type Scenario struct {
	// UUID is the unique identifier for the scenario
	// If not provided when creating, a UUID will be generated automatically
	UUID string `json:"uuid,omitempty"`

	// RequestPath defines which requests this scenario handles
	// Format: "METHOD /path" (e.g., "GET /api/users" or "POST /orders")
	// The path portion can contain wildcards (e.g., "GET /users/*")
	RequestPath string `json:"requestPath"`

	// StatusCode is the HTTP status code to return (e.g., 200, 201, 404, 500)
	StatusCode int `json:"statusCode"`

	// ContentType is the MIME type of the response (e.g., "application/json")
	ContentType string `json:"contentType"`

	// Location is the optional Location header value
	// Usually used with 201 Created responses
	Location string `json:"location,omitempty"`

	// Data is the response body to return
	// For JSON responses, this should be a valid JSON string
	Data string `json:"data"`

	// Headers is a map of HTTP headers to return with the scenario response
	Headers map[string]string `json:"headers,omitempty"`

	// Webhook is the optional outbound webhook to fire after this scenario matches.
	// When nil, no webhook is dispatched. v1 fires exactly one async HTTP request
	// after the scenario response is sent.
	Webhook *WebhookConfig `json:"webhook,omitempty"`

	// Stream is the optional server-generated stream response. When non-nil, the
	// router emits interval-spaced SSE or NDJSON frames instead of writing Data.
	// Mutually exclusive with Data at config validation time.
	Stream *StreamConfig `json:"stream,omitempty"`
}

// StreamConfig describes a server-generated stream response. v1 supports two
// formats (sse, ndjson) and two termination modes (EventCount frames OR HoldOpen
// until client disconnect). The Template is rendered per frame with the
// following placeholders substituted:
//   - {{index}}     0-based frame index (int)
//   - {{timestamp}} current time formatted as RFC3339Nano
type StreamConfig struct {
	// Format is "sse" (text/event-stream) or "ndjson" (application/x-ndjson).
	Format string `json:"format"`

	// IntervalMS is the delay between frames in milliseconds. Zero means no sleep.
	IntervalMS int `json:"intervalMs,omitempty"`

	// EventCount is the number of frames to emit before closing the stream.
	// Exactly one of EventCount>0 or HoldOpen=true must be set at config validation.
	EventCount int `json:"eventCount,omitempty"`

	// HoldOpen, when true, keeps the stream open until the client disconnects
	// (the request context is canceled).
	HoldOpen bool `json:"holdOpen,omitempty"`

	// Template is the per-frame payload template. Must be non-empty at config
	// validation. The template is rendered verbatim per frame; only the
	// documented placeholders are substituted.
	Template string `json:"template"`
}

// WebhookConfig describes the outbound webhook triggered when a scenario matches.
// The dispatcher reads the secret from the environment variable named by SecretEnv
// at delivery time, never accepting the secret value inline.
type WebhookConfig struct {
	// URL is the absolute target URL the webhook is delivered to.
	// Must be a syntactically valid URL (validated at config load).
	URL string `json:"url"`

	// Method is the HTTP method used for delivery.
	// Defaults to POST when empty. Only POST/PUT/PATCH are allowed.
	Method string `json:"method,omitempty"`

	// Headers are extra HTTP headers sent with each delivery attempt.
	Headers map[string]string `json:"headers,omitempty"`

	// Body is the request body sent to the target URL.
	// When non-empty, "{{uuid}}" is replaced per attempt with a fresh UUID.
	// When empty, the matched request body is used.
	Body string `json:"body,omitempty"`

	// SecretEnv is the name of the environment variable that holds the HMAC
	// secret used for Standard Webhooks signing. The secret value itself is
	// never stored inline. When empty or the env var is unset, signing is skipped.
	SecretEnv string `json:"secretEnv,omitempty"`

	// MaxAttempts is the maximum total delivery attempts including the first.
	// Zero value is treated as the default (3).
	MaxAttempts int `json:"maxAttempts,omitempty"`

	// BaseMS is the base backoff in milliseconds. Zero value is treated as default (500).
	BaseMS int `json:"baseMs,omitempty"`

	// MaxMS caps the backoff delay between attempts in milliseconds.
	// Zero value is treated as default (30000).
	MaxMS int `json:"maxMs,omitempty"`
}
