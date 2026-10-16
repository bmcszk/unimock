package handler_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/webhooks"
	"github.com/bmcszk/unimock/pkg/model"
)

func newTechHandlerWithDeliveries(t *testing.T) (*handler.TechHandler, *webhooks.Ring) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	techService := service.NewTechService(time.Now())
	ring := webhooks.NewRing(10)
	h := handler.NewTechHandler(techService, logger, ring)
	return h, ring
}

func TestTechHandler_WebhookDeliveries_EmptyReturnsArray(t *testing.T) {
	h, _ := newTechHandlerWithDeliveries(t)

	req, err := http.NewRequest("GET", "/_uni/webhooks/deliveries", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type: got %q want application/json", got)
	}

	// Must decode as a JSON array, not null.
	var got []webhooks.Delivery
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not a JSON array: %v; raw=%q", err, rr.Body.String())
	}
	if got == nil {
		t.Fatal("expected empty array, got nil")
	}
	if len(got) != 0 {
		t.Errorf("expected empty array, got %d items", len(got))
	}
}

func TestTechHandler_WebhookDeliveries_ReturnsRecords(t *testing.T) {
	h, ring := newTechHandlerWithDeliveries(t)

	ring.Add(webhooks.Delivery{
		ID:         "wh-1",
		URL:        "http://example.com/hook",
		Attempt:    1,
		StatusCode: 200,
		TS:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	ring.Add(webhooks.Delivery{
		ID:      "wh-2",
		URL:     "http://example.com/hook2",
		Attempt: 2,
		Error:   "boom",
		TS:      time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC),
	})

	req, err := http.NewRequest("GET", "/_uni/webhooks/deliveries", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}
	var got []webhooks.Delivery
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if got[0].ID != "wh-1" || got[1].ID != "wh-2" {
		t.Errorf("order/IDs: %+v", got)
	}
	if got[1].Error != "boom" {
		t.Errorf("error not propagated: %+v", got[1])
	}
}

func TestTechHandler_WebhookDeliveries_MethodNotAllowed(t *testing.T) {
	h, _ := newTechHandlerWithDeliveries(t)

	req, err := http.NewRequest("POST", "/_uni/webhooks/deliveries", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d want 405", rr.Code)
	}
}

func TestTechHandler_WebhookDeliveries_PathRoutingPreserved(t *testing.T) {
	// Sanity check that adding the new route did not break existing endpoints.
	h, _ := newTechHandlerWithDeliveries(t)

	tests := []struct {
		path string
		want int
	}{
		{"/_uni/health", http.StatusOK},
		{"/_uni/metrics", http.StatusOK},
		{"/_uni/webhooks/deliveries", http.StatusOK},
		{"/_uni/something-else", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tt.want {
				t.Errorf("path %s: status %d want %d", tt.path, rr.Code, tt.want)
			}
		})
	}
}

// silence unused import warnings
var _ = model.Scenario{}
var _ = os.Stdout
