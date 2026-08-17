## Context

`internal/fetch/pages.go` emits `Done` inside `FetchOwner` (before it returns).
`defaultProgressFetcher`'s goroutine then does `MarkOwnerFetched` + `cache.Save`
*after* `FetchOwner` returns. The Bubble Tea loop can process the buffered `Done`
(→ `reloadOwnerFromCache` → `ReposFor`) before the save completes, yielding an
empty list until restart.

## Goals / Non-Goals

**Goals:** a completed lazy fetch shows its repos without restart; keep the
generic `fetch.Event` model and the commit-only-on-complete guarantee.

**Non-Goals:** changing the CLI fetch, the event types, the cache format, or the
separate self-owner over-scope issue (tracked separately).

## Decisions

- **Emit `Done` after the commit.** In the TUI fetcher, wrap emission so the
  inner `Done` from `FetchOwner` is suppressed; after a successful fetch, save
  the cache, then emit `Done` ourselves via a raw (unfiltered) emitter. This
  guarantees the UI's reload-from-cache observes the saved repos. `Failed` and
  `Canceled` still pass straight through (they must reach the UI and don't touch
  the cache).
- **Always terminate.** On a cache-load error, emit `Failed` (not a silent
  `Warning`) so the UI clears its "fetching…" state instead of hanging.
- Keep the cache as the transport (reload-from-cache) rather than threading
  repos through the event — smaller change, preserves the generic event model.

## Risks / Trade-offs

- [Save failure still yields an empty reload] → emit a `Warning` then `Done`;
  rare, and no worse than today. Not persisting a partial owner is intentional.
- [The race is timing-dependent and hard to unit-test] → covered by an ordering
  test that asserts, at the moment `Done` is delivered, the cache already
  contains the owner's repos (via an injectable fake fetcher).
