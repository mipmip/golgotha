## Why

When the repository (or owner/provider) list is shorter than the viewport, the
footer chrome floats up directly beneath the last row instead of sitting at the
bottom of the terminal (bean `skull2-q0ik`). `View()` concatenates
header → body → footer with no vertical fill, so a short body leaves empty space
*below* the footer. The footer should be pinned to the bottom of the viewport.

## What Changes

- Pad between the body and the footer in `View()` so the footer block is pinned
  to the bottom of the viewport: insert `m.height - usedLines` blank lines before
  the footer when the content is shorter than the terminal height.
- Behavior is unchanged when the list already fills the viewport (pad is zero)
  and when the terminal height is unknown (`m.height == 0`, e.g. before the first
  window-size message) — no padding is applied.

`tui`-only; no config, schema, or gate impact.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `tui`: the footer chrome is pinned to the bottom of the viewport; a short list
  leaves the empty space between the body and the footer rather than below it.

## Impact

- **TUI**: `internal/tui/view.go` — `View()` computes the rendered height of the
  header/body/footer and inserts blank lines before the footer to reach
  `m.height`. `chrome()`/windowing are unaffected (they already reserve the
  footer rows).
- **Tests**: a TUI test asserting the rendered view height equals `m.height` and
  the footer is on the last line for a short list; and that a full list is
  unchanged.
- No new dependencies.
