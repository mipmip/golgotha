package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

func TestPinSelfFirst(t *testing.T) {
	cases := []struct {
		name   string
		owners []string
		self   string
		want   []string
	}{
		{"self present middle", []string{"acme", "mipmip", "zeta"}, "mipmip", []string{"mipmip", "acme", "zeta"}},
		{"self already first", []string{"mipmip", "acme"}, "mipmip", []string{"mipmip", "acme"}},
		{"self absent", []string{"acme", "beta"}, "mipmip", []string{"acme", "beta"}},
		{"empty self", []string{"acme"}, "", []string{"acme"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := pinSelfFirst(tc.owners, tc.self)
			if len(got) != len(tc.want) {
				t.Fatalf("pinSelfFirst = %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("pinSelfFirst = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// TestOwnersForPinsSelf verifies the self account (provider Username) leads the
// owner list even though the index is otherwise sorted.
func TestOwnersForPinsSelf(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", Username: "mipmip"}
	m := &Model{
		providers:        []*config.Provider{p},
		ownersByProvider: map[string][]string{"github": {"acme", "mipmip", "zeta"}},
		reposByProvider:  map[string][]repoItem{},
	}
	got := m.ownersFor(p)
	if len(got) != 3 || got[0] != "mipmip" {
		t.Fatalf("ownersFor = %v, want self (mipmip) first", got)
	}
	// Remaining owners stay sorted after self.
	if got[1] != "acme" || got[2] != "zeta" {
		t.Fatalf("ownersFor = %v, want [mipmip acme zeta]", got)
	}
}

// TestVisibleReposScopesToSelfOwner verifies that selecting the self owner (its
// real login) scopes the repo list to only that owner's repos — the bug from
// skull2-qp5y where the self view over-showed.
func TestVisibleReposScopesToSelfOwner(t *testing.T) {
	p := &config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", Username: "mipmip"}
	m := &Model{
		providers:       []*config.Provider{p},
		selProvider:     p,
		selOwner:        "mipmip",
		nav:             levelRepos,
		reposByProvider: map[string][]repoItem{},
	}
	m.filter = textinput.New()
	m.reposByProvider["github"] = []repoItem{
		{Repo: provider.Repo{Owner: "mipmip", Name: "mine"}, Provider: p},
		{Repo: provider.Repo{Owner: "acme", Name: "theirs"}, Provider: p},
	}
	got := m.visibleRepos()
	if len(got) != 1 || got[0].Repo.Owner != "mipmip" {
		t.Fatalf("visibleRepos for self = %v, want only mipmip-owned repos", got)
	}
}
