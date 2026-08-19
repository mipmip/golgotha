## Context

`add-tui-modes` shipped a working multiplex mode: Enter clones-if-needed then
runs a shell-safe `switch_command`. `add-combined-repo-view` shipped a flat
cross-provider list reachable from an "All repositories" entry. Bean
`skull2-9ftr` asks for four refinements that make the tmux quick-jump flow clean.

Current gaps in the shipped code:
- `renderRepos` always draws the `[ ]`/`[x]` checkbox; `space` toggles selection.
- `multiplexActivate` clones **synchronously** (blocking) with only a status line
  and, after a switch, returns `(m, nil)` — the TUI lingers.
- The `Cloner`/`syncer` clone is synchronous and emits no progress.
- The TUI always starts at `levelProviders`.

## Goals / Non-Goals

**Goals:**

- Multiplex: no checkbox, inert selection toggle.
- Exit 0 after a successful switch.
- Async clone with a cancellable centered progress popup and a determinate bar
  fed by real git progress.
- `--flatlist` to start in the combined flat view; composes with `--mode`.

**Non-Goals:**

- Non-zero exit codes on failure (keep the TUI open with an error instead).
- Progress for pull/ff-update (only clone).
- Changing management-mode behavior, the combined view, or the config schema.

## Decisions

### Decision: Multiselect is gated on multiplex, not the flat view

`renderRepos` omits the checkbox and `toggleSelect` no-ops when
`m.activeSwitchCommand() != ""` (multiplex). The combined flat view under
management keeps bulk-clone, so the gate is the mode, not `flatAll`.

### Decision: Exit 0 after a successful switch

`multiplexActivate` (async, see below) sets `m.quitting = true` and returns
`tea.Quit` after `runSwitch` succeeds. Failures set a status message and stay.
`tea.Quit` exits 0; carrying a non-zero code would require threading an error out
through `Run`, which is out of scope.

### Decision: Async clone flow with message round-trips

```
Enter (multiplex)
  ├─ already cloned ─▶ runSwitch ─▶ tea.Quit
  └─ not cloned ─▶ start cloneProgress cmd, show popup (m.cloning = true)
                    ├─ cloneProgressMsg{percent,phase} ─▶ update bar
                    ├─ cloneDoneMsg{ok}  ─▶ runSwitch ─▶ tea.Quit
                    ├─ cloneDoneMsg{err} ─▶ m.cloning=false, status=error, stay
                    └─ Esc ─▶ cancel ctx, m.cloning=false, stay
```

Mirrors the existing lazy per-owner fetch (spinner/bar + cancel), so the popup
reuses the `bubbles/progress` bar and a cancel context.

### Decision: Progress-emitting clone in the syncer (parse `git clone --progress`)

Add a syncer clone variant that runs `git clone --progress <url> <dir>` and
streams stderr. Git writes progress with `\r`, phases like `Counting objects`,
`Compressing objects`, `Receiving objects: N% (x/y)`, `Resolving deltas`. A
scanner splits on `\r`/`\n`, extracts the trailing `NN%` and phase, and emits
`(percent, phase)` through an emit callback (same shape as `fetch.Emit`). The
existing synchronous `CloneRepo` is kept for `hup sync`; the new one is used by
the TUI. Context cancellation kills the git process.

- **Alternative — indeterminate spinner (no syncer change):** simpler but no real
  percentage. Rejected: the bean explicitly wants a determinate bar.

### Decision: Centered modal overlay via line compositing

Render the normal view, then splice a bordered lipgloss box (the clone popup)
into the center of that background by overwriting the middle lines (dimming the
rest), using `m.width`/`m.height`. This gives a true floating popup rather than a
full-screen replacement.

- **Alternative — full-screen progress view (like `fetchProgressView`):** less
  code but not a "popup". Rejected per the chosen modal style.

### Decision: `--flatlist` sets initial state after `New`

`hup tui --flatlist` parses a bool flag; `Run` sets `flatAll = true`,
`nav = levelRepos` (and clears provider/owner selection) before starting. Caches
already load in `New`, so the flat list has data. Orthogonal to `--mode`.

## Risks / Trade-offs

- **[git progress format drift]** → parsing `git clone --progress` depends on
  git's human output. Mitigation: parse defensively (any `NN%` on a progress
  line), and treat unparseable lines as "no update"; the clone still completes
  and reports its terminal result even if the bar is coarse.
- **[Overlay compositing correctness]** → manual line splicing with ANSI/width is
  fiddly (wide runes, escape codes). Mitigation: keep the popup ASCII-simple,
  size it from `m.width`, and unit-test the composite for line count/placement.
- **[Blocking removed]** → async clone changes the multiplex control flow.
  Mitigation: mirror the tested lazy-fetch pattern (cmd + msgs + cancel) and add
  flow tests with a fake progress-clone.
- **[Quit-on-switch surprises interactive use]** → outside tmux, exiting after a
  switch may be unexpected. Accepted: it is multiplex-mode-only and matches the
  quick-jump intent.

## Migration Plan

1. Add the progress-emitting clone to `internal/syncer` (keep `CloneRepo`).
2. Gate the checkbox/selection on multiplex; make `multiplexActivate` async with
   the popup and exit-on-switch.
3. Add the centered-modal overlay renderer with the determinate bar + Esc cancel.
4. Add `--flatlist` to `hup tui` and the initial-state wiring in `Run`.
5. `gofmt` clean; `nix flake check` passes (coverage gate).

## Open Questions

- Should the popup show the phase label (e.g. "Receiving objects") alongside the
  bar, or just the percentage? Leaning: show phase + percent. Resolve in
  implementation.
