package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	cursorStyle   = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	footerStyle   = lipgloss.NewStyle().Faint(true)
)

// View implements tea.Model.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	// bodyText mutates m.offset via applyWindow and records the indicator range.
	// Bubble Tea calls Update then View, so writing offset here is safe.
	body := m.bodyText()

	var b strings.Builder
	b.WriteString(titleStyle.Render(m.headerText()))
	b.WriteString("\n\n")
	b.WriteString(body)
	b.WriteString("\n")
	if m.filtering || m.filter.Value() != "" {
		b.WriteString(m.filter.View())
		b.WriteString("\n")
	}
	if m.status != "" {
		b.WriteString(dimStyle.Render(m.status))
		b.WriteString("\n")
	}
	b.WriteString(dimStyle.Render(m.indicatorText))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render(m.footerText()))
	b.WriteString("\n")
	return b.String()
}

// headerText renders the breadcrumb / title for the current scope.
func (m *Model) headerText() string {
	crumb := "skull2"
	if m.selProvider != nil {
		crumb += " > " + m.selProvider.Name
	}
	if m.selOwner != "" {
		crumb += " > " + m.selOwner
	}
	return crumb
}

// bodyText renders the list of rows for the current level.
func (m *Model) bodyText() string {
	if m.filtering || m.filter.Value() != "" {
		return m.renderRepos(m.visibleRepos())
	}
	switch m.nav {
	case levelProviders:
		return m.renderStrings(m.providerNames())
	case levelOwners:
		owners := m.ownersFor(m.selProvider)
		labels := make([]string, len(owners))
		for i, o := range owners {
			labels[i] = ownerLabel(o)
			if !m.fetchedOwners[m.selProvider.Name][o] {
				labels[i] += dimStyle.Render(" (not fetched)")
			}
		}
		return m.renderStrings(labels)
	default:
		if m.fetchingOwner != "" && m.fetchingOwner == m.selOwner {
			m.indicatorText = ""
			return dimStyle.Render(fmt.Sprintf("fetching %s...", ownerLabel(m.selOwner)))
		}
		return m.renderRepos(m.visibleRepos())
	}
}

// renderStrings renders a simple string list with cursor highlighting, windowed
// to the visible slice. It sets m.indicatorText for the current window.
func (m *Model) renderStrings(rows []string) string {
	n := len(rows)
	if n == 0 {
		m.indicatorText = ""
		return dimStyle.Render("(nothing here)")
	}
	first, last := m.applyWindow(n)
	m.indicatorText = indicator(first, last, n)
	var b strings.Builder
	// i is the absolute row index so cursor highlighting tracks correctly.
	for i := first; i < last; i++ {
		prefix := "  "
		line := rows[i]
		if i == m.cursor {
			prefix = "> "
			line = cursorStyle.Render(line)
		}
		b.WriteString(prefix + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// renderRepos renders repo rows with cloned/selected indicators, windowed to the
// visible slice. It sets m.indicatorText for the current window.
func (m *Model) renderRepos(rows []repoItem) string {
	n := len(rows)
	if n == 0 {
		m.indicatorText = ""
		return dimStyle.Render("(no repositories)")
	}
	first, last := m.applyWindow(n)
	m.indicatorText = indicator(first, last, n)
	var b strings.Builder
	// i is the absolute row index so cursor highlighting tracks correctly.
	for i := first; i < last; i++ {
		it := rows[i]
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		mark := "[ ]"
		if it.Selected {
			mark = "[x]"
		}
		cloned := " "
		if it.Cloned {
			cloned = "*"
		}
		line := fmt.Sprintf("%s %s %s", mark, cloned, it.key())
		if i == m.cursor {
			line = cursorStyle.Render(line)
		} else if it.Selected {
			line = selectedStyle.Render(line)
		}
		b.WriteString(prefix + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// indicator formats the position indicator "first-last of n" (1-based) for the
// [first, last) window over n rows.
func indicator(first, last, n int) string {
	return fmt.Sprintf("%d-%d of %d", first+1, last, n)
}

// footerText renders the keybinding help bar.
func (m *Model) footerText() string {
	if m.filtering {
		return "enter: apply  esc: cancel filter"
	}
	return "up/down: move  pgup/pgdn ^u/^d: page  home/end: ends  " +
		"enter: drill/clone  /: filter  space: select  " +
		"c: clone  o: open  r: refresh  esc: back  q: quit"
}
