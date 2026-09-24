package model

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
