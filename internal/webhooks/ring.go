// Package webhooks implements the scenario-triggered outbound webhook dispatcher
// plus an in-memory ring buffer of delivery records exposed via
// GET /_uni/webhooks/deliveries.
package webhooks

import (
	"sync"
	"time"
)

// defaultRingCapacity is used when a non-positive capacity is supplied to NewRing.
const defaultRingCapacity = 100

// Delivery captures a single webhook attempt outcome. A non-zero StatusCode reflects
// an HTTP response (with any Error describing the response body or transport issue);
// a zero StatusCode indicates a transport-level failure with Error populated.
type Delivery struct {
	// ID is the Standard Webhooks message id sent in this attempt (UUID).
	ID string `json:"id"`
	// URL is the webhook target that was called.
	URL string `json:"url"`
	// Attempt is the 1-based attempt index.
	Attempt int `json:"attempt"`
	// StatusCode is the HTTP status received, 0 when no response was received.
	StatusCode int `json:"statusCode"`
	// Error is the transport-level or response-decoding error, if any.
	Error string `json:"error,omitempty"`
	// TS is when the attempt completed.
	TS time.Time `json:"ts"`
}

// Ring is a thread-safe bounded buffer of Delivery records. When full, the oldest
// record is evicted to make room for a new one. The zero value is NOT usable;
// callers must call NewRing.
type Ring struct {
	mu       sync.Mutex
	cap      int
	items    []Delivery
	next     int // next write index when full
	full     bool
}

// NewRing creates a Ring with the given capacity. A non-positive capacity falls back
// to defaultRingCapacity.
func NewRing(capacity int) *Ring {
	if capacity <= 0 {
		capacity = defaultRingCapacity
	}
	return &Ring{
		cap:   capacity,
		items: make([]Delivery, 0, capacity),
	}
}

// Add records a delivery. Safe for concurrent use.
func (r *Ring) Add(d Delivery) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.full {
		r.items[r.next] = d
		r.next = (r.next + 1) % r.cap
		return
	}
	if len(r.items) < r.cap {
		r.items = append(r.items, d)
		if len(r.items) == r.cap {
			r.full = true
		}
		return
	}
	// Defensive: should be unreachable since full==false and len==cap is the same
	// condition, but keep a consistent fallback.
	r.items[0] = d
	r.next = 1
}

// Snapshot returns a defensive copy of the buffer in insertion order (oldest first).
// The returned slice is safe for the caller to mutate without affecting the ring.
func (r *Ring) Snapshot() []Delivery {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) == 0 {
		return []Delivery{}
	}
	if !r.full {
		out := make([]Delivery, len(r.items))
		copy(out, r.items)
		return out
	}
	out := make([]Delivery, r.cap)
	copy(out, r.items[r.next:])
	copy(out[r.cap-r.next:], r.items[:r.next])
	return out
}

// Capacity returns the ring capacity.
func (r *Ring) Capacity() int {
	return r.cap
}
