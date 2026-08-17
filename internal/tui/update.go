package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/skull2/internal/config"
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
		// Reload the refreshed provider's cache into the model (repos + owner
		// index + fetched state).
		for _, p := range m.providers {
			if p.Name == msg.Provider {
				m.loadProviderInto(p)
				break
			}
		}
		m.cursor = 0
		m.offset = 0
		m.clampCursor()
		m.status = fmt.Sprintf("refreshed %s: %d repos", msg.Provider, msg.Count)
		return m, nil

	case ownerFetchedMsg:
		if m.fetchingOwner == msg.Owner {
			m.fetchingOwner = ""
		}
		if msg.Err != nil {
			m.status = fmt.Sprintf("fetch failed %s/%s: %v", msg.Provider, ownerLabel(msg.Owner), msg.Err)
			return m, nil
		}
		// Populate the model in-memory from the message so the result is applied
		// whether or not the caller persisted a cache (the production fetcher does;
		// tests inject the message). The owner is marked fetched and its repos
		// replace any previously-cached repos for that owner.
		m.applyOwnerRepos(msg.Provider, msg.Owner, msg.Repos)
		m.clampCursor()
		m.status = fmt.Sprintf("fetched %s: %d repos", ownerLabel(msg.Owner), len(msg.Repos))
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
			return m, m.maybeFetchOwner(m.selProvider, m.selOwner)
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

// maybeFetchOwner returns a lazy-fetch command when the owner's repos are not
// yet cached, or nil when they are (instant display) or fetching is disabled.
func (m *Model) maybeFetchOwner(p *config.Provider, owner string) tea.Cmd {
	if p == nil || m.ownerFetcher == nil {
		return nil
	}
	if m.fetchedOwners[p.Name][owner] {
		return nil // already fetched: instant.
	}
	m.fetchingOwner = owner
	m.status = fmt.Sprintf("fetching %s...", ownerLabel(owner))
	return m.ownerFetcher(context.Background(), p, owner)
}

// forceFetchOwner returns a lazy-fetch command for the owner regardless of its
// cached state (used by `r` to re-fetch the current owner).
func (m *Model) forceFetchOwner(p *config.Provider, owner string) tea.Cmd {
	if p == nil || m.ownerFetcher == nil {
		return nil
	}
	m.fetchingOwner = owner
	m.status = fmt.Sprintf("re-fetching %s...", ownerLabel(owner))
	return m.ownerFetcher(context.Background(), p, owner)
}

// refresh re-fetches the current scope. When viewing an owner's repos it
// re-fetches just that owner (lazy model); otherwise it refreshes the selected
// or highlighted provider. It is a no-op when the relevant seam is disabled
// (tests) or nothing is selected.
func (m *Model) refresh() (tea.Model, tea.Cmd) {
	// At the repos level, `r` re-fetches the current owner.
	if m.nav == levelRepos && m.selProvider != nil && m.filter.Value() == "" {
		if cmd := m.forceFetchOwner(m.selProvider, m.selOwner); cmd != nil {
			return m, cmd
		}
		return m, nil
	}

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
