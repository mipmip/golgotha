## Why

HupHop has a single hardcoded TUI. Bean `skull2-wzbf` wants a second way to use
it: a **multiplex mode** for inside tmux that strips the interface down and, on
selecting a repo, ensures it is cloned and then runs a user-defined
`switch_command` (e.g. open/switch to a tmux session at the checkout). This is
BRIEFING goal #2 ("quick navigation"), a milestone-07 later-stretch — and it is
*not* the "tmux config generator" non-goal: we run a user-supplied command, we
don't generate config.

Introducing modes also generalizes the TUI chrome: making the header and footer
configurable per mode subsumes the standalone `configurable-tui-chrome` work
(management becomes just one mode), so this change **supersedes** it.

## What Changes

- **BREAKING (config)**: add a `default_mode` and a `modes:` map to the config.
  Each mode defines its chrome (ordered `header`/`footer` element lists) and any
  mode-specific settings. `--mode <name>` overrides `default_mode` at launch.
- Refactor the current TUI into the built-in **`management`** mode (unchanged
  behavior): Enter on a repo opens the detail view.
- Add the **`multiplex`** mode: minimal chrome; Enter on a repo **clones it if
  needed** (reusing the existing cloner) and then runs the mode's
  `switch_command`.
- **Mode-aware chrome (absorbs `configurable-tui-chrome`)**: header/footer are
  ordered element slots per mode, drawn from a vocabulary
  (`breadcrumb`, `action_menu`, `filter`, `facet_status`, `status_message`,
  `position_indicator`, plus multiplex-dedicated elements such as `switch_hint`
  and `clone_status`). The repository list is always the fixed body. Empty slots
  render no chrome. Window height derives from the rendered chrome.
- **`switch_command`**: a Go text/template rendered per selected repo, then
  **split into argv via shell-words rules and executed without a shell**
  (injection-safe for repo/owner names). An explicit shell form is available for
  commands that genuinely need a shell.
- Extend the command template context with **`Target`** (the resolved local
  clone path), alongside the existing clone-path fields.
- Validate the new config: known mode names, known element names, no duplicate
  element across a mode's header+footer, and a `switch_command` required for
  modes that need one.
- Update `config.example.yaml`, `BRIEFING.md`, and `openspec/config.yaml`.

Out of scope: in-app mode switching (mode is fixed at launch; switch later),
self-account tint (rides `fix-self-owner-resolution`), columns (`skull2-n3i2`),
star sort (`skull2-2h8p`).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `configuration`: add `default_mode` and the `modes:` map (per-mode chrome +
  mode settings), a `switch_command` field, and their validation.
- `tui`: introduce launch-selected modes (`management`, `multiplex`); make the
  header/footer configurable element slots per mode; add the multiplex primary
  action (clone-if-needed then run a templated, shell-safe `switch_command`).

## Impact

- **Config**: `internal/config` — `Config.DefaultMode`, `Config.Modes` (map of
  mode → chrome + settings), validation. **BREAKING**: config gains a modes
  model; a config without `modes:` uses a built-in default management mode so
  existing setups keep working.
- **TUI**: `internal/tui` — mode selection at startup; element-registry chrome
  rendering per mode (as designed for the retired `configurable-tui-chrome`);
  multiplex primary action; command render/exec; dynamic `chrome()`.
- **Template**: expose `Target` to the `switch_command` context (reusing
  `clonepath.Render` to compute the path).
- **CLI**: `cmd/hup` — a `--mode` flag on the TUI entrypoint.
- **Docs/example**: `config.example.yaml` (+ gate `skull2-cqi8` enforces it),
  `BRIEFING.md`, `openspec/config.yaml`.
- **Supersedes**: `configurable-tui-chrome` (retired) and bean `skull2-rkyi`
  (scrapped, folded here).
- **Sequencing**: touches the config schema — sequence **after**
  `fix-self-owner-resolution`; guarded by the config-example gate.
- No new third-party dependency beyond a small shell-words splitter (or a
  vendored helper); update `vendorHash` if one is added.
