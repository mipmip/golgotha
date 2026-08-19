## Why

The multiplex mode (shipped in `add-tui-modes`) works but its UX is rough for its
tmux quick-jump purpose (bean `skull2-9ftr`): the multiselect checkbox is
meaningless there, the TUI lingers after a switch, the clone-first step blocks
with only a status line, and there's no way to launch straight into the combined
flat list. These refinements make the "find a repo anywhere → jump to it" flow a
single, clean command.

## What Changes

- **Hide the checkbox / disable multiselect in multiplex mode.** The `[ ]`/`[x]`
  column is omitted and `space` is a no-op when the active mode is multiplex
  (bulk-clone stays available in the combined view under `management`).
- **Exit 0 after a successful switch.** Once `switch_command` runs successfully
  the TUI quits cleanly (ideal for a tmux `display-popup` launch). Clone/switch
  failures keep the TUI open with an error.
- **Clone-first shows a centered modal popup with a determinate progress bar.**
  The clone becomes asynchronous; a centered overlay renders over the dimmed list
  with a real percentage bar, cancellable with Esc. On success the switch runs
  (then exit 0); on failure the popup closes and the error is shown.
- **Progress-emitting clone in the syncer.** A new clone path runs
  `git clone --progress` and parses its stderr (`Receiving objects: N%`) into
  percent/phase events (mirroring the existing fetch-progress pattern), so the
  popup bar reflects real git progress.
- **`--flatlist` startup flag.** `hup tui --flatlist` starts one level deeper, in
  the combined cross-provider flat list, instead of the provider list. Composes
  with `--mode`, so `hup tui --mode multiplex --flatlist` is the full quick-jump.

No config-schema change (no example/gate impact).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `tui`: multiplex hides multiselect; a successful switch exits the program; the
  clone-first step shows a cancellable centered progress popup; the TUI can start
  directly in the combined flat view.
- `sync`: the clone engine can emit real git clone progress (percent/phase) for a
  determinate progress bar.

## Impact

- **Syncer**: `internal/syncer` — a progress-emitting clone that runs
  `git clone --progress` and parses stderr; the existing synchronous `CloneRepo`
  stays for `hup sync`.
- **TUI**: `internal/tui` — `renderRepos` drops the checkbox in multiplex and
  `space` no-ops; `multiplexActivate` becomes async (clone → popup → switch →
  quit); a new centered-modal overlay renderer with a determinate bar and Esc
  cancel; `New`/`Run` accept an initial flat-view state.
- **CLI**: `cmd/hup` + `internal/tui` `Run` — a `--flatlist` flag on `hup tui`.
- **Tests**: syncer progress parsing (fake git output), TUI checkbox/space
  gating, async clone→switch→quit flow, `--flatlist` initial state.
- Builds on the shipped `add-tui-modes` and `add-combined-repo-view`. No new
  third-party dependency.
