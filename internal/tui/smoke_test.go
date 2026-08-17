package tui

import (
	"testing"
	"time"

	"github.com/mipmip/golgotha/internal/cache"
	"github.com/mipmip/golgotha/internal/config"
	"github.com/mipmip/golgotha/internal/provider"
)

// TestNewFromCacheSmoke constructs the initial model from a temp cache and calls
// Init/View without panicking.
func TestNewFromCacheSmoke(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	if err := cache.Save("github", cache.Cache{
		FetchedAt: time.Now().UTC(),
		Repos: []provider.Repo{
			{Owner: "mipmip", Name: "golgotha", WebURL: "https://github.com/mipmip/golgotha"},
			{Owner: "acme", Name: "widgets", WebURL: "https://github.com/acme/widgets"},
		},
	}); err != nil {
		t.Fatalf("saving cache: %v", err)
	}

	cfg := &config.Config{
		BaseDir:         dir, // keep clone paths within base
		ClonePatternTpl: config.DefaultClonePatternTpl,
		Providers: []config.Provider{
			{Name: "github", Type: config.ProviderGitHub, Short: "gh", WebURL: "https://github.com"},
		},
	}

	m := New(cfg)
	if got := len(m.reposByProvider["github"]); got != 2 {
		t.Fatalf("expected 2 repos loaded, got %d", got)
	}

	if cmd := m.Init(); cmd != nil {
		t.Fatal("expected nil Init command")
	}

	// View at each level must not panic.
	if m.View() == "" {
		t.Fatal("expected non-empty provider view")
	}
	m.nav = levelOwners
	m.selProvider = &cfg.Providers[0]
	_ = m.View()
	m.nav = levelRepos
	m.selOwner = "mipmip"
	_ = m.View()

	// Filtering view.
	m.filtering = true
	m.filter.SetValue("sk")
	_ = m.View()
}
