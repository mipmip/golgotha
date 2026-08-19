package tui

import (
	"strings"
	"testing"
)

func TestFooterPinnedToBottom(t *testing.T) {
	m := drillToMipmipRepos(t) // 2 repos: a short list
	m.width = 80
	m.height = 20
	view := m.View()

	// Every displayed line is newline-terminated, so the count equals the
	// viewport height when the footer is pinned to the bottom.
	if got := strings.Count(view, "\n"); got != 20 {
		t.Fatalf("expected 20 lines (footer pinned to bottom), got %d:\n%s", got, view)
	}
	// A short body means a visible gap of blank lines before the footer.
	if !strings.Contains(view, "\n\n\n\n") {
		t.Fatalf("expected a padding gap above the footer:\n%s", view)
	}
}

func TestFooterNoPadWhenHeightUnknown(t *testing.T) {
	m := drillToMipmipRepos(t)
	m.height = 0 // height not yet known
	view := m.View()
	if got := strings.Count(view, "\n"); got >= 20 {
		t.Fatalf("expected a short, unpadded view when height is unknown, got %d lines", got)
	}
}

func TestFooterFullListNoExtraPad(t *testing.T) {
	m := drillToMipmipRepos(t)
	m.width = 80
	// Height just large enough for header(1)+sep(1)+2 body+footer(2) = 6 lines.
	m.height = 6
	view := m.View()
	if got := strings.Count(view, "\n"); got != 6 {
		t.Fatalf("expected exactly 6 lines with no extra pad, got %d:\n%s", got, view)
	}
}
