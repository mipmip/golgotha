## Why

In the TUI, entering an owner that hasn't been fetched shows **no repositories**
after the fetch completes; the repos only appear after restarting `gol`. This is
a race: the fetch emits its terminal `Done` event *before* the wrapper commits
the repos to the cache, and the UI's `Done` handler reloads from the (not-yet-
written) cache — so it sees nothing. On restart the cache has been saved, so the
repos show.

## What Changes

- In the TUI progress fetcher (`internal/tui/commands.go`), commit the fetched
  owner to the cache **before** the `Done` event reaches the UI: suppress the
  inner `Done` emitted by `FetchOwner`, save the cache, then emit `Done`. The
  `Done` handler's reload-from-cache then always sees the committed repos.
- Ensure a terminal event is always emitted (emit `Failed` on cache-load error)
  so the UI never gets stuck showing "fetching…".

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `tui`: a completed lazy owner fetch displays its repositories immediately,
  without a restart.

## Impact

- `internal/tui/commands.go` (event/commit ordering in `defaultProgressFetcher`).
  No behavior change to the CLI, providers, cache format, or other TUI logic.
