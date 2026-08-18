---
# skull2-ci84
title: 'TUI: lazy-fetched org shows no repos until restart'
status: completed
type: bug
priority: high
created_at: 2026-08-17T21:50:46Z
updated_at: 2026-08-17T21:52:43Z
parent: skull2-qati
---

Race: fetch Done emitted before cache save; UI reload sees empty. Fixed by committing cache before emitting Done.

## Summary of Changes

Shipped via fix-lazy-fetch-race (commit dbfe4a69). Root cause: FetchOwner emits its terminal Done before the TUI wrapper commits the cache, so the Done handler's reload-from-cache saw an empty cache (repos only appeared after restart). Fix: in progressFetcherWith, suppress the inner Done, commit the cache (MarkOwnerFetched + Save), then emit Done; emit Failed on cache-load error so the UI never hangs in "fetching". Added an ordering test (fake OwnerFetcher) asserting the cache is committed before Done and exactly one Done is delivered.
