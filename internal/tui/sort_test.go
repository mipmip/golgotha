package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

// newSortModel builds a Model drilled into one owner whose repos have distinct
// names and update times (deliberately not in sorted fetch order) so that both
// the sort ordering and the fetch-order default are observable.
func newSortModel(t *testing.T, p *config.Provider) *Model {
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
	d := func(s string) time.Time {
		tm, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatalf("bad date %q: %v", s, err)
		}
		return tm
	}
	// Fetch order: Charlie, alpha, Bravo, delta (intentionally unsorted, mixed
	// case to exercise case-insensitive name compare).
	m.reposByProvider[p.Name] = []repoItem{
		{Repo: provider.Repo{Owner: "acme", Name: "Charlie", UpdatedAt: d("2024-03-01")}, Provider: p},
		{Repo: provider.Repo{Owner: "acme", Name: "alpha", UpdatedAt: d("2024-01-01")}, Provider: p},
		{Repo: provider.Repo{Owner: "acme", Name: "Bravo", UpdatedAt: d("2024-04-01")}, Provider: p},
		{Repo: provider.Repo{Owner: "acme", Name: "delta", UpdatedAt: d("2024-02-01")}, Provider: p},
	}
	return m
}

func TestSortNameAscDesc(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newSortModel(t, p)

	m.sortKey = sortName
	m.sortDir = sortAsc
	if got := names(m.visibleRepos()); !equal(got, []string{"alpha", "Bravo", "Charlie", "delta"}) {
		t.Fatalf("name asc: got %v", got)
	}

	m.sortDir = sortDesc
	if got := names(m.visibleRepos()); !equal(got, []string{"delta", "Charlie", "Bravo", "alpha"}) {
		t.Fatalf("name desc: got %v", got)
	}
}

func TestSortUpdatedAscDesc(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newSortModel(t, p)

	m.sortKey = sortUpdated
	m.sortDir = sortAsc
	// oldest -> newest: alpha(01) delta(02) Charlie(03) Bravo(04)
	if got := names(m.visibleRepos()); !equal(got, []string{"alpha", "delta", "Charlie", "Bravo"}) {
		t.Fatalf("updated asc: got %v", got)
	}

	m.sortDir = sortDesc
	// newest -> oldest
	if got := names(m.visibleRepos()); !equal(got, []string{"Bravo", "Charlie", "delta", "alpha"}) {
		t.Fatalf("updated desc: got %v", got)
	}
}

func TestSortNoneIsFetchOrder(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newSortModel(t, p)

	// Default (zero value) is fetch order.
	if got := names(m.visibleRepos()); !equal(got, []string{"Charlie", "alpha", "Bravo", "delta"}) {
		t.Fatalf("default should be fetch order, got %v", got)
	}

	// Sorting and then returning to none must restore the original fetch order
	// (the underlying cache slice must not have been mutated).
	m.sortKey = sortName
	_ = m.visibleRepos()
	m.sortKey = sortNone
	if got := names(m.visibleRepos()); !equal(got, []string{"Charlie", "alpha", "Bravo", "delta"}) {
		t.Fatalf("none after sort should restore fetch order, got %v", got)
	}
}

func TestSortComposesAfterFilter(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	m := newSortModel(t, p)

	// Fuzzy for "a" matches Charlie, alpha, Bravo, delta all contain 'a' — pick
	// a narrower query. "l" matches Charlie, alpha, delta (not Bravo).
	m.filter.SetValue("l")
	m.sortKey = sortName
	m.sortDir = sortAsc
	got := names(m.visibleRepos())
	if !equal(got, []string{"alpha", "Charlie", "delta"}) {
		t.Fatalf("sort should order only the filtered subset, got %v", got)
	}
}

func TestSortKeyCycleAndDirToggleUnit(t *testing.T) {
	// Pure cycle-wraparound on the sortKey type.
	if got := sortNone.cycle(); got != sortName {
		t.Fatalf("none -> name, got %v", got)
	}
	if got := sortName.cycle(); got != sortUpdated {
		t.Fatalf("name -> updated, got %v", got)
	}
	if got := sortUpdated.cycle(); got != sortNone {
		t.Fatalf("updated -> none, got %v", got)
	}

	if got := sortAsc.toggle(); got != sortDesc {
		t.Fatalf("asc -> desc, got %v", got)
	}
	if got := sortDesc.toggle(); got != sortAsc {
		t.Fatalf("desc -> asc, got %v", got)
	}
}

// equal is a small slice-equality helper for ordered string comparisons.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
