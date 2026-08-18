package syncer

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

// gitEnv returns an environment that makes git commits deterministic and
// hermetic (no user config, agent, or global hooks bleeding in).
func gitEnv() []string {
	return append(os.Environ(),
		"GIT_AUTHOR_NAME=Huphop Test",
		"GIT_AUTHOR_EMAIL=test@huphop.invalid",
		"GIT_COMMITTER_NAME=Huphop Test",
		"GIT_COMMITTER_EMAIL=test@huphop.invalid",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_TERMINAL_PROMPT=0",
	)
}

// runGit runs git in dir and fails the test on error.
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

// makeBareRemote creates a bare "remote" repo with one commit on branch main and
// returns its path.
func makeBareRemote(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	runGit(t, root, "init", "--bare", "-b", "main", bare)

	// Seed the bare repo via a scratch working clone.
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

// addRemoteCommit adds a new commit to the bare remote so clones fall behind.
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

// testEngine returns an Engine using the real ExecGit runner, with hermetic
// GIT_* env vars set on the process so clones/commits work in the sandbox
// without touching the user's global config.
func testEngine(t *testing.T, cfg *config.Config) *Engine {
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
	return NewEngine(NewExecGit(), cfg)
}

// newConfig returns a config whose base_dir is base and one ssh (via file URL)
// provider using the default template.
func newConfig(base string) (*config.Config, *config.Provider) {
	cfg := &config.Config{
		BaseDir:         base,
		ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}",
		Providers: []config.Provider{{
			Name:          "test",
			Type:          config.ProviderGitHub,
			Short:         "gh",
			CloneProtocol: config.ProtocolSSH,
		}},
	}
	return cfg, &cfg.Providers[0]
}

func TestSyncClonesMissingRepo(t *testing.T) {
	bare := makeBareRemote(t)
	base := t.TempDir()
	cfg, p := newConfig(base)
	eng := testEngine(t, cfg)

	repos := []provider.Repo{{
		Owner: "acme", Name: "widget",
		SSHURL: bare, HTTPSURL: bare, DefaultBranch: "main",
	}}

	sum := eng.SyncProvider(context.Background(), p, repos)
	if sum.Cloned != 1 || sum.Failed != 0 {
		t.Fatalf("want 1 cloned 0 failed, got %+v", sum)
	}
	target := filepath.Join(base, "acme", "widget")
	if _, err := os.Stat(filepath.Join(target, ".git")); err != nil {
		t.Fatalf("expected clone at %s: %v", target, err)
	}
	if _, err := os.Stat(filepath.Join(target, "README.md")); err != nil {
		t.Fatalf("expected README in clone: %v", err)
	}
	if got := sum.Results[0].Action; got != ActionCloned {
		t.Fatalf("want ActionCloned, got %q", got)
	}
}

func TestSyncFastForwardsCleanRepo(t *testing.T) {
	bare := makeBareRemote(t)
	base := t.TempDir()
	cfg, p := newConfig(base)
	eng := testEngine(t, cfg)

	repos := []provider.Repo{{
		Owner: "acme", Name: "widget",
		SSHURL: bare, DefaultBranch: "main",
	}}

	// First sync clones.
	if sum := eng.SyncProvider(context.Background(), p, repos); sum.Cloned != 1 {
		t.Fatalf("initial clone failed: %+v", sum)
	}

	// Remote moves ahead; second sync must fast-forward.
	addRemoteCommit(t, bare)

	sum := eng.SyncProvider(context.Background(), p, repos)
	if sum.Updated != 1 || sum.Failed != 0 {
		t.Fatalf("want 1 updated 0 failed, got %+v", sum)
	}
	target := filepath.Join(base, "acme", "widget")
	if _, err := os.Stat(filepath.Join(target, "NEW.md")); err != nil {
		t.Fatalf("expected fast-forwarded NEW.md: %v", err)
	}
}

func TestSyncSkipsDirtyRepo(t *testing.T) {
	bare := makeBareRemote(t)
	base := t.TempDir()
	cfg, p := newConfig(base)
	eng := testEngine(t, cfg)

	repos := []provider.Repo{{
		Owner: "acme", Name: "widget",
		SSHURL: bare, DefaultBranch: "main",
	}}
	if sum := eng.SyncProvider(context.Background(), p, repos); sum.Cloned != 1 {
		t.Fatalf("initial clone failed: %+v", sum)
	}

	target := filepath.Join(base, "acme", "widget")
	// Introduce an uncommitted change.
	if err := os.WriteFile(filepath.Join(target, "README.md"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Remote also advances; a dirty repo must still be skipped, not modified.
	addRemoteCommit(t, bare)

	sum := eng.SyncProvider(context.Background(), p, repos)
	if sum.Skipped != 1 || sum.Updated != 0 || sum.Failed != 0 {
		t.Fatalf("want 1 skipped, got %+v", sum)
	}
	if sum.Results[0].Warning == "" {
		t.Fatalf("expected a warning on skip")
	}
	// The dirty content must be untouched and NEW.md must not have appeared.
	got, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil || string(got) != "dirty\n" {
		t.Fatalf("dirty file modified: %q err=%v", got, err)
	}
	if _, err := os.Stat(filepath.Join(target, "NEW.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dirty repo was fast-forwarded (NEW.md present)")
	}
}

func TestSyncFallsBackToRemoteHead(t *testing.T) {
	bare := makeBareRemote(t)
	base := t.TempDir()
	cfg, p := newConfig(base)
	eng := testEngine(t, cfg)

	// DefaultBranch empty forces CurrentDefaultBranch fallback.
	repos := []provider.Repo{{
		Owner: "acme", Name: "widget",
		SSHURL: bare, DefaultBranch: "",
	}}
	if sum := eng.SyncProvider(context.Background(), p, repos); sum.Cloned != 1 {
		t.Fatalf("initial clone failed: %+v", sum)
	}
	addRemoteCommit(t, bare)

	sum := eng.SyncProvider(context.Background(), p, repos)
	if sum.Updated != 1 {
		t.Fatalf("want 1 updated via remote-HEAD fallback, got %+v", sum)
	}
}

// --- Fake Git runner tests for decision logic ---

type fakeGit struct {
	repos   map[string]bool // dir -> exists
	dirty   map[string]bool
	cloned  []string
	fetched []string
	ffed    []string

	cloneErr error
	dirtyErr error
	fetchErr error
	ffErr    error
	defBr    string
	defErr   error
}

func (f *fakeGit) Clone(_ context.Context, url, dir string) error {
	if f.cloneErr != nil {
		return f.cloneErr
	}
	f.cloned = append(f.cloned, dir)
	if f.repos == nil {
		f.repos = map[string]bool{}
	}
	f.repos[dir] = true
	return nil
}
func (f *fakeGit) Fetch(_ context.Context, dir string) error {
	f.fetched = append(f.fetched, dir)
	return f.fetchErr
}
func (f *fakeGit) FastForward(_ context.Context, dir, branch string) error {
	f.ffed = append(f.ffed, dir+"@"+branch)
	return f.ffErr
}
func (f *fakeGit) IsDirty(_ context.Context, dir string) (bool, error) {
	return f.dirty[dir], f.dirtyErr
}
func (f *fakeGit) IsRepo(dir string) bool { return f.repos[dir] }
func (f *fakeGit) CurrentDefaultBranch(_ context.Context, dir string) (string, error) {
	return f.defBr, f.defErr
}

func fakeCfg() (*config.Config, *config.Provider) {
	cfg := &config.Config{
		BaseDir:         "/base",
		ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}",
		Providers: []config.Provider{{
			Name: "test", Type: config.ProviderGitHub, Short: "gh",
			CloneProtocol: config.ProtocolSSH,
		}},
	}
	return cfg, &cfg.Providers[0]
}

func TestFakeCloneFailureRecordsFailed(t *testing.T) {
	cfg, p := fakeCfg()
	f := &fakeGit{cloneErr: errors.New("boom")}
	eng := NewEngine(f, cfg)
	sum := eng.SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", SSHURL: "url", DefaultBranch: "main",
	}})
	if sum.Failed != 1 || sum.Results[0].Err == nil {
		t.Fatalf("want failed with err, got %+v", sum)
	}
}

func TestFakeMissingCloneURLFails(t *testing.T) {
	cfg, p := fakeCfg()
	f := &fakeGit{}
	eng := NewEngine(f, cfg)
	sum := eng.SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", DefaultBranch: "main", // no SSHURL
	}})
	if sum.Failed != 1 {
		t.Fatalf("want failed on missing URL, got %+v", sum)
	}
}

func TestFakeHTTPSProtocolUsesHTTPSURL(t *testing.T) {
	cfg, p := fakeCfg()
	p.CloneProtocol = config.ProtocolHTTPS
	f := &fakeGit{}
	eng := NewEngine(f, cfg)
	sum := eng.SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", HTTPSURL: "https://x", DefaultBranch: "main",
	}})
	if sum.Cloned != 1 {
		t.Fatalf("want cloned via https, got %+v", sum)
	}
	if u := cloneURL(p, provider.Repo{HTTPSURL: "https://x", SSHURL: "ssh://y"}); u != "https://x" {
		t.Fatalf("cloneURL https wrong: %q", u)
	}
}

func TestFakeDirtyCheckErrorFails(t *testing.T) {
	cfg, p := fakeCfg()
	dir := "/base/o/r"
	f := &fakeGit{repos: map[string]bool{dir: true}, dirtyErr: errors.New("nope")}
	eng := NewEngine(f, cfg)
	sum := eng.SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", SSHURL: "url", DefaultBranch: "main",
	}})
	if sum.Failed != 1 {
		t.Fatalf("want failed on dirty check error, got %+v", sum)
	}
}

func TestFakeFetchAndFFErrors(t *testing.T) {
	cfg, p := fakeCfg()
	dir := "/base/o/r"

	f := &fakeGit{repos: map[string]bool{dir: true}, fetchErr: errors.New("net")}
	sum := NewEngine(f, cfg).SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", SSHURL: "url", DefaultBranch: "main",
	}})
	if sum.Failed != 1 {
		t.Fatalf("want failed on fetch error, got %+v", sum)
	}

	f2 := &fakeGit{repos: map[string]bool{dir: true}, ffErr: errors.New("diverged")}
	sum2 := NewEngine(f2, cfg).SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", SSHURL: "url", DefaultBranch: "main",
	}})
	if sum2.Failed != 1 {
		t.Fatalf("want failed on ff error, got %+v", sum2)
	}
}

func TestFakeDefaultBranchFallback(t *testing.T) {
	cfg, p := fakeCfg()
	dir := "/base/o/r"
	f := &fakeGit{repos: map[string]bool{dir: true}, defBr: "trunk"}
	eng := NewEngine(f, cfg)
	sum := eng.SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", SSHURL: "url", // empty DefaultBranch
	}})
	if sum.Updated != 1 {
		t.Fatalf("want updated, got %+v", sum)
	}
	if len(f.ffed) != 1 || f.ffed[0] != dir+"@trunk" {
		t.Fatalf("expected ff on trunk, got %v", f.ffed)
	}

	// Fallback error path.
	f2 := &fakeGit{repos: map[string]bool{dir: true}, defErr: errors.New("no HEAD")}
	sum2 := NewEngine(f2, cfg).SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", SSHURL: "url",
	}})
	if sum2.Failed != 1 {
		t.Fatalf("want failed on default-branch fallback error, got %+v", sum2)
	}
}

func TestFakeBadTemplateFails(t *testing.T) {
	cfg, p := fakeCfg()
	cfg.ClonePatternTpl = "{{.Nope}}"
	f := &fakeGit{}
	sum := NewEngine(f, cfg).SyncProvider(context.Background(), p, []provider.Repo{{
		Owner: "o", Name: "r", SSHURL: "url", DefaultBranch: "main",
	}})
	if sum.Failed != 1 {
		t.Fatalf("want failed on bad template, got %+v", sum)
	}
}

func TestCloneRepoClonesMissing(t *testing.T) {
	bare := makeBareRemote(t)
	base := t.TempDir()
	cfg, p := newConfig(base)
	eng := testEngine(t, cfg)

	res := eng.CloneRepo(context.Background(), p, provider.Repo{
		Owner: "acme", Name: "widget", SSHURL: bare, DefaultBranch: "main",
	})
	if res.Action != ActionCloned || res.Err != nil {
		t.Fatalf("want cloned, got %+v", res)
	}
	target := filepath.Join(base, "acme", "widget")
	if res.Path != target {
		t.Fatalf("path = %q want %q", res.Path, target)
	}
	if _, err := os.Stat(filepath.Join(target, ".git")); err != nil {
		t.Fatalf("expected clone at %s: %v", target, err)
	}
}

func TestCloneRepoSkipsExisting(t *testing.T) {
	cfg, p := fakeCfg()
	dir := "/base/o/r"
	f := &fakeGit{repos: map[string]bool{dir: true}}
	res := NewEngine(f, cfg).CloneRepo(context.Background(), p, provider.Repo{
		Owner: "o", Name: "r", SSHURL: "url", DefaultBranch: "main",
	})
	if res.Action != ActionSkipped || res.Warning == "" {
		t.Fatalf("want skipped with warning, got %+v", res)
	}
	if len(f.cloned) != 0 {
		t.Fatalf("existing repo should not be cloned, got %v", f.cloned)
	}
}

func TestCloneRepoBadTemplateFails(t *testing.T) {
	cfg, p := fakeCfg()
	cfg.ClonePatternTpl = "{{.Nope}}"
	res := NewEngine(&fakeGit{}, cfg).CloneRepo(context.Background(), p, provider.Repo{
		Owner: "o", Name: "r", SSHURL: "url",
	})
	if res.Action != ActionFailed || res.Err == nil {
		t.Fatalf("want failed on bad template, got %+v", res)
	}
}

func TestCloneRepoMissingURLFails(t *testing.T) {
	cfg, p := fakeCfg()
	res := NewEngine(&fakeGit{}, cfg).CloneRepo(context.Background(), p, provider.Repo{
		Owner: "o", Name: "r", // no SSHURL
	})
	if res.Action != ActionFailed || res.Err == nil {
		t.Fatalf("want failed on missing URL, got %+v", res)
	}
}

func TestCloneRepoCloneErrorFails(t *testing.T) {
	cfg, p := fakeCfg()
	f := &fakeGit{cloneErr: errors.New("boom")}
	res := NewEngine(f, cfg).CloneRepo(context.Background(), p, provider.Repo{
		Owner: "o", Name: "r", SSHURL: "url",
	})
	if res.Action != ActionFailed || res.Err == nil {
		t.Fatalf("want failed on clone error, got %+v", res)
	}
}

func TestSummaryAggregates(t *testing.T) {
	s := &Summary{Providers: []ProviderSummary{
		{Provider: "a", Cloned: 1, Updated: 2, Skipped: 0, Failed: 1},
		{Provider: "b", Cloned: 0, Updated: 1, Skipped: 3, Failed: 0},
	}}
	c, u, sk, fl := s.Totals()
	if c != 1 || u != 3 || sk != 3 || fl != 1 {
		t.Fatalf("totals wrong: %d %d %d %d", c, u, sk, fl)
	}
	if !s.HasFailures() {
		t.Fatalf("expected failures")
	}
	ok := &Summary{Providers: []ProviderSummary{{Provider: "a", Cloned: 1}}}
	if ok.HasFailures() {
		t.Fatalf("did not expect failures")
	}
}

func TestExecGitIsRepoAndDefaultBranch(t *testing.T) {
	bare := makeBareRemote(t)
	base := t.TempDir()
	target := filepath.Join(base, "clone")

	for _, kv := range [][2]string{
		{"GIT_CONFIG_GLOBAL", "/dev/null"}, {"GIT_CONFIG_SYSTEM", "/dev/null"},
	} {
		t.Setenv(kv[0], kv[1])
	}
	g := NewExecGit()
	ctx := context.Background()

	if g.IsRepo(target) {
		t.Fatalf("empty target should not be a repo")
	}
	if err := g.Clone(ctx, bare, target); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if !g.IsRepo(target) {
		t.Fatalf("cloned target should be a repo")
	}
	dirty, err := g.IsDirty(ctx, target)
	if err != nil || dirty {
		t.Fatalf("fresh clone should be clean: dirty=%v err=%v", dirty, err)
	}
	br, err := g.CurrentDefaultBranch(ctx, target)
	if err != nil || br != "main" {
		t.Fatalf("default branch want main, got %q err=%v", br, err)
	}
	if err := g.Fetch(ctx, target); err != nil {
		t.Fatalf("fetch: %v", err)
	}
}

func TestExecGitBinDefault(t *testing.T) {
	if got := (&ExecGit{}).bin(); got != "git" {
		t.Fatalf("empty Bin should default to git, got %q", got)
	}
	if got := (&ExecGit{Bin: "custom-git"}).bin(); got != "custom-git" {
		t.Fatalf("explicit Bin ignored, got %q", got)
	}
}

func TestExecGitRunErrorIncludesStderr(t *testing.T) {
	dir := t.TempDir() // not a git repo
	g := NewExecGit()
	// `git status` in a non-repo dir fails and writes to stderr.
	_, err := g.run(context.Background(), dir, "status", "--porcelain")
	if err == nil {
		t.Fatal("expected error running git status outside a repo")
	}
	if !strings.Contains(err.Error(), "git status") {
		t.Fatalf("error should name the command: %v", err)
	}
}

func TestExecGitIsDirtyErrorPropagates(t *testing.T) {
	dir := t.TempDir() // not a git repo
	g := NewExecGit()
	if _, err := g.IsDirty(context.Background(), dir); err == nil {
		t.Fatal("expected IsDirty error outside a repo")
	}
}

func TestExecGitCloneParentDirError(t *testing.T) {
	// Make the parent path a regular file so MkdirAll fails.
	root := t.TempDir()
	file := filepath.Join(root, "afile")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := NewExecGit()
	// dir's parent is <file>/sub which cannot be created.
	err := g.Clone(context.Background(), "irrelevant", filepath.Join(file, "sub", "clone"))
	if err == nil {
		t.Fatal("expected clone to fail creating parent dir under a file")
	}
}

func TestExecGitCurrentDefaultBranchError(t *testing.T) {
	dir := t.TempDir() // not a git repo
	g := NewExecGit()
	if _, err := g.CurrentDefaultBranch(context.Background(), dir); err == nil {
		t.Fatal("expected error resolving default branch outside a repo")
	}
}
