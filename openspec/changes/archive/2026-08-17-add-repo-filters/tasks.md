## 1. Visibility data

- [x] 1.1 Add `Visibility string` to the `Repo` model (public/private/internal)
- [x] 1.2 Map visibility in GitHub, Codeberg (private bool→public/private) and GitLab (visibility) clients; normalize unknown→public
- [x] 1.3 Client mapping tests for visibility

## 2. Filter model

- [x] 2.1 In-memory filter state: fork (all/only/hide), archived (all/only/hide), visibility (all/public/private/internal)
- [x] 2.2 Apply facets in the TUI `visibleRepos` pipeline (AND together, then fuzzy); defaults mirror config
- [x] 2.3 Pure filter-predicate tests (each facet, combinations, with fuzzy)

## 3. TUI wiring

- [x] 3.1 Keys to cycle facets (e.g. `f`/`a`/`v`); status line shows active facets; footer help updated
- [x] 3.2 Reset the scroll window when the filter set changes (same rule as fuzzy)
- [x] 3.3 Hint when a facet matches nothing due to fetch-time exclusion (archived/fork not cached)
- [x] 3.4 Update-driven tests (facet cycling narrows; compose with fuzzy; window reset; hint)

## 4. Verify

- [x] 4.1 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
- [x] 4.2 Update config.example.yaml note clarifying include_* is the cache superset for the facets
