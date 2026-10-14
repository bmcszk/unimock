package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bmcszk/unimock/pkg/model"
)

const (
	streamFormatSSE    = "sse"
	streamFormatNDJSON = "ndjson"
)

// streamContentType maps a stream Format to its canonical Content-Type.
var streamContentType = map[string]string{
	streamFormatSSE:    "text/event-stream",
	streamFormatNDJSON: "application/x-ndjson",
}

const (
	placeholderIndex     = "{{index}}"
	placeholderTimestamp = "{{timestamp}}"
	timestampLayout      = time.RFC3339Nano
)

// StreamWriter writes server-generated stream responses (SSE or NDJSON) to an
// http.ResponseWriter. The writer is safe for concurrent use only if the caller
// ensures a single goroutine writes at a time — by construction each handler
// invocation owns its own WriteStream call.
type StreamWriter struct {
	logger *slog.Logger
	now    func() time.Time
	wait   func(context.Context, time.Duration)
}

// NewStreamWriter creates a StreamWriter using real-time clock and a
// context-aware wait seam for the per-frame sleep.
func NewStreamWriter(logger *slog.Logger) *StreamWriter {
	return &StreamWriter{
		logger: logger,
		now:    time.Now,
		wait:   waitCtxTimeout,
	}
}

// waitCtxTimeout blocks for d or until ctx is canceled, whichever comes first.
// It is the default `wait` seam for StreamWriter. A zero or negative duration
// returns without blocking (preserves the no-sleep contract for interval_ms<=0).
func waitCtxTimeout(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

// WriteStream renders sc per frame into w, flushing after every frame. It
// honours r.Context().Done() for hold-open mode and aborts immediately if the
// client disconnects in event-count mode.
//
// Contract:
//   - Sets Content-Type, Cache-Control: no-cache, X-Accel-Buffering: no.
//   - Calls WriteHeader(200) BEFORE the first frame is written.
//   - Flushes after every frame (requires http.Flusher).
//   - Frame terminators: SSE — "data: <frame>\n\n"; NDJSON — "<frame>\n".
//   - EventCount mode: ends after N frames.
//   - HoldOpen mode: ends when ctx is canceled.
func (w *StreamWriter) WriteStream(ctx context.Context, rw http.ResponseWriter, sc *model.StreamConfig) {
	if sc == nil {
		return
	}
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "streaming unsupported by response writer", http.StatusInternalServerError)
		return
	}
	contentType := streamContentType[sc.Format]
	if contentType == "" {
		http.Error(rw, fmt.Sprintf("unsupported stream format %q", sc.Format),
			http.StatusInternalServerError)
		return
	}
	w.writeHeaders(rw, contentType)
	rw.WriteHeader(http.StatusOK)
	w.streamLoop(ctx, rw, flusher, sc)
}

// writeHeaders sets the v1 contract headers.
func (*StreamWriter) writeHeaders(rw http.ResponseWriter, contentType string) {
	rw.Header().Set("Content-Type", contentType)
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("X-Accel-Buffering", "no")
}

// streamLoop is the per-frame loop body. Sleeping is the caller-configured wait
// seam (tests can replace it); the default wait aborts promptly when ctx is
// canceled (MINOR7).
func (w *StreamWriter) streamLoop(
	ctx context.Context, rw http.ResponseWriter, flusher http.Flusher, sc *model.StreamConfig,
) {
	interval := time.Duration(sc.IntervalMS) * time.Millisecond
	for i := 0; shouldContinue(ctx, sc, i); i++ {
		frame := renderTemplate(sc.Template, i, w.now())
		if err := w.writeFrame(rw, sc.Format, frame); err != nil {
			return
		}
		flusher.Flush()
		w.wait(ctx, interval)
	}
}

// shouldContinue reports whether the stream should emit another frame.
// HoldOpen mode only stops on context cancel; otherwise EventCount bounds the
// loop to frames 0..EventCount-1 inclusive.
func shouldContinue(ctx context.Context, sc *model.StreamConfig, currentIndex int) bool {
	if err := ctx.Err(); err != nil {
		return false
	}
	if sc.HoldOpen {
		return true
	}
	return currentIndex < sc.EventCount
}

// writeFrame encodes frame per the format's wire protocol.
func (*StreamWriter) writeFrame(rw io.Writer, format, frame string) error {
	switch format {
	case streamFormatSSE:
		_, err := rw.Write([]byte("data: " + frame + "\n\n"))
		return err
	case streamFormatNDJSON:
		_, err := rw.Write([]byte(frame + "\n"))
		return err
	default:
		return errors.New("unsupported format")
	}
}

// renderTemplate substitutes the documented placeholders. Unknown placeholders
// are left intact so user-authored templates can carry braces for JSON or
// other reasons.
func renderTemplate(template string, index int, ts time.Time) string {
	out := template
	if strings.Contains(out, placeholderIndex) {
		out = strings.ReplaceAll(out, placeholderIndex, fmt.Sprintf("%d", index))
	}
	if strings.Contains(out, placeholderTimestamp) {
		out = strings.ReplaceAll(out, placeholderTimestamp, ts.UTC().Format(timestampLayout))
	}
	return out
}
