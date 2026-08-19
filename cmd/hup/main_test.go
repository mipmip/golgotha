package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mipmip/huphop/internal/cache"
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

func gitEnvT(t *testing.T) {
	t.Helper()
	for _, kv := range [][2]string{
		{"GIT_AUTHOR_NAME", "Huphop Test"},
		{"GIT_AUTHOR_EMAIL", "test@huphop.invalid"},
		{"GIT_COMMITTER_NAME", "Huphop Test"},
		{"GIT_COMMITTER_EMAIL", "test@huphop.invalid"},
		{"GIT_CONFIG_GLOBAL", "/dev/null"},
		{"GIT_CONFIG_SYSTEM", "/dev/null"},
		{"GIT_TERMINAL_PROMPT", "0"},
	} {
		t.Setenv(kv[0], kv[1])
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func makeBare(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	git(t, root, "init", "--bare", "-b", "main", bare)
	work := filepath.Join(root, "seed")
	git(t, root, "clone", bare, work)
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, work, "add", ".")
	git(t, work, "commit", "-m", "init")
	git(t, work, "push", "origin", "main")
	return bare
}

// TestRunSyncEndToEnd drives runSync with XDG dirs pointed at temp locations and
// --no-refresh so no network client is built.
func TestRunSyncEndToEnd(t *testing.T) {
	gitEnvT(t)
	bare := makeBare(t)

	cfgHome := t.TempDir()
	cacheHome := t.TempDir()
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	cfgDir := filepath.Join(cfgHome, "huphop")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgYAML := "base_dir: " + base + "\n" +
		"clone_pattern_tpl: \"{{.BaseDir}}/{{.Owner}}/{{.Repo}}\"\n" +
		"providers:\n" +
		"  - name: test\n" +
		"    type: github\n" +
		"    short: gh\n" +
		"    username: me\n" +
		"    clone_protocol: ssh\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Seed the cache so --no-refresh has something to act on.
	c := cache.Cache{FetchedAt: time.Now().UTC(), Repos: []provider.Repo{{
		Owner: "acme", Name: "widget", SSHURL: bare, DefaultBranch: "main",
	}}}
	if err := cache.Save("test", c); err != nil {
		t.Fatal(err)
	}
	// Sanity: the cache round-trips as JSON.
	var check cache.Cache
	raw, _ := os.ReadFile(filepath.Join(cacheHome, "huphop", "test.json"))
	if err := json.Unmarshal(raw, &check); err != nil {
		t.Fatalf("cache json invalid: %v", err)
	}

	if err := runSync([]string{"--no-refresh"}); err != nil {
		t.Fatalf("runSync: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "acme", "widget", ".git")); err != nil {
		t.Fatalf("expected clone: %v", err)
	}
}

func TestRunSyncUnknownProvider(t *testing.T) {
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)
	cfgDir := filepath.Join(cfgHome, "huphop")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgYAML := "providers:\n  - name: test\n    type: github\n    short: gh\n    username: me\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runSync([]string{"--no-refresh", "--provider", "nope"}); err == nil {
		t.Fatalf("expected error for unknown provider")
	}
}

func TestRunSyncNoCacheNoFailure(t *testing.T) {
	cfgHome := t.TempDir()
	cacheHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)
	t.Setenv("XDG_CACHE_HOME", cacheHome)
	cfgDir := filepath.Join(cfgHome, "huphop")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgYAML := "providers:\n  - name: test\n    type: github\n    short: gh\n    username: me\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	// No cache present, --no-refresh: logs a warning but no repo failed, so exit 0.
	if err := runSync([]string{"--no-refresh"}); err != nil {
		t.Fatalf("expected nil error when cache absent, got %v", err)
	}
}

func TestRunReturnsUnknownCommand(t *testing.T) {
	if err := run([]string{"bogus"}); err == nil {
		t.Fatalf("expected unknown command error")
	}
}

// fakeDiscoverClient is a hermetic provider.Provider for eager-sweep tests. It
// records which owner sets ListRepos was called with and returns one repo per
// non-self owner.
type fakeDiscoverClient struct {
	owners     []string
	mu         sync.Mutex // guards listCalls (fetch pool calls ListRepos concurrently)
	listCalls  [][]string
	reposByReq map[string][]provider.Repo
}

func (f *fakeDiscoverClient) ListOwners(_ context.Context) ([]string, error) {
	return f.owners, nil
}

func (f *fakeDiscoverClient) ListRepos(_ context.Context, owners []string) ([]provider.Repo, error) {
	f.mu.Lock()
	f.listCalls = append(f.listCalls, owners)
	f.mu.Unlock()
	if len(owners) == 0 {
		// Self (defensive; callers now pass the real username).
		return []provider.Repo{{Owner: "me", Name: "personal"}}, nil
	}
	owner := owners[0]
	return []provider.Repo{{Owner: owner, Name: owner + "-repo"}}, nil
}

func (f *fakeDiscoverClient) RepoDetails(_ context.Context, _, _ string) (provider.Details, error) {
	return provider.Details{}, nil
}

func (f *fakeDiscoverClient) Readme(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func TestRefreshAllOwnersEagerSweep(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	p := &config.Provider{
		Name:      "gh",
		Type:      config.ProviderGitHub,
		Short:     "gh",
		Username:  "me",
		AllOwners: true,
		Owners:    []string{"explicit"},
	}
	client := &fakeDiscoverClient{owners: []string{"acme", "beta"}}

	if err := refreshAllOwners(context.Background(), client, p, &cliProgressPrinter{w: io.Discard}); err != nil {
		t.Fatal(err)
	}

	c, err := cache.Load("gh")
	if err != nil {
		t.Fatal(err)
	}
	// Resolved: self, acme, beta, explicit -> all fetched, all in the index.
	names := map[string]bool{}
	for _, o := range c.Owners {
		names[o.Name] = true
		if o.FetchedAt == nil {
			t.Fatalf("owner %q should be fetched after eager sweep", o.Name)
		}
	}
	for _, want := range []string{"me", "acme", "beta", "explicit"} {
		if !names[want] {
			t.Fatalf("owner %q missing from index %+v", want, c.Owners)
		}
	}
	// Self was fetched under its real username ("me").
	sawSelf := false
	for _, call := range client.listCalls {
		if len(call) == 1 && call[0] == "me" {
			sawSelf = true
		}
	}
	if !sawSelf {
		t.Fatalf("expected a ListRepos call for self (username %q), calls=%v", "me", client.listCalls)
	}
}

func TestRefreshAllOwnersSkipsExcluded(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	p := &config.Provider{
		Name:          "gh",
		Type:          config.ProviderGitHub,
		Short:         "gh",
		Username:      "me",
		AllOwners:     true,
		ExcludeOwners: []string{"noisy", "me"},
	}
	client := &fakeDiscoverClient{owners: []string{"acme", "noisy"}}

	if err := refreshAllOwners(context.Background(), client, p, &cliProgressPrinter{w: io.Discard}); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Load("gh")
	if err != nil {
		t.Fatal(err)
	}
	// noisy excluded; self excluded via token; only acme remains.
	if len(c.Owners) != 1 || c.Owners[0].Name != "acme" {
		t.Fatalf("owners = %+v, want only acme", c.Owners)
	}
	// The excluded owner was never fetched/cloned: no repos for it.
	if len(c.ReposFor("noisy")) != 0 {
		t.Fatal("excluded owner should not be fetched")
	}
	for _, call := range client.listCalls {
		if len(call) == 1 && call[0] == "noisy" {
			t.Fatal("ListRepos should not be called for excluded owner")
		}
		if len(call) == 0 {
			t.Fatal("self was excluded; ListRepos should not be called for self")
		}
	}
}
