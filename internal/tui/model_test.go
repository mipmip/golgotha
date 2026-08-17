package tui

import (
	"context"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/skull2/internal/config"
	"github.com/mipmip/skull2/internal/provider"
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
		{Repo: provider.Repo{Owner: "mipmip", Name: "skull2", WebURL: "https://github.com/mipmip/skull2"}, Provider: pGH, Target: "/tmp/a"},
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
	// move past end
	for i := 0; i < 10; i++ {
		send(m, key("j"))
	}
	if m.cursor != len(m.providers)-1 {
		t.Fatalf("cursor should clamp to last provider, got %d", m.cursor)
	}
}

func TestFilterNarrowsList(t *testing.T) {
	m, _ := newTestModel(t)
	// activate filter
	send(m, key("/"))
	if !m.filtering {
		t.Fatal("expected filtering active")
	}
	// type "skull" -> only mipmip/skull2 across all providers
	for _, r := range "skull" {
		send(m, key(string(r)))
	}
	repos := m.visibleRepos()
	if len(repos) != 1 || repos[0].key() != "mipmip/skull2" {
		t.Fatalf("expected only mipmip/skull2, got %v", repos)
	}
	// apply filter (enter leaves filtering mode but keeps query)
	send(m, key("enter"))
	if m.filtering {
		t.Fatal("expected filtering mode to end on enter")
	}
	if m.filter.Value() != "skull" {
		t.Fatalf("expected filter kept, got %q", m.filter.Value())
	}
	// esc clears the query
	send(m, key("esc"))
	if m.filter.Value() != "" {
		t.Fatalf("expected filter cleared, got %q", m.filter.Value())
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
	send(m, key("enter")) // repos (skull2, dotfiles)
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
