## 1. Windowing core

- [x] 1.1 Add `offset int` to the model; reset it wherever the cursor resets (drill in/out, filter keystroke, refresh)
- [x] 1.2 Pure `window(cursor, offset, height, chrome, n) -> (offset, first, last)` with margin=2 capped to `(visible-1)/2` and `visible>=1`
- [x] 1.3 Sentinel: `height <= 0` renders all rows

## 2. View integration

- [x] 2.1 Compute `chrome` per frame from actual state (header/filter/status/footer)
- [x] 2.2 Window `renderStrings` and `renderRepos` to the visible slice, preserving cursor/selection/cloned styling
- [x] 2.3 Render the position indicator (`first-last of n`)

## 3. Keys

- [x] 3.1 `PgUp`/`PgDn` (±visible), `Ctrl-U`/`Ctrl-D` (±visible/2), `Home`/`End` (0 / n-1)
- [x] 3.2 Route all movement through the scroll rule + clampCursor; update footer help text

## 4. Tests & verify

- [x] 4.1 Table-driven tests for the pure window function (top, middle, end, tiny height, margin cap, unknown height)
- [x] 4.2 Update-driven tests: WindowSizeMsg + movement keys assert offset/visible slice and indicator
- [x] 4.3 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
