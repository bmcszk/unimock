package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
)

// newTestStreamWriter builds a StreamWriter via the exported constructor.
// The default wait seam is ctx-aware and no-op at interval_ms<=0, so
// event_count tests emit frames instantly.
func newTestStreamWriter() *handler.StreamWriter {
	return handler.NewStreamWriter(nil)
}

func TestStreamWriter_SSEFrameBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 0,
		EventCount: 2,
		Template:   `{"n":{{index}}}`,
	}
	w.WriteStream(context.Background(), rec, sc)

	body := rec.Body.String()
	assert.Equal(t, "data: {\"n\":0}\n\ndata: {\"n\":1}\n\n", body)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "no", rec.Header().Get("X-Accel-Buffering"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStreamWriter_NDJSONFrameBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "ndjson",
		IntervalMS: 0,
		EventCount: 2,
		Template:   `{"n":{{index}}}`,
	}
	w.WriteStream(context.Background(), rec, sc)

	assert.Equal(t, "{\"n\":0}\n{\"n\":1}\n", rec.Body.String())
	assert.Equal(t, "application/x-ndjson", rec.Header().Get("Content-Type"))
}

func TestStreamWriter_TimestampPlaceholder(t *testing.T) {
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 0,
		EventCount: 1,
		Template:   `{"t":"{{timestamp}}"}`,
	}
	w.WriteStream(context.Background(), rec, sc)

	// {{timestamp}} renders RFC3339 of the wall clock (the exact instant is
	// not injectable through the exported API) — assert the RFC3339 shape.
	assert.Contains(t, rec.Body.String(), `{"t":"`)
	assert.Contains(t, rec.Body.String(), time.Now().Format("2006-01-02"))
}

func TestStreamWriter_HoldOpenEndsOnContextCancel(t *testing.T) {
	rec := httptest.NewRecorder()
	// Exported constructor: default wait seam is already ctx-aware, so the
	// loop returns promptly when the context is canceled.
	w := handler.NewStreamWriter(nil)

	ctx, cancel := context.WithCancel(context.Background())
	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 1,
		HoldOpen:   true,
		Template:   "x",
	}
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	done := make(chan struct{})
	go func() {
		w.WriteStream(ctx, rec, sc)
		close(done)
	}()

	select {
	case <-done:
		// loop returned promptly after cancel — no goroutine leak
	case <-time.After(2 * time.Second):
		t.Fatal("hold_open stream did not end after context cancel")
	}
}

func TestStreamWriter_UnsupportedFormat500(t *testing.T) {
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "websocket",
		IntervalMS: 0,
		EventCount: 1,
		Template:   "x",
	}
	w.WriteStream(context.Background(), rec, sc)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "data: ", "no SSE frames written on unsupported format")
}

func TestStreamWriter_NonFlusherWriter500(t *testing.T) {
	// httptest.NewRecorder implements Flusher, so use a plain wrapper that
	// drops Flush to exercise the non-Flusher error path.
	rec := httptest.NewRecorder()
	plain := &plainResponseWriter{ResponseWriter: rec}
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 0,
		EventCount: 1,
		Template:   "x",
	}
	w.WriteStream(context.Background(), plain, sc)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// plainResponseWriter hides the Flusher implementation of the underlying
// recorder so WriteStream's type assertion fails.
type plainResponseWriter struct {
	http.ResponseWriter
}

func TestStreamWriter_EventCountTermination(t *testing.T) {
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "ndjson",
		IntervalMS: 0,
		EventCount: 5,
		Template:   "f",
	}
	w.WriteStream(context.Background(), rec, sc)

	assert.Equal(t, "f\nf\nf\nf\nf\n", rec.Body.String(), "exactly EventCount frames")
}
