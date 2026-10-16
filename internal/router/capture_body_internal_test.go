package router

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCaptureRequestBody_TruncatesAtCap(t *testing.T) {
	// given
	body := bytes.Repeat([]byte("a"), maxDispatchBodyCapture+1024)
	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(body))

	// when
	got := captureRequestBody(req)

	// then
	if len(got) != maxDispatchBodyCapture {
		t.Fatalf("captured %d bytes, want cap %d", len(got), maxDispatchBodyCapture)
	}
	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read restored body: %v", err)
	}
	if len(restored) != maxDispatchBodyCapture {
		t.Fatalf("restored %d bytes, want cap %d", len(restored), maxDispatchBodyCapture)
	}
}
