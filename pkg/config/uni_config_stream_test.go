package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

func TestScenarioConfig_ToModelScenario_StreamPassThrough(t *testing.T) {
	sc := config.ScenarioConfig{
		UUID:        "st-1",
		Method:      "GET",
		Path:        "/events",
		StatusCode:  200,
		ContentType: "text/event-stream",
		Stream: &config.StreamConfig{
			Format:     "sse",
			IntervalMS: 50,
			EventCount: 4,
			Template:   "data: {\"i\":{{index}},\"t\":\"{{timestamp}}\"}\n\n",
		},
	}

	got := sc.ToModelScenario(nil)

	if got.Stream == nil {
		t.Fatal("stream not propagated to model")
	}
	if got.Stream.Format != "sse" {
		t.Errorf("Format: got %q", got.Stream.Format)
	}
	if got.Stream.IntervalMS != 50 {
		t.Errorf("IntervalMS: got %d", got.Stream.IntervalMS)
	}
	if got.Stream.EventCount != 4 {
		t.Errorf("EventCount: got %d", got.Stream.EventCount)
	}
	if !strings.HasPrefix(got.Stream.Template, "data: ") {
		t.Errorf("Template prefix: got %q", got.Stream.Template)
	}
}

func TestScenarioConfig_ToModelScenario_StreamNil(t *testing.T) {
	sc := config.ScenarioConfig{
		Method:     "GET",
		Path:       "/x",
		StatusCode: 200,
	}
	got := sc.ToModelScenario(nil)
	if got.Stream != nil {
		t.Errorf("expected nil stream, got %+v", got.Stream)
	}
}

func TestUniConfig_LoadFromYAML_StreamValid_SSEEventCount(t *testing.T) {
	yaml := `
scenarios:
  - uuid: "st-sse"
    method: "GET"
    path: "/events"
    status_code: 200
    stream:
      format: "sse"
      interval_ms: 100
      event_count: 5
      template: "data: {\"i\":{{index}}}\n\n"
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
	st := cfg.Scenarios[0].Stream
	if st == nil {
		t.Fatal("stream missing from loaded config")
	}
	if st.Format != "sse" {
		t.Errorf("Format: got %q", st.Format)
	}
	if st.IntervalMS != 100 {
		t.Errorf("IntervalMS: got %d", st.IntervalMS)
	}
	if st.EventCount != 5 {
		t.Errorf("EventCount: got %d", st.EventCount)
	}
	if st.Template == "" {
		t.Error("Template empty")
	}
}

func TestUniConfig_LoadFromYAML_StreamValid_NDJSONHoldOpen(t *testing.T) {
	yaml := `
scenarios:
  - uuid: "st-ndjson"
    method: "GET"
    path: "/feed"
    status_code: 200
    stream:
      format: "ndjson"
      interval_ms: 250
      hold_open: true
      template: "{\"i\":{{index}},\"t\":\"{{timestamp}}\"}"
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
	st := cfg.Scenarios[0].Stream
	if st == nil {
		t.Fatal("stream missing")
	}
	if st.Format != "ndjson" {
		t.Errorf("Format: got %q", st.Format)
	}
	if !st.HoldOpen {
		t.Errorf("HoldOpen: got %v", st.HoldOpen)
	}
	if st.IntervalMS != 250 {
		t.Errorf("IntervalMS: got %d", st.IntervalMS)
	}
}

func TestUniConfig_LoadFromYAML_StreamValidationErrors(t *testing.T) {
	type tc struct {
		name, stream, errSubstr string
	}
	tests := []tc{
		{name: "format grpc", stream: streamFormatGRPC, errSubstr: "format"},
		{name: "both event_count and hold_open", stream: streamBothSet, errSubstr: "exactly one"},
		{name: "neither event_count nor hold_open", stream: streamNeitherSet, errSubstr: "exactly one"},
		{name: "negative interval_ms", stream: streamNegativeInterval, errSubstr: "interval_ms"},
		{name: "empty template", stream: streamEmptyTemplate, errSubstr: "template"},
		{name: "stream and data both set", stream: streamPlusData, errSubstr: "mutually exclusive"},
		{name: "unknown stream key websocket", stream: streamUnknownKey, errSubstr: "unknown field"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runStreamLoadError(t, tt.stream, tt.errSubstr)
		})
	}
}

// Per-test YAML fragments. Kept as package-level vars to keep the test table
// short. Each fragment carries a valid `format:` and `template:` so that the
// rejection being exercised is the target one, not an earlier failure.
const (
	streamFormatGRPC = "" +
		"\n      format: \"grpc\"" +
		"\n      interval_ms: 0" +
		"\n      event_count: 1" +
		"\n      template: \"data: hi\\n\\n\""
	streamBothSet = "" +
		"\n      format: \"sse\"" +
		"\n      interval_ms: 0" +
		"\n      event_count: 3" +
		"\n      hold_open: true" +
		"\n      template: \"data: hi\\n\\n\""
	streamNeitherSet = "" +
		"\n      format: \"sse\"" +
		"\n      interval_ms: 0" +
		"\n      template: \"data: hi\\n\\n\""
	streamNegativeInterval = "" +
		"\n      format: \"sse\"" +
		"\n      interval_ms: -10" +
		"\n      event_count: 3" +
		"\n      template: \"data: hi\\n\\n\""
	streamEmptyTemplate = "" +
		"\n      format: \"sse\"" +
		"\n      event_count: 3" +
		"\n      template: \"  \""
	streamPlusData = "" +
		"\n      format: \"sse\"" +
		"\n      event_count: 3" +
		"\n      template: \"data: hi\\n\\n\""
	streamUnknownKey = "" +
		"\n      format: \"sse\"" +
		"\n      event_count: 3" +
		"\n      template: \"data: hi\\n\\n\"" +
		"\n      websocket: true"
)

func runStreamLoadError(t *testing.T, stream, errSubstr string) {
	t.Helper()
	yaml := `
scenarios:
  - method: GET
    path: "/events"
    status_code: 200
    data: "{\"ok\":true}"
    stream:` + stream
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
