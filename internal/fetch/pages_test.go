package fetch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

// recorder collects emitted events (thread-safe).
type recorder struct {
	mu     sync.Mutex
	events []Event
}

func (r *recorder) emit(ev Event) {
	r.mu.Lock()
	r.events = append(r.events, ev)
	r.mu.Unlock()
}

func (r *recorder) kinds() []Kind {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Kind, len(r.events))
	for i, e := range r.events {
		out[i] = e.Kind
	}
	return out
}

func (r *recorder) count(k Kind) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, e := range r.events {
		if e.Kind == k {
			n++
		}
	}
	return n
}

const perPage = 2

// pageData builds `count` items on the given page as "p<page>-<i>" strings.
func pageData(page, count int) []string {
	out := make([]string, count)
	for i := 0; i < count; i++ {
		out[i] = fmt.Sprintf("p%d-%d", page, i)
	}
	return out
}

func strKey(s string) string { return s }

func TestPagesDeterminateFanOut(t *testing.T) {
	// 3 pages known via total on page 1.
	rec := &recorder{}
	fetchPage := func(_ context.Context, page int) (Page[string], error) {
		p := Page[string]{Items: pageData(page, perPage)}
		if page == 1 {
			p.TotalPages = 3
		}
		return p, nil
	}
	items, err := Pages(context.Background(), rec.emit, "prov", "owner", perPage, fetchPage, strKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 6 {
		t.Fatalf("got %d items, want 6: %v", len(items), items)
	}
	if rec.count(KindPageDone) != 3 {
		t.Fatalf("PageDone count = %d, want 3", rec.count(KindPageDone))
	}
	if rec.count(KindStarted) != 1 || rec.count(KindDone) != 1 {
		t.Fatalf("expected 1 started + 1 done, kinds=%v", rec.kinds())
	}
	// First event Started, last event Done.
	ks := rec.kinds()
	if ks[0] != KindStarted || ks[len(ks)-1] != KindDone {
		t.Fatalf("ordering wrong: %v", ks)
	}
}

func TestPagesMergeOrderAndDedupe(t *testing.T) {
	rec := &recorder{}
	fetchPage := func(_ context.Context, page int) (Page[string], error) {
		p := Page[string]{}
		if page == 1 {
			p.TotalPages = 2
			p.Items = []string{"a", "b"}
		} else {
			p.Items = []string{"b", "c"} // "b" duplicates page 1
		}
		return p, nil
	}
	items, err := Pages(context.Background(), rec.emit, "p", "o", perPage, fetchPage, strKey)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a", "b", "c"}
	if len(items) != 3 || items[0] != want[0] || items[1] != want[1] || items[2] != want[2] {
		t.Fatalf("merge/dedupe wrong: %v, want %v", items, want)
	}
}

func TestPagesSequentialFallback(t *testing.T) {
	// No total on page 1; full pages until a short one.
	rec := &recorder{}
	fetchPage := func(_ context.Context, page int) (Page[string], error) {
		switch page {
		case 1, 2:
			return Page[string]{Items: pageData(page, perPage)}, nil // full page
		case 3:
			return Page[string]{Items: pageData(page, 1)}, nil // short page -> last
		default:
			t.Errorf("unexpected page %d", page)
			return Page[string]{}, nil
		}
	}
	items, err := Pages(context.Background(), rec.emit, "p", "o", perPage, fetchPage, strKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 5 {
		t.Fatalf("got %d items, want 5", len(items))
	}
	if rec.count(KindPageDone) != 3 {
		t.Fatalf("PageDone = %d, want 3", rec.count(KindPageDone))
	}
	// Total unknown: PageDone events carry TotalPages 0.
	rec.mu.Lock()
	for _, e := range rec.events {
		if e.Kind == KindPageDone && e.TotalPages != 0 {
			t.Fatalf("expected indeterminate total, got %d", e.TotalPages)
		}
	}
	rec.mu.Unlock()
}

func TestPagesSingleShortFirstPage(t *testing.T) {
	rec := &recorder{}
	calls := 0
	fetchPage := func(_ context.Context, page int) (Page[string], error) {
		calls++
		return Page[string]{Items: pageData(1, 1)}, nil // short, no total
	}
	items, err := Pages(context.Background(), rec.emit, "p", "o", perPage, fetchPage, strKey)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(items) != 1 {
		t.Fatalf("short first page should stop: calls=%d items=%d", calls, len(items))
	}
}

func TestPagesFirstPageError(t *testing.T) {
	rec := &recorder{}
	wantErr := errors.New("boom")
	fetchPage := func(_ context.Context, page int) (Page[string], error) {
		return Page[string]{}, wantErr
	}
	_, err := Pages(context.Background(), rec.emit, "p", "o", perPage, fetchPage, strKey)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want boom", err)
	}
	if rec.count(KindFailed) != 1 {
		t.Fatalf("expected 1 Failed event, kinds=%v", rec.kinds())
	}
}

func TestPagesLaterPageError(t *testing.T) {
	rec := &recorder{}
	wantErr := errors.New("page2 boom")
	fetchPage := func(_ context.Context, page int) (Page[string], error) {
		if page == 1 {
			return Page[string]{Items: []string{"a", "b"}, TotalPages: 3}, nil
		}
		if page == 2 {
			return Page[string]{}, wantErr
		}
		return Page[string]{Items: []string{"x"}}, nil
	}
	_, err := Pages(context.Background(), rec.emit, "p", "o", perPage, fetchPage, strKey)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want page2 boom", err)
	}
	if rec.count(KindFailed) != 1 {
		t.Fatalf("expected Failed, kinds=%v", rec.kinds())
	}
}

func TestPagesCancelDuringFirstPage(t *testing.T) {
	rec := &recorder{}
	ctx, cancel := context.WithCancel(context.Background())
	fetchPage := func(c context.Context, page int) (Page[string], error) {
		cancel()
		return Page[string]{}, c.Err()
	}
	_, err := Pages(ctx, rec.emit, "p", "o", perPage, fetchPage, strKey)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if rec.count(KindCanceled) != 1 {
		t.Fatalf("expected Canceled event, kinds=%v", rec.kinds())
	}
}

func TestPagesCancelDuringFanOut(t *testing.T) {
	rec := &recorder{}
	ctx, cancel := context.WithCancel(context.Background())
	fetchPage := func(c context.Context, page int) (Page[string], error) {
		if page == 1 {
			return Page[string]{Items: []string{"a", "b"}, TotalPages: 5}, nil
		}
		// Cancel before returning subsequent pages.
		cancel()
		return Page[string]{}, c.Err()
	}
	_, err := Pages(ctx, rec.emit, "p", "o", perPage, fetchPage, strKey)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if rec.count(KindCanceled) != 1 {
		t.Fatalf("expected Canceled event, kinds=%v", rec.kinds())
	}
	if rec.count(KindDone) != 0 {
		t.Fatal("Done must not be emitted on cancel")
	}
}

func TestNilEmitIsSafe(t *testing.T) {
	fetchPage := func(_ context.Context, page int) (Page[string], error) {
		return Page[string]{Items: []string{"a"}}, nil
	}
	items, err := Pages(context.Background(), nil, "p", "o", perPage, fetchPage, strKey)
	if err != nil || len(items) != 1 {
		t.Fatalf("nil emit: items=%v err=%v", items, err)
	}
}

func TestKindString(t *testing.T) {
	for _, k := range []Kind{KindStarted, KindPageDone, KindDone, KindFailed, KindCanceled, KindWarning} {
		if k.String() == "" {
			t.Fatalf("kind %d has empty string", k)
		}
	}
	if Kind(99).String() == "" {
		t.Fatal("unknown kind should still stringify")
	}
}

func TestEmitHelpers(t *testing.T) {
	rec := &recorder{}
	e := Emit(rec.emit)
	e.Started("p", "o")
	e.PageDone("p", "o", 1, 2, 3)
	e.Done("p", "o", 3)
	e.Failed("p", "o", errors.New("x"))
	e.Canceled("p", "o")
	e.Warning("p", "o", "hi")
	if len(rec.events) != 6 {
		t.Fatalf("got %d events, want 6", len(rec.events))
	}
	if rec.events[1].Page != 1 || rec.events[1].TotalPages != 2 || rec.events[1].ReposSoFar != 3 {
		t.Fatalf("PageDone fields wrong: %+v", rec.events[1])
	}
	if rec.events[5].Msg != "hi" {
		t.Fatalf("Warning msg wrong: %+v", rec.events[5])
	}
}
