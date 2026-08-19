## 1. Multiplex: hide multiselect

- [ ] 1.1 In `renderRepos`, omit the `[ ]`/`[x]` checkbox column when the active mode is multiplex (`activeSwitchCommand() != ""`); keep it in management (incl. the combined view)
- [ ] 1.2 Make `toggleSelect` (space) a no-op in multiplex mode
- [ ] 1.3 TUI tests: no checkbox + inert space in multiplex; checkbox/toggle still work in management

## 2. Exit 0 after switch

- [ ] 2.1 After a successful `runSwitch`, set `m.quitting` and return `tea.Quit` (clean exit 0); failures keep the TUI open with a status error
- [ ] 2.2 TUI test: successful switch returns a quit command; failed switch does not

## 3. Progress-emitting clone (syncer)

- [ ] 3.1 Add a progress-emitting clone in `internal/syncer` that runs `git clone --progress` and streams stderr, parsing `\r`-delimited lines into `(percent, phase)` events via an emit callback; keep the synchronous `CloneRepo` for `hup sync`
- [ ] 3.2 Honor context cancellation by killing the git process; return a terminal result (target or error) after progress events
- [ ] 3.3 Syncer unit tests: parse representative `git clone --progress` output (fake) into percentages/phases; cancellation stops promptly

## 4. Async clone flow + centered modal popup

- [ ] 4.1 Rewrite `multiplexActivate` as async: uncloned → start progress-clone cmd and show the popup; `cloneProgressMsg` updates the bar; `cloneDoneMsg{ok}` → switch → quit; `cloneDoneMsg{err}` → close popup, status error, stay; already-cloned → switch → quit
- [ ] 4.2 Add a cancellable clone context; Esc while the popup is shown cancels the clone, closes the popup, runs no switch, stays open
- [ ] 4.3 Add a centered-modal overlay renderer (bordered box with a `bubbles/progress` determinate bar + phase/percent), composited over the dimmed list using `m.width`/`m.height`
- [ ] 4.4 TUI tests: async flow (clone→switch→quit) with a fake progress-clone; clone-failure aborts + stays; Esc cancels; overlay composite line count/placement

## 5. --flatlist startup flag

- [ ] 5.1 Add a `--flatlist` bool flag to `hup tui` (`cmd/hup`); pass it into `tui.Run`
- [ ] 5.2 In `Run`, when set, initialize the model into the combined flat view (`flatAll=true`, `nav=levelRepos`, no provider/owner selection); composes with `--mode`
- [ ] 5.3 Tests: `--flatlist` starts in the flat view; back returns to the provider list; composes with `--mode multiplex`

## 6. Verification

- [ ] 6.1 `gofmt -l .` empty; `go test ./...` passes; `nix flake check` passes (coverage gate)
- [ ] 6.2 Keep the `skull2-9ftr` beans checklist current as tasks complete
