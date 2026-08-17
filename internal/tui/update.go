package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model. It is pure with respect to side effects: all
// clone/open/refresh work is deferred to tea.Cmd values returned here.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case cloneResultMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("clone failed %s: %v", msg.Key, msg.Err)
		} else {
			m.markCloned(msg.Provider, msg.Key)
			m.status = fmt.Sprintf("cloned %s", msg.Key)
		}
		return m, nil

	case bulkDoneMsg:
		ok := 0
		for _, r := range msg.Results {
			if r.Err == nil {
				m.markCloned(r.Provider, r.Key)
				m.setSelected(r.Provider, r.Key, false)
				ok++
			}
		}
		m.status = fmt.Sprintf("bulk clone done: %s", fmtCounts(ok, len(msg.Results)))
		return m, nil

	case openResultMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("open failed: %v", msg.Err)
		} else {
			m.status = fmt.Sprintf("opened %s", msg.URL)
		}
		return m, nil

	case refreshResultMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("refresh failed %s: %v", msg.Provider, msg.Err)
			return m, nil
		}
		// Reload the refreshed provider's cache into the model.
		for _, p := range m.providers {
			if p.Name == msg.Provider {
				m.reposByProvider[p.Name] = m.loadProvider(p)
				break
			}
		}
		m.cursor = 0
		m.offset = 0
		m.clampCursor()
		m.status = fmt.Sprintf("refreshed %s: %d repos", msg.Provider, msg.Count)
		return m, nil
	}

	return m, nil
}

// handleKey processes a key message. When the filter input is active most keys
// are routed to the textinput.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Filter-input mode: capture typing; Enter/Esc leave the mode.
	if m.filtering {
		switch msg.Type {
		case tea.KeyEnter:
			m.filtering = false
			m.filter.Blur()
			m.clampCursor()
			return m, nil
		case tea.KeyEsc:
			m.filtering = false
			m.filter.Blur()
			m.filter.SetValue("")
			m.clampCursor()
			return m, nil
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		m.cursor = 0
		m.offset = 0
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		m.moveCursor(-1)
		return m, nil

	case "down", "j":
		m.moveCursor(1)
		return m, nil

	case "pgdown":
		m.moveCursor(m.pageStep())
		return m, nil

	case "pgup":
		m.moveCursor(-m.pageStep())
		return m, nil

	case "ctrl+d":
		m.moveCursor(m.halfPageStep())
		return m, nil

	case "ctrl+u":
		m.moveCursor(-m.halfPageStep())
		return m, nil

	case "home":
		m.cursor = 0
		m.clampCursor()
		return m, nil

	case "end":
		m.cursor = m.rowCount() - 1
		m.clampCursor()
		return m, nil

	case "/":
		m.filtering = true
		m.filter.Focus()
		return m, nil

	case "esc", "backspace":
		return m.goBack()

	case "enter":
		return m.enter()

	case " ":
		return m.toggleSelect()

	case "c":
		return m.clone()

	case "o":
		return m.open()

	case "r":
		return m.refresh()
	}

	return m, nil
}

// goBack pops the navigation stack, or clears an active filter first.
func (m *Model) goBack() (tea.Model, tea.Cmd) {
	if m.filter.Value() != "" {
		m.filter.SetValue("")
		m.cursor = 0
		m.offset = 0
		m.clampCursor()
		return m, nil
	}
	switch m.nav {
	case levelRepos:
		m.nav = levelOwners
		m.selOwner = ""
		m.cursor = 0
		m.offset = 0
	case levelOwners:
		m.nav = levelProviders
		m.selProvider = nil
		m.cursor = 0
		m.offset = 0
	}
	m.clampCursor()
	return m, nil
}

// enter drills into the highlighted row, or clones when already at repos.
func (m *Model) enter() (tea.Model, tea.Cmd) {
	// With an active filter we operate on the flattened repo view directly.
	if m.filter.Value() != "" {
		return m.clone()
	}
	switch m.nav {
	case levelProviders:
		if m.cursor >= 0 && m.cursor < len(m.providers) {
			m.selProvider = m.providers[m.cursor]
			m.nav = levelOwners
			m.cursor = 0
			m.offset = 0
			m.clampCursor()
		}
	case levelOwners:
		owners := m.ownersFor(m.selProvider)
		if m.cursor >= 0 && m.cursor < len(owners) {
			m.selOwner = owners[m.cursor]
			m.nav = levelRepos
			m.cursor = 0
			m.offset = 0
			m.clampCursor()
		}
	case levelRepos:
		return m.clone()
	}
	return m, nil
}

// toggleSelect flips the multi-select flag on the highlighted repo.
func (m *Model) toggleSelect() (tea.Model, tea.Cmd) {
	it, ok := m.currentRepo()
	if !ok {
		return m, nil
	}
	items := m.reposByProvider[it.Provider.Name]
	for i := range items {
		if items[i].key() == it.key() {
			items[i].Selected = !items[i].Selected
			break
		}
	}
	return m, nil
}

// clone clones the multi-selection if any, otherwise the highlighted repo.
func (m *Model) clone() (tea.Model, tea.Cmd) {
	if sel := m.selectedItems(); len(sel) > 0 {
		m.status = fmt.Sprintf("cloning %d selected...", len(sel))
		return m, bulkCloneCmd(m.cloner, sel)
	}
	it, ok := m.currentRepo()
	if !ok {
		return m, nil
	}
	if it.Cloned {
		m.status = fmt.Sprintf("%s already cloned", it.key())
		return m, nil
	}
	m.status = fmt.Sprintf("cloning %s...", it.key())
	return m, cloneCmd(m.cloner, it.Provider, it.Repo)
}

// open opens the highlighted repo's web URL in the browser.
func (m *Model) open() (tea.Model, tea.Cmd) {
	it, ok := m.currentRepo()
	if !ok {
		return m, nil
	}
	if it.Repo.WebURL == "" {
		m.status = fmt.Sprintf("%s has no web URL", it.key())
		return m, nil
	}
	m.status = fmt.Sprintf("opening %s", it.Repo.WebURL)
	return m, openCmd(it.Repo.WebURL)
}

// refresh re-fetches the current provider's cache. It is a no-op when refresh is
// disabled (tests) or no provider is selected.
func (m *Model) refresh() (tea.Model, tea.Cmd) {
	p := m.selProvider
	if p == nil {
		// At the top level, refresh the highlighted provider.
		if m.cursor >= 0 && m.cursor < len(m.providers) {
			p = m.providers[m.cursor]
		}
	}
	if p == nil || m.refresher == nil {
		return m, nil
	}
	m.status = fmt.Sprintf("refreshing %s...", p.Name)
	return m, m.refresher(context.Background(), p)
}
