// Package tui implements the Bubble Tea interactive browser for huphop: it reads
// the per-provider cache and lets the user navigate provider -> owner -> repos,
// fuzzy-filter, clone (single or bulk), open a repo in the browser and refresh a
// provider's cache.
//
// Side effects (clone, open, refresh) run inside tea.Cmd functions that return
// result messages, keeping Update pure and unit-testable without a terminal.
package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/cache"
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/fetch"
	"github.com/mipmip/huphop/internal/provider"
)

// sortKey selects the repo-list ordering. The zero value (none) preserves the
// provider/cache fetch order, so the default behavior is unchanged.
type sortKey int

const (
	sortNone    sortKey = iota // provider/cache fetch order (no sort)
	sortName                   // alphabetical by Repo.Name (case-insensitive)
	sortUpdated                // by Repo.UpdatedAt
)

// cycle advances the sort key none -> name -> updated -> none.
func (k sortKey) cycle() sortKey { return (k + 1) % 3 }

// label renders the sort key for the footer legend.
func (k sortKey) label() string {
	switch k {
	case sortName:
		return "name"
	case sortUpdated:
		return "updated"
	default:
		return "none"
	}
}

// sortDir is the sort direction. The zero value is ascending.
type sortDir int

const (
	sortAsc  sortDir = iota // ascending
	sortDesc                // descending
)

// toggle flips the direction asc <-> desc.
func (d sortDir) toggle() sortDir {
	if d == sortAsc {
		return sortDesc
	}
	return sortAsc
}

// label renders the sort direction for the footer legend.
func (d sortDir) label() string {
	if d == sortDesc {
		return "desc"
	}
	return "asc"
}

// level identifies the current navigation depth.
type level int

const (
	levelProviders level = iota // list of configured providers
	levelOwners                 // owners within the selected provider
	levelRepos                  // repos within the selected owner
	levelDetail                 // single-repo detail view (metadata + README)
)

// repoItem is a repository plus its resolved display/clone state.
type repoItem struct {
	Repo     provider.Repo
	Provider *config.Provider
	// Target is the resolved clone target path (empty if it could not resolve).
	Target string
	// Cloned reports whether the target already exists as a git repo.
	Cloned bool
	// Selected reports whether the row is multi-selected for bulk clone.
	Selected bool
}

// key returns the "owner/name" key used for display and fuzzy matching.
func (r repoItem) key() string { return r.Repo.Owner + "/" + r.Repo.Name }

// Cloner clones a single repository to its templated target. The syncer Engine
// satisfies this; tests can inject a fake.
type Cloner interface {
	CloneRepo(ctx context.Context, p *config.Provider, r provider.Repo) syncerResult
}

// syncerResult mirrors syncer.Result for the fields the TUI needs, decoupling
// the model from the concrete engine so it stays testable. Bridging happens in
// New via engineCloner.
type syncerResult struct {
	Target  string
	Cloned  bool
	Err     error
	Warning string
}

// Model is the root Bubble Tea model.
type Model struct {
	cfg *config.Config

	// providers is the configured provider list.
	providers []*config.Provider
	// reposByProvider holds cache-loaded repo items keyed by provider name.
	reposByProvider map[string][]repoItem
	// ownersByProvider holds the cached owner index (incl. discovered-but-
	// unfetched owners) keyed by provider name. Empty for a legacy/eager cache
	// with no owner index, in which case ownersFor falls back to repo-derived
	// owners.
	ownersByProvider map[string][]string
	// fetchedOwners tracks, per provider, which owners have had their repos
	// fetched (so the TUI knows whether entering an owner needs a lazy fetch).
	fetchedOwners map[string]map[string]bool

	// nav is the current navigation level.
	nav level
	// mode is the active TUI mode name (management, multiplex, ...). Empty falls
	// back to the built-in management chrome.
	mode string
	// flatAll is true in the combined "All repositories" view: a flat list of
	// every cached repo across all providers and owners (selProvider is nil).
	flatAll bool
	// runCommand executes a multiplex switch_command's argv (no shell). Injectable
	// for tests; defaults to running the process.
	runCommand func(argv []string) error
	// selProvider / selOwner track the drilled-down selection.
	selProvider *config.Provider
	selOwner    string

	// cursor is the highlighted row index into the currently visible list.
	cursor int
	// offset is the index of the top visible row (scroll position) for the
	// current level. Reset to 0 wherever the cursor resets.
	offset int
	// indicatorText is the position indicator (e.g. "41-60 of 213") computed by
	// the render pass and shown by View. It is a transient render artifact.
	indicatorText string

	// filtering reports whether the fuzzy filter input is active.
	filtering bool
	filter    textinput.Model

	// facets holds the interactive tri-state/value-cycle facet filters applied
	// to the repo list, composed (AND) with the fuzzy query. In-memory for the
	// session; defaults to all (no constraint).
	facets facets

	// sortKey / sortDir hold the interactive repo-list sort. Zero values
	// (none/asc) preserve the provider/cache fetch order, so the default is
	// unchanged. The sort is applied after filtering, ordering only the visible
	// subset. In-memory for the session.
	sortKey sortKey
	sortDir sortDir

	// status is a one-line status/progress message shown in the footer.
	status string

	// cloner performs clones (behind an interface for tests).
	cloner Cloner
	// refresher re-fetches a provider's cache; nil disables refresh (tests).
	refresher func(ctx context.Context, p *config.Provider) tea.Cmd
	// ownerFetcher lazily fetches one owner's repos and caches them; nil disables
	// lazy fetch (tests inject a fake or the ownerFetchedMsg directly). Used as a
	// fallback when progressFetcher is nil.
	ownerFetcher func(ctx context.Context, p *config.Provider, owner string) tea.Cmd
	// progressFetcher starts a streaming, page-aware fetch for one owner, returning
	// the event channel and a cancel func. nil falls back to ownerFetcher. It is
	// the production path; tests feed progressMsg directly.
	progressFetcher func(ctx context.Context, p *config.Provider, owner string) (<-chan fetch.Event, context.CancelFunc)
	// fetchingOwner is the owner currently being lazily fetched (for the loading
	// line), or "" when none.
	fetchingOwner string

	// spinner is the indeterminate indicator shown until the total is known.
	spinner spinner.Model
	// progress is the determinate bar shown once the total page count is known.
	progress progress.Model
	// fetchTotal is the known total page count for the current fetch (0 = unknown).
	fetchTotal int
	// fetchPage is the number of pages completed so far for the current fetch.
	fetchPage int
	// fetchRepos is the running repo count for the current fetch.
	fetchRepos int
	// fetchCancel cancels the in-flight streaming fetch; nil when none.
	fetchCancel context.CancelFunc

	// --- Detail view (levelDetail) state ---
	// detailRepo is the repo whose detail view is open (tier-1 metadata source).
	detailRepo repoItem
	// detailReturnCursor / detailReturnOffset restore the repo-list position when
	// Esc backs out of the detail view.
	detailReturnCursor int
	detailReturnOffset int
	// detailLoading reports that a detail fetch is in flight (loading indicator).
	detailLoading bool
	// detailLoaded reports whether details (tier-2 + README) are available for the
	// open repo (from cache or a completed fetch).
	detailLoaded bool
	// detail holds the loaded tier-2 metadata + raw README for the open repo.
	detail cache.Details
	// detailUnavailable reports that the fetch failed with no cache, so the view
	// shows metadata plus a "README unavailable" note (graceful offline).
	detailUnavailable bool
	// readme is the scrollable rendered-README viewport.
	readme viewport.Model
	// readmeRendered caches the rendered README for the width it was rendered at,
	// so View only re-renders on a width change.
	readmeRenderedWidth int
	// detailFetcher lazily fetches one repo's details+README and caches them,
	// returning a detailLoadedMsg; nil disables lazy fetch (tests inject the msg).
	detailFetcher func(ctx context.Context, p *config.Provider, r provider.Repo) tea.Cmd

	// checkCloned reports whether target already exists as a git repo. A var so
	// tests can stub the filesystem check.
	checkCloned func(target string) bool

	// width/height track the terminal size for rendering.
	width  int
	height int

	quitting bool
}

// New builds a Model from the configuration, loading the per-provider caches.
// It resolves each repo's clone target and cloned state up front.
func New(cfg *config.Config) *Model {
	m := &Model{
		cfg:              cfg,
		reposByProvider:  map[string][]repoItem{},
		ownersByProvider: map[string][]string{},
		fetchedOwners:    map[string]map[string]bool{},
		nav:              levelProviders,
		mode:             cfg.DefaultMode,
		runCommand:       runArgv,
		checkCloned:      isGitRepo,
	}
	for i := range cfg.Providers {
		m.providers = append(m.providers, &cfg.Providers[i])
	}

	m.filter = textinput.New()
	m.filter.Placeholder = "fuzzy filter (owner/name)"
	m.filter.Prompt = "/ "

	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Dot
	m.progress = progress.New(progress.WithoutPercentage())
	m.readme = viewport.New(0, 0)

	m.cloner = newEngineCloner(cfg)
	m.refresher = defaultRefresher(cfg, m)
	m.ownerFetcher = defaultOwnerFetcher(cfg)
	m.progressFetcher = defaultProgressFetcher(cfg)
	m.detailFetcher = defaultDetailFetcher(cfg)

	m.loadCaches()
	return m
}

// loadCaches (re)loads all provider caches into reposByProvider.
func (m *Model) loadCaches() {
	for _, p := range m.providers {
		m.loadProviderInto(p)
	}
}

// loadProviderInto loads one provider's cache into the repo items, owner index
// and fetched-owner state.
func (m *Model) loadProviderInto(p *config.Provider) {
	if m.ownersByProvider == nil {
		m.ownersByProvider = map[string][]string{}
	}
	if m.fetchedOwners == nil {
		m.fetchedOwners = map[string]map[string]bool{}
	}
	m.reposByProvider[p.Name] = m.loadProvider(p)

	c, _, err := cache.LoadOrEmpty(p.Name)
	if err != nil {
		m.ownersByProvider[p.Name] = nil
		m.fetchedOwners[p.Name] = map[string]bool{}
		return
	}
	m.ownersByProvider[p.Name] = c.OwnerNames()
	fetched := map[string]bool{}
	for _, name := range c.OwnerNames() {
		fetched[name] = c.OwnerFetched(name)
	}
	m.fetchedOwners[p.Name] = fetched
}

// loadProvider loads one provider's cache into repoItems with resolved state.
func (m *Model) loadProvider(p *config.Provider) []repoItem {
	c, _, err := cache.LoadOrEmpty(p.Name)
	if err != nil {
		return nil
	}
	repos := provider.FilterRepos(p, c.Repos)
	items := make([]repoItem, 0, len(repos))
	for _, r := range repos {
		it := repoItem{Repo: r, Provider: p}
		if target, terr := resolveTarget(m.cfg, p, r); terr == nil {
			it.Target = target
			if m.checkCloned != nil {
				it.Cloned = m.checkCloned(target)
			}
		}
		items = append(items, it)
	}
	return items
}

// applyOwnerRepos replaces a provider owner's repo items with freshly fetched
// repos (resolving clone targets/state), marks the owner fetched and ensures it
// is in the owner index. It updates only that owner's rows.
func (m *Model) applyOwnerRepos(providerName, owner string, repos []provider.Repo) {
	var p *config.Provider
	for _, cand := range m.providers {
		if cand.Name == providerName {
			p = cand
			break
		}
	}
	if p == nil {
		return
	}

	// Rebuild the provider's items: keep other owners, replace this owner's.
	filtered := provider.FilterRepos(p, repos)
	kept := make([]repoItem, 0, len(m.reposByProvider[providerName])+len(filtered))
	for _, it := range m.reposByProvider[providerName] {
		if it.Repo.Owner != owner {
			kept = append(kept, it)
		}
	}
	for _, r := range filtered {
		it := repoItem{Repo: r, Provider: p}
		if target, terr := resolveTarget(m.cfg, p, r); terr == nil {
			it.Target = target
			if m.checkCloned != nil {
				it.Cloned = m.checkCloned(target)
			}
		}
		kept = append(kept, it)
	}
	m.reposByProvider[providerName] = kept

	if m.ownersByProvider == nil {
		m.ownersByProvider = map[string][]string{}
	}
	if m.fetchedOwners == nil {
		m.fetchedOwners = map[string]map[string]bool{}
	}
	if m.fetchedOwners[providerName] == nil {
		m.fetchedOwners[providerName] = map[string]bool{}
	}
	m.fetchedOwners[providerName][owner] = true

	// Ensure the owner is in the index.
	found := false
	for _, o := range m.ownersByProvider[providerName] {
		if o == owner {
			found = true
			break
		}
	}
	if !found {
		m.ownersByProvider[providerName] = append(m.ownersByProvider[providerName], owner)
	}
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// providerNames returns provider display names in config order.
func (m *Model) providerNames() []string {
	names := make([]string, 0, len(m.providers))
	for _, p := range m.providers {
		names = append(names, p.Name)
	}
	return names
}

// ownerLabel renders an owner name for display. Owners are real logins (the self
// account is the configured Username), so this is a plain passthrough kept as the
// single display seam for owner names.
func ownerLabel(owner string) string {
	return owner
}

// pinSelfFirst moves self (the provider's Username) to the front of a sorted
// owner list so the user's own account leads, behaving like an ordinary owner
// otherwise. It is a no-op when self is absent.
func pinSelfFirst(owners []string, self string) []string {
	for i, o := range owners {
		if o == self {
			out := make([]string, 0, len(owners))
			out = append(out, o)
			out = append(out, owners[:i]...)
			out = append(out, owners[i+1:]...)
			return out
		}
	}
	return owners
}

// ownersFor returns the owner names for a provider, with the self account (the
// provider's Username) pinned first. It uses the cached owner index (including
// discovered-but-unfetched owners) when present, and falls back to distinct
// repo-derived owners for a legacy/eager cache without an index.
func (m *Model) ownersFor(p *config.Provider) []string {
	if idx := m.ownersByProvider[p.Name]; len(idx) > 0 {
		// Preserve the index order but keep it stable and de-duplicated.
		seen := map[string]struct{}{}
		owners := make([]string, 0, len(idx))
		for _, o := range idx {
			if _, ok := seen[o]; ok {
				continue
			}
			seen[o] = struct{}{}
			owners = append(owners, o)
		}
		sort.Strings(owners)
		return pinSelfFirst(owners, p.Username)
	}

	seen := map[string]struct{}{}
	var owners []string
	for _, it := range m.reposByProvider[p.Name] {
		if _, ok := seen[it.Repo.Owner]; ok {
			continue
		}
		seen[it.Repo.Owner] = struct{}{}
		owners = append(owners, it.Repo.Owner)
	}
	sort.Strings(owners)
	return pinSelfFirst(owners, p.Username)
}

// filtered returns the subset of items whose value fuzzy-matches the active
// filter query, using match to derive the string matched against for each item.
// An empty query returns the input unchanged. It keeps windowing/rendering
// uniform across levels by centralizing the per-level narrowing.
func filtered[T any](items []T, query string, match func(T) string) []T {
	if query == "" {
		return items
	}
	out := make([]T, 0, len(items))
	for _, it := range items {
		if fuzzyMatch(match(it), query) {
			out = append(out, it)
		}
	}
	return out
}

// visibleProviders returns the provider names for the providers level, narrowed
// by the active fuzzy filter (matched against the raw provider name).
func (m *Model) visibleProviders() []string {
	return filtered(m.providerNames(), m.filter.Value(), func(s string) string { return s })
}

// visibleOwners returns the owner names for the selected provider, narrowed by
// the active fuzzy filter. Matching is against the raw owner name, not any
// decorated display label (e.g. "(not fetched)").
func (m *Model) visibleOwners() []string {
	if m.selProvider == nil {
		return nil
	}
	return filtered(m.ownersFor(m.selProvider), m.filter.Value(), func(s string) string { return s })
}

// visibleRepos returns the repo items visible at the current scope. When the
// filter query is non-empty it matches across the current scope (the selected
// owner's repos, or the whole provider when no owner is chosen, or all
// providers at the top level). Otherwise it returns exactly the selected owner's
// repos.
func (m *Model) visibleRepos() []repoItem {
	query := m.filter.Value()

	var scope []repoItem
	switch {
	case m.selOwner != "" && m.selProvider != nil:
		for _, it := range m.reposByProvider[m.selProvider.Name] {
			if it.Repo.Owner == m.selOwner {
				scope = append(scope, it)
			}
		}
	case m.selProvider != nil:
		scope = m.reposByProvider[m.selProvider.Name]
	default:
		for _, p := range m.providers {
			scope = append(scope, m.reposByProvider[p.Name]...)
		}
	}

	// Apply the facets first (AND together), then the fuzzy query narrows the
	// result further. Facets are independent of the fuzzy input.
	if m.facets.active() {
		narrowed := scope[:0:0]
		for _, it := range scope {
			if m.facets.match(it.Repo) {
				narrowed = append(narrowed, it)
			}
		}
		scope = narrowed
	}

	if query != "" {
		out := scope[:0:0]
		for _, it := range scope {
			match := it.key()
			if m.flatAll && it.Provider != nil {
				// Let the provider short participate in fuzzy matching so users
				// can filter the combined list by provider (e.g. "gh").
				match = it.Provider.Short + " " + match
			}
			if fuzzyMatch(match, query) {
				out = append(out, it)
			}
		}
		scope = out
	}

	// Sort last so it orders exactly the visible (filtered) subset, at every
	// scope. sortNone leaves the slice in fetch order.
	return m.sortRepos(scope)
}

// sortRepos returns items ordered by the active sort key and direction. It is
// stable (fetch order is the tiebreaker) so equal keys don't jump around. When
// sortKey is sortNone the slice is returned unchanged (fetch order).
func (m *Model) sortRepos(items []repoItem) []repoItem {
	if m.sortKey == sortNone {
		return items
	}
	// Sort a copy: the scope slice may alias the cached fetch-order slice
	// (provider-level scope returns it directly), and sorting in place would
	// destroy the fetch order that sortNone must be able to restore.
	sorted := make([]repoItem, len(items))
	copy(sorted, items)
	items = sorted
	sort.SliceStable(items, func(i, j int) bool {
		var less bool
		switch m.sortKey {
		case sortName:
			less = strings.ToLower(items[i].Repo.Name) < strings.ToLower(items[j].Repo.Name)
		case sortUpdated:
			less = items[i].Repo.UpdatedAt.Before(items[j].Repo.UpdatedAt)
		}
		if m.sortDir == sortDesc {
			return !less
		}
		return less
	})
	return items
}

// allReposLabel is the synthetic top-of-providers entry that opens the combined
// cross-provider flat repository view.
const allReposLabel = "All repositories"

// providerRows returns the rows shown at the providers level: the fuzzy-filtered
// provider names followed by the synthetic "All repositories" entry.
func (m *Model) providerRows() []string {
	provs := m.visibleProviders()
	rows := make([]string, 0, len(provs)+1)
	rows = append(rows, provs...)
	rows = append(rows, allReposLabel)
	return rows
}

// combinedBadge summarizes the combined view: total cached repos and how many
// owners are loaded out of the total known owners across all providers.
func (m *Model) combinedBadge() string {
	repos := 0
	for _, p := range m.providers {
		repos += len(m.reposByProvider[p.Name])
	}
	total, loaded := 0, 0
	for _, p := range m.providers {
		owners := m.ownersFor(p)
		total += len(owners)
		for _, o := range owners {
			if m.fetchedOwners[p.Name][o] {
				loaded++
			}
		}
	}
	if total == 0 {
		return fmt.Sprintf("%d repos", repos)
	}
	badge := fmt.Sprintf("%d repos · %d/%d owners loaded", repos, loaded, total)
	if loaded < total {
		badge += " (r: refresh all)"
	}
	return badge
}

// rowCount returns the number of rows shown at the current level, honoring the
// active fuzzy filter which narrows the current level's items.
func (m *Model) rowCount() int {
	switch m.nav {
	case levelProviders:
		return len(m.providerRows())
	case levelOwners:
		return len(m.visibleOwners())
	default:
		return len(m.visibleRepos())
	}
}

// clampCursor keeps the cursor within [0, rowCount).
func (m *Model) clampCursor() {
	n := m.rowCount()
	if n == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
}

// moveCursor moves the cursor by delta rows and clamps it. The window follows
// on the next render via applyWindow.
func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	m.clampCursor()
}

// pageStep is the number of rows a PgUp/PgDn moves: one visible screen. It falls
// back to the current list length when the height is unknown (sentinel).
func (m *Model) pageStep() int {
	v := m.visibleRows()
	if v <= 0 {
		if n := m.rowCount(); n > 0 {
			return n
		}
		return 1
	}
	return v
}

// halfPageStep is the number of rows a Ctrl-U/Ctrl-D moves: half a visible
// screen, at least 1.
func (m *Model) halfPageStep() int {
	step := m.pageStep() / 2
	if step < 1 {
		step = 1
	}
	return step
}

// currentRepo returns the highlighted repo item when at the repos level. ok is
// false when there is no such row. The fuzzy filter narrows the repo list at
// this level but does not flatten other levels into repos.
func (m *Model) currentRepo() (repoItem, bool) {
	if m.nav != levelRepos {
		return repoItem{}, false
	}
	repos := m.visibleRepos()
	if m.cursor < 0 || m.cursor >= len(repos) {
		return repoItem{}, false
	}
	return repos[m.cursor], true
}

// selectedItems returns all multi-selected repo items across the whole model.
func (m *Model) selectedItems() []repoItem {
	var out []repoItem
	for _, p := range m.providers {
		for _, it := range m.reposByProvider[p.Name] {
			if it.Selected {
				out = append(out, it)
			}
		}
	}
	return out
}

// setSelectedByKey sets the Selected flag for the item matching provider+key.
func (m *Model) setSelected(prov, key string, sel bool) {
	items := m.reposByProvider[prov]
	for i := range items {
		if items[i].key() == key {
			items[i].Selected = sel
			return
		}
	}
}

// markCloned marks the item matching provider+key as cloned.
func (m *Model) markCloned(prov, key string) {
	items := m.reposByProvider[prov]
	for i := range items {
		if items[i].key() == key {
			items[i].Cloned = true
			return
		}
	}
}

// fmtCounts is a tiny helper for status lines.
func fmtCounts(done, total int) string {
	return fmt.Sprintf("%d/%d", done, total)
}
