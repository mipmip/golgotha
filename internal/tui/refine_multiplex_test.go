package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestMultiplexHidesCheckbox(t *testing.T) {
	m := modeModel(t, "echo {{.Repo}}", false, func([]string) error { return nil }, &recordCloner{})
	body := m.bodyText()
	if strings.Contains(body, "[ ]") || strings.Contains(body, "[x]") {
		t.Fatalf("multiplex body should have no checkbox:\n%s", body)
	}
	if !strings.Contains(body, "me/proj") {
		t.Fatalf("expected the repo row, got:\n%s", body)
	}
}

func TestManagementShowsCheckbox(t *testing.T) {
	m, _ := newTestModel(t)
	send(m, key("enter")) // owners
	send(m, key("j"))     // mipmip
	send(m, key("enter")) // repos
	body := m.bodyText()
	if !strings.Contains(body, "[ ]") {
		t.Fatalf("management body should show the checkbox:\n%s", body)
	}
}

func TestMultiplexSpaceInert(t *testing.T) {
	m := modeModel(t, "echo {{.Repo}}", false, func([]string) error { return nil }, &recordCloner{})
	send(m, key("space"))
	for _, it := range m.reposByProvider["gh"] {
		if it.Selected {
			t.Fatal("space must not select in multiplex mode")
		}
	}
}

func TestMultiplexClonedSwitchesAndQuits(t *testing.T) {
	var ran []string
	m := modeModel(t, "echo {{.Repo}}", true, func(argv []string) error { ran = argv; return nil }, &recordCloner{})
	_, cmd := m.multiplexActivate(m.activeSwitchCommand())
	if len(ran) == 0 {
		t.Fatal("expected switch to run for an already-cloned repo")
	}
	if !m.quitting || !isQuit(cmd) {
		t.Fatalf("expected quit after switch, quitting=%v isQuit=%v", m.quitting, isQuit(cmd))
	}
}

func TestMultiplexAsyncCloneStartsPopup(t *testing.T) {
	m := modeModel(t, "echo {{.Repo}}", false, func([]string) error { return nil }, &recordCloner{})
	canceled := false
	ch := make(chan cloneEvent, 4)
	m.progressCloner = func(_ context.Context, _ *config.Provider, _ provider.Repo) (<-chan cloneEvent, context.CancelFunc) {
		return ch, func() { canceled = true }
	}
	m.multiplexActivate(m.activeSwitchCommand())
	if !m.cloning {
		t.Fatal("expected clone popup active")
	}
	// Progress event updates the bar and keeps cloning.
	send(m, cloneMsg{ev: cloneEvent{Frac: 0.5, Phase: "Receiving objects"}, ch: ch})
	if !m.cloning || m.cloneFrac != 0.5 {
		t.Fatalf("expected cloning with frac 0.5, got cloning=%v frac=%v", m.cloning, m.cloneFrac)
	}
	// Esc cancels.
	send(m, key("esc"))
	if m.cloning || !canceled {
		t.Fatalf("esc should cancel the clone: cloning=%v canceled=%v", m.cloning, canceled)
	}
}

func TestMultiplexAsyncCloneDoneSwitchesAndQuits(t *testing.T) {
	var ran []string
	m := modeModel(t, "echo {{.Repo}}", false, func(argv []string) error { ran = argv; return nil }, &recordCloner{})
	m.cloning = true
	m.cloneItem = m.reposByProvider["gh"][0]
	m.cloneSwitch = m.activeSwitchCommand()
	ch := make(chan cloneEvent)
	cmd := send(m, cloneMsg{ev: cloneEvent{Done: true}, ch: ch})
	if m.cloning {
		t.Fatal("clone done should end the popup")
	}
	if len(ran) == 0 {
		t.Fatal("switch should run after a successful clone")
	}
	if !m.quitting || !isQuit(cmd) {
		t.Fatalf("expected quit after clone+switch, quitting=%v isQuit=%v", m.quitting, isQuit(cmd))
	}
}

func TestMultiplexAsyncCloneFailureAborts(t *testing.T) {
	ranSwitch := false
	m := modeModel(t, "echo {{.Repo}}", false, func([]string) error { ranSwitch = true; return nil }, &recordCloner{})
	m.cloning = true
	m.cloneItem = m.reposByProvider["gh"][0]
	m.cloneSwitch = m.activeSwitchCommand()
	ch := make(chan cloneEvent)
	send(m, cloneMsg{ev: cloneEvent{Done: true, Err: context.DeadlineExceeded}, ch: ch})
	if m.cloning {
		t.Fatal("failed clone should end the popup")
	}
	if ranSwitch {
		t.Fatal("switch must not run after a failed clone")
	}
	if !strings.Contains(m.status, "clone failed") {
		t.Fatalf("expected clone-failure status, got %q", m.status)
	}
	if m.quitting {
		t.Fatal("a failed clone should not quit")
	}
}

func TestStartFlat(t *testing.T) {
	m, _ := newTestModel(t)
	m.startFlat()
	if !m.flatAll || m.nav != levelRepos || m.selProvider != nil {
		t.Fatalf("startFlat: flatAll=%v nav=%d selProvider=%v", m.flatAll, m.nav, m.selProvider)
	}
}
