## Context

`internal/tui` renders lists by looping over every row in `view.go`
(`renderStrings`, `renderRepos`); `m.width`/`m.height` are stored from
`tea.WindowSizeMsg` but never used. The program already runs with
`tea.WithAltScreen()`. This change adds a scrolling viewport without disturbing
the existing multi-select, cloned indicator, hierarchic navigation or fuzzy
filter (approach B: hand-rolled, not a migration to `bubbles/list`).

## Goals / Non-Goals

**Goals:**
- Keep the cursor and breadcrumb visible for lists of any length.
- Pure, unit-testable windowing math (no TTY needed).
- Reuse existing rendering/selection logic.

**Non-Goals:**
- Migrating to `bubbles/list` (tracked separately as a possible future bean).
- Horizontal scrolling or column layouts.
- A configurable scroll-off (fixed at 2 for now).

## Decisions

- **State**: add a single `offset int` (top visible row for the current level).
  Reset to 0 wherever the cursor already resets (drill in/out, filter keystroke,
  refresh).
- **Visible budget**: `visible = height - chrome`, where `chrome` is recomputed
  per frame from actual state (header 2 lines + filter line when filtering/has
  value + status line when set + footer 1). Never hardcode chrome.
- **Sentinel for unknown size**: if `height <= 0`, render all rows (matches
  today's behavior before the first `WindowSizeMsg` and in tests). Alt-screen
  makes any first-frame over-render harmless (off-screen, overwritten).
- **Scroll rule** (single function, fed by every movement key):
  ```
  if cursor - margin <  offset            → offset = cursor - margin
  if cursor + margin >= offset + visible  → offset = cursor + margin - visible + 1
  offset = clamp(offset, 0, max(0, n - visible))
  ```
- **Margin**: fixed 2, capped to `(visible-1)/2` for short terminals.
- **Keys**: `PgUp`/`PgDn` move the cursor by `visible`; `Ctrl-U`/`Ctrl-D` by
  `visible/2`; `Home`/`End` to `0`/`n-1`. All then call the scroll rule +
  `clampCursor`. `up`/`down` unchanged.
- **Indicator**: render `first-last of n` (1-based) from `offset`, `visible`, `n`.
- **Testing**: exercise the pure windowing function directly, and drive
  `Update` with a `tea.WindowSizeMsg` followed by movement keys, asserting
  `offset`/visible slice. Existing View tests (no size sent) keep seeing all
  rows via the sentinel.

## Risks / Trade-offs

- [Chrome height is variable] → recompute per frame from real state; a wrong
  budget clips the last row or flickers.
- [Tiny terminals] → guard `visible >= 1` and cap the margin.
- [Off-by-one in range/indicator] → covered by table-driven tests on the pure
  function.
