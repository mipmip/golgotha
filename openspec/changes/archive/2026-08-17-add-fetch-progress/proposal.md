## Why

Fetching repositories can be slow — a big org has many paginated pages, and a
`sync` across many orgs is slower still — yet the app gives no feedback. In the
TUI, entering an owner shows a silent loading line; on the CLI a long refresh
looks hung. Users need to see what the app is doing, and long fetches should be
cancelable and faster.

## What Changes

- Introduce one progress-event model emitted by the fetch pipeline, consumed by
  two frontends: the TUI (spinner → bar → textual description) and the CLI
  (`sync`/`refresh` printed progress lines). Tests assert on the event stream.
- Make repository fetching page-aware and bounded-parallel: fetch page 1 to
  learn the total, then fan out the remaining pages with a bounded worker pool
  (cap 6). The CLI additionally fetches owners with a bounded pool.
- Show progress in the TUI on lazy per-owner fetch (owner entry): a spinner
  during discovery of totals, then a determinate bar and a "fetching <owner>
  page i/n — N repos" line.
- Support cancellation via context (`Esc` cancels the current fetch). Guarantee
  no cache corruption: an owner is committed to the cache only on a complete
  fetch; cancel or any page failure leaves the owner unfetched and the cache
  untouched. Partial results are shown as transient, never cached as complete.

## Capabilities

### New Capabilities
- `fetch-progress`: the progress-event stream contract (event kinds, emission,
  bounded-parallel fetch, cancellation semantics) shared by all fetch callers.

### Modified Capabilities
- `provider-clients`: page-aware, bounded-parallel fetch that emits progress
  events.
- `tui`: progress UI (spinner/bar/text) on owner entry; cancel the current
  fetch; transient partial results.
- `sync`: `sync`/`refresh` consume progress events and print them; owners fetched
  bounded-parallel.
- `repo-cache`: commit an owner only on complete fetch (all-or-nothing per
  owner).

## Impact

- `internal/provider` (fetch refactor), `internal/tui`, `internal/syncer` +
  `cmd/skull2`, `internal/cache`. Uses `bubbles/spinner` and `bubbles/progress`
  (already a dependency). No new dependencies expected.
