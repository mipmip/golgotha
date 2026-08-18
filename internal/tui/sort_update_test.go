package tui

import (
	"testing"

	"github.com/mipmip/huphop/internal/config"
)

func TestSortKeyPressCycles(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newSortModel(t, p)

	if m.sortKey != sortNone {
		t.Fatalf("expected sortNone initially, got %v", m.sortKey)
	}

	// s: none -> name, list becomes alphabetical.
	send(m, key("s"))
	if m.sortKey != sortName {
		t.Fatalf("after 1st s expected name, got %v", m.sortKey)
	}
	if got := names(m.visibleRepos()); !equal(got, []string{"alpha", "Bravo", "Charlie", "delta"}) {
		t.Fatalf("name asc after s: got %v", got)
	}

	// s: name -> updated.
	send(m, key("s"))
	if m.sortKey != sortUpdated {
		t.Fatalf("after 2nd s expected updated, got %v", m.sortKey)
	}

	// s: updated -> none (fetch order restored).
	send(m, key("s"))
	if m.sortKey != sortNone {
		t.Fatalf("after 3rd s expected none, got %v", m.sortKey)
	}
	if got := names(m.visibleRepos()); !equal(got, []string{"Charlie", "alpha", "Bravo", "delta"}) {
		t.Fatalf("none should restore fetch order, got %v", got)
	}
}

func TestSortReverseTogglesAndNoOpUnderNone(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newSortModel(t, p)

	// S under none is a no-op (direction stays asc, no reorder).
	send(m, key("S"))
	if m.sortDir != sortAsc {
		t.Fatalf("S under none should not change direction, got %v", m.sortDir)
	}
	if got := names(m.visibleRepos()); !equal(got, []string{"Charlie", "alpha", "Bravo", "delta"}) {
		t.Fatalf("S under none should not reorder, got %v", got)
	}

	// Activate name sort, then S toggles to desc.
	send(m, key("s"))
	send(m, key("S"))
	if m.sortDir != sortDesc {
		t.Fatalf("S should toggle to desc, got %v", m.sortDir)
	}
	if got := names(m.visibleRepos()); !equal(got, []string{"delta", "Charlie", "Bravo", "alpha"}) {
		t.Fatalf("name desc after S: got %v", got)
	}

	// S again -> back to asc.
	send(m, key("S"))
	if m.sortDir != sortAsc {
		t.Fatalf("second S should toggle back to asc, got %v", m.sortDir)
	}
}

func TestSortChangeResetsWindow(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newSortModel(t, p)
	m.cursor = 3
	m.offset = 2

	send(m, key("s")) // re-sort
	if m.cursor != 0 || m.offset != 0 {
		t.Fatalf("sort change should reset window, got cursor=%d offset=%d", m.cursor, m.offset)
	}
}
