package syncer

import (
	"bufio"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

func TestScanCRorLF(t *testing.T) {
	sc := bufio.NewScanner(strings.NewReader("a\rb\nc\rd"))
	sc.Split(scanCRorLF)
	var got []string
	for sc.Scan() {
		got = append(got, sc.Text())
	}
	want := []string{"a", "b", "c", "d"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("scanCRorLF = %v, want %v", got, want)
	}
}

func TestExecGitCloneProgress(t *testing.T) {
	bare := makeBareRemote(t)
	for _, kv := range [][2]string{
		{"GIT_CONFIG_GLOBAL", "/dev/null"}, {"GIT_CONFIG_SYSTEM", "/dev/null"},
	} {
		t.Setenv(kv[0], kv[1])
	}
	g := NewExecGit()
	dst := filepath.Join(t.TempDir(), "clone")
	// file:// disables git's local-clone optimization, so --progress streams.
	if err := g.CloneProgress(context.Background(), "file://"+bare, dst, func(float64, string) {}); err != nil {
		t.Fatalf("CloneProgress: %v", err)
	}
	if !g.IsRepo(dst) {
		t.Fatalf("expected a cloned repo at %s", dst)
	}
}

func TestParseGitCloneProgress(t *testing.T) {
	cases := []struct {
		line     string
		wantFrac float64
		wantOk   bool
	}{
		{"Receiving objects:  58% (580/1000), 1.5 MiB", 0.58, true},
		{"remote: Counting objects:  50% (5/10)", 0.50, true},
		{"Resolving deltas: 100% (100/100), done.", 1.0, true},
		{"Cloning into 'foo'...", 0, false},
		{"", 0, false},
	}
	for _, tc := range cases {
		frac, phase, ok := parseGitCloneProgress(tc.line)
		if ok != tc.wantOk {
			t.Fatalf("parse(%q) ok=%v, want %v", tc.line, ok, tc.wantOk)
		}
		if ok && (frac < tc.wantFrac-0.001 || frac > tc.wantFrac+0.001) {
			t.Fatalf("parse(%q) frac=%v, want %v (phase %q)", tc.line, frac, tc.wantFrac, phase)
		}
	}
}

// fakeProgressGit embeds fakeGit and adds a progress-emitting clone.
type fakeProgressGit struct {
	*fakeGit
}

func (f *fakeProgressGit) CloneProgress(ctx context.Context, url, dir string, emit func(frac float64, phase string)) error {
	emit(0.5, "Receiving objects")
	emit(1.0, "Resolving deltas")
	return f.fakeGit.Clone(ctx, url, dir)
}

func TestCloneRepoProgressEmits(t *testing.T) {
	cfg := &config.Config{BaseDir: t.TempDir(), ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}"}
	fg := &fakeProgressGit{fakeGit: &fakeGit{}}
	eng := NewEngine(fg, cfg)
	p := &config.Provider{Name: "gh", Type: config.ProviderGitHub, Short: "gh", CloneProtocol: "ssh"}
	r := provider.Repo{Owner: "me", Name: "proj", SSHURL: "git@github.com:me/proj.git"}

	var fracs []float64
	res := eng.CloneRepoProgress(context.Background(), p, r, func(frac float64, phase string) {
		fracs = append(fracs, frac)
	})
	if res.Err != nil || res.Action != ActionCloned {
		t.Fatalf("CloneRepoProgress: action=%v err=%v", res.Action, res.Err)
	}
	if len(fracs) != 2 || fracs[0] != 0.5 || fracs[1] != 1.0 {
		t.Fatalf("emitted fracs = %v, want [0.5 1.0]", fracs)
	}
}

func TestCloneRepoProgressFallsBackWithoutProgressGit(t *testing.T) {
	cfg := &config.Config{BaseDir: t.TempDir(), ClonePatternTpl: "{{.BaseDir}}/{{.Owner}}/{{.Repo}}"}
	fg := &fakeGit{}
	eng := NewEngine(fg, cfg)
	p := &config.Provider{Name: "gh", Type: config.ProviderGitHub, Short: "gh", CloneProtocol: "ssh"}
	r := provider.Repo{Owner: "me", Name: "proj", SSHURL: "git@github.com:me/proj.git"}

	called := false
	res := eng.CloneRepoProgress(context.Background(), p, r, func(float64, string) { called = true })
	if res.Err != nil || res.Action != ActionCloned {
		t.Fatalf("fallback clone: action=%v err=%v", res.Action, res.Err)
	}
	if called {
		t.Fatal("plain Git should not emit progress")
	}
	if len(fg.cloned) != 1 {
		t.Fatalf("expected one plain Clone, got %v", fg.cloned)
	}
}
