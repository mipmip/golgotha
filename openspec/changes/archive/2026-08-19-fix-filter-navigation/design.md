## Context

`handleKey` (internal/tui/update.go) has an early `if m.filtering` block. Today it
handles Enter/Esc/Ctrl+C, then feeds every other key to the `bubbles/textinput`
and unconditionally sets `m.cursor = 0; m.offset = 0`. Consequences:

- Up/Down reach the single-line text input (caret only) and then the cursor is
  reset to 0 — the selection is pinned to the top while typing.
- Enter merely exits edit mode (keeping the query); a second Enter drills. This
  contradicts the archived spec "Enter drills the filtered item at its level."

Normal-mode navigation already exists (`moveCursor`, `pageStep`) and the
level Enter action is the switch handled after the filtering block.

## Goals / Non-Goals

**Goals:**

- Move/page the selection while the filter input is focused (fzf-style).
- Reset the selection to the top only when the query text changes.
- One Enter acts on the highlighted filtered item.

**Non-Goals:**

- Changing non-filtering key bindings or the filter's scoping/matching.
- Any config, schema, or non-TUI change.

## Decisions

### Decision: Intercept navigation keys before the text input

Inside the `m.filtering` block, add cases that route to the existing movement
helpers and return before `filter.Update`:

- `KeyUp` → `moveCursor(-1)`, `KeyDown` → `moveCursor(1)`
- `KeyPgUp` → `moveCursor(-pageStep())`, `KeyPgDown` → `moveCursor(pageStep())`
- `KeyCtrlP` → `moveCursor(-1)`, `KeyCtrlN` → `moveCursor(1)`

`j`/`k` are printable and fall through to the input (typing). `Home`/`End` and
`Ctrl+U` are left to the input for query editing (caret/kill-line), so half-page
nav is not wired during filtering.

### Decision: Reset the cursor only when the query changes

Replace the unconditional `cursor = 0` with a compare around `filter.Update`:

```
old := m.filter.Value()
m.filter, cmd = m.filter.Update(msg)
if m.filter.Value() != old {
    m.cursor = 0
    m.offset = 0
}
```

Navigation keys never reach this path (they returned early), so the highlight is
preserved while navigating and only snaps to the top when the query text changes.

### Decision: One Enter acts, then leaves filter mode

On `KeyEnter` while filtering: blur the input (keep the query so the list stays
narrowed), set `m.filtering = false`, and delegate to the same level action the
non-filtering Enter runs (the existing select/drill switch). This performs the
drill/open/switch in one press and matches the spec. Drilling clears the filter
via the existing "filter clears on level change" behavior.

- **Alternative — keep two-step Enter:** smallest change but leaves the
  spec/behavior mismatch and a clunky picker. Rejected.

## Risks / Trade-offs

- **[Ctrl+N/Ctrl+P collide with input bindings]** → some text inputs bind these.
  Mitigation: the project's filter input does not use them for editing; binding
  them to navigation is safe and matches fzf/emacs expectations.
- **[Enter-acts changes a tested behavior]** → `TestFilterEnterDrillsFilteredOwner`
  expects two Enters. Mitigation: update it to one Enter (the intended behavior).
- **[Cursor out of range after re-filter]** → resetting to 0 on query change and
  `moveCursor`'s clamping keep the cursor valid as the filtered list shrinks.

## Migration Plan

1. Add the navigation cases and the query-changed guard to the filtering block.
2. Make Enter blur + delegate to the level action.
3. Update `TestFilterEnterDrillsFilteredOwner`; add navigation-while-filtering
   tests.
4. `gofmt` clean; `go test ./...` and `nix flake check` pass.

## Open Questions

- None.
