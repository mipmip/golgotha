## Context

The cache, provider clients, clone-path engine and (from milestone 03) the
clone operation exist. This change builds the interactive front end.

## Goals / Non-Goals

**Goals:**
- A responsive Bubble Tea UI to browse, filter, clone and open repos.
- Testable update logic.

**Non-Goals:**
- New sync semantics — cloning reuses the milestone 03 engine.
- Editing configuration from the UI.

## Decisions

- **Framework**: `charmbracelet/bubbletea` with `bubbles` (list, textinput) and
  `lipgloss` for styling — the stack named in BRIEFING.
- **Model**: a single root model with a navigation stack (provider → owner →
  repos) and a filter mode; repositories come from the cache loaded at start
  and after refresh.
- **Fuzzy filter**: filter over a flattened repo view; use the `bubbles` list
  filtering or a small fuzzy match helper.
- **Actions as commands**: clone/open/refresh run as Bubble Tea `tea.Cmd`s
  returning result messages, so the UI stays responsive and the update function
  is unit-testable without a terminal.
- **Testing**: drive `Update` with synthetic messages/key events and assert on
  model state and emitted commands; no real TTY.

## Risks / Trade-offs

- [Bubble Tea deps enlarge the closure and change vendorHash] → update
  `flake.nix` vendorHash in this change; `nix flake check` reports the value.
- [Terminal UIs are hard to fully test] → keep side effects in commands and
  test the pure update logic; smoke-test model construction.
