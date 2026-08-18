package tui

import (
	"context"
	"testing"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/fetch"
)

// newProgressModel builds a lazy model wired to a streaming progressFetcher whose
// channel and cancel are captured so the test can drive events and assert cancel.
func newProgressModel(t *testing.T) (*Model, chan fetch.Event, *bool) {
	t.Helper()
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", WebURL: "https://github.com"}
	ch := make(chan fetch.Event, 16)
	canceled := false

	ti := textinput.New()
	m := &Model{
		cfg:              &config.Config{BaseDir: "/tmp", ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}", Providers: []config.Provider{*p}},
		providers:        []*config.Provider{p},
		reposByProvider:  map[string][]repoItem{"github": nil},
		ownersByProvider: map[string][]string{"github": {"acme", "beta"}},
		fetchedOwners:    map[string]map[string]bool{"github": {"acme": true, "beta": false}},
		nav:              levelProviders,
		filter:           ti,
		cloner:           &fakeCloner{},
		spinner:          spinner.New(),
		progress:         progress.New(),
		checkCloned:      func(string) bool { return false },
	}
	m.progressFetcher = func(_ context.Context, _ *config.Provider, _ string) (<-chan fetch.Event, context.CancelFunc) {
		return ch, func() { canceled = true }
	}
	return m, ch, &canceled
}

// drillToBeta navigates providers -> owners -> beta (unfetched), starting a
// streaming fetch.
func drillToBeta(t *testing.T, m *Model) tea.Cmd {
	t.Helper()
	send(m, key("enter")) // into owners (cursor acme)
	send(m, key("j"))     // move to beta
	cmd := send(m, key("enter"))
	if cmd == nil {
		t.Fatal("expected a fetch command starting the stream")
	}
	if m.fetchingOwner != "beta" {
		t.Fatalf("fetchingOwner = %q, want beta", m.fetchingOwner)
	}
	return cmd
}

func TestProgressAdvancesSpinnerThenBar(t *testing.T) {
	m, _, _ := newProgressModel(t)
	drillToBeta(t, m)

	// Before any page, total is unknown -> spinner path.
	if m.fetchTotal != 0 {
		t.Fatalf("fetchTotal should start unknown, got %d", m.fetchTotal)
	}

	// Started event.
	send(m, progressMsg{Provider: "github", Owner: "beta", Event: fetch.Event{Kind: fetch.KindStarted, Owner: "beta"}})

	// Page 1 of 3: total becomes known -> determinate.
	send(m, progressMsg{Provider: "github", Owner: "beta", Event: fetch.Event{
		Kind: fetch.KindPageDone, Owner: "beta", Page: 1, TotalPages: 3, ReposSoFar: 10,
	}})
	if m.fetchTotal != 3 {
		t.Fatalf("fetchTotal = %d, want 3", m.fetchTotal)
	}
	if m.fetchPage != 1 || m.fetchRepos != 10 {
		t.Fatalf("page/repos = %d/%d, want 1/10", m.fetchPage, m.fetchRepos)
	}

	// Page 2 advances.
	send(m, progressMsg{Provider: "github", Owner: "beta", Event: fetch.Event{
		Kind: fetch.KindPageDone, Owner: "beta", Page: 2, TotalPages: 3, ReposSoFar: 20,
	}})
	if m.fetchPage != 2 || m.fetchRepos != 20 {
		t.Fatalf("page/repos = %d/%d, want 2/20", m.fetchPage, m.fetchRepos)
	}

	// The determinate view renders without panicking.
	m.nav = levelRepos
	m.selOwner = "beta"
	_ = m.fetchProgressView()
}

func TestProgressDoneReloadsAndClears(t *testing.T) {
	m, _, _ := newProgressModel(t)
	drillToBeta(t, m)

	// Done clears fetching state and applies the (empty cache) result.
	send(m, progressMsg{Provider: "github", Owner: "beta", Event: fetch.Event{
		Kind: fetch.KindDone, Owner: "beta", Provider: "github", Count: 0,
	}})
	if m.fetchingOwner != "" {
		t.Fatalf("fetchingOwner should clear on Done, got %q", m.fetchingOwner)
	}
	if !m.fetchedOwners["github"]["beta"] {
		t.Fatal("beta should be marked fetched after Done")
	}
}

func TestProgressEscCancelsAndBacksOut(t *testing.T) {
	m, _, canceled := newProgressModel(t)
	drillToBeta(t, m)
	if m.nav != levelRepos {
		t.Fatalf("nav = %d, want levelRepos", m.nav)
	}

	// Feed a partial page (transient) before cancel.
	send(m, progressMsg{Provider: "github", Owner: "beta", Event: fetch.Event{
		Kind: fetch.KindPageDone, Owner: "beta", Page: 1, TotalPages: 5, ReposSoFar: 3,
	}})

	// Esc cancels the fetch and backs out to owners.
	send(m, key("esc"))
	if !*canceled {
		t.Fatal("Esc should cancel the in-flight fetch")
	}
	if m.fetchingOwner != "" {
		t.Fatalf("fetchingOwner should clear on cancel, got %q", m.fetchingOwner)
	}
	if m.nav != levelOwners {
		t.Fatalf("nav = %d, want levelOwners after back out", m.nav)
	}
	// Partial results were NOT committed: beta stays unfetched.
	if m.fetchedOwners["github"]["beta"] {
		t.Fatal("cancelled fetch must not mark beta fetched (no partial cache)")
	}
}

func TestProgressCanceledEventClears(t *testing.T) {
	m, _, _ := newProgressModel(t)
	drillToBeta(t, m)
	send(m, progressMsg{Provider: "github", Owner: "beta", Event: fetch.Event{
		Kind: fetch.KindCanceled, Owner: "beta",
	}})
	if m.fetchingOwner != "" {
		t.Fatal("Canceled event should clear fetchingOwner")
	}
	if m.fetchedOwners["github"]["beta"] {
		t.Fatal("Canceled must not mark beta fetched")
	}
}

func TestProgressFailedClearsAndReportsTransient(t *testing.T) {
	m, _, _ := newProgressModel(t)
	drillToBeta(t, m)
	send(m, progressMsg{Provider: "github", Owner: "beta", Event: fetch.Event{
		Kind: fetch.KindFailed, Owner: "beta", Provider: "github", Err: context.DeadlineExceeded,
	}})
	if m.fetchingOwner != "" {
		t.Fatal("Failed should clear fetchingOwner")
	}
	if m.fetchedOwners["github"]["beta"] {
		t.Fatal("Failed must not mark beta fetched (partial is transient)")
	}
}

func TestProgressStaleEventIgnored(t *testing.T) {
	m, _, _ := newProgressModel(t)
	drillToBeta(t, m)
	// An event for a different owner than the one currently fetching is ignored.
	before := m.fetchPage
	send(m, progressMsg{Provider: "github", Owner: "someoneelse", Event: fetch.Event{
		Kind: fetch.KindPageDone, Owner: "someoneelse", Page: 9, TotalPages: 9, ReposSoFar: 99,
	}})
	if m.fetchPage != before {
		t.Fatal("stale event for another owner should be ignored")
	}
}

func TestProgressChannelClosedClears(t *testing.T) {
	m, _, _ := newProgressModel(t)
	drillToBeta(t, m)
	send(m, progressMsg{Provider: "github", Owner: "beta", closed: true})
	if m.fetchingOwner != "" {
		t.Fatal("closed channel should clear fetchingOwner")
	}
}

func TestSpinnerTickIgnoredWhenIdle(t *testing.T) {
	m, _, _ := newProgressModel(t)
	// No fetch in flight: spinner tick is a no-op (no command).
	cmd := send(m, spinner.TickMsg{})
	if cmd != nil {
		t.Fatal("spinner tick while idle should not re-issue")
	}
}
