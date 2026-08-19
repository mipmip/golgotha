package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

func combinedModel() *Model {
	gh := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", Username: "me"}
	cb := &config.Provider{Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", Username: "me"}
	m := &Model{
		cfg:              &config.Config{},
		providers:        []*config.Provider{gh, cb},
		reposByProvider:  map[string][]repoItem{},
		ownersByProvider: map[string][]string{"github": {"me", "acme"}, "codeberg": {"me"}},
		fetchedOwners:    map[string]map[string]bool{"github": {"me": true, "acme": false}, "codeberg": {"me": true}},
		nav:              levelProviders,
	}
	m.filter = textinput.New()
	m.reposByProvider["github"] = []repoItem{
		{Repo: provider.Repo{Owner: "me", Name: "g1"}, Provider: gh},
		{Repo: provider.Repo{Owner: "acme", Name: "g2"}, Provider: gh},
	}
	m.reposByProvider["codeberg"] = []repoItem{
		{Repo: provider.Repo{Owner: "me", Name: "c1"}, Provider: cb},
	}
	return m
}

func TestCombinedViewAggregatesAcrossProviders(t *testing.T) {
	m := combinedModel()
	m.flatAll = true
	m.selProvider = nil
	m.nav = levelRepos
	got := m.visibleRepos()
	if len(got) != 3 {
		t.Fatalf("combined visibleRepos = %d, want 3 (all providers)", len(got))
	}
}

func TestCombinedProviderRowsHasAllEntryLast(t *testing.T) {
	m := combinedModel()
	rows := m.providerRows()
	if len(rows) != 3 || rows[len(rows)-1] != allReposLabel {
		t.Fatalf("providerRows = %v, want providers then %q", rows, allReposLabel)
	}
}

func TestCombinedBadge(t *testing.T) {
	m := combinedModel()
	badge := m.combinedBadge()
	// 3 repos total; owners loaded: github me(✓) acme(✗) + codeberg me(✓) = 2/3.
	if !strings.Contains(badge, "3 repos") || !strings.Contains(badge, "2/3 owners loaded") {
		t.Fatalf("badge = %q", badge)
	}
	if !strings.Contains(badge, "refresh all") {
		t.Fatalf("incomplete cache should hint refresh: %q", badge)
	}
}

func TestCombinedEnterAndBack(t *testing.T) {
	m := combinedModel()
	m.cursor = len(m.providers) // the all-entry row
	send(m, key("enter"))
	if !m.flatAll || m.nav != levelRepos || m.selProvider != nil {
		t.Fatalf("after enter: flatAll=%v nav=%d selProvider=%v", m.flatAll, m.nav, m.selProvider)
	}
	send(m, key("esc"))
	if m.flatAll || m.nav != levelProviders {
		t.Fatalf("after back: flatAll=%v nav=%d, want providers", m.flatAll, m.nav)
	}
}

func TestCombinedRefreshAllRefreshesEveryProvider(t *testing.T) {
	m := combinedModel()
	m.flatAll = true
	m.nav = levelRepos
	var refreshed []string
	m.refresher = func(_ context.Context, p *config.Provider) tea.Cmd {
		refreshed = append(refreshed, p.Name)
		return nil
	}
	m.refresh()
	if len(refreshed) != 2 {
		t.Fatalf("refresh all called %d providers, want 2: %v", len(refreshed), refreshed)
	}
}
