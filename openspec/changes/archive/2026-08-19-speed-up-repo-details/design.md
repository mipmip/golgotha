## Context

Opening a repo's detail view runs `openDetail` → `cache.LoadDetailsOrEmpty`
(instant on hit) → on miss, `detailFetcher` calls `client.RepoDetails` then
`client.Readme` **sequentially**, then `cache.RefreshDetails` writes the per-repo
cache. Investigation (bean `skull2-egce`) established:

- No git clone in the path — two REST calls only.
- The cache works (read on open, atomic write, no TTL, symmetric paths); repeat
  opens are instant.
- glamour render is ~0.18ms — not the cost.

So the only real latency is the **cold open**: two sequential network
round-trips. This change removes/hides that latency.

## Goals / Non-Goals

**Goals:**

- Cut cold-open latency by fetching details and README concurrently.
- Make opening feel instant while browsing, via background prefetch.
- Prove the cache is hit on repeat opens with a regression test.

**Non-Goals:**

- Changing provider APIs, the cache format, or adding a TTL.
- Eliminating the very first fetch of a never-seen repo (impossible; prefetch
  just moves it earlier).
- Rendering changes (glamour is already fast).

## Decisions

### Decision: Parallelize the two detail requests

In `defaultDetailFetcher`, run `RepoDetails` and `Readme` in goroutines and wait
for both. Cold open ≈ `max(a, b)` instead of `a + b`. If `RepoDetails` fails,
fall back as today (error/existing cache). If only `Readme` fails, still return
the details with an empty README rather than failing the open.

### Decision: Debounced, cancellable prefetch on cursor settle

At `levelRepos`, cursor movement schedules a prefetch of the highlighted repo:

```
moveCursor → prefetchSeq++ ; return tea.Tick(debounce, prefetchTickMsg{seq})
prefetchTickMsg{seq}:
   if seq == current prefetchSeq                     (cursor settled)
      and repo uncached and not already prefetching  (skip warm/duplicate)
   → launch prefetch cmd (ctx with cancel stored in m.prefetchCancel)
cursor moves again → cancel the in-flight prefetch (supersede)
```

The prefetch reuses the same concurrent fetch + `RefreshDetails` path, so it
warms exactly what an on-open fetch would. Its completion message only mutates
the view if it matches the repo currently open in the detail view (otherwise it
just warmed the cache). At most one prefetch is in flight; a superseded one is
cancelled. Debounce (~200ms) keeps rapid `j`/`k` scrolling from firing a fetch
per row.

- **Why debounce + single-flight:** avoids hammering the provider API (and rate
  limits) while scrolling; only the row the cursor rests on is fetched.

### Decision: Cache-hit regression test

Seed the detail cache with `SaveDetails`, open the detail view, and assert
`detailLoaded` is true, the status reads "(cached)", and a counting-fake
`detailFetcher` is **not** called — locking in that repeat opens make no network
call.

## Risks / Trade-offs

- **[Prefetch races the open]** → the user opens a repo mid-prefetch. Mitigation:
  `openDetail` still checks the cache and, on miss, fetches as today; a completing
  prefetch that matches the open repo can populate it. No correctness issue, at
  worst a duplicate fetch.
- **[API rate limits from prefetch]** → mitigated by debounce + single-flight +
  skip-if-cached; only settled rows are fetched.
- **[Concurrency in the fetcher]** → two goroutines + a combine; keep it simple
  (wait for both, no shared mutable state) and unit-test the README-fail path.
- **[Stale prefetch result mutating the view]** → guard: apply a
  `detailLoadedMsg` only when it matches the currently-open detail repo.

## Migration Plan

1. Parallelize `defaultDetailFetcher` (+ graceful README-fail); tests.
2. Add debounced, cancellable prefetch-on-navigate wiring in the update loop;
   ignore non-matching prefetch results.
3. Add the cache-hit regression test.
4. `gofmt`, `go test ./...`, `nix flake check` (coverage gate).

## Open Questions

- Debounce duration (~200ms) — tune during implementation; not load-bearing.
