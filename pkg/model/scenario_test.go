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
	if (tt.scenario.Webhook == nil) != tt.wantNil {
		t.Fatalf("Webhook nil mismatch: got %v, want nil=%v", tt.scenario.Webhook, tt.wantNil)
	}
	if tt.scenario.Webhook == nil {
		return
	}
	if tt.scenario.Webhook.URL != tt.wantURL {
		t.Errorf("URL: got %q, want %q", tt.scenario.Webhook.URL, tt.wantURL)
	}
	if tt.wantMethod != "" && tt.scenario.Webhook.Method != tt.wantMethod {
		t.Errorf("Method: got %q, want %q", tt.scenario.Webhook.Method, tt.wantMethod)
	}
	if tt.wantBody != "" && tt.scenario.Webhook.Body != tt.wantBody {
		t.Errorf("Body: got %q, want %q", tt.scenario.Webhook.Body, tt.wantBody)
	}
	if tt.wantMaxAttempt != 0 && tt.scenario.Webhook.MaxAttempts != tt.wantMaxAttempt {
		t.Errorf("MaxAttempts: got %d, want %d", tt.scenario.Webhook.MaxAttempts, tt.wantMaxAttempt)
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
