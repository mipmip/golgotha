---
# skull2-egce
title: show details take long
status: completed
type: bug
priority: normal
created_at: 2026-08-19T11:00:51Z
updated_at: 2026-08-19T22:33:51Z
parent: skull2-ok4c
---

I looks like it's using git to clone the whole repo. And it look like it does not use cache at all.



## Diagnosis (explore + measure, 2026-08-19)

Both hypotheses in the report are disproven by tracing + measuring:
- **NOT git clone:** the detail path is `openDetail → detailFetcher → RepoDetails
  (GET /repos/{o}/{n}) + Readme (GET /repos/{o}/{n}/readme)`. Two REST calls; the
  README is base64 in one response. No git, no tree walk.
- **Cache DOES work:** openDetail reads `cache.LoadDetailsOrEmpty` first (instant,
  no network) and writes via `RefreshDetails`/`SaveDetails` (atomic, no TTL).
  Save/load paths are symmetric (`details/<provider>/<owner>__<repo>.json`).
  Repeat opens of the same repo are instant. No miss bug found.
- **glamour is NOT the bottleneck:** measured `NewTermRenderer+Render` ≈ 180µs
  (fresh) / 100µs (reused). Sub-millisecond.

**Real cause:** a COLD (uncached) open does TWO SEQUENTIAL network round-trips
(RepoDetails then Readme). Browsing many never-opened repos = a cold fetch each
time, which feels like "always slow / no cache."

## Fix (decided): parallelize + prefetch-on-navigate

1. **Parallelize** RepoDetails + Readme (concurrent) in the detail fetcher so a
   cold open ≈ max(a,b) instead of a+b. Consider showing details even if the
   README call fails (degrade gracefully).
2. **Prefetch on navigate:** when the cursor settles on a repo at levelRepos
   (debounced ~200ms, cancel on further movement), warm that repo's detail cache
   in the background if uncached and not already fetching. Enter then hits the
   warm cache → instant. One prefetch at a time; cancel superseded ones; the
   emitted msg is ignored unless it matches the currently-open detail.
3. **Cache-hit regression test:** seed the detail cache (SaveDetails) for a repo,
   openDetail → assert detailLoaded from cache, status "(cached)", and the
   (counting-fake) detailFetcher is NOT called — proving repeat opens make no
   network call.

Honest caveat: the very first open of a never-seen repo always needs one fetch;
prefetch just moves it to idle cursor time. Change name: `speed-up-repo-details`.
Capabilities: repo-details (parallel fetch + prefetch) and tui (navigate-prefetch
wiring). No config/schema change.



## Summary of Changes

Sped up the detail view. defaultDetailFetcher refactored to detailFetcherWith(build)
and now fetches RepoDetails + Readme concurrently (goroutines + WaitGroup); a
README failure is non-fatal (details shown with empty README). Added
navigate-prefetch: cursor-movement keys at levelRepos call schedulePrefetch
(debounced 200ms via tea.Tick with a seq counter); handlePrefetchTick warms the
highlighted repo's cache in the background if uncached (cancellable, superseded
prefetches cancelled), reusing detailFetcher — whose result handleDetailLoaded
already ignores unless it matches the open detail. Confirmed via a cache-hit
regression test that a seeded cache makes openDetail hit disk with no fetcher
call. (Diagnosis: not git clone, cache works, glamour ~0.18ms; the cost was two
sequential cold-open round-trips.) Coverage: tui 80.2%, overall 80.9%. Shipped as
2026-08-19-speed-up-repo-details (commit 541bec33).
