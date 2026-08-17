package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// minReadmeWidth / maxReadmeWidth bound the glamour word-wrap width so a very
// narrow or very wide terminal still renders reasonably.
const (
	minReadmeWidth = 40
	maxReadmeWidth = 120
)

// renderMarkdown renders raw markdown to styled terminal text word-wrapped to
// width. It falls back to the raw markdown when glamour cannot build a renderer
// or rendering fails, so the README is always shown in some form. An empty input
// yields "".
func renderMarkdown(md string, width int) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	w := width
	if w < minReadmeWidth {
		w = minReadmeWidth
	}
	if w > maxReadmeWidth {
		w = maxReadmeWidth
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(w),
	)
	if err != nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return out
}
