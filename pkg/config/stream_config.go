package config

import (
	"fmt"

	"github.com/bmcszk/unimock/pkg/model"
)

// StreamConfig mirrors model.StreamConfig for YAML deserialization and validation.
type StreamConfig struct {
	// Format is the streaming format: "sse" (text/event-stream) or "ndjson"
	// (application/x-ndjson).
	Format string `yaml:"format" json:"format"`

	// IntervalMS is the delay between frames in milliseconds. Must be >= 0.
	IntervalMS int `yaml:"interval_ms,omitempty" json:"interval_ms,omitempty"`

	// EventCount is the number of frames to emit before closing the stream.
	// Exactly one of EventCount > 0 or HoldOpen == true must be set.
	EventCount int `yaml:"event_count,omitempty" json:"event_count,omitempty"`

	// HoldOpen, when true, keeps the stream open until the client disconnects.
	HoldOpen bool `yaml:"hold_open,omitempty" json:"hold_open,omitempty"`

	// Template is the per-frame payload template with {{index}} and {{timestamp}}
	// placeholders. Must be non-empty.
	Template string `yaml:"template" json:"template"`
}

// streamYAMLKeys is the closed allowlist of YAML keys under `stream:`. Any key
// not listed is rejected by the strict YAML decoder to prevent typos and
// unintended inline fields from sneaking in.
var streamYAMLKeys = map[string]struct{}{
	"format":      {},
	"interval_ms": {},
	"event_count": {},
	"hold_open":   {},
	"template":    {},
}

// validate enforces the stream configuration contract. It returns the first error
// found. The caller should already have rejected unknown YAML keys via the
// strict decoder. The error strings MUST remain stable (prefix "stream: " +
// body from model) because pkg/config tests assert them by substring.
//
// Delegates to model.StreamConfig.Validate so the model owns the rule; we only
// add the "stream: " prefix the YAML error messages must include.
func (s *StreamConfig) validate() error {
	if s == nil {
		return nil
	}
	err := s.toModel().Validate()
	if err == nil {
		return nil
	}
	return fmt.Errorf("stream: %w", err)
}

// toModel converts the config stream into the runtime model representation.
func (s *StreamConfig) toModel() *model.StreamConfig {
	return &model.StreamConfig{
		Format:     s.Format,
		IntervalMS: s.IntervalMS,
		EventCount: s.EventCount,
		HoldOpen:   s.HoldOpen,
		Template:   s.Template,
	}
}