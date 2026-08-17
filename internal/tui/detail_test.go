package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/skull2/internal/cache"
	"github.com/mipmip/skull2/internal/config"
	"github.com/mipmip/skull2/internal/provider"
)

// newDetailTestModel builds a model at the repos level with one repo and an
// injectable detail fetcher. XDG_CACHE_HOME is redirected to a temp dir so the
// detail cache never touches the real filesystem.
func newDetailTestModel(t *testing.T) (*Model, *config.Provider) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", WebURL: "https://github.com"}
	m, _ := newTestModel(t)
	m.spinner = spinner.New()
	m.readme = viewport.New(0, 0)
	m.detailFetcher = nil // default: no network; tests inject the msg
	m.width = 80
	m.height = 40

	// Navigate to github/mipmip repos (acme, mipmip sorted; move to mipmip).
	send(m, key("enter")) // owners
	send(m, key("j"))     // mipmip
	send(m, key("enter")) // repos
	if m.nav != levelRepos {
		t.Fatalf("expected repos level, got %d", m.nav)
	}
	_ = p
	return m, m.selProvider
}

func TestEnterOpensDetailView(t *testing.T) {
	m, _ := newDetailTestModel(t)

	// Wire a fetcher that records the call and returns a canned msg.
	var fetched string
	m.detailFetcher = func(_ context.Context, p *config.Provider, r provider.Repo) tea.Cmd {
		fetched = r.Owner + "/" + r.Name
		return func() tea.Msg {
			return detailLoadedMsg{
				Provider: p.Name, Owner: r.Owner, Name: r.Name,
				Details: cache.Details{Stars: 3, Language: "Go", Topics: []string{"cli"}, ReadmeMarkdown: "# Hi\n"},
			}
		}
	}

	cmd := send(m, key("enter"))
	if m.nav != levelDetail {
		t.Fatalf("enter should open detail view, nav=%d", m.nav)
	}
	if !m.detailLoading {
		t.Fatal("expected detailLoading on first open with no cache")
	}
	if cmd == nil {
		t.Fatal("expected a lazy-fetch command")
	}
	if fetched == "" {
		t.Fatal("fetcher was not invoked")
	}

	// Loading indicator visible before the fetch completes.
	if !strings.Contains(m.View(), "loading") {
		t.Fatalf("expected loading indicator, view=\n%s", m.View())
	}

	// Deliver the fetched detail.
	send(m, cmd())
	if m.detailLoading {
		t.Fatal("detailLoading should clear once loaded")
	}
	if !m.detailLoaded || m.detail.Stars != 3 {
		t.Fatalf("detail not populated: loaded=%v detail=%+v", m.detailLoaded, m.detail)
	}
	view := m.View()
	if !strings.Contains(view, "Go") || !strings.Contains(view, "cli") {
		t.Fatalf("expected metadata in view, got:\n%s", view)
	}
}

func TestDetailFetchCachesAndReuses(t *testing.T) {
	m, p := newDetailTestModel(t)
	it, _ := m.currentRepo()

	// Pre-populate the detail cache; opening should reuse it (no fetch command).
	_, err := cache.RefreshDetails(p.Name, it.Repo.Owner, it.Repo.Name,
		provider.Details{Stars: 99, Language: "Rust"}, "# Cached\n", nowUTC())
	if err != nil {
		t.Fatal(err)
	}
	fetchCalled := false
	m.detailFetcher = func(_ context.Context, _ *config.Provider, _ provider.Repo) tea.Cmd {
		fetchCalled = true
		return func() tea.Msg { return nil }
	}

	cmd := send(m, key("enter"))
	if m.nav != levelDetail {
		t.Fatalf("expected detail view, nav=%d", m.nav)
	}
	if cmd != nil {
		t.Fatal("cached detail should not trigger a fetch command")
	}
	if fetchCalled {
		t.Fatal("fetcher must not be called when cache exists")
	}
	if !m.detailLoaded || m.detail.Stars != 99 {
		t.Fatalf("expected cached detail reused: %+v", m.detail)
	}
}

func TestDetailGracefulOffline(t *testing.T) {
	m, _ := newDetailTestModel(t)
	m.detailFetcher = func(_ context.Context, p *config.Provider, r provider.Repo) tea.Cmd {
		return func() tea.Msg {
			return detailLoadedMsg{Provider: p.Name, Owner: r.Owner, Name: r.Name, Err: context.DeadlineExceeded}
		}
	}

	cmd := send(m, key("enter"))
	send(m, cmd())

	if m.detailLoaded {
		t.Fatal("should not be loaded on fetch failure with no cache")
	}
	if !m.detailUnavailable {
		t.Fatal("expected detailUnavailable on offline fetch failure")
	}
	view := m.View()
	if !strings.Contains(view, "README unavailable") {
		t.Fatalf("expected 'README unavailable' note, got:\n%s", view)
	}
	// Metadata is still shown (visibility line), no error screen.
	if strings.Contains(strings.ToLower(view), "error") {
		t.Fatalf("offline should not show an error screen:\n%s", view)
	}
}

func TestDetailRefreshReFetches(t *testing.T) {
	m, p := newDetailTestModel(t)
	it, _ := m.currentRepo()
	// Cache exists so open is instant.
	cache.RefreshDetails(p.Name, it.Repo.Owner, it.Repo.Name, provider.Details{Stars: 1}, "old", nowUTC())

	calls := 0
	m.detailFetcher = func(_ context.Context, pp *config.Provider, r provider.Repo) tea.Cmd {
		calls++
		return func() tea.Msg {
			return detailLoadedMsg{Provider: pp.Name, Owner: r.Owner, Name: r.Name, Details: cache.Details{Stars: 2, ReadmeMarkdown: "new"}}
		}
	}

	send(m, key("enter")) // open (cached, no fetch)
	if calls != 0 {
		t.Fatalf("cached open should not fetch, calls=%d", calls)
	}
	cmd := send(m, key("r")) // re-fetch
	if cmd == nil || calls != 1 {
		t.Fatalf("r should re-fetch: cmd=%v calls=%d", cmd, calls)
	}
	if !m.detailLoading {
		t.Fatal("expected loading during re-fetch")
	}
	send(m, cmd())
	if m.detail.Stars != 2 {
		t.Fatalf("re-fetch did not update detail: %+v", m.detail)
	}
}

func TestDetailEscReturnsToRepoList(t *testing.T) {
	m, _ := newDetailTestModel(t)
	m.detailFetcher = func(_ context.Context, p *config.Provider, r provider.Repo) tea.Cmd {
		return func() tea.Msg { return detailLoadedMsg{Provider: p.Name, Owner: r.Owner, Name: r.Name} }
	}
	// Move cursor to second repo so we can assert position restore.
	send(m, key("j"))
	priorCursor := m.cursor

	send(m, key("enter")) // open detail
	if m.nav != levelDetail {
		t.Fatalf("expected detail, nav=%d", m.nav)
	}
	send(m, key("esc"))
	if m.nav != levelRepos {
		t.Fatalf("esc should return to repos, nav=%d", m.nav)
	}
	if m.cursor != priorCursor {
		t.Fatalf("esc should restore cursor %d, got %d", priorCursor, m.cursor)
	}
}

func TestDetailCloneUsesC(t *testing.T) {
	m, _ := newDetailTestModel(t)
	m.detailFetcher = func(_ context.Context, p *config.Provider, r provider.Repo) tea.Cmd {
		return func() tea.Msg { return detailLoadedMsg{Provider: p.Name, Owner: r.Owner, Name: r.Name} }
	}
	send(m, key("enter")) // open detail
	cmd := send(m, key("c"))
	if cmd == nil {
		t.Fatal("c should clone from the detail view")
	}
	msg := cmd()
	if _, ok := msg.(cloneResultMsg); !ok {
		t.Fatalf("expected cloneResultMsg, got %T", msg)
	}
}
