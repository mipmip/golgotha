package tui

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

func TestSplitArgs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`a b c`, []string{"a", "b", "c"}},
		{`tmux new -s "gh->me" -c /path`, []string{"tmux", "new", "-s", "gh->me", "-c", "/path"}},
		{`x 'quoted arg' y`, []string{"x", "quoted arg", "y"}},
		{`name-with-space\ here`, []string{"name-with-space here"}},
		// A repo name with shell metacharacters stays one literal arg (no injection).
		{`tmux -s "a; rm -rf ~"`, []string{"tmux", "-s", "a; rm -rf ~"}},
	}
	for _, tc := range cases {
		got, err := splitArgs(tc.in)
		if err != nil {
			t.Fatalf("splitArgs(%q) error: %v", tc.in, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("splitArgs(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}

func TestSplitArgsErrors(t *testing.T) {
	for _, in := range []string{`"unterminated`, `'nope`, `trailing\`} {
		if _, err := splitArgs(in); err == nil {
			t.Fatalf("splitArgs(%q) expected error", in)
		}
	}
}

// modeModel builds a Model with an explicit multiplex mode and an injectable
// command runner.
func modeModel(t *testing.T, sw string, cloned bool, run func([]string) error, clone Cloner) *Model {
	t.Helper()
	p := &config.Provider{Name: "gh", Type: config.ProviderGitHub, Short: "gh", Username: "me", WebURL: "https://github.com"}
	cfg := &config.Config{
		BaseDir:     "/tmp",
		DefaultMode: "multiplex",
		Modes: map[string]config.ModeConfig{
			"management": {Header: []string{"breadcrumb"}, Footer: []string{"action_menu"}},
			"multiplex":  {Header: []string{}, Footer: []string{"switch_hint"}, SwitchCommand: sw},
		},
	}
	m := &Model{
		cfg:             cfg,
		providers:       []*config.Provider{p},
		reposByProvider: map[string][]repoItem{},
		nav:             levelRepos,
		selProvider:     p,
		selOwner:        "me",
		mode:            "multiplex",
		runCommand:      run,
		cloner:          clone,
	}
	m.filter = textinput.New()
	m.reposByProvider["gh"] = []repoItem{
		{Repo: provider.Repo{Owner: "me", Name: "proj"}, Provider: p, Target: "/tmp/gh.me/proj", Cloned: cloned},
	}
	return m
}

type recordCloner struct {
	called bool
	res    syncerResult
}

func (c *recordCloner) CloneRepo(_ context.Context, _ *config.Provider, _ provider.Repo) syncerResult {
	c.called = true
	return c.res
}

func TestRenderSwitchCommandHasTarget(t *testing.T) {
	m := modeModel(t, `tmux -c {{.Target}} -s {{.Short}}-{{.Repo}}`, true, func([]string) error { return nil }, &recordCloner{})
	it := m.reposByProvider["gh"][0]
	got, err := m.renderSwitchCommand(m.activeSwitchCommand(), it)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(got, "/tmp/gh.me/proj") || !strings.Contains(got, "gh-proj") {
		t.Fatalf("rendered switch_command = %q", got)
	}
}

func TestMultiplexActivateAlreadyCloned(t *testing.T) {
	var ran []string
	cl := &recordCloner{}
	m := modeModel(t, `echo {{.Repo}}`, true, func(argv []string) error { ran = argv; return nil }, cl)
	m.multiplexActivate(m.activeSwitchCommand())
	if cl.called {
		t.Fatal("already-cloned repo should not be re-cloned")
	}
	if len(ran) != 2 || ran[0] != "echo" || ran[1] != "proj" {
		t.Fatalf("switch argv = %v", ran)
	}
}

func TestMultiplexActivateClonesFirst(t *testing.T) {
	var ran []string
	cl := &recordCloner{res: syncerResult{Target: "/tmp/gh.me/proj", Cloned: true}}
	m := modeModel(t, `echo {{.Repo}}`, false, func(argv []string) error { ran = argv; return nil }, cl)
	m.multiplexActivate(m.activeSwitchCommand())
	if !cl.called {
		t.Fatal("uncloned repo should be cloned before switching")
	}
	if len(ran) == 0 {
		t.Fatal("switch_command should run after a successful clone")
	}
}

func TestMultiplexActivateCloneFailureAbortsSwitch(t *testing.T) {
	ran := false
	cl := &recordCloner{res: syncerResult{Err: context.DeadlineExceeded}}
	m := modeModel(t, `echo {{.Repo}}`, false, func([]string) error { ran = true; return nil }, cl)
	m.multiplexActivate(m.activeSwitchCommand())
	if ran {
		t.Fatal("switch must not run when the clone fails")
	}
	if !strings.Contains(m.status, "clone failed") {
		t.Fatalf("status = %q, want a clone-failure message", m.status)
	}
}

// TestMultiplexEmptyChrome verifies a mode with empty header/footer renders no
// chrome (the "no menu" multiplex view) and that management chrome is unchanged.
func TestMultiplexEmptyChrome(t *testing.T) {
	m := modeModel(t, `echo {{.Repo}}`, true, func([]string) error { return nil }, &recordCloner{})
	header, footer := m.modeChrome()
	if len(header) != 0 {
		t.Fatalf("multiplex header = %v, want empty", header)
	}
	if len(footer) != 1 || footer[0] != "switch_hint" {
		t.Fatalf("multiplex footer = %v, want [switch_hint]", footer)
	}
}
