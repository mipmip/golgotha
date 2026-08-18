package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

// newFacetModel builds a Model already drilled into one owner whose repos vary
// by fork/archived/visibility, so facet cycling is directly observable.
func newFacetModel(t *testing.T, p *config.Provider) *Model {
	t.Helper()
	ti := textinput.New()
	m := &Model{
		cfg:             &config.Config{BaseDir: "/tmp", Providers: []config.Provider{*p}},
		providers:       []*config.Provider{p},
		reposByProvider: map[string][]repoItem{},
		nav:             levelRepos,
		selProvider:     p,
		selOwner:        "acme",
		filter:          ti,
		checkCloned:     func(string) bool { return false },
	}
	m.reposByProvider[p.Name] = []repoItem{
		{Repo: provider.Repo{Owner: "acme", Name: "pub", Visibility: "public"}, Provider: p},
		{Repo: provider.Repo{Owner: "acme", Name: "priv", Visibility: "private"}, Provider: p},
		{Repo: provider.Repo{Owner: "acme", Name: "afork", Visibility: "public", Fork: true}, Provider: p},
		{Repo: provider.Repo{Owner: "acme", Name: "old", Visibility: "public", Archived: true}, Provider: p},
	}
	return m
}

func names(items []repoItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Repo.Name
	}
	return out
}

func TestFacetCyclingNarrows(t *testing.T) {
	inclArch := true
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", IncludeArchived: &inclArch}
	m := newFacetModel(t, p)

	if len(m.visibleRepos()) != 4 {
		t.Fatalf("expected 4 repos initially, got %v", names(m.visibleRepos()))
	}

	// f cycles fork all -> only: keep only the fork.
	send(m, key("f"))
	if got := names(m.visibleRepos()); len(got) != 1 || got[0] != "afork" {
		t.Fatalf("fork:only expected [afork], got %v", got)
	}
	// f again -> hide: drop the fork.
	send(m, key("f"))
	got := names(m.visibleRepos())
	for _, n := range got {
		if n == "afork" {
			t.Fatalf("fork:hide should drop afork, got %v", got)
		}
	}
	if len(got) != 3 {
		t.Fatalf("fork:hide expected 3 repos, got %v", got)
	}
	// f again -> all.
	send(m, key("f"))
	if len(m.visibleRepos()) != 4 {
		t.Fatalf("fork:all expected 4, got %v", names(m.visibleRepos()))
	}

	// v cycles all -> public: pub, afork, old are public.
	send(m, key("v"))
	if got := names(m.visibleRepos()); len(got) != 3 {
		t.Fatalf("vis:public expected 3, got %v", got)
	}
	// v -> private: only priv.
	send(m, key("v"))
	if got := names(m.visibleRepos()); len(got) != 1 || got[0] != "priv" {
		t.Fatalf("vis:private expected [priv], got %v", got)
	}

	// a cycles archived all -> only (combined with vis:private -> nothing here).
	send(m, key("a"))
	if got := m.visibleRepos(); len(got) != 0 {
		t.Fatalf("vis:private + archived:only expected empty, got %v", names(got))
	}
}

func TestFacetComposesWithFuzzy(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newFacetModel(t, p)

	// vis:public -> pub, afork, old.
	send(m, key("v"))
	// Now fuzzy for "pub".
	send(m, key("/"))
	for _, r := range "pub" {
		send(m, key(string(r)))
	}
	got := names(m.visibleRepos())
	if len(got) != 1 || got[0] != "pub" {
		t.Fatalf("vis:public + fuzzy 'pub' expected [pub], got %v", got)
	}
}

func TestFacetChangeResetsWindow(t *testing.T) {
	inclArch := true
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", IncludeArchived: &inclArch}
	m := newFacetModel(t, p)
	m.cursor = 3
	m.offset = 2

	send(m, key("v")) // change filter set
	if m.cursor != 0 || m.offset != 0 {
		t.Fatalf("facet change should reset window, got cursor=%d offset=%d", m.cursor, m.offset)
	}
}

func TestFacetHintWhenDataNotCached(t *testing.T) {
	// include_archived defaults to false -> archived:only can't match cached data.
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newFacetModel(t, p)
	// Remove the archived repo to simulate it never being cached.
	m.reposByProvider[p.Name] = m.reposByProvider[p.Name][:3]

	send(m, key("a")) // archived:only
	if hint := m.facetHint(); hint == "" {
		t.Fatal("expected hint when archived not cached")
	}
	// The body should surface the hint rather than a bare empty list.
	body := m.renderRepos(m.visibleRepos())
	if body == "" || body == dimStyle.Render("(no repositories)") {
		t.Fatalf("expected hinted empty body, got %q", body)
	}
	if !strings.Contains(body, "not cached") {
		t.Fatalf("expected hint text in body, got %q", body)
	}
}

func TestFacetNoHintWhenCached(t *testing.T) {
	inclArch := true
	inclForks := false
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh",
		IncludeArchived: &inclArch, IncludeForks: &inclForks}
	m := newFacetModel(t, p)

	send(m, key("a")) // archived:only — archived IS cached, so no hint.
	if hint := m.facetHint(); hint != "" {
		t.Fatalf("expected no archived hint, got %q", hint)
	}
	// Reset to all, then fork:only — forks NOT cached -> hint.
	send(m, key("a"))
	send(m, key("a"))
	send(m, key("f"))
	if hint := m.facetHint(); hint == "" {
		t.Fatal("expected fork hint when include_forks:false")
	}
}
