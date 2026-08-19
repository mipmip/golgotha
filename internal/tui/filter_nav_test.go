package tui

import "testing"

// drillToMipmipRepos navigates to github > mipmip repos (huphop, dotfiles).
func drillToMipmipRepos(t *testing.T) *Model {
	t.Helper()
	m, _ := newTestModel(t)
	send(m, key("enter")) // owners
	send(m, key("j"))     // mipmip
	send(m, key("enter")) // repos of mipmip
	if m.selOwner != "mipmip" {
		t.Fatalf("setup: expected mipmip, got %q", m.selOwner)
	}
	if len(m.visibleRepos()) < 2 {
		t.Fatalf("setup: expected >= 2 repos, got %d", len(m.visibleRepos()))
	}
	return m
}

func TestFilterNavArrowsMoveSelection(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("/"))
	if !m.filtering {
		t.Fatal("expected filtering active")
	}
	// Empty query keeps all repos; arrows move the selection without leaving
	// filter-input mode and without changing the query.
	send(m, key("down"))
	if m.cursor != 1 {
		t.Fatalf("down: cursor=%d, want 1", m.cursor)
	}
	if !m.filtering || m.filter.Value() != "" {
		t.Fatalf("expected still filtering with empty query, filtering=%v q=%q", m.filtering, m.filter.Value())
	}
	send(m, key("up"))
	if m.cursor != 0 {
		t.Fatalf("up: cursor=%d, want 0", m.cursor)
	}
	// Ctrl+N / Ctrl+P also navigate.
	send(m, key("ctrl+n"))
	if m.cursor != 1 {
		t.Fatalf("ctrl+n: cursor=%d, want 1", m.cursor)
	}
	send(m, key("ctrl+p"))
	if m.cursor != 0 {
		t.Fatalf("ctrl+p: cursor=%d, want 0", m.cursor)
	}
}

func TestFilterNavPreservesHighlightAcrossNavKeys(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("/"))
	send(m, key("down")) // cursor -> 1
	send(m, key("down")) // clamps at last (still 1 for 2 repos)
	if m.cursor == 0 {
		t.Fatal("expected highlight preserved/advanced, got 0")
	}
}

func TestFilterTypingResetsToTop(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("/"))
	send(m, key("down")) // move off row 0
	if m.cursor != 1 {
		t.Fatalf("precondition: cursor=%d, want 1", m.cursor)
	}
	// Changing the query resets the selection to the top.
	send(m, key("o")) // matches both huphop/dotfiles; query changed
	if m.cursor != 0 {
		t.Fatalf("typing should reset cursor to 0, got %d", m.cursor)
	}
	if m.filter.Value() != "o" {
		t.Fatalf("expected query 'o', got %q", m.filter.Value())
	}
}

func TestFilterLetterKeysStillType(t *testing.T) {
	m := drillToMipmipRepos(t)
	send(m, key("/"))
	// 'j' and 'k' are letters here: they type, they do not navigate.
	send(m, key("j"))
	send(m, key("k"))
	if m.filter.Value() != "jk" {
		t.Fatalf("expected 'jk' typed into filter, got %q", m.filter.Value())
	}
}
