package model_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/pkg/model"
)

func TestScenario_Webhook_FieldExists(t *testing.T) {
	tests := []struct {
		name           string
		scenario       model.Scenario
		wantNil        bool
		wantURL        string
		wantMethod     string
		wantBody       string
		wantMaxAttempt int
	}{
		{
			name:     "no webhook configured",
			scenario: model.Scenario{UUID: "x", RequestPath: "GET /x"},
			wantNil:  true,
		},
		{
			name: "webhook with url only",
			scenario: model.Scenario{
				UUID:        "x",
				RequestPath: "GET /x",
				Webhook:     &model.WebhookConfig{URL: "http://example.com/hook"},
			},
			wantNil: false,
			wantURL: "http://example.com/hook",
		},
		{
			name: "webhook with full configuration",
			scenario: model.Scenario{
				UUID:        "x",
				RequestPath: "GET /x",
				Webhook: &model.WebhookConfig{
					URL:         "http://example.com/hook",
					Method:      "PUT",
					Headers:     map[string]string{"X-Foo": "bar"},
					Body:        `{"id":"{{uuid}}"}`,
					SecretEnv:   "MY_SECRET",
					MaxAttempts: 5,
					BaseMS:      200,
					MaxMS:       10000,
				},
			},
			wantNil:        false,
			wantURL:        "http://example.com/hook",
			wantMethod:     "PUT",
			wantBody:       `{"id":"{{uuid}}"}`,
			wantMaxAttempt: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runScenarioWebhookFieldChecks(t, tt)
		})
	}
}

func runScenarioWebhookFieldChecks(t *testing.T, tt struct {
	name           string
	scenario       model.Scenario
	wantNil        bool
	wantURL        string
	wantMethod     string
	wantBody       string
	wantMaxAttempt int
}) {
	t.Helper()
	checkScenarioWebhookNil(t, tt.scenario.Webhook, tt.wantNil)
	if tt.scenario.Webhook == nil {
		return
	}
	checkScenarioWebhookURL(t, tt.scenario.Webhook.URL, tt.wantURL)
	checkScenarioWebhookOptionalFields(
		t, tt.scenario.Webhook, tt.wantMethod, tt.wantBody, tt.wantMaxAttempt,
	)
}

func checkScenarioWebhookNil(t *testing.T, got *model.WebhookConfig, wantNil bool) {
	t.Helper()
	if (got == nil) != wantNil {
		t.Fatalf("Webhook nil mismatch: got %v, want nil=%v", got, wantNil)
	}
}

func checkScenarioWebhookURL(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("URL: got %q, want %q", got, want)
	}
}

func checkScenarioWebhookOptionalFields(
	t *testing.T, w *model.WebhookConfig,
	wantMethod, wantBody string, wantMaxAttempt int,
) {
	t.Helper()
	if wantMethod != "" && w.Method != wantMethod {
		t.Errorf("Method: got %q, want %q", w.Method, wantMethod)
	}
	if wantBody != "" && w.Body != wantBody {
		t.Errorf("Body: got %q, want %q", w.Body, wantBody)
	}
	if wantMaxAttempt != 0 && w.MaxAttempts != wantMaxAttempt {
		t.Errorf("MaxAttempts: got %d, want %d", w.MaxAttempts, wantMaxAttempt)
	}
}

func TestScenario_Webhook_JSONRoundTrip(t *testing.T) {
	scenario := model.Scenario{
		UUID:        "s1",
		RequestPath: "POST /orders",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"ok":true}`,
		Webhook: &model.WebhookConfig{
			URL:         "http://example.com/hook",
			Method:      "POST",
			Headers:     map[string]string{"X-Foo": "bar"},
			Body:        `{"id":"{{uuid}}"}`,
			SecretEnv:   "WH_SECRET",
			MaxAttempts: 4,
			BaseMS:      250,
			MaxMS:       5000,
		},
	}

	data, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded model.Scenario
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Webhook == nil {
		t.Fatal("webhook lost in round-trip")
	}
	if decoded.Webhook.URL != scenario.Webhook.URL {
		t.Errorf("URL: got %q, want %q", decoded.Webhook.URL, scenario.Webhook.URL)
	}
	if decoded.Webhook.SecretEnv != scenario.Webhook.SecretEnv {
		t.Errorf("SecretEnv: got %q, want %q", decoded.Webhook.SecretEnv, scenario.Webhook.SecretEnv)
	}
	if decoded.Webhook.Headers["X-Foo"] != "bar" {
		t.Errorf("Headers: got %v", decoded.Webhook.Headers)
	}
	if decoded.Webhook.MaxAttempts != 4 {
		t.Errorf("MaxAttempts: got %d, want 4", decoded.Webhook.MaxAttempts)
	}
}

func TestScenario_Webhook_JSONOmitempty(t *testing.T) {
	scenario := model.Scenario{
		UUID:        "s2",
		RequestPath: "GET /x",
	}

	data, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if strings.Contains(string(data), `"webhook"`) {
		t.Errorf("expected no webhook key in JSON when nil, got: %s", string(data))
	}
}

type streamFieldCase struct {
	name         string
	scenario     model.Scenario
	wantNil      bool
	wantFormat   string
	wantTemplate string
	wantCount    int
	wantHold     bool
	wantInterval int
}

func TestScenario_Stream_FieldExists(t *testing.T) {
	tests := []streamFieldCase{
		{
			name:     "no stream configured",
			scenario: model.Scenario{UUID: "x", RequestPath: "GET /x"},
			wantNil:  true,
		},
		{
			name: "stream sse with event count",
			scenario: model.Scenario{
				UUID:        "x",
				RequestPath: "GET /events",
				Stream: &model.StreamConfig{
					Format:     "sse",
					EventCount: 5,
					IntervalMS: 100,
					Template:   "frame-{{index}}",
				},
			},
			wantNil:      false,
			wantFormat:   "sse",
			wantTemplate: "frame-{{index}}",
			wantCount:    5,
			wantInterval: 100,
		},
		{
			name: "stream ndjson hold open",
			scenario: model.Scenario{
				UUID:        "x",
				RequestPath: "GET /ticks",
				Stream: &model.StreamConfig{
					Format:   "ndjson",
					HoldOpen: true,
					Template: `{"n":{{index}}}`,
				},
			},
			wantNil:      false,
			wantFormat:   "ndjson",
			wantTemplate: `{"n":{{index}}}`,
			wantHold:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runStreamFieldChecks(t, tt)
		})
	}
}

func runStreamFieldChecks(t *testing.T, tt streamFieldCase) {
	t.Helper()
	if (tt.scenario.Stream == nil) != tt.wantNil {
		t.Fatalf("Stream nil mismatch: got %v, want nil=%v", tt.scenario.Stream, tt.wantNil)
	}
	if tt.scenario.Stream == nil {
		return
	}
	if tt.scenario.Stream.Format != tt.wantFormat {
		t.Errorf("Format: got %q, want %q", tt.scenario.Stream.Format, tt.wantFormat)
	}
	if tt.scenario.Stream.Template != tt.wantTemplate {
		t.Errorf("Template: got %q, want %q", tt.scenario.Stream.Template, tt.wantTemplate)
	}
	if tt.scenario.Stream.EventCount != tt.wantCount {
		t.Errorf("EventCount: got %d, want %d", tt.scenario.Stream.EventCount, tt.wantCount)
	}
	if tt.scenario.Stream.HoldOpen != tt.wantHold {
		t.Errorf("HoldOpen: got %v, want %v", tt.scenario.Stream.HoldOpen, tt.wantHold)
	}
	if tt.scenario.Stream.IntervalMS != tt.wantInterval {
		t.Errorf("IntervalMS: got %d, want %d", tt.scenario.Stream.IntervalMS, tt.wantInterval)
	}
}

func TestScenario_Stream_JSONRoundTrip(t *testing.T) {
	scenario := model.Scenario{
		UUID:        "s-stream",
		RequestPath: "GET /stream",
		StatusCode:  200,
		Stream: &model.StreamConfig{
			Format:     "sse",
			IntervalMS: 250,
			EventCount: 3,
			HoldOpen:   false,
			Template:   "tick-{{index}}-{{timestamp}}",
		},
	}

	data, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded model.Scenario
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Stream == nil {
		t.Fatal("stream lost in round-trip")
	}
	if decoded.Stream.Format != scenario.Stream.Format {
		t.Errorf("Format: got %q, want %q", decoded.Stream.Format, scenario.Stream.Format)
	}
	if decoded.Stream.Template != scenario.Stream.Template {
		t.Errorf("Template: got %q, want %q", decoded.Stream.Template, scenario.Stream.Template)
	}
	if decoded.Stream.IntervalMS != scenario.Stream.IntervalMS {
		t.Errorf("IntervalMS: got %d, want %d", decoded.Stream.IntervalMS, scenario.Stream.IntervalMS)
	}
	if decoded.Stream.EventCount != scenario.Stream.EventCount {
		t.Errorf("EventCount: got %d, want %d", decoded.Stream.EventCount, scenario.Stream.EventCount)
	}
}

func TestScenario_Stream_JSONOmitempty(t *testing.T) {
	scenario := model.Scenario{
		UUID:        "s-no-stream",
		RequestPath: "GET /x",
	}

	data, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if strings.Contains(string(data), `"stream"`) {
		t.Errorf("expected no stream key in JSON when nil, got: %s", string(data))
	}
}
