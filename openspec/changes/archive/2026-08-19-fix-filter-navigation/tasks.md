## 1. Navigation while filtering

- [x] 1.1 In the `m.filtering` block of `handleKey`, add navigation cases before `filter.Update`: Up/Down → `moveCursor(∓1)`, PgUp/PgDn → `moveCursor(±pageStep())`, Ctrl+P/Ctrl+N → up/down
- [x] 1.2 Replace the unconditional `cursor = 0` reset with a query-changed guard: capture `filter.Value()` before `filter.Update`, and only reset `cursor`/`offset` to 0 when it changed
- [x] 1.3 Leave `j`/`k`, Home/End, and Ctrl+U to the text input (typing / caret editing)

## 2. One-Enter acts

- [x] 2.1 On Enter while filtering, blur the input (keep the query), set `filtering = false`, and delegate to the normal level action so it drills/opens/switches the highlighted item in one press

## 3. Tests & verification

- [x] 3.1 TUI tests: Up/Down and PgUp/PgDn and Ctrl+N/Ctrl+P move the selection while filtering; the input stays focused and the query is unchanged
- [x] 3.2 TUI tests: typing resets the selection to the top; navigating (no query change) preserves the highlight; `j`/`k` still type
- [x] 3.3 Update `TestFilterEnterDrillsFilteredOwner` to expect a single Enter to drill the filtered owner
- [x] 3.4 `gofmt -l .` empty; `go test ./...` passes; `nix flake check` passes (coverage gate)
- [x] 3.5 Keep the `skull2-x4ej` beans checklist current as tasks complete
