package syncer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

func TestParseJJProgress(t *testing.T) {
	cases := []struct {
		in       string
		wantFrac float64
		wantOk   bool
	}{
		{"\x1b[2K 58% [\x1b[38;5;2m###\x1b[39m]", 0.58, true},
		{" 100% []", 1.0, true},
		{"\x1b[?25lremote: Enumerating objects: 102, done.", 0, false},
		{"Fetching into new repo in \"x\"", 0, false},
		{"", 0, false},
	}
	for _, tc := range cases {
		frac, ok := parseJJProgress(tc.in)
		if ok != tc.wantOk {
			t.Fatalf("parseJJProgress(%q) ok=%v want %v", tc.in, ok, tc.wantOk)
		}
		if ok && (frac < tc.wantFrac-0.001 || frac > tc.wantFrac+0.001) {
			t.Fatalf("parseJJProgress(%q) frac=%v want %v", tc.in, frac, tc.wantFrac)
		}
	}
}

func TestStripANSI(t *testing.T) {
	in := "\x1b[2K\x1b[?25l 42% [\x1b[38;5;2mX\x1b[39m]\x1b[?25h"
	if got := stripANSI(in); got != " 42% [X]" {
		t.Fatalf("stripANSI = %q, want %q", got, " 42% [X]")
	}
}

func jjTestEngine(vcs string, jjClone func(ctx context.Context, url, dir string, emit func(float64, string)) error) (*Engine, *config.Provider, provider.Repo) {
	cfg := &config.Config{BaseDir: "/tmp", ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}", CloneVCS: vcs}
	e := NewEngine(&fakeGit{}, cfg)
	e.JJClone = jjClone
	p := &config.Provider{Name: "gh", Type: config.ProviderGitHub, Short: "gh", CloneProtocol: "ssh"}
	r := provider.Repo{Owner: "me", Name: "proj", SSHURL: "git@github.com:me/proj.git"}
	return e, p, r
}

func TestCloneRepoUsesJJWhenConfigured(t *testing.T) {
	var got struct {
		url, dir string
		hadEmit  bool
	}
	e, p, r := jjTestEngine("jj", func(_ context.Context, url, dir string, emit func(float64, string)) error {
		got.url, got.dir, got.hadEmit = url, dir, emit != nil
		return nil
	})
	res := e.CloneRepo(context.Background(), p, r)
	if res.Err != nil || res.Action != ActionCloned {
		t.Fatalf("CloneRepo jj: action=%v err=%v", res.Action, res.Err)
	}
	if got.url != r.SSHURL {
		t.Fatalf("jj clone url = %q, want %q", got.url, r.SSHURL)
	}
	if got.hadEmit {
		t.Fatal("plain CloneRepo should pass a nil emit to jj")
	}
	// git backend was NOT used.
	if len(e.Git.(*fakeGit).cloned) != 0 {
		t.Fatal("jj clone must not call the git backend")
	}
}

func TestCloneRepoUsesGitByDefault(t *testing.T) {
	e, p, r := jjTestEngine("git", func(context.Context, string, string, func(float64, string)) error {
		t.Fatal("git clone must not call JJClone")
		return nil
	})
	res := e.CloneRepo(context.Background(), p, r)
	if res.Err != nil || res.Action != ActionCloned {
		t.Fatalf("CloneRepo git: action=%v err=%v", res.Action, res.Err)
	}
	if len(e.Git.(*fakeGit).cloned) != 1 {
		t.Fatalf("expected one git clone, got %v", e.Git.(*fakeGit).cloned)
	}
}

func TestCloneRepoProgressJJEmits(t *testing.T) {
	var fracs []float64
	e, p, r := jjTestEngine("jj", func(_ context.Context, _, _ string, emit func(float64, string)) error {
		emit(0.3, "cloning")
		emit(0.9, "cloning")
		return nil
	})
	res := e.CloneRepoProgress(context.Background(), p, r, func(f float64, _ string) { fracs = append(fracs, f) })
	if res.Err != nil || res.Action != ActionCloned {
		t.Fatalf("CloneRepoProgress jj: action=%v err=%v", res.Action, res.Err)
	}
	if len(fracs) != 2 || fracs[0] != 0.3 || fracs[1] != 0.9 {
		t.Fatalf("jj progress emits = %v, want [0.3 0.9]", fracs)
	}
}

func TestEnsureJJMissing(t *testing.T) {
	t.Setenv("PATH", "") // no jj resolvable
	if err := ensureJJ(); err == nil {
		t.Fatal("expected an error when jj is not on PATH")
	}
}

// TestExecJJClonePlainReal exercises the real (default) execJJClone plain path
// against a local bare repo, so the jj-clone code is covered. Skips when jj is
// not installed.
func TestExecJJClonePlainReal(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not on PATH")
	}
	bare := makeBareRemote(t)

	// Give jj an identity in a clean environment.
	home := t.TempDir()
	cfgPath := filepath.Join(home, "jjconfig.toml")
	if err := os.WriteFile(cfgPath, []byte("[user]\nname = \"t\"\nemail = \"t@t\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("JJ_CONFIG", cfgPath)

	dst := filepath.Join(t.TempDir(), "clone")
	if err := execJJClone(context.Background(), "file://"+bare, dst, nil); err != nil {
		t.Fatalf("execJJClone plain: %v", err)
	}
	// Colocated: both a git dir and a jj dir exist.
	for _, sub := range []string{".git", ".jj"} {
		if _, err := os.Stat(filepath.Join(dst, sub)); err != nil {
			t.Fatalf("expected %s in colocated clone: %v", sub, err)
		}
	}
}
