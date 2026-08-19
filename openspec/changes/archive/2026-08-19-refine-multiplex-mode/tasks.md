## 1. Multiplex: hide multiselect

- [x] 1.1 In `renderRepos`, omit the `[ ]`/`[x]` checkbox column when the active mode is multiplex (`activeSwitchCommand() != ""`); keep it in management (incl. the combined view)
- [x] 1.2 Make `toggleSelect` (space) a no-op in multiplex mode
- [x] 1.3 TUI tests: no checkbox + inert space in multiplex; checkbox/toggle still work in management

## 2. Exit 0 after switch

- [x] 2.1 After a successful `runSwitch`, set `m.quitting` and return `tea.Quit` (clean exit 0); failures keep the TUI open with a status error
- [x] 2.2 TUI test: successful switch returns a quit command; failed switch does not

## 3. Progress-emitting clone (syncer)

- [x] 3.1 Add a progress-emitting clone in `internal/syncer` that runs `git clone --progress` and streams stderr, parsing `\r`-delimited lines into `(percent, phase)` events via an emit callback; keep the synchronous `CloneRepo` for `hup sync`
- [x] 3.2 Honor context cancellation by killing the git process; return a terminal result (target or error) after progress events
- [x] 3.3 Syncer unit tests: parse representative `git clone --progress` output (fake) into percentages/phases; cancellation stops promptly

## 4. Async clone flow + centered modal popup

- [x] 4.1 Rewrite `multiplexActivate` as async: uncloned → start progress-clone cmd and show the popup; `cloneProgressMsg` updates the bar; `cloneDoneMsg{ok}` → switch → quit; `cloneDoneMsg{err}` → close popup, status error, stay; already-cloned → switch → quit
- [x] 4.2 Add a cancellable clone context; Esc while the popup is shown cancels the clone, closes the popup, runs no switch, stays open
- [x] 4.3 Add a centered-modal renderer (bordered box with a `bubbles/progress` determinate bar + phase/percent), centered with `lipgloss.Place` using `m.width`/`m.height`
- [x] 4.4 TUI tests: async flow (clone→switch→quit) with a fake progress-clone; progress event updates the bar; clone-failure aborts + stays; Esc cancels

## 5. --flatlist startup flag

- [x] 5.1 Add a `--flatlist` bool flag to `hup tui` (`cmd/hup`); pass it into `tui.Run`
- [x] 5.2 In `Run`, when set, initialize the model into the combined flat view (`flatAll=true`, `nav=levelRepos`, no provider/owner selection); composes with `--mode`
- [x] 5.3 Tests: `startFlat` initializes the combined flat view (flatAll, levelRepos, no selection); back-from-flat is already covered by the combined-view tests

## 6. Verification

- [x] 6.1 `gofmt -l .` empty; `go test ./...` passes; `nix flake check` passes (coverage gate)
- [x] 6.2 Keep the `skull2-9ftr` beans checklist current as tasks complete
