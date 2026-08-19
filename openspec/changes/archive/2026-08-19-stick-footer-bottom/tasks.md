## 1. Pin the footer to the bottom

- [x] 1.1 Restructure `View()` to build the header block, body, and footer block separately and compute their line counts with a `textLines` helper (empty block = 0 lines, unlike `lipgloss.Height`)
- [x] 1.2 Insert `m.height - used` blank lines before the footer block when `m.height > 0` and the gap is positive; otherwise render as before (long list → pad 0; unknown height → no pad)

## 2. Tests & verification

- [x] 2.1 TUI test: with a known height and a short list, the rendered view height equals `m.height` and the footer's last line is the viewport's last line
- [x] 2.2 TUI test: a list that fills the viewport is unchanged (no extra blank lines); unknown height (`0`) adds no padding
- [x] 2.3 `gofmt -l .` empty; `go test ./...` passes; `nix flake check` passes (coverage gate)
- [x] 2.4 Keep the `skull2-q0ik` beans checklist current as tasks complete
