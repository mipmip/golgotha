package tui

import (
	"strings"
	"testing"
)

// viewLineCount returns the number of rendered lines in a view frame (which has
// no trailing newline).
func viewLineCount(view string) int {
	if view == "" {
		return 0
	}
	return strings.Count(view, "\n") + 1
}

func TestFooterPinnedToBottom(t *testing.T) {
	m := drillToMipmipRepos(t) // 2 repos: a short list
	m.width = 80
	m.height = 20
	view := m.View()

	// The frame is exactly m.height lines (footer pinned to the bottom) with no
	// trailing newline that would scroll the top row off-screen.
	if got := viewLineCount(view); got != 20 {
		t.Fatalf("expected 20 lines (footer pinned to bottom), got %d:\n%s", got, view)
	}
	// A short body means a visible gap of blank lines before the footer.
	if !strings.Contains(view, "\n\n\n\n") {
		t.Fatalf("expected a padding gap above the footer:\n%s", view)
	}
	// The top content (first repo row) must remain the first line.
	firstLine := strings.SplitN(view, "\n", 2)[0]
	if strings.TrimSpace(firstLine) == "" {
		t.Fatalf("top line should hold content, got blank:\n%s", view)
	}
}

func TestFooterNoPadWhenHeightUnknown(t *testing.T) {
	m := drillToMipmipRepos(t)
	m.height = 0 // height not yet known
	view := m.View()
	if got := viewLineCount(view); got >= 20 {
		t.Fatalf("expected a short, unpadded view when height is unknown, got %d lines", got)
	}
}

func TestFooterFullListNoExtraPad(t *testing.T) {
	m := drillToMipmipRepos(t)
	m.width = 80
	// Height just large enough for breadcrumb(1)+sep(1)+2 body+footer(2) = 6 lines.
	m.height = 6
	view := m.View()
	if got := viewLineCount(view); got != 6 {
		t.Fatalf("expected exactly 6 lines with no extra pad, got %d:\n%s", got, view)
	}
}
