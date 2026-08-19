## 1. Parallelize the cold detail fetch

- [ ] 1.1 In `defaultDetailFetcher`, run `RepoDetails` and `Readme` concurrently (goroutines + wait) instead of sequentially, then `RefreshDetails`
- [ ] 1.2 Degrade gracefully: if `RepoDetails` fails, fall back as today; if only `Readme` fails, show details with an empty README
- [ ] 1.3 TUI tests: both requests are issued concurrently (fake client records calls); README-only failure still yields a shown detail with empty README

## 2. Prefetch on navigate

- [ ] 2.1 Add prefetch state to the Model (a debounce sequence counter + a `prefetchCancel context.CancelFunc`)
- [ ] 2.2 On cursor movement at `levelRepos`, bump the sequence and schedule a debounced `prefetchTickMsg{seq}` (~200ms) via `tea.Tick`
- [ ] 2.3 On `prefetchTickMsg`: if the sequence still current, the highlighted repo is uncached, and none is in flight, launch a cancellable prefetch that reuses the concurrent fetch + `RefreshDetails` path; cancel any superseded in-flight prefetch on further movement
- [ ] 2.4 Ignore/skip prefetch for already-cached repos; only apply a completed prefetch's `detailLoadedMsg` to the view when it matches the currently-open detail repo
- [ ] 2.5 TUI tests: settling on an uncached repo warms the cache so a subsequent open is a cache hit; rapid movement debounces to a single prefetch; a non-matching prefetch result does not change the visible view

## 3. Cache-hit regression test

- [ ] 3.1 Seed the detail cache with `SaveDetails`, open the detail view, and assert `detailLoaded` + "(cached)" status and that a counting-fake `detailFetcher` is not called

## 4. Verification

- [ ] 4.1 `gofmt -l .` empty; `go test ./...` passes; `nix flake check` passes (coverage gate)
- [ ] 4.2 Keep the `skull2-egce` beans checklist current as tasks complete
