package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/skull2/internal/provider"
)

func TestWindow(t *testing.T) {
	tests := []struct {
		name                   string
		cursor, offset, height int
		chrome, n              int
		wantOffset, wantFirst  int
		wantLast               int
	}{
		{
			name: "sentinel unknown height renders all",
			// height <= 0 -> no constraint
			cursor: 5, offset: 0, height: 0, chrome: 5, n: 20,
			wantOffset: 0, wantFirst: 0, wantLast: 20,
		},
		{
			name:   "negative height sentinel",
			cursor: 10, offset: 3, height: -1, chrome: 4, n: 30,
			wantOffset: 0, wantFirst: 0, wantLast: 30,
		},
		{
			name: "top of list no scroll",
			// visible = 10-5 = 5, margin capped to (5-1)/2=2
			cursor: 0, offset: 0, height: 10, chrome: 5, n: 20,
			wantOffset: 0, wantFirst: 0, wantLast: 5,
		},
		{
			name:   "list shorter than window shows all",
			cursor: 1, offset: 0, height: 20, chrome: 5, n: 3, // visible=15
			wantOffset: 0, wantFirst: 0, wantLast: 3,
		},
		{
			name: "middle of list scrolls to keep margin",
			// visible=5, margin=2, cursor=10 -> offset = 10+2-5+1 = 8
			cursor: 10, offset: 0, height: 10, chrome: 5, n: 20,
			wantOffset: 8, wantFirst: 8, wantLast: 13,
		},
		{
			name: "near end clamps offset",
			// visible=5, n=20 -> maxOffset=15. cursor=19 wants 19+2-5+1=17 -> clamp 15
			cursor: 19, offset: 0, height: 10, chrome: 5, n: 20,
			wantOffset: 15, wantFirst: 15, wantLast: 20,
		},
		{
			name:   "exact end",
			cursor: 19, offset: 15, height: 10, chrome: 5, n: 20,
			wantOffset: 15, wantFirst: 15, wantLast: 20,
		},
		{
			name: "scroll up past top",
			// cursor moved above offset; visible=5 margin=2
			// cursor=3 offset=10 -> cursor-margin=1 < 10 -> offset=1
			cursor: 3, offset: 10, height: 10, chrome: 5, n: 20,
			wantOffset: 1, wantFirst: 1, wantLast: 6,
		},
		{
			name: "tiny height visible clamps to 1",
			// height-chrome = 3-5 = -2 -> visible=1, margin capped to 0
			cursor: 7, offset: 0, height: 3, chrome: 5, n: 20,
			wantOffset: 7, wantFirst: 7, wantLast: 8,
		},
		{
			name: "margin behavior mid list stable window",
			// visible=5 margin=2; cursor=6 offset=4: cursor-margin=4 not <4;
			// cursor+margin=8 >= 4+5=9? no. stays offset 4
			cursor: 6, offset: 4, height: 10, chrome: 5, n: 20,
			wantOffset: 4, wantFirst: 4, wantLast: 9,
		},
		{
			name: "margin cap on short terminal visible 3",
			// height=8 chrome=5 -> visible=3, margin capped to (3-1)/2=1
			// cursor=10 -> offset=10+1-3+1=9
			cursor: 10, offset: 0, height: 8, chrome: 5, n: 20,
			wantOffset: 9, wantFirst: 9, wantLast: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOffset, gotFirst, gotLast := window(tt.cursor, tt.offset, tt.height, tt.chrome, tt.n)
			if gotOffset != tt.wantOffset || gotFirst != tt.wantFirst || gotLast != tt.wantLast {
				t.Fatalf("window(%d,%d,%d,%d,%d) = (%d,%d,%d), want (%d,%d,%d)",
					tt.cursor, tt.offset, tt.height, tt.chrome, tt.n,
					gotOffset, gotFirst, gotLast,
					tt.wantOffset, tt.wantFirst, tt.wantLast)
			}
			// Invariant: cursor must be within the returned slice (non-sentinel).
			if tt.height > 0 && tt.n > 0 {
				if tt.cursor < gotFirst || tt.cursor >= gotLast {
					t.Fatalf("cursor %d not in visible slice [%d,%d)", tt.cursor, gotFirst, gotLast)
				}
			}
		})
	}
}

func TestIndicatorText(t *testing.T) {
	if got := indicator(0, 5, 20); got != "1-5 of 20" {
		t.Fatalf("got %q", got)
	}
	if got := indicator(40, 60, 213); got != "41-60 of 213" {
		t.Fatalf("got %q", got)
	}
	if got := indicator(0, 3, 3); got != "1-3 of 3" {
		t.Fatalf("got %q", got)
	}
}

// newBigProviderModel builds a model with a single provider holding n repos under
// one owner, drilled into the repos level, sized to Height with the given filter
// state. Used for update-driven windowing tests.
func newBigProviderModel(t *testing.T, n int) *Model {
	t.Helper()
	m, _ := newTestModel(t)
	p := m.providers[0] // github
	items := make([]repoItem, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, repoItem{
			Repo:     provider.Repo{Owner: "big", Name: nameFor(i)},
			Provider: p,
		})
	}
	m.reposByProvider["github"] = items
	m.reposByProvider["codeberg"] = nil
	m.nav = levelRepos
	m.selProvider = p
	m.selOwner = "big"
	return m
}

func nameFor(i int) string {
	// zero-padded so sort order matches insertion order for assertions.
	s := "0000" + itoa(i)
	return "repo-" + s[len(s)-4:]
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

func TestUpdateWindowingDownPastFold(t *testing.T) {
	m := newBigProviderModel(t, 50)
	send(m, tea.WindowSizeMsg{Width: 80, Height: 10})
	// chrome at repos level, no filter/status: 2 + 0 + 0 + 1(indicator) + 1(footer) = 4
	// visible = 10 - 4 = 6, margin = 2
	if got := m.visibleRows(); got != 6 {
		t.Fatalf("visibleRows = %d, want 6", got)
	}
	// Render once to establish offset at top.
	_ = m.View()
	if m.offset != 0 {
		t.Fatalf("offset should start 0, got %d", m.offset)
	}
	// Move down 5 times -> cursor 5. cursor+margin=7 >= offset+visible? 0+6=6, 7>=6 yes
	for i := 0; i < 5; i++ {
		send(m, key("j"))
	}
	_ = m.View()
	if m.cursor != 5 {
		t.Fatalf("cursor = %d, want 5", m.cursor)
	}
	// offset = cursor+margin-visible+1 = 5+2-6+1 = 2
	if m.offset != 2 {
		t.Fatalf("offset = %d, want 2", m.offset)
	}
	if !strings.Contains(m.indicatorText, "of 50") {
		t.Fatalf("indicator missing total: %q", m.indicatorText)
	}
}

func TestUpdateEndHomePgDnPgUpCtrlDU(t *testing.T) {
	m := newBigProviderModel(t, 50)
	send(m, tea.WindowSizeMsg{Width: 80, Height: 10}) // visible = 6

	// End -> cursor 49, offset clamps to 44 (n-visible=44).
	send(m, key("end"))
	_ = m.View()
	if m.cursor != 49 {
		t.Fatalf("End cursor = %d, want 49", m.cursor)
	}
	if m.offset != 44 {
		t.Fatalf("End offset = %d, want 44", m.offset)
	}
	if m.indicatorText != "45-50 of 50" {
		t.Fatalf("End indicator = %q, want 45-50 of 50", m.indicatorText)
	}

	// Home -> cursor 0, offset 0.
	send(m, key("home"))
	_ = m.View()
	if m.cursor != 0 || m.offset != 0 {
		t.Fatalf("Home cursor/offset = %d/%d, want 0/0", m.cursor, m.offset)
	}
	if m.indicatorText != "1-6 of 50" {
		t.Fatalf("Home indicator = %q, want 1-6 of 50", m.indicatorText)
	}

	// PgDn -> cursor += visible(6) => 6.
	send(m, tea.KeyMsg{Type: tea.KeyPgDown})
	_ = m.View()
	if m.cursor != 6 {
		t.Fatalf("PgDn cursor = %d, want 6", m.cursor)
	}

	// PgUp -> cursor -= 6 => 0.
	send(m, tea.KeyMsg{Type: tea.KeyPgUp})
	_ = m.View()
	if m.cursor != 0 {
		t.Fatalf("PgUp cursor = %d, want 0", m.cursor)
	}

	// Ctrl+D -> cursor += visible/2 = 3.
	send(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	_ = m.View()
	if m.cursor != 3 {
		t.Fatalf("Ctrl+D cursor = %d, want 3", m.cursor)
	}

	// Ctrl+U -> cursor -= 3 => 0.
	send(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	_ = m.View()
	if m.cursor != 0 {
		t.Fatalf("Ctrl+U cursor = %d, want 0", m.cursor)
	}
}

func TestUpdateOffsetResetsOnDrill(t *testing.T) {
	m := newBigProviderModel(t, 50)
	send(m, tea.WindowSizeMsg{Width: 80, Height: 10})
	send(m, key("end"))
	_ = m.View()
	if m.offset == 0 {
		t.Fatalf("expected non-zero offset after End")
	}
	// Go back to owners: offset must reset.
	send(m, key("esc"))
	if m.offset != 0 || m.cursor != 0 {
		t.Fatalf("expected offset/cursor reset on back, got %d/%d", m.offset, m.cursor)
	}
}

func TestSentinelRendersAllRows(t *testing.T) {
	m := newBigProviderModel(t, 50)
	// No WindowSizeMsg sent -> height 0 -> sentinel.
	out := m.View()
	// All 50 repo names should be present.
	for _, want := range []string{nameFor(0), nameFor(25), nameFor(49)} {
		if !strings.Contains(out, want) {
			t.Fatalf("sentinel view missing %q", want)
		}
	}
	if m.offset != 0 {
		t.Fatalf("sentinel offset should be 0, got %d", m.offset)
	}
	if m.indicatorText != "1-50 of 50" {
		t.Fatalf("sentinel indicator = %q, want 1-50 of 50", m.indicatorText)
	}
}

func TestVisibleSliceInView(t *testing.T) {
	m := newBigProviderModel(t, 50)
	send(m, tea.WindowSizeMsg{Width: 80, Height: 10}) // visible 6
	send(m, key("end"))
	out := m.View()
	// Window should show rows 44..49 only; row 0 must be absent.
	if strings.Contains(out, nameFor(0)) {
		t.Fatalf("expected windowed view to exclude first row")
	}
	if !strings.Contains(out, nameFor(49)) {
		t.Fatalf("expected windowed view to include last row")
	}
}
