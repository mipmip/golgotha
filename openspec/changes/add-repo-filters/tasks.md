## 1. Visibility data

- [ ] 1.1 Add `Visibility string` to the `Repo` model (public/private/internal)
- [ ] 1.2 Map visibility in GitHub, Codeberg (private bool→public/private) and GitLab (visibility) clients; normalize unknown→public
- [ ] 1.3 Client mapping tests for visibility

## 2. Filter model

- [ ] 2.1 In-memory filter state: fork (all/only/hide), archived (all/only/hide), visibility (all/public/private/internal)
- [ ] 2.2 Apply facets in the TUI `visibleRepos` pipeline (AND together, then fuzzy); defaults mirror config
- [ ] 2.3 Pure filter-predicate tests (each facet, combinations, with fuzzy)

## 3. TUI wiring

- [ ] 3.1 Keys to cycle facets (e.g. `f`/`a`/`v`); status line shows active facets; footer help updated
- [ ] 3.2 Reset the scroll window when the filter set changes (same rule as fuzzy)
- [ ] 3.3 Hint when a facet matches nothing due to fetch-time exclusion (archived/fork not cached)
- [ ] 3.4 Update-driven tests (facet cycling narrows; compose with fuzzy; window reset; hint)

## 4. Verify

- [ ] 4.1 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
- [ ] 4.2 Update config.example.yaml note clarifying include_* is the cache superset for the facets
