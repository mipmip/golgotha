package tui

import (
	"context"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

// fakeCloner records clone calls and returns a canned result.
type fakeCloner struct {
	calls []string
	err   error
}

func (f *fakeCloner) CloneRepo(_ context.Context, p *config.Provider, r provider.Repo) syncerResult {
	f.calls = append(f.calls, p.Name+":"+r.Owner+"/"+r.Name)
	return syncerResult{Cloned: f.err == nil, Err: f.err}
}

// newTestModel builds a Model directly (no cache/filesystem) with two providers
// and a handful of repos, using a fake cloner and disabled refresher.
func newTestModel(t *testing.T) (*Model, *fakeCloner) {
	t.Helper()
	pGH := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", WebURL: "https://github.com"}
	pCB := &config.Provider{Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", WebURL: "https://codeberg.org"}

	fc := &fakeCloner{}
	ti := textinput.New()
	m := &Model{
		cfg:             &config.Config{BaseDir: "/tmp", Providers: []config.Provider{*pGH, *pCB}},
		providers:       []*config.Provider{pGH, pCB},
		reposByProvider: map[string][]repoItem{},
		nav:             levelProviders,
		filter:          ti,
		cloner:          fc,
		refresher:       nil, // disabled: no network in tests
		checkCloned:     func(string) bool { return false },
	}
	m.reposByProvider["github"] = []repoItem{
		{Repo: provider.Repo{Owner: "mipmip", Name: "huphop", WebURL: "https://github.com/mipmip/huphop"}, Provider: pGH, Target: "/tmp/a"},
		{Repo: provider.Repo{Owner: "mipmip", Name: "dotfiles", WebURL: "https://github.com/mipmip/dotfiles"}, Provider: pGH, Target: "/tmp/b"},
		{Repo: provider.Repo{Owner: "acme", Name: "widgets", WebURL: "https://github.com/acme/widgets"}, Provider: pGH, Target: "/tmp/c"},
	}
	m.reposByProvider["codeberg"] = []repoItem{
		{Repo: provider.Repo{Owner: "pim", Name: "notes", WebURL: "https://codeberg.org/pim/notes"}, Provider: pCB, Target: "/tmp/d"},
	}
	return m, fc
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "ctrl+n":
		return tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func send(m *Model, msg tea.Msg) tea.Cmd {
	_, cmd := m.Update(msg)
	return cmd
}

func TestNavigationDrillDownAndBack(t *testing.T) {
	m, _ := newTestModel(t)

	// providers -> owners (select github)
	send(m, key("enter"))
	if m.nav != levelOwners || m.selProvider == nil || m.selProvider.Name != "github" {
		t.Fatalf("expected owners level under github, got nav=%d prov=%v", m.nav, m.selProvider)
	}

	// owners: acme, mipmip (sorted). Select first -> acme.
	owners := m.ownersFor(m.selProvider)
	if len(owners) != 2 || owners[0] != "acme" {
		t.Fatalf("unexpected owners %v", owners)
	}
	send(m, key("enter"))
	if m.nav != levelRepos || m.selOwner != "acme" {
		t.Fatalf("expected repos level for acme, got nav=%d owner=%q", m.nav, m.selOwner)
	}
	repos := m.visibleRepos()
	if len(repos) != 1 || repos[0].Repo.Name != "widgets" {
		t.Fatalf("expected only acme/widgets, got %v", repos)
	}

	// back to owners
	send(m, key("esc"))
	if m.nav != levelOwners || m.selOwner != "" {
		t.Fatalf("expected back at owners, got nav=%d owner=%q", m.nav, m.selOwner)
	}
	// back to providers
	send(m, key("backspace"))
	if m.nav != levelProviders || m.selProvider != nil {
		t.Fatalf("expected back at providers, got nav=%d prov=%v", m.nav, m.selProvider)
	}
}

func TestCursorMovementClamps(t *testing.T) {
	m, _ := newTestModel(t)
	// up at top stays at 0
	send(m, key("k"))
	if m.cursor != 0 {
		t.Fatalf("cursor should clamp to 0, got %d", m.cursor)
	}
	send(m, key("j"))
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", m.cursor)
	}
	// move past end: clamps to the last row, which is the synthetic
	// "All repositories" entry appended after the providers.
	for i := 0; i < 10; i++ {
		send(m, key("j"))
	}
	if m.cursor != len(m.providers) {
		t.Fatalf("cursor should clamp to the all-repos entry, got %d", m.cursor)
	}
}

func TestFilterNarrowsReposAtReposLevel(t *testing.T) {
	m, _ := newTestModel(t)
	// Drill into github > mipmip (huphop, dotfiles).
	send(m, key("enter")) // owners
	send(m, key("j"))     // mipmip
	send(m, key("enter")) // repos of mipmip
	if m.selOwner != "mipmip" {
		t.Fatalf("expected mipmip owner, got %q", m.selOwner)
	}
	// activate filter
	send(m, key("/"))
	if !m.filtering {
		t.Fatal("expected filtering active")
	}
	// type "huph" -> only mipmip/huphop
	for _, r := range "huph" {
		send(m, key(string(r)))
	}
	repos := m.visibleRepos()
	if len(repos) != 1 || repos[0].key() != "mipmip/huphop" {
		t.Fatalf("expected only mipmip/huphop, got %v", repos)
	}
	// esc while filtering clears the query and exits filter mode
	send(m, key("esc"))
	if m.filtering {
		t.Fatal("expected filtering mode to end on esc")
	}
	if m.filter.Value() != "" {
		t.Fatalf("expected filter cleared on esc, got %q", m.filter.Value())
	}
}

func TestFilterNarrowsProvidersAtProvidersLevel(t *testing.T) {
	m, _ := newTestModel(t)
	send(m, key("/"))
	for _, r := range "code" {
		send(m, key(string(r)))
	}
	got := m.visibleProviders()
	if len(got) != 1 || got[0] != "codeberg" {
		t.Fatalf("expected only codeberg, got %v", got)
	}
}

func TestFilterNarrowsOwnersAtOwnersLevel(t *testing.T) {
	m, _ := newTestModel(t)
	send(m, key("enter")) // into github owners (acme, mipmip)
	send(m, key("/"))
	for _, r := range "mip" {
		send(m, key(string(r)))
	}
	got := m.visibleOwners()
	if len(got) != 1 || got[0] != "mipmip" {
		t.Fatalf("expected only mipmip, got %v", got)
	}
}

func TestFilterEnterDrillsFilteredOwner(t *testing.T) {
	m, _ := newTestModel(t)
	send(m, key("enter")) // into github owners (acme, mipmip)
	send(m, key("/"))
	for _, r := range "mip" {
		send(m, key(string(r)))
	}
	// One Enter acts on the highlighted filtered owner (mipmip) and drills in.
	send(m, key("enter"))
	if m.filtering {
		t.Fatal("expected filtering mode to end on enter")
	}
	if m.nav != levelRepos || m.selOwner != "mipmip" {
		t.Fatalf("expected repos level for mipmip, got nav=%d owner=%q", m.nav, m.selOwner)
	}
	// Drilling in clears the filter.
	if m.filter.Value() != "" {
		t.Fatalf("expected filter cleared on drill-in, got %q", m.filter.Value())
	}
	repos := m.visibleRepos()
	if len(repos) != 2 {
		t.Fatalf("expected mipmip's 2 repos, got %v", repos)
	}
}

func TestFilterMatchesRawOwnerNotDecoration(t *testing.T) {
	m, _ := newLazyModel(t)
	send(m, key("enter")) // into provider owners (acme fetched, beta not fetched)
	// "beta" carries a "(not fetched)" decoration in the rendered label; the
	// filter must match the raw name, and must NOT be satisfiable via the
	// decoration text (e.g. "notfetched").
	send(m, key("/"))
	for _, r := range "beta" {
		send(m, key(string(r)))
	}
	got := m.visibleOwners()
	if len(got) != 1 || got[0] != "beta" {
		t.Fatalf("expected only beta, got %v", got)
	}
	// A query that only exists in the decoration must not match anything.
	m.filter.SetValue("notfetched")
	if got := m.visibleOwners(); len(got) != 0 {
		t.Fatalf("expected no owners for decoration-only query, got %v", got)
	}
}

func TestFilterClearsOnBack(t *testing.T) {
	m, _ := newTestModel(t)
	send(m, key("enter")) // into github owners
	send(m, key("/"))
	for _, r := range "mip" {
		send(m, key(string(r)))
	}
	if m.filter.Value() == "" {
		t.Fatal("expected filter query while typing")
	}
	// Esc while filtering clears the query and exits filter mode (stays at owners).
	send(m, key("esc"))
	if m.filtering || m.filter.Value() != "" {
		t.Fatalf("expected filter cleared and mode ended, filtering=%v q=%q", m.filtering, m.filter.Value())
	}
	if m.nav != levelOwners {
		t.Fatalf("expected still at owners level, got nav=%d", m.nav)
	}
	// Back again pops to providers, unfiltered.
	send(m, key("esc"))
	if m.nav != levelProviders {
		t.Fatalf("expected providers level, got nav=%d", m.nav)
	}
}

func TestSpaceTogglesSelection(t *testing.T) {
	m, _ := newTestModel(t)
	// Drill into github/mipmip.
	send(m, key("enter")) // owners
	// owners sorted: acme, mipmip -> move to mipmip
	send(m, key("j"))
	send(m, key("enter")) // repos of mipmip
	if m.selOwner != "mipmip" {
		t.Fatalf("expected mipmip owner, got %q", m.selOwner)
	}
	// toggle selection on first repo
	send(m, key("space"))
	sel := m.selectedItems()
	if len(sel) != 1 {
		t.Fatalf("expected 1 selected, got %d", len(sel))
	}
	// toggle off
	send(m, key("space"))
	if len(m.selectedItems()) != 0 {
		t.Fatalf("expected 0 selected after toggle off")
	}
}

func TestSingleCloneEmitsCommandAndMarksCloned(t *testing.T) {
	m, fc := newTestModel(t)
	send(m, key("enter")) // owners
	send(m, key("j"))     // mipmip
	send(m, key("enter")) // repos
	cmd := send(m, key("c"))
	if cmd == nil {
		t.Fatal("expected a clone command")
	}
	msg := cmd()
	cr, ok := msg.(cloneResultMsg)
	if !ok {
		t.Fatalf("expected cloneResultMsg, got %T", msg)
	}
	if len(fc.calls) != 1 {
		t.Fatalf("expected 1 clone call, got %v", fc.calls)
	}
	// feed the result back into Update
	send(m, cr)
	// find the item; it should now be cloned
	found := false
	for _, it := range m.reposByProvider[cr.Provider] {
		if it.key() == cr.Key {
			found = true
			if !it.Cloned {
				t.Fatalf("expected %s marked cloned", cr.Key)
			}
		}
	}
	if !found {
		t.Fatalf("cloned repo %s not found", cr.Key)
	}
}

func TestBulkCloneMarksAllAndClearsSelection(t *testing.T) {
	m, fc := newTestModel(t)
	send(m, key("enter")) // owners
	send(m, key("j"))     // mipmip
	send(m, key("enter")) // repos (huphop, dotfiles)
	send(m, key("space")) // select first
	send(m, key("j"))
	send(m, key("space")) // select second
	if len(m.selectedItems()) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(m.selectedItems()))
	}
	cmd := send(m, key("c"))
	if cmd == nil {
		t.Fatal("expected bulk clone command")
	}
	msg := cmd()
	bd, ok := msg.(bulkDoneMsg)
	if !ok {
		t.Fatalf("expected bulkDoneMsg, got %T", msg)
	}
	if len(fc.calls) != 2 {
		t.Fatalf("expected 2 clone calls, got %v", fc.calls)
	}
	send(m, bd)
	if len(m.selectedItems()) != 0 {
		t.Fatalf("expected selection cleared after bulk clone")
	}
	cloned := 0
	for _, it := range m.reposByProvider["github"] {
		if it.Cloned {
			cloned++
		}
	}
	if cloned != 2 {
		t.Fatalf("expected 2 cloned, got %d", cloned)
	}
}

func TestAlreadyClonedNotReCloned(t *testing.T) {
	m, fc := newTestModel(t)
	// mark all github repos cloned
	items := m.reposByProvider["github"]
	for i := range items {
		items[i].Cloned = true
	}
	send(m, key("enter")) // owners
	send(m, key("enter")) // repos of acme (widgets)
	cmd := send(m, key("c"))
	if cmd != nil {
		t.Fatal("expected no clone command for already-cloned repo")
	}
	if len(fc.calls) != 0 {
		t.Fatalf("expected no clone calls, got %v", fc.calls)
	}
}

func TestOpenCallsStubbedOpener(t *testing.T) {
	m, _ := newTestModel(t)

	var opened string
	orig := openURL
	openURL = func(_ context.Context, url string) error {
		opened = url
		return nil
	}
	defer func() { openURL = orig }()

	send(m, key("enter")) // owners (github)
	send(m, key("enter")) // repos of acme
	cmd := send(m, key("o"))
	if cmd == nil {
		t.Fatal("expected an open command")
	}
	msg := cmd()
	if _, ok := msg.(openResultMsg); !ok {
		t.Fatalf("expected openResultMsg, got %T", msg)
	}
	if opened != "https://github.com/acme/widgets" {
		t.Fatalf("expected acme/widgets web URL opened, got %q", opened)
	}
	// feed result back for status coverage
	send(m, msg)
}

func TestRefreshDisabledIsNoop(t *testing.T) {
	m, _ := newTestModel(t)
	// refresher is nil in the test model; pressing r must not panic or emit a cmd
	cmd := send(m, key("r"))
	if cmd != nil {
		t.Fatal("expected no refresh command when refresher disabled")
	}
}

func TestRefreshResultReloadsProvider(t *testing.T) {
	m, _ := newTestModel(t)
	// Simulate a refresh result message directly (no network).
	cmd := send(m, refreshResultMsg{Provider: "github", Count: 3})
	if cmd != nil {
		t.Fatal("expected no follow-up command")
	}
	if m.status == "" {
		t.Fatal("expected status update after refresh")
	}
}

func TestQuit(t *testing.T) {
	m, _ := newTestModel(t)
	cmd := send(m, key("q"))
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if !m.quitting {
		t.Fatal("expected quitting flag set")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m, _ := newTestModel(t)
	send(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	if m.width != 80 || m.height != 24 {
		t.Fatalf("expected size stored, got %dx%d", m.width, m.height)
	}
}

// fakeFetcher records lazy owner-fetch calls and returns a canned message.
type fakeFetcher struct {
	calls []string
	repos []provider.Repo
	err   error
}

func (f *fakeFetcher) cmd(_ context.Context, p *config.Provider, owner string) tea.Cmd {
	f.calls = append(f.calls, p.Name+":"+owner)
	repos := f.repos
	err := f.err
	return func() tea.Msg {
		return ownerFetchedMsg{Provider: p.Name, Owner: owner, Repos: repos, Err: err}
	}
}

// newLazyModel builds a Model with a single provider whose owner index has one
// fetched owner (acme) and one unfetched owner (beta), plus an injectable fake
// fetcher.
func newLazyModel(t *testing.T) (*Model, *fakeFetcher) {
	t.Helper()
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", WebURL: "https://github.com"}
	ff := &fakeFetcher{repos: []provider.Repo{{Owner: "beta", Name: "b1", WebURL: "https://github.com/beta/b1"}}}
	ti := textinput.New()
	m := &Model{
		cfg:              &config.Config{BaseDir: "/tmp", ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}", Providers: []config.Provider{*p}},
		providers:        []*config.Provider{p},
		reposByProvider:  map[string][]repoItem{},
		ownersByProvider: map[string][]string{"github": {"acme", "beta"}},
		fetchedOwners:    map[string]map[string]bool{"github": {"acme": true, "beta": false}},
		nav:              levelProviders,
		filter:           ti,
		cloner:           &fakeCloner{},
		refresher:        nil,
		ownerFetcher:     ff.cmd,
		checkCloned:      func(string) bool { return false },
	}
	m.reposByProvider["github"] = []repoItem{
		{Repo: provider.Repo{Owner: "acme", Name: "widgets", WebURL: "https://github.com/acme/widgets"}, Provider: p, Target: "/tmp/a"},
	}
	return m, ff
}

func TestOwnerLevelListsDiscoveredUnfetched(t *testing.T) {
	m, _ := newLazyModel(t)
	send(m, key("enter")) // into owners
	owners := m.ownersFor(m.selProvider)
	if len(owners) != 2 || owners[0] != "acme" || owners[1] != "beta" {
		t.Fatalf("owner index not listed incl. unfetched: %v", owners)
	}
}

func TestEnterUnfetchedOwnerEmitsFetchAndPopulates(t *testing.T) {
	m, ff := newLazyModel(t)
	send(m, key("enter")) // into owners (cursor on acme)
	send(m, key("j"))     // move to beta
	cmd := send(m, key("enter"))
	if cmd == nil {
		t.Fatal("expected a fetch command for unfetched owner")
	}
	if m.selOwner != "beta" {
		t.Fatalf("selOwner = %q, want beta", m.selOwner)
	}
	if m.fetchingOwner != "beta" {
		t.Fatalf("fetchingOwner = %q, want beta", m.fetchingOwner)
	}
	if len(ff.calls) != 1 || ff.calls[0] != "github:beta" {
		t.Fatalf("fetch calls = %v", ff.calls)
	}
	// Run the command and feed the message back.
	msg := cmd()
	if _, ok := msg.(ownerFetchedMsg); !ok {
		t.Fatalf("expected ownerFetchedMsg, got %T", msg)
	}
	send(m, msg)
	if m.fetchingOwner != "" {
		t.Fatal("fetchingOwner should be cleared after result")
	}
	// beta now cached and populated.
	if !m.fetchedOwners["github"]["beta"] {
		t.Fatal("beta should be marked fetched after result")
	}
}

func TestEnterFetchedOwnerDoesNotFetch(t *testing.T) {
	m, ff := newLazyModel(t)
	send(m, key("enter")) // into owners (cursor on acme, which is fetched)
	cmd := send(m, key("enter"))
	if cmd != nil {
		t.Fatal("expected no fetch command for already-fetched owner")
	}
	if len(ff.calls) != 0 {
		t.Fatalf("expected no fetch calls, got %v", ff.calls)
	}
	if m.selOwner != "acme" {
		t.Fatalf("selOwner = %q, want acme", m.selOwner)
	}
}

func TestRefreshReFetchesCurrentOwner(t *testing.T) {
	m, ff := newLazyModel(t)
	send(m, key("enter")) // owners
	send(m, key("enter")) // into acme (fetched, no fetch)
	if m.nav != levelRepos || m.selOwner != "acme" {
		t.Fatalf("expected repos of acme, got nav=%d owner=%q", m.nav, m.selOwner)
	}
	// `r` re-fetches acme even though it is already fetched.
	cmd := send(m, key("r"))
	if cmd == nil {
		t.Fatal("expected re-fetch command from r at repos level")
	}
	if len(ff.calls) != 1 || ff.calls[0] != "github:acme" {
		t.Fatalf("re-fetch calls = %v", ff.calls)
	}
}
