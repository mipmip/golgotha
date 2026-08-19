## Why

While the fuzzy filter is being typed, you can't move the selected row
(bean `skull2-x4ej`). In `handleKey`, every key except Enter/Esc/Ctrl+C is fed to
the text input and `m.cursor`/`m.offset` are reset to 0 on every keystroke, so
arrows do nothing and the highlight is pinned to row 0. You can only navigate
after pressing Enter to leave edit mode — the opposite of the expected fzf-style
"type and move at the same time." The archived spec even says *"Enter drills the
filtered item at its level,"* yet the code needs two Enters.

## What Changes

- **Live navigation while filtering.** Route Up/Down (±1 row), PgUp/PgDn (±page),
  and Ctrl+N/Ctrl+P (down/up) to cursor movement while the filter input is
  focused. `j`/`k` keep typing (they're letters); Home/End/Ctrl+U stay with the
  text input for editing the query.
- **Reset selection only on query change.** The cursor snaps back to the top only
  when the filter text actually changes; plain navigation preserves the highlight.
- **One Enter acts (fzf-style).** While filtering, Enter blurs the input (keeping
  the query) and immediately performs the current level's normal action on the
  highlighted item (drill into an owner/provider, or open/switch a repo) in a
  single press — matching the existing spec requirement. Drilling clears the
  filter as today.

`tui`-only; no config, schema, or gate impact.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `tui`: the selection can be moved (and paged) while the fuzzy filter is being
  typed, and Enter acts on the highlighted filtered item in one press.

## Impact

- **TUI**: `internal/tui/update.go` — the `m.filtering` key block gains navigation
  routing (arrows/page/Ctrl+N/P), a query-changed guard around the cursor reset,
  and an Enter path that blurs then delegates to the normal level action.
- **Tests**: new coverage for moving/paging the selection while filtering and for
  cursor-preservation vs reset-on-query-change; `TestFilterEnterDrillsFilteredOwner`
  updated from two Enters to one.
- Independent of the in-flight `refine-multiplex-mode` (different code paths). No
  new dependencies.
