## 1. Entry and flat scope

- [ ] 1.1 Add a `flatAll` scope flag to the Model and a synthetic "All repositories" entry pinned at the top of the provider list (`levelProviders` rendering + selection)
- [ ] 1.2 On selecting the entry, enter `levelRepos` with `flatAll` set (`selProvider=nil`, `selOwner=""`); Esc/back returns to the provider list
- [ ] 1.3 Route `visibleRepos()` to its all-providers branch when `flatAll` is set; audit `levelRepos` code paths that assume `selProvider != nil` and branch on `flatAll`
- [ ] 1.4 Set the breadcrumb/header for the flat view (e.g. "all repositories")

## 2. Rendering

- [ ] 2.1 In `flatAll`, prefix each row with the provider short code before `owner/name`, keeping alignment
- [ ] 2.2 Include the provider short in fuzzy matching so users can filter by provider token
- [ ] 2.3 Compute and render the completeness badge (`N repos · X/Y owners loaded`) from `ownersByProvider` vs `fetchedOwners`; fall back to `N repos` when no owner index exists

## 3. Full refresh-all

- [ ] 3.1 Add a refresh command in the flat view that re-fetches every owner across every provider, reusing the existing eager per-owner fetch and updating `reposByProvider` + the cache
- [ ] 3.2 Aggregate progress feedback across the multi-owner fetch; make it cancellable, leaving the prior cache intact on cancel

## 4. Tests & verification

- [ ] 4.1 TUI unit tests: entering/leaving the flat view, all-providers aggregation, provider-prefixed rows, provider-token fuzzy match
- [ ] 4.2 TUI unit tests: completeness badge counts (partial vs complete, legacy no-index fallback)
- [ ] 4.3 TUI unit tests: refresh-all orchestration across providers×owners, and cancel leaves cache intact
- [ ] 4.4 `gofmt -l .` is empty and `nix flake check` passes (build + tests + coverage gate)
- [ ] 4.5 Keep the `skull2-39es` beans checklist current as tasks complete
