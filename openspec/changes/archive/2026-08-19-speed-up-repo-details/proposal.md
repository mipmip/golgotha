## Why

Opening a repository's detail view feels slow (bean `skull2-egce`). Investigation
disproved the two guesses in the report: there is **no git clone** in the detail
path (it is two REST calls — `/repos/{o}/{n}` and `/repos/{o}/{n}/readme`), the
per-repo detail **cache does work** (read on open, written atomically, no TTL,
symmetric paths), and glamour rendering is **not** the cost (measured ~0.18ms).

The real cause is a **cold open**: the details and README are fetched **one after
the other**, so a first open pays two sequential network round-trips. Browsing
many never-opened repos makes every open a cold fetch — which reads as "always
slow / no cache."

## What Changes

- **Parallelize** the two detail requests (`RepoDetails` and `Readme`) so a cold
  open takes about `max(a, b)` instead of `a + b`. If the README call fails but
  details succeed, still show the details (degrade gracefully).
- **Prefetch on navigate**: when the cursor settles on a repository row
  (debounced, cancelled on further movement), warm that repo's detail cache in
  the background if it is not already cached or being fetched — so pressing Enter
  is usually instant. At most one prefetch runs at a time; a superseded prefetch
  is cancelled; its result only updates the view if it matches the repo currently
  open.
- **Cache-hit regression test**: seed the detail cache and assert a subsequent
  open is served from disk with no network fetch.

No config, schema, or provider-API changes. The very first open of a never-seen
repo still needs one fetch; prefetch just moves it to idle cursor time.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `repo-details`: the detail fetch runs its two requests concurrently and shows
  details even if the README fetch fails; details may be prefetched to warm the
  cache before the view is opened.
- `tui`: navigating the repository list prefetches the highlighted repo's details
  in the background (debounced, cancellable) so opening is instant when warm.

## Impact

- **TUI**: `internal/tui` — `defaultDetailFetcher` fetches details+README
  concurrently; a debounced prefetch is triggered on cursor settle at the repos
  level, cancellable and de-duplicated, reusing the same fetch/caching path; the
  prefetch result is ignored unless it matches the open detail.
- **Tests**: concurrent-fetch behavior (both calls issued, README-fail still
  shows details); prefetch warms the cache and Enter is a cache hit; a cache-hit
  regression test (seeded cache → open → no fetch).
- No new dependencies; provider `RepoDetails`/`Readme` and the cache format are
  unchanged.
