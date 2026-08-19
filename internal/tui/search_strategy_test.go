package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

// newStrategyModel builds a repos-level Model for one owner with the given
// search strategy and repo names, ready for visibleRepos filtering.
func newStrategyModel(strategy string, names ...string) *Model {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh"}
	ti := textinput.New()
	m := &Model{
		cfg:             &config.Config{SearchStrategy: strategy, Providers: []config.Provider{*p}},
		providers:       []*config.Provider{p},
		reposByProvider: map[string][]repoItem{},
		nav:             levelRepos,
		selProvider:     p,
		selOwner:        "mip",
		filter:          ti,
	}
	for _, n := range names {
		m.reposByProvider["github"] = append(m.reposByProvider["github"],
			repoItem{Repo: provider.Repo{Owner: "mip", Name: n}, Provider: p})
	}
	return m
}

func visibleNames(m *Model) []string {
	var out []string
	for _, it := range m.visibleRepos() {
		out = append(out, it.Repo.Name)
	}
	return out
}

func TestSearchStrategyAtReposLevel(t *testing.T) {
	// "nix-services" is a fuzzy (subsequence) match for "nvs" but not a
	// substring match; "dotfiles" matches neither.
	cases := []struct {
		name     string
		strategy string
		query    string
		want     []string
	}{
		{"fuzzy default matches subsequence", config.SearchFuzzy, "nvs", []string{"nix-services"}},
		{"substring rejects subsequence", config.SearchSubstring, "nvs", nil},
		{"substring matches literal", config.SearchSubstring, "nix", []string{"nix-services"}},
		{"quote toggles to fuzzy in substring mode", config.SearchSubstring, "'nvs", []string{"nix-services"}},
		{"quote toggles to substring in fuzzy mode", config.SearchFuzzy, "'nvs", nil},
		{"prefix anchor", config.SearchFuzzy, "^mip/nix", []string{"nix-services"}},
		{"suffix anchor", config.SearchFuzzy, "services$", []string{"nix-services"}},
		{"negation excludes", config.SearchSubstring, "!nix", []string{"dotfiles"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newStrategyModel(tc.strategy, "nix-services", "dotfiles")
			m.filter.SetValue(tc.query)
			got := visibleNames(m)
			if len(got) != len(tc.want) {
				t.Fatalf("query %q strategy %q: got %v, want %v", tc.query, tc.strategy, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("query %q strategy %q: got %v, want %v", tc.query, tc.strategy, got, tc.want)
				}
			}
		})
	}
}

// TestSearchStrategyAppliesAtEveryLevel confirms the configured strategy also
// governs the providers and owners levels, not just repos.
func TestSearchStrategyProvidersAndOwners(t *testing.T) {
	m := newStrategyModel(config.SearchSubstring, "nix-services")
	// Providers level: "gh" subsequence of "github" but query "gthb" is only a
	// subsequence, not a substring, so substring mode must reject it.
	m.nav = levelProviders
	m.filter.SetValue("gthb")
	if got := m.visibleProviders(); len(got) != 0 {
		t.Fatalf("substring providers: got %v, want none", got)
	}
	m.filter.SetValue("git")
	if got := m.visibleProviders(); len(got) != 1 || got[0] != "github" {
		t.Fatalf("substring providers literal: got %v, want [github]", got)
	}
}
