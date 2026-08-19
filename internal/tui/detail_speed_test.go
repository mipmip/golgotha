package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/cache"
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

// fakeDetailProvider is a provider.Provider with controllable RepoDetails/Readme.
type fakeDetailProvider struct {
	details      provider.Details
	detailsErr   error
	readme       string
	readmeErr    error
	detailsCalls int
	readmeCalls  int
}

func (f *fakeDetailProvider) ListRepos(context.Context, []string) ([]provider.Repo, error) {
	return nil, nil
}
func (f *fakeDetailProvider) ListOwners(context.Context) ([]string, error) { return nil, nil }
func (f *fakeDetailProvider) RepoDetails(context.Context, string, string) (provider.Details, error) {
	f.detailsCalls++
	return f.details, f.detailsErr
}
func (f *fakeDetailProvider) Readme(context.Context, string, string) (string, error) {
	f.readmeCalls++
	return f.readme, f.readmeErr
}

func TestDetailFetcherConcurrentAndReadmeFailNonFatal(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	fp := &fakeDetailProvider{
		details:   provider.Details{Stars: 7, Language: "Go"},
		readme:    "",
		readmeErr: errors.New("no readme"),
	}
	build := func(*config.Provider) (provider.Provider, error) { return fp, nil }
	p := &config.Provider{Name: "gh", Type: "github"}
	r := provider.Repo{Owner: "me", Name: "proj"}

	msg := detailFetcherWith(build)(context.Background(), p, r)().(detailLoadedMsg)

	if fp.detailsCalls != 1 || fp.readmeCalls != 1 {
		t.Fatalf("both requests should be issued: details=%d readme=%d", fp.detailsCalls, fp.readmeCalls)
	}
	if msg.Err != nil {
		t.Fatalf("README failure must be non-fatal, got err %v", msg.Err)
	}
	if msg.Details.Stars != 7 || msg.Details.Language != "Go" {
		t.Fatalf("details not carried through: %+v", msg.Details)
	}
	if msg.Details.ReadmeMarkdown != "" {
		t.Fatalf("expected empty README on readme failure, got %q", msg.Details.ReadmeMarkdown)
	}
}

func TestDetailFetcherDetailsFailFallsBack(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	fp := &fakeDetailProvider{detailsErr: errors.New("boom")}
	build := func(*config.Provider) (provider.Provider, error) { return fp, nil }
	p := &config.Provider{Name: "gh", Type: "github"}
	msg := detailFetcherWith(build)(context.Background(), p, provider.Repo{Owner: "me", Name: "proj"})().(detailLoadedMsg)
	if msg.Err == nil {
		t.Fatal("details failure with no cache should carry the error")
	}
}

func detailModel(t *testing.T) (*Model, *config.Provider, provider.Repo, *int) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	p := &config.Provider{Name: "gh", Type: config.ProviderGitHub, Short: "gh", Username: "me", WebURL: "https://github.com"}
	r := provider.Repo{Owner: "me", Name: "proj", WebURL: "https://github.com/me/proj"}
	calls := 0
	m := &Model{
		cfg:             &config.Config{BaseDir: "/tmp", ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}"},
		providers:       []*config.Provider{p},
		reposByProvider: map[string][]repoItem{"gh": {{Repo: r, Provider: p}}},
		selProvider:     p,
		selOwner:        "me",
		nav:             levelRepos,
	}
	m.filter = textinput.New()
	m.detailFetcher = func(context.Context, *config.Provider, provider.Repo) tea.Cmd {
		calls++
		return nil
	}
	return m, p, r, &calls
}

func TestOpenDetailCacheHitNoFetch(t *testing.T) {
	m, p, r, calls := detailModel(t)
	// Seed the detail cache.
	if err := cache.SaveDetails(p.Name, r.Owner, r.Name, cache.Details{Stars: 3, ReadmeMarkdown: "# hi"}); err != nil {
		t.Fatal(err)
	}
	m.openDetail()
	if !m.detailLoaded {
		t.Fatal("expected details loaded from cache")
	}
	if *calls != 0 {
		t.Fatalf("cache hit must not call the fetcher, got %d calls", *calls)
	}
	if m.detail.Stars != 3 {
		t.Fatalf("wrong cached details: %+v", m.detail)
	}
}

func TestPrefetchWarmsUncachedAndSkipsCached(t *testing.T) {
	m, p, r, calls := detailModel(t)

	// Schedule bumps the sequence at repos level.
	if cmd := m.schedulePrefetch(); cmd == nil {
		t.Fatal("expected a debounced prefetch command at repos level")
	}
	seq := m.prefetchSeq

	// A stale tick does nothing.
	m.handlePrefetchTick(prefetchTickMsg{seq: seq - 1})
	if *calls != 0 {
		t.Fatalf("stale tick should not prefetch, got %d", *calls)
	}
	// A current tick on an uncached repo prefetches.
	m.handlePrefetchTick(prefetchTickMsg{seq: seq})
	if *calls != 1 {
		t.Fatalf("settled tick should prefetch once, got %d", *calls)
	}
	// Now cached: a fresh settle does not prefetch again.
	if err := cache.SaveDetails(p.Name, r.Owner, r.Name, cache.Details{Stars: 1}); err != nil {
		t.Fatal(err)
	}
	m.schedulePrefetch()
	m.handlePrefetchTick(prefetchTickMsg{seq: m.prefetchSeq})
	if *calls != 1 {
		t.Fatalf("cached repo must not be prefetched again, got %d", *calls)
	}
}

func TestSchedulePrefetchNoopOffReposLevel(t *testing.T) {
	m, _, _, _ := detailModel(t)
	m.nav = levelProviders
	if cmd := m.schedulePrefetch(); cmd != nil {
		t.Fatal("schedulePrefetch should be a no-op off the repos level")
	}
}
