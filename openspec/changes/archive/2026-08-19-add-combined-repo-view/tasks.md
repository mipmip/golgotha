## 1. Entry and flat scope

- [x] 1.1 Add a `flatAll` scope flag to the Model and a synthetic "All repositories" entry appended to the provider list (`levelProviders` rendering + selection). Placed last to keep provider cursor indices stable.
- [x] 1.2 On selecting the entry, enter `levelRepos` with `flatAll` set (`selProvider=nil`, `selOwner=""`); Esc/back returns to the provider list
- [x] 1.3 Route `visibleRepos()` to its all-providers branch when `flatAll` is set; audit `levelRepos` code paths that assume `selProvider != nil` and branch on `flatAll`
- [x] 1.4 Set the breadcrumb/header for the flat view (e.g. "all repositories")

## 2. Rendering

- [x] 2.1 In `flatAll`, prefix each row with the provider short code before `owner/name`, keeping alignment
- [x] 2.2 Include the provider short in fuzzy matching so users can filter by provider token
- [x] 2.3 Compute and render the completeness badge (`N repos · X/Y owners loaded`) from `ownersByProvider` vs `fetchedOwners`; fall back to `N repos` when no owner index exists

## 3. Full refresh-all

- [x] 3.1 Add a refresh command in the flat view that re-fetches every provider (each a full per-owner sweep), reusing the existing per-provider `refresher` and updating `reposByProvider` + the cache via `refreshResultMsg`
- [x] 3.2 Each provider's refresh is atomic (commit-only-on-complete via the existing refresher), so a failed/interrupted provider keeps its prior cache; progress shown via the status line

## 4. Tests & verification

- [x] 4.1 TUI unit tests: entering/leaving the flat view, all-providers aggregation, provider-prefixed rows, provider-token fuzzy match
- [x] 4.2 TUI unit tests: completeness badge counts (partial vs complete, legacy no-index fallback)
- [x] 4.3 TUI unit tests: refresh-all orchestration across providers×owners, and cancel leaves cache intact
- [x] 4.4 `gofmt -l .` is empty and `nix flake check` passes (build + tests + coverage gate)
- [x] 4.5 Keep the `skull2-39es` beans checklist current as tasks complete
