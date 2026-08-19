## 1. Config: modes model + validation

- [x] 1.1 Add `Config.DefaultMode` and `Config.Modes` (map of mode name → chrome + settings) with a `ModeConfig{ Header, Footer []string, SwitchCommand string }`, distinguishing unset from empty
- [x] 1.2 Provide a built-in default `management` mode when `modes:` is omitted (standard header/footer), so existing configs keep working
- [x] 1.3 Validate: `default_mode` names a known mode; header/footer element names are known; no element appears twice within a mode; `switch_command` required for multiplex
- [x] 1.4 Config unit tests: default when omitted, unknown mode/element, duplicate element, missing multiplex `switch_command`

## 2. TUI: mode selection + element-registry chrome

- [x] 2.1 Add a `--mode <name>` flag to the TUI entrypoint (`cmd/hup`); resolve active mode as flag > `default_mode` > built-in management; error on unknown mode
- [x] 2.2 Introduce an element registry `map[string]func(*Model) string` (breadcrumb, action_menu, filter, facet_status, status_message, position_indicator, switch_hint, clone_status), each returning "" when inactive
- [x] 2.3 Refactor `view.go` to render header/footer from the active mode's element lists (in order, skipping empty); body between them
- [x] 2.4 Make `chrome()`/`viewport.go` derive height from rendered header+footer lines (wrap-aware)
- [x] 2.5 Refactor the current TUI behavior into the `management` mode (Enter → detail); no behavior change in that mode

## 3. Multiplex mode: primary action + switch_command

- [x] 3.1 Model modes as a strategy: per-mode primary action keyed by active mode (management → detail; multiplex → clone-then-switch)
- [x] 3.2 Add `Target` (resolved local clone path via `clonepath.Render`) to the `switch_command` template context alongside the clone-path fields
- [x] 3.3 Render `switch_command`, split into argv via shell-quoting rules, and execute without a shell; surface non-zero exit as an error
- [x] 3.4 Multiplex activate flow: clone-if-needed via existing `Cloner` (with progress), abort switch on clone failure, then run the command
- [x] 3.5 Add multiplex-dedicated elements (`switch_hint`, `clone_status`) to the registry

## 4. Docs, example & verification

- [x] 4.1 Update `config.example.yaml` with `default_mode` + `modes:` (management + a multiplex example), keeping it valid + complete for the config-example gate
- [x] 4.2 Update `BRIEFING.md` and `openspec/config.yaml` `context:` for modes + `switch_command`
- [x] 4.3 TUI unit tests: mode selection/flag, per-mode chrome (incl. empty multiplex), clone-then-switch flow, shell-safe arg splitting (metacharacter names), clone-failure aborts switch
- [x] 4.4 If a shell-words dependency is added, update `flake.nix` `vendorHash`; `gofmt -l .` empty; `nix flake check` passes
- [x] 4.5 Keep the `skull2-wzbf` beans checklist current as tasks complete
