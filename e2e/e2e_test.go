// Package e2e contains hermetic end-to-end tests for the primary skull2 flows:
// refresh -> browse -> clone, and sync (clone-missing then fast-forward). They
// stand up an httptest.Server for the provider API and use local bare git repos
// (file:// clone sources) so no real network or SSH is required. Everything runs
// under t.TempDir() with t.Setenv for HOME/XDG and tokens.
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mipmip/skull2/internal/cache"
	"github.com/mipmip/skull2/internal/clonepath"
	"github.com/mipmip/skull2/internal/config"
	"github.com/mipmip/skull2/internal/provider"
	"github.com/mipmip/skull2/internal/syncer"
)

// gitEnv returns a hermetic environment for git so commits are deterministic and
// no user/global/system config or credential prompts bleed into the tests.
func gitEnv() []string {
	return append(os.Environ(),
		"GIT_AUTHOR_NAME=Skull2 E2E",
		"GIT_AUTHOR_EMAIL=e2e@skull2.invalid",
		"GIT_COMMITTER_NAME=Skull2 E2E",
		"GIT_COMMITTER_EMAIL=e2e@skull2.invalid",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_TERMINAL_PROMPT=0",
	)
}

// runGit runs git in dir, failing the test on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// hermeticGitEnv sets the GIT_* variables on the process so the real ExecGit
// runner (used by the syncer) also runs hermetically.
func hermeticGitEnv(t *testing.T) {
	t.Helper()
	for _, kv := range [][2]string{
		{"GIT_AUTHOR_NAME", "Skull2 E2E"},
		{"GIT_AUTHOR_EMAIL", "e2e@skull2.invalid"},
		{"GIT_COMMITTER_NAME", "Skull2 E2E"},
		{"GIT_COMMITTER_EMAIL", "e2e@skull2.invalid"},
		{"GIT_CONFIG_GLOBAL", "/dev/null"},
		{"GIT_CONFIG_SYSTEM", "/dev/null"},
		{"GIT_TERMINAL_PROMPT", "0"},
	} {
		t.Setenv(kv[0], kv[1])
	}
}

// makeBareRemote creates a bare "remote" repo with one commit on branch main and
// returns its filesystem path. The path doubles as a file:// clone URL.
func makeBareRemote(t *testing.T, name string) string {
	t.Helper()
	root := t.TempDir()
	bare := filepath.Join(root, name+".git")
	runGit(t, root, "init", "--bare", "-b", "main", bare)

	work := filepath.Join(root, "seed")
	runGit(t, root, "clone", bare, work)
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-m", "initial")
	runGit(t, work, "push", "origin", "main")
	return bare
}

// addRemoteCommit pushes a new commit to the bare remote so existing clones fall
// behind and can be fast-forwarded.
func addRemoteCommit(t *testing.T, bare string) {
	t.Helper()
	root := t.TempDir()
	work := filepath.Join(root, "push")
	runGit(t, root, "clone", bare, work)
	if err := os.WriteFile(filepath.Join(work, "NEW.md"), []byte("more\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-m", "second")
	runGit(t, work, "push", "origin", "main")
}

// ghRepoFixture is the subset of GitHub repository JSON the client consumes.
// bareURL is placed in both ssh_url and clone_url so whichever protocol the
// config selects points at the local bare repo.
type ghRepoFixture struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
	SSHURL        string    `json:"ssh_url"`
	CloneURL      string    `json:"clone_url"`
	HTMLURL       string    `json:"html_url"`
	DefaultBranch string    `json:"default_branch"`
	Archived      bool      `json:"archived"`
	Fork          bool      `json:"fork"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// newGitHubFixtureServer serves the GitHub REST org-repos endpoint for owner,
// returning a single repository whose clone URLs point at bareURL. It asserts
// the Bearer token matches wantToken.
func newGitHubFixtureServer(t *testing.T, owner, repo, bareURL, wantToken string) *httptest.Server {
	t.Helper()
	wantPath := "/orgs/" + owner + "/repos"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+wantToken {
			http.Error(w, "bad auth: "+got, http.StatusUnauthorized)
			return
		}
		if r.URL.Path != wantPath {
			http.Error(w, "unexpected path "+r.URL.Path, http.StatusNotFound)
			return
		}
		var fx ghRepoFixture
		fx.Name = repo
		fx.Description = "fixture repo"
		fx.Owner.Login = owner
		fx.SSHURL = bareURL
		fx.CloneURL = bareURL
		fx.HTMLURL = "https://example.test/" + owner + "/" + repo
		fx.DefaultBranch = "main"
		fx.UpdatedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]ghRepoFixture{fx}); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// writeConfig writes a minimal config.yaml under XDG_CONFIG_HOME/skull2 pointing
// api_url at apiURL and returns the loaded, validated Config.
func writeConfig(t *testing.T, baseDir, apiURL, owner, tokenEnv string) *config.Config {
	t.Helper()
	xdgConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	cfgDir := filepath.Join(xdgConfig, "skull2")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := fmt.Sprintf(`base_dir: %q
clone_pattern_tpl: "{{.BaseDir}}/{{.Short}}.{{.OwnerLower}}/{{.Repo}}"
providers:
  - name: github-personal
    type: github
    short: gh
    api_url: %q
    web_url: https://example.test
    clone_protocol: https
    auth:
      env: %s
    owners:
      - %s
`, baseDir, apiURL, tokenEnv, owner)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	return cfg
}

// TestRefreshBrowseClone proves the full refresh -> browse -> clone flow:
//   - refresh from a mocked GitHub API into the JSON cache,
//   - "browse" the cache to select a repo,
//   - resolve its clone target via the clone-path template,
//   - clone it (from a local bare repo) via the syncer engine,
//   - assert the repo lands at exactly the templated path.
func TestRefreshBrowseClone(t *testing.T) {
	hermeticGitEnv(t)

	const owner = "acme"
	const repo = "widget"
	const tokenEnv = "SKULL2_GITHUB_TOKEN"
	const token = "e2e-token"
	t.Setenv(tokenEnv, token)

	bare := makeBareRemote(t, repo)
	// file:// URL so git clones from the local bare repo without SSH/network.
	bareURL := "file://" + bare

	srv := newGitHubFixtureServer(t, owner, repo, bareURL, token)

	// XDG_CACHE_HOME so the cache lands in a temp dir.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	base := t.TempDir()
	cfg := writeConfig(t, base, srv.URL, owner, tokenEnv)
	p := &cfg.Providers[0]

	ctx := context.Background()

	// --- Refresh: fetch from the mocked provider and populate the cache. ---
	client := provider.NewGitHub(*p, srv.Client())
	repos, err := client.ListRepos(ctx, p.Owners)
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != repo || repos[0].Owner != owner {
		t.Fatalf("unexpected repos from refresh: %+v", repos)
	}
	if err := cache.Save(p.Name, cache.Cache{FetchedAt: time.Now().UTC(), Repos: repos}); err != nil {
		t.Fatalf("cache.Save: %v", err)
	}

	// --- Browse: read the cache back, as the TUI would. ---
	c, ok, err := cache.LoadOrEmpty(p.Name)
	if err != nil || !ok {
		t.Fatalf("cache.LoadOrEmpty: ok=%v err=%v", ok, err)
	}
	if len(c.Repos) != 1 {
		t.Fatalf("expected 1 cached repo, got %d", len(c.Repos))
	}
	selected := c.Repos[0]

	// --- Resolve the templated clone target. ---
	want, err := clonepath.RenderFor(cfg, p, p.WebURL, selected.Owner, selected.Name)
	if err != nil {
		t.Fatalf("clonepath.RenderFor: %v", err)
	}
	expected := filepath.Join(base, "gh.acme", "widget")
	if want != expected {
		t.Fatalf("templated path = %q want %q", want, expected)
	}

	// --- Clone via the syncer engine (https protocol -> HTTPSURL == file://). ---
	eng := syncer.NewEngine(syncer.NewExecGit(), cfg)
	res := eng.CloneRepo(ctx, p, selected)
	if res.Action != syncer.ActionCloned || res.Err != nil {
		t.Fatalf("clone failed: %+v", res)
	}
	if res.Path != expected {
		t.Fatalf("clone path = %q want %q", res.Path, expected)
	}

	// --- Assert the repo landed at exactly the templated path. ---
	if _, err := os.Stat(filepath.Join(expected, ".git")); err != nil {
		t.Fatalf("no .git at templated path %s: %v", expected, err)
	}
	if _, err := os.Stat(filepath.Join(expected, "README.md")); err != nil {
		t.Fatalf("README.md missing from clone: %v", err)
	}
}

// TestSyncCloneThenFastForward proves the sync flow: a first run clones the
// missing repo, and after an upstream commit a second run fast-forwards it.
func TestSyncCloneThenFastForward(t *testing.T) {
	hermeticGitEnv(t)

	bare := makeBareRemote(t, "widget")
	bareURL := "file://" + bare

	base := t.TempDir()
	cfg := &config.Config{
		BaseDir:         base,
		ClonePatternTpl: "{{.BaseDir}}/{{.Short}}.{{.OwnerLower}}/{{.Repo}}",
		Providers: []config.Provider{{
			Name:          "github-personal",
			Type:          config.ProviderGitHub,
			Short:         "gh",
			CloneProtocol: config.ProtocolHTTPS,
		}},
	}
	p := &cfg.Providers[0]

	// Write the cache JSON directly, as a prior refresh would have.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	repos := []provider.Repo{{
		Owner:         "acme",
		Name:          "widget",
		SSHURL:        bareURL,
		HTTPSURL:      bareURL,
		WebURL:        "https://example.test/acme/widget",
		DefaultBranch: "main",
	}}
	if err := cache.Save(p.Name, cache.Cache{FetchedAt: time.Now().UTC(), Repos: repos}); err != nil {
		t.Fatalf("cache.Save: %v", err)
	}

	eng := syncer.NewEngine(syncer.NewExecGit(), cfg)
	ctx := context.Background()

	// Load repos from the cache, as the sync command does.
	c, ok, err := cache.LoadOrEmpty(p.Name)
	if err != nil || !ok {
		t.Fatalf("cache.LoadOrEmpty: ok=%v err=%v", ok, err)
	}
	cached := provider.FilterRepos(p, c.Repos)

	// --- First run: clone the missing repo. ---
	sum1 := eng.SyncProvider(ctx, p, cached)
	if sum1.Cloned != 1 || sum1.Updated != 0 || sum1.Failed != 0 {
		t.Fatalf("first run: want 1 cloned, got %+v", sum1)
	}
	target := filepath.Join(base, "gh.acme", "widget")
	if _, err := os.Stat(filepath.Join(target, "README.md")); err != nil {
		t.Fatalf("first run did not clone to %s: %v", target, err)
	}
	if _, err := os.Stat(filepath.Join(target, "NEW.md")); !os.IsNotExist(err) {
		t.Fatalf("NEW.md should not exist before upstream commit: err=%v", err)
	}

	// --- Upstream advances. ---
	addRemoteCommit(t, bare)

	// --- Second run: fast-forward the existing clone. ---
	sum2 := eng.SyncProvider(ctx, p, cached)
	if sum2.Updated != 1 || sum2.Cloned != 0 || sum2.Failed != 0 {
		t.Fatalf("second run: want 1 updated, got %+v", sum2)
	}
	if _, err := os.Stat(filepath.Join(target, "NEW.md")); err != nil {
		t.Fatalf("second run did not fast-forward (NEW.md missing): %v", err)
	}
}
