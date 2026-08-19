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
	// selfStyle tints the user's own account in the owner list so it reads as
	// "you" while behaving like an ordinary owner.
	selfStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

// View implements tea.Model.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	if m.nav == levelDetail {
		return m.detailView()
	}

	// bodyText mutates m.offset via applyWindow and records the indicator range.
	// Bubble Tea calls Update then View, so writing offset here is safe.
	body := m.bodyText()
	header, footer := m.modeChrome()

	var b strings.Builder
	var hdr []string
	for _, el := range header {
		if s, ok := m.renderElement(el); ok {
			hdr = append(hdr, s)
		}
	}
	if len(hdr) > 0 {
		b.WriteString(strings.Join(hdr, "\n"))
		b.WriteString("\n\n")
	}
	b.WriteString(body)
	b.WriteString("\n")
	for _, el := range footer {
		if s, ok := m.renderElement(el); ok {
			b.WriteString(s)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// headerText renders the breadcrumb / title for the current scope.
func (m *Model) headerText() string {
	crumb := "hup"
	if m.flatAll {
		return crumb + " > all repositories · " + m.combinedBadge()
	}
	if m.selProvider != nil {
		crumb += " > " + m.selProvider.Name
	}
	if m.selOwner != "" {
		crumb += " > " + m.selOwner
	}
	return crumb
}

// bodyText renders the list of rows for the current level. The fuzzy filter is
// level-aware: it narrows the current level's items rather than flattening to a
// repository search.
func (m *Model) bodyText() string {
	switch m.nav {
	case levelProviders:
		return m.renderStrings(m.providerRows())
	case levelOwners:
		owners := m.visibleOwners()
		labels := make([]string, len(owners))
		for i, o := range owners {
			label := ownerLabel(o)
			if o == m.selProvider.Username {
				label = selfStyle.Render(label)
			}
			labels[i] = label
			if !m.fetchedOwners[m.selProvider.Name][o] {
				labels[i] += dimStyle.Render(" (not fetched)")
			}
		}
		return m.renderStrings(labels)
	default:
		if m.fetchingOwner != "" && m.fetchingOwner == m.selOwner {
			m.indicatorText = ""
			return m.fetchProgressView()
		}
		return m.renderRepos(m.visibleRepos())
	}
}

// fetchProgressView renders the in-flight fetch indicator: an indeterminate
// spinner until the total page count is known, then a determinate bar plus a
// "fetching <owner> page i/n — N repos" line.
func (m *Model) fetchProgressView() string {
	owner := ownerLabel(m.selOwner)
	if m.fetchTotal <= 0 {
		// Total unknown: spinner + running counts.
		line := fmt.Sprintf("fetching %s — %d repos", owner, m.fetchRepos)
		if m.fetchPage > 0 {
			line = fmt.Sprintf("fetching %s page %d — %d repos", owner, m.fetchPage, m.fetchRepos)
		}
		return m.spinner.View() + " " + dimStyle.Render(line)
	}
	// Total known: determinate bar + descriptive line.
	frac := float64(m.fetchPage) / float64(m.fetchTotal)
	if frac > 1 {
		frac = 1
	}
	bar := m.progress.ViewAs(frac)
	line := fmt.Sprintf("fetching %s page %d/%d — %d repos",
		owner, m.fetchPage, m.fetchTotal, m.fetchRepos)
	return bar + "\n" + dimStyle.Render(line)
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
		if hint := m.facetHint(); hint != "" {
			return dimStyle.Render("(no repositories) — " + hint)
		}
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
		label := it.key()
		if m.flatAll && it.Provider != nil {
			// Disambiguate across providers in the combined view.
			label = it.Provider.Short + "  " + label
		}
		line := fmt.Sprintf("%s %s %s", mark, cloned, label)
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

// detailView renders the single-repository detail view: a metadata header
// (description, stars, topics, language, updated, visibility) followed by the
// scrollable rendered README (or a loading / unavailable note).
func (m *Model) detailView() string {
	it := m.detailRepo
	var b strings.Builder

	// Breadcrumb / title.
	crumb := "hup"
	if it.Provider != nil {
		crumb += " > " + it.Provider.Name
	}
	crumb += " > " + it.key()
	b.WriteString(titleStyle.Render(crumb))
	b.WriteString("\n\n")

	// Metadata header.
	b.WriteString(m.detailHeader())
	b.WriteString("\n")

	// README body.
	switch {
	case m.detailLoading:
		b.WriteString(m.spinner.View() + " " + dimStyle.Render("loading details..."))
		b.WriteString("\n")
	case m.detailUnavailable:
		b.WriteString(dimStyle.Render("README unavailable"))
		b.WriteString("\n")
	default:
		b.WriteString(m.readmeBody())
		b.WriteString("\n")
	}

	if m.status != "" {
		b.WriteString(dimStyle.Render(m.status))
		b.WriteString("\n")
	}
	b.WriteString(footerStyle.Render(m.detailFooterText()))
	b.WriteString("\n")
	return b.String()
}

// detailHeader renders the tier-1/tier-2 metadata lines for the detail view.
func (m *Model) detailHeader() string {
	it := m.detailRepo
	r := it.Repo
	var b strings.Builder

	if r.Description != "" {
		b.WriteString(r.Description)
		b.WriteString("\n")
	}

	// stars/language/visibility line (tier-2 shown once loaded).
	parts := []string{}
	if m.detailLoaded {
		parts = append(parts, fmt.Sprintf("★ %d", m.detail.Stars))
		if m.detail.Language != "" {
			parts = append(parts, "lang: "+m.detail.Language)
		}
	}
	parts = append(parts, "visibility: "+r.Visibility)
	if !r.UpdatedAt.IsZero() {
		parts = append(parts, "updated: "+r.UpdatedAt.Format("2006-01-02"))
	}
	if it.Cloned {
		parts = append(parts, "cloned")
	}
	b.WriteString(dimStyle.Render(strings.Join(parts, "  ")))
	b.WriteString("\n")

	if m.detailLoaded && len(m.detail.Topics) > 0 {
		b.WriteString(dimStyle.Render("topics: " + strings.Join(m.detail.Topics, ", ")))
		b.WriteString("\n")
	}
	return b.String()
}

// readmeBody renders (once per width) the raw README into the viewport and
// returns the viewport view. An empty README shows a placeholder.
func (m *Model) readmeBody() string {
	width := m.width
	if width <= 0 {
		width = maxReadmeWidth
	}

	if strings.TrimSpace(m.detail.ReadmeMarkdown) == "" {
		return dimStyle.Render("(no README)")
	}

	// Re-render only when the width changed since the last render.
	if m.readmeRenderedWidth != width {
		rendered := renderMarkdown(m.detail.ReadmeMarkdown, width)
		m.readme.Width = width
		height := m.detailViewportHeight()
		if height > 0 {
			m.readme.Height = height
		}
		m.readme.SetContent(rendered)
		m.readmeRenderedWidth = width
	}
	return m.readme.View()
}

// detailViewportHeight returns the number of terminal rows available to the
// README viewport (total height minus the detail chrome), or 0 when the height
// is unknown (render everything).
func (m *Model) detailViewportHeight() int {
	if m.height <= 0 {
		return 0
	}
	// chrome: title(1) + blank(1) + header lines + status(1) + footer(1).
	chrome := 4
	chrome += strings.Count(m.detailHeader(), "\n")
	h := m.height - chrome
	if h < 1 {
		h = 1
	}
	return h
}

// detailFooterText renders the detail-view keybinding help bar.
func (m *Model) detailFooterText() string {
	return "up/down/pgup/pgdn: scroll  c: clone  o: open  r: refresh  esc: back  q: quit"
}

// footerText renders the keybinding help bar.
func (m *Model) footerText() string {
	if m.filtering {
		return "enter: apply  esc: cancel filter"
	}
	sortHelp := "s: sort  S: reverse"
	if m.sortKey != sortNone {
		sortHelp = fmt.Sprintf("s: sort [%s %s]  S: reverse", m.sortKey.label(), m.sortDir.label())
	}
	return "up/down: move  pgup/pgdn ^u/^d: page  home/end: ends  " +
		"enter: drill/details  /: filter  f/a/v: fork/archived/vis  " + sortHelp + "  space: select  " +
		"c: clone  o: open  r: refresh  esc: back  q: quit"
}
