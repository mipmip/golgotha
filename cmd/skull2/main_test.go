package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mipmip/skull2/internal/cache"
	"github.com/mipmip/skull2/internal/provider"
)

func gitEnvT(t *testing.T) {
	t.Helper()
	for _, kv := range [][2]string{
		{"GIT_AUTHOR_NAME", "Skull2 Test"},
		{"GIT_AUTHOR_EMAIL", "test@skull2.invalid"},
		{"GIT_COMMITTER_NAME", "Skull2 Test"},
		{"GIT_COMMITTER_EMAIL", "test@skull2.invalid"},
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

	cfgDir := filepath.Join(cfgHome, "skull2")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgYAML := "base_dir: " + base + "\n" +
		"clone_pattern_tpl: \"{{.BaseDir}}/{{.Owner}}/{{.Repo}}\"\n" +
		"providers:\n" +
		"  - name: test\n" +
		"    type: github\n" +
		"    short: gh\n" +
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
	raw, _ := os.ReadFile(filepath.Join(cacheHome, "skull2", "test.json"))
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
	cfgDir := filepath.Join(cfgHome, "skull2")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgYAML := "providers:\n  - name: test\n    type: github\n    short: gh\n"
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
	cfgDir := filepath.Join(cfgHome, "skull2")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgYAML := "providers:\n  - name: test\n    type: github\n    short: gh\n"
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
