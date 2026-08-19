package tui

import (
	"fmt"
	"strings"
)

// builtinManagementChrome is the fallback chrome (used when no modes are
// configured or the active mode is unknown): it reproduces the standard layout.
func builtinManagementChrome() (header, footer []string) {
	return []string{"breadcrumb"},
		[]string{"filter", "facet_status", "status_message", "position_indicator", "action_menu"}
}

// modeChrome returns the active mode's ordered header and footer element lists,
// falling back to the built-in management chrome when the mode is absent.
func (m *Model) modeChrome() (header, footer []string) {
	bh, bf := builtinManagementChrome()
	if m.cfg == nil {
		return bh, bf
	}
	name := m.mode
	if name == "" {
		name = m.cfg.DefaultMode
	}
	if name == "" {
		return bh, bf
	}
	mc, ok := m.cfg.Modes[name]
	if !ok {
		return bh, bf
	}
	return mc.Header, mc.Footer
}

// SetMode selects the active TUI mode, validating it against the configured
// modes. It is used by the CLI --mode flag.
func (m *Model) SetMode(name string) error {
	if m.cfg != nil && len(m.cfg.Modes) > 0 {
		if _, ok := m.cfg.Modes[name]; !ok {
			return fmt.Errorf("unknown mode %q", name)
		}
	}
	m.mode = name
	return nil
}

// renderElement renders a named chrome element for the current state. The bool
// reports whether the element occupies space; false means it is skipped
// entirely (e.g. an inactive filter). Always-on elements return true even when
// their text is empty so they hold a stable line (and stable window height).
func (m *Model) renderElement(name string) (string, bool) {
	switch name {
	case "breadcrumb":
		return titleStyle.Render(m.headerText()), true
	case "filter":
		if m.filtering || m.filter.Value() != "" {
			return m.filter.View(), true
		}
		return "", false
	case "facet_status":
		if fs := m.facets.status(); fs != "" {
			return dimStyle.Render("filters: " + fs), true
		}
		return "", false
	case "status_message", "clone_status":
		if m.status != "" {
			return dimStyle.Render(m.status), true
		}
		return "", false
	case "position_indicator":
		return dimStyle.Render(m.indicatorText), true
	case "action_menu":
		return footerStyle.Render(m.footerText()), true
	case "switch_hint":
		return footerStyle.Render("enter: clone & switch  esc: back"), true
	default:
		return "", false
	}
}

// countLines returns the number of terminal lines a rendered string occupies.
func countLines(s string) int { return 1 + strings.Count(s, "\n") }
