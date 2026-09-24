package webhooks_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/webhooks"
)

func TestRing_AddAndSnapshot(t *testing.T) {
	r := webhooks.NewRing(3)

	r.Add(webhooks.Delivery{ID: "a", URL: "u1", Attempt: 1, StatusCode: 200, TS: time.Now()})
	r.Add(webhooks.Delivery{ID: "b", URL: "u2", Attempt: 1, StatusCode: 500, Error: "boom", TS: time.Now()})

	got := r.Snapshot()
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if got[0].ID != "a" || got[1].ID != "b" {
		t.Errorf("order: got [%s, %s], want [a, b]", got[0].ID, got[1].ID)
	}
}

func TestRing_OverflowEvictsOldest(t *testing.T) {
	r := webhooks.NewRing(3)
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		r.Add(webhooks.Delivery{ID: id, URL: "u", Attempt: 1, StatusCode: 200, TS: time.Now()})
	}
	got := r.Snapshot()
	if len(got) != 3 {
		t.Fatalf("expected cap 3 records, got %d", len(got))
	}
	want := []string{"c", "d", "e"}
	for i, w := range want {
		if got[i].ID != w {
			t.Errorf("position %d: got %q, want %q", i, got[i].ID, w)
		}
	}
}

func TestRing_EmptySnapshot(t *testing.T) {
	r := webhooks.NewRing(5)
	got := r.Snapshot()
	if got == nil {
		t.Fatal("snapshot of empty ring must not be nil (handler expects empty array)")
	}
	if len(got) != 0 {
		t.Errorf("expected empty snapshot, got %d items", len(got))
	}
}

func TestRing_ConcurrentSafe(t *testing.T) {
	r := webhooks.NewRing(100)
	const writers = 8
	const perWriter = 50

	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		w := w
		wg.Go(func() {
			for i := 0; i < perWriter; i++ {
				r.Add(webhooks.Delivery{
					ID:         fmt.Sprintf("w%d-i%d", w, i),
					URL:        "u",
					Attempt:    1,
					StatusCode: 200,
					TS:         time.Now(),
				})
			}
		})
	}
	wg.Wait()

	got := r.Snapshot()
	if len(got) != 100 {
		t.Errorf("expected 100 records after writes, got %d", len(got))
	}
}

func TestRing_CapacityZeroDefaultsTo100(t *testing.T) {
	r := webhooks.NewRing(0)
	for i := 0; i < 200; i++ {
		r.Add(webhooks.Delivery{ID: fmt.Sprintf("id%d", i), TS: time.Now()})
	}
	got := r.Snapshot()
	if len(got) != 100 {
		t.Errorf("expected default cap 100, got %d", len(got))
	}
}

func TestRing_SnapshotIsCopy(t *testing.T) {
	r := webhooks.NewRing(3)
	r.Add(webhooks.Delivery{ID: "a", TS: time.Now()})
	snap := r.Snapshot()
	snap[0].ID = "mutated"

	again := r.Snapshot()
	if again[0].ID != "a" {
		t.Errorf("snapshot must be a defensive copy; got ID %q", again[0].ID)
	}
}
