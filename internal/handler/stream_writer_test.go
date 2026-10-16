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

// newTestStreamWriter: the default wait seam no-ops at interval_ms<=0, so event_count tests emit instantly.
func newTestStreamWriter() *handler.StreamWriter {
	return handler.NewStreamWriter(nil)
}

func TestStreamWriter_SSEFrameBytes(t *testing.T) {
	// given
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 0,
		EventCount: 2,
		Template:   `{"n":{{index}}}`,
	}

	// when
	w.WriteStream(context.Background(), rec, sc, 0)

	// then
	body := rec.Body.String()
	assert.Equal(t, "data: {\"n\":0}\n\ndata: {\"n\":1}\n\n", body)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "no", rec.Header().Get("X-Accel-Buffering"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStreamWriter_NDJSONFrameBytes(t *testing.T) {
	// given
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "ndjson",
		IntervalMS: 0,
		EventCount: 2,
		Template:   `{"n":{{index}}}`,
	}

	// when
	w.WriteStream(context.Background(), rec, sc, 0)

	// then
	assert.Equal(t, "{\"n\":0}\n{\"n\":1}\n", rec.Body.String())
	assert.Equal(t, "application/x-ndjson", rec.Header().Get("Content-Type"))
}

func TestStreamWriter_TimestampPlaceholder(t *testing.T) {
	// given
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 0,
		EventCount: 1,
		Template:   `{"t":"{{timestamp}}"}`,
	}

	// when
	w.WriteStream(context.Background(), rec, sc, 0)

	// then
	// {{timestamp}} renders RFC3339 of the wall clock (the exact instant is
	// not injectable through the exported API) — assert the RFC3339 shape.
	assert.Contains(t, rec.Body.String(), `{"t":"`)
	assert.Contains(t, rec.Body.String(), time.Now().Format("2006-01-02"))
}

func TestStreamWriter_HoldOpenEndsOnContextCancel(t *testing.T) {
	// given
	rec := httptest.NewRecorder()
	w := handler.NewStreamWriter(nil)

	ctx, cancel := context.WithCancel(context.Background())
	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 1,
		HoldOpen:   true,
		Template:   "x",
	}

	// when
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	done := make(chan struct{})
	go func() {
		w.WriteStream(ctx, rec, sc, 0)
		close(done)
	}()

	// then
	select {
	case <-done:
		// loop returned promptly after cancel — no goroutine leak
	case <-time.After(2 * time.Second):
		t.Fatal("hold_open stream did not end after context cancel")
	}
}

func TestStreamWriter_SetWriteDeadlineRefreshedPerFrame(t *testing.T) {
	// given
	base := httptest.NewRecorder()
	rec := &deadlineRecordingResponseWriter{ResponseWriter: base}
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 0,
		EventCount: 3,
		Template:   `{"n":{{index}}}`,
	}

	// when
	w.WriteStream(context.Background(), rec, sc, 0)

	// then
	assert.Equal(t, "data: {\"n\":0}\n\ndata: {\"n\":1}\n\ndata: {\"n\":2}\n\n",
		base.Body.String(), "stream frames unchanged by deadline bookkeeping")
	assert.GreaterOrEqual(t, rec.setWriteDeadlineCalls, 3,
		"every frame must be preceded by a SetWriteDeadline refresh")
	for i, call := range rec.deadlines {
		assert.True(t, call.value.After(call.atCallTime),
			"deadline call %d (%v) must land in the future at call time", i, call.value)
	}
}

func TestStreamWriter_HoldOpenAlsoRefreshesWriteDeadline(t *testing.T) {
	// given
	base := httptest.NewRecorder()
	rec := &deadlineRecordingResponseWriter{ResponseWriter: base}
	w := handler.NewStreamWriter(nil)

	ctx, cancel := context.WithCancel(context.Background())
	sc := &model.StreamConfig{
		Format:     "sse",
		IntervalMS: 1,
		HoldOpen:   true,
		Template:   "x",
	}

	// when
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	done := make(chan struct{})
	go func() {
		w.WriteStream(ctx, rec, sc, 0)
		close(done)
	}()

	// then
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("hold_open stream did not end after context cancel")
	}
	assert.GreaterOrEqual(t, rec.setWriteDeadlineCalls, 1,
		"hold_open frames must refresh the write deadline too")
}

// deadlineRecord pairs a SetWriteDeadline argument with the wall clock at the
// moment of the call, so tests can assert the deadline lies in the future.
type deadlineRecord struct {
	value      time.Time
	atCallTime time.Time
}

// deadlineRecordingResponseWriter captures SetWriteDeadline calls exactly the
// way net/http's internal connection writer receives them.
type deadlineRecordingResponseWriter struct {
	http.ResponseWriter
	setWriteDeadlineCalls int
	deadlines             []deadlineRecord
}

func (w *deadlineRecordingResponseWriter) SetWriteDeadline(deadline time.Time) error {
	w.setWriteDeadlineCalls++
	w.deadlines = append(w.deadlines, deadlineRecord{value: deadline, atCallTime: time.Now()})
	return nil
}

func (w *deadlineRecordingResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func TestStreamWriter_UnsupportedFormat500(t *testing.T) {
	// given
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "websocket",
		IntervalMS: 0,
		EventCount: 1,
		Template:   "x",
	}

	// when
	w.WriteStream(context.Background(), rec, sc, 0)

	// then
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "data: ", "no SSE frames written on unsupported format")
}

func TestStreamWriter_NonFlusherWriter500(t *testing.T) {
	// given
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

	// when
	w.WriteStream(context.Background(), plain, sc, 0)

	// then
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// plainResponseWriter hides the Flusher implementation of the underlying
// recorder so WriteStream's type assertion fails.
type plainResponseWriter struct {
	http.ResponseWriter
}

func TestStreamWriter_EventCountTermination(t *testing.T) {
	// given
	rec := httptest.NewRecorder()
	w := newTestStreamWriter()

	sc := &model.StreamConfig{
		Format:     "ndjson",
		IntervalMS: 0,
		EventCount: 5,
		Template:   "f",
	}

	// when
	w.WriteStream(context.Background(), rec, sc, 0)

	// then
	assert.Equal(t, "f\nf\nf\nf\nf\n", rec.Body.String(), "exactly EventCount frames")
}

func TestWriteStream_HonorsScenarioStatusCode(t *testing.T) {
	// given
	w := handler.NewStreamWriter(nil)
	rec := httptest.NewRecorder()

	// when
	w.WriteStream(context.Background(), rec, &model.StreamConfig{
		Format:     "sse",
		EventCount: 1,
		IntervalMS: 0,
		Template:   `{"i":{{index}}}`,
	}, http.StatusTeapot)

	// then
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Contains(t, rec.Body.String(), "data: {\"i\":0}")
}

func TestWriteStream_ZeroStatusCodeDefaultsOK(t *testing.T) {
	// given
	w := handler.NewStreamWriter(nil)
	rec := httptest.NewRecorder()

	// when
	w.WriteStream(context.Background(), rec, &model.StreamConfig{
		Format:     "ndjson",
		EventCount: 1,
		IntervalMS: 0,
		Template:   `{"i":{{index}}}`,
	}, 0)

	// then
	assert.Equal(t, http.StatusOK, rec.Code)
}
