## Why

The TUI renders every row of the current list on every frame with no regard for
terminal height. With 200+ repositories the list overruns the screen: the
breadcrumb scrolls off the top and the cursor can sit below the visible area
where it cannot be seen. Browsing large portfolios is the exact use case skull2
exists for, so the list must stay usable at scale.

## What Changes

- Add a scrolling viewport to all three list levels (providers, owners, repos):
  render only the rows that fit in the terminal height and slide the window to
  keep the cursor visible.
- Keep a fixed scroll-off margin of 2 rows so the window starts scrolling
  before the cursor reaches the very edge (margin collapses at list ends).
- Add paging/navigation keys: `PgUp`/`PgDn` (screen), `Ctrl-U`/`Ctrl-D`
  (half-screen), `Home`/`End` (jump to first/last); existing `up`/`down` keep
  moving the cursor.
- Show a position indicator (e.g. `41-60 of 213`) for the current list.
- Treat unknown terminal height (`height <= 0`) as "no constraint" and render
  all rows, preserving current behavior before the first size event and in
  tests.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `tui`: adds scrolling-viewport, paging-key and position-indicator
  requirements to the browsing behavior (existing requirements unchanged).

## Impact

- `internal/tui` (model gains a scroll offset; view windows the row slice; key
  handling gains the paging keys). No new dependencies; alt-screen startup
  unchanged.
