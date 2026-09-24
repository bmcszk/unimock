package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "unimock-webhook-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

func TestScenarioConfig_ToModelScenario_WebhookPassThrough(t *testing.T) {
	sc := config.ScenarioConfig{
		UUID:        "wh-1",
		Method:      "POST",
		Path:        "/orders",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"ok":true}`,
		Webhook: &config.WebhookConfig{
			URL:         "http://example.com/hook",
			Method:      "PUT",
			Body:        `{"id":"{{uuid}}"}`,
			SecretEnv:   "MY_SECRET",
			MaxAttempts: 5,
			BaseMS:      100,
			MaxMS:       2000,
			Headers:     map[string]string{"X-Foo": "bar"},
		},
	}

	got := sc.ToModelScenario(nil)

	if got.Webhook == nil {
		t.Fatal("webhook not propagated to model")
	}
	if got.Webhook.URL != "http://example.com/hook" {
		t.Errorf("URL: got %q", got.Webhook.URL)
	}
	if got.Webhook.Method != "PUT" {
		t.Errorf("Method: got %q", got.Webhook.Method)
	}
	if got.Webhook.Headers["X-Foo"] != "bar" {
		t.Errorf("Headers: got %v", got.Webhook.Headers)
	}
}

func TestScenarioConfig_ToModelScenario_WebhookNil(t *testing.T) {
	sc := config.ScenarioConfig{
		Method:     "GET",
		Path:       "/x",
		StatusCode: 200,
	}
	got := sc.ToModelScenario(nil)
	if got.Webhook != nil {
		t.Errorf("expected nil webhook, got %+v", got.Webhook)
	}
}

func TestUniConfig_LoadFromYAML_WebhookValid(t *testing.T) {
	yaml := `
scenarios:
  - uuid: "wh-ok"
    method: "POST"
    path: "/orders"
    status_code: 201
    data: '{"ok":true}'
    webhook:
      url: "http://example.com/hook"
      method: POST
      body: '{"id":"{{uuid}}"}'
      secret_env: MY_SECRET
      max_attempts: 4
      base_ms: 250
      max_ms: 5000
      headers:
        X-Foo: bar
`
	path := writeYAML(t, yaml)
	defer os.Remove(path)

	cfg, err := config.LoadFromYAML(path)
	if err != nil {
		t.Fatalf("LoadFromYAML: %v", err)
	}
	if len(cfg.Scenarios) != 1 {
		t.Fatalf("scenarios len = %d", len(cfg.Scenarios))
	}
	wh := cfg.Scenarios[0].Webhook
	if wh == nil {
		t.Fatal("webhook missing from loaded config")
	}
	if wh.URL != "http://example.com/hook" {
		t.Errorf("URL: got %q", wh.URL)
	}
	if wh.MaxAttempts != 4 {
		t.Errorf("MaxAttempts: got %d", wh.MaxAttempts)
	}
	if wh.Headers["X-Foo"] != "bar" {
		t.Errorf("Headers: got %v", wh.Headers)
	}
}

func TestUniConfig_LoadFromYAML_WebhookValidationErrors(t *testing.T) {
	type tc struct {
		name, webhook, errSubstr string
	}
	tests := []tc{
		{name: "missing url", webhook: webhookMissingURL, errSubstr: "url"},
		{name: "invalid url", webhook: webhookInvalidURL, errSubstr: "url"},
		{name: "disallowed method GET", webhook: webhookMethodGET, errSubstr: "method"},
		{name: "disallowed method DELETE", webhook: webhookMethodDELETE, errSubstr: "method"},
		{name: "negative max_attempts", webhook: webhookNegativeAttempts, errSubstr: "maxAttempts"},
		{name: "negative base_ms", webhook: webhookNegativeBase, errSubstr: "baseMs"},
		{name: "negative max_ms", webhook: webhookNegativeMax, errSubstr: "maxMs"},
		{name: "inline secret forbidden", webhook: webhookInlineSecret, errSubstr: "secret"},
		{name: "unknown field", webhook: webhookUnknownField, errSubstr: "strange"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWebhookLoadError(t, tt.webhook, tt.errSubstr)
		})
	}
}

// Per-test YAML fragments. Kept as package-level vars to keep the test table short.
const (
	webhookMissingURL     = "\n      method: POST\n      body: '{}'\n"
	webhookInvalidURL     = "\n      url: \"://no-scheme\"\n      method: POST\n"
	webhookMethodGET      = "\n      url: \"http://example.com\"\n      method: GET\n"
	webhookMethodDELETE   = "\n      url: \"http://example.com\"\n      method: DELETE\n"
	webhookNegativeAttempts = "\n      url: \"http://example.com\"\n      max_attempts: -1\n"
	webhookNegativeBase   = "\n      url: \"http://example.com\"\n      base_ms: -10\n"
	webhookNegativeMax    = "\n      url: \"http://example.com\"\n      max_ms: -1\n"
	webhookInlineSecret   = "\n      url: \"http://example.com\"\n      secret: \"shhh\"\n"
	webhookUnknownField   = "\n      url: \"http://example.com\"\n      strange: true\n"
)

func runWebhookLoadError(t *testing.T, webhook, errSubstr string) {
	t.Helper()
	yaml := `
scenarios:
  - method: POST
    path: "/orders"
    status_code: 201
    webhook:` + webhook
	path := writeYAML(t, yaml)
	defer os.Remove(path)
	_, err := config.LoadFromYAML(path)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(errSubstr)) {
		t.Errorf("error %q does not contain %q", err.Error(), errSubstr)
	}
}

func TestUniConfig_LoadFromYAML_WebhookValid_AcceptsPOSTPUTPATCH(t *testing.T) {
	for _, m := range []string{"POST", "PUT", "PATCH", "post", "put"} {
		t.Run(m, func(t *testing.T) {
			yaml := `
scenarios:
  - method: POST
    path: "/x"
    webhook:
      url: "http://example.com"
      method: ` + m + `
`
			path := writeYAML(t, yaml)
			defer os.Remove(path)
			if _, err := config.LoadFromYAML(path); err != nil {
				t.Errorf("method %q unexpectedly rejected: %v", m, err)
			}
		})
	}
}

func TestScenarioConfig_ToModelScenario_HeadersAndBodyFlow(t *testing.T) {
	wh := &config.WebhookConfig{
		URL:     "http://example.com",
		Body:    `{"a":"{{uuid}}"}`,
		Headers: map[string]string{"X-H": "1"},
	}
	sc := config.ScenarioConfig{Method: "POST", Path: "/x", Webhook: wh}
	got := sc.ToModelScenario(nil)
	if got.Webhook == nil {
		t.Fatal("nil webhook")
	}
	if got.Webhook.Headers["X-H"] != "1" {
		t.Errorf("Headers: %v", got.Webhook.Headers)
	}
	if got.Webhook.Body != `{"a":"{{uuid}}"}` {
		t.Errorf("Body: %q", got.Webhook.Body)
	}
}

// silence unused import warnings when the file is stripped
var _ = model.Scenario{}
