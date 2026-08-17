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

	var b strings.Builder
	b.WriteString(titleStyle.Render(m.headerText()))
	b.WriteString("\n\n")
	b.WriteString(m.bodyText())
	b.WriteString("\n")
	if m.filtering || m.filter.Value() != "" {
		b.WriteString(m.filter.View())
		b.WriteString("\n")
	}
	if m.status != "" {
		b.WriteString(dimStyle.Render(m.status))
		b.WriteString("\n")
	}
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
		return m.renderStrings(m.ownersFor(m.selProvider))
	default:
		return m.renderRepos(m.visibleRepos())
	}
}

// renderStrings renders a simple string list with cursor highlighting.
func (m *Model) renderStrings(rows []string) string {
	if len(rows) == 0 {
		return dimStyle.Render("(nothing here)")
	}
	var b strings.Builder
	for i, r := range rows {
		prefix := "  "
		line := r
		if i == m.cursor {
			prefix = "> "
			line = cursorStyle.Render(line)
		}
		b.WriteString(prefix + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// renderRepos renders repo rows with cloned/selected indicators.
func (m *Model) renderRepos(rows []repoItem) string {
	if len(rows) == 0 {
		return dimStyle.Render("(no repositories)")
	}
	var b strings.Builder
	for i, it := range rows {
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

// footerText renders the keybinding help bar.
func (m *Model) footerText() string {
	if m.filtering {
		return "enter: apply  esc: cancel filter"
	}
	return "up/down: move  enter: drill/clone  /: filter  space: select  " +
		"c: clone  o: open  r: refresh  esc: back  q: quit"
}
