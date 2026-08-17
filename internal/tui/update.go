package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/golgotha/internal/cache"
	"github.com/mipmip/golgotha/internal/config"
	"github.com/mipmip/golgotha/internal/fetch"
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
			// Keep the open detail view's cloned flag in sync.
			if m.detailRepo.Provider != nil &&
				m.detailRepo.Provider.Name == msg.Provider &&
				m.detailRepo.key() == msg.Key {
				m.detailRepo.Cloned = true
			}
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

	case detailLoadedMsg:
		return m.handleDetailLoaded(msg)

	case progressMsg:
		return m.handleProgress(msg)

	case spinner.TickMsg:
		// Keep the spinner animating only while a fetch is in flight.
		if m.fetchingOwner == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleDetailLoaded folds a resolved repository detail into the model. A stale
// message (for a repo other than the one currently open) is ignored. On error
// with no cache the view degrades gracefully (metadata + "README unavailable").
func (m *Model) handleDetailLoaded(msg detailLoadedMsg) (tea.Model, tea.Cmd) {
	// Ignore results for a repo the user already navigated away from.
	if m.nav != levelDetail ||
		m.detailRepo.Provider == nil ||
		m.detailRepo.Provider.Name != msg.Provider ||
		m.detailRepo.Repo.Owner != msg.Owner ||
		m.detailRepo.Repo.Name != msg.Name {
		return m, nil
	}

	m.detailLoading = false
	if msg.Err != nil {
		// Graceful offline: no cache and the fetch failed.
		m.detailUnavailable = true
		m.detailLoaded = false
		m.status = "README unavailable"
		return m, nil
	}
	m.detail = msg.Details
	m.detailLoaded = true
	m.detailUnavailable = false
	m.readmeRenderedWidth = -1 // force re-render on next View
	if msg.Cached {
		m.status = fmt.Sprintf("%s (cached)", m.detailRepo.key())
	} else {
		m.status = fmt.Sprintf("loaded %s", m.detailRepo.key())
	}
	return m, nil
}

// handleProgress folds one fetch progress event into the model and re-issues a
// wait for the next event (until the channel closes). It is pure: cache commits
// happen in the fetch goroutine, not here.
func (m *Model) handleProgress(msg progressMsg) (tea.Model, tea.Cmd) {
	// Ignore stale events from a fetch the user already backed out of.
	if m.fetchingOwner != msg.Owner {
		return m, nil
	}

	if msg.closed {
		// Channel drained; if we are still marked fetching (no terminal event
		// arrived, e.g. cancel) clear the state.
		if m.fetchingOwner == msg.Owner {
			m.fetchingOwner = ""
			m.fetchCancel = nil
		}
		return m, nil
	}

	ev := msg.Event
	switch ev.Kind {
	case fetch.KindStarted:
		m.status = fmt.Sprintf("fetching %s...", ownerLabel(ev.Owner))

	case fetch.KindPageDone:
		m.fetchPage++
		if ev.TotalPages > 0 {
			m.fetchTotal = ev.TotalPages
		}
		m.fetchRepos = ev.ReposSoFar
		if m.fetchTotal > 0 {
			m.status = fmt.Sprintf("fetching %s page %d/%d — %d repos",
				ownerLabel(ev.Owner), m.fetchPage, m.fetchTotal, m.fetchRepos)
		} else {
			m.status = fmt.Sprintf("fetching %s page %d — %d repos",
				ownerLabel(ev.Owner), m.fetchPage, m.fetchRepos)
		}

	case fetch.KindWarning:
		m.status = fmt.Sprintf("%s: %s", ownerLabel(ev.Owner), ev.Msg)

	case fetch.KindDone:
		m.fetchingOwner = ""
		m.fetchCancel = nil
		m.reloadOwnerFromCache(ev.Provider, ev.Owner)
		m.clampCursor()
		m.status = fmt.Sprintf("fetched %s: %d repos", ownerLabel(ev.Owner), ev.Count)
		return m, waitForProgress(msg.Provider, msg.Owner, msg.ch)

	case fetch.KindFailed:
		m.cancelFetch()
		m.status = fmt.Sprintf("fetch failed %s/%s: %v", ev.Provider, ownerLabel(ev.Owner), ev.Err)
		return m, waitForProgress(msg.Provider, msg.Owner, msg.ch)

	case fetch.KindCanceled:
		m.fetchingOwner = ""
		m.fetchCancel = nil
		m.status = fmt.Sprintf("fetch canceled %s", ownerLabel(ev.Owner))
		return m, waitForProgress(msg.Provider, msg.Owner, msg.ch)
	}

	return m, waitForProgress(msg.Provider, msg.Owner, msg.ch)
}

// reloadOwnerFromCache reloads the just-fetched owner's repos from the persisted
// cache into the model (repos + fetched state). The streaming fetch goroutine
// commits the cache; this makes the result visible in the UI.
func (m *Model) reloadOwnerFromCache(providerName, owner string) {
	c, _, err := cache.LoadOrEmpty(providerName)
	if err != nil {
		return
	}
	m.applyOwnerRepos(providerName, owner, c.ReposFor(owner))
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

	// The detail view has its own key map (scroll the README, clone, open,
	// refresh, back). Handle it before the list-navigation key map.
	if m.nav == levelDetail {
		return m.handleDetailKey(msg)
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

	case "f":
		m.facets.fork = m.facets.fork.cycle()
		return m.applyFacetChange()

	case "a":
		m.facets.archived = m.facets.archived.cycle()
		return m.applyFacetChange()

	case "v":
		m.facets.vis = m.facets.vis.cycle()
		return m.applyFacetChange()

	case "s":
		m.sortKey = m.sortKey.cycle()
		return m.applySortChange()

	case "S":
		// Reversing direction is meaningless with no active sort.
		if m.sortKey == sortNone {
			return m, nil
		}
		m.sortDir = m.sortDir.toggle()
		return m.applySortChange()
	}

	return m, nil
}

// applyFacetChange resets the scroll window after a facet toggles the filter set
// (same rule as a fuzzy keystroke) and refreshes the status line, adding a hint
// when the active facets can't match because the data was excluded at fetch.
func (m *Model) applyFacetChange() (tea.Model, tea.Cmd) {
	m.cursor = 0
	m.offset = 0
	m.clampCursor()
	m.status = m.facetStatus()
	return m, nil
}

// applySortChange resets the scroll window after a re-sort (same rule as a facet
// change), keeps the cursor valid against the new row order, and reflects the
// active sort in the status line.
func (m *Model) applySortChange() (tea.Model, tea.Cmd) {
	m.cursor = 0
	m.offset = 0
	m.clampCursor()
	if m.sortKey == sortNone {
		m.status = "sort: none (fetch order)"
	} else {
		m.status = fmt.Sprintf("sort: %s %s", m.sortKey.label(), m.sortDir.label())
	}
	return m, nil
}

// facetStatus renders the status line for the current facet state, appending a
// hint when a facet was set to reveal data that config excluded before caching
// (Model A can only narrow what is cached).
func (m *Model) facetStatus() string {
	s := m.facets.status()
	if hint := m.facetHint(); hint != "" {
		if s != "" {
			s += "  "
		}
		s += hint
	}
	return s
}

// facetHint returns a hint when a facet selection asks for data that was
// excluded at fetch time (archived/forks not cached because config disabled
// include_archived/include_forks), or "" when no hint applies.
func (m *Model) facetHint() string {
	p := m.selProvider
	if m.facets.archived == triOnly && !cacheIncludesArchived(p) {
		return "archived repos are not cached (include_archived: false)"
	}
	if m.facets.fork == triOnly && !cacheIncludesForks(p) {
		return "forks are not cached (include_forks: false)"
	}
	return ""
}

// cacheIncludesArchived reports whether archived repos are part of the cache
// superset for the provider (config include_archived, default false). A nil
// provider (top-level scope) is treated as not-excluded so no spurious hint.
func cacheIncludesArchived(p *config.Provider) bool {
	if p == nil {
		return true
	}
	if p.IncludeArchived != nil {
		return *p.IncludeArchived
	}
	return false // documented default: archived excluded.
}

// cacheIncludesForks reports whether forks are part of the cache superset for
// the provider (config include_forks, default true).
func cacheIncludesForks(p *config.Provider) bool {
	if p == nil {
		return true
	}
	if p.IncludeForks != nil {
		return *p.IncludeForks
	}
	return true // documented default: forks included.
}

// goBack pops the navigation stack, or clears an active filter first.
func (m *Model) goBack() (tea.Model, tea.Cmd) {
	if m.filter.Value() != "" {
		m.clearFilter()
		m.cursor = 0
		m.offset = 0
		m.clampCursor()
		return m, nil
	}
	switch m.nav {
	case levelRepos:
		// Cancel any in-flight owner fetch; partial results are not cached.
		if m.fetchingOwner != "" {
			m.cancelFetch()
		}
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

// enter drills into the highlighted row, opening the detail view at the repo
// level. Enter consistently means "drill deeper"; cloning is on `c`. The fuzzy
// filter is level-aware: Enter drills the highlighted item from the filtered
// list at the current level, and the filter clears on any level change.
func (m *Model) enter() (tea.Model, tea.Cmd) {
	switch m.nav {
	case levelProviders:
		providers := m.visibleProviders()
		if m.cursor >= 0 && m.cursor < len(providers) {
			m.selProvider = m.providerByName(providers[m.cursor])
			m.nav = levelOwners
			m.clearFilter()
			m.cursor = 0
			m.offset = 0
			m.clampCursor()
		}
	case levelOwners:
		owners := m.visibleOwners()
		if m.cursor >= 0 && m.cursor < len(owners) {
			m.selOwner = owners[m.cursor]
			m.nav = levelRepos
			m.clearFilter()
			m.cursor = 0
			m.offset = 0
			m.clampCursor()
			return m, m.maybeFetchOwner(m.selProvider, m.selOwner)
		}
	case levelRepos:
		return m.openDetail()
	}
	return m, nil
}

// providerByName returns the configured provider with the given name, or nil.
func (m *Model) providerByName(name string) *config.Provider {
	for _, p := range m.providers {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// clearFilter resets the active fuzzy filter (query and input state). It is
// called on any navigation level change so each level starts unfiltered.
func (m *Model) clearFilter() {
	m.filtering = false
	m.filter.SetValue("")
	m.filter.Blur()
}

// openDetail opens the detail view for the highlighted repository. It restores
// tier-2 metadata + README from the detail cache when present (instant, no
// network) and otherwise starts a lazy fetch showing a loading indicator.
func (m *Model) openDetail() (tea.Model, tea.Cmd) {
	it, ok := m.currentRepo()
	if !ok {
		return m, nil
	}

	// Remember the repo-list position so Esc restores it.
	m.detailReturnCursor = m.cursor
	m.detailReturnOffset = m.offset

	m.detailRepo = it
	m.nav = levelDetail
	m.detail = cache.Details{}
	m.detailLoaded = false
	m.detailUnavailable = false
	m.detailLoading = false
	m.readmeRenderedWidth = -1
	m.readme.GotoTop()

	// Reuse the cached detail if present (no network).
	if d, cok, err := cache.LoadDetailsOrEmpty(it.Provider.Name, it.Repo.Owner, it.Repo.Name); err == nil && cok {
		m.detail = d
		m.detailLoaded = true
		m.status = fmt.Sprintf("%s (cached)", it.key())
		return m, nil
	}

	// Lazy fetch on first open.
	if m.detailFetcher == nil {
		return m, nil
	}
	m.detailLoading = true
	m.status = fmt.Sprintf("loading %s...", it.key())
	return m, m.detailFetcher(context.Background(), it.Provider, it.Repo)
}

// closeDetail returns from the detail view to the repo list at the prior
// position, clearing detail state.
func (m *Model) closeDetail() {
	m.nav = levelRepos
	m.cursor = m.detailReturnCursor
	m.offset = m.detailReturnOffset
	m.detailRepo = repoItem{}
	m.detail = cache.Details{}
	m.detailLoaded = false
	m.detailUnavailable = false
	m.detailLoading = false
	m.clampCursor()
}

// handleDetailKey processes keys while the detail view is open: Esc backs out,
// `c` clones, `o` opens in the browser, `r` re-fetches; movement/paging keys
// scroll the README viewport.
func (m *Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc", "backspace":
		m.closeDetail()
		return m, nil

	case "c":
		return m.cloneDetail()

	case "o":
		return m.openDetailRepo()

	case "r":
		return m.refreshDetail()
	}

	// Everything else scrolls the README viewport (up/down/pgup/pgdn/home/end...).
	var cmd tea.Cmd
	m.readme, cmd = m.readme.Update(msg)
	return m, cmd
}

// cloneDetail clones the repo whose detail view is open.
func (m *Model) cloneDetail() (tea.Model, tea.Cmd) {
	it := m.detailRepo
	if it.Provider == nil {
		return m, nil
	}
	if it.Cloned {
		m.status = fmt.Sprintf("%s already cloned", it.key())
		return m, nil
	}
	m.status = fmt.Sprintf("cloning %s...", it.key())
	return m, cloneCmd(m.cloner, it.Provider, it.Repo)
}

// openDetailRepo opens the detail-view repo's web URL in the browser.
func (m *Model) openDetailRepo() (tea.Model, tea.Cmd) {
	it := m.detailRepo
	if it.Repo.WebURL == "" {
		m.status = fmt.Sprintf("%s has no web URL", it.key())
		return m, nil
	}
	m.status = fmt.Sprintf("opening %s", it.Repo.WebURL)
	return m, openCmd(it.Repo.WebURL)
}

// refreshDetail re-fetches the open repo's details and README, bypassing the
// cache. It is a no-op when the fetch seam is disabled (tests).
func (m *Model) refreshDetail() (tea.Model, tea.Cmd) {
	it := m.detailRepo
	if it.Provider == nil || m.detailFetcher == nil {
		return m, nil
	}
	m.detailLoading = true
	m.detailUnavailable = false
	m.status = fmt.Sprintf("re-fetching %s...", it.key())
	return m, m.detailFetcher(context.Background(), it.Provider, it.Repo)
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
	if p == nil {
		return nil
	}
	if m.fetchedOwners[p.Name][owner] {
		return nil // already fetched: instant.
	}
	return m.startFetch(p, owner, "fetching %s...")
}

// forceFetchOwner returns a lazy-fetch command for the owner regardless of its
// cached state (used by `r` to re-fetch the current owner).
func (m *Model) forceFetchOwner(p *config.Provider, owner string) tea.Cmd {
	if p == nil {
		return nil
	}
	return m.startFetch(p, owner, "re-fetching %s...")
}

// startFetch begins a per-owner fetch. It prefers the streaming progressFetcher
// (spinner/bar + cancel); when that is nil it falls back to the one-shot
// ownerFetcher. It returns nil when both seams are disabled (tests).
func (m *Model) startFetch(p *config.Provider, owner, statusFmt string) tea.Cmd {
	m.fetchingOwner = owner
	m.fetchTotal = 0
	m.fetchPage = 0
	m.fetchRepos = 0
	m.status = fmt.Sprintf(statusFmt, ownerLabel(owner))

	if m.progressFetcher != nil {
		ch, cancel := m.progressFetcher(context.Background(), p, owner)
		m.fetchCancel = cancel
		return tea.Batch(m.spinner.Tick, waitForProgress(p.Name, owner, ch))
	}
	if m.ownerFetcher != nil {
		return m.ownerFetcher(context.Background(), p, owner)
	}
	m.fetchingOwner = ""
	return nil
}

// cancelFetch cancels the in-flight streaming fetch (if any) and clears the
// fetch-in-progress state without touching the cache.
func (m *Model) cancelFetch() {
	if m.fetchCancel != nil {
		m.fetchCancel()
		m.fetchCancel = nil
	}
	m.fetchingOwner = ""
	m.fetchTotal = 0
	m.fetchPage = 0
	m.fetchRepos = 0
}

// refresh re-fetches the current scope. When viewing an owner's repos it
// re-fetches just that owner (lazy model); otherwise it refreshes the selected
// or highlighted provider. It is a no-op when the relevant seam is disabled
// (tests) or nothing is selected.
func (m *Model) refresh() (tea.Model, tea.Cmd) {
	// At the repos level, `r` re-fetches the current owner (the level-aware
	// filter only narrows this owner's repos, so it stays the current owner).
	if m.nav == levelRepos && m.selProvider != nil {
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
